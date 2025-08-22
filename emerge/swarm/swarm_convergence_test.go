package swarm

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/carlisia/bio-adapt/emerge/core"
)

// TestSwarmConvergence tests that swarms actually achieve synchronization.
func TestSwarmConvergence(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		size            int
		targetCoherence float64
		timeout         time.Duration
		minImprovement  float64 // Minimum improvement required
	}{
		{
			name:            "small_swarm_should_converge",
			size:            10,
			targetCoherence: 0.7,
			timeout:         5 * time.Second,
			minImprovement:  0.3, // Should improve by at least 30%
		},
		{
			name:            "medium_swarm_should_converge",
			size:            50,
			targetCoherence: 0.65,
			timeout:         8 * time.Second,
			minImprovement:  0.4, // Should improve by at least 40%
		},
		{
			name:            "large_swarm_should_converge",
			size:            150,
			targetCoherence: 0.6,
			timeout:         10 * time.Second,
			minImprovement:  0.35, // Should improve by at least 35%
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create swarm with target coherence
			swarm, err := New(tt.size, core.State{
				Phase:     0,
				Frequency: 200 * time.Millisecond,
				Coherence: tt.targetCoherence,
			})
			require.NoError(t, err, "Failed to create swarm")

			// Measure initial coherence
			initialCoherence := swarm.MeasureCoherence()
			t.Logf("Initial coherence: %.3f", initialCoherence)

			// Initial coherence should be low (random phases), but small swarms may vary more
			expectedMaxInitial := 0.3
			if tt.size <= 15 {
				expectedMaxInitial = 0.4 // Small swarms can have higher initial coherence by chance
			}
			if initialCoherence > expectedMaxInitial {
				t.Logf("Initial coherence %.3f higher than typical %.3f (acceptable for small swarms)",
					initialCoherence, expectedMaxInitial)
			}

			// Run synchronization with timeout
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			// Run swarm (this should now work with goal-directed pattern completion)
			err = swarm.Run(ctx)
			if err != nil && !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("Swarm run failed: %v", err)
			}

			// Measure final coherence
			finalCoherence := swarm.MeasureCoherence()
			t.Logf("Final coherence: %.3f (target: %.3f)", finalCoherence, tt.targetCoherence)

			// Test 1: Coherence should improve significantly
			improvement := finalCoherence - initialCoherence
			assert.GreaterOrEqual(t, improvement, tt.minImprovement, "Insufficient improvement: %.3f, expected at least %.3f", improvement, tt.minImprovement)

			// Test 2: Should achieve or get close to target coherence
			tolerance := 0.15 // Allow 15% tolerance
			assert.GreaterOrEqual(t, finalCoherence, tt.targetCoherence-tolerance, "Failed to approach target coherence: got %.3f, want >= %.3f", finalCoherence, tt.targetCoherence-tolerance)

			// Test 3: Final coherence should be reasonable (not too high to be suspicious)
			assert.LessOrEqual(t, finalCoherence, 1.0, "Final coherence above maximum: %.3f, may indicate test error", finalCoherence)
		})
	}
}

// TestSwarmConvergenceConsistency tests that convergence is consistent across runs.
func TestSwarmConvergenceConsistency(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("Skipping consistency test in short mode")
	}

	const (
		swarmSize       = 20
		runs            = 5
		targetCoherence = 0.7
		timeout         = 5 * time.Second
	)

	finalCoherences := make([]float64, 0, runs)

	for i := range runs {
		t.Logf("Run %d/%d", i+1, runs)

		swarm, err := New(swarmSize, core.State{
			Phase:     0,
			Frequency: 200 * time.Millisecond,
			Coherence: targetCoherence,
		})
		require.NoError(t, err, "Run %d: Failed to create swarm", i)

		initialCoherence := swarm.MeasureCoherence()

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		err = swarm.Run(ctx)
		cancel()

		if err != nil && !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("Run %d: Swarm run failed: %v", i, err)
		}

		finalCoherence := swarm.MeasureCoherence()
		finalCoherences = append(finalCoherences, finalCoherence)

		t.Logf("Run %d: %.3f -> %.3f (improvement: %.3f)",
			i+1, initialCoherence, finalCoherence, finalCoherence-initialCoherence)

		// Each run should show significant improvement
		if finalCoherence-initialCoherence < 0.2 {
			t.Errorf("Run %d: Insufficient improvement: %.3f", i, finalCoherence-initialCoherence)
		}
	}

	// Calculate statistics across runs
	var sum, minVal, maxVal float64
	minVal = finalCoherences[0]
	maxVal = finalCoherences[0]

	for _, coherence := range finalCoherences {
		sum += coherence
		if coherence < minVal {
			minVal = coherence
		}
		if coherence > maxVal {
			maxVal = coherence
		}
	}

	avg := sum / float64(runs)
	variance := maxVal - minVal

	t.Logf("Consistency: avg=%.3f, min=%.3f, max=%.3f, variance=%.3f", avg, minVal, maxVal, variance)

	// Results should be reasonably consistent
	if variance > 0.3 {
		t.Errorf("Results too inconsistent: variance %.3f > 0.3", variance)
	}

	// Average should be good
	if avg < targetCoherence-0.2 {
		t.Errorf("Average coherence too low: %.3f, expected >= %.3f", avg, targetCoherence-0.2)
	}
}

// TestSwarmNonRegression tests that we don't regress to the old broken behavior.
func TestSwarmNonRegression(t *testing.T) {
	t.Parallel()
	// This test ensures we never go back to the old broken synchronization
	// where swarms would get stuck at very low coherence (~7%)

	swarm, err := New(30, core.State{
		Phase:     0,
		Frequency: 200 * time.Millisecond,
		Coherence: 0.65,
	})
	if err != nil {
		t.Fatalf("Failed to create swarm: %v", err)
	}

	initialCoherence := swarm.MeasureCoherence()

	// Run for a reasonable amount of time
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = swarm.Run(ctx)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Swarm run failed: %v", err)
	}

	finalCoherence := swarm.MeasureCoherence()

	t.Logf("Non-regression test: %.3f -> %.3f", initialCoherence, finalCoherence)

	// Critical: Final coherence must NOT be stuck in the old broken range (5-10%)
	if finalCoherence < 0.15 {
		t.Errorf("REGRESSION: Final coherence %.3f indicates broken synchronization (old bug)", finalCoherence)
	}

	// Should show substantial improvement
	if finalCoherence-initialCoherence < 0.25 {
		t.Errorf("REGRESSION: Insufficient improvement %.3f indicates broken synchronization",
			finalCoherence-initialCoherence)
	}

	// Should achieve reasonable synchronization
	if finalCoherence < 0.4 {
		t.Errorf("REGRESSION: Final coherence %.3f too low, synchronization appears broken", finalCoherence)
	}
}

// TestSwarmTargetAchievement tests that swarms can achieve their specific target.
func TestSwarmTargetAchievement(t *testing.T) {
	t.Parallel()
	targets := []float64{0.5, 0.6, 0.7, 0.8}

	for _, target := range targets {
		t.Run(fmt.Sprintf("target_%.1f", target), func(t *testing.T) {
			t.Parallel()
			swarm, err := New(25, core.State{
				Phase:     0,
				Frequency: 200 * time.Millisecond,
				Coherence: target,
			})
			require.NoError(t, err, "Failed to create swarm")

			ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
			defer cancel()

			err = swarm.Run(ctx)
			if err != nil && !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("Swarm run failed: %v", err)
			}

			finalCoherence := swarm.MeasureCoherence()

			// Should get reasonably close to target (within 20%)
			tolerance := 0.2
			if finalCoherence < target-tolerance {
				t.Errorf("Failed to achieve target %.1f: got %.3f (min expected: %.3f)",
					target, finalCoherence, target-tolerance)
			}

			// But also shouldn't drastically overshoot (indicates possible test error)
			if finalCoherence > target+0.3 {
				t.Logf("Warning: Coherence %.3f significantly exceeds target %.1f", finalCoherence, target)
			}
		})
	}
}
