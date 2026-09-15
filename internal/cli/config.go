package cli

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/obliviousorion/kawaii-wify/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View and manage local settings",
	Long:  "View and update non-sensitive configuration settings such as check interval and default username.",
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Display one or all configuration values",
	Args:  cobra.MaximumNArgs(1),
	Run:   runConfigGet,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Update a configuration setting",
	Args:  cobra.ExactArgs(2),
	Run:   runConfigSet,
}

func runConfigGet(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("[WARN] Failed to load config, showing defaults: %v", err)
		cfg = &config.Config{CheckInterval: "10s"}
	}

	if len(args) == 0 {
		fmt.Printf("username: %s\n", cfg.Username)
		fmt.Printf("check_interval: %s\n", cfg.CheckInterval)
		return
	}

	key := strings.ToLower(args[0])
	switch key {
	case "username":
		fmt.Println(cfg.Username)
	case "check_interval", "interval":
		fmt.Println(cfg.CheckInterval)
	default:
		log.Fatalf("[ERROR] Unknown configuration key '%s'. Supported keys: username, check_interval", args[0])
	}
}

func runConfigSet(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("[WARN] Failed to load config, initializing fresh: %v", err)
		cfg = &config.Config{CheckInterval: "10s"}
	}

	key := strings.ToLower(args[0])
	val := args[1]

	switch key {
	case "check_interval", "interval":
		d, err := time.ParseDuration(val)
		if err != nil || d <= 0 {
			log.Fatalf("[ERROR] Invalid duration format '%s'. Use values like '5s', '10s', or '1m'.", val)
		}
		cfg.CheckInterval = val
	case "username":
		cfg.Username = val
		log.Printf("[INFO] Updated default user to %s. Ensure credentials are saved via 'kawaii-wify login'.", val)
	default:
		log.Fatalf("[ERROR] Unknown configuration key '%s'. Supported keys: username, check_interval", args[0])
	}

	if err := config.Save(cfg); err != nil {
		log.Fatalf("[FATAL] Failed to save configuration: %v", err)
	}

	log.Printf("[SUCCESS] Configuration updated: %s = %s", key, val)
}

func init() {
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}
