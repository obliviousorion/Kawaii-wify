package cli

import (
	"log"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kawaii-wify",
	Short: "FortiGate (FortiOS) Captive Portal login assistant",
	Long: `A CLI tool to manage your WiFi portal login.
	Features include: Captive portal detection, automated login, and keep-alive checks.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatalf("[ERROR] Failed executing root command: %v", err)
	}
}
