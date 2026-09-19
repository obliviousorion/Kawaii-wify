package cli

import (
	"log"

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
		log.Fatalf("[FATAL] %v", err)
	}

	log.Printf("[INFO] Disconnect request engaged: %v", err)

	resp, err := client.Disconnect()
	if err != nil {
		log.Fatalf("[ERROR] Disconnect request failed: %v", err)
	}

	if !resp.Success {
		log.Fatalf("[ERROR] %s", resp.Message)
	}

	log.Printf("[SUCCESS] %s", resp.Message)
}