package biofield

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// Example demonstrates distributed biological synchronization.
func Example() {
	// Create target pattern
	goal := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	// Create swarm of autonomous agents
	swarm := NewSwarm(100, goal) // Smaller for simplicity

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
	go swarm.Run(ctx)

	// Observe emergent synchronization
	time.Sleep(3 * time.Second)

	finalCoherence := swarm.MeasureCoherence()
	fmt.Printf("Final coherence: %.3f in %v\n", finalCoherence, time.Since(start))

	// Check energy levels (metabolic state)
	var totalEnergy float64
	var count int
	swarm.agents.Range(func(key, value any) bool {
		agent := value.(*Agent)
		totalEnergy += agent.GetEnergy()
		count++
		return true
	})

	fmt.Printf("Average energy: %.1f/100\n", totalEnergy/float64(count))
}

// DemonstrateAutonomy shows how agents make independent decisions.
func DemonstrateAutonomy() {
	// Create two agents with different preferences
	agent1 := NewAgent("stubborn")
	agent1.stubbornness.Store(0.8) // Very stubborn
	agent1.LocalGoal.Store(0)      // Wants phase 0

	agent2 := NewAgent("flexible")
	agent2.stubbornness.Store(0.1)  // Flexible
	agent2.LocalGoal.Store(3.14159) // Wants phase π

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
	fmt.Printf("Stubborn agent (wants 0): phase=%.2f\n", agent1.GetPhase())
	fmt.Printf("Flexible agent (wants π): phase=%.2f\n", agent2.GetPhase())

	// Let them negotiate
	for range 20 {
		action1, accepted1 := agent1.ProposeAdjustment(globalGoal)
		if accepted1 {
			agent1.ApplyAction(action1)
		}

		action2, accepted2 := agent2.ProposeAdjustment(globalGoal)
		if accepted2 {
			agent2.ApplyAction(action2)
		}
	}

	fmt.Println("\nAfter negotiation:")
	fmt.Printf("Stubborn agent: phase=%.2f (resisted global goal)\n", agent1.GetPhase())
	fmt.Printf("Flexible agent: phase=%.2f (moved toward goal)\n", agent2.GetPhase())
}

// Benchmark measures convergence time for different swarm sizes.
func Benchmark() {
	sizes := []int{10, 50, 100, 500}

	for _, size := range sizes {
		goal := State{
			Phase:     0,
			Frequency: 100 * time.Millisecond,
			Coherence: 0.8,
		}

		swarm := NewSwarm(size, goal)

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
		go swarm.Run(ctx)

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
}
