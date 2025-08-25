package simulation

import (
	"context"
	"testing"
	"time"

	"github.com/carlisia/bio-adapt/emerge/goal"
	"github.com/carlisia/bio-adapt/emerge/scale"
)

func TestSimulation(t *testing.T) {
	t.Parallel()
	// Create simulation using configuration
	buildConfig := BuildConfig{
		Goal:            goal.MinimizeAPICalls,
		Scale:           scale.Tiny,
		TargetCoherence: 0.75,
	}
	sim, err := New(buildConfig)
	if err != nil {
		t.Fatalf("Failed to create simulation: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Start simulation
	go func() {
		_ = sim.Start(ctx) // Error handling not needed in test goroutine
	}()

	// Let it run briefly
	time.Sleep(500 * time.Millisecond)

	// Get snapshot
	snapshot := sim.Snapshot()

	// Verify we have agents (default for Tiny scale is 20)
	if len(snapshot.Agents) != 20 {
		t.Errorf("Expected %d agents, got %d", 20, len(snapshot.Agents))
	}

	// Verify coherence is being measured
	if snapshot.Coherence < 0 || snapshot.Coherence > 1 {
		t.Errorf("Invalid coherence: %f", snapshot.Coherence)
	}

	t.Logf("Simulation running with %d agents, coherence: %.2f%%",
		len(snapshot.Agents), snapshot.Coherence*100)
}
