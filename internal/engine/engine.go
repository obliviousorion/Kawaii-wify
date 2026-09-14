package engine

import (
	"log"
	"net/http"
	"sync"
)


const MaxAuthFailures = 3

type Engine struct {
// 	client (*http.Client)
	client		*http.Client
	username	string
	password	string

	mu			sync.RWMutex
	state		State
	sessionToken string
	failCount	int

}

func NewEngine(client *http.Client, username string, password string) *Engine {
	return &Engine{
		client: client,
		username: username,
		password: password,
		state: StateOffline,
		failCount: 0,
	}
}


func (e *Engine) State() State {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.state
}

func (e *Engine) transitionTo(next State) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.state == next {
		return nil
	}

	log.Printf("[STATE] Transition from %s → %s", e.state, next)

	if next == StateOnline {
		e.failCount = 0 // resettting the circuit breaker
	}
	
	e.state = next
	return nil
}

func (e *Engine) tick() {
	// TODO
}