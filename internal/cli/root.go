package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kawaii-wify",
	Short: "FortiGate (FortiOS) Captive Portal login assistant",
	Long:  "A lightweight background session manager and automated login daemon for FortiOS (FortiGate) captive portals, specifically designed for campus WLAN environments like BITS Pilani.\nFeatures include: Captive portal detection, automated login, and keep-alive checks.",
}

var AppVersion = "dev"

func Execute(version string) {
	if version != "" {
		AppVersion = version
	}
	rootCmd.Version = AppVersion
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
