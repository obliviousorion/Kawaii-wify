package main

import (
	"fmt"

	"github.com/obliviousorion/bits-wlan-manager/internal/auth"
	"github.com/obliviousorion/bits-wlan-manager/internal/engine"
)	

func main() {
	client := auth.NewClient()
	eng := engine.New(client, "testuser", "testpass")

	fmt.Printf("[TEST] Initial State: %s\n", eng.State())
}