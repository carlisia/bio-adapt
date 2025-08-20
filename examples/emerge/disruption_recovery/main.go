// Package main demonstrates system resilience to disruptions.
// This example demonstrates the system's resilience to various types of
// disruptions and its ability to recover and maintain coherence.
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/carlisia/bio-adapt/emerge/agent"
	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/monitoring"
	"github.com/carlisia/bio-adapt/emerge/swarm"
	"github.com/carlisia/bio-adapt/internal/random"
)

// disruption represents a disruption test case.
type disruption struct {
	name        string
	disruptFunc func(*swarm.Swarm)
	severity    string
}

// setupSwarmAndMonitoring creates and configures the swarm and monitoring.
func setupSwarmAndMonitoring() (*swarm.Swarm, *monitoring.Monitor, context.Context, context.CancelFunc) {
	target := core.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.85,
	}

	swarmSize := 50
	s, err := swarm.New(swarmSize, target)
	if err != nil {
		fmt.Printf("Error creating swarm: %v\n", err)
		panic(err)
	}
	fmt.Printf("Created swarm with %d agents\n", swarmSize)
	fmt.Printf("Target coherence: %.2f\n\n", target.Coherence)

	monitor := monitoring.New()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	return s, monitor, ctx, cancel
}

// startSwarm starts the swarm in a goroutine.
func startSwarm(ctx context.Context, s *swarm.Swarm, cancel context.CancelFunc) {
	errChan := make(chan error, 1)
	go func() {
		if err := s.Run(ctx); err != nil {
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
			// Context canceled
		}
	}()
}

// runInitialStabilization runs the initial stabilization phase.
func runInitialStabilization(s *swarm.Swarm, monitor *monitoring.Monitor) float64 {
	fmt.Println("Phase 1: Initial Stabilization")
	fmt.Println("------------------------------")

	for i := range 5 {
		time.Sleep(500 * time.Millisecond)
		coherence := s.MeasureCoherence()
		monitor.RecordSample(coherence)
		fmt.Printf("  %.1fs: Coherence = %.3f\n", float64(i+1)*0.5, coherence)
	}

	initialStableCoherence := monitor.Latest()
	fmt.Printf("Stabilized at coherence: %.3f\n\n", initialStableCoherence)
	return initialStableCoherence
}

// getDisruptions returns the list of disruption test cases.
func getDisruptions() []disruption {
	return []disruption{
		{
			name: "Random Phase Disruption",
			disruptFunc: func(s *swarm.Swarm) {
				randomPhaseDisruption(s, 0.2) // Disrupt 20% of agents
			},
			severity: "Moderate",
		},
		{
			name: "Energy Depletion",
			disruptFunc: func(s *swarm.Swarm) {
				energyDepletionDisruption(s, 0.3) // Deplete 30% of agents
			},
			severity: "High",
		},
		{
			name: "Network Partition",
			disruptFunc: func(s *swarm.Swarm) {
				networkPartitionDisruption(s)
			},
			severity: "Severe",
		},
		{
			name: "Stubborn Agents",
			disruptFunc: func(s *swarm.Swarm) {
				stubbornAgentDisruption(s, 0.15) // Make 15% stubborn
			},
			severity: "Low",
		},
		{
			name: "Cascade Failure",
			disruptFunc: func(s *swarm.Swarm) {
				cascadeFailureDisruption(s)
			},
			severity: "Critical",
		},
	}
}

// testDisruptions tests each disruption and monitors recovery.
func testDisruptions(s *swarm.Swarm, monitor *monitoring.Monitor, disruptions []disruption) {
	for idx, d := range disruptions {
		testSingleDisruption(s, monitor, d, idx+2)

		// Let system stabilize before next disruption
		time.Sleep(1 * time.Second)
		fmt.Println()
	}
}

// testSingleDisruption tests a single disruption.
func testSingleDisruption(s *swarm.Swarm, monitor *monitoring.Monitor, d disruption, phase int) {
	fmt.Printf("Phase %d: %s (Severity: %s)\n", phase, d.name, d.severity)
	fmt.Println("------------------------------")

	// Measure before disruption
	beforeCoherence := s.MeasureCoherence()
	fmt.Printf("Before disruption: %.3f\n", beforeCoherence)

	// Apply disruption
	d.disruptFunc(s)

	// Measure immediately after disruption
	afterCoherence := s.MeasureCoherence()
	monitor.RecordSample(afterCoherence)
	fmt.Printf("After disruption:  %.3f (Δ = %.3f)\n",
		afterCoherence, afterCoherence-beforeCoherence)

	// Monitor recovery
	monitorRecovery(s, monitor, beforeCoherence)
}

// monitorRecovery monitors the recovery process.
func monitorRecovery(s *swarm.Swarm, monitor *monitoring.Monitor, beforeCoherence float64) {
	fmt.Println("Recovery progress:")
	recoveryStart := time.Now()
	recovered := false

	for range 10 {
		time.Sleep(500 * time.Millisecond)
		coherence := s.MeasureCoherence()
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
}

// printFinalAnalysis prints the final recovery analysis.
func printFinalAnalysis(monitor *monitoring.Monitor, targetCoherence float64) {
	fmt.Println("=== Recovery Analysis ===")
	fmt.Println("------------------------")

	history := monitor.History()
	avgCoherence := monitor.Average()

	// Find min and max from history
	minCoherence, maxCoherence := findMinMax(history)

	fmt.Printf("Average coherence:  %.3f\n", avgCoherence)
	fmt.Printf("Minimum coherence:  %.3f\n", minCoherence)
	fmt.Printf("Maximum coherence:  %.3f\n", maxCoherence)
	fmt.Printf("Resilience factor:  %.3f\n", avgCoherence/targetCoherence)
}

// findMinMax finds minimum and maximum values in a slice.
func findMinMax(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}

	minVal, maxVal := values[0], values[0]
	for _, v := range values[1:] {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}
	return minVal, maxVal
}

// runExtremeStressTest runs the extreme stress test.
func runExtremeStressTest(s *swarm.Swarm) {
	fmt.Println("\n=== Extreme Stress Test ===")
	fmt.Println("--------------------------")
	fmt.Println("Applying multiple simultaneous disruptions...")

	preStressCoherence := s.MeasureCoherence()
	fmt.Printf("Before stress test: %.3f\n", preStressCoherence)

	// Apply multiple disruptions at once
	randomPhaseDisruption(s, 0.3)
	energyDepletionDisruption(s, 0.2)
	stubbornAgentDisruption(s, 0.1)

	stressedCoherence := s.MeasureCoherence()
	fmt.Printf("After stress test:  %.3f (Δ = %.3f)\n",
		stressedCoherence, stressedCoherence-preStressCoherence)

	// Monitor recovery from extreme stress
	monitorExtremeRecovery(s, preStressCoherence)
}

// monitorExtremeRecovery monitors recovery from extreme stress.
func monitorExtremeRecovery(s *swarm.Swarm, preStressCoherence float64) {
	fmt.Println("\nMonitoring recovery from extreme stress:")
	for i := range 8 {
		time.Sleep(1 * time.Second)
		coherence := s.MeasureCoherence()
		fmt.Printf("  %ds: %.3f", i+1, coherence)
		if coherence >= preStressCoherence*0.8 {
			fmt.Printf(" ✓ [System resilient!]")
			break
		}
		fmt.Println()
	}
}

func main() {
	fmt.Println("=== Disruption Recovery and Resilience Example ===")
	fmt.Println()

	// Setup swarm and monitoring
	s, monitor, ctx, cancel := setupSwarmAndMonitoring()
	defer cancel()

	// Start the swarm
	startSwarm(ctx, s, cancel)

	// Initial stabilization phase
	_ = runInitialStabilization(s, monitor)

	// Test different types of disruptions
	disruptions := getDisruptions()
	testDisruptions(s, monitor, disruptions)

	// Final analysis
	printFinalAnalysis(monitor, 0.85) // target coherence

	// Test extreme scenario
	runExtremeStressTest(s)

	fmt.Println("\n\nDemo complete!")
}

// randomPhaseDisruption randomly changes the phase of a fraction of agents.
func randomPhaseDisruption(s *swarm.Swarm, fraction float64) {
	count := 0
	for _, a := range s.Agents() {
		if random.Float64() < fraction {
			// Random phase between 0 and 2π
			a.SetPhase(random.Float64() * 2 * 3.14159)
			count++
		}
	}
	fmt.Printf("  Disrupted %d agents with random phases\n", count)
}

// energyDepletionDisruption depletes energy from a fraction of agents.
func energyDepletionDisruption(s *swarm.Swarm, fraction float64) {
	count := 0
	for _, a := range s.Agents() {
		if random.Float64() < fraction {
			// Reduce energy to critical levels
			a.SetEnergy(5 + random.Float64()*10)
			count++
		}
	}
	fmt.Printf("  Depleted energy in %d agents\n", count)
}

// networkPartitionDisruption simulates network partition by removing connections.
func networkPartitionDisruption(s *swarm.Swarm) {
	// Remove connections between two halves of the network
	swarmAgents := s.Agents()
	agents := make([]*agent.Agent, 0, len(swarmAgents))
	for _, a := range swarmAgents {
		agents = append(agents, a)
	}

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
				if random.Float64() < 0.1 { // Restore 10% of connections
					agents[i].Neighbors().Store(agents[j].ID, agents[j])
					agents[j].Neighbors().Store(agents[i].ID, agents[i])
				}
			}
		}
		fmt.Println("  [Network partition partially healed]")
	}()
}

// stubbornAgentDisruption makes some agents very stubborn.
func stubbornAgentDisruption(s *swarm.Swarm, fraction float64) {
	count := 0
	for _, a := range s.Agents() {
		if random.Float64() < fraction {
			// Make agent very stubborn
			a.SetStubbornness(0.9 + random.Float64()*0.1)
			// Give them different local goals
			a.SetLocalGoal(random.Float64() * 2 * 3.14159)
			count++
		}
	}
	fmt.Printf("  Made %d agents stubborn with conflicting goals\n", count)
}

// cascadeFailureDisruption simulates a cascade failure starting from one agent.
func cascadeFailureDisruption(s *swarm.Swarm) {
	// Pick a random agent to start the cascade
	var seedAgent *agent.Agent
	for _, a := range s.Agents() {
		if seedAgent == nil && random.Float64() < 0.1 {
			seedAgent = a
			break
		}
	}

	if seedAgent == nil {
		return
	}

	// Disrupt the seed agent severely
	seedAgent.SetPhase(random.Float64() * 2 * 3.14159)
	seedAgent.SetEnergy(1)
	seedAgent.SetInfluence(0.1)

	affected := 1

	// Cascade to neighbors
	seedAgent.Neighbors().Range(func(_, value any) bool {
		neighbor, ok := value.(*agent.Agent)
		if !ok {
			return true
		}
		// Neighbors are partially affected
		neighbor.SetPhase(neighbor.Phase() + (random.Float64()-0.5)*2)
		neighbor.SetEnergy(neighbor.Energy() * 0.5)
		affected++

		// Secondary cascade (with lower probability)
		if random.Float64() < 0.3 {
			neighbor.Neighbors().Range(func(_, v2 any) bool {
				if secondary, ok := v2.(*agent.Agent); ok {
					secondary.SetEnergy(secondary.Energy() * 0.7)
					affected++
				}
				return random.Float64() < 0.5 // Continue with 50% probability
			})
		}
		return true
	})

	fmt.Printf("  Cascade failure affected %d agents\n", affected)
}
