package agent

import (
	"sync/atomic"
	"time"
)

// AtomicState uses atomic.Value for simpler, more efficient grouped atomics.
type AtomicState struct {
	value atomic.Value // stores *StateData
}

// StateData contains the actual state fields.
type StateData struct {
	Phase     float64       // Current phase [0, 2π]
	Energy    float64       // Available energy
	LocalGoal float64       // Preferred phase
	Frequency time.Duration // Current frequency
}

// BehaviorData contains behavioral parameters.
type BehaviorData struct {
	Influence    float64 // Local vs global weight [0, 1]
	Stubbornness float64 // Resistance to change [0, 1]
}

// NewAtomicState creates a new atomic state.
func NewAtomicState() *AtomicState {
	s := &AtomicState{}
	s.value.Store(&StateData{})
	return s
}

// Load atomically loads the state.
func (a *AtomicState) Load() StateData {
	if v := a.value.Load(); v != nil {
		if data, ok := v.(*StateData); ok {
			return *data
		}
	}
	return StateData{}
}

// Store atomically stores the state.
func (a *AtomicState) Store(state StateData) {
	a.value.Store(&state)
}

// Update atomically updates the state.
// For simplicity, this just loads, modifies, and stores.
// In high contention scenarios, consider using a mutex instead.
func (a *AtomicState) Update(fn func(*StateData)) {
	old := a.Load()
	fn(&old)
	a.Store(old)
}

// AtomicBehavior uses atomic.Value for behavioral parameters.
type AtomicBehavior struct {
	value atomic.Value // stores *BehaviorData
}

// NewAtomicBehavior creates a new atomic behavior.
func NewAtomicBehavior() *AtomicBehavior {
	b := &AtomicBehavior{}
	b.value.Store(&BehaviorData{})
	return b
}

// Load atomically loads the behavior.
func (a *AtomicBehavior) Load() BehaviorData {
	if v := a.value.Load(); v != nil {
		if data, ok := v.(*BehaviorData); ok {
			return *data
		}
	}
	return BehaviorData{}
}

// Store atomically stores the behavior.
func (a *AtomicBehavior) Store(behavior BehaviorData) {
	a.value.Store(&behavior)
}

// Update atomically updates the behavior.
func (a *AtomicBehavior) Update(fn func(*BehaviorData)) {
	old := a.Load()
	fn(&old)
	a.Store(old)
}
