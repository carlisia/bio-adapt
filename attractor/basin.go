package attractor

import (
	"math"
	"sync"
)

// AttractorBasin represents a stable state that the system tends toward.
// In biological systems, these are like "morphogenetic fields" that guide
// development and behavior toward specific patterns.
//
// The basin metaphor comes from dynamical systems theory where states
// naturally "roll down" into stable configurations like a ball rolling
// into a valley.
type AttractorBasin struct {
	TargetState State   // The attractor point
	Strength    float64 // Basin depth/attraction strength [0, 1]
	Radius      float64 // Basin of attraction radius

	mu sync.RWMutex
}

// NewAttractorBasin creates a new attractor basin with the given target state.
func NewAttractorBasin(target State, strength, radius float64) *AttractorBasin {
	return &AttractorBasin{
		TargetState: target,
		Strength:    math.Max(0, math.Min(1, strength)), // Clamp to [0, 1]
		Radius:      math.Max(0, radius),
	}
}

// DistanceToTarget calculates the phase distance from current state to target.
// Returns a value between 0 (at target) and π (maximally distant).
func (ab *AttractorBasin) DistanceToTarget(current State) float64 {
	ab.mu.RLock()
	defer ab.mu.RUnlock()

	// Calculate phase difference (wrapped to [-π, π])
	diff := current.Phase - ab.TargetState.Phase

	// Wrap to [-π, π]
	for diff > math.Pi {
		diff -= 2 * math.Pi
	}
	for diff < -math.Pi {
		diff += 2 * math.Pi
	}

	return math.Abs(diff)
}

// AttractionForce calculates the force pulling toward the basin center.
// Force is stronger when closer to the target (within radius).
// Returns a value between 0 (no force) and Strength (maximum force).
func (ab *AttractorBasin) AttractionForce(distance float64) float64 {
	ab.mu.RLock()
	defer ab.mu.RUnlock()

	if distance > ab.Radius {
		// Outside the basin of attraction
		return 0
	}

	// Linear decay within radius (could be made nonlinear)
	normalizedDist := distance / ab.Radius
	return ab.Strength * (1 - normalizedDist)
}

// IsInBasin checks if a state is within the basin of attraction.
func (ab *AttractorBasin) IsInBasin(state State) bool {
	return ab.DistanceToTarget(state) <= ab.Radius
}

// ConvergenceRate estimates how quickly states converge to the target.
// Higher values mean faster convergence.
func (ab *AttractorBasin) ConvergenceRate() float64 {
	ab.mu.RLock()
	defer ab.mu.RUnlock()

	// Rate depends on both strength and radius
	// Stronger basins with smaller radius converge faster
	if ab.Radius == 0 {
		return 0
	}
	return ab.Strength / ab.Radius
}

// SetStrength updates the basin's attraction strength.
func (ab *AttractorBasin) SetStrength(strength float64) {
	ab.mu.Lock()
	defer ab.mu.Unlock()
	ab.Strength = math.Max(0, math.Min(1, strength))
}

// SetRadius updates the basin's radius of attraction.
func (ab *AttractorBasin) SetRadius(radius float64) {
	ab.mu.Lock()
	defer ab.mu.Unlock()
	ab.Radius = math.Max(0, radius)
}

// MeasureCoherence calculates the Kuramoto order parameter for a collection of phases.
// This measures how synchronized the phases are (0 = random, 1 = perfect sync).
// This is moved here from swarm.go as it's fundamentally about measuring
// convergence within an attractor basin.
func MeasureCoherence(phases []float64) float64 {
	if len(phases) == 0 {
		return 0
	}

	// Kuramoto order parameter: R = |Σ e^(iφ)| / N
	var sumCos, sumSin float64
	for _, phase := range phases {
		sumCos += math.Cos(phase)
		sumSin += math.Sin(phase)
	}

	n := float64(len(phases))
	return math.Sqrt(sumCos*sumCos+sumSin*sumSin) / n
}

// MeasureCoherenceWeighted calculates weighted Kuramoto order parameter.
// Useful when some agents have more influence than others.
func MeasureCoherenceWeighted(phases []float64, weights []float64) float64 {
	if len(phases) == 0 || len(phases) != len(weights) {
		return 0
	}

	var sumCos, sumSin, sumWeights float64
	for i, phase := range phases {
		w := weights[i]
		sumCos += w * math.Cos(phase)
		sumSin += w * math.Sin(phase)
		sumWeights += w
	}

	if sumWeights == 0 {
		return 0
	}

	return math.Sqrt(sumCos*sumCos+sumSin*sumSin) / sumWeights
}

// PhaseDifference calculates the wrapped phase difference between two phases.
// Returns a value in [-π, π].
func PhaseDifference(phase1, phase2 float64) float64 {
	diff := phase1 - phase2

	// Wrap to [-π, π]
	for diff > math.Pi {
		diff -= 2 * math.Pi
	}
	for diff < -math.Pi {
		diff += 2 * math.Pi
	}

	return diff
}

// WrapPhase wraps a phase value to [0, 2π].
func WrapPhase(phase float64) float64 {
	wrapped := math.Mod(phase, 2*math.Pi)
	if wrapped < 0 {
		wrapped += 2 * math.Pi
	}
	return wrapped
}

