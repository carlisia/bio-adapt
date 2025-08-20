// Package main demonstrates how bio-inspired synchronization solves real-world
// API batching challenges without centralized coordination.
package main

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/swarm"
	"github.com/carlisia/bio-adapt/internal/analysis"
	"github.com/carlisia/bio-adapt/internal/display"
	"github.com/jedib0t/go-pretty/v6/table"
)

func main() {
	// Configuration
	numAgents := 20
	maxIterations := 15
	checkInterval := 500 * time.Millisecond
	timeout := 15 * time.Second
	targetState := core.State{
		Phase:     0,
		Frequency: 200 * time.Millisecond, // Batch window size
		Coherence: 0.75,                   // Target synchronization
	}

	// 1. Purpose and context
	display.Banner("LLM API REQUEST BATCHING DEMO")

	display.Section("Purpose and Context")
	fmt.Println("This demo shows timing coordination for API request batching.")
	fmt.Println("Bio-inspired synchronization creates natural batch windows.")
	fmt.Println("No central coordinator or shared state required.")
	fmt.Println("Note: Simulates timing only - no actual API calls made.")
	fmt.Println()

	// 2. What to observe
	display.Section("What to Observe")
	display.Bullet(
		"20 independent agents start with random request timing",
		"Agents gradually align their request phases",
		"Natural batch windows emerge as phases converge",
		fmt.Sprintf("Target: %.0f%% synchronization for optimal batching", targetState.Coherence*100),
		"Typical 70-85% reduction in total API calls",
	)
	fmt.Println()

	// 3. Key concepts
	display.Section("Key Concepts")
	fmt.Println("REQUEST BATCHING = Multiple requests in single API call")
	fmt.Println("• Reduces rate limiting pressure")
	fmt.Println("• Lower per-request overhead")
	fmt.Println("• Cost savings through batch pricing")
	fmt.Println()

	fmt.Println("BATCH WINDOW = Time period for collecting requests")
	fmt.Printf("• Window size: %dms\n", targetState.Frequency.Milliseconds())
	fmt.Println("• Agents aligning = requests clustering")
	fmt.Println("• Higher coherence = better batching")
	fmt.Println()

	// 4. Simulation setup
	display.Section("Simulation Setup")
	fmt.Printf("• Agents: %d independent workloads\n", numAgents)
	fmt.Printf("• Batch window: %v\n", targetState.Frequency)
	fmt.Printf("• Target coherence: %.0f%%\n", targetState.Coherence*100)
	fmt.Printf("• Max iterations: %d\n", maxIterations)
	fmt.Printf("• Check interval: %v\n", checkInterval)
	fmt.Println("• Scenario: Each agent represents a service needing LLM API access")
	fmt.Println()

	// 5. Parameter tradeoffs (using go-pretty table)
	display.Section("Parameter Tradeoffs")
	t := display.NewTable()
	t.AppendHeader(table.Row{"Parameter", "Lower Value", "Higher Value", "Sweet Spot"})
	t.AppendRows([]table.Row{
		{"Agent Count", "5-10 (simple system)", "50-100 (complex system)", "15-30"},
		{"Batch Window", "100ms (low latency)", "500ms (max batching)", "200-300ms"},
		{"Target Coherence", "0.6 (loose batching)", "0.9 (tight batching)", "0.7-0.8"},
		{"Update Rate", "Fast (responsive)", "Slow (stable)", "Balanced"},
	})
	t.Render()
	fmt.Println()

	// Create swarm
	s, err := swarm.New(numAgents, targetState)
	if err != nil {
		fmt.Printf("Error: failed to create swarm: %v\n", err)
		return
	}

	// 6. Run loop and monitoring
	display.Section("Run Loop and Monitoring")

	// Initial state
	initialCoherence := s.MeasureCoherence()
	fmt.Println("Initial Request Distribution:")
	visualizeRequestTimeline(s, targetState.Frequency)
	fmt.Printf("Batching Quality: %.1f%% %s\n", initialCoherence*100,
		analysis.DescribeSyncQuality(initialCoherence, "batch"))

	fmt.Printf("Current API calls: %d (no batching)\n\n", numAgents)

	// Start synchronization
	fmt.Println("Starting batch alignment...")
	fmt.Printf("(Each step = %v of coordination)\n\n", checkInterval)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		if err := s.Run(ctx); err != nil {
			errChan <- err
		}
	}()

	// Monitor progress
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	iterations := 0
	lastCoherence := initialCoherence
	stuckCount := 0

	for {
		select {
		case <-ticker.C:
			iterations++
			coherence := s.MeasureCoherence()
			batches := countBatches(s, 3)

			fmt.Printf("Step %2d/%d: ", iterations, maxIterations)
			display.DrawProgressBar(coherence, targetState.Coherence, 30)
			fmt.Printf(" %5.1f%%", coherence*100)
			fmt.Printf(" (%d batches)", batches)

			// Trend
			switch {
			case coherence > lastCoherence+0.01:
				fmt.Print(" [improving]")
			case coherence < lastCoherence-0.01:
				fmt.Print(" [degrading]")
			default:
				fmt.Print(" [stable]")
				stuckCount++
			}

			if coherence >= targetState.Coherence {
				if display.UseEmoji() {
					fmt.Print(" ✅ TARGET REACHED!")
				} else {
					fmt.Print(" [OK] TARGET REACHED!")
				}
				fmt.Println()
				goto done
			}

			if stuckCount > 5 {
				fmt.Print(" (plateau)")
			}
			fmt.Println()

			lastCoherence = coherence

			if iterations >= maxIterations {
				fmt.Printf("\nMax iterations reached\n")
				goto done
			}

		case err := <-errChan:
			fmt.Printf("\nSwarm error: %v\n", err)
			goto done

		case <-ctx.Done():
			fmt.Println("\nTimeout reached")
			goto done
		}
	}

done:
	// 7. Diagnostics and fixes
	finalCoherence := s.MeasureCoherence()
	finalBatches := countBatches(s, 3)
	improvement := ((finalCoherence - initialCoherence) / initialCoherence) * 100

	display.Section("Diagnostics and Fixes")
	stats := analysis.RunStats{
		Initial:    initialCoherence,
		Final:      finalCoherence,
		Iter:       iterations,
		MaxIter:    maxIterations,
		StuckCount: stuckCount,
	}

	summary, suggestions := analysis.Diagnose(stats, targetState.Coherence, "batch")
	fmt.Println(summary)
	if len(suggestions) > 0 {
		fmt.Println("\nSuggested optimizations:")
		for _, s := range suggestions {
			fmt.Printf("• %s\n", s)
		}
	}
	fmt.Println()

	// 8. Results and interpretation
	display.Section("Results and Interpretation")
	display.PrintResultsSummary(
		initialCoherence,
		finalCoherence,
		targetState.Coherence,
		improvement,
		display.WithBatchReduction(numAgents, finalBatches),
	)

	// Final visualization
	fmt.Println("\nFinal Request Distribution:")
	visualizeRequestTimeline(s, targetState.Frequency)
	fmt.Printf("Batching Quality: %s\n", analysis.DescribeSyncQuality(finalCoherence, "batch"))
	fmt.Println()

	// Show batch efficiency
	if finalBatches < numAgents {
		reduction := float64(numAgents-finalBatches) / float64(numAgents) * 100
		fmt.Printf("Batch Efficiency: %.0f%% reduction in API calls\n", reduction)
		fmt.Printf("Cost Savings: Estimated %.0f%% lower API costs\n", reduction*0.7)
	}
	fmt.Println()

	// 9. Real-world mappings
	display.Section("Real-World Mappings")
	fmt.Println("This batching pattern applies to:")
	display.Bullet(
		"OpenAI/Anthropic API calls from multiple services",
		"Database writes from distributed workers",
		"Message queue batch processing",
		"Elasticsearch bulk indexing",
		"Cloud storage batch uploads",
		"Monitoring metric aggregation",
	)
	fmt.Println("\nProduction considerations:")
	display.Bullet(
		"Network topology affects convergence speed",
		"Partial sync (70-80%) often better than perfect sync",
		"Monitor batch size vs latency tradeoff",
		"Use circuit breakers for API failures",
		"Consider priority queues for urgent requests",
	)
	fmt.Println()

	// 10. API/how to apply
	display.Section("How to Apply in Your System")
	fmt.Println("1. Wrap your API client with an Agent")
	fmt.Println("2. Set batch window (frequency) based on latency tolerance")
	fmt.Println("3. Configure target coherence (0.7-0.8 recommended)")
	fmt.Println("4. Connect agents via service discovery or mesh")
	fmt.Println("5. Collect requests when phases align")
	fmt.Println()
	fmt.Println("Example implementation:")
	fmt.Println("  agent := emerge.NewAgent(\"api-worker-1\")")
	fmt.Println("  swarm.AddAgent(agent)")
	fmt.Println("  ")
	fmt.Println("  // In request handler")
	fmt.Println("  if agent.IsNearPhaseZero() {")
	fmt.Println("      batch := collectPendingRequests()")
	fmt.Println("      response := llmClient.BatchCall(batch)")
	fmt.Println("      distributResponses(response)")
	fmt.Println("  }")
	fmt.Println()
	fmt.Println("Benefits over traditional approaches:")
	display.Bullet(
		"No central batch coordinator needed",
		"Self-healing if services restart",
		"Works across languages/platforms",
		"Scales horizontally without configuration",
		"Natural load distribution",
	)
}

// visualizeRequestTimeline shows request distribution over time windows.
func visualizeRequestTimeline(s *swarm.Swarm, frequency time.Duration) {
	agents := s.Agents()
	phases := make([]float64, 0, len(agents))
	for _, agent := range agents {
		phases = append(phases, agent.Phase())
	}

	// Create 10 time bins for the batch window
	bins := display.BinPhases(phases, 10)
	msPerWindow := int(frequency.Milliseconds() / 10)
	display.PrintTimeline(bins, msPerWindow, 3)
}

// countBatches estimates number of batches based on phase clustering.
func countBatches(s *swarm.Swarm, threshold int) int {
	agents := s.Agents()
	phases := make([]float64, 0, len(agents))
	for _, agent := range agents {
		phases = append(phases, agent.Phase())
	}

	// Bin phases
	bins := display.BinPhases(phases, 10)

	// Count bins that would form batches
	batches := 0
	for _, count := range bins {
		if count >= threshold {
			batches++
		} else if count > 0 {
			batches += count // Individual requests
		}
	}

	// Ensure at least 1 batch
	if batches == 0 && len(agents) > 0 {
		batches = 1
	}

	// Cap at reasonable minimum
	minBatches := int(math.Ceil(float64(len(agents)) / 10.0))
	if batches < minBatches {
		batches = minBatches
	}

	return batches
}
