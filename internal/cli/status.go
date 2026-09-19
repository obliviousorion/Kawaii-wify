package cli

import (
	"fmt"
	"os"

	"github.com/obliviousorion/kawaii-wify/internal/ipc"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current daemon connectivity and telemetry",
	Long:  "Queries the running kawaii-wify background daemon over IPC to display active network state, uptime, and session details.",
	Run:   runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) {
	client, err := ipc.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	status, err := client.GetStatus()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error querying daemon: %v\n", err)
		os.Exit(1)
	}

	lastProbeStr := "In progress..."
	if !status.LastProbe.IsZero() {
		lastProbeStr = status.LastProbe.Format("15:04:05")
	}

	sessionTokenStr := status.SessionToken
	if sessionTokenStr == "" {
		sessionTokenStr = "None"
	}

	fmt.Println("kawaii-wify Daemon Status")
	fmt.Println("─────────────────────────")
	fmt.Printf("  State:         %s\n", status.State)
	fmt.Printf("  User:          %s\n", status.Username)
	fmt.Printf("  Paused:        %t\n", status.Paused)
	fmt.Printf("  Uptime:        %s\n", status.Uptime)
	fmt.Printf("  Last Probe:    %s\n", lastProbeStr)
	fmt.Printf("  Session:       %s\n", sessionTokenStr)
}