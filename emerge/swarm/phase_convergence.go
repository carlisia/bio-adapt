package swarm

import (
	"math"
)

// MeasurePhaseConvergence calculates how well agents have converged to a target phase.
// Returns a value between 0 and 1, where 1 means perfect convergence.
func (s *Swarm) MeasurePhaseConvergence(targetPhase float64) float64 {
	agents := s.collectAgents()
	if len(agents) == 0 {
		return 0
	}

	// Calculate mean phase distance from target
	totalDistance := 0.0
	for _, agent := range agents {
		phase := agent.Phase()
		// Calculate circular distance
		diff := math.Abs(phase - targetPhase)
		// Handle wraparound
		if diff > math.Pi {
			diff = 2*math.Pi - diff
		}
		totalDistance += diff
	}

	avgDistance := totalDistance / float64(len(agents))
	// Convert to convergence score (1 = perfect, 0 = worst)
	convergence := 1.0 - (avgDistance / math.Pi)
	if convergence < 0 {
		convergence = 0
	}
	return convergence
}

// MeasurePhaseVariance calculates the variance of agent phases.
// Lower values indicate agents are closer to the same phase.
func (s *Swarm) MeasurePhaseVariance() float64 {
	agents := s.collectAgents()
	if len(agents) == 0 {
		return 0
	}

	// Calculate circular mean using vector addition
	sumCos := 0.0
	sumSin := 0.0
	for _, agent := range agents {
		phase := agent.Phase()
		sumCos += math.Cos(phase)
		sumSin += math.Sin(phase)
	}

	// Mean resultant vector
	meanCos := sumCos / float64(len(agents))
	meanSin := sumSin / float64(len(agents))

	// Mean phase
	meanPhase := math.Atan2(meanSin, meanCos)

	// Calculate circular variance
	variance := 0.0
	for _, agent := range agents {
		phase := agent.Phase()
		diff := phase - meanPhase
		// Normalize to [-π, π]
		for diff > math.Pi {
			diff -= 2 * math.Pi
		}
		for diff < -math.Pi {
			diff += 2 * math.Pi
		}
		variance += diff * diff
	}

	return variance / float64(len(agents))
}
