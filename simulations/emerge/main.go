// Package main demonstrates how independent AI agents can minimize API calls
// through emergent synchronization, without any central coordinator.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/term"

	"github.com/carlisia/bio-adapt/emerge/goal"
	"github.com/carlisia/bio-adapt/emerge/scale"
	"github.com/carlisia/bio-adapt/simulations/display"
	"github.com/carlisia/bio-adapt/simulations/emerge/simulation"
	"github.com/carlisia/bio-adapt/simulations/emerge/simulation/pattern"
	"github.com/carlisia/bio-adapt/simulations/emerge/ui"
)

func main() {
	// Parse command-line flags
	scaleName := flag.String("scale", "tiny", "Swarm scale: tiny, small, medium, large, huge")
	listScales := flag.Bool("list", false, "List available scales")
	updateInterval := flag.Duration("interval", 100*time.Millisecond, "Display update interval")
	timeout := flag.Duration("timeout", 5*time.Minute, "Simulation timeout (0 for no timeout)")
	flag.Parse()

	// Handle list scales
	if *listScales {
		fmt.Println("Available scales for minimize_api_calls simulation:")
		fmt.Println("  tiny   - 20 agents, tight coordination")
		fmt.Println("  small  - 50 agents, team-sized")
		fmt.Println("  medium - 200 agents, department-scale")
		fmt.Println("  large  - 1000 agents, enterprise-scale")
		fmt.Println("  huge   - 2000 agents, cloud-scale")
		return
	}

	// Parse scale
	swarmScale, ok := ParseScale(*scaleName)
	if !ok {
		fmt.Printf("Unknown scale: %s\n", *scaleName)
		fmt.Println("Available scales: tiny, small, medium, large, huge")
		fmt.Println("Use -list for descriptions")
		os.Exit(1)
	}

	// Create configuration with defaults
	config := DefaultConfig()
	config.Scale = swarmScale
	config.UpdateInterval = *updateInterval
	config.Timeout = *timeout

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Handle graceful shutdown
	go handleSignals(cancel)

	// Run simulation loop (allows scale switching)
	for {
		result := runSimulation(ctx, config)

		if result.Error != nil {
			cancel()
			fmt.Printf("Error: %v\n", result.Error)
			os.Exit(1)
		}

		// Check if user requested a scale, goal, or pattern switch
		var newScale scale.Size
		var newGoal goal.Type
		var newPattern pattern.Type
		scaleChanged := false
		goalChanged := false
		patternChanged := false

		switch result.ScaleEvent {
		// Scale switches
		case display.ScaleTiny:
			newScale = scale.Tiny
			scaleChanged = (config.Scale != newScale)
		case display.ScaleSmall:
			newScale = scale.Small
			scaleChanged = (config.Scale != newScale)
		case display.ScaleMedium:
			newScale = scale.Medium
			scaleChanged = (config.Scale != newScale)
		case display.ScaleLarge:
			newScale = scale.Large
			scaleChanged = (config.Scale != newScale)
		case display.ScaleHuge:
			newScale = scale.Huge
			scaleChanged = (config.Scale != newScale)
		// Goal switches
		case display.GoalBatch:
			newGoal = goal.MinimizeAPICalls
			goalChanged = (config.GoalType != newGoal)
		case display.GoalLoad:
			newGoal = goal.DistributeLoad
			goalChanged = (config.GoalType != newGoal)
		case display.GoalConsensus:
			newGoal = goal.ReachConsensus
			goalChanged = (config.GoalType != newGoal)
		case display.GoalLatency:
			newGoal = goal.MinimizeLatency
			goalChanged = (config.GoalType != newGoal)
		case display.GoalEnergy:
			newGoal = goal.SaveEnergy
			goalChanged = (config.GoalType != newGoal)
		case display.GoalRhythm:
			newGoal = goal.MaintainRhythm
			goalChanged = (config.GoalType != newGoal)
		case display.GoalFailure:
			newGoal = goal.RecoverFromFailure
			goalChanged = (config.GoalType != newGoal)
		case display.GoalTraffic:
			newGoal = goal.AdaptToTraffic
			goalChanged = (config.GoalType != newGoal)
		// Pattern switches
		case display.PatternHighFreq:
			newPattern = pattern.HighFrequency
			patternChanged = (config.Pattern != newPattern)
		case display.PatternBurst:
			newPattern = pattern.Burst
			patternChanged = (config.Pattern != newPattern)
		case display.PatternSteady:
			newPattern = pattern.Steady
			patternChanged = (config.Pattern != newPattern)
		case display.PatternMixed:
			newPattern = pattern.Mixed
			patternChanged = (config.Pattern != newPattern)
		case display.PatternSparse:
			newPattern = pattern.Sparse
			patternChanged = (config.Pattern != newPattern)
		// Other events
		case display.EventQuit:
			// Exit the application
			cancel()
			return
		case display.EventReset, display.EventDisrupt, display.EventPause, display.EventResume:
			// These are handled within the simulation, not here
		default:
			// No change
		}

		// If scale, goal, or pattern was switched, update config and restart
		if scaleChanged || goalChanged || patternChanged {
			cancel() // Cancel old context
			if scaleChanged {
				config.Scale = newScale
				fmt.Printf("\nSwitching to %s scale...\n\n", newScale.String())
			}
			if goalChanged {
				config.GoalType = newGoal
				fmt.Printf("\nSwitching to %s optimization...\n\n", newGoal.String())
			}
			if patternChanged {
				config.Pattern = newPattern
				fmt.Printf("\nSwitching to %s pattern...\n\n", newPattern.String())
			}
			// Create new context for restart
			ctx, cancel = context.WithCancel(context.Background())
			continue
		}
		break
	}
}

// runSimulation creates and runs the simulation with the given configuration.
func runSimulation(ctx context.Context, cfg Config) *ui.RunResult {
	// Create simulation using emerge's pattern: For(goal).With(scale)
	// Adjust target coherence based on goal requirements
	targetCoherence := cfg.Scale.DefaultTargetCoherence()

	// Override for goals that need different coherence targets
	switch cfg.GoalType {
	case goal.DistributeLoad, goal.RecoverFromFailure:
		// These goals want ANTI-PHASE (low coherence)
		// Lower values = more distributed phases
		targetCoherence = 0.2 // Want agents spread out
	case goal.ReachConsensus:
		// Wants partial coherence (voting blocs)
		targetCoherence = 0.6 // Medium coherence for clusters
	case goal.SaveEnergy:
		// Wants sparse synchronization
		targetCoherence = 0.4 // Some coordination but not tight
	case goal.MinimizeAPICalls, goal.MinimizeLatency, goal.MaintainRhythm, goal.AdaptToTraffic:
		// These use the default target coherence
	default:
		// Use default
	}

	buildConfig := simulation.BuildConfig{
		Goal:            cfg.Goal(),
		Scale:           cfg.Scale,
		Pattern:         cfg.Pattern,
		TargetCoherence: targetCoherence,
	}
	sim, err := simulation.New(buildConfig)
	if err != nil {
		return &ui.RunResult{Error: fmt.Errorf("failed to create simulation: %w", err)}
	}

	// Create display based on terminal availability
	var disp display.Display
	if isTerminal() {
		termDisplay := display.NewTerminalDisplay()
		if td, ok := termDisplay.(*display.TerminalDisplay); ok {
			td.SetSimulationName(cfg.DisplayName())
			td.SetGoalName(cfg.Goal().String())
			td.SetGoalAndScale(cfg.Goal(), cfg.Scale)
			// Set pattern from config (or auto-select if not set)
			if cfg.Pattern == pattern.Unset {
				cfg.Pattern = pattern.BestPatternForGoal(int(cfg.Goal()))
			}
			td.SetPattern(cfg.Pattern)
		}
		disp = termDisplay
	} else {
		textDisplay := display.NewTextDisplay()
		if td, ok := textDisplay.(*display.TextDisplay); ok {
			td.SetSimulationName(cfg.DisplayName())
		}
		disp = textDisplay
	}

	// Create controller for keyboard input
	controller := display.NewKeyboardController()

	// Create runner configuration with current state for comparison
	// Note: cfg.Pattern has been set to actual pattern by this point (not Unset)
	runnerConfig := ui.RunnerConfig{
		UpdateInterval: cfg.UpdateInterval,
		Timeout:        cfg.Timeout,
		CurrentGoal:    cfg.GoalType,
		CurrentScale:   cfg.Scale,
		CurrentPattern: cfg.Pattern,
	}

	// Create and run runner
	runner := ui.NewRunner(runnerConfig, sim, disp, controller)
	return runner.Run(ctx)
}

// handleSignals handles OS signals for graceful shutdown.
func handleSignals(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	cancel()
}

// isTerminal checks if we're running in a terminal.
func isTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
