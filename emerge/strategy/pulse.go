package strategy

import (
	"math"
	"time"

	"github.com/carlisia/bio-adapt/emerge/core"
)

// Pulse sends periodic synchronization pulses.
// This mimics pacemaker cells in biological systems.
type Pulse struct {
	Period    time.Duration
	Amplitude float64
	lastPulse time.Time
}

// NewPulse creates a pulsing synchronization strategy.
func NewPulse(period time.Duration, amplitude float64) *Pulse {
	return &Pulse{
		Period:    period,
		Amplitude: math.Max(0, math.Min(1, amplitude)),
		lastPulse: time.Now(),
	}
}

// Propose suggests a pulse action when appropriate.
func (s *Pulse) Propose(current, target core.State, context core.Context) (core.Action, float64) {
	now := time.Now()
	timeSincePulse := now.Sub(s.lastPulse)

	// Check if it's time for a pulse
	if timeSincePulse < s.Period {
		// Between pulses - maintain
		return core.Action{
			Type:    "maintain",
			Value:   0,
			Cost:    0.1,
			Benefit: context.Stability * 0.5,
		}, 0.3
	}

	// Time for a pulse!
	s.lastPulse = now

	// Strong adjustment toward target
	diff := core.PhaseDifference(target.Phase, current.Phase)
	adjustment := diff * s.Amplitude

	return core.Action{
		Type:    "pulse",
		Value:   adjustment,
		Cost:    math.Abs(adjustment) * 4.0, // High cost for pulses
		Benefit: 2.0,                        // High benefit to overcome cost
	}, s.Amplitude
}

// Name returns the strategy's identifier.
func (*Pulse) Name() string {
	return "pulse"
}
