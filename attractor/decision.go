package attractor

import "math"

// DecisionMaker allows agents to make autonomous choices.
// This interface enables upgrading from simple to complex decision-making
// (e.g., neural networks) without changing the agent structure.
type DecisionMaker interface {
	// Decide chooses an action based on current state and options.
	// Returns confidence level [0, 1] with the decision.
	Decide(state State, options []Action) (Action, float64)
}

// SimpleDecisionMaker provides basic autonomous decision-making.
// Chooses actions based on benefit/cost ratio with context weighting.
type SimpleDecisionMaker struct{}

// Decide selects action with best benefit/cost ratio adjusted by context.
func (s *SimpleDecisionMaker) Decide(state State, options []Action) (Action, float64) {
	if len(options) == 0 {
		return Action{Type: "maintain"}, 0.5
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
