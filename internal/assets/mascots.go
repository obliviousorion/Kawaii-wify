package assets

import (
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

//go:embed mascots/*.jpg
var mascotFS embed.FS

const (
	MascotOnline     = "wify_mascot_online.jpg"
	MascotOffline    = "wify_mascot_offline.jpg"
	MascotSecurity   = "wify_mascot_security.jpg"
	MascotAuthFailed = "wify_mascot_auth_failed.jpg"
	MascotUpdate     = "wify_mascot_update.jpg"
)

// GetMascotPath extracts the embedded mascot into the system app cache directory
// (if not already extracted) and returns the absolute file path on disk.
// This is required because native OS toast notification APIs (Windows WinRT, Linux notify-send)
// require absolute filesystem paths rather than in-memory byte slices.
func GetMascotPath(filename string) (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}

	targetDir := filepath.Join(cacheDir, "kawaii-wify", "assets")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create mascot cache dir: %w", err)
	}

	targetPath := filepath.Join(targetDir, filename)

	// Check if already extracted and non-empty
	if info, err := os.Stat(targetPath); err == nil && info.Size() > 0 {
		return targetPath, nil
	}

	// Read from embedded FS under mascots/
	embeddedPath := "mascots/" + filename
	srcFile, err := mascotFS.Open(embeddedPath)
	if err != nil {
		return "", fmt.Errorf("mascot %q not found in embedded assets: %w", filename, err)
	}
	defer srcFile.Close()

	// Write to cache file
	dstFile, err := os.Create(targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to create cache file %q: %w", targetPath, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return "", fmt.Errorf("failed to write mascot cache: %w", err)
	}

	return targetPath, nil
}
