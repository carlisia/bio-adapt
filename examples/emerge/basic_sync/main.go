// Package main demonstrates the fundamental concepts of bio-inspired attractor basin
// synchronization using a minimal example.
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/swarm"
	"github.com/carlisia/bio-adapt/internal/analysis"
	"github.com/carlisia/bio-adapt/internal/display"
	"github.com/jedib0t/go-pretty/v6/table"
)

func main() {
	// Configuration
	swarmSize := 1500 // Standard demo size
	maxIterations := 20
	checkInterval := 500 * time.Millisecond
	timeout := 10 * time.Second
	targetState := core.State{
		Phase:     0,
		Frequency: 250 * time.Millisecond,
		Coherence: 0.65,
	}

	// 1. Purpose and context
	display.Banner("ATTRACTOR BASIN SYNCHRONIZATION DEMO")

	display.Section("Purpose and Context")
	fmt.Println("This demo shows core synchronization algorithm operation.")
	fmt.Println("Based on bio-inspired phase coupling (Kuramoto model).")
	fmt.Println("Pure algorithm demonstration - no external systems involved.")
	fmt.Println()

	// 2. What to observe
	display.Section("What to Observe")
	display.Bullet(
		"Initial coherence is low (agents start with random phases)",
		"Coherence gradually increases as agents influence each other",
		fmt.Sprintf("System converges to target coherence (%.0f%% in this example)", targetState.Coherence*100),
		"Convergence typically occurs within 10 iterations",
	)
	fmt.Println()

	// 3. Key concepts
	display.Section("Key Concepts")
	fmt.Println("COHERENCE = How synchronized the agents are (0-100%)")
	fmt.Println("• 0-20%:  Chaos - no coordination")
	fmt.Println("• 20-40%: Groups forming - multiple rhythms")
	fmt.Println("• 40-60%: Partial coordination - groups merging")
	fmt.Println("• 60-80%: Good synchronization - single dominant rhythm")
	fmt.Println("• 80-100%: Excellent synchronization - unified rhythm")
	fmt.Println()

	fmt.Println("LOCAL MINIMA = Stuck in a 'good enough' state")
	fmt.Println("• Like a ball stuck in a shallow dip")
	fmt.Println("• Example: 2 groups synced internally but not together")
	fmt.Println()

	// 4. Simulation setup
	display.Section("Simulation Setup")
	fmt.Printf("• Agents: %d independent entities\n", swarmSize)
	fmt.Printf("• Target: %.0f%% synchronization\n", targetState.Coherence*100)
	fmt.Printf("• Window: %v oscillation period\n", targetState.Frequency)
	fmt.Printf("• Max iterations: %d (checked every %v)\n", maxIterations, checkInterval)
	fmt.Printf("• Timeout: %v\n", timeout)
	fmt.Println("• Method: Local interactions only (no coordinator)")
	fmt.Println()

	// 5. Parameter tradeoffs (using go-pretty table)
	display.Section("Parameter Tradeoffs")
	t := display.NewTable()
	t.AppendHeader(table.Row{"Parameter", "Lower Value", "Higher Value", "Sweet Spot"})
	t.AppendRows([]table.Row{
		{"Swarm Size", "3-5 (easier to converge)", "10-20 (more realistic)", "6-8"},
		{"Target Coherence", "0.5-0.6 (easy, loose)", "0.8-0.95 (tight, risky)", "0.65-0.75"},
		{"Frequency", "250-500ms (stable, slower)", "50-100ms (fast, jittery)", "150-200ms"},
		{"Check Interval", "1000ms (less CPU)", "100ms (fine monitoring)", fmt.Sprintf("%dms", checkInterval.Milliseconds())},
	})
	t.Render()
	fmt.Println()

	// Create swarm
	config := swarm.AutoScaleConfig(swarmSize)
	s, err := swarm.New(swarmSize, targetState, swarm.WithConfig(config))
	if err != nil {
		fmt.Printf("Error: failed to create swarm: %v\n", err)
		return
	}

	// 6. Run loop and monitoring
	display.Section("Run Loop and Monitoring")

	// Initial state
	initialCoherence := s.MeasureCoherence()
	fmt.Printf("Initial State: ")
	visualizeAgents(s)
	fmt.Printf("Coherence: %.1f%% %s\n", initialCoherence*100,
		analysis.DescribeSyncQuality(initialCoherence, "sync"))
	fmt.Println()

	// Start synchronization
	fmt.Println("Starting synchronization...")
	fmt.Printf("(Each step = %v of algorithm iterations)\n\n", checkInterval)

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

			fmt.Printf("Step %2d/%d: ", iterations, maxIterations)
			display.DrawProgressBar(coherence, targetState.Coherence, 30)
			fmt.Printf(" %5.1f%%", coherence*100)

			// Trend indicator
			if coherence > lastCoherence+0.01 {
				fmt.Print(" [rising]")
			} else if coherence < lastCoherence-0.01 {
				fmt.Print(" [falling]")
			} else {
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
				fmt.Print(" (stuck)")
			}
			fmt.Println()

			lastCoherence = coherence

			if iterations >= maxIterations {
				fmt.Printf("\nMax iterations (%d) reached\n", maxIterations)
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
	improvement := ((finalCoherence - initialCoherence) / initialCoherence) * 100

	display.Section("Diagnostics and Fixes")
	stats := analysis.RunStats{
		Initial:    initialCoherence,
		Final:      finalCoherence,
		Iter:       iterations,
		MaxIter:    maxIterations,
		StuckCount: stuckCount,
	}

	summary, suggestions := analysis.Diagnose(stats, targetState.Coherence, "sync")
	fmt.Println(summary)
	if len(suggestions) > 0 {
		fmt.Println("\nSuggested fixes:")
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
	)

	// Final visualization
	fmt.Println("\nFinal State:")
	visualizeAgents(s)
	fmt.Printf("Quality: %s\n", analysis.DescribeSyncQuality(finalCoherence, "sync"))
	fmt.Println()

	// 9. Real-world mappings
	display.Section("Real-World Mappings")
	fmt.Println("This synchronization pattern applies to:")
	display.Bullet(
		"Microservice coordination",
		"Distributed cache invalidation",
		"Load balancer health checks",
		"Database replica sync",
		"IoT device coordination",
	)
	fmt.Println()

	// 10. API/how to apply
	display.Section("How to Apply in Your System")
	fmt.Println("1. Define your 'phase' (timing/state to synchronize)")
	fmt.Println("2. Set target coherence (how tight sync should be)")
	fmt.Println("3. Configure agent connections (who observes whom)")
	fmt.Println("4. Run swarm.Run(ctx) and monitor convergence")
	fmt.Println("5. Use coherence metric to trigger actions")
	fmt.Println()
	fmt.Println("Example integration:")
	fmt.Println("  if swarm.MeasureCoherence() > 0.7 {")
	fmt.Println("      // Trigger coordinated action")
	fmt.Println("      performBatchOperation()")
	fmt.Println("  }")
}

// visualizeAgents shows agent phase distribution.
func visualizeAgents(s *swarm.Swarm) {
	agents := s.Agents()
	phases := make([]float64, 0, len(agents))
	for _, agent := range agents {
		phases = append(phases, agent.Phase())
	}

	// Create 12 bins for clock positions
	bins := display.BinPhases(phases, 12)
	display.PrintClockBins(bins)
}
