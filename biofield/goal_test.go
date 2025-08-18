package biofield

import (
	"math"
	"math/rand"
	"testing"
	"time"
)

func TestWeightedGoalManager(t *testing.T) {
	gm := &WeightedGoalManager{}

	local := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.3,
	}

	global := State{
		Phase:     math.Pi,
		Frequency: 200 * time.Millisecond,
		Coherence: 0.9,
	}

	// Test pure local (weight = 0)
	blended := gm.Blend(local, global, 0)
	if math.Abs(blended.Phase-local.Phase) > 0.01 {
		t.Errorf("Weight 0 should give local phase, got %f", blended.Phase)
	}
	if math.Abs(blended.Coherence-local.Coherence) > 0.01 {
		t.Errorf("Weight 0 should give local coherence, got %f", blended.Coherence)
	}

	// Test pure global (weight = 1)
	blended = gm.Blend(local, global, 1)
	if math.Abs(blended.Phase-global.Phase) > 0.01 {
		t.Errorf("Weight 1 should give global phase, got %f", blended.Phase)
	}
	if math.Abs(blended.Coherence-global.Coherence) > 0.01 {
		t.Errorf("Weight 1 should give global coherence, got %f", blended.Coherence)
	}

	// Test 50/50 blend (weight = 0.5)
	blended = gm.Blend(local, global, 0.5)
	expectedPhase := math.Pi / 2 // Halfway between 0 and π
	if math.Abs(blended.Phase-expectedPhase) > 0.01 {
		t.Errorf("Weight 0.5 should blend phases equally, expected %f, got %f",
			expectedPhase, blended.Phase)
	}
	expectedCoherence := 0.6 // (0.3 + 0.9) / 2
	if math.Abs(blended.Coherence-expectedCoherence) > 0.01 {
		t.Errorf("Weight 0.5 should blend coherence equally, expected %f, got %f",
			expectedCoherence, blended.Coherence)
	}

	// Test frequency preservation
	if blended.Frequency != local.Frequency {
		t.Error("Frequency should be preserved from local state")
	}
}

func TestWeightedGoalManagerPhaseWrapping(t *testing.T) {
	gm := &WeightedGoalManager{}

	// Test phase wrapping across 0/2π boundary
	local := State{
		Phase:     0.1,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.5,
	}

	global := State{
		Phase:     2*math.Pi - 0.1,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.5,
	}

	// Should take shortest path across boundary
	blended := gm.Blend(local, global, 0.5)

	// The shortest path from 0.1 to 2π-0.1 is backwards across 0
	// So halfway should be around 0 (or 2π)
	if blended.Phase > 0.2 && blended.Phase < 2*math.Pi-0.2 {
		t.Errorf("Phase wrapping not working correctly, got %f", blended.Phase)
	}
}

func TestWeightedGoalManagerWeightClamping(t *testing.T) {
	gm := &WeightedGoalManager{}

	local := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.5,
	}

	global := State{
		Phase:     math.Pi,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.5,
	}

	// Test weight > 1 (should clamp to 1)
	blended := gm.Blend(local, global, 1.5)
	if math.Abs(blended.Phase-global.Phase) > 0.01 {
		t.Error("Weight > 1 should be clamped to 1")
	}

	// Test weight < 0 (should clamp to 0)
	blended = gm.Blend(local, global, -0.5)
	if math.Abs(blended.Phase-local.Phase) > 0.01 {
		t.Error("Weight < 0 should be clamped to 0")
	}
}

func TestWeightedGoalManagerSymmetry(t *testing.T) {
	gm := &WeightedGoalManager{}

	// Test that blending is symmetric
	state1 := State{
		Phase:     math.Pi / 4,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.4,
	}

	state2 := State{
		Phase:     3 * math.Pi / 4,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.6,
	}

	// Blend 1->2 with weight 0.3
	blend12 := gm.Blend(state1, state2, 0.3)

	// Blend 2->1 with weight 0.7 (opposite)
	blend21 := gm.Blend(state2, state1, 0.7)

	// Results should be the same
	if math.Abs(blend12.Phase-blend21.Phase) > 0.01 {
		t.Errorf("Blending not symmetric for phase: %f vs %f", blend12.Phase, blend21.Phase)
	}
	if math.Abs(blend12.Coherence-blend21.Coherence) > 0.01 {
		t.Errorf("Blending not symmetric for coherence: %f vs %f",
			blend12.Coherence, blend21.Coherence)
	}
}

func BenchmarkWeightedGoalManager(b *testing.B) {
	gm := &WeightedGoalManager{}

	local := State{
		Phase:     rand.Float64() * 2 * math.Pi,
		Frequency: 100 * time.Millisecond,
		Coherence: rand.Float64(),
	}

	global := State{
		Phase:     rand.Float64() * 2 * math.Pi,
		Frequency: 100 * time.Millisecond,
		Coherence: rand.Float64(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		weight := float64(i%100) / 100.0
		gm.Blend(local, global, weight)
	}
}