package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/obliviousorion/kawaii-wify/internal/config"
	"github.com/obliviousorion/kawaii-wify/internal/notify"
	"github.com/spf13/cobra"
)

var notifyCmd = &cobra.Command{
	Use:   "notify",
	Short: "Manage desktop toast notifications and alerts",
	Long:  "Configure and test desktop OS toast notifications and mascot alerts across security halts, auth errors, and updates.",
	Run:   runNotifyStatus,
}

var notifyEnableCmd = &cobra.Command{
	Use:     "enable",
	Aliases: []string{"on"},
	Short:   "Enable desktop OS toast notifications",
	Run:     runNotifyEnable,
}

var notifyDisableCmd = &cobra.Command{
	Use:     "disable",
	Aliases: []string{"off"},
	Short:   "Disable desktop OS toast notifications",
	Run:     runNotifyDisable,
}

var notifyStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current desktop notification setting",
	Run:   runNotifyStatus,
}

var notifyTestCmd = &cobra.Command{
	Use:   "test [category]",
	Short: "Send a test toast notification with mascot avatar",
	Long:  "Send a test notification to verify OS toast integration. Supported categories: security, auth_failed, update, online, offline (default: security).",
	Args:  cobra.MaximumNArgs(1),
	Run:   runNotifyTest,
}

func init() {
	notifyCmd.AddCommand(notifyEnableCmd)
	notifyCmd.AddCommand(notifyDisableCmd)
	notifyCmd.AddCommand(notifyStatusCmd)
	notifyCmd.AddCommand(notifyTestCmd)
	rootCmd.AddCommand(notifyCmd)
}

func runNotifyStatus(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	status := "enabled"
	if !cfg.IsNotificationsEnabled() {
		status = "disabled"
	}

	fmt.Printf("Desktop Notifications: %s\n", status)
	fmt.Println("To toggle: 'kawaii-wify notify enable' or 'kawaii-wify notify disable'")
	fmt.Println("To test:   'kawaii-wify notify test'")
}

func runNotifyEnable(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}

	cfg.SetNotificationsEnabled(true)
	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Desktop notifications enabled successfully.")
	notifyDaemonConfigReload()
}

func runNotifyDisable(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}

	cfg.SetNotificationsEnabled(false)
	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Desktop notifications disabled.")
	notifyDaemonConfigReload()
}

func runNotifyTest(cmd *cobra.Command, args []string) {
	category := notify.CategorySecurity
	if len(args) > 0 {
		switch strings.ToLower(args[0]) {
		case "security", "sec", "halt":
			category = notify.CategorySecurity
		case "auth_failed", "auth", "failed", "denied":
			category = notify.CategoryAuthFailed
		case "update", "new":
			category = notify.CategoryUpdate
		case "online", "connected":
			category = notify.CategoryOnline
		case "offline", "disconnected":
			category = notify.CategoryOffline
		default:
			fmt.Fprintf(os.Stderr, "Unknown category %q. Choose from: security, auth_failed, update, online, offline\n", args[0])
			os.Exit(1)
		}
	}

	var title, msg string
	switch category {
	case notify.CategorySecurity:
		title = "Kawaii-Wify: Security Alert"
		msg = "Gateway certificate mismatch detected. Auto-login halted for safety."
	case notify.CategoryAuthFailed:
		title = "Kawaii-Wify: Login Failed"
		msg = "Campus firewall rejected credentials. Please update your password."
	case notify.CategoryUpdate:
		title = "Kawaii-Wify: Update Available"
		msg = "Version v0.2.0 is out with security enhancements and improvements!"
	case notify.CategoryOnline:
		title = "Kawaii-Wify: Connected"
		msg = "Connection to campus network established."
	case notify.CategoryOffline:
		title = "Kawaii-Wify: Disconnected"
		msg = "Network offline or daemon paused."
	}

	fmt.Printf("Dispatching test notification [%s]...\n", category)
	notifier := notify.NewNotifier()
	if err := notifier.Send(notify.Notification{
		Title:    title,
		Message:  msg,
		Category: category,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send notification: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Notification dispatched successfully! Check your desktop notifications.")
}
