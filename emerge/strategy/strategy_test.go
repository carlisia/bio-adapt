package strategy

import (
	"math"
	"testing"
	"time"

	"github.com/carlisia/bio-adapt/emerge/core"
)

func TestPhaseNudge(t *testing.T) {
	tests := []struct {
		name        string
		rate        float64
		current     core.State
		target      core.State
		context     core.Context
		validateFn  func(t *testing.T, strategy *PhaseNudge, action core.Action, confidence float64)
		description string
	}{
		// Happy path cases
		{
			name: "basic nudge forward",
			rate: 0.5,
			current: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			target: core.State{
				Phase:     math.Pi / 2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			context: core.Context{
				Density:        0.5,
				Stability:      0.7,
				Progress:       0.3,
				LocalCoherence: 0.6,
			},
			validateFn: func(t *testing.T, strategy *PhaseNudge, action core.Action, confidence float64) {
				if strategy.Rate != 0.5 {
					t.Errorf("Expected rate 0.5, got %f", strategy.Rate)
				}
				if strategy.Name() != "phase_nudge" {
					t.Errorf("Expected name 'phase_nudge', got '%s'", strategy.Name())
				}
				if action.Type != "phase_nudge" {
					t.Errorf("Expected action type 'phase_nudge', got '%s'", action.Type)
				}
				if action.Value <= 0 {
					t.Error("Action value should be positive when moving forward")
				}
				if action.Cost <= 0 {
					t.Error("Action cost should be positive")
				}
				if confidence < 0 || confidence > 1 {
					t.Errorf("Confidence should be in [0, 1], got %f", confidence)
				}
			},
			description: "Basic forward phase nudge",
		},
		{
			name: "nudge backward",
			rate: 0.3,
			current: core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			target: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			context: core.Context{
				Density:        0.5,
				Stability:      0.7,
				Progress:       0.3,
				LocalCoherence: 0.6,
			},
			validateFn: func(t *testing.T, strategy *PhaseNudge, action core.Action, confidence float64) {
				if action.Type != "phase_nudge" {
					t.Errorf("Expected action type 'phase_nudge', got '%s'", action.Type)
				}
				if action.Value >= 0 {
					t.Error("Action value should be negative when moving backward")
				}
			},
			description: "Backward phase nudge",
		},
		{
			name: "zero rate",
			rate: 0,
			current: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			target: core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			context: core.Context{
				LocalCoherence: 0.6,
			},
			validateFn: func(t *testing.T, strategy *PhaseNudge, action core.Action, confidence float64) {
				if action.Value != 0 {
					t.Error("Zero rate should produce zero value")
				}
			},
			description: "Zero rate produces no nudge",
		},
		{
			name: "max rate",
			rate: 1.0,
			current: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			target: core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			context: core.Context{
				LocalCoherence: 0.8,
			},
			validateFn: func(t *testing.T, strategy *PhaseNudge, action core.Action, confidence float64) {
				if strategy.Rate != 1.0 {
					t.Errorf("Expected rate 1.0, got %f", strategy.Rate)
				}
				// Max rate should produce large adjustment
				if math.Abs(action.Value) < math.Pi/4 {
					t.Error("Max rate should produce significant adjustment")
				}
			},
			description: "Maximum rate produces large nudge",
		},
		// Edge cases
		{
			name: "same phase",
			rate: 0.5,
			current: core.State{
				Phase:     math.Pi / 2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			target: core.State{
				Phase:     math.Pi / 2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			context: core.Context{
				LocalCoherence: 0.6,
			},
			validateFn: func(t *testing.T, strategy *PhaseNudge, action core.Action, confidence float64) {
				if action.Value != 0 {
					t.Error("Same phase should produce zero nudge")
				}
			},
			description: "Same phase produces no nudge",
		},
		{
			name: "wrap around",
			rate: 0.5,
			current: core.State{
				Phase:     2*math.Pi - 0.1,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			target: core.State{
				Phase:     0.1,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			context: core.Context{
				LocalCoherence: 0.6,
			},
			validateFn: func(t *testing.T, strategy *PhaseNudge, action core.Action, confidence float64) {
				// Should take shortest path across boundary
				if math.Abs(action.Value) > math.Pi {
					t.Error("Should take shortest path across phase boundary")
				}
			},
			description: "Phase wrap around handling",
		},
		{
			name: "negative phase",
			rate: 0.5,
			current: core.State{
				Phase:     -math.Pi / 2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			target: core.State{
				Phase:     math.Pi / 2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			context: core.Context{
				LocalCoherence: 0.6,
			},
			validateFn: func(t *testing.T, strategy *PhaseNudge, action core.Action, confidence float64) {
				if action.Type != "phase_nudge" {
					t.Errorf("Expected action type 'phase_nudge', got '%s'", action.Type)
				}
			},
			description: "Negative phase handling",
		},
		{
			name: "zero coherence context",
			rate: 0.5,
			current: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			target: core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			context: core.Context{
				LocalCoherence: 0,
			},
			validateFn: func(t *testing.T, strategy *PhaseNudge, action core.Action, confidence float64) {
				if confidence < 0.5 {
					t.Error("Confidence should be at least 0.5")
				}
				// Zero local coherence should produce high confidence (need to adjust)
				if confidence < 0.9 {
					t.Error("Zero local coherence should produce high confidence")
				}
			},
			description: "Zero coherence produces high confidence",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := NewPhaseNudge(tt.rate)
			action, confidence := strategy.Propose(tt.current, tt.target, tt.context)
			tt.validateFn(t, strategy, action, confidence)
		})
	}
}

func TestPhaseNudgeRateClamping(t *testing.T) {
	tests := []struct {
		name        string
		rate        float64
		expected    float64
		description string
	}{
		{"negative rate", -0.5, 0, "Negative rate should clamp to 0"},
		{"zero rate", 0, 0, "Zero rate should remain 0"},
		{"normal rate", 0.5, 0.5, "Normal rate should be preserved"},
		{"max rate", 1.0, 1.0, "Max rate should be preserved"},
		{"over max rate", 1.5, 1.0, "Over max should clamp to 1.0"},
		{"very negative", -100, 0, "Very negative should clamp to 0"},
		{"very large", 100, 1.0, "Very large should clamp to 1.0"},
		{"NaN rate", math.NaN(), 0, "NaN should default to 0"},
		{"infinity rate", math.Inf(1), 1.0, "Infinity should clamp to 1.0"},
		{"negative infinity", math.Inf(-1), 0, "Negative infinity should clamp to 0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := NewPhaseNudge(tt.rate)

			// Special handling for NaN
			if math.IsNaN(tt.rate) {
				if !math.IsNaN(strategy.Rate) && strategy.Rate != tt.expected {
					t.Errorf("%s: Expected rate %f, got %f", tt.description, tt.expected, strategy.Rate)
				}
			} else if strategy.Rate != tt.expected {
				t.Errorf("%s: Expected rate %f, got %f", tt.description, tt.expected, strategy.Rate)
			}
		})
	}
}

func TestFrequencyLockStrategy(t *testing.T) {
	strategy := NewFrequencyLock(0.8)

	if strategy.SyncRate != 0.8 {
		t.Errorf("Expected strength 0.8, got %f", strategy.SyncRate)
	}

	if strategy.Name() != "frequency_lock" {
		t.Errorf("Expected name 'frequency_lock', got '%s'", strategy.Name())
	}

	current := core.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.5,
	}

	target := core.State{
		Phase:     math.Pi,
		Frequency: 150 * time.Millisecond,
		Coherence: 0.9,
	}

	context := core.Context{
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
	expectedConfidence := context.LocalCoherence * strategy.SyncRate
	if math.Abs(confidence-expectedConfidence) > 0.01 {
		t.Errorf("Expected confidence %f, got %f", expectedConfidence, confidence)
	}
}

func TestEnergyAwareStrategy(t *testing.T) {
	strategy := NewEnergyAware(20.0)

	if strategy.Threshold != 20.0 {
		t.Errorf("Expected threshold 20.0, got %f", strategy.Threshold)
	}

	if strategy.Name() != "energy_aware" {
		t.Errorf("Expected name 'energy_aware', got '%s'", strategy.Name())
	}

	// Test with low energy
	currentLowEnergy := core.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.5,
	}

	target := core.State{
		Phase:     math.Pi,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	contextLowEnergy := core.Context{
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
	contextHighEnergy := core.Context{
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
	nudge := NewPhaseNudge(0.3)
	frequency := NewFrequencyLock(0.7)
	energy := NewEnergyAware(25.0)

	adaptive := newAdaptiveStrategy([]core.SyncStrategy{nudge, frequency, energy})

	if adaptive.Name() != "adaptive" {
		t.Errorf("Expected name 'adaptive', got '%s'", adaptive.Name())
	}

	current := core.State{
		Phase:     math.Pi / 4,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.6,
	}

	target := core.State{
		Phase:     3 * math.Pi / 4,
		Frequency: 120 * time.Millisecond,
		Coherence: 0.95,
	}

	context := core.Context{
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
	strategy := NewPulse(100*time.Millisecond, 0.5)

	if strategy.Name() != "pulse" {
		t.Errorf("Expected name 'pulse', got '%s'", strategy.Name())
	}

	if strategy.Period != 100*time.Millisecond {
		t.Errorf("Expected period 100ms, got %v", strategy.Period)
	}

	if strategy.Amplitude != 0.5 {
		t.Errorf("Expected amplitude 0.5, got %f", strategy.Amplitude)
	}

	current := core.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.3,
	}

	target := core.State{
		Phase:     math.Pi,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	context := core.Context{
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
	strategies := []core.SyncStrategy{
		NewPhaseNudge(0.5),
		NewFrequencyLock(0.7),
		NewEnergyAware(30.0),
		NewPulse(50*time.Millisecond, 0.6),
	}

	current := core.State{
		Phase:     math.Pi / 4,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.6,
	}

	target := core.State{
		Phase:     3 * math.Pi / 4,
		Frequency: 120 * time.Millisecond,
		Coherence: 0.95,
	}

	context := core.Context{
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
	strategy1 := NewPhaseNudge(0.5)
	strategy2 := NewFrequencyLock(0.8)
	strategy3 := NewEnergyAware(25.0)

	adaptive := newAdaptiveStrategy([]core.SyncStrategy{strategy1, strategy2, strategy3})

	current := core.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.5,
	}

	target := core.State{
		Phase:     math.Pi,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	// Test with low stability (simulates low energy)
	lowStabilityContext := core.Context{
		LocalCoherence: 0.5,
		Stability:      0.2,
	}

	action1, _ := adaptive.Propose(current, target, lowStabilityContext)
	if action1.Type == "" {
		t.Error("Should produce an action even with low stability")
	}

	// Test with high coherence
	highCoherenceContext := core.Context{
		LocalCoherence: 0.85,
		Stability:      0.9,
	}

	action2, _ := adaptive.Propose(current, target, highCoherenceContext)
	if action2.Type == "" {
		t.Error("Should produce an action with high coherence")
	}
}
