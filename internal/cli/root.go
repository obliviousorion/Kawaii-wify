package cli

import (
	"log"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kawaii-wify",
	Short: "FortiGate (FortiOS) Captive Portal login assistant",
	Long:  "A lightweight background session manager and automated login daemon for FortiOS (FortiGate) captive portals, specifically designed for campus WLAN environments like BITS Pilani.\nFeatures include: Captive portal detection, automated login, and keep-alive checks.",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatalf("[ERROR] Failed executing root command: %v", err)
	}
}
