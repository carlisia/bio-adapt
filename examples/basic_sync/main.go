// Basic Synchronization Example
// This example demonstrates the fundamental bioelectric attractor basin synchronization
// with a small swarm of agents converging to a target state.

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/carlisia/bio-adapt/biofield"
)

func main() {
	fmt.Println("=== Basic Biofield Synchronization Example ===")
	fmt.Println()

	// Define the target state we want the swarm to converge to
	targetState := biofield.State{
		Phase:     0,                      // Target phase (0 radians)
		Frequency: 100 * time.Millisecond, // Oscillation frequency
		Coherence: 0.8,                    // 80% synchronization target
	}

	// Create a small swarm of 10 autonomous agents
	swarmSize := 10
	fmt.Printf("Creating swarm of %d agents...\n", swarmSize)
	swarm, err := biofield.NewSwarm(swarmSize, targetState)
	if err != nil {
		fmt.Printf("Error creating swarm: %v\n", err)
		return
	}

	// Measure initial coherence (should be low due to random initialization)
	initialCoherence := swarm.MeasureCoherence()
	fmt.Printf("Initial coherence: %.3f\n", initialCoherence)

	// Create a context with timeout for the synchronization process
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start the autonomous synchronization process
	fmt.Println("\nStarting synchronization process...")
	errChan := make(chan error, 1)
	go func() {
		if err := swarm.Run(ctx); err != nil {
			errChan <- err
		}
	}()

	// Monitor the convergence process
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	iterations := 0
	for {
		select {
		case <-ticker.C:
			iterations++
			coherence := swarm.MeasureCoherence()
			fmt.Printf("Iteration %d: Coherence = %.3f", iterations, coherence)

			// Check if we've reached the target
			if coherence >= targetState.Coherence {
				fmt.Printf(" ✓ [Target reached!]\n")
				goto done
			}
			fmt.Println()

			if iterations >= 20 {
				fmt.Println("Maximum iterations reached")
				goto done
			}

		case err := <-errChan:
			fmt.Printf("Error in swarm: %v\n", err)
			goto done
		case <-ctx.Done():
			fmt.Println("Context timeout")
			goto done
		}
	}

done:
	// Final measurements
	finalCoherence := swarm.MeasureCoherence()
	improvement := ((finalCoherence - initialCoherence) / initialCoherence) * 100

	fmt.Println("\n=== Results ===")
	fmt.Printf("Initial coherence: %.3f\n", initialCoherence)
	fmt.Printf("Final coherence:   %.3f\n", finalCoherence)
	fmt.Printf("Improvement:       %.1f%%\n", improvement)
	fmt.Printf("Target achieved:   %v\n", finalCoherence >= targetState.Coherence)
}

