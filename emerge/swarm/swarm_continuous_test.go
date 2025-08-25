package swarm_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/swarm"
	"github.com/carlisia/bio-adapt/internal/config"
)

// TestRunContinuous tests the RunContinuous method
func TestRunContinuous(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		swarmSize      int
		disruptionTime time.Duration
		disruptionPct  float64
		testDuration   time.Duration
		expectRecovery bool
	}{
		{
			name:           "basic_continuous_run",
			swarmSize:      20,
			disruptionTime: 0, // No disruption
			disruptionPct:  0,
			testDuration:   2 * time.Second,
			expectRecovery: false, // No recovery needed
		},
		{
			name:           "recovers_from_disruption",
			swarmSize:      20,
			disruptionTime: 500 * time.Millisecond,
			disruptionPct:  0.5,
			testDuration:   3 * time.Second,
			expectRecovery: true,
		},
		{
			name:           "recovers_from_heavy_disruption",
			swarmSize:      30,
			disruptionTime: 800 * time.Millisecond,
			disruptionPct:  0.8,
			testDuration:   4 * time.Second,
			expectRecovery: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			goal := core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.85,
			}

			cfg := config.AutoScaleConfig(tt.swarmSize)
			s, err := swarm.New(tt.swarmSize, goal, swarm.WithConfig(cfg))
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), tt.testDuration)
			defer cancel()

			// Track coherence measurements
			var measurements []float64
			measurementDone := make(chan struct{})

			// Monitor coherence in background
			go func() {
				ticker := time.NewTicker(100 * time.Millisecond)
				defer ticker.Stop()

				for {
					select {
					case <-ticker.C:
						coherence := s.MeasureCoherence()
						measurements = append(measurements, coherence)
					case <-ctx.Done():
						close(measurementDone)
						return
					}
				}
			}()

			// Apply disruption if specified
			if tt.disruptionTime > 0 {
				go func() {
					time.Sleep(tt.disruptionTime)
					t.Logf("Applying %.0f%% disruption at %v", tt.disruptionPct*100, tt.disruptionTime)
					s.DisruptAgents(tt.disruptionPct)
				}()
			}

			// Run continuously
			err = s.RunContinuous(ctx)
			assert.ErrorIs(t, err, context.DeadlineExceeded, "Should exit with deadline exceeded")

			// Wait for measurements to complete
			<-measurementDone

			// Analyze results
			require.Greater(t, len(measurements), 10, "Should have enough measurements")

			if tt.expectRecovery {
				// Find the disruption point (lowest coherence)
				minCoherence := measurements[0]
				minIdx := 0
				for i, c := range measurements {
					if c < minCoherence {
						minCoherence = c
						minIdx = i
					}
				}

				// Check recovery after disruption
				if minIdx < len(measurements)-5 {
					// Average of last few measurements
					var finalAvg float64
					count := 0
					for i := len(measurements) - 5; i < len(measurements); i++ {
						finalAvg += measurements[i]
						count++
					}
					finalAvg /= float64(count)

					t.Logf("Min coherence: %.3f at index %d, Final avg: %.3f",
						minCoherence, minIdx, finalAvg)

					assert.Greater(t, finalAvg, minCoherence+0.1,
						"Should recover from disruption")
				}
			}
		})
	}
}

// TestRunContinuousMultipleDisruptions tests recovery from multiple disruptions
func TestRunContinuousMultipleDisruptions(t *testing.T) {
	t.Parallel()

	goal := core.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.80,
	}

	s, err := swarm.New(25, goal)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Count disruptions applied
	disruptionCount := int32(0)

	// Apply multiple disruptions with absolute times
	go func() {
		startTime := time.Now()
		disruptions := []struct {
			atTime time.Duration
			pct    float64
		}{
			{500 * time.Millisecond, 0.3},
			{1000 * time.Millisecond, 0.4},
			{1500 * time.Millisecond, 0.2},
			{2000 * time.Millisecond, 0.5},
		}

		for _, d := range disruptions {
			waitTime := d.atTime - time.Since(startTime)
			if waitTime > 0 {
				select {
				case <-time.After(waitTime):
					s.DisruptAgents(d.pct)
					atomic.AddInt32(&disruptionCount, 1)
					t.Logf("Applied disruption %d: %.0f%% at %v",
						atomic.LoadInt32(&disruptionCount), d.pct*100, time.Since(startTime))
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	// Run continuously
	err = s.RunContinuous(ctx)
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	// Verify we applied disruptions
	finalCount := atomic.LoadInt32(&disruptionCount)
	assert.Equal(t, int32(4), finalCount, "Should have applied all disruptions")

	// Check final coherence - should have recovered
	finalCoherence := s.MeasureCoherence()
	t.Logf("Final coherence after %d disruptions: %.3f", finalCount, finalCoherence)
	assert.Greater(t, finalCoherence, 0.4,
		"Should maintain reasonable coherence despite multiple disruptions")
}

// TestRunContinuousVsRun compares RunContinuous with Run behavior
func TestRunContinuousVsRun(t *testing.T) {
	t.Parallel()

	goal := core.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.85,
	}

	// Test with Run()
	t.Run("with_Run", func(t *testing.T) {
		t.Parallel()
		s, err := swarm.New(20, goal)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		// Track when Run exits
		runExited := make(chan struct{})
		go func() {
			_ = s.Run(ctx)
			close(runExited)
		}()

		// Let it converge initially
		time.Sleep(1 * time.Second)

		// Apply disruption after initial convergence
		beforeDisruption := s.MeasureCoherence()
		s.DisruptAgents(0.5)
		afterDisruption := s.MeasureCoherence()
		t.Logf("Disrupted swarm using Run(): %.3f -> %.3f",
			beforeDisruption, afterDisruption)

		// Wait a bit more
		time.Sleep(1 * time.Second)

		// Measure coherence - should not recover well
		finalCoherence := s.MeasureCoherence()
		t.Logf("Final coherence with Run(): %.3f", finalCoherence)

		// Run() may or may not exit depending on whether it achieved initial convergence
		select {
		case <-runExited:
			t.Log("Run() exited")
		case <-time.After(100 * time.Millisecond):
			t.Log("Run() still running")
		}
	})

	// Test with RunContinuous()
	t.Run("with_RunContinuous", func(t *testing.T) {
		t.Parallel()

		// Create swarm with explicit recovery config
		recoveryConfig := swarm.DefaultRecoveryConfig(goal.Coherence)
		s, err := swarm.New(20, goal, swarm.WithRecoveryConfig(recoveryConfig))
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		// Apply disruption after 1 second
		go func() {
			time.Sleep(1 * time.Second)
			beforeDisruption := s.MeasureCoherence()
			s.DisruptAgents(0.5)
			afterDisruption := s.MeasureCoherence()
			t.Logf("Disrupted swarm using RunContinuous(): %.3f -> %.3f",
				beforeDisruption, afterDisruption)
		}()

		// Track when RunContinuous exits
		runExited := make(chan struct{})
		go func() {
			_ = s.RunContinuous(ctx)
			close(runExited)
		}()

		select {
		case <-runExited:
			// Should only exit when context is done
			assert.True(t, ctx.Err() != nil,
				"RunContinuous should only exit when context is done")
		case <-time.After(100 * time.Millisecond):
			// RunContinuous should still be running
			t.Log("RunContinuous still running as expected")
		}

		// Wait for context to expire
		<-ctx.Done()
		<-runExited

		// Measure final coherence
		finalCoherence := s.MeasureCoherence()
		t.Logf("Final coherence with RunContinuous(): %.3f", finalCoherence)

		// Check recovery based on configured minimum viable coherence
		// Target is 0.85, which gets MinViableCoherenceMedium = 0.4
		assert.Greater(t, finalCoherence, recoveryConfig.MinimumViableCoherence,
			"RunContinuous should maintain coherence above minimum viable threshold")
	})
}

// TestRunContinuousContextCancellation tests proper context handling
func TestRunContinuousContextCancellation(t *testing.T) {
	t.Parallel()

	goal := core.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.85,
	}

	s, err := swarm.New(20, goal)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	// Run in background
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.RunContinuous(ctx)
	}()

	// Let it run briefly
	time.Sleep(500 * time.Millisecond)

	// Cancel context
	cancel()

	// Should exit quickly
	select {
	case err := <-errCh:
		assert.ErrorIs(t, err, context.Canceled,
			"Should return context.Canceled error")
	case <-time.After(1 * time.Second):
		t.Error("RunContinuous did not exit quickly after cancellation")
	}
}

// TestRunContinuousDetectsDisruption tests disruption detection logic
func TestRunContinuousDetectsDisruption(t *testing.T) {
	t.Parallel()

	goal := core.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.90,
	}

	cfg := config.AutoScaleConfig(30)
	s, err := swarm.New(30, goal, swarm.WithConfig(cfg))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	coherenceDropDetected := false
	recoveryDetected := false

	// Monitor coherence changes
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		var lastCoherence float64
		highCoherenceReached := false

		for {
			select {
			case <-ticker.C:
				coherence := s.MeasureCoherence()

				// Detect when we reach high coherence
				if coherence > 0.85 && !highCoherenceReached {
					highCoherenceReached = true
					lastCoherence = coherence
					t.Logf("High coherence reached: %.3f", coherence)
				}

				// Detect significant drop
				if highCoherenceReached && coherence < lastCoherence-0.1 {
					coherenceDropDetected = true
					t.Logf("Coherence drop detected: %.3f -> %.3f",
						lastCoherence, coherence)
				}

				// Detect recovery
				if coherenceDropDetected && coherence > 0.8 {
					recoveryDetected = true
					t.Logf("Recovery detected: %.3f", coherence)
				}

				lastCoherence = coherence

			case <-ctx.Done():
				return
			}
		}
	}()

	// Apply disruption after initial convergence
	go func() {
		time.Sleep(1500 * time.Millisecond)
		t.Log("Applying disruption")
		s.DisruptAgents(0.6)
	}()

	// Run continuously
	_ = s.RunContinuous(ctx)

	// Verify disruption detection and recovery
	assert.True(t, coherenceDropDetected,
		"Should detect coherence drop after disruption")
	assert.True(t, recoveryDetected,
		"Should detect recovery after disruption")
}
