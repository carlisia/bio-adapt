package strategy

import (
	"math"

	"github.com/carlisia/bio-adapt/emerge/core"
)

// EnergyAware conserves energy while synchronizing.
// This implements resource-aware synchronization that balances coordination with efficiency.
type EnergyAware struct {
	Threshold float64 // Minimum energy threshold for action
}

// NewEnergyAware creates an energy-conscious strategy.
func NewEnergyAware(threshold float64) *EnergyAware {
	return &EnergyAware{
		Threshold: math.Max(0, threshold),
	}
}

// Propose suggests an energy-aware action.
func (*EnergyAware) Propose(current, target core.State, context core.Context) (core.Action, float64) {
	// Calculate phase difference
	diff := core.PhaseDifference(target.Phase, current.Phase)

	// Only act if difference is significant
	if math.Abs(diff) < 0.1 {
		return core.Action{
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

	return core.Action{
		Type:    "energy_save",
		Value:   adjustment,
		Cost:    math.Max(0.5, math.Abs(adjustment)), // Minimum cost
		Benefit: context.Progress * 0.5,              // Reduced benefit
	}, confidence
}

// Name returns the strategy's identifier.
func (*EnergyAware) Name() string {
	return "energy_aware"
}
