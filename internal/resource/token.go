package resource

import (
	"math"
	"sync"

	"go.uber.org/atomic"
)

// TokenManager implements simple energy management.
// Agents have limited energy that depletes through actions.
type TokenManager struct {
	tokens    atomic.Float64
	maxTokens float64
	mu        sync.Mutex
}

// NewTokenManager creates a new resource manager with the specified maximum tokens.
func NewTokenManager(maxTokens float64) *TokenManager {
	// Handle negative max tokens - treat as 0
	if maxTokens < 0 {
		maxTokens = 0
	}
	t := &TokenManager{
		maxTokens: maxTokens,
	}
	t.tokens.Store(maxTokens)
	return t
}

// Request attempts to allocate resources for an action.
func (t *TokenManager) Request(amount float64) float64 {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Handle negative or zero requests
	if amount <= 0 {
		return 0
	}

	available := t.tokens.Load()
	// Can't allocate from negative pool
	if available <= 0 {
		return 0
	}

	allocated := math.Min(amount, available)
	t.tokens.Store(available - allocated)

	return allocated
}

// Release returns unused resources to the pool.
func (t *TokenManager) Release(amount float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Ignore negative releases to maintain resource integrity
	if amount <= 0 {
		return
	}

	current := t.tokens.Load()
	t.tokens.Store(math.Min(current+amount, t.maxTokens))
}

// Available returns current resource level.
func (t *TokenManager) Available() float64 {
	return t.tokens.Load()
}
