package engine

import (
	"context"
	"errors"
	"fmt"
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

var (
	ErrCooldown          = errors.New("circuit breaker active: engine in cooldown")
	ErrMaxAuthFailures   = errors.New("maximum auth failures reached, cooldown active")
	ErrTimeout           = errors.New("connection attempt timed out")
	ErrAlreadyInProgress = errors.New("connection attempt already in progress")
)

type Engine struct {
	client    *http.Client
	username  string
	password  string
	keepalive bool

	triggerChan	  chan chan error
	startTime	time.Time
	lastProbe	time.Time
	paused		bool
	
	mu            sync.RWMutex
	state         State
	sessionToken  string
	failCount     int
	cooldownStart time.Time
	
}

// Constructor for the Engine

func New(client *http.Client, username string, password string, keepalive bool, autoConnect bool) *Engine {
	return &Engine{
		client:    client,
		username:  username,
		password:  password,
		keepalive: keepalive,
		paused:    !autoConnect,
		state:     StateOffline,
		failCount: 0,
		triggerChan: make(chan chan error, 1),
		startTime: time.Now(),
	}
}

// Main Run logic

func (e *Engine) Run(ctx context.Context, interval time.Duration) error {

	// execute a tick immediately
	if !e.IsPaused() {
		e.tick(ctx)
	}


	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[INFO] Shutting down the engine")
			return ctx.Err()

		case <-ticker.C:
			if !e.IsPaused() {
				e.tick(ctx)
			}

		case respChan := <- e.triggerChan:
			log.Printf("[INFO] Check triggered by IPC protocol")
			err := e.tick(ctx)
			if respChan != nil {
				respChan <- err
			}
		}
	}

}

// Main Tick implementation for the Engine
func (e *Engine) tick(ctx context.Context) error {
    

    if e.IsPaused() {
        return nil
    }
	
	e.updateLastProbe()
	
    if e.inCooldown() {
        return ErrCooldown
    }

    isCaptive, magicToken, err := auth.Probe(ctx, e.client)
    if err != nil {
        log.Printf("[ERROR] Probe failed: %v", err)
        e.transition(StateOffline)
        return fmt.Errorf("probe failed: %w", err)
    }

    if !isCaptive {
        e.transition(StateOnline)
        if e.keepalive {
            token := e.SessionToken()
            if token != "" {
                if err := auth.Keepalive(ctx, e.client, token); err != nil {
                    log.Printf("[WARN] Keepalive failed: %v", err)
                    return err
                }
                log.Printf("[INFO] Keepalive successful")
            }
        }
        return nil
    }

    if isCaptive {
        e.transition(StateCaptive)

        if e.FailureCount() >= MaxAuthFailures {
            e.transition(StateCooldown)
            return ErrMaxAuthFailures
        }

        if err := auth.Prime(ctx, e.client, magicToken); err != nil {
            log.Printf("[WARN] Gateway priming failed, will retry on next tick: %v", err)
            return err
        }

        sessionToken, err := auth.Login(ctx, e.client, e.username, e.password, magicToken)
        if err != nil {
            currentFails := e.incrementFailures()
            log.Printf("[ERROR] Login failed (attempt %d/%d): %v", currentFails, MaxAuthFailures, err)
            if currentFails >= MaxAuthFailures {
                e.transition(StateCooldown)
            }
            return err
        }

        e.setSessionToken(sessionToken)
        log.Printf("[SUCCESS] Logged in! Session token: %s", sessionToken)
        e.transition(StateOnline)
    }

    return nil
}

// State Controllers for IPC

func (e *Engine) Connect(timeout time.Duration) error {
    e.mu.Lock()
    e.paused = false
    e.failCount = 0
    if e.state == StateCooldown {
        e.transitionLocked(StateOffline)
    }
    e.mu.Unlock()

    respChan := make(chan error, 1)

    select {
    case e.triggerChan <- respChan:
        select {
        case err := <-respChan:
            return err
        case <-time.After(timeout):
            return ErrTimeout
        }
    default:
        return ErrAlreadyInProgress
    }
}

func (e *Engine) Disconnect() {
	e.mu.Lock()
	e.paused = true
	e.sessionToken = ""
	e.transitionLocked(StateOffline)
	e.mu.Unlock()

	log.Printf("[INFO] Disconnected and Engine paused by user command")
}


// Accessor Methods for the IPC
func (e *Engine) TriggerCheck() {
	select {
		case e.triggerChan <- nil:
		log.Printf("[INFO] Background Check Triggered")
		default:
	}
}

func (e *Engine) Username() string {
	return e.username
}

func (e *Engine) Uptime() time.Duration {
	return time.Since(e.startTime)
}

func (e *Engine) LastProbe() time.Time {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.lastProbe
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

func (e *Engine) IsPaused() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.paused
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
	e.transitionLocked(next)
}

func (e *Engine) transitionLocked(next State) {
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
		e.transitionLocked(StateOffline)
		return false // Cooldown is over, let the tick proceed!
	}

	waitTime := CooldownDuration - time.Since(e.cooldownStart)
	log.Printf("[INFO] Engine in cooldown. Retrying in %s", waitTime.Round(time.Second))
	return true // Still in cooldown, skip this tick
}


func (e *Engine) updateLastProbe() {
	e.mu.Lock()
	e.lastProbe = time.Now()
	e.mu.Unlock()
}