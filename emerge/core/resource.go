package core

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
