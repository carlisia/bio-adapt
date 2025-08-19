// LLM Request Batching Example
// This example demonstrates how bio-inspired synchronization can efficiently
// batch LLM API requests from multiple independent workloads (microservices,
// worker threads, serverless functions, batch jobs, or any concurrent entities).

package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/carlisia/bio-adapt/attractor"
)

func main() {
	fmt.Println("=== LLM Request Batching via Bio-Synchronization ===")
	fmt.Println()
	
	// Problem statement
	fmt.Println("SCENARIO: 20 independent workloads need to make LLM API calls")
	fmt.Println("          (microservices, workers, lambdas, threads, etc.)")
	fmt.Println("PROBLEM:  Uncoordinated requests cause:")
	fmt.Println("          • Rate limiting and throttling")
	fmt.Println("          • Increased latency from queue buildup")
	fmt.Println("          • Inefficient resource utilization")
	fmt.Println("SOLUTION: Bio-inspired synchronization creates natural batching")
	fmt.Println()

	// Define target state for batch synchronization
	targetState := attractor.State{
		Phase:     0,                       // Alignment point for batches
		Frequency: 200 * time.Millisecond,  // 5 batch windows per second
		Coherence: 0.9,                     // 90% synchronization target
	}

	// Create swarm of workloads
	numWorkloads := 20
	fmt.Printf("Creating swarm of %d workloads...\n", numWorkloads)
	swarm, err := attractor.NewSwarm(numWorkloads, targetState)
	if err != nil {
		fmt.Printf("Error creating swarm: %v\n", err)
		return
	}

	// Configure agents for realistic workload behavior
	fmt.Println("\nConfiguring workload agents with varying characteristics:")
	configureWorkloads(swarm)

	// Show initial uncoordinated state
	fmt.Println("\n════ Initial State: Uncoordinated Requests ════")
	initialCoherence := swarm.MeasureCoherence()
	showRequestPattern(swarm, initialCoherence)
	initialBatches := estimateBatches(swarm)
	fmt.Printf("Effective API calls needed: %d (one per workload)\n", numWorkloads)

	// Run synchronization
	fmt.Println("\n════ Synchronization Process ════")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start swarm convergence
	errChan := make(chan error, 1)
	go func() {
		if err := swarm.Run(ctx); err != nil && err != context.Canceled {
			errChan <- err
		}
	}()

	// Monitor convergence with visual feedback
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var finalCoherence float64
	iteration := 0

	for {
		select {
		case <-ticker.C:
			iteration++
			coherence := swarm.MeasureCoherence()
			
			// Show progress
			fmt.Printf("Step %d: ", iteration)
			drawProgressBar(coherence, 30)
			fmt.Printf(" %.3f", coherence)
			
			if coherence >= targetState.Coherence {
				fmt.Printf(" ✓ [Synchronized!]\n")
				finalCoherence = coherence
				cancel()
				goto results
			}
			fmt.Println()

			if iteration >= 15 {
				fmt.Println("\nMax iterations reached")
				finalCoherence = coherence
				cancel()
				goto results
			}

		case err := <-errChan:
			if err != nil && err != context.Canceled {
				fmt.Printf("Swarm error: %v\n", err)
			}
			goto results

		case <-ctx.Done():
			finalCoherence = swarm.MeasureCoherence()
			goto results
		}
	}

results:
	// Show synchronized state
	fmt.Println("\n════ Final State: Synchronized Batching ════")
	showRequestPattern(swarm, finalCoherence)
	finalBatches := estimateBatches(swarm)
	fmt.Printf("Effective API calls needed: %d (batched requests)\n", finalBatches)

	// Test resilience
	fmt.Println("\n════ Resilience Test ════")
	testResilience(swarm, targetState)

	// Summary metrics
	fmt.Println("\n════ Performance Summary ════")
	printSummary(numWorkloads, initialBatches, finalBatches, initialCoherence, finalCoherence)

	// Explain the mechanism
	fmt.Println("\n════ How It Works ════")
	fmt.Println("1. Workloads start with random request timing (uncoordinated)")
	fmt.Println("2. Each workload acts as an autonomous agent with a phase")
	fmt.Println("3. Agents influence neighbors through local interactions")
	fmt.Println("4. Attractor basins guide convergence to synchronized states")
	fmt.Println("5. Result: Workloads naturally batch their requests")
	fmt.Println("6. System self-heals after disruptions (no coordinator needed)")
	
	// Show applicability
	fmt.Println("\n════ Where This Applies ════")
	fmt.Println("This pattern works for ANY concurrent workloads that need coordination:")
	fmt.Println("• Microservices making API calls")
	fmt.Println("• Worker threads in a pool")
	fmt.Println("• Serverless functions (AWS Lambda, Google Cloud Functions)")
	fmt.Println("• Batch processing jobs")
	fmt.Println("• Database connection pools")
	fmt.Println("• IoT devices sending telemetry")
	fmt.Println("• Mobile apps syncing with backend")
	fmt.Println("• Distributed crawlers or scrapers")
	fmt.Println("• Any scenario where multiple entities access a rate-limited resource")
}

// configureWorkloads sets up agents with varying workload characteristics
func configureWorkloads(swarm *attractor.Swarm) {
	workloadTypes := []string{"Fast", "Normal", "Slow", "Bursty"}
	typeCount := make(map[string]int)
	
	i := 0
	swarm.Agents().Range(func(key, value any) bool {
		agent := value.(*attractor.Agent)
		
		// Assign workload characteristics
		workloadType := workloadTypes[i%len(workloadTypes)]
		typeCount[workloadType]++
		
		switch workloadType {
		case "Fast":
			agent.SetPhase(rand.Float64() * math.Pi)
			agent.SetInfluence(0.7)     // Strong influencer
			agent.SetStubbornness(0.05)  // Adapts quickly
		case "Normal":
			agent.SetPhase(rand.Float64() * 2 * math.Pi)
			agent.SetInfluence(0.5)      // Average influence
			agent.SetStubbornness(0.1)   // Normal adaptation
		case "Slow":
			agent.SetPhase(math.Pi + rand.Float64()*math.Pi)
			agent.SetInfluence(0.3)      // Weak influence
			agent.SetStubbornness(0.15)  // Slower to adapt
		case "Bursty":
			agent.SetPhase(rand.Float64() * 2 * math.Pi)
			agent.SetInfluence(0.6)      // Variable influence
			agent.SetStubbornness(0.2)   // More independent
		}
		
		i++
		return true
	})
	
	for wType, count := range typeCount {
		fmt.Printf("  %s workloads: %d\n", wType, count)
	}
}

// showRequestPattern visualizes the request timing distribution
func showRequestPattern(swarm *attractor.Swarm, coherence float64) {
	// Collect phases
	phases := make([]float64, 0, swarm.Size())
	swarm.Agents().Range(func(key, value any) bool {
		agent := value.(*attractor.Agent)
		phases = append(phases, agent.GetPhase())
		return true
	})
	
	// Create time bins (like timeline slots)
	numBins := 20
	bins := make([]int, numBins)
	for _, phase := range phases {
		// Map phase to time bin
		bin := int(phase / (2 * math.Pi) * float64(numBins))
		if bin >= numBins {
			bin = numBins - 1
		}
		bins[bin]++
	}
	
	// Draw timeline
	fmt.Print("Request Timeline: ")
	maxCount := 0
	for _, count := range bins {
		if count > maxCount {
			maxCount = count
		}
	}
	
	for _, count := range bins {
		if count == 0 {
			fmt.Print("·")
		} else if count <= maxCount/4 {
			fmt.Print("▁")
		} else if count <= maxCount/2 {
			fmt.Print("▃")
		} else if count <= 3*maxCount/4 {
			fmt.Print("▅")
		} else {
			fmt.Print("█")
		}
	}
	
	fmt.Printf(" (Coherence: %.3f)\n", coherence)
	
	// Interpret the pattern
	if coherence < 0.3 {
		fmt.Println("Pattern: Random/uncoordinated - requests scattered across time")
	} else if coherence < 0.7 {
		fmt.Println("Pattern: Partial clustering - some natural batching emerging")
	} else {
		fmt.Println("Pattern: Synchronized batches - requests aligned in time windows")
	}
}

// drawProgressBar creates a visual progress indicator
func drawProgressBar(progress float64, width int) {
	filled := int(progress * float64(width))
	fmt.Print("[")
	for i := range width {
		if i < filled {
			fmt.Print("█")
		} else {
			fmt.Print("░")
		}
	}
	fmt.Print("]")
}

// estimateBatches counts distinct request clusters
func estimateBatches(swarm *attractor.Swarm) int {
	// Collect all phases
	phases := make([]float64, 0, swarm.Size())
	swarm.Agents().Range(func(key, value any) bool {
		agent := value.(*attractor.Agent)
		phases = append(phases, agent.GetPhase())
		return true
	})
	
	if len(phases) == 0 {
		return 0
	}
	
	// Count clusters (phases within π/4 are same batch)
	threshold := math.Pi / 4
	clusters := 0
	used := make([]bool, len(phases))
	
	for i, phase1 := range phases {
		if used[i] {
			continue
		}
		
		// Start new cluster
		clusters++
		used[i] = true
		
		// Find all phases in this cluster
		for j, phase2 := range phases {
			if !used[j] {
				diff := math.Abs(attractor.PhaseDifference(phase1, phase2))
				if diff < threshold {
					used[j] = true
				}
			}
		}
	}
	
	return clusters
}

// testResilience demonstrates recovery from disruption
func testResilience(swarm *attractor.Swarm, targetState attractor.State) {
	fmt.Println("Simulating workload disruption (30% of workloads)...")
	
	beforeCoherence := swarm.MeasureCoherence()
	fmt.Printf("Before disruption: %.3f\n", beforeCoherence)
	
	// Disrupt random workloads
	swarm.DisruptAgents(0.3)
	
	afterDisruption := swarm.MeasureCoherence()
	fmt.Printf("After disruption:  %.3f (degraded)\n", afterDisruption)
	
	// Allow self-recovery
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	go func() {
		_ = swarm.Run(ctx) // Context cancellation is expected
	}()
	
	// Wait and measure recovery
	time.Sleep(2 * time.Second)
	afterRecovery := swarm.MeasureCoherence()
	
	fmt.Printf("After recovery:    %.3f", afterRecovery)
	if afterRecovery >= targetState.Coherence*0.85 {
		fmt.Printf(" ✓ [Self-healed without intervention]\n")
	} else {
		fmt.Printf(" [Partial recovery]\n")
	}
}

// printSummary shows the key metrics and benefits
func printSummary(workloads, initialBatches, finalBatches int, initialCoherence, finalCoherence float64) {
	// Calculate improvements
	apiReduction := float64(workloads-finalBatches) / float64(workloads) * 100
	coherenceGain := (finalCoherence - initialCoherence) / (1 - initialCoherence) * 100
	
	fmt.Println("\n┌─────────────────────────┬────────────┬────────────┐")
	fmt.Println("│ Metric                  │ Before     │ After      │")
	fmt.Println("├─────────────────────────┼────────────┼────────────┤")
	fmt.Printf("│ API Calls Required      │ %10d │ %10d │\n", workloads, finalBatches)
	fmt.Printf("│ Coherence              │ %10.3f │ %10.3f │\n", initialCoherence, finalCoherence)
	fmt.Printf("│ Call Reduction         │ %10s │ %9.0f%% │\n", "—", apiReduction)
	fmt.Printf("│ Sync Improvement       │ %10s │ %9.0f%% │\n", "—", coherenceGain)
	fmt.Println("└─────────────────────────┴────────────┴────────────┘")
	
	fmt.Println("\n✅ Key Benefits Achieved:")
	fmt.Printf("   • Reduced API calls by %.0f%% through natural batching\n", apiReduction)
	fmt.Println("   • No central coordinator or scheduler required")
	fmt.Println("   • Self-organizing and self-healing system")
	fmt.Println("   • Scales naturally with number of workloads")
	fmt.Println("   • Resilient to failures and network issues")
}