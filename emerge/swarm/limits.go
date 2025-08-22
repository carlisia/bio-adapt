package swarm

import (
	"math"
)

// CoherenceLimits defines theoretical and practical coherence limits based on swarm size
type CoherenceLimits struct {
	Theoretical float64 // Maximum theoretically achievable
	Practical   float64 // Maximum practically achievable in reasonable time
	Recommended float64 // Recommended target for reliable convergence
}

// GetCoherenceLimits returns the coherence limits for a given swarm size.
// Based on Kuramoto model theory and empirical observations.
func GetCoherenceLimits(swarmSize int) CoherenceLimits {
	if swarmSize <= 0 {
		return CoherenceLimits{0, 0, 0}
	}

	// For very small swarms, finite-size effects dominate
	// The theoretical maximum is limited by 1/sqrt(N) fluctuations
	switch {
	case swarmSize == 1:
		// Single agent is always perfectly coherent with itself
		return CoherenceLimits{
			Theoretical: 1.0,
			Practical:   1.0,
			Recommended: 1.0,
		}
	case swarmSize < 5:
		// Very small swarms have high variance
		// Theoretical max approaches 1 but with large fluctuations
		return CoherenceLimits{
			Theoretical: 0.95,
			Practical:   0.85,
			Recommended: 0.75,
		}
	case swarmSize < 10:
		// Small swarms can achieve good coherence but not perfect
		// Finite size effects: ~1/sqrt(N) = ~0.32 for N=10
		return CoherenceLimits{
			Theoretical: 0.98,
			Practical:   0.92,
			Recommended: 0.85,
		}
	case swarmSize < 20:
		// Better coherence possible with more agents
		return CoherenceLimits{
			Theoretical: 0.99,
			Practical:   0.95,
			Recommended: 0.90,
		}
	case swarmSize < 50:
		// Medium swarms can achieve high coherence
		return CoherenceLimits{
			Theoretical: 0.995,
			Practical:   0.97,
			Recommended: 0.92,
		}
	case swarmSize < 100:
		// Large swarms approach theoretical limits
		return CoherenceLimits{
			Theoretical: 0.998,
			Practical:   0.98,
			Recommended: 0.95,
		}
	default:
		// Very large swarms can achieve near-perfect coherence
		// Limited mainly by numerical precision and coupling strength
		return CoherenceLimits{
			Theoretical: 0.999,
			Practical:   0.99,
			Recommended: 0.97,
		}
	}
}

// GetMinimumSwarmSize returns the minimum swarm size needed to reliably achieve
// a target coherence level within reasonable time.
func GetMinimumSwarmSize(targetCoherence float64) int {
	// Based on empirical observations and Kuramoto model theory
	switch {
	case targetCoherence >= 0.99:
		return 100 // Need large swarm for very high coherence
	case targetCoherence >= 0.97:
		return 50
	case targetCoherence >= 0.95:
		return 20
	case targetCoherence >= 0.90:
		return 10
	case targetCoherence >= 0.85:
		return 5
	default:
		return 2 // Any swarm size > 1 can achieve moderate coherence
	}
}

// GetConvergenceTimeFactor returns a multiplication factor for expected convergence time
// based on how close the target is to the practical limit.
func GetConvergenceTimeFactor(swarmSize int, targetCoherence float64) float64 {
	limits := GetCoherenceLimits(swarmSize)

	// If target exceeds practical limit, convergence will be very slow or impossible
	if targetCoherence > limits.Practical {
		return math.Inf(1) // Infinite time expected
	}

	// Calculate how close we are to the practical limit
	if limits.Practical <= 0 {
		return 1.0
	}

	// The closer to the limit, the longer it takes (exponential increase)
	utilization := targetCoherence / limits.Practical
	switch {
	case utilization < 0.8:
		return 1.0 // Normal convergence time
	case utilization < 0.9:
		return 2.0 // 2x longer
	case utilization < 0.95:
		return 4.0 // 4x longer
	default:
		return 8.0 // 8x longer for very high targets
	}
}

// ValidateCoherenceTarget checks if a target coherence is achievable for a given swarm size
// and returns an adjusted target if necessary.
func ValidateCoherenceTarget(swarmSize int, requestedCoherence float64) (adjustedCoherence float64, warning string) {
	limits := GetCoherenceLimits(swarmSize)

	if requestedCoherence > limits.Theoretical {
		return limits.Theoretical,
			"Target coherence exceeds theoretical maximum for this swarm size"
	}

	if requestedCoherence > limits.Practical {
		return requestedCoherence,
			"Target coherence may not be achievable in reasonable time"
	}

	if requestedCoherence > limits.Recommended {
		return requestedCoherence,
			"Target coherence is ambitious for this swarm size"
	}

	return requestedCoherence, ""
}
