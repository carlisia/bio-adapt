package emerge

import "math"

// GoalManager blends local and global objectives hierarchically.
// This creates multi-scale goal structure where individual goals
// contribute to but don't override collective goals.
type GoalManager interface {
	// Blend combines local and global goals based on context.
	// Weight parameter determines local (0) vs global (1) influence.
	Blend(local, global State, weight float64) State
}

// WeightedGoalManager blends local and global goals hierarchically.
type WeightedGoalManager struct{}

// Blend combines goals with smooth weighting.
func (w *WeightedGoalManager) Blend(local, global State, weight float64) State {
	// Ensure weight is in [0, 1]
	weight = math.Max(0, math.Min(1, weight))

	// Blend phases using circular interpolation
	localPhase := local.Phase
	globalPhase := global.Phase

	// Handle phase wrapping
	diff := globalPhase - localPhase
	for diff > math.Pi {
		diff -= 2 * math.Pi
	}
	for diff < -math.Pi {
		diff += 2 * math.Pi
	}

	blendedPhase := math.Mod(localPhase+weight*diff, 2*math.Pi)

	return State{
		Phase:     blendedPhase,
		Frequency: local.Frequency, // Keep local frequency for now
		Coherence: weight*global.Coherence + (1-weight)*local.Coherence,
	}
}
