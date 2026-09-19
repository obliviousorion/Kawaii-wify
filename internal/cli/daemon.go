package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/obliviousorion/kawaii-wify/internal/auth"
	"github.com/obliviousorion/kawaii-wify/internal/config"
	"github.com/obliviousorion/kawaii-wify/internal/credentials"
	"github.com/obliviousorion/kawaii-wify/internal/engine"
	"github.com/obliviousorion/kawaii-wify/internal/ipc"
	"github.com/obliviousorion/kawaii-wify/internal/logger"
	"github.com/spf13/cobra"
)

var (
	userOverride    string
	keepaliveFlag   bool
	noKeepalive     bool
	autoConnectFlag bool
	noAutoConnect   bool
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Runs the background authentication and keepalive services",
	Long:  "Runs the background session manager. Continuously monitors connectivity and automatically logs in when a captive portal is encountered.",
	Run:   runDaemon,
}

func runDaemon(cmd *cobra.Command, args []string) {
	cleanup, err := logger.Setup()
	if err != nil {
		logger.Warn("Failed to setup file logging (falling back to stdout): %v", err)
	} else {
		defer cleanup()
	}

	cfg, err := config.Load()
	if err != nil {
		logger.Warn("Failed to load config, using defaults: %v", err)
		cfg = config.Default()
	}

	user, pass, err := credentials.Resolve(userOverride, cfg.Username)
	if err != nil {
		logger.Fatal("Could not resolve user credentials: %v", err)
	}

	if cfg.Username != user {
		cfg.Username = user
		if err := config.Save(cfg); err != nil {
			logger.Warn("Failed to save config: %v", err)
		} else {
			logger.Boot("Saved default username %s to config", user)
		}
	}

	keepalive := cfg.Keepalive
	if cmd.Flags().Changed("no-keepalive") {
		keepalive = false
	} else if cmd.Flags().Changed("keepalive") {
		keepalive = keepaliveFlag
	}

	autoConnect := cfg.AutoConnect
	if cmd.Flags().Changed("no-auto-connect") || cmd.Flags().Changed("paused") {
		autoConnect = false
	} else if cmd.Flags().Changed("auto-connect") {
		autoConnect = autoConnectFlag
	}

	logger.Boot("Starting kawaii-wify daemon (PID: %d, user: %s, interval: %s, keepalive: %t, auto_connect: %t)", 
		os.Getpid(), user, cfg.Interval(), keepalive, autoConnect)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// ipc listener
	listener, err := ipc.Listen()
	if err != nil {
		logger.Fatal("Cannot start IPC Listener: %v", err)
	}
	defer listener.Close()

	client := auth.NewClient()
	eng := engine.New(client, user, pass, keepalive, autoConnect)

	go func() {
		if err := ipc.Serve(ctx, cancel, listener, eng); err != nil {
			logger.Warn("IPC server stopped: %v", err)
		}
	}()

	if err := eng.Run(ctx, cfg.Interval()); err != nil && err != context.Canceled {
		logger.Fatal("Engine crashed: %v", err)
	}

	logger.Boot("kawaii-wify gracefully stopped")
}

func init() {
	daemonCmd.Flags().StringVarP(&userOverride, "user", "u", "", "Set or override kawaii-wify username")
	daemonCmd.Flags().BoolVar(&keepaliveFlag, "keepalive", true, "Enable periodic keepalive pings")
	daemonCmd.Flags().BoolVar(&noKeepalive, "no-keepalive", false, "Disable periodic keepalive pings")
	daemonCmd.Flags().BoolVar(&autoConnectFlag, "auto-connect", true, "Automatically connect and authenticate on launch")
	daemonCmd.Flags().BoolVar(&noAutoConnect, "no-auto-connect", false, "Start daemon without automatically connecting")
	daemonCmd.Flags().BoolVarP(&noAutoConnect, "paused", "p", false, "Start daemon in paused state (alias for --no-auto-connect)")
	rootCmd.AddCommand(daemonCmd)
}
