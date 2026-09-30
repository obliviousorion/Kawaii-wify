package cli

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/obliviousorion/kawaii-wify/internal/ipc"
	"github.com/obliviousorion/kawaii-wify/internal/updater"
	"github.com/spf13/cobra"
)

var (
	updateCheckOnly bool
	updateForce     bool
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for updates and update kawaii-wify in place",
	Long:  "Queries GitHub Releases for newer releases of kawaii-wify, downloads the appropriate binary for your platform, and updates your current installation safely in place.",
	Run:   runUpdate,
}

func init() {
	updateCmd.Flags().BoolVar(&updateCheckOnly, "check", false, "Check for newer version without downloading")
	updateCmd.Flags().BoolVarP(&updateForce, "force", "f", false, "Force re-downloading even if already on the latest version")
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) {
	fmt.Printf("Checking for updates (current version: %s)...\n", AppVersion)

	checker := updater.NewChecker(AppVersion)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rel, isNewer, err := checker.CheckLatest(ctx)
	if err != nil && !updateForce {
		if err == updater.ErrDevVersion {
			fmt.Println("Running a development build. Run with --force to overwrite with the latest GitHub release.")
			return
		}
		fmt.Fprintf(os.Stderr, "Error checking for updates: %v\n", err)
		os.Exit(1)
	}

	if !isNewer && !updateForce {
		fmt.Printf("Kawaii-Wify is already up to date (%s)!\n", AppVersion)
		return
	}

	targetTag := ""
	if rel != nil {
		targetTag = rel.TagName
	}
	if targetTag == "" {
		targetTag = "latest"
	}
	fmt.Printf("Found available release: %s\n", targetTag)

	if updateCheckOnly {
		fmt.Println("Run 'kawaii-wify update' to install it.")
		return
	}

	// 1. Identify matching binary asset for current OS/Arch
	expectedAssetName := fmt.Sprintf("kawaii-wify-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		expectedAssetName += ".exe"
	}

	var downloadURL string
	if rel != nil {
		for _, asset := range rel.Assets {
			if asset.Name == expectedAssetName {
				downloadURL = asset.BrowserDownloadURL
				break
			}
		}
	}

	if downloadURL == "" {
		fmt.Fprintf(os.Stderr, "Error: Release %s does not contain prebuilt binary asset '%s' for your system.\n", targetTag, expectedAssetName)
		os.Exit(1)
	}

	// 2. Resolve target executable path
	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error determining executable path: %v\n", err)
		os.Exit(1)
	}
	realExePath, err := filepath.EvalSymlinks(exePath)
	if err == nil && realExePath != "" {
		exePath = realExePath
	}

	installDir := filepath.Dir(exePath)
	// Check directory write permission
	testTmpFile := filepath.Join(installDir, fmt.Sprintf(".perm_test_%d", time.Now().UnixNano()))
	if f, err := os.OpenFile(testTmpFile, os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Insufficient write permissions in '%s'. Please run as Administrator or with sudo.\n", installDir)
		os.Exit(1)
	} else {
		f.Close()
		_ = os.Remove(testTmpFile)
	}

	// 3. Download to sibling temporary file (.tmp)
	tmpPath := exePath + ".tmp"
	fmt.Printf("Downloading %s from GitHub...\n", expectedAssetName)
	if err := downloadFile(ctx, downloadURL, tmpPath); err != nil {
		_ = os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
		os.Exit(1)
	}

	// Ensure executable permissions on Unix
	if runtime.GOOS != "windows" {
		_ = os.Chmod(tmpPath, 0755)
	}

	// 4. Daemon coordination: If background daemon is running, gracefully stop it
	var restartDaemon bool
	client, err := ipc.NewClient()
	if err == nil && client != nil {
		status, sErr := client.GetStatus()
		if sErr == nil && status != nil {
			fmt.Println("Active background daemon detected. Stopping daemon temporarily for update...")
			_, _ = client.Stop()
			restartDaemon = true
			time.Sleep(1 * time.Second)
		}
		client.Close()
	}

	// 5. In-place binary swap with Windows locking protection and rollback
	oldPath := exePath + ".old"
	_ = os.Remove(oldPath) // Remove any leftover .old from previous updates

	if err := os.Rename(exePath, oldPath); err != nil {
		_ = os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Failed to prepare executable for replacement: %v\n", err)
		os.Exit(1)
	}

	if err := os.Rename(tmpPath, exePath); err != nil {
		// Roll back
		_ = os.Rename(oldPath, exePath)
		_ = os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Failed to install new executable (rolled back): %v\n", err)
		os.Exit(1)
	}

	// Best-effort cleanup of .old binary
	_ = os.Remove(oldPath)

	fmt.Printf("Successfully updated Kawaii-Wify to %s!\n", targetTag)

	// 6. Resume daemon if it was running
	if restartDaemon {
		fmt.Println("Restarting background daemon with updated binary...")
		subCmd := exec.Command(exePath, "daemon")
		detachCmd(subCmd)
		subCmd.Stdin = nil
		subCmd.Stdout = nil
		subCmd.Stderr = nil
		if err := subCmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to restart daemon automatically: %v\nRun 'kawaii-wify start' to start it.\n", err)
		} else {
			fmt.Println("Daemon restarted successfully.")
		}
	}
}

func downloadFile(ctx context.Context, url string, destPath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Kawaii-Wify-Updater")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error %d", resp.StatusCode)
	}

	out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
