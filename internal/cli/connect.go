package cli

import (
	"log"

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
		log.Fatalf("[FATAL] %v", err)
	}
	defer client.Close()

	log.Println("[INFO] Requesting network connection via daemon...")

	resp, err := client.Connect()
	if err != nil {
		log.Fatalf("[ERROR] Connect request failed: %v", err)
	}

	if !resp.Success {
		log.Fatalf("[ERROR] %s", resp.Message)
	}

	log.Printf("[SUCCESS] %s", resp.Message)
}