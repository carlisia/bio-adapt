package emerge

import "math"

// WrapPhase normalizes a phase value to [0, 2π]
func WrapPhase(phase float64) float64 {
	// Normalize to [0, 2π]
	for phase < 0 {
		phase += 2 * math.Pi
	}
	for phase >= 2*math.Pi {
		phase -= 2 * math.Pi
	}
	return phase
}

// PhaseDifference calculates the minimal phase difference between two phases
// Returns a value in [-π, π]
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

// MeasureCoherence calculates the Kuramoto order parameter for phase synchronization
func MeasureCoherence(phases []float64) float64 {
	if len(phases) == 0 {
		return 0
	}

	sumCos := 0.0
	sumSin := 0.0

	for _, phase := range phases {
		sumCos += math.Cos(phase)
		sumSin += math.Sin(phase)
	}

	n := float64(len(phases))
	return math.Sqrt(sumCos*sumCos+sumSin*sumSin) / n
}

// MeasureCoherenceWeighted calculates weighted coherence
func MeasureCoherenceWeighted(phases []float64, weights []float64) float64 {
	if len(phases) == 0 || len(weights) == 0 {
		return 0
	}

	sumCos := 0.0
	sumSin := 0.0
	sumWeights := 0.0

	minLen := len(phases)
	if len(weights) < minLen {
		minLen = len(weights)
	}

	for i := range minLen {
		w := weights[i]
		sumCos += w * math.Cos(phases[i])
		sumSin += w * math.Sin(phases[i])
		sumWeights += w
	}

	if sumWeights == 0 {
		return 0
	}

	return math.Sqrt(sumCos*sumCos+sumSin*sumSin) / sumWeights
}