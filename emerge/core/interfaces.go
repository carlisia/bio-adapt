package core

// DecisionMaker allows agents to make autonomous choices.
// This interface enables upgrading from simple to complex decision-making
// (e.g., neural networks) without changing the agent structure.
type DecisionMaker interface {
	// Decide chooses an action based on current state and options.
	// Returns confidence level [0, 1] with the decision.
	Decide(state State, options []Action) (Action, float64)
}

// ResourceManager handles energy/resource allocation.
// This implements metabolic-like constraints where actions have costs.
type ResourceManager interface {
	// Request attempts to allocate resources for an action.
	// Returns actual amount available (may be less than requested).
	Request(amount float64) float64

	// Release returns unused resources to the pool.
	Release(amount float64)

	// Available returns current resource level.
	Available() float64
}

// SyncStrategy defines how agents adjust toward synchronization.
// Different strategies represent different biological synchronization mechanisms.
type SyncStrategy interface {
	// Propose suggests an action to move toward the target state.
	// Returns the proposed action and a confidence level [0, 1].
	Propose(current, target State, context Context) (Action, float64)

	// Name returns the strategy's identifier.
	Name() string
}
