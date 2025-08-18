package biofield

import "time"

// State represents a system configuration that agents work toward.
// In biological terms, this is like a morphological target.
type State struct {
	Phase     float64       // Target phase in radians [0, 2π]
	Frequency time.Duration // Target oscillation period
	Coherence float64       // Target synchronization level [0, 1]
}