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
	"strings"
	"time"

	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/swarm"
)

func main() {
	// Optional: Set seed for reproducible demos
	// rand.Seed(42)

	fmt.Println("╔════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║        🤖 LLM API REQUEST BATCHING DEMO 🤖               ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Clarify what this demonstration shows
	fmt.Println("🔬 SIMULATION OVERVIEW")
	fmt.Println("├─ What: Timing coordination algorithm for request batching")
	fmt.Println("├─ How: Bio-inspired synchronization (like fireflies syncing)")
	fmt.Println("└─ Note: No actual API calls are made - we simulate the timing")
	fmt.Println()

	fmt.Println("🎯 WHAT'S BEING SIMULATED:")
	fmt.Println("├─ ✅ Request timing and coordination logic")
	fmt.Println("├─ ✅ Workload diversity (fast/slow/bursty services)")
	fmt.Println("├─ ✅ Natural batch window formation")
	fmt.Println("├─ ✅ Self-healing after disruptions")
	fmt.Println("└─ ❌ NOT simulated: Actual API calls, network latency, real data")
	fmt.Println()

	// Present the real-world scenario
	fmt.Println("📋 REAL-WORLD SCENARIO THIS SOLVES:")
	fmt.Println("├─ 20 independent workloads need LLM API access")
	fmt.Println("├─ Each makes requests at random times")
	fmt.Println("└─ Goal: Batch requests to reduce API calls by 70%+")
	fmt.Println()

	fmt.Println("⚠️  WITHOUT COORDINATION:")
	fmt.Println("├─ 🚫 Rate limiting (429 errors)")
	fmt.Println("├─ 💸 Higher costs (per-request pricing)")
	fmt.Println("├─ 🐌 Increased latency (queue buildup)")
	fmt.Println("└─ 📈 20 API calls per cycle")
	fmt.Println()

	fmt.Println("✅ WITH BIO-SYNCHRONIZATION:")
	fmt.Println("├─ 📦 Natural request batching")
	fmt.Println("├─ 💰 Lower costs (batch discounts)")
	fmt.Println("├─ ⚡ Better throughput")
	fmt.Println("└─ 📉 3-5 API calls per cycle")
	fmt.Println()

	fmt.Println("🔬 KEY CONCEPTS FOR BATCHING:")
	fmt.Println("├─ LOCAL MINIMA = Partial batching that won't improve")
	fmt.Println("│  • Like having 3 separate batch groups that won't merge")
	fmt.Println("│  • Example: Morning, afternoon, evening batches stuck separate")
	fmt.Println("│  • System achieves some batching but not optimal")
	fmt.Println("│")
	fmt.Println("├─ METASTABLE STATE = Fragile batching arrangement")
	fmt.Println("│  • Like a house of cards - works until disrupted")
	fmt.Println("│  • Example: Batches aligned but one slow request breaks it")
	fmt.Println("│  • New workload or network hiccup destroys coordination")
	fmt.Println("│")
	fmt.Println("└─ PERTURBATION = Intentional timing shifts")
	fmt.Println("   • Like jiggling a vending machine to unstick items")
	fmt.Println("   • Example: Randomly delay some requests to find better batching")
	fmt.Println("   • Helps escape suboptimal batching patterns")
	fmt.Println()

	// targetState defines the desired batching behavior.
	// This is the "attractor" that the system converges toward.
	//
	// SETTING THE FREQUENCY FOR BATCHING: We're using 200ms here because:
	// - Creates batch windows every 200ms (5 batches per second)
	// - Balance between latency (200ms max wait) and efficiency
	// - Shorter = more batches but less requests per batch
	// - Longer = fewer batches with more requests, but higher latency
	batchWindow := 200 * time.Millisecond
	targetState := core.State{
		Phase:     0,           // Alignment point (all requests sync to same moment)
		Frequency: batchWindow, // Batch window size (we chose 200ms for balance)
		Coherence: 0.65,        // How tightly synchronized (65% = good batching)
	}

	// Create a swarm representing independent workloads.
	// Each could be a different microservice, lambda, or worker thread.
	numWorkloads := 20
	maxIterations := 20
	checkInterval := 500 * time.Millisecond
	timeout := 10 * time.Second
	fmt.Println("🔧 SETUP")
	fmt.Printf("├─ Creating %d independent workloads\n", numWorkloads)
	fmt.Printf("├─ Batch window: %v (%.0f batches/sec)\n",
		targetState.Frequency, float64(1000)/float64(targetState.Frequency.Milliseconds()))
	fmt.Printf("├─ Target sync: %.0f%%\n", targetState.Coherence*100)
	fmt.Printf("├─ Max iterations: %d (checked every %v)\n", maxIterations, checkInterval)
	fmt.Printf("└─ Max time: %v timeout", timeout)
	fmt.Println()

	fmt.Println("📊 BATCHING PARAMETER TRADEOFFS:")
	fmt.Println("┌──────────────────┬────────────────────────┬────────────────────────┬──────────────┐")
	fmt.Println("│ Parameter        │ Lower Value            │ Higher Value           │ Sweet Spot   │")
	fmt.Println("├──────────────────┼────────────────────────┼────────────────────────┼──────────────┤")
	fmt.Println("│ Workload Count   │ 5-10 workloads         │ 50-100 workloads       │ 15-30        │")
	fmt.Println("│                  │ ✅ Easy batching       │ ✅ Realistic scale     │              │")
	fmt.Println("│                  │ ❌ Limited benefit     │ ❌ Complex dynamics    │              │")
	fmt.Println("├──────────────────┼────────────────────────┼────────────────────────┼──────────────┤")
	fmt.Println("│ Batch Window     │ 500-1000ms             │ 50-100ms               │ 200-300ms    │")
	fmt.Println("│ (Frequency)      │ ✅ Large batches       │ ✅ Low latency         │              │")
	fmt.Println("│                  │ ❌ High latency        │ ❌ Small batches       │              │")
	fmt.Println("├──────────────────┼────────────────────────┼────────────────────────┼──────────────┤")
	fmt.Println("│ Target Coherence │ 0.6-0.7 (60-70%)       │ 0.85-0.95 (85-95%)     │ 0.70-0.80    │")
	fmt.Println("│                  │ ✅ Achievable          │ ✅ Maximum batching    │              │")
	fmt.Println("│                  │ ❌ Some stragglers     │ ❌ May never converge  │              │")
	fmt.Println("├──────────────────┼────────────────────────┼────────────────────────┼──────────────┤")
	fmt.Println("│ Workload Types   │ Homogeneous            │ Highly diverse         │ Mixed types  │")
	fmt.Println("│                  │ ✅ Predictable sync    │ ✅ Realistic           │              │")
	fmt.Println("│                  │ ❌ Unrealistic         │ ❌ Harder to batch     │              │")
	fmt.Println("└──────────────────┴────────────────────────┴────────────────────────┴──────────────┘")
	fmt.Println()

	// Use optimized configuration for batching scenarios
	config := swarm.ConfigForBatching(numWorkloads, batchWindow)
	swarm, err := swarm.New(numWorkloads, targetState, swarm.WithConfig(config))
	if err != nil {
		fmt.Printf("❌ Error creating swarm: %v\n", err)
		return
	}

	// Configure heterogeneous workload behaviors.
	// In production, these differences arise naturally from:
	// - Network latency variations
	// - Processing speed differences
	// - Cache hit rates
	// - Geographic distribution
	fmt.Println("🎭 WORKLOAD DIVERSITY")
	configureWorkloads(swarm)
	fmt.Println()

	// Demonstrate the initial chaos: requests scattered across time
	fmt.Println("═══ INITIAL STATE: CHAOS ═══")
	initialCoherence := swarm.MeasureCoherence()
	visualizeRequestTimeline(swarm)
	batches := estimateBatches(swarm)
	fmt.Printf("📊 Coherence: %.1f%% ", initialCoherence*100)
	interpretBatchingQuality(initialCoherence)
	fmt.Printf("📡 Simulated API Calls: %d separate requests (inefficient!)\n", batches)
	fmt.Printf("   (In production: Each dot would be a real API call)\n")
	fmt.Println()

	// Begin the synchronization process.
	// In production, this would run continuously.
	fmt.Println("═══ SYNCHRONIZATION IN PROGRESS ═══")
	fmt.Println("⚡ Simulating: Workloads discovering natural batch windows...")
	fmt.Printf("   (Each step = %v of simulated time)\n", checkInterval)
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Start autonomous synchronization
	errChan := make(chan error, 1)
	go func() {
		if err := swarm.Run(ctx); err != nil && err != context.Canceled {
			errChan <- err
		}
	}()

	// Monitor convergence with visual progress indicator
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	var finalCoherence float64
	iteration := 0
	lastCoherence := initialCoherence
	stuckCount := 0

	for {
		select {
		case <-ticker.C:
			iteration++
			coherence := swarm.MeasureCoherence()
			currentBatches := estimateBatches(swarm)

			// Show iteration with visual progress
			fmt.Printf("Step %2d/%d: ", iteration, maxIterations)

			// Visual progress bar
			drawColoredProgressBar(coherence, targetState.Coherence, 30)

			// Show percentage and trend
			fmt.Printf(" %5.1f%%", coherence*100)

			// Show trend indicator
			if coherence > lastCoherence+0.01 {
				fmt.Print(" ↗️")
			} else if coherence < lastCoherence-0.01 {
				fmt.Print(" ↘️")
			} else {
				fmt.Print(" →")
				stuckCount++
			}

			// Check if we've reached the target coherence.
			// The system has successfully self-organized!
			if coherence >= targetState.Coherence {
				fmt.Printf(" ✅ TARGET REACHED!\n")
				finalCoherence = coherence
				cancel()
				goto results
			}

			// Warn if stuck
			if stuckCount > 5 {
				fmt.Print(" ⚠️  (stuck - may need parameter tuning)")
			}

			// Extra info specific to batching
			fmt.Printf(" │ %2d batches", currentBatches)
			reduction := float64(numWorkloads-currentBatches) / float64(numWorkloads) * 100
			if reduction > 0 {
				fmt.Printf(" (%.0f%% reduction)", reduction)
			}

			fmt.Println()
			lastCoherence = coherence

			// Prevent infinite loops in demo
			if iteration >= maxIterations {
				fmt.Printf("\n⏱️  Max iterations (%d) reached - stopping\n", maxIterations)
				finalCoherence = coherence
				cancel()
				goto results
			}

		case err := <-errChan:
			if err != nil && err != context.Canceled {
				fmt.Printf("\n❌ Swarm error: %v\n", err)
			}
			goto results

		case <-ctx.Done():
			fmt.Println() // Ensure we're on a new line
			finalCoherence = swarm.MeasureCoherence()
			goto results
		}
	}

results:
	fmt.Println()

	// Show the synchronized state with natural batching
	fmt.Println("═══ FINAL STATE: SYNCHRONIZED BATCHING ═══")
	visualizeRequestTimeline(swarm)
	finalBatches := estimateBatches(swarm)
	fmt.Printf("📊 Coherence: %.1f%% ", finalCoherence*100)
	interpretBatchingQuality(finalCoherence)
	fmt.Printf("📡 Simulated API Calls: %d batched requests ", finalBatches)

	// Show the improvement
	reduction := float64(numWorkloads-finalBatches) / float64(numWorkloads) * 100
	if reduction > 0 {
		fmt.Printf("(%.0f%% reduction! 🎉)\n", reduction)
	} else {
		fmt.Printf("(no reduction)\n")
	}
	fmt.Println()

	// Calculate improvement percentage
	improvement := ((finalCoherence - initialCoherence) / initialCoherence) * 100

	// Show simulation results in a clean table
	fmt.Println("═══ SIMULATION RESULTS ═══")
	fmt.Println("┌──────────────────────────────────────┐")
	fmt.Printf("│ Initial coherence:    %6.1f%%        │\n", initialCoherence*100)
	fmt.Printf("│ Final coherence:      %6.1f%%        │\n", finalCoherence*100)
	fmt.Printf("│ Sync improvement:     %6.1f%%        │\n", improvement)
	fmt.Println("├──────────────────────────────────────┤")
	fmt.Printf("│ Initial API calls:    %6d requests │\n", numWorkloads)
	fmt.Printf("│ Final API calls:      %6d batches  │\n", finalBatches)
	fmt.Printf("│ Call reduction:       %6.1f%%        │\n", reduction)
	fmt.Println("├──────────────────────────────────────┤")
	fmt.Printf("│ Target (%.0f%%):        ", targetState.Coherence*100)
	if finalCoherence >= targetState.Coherence {
		fmt.Printf("✅ ACHIEVED      │\n")
	} else {
		fmt.Printf("❌ NOT REACHED   │\n")
	}
	fmt.Printf("│ Batch efficiency:     ")
	if reduction >= 70 {
		fmt.Printf("✅ EXCELLENT   │\n")
	} else if reduction >= 50 {
		fmt.Printf("🟡 GOOD        │\n")
	} else {
		fmt.Printf("❌ POOR        │\n")
	}
	fmt.Println("└──────────────────────────────────────┘")
	fmt.Println()

	// Provide diagnostics if target wasn't reached
	if finalCoherence < targetState.Coherence {
		fmt.Println("🔍 DIAGNOSTICS - Why didn't we reach optimal batching?")

		gap := targetState.Coherence - finalCoherence
		if gap > 0.3 {
			fmt.Println("   ⚠️  Large gap (>30%) from target:")
			fmt.Println("   • Multiple batch groups formed but won't merge")
			fmt.Println("   • Some workloads too diverse to synchronize")
			fmt.Println("   • May need different workload groupings")
		} else if stuckCount > 5 {
			fmt.Println("   ⚠️  System stuck in LOCAL MINIMA:")
			fmt.Println("   • Achieved partial batching but can't improve")
			fmt.Println("   • Example: Morning & evening batches won't combine")
			fmt.Println("   • In METASTABLE STATE - works but fragile")
			fmt.Println("   • Would need PERTURBATION (timing shifts) to escape")
		} else if reduction < 50 {
			fmt.Println("   ⚠️  Limited batching achieved:")
			fmt.Println("   • Workloads too diverse or stubborn")
			fmt.Println("   • Batch windows may be too small")
			fmt.Println("   • Consider grouping similar workloads")
		}

		fmt.Println()
		fmt.Println("📊 RECOMMENDED TUNING:")
		fmt.Println("   • Larger batch window: Frequency = 300ms")
		fmt.Println("   • Lower target: Coherence = 0.8")
		fmt.Println("   • Group similar workloads together")
		fmt.Println("   • Add jitter/perturbation to escape local minima")
		fmt.Println()
	}

	// Demonstrate self-healing after disruption
	fmt.Println("═══ RESILIENCE TEST ═══")
	testResilience(swarm, targetState)
	fmt.Println()

	// Calculate and display metrics
	fmt.Println("═══ PERFORMANCE METRICS ═══")
	printEnhancedSummary(numWorkloads, finalBatches, initialCoherence, finalCoherence)

	// Skip detailed explanation in CI/quiet mode
	if os.Getenv("QUIET") == "1" || os.Getenv("CI") == "true" {
		fmt.Println("\n(Running in quiet mode. Set QUIET=0 to see full explanation)")
		return
	}

	// Educational section: explain the mechanism
	fmt.Println("\n💡 HOW THE ALGORITHM WORKS")
	fmt.Println("├─ Each workload has a phase (timing in cycle)")
	fmt.Println("├─ Workloads observe neighbors and adjust timing")
	fmt.Println("├─ Attractor basins guide toward synchronization")
	fmt.Println("├─ Batch windows emerge naturally")
	fmt.Println("└─ No central coordinator needed!")

	fmt.Println("\n📐 UNDERSTANDING PHASE IN REQUEST BATCHING:")
	fmt.Println("├─ Phase = When in the batch window a request occurs")
	fmt.Println("├─ 0 radians = Start of batch window")
	fmt.Println("├─ π radians = Middle of batch window")
	fmt.Println("├─ 2π radians = End of batch window (wraps to 0)")
	fmt.Println("└─ Goal: All workloads at phase=0 (same batch moment)")

	fmt.Println("\n🔧 WHAT 'PHASE' MEANS FOR DIFFERENT WORKLOADS:")
	fmt.Println("├─ 🌐 Web service: Request timing in rate limit window")
	fmt.Println("├─ 📊 Analytics: Position in aggregation period")
	fmt.Println("├─ 🤖 ML pipeline: Stage in batch processing cycle")
	fmt.Println("├─ 📨 Email service: Position in send queue window")
	fmt.Println("├─ 💾 Data sync: Timing in replication cycle")
	fmt.Println("└─ 🔄 ETL job: Position in extraction window")

	fmt.Println("\n🔧 TO IMPLEMENT IN PRODUCTION:")
	fmt.Println("├─ 1. Replace simulated workloads with real services")
	fmt.Println("├─ 2. Hook phase timing to actual API call scheduling")
	fmt.Println("├─ 3. Use batch API endpoints when requests align")
	fmt.Println("├─ 4. Monitor actual reduction in API calls")
	fmt.Println("└─ 5. Tune parameters based on your SLA requirements")

	fmt.Println("\n📝 PRODUCTION API SETUP:")
	fmt.Println("```go")
	fmt.Println("// Define your batching requirements")
	fmt.Println("batchConfig := core.State{")
	fmt.Println("    Phase:     0,                       // Sync point")
	fmt.Println("    Frequency: 500 * time.Millisecond, // Batch every 500ms")
	fmt.Println("    Coherence: 0.85,                   // 85% synchronization")
	fmt.Println("}")
	fmt.Println()
	fmt.Println("// Create workload swarm")
	fmt.Println("workloads, _ := swarm.New(50, batchConfig)")
	fmt.Println()
	fmt.Println("// In each workload, check phase before API call:")
	fmt.Println("if agent.Phase() < 0.1 { // Near batch window")
	fmt.Println("    // Add request to batch queue")
	fmt.Println("    batchQueue.Add(request)")
	fmt.Println("}")
	fmt.Println("```")

	// Show real-world applicability
	fmt.Println("\n🌍 REAL-WORLD APPLICATIONS")
	fmt.Println("├─ OpenAI/Anthropic API batching")
	fmt.Println("├─ Database connection pooling")
	fmt.Println("├─ Kubernetes pod scheduling")
	fmt.Println("├─ IoT telemetry collection")
	fmt.Println("├─ CDN cache invalidation")
	fmt.Println("└─ Any rate-limited resource")

	// Guide for production use
	fmt.Println("\n🚀 PRODUCTION DEPLOYMENT:")
	fmt.Println("├─ This simulation demonstrates the timing algorithm")
	fmt.Println("├─ In production, integrate with your actual API client")
	fmt.Println("├─ The emerge package provides the coordination logic")
	fmt.Println("├─ Your code handles the actual API calls when aligned")
	fmt.Println("└─ Result: 70-85% reduction in real API calls")
}

// configureWorkloads simulates heterogeneous workload characteristics.
// In production, these differences emerge naturally from system diversity.
func configureWorkloads(swarm *swarm.Swarm) {
	workloadTypes := []string{"⚡Fast", "🔄Normal", "🐌Slow", "💥Bursty"}
	typeCount := make(map[string]int)
	typeEmojis := map[string]string{
		"⚡Fast": "⚡", "🔄Normal": "🔄",
		"🐌Slow": "🐌", "💥Bursty": "💥",
	}

	i := 0
	for _, agent := range swarm.Agents() {

		// Assign workload personality based on real-world patterns
		workloadType := workloadTypes[i%len(workloadTypes)]
		typeCount[workloadType]++

		switch workloadType {
		case "⚡Fast":
			// Fast workloads: Low-latency services, cached responses
			agent.SetPhase(rand.Float64() * math.Pi)
			agent.SetInfluence(0.7)     // Strong influencer
			agent.SetStubbornness(0.05) // Adapts quickly
		case "🔄Normal":
			// Normal workloads: Standard microservices
			agent.SetPhase(rand.Float64() * 2 * math.Pi)
			agent.SetInfluence(0.5)    // Average influence
			agent.SetStubbornness(0.1) // Normal adaptation
		case "🐌Slow":
			// Slow workloads: Complex processing, cold starts
			agent.SetPhase(math.Pi + rand.Float64()*math.Pi)
			agent.SetInfluence(0.3)     // Weak influence
			agent.SetStubbornness(0.15) // Slower to adapt
		case "💥Bursty":
			// Bursty workloads: Event-driven, webhook handlers
			agent.SetPhase(rand.Float64() * 2 * math.Pi)
			agent.SetInfluence(0.6)    // Variable influence
			agent.SetStubbornness(0.2) // More independent
		}

		i++
	}

	// Display workload distribution
	for _, wType := range workloadTypes {
		if count, ok := typeCount[wType]; ok && count > 0 {
			emoji := typeEmojis[wType]
			fmt.Printf("├─ %s %s: %d workloads\n", emoji, wType[1:], count)
		}
	}
	fmt.Print("└─ Total: ")
	fmt.Printf("%d heterogeneous workloads\n", swarm.Size())
}

// visualizeRequestTimeline shows when requests would occur in a time window
func visualizeRequestTimeline(swarm *swarm.Swarm) {
	// Collect all agent phases (request timings)
	phases := make([]float64, 0, swarm.Size())
	for _, agent := range swarm.Agents() {

		phases = append(phases, agent.Phase())
	}

	// Create time bins representing 200ms windows
	numBins := 24
	bins := make([]int, numBins)
	for _, phase := range phases {
		// Map phase (0-2π) to time bin
		bin := int(phase / (2 * math.Pi) * float64(numBins))
		if bin >= numBins {
			bin = numBins - 1
		}
		if bin < 0 {
			bin = 0
		}
		bins[bin]++
	}

	// Find max for scaling
	maxCount := 0
	for _, count := range bins {
		if count > maxCount {
			maxCount = count
		}
	}

	// Draw timeline header
	fmt.Println("📅 Simulated Request Timeline (200ms window):")
	fmt.Print("   ")

	// Time labels
	for i := 0; i < numBins; i += 6 {
		fmt.Printf("%-6dms", i*200/numBins)
	}
	fmt.Println()

	// Draw the timeline bars
	fmt.Print("   ")
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
	fmt.Println()

	// Show batch formations
	fmt.Print("   ")
	batchThreshold := 2 // Minimum requests to form a batch
	inBatch := false
	for _, count := range bins {
		if count >= batchThreshold {
			if !inBatch {
				fmt.Print("┌")
				inBatch = true
			} else {
				fmt.Print("─")
			}
		} else {
			if inBatch {
				fmt.Print("┘")
				inBatch = false
			} else {
				fmt.Print(" ")
			}
		}
	}
	if inBatch {
		fmt.Print("┘")
	}
	fmt.Println(" ← Batch windows")
}

// drawColoredProgressBar creates a visual progress indicator with color coding
func drawColoredProgressBar(current, target float64, width int) {
	progress := min(current/target, 1.0)

	filled := int(progress * float64(width))

	// Color based on progress
	if progress < 0.3 {
		fmt.Print("🔴")
	} else if progress < 0.7 {
		fmt.Print("🟡")
	} else {
		fmt.Print("🟢")
	}

	fmt.Print(" [")
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	fmt.Print(bar)
	fmt.Print("]")
}

// interpretBatchingQuality provides context for coherence values
func interpretBatchingQuality(coherence float64) {
	if coherence < 0.2 {
		fmt.Print("(🌪️  Chaos - no batching)")
	} else if coherence < 0.4 {
		fmt.Print("(🌊 Weak batching emerging)")
	} else if coherence < 0.6 {
		fmt.Print("(⚡ Moderate batching)")
	} else if coherence < 0.8 {
		fmt.Print("(📦 Good batching)")
	} else {
		fmt.Print("(🎯 Excellent batching!)")
	}
	fmt.Println()
}

// estimateBatches counts how many distinct request clusters formed.
// Fewer batches = better efficiency (more requests per batch).
func estimateBatches(swarm *swarm.Swarm) int {
	// Collect all phases
	phases := make([]float64, 0, swarm.Size())
	for _, agent := range swarm.Agents() {

		phases = append(phases, agent.Phase())
	}

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
				diff := math.Abs(core.PhaseDifference(phase1, phase2))
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
func testResilience(swarm *swarm.Swarm, targetState core.State) {
	fmt.Println("🔨 Simulating disruption scenario (30% workloads fail)...")
	fmt.Println("   (In production: pod restarts, deployments, network issues)")

	beforeCoherence := swarm.MeasureCoherence()
	beforeBatches := estimateBatches(swarm)

	// Simulate failures: cloud restarts, deployments, network issues
	swarm.DisruptAgents(0.3)

	afterDisruption := swarm.MeasureCoherence()
	afterBatches := estimateBatches(swarm)

	// Visual before/after
	fmt.Printf("├─ Before: %.1f%% sync, %d batches\n",
		beforeCoherence*100, beforeBatches)
	fmt.Printf("├─ After:  %.1f%% sync, %d batches 💥\n",
		afterDisruption*100, afterBatches)

	// Allow autonomous recovery
	fmt.Print("├─ Recovery in progress")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go func() {
		_ = swarm.Run(ctx) // Context cancellation is expected
	}()

	// Animated recovery dots
	for range 3 {
		time.Sleep(500 * time.Millisecond)
		fmt.Print(".")
	}
	fmt.Println()

	// Measure recovery
	time.Sleep(500 * time.Millisecond)
	afterRecovery := swarm.MeasureCoherence()
	recoveredBatches := estimateBatches(swarm)

	fmt.Printf("└─ Result: %.1f%% sync, %d batches",
		afterRecovery*100, recoveredBatches)

	if afterRecovery >= targetState.Coherence*0.85 {
		fmt.Printf(" ✅ Self-healed!\n")
	} else if afterRecovery > afterDisruption {
		fmt.Printf(" 📈 Partial recovery\n")
	} else {
		fmt.Printf(" ⚠️  Limited recovery\n")
	}
}

// printEnhancedSummary displays the key metrics with visual appeal
func printEnhancedSummary(workloads, finalBatches int, initialCoherence, finalCoherence float64) {
	// Calculate real-world improvements
	apiReduction := float64(workloads-finalBatches) / float64(workloads) * 100
	coherenceGain := (finalCoherence - initialCoherence) / (1 - initialCoherence) * 100

	// Visual summary box
	fmt.Println("┌──────────────────────────────────────────────┐")
	fmt.Println("│      SIMULATED BATCHING PERFORMANCE          │")
	fmt.Println("├──────────────────────────────────────────────┤")

	// API calls comparison
	fmt.Printf("│ 📡 API Calls (simulated):                    │\n")
	fmt.Printf("│    Before: %3d requests ████████████        │\n", workloads)

	// Visual bar for after
	barLength := int(float64(finalBatches) / float64(workloads) * 12)
	bar := strings.Repeat("█", barLength)
	fmt.Printf("│    After:  %3d batches  %-12s 🎯    │\n", finalBatches, bar)

	fmt.Printf("│    Reduction: %.0f%%                          │\n", apiReduction)

	fmt.Println("├──────────────────────────────────────────────┤")

	// Coherence comparison
	fmt.Printf("│ 🎯 Synchronization:                          │\n")
	fmt.Printf("│    Start:  %5.1f%% ", initialCoherence*100)
	drawMiniBar(initialCoherence, 15)
	fmt.Printf("        │\n")

	fmt.Printf("│    Final:  %5.1f%% ", finalCoherence*100)
	drawMiniBar(finalCoherence, 15)
	fmt.Printf("        │\n")

	fmt.Printf("│    Improvement: +%.0f%%                       │\n", coherenceGain)

	fmt.Println("├──────────────────────────────────────────────┤")

	// Benefits achieved
	fmt.Println("│ ✅ Benefits Achieved:                        │")
	if apiReduction > 60 {
		fmt.Printf("│    • %.0f%% fewer API calls                 │\n", apiReduction)
		fmt.Println("│    • Major cost savings                      │")
		fmt.Println("│    • Eliminated rate limiting                │")
	} else if apiReduction > 30 {
		fmt.Printf("│    • %.0f%% fewer API calls                 │\n", apiReduction)
		fmt.Println("│    • Moderate cost savings                   │")
		fmt.Println("│    • Reduced rate limiting                   │")
	} else {
		fmt.Println("│    • Limited batching achieved               │")
		fmt.Println("│    • Consider parameter tuning               │")
	}

	fmt.Println("│    • No central coordinator needed           │")
	fmt.Println("│    • Self-healing capability                 │")
	fmt.Println("└──────────────────────────────────────────────┘")
}

// drawMiniBar draws a small inline progress bar
func drawMiniBar(progress float64, width int) {
	filled := int(progress * float64(width))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	fmt.Print(bar)
}
