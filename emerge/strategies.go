package emerge

import (
	"math"
	"time"
)

// SyncStrategy defines how agents adjust toward synchronization.
// Different strategies represent different biological synchronization mechanisms.
type SyncStrategy interface {
	// Propose suggests an action to move toward the target state.
	// Returns the proposed action and a confidence level [0, 1].
	Propose(current, target State, context Context) (Action, float64)

	// Name returns the strategy's identifier.
	Name() string
}

// PhaseNudgeStrategy gently adjusts phase toward target.
// This mimics gradual biological entrainment like circadian rhythm adjustment.
type PhaseNudgeStrategy struct {
	Rate float64 // Adjustment rate [0, 1]
}

// NewPhaseNudgeStrategy creates a phase nudging strategy.
func NewPhaseNudgeStrategy(rate float64) *PhaseNudgeStrategy {
	return &PhaseNudgeStrategy{
		Rate: math.Max(0, math.Min(1, rate)),
	}
}

func (s *PhaseNudgeStrategy) Propose(current, target State, context Context) (Action, float64) {
	// Calculate phase difference
	diff := PhaseDifference(target.Phase, current.Phase)

	// Adjust by rate factor
	adjustment := diff * s.Rate

	// Higher confidence when we need to make changes (lower coherence)
	// Use 1 - LocalCoherence so we're more confident when less synchronized
	confidence := math.Max(0.5, 1.0-context.LocalCoherence)

	return Action{
		Type:    "phase_nudge",
		Value:   adjustment,
		Cost:    math.Abs(adjustment) * 2.0,           // Energy cost proportional to change
		Benefit: (1.0 - math.Abs(diff)/math.Pi) * 1.5, // Increase benefit to encourage convergence
	}, confidence
}

func (s *PhaseNudgeStrategy) Name() string {
	return "phase_nudge"
}

// FrequencyLockStrategy synchronizes oscillation frequency.
// This represents frequency locking in biological oscillators.
type FrequencyLockStrategy struct {
	Strength float64 // Locking strength [0, 1]
}

// NewFrequencyLockStrategy creates a frequency locking strategy.
func NewFrequencyLockStrategy(strength float64) *FrequencyLockStrategy {
	return &FrequencyLockStrategy{
		Strength: math.Max(0, math.Min(1, strength)),
	}
}

func (s *FrequencyLockStrategy) Propose(current, target State, context Context) (Action, float64) {
	// For now, we focus on phase since frequency is less dynamic
	// In a full implementation, this would adjust oscillation frequency

	// Calculate phase adjustment to achieve frequency lock
	diff := PhaseDifference(target.Phase, current.Phase)

	// Stronger adjustment for frequency locking
	adjustment := diff * s.Strength

	// Confidence based on local coherence
	confidence := context.LocalCoherence * s.Strength

	return Action{
		Type:    "frequency_lock",
		Value:   adjustment,
		Cost:    math.Abs(adjustment) * 3.0,   // Higher cost for frequency changes
		Benefit: context.LocalCoherence * 2.0, // Higher benefit when neighbors are coherent
	}, confidence
}

func (s *FrequencyLockStrategy) Name() string {
	return "frequency_lock"
}

// EnergyAwareStrategy conserves energy while synchronizing.
// This mimics biological systems that must balance synchronization with resource conservation.
type EnergyAwareStrategy struct {
	Threshold float64 // Minimum energy threshold for action
}

// NewEnergyAwareStrategy creates an energy-conscious strategy.
func NewEnergyAwareStrategy(threshold float64) *EnergyAwareStrategy {
	return &EnergyAwareStrategy{
		Threshold: math.Max(0, threshold),
	}
}

func (s *EnergyAwareStrategy) Propose(current, target State, context Context) (Action, float64) {
	// Calculate phase difference
	diff := PhaseDifference(target.Phase, current.Phase)

	// Only act if difference is significant
	if math.Abs(diff) < 0.1 {
		return Action{
			Type:    "maintain",
			Value:   0,
			Cost:    0.1, // Small maintenance cost
			Benefit: context.Stability,
		}, 0.5
	}

	// Conservative adjustment based on available resources
	adjustment := diff * 0.1 // Very conservative

	// Low confidence to preserve energy
	confidence := 0.3

	return Action{
		Type:    "energy_save",
		Value:   adjustment,
		Cost:    math.Max(0.5, math.Abs(adjustment)), // Minimum cost
		Benefit: context.Progress * 0.5,              // Reduced benefit
	}, confidence
}

func (s *EnergyAwareStrategy) Name() string {
	return "energy_aware"
}

// AdaptiveStrategy switches between strategies based on context.
// This represents biological systems that change behavior based on conditions.
type AdaptiveStrategy struct {
	strategies []SyncStrategy
	selector   func(Context) int // Returns index of strategy to use
}

// NewAdaptiveStrategy creates a strategy that adapts to context.
func NewAdaptiveStrategy(strategies []SyncStrategy) *AdaptiveStrategy {
	return &AdaptiveStrategy{
		strategies: strategies,
		selector: func(ctx Context) int {
			// Default selection logic based on context
			if ctx.Stability > 0.7 {
				return 0 // Use first strategy when stable
			} else if ctx.LocalCoherence > 0.5 {
				return min(1, len(strategies)-1) // Use second when locally coherent
			}
			return min(2, len(strategies)-1) // Use third otherwise
		},
	}
}

func (s *AdaptiveStrategy) Propose(current, target State, context Context) (Action, float64) {
	if len(s.strategies) == 0 {
		return Action{Type: "maintain"}, 0.5
	}

	// Select strategy based on context
	idx := s.selector(context)
	if idx < 0 || idx >= len(s.strategies) {
		idx = 0
	}

	return s.strategies[idx].Propose(current, target, context)
}

func (s *AdaptiveStrategy) Name() string {
	return "adaptive"
}

// PulseStrategy sends periodic synchronization pulses.
// This mimics pacemaker cells in biological systems.
type PulseStrategy struct {
	Period    time.Duration
	Amplitude float64
	lastPulse time.Time
}

// NewPulseStrategy creates a pulsing synchronization strategy.
func NewPulseStrategy(period time.Duration, amplitude float64) *PulseStrategy {
	return &PulseStrategy{
		Period:    period,
		Amplitude: math.Max(0, math.Min(1, amplitude)),
		lastPulse: time.Now(),
	}
}

func (s *PulseStrategy) Propose(current, target State, context Context) (Action, float64) {
	now := time.Now()
	timeSincePulse := now.Sub(s.lastPulse)

	// Check if it's time for a pulse
	if timeSincePulse < s.Period {
		// Between pulses - maintain
		return Action{
			Type:    "maintain",
			Value:   0,
			Cost:    0.1,
			Benefit: context.Stability * 0.5,
		}, 0.3
	}

	// Time for a pulse!
	s.lastPulse = now

	// Strong adjustment toward target
	diff := PhaseDifference(target.Phase, current.Phase)
	adjustment := diff * s.Amplitude

	return Action{
		Type:    "pulse",
		Value:   adjustment,
		Cost:    math.Abs(adjustment) * 4.0, // High cost for pulses
		Benefit: 2.0,                        // High benefit to overcome cost
	}, s.Amplitude
}

func (s *PulseStrategy) Name() string {
	return "pulse"
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
