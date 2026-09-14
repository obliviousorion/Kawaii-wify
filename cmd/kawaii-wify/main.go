package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/obliviousorion/kawaii-wify/internal/auth"
	"github.com/obliviousorion/kawaii-wify/internal/config"
	"github.com/obliviousorion/kawaii-wify/internal/credentials"
	"github.com/obliviousorion/kawaii-wify/internal/engine"
)

func main() {
	userFlag := flag.String("u", "", "Campus student ID override (e.g. F20230814)")
	flag.Parse()

	// 1. Load saved configuration
	cfg, err := config.Load()
	if err != nil {
		log.Printf("[WARN] Failed to load config, using defaults: %v", err)
		cfg = &config.Config{CheckInterval: "10s"}
	}

	// 2. Flags take priority over config file
	targetUser := cfg.Username
	if *userFlag != "" {
		targetUser = *userFlag
	}

	// 3. Resolve credentials (Env -> Keyring -> Prompt)
	user, pass, err := credentials.Resolve(targetUser)
	if err != nil {
		log.Fatalf("[FATAL] Could not resolve credentials: %v", err)
	}

	// 4. Save username if it was missing or updated
	if cfg.Username != user {
		cfg.Username = user
		if err := config.Save(cfg); err != nil {
			log.Printf("[WARN] Failed to save config: %v", err)
		} else {
			log.Printf("[INFO] Saved default username %s to config", user)
		}
	}

	log.Printf("[INFO] Starting kawaii-wify for user: %s (interval: %s)", user, cfg.Interval())

	// 5. Initialize network client and engine
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client := auth.NewClient()
	eng := engine.New(client, user, pass)

	// 6. Run background daemon
	if err := eng.Run(ctx, cfg.Interval()); err != nil && err != context.Canceled {
		log.Fatalf("[FATAL] Engine crashed: %v", err)
	}

	log.Println("[INFO] kawaii-wify gracefully stopped.")
}
