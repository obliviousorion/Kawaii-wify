package main

import (
	"fmt"

	"github.com/obliviousorion/bits-wlan-manager/internal/auth"
	"github.com/obliviousorion/bits-wlan-manager/internal/engine"
)	

func main() {
	client := auth.NewClient()
	eng := engine.New(client, "F20230814", "F20237057#")

	fmt.Printf("[TEST] Initial State: %s\n", eng.State())

	eng.Tick()

	fmt.Printf("[TEST] Final State: %s\n", eng.State())
}