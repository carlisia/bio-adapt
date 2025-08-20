package goal_test

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/goal"
)

func TestWeightedGoalManagerBlend(t *testing.T) {
	tests := []struct {
		name              string
		local             core.State
		global            core.State
		weight            float64
		expectedPhase     float64
		expectedCoherence float64
		phaseTolerance    float64
		cohTolerance      float64
		description       string
	}{
		// Happy path cases
		{
			name: "pure local (weight = 0)",
			local: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.3,
			},
			global: core.State{
				Phase:     math.Pi,
				Frequency: 200 * time.Millisecond,
				Coherence: 0.9,
			},
			weight:            0,
			expectedPhase:     0,
			expectedCoherence: 0.3,
			phaseTolerance:    0.01,
			cohTolerance:      0.01,
			description:       "Weight 0 should give pure local state",
		},
		{
			name: "pure global (weight = 1)",
			local: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.3,
			},
			global: core.State{
				Phase:     math.Pi,
				Frequency: 200 * time.Millisecond,
				Coherence: 0.9,
			},
			weight:            1,
			expectedPhase:     math.Pi,
			expectedCoherence: 0.9,
			phaseTolerance:    0.01,
			cohTolerance:      0.01,
			description:       "Weight 1 should give pure global state",
		},
		{
			name: "50/50 blend (weight = 0.5)",
			local: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.3,
			},
			global: core.State{
				Phase:     math.Pi,
				Frequency: 200 * time.Millisecond,
				Coherence: 0.9,
			},
			weight:            0.5,
			expectedPhase:     math.Pi / 2,
			expectedCoherence: 0.6,
			phaseTolerance:    0.01,
			cohTolerance:      0.01,
			description:       "Weight 0.5 should blend equally",
		},
		{
			name: "25/75 blend (weight = 0.25)",
			local: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.2,
			},
			global: core.State{
				Phase:     math.Pi,
				Frequency: 200 * time.Millisecond,
				Coherence: 0.8,
			},
			weight:            0.25,
			expectedPhase:     math.Pi / 4,
			expectedCoherence: 0.35,
			phaseTolerance:    0.01,
			cohTolerance:      0.01,
			description:       "Weight 0.25 should favor local",
		},
		{
			name: "75/25 blend (weight = 0.75)",
			local: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.2,
			},
			global: core.State{
				Phase:     math.Pi,
				Frequency: 200 * time.Millisecond,
				Coherence: 0.8,
			},
			weight:            0.75,
			expectedPhase:     3 * math.Pi / 4,
			expectedCoherence: 0.65,
			phaseTolerance:    0.01,
			cohTolerance:      0.01,
			description:       "Weight 0.75 should favor global",
		},
		// Edge cases
		{
			name: "identical states",
			local: core.State{
				Phase:     math.Pi / 2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			global: core.State{
				Phase:     math.Pi / 2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			weight:            0.5,
			expectedPhase:     math.Pi / 2,
			expectedCoherence: 0.5,
			phaseTolerance:    0.01,
			cohTolerance:      0.01,
			description:       "Identical states should remain unchanged",
		},
		{
			name: "zero coherence blend",
			local: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0,
			},
			global: core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0,
			},
			weight:            0.5,
			expectedPhase:     math.Pi / 2,
			expectedCoherence: 0,
			phaseTolerance:    0.01,
			cohTolerance:      0.01,
			description:       "Zero coherence should blend to zero",
		},
		{
			name: "maximum coherence blend",
			local: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 1.0,
			},
			global: core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 1.0,
			},
			weight:            0.5,
			expectedPhase:     math.Pi / 2,
			expectedCoherence: 1.0,
			phaseTolerance:    0.01,
			cohTolerance:      0.01,
			description:       "Maximum coherence should blend to maximum",
		},
		{
			name: "very small weight",
			local: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.1,
			},
			global: core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			weight:            0.001,
			expectedPhase:     0.001 * math.Pi,
			expectedCoherence: 0.1008,
			phaseTolerance:    0.01,
			cohTolerance:      0.01,
			description:       "Very small weight should barely affect local",
		},
		{
			name: "very large weight",
			local: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.1,
			},
			global: core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			weight:            0.999,
			expectedPhase:     0.999 * math.Pi,
			expectedCoherence: 0.8992,
			phaseTolerance:    0.01,
			cohTolerance:      0.01,
			description:       "Very large weight should nearly match global",
		},
		{
			name: "negative phases",
			local: core.State{
				Phase:     -math.Pi / 2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			global: core.State{
				Phase:     math.Pi / 2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			weight:            0.5,
			expectedPhase:     0,
			expectedCoherence: 0.5,
			phaseTolerance:    0.01,
			cohTolerance:      0.01,
			description:       "Should handle negative phases",
		},
		{
			name: "large phase values",
			local: core.State{
				Phase:     10 * math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			global: core.State{
				Phase:     11 * math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			weight:            0.5,
			expectedPhase:     10.5 * math.Pi,
			expectedCoherence: 0.5,
			phaseTolerance:    0.01,
			cohTolerance:      0.01,
			description:       "Should handle large phase values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gm := &goal.WeightedManager{}
			blended := gm.Blend(tt.local, tt.global, tt.weight)

			// Check phase
			phaseDiff := math.Abs(blended.Phase - tt.expectedPhase)
			// Handle phase wrapping
			if phaseDiff > math.Pi {
				phaseDiff = 2*math.Pi - phaseDiff
			}
			assert.LessOrEqual(t, phaseDiff, tt.phaseTolerance, "%s: Phase = %f, expected %f±%f", tt.description, blended.Phase, tt.expectedPhase, tt.phaseTolerance)

			// Check coherence
			assert.InDelta(t, tt.expectedCoherence, blended.Coherence, tt.cohTolerance, "%s: Coherence should match expected", tt.description)

			// Check frequency preservation
			assert.Equal(t, tt.local.Frequency, blended.Frequency, "%s: Frequency should be preserved from local state", tt.description)
		})
	}
}

func TestWeightedGoalManagerPhaseWrapping(t *testing.T) {
	tests := []struct {
		name       string
		local      core.State
		global     core.State
		weight     float64
		validateFn func(t *testing.T, blended core.State)
	}{
		{
			name: "phase wrapping across 0/2π boundary",
			local: core.State{
				Phase:     0.1,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			global: core.State{
				Phase:     2*math.Pi - 0.1,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			weight: 0.5,
			validateFn: func(t *testing.T, blended core.State) {
				// Shortest path from 0.1 to 2π-0.1 is backwards across 0
				// So halfway should be around 0 (or 2π)
				assert.True(t, blended.Phase <= 0.2 || blended.Phase >= 2*math.Pi-0.2, "Phase wrapping not working correctly, got %f", blended.Phase)
			},
		},
		{
			name: "phase wrapping at π boundary",
			local: core.State{
				Phase:     math.Pi - 0.1,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			global: core.State{
				Phase:     math.Pi + 0.1,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			weight: 0.5,
			validateFn: func(t *testing.T, blended core.State) {
				// Should be very close to π
				assert.InDelta(t, math.Pi, blended.Phase, 0.01, "Expected phase near π, got %f", blended.Phase)
			},
		},
		{
			name: "large phase difference (opposite sides)",
			local: core.State{
				Phase:     math.Pi / 4,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			global: core.State{
				Phase:     5 * math.Pi / 4,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			weight: 0.5,
			validateFn: func(t *testing.T, blended core.State) {
				// Should take shortest path
				expected := 3 * math.Pi / 4
				assert.InDelta(t, expected, blended.Phase, 0.01, "Expected phase %f, got %f", expected, blended.Phase)
			},
		},
		{
			name: "wrapping with negative phase",
			local: core.State{
				Phase:     -0.1,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			global: core.State{
				Phase:     0.1,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			weight: 0.5,
			validateFn: func(t *testing.T, blended core.State) {
				// Should be around 0
				assert.True(t, math.Abs(blended.Phase) <= 0.01 || math.Abs(blended.Phase-2*math.Pi) <= 0.01, "Expected phase near 0, got %f", blended.Phase)
			},
		},
		{
			name: "multiple wraps",
			local: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			global: core.State{
				Phase:     4 * math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			weight: 0.5,
			validateFn: func(t *testing.T, blended core.State) {
				// 4π is same as 0, so blend should be 0
				assert.True(t, math.Abs(blended.Phase) <= 0.01 || math.Abs(blended.Phase-2*math.Pi) <= 0.01, "Expected phase near 0, got %f", blended.Phase)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gm := &goal.WeightedManager{}
			blended := gm.Blend(tt.local, tt.global, tt.weight)
			tt.validateFn(t, blended)
		})
	}
}

func TestWeightedGoalManagerWeightClamping(t *testing.T) {
	tests := []struct {
		name              string
		local             core.State
		global            core.State
		weight            float64
		expectedPhase     float64
		expectedCoherence float64
		description       string
	}{
		{
			name: "weight > 1 (should clamp to 1)",
			local: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			global: core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			weight:            1.5,
			expectedPhase:     math.Pi,
			expectedCoherence: 0.5,
			description:       "Weight > 1 should be clamped to 1",
		},
		{
			name: "weight < 0 (should clamp to 0)",
			local: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			global: core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			weight:            -0.5,
			expectedPhase:     0,
			expectedCoherence: 0.5,
			description:       "Weight < 0 should be clamped to 0",
		},
		{
			name: "very large positive weight",
			local: core.State{
				Phase:     math.Pi / 4,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.3,
			},
			global: core.State{
				Phase:     3 * math.Pi / 4,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.7,
			},
			weight:            100.0,
			expectedPhase:     3 * math.Pi / 4,
			expectedCoherence: 0.7,
			description:       "Very large weight should clamp to 1",
		},
		{
			name: "very large negative weight",
			local: core.State{
				Phase:     math.Pi / 4,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.3,
			},
			global: core.State{
				Phase:     3 * math.Pi / 4,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.7,
			},
			weight:            -100.0,
			expectedPhase:     math.Pi / 4,
			expectedCoherence: 0.3,
			description:       "Very negative weight should clamp to 0",
		},
		{
			name: "infinity weight",
			local: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.2,
			},
			global: core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			weight:            math.Inf(1),
			expectedPhase:     math.Pi,
			expectedCoherence: 0.8,
			description:       "Infinity should clamp to 1",
		},
		{
			name: "negative infinity weight",
			local: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.2,
			},
			global: core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			weight:            math.Inf(-1),
			expectedPhase:     0,
			expectedCoherence: 0.2,
			description:       "Negative infinity should clamp to 0",
		},
		{
			name: "NaN weight (edge case)",
			local: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			global: core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			weight:            math.NaN(),
			expectedPhase:     0, // Implementation specific - might default to local
			expectedCoherence: 0.5,
			description:       "NaN weight behavior",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gm := &goal.WeightedManager{}
			blended := gm.Blend(tt.local, tt.global, tt.weight)

			// Skip NaN test if result is also NaN
			if math.IsNaN(tt.weight) && math.IsNaN(blended.Phase) {
				t.Skip("NaN weight produces NaN result")
			}

			assert.InDelta(t, tt.expectedPhase, blended.Phase, 0.01, "%s: Phase should match expected", tt.description)
			assert.InDelta(t, tt.expectedCoherence, blended.Coherence, 0.01, "%s: Coherence should match expected", tt.description)
		})
	}
}

func TestWeightedGoalManagerSymmetry(t *testing.T) {
	tests := []struct {
		name   string
		state1 core.State
		state2 core.State
		weight float64
	}{
		{
			name: "basic symmetry",
			state1: core.State{
				Phase:     math.Pi / 4,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.4,
			},
			state2: core.State{
				Phase:     3 * math.Pi / 4,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.6,
			},
			weight: 0.3,
		},
		{
			name: "opposite phases",
			state1: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.2,
			},
			state2: core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			weight: 0.4,
		},
		{
			name: "near boundary",
			state1: core.State{
				Phase:     0.1,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			state2: core.State{
				Phase:     2*math.Pi - 0.1,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			weight: 0.2,
		},
		{
			name: "equal coherence",
			state1: core.State{
				Phase:     math.Pi / 3,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.7,
			},
			state2: core.State{
				Phase:     2 * math.Pi / 3,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.7,
			},
			weight: 0.6,
		},
		{
			name: "zero coherence",
			state1: core.State{
				Phase:     1.0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0,
			},
			state2: core.State{
				Phase:     2.0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0,
			},
			weight: 0.45,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gm := &goal.WeightedManager{}

			// Blend 1->2 with weight w
			blend12 := gm.Blend(tt.state1, tt.state2, tt.weight)

			// Blend 2->1 with weight (1-w)
			blend21 := gm.Blend(tt.state2, tt.state1, 1-tt.weight)

			// Results should be the same
			phaseDiff := math.Abs(blend12.Phase - blend21.Phase)
			// Handle phase wrapping
			if phaseDiff > math.Pi {
				phaseDiff = 2*math.Pi - phaseDiff
			}

			assert.LessOrEqual(t, phaseDiff, 0.01, "Blending not symmetric for phase: %f vs %f (diff: %f)", blend12.Phase, blend21.Phase, phaseDiff)
			assert.InDelta(t, blend12.Coherence, blend21.Coherence, 0.01, "Blending not symmetric for coherence: %f vs %f", blend12.Coherence, blend21.Coherence)

			// Note: Frequency might not be symmetric since it's taken from local state
		})
	}
}

func TestWeightedGoalManagerEdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		local      core.State
		global     core.State
		weight     float64
		validateFn func(t *testing.T, blended core.State)
	}{
		{
			name: "zero frequency",
			local: core.State{
				Phase:     0,
				Frequency: 0,
				Coherence: 0.5,
			},
			global: core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			weight: 0.5,
			validateFn: func(t *testing.T, blended core.State) {
				assert.Equal(t, time.Duration(0), blended.Frequency, "Frequency should be preserved from local state")
			},
		},
		{
			name: "negative frequency (invalid)",
			local: core.State{
				Phase:     0,
				Frequency: -100 * time.Millisecond,
				Coherence: 0.5,
			},
			global: core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			weight: 0.5,
			validateFn: func(t *testing.T, blended core.State) {
				assert.Equal(t, -100*time.Millisecond, blended.Frequency, "Frequency should be preserved even if negative")
			},
		},
		{
			name: "coherence > 1 (invalid)",
			local: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 1.5,
			},
			global: core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 2.0,
			},
			weight: 0.5,
			validateFn: func(t *testing.T, blended core.State) {
				// Should blend even invalid coherence values
				expected := 1.75
				assert.InDelta(t, expected, blended.Coherence, 0.01, "Coherence should match expected")
			},
		},
		{
			name: "negative coherence (invalid)",
			local: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: -0.5,
			},
			global: core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			weight: 0.5,
			validateFn: func(t *testing.T, blended core.State) {
				// Should blend to 0
				assert.InDelta(t, 0.0, blended.Coherence, 0.01, "Coherence should be 0")
			},
		},
		{
			name: "all zero state",
			local: core.State{
				Phase:     0,
				Frequency: 0,
				Coherence: 0,
			},
			global: core.State{
				Phase:     0,
				Frequency: 0,
				Coherence: 0,
			},
			weight: 0.5,
			validateFn: func(t *testing.T, blended core.State) {
				assert.Equal(t, 0.0, blended.Phase, "Phase should be 0")
				assert.Equal(t, time.Duration(0), blended.Frequency, "Frequency should be 0")
				assert.Equal(t, 0.0, blended.Coherence, "Coherence should be 0")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gm := &goal.WeightedManager{}
			blended := gm.Blend(tt.local, tt.global, tt.weight)
			tt.validateFn(t, blended)
		})
	}
}

func TestWeightedGoalManagerConcurrency(t *testing.T) {
	gm := &goal.WeightedManager{}

	local := core.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.5,
	}

	global := core.State{
		Phase:     math.Pi,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.5,
	}

	// Run concurrent blends
	done := make(chan bool, 100)
	for i := range 100 {
		go func(weight float64) {
			_ = gm.Blend(local, global, weight)
			done <- true
		}(float64(i) / 100.0)
	}

	// Wait for all goroutines
	for range 100 {
		<-done
	}

	// If we get here without panic, concurrent access is safe
}

func BenchmarkWeightedGoalManager(b *testing.B) {
	gm := &goal.WeightedManager{}

	local := core.State{
		Phase:     rand.Float64() * 2 * math.Pi,
		Frequency: 100 * time.Millisecond,
		Coherence: rand.Float64(),
	}

	global := core.State{
		Phase:     rand.Float64() * 2 * math.Pi,
		Frequency: 100 * time.Millisecond,
		Coherence: rand.Float64(),
	}

	b.ResetTimer()
	for i := range b.N {
		weight := float64(i%100) / 100.0
		gm.Blend(local, global, weight)
	}
}

func BenchmarkWeightedGoalManagerWrapping(b *testing.B) {
	gm := &goal.WeightedManager{}

	// Test with phases that require wrapping
	local := core.State{
		Phase:     0.1,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.5,
	}

	global := core.State{
		Phase:     2*math.Pi - 0.1,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.5,
	}

	b.ResetTimer()
	for i := range b.N {
		weight := float64(i%100) / 100.0
		gm.Blend(local, global, weight)
	}
}
