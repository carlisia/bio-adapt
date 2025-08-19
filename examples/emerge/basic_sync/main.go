// Package main demonstrates the fundamental concepts of bio-inspired attractor basin
// synchronization using a minimal example.
//
// This example shows how independent agents can achieve synchronization through
// local interactions, without any central coordinator. The pattern mimics biological
// systems like firefly synchronization or cardiac pacemaker cells.
//
// Key Concepts Demonstrated:
//   - Attractor Basin: A stable state that agents naturally converge toward
//   - Phase Coherence: Measure of how synchronized agents are (0=chaos, 1=perfect sync)
//   - Emergent Behavior: Global synchronization emerges from local rules
//   - Self-Organization: No central control needed
//
// What to Observe:
//  1. Initial coherence is low (agents start with random phases)
//  2. Coherence gradually increases as agents influence each other
//  3. System converges to target coherence (80% in this example)
//  4. Convergence typically occurs within 10 iterations
//
// Try Modifying (if system doesn't converge):
//   - swarmSize: Start with 5-8 agents (fewer = simpler dynamics)
//   - targetState.Coherence: Try 0.6-0.7 (lower = easier to achieve)
//   - targetState.Frequency: Try 150-300ms (slower = gentler convergence)
//   - Ticker duration: Keep at 500ms (monitoring frequency)
//   - Context timeout: Increase to 20s (more time to converge)
//
// Advanced tuning (requires modifying agent creation):
//   - Agent.SetStubbornness(0.05): Lower = faster adaptation
//   - Agent.SetInfluence(0.7): Higher = stronger mutual influence
package main

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/carlisia/bio-adapt/emerge"
)

func main() {
	// Define configuration first
	swarmSize := 150
	maxIterations := 20
	checkInterval := 500 * time.Millisecond
	timeout := 10 * time.Second
	targetState := emerge.State{
		Phase:     0,                      // Target phase (0 radians = all aligned at "12 o'clock")
		Frequency: 250 * time.Millisecond, // How fast agents cycle
		Coherence: 0.65,                   // Target sync level (increased for better demo)
	}

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║     🧲 ATTRACTOR BASIN SYNCHRONIZATION DEMO 🧲            ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Println("🔬 SIMULATION OVERVIEW")
	fmt.Println("├─ What: Core synchronization algorithm demonstration")
	fmt.Println("├─ How: Bio-inspired phase coupling (Kuramoto model)")
	fmt.Println("└─ Note: Pure algorithm - no external systems involved")
	fmt.Println()

	fmt.Println("🎯 THIS SIMULATION")
	fmt.Printf("✅ %d Autonomous Agents with Random Start Times\n", swarmSize)
	fmt.Println("    • Each agent has a phase in [0 – 2π]:")
	fmt.Printf("    • 0 → start of its %vms cycle\n", targetState.Frequency.Milliseconds())
	fmt.Printf("    • π → halfway (%vms in)\n", targetState.Frequency.Milliseconds()/2)
	fmt.Println("    • 2π → end of cycle (ready to restart)")
	fmt.Println("    • Phases are randomized → agents begin out of sync")
	fmt.Println("✅ Local Synchronization Mechanics")
	fmt.Println("    • Phase difference → how misaligned two agents are")
	fmt.Println("    • Coupling strength → how much an agent pulls neighbors toward its phase")
	fmt.Println("    • Based on the Kuramoto model (bio-inspired synchronization dynamics)")
	fmt.Println("✅ Goal: Emergent Global Sync")
	fmt.Println("    • Agents converge toward phase = 0 (fully aligned)")
	fmt.Println("    • No central coordinator required — sync emerges organically")
	fmt.Println("🔄 Agent = Abstract Entity")
	fmt.Println("    • Could represent a service, device, thread, or function — anything needing coordination")
	fmt.Println()

	fmt.Println("📊 UNDERSTANDING COHERENCE:")
	fmt.Println("├─ Coherence = How synchronized the agents are (0-100%)")
	fmt.Println("├─ Think of it like dancers trying to move in unison:")
	fmt.Println("│")
	fmt.Println("├─ 0-20%  = CHAOS")
	fmt.Println("│  • Everyone doing their own thing")
	fmt.Println("│  • No pattern, completely random")
	fmt.Println("│")
	fmt.Println("├─ 20-40% = GROUPS FORMING")
	fmt.Println("│  • Small clusters with same timing")
	fmt.Println("│  • Multiple separate rhythms")
	fmt.Println("│")
	fmt.Println("├─ 40-60% = PARTIAL COORDINATION")
	fmt.Println("│  • Groups starting to merge")
	fmt.Println("│  • Some agents between groups")
	fmt.Println("│")
	fmt.Println("├─ 60-80% = GOOD SYNCHRONIZATION")
	fmt.Println("│  • Most agents in same rhythm")
	fmt.Println("│  • Few outliers remaining")
	fmt.Println("│")
	fmt.Println("└─ 80-100% = EXCELLENT SYNCHRONIZATION")
	fmt.Println("   • All agents moving as one")
	fmt.Println("   • Like a flock of birds turning together")
	fmt.Println()

	fmt.Println("🔬 KEY CONCEPTS:")
	fmt.Println("├─ LOCAL MINIMA = Stuck in a 'good enough' state")
	fmt.Println("│  • Like a ball stuck in a shallow dip, not the deepest valley")
	fmt.Println("│  • Example: 2 groups synced internally but not with each other")
	fmt.Println("│  • System thinks it's optimal but isn't globally optimal")
	fmt.Println("│")
	fmt.Println("├─ METASTABLE STATE = Temporarily stable but fragile")
	fmt.Println("│  • Like a pencil balanced on its tip - stable until disturbed")
	fmt.Println("│  • Example: Agents loosely coordinated, easily disrupted")
	fmt.Println("│  • Small changes could improve OR worsen synchronization")
	fmt.Println("│")
	fmt.Println("└─ PERTURBATION = Intentional disruption to escape stuck states")
	fmt.Println("   • Like shaking a stuck gear to make it move")
	fmt.Println("   • Example: Randomly shifting some agents' phases")
	fmt.Println("   • Helps escape local minima to find better solutions")
	fmt.Println()
	fmt.Println("🔧 SIMULATION SETUP")
	fmt.Printf("├─ Agents: %d independent entities\n", swarmSize)
	fmt.Printf("├─ Target: %.0f%% synchronization (coherence)\n", targetState.Coherence*100)
	fmt.Printf("├─ Window: %v oscillation period\n", targetState.Frequency)
	fmt.Printf("├─ Max iterations: %d (checked every %v)\n", maxIterations, checkInterval)
	fmt.Printf("├─ Max time: %v timeout\n", timeout)
	fmt.Println("└─ Method: Local interactions only (no coordinator)")
	fmt.Println()

	fmt.Println("📊 PARAMETER TRADEOFFS TABLE:")
	fmt.Println("┌─────────────────┬──────────────────────┬──────────────────────┬─────────────┐")
	fmt.Println("│ Parameter       │ Lower Value          │ Higher Value         │ Sweet Spot  │")
	fmt.Println("├─────────────────┼──────────────────────┼──────────────────────┼─────────────┤")
	fmt.Println("│ Swarm Size      │ 3-5 agents           │ 10-20 agents         │ 6-8 agents  │")
	fmt.Println("│                 │ ✅ Converges easily  │ ✅ More realistic    │             │")
	fmt.Println("│                 │ ❌ Too simple        │ ❌ Harder to sync    │             │")
	fmt.Println("├─────────────────┼──────────────────────┼──────────────────────┼─────────────┤")
	fmt.Println("│ Target Coherence│ 0.5-0.6 (50-60%)     │ 0.8-0.95 (80-95%)    │ 0.65-0.75   │")
	fmt.Println("│                 │ ✅ Easy to achieve   │ ✅ Tight sync        │             │")
	fmt.Println("│                 │ ❌ Loose coordination│ ❌ May never reach   │             │")
	fmt.Println("├─────────────────┼──────────────────────┼──────────────────────┼─────────────┤")
	fmt.Println("│ Frequency       │ 250-500ms            │ 50-100ms             │ 150-200ms   │")
	fmt.Println("│                 │ ✅ Stable, gentle    │ ✅ Fast convergence  │             │")
	fmt.Println("│                 │ ❌ Slow to converge  │ ❌ Unstable, jittery │             │")
	fmt.Println("├─────────────────┼──────────────────────┼──────────────────────┼─────────────┤")
	fmt.Printf("│ Check Interval  │ 1000ms               │ 100ms                │ %vms       │\n", checkInterval.Milliseconds())
	fmt.Println("│ (ticker)        │ ✅ Less CPU usage    │ ✅ Fine monitoring   │             │")
	fmt.Println("│                 │ ❌ Miss details      │ ❌ High overhead     │             │")
	fmt.Println("└─────────────────┴──────────────────────┴──────────────────────┴─────────────┘")
	fmt.Println()

	// Create swarm with auto-scaled configuration
	config := emerge.AutoScaleConfig(swarmSize)

	// Optionally customize configuration for demo purposes
	// You can uncomment these lines to see different behaviors:
	// config.MaxConcurrentAgents = 50  // Limit concurrent goroutines
	// config.UseBatchProcessing = true // Enable batch processing
	// config.BatchSize = 25            // Process 25 agents per batch
	// config.WorkerPoolSize = 10       // Use 10 worker goroutines
	// config.AgentUpdateInterval = 100 * time.Millisecond // Slower updates
	// config.MonitoringInterval = 200 * time.Millisecond  // Less frequent monitoring

	swarm, err := emerge.NewSwarm(swarmSize, targetState, emerge.WithConfig(config))
	if err != nil {
		fmt.Printf("❌ Error: failed to create swarm: %v\n", err)
		return
	}

	fmt.Println("⚡ SCALABILITY CONFIGURATION")
	fmt.Printf("├─ Batch Processing: %v\n", config.UseBatchProcessing)
	if config.UseBatchProcessing {
		fmt.Printf("├─ Batch Size: %d agents per batch\n", config.BatchSize)
		fmt.Printf("├─ Worker Pool: %d goroutines\n", config.WorkerPoolSize)
	}
	fmt.Printf("├─ Max Concurrent: %d goroutines\n", config.MaxConcurrentAgents)
	fmt.Printf("├─ Update Interval: %v\n", config.AgentUpdateInterval)
	fmt.Printf("├─ Monitor Interval: %v\n", config.MonitoringInterval)
	if config.EnableConnectionOptim && swarmSize > config.ConnectionOptimThreshold {
		fmt.Printf("├─ Connection Optimization: ENABLED (threshold: %d)\n", config.ConnectionOptimThreshold)
	} else {
		fmt.Println("├─ Connection Optimization: DISABLED")
	}
	fmt.Printf("└─ Max Swarm Size: %d agents\n", config.MaxSwarmSize)
	fmt.Println()

	// Measure initial coherence using Kuramoto order parameter.
	// This will be low (~0.1-0.3) due to random initialization.
	initialCoherence := swarm.MeasureCoherence()

	fmt.Println("═══ INITIAL STATE (SIMULATED) ═══")
	visualizeAgents(swarm)
	fmt.Printf("📊 Coherence Score: %.1f%% ", initialCoherence*100)
	interpretCoherence(initialCoherence)
	fmt.Println()

	// Create a bounded context to prevent infinite execution.
	// In production, you might use a cancel signal or deadline.
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Start the swarm's autonomous synchronization process.
	// Each agent will:
	// 1. Observe neighbors' phases
	// 2. Calculate local coherence
	// 3. Adjust toward the attractor basin
	// 4. Repeat until convergence
	fmt.Println("\n═══ SYNCHRONIZATION PROCESS (SIMULATED) ═══")
	fmt.Println("⚡ Simulating: Agents discovering common rhythm...")
	fmt.Printf("   (Each step = %v of algorithm iterations)\n", checkInterval)
	fmt.Println()

	errChan := make(chan error, 1)
	go func() {
		if err := swarm.Run(ctx); err != nil {
			errChan <- err
		}
	}()

	// Monitor convergence progress at regular intervals.
	// This demonstrates the gradual emergence of synchronization.
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	iterations := 0
	lastCoherence := initialCoherence
	stuckCount := 0

	for {
		select {
		case <-ticker.C:
			iterations++
			coherence := swarm.MeasureCoherence()

			// Show iteration with visual progress (new line each time for compatibility)
			fmt.Printf("Step %2d/%d: ", iterations, maxIterations)

			// Visual progress bar
			drawProgressBar(coherence, targetState.Coherence, 30)

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
				goto done
			}

			// Warn if stuck
			if stuckCount > 5 {
				fmt.Print(" ⚠️  (stuck - may need parameter tuning)")
			}

			fmt.Println()
			lastCoherence = coherence

			// Safety limit to prevent excessive iterations.
			// If convergence hasn't occurred by now, something may be wrong.
			if iterations >= maxIterations {
				fmt.Printf("\n⏱️  Max iterations (%d) reached - stopping\n", maxIterations)
				fmt.Printf("   (%d checks over %v, every %v)\n", maxIterations, timeout, checkInterval)
				fmt.Println("   Tip: System may be stuck in local minima")
				goto done
			}

		case err := <-errChan:
			fmt.Printf("\n❌ Swarm error: %v\n", err)
			goto done

		case <-ctx.Done():
			fmt.Println("\n⏱️  Timeout: context cancelled")
			goto done
		}
	}

done:
	fmt.Println()

	// Calculate and display final metrics.
	// Improvement shows how much the system self-organized from chaos.
	finalCoherence := swarm.MeasureCoherence()
	improvement := ((finalCoherence - initialCoherence) / initialCoherence) * 100

	fmt.Println("═══ FINAL STATE (SIMULATED) ═══")
	visualizeAgents(swarm)
	fmt.Printf("\n📊 Coherence Score: %.1f%% ", finalCoherence*100)
	interpretCoherence(finalCoherence)
	fmt.Println()

	fmt.Println("\n═══ SIMULATION RESULTS ═══")
	fmt.Println("┌─────────────────────────────────────┐")
	fmt.Printf("│ Initial chaos:      %6.1f%%         │\n", initialCoherence*100)
	fmt.Printf("│ Final sync:         %6.1f%%         │\n", finalCoherence*100)
	fmt.Printf("│ Improvement:        %6.1f%%         │\n", improvement)
	fmt.Printf("│ Target (%.0f%%):      ", targetState.Coherence*100)
	if finalCoherence >= targetState.Coherence {
		fmt.Printf("✅ ACHIEVED      │\n")
	} else {
		fmt.Printf("❌ NOT REACHED   │\n")
	}
	fmt.Println("└─────────────────────────────────────┘")

	// Explain what happened
	fmt.Println()
	if finalCoherence >= targetState.Coherence {
		fmt.Println("🎉 Success! The agents synchronized through local interactions.")
		fmt.Println("   No central coordinator was needed - emergence in action!")
	} else if finalCoherence > initialCoherence*2 {
		fmt.Println("📈 Partial success - significant synchronization achieved.")
		fmt.Println()
		fmt.Println("🔍 DIAGNOSTICS - Why didn't we reach target?")

		// Analyze the specific situation
		gap := targetState.Coherence - finalCoherence
		if gap > 0.3 {
			fmt.Println("   ⚠️  Large gap (>30%) suggests fundamental issues:")
			fmt.Println("   • Agents may have formed multiple stable groups")
			fmt.Println("   • Some agents might be too stubborn to adapt")
			fmt.Println("   • Coupling strength may be too weak")
		} else if stuckCount > 5 {
			fmt.Println("   ⚠️  System got stuck (no progress for 5+ iterations):")
			fmt.Println("   • Likely trapped in LOCAL MINIMA (see KEY CONCEPTS above)")
			fmt.Println("   • Agents reached a METASTABLE STATE")
			fmt.Println("   • Would need PERTURBATION to escape (random phase shifts)")
		} else {
			fmt.Println("   ⚠️  Slow convergence detected:")
			fmt.Println("   • More time might help (extend timeout)")
			fmt.Println("   • Frequency might be too fast for stable sync")
		}

		fmt.Println()
		fmt.Println("📊 RECOMMENDED FIXES:")
		fmt.Println("   • Lower target: targetState.Coherence = 0.65")
		fmt.Println("   • Slower cycles: targetState.Frequency = 200ms")
		fmt.Println("   • Fewer agents: swarmSize = 7")
		fmt.Println("   • More time: timeout = 20*time.Second")
	} else {
		fmt.Println("🔄 Limited synchronization achieved.")
		fmt.Println()
		fmt.Println("🔍 DIAGNOSTICS - What went wrong?")

		// Detailed analysis of failure
		if finalCoherence < 0.3 {
			fmt.Println("   ❌ Very low coherence (<30%) indicates:")
			fmt.Println("   • Agents remain essentially random")
			fmt.Println("   • No effective coupling occurring")
			fmt.Println("   • Parameters may be incompatible")
		} else if finalCoherence < initialCoherence*1.5 {
			fmt.Println("   ❌ Minimal improvement suggests:")
			fmt.Println("   • Coupling too weak to overcome randomness")
			fmt.Println("   • Agents too stubborn to adapt")
			fmt.Println("   • Frequency mismatch preventing sync")
		}

		if iterations >= maxIterations {
			fmt.Println("   ❌ Hit iteration limit:")
			fmt.Println("   • System needs more time")
			fmt.Println("   • Or parameters prevent convergence")
		}

		fmt.Println()
		fmt.Println("📊 DEBUGGING STEPS:")
		fmt.Println("   1. Start simple: swarmSize = 3")
		fmt.Println("   2. Easy target: targetState.Coherence = 0.5")
		fmt.Println("   3. Slow frequency: targetState.Frequency = 250ms")
		fmt.Println("   4. If that works, gradually increase complexity")
	}

	fmt.Println()
	fmt.Println("💡 REAL-WORLD APPLICATIONS:")
	fmt.Println("├─ 🔄 Distributed system coordination")
	fmt.Println("├─ 📡 IoT device synchronization")
	fmt.Println("├─ 🎵 Audio/video stream alignment")
	fmt.Println("├─ 💓 Cardiac pacemaker networks")
	fmt.Println("├─ 🚦 Traffic light timing")
	fmt.Println("└─ 🤖 Robot swarm coordination")

	fmt.Println()
	fmt.Println("📐 UNDERSTANDING PHASE IN THIS SIMULATION:")
	fmt.Printf("├─ Phase = Where an agent is in its %vms repeating cycle\n", targetState.Frequency.Milliseconds())
	fmt.Println("├─ Think of it like runners on a circular track:")
	fmt.Println("│  • Phase 0 = At the starting line")
	fmt.Printf("│  • Phase π/2 = Quarter way around (%vms into cycle)\n", targetState.Frequency.Milliseconds()/4)
	fmt.Printf("│  • Phase π = Halfway around (%vms into cycle)\n", targetState.Frequency.Milliseconds()/2)
	fmt.Printf("│  • Phase 3π/2 = Three quarters around (%vms into cycle)\n", targetState.Frequency.Milliseconds()*3/4)
	fmt.Printf("│  • Phase 2π = Back at start (%vms, cycle repeats)\n", targetState.Frequency.Milliseconds())
	fmt.Println("├─ Random initial phases = like runners starting at random")
	fmt.Println("│  positions around the track")
	fmt.Println("└─ Goal: Get all runners to cross the starting line together")

	fmt.Println()
	fmt.Println("🔧 WHAT 'PHASE' COULD MEAN IN YOUR SYSTEM:")
	fmt.Println("├─ 📊 Database backup: Position in backup schedule (0=start, π=halfway)")
	fmt.Println("├─ 🔄 Cache refresh: Timing in refresh cycle")
	fmt.Println("├─ 📡 API calls: Position in request window")
	fmt.Println("├─ 💾 Log rotation: Point in rotation schedule")
	fmt.Println("├─ 🎮 Game loop: Frame timing in update cycle")
	fmt.Println("├─ 📈 Metrics collection: Position in sampling period")
	fmt.Println("└─ 🔐 Token refresh: Timing in auth renewal cycle")

	fmt.Println()
	fmt.Println("⏰ HOW WE SET FREQUENCY IN THIS SIMULATION:")
	fmt.Printf("├─ Frequency: %v (set in targetState)\n", targetState.Frequency)
	fmt.Printf("├─ This means agents complete a full cycle every %vms\n", targetState.Frequency.Milliseconds())
	fmt.Printf("├─ We chose %vms for optimal stability and convergence\n", targetState.Frequency.Milliseconds())
	fmt.Println("├─ Faster frequency = more rapid but potentially unstable sync")
	fmt.Println("└─ Slower frequency = gentler, more reliable convergence")

	fmt.Println()
	fmt.Println("🔧 API TO SET UP YOUR OWN SYSTEM:")
	fmt.Println("```go")
	fmt.Println("// 1. Define your target state")
	fmt.Println("targetState := emerge.State{")
	fmt.Println("    Phase:     0,                      // Where to sync (0 = aligned)")
	fmt.Println("    Frequency: 200 * time.Millisecond, // Your cycle time")
	fmt.Println("    Coherence: 0.8,                    // How tight (0.8 = 80%)")
	fmt.Println("}")
	fmt.Println()
	fmt.Println("// 2. Create your swarm")
	fmt.Println("swarm, err := emerge.NewSwarm(10, targetState)")
	fmt.Println()
	fmt.Println("// 3. Run synchronization")
	fmt.Println("ctx := context.Background()")
	fmt.Println("go swarm.Run(ctx)")
	fmt.Println()
	fmt.Println("// 4. Monitor progress")
	fmt.Println("coherence := swarm.MeasureCoherence()")
	fmt.Println("```")
}

// visualizeAgents shows the current phase distribution of agents
func visualizeAgents(swarm *emerge.Swarm) {
	phases := make([]float64, 0, swarm.Size())
	swarm.Agents().Range(func(key, value any) bool {
		agent := value.(*emerge.Agent)
		phases = append(phases, agent.Phase())
		return true
	})

	// First show what the visualization means
	fmt.Println("🕰️ Phase Distribution (like positions on a clock):")
	fmt.Print("   ")

	// Create 12 bins (like clock positions)
	bins := make([]int, 12)
	for _, phase := range phases {
		bin := int(phase / (2 * math.Pi) * 12)
		if bin >= 12 {
			bin = 11
		}
		if bin < 0 {
			bin = 0
		}
		bins[bin]++
	}

	// Show as clock positions with better explanation
	symbols := []string{"🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕗", "🕘", "🕙", "🕚", "🕛"}
	clockLabels := []string{"12:00", "1:00", "2:00", "3:00", "4:00", "5:00", "6:00", "7:00", "8:00", "9:00", "10:00", "11:00"}

	// Count how many unique positions have agents
	uniquePositions := 0
	for _, count := range bins {
		if count > 0 {
			uniquePositions++
		}
	}

	// Show the distribution
	for i, count := range bins {
		if count > 0 {
			fmt.Printf("%s(%s)=%d ", symbols[i], clockLabels[i], count)
		}
	}
	fmt.Println()

	// Interpret the pattern
	fmt.Print("   Pattern: ")
	if uniquePositions <= 2 {
		fmt.Println("✅ Single cluster - all agents at similar times!")
	} else if uniquePositions <= 4 {
		fmt.Println("🟡 Few groups - 2-3 different timings")
	} else if uniquePositions <= 6 {
		fmt.Println("🟠 Multiple groups - 4-6 different timings")
	} else {
		fmt.Println("🔴 Scattered - many different timings (no coordination)")
	}
}

// drawProgressBar creates a visual progress indicator
func drawProgressBar(current, target float64, width int) {
	progress := min(current/target, 1.0)

	filled := int(progress * float64(width))

	// Use different colors based on progress
	if progress < 0.3 {
		fmt.Print("🔴 [")
	} else if progress < 0.7 {
		fmt.Print("🟡 [")
	} else {
		fmt.Print("🟢 [")
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	fmt.Print(bar)
	fmt.Print("]")
}

// interpretCoherence provides human-readable interpretation
func interpretCoherence(coherence float64) {
	if coherence < 0.2 {
		fmt.Print("(🌪️  Chaos - no coordination)")
	} else if coherence < 0.4 {
		fmt.Print("(🌊 Groups forming - multiple rhythms)")
	} else if coherence < 0.6 {
		fmt.Print("(⚡ Partial coordination - groups merging)")
	} else if coherence < 0.8 {
		fmt.Print("(🎵 Good sync - single dominant rhythm)")
	} else {
		fmt.Print("(✨ Excellent - synchronized as one!)")
	}
}
