//nolint:paralleltest // Tests modify shared config state
package swarm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/carlisia/bio-adapt/emerge/goal"
	"github.com/carlisia/bio-adapt/emerge/scale"
	"github.com/carlisia/bio-adapt/emerge/trait"
)

func TestConfigFor(t *testing.T) {
	tests := []struct {
		name  string
		goal  goal.Type
		check func(t *testing.T, cfg *Config)
	}{
		{
			name: "minimize_api_calls",
			goal: goal.MinimizeAPICalls,
			check: func(t *testing.T, cfg *Config) {
				t.Helper()
				assert.Equal(t, 0.90, cfg.Convergence.PhaseConvergenceGoal)
				assert.Equal(t, 0.75, cfg.Convergence.BaseAdjustmentScale)
				assert.Equal(t, 0.3, cfg.Thresholds.PhaseVariance)
			},
		},
		{
			name: "reach_consensus",
			goal: goal.ReachConsensus,
			check: func(t *testing.T, cfg *Config) {
				t.Helper()
				assert.Equal(t, 0.95, cfg.Convergence.PhaseConvergenceGoal)
				assert.Equal(t, 0.80, cfg.Convergence.BaseAdjustmentScale)
				assert.Equal(t, 0.2, cfg.Thresholds.PhaseVariance)
			},
		},
		{
			name: "distribute_load",
			goal: goal.DistributeLoad,
			check: func(t *testing.T, cfg *Config) {
				t.Helper()
				assert.Equal(t, 0.30, cfg.Convergence.PhaseConvergenceGoal)
				assert.Equal(t, 0.8, cfg.Thresholds.PhaseVariance)
				assert.Greater(t, cfg.Variation.BaseRange[1], cfg.Variation.BaseRange[0])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := For(tt.goal)
			require.NotNil(t, cfg)
			tt.check(t, cfg)

			// Validate the configuration
			err := cfg.Validate()
			assert.NoError(t, err)
		})
	}
}

func TestConfigTuneFor(t *testing.T) {
	tests := []struct {
		name   string
		target trait.Target
		check  func(t *testing.T, before, after *Config)
	}{
		{
			name:   "stability",
			target: trait.Stability,
			check: func(t *testing.T, before, after *Config) {
				t.Helper()
				// Stability makes adjustments gentler
				assert.Less(t, after.Convergence.BaseAdjustmentScale, before.Convergence.BaseAdjustmentScale)
				assert.Greater(t, after.Variation.BaseRange[1], before.Variation.BaseRange[1])
				assert.Greater(t, after.Strategy.UpdateInterval, before.Strategy.UpdateInterval)
			},
		},
		{
			name:   "speed",
			target: trait.Speed,
			check: func(t *testing.T, before, after *Config) {
				t.Helper()
				// Speed makes convergence faster
				assert.Greater(t, after.Convergence.BaseAdjustmentScale, before.Convergence.BaseAdjustmentScale)
				assert.Less(t, after.Strategy.UpdateInterval, before.Strategy.UpdateInterval)
			},
		},
		{
			name:   "efficiency",
			target: trait.Efficiency,
			check: func(t *testing.T, before, after *Config) {
				t.Helper()
				// Efficiency reduces resource usage
				assert.Greater(t, after.Strategy.UpdateInterval, before.Strategy.UpdateInterval)
				assert.Less(t, after.Variation.PerturbationChance, before.Variation.PerturbationChance)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Start with a base config
			before := For(goal.ReachConsensus)

			// Make a copy for comparison
			after := &Config{}
			*after = *before

			// Apply tuning
			after.TuneFor(tt.target)

			tt.check(t, before, after)

			// Validate the modified configuration
			err := after.Validate()
			assert.NoError(t, err)
		})
	}
}

func TestConfigWith(t *testing.T) {
	tests := []struct {
		name  string
		scale scale.Size
		check func(t *testing.T, cfg *Config)
	}{
		{
			name:  "tiny",
			scale: scale.Tiny,
			check: func(t *testing.T, cfg *Config) {
				t.Helper()
				assert.Equal(t, 0.020, cfg.Convergence.ToleranceSmall)
				assert.Equal(t, 20*time.Millisecond, cfg.Strategy.UpdateInterval)
			},
		},
		{
			name:  "large",
			scale: scale.Large,
			check: func(t *testing.T, cfg *Config) {
				t.Helper()
				// Large swarms need more tolerance
				// Note: MinimizeAPICalls starts with 0.01, * 1.5 = 0.015
				assert.GreaterOrEqual(t, cfg.Convergence.ToleranceSmall, 0.015)
				assert.GreaterOrEqual(t, cfg.Strategy.UpdateInterval, 150*time.Millisecond)
			},
		},
		{
			name:  "huge",
			scale: scale.Huge,
			check: func(t *testing.T, cfg *Config) {
				t.Helper()
				// Huge swarms need significant adjustments
				// Note: MinimizeAPICalls starts with 0.01, * 2.0 = 0.020
				assert.GreaterOrEqual(t, cfg.Convergence.ToleranceSmall, 0.020)
				assert.GreaterOrEqual(t, cfg.Strategy.UpdateInterval, 200*time.Millisecond)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := For(goal.MinimizeAPICalls).WithSize(tt.scale.DefaultAgentCount())
			require.NotNil(t, cfg)
			tt.check(t, cfg)

			// Validate the configuration
			err := cfg.Validate()
			assert.NoError(t, err)
		})
	}
}

func TestConfigChaining(t *testing.T) {
	// Test the full fluent API
	cfg := For(goal.MinimizeAPICalls).
		TuneFor(trait.Stability).
		WithSize(scale.Large.DefaultAgentCount())

	require.NotNil(t, cfg)

	// Should have properties from all modifiers
	assert.Equal(t, 0.90, cfg.Convergence.PhaseConvergenceGoal) // From MinimizeAPICalls
	assert.Less(t, cfg.Convergence.BaseAdjustmentScale, 0.75)   // Reduced by Stability
	assert.Greater(t, cfg.Convergence.ToleranceSmall, 0.010)    // Increased by Large scale

	// Validate the final configuration
	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*Config)
		wantErr bool
	}{
		{
			name: "valid_config",
			modify: func(_ *Config) {
				// No modifications - should be valid
			},
			wantErr: false,
		},
		{
			name: "invalid_adjustment_scale",
			modify: func(c *Config) {
				c.Convergence.BaseAdjustmentScale = 1.5 // > 1
			},
			wantErr: true,
		},
		{
			name: "invalid_phase_convergence_goal",
			modify: func(c *Config) {
				c.Convergence.PhaseConvergenceGoal = -0.1 // < 0
			},
			wantErr: true,
		},
		{
			name: "invalid_threshold_order",
			modify: func(c *Config) {
				c.Thresholds.HighCoherence = 0.95
				c.Thresholds.VeryHighCoherence = 0.90 // Should be >= HighCoherence
			},
			wantErr: true,
		},
		{
			name: "invalid_variation_range",
			modify: func(c *Config) {
				c.Variation.BaseRange = [2]float64{0.5, 0.3} // min > max
			},
			wantErr: true,
		},
		{
			name: "invalid_perturbation_chance",
			modify: func(c *Config) {
				c.Variation.PerturbationChance = 1.5 // > 1
			},
			wantErr: true,
		},
		{
			name: "invalid_update_interval",
			modify: func(c *Config) {
				c.Strategy.UpdateInterval = 0 // Must be positive
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := For(goal.ReachConsensus)
			tt.modify(cfg)

			err := cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigUsageExamples(t *testing.T) {
	t.Run("consensus_scenario", func(t *testing.T) {
		// Simple consensus scenario
		cfg := For(goal.ReachConsensus)
		assert.Equal(t, 0.95, cfg.Convergence.PhaseConvergenceGoal)
		assert.NoError(t, cfg.Validate())
	})

	t.Run("fast_convergence_small_swarm", func(t *testing.T) {
		// Fast convergence for small swarm
		cfg := For(goal.MinimizeAPICalls).
			TuneFor(trait.Speed).
			WithSize(scale.Small.DefaultAgentCount())

		assert.Less(t, cfg.Strategy.UpdateInterval, 100*time.Millisecond)
		assert.NoError(t, cfg.Validate())
	})

	t.Run("large_swarm_with_robustness", func(t *testing.T) {
		// Large swarm with robustness
		cfg := For(goal.DistributeLoad).
			TuneFor(trait.Resilience).
			WithSize(scale.Large.DefaultAgentCount())

		assert.Greater(t, cfg.Resonance.NoiseMagnitude, 0.5)
		assert.NoError(t, cfg.Validate())
	})

	t.Run("energy_efficient_operation", func(t *testing.T) {
		// Energy-efficient operation
		cfg := For(goal.SaveEnergy).
			TuneFor(trait.Efficiency)

		assert.Greater(t, cfg.Strategy.UpdateInterval, 200*time.Millisecond)
		assert.NoError(t, cfg.Validate())
	})
}
