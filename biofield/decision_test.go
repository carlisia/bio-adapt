package biofield

import (
	"math"
	"testing"
	"time"
)

func TestSimpleDecisionMaker(t *testing.T) {
	dm := &SimpleDecisionMaker{}

	state := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.5,
	}

	// Test with no options
	action, confidence := dm.Decide(state, []Action{})
	if action.Type != "maintain" {
		t.Errorf("Expected maintain action for no options, got %s", action.Type)
	}
	if confidence != 0.5 {
		t.Errorf("Expected confidence 0.5 for no options, got %f", confidence)
	}

	// Test with multiple options
	options := []Action{
		{
			Type:    "adjust_phase",
			Value:   0.1,
			Cost:    2.0,
			Benefit: 1.0,
		},
		{
			Type:    "adjust_phase",
			Value:   0.3,
			Cost:    1.0,
			Benefit: 2.0, // Best benefit/cost ratio
		},
		{
			Type:    "maintain",
			Value:   0,
			Cost:    0.1,
			Benefit: 0.05, // Lower benefit to avoid being selected
		},
	}

	action, confidence = dm.Decide(state, options)

	// Should choose action with best benefit/cost ratio (2.0/1.0 = 2.0)
	if action.Value != 0.3 {
		t.Errorf("Expected action with value 0.3 (best ratio), got %f", action.Value)
	}

	// Confidence should be based on score
	expectedConfidence := 1.0 // score (2.0) / 2.0 = 1.0
	if math.Abs(confidence-expectedConfidence) > 0.01 {
		t.Errorf("Expected confidence %f, got %f", expectedConfidence, confidence)
	}
}

func TestSimpleDecisionMakerZeroCost(t *testing.T) {
	dm := &SimpleDecisionMaker{}

	state := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.5,
	}

	// Test with zero cost action (should use minimum cost of 0.1)
	options := []Action{
		{
			Type:    "free_action",
			Value:   1.0,
			Cost:    0, // Zero cost
			Benefit: 1.0,
		},
	}

	action, confidence := dm.Decide(state, options)

	if action.Type != "free_action" {
		t.Errorf("Expected free_action, got %s", action.Type)
	}

	// Score should be 1.0/0.1 = 10.0, confidence = min(10.0/2.0, 1.0) = 1.0
	if confidence != 1.0 {
		t.Errorf("Expected confidence 1.0 for high score, got %f", confidence)
	}
}

func TestSimpleDecisionMakerNegativeBenefit(t *testing.T) {
	dm := &SimpleDecisionMaker{}

	state := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.5,
	}

	// Test with negative benefit actions
	options := []Action{
		{
			Type:    "bad_action",
			Value:   1.0,
			Cost:    1.0,
			Benefit: -1.0, // Negative benefit
		},
		{
			Type:    "worse_action",
			Value:   2.0,
			Cost:    1.0,
			Benefit: -2.0, // More negative
		},
	}

	action, _ := dm.Decide(state, options)

	// Should choose the least bad option
	if action.Type != "bad_action" {
		t.Errorf("Expected bad_action (least negative), got %s", action.Type)
	}
}

func TestSimpleDecisionMakerConsistency(t *testing.T) {
	dm := &SimpleDecisionMaker{}

	state := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.5,
	}

	options := []Action{
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
	}

	// Both have same benefit/cost ratio, should consistently choose one
	action1, _ := dm.Decide(state, options)
	action2, _ := dm.Decide(state, options)

	if action1.Type != action2.Type {
		t.Error("Decision maker should be consistent for same inputs")
	}
}

func BenchmarkSimpleDecisionMaker(b *testing.B) {
	dm := &SimpleDecisionMaker{}

	state := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.5,
	}

	options := make([]Action, 10)
	for i := range options {
		options[i] = Action{
			Type:    "adjust_phase",
			Value:   float64(i) * 0.1,
			Cost:    float64(i+1) * 0.5,
			Benefit: float64(10-i) * 0.3,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dm.Decide(state, options)
	}
}