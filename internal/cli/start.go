package cli

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/obliviousorion/kawaii-wify/internal/ipc"
	"github.com/spf13/cobra"
)

var (
	startUser            string
	startGateway         string
	startKeepalive       bool
	startNoKeepalive     bool
	startAutoConnect     bool
	startNoAutoConnect   bool
	startPaused          bool
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the background daemon process detached from the terminal",
	Long:  "Spawns the kawaii-wify daemon in the background as a detached process. The daemon continues running even after the terminal is closed.",
	Run:   runStart,
}

func init() {
	startCmd.Flags().StringVarP(&startUser, "user", "u", "", "Set or override kawaii-wify username")
	startCmd.Flags().StringVar(&startGateway, "gateway", "", "Override FortiOS captive portal gateway endpoint")
	startCmd.Flags().BoolVar(&startKeepalive, "keepalive", true, "Enable periodic keepalive pings")
	startCmd.Flags().BoolVar(&startNoKeepalive, "no-keepalive", false, "Disable periodic keepalive pings")
	startCmd.Flags().BoolVar(&startAutoConnect, "auto-connect", true, "Automatically connect and authenticate on launch")
	startCmd.Flags().BoolVar(&startNoAutoConnect, "no-auto-connect", false, "Start daemon without automatically connecting")
	startCmd.Flags().BoolVarP(&startPaused, "paused", "p", false, "Start daemon in paused state (alias for --no-auto-connect)")
	rootCmd.AddCommand(startCmd)
}

func runStart(cmd *cobra.Command, args []string) {
	// 1. Check if daemon is already running
	if client, err := ipc.NewClient(); err == nil {
		defer client.Close()
		status, err := client.GetStatus()
		if err == nil {
			fmt.Fprintf(os.Stderr, "Daemon is already running (User: %s, State: %s).\nUse 'kawaii-wify status' or 'kawaii-wify stop'.\n", status.Username, status.State)
			return
		}
		fmt.Fprintln(os.Stderr, "Daemon is already running. Use 'kawaii-wify status' or 'kawaii-wify stop'.")
		return
	}

	// 2. Resolve executable path
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not resolve executable path: %v\n", err)
		os.Exit(1)
	}

	// 3. Assemble arguments to forward to 'daemon'
	daemonArgs := []string{"daemon"}
	if cmd.Flags().Changed("user") {
		daemonArgs = append(daemonArgs, "-u", startUser)
	}
	if cmd.Flags().Changed("gateway") {
		daemonArgs = append(daemonArgs, "--gateway", startGateway)
	}
	if cmd.Flags().Changed("no-keepalive") {
		daemonArgs = append(daemonArgs, "--no-keepalive")
	} else if cmd.Flags().Changed("keepalive") {
		daemonArgs = append(daemonArgs, "--keepalive")
	}
	if cmd.Flags().Changed("no-auto-connect") {
		daemonArgs = append(daemonArgs, "--no-auto-connect")
	} else if cmd.Flags().Changed("paused") {
		daemonArgs = append(daemonArgs, "-p")
	} else if cmd.Flags().Changed("auto-connect") {
		daemonArgs = append(daemonArgs, "--auto-connect")
	}

	// 4. Configure detached background process
	subCmd := exec.Command(exe, daemonArgs...)
	detachCmd(subCmd)
	subCmd.Stdin = nil
	subCmd.Stdout = nil
	subCmd.Stderr = nil

	if err := subCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to launch background daemon: %v\n", err)
		os.Exit(1)
	}

	// 5. Handshake: wait briefly for the daemon IPC endpoint to become reachable
	pid := subCmd.Process.Pid
	ready := false
	for i := 0; i < 20; i++ {
		time.Sleep(100 * time.Millisecond)
		if client, err := ipc.NewClient(); err == nil {
			client.Close()
			ready = true
			break
		}
	}

	if !ready {
		fmt.Fprintf(os.Stderr, "Warning: daemon spawned (PID: %d), but IPC endpoint did not respond yet.\nCheck 'kawaii-wify logs' for startup status.\n", pid)
		return
	}

	fmt.Printf("✓ kawaii-wify daemon started in background (PID: %d).\n", pid)
	fmt.Println("  Status: kawaii-wify status")
	fmt.Println("  Logs:   kawaii-wify logs")
	fmt.Println("  Stop:   kawaii-wify stop")
}
