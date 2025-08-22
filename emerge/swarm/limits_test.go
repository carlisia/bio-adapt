//nolint:paralleltest // Limits tests don't need parallelization
package swarm

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCoherenceLimits(t *testing.T) {
	tests := []struct {
		name      string
		swarmSize int
		wantMin   float64 // Minimum expected practical limit
		wantMax   float64 // Maximum expected theoretical limit
	}{
		{
			name:      "single_agent",
			swarmSize: 1,
			wantMin:   1.0,
			wantMax:   1.0,
		},
		{
			name:      "very_small_swarm",
			swarmSize: 3,
			wantMin:   0.8,
			wantMax:   0.95,
		},
		{
			name:      "small_swarm",
			swarmSize: 10,
			wantMin:   0.9,
			wantMax:   0.99,
		},
		{
			name:      "medium_swarm",
			swarmSize: 50,
			wantMin:   0.95,
			wantMax:   0.999,
		},
		{
			name:      "large_swarm",
			swarmSize: 100,
			wantMin:   0.97,
			wantMax:   0.999,
		},
		{
			name:      "very_large_swarm",
			swarmSize: 1000,
			wantMin:   0.98,
			wantMax:   0.999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limits := GetCoherenceLimits(tt.swarmSize)

			// Check that limits are ordered correctly
			assert.LessOrEqual(t, limits.Recommended, limits.Practical,
				"Recommended should be <= Practical")
			assert.LessOrEqual(t, limits.Practical, limits.Theoretical,
				"Practical should be <= Theoretical")

			// Check minimum practical limit
			assert.GreaterOrEqual(t, limits.Practical, tt.wantMin,
				"Practical limit should be >= %f", tt.wantMin)

			// Check maximum theoretical limit
			assert.LessOrEqual(t, limits.Theoretical, tt.wantMax,
				"Theoretical limit should be <= %f", tt.wantMax)
		})
	}
}

func TestGetMinimumSwarmSize(t *testing.T) {
	tests := []struct {
		coherence    float64
		minSwarmSize int
	}{
		{0.80, 2},
		{0.85, 5},
		{0.90, 10},
		{0.95, 20},
		{0.97, 50},
		{0.99, 100},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			size := GetMinimumSwarmSize(tt.coherence)
			assert.GreaterOrEqual(t, size, tt.minSwarmSize,
				"For coherence %f, minimum swarm size should be >= %d",
				tt.coherence, tt.minSwarmSize)
		})
	}
}

func TestGetConvergenceTimeFactor(t *testing.T) {
	tests := []struct {
		name      string
		swarmSize int
		coherence float64
		wantMin   float64
		wantMax   float64
	}{
		{
			name:      "easy_target",
			swarmSize: 100,
			coherence: 0.7, // Well below practical limit
			wantMin:   1.0,
			wantMax:   1.0,
		},
		{
			name:      "moderate_target",
			swarmSize: 100,
			coherence: 0.9,
			wantMin:   1.0,
			wantMax:   4.0,
		},
		{
			name:      "hard_target",
			swarmSize: 100,
			coherence: 0.97,
			wantMin:   3.5,
			wantMax:   8.5,
		},
		{
			name:      "impossible_target",
			swarmSize: 10,
			coherence: 0.99,
			wantMin:   math.Inf(1),
			wantMax:   math.Inf(1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factor := GetConvergenceTimeFactor(tt.swarmSize, tt.coherence)
			if math.IsInf(tt.wantMin, 1) {
				assert.True(t, math.IsInf(factor, 1),
					"Should return infinite time for impossible target")
			} else {
				assert.GreaterOrEqual(t, factor, tt.wantMin,
					"Time factor should be >= %f", tt.wantMin)
				assert.LessOrEqual(t, factor, tt.wantMax,
					"Time factor should be <= %f", tt.wantMax)
			}
		})
	}
}

func TestValidateCoherenceTarget(t *testing.T) {
	tests := []struct {
		name         string
		swarmSize    int
		requested    float64
		expectAdjust bool
		expectWarn   bool
	}{
		{
			name:         "reasonable_target",
			swarmSize:    100,
			requested:    0.90,
			expectAdjust: false,
			expectWarn:   false,
		},
		{
			name:         "ambitious_target",
			swarmSize:    50,
			requested:    0.96,
			expectAdjust: false,
			expectWarn:   true,
		},
		{
			name:         "impossible_target",
			swarmSize:    10,
			requested:    0.99,
			expectAdjust: false, // We don't adjust, just warn
			expectWarn:   true,
		},
		{
			name:         "beyond_theoretical",
			swarmSize:    5,
			requested:    1.0,
			expectAdjust: true,
			expectWarn:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adjusted, warning := ValidateCoherenceTarget(tt.swarmSize, tt.requested)

			if tt.expectAdjust {
				assert.NotEqual(t, tt.requested, adjusted,
					"Should adjust impossible target")
			} else {
				assert.Equal(t, tt.requested, adjusted,
					"Should not adjust achievable target")
			}

			if tt.expectWarn {
				assert.NotEmpty(t, warning, "Should provide warning")
			} else {
				assert.Empty(t, warning, "Should not warn for reasonable target")
			}
		})
	}
}
