package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/obliviousorion/kawaii-wify/internal/auth"
	"github.com/obliviousorion/kawaii-wify/internal/credentials"
	"github.com/obliviousorion/kawaii-wify/internal/engine"
)

func main() {
	userFlag := flag.String("u", "", "Campus student ID (e.g. F20230814)")
    flag.Parse()

	user, pass, err := credentials.Resolve(*userFlag)
	if err != nil {
		log.Fatalf("[FATAL] could not resolve credentials: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	client := auth.NewClient()
	eng := engine.New(client, user, pass)

	if err := eng.Run(ctx, 10*time.Second); err != nil {
		if err == context.Canceled {
			log.Println("[INFO] Engine gracefully shuts down.")
		} else {
			log.Printf("[ERROR] Engine run failed: %v\n", err)
		}
	}


}


// func main() {
// 	client := auth.NewClient()
// 	eng := engine.New(client, "F20230814", "F20237057#")

// 	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

// 	defer cancel()

// 	if err := eng.Run(ctx, 10*time.Second); err != nil {
// 		if err == context.Canceled {
// 			log.Println("[INFO] Engine gracefully shuts down.")
// 		} else {
// 			log.Printf("[ERROR] Engine run failed: %v\n", err)
// 		}
// 	}

// }