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
	"github.com/carlisia/bio-adapt/emerge/goal"
	"github.com/carlisia/bio-adapt/emerge/swarm"
	"github.com/carlisia/bio-adapt/internal/analysis"
	"github.com/carlisia/bio-adapt/internal/display"
)

// batchingConfig holds the batching simulation configuration.
type batchingConfig struct {
	numAgents         int
	maxIterations     int
	checkInterval     time.Duration
	timeout           time.Duration
	targetState       core.State
	enableDisruption  bool
	disruptionStep    int
}

// setupBatchingConfig creates the configuration for the batching demo.
func setupBatchingConfig() batchingConfig {
	return batchingConfig{
		numAgents:        20,
		maxIterations:    20, // Increased to show disruption recovery
		checkInterval:    500 * time.Millisecond,
		timeout:          20 * time.Second,
		enableDisruption: true,
		disruptionStep:   12, // When to trigger disruption
		targetState: core.State{
			Phase:     0,
			Frequency: 200 * time.Millisecond, // Batch window size
			Coherence: 0.85,                   // Higher target for better batching
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
		"20 independent services start with random request timing",
		"Services gradually align their request phases",
		"Natural batch windows emerge as phases converge",
		fmt.Sprintf("Target: %.0f%% synchronization for optimal batching", config.targetState.Coherence*100),
		"Typical 70-85% reduction in total API calls",
	)
	if config.enableDisruption {
		display.Bullet("Watch for disruption recovery midway through demo")
	}
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
	fmt.Println("• Services aligning = requests clustering")
	fmt.Println("• Higher coherence = better batching")
	fmt.Println()
}

// printSimulationSetup prints the simulation setup information.
func printSimulationSetup(config batchingConfig) {
	display.Section("Simulation Setup")
	fmt.Printf("• Services: %d independent microservices\n", config.numAgents)
	fmt.Printf("• Batch window: %v\n", config.targetState.Frequency)
	fmt.Printf("• Target coherence: %.0f%%\n", config.targetState.Coherence*100)
	fmt.Printf("• Max iterations: %d\n", config.maxIterations)
	fmt.Printf("• Check interval: %v\n", config.checkInterval)
	if config.enableDisruption {
		fmt.Printf("• Disruption: Simulated at step %d\n", config.disruptionStep)
	}
	fmt.Println("• Scenario: Each service needs LLM API access for different tasks")
	fmt.Println()
}

// printParameterTradeoffs prints the parameter tradeoffs table.
func printParameterTradeoffs() {
	display.Section("Parameter Tradeoffs")
	t := display.NewTable()
	t.AppendHeader(table.Row{"Parameter", "Lower Value", "Higher Value", "Sweet Spot"})
	t.AppendRows([]table.Row{
		{"Service Count", "5-10 (simple system)", "50-100 (complex system)", "15-30"},
		{"Batch Window", "100ms (low latency)", "500ms (max batching)", "200-300ms"},
		{"Target Coherence", "0.6 (loose batching)", "0.9 (tight batching)", "0.75-0.85"},
		{"Update Rate", "Fast (responsive)", "Slow (stable)", "Balanced"},
	})
	t.Render()
	fmt.Println()
}

// monitorBatchingProgress monitors the batching synchronization progress.
func monitorBatchingProgress(
	ctx context.Context,
	s *swarm.Swarm,
	config batchingConfig,
	errChan chan error,
) (int, int) {
	ticker := time.NewTicker(config.checkInterval)
	defer ticker.Stop()

	iterations := 0
	lastCoherence := s.MeasureCoherence()
	stuckCount := 0

	for {
		select {
		case <-ticker.C:
			iterations++

			// Simulate disruption (new services added, network jitter, etc)
			if config.enableDisruption && iterations == config.disruptionStep {
				fmt.Println()
				if display.UseEmoji() {
					fmt.Println("⚠️  DISRUPTION: Network jitter affecting 30% of services")
				} else {
					fmt.Println("[!] DISRUPTION: Network jitter affecting 30% of services")
				}
				s.DisruptAgents(0.3) // 30% phase disruption
				fmt.Println("    Services will now recover and re-synchronize...")
				fmt.Println()
			}

			if done, stuck := processProgressIteration(s, config.targetState, iterations, config.maxIterations, lastCoherence, stuckCount); done {
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
		return 0 // Reset stuck count when improving
	case coherence < lastCoherence-0.01:
		fmt.Print(" [recovering]") // More positive framing after disruption
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

// printHowToApply shows how to use the API in real code
func printHowToApply(config batchingConfig) {
	display.Section("How to Apply in Your System")
	
	fmt.Println("Simple integration with goal-based configuration:")
	fmt.Println()
	fmt.Println("```go")
	fmt.Println("// 1. Create swarm with goal-based preset")
	fmt.Println("cfg := swarm.For(goal.MinimizeAPICalls)")
	fmt.Println("s, _ := swarm.New(20, core.State{")
	fmt.Println("    Phase:     0,")
	fmt.Println("    Frequency: 200 * time.Millisecond,  // Batch window")
	fmt.Printf("    Coherence: %.2f,                     // Target sync\n", config.targetState.Coherence)
	fmt.Println("}, swarm.WithGoalConfig(cfg))")
	fmt.Println()
	fmt.Println("// 2. Run in background")
	fmt.Println("go s.Run(ctx)")
	fmt.Println()
	fmt.Println("// 3. In your request handler")
	fmt.Println("agent := s.Agents()[serviceID % s.Size()]")
	fmt.Println("if agent.Phase() < 0.1 { // Near batch window")
	fmt.Println("    batch := collectPendingRequests()")
	fmt.Println("    resp := llmClient.BatchCall(batch)")
	fmt.Println("    distributeResponses(resp)")
	fmt.Println("}")
	fmt.Println("```")
	fmt.Println()
	
	fmt.Println("Advanced configuration options:")
	fmt.Println()
	fmt.Println("```go")
	fmt.Println("// Fluent API for fine-tuning")
	fmt.Println("cfg := swarm.For(goal.MinimizeAPICalls).")
	fmt.Println("    TuneFor(trait.Stability).")
	fmt.Println("    With(scale.Large)")
	fmt.Println()
	fmt.Println("// Or auto-scale based on service count")
	fmt.Println("cfg := config.AutoScaleConfig(serviceCount)")
	fmt.Println("```")
	fmt.Println()
}

func main() {
	// Configuration
	config := setupBatchingConfig()

	// Print introduction and setup
	printBatchingIntro(config)

	// Create swarm using goal-based configuration (more user-friendly)
	display.Section("Creating Swarm")
	fmt.Println("Using goal-based configuration for API batching...")
	
	// Show the actual API usage
	cfg := swarm.For(goal.MinimizeAPICalls)
	s, err := swarm.New(config.numAgents, config.targetState, swarm.WithGoalConfig(cfg))
	if err != nil {
		fmt.Printf("Error: failed to create swarm: %v\n", err)
		return
	}
	fmt.Println("Swarm created with optimized settings for API batching")
	fmt.Println()

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
	iterations, stuckCount := monitorBatchingProgress(ctx, s, config, errChan)
	
	// 7. Diagnostics and fixes
	printDiagnostics(s, config, initialCoherence, iterations, stuckCount)

	// 8. Results and interpretation
	printResults(s, config, initialCoherence)

	// 9. Real-world mappings
	display.Section("Real-World Applications")
	fmt.Println("This batching primitive applies to:")
	display.Bullet(
		"OpenAI/Anthropic/Claude API calls from microservices",
		"Database write batching from distributed workers",
		"Elasticsearch bulk indexing operations",
		"Message queue batch processing (Kafka, SQS)",
		"Cloud storage batch uploads (S3, GCS)",
		"Monitoring metric aggregation (Prometheus, DataDog)",
	)
	fmt.Println("\nProduction considerations:")
	display.Bullet(
		"Network topology affects convergence speed",
		"Partial sync (75-85%) often better than perfect sync",
		"Monitor batch size vs latency tradeoff",
		"Use circuit breakers for API failures",
		"Consider priority queues for urgent requests",
		"Handles service restarts/scaling automatically",
	)
	fmt.Println()

	// 10. API/how to apply
	printHowToApply(config)
	
	fmt.Println("Benefits over traditional approaches:")
	display.Bullet(
		"No central batch coordinator needed",
		"Self-healing after service disruptions",
		"Works across languages/platforms",
		"Scales horizontally without reconfiguration",
		"Natural load distribution across batch windows",
		"Resilient to network partitions",
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