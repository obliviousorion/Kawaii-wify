package cli

import (
	"fmt"
	"os"

	"github.com/obliviousorion/kawaii-wify/internal/ipc"
	"github.com/spf13/cobra"
)

var disconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "Disconnect active session and pause background monitoring",
	Long:  "Commands the background kawaii-wify daemon over IPC to clear the active session and pause automatic reconnects until explicitly resumed.",
	Run:   runDisconnect,
}

func init() {
	rootCmd.AddCommand(disconnectCmd)
}

func runDisconnect(cmd *cobra.Command, args []string) {
	client, err := ipc.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	resp, err := client.Disconnect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: disconnect request failed: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Fprintf(os.Stderr, "✕ %s\n", resp.Message)
		os.Exit(1)
	}

	fmt.Printf("✓ %s\n", resp.Message)
}