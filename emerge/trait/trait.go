// Package trait defines optimization targets for swarm behavior.
// Traits represent HOW you want to achieve your goals.
package trait

// Target represents a behavioral preference or optimization constraint.
type Target int

const (
	// Stability prioritizes gentle changes and predictable behavior.
	// Uses conservative strategies with low adjustment rates.
	Stability Target = iota

	// Speed prioritizes fast convergence over smoothness.
	// Uses aggressive strategies and higher update rates.
	Speed

	// Efficiency prioritizes low resource consumption.
	// Minimizes energy usage and computational overhead.
	Efficiency

	// Throughput prioritizes high processing capacity.
	// Optimizes for handling maximum workload.
	Throughput

	// Resilience prioritizes fault tolerance.
	// Uses redundant strategies and handles failures gracefully.
	Resilience

	// Precision prioritizes accuracy over speed.
	// Uses fine-grained adjustments and tight tolerances.
	Precision
)

// String returns the string representation of the target.
func (t Target) String() string {
	switch t {
	case Stability:
		return "stability"
	case Speed:
		return "speed"
	case Efficiency:
		return "efficiency"
	case Throughput:
		return "throughput"
	case Resilience:
		return "resilience"
	case Precision:
		return "precision"
	default:
		return "unknown"
	}
}
