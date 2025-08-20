package strategy

import "github.com/carlisia/bio-adapt/emerge/core"

// AdaptiveStrategy switches between strategies based on context.
// This represents biological systems that change behavior based on conditions.
type adaptiveStrategy struct {
	strategies []core.SyncStrategy
	selector   func(core.Context) int // Returns index of strategy to use
}

// NewAdaptiveStrategy creates a strategy that adapts to context.
func newAdaptiveStrategy(strategies []core.SyncStrategy) *adaptiveStrategy {
	return &adaptiveStrategy{
		strategies: strategies,
		selector: func(ctx core.Context) int {
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

// WithSelector allows custom strategy selection logic.
func (s *adaptiveStrategy) WithSelector(selector func(core.Context) int) *adaptiveStrategy {
	s.selector = selector
	return s
}

// Propose delegates to the selected strategy based on context.
func (s *adaptiveStrategy) Propose(current, target core.State, context core.Context) (core.Action, float64) {
	if len(s.strategies) == 0 {
		return core.Action{Type: "maintain"}, 0.5
	}

	// Select strategy based on context
	idx := s.selector(context)
	if idx < 0 || idx >= len(s.strategies) {
		idx = 0
	}

	return s.strategies[idx].Propose(current, target, context)
}

// Name returns the strategy's identifier.
func (s *adaptiveStrategy) Name() string {
	return "adaptive"
}

// Helper function for min.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
