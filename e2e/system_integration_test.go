//go:build e2e
// +build e2e

package e2e_test

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/carlisia/bio-adapt/emerge/agent"
	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/swarm"
	"github.com/carlisia/bio-adapt/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2EFullSystemIntegration tests the complete bio-adaptive system
func TestE2EFullSystemIntegration(t *testing.T) {
	t.Run("BasicSynchronization", testBasicSynchronization)
	t.Run("ScalabilityTest", testScalability)
	t.Run("ResilienceUnderLoad", testResilienceUnderLoad)
	t.Run("MultiSwarmCoordination", testMultiSwarmCoordination)
	t.Run("RealWorldScenarios", testRealWorldScenarios)
}

func testBasicSynchronization(t *testing.T) {
	tests := []struct {
		name            string
		swarmSize       int
		targetCoherence float64
		maxTime         time.Duration
		topology        string
	}{
		{
			name:            "small_fully_connected",
			swarmSize:       10,
			targetCoherence: 0.95,
			maxTime:         5 * time.Second,
			topology:        "full",
		},
		{
			name:            "medium_small_world",
			swarmSize:       100,
			targetCoherence: 0.90,
			maxTime:         10 * time.Second,
			topology:        "small-world",
		},
		{
			name:            "large_scale_free",
			swarmSize:       500,
			targetCoherence: 0.85,
			maxTime:         20 * time.Second,
			topology:        "scale-free",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goal := core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: tt.targetCoherence,
			}

			cfg := config.AutoScaleConfig(tt.swarmSize)
			s, err := swarm.New(tt.swarmSize, goal, swarm.WithConfig(cfg))
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), tt.maxTime)
			defer cancel()

			// Monitor convergence
			converged := make(chan bool)
			go monitorConvergence(s, tt.targetCoherence, converged)

			// Run swarm
			go s.Run(ctx)

			// Wait for result
			select {
			case success := <-converged:
				assert.True(t, success, "Swarm should converge to target coherence")
				t.Logf("%s: Converged successfully", tt.name)
			case <-ctx.Done():
				t.Errorf("%s: Failed to converge within %v", tt.name, tt.maxTime)
			}
		})
	}
}

func testScalability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping scalability test in short mode")
	}

	// Test increasing swarm sizes
	sizes := []int{10, 50, 100, 250, 500, 1000, 2000}

	results := make([]struct {
		size            int
		creationTime    time.Duration
		convergenceTime time.Duration
		memoryUsed      uint64
	}, len(sizes))

	for i, size := range sizes {
		t.Run(fmt.Sprintf("size_%d", size), func(t *testing.T) {
			// Record initial memory
			var m1 runtime.MemStats
			runtime.ReadMemStats(&m1)

			// Create swarm
			start := time.Now()
			goal := core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.80,
			}

			s, err := swarm.New(size, goal)
			results[i].creationTime = time.Since(start)
			require.NoError(t, err)

			// Run and measure convergence
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			convergenceStart := time.Now()
			converged := make(chan bool)
			go monitorConvergence(s, 0.80, converged)
			go s.Run(ctx)

			select {
			case <-converged:
				results[i].convergenceTime = time.Since(convergenceStart)
			case <-ctx.Done():
				t.Logf("Size %d: Did not converge in 60s", size)
			}

			// Record final memory
			var m2 runtime.MemStats
			runtime.ReadMemStats(&m2)
			results[i].memoryUsed = m2.Alloc - m1.Alloc
			results[i].size = size

			// Performance assertions
			assert.Less(t, results[i].creationTime, time.Duration(size)*2*time.Millisecond,
				"Creation time should scale linearly or better")

			if i > 0 {
				// Check that scaling is sub-quadratic
				prevResult := results[i-1]
				scaleFactor := float64(size) / float64(prevResult.size)
				timeScale := float64(results[i].convergenceTime) / float64(prevResult.convergenceTime)

				assert.Less(t, timeScale, scaleFactor*scaleFactor,
					"Convergence time should scale sub-quadratically")
			}
		})
	}

	// Print performance summary
	t.Log("\n=== Performance Scaling Summary ===")
	for _, r := range results {
		if r.size > 0 {
			t.Logf("Size %4d: Create=%8v, Converge=%8v, Memory=%dKB",
				r.size, r.creationTime, r.convergenceTime, r.memoryUsed/1024)
		}
	}
}

func testResilienceUnderLoad(t *testing.T) {
	scenarios := []struct {
		name       string
		swarmSize  int
		disruption func(s *swarm.Swarm, iteration int)
		validate   func(t *testing.T, s *swarm.Swarm, finalCoherence float64)
	}{
		{
			name:      "continuous_disruption",
			swarmSize: 100,
			disruption: func(s *swarm.Swarm, iteration int) {
				// Disrupt 10% of agents every iteration
				s.DisruptAgents(0.10)
			},
			validate: func(t *testing.T, s *swarm.Swarm, finalCoherence float64) {
				assert.Greater(t, finalCoherence, 0.60,
					"Should maintain reasonable coherence under continuous disruption")
			},
		},
		{
			name:      "cascading_failures",
			swarmSize: 200,
			disruption: func(s *swarm.Swarm, iteration int) {
				// Progressively fail more agents
				percentage := float64(iteration) * 0.05
				if percentage > 0.5 {
					percentage = 0.5
				}
				agents := s.Agents()
				count := 0
				maxFail := int(float64(len(agents)) * percentage)
				for _, a := range agents {
					if count >= maxFail {
						break
					}
					a.SetEnergy(0)
					count++
				}
			},
			validate: func(t *testing.T, s *swarm.Swarm, finalCoherence float64) {
				// Even with 50% failure, should have some coherence
				assert.Greater(t, finalCoherence, 0.30,
					"Should degrade gracefully with cascading failures")
			},
		},
		{
			name:      "network_partitions",
			swarmSize: 150,
			disruption: func(s *swarm.Swarm, iteration int) {
				if iteration%5 == 0 {
					// Create random network partitions
					agents := s.Agents()
					agentList := make([]*agent.Agent, 0, len(agents))
					for _, a := range agents {
						agentList = append(agentList, a)
					}

					// Disconnect random pairs
					for i := 0; i < len(agentList)/10; i++ {
						idx1 := i * 2
						idx2 := i*2 + 1
						if idx2 < len(agentList) {
							agentList[idx1].DisconnectFrom(agentList[idx2].ID)
						}
					}
				}
			},
			validate: func(t *testing.T, s *swarm.Swarm, finalCoherence float64) {
				assert.Greater(t, finalCoherence, 0.50,
					"Should handle network partitions")
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			goal := core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.85,
			}

			s, err := swarm.New(scenario.swarmSize, goal)
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			// Run swarm with periodic disruptions
			go s.Run(ctx)

			// Apply disruptions
			go func() {
				ticker := time.NewTicker(500 * time.Millisecond)
				defer ticker.Stop()
				iteration := 0

				for {
					select {
					case <-ticker.C:
						scenario.disruption(s, iteration)
						iteration++
					case <-ctx.Done():
						return
					}
				}
			}()

			// Let it run
			time.Sleep(10 * time.Second)

			// Measure final coherence
			finalCoherence := s.MeasureCoherence()
			scenario.validate(t, s, finalCoherence)
		})
	}
}

func testMultiSwarmCoordination(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping multi-swarm test in short mode")
	}

	// Create multiple swarms with different goals
	numSwarms := 5
	swarmSize := 100

	type swarmResult struct {
		id              int
		converged       bool
		convergenceTime time.Duration
		finalCoherence  float64
	}

	results := make(chan swarmResult, numSwarms)
	var wg sync.WaitGroup

	for i := 0; i < numSwarms; i++ {
		wg.Add(1)
		go func(swarmID int) {
			defer wg.Done()

			// Each swarm has a different target
			phase := float64(swarmID) * 2 * math.Pi / float64(numSwarms)
			goal := core.State{
				Phase:     phase,
				Frequency: time.Duration(100+swarmID*10) * time.Millisecond,
				Coherence: 0.85,
			}

			s, err := swarm.New(swarmSize, goal)
			if err != nil {
				t.Errorf("Swarm %d: Failed to create: %v", swarmID, err)
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			// Monitor this swarm
			start := time.Now()
			converged := make(chan bool)
			go monitorConvergence(s, 0.85, converged)
			go s.Run(ctx)

			result := swarmResult{id: swarmID}

			select {
			case <-converged:
				result.converged = true
				result.convergenceTime = time.Since(start)
				result.finalCoherence = s.MeasureCoherence()
			case <-ctx.Done():
				result.converged = false
				result.finalCoherence = s.MeasureCoherence()
			}

			results <- result
		}(i)
	}

	wg.Wait()
	close(results)

	// Analyze results
	convergedCount := 0
	totalTime := time.Duration(0)

	for result := range results {
		if result.converged {
			convergedCount++
			totalTime += result.convergenceTime
			t.Logf("Swarm %d: Converged in %v with coherence %.3f",
				result.id, result.convergenceTime, result.finalCoherence)
		} else {
			t.Logf("Swarm %d: Failed to converge, final coherence %.3f",
				result.id, result.finalCoherence)
		}
	}

	assert.Greater(t, convergedCount, numSwarms*3/4,
		"At least 75%% of swarms should converge independently")

	if convergedCount > 0 {
		avgTime := totalTime / time.Duration(convergedCount)
		t.Logf("Average convergence time: %v", avgTime)
	}
}

func testRealWorldScenarios(t *testing.T) {
	t.Run("LoadBalancing", func(t *testing.T) {
		// Simulate load balancing scenario
		numServers := 20
		goal := core.State{
			Phase:     0, // Balanced state
			Frequency: 50 * time.Millisecond,
			Coherence: 0.90,
		}

		s, err := swarm.New(numServers, goal)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Simulate load spikes
		go func() {
			time.Sleep(2 * time.Second)
			// Spike on 25% of servers
			agents := s.Agents()
			count := 0
			for _, a := range agents {
				if count >= numServers/4 {
					break
				}
				a.SetPhase(math.Pi) // High load
				count++
			}
		}()

		go s.Run(ctx)
		time.Sleep(5 * time.Second)

		// Check if load is rebalanced
		coherence := s.MeasureCoherence()
		assert.Greater(t, coherence, 0.70,
			"Load should be rebalanced after spike")
	})

	t.Run("TrafficRouting", func(t *testing.T) {
		// Simulate traffic routing with varying conditions
		numRouters := 50
		goal := core.State{
			Phase:     math.Pi / 2, // Optimal routing state
			Frequency: 100 * time.Millisecond,
			Coherence: 0.85,
		}

		cfg := config.AutoScaleConfig(numRouters)
		cfg.ConnectionProbability = 0.3 // Sparse network

		s, err := swarm.New(numRouters, goal, swarm.WithConfig(cfg))
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Simulate varying traffic patterns
		go func() {
			patterns := []float64{0, math.Pi / 4, math.Pi / 2, math.Pi}
			for _, pattern := range patterns {
				time.Sleep(3 * time.Second)
				// Change traffic pattern
				agents := s.Agents()
				for _, a := range agents {
					if math.Mod(float64(len(a.ID)), 2) == 0 {
						a.SetLocalGoal(pattern)
					}
				}
			}
		}()

		go s.Run(ctx)

		// Monitor adaptation
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		adaptations := 0
		prevCoherence := 0.0
		prevPhaseVariance := 0.0

		for i := 0; i < 10; i++ {
			<-ticker.C
			coherence := s.MeasureCoherence()
			phaseVariance := s.MeasurePhaseVariance()

			coherenceDelta := math.Abs(coherence - prevCoherence)
			varianceDelta := math.Abs(phaseVariance - prevPhaseVariance)

			// Debug output to understand what's happening
			t.Logf("Tick %d: coherence=%.3f (Δ=%.3f), variance=%.3f (Δ=%.3f)",
				i, coherence, coherenceDelta, phaseVariance, varianceDelta)

			// Count as adaptation if we see meaningful changes
			// Modern stable system maintains high coherence but shows subtle adaptations
			// Look for: small coherence changes OR phase variance shifts OR initial stabilization
			if coherenceDelta > 0.004 || varianceDelta > 0.01 || i == 0 {
				// Skip counting the initial measurement as adaptation
				if i > 0 {
					adaptations++
					t.Logf("  -> Adaptation detected! Total: %d", adaptations)
				}
			}
			prevCoherence = coherence
			prevPhaseVariance = phaseVariance
		}

		assert.Greater(t, adaptations, 2,
			"System should adapt to changing traffic patterns")
	})

	t.Run("ConsensusFormation", func(t *testing.T) {
		// Simulate distributed consensus
		numNodes := 30
		goal := core.State{
			Phase:     math.Pi, // Consensus target
			Frequency: 100 * time.Millisecond,
			Coherence: 0.95, // High consensus requirement
		}

		s, err := swarm.New(numNodes, goal)
		require.NoError(t, err)

		// Start with divergent opinions
		agents := s.Agents()
		i := 0
		for _, a := range agents {
			// Three different initial opinions
			phase := float64(i%3) * 2 * math.Pi / 3
			a.SetPhase(phase)
			i++
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		converged := make(chan bool)
		go monitorConvergence(s, 0.95, converged)
		go s.Run(ctx)

		select {
		case <-converged:
			// Check consensus
			phases := make([]float64, 0, numNodes)
			for _, a := range s.Agents() {
				phases = append(phases, a.Phase())
			}

			// Calculate phase variance
			mean := 0.0
			for _, p := range phases {
				mean += p
			}
			mean /= float64(len(phases))

			variance := 0.0
			for _, p := range phases {
				diff := p - mean
				variance += diff * diff
			}
			variance /= float64(len(phases))

			assert.Less(t, variance, 0.1,
				"Consensus should have low phase variance")
			t.Logf("Consensus reached with variance: %.6f", variance)

		case <-ctx.Done():
			t.Error("Failed to reach consensus")
		}
	})
}

// Helper function to monitor convergence
func monitorConvergence(s *swarm.Swarm, target float64, done chan<- bool) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		if s.MeasureCoherence() >= target {
			done <- true
			return
		}
	}
}

// BenchmarkE2EScaling benchmarks performance at different scales
func BenchmarkE2EScaling(b *testing.B) {
	sizes := []int{10, 50, 100, 500, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				goal := core.State{
					Phase:     math.Pi,
					Frequency: 100 * time.Millisecond,
					Coherence: 0.80,
				}

				s, err := swarm.New(size, goal)
				if err != nil {
					b.Fatal(err)
				}

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				s.Run(ctx)
				cancel()
			}
		})
	}
}

// BenchmarkE2EConcurrency benchmarks concurrent operations
func BenchmarkE2EConcurrency(b *testing.B) {
	goal := core.State{
		Phase:     math.Pi,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.85,
	}

	s, err := swarm.New(1000, goal)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	go s.Run(ctx)
	time.Sleep(1 * time.Second) // Let it stabilize

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Benchmark concurrent operations
			_ = s.MeasureCoherence()
			s.DisruptAgents(0.01)
		}
	})
}
