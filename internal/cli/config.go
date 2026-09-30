package cli

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/obliviousorion/kawaii-wify/internal/config"
	"github.com/obliviousorion/kawaii-wify/internal/ipc"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View and manage local settings",
	Long:  "View and update non-sensitive configuration settings such as check interval, keepalive, default username, and TLS certificate pins.",
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

var configClearPinsCmd = &cobra.Command{
	Use:   "clear-pins [endpoint]",
	Short: "Clear stored TLS certificate pins for a gateway (or default gateway)",
	Args:  cobra.MaximumNArgs(1),
	Run:   runConfigClearPins,
}

func runConfigGet(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("[WARN] Failed to load config, showing defaults: %v", err)
		cfg = config.Default()
	}

	if len(args) == 0 {
		fmt.Printf("username: %s\n", cfg.Username)
		fmt.Printf("gateway: %s\n", cfg.GatewayEndpoint())
		fmt.Printf("check_interval: %s\n", cfg.CheckInterval)
		fmt.Printf("keepalive: %t\n", cfg.Keepalive)
		fmt.Printf("auto_connect: %t\n", cfg.AutoConnect)
		fmt.Printf("verify_tls: %t\n", cfg.IsTLSVerificationEnabled())
		fmt.Printf("notifications: %t\n", cfg.IsNotificationsEnabled())
		pins := cfg.GetPins(cfg.GatewayEndpoint())
		if len(pins) > 0 {
			fmt.Printf("cert_pins (%s): %d stored\n", cfg.GatewayEndpoint(), len(pins))
			for _, p := range pins {
				fmt.Printf("  - %s\n", p)
			}
		} else {
			fmt.Printf("cert_pins (%s): None (TOFU mode)\n", cfg.GatewayEndpoint())
		}
		return
	}

	key := strings.ToLower(args[0])
	switch key {
	case "username":
		fmt.Println(cfg.Username)
	case "gateway":
		fmt.Println(cfg.GatewayEndpoint())
	case "check_interval", "interval":
		fmt.Println(cfg.CheckInterval)
	case "keepalive":
		fmt.Println(cfg.Keepalive)
	case "auto_connect", "autoconnect":
		fmt.Println(cfg.AutoConnect)
	case "verify_tls", "verifytls":
		fmt.Println(cfg.IsTLSVerificationEnabled())
	case "notifications", "notify":
		fmt.Println(cfg.IsNotificationsEnabled())
	case "cert_pins", "pins":
		endpoint := cfg.GatewayEndpoint()
		pins := cfg.GetPins(endpoint)
		if len(pins) == 0 {
			fmt.Printf("No certificate pins recorded for %s (TOFU mode)\n", endpoint)
		} else {
			for _, p := range pins {
				fmt.Println(p)
			}
		}
	default:
		log.Fatalf("[ERROR] Unknown configuration key '%s'. Supported keys: username, gateway, check_interval, keepalive, auto_connect, verify_tls, notifications, cert_pins", args[0])
	}
}

func runConfigSet(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("[WARN] Failed to load config, initializing fresh: %v", err)
		cfg = config.Default()
	}

	key := strings.ToLower(args[0])
	val := args[1]

	switch key {
	case "gateway":
		if strings.ToLower(val) == "default" {
			cfg.Gateway = config.DefaultGateway
		} else {
			cfg.Gateway = val
			cfg.Gateway = cfg.GatewayEndpoint()
		}
		val = cfg.Gateway
	case "check_interval", "interval":
		d, err := time.ParseDuration(val)
		if err != nil {
			log.Fatalf("[ERROR] Invalid duration format '%s'. Use values like '5s', '10s', or '1m'.", val)
		}
		if d < 2*time.Second {
			log.Fatalf("[ERROR] Check interval must be at least 2s (got %s)", val)
		}
		cfg.CheckInterval = val
	case "username":
		cfg.Username = val
		log.Printf("[INFO] Updated default user to %s. Ensure credentials are saved via 'kawaii-wify login'.", val)
	case "keepalive":
		b, err := strconv.ParseBool(val)
		if err != nil {
			log.Fatalf("[ERROR] Invalid boolean value '%s'. Use 'true' or 'false'.", val)
		}
		cfg.Keepalive = b
	case "auto_connect", "autoconnect":
		b, err := strconv.ParseBool(val)
		if err != nil {
			log.Fatalf("[ERROR] Invalid boolean value '%s'. Use 'true' or 'false'.", val)
		}
		cfg.AutoConnect = b
	case "verify_tls", "verifytls":
		b, err := strconv.ParseBool(val)
		if err != nil {
			log.Fatalf("[ERROR] Invalid boolean value '%s'. Use 'true' or 'false'.", val)
		}
		cfg.VerifyTLS = &b
	case "notifications", "notify":
		b, err := strconv.ParseBool(val)
		if err != nil {
			log.Fatalf("[ERROR] Invalid boolean value '%s'. Use 'true' or 'false'.", val)
		}
		cfg.SetNotificationsEnabled(b)
	default:
		log.Fatalf("[ERROR] Unknown configuration key '%s'. Supported keys: username, gateway, check_interval, keepalive, auto_connect, verify_tls, notifications", args[0])
	}

	if err := config.Save(cfg); err != nil {
		log.Fatalf("[FATAL] Failed to save configuration: %v", err)
	}

	log.Printf("[SUCCESS] Configuration updated: %s = %s", key, val)
	notifyDaemonConfigReload()
}

func runConfigClearPins(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[ERROR] Failed to load config: %v", err)
	}

	target := cfg.GatewayEndpoint()
	if len(args) > 0 {
		target = args[0]
	}

	if strings.EqualFold(target, "all") {
		cfg.CertPins = make(map[string][]string)
		if err := config.Save(cfg); err != nil {
			log.Fatalf("[ERROR] Failed to save config: %v", err)
		}
		fmt.Println("✓ Cleared all stored certificate pins.")
		notifyDaemonConfigReload()
		return
	}

	cfg.ClearPins(target)
	if err := config.Save(cfg); err != nil {
		log.Fatalf("[ERROR] Failed to save config: %v", err)
	}
	fmt.Printf("✓ Cleared certificate pins for %s. Next connection will re-pin in TOFU mode.\n", target)
	notifyDaemonConfigReload()
}

func notifyDaemonConfigReload() {
	client, err := ipc.NewClient()
	if err == nil {
		defer client.Close()
		_, _ = client.ReloadConfig()
	}
}

var configReloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Hot-reload active configuration into running daemon",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := ipc.NewClient()
		if err != nil {
			fmt.Println("✕ Daemon is not running. Configuration will be loaded on next startup.")
			return
		}
		defer client.Close()
		resp, err := client.ReloadConfig()
		if err != nil {
			fmt.Printf("✕ Reload failed: %v\n", err)
			return
		}
		fmt.Printf("✓ %s\n", resp.Message)
	},
}

func init() {
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configClearPinsCmd)
	configCmd.AddCommand(configReloadCmd)
	rootCmd.AddCommand(configCmd)
}

