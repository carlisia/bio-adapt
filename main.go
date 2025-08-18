package main

import (
	"context"
	"fmt"
	"time"

	"github.com/carlisia/bio-adapt/biofield"
)

func main() {
	fmt.Println("Bio-Adapt: Bioelectric Attractor Basin Synchronization Demo")
	fmt.Println("=" + string(make([]byte, 59)) + "=")

	// Define target state
	goal := biofield.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.8, // Target 80% synchronization
	}

	// Create swarm of autonomous agents
	swarmSize := 50
	fmt.Printf("\nCreating swarm of %d autonomous agents...\n", swarmSize)
	swarm := biofield.NewSwarm(swarmSize, goal)

	// Measure initial coherence
	initialCoherence := swarm.MeasureCoherence()
	fmt.Printf("Initial coherence: %.3f\n", initialCoherence)

	// Run autonomous synchronization
	fmt.Println("\nRunning distributed synchronization (no central control)...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start swarm
	go swarm.Run(ctx)

	// Monitor progress
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	start := time.Now()
	for range 6 {
		<-ticker.C
		coherence := swarm.MeasureCoherence()
		elapsed := time.Since(start).Seconds()
		fmt.Printf("  %.1fs: Coherence = %.3f", elapsed, coherence)
		
		if coherence >= goal.Coherence {
			fmt.Printf(" ✓ [Target reached!]\n")
		} else {
			fmt.Println()
		}
	}

	// Final statistics
	finalCoherence := swarm.MeasureCoherence()
	improvement := (finalCoherence - initialCoherence) * 100
	
	fmt.Println("\n" + string(make([]byte, 61)))
	fmt.Println("Summary:")
	fmt.Printf("  Initial coherence: %.3f\n", initialCoherence)
	fmt.Printf("  Final coherence:   %.3f\n", finalCoherence)
	fmt.Printf("  Improvement:       %.1f%%\n", improvement)
	fmt.Printf("  Time elapsed:      %.1fs\n", time.Since(start).Seconds())

	// Demonstrate disruption and recovery
	fmt.Println("\nTesting disruption recovery...")
	swarm.DisruptAgents(0.2) // Disrupt 20% of agents
	disruptedCoherence := swarm.MeasureCoherence()
	fmt.Printf("  After disruption (20%% agents): %.3f\n", disruptedCoherence)
	
	// Let it recover
	time.Sleep(2 * time.Second)
	recoveredCoherence := swarm.MeasureCoherence()
	fmt.Printf("  After 2s recovery:             %.3f\n", recoveredCoherence)

	fmt.Println("\nDemo complete!")
}