package cli

import (
	"fmt"
	"os"

	"github.com/obliviousorion/kawaii-wify/internal/autostart"
	"github.com/spf13/cobra"
)

var autostartCmd = &cobra.Command{
	Use:   "autostart",
	Short: "Manage automatic background startup on system boot or login",
	Long:  "Configure whether the kawaii-wify daemon starts automatically in the background when the user logs into the desktop.",
	Run:   runAutostartStatus,
}

var autostartEnableCmd = &cobra.Command{
	Use:     "enable",
	Aliases: []string{"on"},
	Short:   "Enable background autostart on system boot or login",
	Run:     runAutostartEnable,
}

var autostartDisableCmd = &cobra.Command{
	Use:     "disable",
	Aliases: []string{"off"},
	Short:   "Disable background autostart on system boot or login",
	Run:     runAutostartDisable,
}

var autostartStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check whether background autostart is enabled",
	Run:   runAutostartStatus,
}

func init() {
	autostartCmd.AddCommand(autostartEnableCmd)
	autostartCmd.AddCommand(autostartDisableCmd)
	autostartCmd.AddCommand(autostartStatusCmd)
	rootCmd.AddCommand(autostartCmd)
}

func runAutostartEnable(cmd *cobra.Command, args []string) {
	mgr, err := autostart.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: autostart is not supported on this platform: %v\n", err)
		os.Exit(1)
	}

	if err := mgr.Enable(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to enable autostart: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Autostart enabled successfully. kawaii-wify daemon will run in the background upon login.")
}

func runAutostartDisable(cmd *cobra.Command, args []string) {
	mgr, err := autostart.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: autostart is not supported on this platform: %v\n", err)
		os.Exit(1)
	}

	if err := mgr.Disable(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to disable autostart: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Autostart disabled successfully.")
}

func runAutostartStatus(cmd *cobra.Command, args []string) {
	mgr, err := autostart.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: autostart is not supported on this platform: %v\n", err)
		os.Exit(1)
	}

	enabled, err := mgr.IsEnabled()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to check autostart status: %v\n", err)
		os.Exit(1)
	}

	if enabled {
		fmt.Println("Autostart: Enabled")
		fmt.Println("To disable, run: kawaii-wify autostart disable")
	} else {
		fmt.Println("Autostart: Disabled")
		fmt.Println("To enable, run: kawaii-wify autostart enable")
	}
}
