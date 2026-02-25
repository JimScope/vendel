package services

import (
	"sync"
	"time"
)

type circuitState int

const (
	stateClosed   circuitState = iota
	stateOpen
	stateHalfOpen
)

// CircuitBreaker prevents cascading failures by short-circuiting calls
// to an unhealthy dependency after repeated failures.
type CircuitBreaker struct {
	name       string
	maxFails   int
	cooldown   time.Duration

	mu         sync.Mutex
	state      circuitState
	failures   int
	openedAt   time.Time
}

func NewCircuitBreaker(name string, maxFails int, cooldown time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:     name,
		maxFails: maxFails,
		cooldown: cooldown,
	}
}

// Allow returns true if the call should proceed.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case stateClosed:
		return true
	case stateOpen:
		if time.Since(cb.openedAt) >= cb.cooldown {
			cb.state = stateHalfOpen
			return true
		}
		return false
	case stateHalfOpen:
		return false // only one probe at a time; wait for result
	}
	return true
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	cb.state = stateClosed
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	if cb.state == stateHalfOpen || cb.failures >= cb.maxFails {
		cb.state = stateOpen
		cb.openedAt = time.Now()
	}
}
