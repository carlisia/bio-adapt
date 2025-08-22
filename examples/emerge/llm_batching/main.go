// Package main demonstrates how goal-directed synchronization solves real-world
// API batching challenges without centralized coordination.
package main

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/swarm"
	"github.com/carlisia/bio-adapt/internal/analysis"
	"github.com/carlisia/bio-adapt/internal/display"
)

// batchingConfig holds the batching simulation configuration.
type batchingConfig struct {
	numAgents     int
	maxIterations int
	checkInterval time.Duration
	timeout       time.Duration
	targetState   core.State
}

// setupBatchingConfig creates the configuration for the batching demo.
func setupBatchingConfig() batchingConfig {
	return batchingConfig{
		numAgents:     20,
		maxIterations: 15,
		checkInterval: 500 * time.Millisecond,
		timeout:       15 * time.Second,
		targetState: core.State{
			Phase:     0,
			Frequency: 200 * time.Millisecond, // Batch window size
			Coherence: 0.75,                   // Target synchronization
		},
	}
}

// printBatchingIntro prints the introduction and setup information.
func printBatchingIntro(config batchingConfig) {
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
		fmt.Sprintf("Target: %.0f%% synchronization for optimal batching", config.targetState.Coherence*100),
		"Typical 70-85% reduction in total API calls",
	)
	fmt.Println()

	// 3. Key concepts
	printKeyConcepts(config)

	// 4. Simulation setup
	printSimulationSetup(config)

	// 5. Parameter tradeoffs
	printParameterTradeoffs()
}

// printKeyConcepts prints the key concepts section.
func printKeyConcepts(config batchingConfig) {
	display.Section("Key Concepts")
	fmt.Println("REQUEST BATCHING = Multiple requests in single API call")
	fmt.Println("• Reduces rate limiting pressure")
	fmt.Println("• Lower per-request overhead")
	fmt.Println("• Cost savings through batch pricing")
	fmt.Println()

	fmt.Println("BATCH WINDOW = Time period for collecting requests")
	fmt.Printf("• Window size: %dms\n", config.targetState.Frequency.Milliseconds())
	fmt.Println("• Agents aligning = requests clustering")
	fmt.Println("• Higher coherence = better batching")
	fmt.Println()
}

// printSimulationSetup prints the simulation setup information.
func printSimulationSetup(config batchingConfig) {
	display.Section("Simulation Setup")
	fmt.Printf("• Agents: %d independent workloads\n", config.numAgents)
	fmt.Printf("• Batch window: %v\n", config.targetState.Frequency)
	fmt.Printf("• Target coherence: %.0f%%\n", config.targetState.Coherence*100)
	fmt.Printf("• Max iterations: %d\n", config.maxIterations)
	fmt.Printf("• Check interval: %v\n", config.checkInterval)
	fmt.Println("• Scenario: Each agent represents a service needing LLM API access")
	fmt.Println()
}

// printParameterTradeoffs prints the parameter tradeoffs table.
func printParameterTradeoffs() {
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
}

// monitorBatchingProgress monitors the batching synchronization progress.
func monitorBatchingProgress(
	ctx context.Context,
	s *swarm.Swarm,
	targetState core.State,
	errChan chan error,
	checkInterval time.Duration,
	maxIterations int,
) (int, int) {
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	iterations := 0
	lastCoherence := s.MeasureCoherence()
	stuckCount := 0

	for {
		select {
		case <-ticker.C:
			iterations++

			if done, stuck := processProgressIteration(s, targetState, iterations, maxIterations, lastCoherence, stuckCount); done {
				return iterations, stuck
			} else {
				stuckCount = stuck
				lastCoherence = s.MeasureCoherence()
			}

		case err := <-errChan:
			fmt.Printf("\nSwarm error: %v\n", err)
			return iterations, stuckCount

		case <-ctx.Done():
			fmt.Println("\nTimeout reached")
			return iterations, stuckCount
		}
	}
}

// processProgressIteration processes a single progress iteration.
func processProgressIteration(
	s *swarm.Swarm,
	targetState core.State,
	iterations, maxIterations int,
	lastCoherence float64,
	stuckCount int,
) (bool, int) {
	coherence := s.MeasureCoherence()
	batches := countBatches(s, 3)

	fmt.Printf("Step %2d/%d: ", iterations, maxIterations)
	display.DrawProgressBar(coherence, targetState.Coherence, 30)
	fmt.Printf(" %5.1f%%", coherence*100)
	fmt.Printf(" (%d batches)", batches)

	// Analyze trend
	newStuckCount := analyzeTrend(coherence, lastCoherence, stuckCount)

	// Check if target reached
	if coherence >= targetState.Coherence {
		if display.UseEmoji() {
			fmt.Print(" ✅ TARGET REACHED!")
		} else {
			fmt.Print(" [OK] TARGET REACHED!")
		}
		fmt.Println()
		return true, newStuckCount
	}

	if newStuckCount > 5 {
		fmt.Print(" (plateau)")
	}
	fmt.Println()

	if iterations >= maxIterations {
		fmt.Printf("\nMax iterations reached\n")
		return true, newStuckCount
	}

	return false, newStuckCount
}

// analyzeTrend analyzes the coherence trend.
func analyzeTrend(coherence, lastCoherence float64, stuckCount int) int {
	switch {
	case coherence > lastCoherence+0.01:
		fmt.Print(" [improving]")
		return stuckCount
	case coherence < lastCoherence-0.01:
		fmt.Print(" [degrading]")
		return stuckCount
	default:
		fmt.Print(" [stable]")
		return stuckCount + 1
	}
}

// printDiagnostics prints diagnostics and suggestions.
func printDiagnostics(s *swarm.Swarm, config batchingConfig, initialCoherence float64, iterations, stuckCount int) {
	finalCoherence := s.MeasureCoherence()

	display.Section("Diagnostics and Fixes")
	stats := analysis.RunStats{
		Initial:    initialCoherence,
		Final:      finalCoherence,
		Iter:       iterations,
		MaxIter:    config.maxIterations,
		StuckCount: stuckCount,
	}

	summary, suggestions := analysis.Diagnose(stats, config.targetState.Coherence, "batch")
	fmt.Println(summary)
	if len(suggestions) > 0 {
		fmt.Println("\nSuggested optimizations:")
		for _, s := range suggestions {
			fmt.Printf("• %s\n", s)
		}
	}
	fmt.Println()
}

// printResults prints the results and interpretation.
func printResults(s *swarm.Swarm, config batchingConfig, initialCoherence float64) {
	finalCoherence := s.MeasureCoherence()
	finalBatches := countBatches(s, 3)
	improvement := ((finalCoherence - initialCoherence) / initialCoherence) * 100

	display.Section("Results and Interpretation")
	display.PrintResultsSummary(
		initialCoherence,
		finalCoherence,
		config.targetState.Coherence,
		improvement,
		display.WithBatchReduction(config.numAgents, finalBatches),
	)

	// Final visualization
	fmt.Println("\nFinal Request Distribution:")
	visualizeRequestTimeline(s, config.targetState.Frequency)
	fmt.Printf("Batching Quality: %s\n", analysis.DescribeSyncQuality(finalCoherence, "batch"))
	fmt.Println()

	// Show batch efficiency
	if finalBatches < config.numAgents {
		reduction := float64(config.numAgents-finalBatches) / float64(config.numAgents) * 100
		fmt.Printf("Batch Efficiency: %.0f%% reduction in API calls\n", reduction)
		fmt.Printf("Cost Savings: Estimated %.0f%% lower API costs\n", reduction*0.7)
	}
	fmt.Println()
}

func main() {
	// Configuration
	config := setupBatchingConfig()

	// Print introduction and setup
	printBatchingIntro(config)

	// Create swarm
	s, err := swarm.New(config.numAgents, config.targetState)
	if err != nil {
		fmt.Printf("Error: failed to create swarm: %v\n", err)
		return
	}

	// 6. Run loop and monitoring
	display.Section("Run Loop and Monitoring")

	// Initial state
	initialCoherence := s.MeasureCoherence()
	fmt.Println("Initial Request Distribution:")
	visualizeRequestTimeline(s, config.targetState.Frequency)
	fmt.Printf("Batching Quality: %.1f%% %s\n", initialCoherence*100,
		analysis.DescribeSyncQuality(initialCoherence, "batch"))

	fmt.Printf("Current API calls: %d (no batching)\n\n", config.numAgents)

	// Start synchronization
	fmt.Println("Starting batch alignment...")
	fmt.Printf("(Each step = %v of coordination)\n\n", config.checkInterval)

	ctx, cancel := context.WithTimeout(context.Background(), config.timeout)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		if err := s.Run(ctx); err != nil {
			errChan <- err
		}
	}()

	// Monitor progress
	iterations, stuckCount := monitorBatchingProgress(
		ctx, s, config.targetState, errChan, config.checkInterval, config.maxIterations)
	// 7. Diagnostics and fixes
	printDiagnostics(s, config, initialCoherence, iterations, stuckCount)

	// 8. Results and interpretation
	printResults(s, config, initialCoherence)

	// 9. Real-world mappings
	display.Section("Real-World Mappings")
	fmt.Println("This batching primitive applies to:")
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
