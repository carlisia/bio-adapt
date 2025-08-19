package emerge

import (
	"math"
	"sync"

	"go.uber.org/atomic"
)

// ResourceManager handles energy/resource allocation.
// This implements metabolic-like constraints where actions have costs.
type ResourceManager interface {
	// Request attempts to allocate resources for an action.
	// Returns actual amount available (may be less than requested).
	Request(amount float64) float64

	// Release returns unused resources to the pool.
	Release(amount float64)

	// Available returns current resource level.
	Available() float64
}

// TokenResourceManager implements simple energy management.
// Agents have limited energy that depletes through actions.
type TokenResourceManager struct {
	tokens    atomic.Float64
	maxTokens float64
	mu        sync.Mutex
}

// NewTokenResourceManager creates a new resource manager with the specified maximum tokens.
func NewTokenResourceManager(maxTokens float64) *TokenResourceManager {
	t := &TokenResourceManager{
		maxTokens: maxTokens,
	}
	t.tokens.Store(maxTokens)
	return t
}

func (t *TokenResourceManager) Request(amount float64) float64 {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Handle negative or zero requests
	if amount <= 0 {
		return 0
	}

	available := t.tokens.Load()
	allocated := math.Min(amount, available)
	t.tokens.Store(available - allocated)

	return allocated
}

func (t *TokenResourceManager) Release(amount float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	current := t.tokens.Load()
	t.tokens.Store(math.Min(current+amount, t.maxTokens))
}

func (t *TokenResourceManager) Available() float64 {
	return t.tokens.Load()
}
