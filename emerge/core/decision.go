package core

// DecisionMaker allows agents to make autonomous choices.
// This interface enables upgrading from simple to complex decision-making
// (e.g., machine learning models) without changing the agent structure.
type DecisionMaker interface {
	// Decide chooses an action based on current state and options.
	// Returns confidence level [0, 1] with the decision.
	Decide(state State, options []Action) (Action, float64)
}
