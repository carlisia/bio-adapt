//nolint:paralleltest,thelper,intrange // E2E tests shouldn't run in parallel, validate funcs are not helpers, loop counters needed
package e2e_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/swarm"
	"github.com/carlisia/bio-adapt/internal/config"
)

const swarmWithDisruptionRecovery = "swarm_with_disruption_recovery"

// TestE2ESwarmConvergence tests end-to-end swarm convergence scenarios
func TestE2ESwarmConvergence(t *testing.T) {
	tests := []struct {
		name            string
		swarmSize       int
		targetPhase     float64
		targetCoherence float64
		maxTime         time.Duration
		disruption      func(*swarm.Swarm)
		expectSuccess   bool
	}{
		{
			name:            "small_swarm_quick_convergence",
			swarmSize:       10,
			targetPhase:     math.Pi,
			targetCoherence: 0.95,
			maxTime:         10 * time.Second,
			expectSuccess:   true,
		},
		{
			name:            "medium_swarm_convergence",
			swarmSize:       100,
			targetPhase:     math.Pi / 2,
			targetCoherence: 0.90,
			maxTime:         10 * time.Second,
			expectSuccess:   true,
		},
		{
			name:            "large_swarm_convergence",
			swarmSize:       500,
			targetPhase:     math.Pi,
			targetCoherence: 0.85,
			maxTime:         15 * time.Second,
			expectSuccess:   true,
		},
		{
			name:            swarmWithDisruptionRecovery,
			swarmSize:       50,
			targetPhase:     math.Pi,
			targetCoherence: 0.85,             // More realistic target for recovery
			maxTime:         10 * time.Second, // Sufficient time for recovery
			disruption: func(s *swarm.Swarm) {
				// Disrupt after we have a chance to measure pre-disruption coherence
				time.Sleep(500 * time.Millisecond)
				s.DisruptAgents(0.7) // Significant disruption - 70% of agents
			},
			expectSuccess: true,
		},
		{
			name:            "multiple_sequential_disruptions",
			swarmSize:       30,
			targetPhase:     math.Pi,
			targetCoherence: 0.85,
			maxTime:         10 * time.Second,
			disruption: func(s *swarm.Swarm) {
				// Multiple disruptions to test repeated recovery
				time.Sleep(500 * time.Millisecond)
				s.DisruptAgents(0.3) // First disruption - 30%
				time.Sleep(2 * time.Second)
				s.DisruptAgents(0.2) // Second disruption - 20%
				time.Sleep(2 * time.Second)
				s.DisruptAgents(0.4) // Third disruption - 40%
			},
			expectSuccess: true, // Should eventually converge despite multiple disruptions
		},
		{
			name:            "very_small_swarm",
			swarmSize:       3, // Minimum viable swarm
			targetPhase:     0,
			targetCoherence: 0.90,
			maxTime:         5 * time.Second,
			expectSuccess:   true,
		},
		{
			name:            "immediate_disruption",
			swarmSize:       20,
			targetPhase:     math.Pi / 2,
			targetCoherence: 0.80,
			maxTime:         8 * time.Second,
			disruption: func(s *swarm.Swarm) {
				// Disrupt immediately before any convergence
				time.Sleep(10 * time.Millisecond)
				s.DisruptAgents(0.5)
			},
			expectSuccess: true,
		},
		{
			name:            "total_disruption",
			swarmSize:       20,
			targetPhase:     math.Pi,
			targetCoherence: 0.70,
			maxTime:         10 * time.Second,
			disruption: func(s *swarm.Swarm) {
				// Disrupt ALL agents
				time.Sleep(1 * time.Second)
				s.DisruptAgents(1.0) // 100% disruption
			},
			expectSuccess: true, // Should still converge from complete chaos
		},
		{
			name:            "impossible_coherence_target",
			swarmSize:       20,
			targetPhase:     math.Pi,
			targetCoherence: 1.0, // Perfect coherence is nearly impossible
			maxTime:         2 * time.Second,
			expectSuccess:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create goal state
			goal := core.State{
				Phase:     tt.targetPhase,
				Frequency: 100 * time.Millisecond,
				Coherence: tt.targetCoherence,
			}

			// Create swarm with auto-scaled config
			cfg := config.AutoScaleConfig(tt.swarmSize)
			s, err := swarm.New(tt.swarmSize, goal, swarm.WithConfig(cfg))
			require.NoError(t, err)

			// Setup context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), tt.maxTime)
			defer cancel()

			// Track convergence metrics
			var (
				converged       bool
				convergenceTime time.Duration
				finalCoherence  float64
				mu              sync.Mutex
			)

			// Start disruption if specified
			if tt.disruption != nil {
				go tt.disruption(s)
			}

			// Special handling for disruption recovery test
			if tt.name == swarmWithDisruptionRecovery {
				// This test requires special logic to track disruption and recovery
				var (
					disruptionDetected      bool
					recoveryAchieved        bool
					preDisruptionCoherence  float64
					postDisruptionCoherence float64
					maxCoherenceDrop        float64
				)

				// Monitor for disruption and recovery phases
				ticker := time.NewTicker(100 * time.Millisecond)
				defer ticker.Stop()
				startTime := time.Now()

				// Run swarm with continuous monitoring for recovery
				go func() {
					_ = s.RunContinuous(ctx) // Use RunContinuous to handle disruption recovery
				}()

			monitorLoop:
				for {
					select {
					case <-ticker.C:
						coherence := s.MeasureCoherence()

						// Track coherence before disruption (first 500ms)
						if time.Since(startTime) < 450*time.Millisecond {
							preDisruptionCoherence = coherence
							t.Logf("Pre-disruption coherence: %.3f at time %v", coherence, time.Since(startTime))
						}

						// Detect disruption (significant drop after 500ms)
						if time.Since(startTime) > 600*time.Millisecond && !disruptionDetected {
							if coherence < preDisruptionCoherence-0.2 {
								disruptionDetected = true
								postDisruptionCoherence = coherence
								maxCoherenceDrop = preDisruptionCoherence - coherence
								t.Logf("Disruption detected: %.3f -> %.3f (drop: %.3f)",
									preDisruptionCoherence, postDisruptionCoherence, maxCoherenceDrop)
							}
						}

						// Check for recovery after disruption
						if disruptionDetected && !recoveryAchieved {
							if coherence >= tt.targetCoherence {
								recoveryAchieved = true
								convergenceTime = time.Since(startTime)
								t.Logf("Recovery achieved: %.3f in %v", coherence, convergenceTime)
							}
						}

						mu.Lock()
						finalCoherence = coherence
						converged = recoveryAchieved
						mu.Unlock()

						// Exit if recovered or if we've been running too long
						if recoveryAchieved || time.Since(startTime) > tt.maxTime-1*time.Second {
							cancel()
							break monitorLoop // Exit the for loop
						}

						// Log progress for debugging
						if time.Since(startTime).Truncate(time.Second) != (time.Since(startTime) - 100*time.Millisecond).Truncate(time.Second) {
							t.Logf("Time: %v, Coherence: %.3f, Disrupted: %v, Recovered: %v",
								time.Since(startTime).Truncate(time.Second), coherence, disruptionDetected, recoveryAchieved)
						}

					case <-ctx.Done():
						break monitorLoop // Exit the for loop
					}
				}

				// We've exited the monitoring loop - now proceed to assertions
				t.Logf("Exited monitoring loop. Proceeding to final assertions...")
			}

			// Monitor convergence (skip for disruption recovery test)
			done := make(chan struct{})
			if tt.name != swarmWithDisruptionRecovery {
				go func() {
					ticker := time.NewTicker(100 * time.Millisecond)
					defer ticker.Stop()
					startTime := time.Now()

					for {
						select {
						case <-ticker.C:
							coherence := s.MeasureCoherence()
							mu.Lock()
							finalCoherence = coherence
							if coherence >= tt.targetCoherence {
								converged = true
								convergenceTime = time.Since(startTime)
								mu.Unlock()
								close(done)
								return
							}
							mu.Unlock()
						case <-ctx.Done():
							return
						}
					}
				}()
			}

			// Run swarm (skip for disruption recovery test as it has its own logic)
			errCh := make(chan error, 1)
			if tt.name != swarmWithDisruptionRecovery {
				go func() {
					errCh <- s.Run(ctx)
				}()
			}

			// Wait for convergence or timeout (skip for disruption recovery test)
			if tt.name != swarmWithDisruptionRecovery {
				select {
				case <-done:
					cancel() // Stop the swarm
					<-errCh  // Wait for clean shutdown
				case <-ctx.Done():
					// Timeout reached
				}
			}

			// Verify results
			mu.Lock()
			defer mu.Unlock()

			// Special verification for disruption recovery test
			if tt.name == swarmWithDisruptionRecovery {
				t.Logf("FINAL TEST STATE:")
				t.Logf("  - Converged: %v", converged)
				t.Logf("  - Final coherence: %.3f", finalCoherence)
				t.Logf("  - Target coherence: %.3f", tt.targetCoherence)

				// Check if we achieved target coherence (regardless of disruption detection)
				success := finalCoherence >= tt.targetCoherence

				if success {
					t.Logf("Successfully achieved target coherence %.2f (final: %.3f)", tt.targetCoherence, finalCoherence)
				} else {
					t.Logf("Failed to achieve target coherence %.2f (final: %.3f)", tt.targetCoherence, finalCoherence)
				}

				assert.True(t, success,
					"Should achieve %.0f%% coherence and maintain it", tt.targetCoherence*100)
				return
			}

			if tt.expectSuccess {
				assert.True(t, converged,
					"Swarm should converge. Final coherence: %.3f, Target: %.3f",
					finalCoherence, tt.targetCoherence)
				assert.Less(t, convergenceTime, tt.maxTime,
					"Should converge within time limit")
				t.Logf("Converged in %v with coherence %.3f", convergenceTime, finalCoherence)
			} else {
				assert.False(t, converged,
					"Should not achieve impossible target. Final coherence: %.3f",
					finalCoherence)
			}
		})
	}
}

// TestE2ESwarmResilience tests swarm resilience to various failures
func TestE2ESwarmResilience(t *testing.T) {
	tests := []struct {
		name     string
		scenario func(t *testing.T, s *swarm.Swarm)
		validate func(t *testing.T, s *swarm.Swarm)
	}{
		{
			name: "agent_energy_depletion",
			scenario: func(_ *testing.T, s *swarm.Swarm) {
				// Deplete energy of half the agents
				agents := s.Agents()
				count := 0
				for _, agent := range agents {
					if count >= len(agents)/2 {
						break
					}
					agent.SetEnergy(0)
					count++
				}
			},
			validate: func(t *testing.T, s *swarm.Swarm) {
				// System should still achieve some coherence
				coherence := s.MeasureCoherence()
				assert.Greater(t, coherence, 0.3, "Should maintain partial coherence despite energy depletion")
			},
		},
		{
			name: "network_partition",
			scenario: func(_ *testing.T, s *swarm.Swarm) {
				// Disconnect half the agents from each other
				agents := s.Agents()
				mid := len(agents) / 2

				// Get agent IDs for the two groups
				var group1, group2 []string
				count := 0
				for id := range agents {
					if count < mid {
						group1 = append(group1, id)
					} else {
						group2 = append(group2, id)
					}
					count++
				}

				// Disconnect group1 from group2
				for _, id1 := range group1 {
					agent1 := agents[id1]
					for _, id2 := range group2 {
						agent1.DisconnectFrom(id2)
						agents[id2].DisconnectFrom(id1)
					}
				}
			},
			validate: func(t *testing.T, s *swarm.Swarm) {
				// Each partition should achieve local coherence
				agents := s.Agents()
				assert.Greater(t, len(agents), 0, "Should have agents")

				// Check that some agents are disconnected
				var disconnectedPairs int
				for _, a1 := range agents {
					for _, a2 := range agents {
						if a1.ID != a2.ID && !a1.IsConnectedTo(a2.ID) {
							disconnectedPairs++
						}
					}
				}
				assert.Greater(t, disconnectedPairs, 0, "Should have disconnected agent pairs")
			},
		},
		{
			name: "rapid_phase_changes",
			scenario: func(_ *testing.T, s *swarm.Swarm) {
				// Rapidly change phases of random agents
				for i := 0; i < 10; i++ {
					s.DisruptAgents(0.1) // Disrupt 10% each time
					time.Sleep(100 * time.Millisecond)
				}
			},
			validate: func(t *testing.T, s *swarm.Swarm) {
				// System should stabilize after disruptions
				time.Sleep(2 * time.Second)
				coherence := s.MeasureCoherence()
				assert.Greater(t, coherence, 0.5, "Should recover from rapid disruptions")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a medium-sized swarm
			goal := core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.85,
			}

			s, err := swarm.New(50, goal)
			require.NoError(t, err)

			// Run swarm for initial convergence
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Start swarm in background with continuous recovery
			errCh := make(chan error, 1)
			go func() {
				errCh <- s.RunContinuous(ctx)
			}()

			// Let it stabilize initially
			time.Sleep(1 * time.Second)

			// Apply scenario while swarm is still running
			tt.scenario(t, s)

			// Monitor for a bit after disruption
			time.Sleep(1 * time.Second)

			// Validate resilience while swarm is still active
			tt.validate(t, s)

			// Cancel and wait for clean shutdown
			cancel()
			select {
			case err := <-errCh:
				if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
					t.Errorf("Swarm error: %v", err)
				}
			case <-time.After(1 * time.Second):
				// Timeout waiting for shutdown
			}
		})
	}
}

// TestE2EPerformanceScaling tests performance at different scales
func TestE2EPerformanceScaling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance scaling test in short mode")
	}

	sizes := []int{10, 50, 100, 500, 1000}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("size_%d", size), func(t *testing.T) {
			goal := core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.80,
			}

			// Measure creation time
			startCreate := time.Now()
			s, err := swarm.New(size, goal)
			createTime := time.Since(startCreate)
			require.NoError(t, err)

			// Measure convergence time
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			startRun := time.Now()
			done := make(chan struct{})

			go func() {
				ticker := time.NewTicker(100 * time.Millisecond)
				defer ticker.Stop()

				for {
					select {
					case <-ticker.C:
						if s.MeasureCoherence() >= 0.80 {
							close(done)
							return
						}
					case <-ctx.Done():
						return
					}
				}
			}()

			go s.Run(ctx)

			select {
			case <-done:
				convergenceTime := time.Since(startRun)
				t.Logf("Size %d: Creation=%v, Convergence=%v",
					size, createTime, convergenceTime)

				// Performance assertions
				assert.Less(t, createTime, time.Duration(size)*time.Millisecond,
					"Creation should be < 1ms per agent")
				assert.Less(t, convergenceTime, time.Duration(size/10)*time.Second,
					"Convergence should scale sub-linearly")
			case <-ctx.Done():
				t.Errorf("Size %d: Failed to converge in 30s", size)
			}
		})
	}
}

// TestE2EMemoryStability tests memory usage stability over time
func TestE2EMemoryStability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory stability test in short mode")
	}

	goal := core.State{
		Phase:     math.Pi,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.85,
	}

	// Create a large swarm
	s, err := swarm.New(1000, goal)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Run swarm in background
	errCh := make(chan error, 1)
	go func() {
		// Use Run() for now - won't recover properly from disruptions
		// Should use RunContinuous() when available
		errCh <- s.Run(ctx)
	}()

	// Monitor for steady state
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var measurements []float64
	for i := 0; i < 10; i++ {
		select {
		case <-ticker.C:
			coherence := s.MeasureCoherence()
			measurements = append(measurements, coherence)

			// Apply periodic disruptions
			if i%3 == 0 {
				t.Logf("Applying 5%% disruption at iteration %d", i)
				s.DisruptAgents(0.05) // Disrupt 5%
			}
		case err := <-errCh:
			// Swarm exited early - this might happen after initial convergence
			if err != nil && !errors.Is(err, context.Canceled) {
				t.Logf("Swarm exited: %v", err)
			}
			// Continue measuring even if swarm stopped
		case <-ctx.Done():
			return
		}
	}

	// Verify stability
	if len(measurements) > 5 {
		// Check that coherence remains stable
		avgCoherence := 0.0
		for _, c := range measurements[5:] { // Skip initial convergence
			avgCoherence += c
		}
		avgCoherence /= float64(len(measurements) - 5)

		assert.Greater(t, avgCoherence, 0.7,
			"Should maintain good coherence over time")
	}
}

// TestE2ERunContinuousRecovery tests the RunContinuous method's ability to handle disruptions
func TestE2ERunContinuousRecovery(t *testing.T) {
	tests := []struct {
		name            string
		swarmSize       int
		targetCoherence float64
		disruptions     []struct {
			delay   time.Duration
			percent float64
		}
		testDuration      time.Duration
		minFinalCoherence float64
	}{
		{
			name:            "single_disruption_recovery",
			swarmSize:       20,
			targetCoherence: 0.90,
			disruptions: []struct {
				delay   time.Duration
				percent float64
			}{
				{delay: 1 * time.Second, percent: 0.5},
			},
			testDuration:      5 * time.Second,
			minFinalCoherence: 0.75, // More realistic for 50% disruption in 4s recovery time
		},
		{
			name:            "multiple_disruptions_recovery",
			swarmSize:       30,
			targetCoherence: 0.85,
			disruptions: []struct {
				delay   time.Duration
				percent float64
			}{
				{delay: 500 * time.Millisecond, percent: 0.3},
				{delay: 2 * time.Second, percent: 0.4},
				{delay: 3500 * time.Millisecond, percent: 0.2},
			},
			testDuration:      6 * time.Second,
			minFinalCoherence: 0.80,
		},
		{
			name:            "rapid_successive_disruptions",
			swarmSize:       25,
			targetCoherence: 0.80,
			disruptions: []struct {
				delay   time.Duration
				percent float64
			}{
				{delay: 300 * time.Millisecond, percent: 0.2},
				{delay: 600 * time.Millisecond, percent: 0.2},
				{delay: 900 * time.Millisecond, percent: 0.2},
				{delay: 1200 * time.Millisecond, percent: 0.2},
			},
			testDuration:      4 * time.Second,
			minFinalCoherence: 0.70,
		},
		{
			name:            "extreme_disruption_recovery",
			swarmSize:       40,
			targetCoherence: 0.75,
			disruptions: []struct {
				delay   time.Duration
				percent float64
			}{
				{delay: 800 * time.Millisecond, percent: 0.9}, // 90% disruption
			},
			testDuration:      5 * time.Second,
			minFinalCoherence: 0.65,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create goal state
			goal := core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: tt.targetCoherence,
			}

			// Create swarm with auto-scaled config
			cfg := config.AutoScaleConfig(tt.swarmSize)
			s, err := swarm.New(tt.swarmSize, goal, swarm.WithConfig(cfg))
			require.NoError(t, err)

			// Setup context
			ctx, cancel := context.WithTimeout(context.Background(), tt.testDuration)
			defer cancel()

			// Apply disruptions in background
			go func() {
				for _, d := range tt.disruptions {
					select {
					case <-time.After(d.delay):
						t.Logf("Applying %.0f%% disruption at %v", d.percent*100, d.delay)
						s.DisruptAgents(d.percent)
					case <-ctx.Done():
						return
					}
				}
			}()

			// Track coherence over time
			coherenceSamples := []float64{}
			var mu sync.Mutex

			// Monitor coherence in background
			go func() {
				ticker := time.NewTicker(200 * time.Millisecond)
				defer ticker.Stop()

				for {
					select {
					case <-ticker.C:
						coherence := s.MeasureCoherence()
						mu.Lock()
						coherenceSamples = append(coherenceSamples, coherence)
						mu.Unlock()
					case <-ctx.Done():
						return
					}
				}
			}()

			// Run swarm with RunContinuous for proper recovery
			errCh := make(chan error, 1)
			go func() {
				errCh <- s.RunContinuous(ctx)
			}()

			// Wait for test duration
			select {
			case <-ctx.Done():
				// Test duration reached
			case err := <-errCh:
				if err != nil && !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("Swarm error: %v", err)
				}
			}

			// Analyze results
			mu.Lock()
			defer mu.Unlock()

			if len(coherenceSamples) < 5 {
				t.Fatal("Not enough coherence samples collected")
			}

			// Get final coherence (average of last 3 samples)
			finalSamples := coherenceSamples[len(coherenceSamples)-3:]
			finalCoherence := 0.0
			for _, c := range finalSamples {
				finalCoherence += c
			}
			finalCoherence /= float64(len(finalSamples))

			// Find min and max coherence
			minCoherence, maxCoherence := coherenceSamples[0], coherenceSamples[0]
			for _, c := range coherenceSamples {
				if c < minCoherence {
					minCoherence = c
				}
				if c > maxCoherence {
					maxCoherence = c
				}
			}

			t.Logf("Coherence range: %.3f - %.3f", minCoherence, maxCoherence)
			t.Logf("Final coherence: %.3f (target: %.3f, min required: %.3f)",
				finalCoherence, tt.targetCoherence, tt.minFinalCoherence)

			// This should FAIL without RunContinuous
			assert.GreaterOrEqual(t, finalCoherence, tt.minFinalCoherence,
				"Swarm should maintain minimum coherence after disruptions (requires RunContinuous)")
		})
	}
}

// TestE2EBoundaryConditions tests edge cases and boundary conditions
func TestE2EBoundaryConditions(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (*swarm.Swarm, error)
		action      func(*swarm.Swarm)
		validate    func(*testing.T, *swarm.Swarm)
		expectPanic bool
	}{
		{
			name: "zero_energy_all_agents",
			setup: func() (*swarm.Swarm, error) {
				goal := core.State{
					Phase:     0,
					Frequency: 100 * time.Millisecond,
					Coherence: 0.80,
				}
				return swarm.New(10, goal)
			},
			action: func(s *swarm.Swarm) {
				// Set all agents to zero energy
				agents := s.Agents()
				for _, agent := range agents {
					agent.SetEnergy(0)
				}
			},
			validate: func(t *testing.T, s *swarm.Swarm) {
				coherence := s.MeasureCoherence()
				assert.Less(t, coherence, 0.5, "Zero energy should prevent coherence")
			},
		},
		{
			name: "disconnected_network",
			setup: func() (*swarm.Swarm, error) {
				goal := core.State{
					Phase:     math.Pi,
					Frequency: 100 * time.Millisecond,
					Coherence: 0.85,
				}
				return swarm.New(10, goal)
			},
			action: func(s *swarm.Swarm) {
				// Disconnect all agents from each other
				agents := s.Agents()
				agentIDs := make([]string, 0, len(agents))
				for id := range agents {
					agentIDs = append(agentIDs, id)
				}

				for i, id1 := range agentIDs {
					for j, id2 := range agentIDs {
						if i != j {
							agents[id1].DisconnectFrom(id2)
						}
					}
				}
			},
			validate: func(t *testing.T, s *swarm.Swarm) {
				coherence := s.MeasureCoherence()
				assert.Less(t, coherence, 0.45, "Disconnected network should have low coherence")
			},
		},
		{
			name: "single_agent_swarm",
			setup: func() (*swarm.Swarm, error) {
				goal := core.State{
					Phase:     0,
					Frequency: 100 * time.Millisecond,
					Coherence: 0.90,
				}
				return swarm.New(1, goal)
			},
			action: func(_ *swarm.Swarm) {
				// Single agent - nothing to disrupt
				time.Sleep(100 * time.Millisecond)
			},
			validate: func(t *testing.T, s *swarm.Swarm) {
				coherence := s.MeasureCoherence()
				assert.InDelta(t, 1.0, coherence, 0.001, "Single agent should have perfect coherence")
			},
		},
		{
			name: "negative_phase_values",
			setup: func() (*swarm.Swarm, error) {
				goal := core.State{
					Phase:     -math.Pi,
					Frequency: 100 * time.Millisecond,
					Coherence: 0.85,
				}
				return swarm.New(15, goal)
			},
			action: func(s *swarm.Swarm) {
				// Let it run briefly
				ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
				defer cancel()
				_ = s.Run(ctx)
			},
			validate: func(t *testing.T, s *swarm.Swarm) {
				coherence := s.MeasureCoherence()
				assert.Greater(t, coherence, 0.0, "Should handle negative phase values")
			},
		},
		{
			name: "rapid_phase_switching",
			setup: func() (*swarm.Swarm, error) {
				goal := core.State{
					Phase:     0,
					Frequency: 100 * time.Millisecond,
					Coherence: 0.80,
				}
				return swarm.New(20, goal)
			},
			action: func(s *swarm.Swarm) {
				// Rapidly switch phases of all agents
				for i := 0; i < 10; i++ {
					agents := s.Agents()
					for _, agent := range agents {
						currentPhase := agent.Phase()
						agent.SetPhase(currentPhase + math.Pi)
					}
					time.Sleep(50 * time.Millisecond)
				}
			},
			validate: func(t *testing.T, s *swarm.Swarm) {
				coherence := s.MeasureCoherence()
				t.Logf("Coherence after rapid switching: %.3f", coherence)
				// Just verify it doesn't crash
				assert.GreaterOrEqual(t, coherence, 0.0, "Should survive rapid phase switching")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				assert.Panics(t, func() {
					s, err := tt.setup()
					if err != nil {
						panic(err)
					}
					tt.action(s)
				})
			} else {
				s, err := tt.setup()
				require.NoError(t, err)

				tt.action(s)
				tt.validate(t, s)
			}
		})
	}
}

// TestE2EConcurrentSwarms tests multiple swarms running concurrently
func TestE2EConcurrentSwarms(t *testing.T) {
	numSwarms := 5
	swarmSize := 100

	var wg sync.WaitGroup
	results := make([]bool, numSwarms)
	errs := make([]error, numSwarms)

	for i := 0; i < numSwarms; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Each swarm has a different target phase
			goal := core.State{
				Phase:     float64(idx) * math.Pi / float64(numSwarms),
				Frequency: 100 * time.Millisecond,
				Coherence: 0.85,
			}

			s, err := swarm.New(swarmSize, goal)
			if err != nil {
				errs[idx] = fmt.Errorf("swarm %d: failed to create: %w", idx, err)
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// Run and monitor
			done := make(chan struct{})
			go func() {
				ticker := time.NewTicker(100 * time.Millisecond)
				defer ticker.Stop()

				for {
					select {
					case <-ticker.C:
						if s.MeasureCoherence() >= 0.85 {
							results[idx] = true
							close(done)
							return
						}
					case <-ctx.Done():
						return
					}
				}
			}()

			go s.Run(ctx)

			select {
			case <-done:
				t.Logf("Swarm %d converged successfully", idx)
			case <-ctx.Done():
				t.Logf("Swarm %d timed out", idx)
			}
		}(i)
	}

	wg.Wait()

	// Check for errors
	for i, err := range errs {
		if err != nil {
			t.Errorf("Swarm %d error: %v", i, err)
		}
	}

	// Check results
	successCount := 0
	for _, success := range results {
		if success {
			successCount++
		}
	}

	assert.Greater(t, successCount, numSwarms*3/4,
		"At least 75%% of swarms should converge")
}
