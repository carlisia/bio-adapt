// Examples of usage:
//
// Simple cases:
//    client := emerge.MinimizeAPICalls(scale.Medium)
//    client := emerge.DistributeLoad(scale.Large)
//    client := emerge.Default()
//
// For custom configuration, see custom.go:
//    client := emerge.New().WithGoal(...).WithScale(...).Build()
//    client := emerge.Custom().WithTargetCoherence(0.95).Build()

package emerge

import (
	"github.com/carlisia/bio-adapt/emerge/goal"
	"github.com/carlisia/bio-adapt/emerge/scale"
)

// Convenience constructors provide simple, one-line creation for common scenarios.
// For custom configuration, use New() or Custom() from custom.go.

// MinimizeAPICalls creates a client optimized for API call minimization.
// This is a convenience method for the most common use case.
// Agents will synchronize their phases to enable batching of operations.
func MinimizeAPICalls(s scale.Size) (*Client, error) {
	return New().
		WithGoal(goal.MinimizeAPICalls).
		WithScale(s).
		Build()
}

// DistributeLoad creates a client optimized for load distribution.
// Uses anti-phase synchronization to spread work evenly across agents.
// This prevents resource contention and hotspots.
func DistributeLoad(s scale.Size) (*Client, error) {
	return New().
		WithGoal(goal.DistributeLoad).
		WithScale(s).
		Build()
}

// Default creates a client with all defaults.
// Uses MinimizeAPICalls goal with Tiny scale.
// Useful for testing or when you just need something that works.
func Default() (*Client, error) {
	return New().Build()
}
