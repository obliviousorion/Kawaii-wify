package cli

import (
	"fmt"
	"os"

	"github.com/obliviousorion/kawaii-wify/internal/ipc"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the background daemon process",
	Long:  "Commands the background kawaii-wify daemon over IPC to gracefully disconnect and terminate.",
	Run:   runStop,
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func runStop(cmd *cobra.Command, args []string) {
	client, err := ipc.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	resp, err := client.Stop()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: stop request failed: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Fprintf(os.Stderr, "✕ %s\n", resp.Message)
		os.Exit(1)
	}

	fmt.Printf("✓ %s\n", resp.Message)
}
