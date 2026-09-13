package main

import (
	"fmt"
	"log"
	"time"

	"github.com/obliviousorion/bits-wlan-manager/internal/auth"
)

func main() {
	client := auth.NewClient()

	// 1. Probe network state
	isCaptive, challengeToken, err := auth.Probe(client)
	if err != nil {
		log.Fatalf("[FATAL] Probe failed: %v", err)
    }

    if !isCaptive {
        fmt.Println("[INFO] Already online! WAN access confirmed.")
        return
    }

    fmt.Printf("[INFO] Trapped in portal. Challenge token: %s\n", challengeToken)

    // 2. Prime the gateway session
    if err := auth.Prime(client, challengeToken); err != nil {
        log.Fatalf("[FATAL] Priming failed: %v", err)
    }
    fmt.Println("[INFO] Gateway primed successfully.")

    // 3. Authenticate
    // (We will replace hardcoded credentials with the secure Keyring in Milestone 3)
    sessionToken, err := auth.Login(client, "F20230814", "F20237057#", challengeToken)
    if err != nil {
        log.Fatalf("[FATAL] Login failed: %v", err)
    }
    fmt.Printf("[SUCCESS] Logged in! Session token: %s\n", sessionToken)

    // 4. Test Keepalive
    fmt.Println("[INFO] Testing keepalive ping...")
    if err := auth.Keepalive(client, sessionToken); err != nil {
        log.Printf("[WARN] Keepalive check failed: %v", err)
    } else {
        fmt.Println("[SUCCESS] Keepalive acknowledged by gateway.")
    }

    // 5. Clean teardown
    isLogout := true
    if isLogout {
        time.Sleep(2 * time.Second)
        if err := auth.Logout(client, sessionToken); err != nil {
            log.Printf("[WARN] Logout error: %v", err)
        } else {
            fmt.Println("[SUCCESS] Logged out cleanly.")
        }
    }
}