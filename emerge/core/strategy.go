package core

// SyncStrategy defines how agents adjust toward synchronization.
// Different strategies represent different synchronization approaches.
type SyncStrategy interface {
	// Propose suggests an action to move toward the target state.
	// Returns the proposed action and a confidence level [0, 1].
	Propose(current, target State, context Context) (Action, float64)

	// Name returns the strategy's identifier.
	Name() string
}
