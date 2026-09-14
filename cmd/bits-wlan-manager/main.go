package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/obliviousorion/bits-wlan-manager/internal/auth"
	"github.com/obliviousorion/bits-wlan-manager/internal/engine"
)	

func main() {
	client := auth.NewClient()
	eng := engine.New(client, "F20230814", "F20237057#")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	defer cancel()

	if err := eng.Run(ctx, 10*time.Second); err != nil {
		if err == context.Canceled {
			log.Println("[INFO] Engine gracefully shuts down.")
		} else {
			log.Printf("[ERROR] Engine run failed: %v\n", err)
		}
	}

}