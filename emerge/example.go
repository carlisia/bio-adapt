package emerge

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// Example demonstrates distributed biological synchronization.
func Example() error {
	// Create target pattern
	goal := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	// Create swarm of autonomous agents
	swarm, err := NewSwarm(100, goal) // Smaller for simplicity
	if err != nil {
		return fmt.Errorf("failed to create swarm: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Printf("Initial coherence: %.3f\n", swarm.MeasureCoherence())

	// Run distributed synchronization - no central control
	start := time.Now()

	// Simulate disruption after 1 second
	go func() {
		time.Sleep(1 * time.Second)

		// Disrupt 10% of agents
		swarm.DisruptAgents(0.1)

		fmt.Println("Disrupted 10 agents - watch autonomous recovery")
	}()

	// Let the swarm run autonomously
	go func() {
		if err := swarm.Run(ctx); err != nil {
			fmt.Printf("Error running swarm: %v\n", err)
		}
	}()

	// Observe emergent synchronization
	time.Sleep(3 * time.Second)

	finalCoherence := swarm.MeasureCoherence()
	fmt.Printf("Final coherence: %.3f in %v\n", finalCoherence, time.Since(start))

	// Check energy levels (metabolic state)
	var totalEnergy float64
	var count int
	swarm.agents.Range(func(key, value any) bool {
		agent := value.(*Agent)
		totalEnergy += agent.Energy()
		count++
		return true
	})

	fmt.Printf("Average energy: %.1f/100\n", totalEnergy/float64(count))
	return nil
}

// DemonstrateAutonomy shows how agents make independent decisions.
// Note: In this demonstration, we intentionally ignore ApplyAction return values
// because we're showing autonomous behavior where actions may fail due to
// stubbornness or energy constraints - this is expected and part of the demo.
func DemonstrateAutonomy() {
	// Create two agents with different preferences
	agent1 := NewAgent("stubborn")
	agent1.stubbornness.Store(0.8) // Very stubborn
	agent1.SetLocalGoal(0)         // Wants phase 0

	agent2 := NewAgent("flexible")
	agent2.stubbornness.Store(0.1) // Flexible
	agent2.SetLocalGoal(3.14159)   // Wants phase π

	// Connect them as neighbors
	agent1.neighbors.Store(agent2.ID, agent2)
	agent2.neighbors.Store(agent1.ID, agent1)

	// Global goal
	globalGoal := State{
		Phase:     1.57, // π/2
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	fmt.Println("Agent autonomy demonstration:")
	fmt.Printf("Stubborn agent (wants 0): phase=%.2f\n", agent1.Phase())
	fmt.Printf("Flexible agent (wants π): phase=%.2f\n", agent2.Phase())

	// Let them negotiate
	for range 20 {
		action1, accepted1 := agent1.ProposeAdjustment(globalGoal)
		if accepted1 {
			// Apply action - success and energy cost are not critical for this demo
			_, _ = agent1.ApplyAction(action1)
		}

		action2, accepted2 := agent2.ProposeAdjustment(globalGoal)
		if accepted2 {
			// Apply action - success and energy cost are not critical for this demo
			_, _ = agent2.ApplyAction(action2)
		}
	}

	fmt.Println("\nAfter negotiation:")
	fmt.Printf("Stubborn agent: phase=%.2f (resisted global goal)\n", agent1.Phase())
	fmt.Printf("Flexible agent: phase=%.2f (moved toward goal)\n", agent2.Phase())
}

// Benchmark measures convergence time for different swarm sizes.
func Benchmark() error {
	sizes := []int{10, 50, 100, 500}

	for _, size := range sizes {
		goal := State{
			Phase:     0,
			Frequency: 100 * time.Millisecond,
			Coherence: 0.8,
		}

		swarm, err := NewSwarm(size, goal)
		if err != nil {
			return fmt.Errorf("failed to create swarm of size %d: %w", size, err)
		}

		// Random initial phases
		swarm.agents.Range(func(key, value any) bool {
			agent := value.(*Agent)
			agent.SetPhase(rand.Float64() * 2 * 3.14159)
			return true
		})

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		start := time.Now()
		initialCoherence := swarm.MeasureCoherence()

		// Run until convergence
		go func() {
			if err := swarm.Run(ctx); err != nil {
				fmt.Printf("Error running swarm %d: %v\n", size, err)
			}
		}()

		// Monitor convergence
		converged := false
		for !converged {
			time.Sleep(100 * time.Millisecond)
			coherence := swarm.MeasureCoherence()
			if coherence > goal.Coherence {
				converged = true
			}
			if time.Since(start) > 5*time.Second {
				break
			}
		}

		convergenceTime := time.Since(start)
		finalCoherence := swarm.MeasureCoherence()

		fmt.Printf("Size=%d: %.3f -> %.3f in %v\n",
			size, initialCoherence, finalCoherence, convergenceTime)

		cancel()
	}
	return nil
}
