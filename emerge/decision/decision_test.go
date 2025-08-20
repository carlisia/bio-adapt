package decision

import (
	"math"
	"testing"
	"time"

	"github.com/carlisia/bio-adapt/emerge/core"
)

func TestSimpleDecisionMaker(t *testing.T) {
	tests := []struct {
		name               string
		state              core.State
		options            []core.Action
		expectedType       string
		expectedValue      float64
		expectedConfidence float64
		confTolerance      float64
		description        string
	}{
		// Happy path cases
		{
			name: "no options returns maintain",
			state: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			options:            []core.Action{},
			expectedType:       "maintain",
			expectedValue:      0,
			expectedConfidence: 0.5,
			confTolerance:      0.01,
			description:        "No options should return maintain with coherence confidence",
		},
		{
			name: "single option selected",
			state: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			options: []core.Action{
				{
					Type:    "adjust_phase",
					Value:   0.1,
					Cost:    1.0,
					Benefit: 2.0,
				},
			},
			expectedType:       "adjust_phase",
			expectedValue:      0.1,
			expectedConfidence: 1.0, // score=2.0, conf=min(2.0/2.0, 1.0)=1.0
			confTolerance:      0.01,
			description:        "Single option should be selected",
		},
		{
			name: "best benefit/cost ratio wins",
			state: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			options: []core.Action{
				{
					Type:    "adjust_phase",
					Value:   0.1,
					Cost:    2.0,
					Benefit: 1.0, // ratio = 0.5
				},
				{
					Type:    "adjust_phase",
					Value:   0.3,
					Cost:    1.0,
					Benefit: 2.0, // ratio = 2.0 (best)
				},
				{
					Type:    "maintain",
					Value:   0,
					Cost:    0.1,
					Benefit: 0.05, // ratio = 0.5
				},
			},
			expectedType:       "adjust_phase",
			expectedValue:      0.3,
			expectedConfidence: 1.0,
			confTolerance:      0.01,
			description:        "Action with best benefit/cost ratio should be selected",
		},
		{
			name: "zero cost uses minimum",
			state: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			options: []core.Action{
				{
					Type:    "free_action",
					Value:   1.0,
					Cost:    0, // Will use 0.1 minimum
					Benefit: 1.0,
				},
			},
			expectedType:       "free_action",
			expectedValue:      1.0,
			expectedConfidence: 1.0, // score=10.0, conf=min(10.0/2.0, 1.0)=1.0
			confTolerance:      0.01,
			description:        "Zero cost should use minimum cost of 0.1",
		},
		{
			name: "negative cost uses minimum",
			state: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			options: []core.Action{
				{
					Type:    "negative_cost",
					Value:   1.0,
					Cost:    -5.0, // Will use 0.1 minimum
					Benefit: 1.0,
				},
			},
			expectedType:       "negative_cost",
			expectedValue:      1.0,
			expectedConfidence: 1.0,
			confTolerance:      0.01,
			description:        "Negative cost should use minimum cost of 0.1",
		},
		// Non-happy path cases
		{
			name: "negative benefit least bad wins",
			state: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			options: []core.Action{
				{
					Type:    "bad_action",
					Value:   1.0,
					Cost:    1.0,
					Benefit: -1.0, // ratio = -1.0 (least bad)
				},
				{
					Type:    "worse_action",
					Value:   2.0,
					Cost:    1.0,
					Benefit: -2.0, // ratio = -2.0
				},
			},
			expectedType:       "bad_action",
			expectedValue:      1.0,
			expectedConfidence: 0, // negative score clamps to 0
			confTolerance:      0.01,
			description:        "Least negative benefit/cost ratio should win",
		},
		{
			name: "all zero benefit first selected",
			state: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			options: []core.Action{
				{
					Type:    "action1",
					Value:   1.0,
					Cost:    1.0,
					Benefit: 0,
				},
				{
					Type:    "action2",
					Value:   2.0,
					Cost:    1.0,
					Benefit: 0,
				},
			},
			expectedType:       "action1",
			expectedValue:      1.0,
			expectedConfidence: 0,
			confTolerance:      0.01,
			description:        "With all zero benefits, first action selected",
		},
		{
			name: "equal ratios first selected",
			state: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			options: []core.Action{
				{
					Type:    "option1",
					Value:   1.0,
					Cost:    1.0,
					Benefit: 2.0, // ratio = 2.0
				},
				{
					Type:    "option2",
					Value:   2.0,
					Cost:    2.0,
					Benefit: 4.0, // ratio = 2.0
				},
			},
			expectedType:       "option1",
			expectedValue:      1.0,
			expectedConfidence: 1.0,
			confTolerance:      0.01,
			description:        "With equal ratios, first option selected",
		},
		// Edge cases
		{
			name: "very small benefit and cost",
			state: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			options: []core.Action{
				{
					Type:    "tiny_action",
					Value:   0.001,
					Cost:    0.00001,
					Benefit: 0.00002,
				},
			},
			expectedType:       "tiny_action",
			expectedValue:      0.001,
			expectedConfidence: 0.001, // score=0.002, conf=min(0.002/2.0, 1.0)=0.001
			confTolerance:      0.01,
			description:        "Very small values should work correctly",
		},
		{
			name: "very large benefit and cost",
			state: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			options: []core.Action{
				{
					Type:    "huge_action",
					Value:   1000,
					Cost:    10000,
					Benefit: 50000, // ratio = 5.0
				},
			},
			expectedType:       "huge_action",
			expectedValue:      1000,
			expectedConfidence: 1.0, // score=5.0, conf=min(5.0/2.0, 1.0)=1.0
			confTolerance:      0.01,
			description:        "Very large values should work correctly",
		},
		{
			name: "mixed positive and negative benefits",
			state: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			options: []core.Action{
				{
					Type:    "negative",
					Value:   1.0,
					Cost:    1.0,
					Benefit: -1.0,
				},
				{
					Type:    "positive",
					Value:   2.0,
					Cost:    2.0,
					Benefit: 1.0, // ratio = 0.5 (best)
				},
				{
					Type:    "zero",
					Value:   3.0,
					Cost:    1.0,
					Benefit: 0,
				},
			},
			expectedType:       "positive",
			expectedValue:      2.0,
			expectedConfidence: 0.25, // score=0.5, conf=min(0.5/2.0, 1.0)=0.25
			confTolerance:      0.01,
			description:        "Positive benefit should win over negative/zero",
		},
		{
			name: "infinity benefit",
			state: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			options: []core.Action{
				{
					Type:    "infinite",
					Value:   1.0,
					Cost:    1.0,
					Benefit: math.Inf(1),
				},
			},
			expectedType:       "infinite",
			expectedValue:      1.0,
			expectedConfidence: 1.0, // Infinity score clamps to 1.0
			confTolerance:      0.01,
			description:        "Infinity benefit should work",
		},
		{
			name: "NaN benefit",
			state: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			options: []core.Action{
				{
					Type:    "nan_action",
					Value:   1.0,
					Cost:    1.0,
					Benefit: math.NaN(),
				},
				{
					Type:    "valid_action",
					Value:   2.0,
					Cost:    1.0,
					Benefit: 1.0,
				},
			},
			expectedType:       "valid_action", // NaN should be handled, valid action selected
			expectedValue:      2.0,
			expectedConfidence: 0.5,
			confTolerance:      0.01,
			description:        "NaN benefit should be handled",
		},
		{
			name: "zero coherence state",
			state: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0,
			},
			options:            []core.Action{},
			expectedType:       "maintain",
			expectedValue:      0,
			expectedConfidence: 0,
			confTolerance:      0.01,
			description:        "Zero coherence should work for no options",
		},
		{
			name: "negative coherence state",
			state: core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: -0.5, // Invalid but should handle
			},
			options:            []core.Action{},
			expectedType:       "maintain",
			expectedValue:      0,
			expectedConfidence: -0.5,
			confTolerance:      0.01,
			description:        "Negative coherence should be passed through",
		},
		{
			name: "coherence > 1",
			state: core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 1.5, // Invalid but should handle
			},
			options:            []core.Action{},
			expectedType:       "maintain",
			expectedValue:      0,
			expectedConfidence: 1.5,
			confTolerance:      0.01,
			description:        "Coherence > 1 should be passed through",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dm := &SimpleDecisionMaker{}
			action, confidence := dm.Decide(tt.state, tt.options)

			if action.Type != tt.expectedType {
				t.Errorf("%s: Type = %s, expected %s",
					tt.description, action.Type, tt.expectedType)
			}

			if math.Abs(action.Value-tt.expectedValue) > 0.001 {
				t.Errorf("%s: Value = %f, expected %f",
					tt.description, action.Value, tt.expectedValue)
			}

			// Skip confidence check for NaN cases
			if math.IsNaN(tt.expectedConfidence) && math.IsNaN(confidence) {
				return
			}

			if math.Abs(confidence-tt.expectedConfidence) > tt.confTolerance {
				t.Errorf("%s: Confidence = %f, expected %f±%f",
					tt.description, confidence, tt.expectedConfidence, tt.confTolerance)
			}
		})
	}
}

func TestSimpleDecisionMakerConsistency(t *testing.T) {
	tests := []struct {
		name        string
		state       core.State
		options     []core.Action
		iterations  int
		description string
	}{
		{
			name: "consistent with equal ratios",
			state: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			options: []core.Action{
				{
					Type:    "option1",
					Value:   1.0,
					Cost:    1.0,
					Benefit: 2.0,
				},
				{
					Type:    "option2",
					Value:   2.0,
					Cost:    2.0,
					Benefit: 4.0,
				},
			},
			iterations:  10,
			description: "Should consistently choose same option with equal ratios",
		},
		{
			name: "consistent with clear winner",
			state: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			options: []core.Action{
				{
					Type:    "bad",
					Value:   1.0,
					Cost:    10.0,
					Benefit: 1.0,
				},
				{
					Type:    "good",
					Value:   2.0,
					Cost:    1.0,
					Benefit: 10.0,
				},
			},
			iterations:  10,
			description: "Should consistently choose clear winner",
		},
		{
			name: "consistent with no options",
			state: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			options:     []core.Action{},
			iterations:  10,
			description: "Should consistently return maintain for no options",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dm := &SimpleDecisionMaker{}

			var firstAction core.Action
			var firstConfidence float64

			for i := 0; i < tt.iterations; i++ {
				action, confidence := dm.Decide(tt.state, tt.options)

				if i == 0 {
					firstAction = action
					firstConfidence = confidence
				} else {
					if action.Type != firstAction.Type {
						t.Errorf("%s: Iteration %d returned different type: %s vs %s",
							tt.description, i, action.Type, firstAction.Type)
					}
					if math.Abs(action.Value-firstAction.Value) > 0.001 {
						t.Errorf("%s: Iteration %d returned different value: %f vs %f",
							tt.description, i, action.Value, firstAction.Value)
					}
					if math.Abs(confidence-firstConfidence) > 0.001 {
						t.Errorf("%s: Iteration %d returned different confidence: %f vs %f",
							tt.description, i, confidence, firstConfidence)
					}
				}
			}
		})
	}
}

func TestSimpleDecisionMakerBoundaryConditions(t *testing.T) {
	tests := []struct {
		name        string
		state       core.State
		options     []core.Action
		validateFn  func(t *testing.T, action core.Action, confidence float64)
		description string
	}{
		{
			name: "many options performance",
			state: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			options: func() []core.Action {
				opts := make([]core.Action, 1000)
				for i := range opts {
					opts[i] = core.Action{
						Type:    "action",
						Value:   float64(i),
						Cost:    float64(i+1) * 0.5,
						Benefit: float64(1000-i) * 0.3,
					}
				}
				// Make last one the best
				opts[999] = core.Action{
					Type:    "best",
					Value:   999,
					Cost:    0.1,
					Benefit: 100,
				}
				return opts
			}(),
			validateFn: func(t *testing.T, action core.Action, confidence float64) {
				if action.Type != "best" {
					t.Errorf("Should select best option even among many")
				}
			},
			description: "Should handle many options efficiently",
		},
		{
			name: "all identical options",
			state: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			options: []core.Action{
				{Type: "same1", Value: 1.0, Cost: 1.0, Benefit: 1.0},
				{Type: "same2", Value: 1.0, Cost: 1.0, Benefit: 1.0},
				{Type: "same3", Value: 1.0, Cost: 1.0, Benefit: 1.0},
			},
			validateFn: func(t *testing.T, action core.Action, confidence float64) {
				if action.Type != "same1" {
					t.Errorf("Should select first of identical options")
				}
				if math.Abs(confidence-0.5) > 0.01 {
					t.Errorf("Confidence should be 0.5 for ratio 1.0")
				}
			},
			description: "Should handle identical options",
		},
		{
			name: "zero state values",
			state: core.State{
				Phase:     0,
				Frequency: 0,
				Coherence: 0,
			},
			options: []core.Action{
				{Type: "action", Value: 1.0, Cost: 1.0, Benefit: 1.0},
			},
			validateFn: func(t *testing.T, action core.Action, confidence float64) {
				if action.Type != "action" {
					t.Errorf("Should work with zero state values")
				}
			},
			description: "Should work with zero state values",
		},
		{
			name: "negative phase state",
			state: core.State{
				Phase:     -math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			options: []core.Action{
				{Type: "action", Value: 1.0, Cost: 1.0, Benefit: 1.0},
			},
			validateFn: func(t *testing.T, action core.Action, confidence float64) {
				if action.Type != "action" {
					t.Errorf("Should work with negative phase")
				}
			},
			description: "Should work with negative phase",
		},
		{
			name: "negative frequency state",
			state: core.State{
				Phase:     0,
				Frequency: -100 * time.Millisecond,
				Coherence: 0.5,
			},
			options: []core.Action{
				{Type: "action", Value: 1.0, Cost: 1.0, Benefit: 1.0},
			},
			validateFn: func(t *testing.T, action core.Action, confidence float64) {
				if action.Type != "action" {
					t.Errorf("Should work with negative frequency")
				}
			},
			description: "Should work with negative frequency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dm := &SimpleDecisionMaker{}
			action, confidence := dm.Decide(tt.state, tt.options)
			tt.validateFn(t, action, confidence)
		})
	}
}

func TestSimpleDecisionMakerConcurrency(t *testing.T) {
	dm := &SimpleDecisionMaker{}

	state := core.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.5,
	}

	options := []core.Action{
		{Type: "action1", Value: 1.0, Cost: 1.0, Benefit: 2.0},
		{Type: "action2", Value: 2.0, Cost: 2.0, Benefit: 3.0},
		{Type: "action3", Value: 3.0, Cost: 1.5, Benefit: 4.0},
	}

	// Run concurrent decisions
	done := make(chan bool, 100)
	for range 100 {
		go func() {
			_, _ = dm.Decide(state, options)
			done <- true
		}()
	}

	// Wait for all goroutines
	for range 100 {
		<-done
	}

	// If we get here without race conditions, concurrent access is safe
}

func BenchmarkSimpleDecisionMaker(b *testing.B) {
	benchmarks := []struct {
		name    string
		options int
	}{
		{"1_option", 1},
		{"10_options", 10},
		{"100_options", 100},
		{"1000_options", 1000},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			dm := &SimpleDecisionMaker{}

			state := core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			}

			options := make([]core.Action, bm.options)
			for i := range options {
				options[i] = core.Action{
					Type:    "adjust_phase",
					Value:   float64(i) * 0.1,
					Cost:    float64(i+1) * 0.5,
					Benefit: float64(bm.options-i) * 0.3,
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dm.Decide(state, options)
			}
		})
	}
}

func BenchmarkSimpleDecisionMakerZeroCost(b *testing.B) {
	dm := &SimpleDecisionMaker{}

	state := core.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.5,
	}

	// All zero cost options
	options := make([]core.Action, 10)
	for i := range options {
		options[i] = core.Action{
			Type:    "free_action",
			Value:   float64(i),
			Cost:    0,
			Benefit: float64(i) * 0.5,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dm.Decide(state, options)
	}
}
