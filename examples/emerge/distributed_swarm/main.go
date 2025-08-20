// Package main demonstrates distributed swarm coordination.
// This example shows how multiple sub-swarms can operate independently
// yet achieve global coherence through local interactions.
package main

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/swarm"
	"github.com/carlisia/bio-adapt/internal/random"
)

func main() {
	fmt.Println("=== Distributed Swarm Coordination Example ===")
	fmt.Println()

	// Global target state for all sub-swarms
	globalTarget := core.State{
		Phase:     1.57, // π/2 radians
		Frequency: 200 * time.Millisecond,
		Coherence: 0.75,
	}

	// Create multiple sub-swarms representing different regions
	numSubSwarms := 3
	swarmSizes := []int{20, 30, 25}
	subSwarms := make([]*swarm.Swarm, numSubSwarms)

	fmt.Printf("Creating %d sub-swarms:\n", numSubSwarms)
	var err error
	for i := range numSubSwarms {
		subSwarms[i], err = swarm.New(swarmSizes[i], globalTarget)
		if err != nil {
			fmt.Printf("Error creating sub-swarm %d: %v\n", i, err)
			return
		}

		// Give each sub-swarm slightly different initial conditions
		for _, agent := range subSwarms[i].Agents() {
			// Add regional bias to initial phase
			regionalPhase := float64(i) * 0.5
			agent.SetPhase(regionalPhase + random.Float64()*0.5)
		}

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
	for i, sw := range subSwarms {
		wg.Add(1)
		go func(idx int, s *swarm.Swarm) {
			defer wg.Done()
			if err := s.Run(ctx); err != nil {
				errChan <- fmt.Errorf("sub-swarm %d: %w", idx, err)
			}
		}(i, sw)
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

			for i, s := range subSwarms {
				coherence := s.MeasureCoherence()
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
	for i, s := range subSwarms {
		coherence := s.MeasureCoherence()
		fmt.Printf("Sub-swarm %d final coherence: %.3f\n", i+1, coherence)
	}

	finalGlobalCoherence := measureGlobalCoherence(subSwarms)
	fmt.Printf("\nGlobal final coherence: %.3f\n", finalGlobalCoherence)
	fmt.Printf("Target achieved: %v\n", finalGlobalCoherence >= globalTarget.Coherence)
}

// measureGlobalCoherence calculates coherence across all sub-swarms.
func measureGlobalCoherence(swarms []*swarm.Swarm) float64 {
	var phases []float64

	for _, s := range swarms {
		for _, a := range s.Agents() {
			phases = append(phases, a.Phase())
		}
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

// connectSubSwarms creates bridge connections between sub-swarms.
func connectSubSwarms(swarms []*swarm.Swarm) {
	// Connect adjacent sub-swarms through a few bridge agents
	for i := range len(swarms) - 1 {
		// Select 2 random agents from each swarm to act as bridges
		bridgeCount := 0
		for _, agent1 := range swarms[i].Agents() {
			if bridgeCount >= 2 {
				break
			}

			connected := 0
			for _, agent2 := range swarms[i+1].Agents() {
				if connected >= 1 {
					break
				}

				// Create bidirectional connection
				agent1.Neighbors().Store(agent2.ID, agent2)
				agent2.Neighbors().Store(agent1.ID, agent1)

				connected++
			}
			bridgeCount++
		}
	}
}
