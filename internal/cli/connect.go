package cli

import (
	"fmt"
	"os"

	"github.com/obliviousorion/kawaii-wify/internal/ipc"
	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Trigger an immediate network probe and login attempt",
	Long:  "Commands the background kawaii-wify daemon over IPC to unpause and immediately authenticate with the captive portal.",
	Run:   runConnect,
}

func init() {
	rootCmd.AddCommand(connectCmd)
}

func runConnect(cmd *cobra.Command, args []string) {
	client, err := ipc.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("Connecting to network via daemon...")

	resp, err := client.Connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: connect request failed: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Fprintf(os.Stderr, "✕ %s\n", resp.Message)
		os.Exit(1)
	}

	fmt.Printf("✓ %s\n", resp.Message)
}