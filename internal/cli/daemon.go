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
	"github.com/spf13/cobra"
)

var userOverride string

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Runs the background authentication and keepalive services",
	Run:   runDaemon,
}

func runDaemon(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("[WARN] Failed to load config, using defaults: %v", err)
		cfg = &config.Config{CheckInterval: "10s"}
	}

	targetUser := cfg.Username
	if userOverride != "" {
		targetUser = userOverride
	}

	user, pass, err := credentials.Resolve(targetUser)
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

	log.Printf("[INFO] Starting kawaii-wify for user: %s (interval: %s)", user, cfg.Interval())

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	client := auth.NewClient()
	eng := engine.New(client, user, pass)

	if err := eng.Run(ctx, cfg.Interval()); err != nil && err != context.Canceled {
		log.Fatalf("[FATAL] Engine crashed: %v", err)
	}

	log.Println("[INFO] kawaii-wify gracefully stopped.")
}

func init() {
	daemonCmd.Flags().StringVarP(&userOverride, "user", "u", "", "Set or override kawaii-wify username")
	rootCmd.AddCommand(daemonCmd)
}
