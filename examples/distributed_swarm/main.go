// Distributed Swarm Example
// This example shows how multiple sub-swarms can operate independently
// yet achieve global coherence through local interactions.

package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/carlisia/bio-adapt/biofield"
)

func main() {
	fmt.Println("=== Distributed Swarm Coordination Example ===")
	fmt.Println()

	// Global target state for all sub-swarms
	globalTarget := biofield.State{
		Phase:     1.57, // π/2 radians
		Frequency: 200 * time.Millisecond,
		Coherence: 0.75,
	}

	// Create multiple sub-swarms representing different regions
	numSubSwarms := 3
	swarmSizes := []int{20, 30, 25}
	subSwarms := make([]*biofield.Swarm, numSubSwarms)

	fmt.Printf("Creating %d sub-swarms:\n", numSubSwarms)
	var err error
	for i := range numSubSwarms {
		subSwarms[i], err = biofield.NewSwarm(swarmSizes[i], globalTarget)
		if err != nil {
			fmt.Printf("Error creating sub-swarm %d: %v\n", i, err)
			return
		}

		// Give each sub-swarm slightly different initial conditions
		subSwarms[i].Agents().Range(func(key, value any) bool {
			agent := value.(*biofield.Agent)
			// Add regional bias to initial phase
			regionalPhase := float64(i) * 0.5
			agent.SetPhase(regionalPhase + rand.Float64()*0.5)
			return true
		})

		fmt.Printf("  Sub-swarm %d: %d agents (regional bias: %.2f)\n",
			i+1, swarmSizes[i], float64(i)*0.5)
	}

	// Measure initial global coherence
	globalCoherence := measureGlobalCoherence(subSwarms)
	fmt.Printf("\nInitial global coherence: %.3f\n", globalCoherence)

	// Create context for all sub-swarms
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Start all sub-swarms concurrently
	fmt.Println("\nStarting distributed synchronization...")
	var wg sync.WaitGroup
	errChan := make(chan error, len(subSwarms))
	for i, swarm := range subSwarms {
		wg.Add(1)
		go func(idx int, s *biofield.Swarm) {
			defer wg.Done()
			if err := s.Run(ctx); err != nil {
				errChan <- fmt.Errorf("sub-swarm %d: %w", idx, err)
			}
		}(i, swarm)
	}

	// Create inter-swarm connections (bridge agents)
	fmt.Println("Establishing inter-swarm connections...")
	connectSubSwarms(subSwarms)

	// Monitor convergence
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	startTime := time.Now()
	iteration := 0

	for {
		select {
		case err := <-errChan:
			fmt.Printf("\nError in sub-swarm: %v\n", err)
			cancel()
			goto done
		case <-ticker.C:
			iteration++

			// Measure individual and global coherence
			fmt.Printf("\n--- Iteration %d (%.1fs) ---\n",
				iteration, time.Since(startTime).Seconds())

			for i, swarm := range subSwarms {
				coherence := swarm.MeasureCoherence()
				fmt.Printf("  Sub-swarm %d: %.3f\n", i+1, coherence)
			}

			globalCoherence = measureGlobalCoherence(subSwarms)
			fmt.Printf("  Global:      %.3f", globalCoherence)

			if globalCoherence >= globalTarget.Coherence {
				fmt.Printf(" ✓ [Target reached!]\n")
				cancel()
				goto done
			}
			fmt.Println()

			if iteration >= 10 {
				fmt.Println("\nMaximum iterations reached")
				cancel()
				goto done
			}

		case <-ctx.Done():
			goto done
		}
	}

done:
	// Wait for all goroutines to finish
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	// Final results
	fmt.Println("\n=== Final Results ===")
	for i, swarm := range subSwarms {
		coherence := swarm.MeasureCoherence()
		fmt.Printf("Sub-swarm %d final coherence: %.3f\n", i+1, coherence)
	}

	finalGlobalCoherence := measureGlobalCoherence(subSwarms)
	fmt.Printf("\nGlobal final coherence: %.3f\n", finalGlobalCoherence)
	fmt.Printf("Target achieved: %v\n", finalGlobalCoherence >= globalTarget.Coherence)
}

// measureGlobalCoherence calculates coherence across all sub-swarms
func measureGlobalCoherence(swarms []*biofield.Swarm) float64 {
	var phases []float64

	for _, swarm := range swarms {
		swarm.Agents().Range(func(key, value any) bool {
			agent := value.(*biofield.Agent)
			phases = append(phases, agent.GetPhase())
			return true
		})
	}

	if len(phases) == 0 {
		return 0
	}

	// Calculate Kuramoto order parameter
	var sumCos, sumSin float64
	for _, phase := range phases {
		sumCos += math.Cos(phase)
		sumSin += math.Sin(phase)
	}

	n := float64(len(phases))
	return math.Sqrt(sumCos*sumCos+sumSin*sumSin) / n
}

// connectSubSwarms creates bridge connections between sub-swarms
func connectSubSwarms(swarms []*biofield.Swarm) {
	// Connect adjacent sub-swarms through a few bridge agents
	for i := range len(swarms) - 1 {
		// Select 2 random agents from each swarm to act as bridges
		bridgeCount := 0
		swarms[i].Agents().Range(func(key1, value1 any) bool {
			if bridgeCount >= 2 {
				return false
			}

			agent1 := value1.(*biofield.Agent)
			connected := 0

			swarms[i+1].Agents().Range(func(key2, value2 any) bool {
				if connected >= 1 {
					return false
				}

				agent2 := value2.(*biofield.Agent)

				// Create bidirectional connection
				agent1.Neighbors().Store(agent2.ID, agent2)
				agent2.Neighbors().Store(agent1.ID, agent1)

				connected++
				return true
			})

			bridgeCount++
			return true
		})
	}
}

