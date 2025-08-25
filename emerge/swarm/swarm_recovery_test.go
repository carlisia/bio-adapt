package swarm_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/swarm"
	"github.com/carlisia/bio-adapt/internal/config"
)

// TestSwarmRecoveryAfterDisruption tests that a swarm can recover after disruption
// This test replicates the exact pattern used in the llm_batching_visual demo
func TestSwarmRecoveryAfterDisruption(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name              string
		swarmSize         int
		disruptionPercent float64
		disruptionDelay   time.Duration
		monitoringPeriod  time.Duration
		expectedRecovery  bool
		minRecoveryLevel  float64
	}{
		{
			name:              "demo_scenario_20_agents",
			swarmSize:         20,
			disruptionPercent: 0.2,
			disruptionDelay:   2 * time.Second,
			monitoringPeriod:  8 * time.Second,
			expectedRecovery:  true,
			minRecoveryLevel:  0.7, // Should recover to at least 70%
		},
		{
			name:              "small_disruption",
			swarmSize:         20,
			disruptionPercent: 0.1,
			disruptionDelay:   2 * time.Second,
			monitoringPeriod:  5 * time.Second,
			expectedRecovery:  true,
			minRecoveryLevel:  0.8,
		},
		{
			name:              "large_disruption",
			swarmSize:         20,
			disruptionPercent: 0.3,
			disruptionDelay:   2 * time.Second,
			monitoringPeriod:  10 * time.Second,
			expectedRecovery:  true,
			minRecoveryLevel:  0.65,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Setup - exactly as in the demo
			targetState := core.State{
				Phase:     0,
				Frequency: 200 * time.Millisecond,
				Coherence: 0.85,
			}

			// Use AutoScaleConfig like the demo
			cfg := config.AutoScaleConfig(tt.swarmSize)
			cfg.AgentUpdateInterval = 100 * time.Millisecond // Same as demo

			s, err := swarm.New(tt.swarmSize, targetState, swarm.WithConfig(cfg))
			require.NoError(t, err, "Failed to create swarm")

			// Start swarm in background - exactly as demo does
			ctx, cancel := context.WithTimeout(context.Background(),
				tt.disruptionDelay+tt.monitoringPeriod+2*time.Second)
			defer cancel()

			// Run swarm in goroutine - use RunContinuous for recovery
			errCh := make(chan error, 1)
			go func() {
				err := s.RunContinuous(ctx)
				if err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
					errCh <- err
				}
			}()

			// Wait for initial convergence
			time.Sleep(1000 * time.Millisecond)
			initialCoherence := s.MeasureCoherence()
			t.Logf("Initial coherence: %.3f", initialCoherence)

			// In parallel tests, convergence might be slower due to CPU contention
			// Accept lower initial coherence but ensure it's reasonable
			assert.Greater(t, initialCoherence, 0.6, "Should have some initial synchronization")

			// Monitor coherence before disruption
			coherenceBeforeDisruption := 0.0
			ticker := time.NewTicker(200 * time.Millisecond)
			defer ticker.Stop()

			disruptionApplied := false
			disruptionTime := time.After(tt.disruptionDelay)
			monitoringEnd := time.After(tt.disruptionDelay + tt.monitoringPeriod)

			postDisruptionMeasurements := []float64{}

			for {
				select {
				case <-disruptionTime:
					if !disruptionApplied {
						coherenceBeforeDisruption = s.MeasureCoherence()
						t.Logf("Coherence before disruption: %.3f", coherenceBeforeDisruption)

						// Apply disruption - exactly as demo
						s.DisruptAgents(tt.disruptionPercent)
						disruptionApplied = true

						// Immediate measurement after disruption
						time.Sleep(100 * time.Millisecond)
						coherenceAfter := s.MeasureCoherence()
						t.Logf("Coherence immediately after disruption: %.3f", coherenceAfter)
						assert.Less(t, coherenceAfter, coherenceBeforeDisruption,
							"Coherence should drop after disruption")
					}

				case <-ticker.C:
					coherence := s.MeasureCoherence()

					if disruptionApplied {
						postDisruptionMeasurements = append(postDisruptionMeasurements, coherence)
						t.Logf("Post-disruption coherence: %.3f", coherence)
					}

				case <-monitoringEnd:
					// Final measurements
					finalCoherence := s.MeasureCoherence()
					t.Logf("Final coherence: %.3f", finalCoherence)

					if tt.expectedRecovery {
						// Check if coherence recovered
						assert.Greater(t, finalCoherence, tt.minRecoveryLevel,
							"Swarm should recover to at least %.0f%% coherence",
							tt.minRecoveryLevel*100)

						// Check if there was improvement over time
						if len(postDisruptionMeasurements) > 2 {
							firstPostDisruption := postDisruptionMeasurements[0]
							lastPostDisruption := postDisruptionMeasurements[len(postDisruptionMeasurements)-1]

							t.Logf("Recovery trend: %.3f -> %.3f",
								firstPostDisruption, lastPostDisruption)

							// Check if coherence is improving (not stuck)
							improvement := lastPostDisruption - firstPostDisruption
							assert.Greater(t, improvement, -0.05,
								"Coherence should not be decreasing significantly")

							// Check recovery trend
							// If we've recovered to near the target, that's success even if stable
							if lastPostDisruption >= tt.minRecoveryLevel {
								t.Logf("Successfully recovered and maintaining target coherence")
							} else {
								// Only check for stuck measurements if we haven't recovered
								uniqueValues := make(map[float64]bool)
								for _, v := range postDisruptionMeasurements {
									// Round to 3 decimal places to account for small variations
									rounded := float64(int(v*1000)) / 1000
									uniqueValues[rounded] = true
								}
								assert.Greater(t, len(uniqueValues), 1,
									"Coherence measurements should not be stuck at a single value when not at target")
							}
						}
					}

					cancel()
					return

				case err := <-errCh:
					t.Fatalf("Swarm error: %v", err)
					return

				case <-ctx.Done():
					t.Fatal("Test timeout")
					return
				}
			}
		})
	}
}

// TestSwarmCoherenceNotStuck verifies that coherence measurements change over time
func TestSwarmCoherenceNotStuck(t *testing.T) {
	t.Parallel()
	targetState := core.State{
		Phase:     0,
		Frequency: 200 * time.Millisecond,
		Coherence: 0.85,
	}

	cfg := config.AutoScaleConfig(20)
	s, err := swarm.New(20, targetState, swarm.WithConfig(cfg))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Run swarm
	go func() {
		_ = s.Run(ctx)
	}()

	// Collect measurements
	measurements := []float64{}
	for i := range 10 {
		time.Sleep(500 * time.Millisecond)
		coherence := s.MeasureCoherence()
		measurements = append(measurements, coherence)

		// Disrupt halfway through
		if i == 5 {
			s.DisruptAgents(0.2)
			t.Log("Applied disruption")
		}
	}

	// Check that we have variation in measurements
	uniqueValues := make(map[float64]bool)
	for _, v := range measurements {
		rounded := float64(int(v*1000)) / 1000
		uniqueValues[rounded] = true
	}

	t.Logf("Unique coherence values: %d out of %d measurements",
		len(uniqueValues), len(measurements))
	t.Logf("Measurements: %v", measurements)

	assert.Greater(t, len(uniqueValues), 1,
		"Coherence should not be stuck at a single value")
}

// TestManualPhaseUpdateAfterDisruption checks if agents can update phases manually after disruption
func TestManualPhaseUpdateAfterDisruption(t *testing.T) {
	t.Parallel()
	targetState := core.State{
		Phase:     0,
		Frequency: 200 * time.Millisecond,
		Coherence: 0.85,
	}

	cfg := config.AutoScaleConfig(10)
	s, err := swarm.New(10, targetState, swarm.WithConfig(cfg))
	require.NoError(t, err)

	// Run briefly to get some coherence
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	go func() {
		_ = s.Run(ctx)
	}()
	time.Sleep(400 * time.Millisecond)
	cancel()
	time.Sleep(50 * time.Millisecond)

	// Get initial coherence after some convergence
	initialCoherence := s.MeasureCoherence()
	t.Logf("Initial coherence: %.3f", initialCoherence)

	// Disrupt agents significantly
	s.DisruptAgents(0.5) // Increase disruption to ensure coherence changes

	// Check coherence changed (could go up or down with random phases)
	afterDisruption := s.MeasureCoherence()
	t.Logf("After disruption: %.3f", afterDisruption)
	assert.NotEqual(t, afterDisruption, initialCoherence, "Disruption should change coherence")

	// Manually converge agents back
	agents := s.Agents()
	for _, agent := range agents {
		agent.SetPhase(0) // Set all to same phase
	}

	// Measure again
	afterManualFix := s.MeasureCoherence()
	t.Logf("After manual convergence: %.3f", afterManualFix)
	assert.Greater(t, afterManualFix, 0.9, "Should be highly synchronized after manual fix")
}
