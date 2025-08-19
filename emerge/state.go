package emerge

import "time"

// State represents a system configuration that agents work toward.
// In biological terms, this is like a morphological target.
type State struct {
	Phase     float64       // Target phase in radians [0, 2π]
	Frequency time.Duration // Target oscillation period
	Coherence float64       // Target synchronization level [0, 1]
}

// StateUpdate represents a change in agent state for gossip protocol.
type StateUpdate struct {
	AgentID   string
	FromID    string // ID of the agent who sent the update
	Phase     float64
	Frequency time.Duration
	Energy    float64 // Energy level of the agent
	Timestamp time.Time
}
