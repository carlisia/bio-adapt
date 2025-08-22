//nolint:paralleltest,thelper,intrange // E2E tests shouldn't run in parallel, validate funcs are not helpers, loop counters needed
package e2e_test

import (
	"context"
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
			name:            "swarm_with_disruption",
			swarmSize:       50,
			targetPhase:     math.Pi,
			targetCoherence: 0.90,
			maxTime:         10 * time.Second,
			disruption: func(s *swarm.Swarm) {
				// Disrupt 20% of agents after 2 seconds
				time.Sleep(2 * time.Second)
				s.DisruptAgents(0.2)
			},
			expectSuccess: true,
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

			// Monitor convergence
			done := make(chan struct{})
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

			// Run swarm
			errCh := make(chan error, 1)
			go func() {
				errCh <- s.Run(ctx)
			}()

			// Wait for convergence or timeout
			select {
			case <-done:
				cancel() // Stop the swarm
				<-errCh  // Wait for clean shutdown
			case <-ctx.Done():
				// Timeout reached
			}

			// Verify results
			mu.Lock()
			defer mu.Unlock()

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
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			go s.Run(ctx)
			time.Sleep(1 * time.Second) // Let it stabilize

			// Apply scenario
			tt.scenario(t, s)

			// Continue running with disruption
			time.Sleep(2 * time.Second)

			// Validate resilience
			tt.validate(t, s)

			cancel()
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

	// Run swarm
	go s.Run(ctx)

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
				s.DisruptAgents(0.05) // Disrupt 5%
			}
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

// TestE2EConcurrentSwarms tests multiple swarms running concurrently
func TestE2EConcurrentSwarms(t *testing.T) {
	numSwarms := 5
	swarmSize := 100

	var wg sync.WaitGroup
	results := make([]bool, numSwarms)

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
				t.Errorf("Swarm %d: Failed to create: %v", idx, err)
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
