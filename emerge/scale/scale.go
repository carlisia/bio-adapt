// Package scale defines swarm size categories for configuration tuning.
// Scale modifiers adjust parameters based on the number of agents.
package scale

// Size represents a swarm size category.
type Size int

const (
	// Tiny represents swarms with fewer than 10 agents.
	// Fully connected topology, very strong coupling.
	Tiny Size = iota

	// Small represents swarms with 10-50 agents.
	// Dense connectivity, strong coupling.
	Small

	// Medium represents swarms with 50-200 agents.
	// Moderate connectivity, balanced coupling.
	Medium

	// Large represents swarms with 200-1000 agents.
	// Sparse connectivity, weaker coupling.
	Large

	// Huge represents swarms with 1000+ agents.
	// Minimal connectivity, very weak coupling.
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
	case count < 10:
		return Tiny
	case count < 50:
		return Small
	case count < 200:
		return Medium
	case count < 1000:
		return Large
	default:
		return Huge
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
