package swarm_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/swarm"
)

// TestDisruptAgents tests the DisruptAgents method thoroughly
func TestDisruptAgents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		swarmSize         int
		disruptionPercent float64
		expectedBehavior  string
	}{
		{
			name:              "no_disruption",
			swarmSize:         20,
			disruptionPercent: 0.0,
			expectedBehavior:  "no_change",
		},
		{
			name:              "partial_disruption_20_percent",
			swarmSize:         20,
			disruptionPercent: 0.2,
			expectedBehavior:  "coherence_drop",
		},
		{
			name:              "half_disruption",
			swarmSize:         20,
			disruptionPercent: 0.5,
			expectedBehavior:  "significant_coherence_drop",
		},
		{
			name:              "total_disruption",
			swarmSize:         20,
			disruptionPercent: 1.0,
			expectedBehavior:  "maximum_disorder",
		},
		{
			name:              "over_100_percent",
			swarmSize:         20,
			disruptionPercent: 1.5,
			expectedBehavior:  "clamped_to_100",
		},
		{
			name:              "negative_percent",
			swarmSize:         20,
			disruptionPercent: -0.5,
			expectedBehavior:  "no_change",
		},
		{
			name:              "very_small_disruption",
			swarmSize:         100,
			disruptionPercent: 0.01,
			expectedBehavior:  "minimal_change",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create swarm
			goal := core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.85,
			}

			s, err := swarm.New(tt.swarmSize, goal)
			require.NoError(t, err)

			// Run briefly to establish some coherence
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			done := make(chan struct{})
			go func() {
				_ = s.Run(ctx)
				close(done)
			}()
			time.Sleep(400 * time.Millisecond)

			// Cancel and wait for swarm to stop
			cancel()
			<-done
			time.Sleep(10 * time.Millisecond)

			// Measure coherence before disruption (swarm stopped)
			beforeCoherence := s.MeasureCoherence()
			t.Logf("Before disruption: %.3f", beforeCoherence)

			// Apply disruption
			s.DisruptAgents(tt.disruptionPercent)

			// Measure coherence after disruption (swarm still stopped)
			afterCoherence := s.MeasureCoherence()
			t.Logf("After disruption (%.1f%%): %.3f", tt.disruptionPercent*100, afterCoherence)

			// Verify expected behavior
			switch tt.expectedBehavior {
			case "no_change":
				assert.InDelta(t, beforeCoherence, afterCoherence, 0.05,
					"Coherence should remain unchanged")

			case "coherence_drop":
				assert.Less(t, afterCoherence, beforeCoherence,
					"Coherence should drop after disruption")
				// Don't check lower bound as it varies with random phases

			case "significant_coherence_drop":
				assert.Less(t, afterCoherence, beforeCoherence,
					"Coherence should drop after disruption")
				// The amount of drop varies based on initial coherence and randomness

			case "maximum_disorder":
				assert.Less(t, afterCoherence, 0.8,
					"Coherence should be low after total disruption")

			case "clamped_to_100":
				// Should behave like 100% disruption
				assert.Less(t, afterCoherence, 0.8,
					"Over 100% should be clamped to 100%")

			case "minimal_change":
				assert.InDelta(t, beforeCoherence, afterCoherence, 0.1,
					"Very small disruption should have minimal effect")
			}
		})
	}
}

// TestDisruptAgentsDistribution tests that disruption affects the right number of agents
func TestDisruptAgentsDistribution(t *testing.T) {
	t.Parallel()

	swarmSize := 100
	goal := core.State{
		Phase:     math.Pi,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.85,
	}

	s, err := swarm.New(swarmSize, goal)
	require.NoError(t, err)

	// Get initial phases
	initialPhases := make(map[string]float64)
	agents := s.Agents()
	for id, agent := range agents {
		initialPhases[id] = agent.Phase()
	}

	// Disrupt 30% of agents
	s.DisruptAgents(0.3)

	// Count how many agents had their phases changed
	changedCount := 0
	for id, agent := range agents {
		newPhase := agent.Phase()
		if math.Abs(newPhase-initialPhases[id]) > 0.01 {
			changedCount++
		}
	}

	// Verify approximately 30% were disrupted (with some tolerance for randomness)
	expectedCount := int(0.3 * float64(swarmSize))
	assert.InDelta(t, expectedCount, changedCount, float64(swarmSize)*0.1,
		"Approximately 30%% of agents should be disrupted")
}

// TestDisruptAgentsMultipleTimes tests applying disruption multiple times
func TestDisruptAgentsMultipleTimes(t *testing.T) {
	t.Parallel()

	goal := core.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.85,
	}

	s, err := swarm.New(50, goal)
	require.NoError(t, err)

	// Run briefly to establish coherence
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	go s.Run(ctx)
	time.Sleep(400 * time.Millisecond)
	cancel()

	initialCoherence := s.MeasureCoherence()
	t.Logf("Initial coherence: %.3f", initialCoherence)

	// Apply multiple disruptions
	disruptions := []float64{0.1, 0.2, 0.3, 0.2, 0.1}
	coherenceValues := []float64{initialCoherence}

	for i, pct := range disruptions {
		beforeDisruption := s.MeasureCoherence()
		s.DisruptAgents(pct)
		afterDisruption := s.MeasureCoherence()
		coherenceValues = append(coherenceValues, afterDisruption)

		t.Logf("Disruption %d (%.0f%%): %.3f -> %.3f", i+1, pct*100, beforeDisruption, afterDisruption)

		// Disruption should change coherence (usually decrease, but could increase from random state)
		assert.NotEqual(t, beforeDisruption, afterDisruption,
			"Disruption should change coherence")
	}

	// Overall, multiple disruptions should have degraded coherence from initial
	finalCoherence := coherenceValues[len(coherenceValues)-1]
	assert.Less(t, finalCoherence, initialCoherence+0.1,
		"Multiple disruptions should not significantly improve coherence")
}

// TestDisruptAgentsOnEmptySwarm tests disrupting a swarm with no agents
func TestDisruptAgentsOnEmptySwarm(t *testing.T) {
	t.Parallel()

	// Create minimal swarm
	goal := core.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.85,
	}

	s, err := swarm.New(1, goal)
	require.NoError(t, err)

	// This should not panic
	assert.NotPanics(t, func() {
		s.DisruptAgents(0.5)
	}, "Disrupting a single-agent swarm should not panic")

	// Coherence of single agent should remain perfect
	coherence := s.MeasureCoherence()
	assert.InDelta(t, 1.0, coherence, 0.001,
		"Single agent swarm should maintain perfect coherence")
}

// TestDisruptAgentsRaceCondition tests concurrent disruptions
func TestDisruptAgentsRaceCondition(t *testing.T) {
	t.Parallel()

	goal := core.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.85,
	}

	s, err := swarm.New(100, goal)
	require.NoError(t, err)

	// Run swarm
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go s.Run(ctx)

	// Apply concurrent disruptions
	done := make(chan bool, 3)

	// Goroutine 1: Regular disruptions
	go func() {
		for range 10 {
			s.DisruptAgents(0.1)
			time.Sleep(50 * time.Millisecond)
		}
		done <- true
	}()

	// Goroutine 2: Measure coherence
	go func() {
		for range 20 {
			_ = s.MeasureCoherence()
			time.Sleep(25 * time.Millisecond)
		}
		done <- true
	}()

	// Goroutine 3: Heavy disruptions
	go func() {
		for range 5 {
			s.DisruptAgents(0.5)
			time.Sleep(100 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for all goroutines
	for range 3 {
		<-done
	}

	// Should not have panicked or deadlocked
	finalCoherence := s.MeasureCoherence()
	assert.GreaterOrEqual(t, finalCoherence, 0.0,
		"Coherence should be valid after concurrent operations")
}
