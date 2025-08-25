// Package scale defines swarm size categories for configuration tuning.
// Scale modifiers adjust parameters based on the number of agents.
package scale

// Size represents a swarm size category.
type Size int

const (
	// Tiny represents a tiny swarm with 20 agents.
	// Use for tightly coupled systems requiring strong synchronization.
	// Fully connected topology allows rapid consensus.
	Tiny Size = iota

	// Small represents a small swarm with 50 agents.
	// Use for team-sized coordination with dense connectivity.
	// Balances coordination overhead with effectiveness.
	Small

	// Medium represents a medium swarm with 200 agents.
	// Use for department-scale systems with moderate connectivity.
	// Optimal for most production workloads.
	Medium

	// Large represents a large swarm with 1000 agents.
	// Use for enterprise-scale distributed systems.
	// Sparse connectivity reduces coordination overhead.
	Large

	// Huge represents a huge swarm with 2000 agents.
	// Use for massive distributed systems with minimal coupling.
	// Designed for cloud-scale deployments.
	Huge
)

// String returns the string representation of the size.
func (s Size) String() string {
	switch s {
	case Tiny:
		return "tiny"
	case Small:
		return "small"
	case Medium:
		return "medium"
	case Large:
		return "large"
	case Huge:
		return "huge"
	default:
		return "unknown"
	}
}

// FromCount returns the appropriate Size for a given agent count.
func FromCount(count int) Size {
	switch {
	case count <= 20:
		return Tiny
	case count <= 50:
		return Small
	case count <= 200:
		return Medium
	case count <= 1000:
		return Large
	default:
		return Huge
	}
}

// DefaultAgentCount returns the default number of agents for this size.
func (s Size) DefaultAgentCount() int {
	switch s {
	case Tiny:
		return 20
	case Small:
		return 50
	case Medium:
		return 200
	case Large:
		return 1000
	case Huge:
		return 2000
	default:
		return 50 // Reasonable default
	}
}

// DefaultTargetCoherence returns the recommended target coherence for this size.
// Larger swarms have lower achievable coherence due to coordination overhead.
func (s Size) DefaultTargetCoherence() float64 {
	switch s {
	case Tiny:
		return 0.90 // Very high coherence possible with few agents
	case Small:
		return 0.85 // High coherence still achievable
	case Medium:
		return 0.75 // Good balance of scale and coherence
	case Large:
		return 0.65 // Lower coherence due to coordination overhead
	case Huge:
		return 0.55 // Minimal coherence in massive swarms
	default:
		return 0.75 // Reasonable default
	}
}

// DefaultUpdateIntervalMs returns the recommended update interval in milliseconds.
// Smaller swarms can update more frequently without overhead.
func (s Size) DefaultUpdateIntervalMs() int {
	switch s {
	case Tiny:
		return 20 // Very fast updates for tiny swarms
	case Small:
		return 50 // Fast updates
	case Medium:
		return 100 // Standard update rate
	case Large:
		return 150 // Slower to reduce overhead
	case Huge:
		return 200 // Minimal update frequency
	default:
		return 100 // Standard default
	}
}

// DefaultConvergenceTimeFactor returns the expected convergence time factor.
// Larger swarms need more iterations to converge.
func (s Size) DefaultConvergenceTimeFactor() float64 {
	switch s {
	case Tiny:
		return 1.0 // Baseline convergence time
	case Small:
		return 1.2 // Slightly longer
	case Medium:
		return 1.5 // Moderate increase
	case Large:
		return 2.0 // Double the baseline
	case Huge:
		return 3.0 // Triple for massive swarms
	default:
		return 1.5 // Moderate default
	}
}

// MaxNeighbors returns the recommended maximum neighbors for this size.
func (s Size) MaxNeighbors() int {
	switch s {
	case Tiny:
		return 9 // Fully connected
	case Small:
		return 15
	case Medium:
		return 20
	case Large:
		return 10
	case Huge:
		return 5
	default:
		return 3
	}
}

// UpdateInterval returns the recommended update interval for this size.
func (s Size) UpdateInterval() int {
	switch s {
	case Tiny:
		return 20 // ms
	case Small:
		return 50
	case Medium:
		return 100
	case Large:
		return 150
	case Huge:
		return 200
	default:
		return 250
	}
}
