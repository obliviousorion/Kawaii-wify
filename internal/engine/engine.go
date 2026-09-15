package engine

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/obliviousorion/kawaii-wify/internal/auth"
)

const (
	MaxAuthFailures  = 3
	CooldownDuration = 10 * time.Second
)

type Engine struct {
	client   *http.Client
	username string
	password string

	mu            sync.RWMutex
	state         State
	sessionToken  string
	failCount     int
	cooldownStart time.Time
}

// Constructor for the Engine

func New(client *http.Client, username string, password string) *Engine {
	return &Engine{
		client:    client,
		username:  username,
		password:  password,
		state:     StateOffline,
		failCount: 0,
	}
}

// Main Run logic

func (e *Engine) Run(ctx context.Context, interval time.Duration) error {

	// execute a tick immediately
	e.Tick()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[INFO] Shutting down the engine")
			return ctx.Err()

		case <-ticker.C:
			e.Tick()
		}
	}

}

// Main Tick implementation for the Engine

func (e *Engine) Tick() {

	if e.inCooldown() {
		return
	}

	isCaptive, magicToken, err := auth.Probe(e.client)
	if err != nil {
		log.Printf("[ERROR] Probe failed: %v", err)
		e.transition(StateOffline)
		return
	}

	if !isCaptive {
		e.transition(StateOnline)
		token := e.SessionToken()
		if token != "" {
			err := auth.Keepalive(e.client, token)
			if err != nil {
				log.Printf("[WARN] Keepalive failed: %v", err)
				return
			}
			log.Printf("[INFO] Keepalive successful")
			return
		}
	}

	if isCaptive {
		e.transition(StateCaptive)

		if e.FailureCount() >= MaxAuthFailures {
			e.transition(StateCooldown)
			return
		}

		err := auth.Prime(e.client, magicToken)
		if err != nil {
			log.Printf("[ERROR] Priming Failed, Retrying on next Tick: %v", err)
			return
		}
		sessionToken, err := auth.Login(e.client, e.username, e.password, magicToken)
		if err != nil {
			currentFails := e.incrementFailures()

			log.Printf("[ERROR] Login failed (attempt %d/%d): %v", currentFails, MaxAuthFailures, err)
			log.Println("[INFO] Retrying Login on next Tick")

			if currentFails >= MaxAuthFailures {
				e.transition(StateCooldown)
			}

			return
		}

		e.setSessionToken(sessionToken)
		log.Printf("[SUCCESS] Logged in! Session token: %s", sessionToken)
		e.transition(StateOnline)
	}
}

// Thread-Safe Getters

func (e *Engine) State() State {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.state
}

func (e *Engine) FailureCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.failCount
}

func (e *Engine) SessionToken() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.sessionToken
}

// Mutex Helpers

func (e *Engine) setSessionToken(token string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.sessionToken = token
}

// transition safely updates the engine state and handles entry actions.
func (e *Engine) transition(next State) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.state == next {
		return
	}

	log.Printf("[STATE] Transition: %s -> %s", e.state, next)
	e.state = next

	// Reset the circuit breaker on a successful connection
	if next == StateOnline {
		e.failCount = 0
	}

	if next == StateCooldown {
		e.cooldownStart = time.Now()
	}
}

func (e *Engine) incrementFailures() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.failCount++
	return e.failCount
}

// inCooldown checks if the circuit breaker is active and self-heals if expired.
// Returns true if tick should be skipped.
func (e *Engine) inCooldown() bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.state != StateCooldown {
		return false
	}

	// Has enough time elapsed?
	if time.Since(e.cooldownStart) >= CooldownDuration {
		log.Println("[ENGINE] Cooldown elapsed. Resetting circuit breaker.")
		e.failCount = 0
		e.state = StateOffline
		log.Printf("[STATE] Transition: %s -> %s", StateCooldown, StateOffline)
		return false // Cooldown is over, let the tick proceed!
	}

	waitTime := CooldownDuration - time.Since(e.cooldownStart)
	log.Printf("[INFO] Engine in cooldown. Retrying in %s", waitTime.Round(time.Second))
	return true // Still in cooldown, skip this tick
}
