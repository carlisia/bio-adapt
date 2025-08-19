// Energy Management Example
// This example demonstrates how agents manage their energy resources
// and how energy constraints affect synchronization behavior.

package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/carlisia/bio-adapt/attractor"
)

func main() {
	fmt.Println("=== Energy-Constrained Synchronization Example ===")
	fmt.Println()

	// Create target state
	target := attractor.State{
		Phase:     0,
		Frequency: 150 * time.Millisecond,
		Coherence: 0.85,
	}

	// Create swarm with varying energy levels
	swarmSize := 30
	swarm, err := attractor.NewSwarm(swarmSize, target)
	if err != nil {
		fmt.Printf("Error creating swarm: %v\n", err)
		return
	}
	fmt.Println("Configuring agents with different energy profiles:")

	// Configure agents with different energy characteristics
	agentCount := 0
	swarm.Agents().Range(func(key, value any) bool {
		agent := value.(*attractor.Agent)

		if agentCount < 10 {
			// High energy agents - can afford more adjustments
			agent.SetEnergy(100)
			agent.SetInfluence(0.8)
			fmt.Printf("  Agent %s: High energy (100), high influence\n", agent.ID[:8])
		} else if agentCount < 20 {
			// Medium energy agents
			agent.SetEnergy(50)
			agent.SetInfluence(0.5)
			fmt.Printf("  Agent %s: Medium energy (50), medium influence\n", agent.ID[:8])
		} else {
			// Low energy agents - must be conservative
			agent.SetEnergy(20)
			agent.SetInfluence(0.3)
			fmt.Printf("  Agent %s: Low energy (20), low influence\n", agent.ID[:8])
		}

		agentCount++
		return agentCount < swarmSize
	})

	fmt.Printf("\nInitial coherence: %.3f\n", swarm.MeasureCoherence())

	// Run synchronization with energy monitoring
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	fmt.Println("\nStarting energy-aware synchronization...")
	errChan := make(chan error, 1)
	go func() {
		if err := swarm.Run(ctx); err != nil {
			errChan <- err
		}
	}()

	// Monitor energy consumption and coherence
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	iteration := 0
	startTime := time.Now()

	for {
		select {
		case err := <-errChan:
			fmt.Printf("\nError in swarm: %v\n", err)
			goto done
		case <-ticker.C:
			iteration++

			// Calculate energy statistics
			var totalEnergy, minEnergy, maxEnergy float64
			var exhaustedAgents int
			minEnergy = 100

			swarm.Agents().Range(func(key, value any) bool {
				agent := value.(*attractor.Agent)
				energy := agent.GetEnergy()

				totalEnergy += energy
				if energy < minEnergy {
					minEnergy = energy
				}
				if energy > maxEnergy {
					maxEnergy = energy
				}
				if energy < 10 {
					exhaustedAgents++
				}

				return true
			})

			avgEnergy := totalEnergy / float64(swarmSize)
			coherence := swarm.MeasureCoherence()

			fmt.Printf("\n--- Iteration %d (%.1fs) ---\n",
				iteration, time.Since(startTime).Seconds())
			fmt.Printf("  Coherence:        %.3f\n", coherence)
			fmt.Printf("  Avg Energy:       %.1f\n", avgEnergy)
			fmt.Printf("  Energy Range:     %.1f - %.1f\n", minEnergy, maxEnergy)
			fmt.Printf("  Exhausted Agents: %d/%d\n", exhaustedAgents, swarmSize)

			// Simulate energy replenishment for some agents
			if iteration%3 == 0 {
				replenishEnergy(swarm, 0.2) // Replenish 20% of agents
				fmt.Println("  [Energy replenished for some agents]")
			}

			if coherence >= target.Coherence {
				fmt.Printf("\n✓ Target coherence reached!\n")
				goto done
			}

			if iteration >= 15 {
				fmt.Println("\nMaximum iterations reached")
				goto done
			}

			if exhaustedAgents > swarmSize/2 {
				fmt.Println("\nToo many exhausted agents - system struggling")
			}

		case <-ctx.Done():
			goto done
		}
	}

done:
	cancel()

	// Final analysis
	fmt.Println("\n=== Energy Analysis ===")

	// Group agents by remaining energy
	var highEnergy, medEnergy, lowEnergy, exhausted int

	swarm.Agents().Range(func(key, value any) bool {
		agent := value.(*attractor.Agent)
		energy := agent.GetEnergy()

		switch {
		case energy >= 70:
			highEnergy++
		case energy >= 30:
			medEnergy++
		case energy >= 10:
			lowEnergy++
		default:
			exhausted++
		}

		return true
	})

	fmt.Printf("High energy (≥70):    %d agents\n", highEnergy)
	fmt.Printf("Medium energy (30-70): %d agents\n", medEnergy)
	fmt.Printf("Low energy (10-30):    %d agents\n", lowEnergy)
	fmt.Printf("Exhausted (<10):       %d agents\n", exhausted)

	finalCoherence := swarm.MeasureCoherence()
	fmt.Printf("\nFinal coherence: %.3f\n", finalCoherence)
	fmt.Printf("Target achieved: %v\n", finalCoherence >= target.Coherence)

	// Demonstrate energy-based decision making
	fmt.Println("\n=== Energy-Based Decision Example ===")
	demonstrateEnergyDecisions()
}

// replenishEnergy simulates energy recovery for a fraction of agents
func replenishEnergy(swarm *attractor.Swarm, fraction float64) {
	count := 0
	swarm.Agents().Range(func(key, value any) bool {
		if rand.Float64() < fraction {
			agent := value.(*attractor.Agent)
			currentEnergy := agent.GetEnergy()
			// Replenish up to 30 energy units
			agent.SetEnergy(currentEnergy + 30)
			if agent.GetEnergy() > 100 {
				agent.SetEnergy(100)
			}
			count++
		}
		return true
	})
}

// demonstrateEnergyDecisions shows how agents make decisions based on energy
func demonstrateEnergyDecisions() {
	// Create two agents with different energy levels
	richAgent := attractor.NewAgent("energy-rich")
	richAgent.SetEnergy(90)

	poorAgent := attractor.NewAgent("energy-poor")
	poorAgent.SetEnergy(15)

	target := attractor.State{
		Phase:     1.0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.8,
	}

	fmt.Println("\nTwo agents face the same adjustment opportunity:")
	fmt.Printf("Rich agent (90 energy): ")

	// Rich agent can afford expensive adjustments
	action1, accepted1 := richAgent.ProposeAdjustment(target)
	if accepted1 {
		fmt.Printf("Accepts adjustment (cost: %.1f)\n", action1.Cost)
		richAgent.ApplyAction(action1)
	} else {
		fmt.Println("Rejects adjustment")
	}

	fmt.Printf("Poor agent (15 energy): ")

	// Poor agent must be conservative
	action2, accepted2 := poorAgent.ProposeAdjustment(target)
	if accepted2 {
		fmt.Printf("Accepts adjustment (cost: %.1f)\n", action2.Cost)
		poorAgent.ApplyAction(action2)
	} else {
		fmt.Println("Rejects adjustment (too expensive)")
	}

	fmt.Printf("\nFinal energy levels:\n")
	fmt.Printf("  Rich agent: %.1f\n", richAgent.GetEnergy())
	fmt.Printf("  Poor agent: %.1f\n", poorAgent.GetEnergy())
}

