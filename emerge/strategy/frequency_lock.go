package strategy

import (
	"math"

	"github.com/carlisia/bio-adapt/emerge/core"
)

// FrequencyLock synchronizes oscillation frequency.
// This represents frequency locking in biological oscillators.
type FrequencyLock struct {
	SyncRate float64 // Locking strength [0, 1]
}

// NewFrequencyLock creates a frequency locking strategy.
func NewFrequencyLock(syncRate float64) *FrequencyLock {
	return &FrequencyLock{
		SyncRate: math.Max(0, math.Min(1, syncRate)),
	}
}

// Propose suggests a frequency locking action.
func (s *FrequencyLock) Propose(current, target core.State, context core.Context) (core.Action, float64) {
	// For now, we focus on phase since frequency is less dynamic
	// In a full implementation, this would adjust oscillation frequency

	// Calculate phase adjustment to achieve frequency lock
	diff := core.PhaseDifference(target.Phase, current.Phase)

	// Stronger adjustment for frequency locking
	adjustment := diff * s.SyncRate

	// Confidence based on local coherence
	confidence := context.LocalCoherence * s.SyncRate

	return core.Action{
		Type:    "frequency_lock",
		Value:   adjustment,
		Cost:    math.Abs(adjustment) * 3.0,   // Higher cost for frequency changes
		Benefit: context.LocalCoherence * 2.0, // Higher benefit when neighbors are coherent
	}, confidence
}

// Name returns the strategy's identifier.
func (*FrequencyLock) Name() string {
	return "frequency_lock"
}
