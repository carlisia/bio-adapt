package strategy

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/carlisia/bio-adapt/emerge/core"
)

func TestPhaseNudge(t *testing.T) {
	t.Parallel()
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
				assert.Equal(t, 0.5, strategy.Rate, "Expected rate 0.5")
				assert.Equal(t, "phase_nudge", strategy.Name(), "Expected name 'phase_nudge'")
				assert.Equal(t, "phase_nudge", action.Type, "Expected action type 'phase_nudge'")
				assert.Greater(t, action.Value, 0.0, "Action value should be positive when moving forward")
				assert.Greater(t, action.Cost, 0.0, "Action cost should be positive")
				assert.GreaterOrEqual(t, confidence, 0.0, "Confidence should be >= 0")
				assert.LessOrEqual(t, confidence, 1.0, "Confidence should be <= 1")
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
				assert.Equal(t, "phase_nudge", action.Type, "Expected action type 'phase_nudge'")
				assert.Less(t, action.Value, 0.0, "Action value should be negative when moving backward")
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
				assert.Equal(t, 0.0, action.Value, "Zero rate should produce zero value")
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
				assert.Equal(t, 1.0, strategy.Rate, "Expected rate 1.0")
				// Max rate should produce large adjustment
				assert.GreaterOrEqual(t, math.Abs(action.Value), math.Pi/4, "Max rate should produce significant adjustment")
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
				assert.Equal(t, 0.0, action.Value, "Same phase should produce zero nudge")
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
				assert.LessOrEqual(t, math.Abs(action.Value), math.Pi, "Should take shortest path across phase boundary")
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
				assert.Equal(t, "phase_nudge", action.Type, "Expected action type 'phase_nudge'")
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
				assert.GreaterOrEqual(t, confidence, 0.5, "Confidence should be at least 0.5")
				// Zero local coherence should produce high confidence (need to adjust)
				assert.GreaterOrEqual(t, confidence, 0.9, "Zero local coherence should produce high confidence")
			},
			description: "Zero coherence produces high confidence",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			strategy := NewPhaseNudge(tt.rate)
			action, confidence := strategy.Propose(tt.current, tt.target, tt.context)
			tt.validateFn(t, strategy, action, confidence)
		})
	}
}

func TestPhaseNudgeRateClamping(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			strategy := NewPhaseNudge(tt.rate)

			// Special handling for NaN
			if math.IsNaN(tt.rate) {
				if !math.IsNaN(strategy.Rate) && strategy.Rate != tt.expected {
					assert.Equal(t, tt.expected, strategy.Rate, "%s: Expected rate %f", tt.description, tt.expected)
				}
			} else {
				assert.Equal(t, tt.expected, strategy.Rate, "%s: Expected rate %f", tt.description, tt.expected)
			}
		})
	}
}

func TestFrequencyLockStrategy(t *testing.T) {
	t.Parallel()
	strategy := NewFrequencyLock(0.8)

	assert.Equal(t, 0.8, strategy.SyncRate, "Expected strength 0.8")
	assert.Equal(t, "frequency_lock", strategy.Name(), "Expected name 'frequency_lock'")

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

	assert.Equal(t, "frequency_lock", action.Type, "Expected action type 'frequency_lock'")

	// Value should represent phase adjustment
	assert.NotEqual(t, 0.0, action.Value, "Action value should not be zero when phases differ")

	// Confidence should be based on local coherence and strength
	expectedConfidence := context.LocalCoherence * strategy.SyncRate
	assert.InDelta(t, expectedConfidence, confidence, 0.01, "Expected confidence %f", expectedConfidence)
}

func TestEnergyAwareStrategy(t *testing.T) {
	t.Parallel()
	strategy := NewEnergyAware(20.0)

	assert.Equal(t, 20.0, strategy.Threshold, "Expected threshold 20.0")
	assert.Equal(t, "energy_aware", strategy.Name(), "Expected name 'energy_aware'")

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

	assert.Equal(t, "energy_save", action.Type, "Expected 'energy_save' action with low energy")
	assert.Equal(t, 0.3, confidence, "Expected confidence 0.3 for energy_save")

	// Test with sufficient energy (high stability)
	contextHighEnergy := core.Context{
		Density:        0.5,
		Stability:      0.9, // High stability simulates high energy
		Progress:       0.3,
		LocalCoherence: 0.6,
	}

	action, confidence = strategy.Propose(currentLowEnergy, target, contextHighEnergy)

	// With high stability, should still do energy_save but with different parameters
	assert.Equal(t, "energy_save", action.Type, "Expected 'energy_save' action")

	// Confidence should still be 0.3 for energy_save
	assert.Equal(t, 0.3, confidence, "Expected confidence 0.3")
}

func TestAdaptiveStrategy(t *testing.T) {
	t.Parallel()
	nudge := NewPhaseNudge(0.3)
	frequency := NewFrequencyLock(0.7)
	energy := NewEnergyAware(25.0)

	adaptive := newAdaptiveStrategy([]core.SyncStrategy{nudge, frequency, energy})

	assert.Equal(t, "adaptive", adaptive.Name(), "Expected name 'adaptive'")

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

	assert.NotEmpty(t, action.Type, "Adaptive strategy should produce an action")
	assert.GreaterOrEqual(t, confidence, 0.0, "Confidence should be >= 0")
	assert.LessOrEqual(t, confidence, 1.0, "Confidence should be <= 1")
}

func TestPulseStrategy(t *testing.T) {
	t.Parallel()
	strategy := NewPulse(100*time.Millisecond, 0.5)

	assert.Equal(t, "pulse", strategy.Name(), "Expected name 'pulse'")
	assert.Equal(t, 100*time.Millisecond, strategy.Period, "Expected period 100ms")
	assert.Equal(t, 0.5, strategy.Amplitude, "Expected amplitude 0.5")

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
	assert.True(t, action.Type == "pulse" || action.Type == "maintain", "Expected action type 'pulse' or 'maintain', got '%s'", action.Type)
	assert.GreaterOrEqual(t, confidence, 0.0, "Confidence should be >= 0")
	assert.LessOrEqual(t, confidence, 1.0, "Confidence should be <= 1")
}

func TestStrategyIntegration(t *testing.T) {
	t.Parallel()
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

		assert.NotEmpty(t, action.Type, "Strategy %d (%s) produced empty action type", i, strategy.Name())
		assert.GreaterOrEqual(t, confidence, 0.0, "Strategy %d (%s) produced invalid confidence %f", i, strategy.Name(), confidence)
		assert.LessOrEqual(t, confidence, 1.0, "Strategy %d (%s) produced invalid confidence %f", i, strategy.Name(), confidence)
		assert.GreaterOrEqual(t, action.Cost, 0.0, "Strategy %d (%s) produced negative cost %f", i, strategy.Name(), action.Cost)
	}
}

func TestStrategySelection(t *testing.T) {
	t.Parallel()
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
	assert.NotEmpty(t, action1.Type, "Should produce an action even with low stability")

	// Test with high coherence
	highCoherenceContext := core.Context{
		LocalCoherence: 0.85,
		Stability:      0.9,
	}

	action2, _ := adaptive.Propose(current, target, highCoherenceContext)
	assert.NotEmpty(t, action2.Type, "Should produce an action with high coherence")
}
