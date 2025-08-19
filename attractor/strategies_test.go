package attractor

import (
	"math"
	"testing"
	"time"
)

func TestPhaseNudgeStrategy(t *testing.T) {
	strategy := NewPhaseNudgeStrategy(0.5)

	if strategy.Rate != 0.5 {
		t.Errorf("Expected rate 0.5, got %f", strategy.Rate)
	}

	if strategy.Name() != "phase_nudge" {
		t.Errorf("Expected name 'phase_nudge', got '%s'", strategy.Name())
	}

	current := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.5,
	}

	target := State{
		Phase:     math.Pi/2,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	context := Context{
		Density:        0.5,
		Stability:      0.7,
		Progress:       0.3,
		LocalCoherence: 0.6,
	}

	action, confidence := strategy.Propose(current, target, context)

	if action.Type != "phase_nudge" {
		t.Errorf("Expected action type 'phase_nudge', got '%s'", action.Type)
	}

	// Value should be positive (moving toward pi/2 from 0)
	if action.Value <= 0 {
		t.Error("Action value should be positive when moving forward")
	}

	// Cost should be proportional to adjustment
	if action.Cost <= 0 {
		t.Error("Action cost should be positive")
	}

	if confidence < 0 || confidence > 1 {
		t.Errorf("Confidence should be in [0, 1], got %f", confidence)
	}
}

func TestPhaseNudgeStrategyRateClamping(t *testing.T) {
	tests := []struct {
		name     string
		rate     float64
		expected float64
	}{
		{"negative rate", -0.5, 0},
		{"zero rate", 0, 0},
		{"normal rate", 0.5, 0.5},
		{"max rate", 1.0, 1.0},
		{"over max rate", 1.5, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := NewPhaseNudgeStrategy(tt.rate)
			if strategy.Rate != tt.expected {
				t.Errorf("Expected rate %f, got %f", tt.expected, strategy.Rate)
			}
		})
	}
}

func TestFrequencyLockStrategy(t *testing.T) {
	strategy := NewFrequencyLockStrategy(0.8)

	if strategy.Strength != 0.8 {
		t.Errorf("Expected strength 0.8, got %f", strategy.Strength)
	}

	if strategy.Name() != "frequency_lock" {
		t.Errorf("Expected name 'frequency_lock', got '%s'", strategy.Name())
	}

	current := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.5,
	}

	target := State{
		Phase:     math.Pi,
		Frequency: 150 * time.Millisecond,
		Coherence: 0.9,
	}

	context := Context{
		Density:        0.5,
		Stability:      0.9,
		Progress:       0.7,
		LocalCoherence: 0.7,
	}

	action, confidence := strategy.Propose(current, target, context)

	if action.Type != "frequency_lock" {
		t.Errorf("Expected action type 'frequency_lock', got '%s'", action.Type)
	}

	// Value should represent phase adjustment
	if action.Value == 0 {
		t.Error("Action value should not be zero when phases differ")
	}

	// Confidence should be based on local coherence and strength
	expectedConfidence := context.LocalCoherence * strategy.Strength
	if math.Abs(confidence-expectedConfidence) > 0.01 {
		t.Errorf("Expected confidence %f, got %f", expectedConfidence, confidence)
	}
}

func TestEnergyAwareStrategy(t *testing.T) {
	strategy := NewEnergyAwareStrategy(20.0)

	if strategy.Threshold != 20.0 {
		t.Errorf("Expected threshold 20.0, got %f", strategy.Threshold)
	}

	if strategy.Name() != "energy_aware" {
		t.Errorf("Expected name 'energy_aware', got '%s'", strategy.Name())
	}

	// Test with low energy
	currentLowEnergy := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.5,
	}

	target := State{
		Phase:     math.Pi,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	contextLowEnergy := Context{
		Density:        0.5,
		Stability:      0.2, // Low stability simulates low energy
		Progress:       0.3,
		LocalCoherence: 0.6,
	}

	action, confidence := strategy.Propose(currentLowEnergy, target, contextLowEnergy)

	if action.Type != "energy_save" {
		t.Errorf("Expected 'energy_save' action with low energy, got '%s'", action.Type)
	}

	if confidence != 0.3 {
		t.Errorf("Expected confidence 0.3 for energy_save, got %f", confidence)
	}

	// Test with sufficient energy (high stability)
	contextHighEnergy := Context{
		Density:        0.5,
		Stability:      0.9, // High stability simulates high energy
		Progress:       0.3,
		LocalCoherence: 0.6,
	}

	action, confidence = strategy.Propose(currentLowEnergy, target, contextHighEnergy)

	// With high stability, should still do energy_save but with different parameters
	if action.Type != "energy_save" {
		t.Errorf("Expected 'energy_save' action, got '%s'", action.Type)
	}
	
	// Confidence should still be 0.3 for energy_save
	if confidence != 0.3 {
		t.Errorf("Expected confidence 0.3, got %f", confidence)
	}
}

func TestAdaptiveStrategy(t *testing.T) {
	nudge := NewPhaseNudgeStrategy(0.3)
	frequency := NewFrequencyLockStrategy(0.7)
	energy := NewEnergyAwareStrategy(25.0)

	adaptive := NewAdaptiveStrategy([]SyncStrategy{nudge, frequency, energy})

	if adaptive.Name() != "adaptive" {
		t.Errorf("Expected name 'adaptive', got '%s'", adaptive.Name())
	}

	current := State{
		Phase:     math.Pi/4,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.6,
	}

	target := State{
		Phase:     3*math.Pi/4,
		Frequency: 120 * time.Millisecond,
		Coherence: 0.95,
	}

	context := Context{
		Density:        0.6,
		Stability:      0.7,
		Progress:       0.4,
		LocalCoherence: 0.65,
	}

	action, confidence := adaptive.Propose(current, target, context)

	if action.Type == "" {
		t.Error("Adaptive strategy should produce an action")
	}

	if confidence < 0 || confidence > 1 {
		t.Errorf("Confidence should be in [0, 1], got %f", confidence)
	}
}

func TestPulseStrategy(t *testing.T) {
	strategy := NewPulseStrategy(100*time.Millisecond, 0.5)

	if strategy.Name() != "pulse" {
		t.Errorf("Expected name 'pulse', got '%s'", strategy.Name())
	}

	if strategy.Period != 100*time.Millisecond {
		t.Errorf("Expected period 100ms, got %v", strategy.Period)
	}

	if strategy.Amplitude != 0.5 {
		t.Errorf("Expected amplitude 0.5, got %f", strategy.Amplitude)
	}

	current := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.3,
	}

	target := State{
		Phase:     math.Pi,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	context := Context{
		LocalCoherence: 0.5,
		Progress:       0.3,
	}

	action, confidence := strategy.Propose(current, target, context)

	// Pulse strategy may return 'maintain' when not pulsing
	if action.Type != "pulse" && action.Type != "maintain" {
		t.Errorf("Expected action type 'pulse' or 'maintain', got '%s'", action.Type)
	}

	if confidence < 0 || confidence > 1 {
		t.Errorf("Confidence should be in [0, 1], got %f", confidence)
	}
}

func TestStrategyIntegration(t *testing.T) {
	// Test that all strategies work together
	strategies := []SyncStrategy{
		NewPhaseNudgeStrategy(0.5),
		NewFrequencyLockStrategy(0.7),
		NewEnergyAwareStrategy(30.0),
		NewPulseStrategy(50*time.Millisecond, 0.6),
	}

	current := State{
		Phase:     math.Pi/4,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.6,
	}

	target := State{
		Phase:     3*math.Pi/4,
		Frequency: 120 * time.Millisecond,
		Coherence: 0.95,
	}

	context := Context{
		Density:        0.6,
		Stability:      0.7,
		Progress:       0.4,
		LocalCoherence: 0.65,
	}

	for i, strategy := range strategies {
		action, confidence := strategy.Propose(current, target, context)
		
		if action.Type == "" {
			t.Errorf("Strategy %d (%s) produced empty action type", i, strategy.Name())
		}
		
		if confidence < 0 || confidence > 1 {
			t.Errorf("Strategy %d (%s) produced invalid confidence %f", i, strategy.Name(), confidence)
		}
		
		if action.Cost < 0 {
			t.Errorf("Strategy %d (%s) produced negative cost %f", i, strategy.Name(), action.Cost)
		}
	}
}

func TestStrategySelection(t *testing.T) {
	// Test adaptive strategy selection
	strategy1 := NewPhaseNudgeStrategy(0.5)
	strategy2 := NewFrequencyLockStrategy(0.8)
	strategy3 := NewEnergyAwareStrategy(25.0)

	adaptive := NewAdaptiveStrategy([]SyncStrategy{strategy1, strategy2, strategy3})

	current := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.5,
	}

	target := State{
		Phase:     math.Pi,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	// Test with low stability (simulates low energy)
	lowStabilityContext := Context{
		LocalCoherence: 0.5,
		Stability:      0.2,
	}

	action1, _ := adaptive.Propose(current, target, lowStabilityContext)
	if action1.Type == "" {
		t.Error("Should produce an action even with low stability")
	}

	// Test with high coherence
	highCoherenceContext := Context{
		LocalCoherence: 0.85,
		Stability:      0.9,
	}

	action2, _ := adaptive.Propose(current, target, highCoherenceContext)
	if action2.Type == "" {
		t.Error("Should produce an action with high coherence")
	}
}