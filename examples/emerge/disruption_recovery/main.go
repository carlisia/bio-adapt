// Disruption Recovery Example
// This example demonstrates the system's resilience to various types of
// disruptions and its ability to recover and maintain coherence.

package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/carlisia/bio-adapt/emerge"
)

func main() {
	fmt.Println("=== Disruption Recovery and Resilience Example ===")
	fmt.Println()

	// Create target state
	target := emerge.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.85,
	}

	// Create a robust swarm
	swarmSize := 50
	swarm, err := emerge.NewSwarm(swarmSize, target)
	if err != nil {
		fmt.Printf("Error creating swarm: %v\n", err)
		return
	}
	fmt.Printf("Created swarm with %d agents\n", swarmSize)
	fmt.Printf("Target coherence: %.2f\n\n", target.Coherence)

	// Start monitoring
	monitor := emerge.NewMonitor()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start the swarm
	errChan := make(chan error, 1)
	go func() {
		if err := swarm.Run(ctx); err != nil {
			errChan <- err
		}
	}()

	// Check for immediate errors
	go func() {
		select {
		case err := <-errChan:
			fmt.Printf("\nError in swarm: %v\n", err)
			cancel()
		case <-ctx.Done():
			// Context cancelled
		}
	}()

	// Initial stabilization phase
	fmt.Println("Phase 1: Initial Stabilization")
	fmt.Println("------------------------------")

	for i := range 5 {
		time.Sleep(500 * time.Millisecond)
		coherence := swarm.MeasureCoherence()
		monitor.RecordSample(coherence)
		fmt.Printf("  %.1fs: Coherence = %.3f\n", float64(i+1)*0.5, coherence)
	}

	initialStableCoherence := monitor.GetLatest()
	fmt.Printf("Stabilized at coherence: %.3f\n\n", initialStableCoherence)

	// Test different types of disruptions
	disruptions := []struct {
		name        string
		disruptFunc func(*emerge.Swarm)
		severity    string
	}{
		{
			name: "Random Phase Disruption",
			disruptFunc: func(s *emerge.Swarm) {
				randomPhaseDisruption(s, 0.2) // Disrupt 20% of agents
			},
			severity: "Moderate",
		},
		{
			name: "Energy Depletion",
			disruptFunc: func(s *emerge.Swarm) {
				energyDepletionDisruption(s, 0.3) // Deplete 30% of agents
			},
			severity: "High",
		},
		{
			name: "Network Partition",
			disruptFunc: func(s *emerge.Swarm) {
				networkPartitionDisruption(s)
			},
			severity: "Severe",
		},
		{
			name: "Stubborn Agents",
			disruptFunc: func(s *emerge.Swarm) {
				stubbornAgentDisruption(s, 0.15) // Make 15% stubborn
			},
			severity: "Low",
		},
		{
			name: "Cascade Failure",
			disruptFunc: func(s *emerge.Swarm) {
				cascadeFailureDisruption(s)
			},
			severity: "Critical",
		},
	}

	for idx, disruption := range disruptions {
		fmt.Printf("Phase %d: %s (Severity: %s)\n", idx+2, disruption.name, disruption.severity)
		fmt.Println("------------------------------")

		// Measure before disruption
		beforeCoherence := swarm.MeasureCoherence()
		fmt.Printf("Before disruption: %.3f\n", beforeCoherence)

		// Apply disruption
		disruption.disruptFunc(swarm)

		// Measure immediately after disruption
		afterCoherence := swarm.MeasureCoherence()
		monitor.RecordSample(afterCoherence)
		fmt.Printf("After disruption:  %.3f (Δ = %.3f)\n",
			afterCoherence, afterCoherence-beforeCoherence)

		// Monitor recovery
		fmt.Println("Recovery progress:")
		recoveryStart := time.Now()
		recovered := false

		for range 10 {
			time.Sleep(500 * time.Millisecond)
			coherence := swarm.MeasureCoherence()
			monitor.RecordSample(coherence)

			recoveryTime := time.Since(recoveryStart).Seconds()
			fmt.Printf("  +%.1fs: %.3f", recoveryTime, coherence)

			// Check if recovered to 90% of pre-disruption level
			if coherence >= beforeCoherence*0.9 && !recovered {
				fmt.Printf(" ✓ [Recovered]")
				recovered = true
			}
			fmt.Println()

			if recovered {
				break
			}
		}

		if !recovered {
			fmt.Println("  [Recovery incomplete]")
		}

		// Let system stabilize before next disruption
		time.Sleep(1 * time.Second)
		fmt.Println()
	}

	// Final analysis
	fmt.Println("=== Recovery Analysis ===")
	fmt.Println("------------------------")

	history := monitor.GetHistory()
	avgCoherence := monitor.GetAverage()

	// Find min and max from history
	minCoherence := 1.0
	maxCoherence := 0.0
	for _, c := range history {
		if c < minCoherence {
			minCoherence = c
		}
		if c > maxCoherence {
			maxCoherence = c
		}
	}

	fmt.Printf("Average coherence:  %.3f\n", avgCoherence)
	fmt.Printf("Minimum coherence:  %.3f\n", minCoherence)
	fmt.Printf("Maximum coherence:  %.3f\n", maxCoherence)
	fmt.Printf("Resilience factor:  %.3f\n", avgCoherence/target.Coherence)

	// Test extreme scenario: Multiple simultaneous disruptions
	fmt.Println("\n=== Extreme Stress Test ===")
	fmt.Println("--------------------------")
	fmt.Println("Applying multiple simultaneous disruptions...")

	preStressCoherence := swarm.MeasureCoherence()
	fmt.Printf("Before stress test: %.3f\n", preStressCoherence)

	// Apply multiple disruptions at once
	randomPhaseDisruption(swarm, 0.3)
	energyDepletionDisruption(swarm, 0.2)
	stubbornAgentDisruption(swarm, 0.1)

	stressedCoherence := swarm.MeasureCoherence()
	fmt.Printf("After stress test:  %.3f (Δ = %.3f)\n",
		stressedCoherence, stressedCoherence-preStressCoherence)

	// Monitor recovery from extreme stress
	fmt.Println("\nMonitoring recovery from extreme stress:")
	for i := range 8 {
		time.Sleep(1 * time.Second)
		coherence := swarm.MeasureCoherence()
		fmt.Printf("  %ds: %.3f", i+1, coherence)
		if coherence >= preStressCoherence*0.8 {
			fmt.Printf(" ✓ [System resilient!]")
			break
		}
		fmt.Println()
	}

	fmt.Println("\n\nDemo complete!")
}

// randomPhaseDisruption randomly changes the phase of a fraction of agents
func randomPhaseDisruption(swarm *emerge.Swarm, fraction float64) {
	count := 0
	swarm.Agents().Range(func(key, value any) bool {
		if rand.Float64() < fraction {
			agent := value.(*emerge.Agent)
			// Random phase between 0 and 2π
			agent.SetPhase(rand.Float64() * 2 * 3.14159)
			count++
		}
		return true
	})
	fmt.Printf("  Disrupted %d agents with random phases\n", count)
}

// energyDepletionDisruption depletes energy from a fraction of agents
func energyDepletionDisruption(swarm *emerge.Swarm, fraction float64) {
	count := 0
	swarm.Agents().Range(func(key, value any) bool {
		if rand.Float64() < fraction {
			agent := value.(*emerge.Agent)
			// Reduce energy to critical levels
			agent.SetEnergy(5 + rand.Float64()*10)
			count++
		}
		return true
	})
	fmt.Printf("  Depleted energy in %d agents\n", count)
}

// networkPartitionDisruption simulates network partition by removing connections
func networkPartitionDisruption(swarm *emerge.Swarm) {
	// Remove connections between two halves of the network
	var agents []*emerge.Agent
	swarm.Agents().Range(func(key, value any) bool {
		agents = append(agents, value.(*emerge.Agent))
		return true
	})

	midpoint := len(agents) / 2
	disconnections := 0

	// Disconnect first half from second half
	for i := range midpoint {
		for j := midpoint; j < len(agents); j++ {
			agents[i].Neighbors().Delete(agents[j].ID)
			agents[j].Neighbors().Delete(agents[i].ID)
			disconnections++
		}
	}

	fmt.Printf("  Created network partition (removed %d connections)\n", disconnections)

	// Reconnect after 2 seconds
	go func() {
		time.Sleep(2 * time.Second)
		// Restore some connections
		for i := range midpoint {
			for j := midpoint; j < len(agents); j++ {
				if rand.Float64() < 0.1 { // Restore 10% of connections
					agents[i].Neighbors().Store(agents[j].ID, agents[j])
					agents[j].Neighbors().Store(agents[i].ID, agents[i])
				}
			}
		}
		fmt.Println("  [Network partition partially healed]")
	}()
}

// stubbornAgentDisruption makes some agents very stubborn
func stubbornAgentDisruption(swarm *emerge.Swarm, fraction float64) {
	count := 0
	swarm.Agents().Range(func(key, value any) bool {
		if rand.Float64() < fraction {
			agent := value.(*emerge.Agent)
			// Make agent very stubborn
			agent.SetStubbornness(0.9 + rand.Float64()*0.1)
			// Give them different local goals
			agent.SetLocalGoal(rand.Float64() * 2 * 3.14159)
			count++
		}
		return true
	})
	fmt.Printf("  Made %d agents stubborn with conflicting goals\n", count)
}

// cascadeFailureDisruption simulates a cascade failure starting from one agent
func cascadeFailureDisruption(swarm *emerge.Swarm) {
	// Pick a random agent to start the cascade
	var seedAgent *emerge.Agent
	swarm.Agents().Range(func(key, value any) bool {
		if seedAgent == nil && rand.Float64() < 0.1 {
			seedAgent = value.(*emerge.Agent)
			return false
		}
		return true
	})

	if seedAgent == nil {
		return
	}

	// Disrupt the seed agent severely
	seedAgent.SetPhase(rand.Float64() * 2 * 3.14159)
	seedAgent.SetEnergy(1)
	seedAgent.SetInfluence(0.1)

	affected := 1

	// Cascade to neighbors
	seedAgent.Neighbors().Range(func(key, value any) bool {
		neighbor := value.(*emerge.Agent)
		// Neighbors are partially affected
		neighbor.SetPhase(neighbor.Phase() + (rand.Float64()-0.5)*2)
		neighbor.SetEnergy(neighbor.Energy() * 0.5)
		affected++

		// Secondary cascade (with lower probability)
		if rand.Float64() < 0.3 {
			neighbor.Neighbors().Range(func(k2, v2 any) bool {
				secondary := v2.(*emerge.Agent)
				secondary.SetEnergy(secondary.Energy() * 0.7)
				affected++
				return rand.Float64() < 0.5 // Continue with 50% probability
			})
		}

		return true
	})

	fmt.Printf("  Cascade failure affected %d agents\n", affected)
}
