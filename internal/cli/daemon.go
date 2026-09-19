package cli

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/obliviousorion/kawaii-wify/internal/auth"
	"github.com/obliviousorion/kawaii-wify/internal/config"
	"github.com/obliviousorion/kawaii-wify/internal/credentials"
	"github.com/obliviousorion/kawaii-wify/internal/engine"
	"github.com/obliviousorion/kawaii-wify/internal/ipc"
	"github.com/spf13/cobra"
)

var (
	userOverride string
	noKeepalive  bool
	noAutoConnect bool
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Runs the background authentication and keepalive services",
	Long:  "Runs the background session manager. Continuously monitors connectivity and automatically logs in when a captive portal is encountered.",
	Run:   runDaemon,
}

func runDaemon(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("[WARN] Failed to load config, using defaults: %v", err)
		cfg = config.Default()
	}

	user, pass, err := credentials.Resolve(userOverride, cfg.Username)
	if err != nil {
		log.Fatalf("[FATAL] Could not resolve user credentials: %v", err)
	}

	if cfg.Username != user {
		cfg.Username = user
		if err := config.Save(cfg); err != nil {
			log.Printf("[WARN] Failed to save config: %v", err)
		} else {
			log.Printf("[INFO] Saved default username %s to config", user)
		}
	}

	keepalive := cfg.Keepalive
	if noKeepalive {
		keepalive = false
	}

	autoConnect := cfg.AutoConnect
	if noAutoConnect {
		autoConnect = false
	}

log.Printf("[INFO] Starting kawaii-wify for user: %s (interval: %s, keepalive: %t, auto_connect: %t)", 
		user, cfg.Interval(), keepalive, autoConnect)


	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// ipc listener
	listener, err := ipc.Listen()
	if err != nil {
		log.Fatalf("[FATAL] Cannot start IPC Listener: %v", err)
	}
	defer listener.Close()


	client := auth.NewClient()
	eng := engine.New(client, user, pass, keepalive, autoConnect )

	go func() {
		if err := ipc.Serve(ctx, listener, eng); err != nil {
			log.Printf("[WARN] IPC server stopped: %v", err)
		}
	} ()


	if err := eng.Run(ctx, cfg.Interval()); err != nil && err != context.Canceled {
		log.Fatalf("[FATAL] Engine crashed: %v", err)
	}

	log.Println("[INFO] kawaii-wify gracefully stopped.")
}

func init() {
	daemonCmd.Flags().StringVarP(&userOverride, "user", "u", "", "Set or override kawaii-wify username")
	daemonCmd.Flags().BoolVar(&noKeepalive, "no-keepalive", false, "Disable periodic keepalive pings")
	daemonCmd.Flags().BoolVar(&noAutoConnect, "no-auto-connect", false, "Start daemon without automatically connecting")
	daemonCmd.Flags().BoolVarP(&noAutoConnect, "paused", "p", false, "Start daemon in paused state (alias for --no-auto-connect)")
	rootCmd.AddCommand(daemonCmd)
}
