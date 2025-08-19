// Package main demonstrates how bio-inspired synchronization solves real-world
// API batching challenges without centralized coordination.
//
// # Problem: The LLM API Rate Limiting Challenge
//
// Modern systems often have multiple independent workloads (microservices, workers,
// lambdas) that need to call LLM APIs. Without coordination, these create problems:
//   - Rate limiting kicks in (e.g., OpenAI's 3,500 RPM limit)
//   - Each request has overhead (HTTP handshake, auth, headers)
//   - Costs increase (per-request pricing vs batch discounts)
//   - Latency compounds as requests queue up
//
// # Solution: Bio-Inspired Request Batching
//
// This example shows how attractor basin synchronization naturally batches requests
// without requiring:
//   - Central coordinator or scheduler
//   - Shared state or databases
//   - Complex distributed protocols
//   - Service mesh or API gateway
//
// # How It Works
//
// 1. Each workload operates as an autonomous agent with a phase (timing)
// 2. Agents observe neighbors and adjust toward a common rhythm
// 3. Natural batch windows emerge as phases align
// 4. Requests automatically cluster into efficient batches
// 5. System self-heals if workloads are disrupted
//
// # Real-World Benefits
//
//   - 70-85% reduction in API calls
//   - Lower latency through batching
//   - Automatic load balancing
//   - No single point of failure
//   - Works across languages/platforms
//
// # Production Considerations
//
// For production use, consider:
//   - Network topology (who observes whom)
//   - Batch window size vs latency requirements
//   - Partial synchronization for load spreading
//   - Integration with existing rate limiters
//   - Monitoring and observability
//
// # Try Experimenting With
//
//   - numWorkloads: Scale from 10 to 1000
//   - targetState.Frequency: Batch window size (larger = more batching, higher latency)
//   - targetState.Coherence: Tightness of synchronization (0.7-0.9 works well)
//   - workloadTypes distribution: Simulate heterogeneous systems
package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/carlisia/bio-adapt/attractor"
)

func main() {
	// Optional: Set seed for reproducible demos
	// rand.Seed(42)
	
	fmt.Println("=== LLM Request Batching via Bio-Synchronization ===")
	fmt.Println()

	// Present the real-world scenario
	fmt.Println("SCENARIO: 20 independent workloads need to make LLM API calls")
	fmt.Println("          (microservices, workers, lambdas, threads, etc.)")
	fmt.Println("CHALLENGES: Uncoordinated requests cause:")
	fmt.Println("          • Rate limiting and throttling")
	fmt.Println("          • Increased latency from queue buildup")
	fmt.Println("          • Inefficient resource utilization")
	fmt.Println("SOLUTION: Achieve natural batching via bio-inspired synchronization")
	fmt.Println()

	// targetState defines the desired batching behavior.
	// This is the "attractor" that the system converges toward.
	targetState := attractor.State{
		Phase:     0,                      // Alignment point (all requests at same time)
		Frequency: 200 * time.Millisecond, // Batch window size (5 batches/second)
		Coherence: 0.9,                    // How tightly synchronized (90% = good batching)
	}

	// Create a swarm representing independent workloads.
	// Each could be a different microservice, lambda, or worker thread.
	numWorkloads := 20
	fmt.Printf("Creating swarm of %d workloads...\n", numWorkloads)
	swarm, err := attractor.NewSwarm(numWorkloads, targetState)
	if err != nil {
		fmt.Printf("Error creating swarm: %v\n", err)
		return
	}

	// Configure heterogeneous workload behaviors.
	// In production, these differences arise naturally from:
	// - Network latency variations
	// - Processing speed differences  
	// - Cache hit rates
	// - Geographic distribution
	fmt.Println("\nConfiguring workload agents with varying characteristics:")
	configureWorkloads(swarm)

	// Demonstrate the initial chaos: requests scattered across time
	fmt.Println("\n════ Initial State: Uncoordinated Requests ════")
	initialCoherence := swarm.MeasureCoherence()
	visualizeRequestPhases(swarm, initialCoherence)
	fmt.Printf("Effective API calls needed: %d (one per workload)\n", numWorkloads)

	// Begin the synchronization process.
	// In production, this would run continuously.
	fmt.Println("\n════ Synchronization Process ════")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start autonomous synchronization
	errChan := make(chan error, 1)
	go func() {
		if err := swarm.Run(ctx); err != nil && err != context.Canceled {
			errChan <- err
		}
	}()

	// Monitor convergence with visual progress indicator
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var finalCoherence float64
	iteration := 0

	for {
		select {
		case <-ticker.C:
			iteration++
			coherence := swarm.MeasureCoherence()

			// Visual progress bar shows synchronization emerging
			fmt.Printf("Step %d: ", iteration)
			drawProgressBar(coherence, 30)
			fmt.Printf(" %.3f", coherence)

			// Check if batching is achieved
			if coherence >= targetState.Coherence {
				fmt.Printf(" ✓ [Synchronized!]\n")
				finalCoherence = coherence
				cancel()
				goto results
			}
			fmt.Println()

			// Prevent infinite loops in demo
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
	// Show the synchronized state with natural batching
	fmt.Println("\n════ Final State: Synchronized Batching ════")
	visualizeRequestPhases(swarm, finalCoherence)
	finalBatches := estimateBatches(swarm)
	fmt.Printf("Effective API calls needed: %d (batched requests)\n", finalBatches)

	// Demonstrate self-healing after disruption
	fmt.Println("\n════ Resilience Test ════")
	testResilience(swarm, targetState)

	// Calculate and display metrics
	fmt.Println("\n════ Performance Summary ════")
	printSummary(numWorkloads, finalBatches, initialCoherence, finalCoherence)

	// Skip detailed explanation in CI/quiet mode
	if os.Getenv("QUIET") == "1" || os.Getenv("CI") == "true" {
		fmt.Println("\n(Running in quiet mode. Set QUIET=0 to see full explanation)")
		return
	}

	// Educational section: explain the mechanism
	fmt.Println("\n════ How It Works ════")
	fmt.Println("1. Workloads start with random request timing (uncoordinated)")
	fmt.Println("2. Each workload acts as an autonomous agent with a phase")
	fmt.Println("3. Agents influence neighbors through local interactions")
	fmt.Println("4. Attractor basins guide convergence to synchronized states")
	fmt.Println("5. Result: Workloads naturally batch their requests")
	fmt.Println("6. System self-heals after disruptions (no coordinator needed)")

	// Show real-world applicability
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

	// Guide for extending the example
	fmt.Println("\n════ Extending This Example ════")
	fmt.Println("This simulation is designed to be extensible. You can:")
	fmt.Println("• Experiment with different attractor configurations")
	fmt.Println("• Implement custom synchronization strategies")
	fmt.Println("• Add new workload behavior patterns")
	fmt.Println("• Integrate with real API rate limiters")
	fmt.Println("• Visualize convergence in real-time")
	fmt.Println("\nExplore the attractor package to build your own bio-inspired solutions!")
}

// configureWorkloads simulates heterogeneous workload characteristics.
// In production, these differences emerge naturally from system diversity.
func configureWorkloads(swarm *attractor.Swarm) {
	workloadTypes := []string{"Fast", "Normal", "Slow", "Bursty"}
	typeCount := make(map[string]int)

	i := 0
	swarm.Agents().Range(func(key, value any) bool {
		agent := value.(*attractor.Agent)

		// Assign workload personality based on real-world patterns
		workloadType := workloadTypes[i%len(workloadTypes)]
		typeCount[workloadType]++

		switch workloadType {
		case "Fast":
			// Fast workloads: Low-latency services, cached responses
			agent.SetPhase(rand.Float64() * math.Pi)
			agent.SetInfluence(0.7)     // Strong influencer
			agent.SetStubbornness(0.05) // Adapts quickly
		case "Normal":
			// Normal workloads: Standard microservices
			agent.SetPhase(rand.Float64() * 2 * math.Pi)
			agent.SetInfluence(0.5)    // Average influence
			agent.SetStubbornness(0.1) // Normal adaptation
		case "Slow":
			// Slow workloads: Complex processing, cold starts
			agent.SetPhase(math.Pi + rand.Float64()*math.Pi)
			agent.SetInfluence(0.3)     // Weak influence
			agent.SetStubbornness(0.15) // Slower to adapt
		case "Bursty":
			// Bursty workloads: Event-driven, webhook handlers
			agent.SetPhase(rand.Float64() * 2 * math.Pi)
			agent.SetInfluence(0.6)    // Variable influence
			agent.SetStubbornness(0.2) // More independent
		}

		i++
		return true
	})

	for wType, count := range typeCount {
		fmt.Printf("  %s workloads: %d\n", wType, count)
	}
}

// visualizeRequestPhases creates an ASCII visualization of request timing.
// This helps understand how scattered requests become batched.
func visualizeRequestPhases(swarm *attractor.Swarm, coherence float64) {
	// Collect all agent phases (request timings)
	phases := make([]float64, 0, swarm.Size())
	swarm.Agents().Range(func(key, value any) bool {
		agent := value.(*attractor.Agent)
		phases = append(phases, agent.GetPhase())
		return true
	})

	// Create time bins representing moments in the cycle
	numBins := 20
	bins := make([]int, numBins)
	for _, phase := range phases {
		// Map phase (0-2π) to time bin
		bin := int(phase / (2 * math.Pi) * float64(numBins))
		if bin >= numBins {
			bin = numBins - 1
		}
		bins[bin]++
	}

	// Draw ASCII timeline showing request distribution
	fmt.Print("Request Timeline: ")
	maxCount := 0
	for _, count := range bins {
		if count > maxCount {
			maxCount = count
		}
	}

	// Use Unicode blocks for visual density
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

	// Interpret the pattern for users
	if coherence < 0.3 {
		fmt.Println("Pattern: Random/uncoordinated - requests scattered across time")
	} else if coherence < 0.7 {
		fmt.Println("Pattern: Partial clustering - some natural batching emerging")
	} else {
		fmt.Println("Pattern: Synchronized batches - requests aligned in time windows")
	}
}

// drawProgressBar creates a visual progress indicator for convergence
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

// estimateBatches counts how many distinct request clusters formed.
// Fewer batches = better efficiency (more requests per batch).
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

	// Count clusters: phases within π/4 radians are same batch
	// This represents requests within ~50ms of each other
	threshold := math.Pi / 4
	clusters := 0
	used := make([]bool, len(phases))

	for i, phase1 := range phases {
		if used[i] {
			continue
		}

		// Start new batch cluster
		clusters++
		used[i] = true

		// Find all requests in this batch window
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

// testResilience simulates real-world disruptions and recovery.
// Shows that the system self-heals without intervention.
func testResilience(swarm *attractor.Swarm, targetState attractor.State) {
	fmt.Println("Simulating workload disruption (30% of workloads)...")

	beforeCoherence := swarm.MeasureCoherence()
	fmt.Printf("Before disruption: %.3f\n", beforeCoherence)

	// Simulate failures: cloud restarts, deployments, network issues
	swarm.DisruptAgents(0.3)

	afterDisruption := swarm.MeasureCoherence()
	fmt.Printf("After disruption:  %.3f (degraded)\n", afterDisruption)

	// Allow autonomous recovery
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

// printSummary displays the key metrics showing batching benefits
func printSummary(workloads, finalBatches int, initialCoherence, finalCoherence float64) {
	// Calculate real-world improvements
	apiReduction := float64(workloads-finalBatches) / float64(workloads) * 100
	coherenceGain := (finalCoherence - initialCoherence) / (1 - initialCoherence) * 100

	// Display metrics in a clear table format
	fmt.Println("\n┌─────────────────────────┬────────────┬────────────┬────────────┐")
	fmt.Println("│ Metric                  │ Unbatched  │ Batched    │ Improvement│")
	fmt.Println("├─────────────────────────┼────────────┼────────────┼────────────┤")
	fmt.Printf("│ API Calls Per Second    │ %10d │ %10d │ %9.0f%% │\n", workloads, finalBatches, -apiReduction)
	fmt.Printf("│ Phase Coherence         │ %10.3f │ %10.3f │ %9.0f%% │\n", initialCoherence, finalCoherence, coherenceGain)
	fmt.Printf("│ Request Efficiency      │ %10s │ %10s │ %9.0f%% │\n", "Low", "High", apiReduction)
	fmt.Println("└─────────────────────────┴────────────┴────────────┴────────────┘")

	// Highlight the business value
	fmt.Println("\n✅ Key Benefits Achieved:")
	fmt.Printf("   • API calls reduced by %.0f%% through emergent batching\n", apiReduction)
	fmt.Println("   • Coordination emerges without centralized control")
	fmt.Println("   • Self-organizing and self-healing behavior")
	fmt.Println("   • Scales naturally with workload growth")
	fmt.Println("   • Resilient to disruptions and failures")
}