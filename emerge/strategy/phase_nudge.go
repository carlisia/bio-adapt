package strategy

import (
	"math"

	"github.com/carlisia/bio-adapt/emerge/core"
)

// PhaseNudge gently adjusts phase toward target.
// This mimics gradual biological entrainment like circadian rhythm adjustment.
type PhaseNudge struct {
	Rate float64 // Adjustment rate [0, 1]
}

// NewPhaseNudge creates a phase nudging strategy.
func NewPhaseNudge(rate float64) *PhaseNudge {
	return &PhaseNudge{
		Rate: math.Max(0, math.Min(1, rate)),
	}
}

// Propose suggests a phase adjustment action.
func (s *PhaseNudge) Propose(current, target core.State, context core.Context) (core.Action, float64) {
	// Calculate phase difference
	diff := core.PhaseDifference(target.Phase, current.Phase)

	// Adjust by rate factor
	adjustment := diff * s.Rate

	// Higher confidence when we need to make changes (lower coherence)
	// Use 1 - LocalCoherence so we're more confident when less synchronized
	confidence := math.Max(0.5, 1.0-context.LocalCoherence)

	return core.Action{
		Type:    "phase_nudge",
		Value:   adjustment,
		Cost:    math.Abs(adjustment) * 2.0,           // Energy cost proportional to change
		Benefit: (1.0 - math.Abs(diff)/math.Pi) * 1.5, // Increase benefit to encourage convergence
	}, confidence
}

// Name returns the strategy's identifier.
func (s *PhaseNudge) Name() string {
	return "phase_nudge"
}
