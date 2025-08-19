package emerge

import "math"

// AttractorBasin represents the mathematical attractor that guides convergence.
// This is internal - users interact through Swarm, not directly with the basin.
type AttractorBasin struct {
	targetPhase     float64
	targetCoherence float64
	strength        float64 // Pull strength [0, 1]
	radius          float64 // Influence radius in phase space
}

// NewAttractorBasin creates a new attractor basin.
func NewAttractorBasin(targetPhase, targetCoherence, strength, radius float64) *AttractorBasin {
	// Clamp strength to [0, 1]
	if strength < 0 {
		strength = 0
	} else if strength > 1 {
		strength = 1
	}

	return &AttractorBasin{
		targetPhase:     targetPhase,
		targetCoherence: targetCoherence,
		strength:        strength,
		radius:          radius,
	}
}

// AttractionForce calculates the pull toward the basin center.
func (b *AttractorBasin) AttractionForce(currentPhase float64) float64 {
	distance := b.PhaseDistance(currentPhase)

	if distance > b.radius {
		return 0
	}

	// Stronger force when closer to center
	normalizedDist := distance / b.radius
	return b.strength * (1 - normalizedDist)
}

// PhaseDistance calculates distance from current phase to target.
func (b *AttractorBasin) PhaseDistance(currentPhase float64) float64 {
	diff := currentPhase - b.targetPhase
	// Normalize to [-π, π] range using math.Remainder
	// This handles both positive and negative phases correctly
	diff = math.Remainder(diff, 2*math.Pi)
	// Return the absolute distance
	return math.Abs(diff)
}

// IsInBasin checks if a phase is within the basin's influence.
func (b *AttractorBasin) IsInBasin(phase float64) bool {
	return b.PhaseDistance(phase) <= b.radius
}

// ConvergenceRate estimates how quickly the system will converge.
func (b *AttractorBasin) ConvergenceRate(currentPhase float64) float64 {
	if b.radius == 0 {
		return 0
	}
	distance := b.PhaseDistance(currentPhase)
	if distance >= b.radius {
		return 0
	}
	return b.strength * (1 - distance/b.radius)
}

