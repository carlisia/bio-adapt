package decision

import (
	"github.com/carlisia/bio-adapt/emerge/core"
	"math"
)

// SimpleDecisionMaker provides basic autonomous decision-making.
// Chooses actions based on benefit/cost ratio with context weighting.
type SimpleDecisionMaker struct{}

// Decide selects action with best benefit/cost ratio adjusted by context.
func (s *SimpleDecisionMaker) Decide(state core.State, options []core.Action) (core.Action, float64) {
	if len(options) == 0 {
		return core.Action{Type: "maintain"}, state.Coherence
	}

	bestAction := options[0]
	bestScore := -math.MaxFloat64

	for _, action := range options {
		// Avoid division by zero
		cost := math.Max(action.Cost, 0.1)
		score := action.Benefit / cost

		if score > bestScore {
			bestScore = score
			bestAction = action
		}
	}

	// Confidence based on score magnitude, but ensure non-negative
	confidence := math.Min(math.Max(bestScore/2.0, 0), 1.0)

	return bestAction, confidence
}
