// Package ui provides the runner that connects simulation with display
package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/carlisia/bio-adapt/emerge/goal"
	"github.com/carlisia/bio-adapt/emerge/scale"
	"github.com/carlisia/bio-adapt/simulations/display"
	"github.com/carlisia/bio-adapt/simulations/emerge/simulation"
	"github.com/carlisia/bio-adapt/simulations/emerge/simulation/pattern"
)

// RunnerConfig holds runner configuration
type RunnerConfig struct {
	UpdateInterval time.Duration
	Timeout        time.Duration
	CurrentGoal    goal.Type    // Current goal for comparison
	CurrentScale   scale.Size   // Current scale for comparison
	CurrentPattern pattern.Type // Current pattern for comparison
}

// RunResult represents the result of running the simulation
type RunResult struct {
	Error      error
	ScaleEvent display.EventType // If non-zero, indicates a scale switch was requested
}

// Runner runs simulation with display
type Runner struct {
	config     RunnerConfig
	simulation *simulation.Simulation
	display    display.Display
	controller display.Controller

	// State tracking
	paused bool
}

// NewRunner creates a new runner
func NewRunner(config RunnerConfig, sim *simulation.Simulation, disp display.Display, ctrl display.Controller) *Runner {
	return &Runner{
		config:     config,
		simulation: sim,
		display:    disp,
		controller: ctrl,
	}
}

// Run runs the simulation with display
func (c *Runner) Run(ctx context.Context) *RunResult {
	// Create cancellable context
	appCtx, appCancel := context.WithCancel(ctx)
	defer appCancel()

	// Add timeout if configured
	if c.config.Timeout > 0 {
		var cancel context.CancelFunc
		appCtx, cancel = context.WithTimeout(appCtx, c.config.Timeout)
		defer cancel()
	}

	// Initialize display
	if err := c.display.Initialize(); err != nil {
		return &RunResult{Error: fmt.Errorf("failed to initialize display: %w", err)}
	}
	defer func() {
		if err := c.display.Close(); err != nil {
			// Ignore close error
			_ = err
		}
	}()

	// Show welcome
	c.display.ShowWelcome()

	// Start simulation
	simCtx, simCancel := context.WithCancel(appCtx)
	defer simCancel()

	go func() {
		if err := c.simulation.Start(simCtx); err != nil {
			fmt.Printf("Simulation error: %v\n", err)
		}
	}()

	// Start display updates
	go c.runUpdates(appCtx)

	// Handle control events
	scaleEventChan := make(chan display.EventType, 10) // Increase buffer size
	go c.handleEvents(appCtx, simCancel, appCancel, scaleEventChan)

	// Run display (use KeyboardController wrapper if needed)
	kbController, ok := c.controller.(*display.KeyboardController)
	if !ok {
		// If not a KeyboardController, create a wrapper or handle differently
		// For now, we'll assume it's always a KeyboardController
		return &RunResult{Error: fmt.Errorf("controller must be a KeyboardController")}
	}
	if err := c.display.Run(appCtx, kbController); err != nil {
		return &RunResult{Error: err}
	}

	// Check if a scale switch was requested
	select {
	case scaleEvent := <-scaleEventChan:
		return &RunResult{ScaleEvent: scaleEvent}
	default:
	}

	// Show summary
	stats := c.convertStats(c.simulation.Statistics())
	c.display.ShowSummary(stats)

	return &RunResult{}
}

// runUpdates handles periodic display updates
func (c *Runner) runUpdates(ctx context.Context) {
	ticker := time.NewTicker(c.config.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			snapshot := c.convertSnapshot(c.simulation.Snapshot())
			c.display.Update(snapshot)
		}
	}
}

// isChangeEvent checks if the event represents an actual change from current state
func (c *Runner) isChangeEvent(eventType display.EventType) bool {
	switch eventType {
	// Scale events
	case display.ScaleTiny:
		return c.config.CurrentScale != scale.Tiny
	case display.ScaleSmall:
		return c.config.CurrentScale != scale.Small
	case display.ScaleMedium:
		return c.config.CurrentScale != scale.Medium
	case display.ScaleLarge:
		return c.config.CurrentScale != scale.Large
	case display.ScaleHuge:
		return c.config.CurrentScale != scale.Huge
	// Goal events
	case display.GoalBatch:
		return c.config.CurrentGoal != goal.MinimizeAPICalls
	case display.GoalLoad:
		return c.config.CurrentGoal != goal.DistributeLoad
	case display.GoalConsensus:
		return c.config.CurrentGoal != goal.ReachConsensus
	case display.GoalLatency:
		return c.config.CurrentGoal != goal.MinimizeLatency
	case display.GoalEnergy:
		return c.config.CurrentGoal != goal.SaveEnergy
	case display.GoalRhythm:
		return c.config.CurrentGoal != goal.MaintainRhythm
	case display.GoalFailure:
		return c.config.CurrentGoal != goal.RecoverFromFailure
	case display.GoalTraffic:
		return c.config.CurrentGoal != goal.AdaptToTraffic
	// Pattern events
	case display.PatternHighFreq:
		return c.config.CurrentPattern != pattern.HighFrequency
	case display.PatternBurst:
		return c.config.CurrentPattern != pattern.Burst
	case display.PatternSteady:
		return c.config.CurrentPattern != pattern.Steady
	case display.PatternMixed:
		return c.config.CurrentPattern != pattern.Mixed
	case display.PatternSparse:
		return c.config.CurrentPattern != pattern.Sparse
	case display.EventQuit, display.EventReset, display.EventDisrupt, display.EventPause, display.EventResume:
		return true // Control events always represent a change
	default:
		return true // Other events always represent a change
	}
}

// handleEvents processes control events
func (c *Runner) handleEvents(ctx context.Context, _, appCancel context.CancelFunc, scaleEventChan chan display.EventType) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Get events from the controller channel
			events := c.controller.Events()
			select {
			case event := <-events:
				switch event.Type {
				case display.EventQuit:
					appCancel()
				case display.EventReset:
					// Only allow reset when not paused
					if !c.paused {
						c.simulation.Reset()
					}
				case display.EventDisrupt:
					// Only allow disrupt when not paused
					if !c.paused {
						c.simulation.Disrupt()
					}
				case display.EventPause:
					// Toggle pause state
					if c.paused {
						c.simulation.Resume()
						c.paused = false
					} else {
						c.simulation.Pause()
						c.paused = true
					}
				case display.EventResume:
					c.simulation.Resume()
					c.paused = false
				case display.ScaleTiny, display.ScaleSmall, display.ScaleMedium,
					display.ScaleLarge, display.ScaleHuge,
					display.GoalBatch, display.GoalLoad, display.GoalConsensus,
					display.GoalLatency, display.GoalEnergy, display.GoalRhythm,
					display.GoalFailure, display.GoalTraffic,
					display.PatternHighFreq, display.PatternBurst, display.PatternSteady,
					display.PatternMixed, display.PatternSparse:
					// Only restart if this represents an actual change
					if c.isChangeEvent(event.Type) {
						// Clear channel and send new event
						select {
						case <-scaleEventChan:
							// Clear any existing event
						default:
						}
						// Send new event
						scaleEventChan <- event.Type
						appCancel()
					}
					// Otherwise ignore - no change needed
				}
			case <-time.After(10 * time.Millisecond):
				// No event, continue
			}
		}
	}
}

// convertSnapshot converts simulation.Snapshot to display.SimulationSnapshot
func (*Runner) convertSnapshot(sim simulation.Snapshot) display.SimulationSnapshot {
	// Convert agent snapshots
	agents := make([]display.AgentSnapshot, len(sim.Agents))
	for i, a := range sim.Agents {
		agents[i] = display.AgentSnapshot{
			ID:            a.ID,
			Type:          a.Type,
			Icon:          a.Icon,
			Phase:         a.Phase,
			PendingTasks:  a.PendingTasks,
			BatchesSent:   a.BatchesSent,
			InBurstMode:   a.InBurstMode,
			ActivityLevel: a.ActivityLevel,
		}
	}

	return display.SimulationSnapshot{
		Timestamp:        sim.Timestamp,
		ElapsedTime:      sim.ElapsedTime,
		Agents:           agents,
		Coherence:        sim.Coherence,
		TargetCoherence:  sim.TargetCoherence,
		PendingTasks:     sim.PendingTasks,
		CurrentBatchSize: sim.CurrentBatchSize,
		BatchesProcessed: sim.BatchesProcessed,
		CostWithoutSync:  sim.CostWithoutSync,
		CostWithSync:     sim.CostWithSync,
		Savings:          sim.Savings,
		SavingsPercent:   sim.SavingsPercent,
		Paused:           sim.Paused,
		Disrupted:        sim.Disrupted,
		Reset:            sim.Reset,
		BatchJustSent:    sim.BatchJustSent,
		LastBatchTime:    sim.LastBatchTime,
		LastBatchSize:    sim.LastBatchSize,
	}
}

// convertStats converts simulation.Statistics to display.Statistics
func (*Runner) convertStats(sim simulation.Statistics) display.Statistics {
	return display.Statistics{
		TotalAPICalls:    sim.TotalAPICalls,
		TotalBatches:     sim.TotalBatches,
		AverageBatchSize: sim.AverageBatchSize,
		CostWithoutSync:  sim.CostWithoutSync,
		CostWithSync:     sim.CostWithSync,
		TotalSavings:     sim.TotalSavings,
		SavingsPercent:   sim.SavingsPercent,
		FinalCoherence:   sim.FinalCoherence,
		PeakCoherence:    sim.PeakCoherence,
		TimeToConverge:   sim.TimeToConverge,
	}
}
