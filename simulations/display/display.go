// Package display provides terminal-based display components for simulations.
// It includes both text-based and terminal UI visualizations that can be easily
// swapped for GUI or web interfaces.
package display

import (
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/gauge"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"

	"github.com/carlisia/bio-adapt/emerge/goal"
	"github.com/carlisia/bio-adapt/emerge/scale"
	"github.com/carlisia/bio-adapt/simulations/emerge/simulation/pattern"
)

// Constants for sync indicators
const (
	syncRed    = "🔴"
	syncGreen  = "🟢"
	syncYellow = "🟡"
)

// Constants for activity patterns
const (
	activityQuiet = "quiet"
	activityBurst = "burst"
)

// TerminalDisplay implements the Display interface using termdash
type TerminalDisplay struct {
	// Terminal components
	terminal   *tcell.Terminal
	container  *container.Container
	controller *KeyboardController

	// Widgets
	titleText        *text.Text
	configText       *text.Text // New: Left box for Goal/Pattern/Scale
	descriptionText  *text.Text // New: Right box for Problem/Solution
	agentsText       *text.Text
	metricsText      *text.Text
	costText         *text.Text
	swarmText        *text.Text
	swarmVisText     *text.Text // New swarm visualization
	instructionsText *text.Text // Menu/instructions widget
	targetGauge      *gauge.Gauge
	coherenceGauge   *gauge.Gauge
	coherenceChart   *linechart.LineChart

	// Data tracking
	coherenceHistory []float64
	targetHistory    []float64
	mu               sync.Mutex

	// Configuration
	simulationName string
	goalName       string
	patternName    string
	currentGoal    goal.Type
	currentScale   scale.Size
	currentPattern pattern.Type
}

// NewTerminalDisplay creates a new terminal-based display
func NewTerminalDisplay() Display {
	return &TerminalDisplay{
		coherenceHistory: make([]float64, 0, 100),
		targetHistory:    make([]float64, 0, 100),
		simulationName:   "MINIMIZE API CALLS", // Default name
	}
}

// SetSimulationName sets the simulation name for display
func (d *TerminalDisplay) SetSimulationName(name string) {
	d.simulationName = name
}

// SetGoalName sets the goal name for display
func (d *TerminalDisplay) SetGoalName(name string) {
	d.goalName = name
}

// SetGoalAndScale sets the current goal and scale for recommendation checking
func (d *TerminalDisplay) SetGoalAndScale(g goal.Type, s scale.Size) {
	d.currentGoal = g
	d.currentScale = s
	d.updateInstructionsWidget() // Update menu to reflect new combination
}

// SetPattern sets the current request pattern for display
func (d *TerminalDisplay) SetPattern(p pattern.Type) {
	d.currentPattern = p
	d.patternName = p.String()
	d.updateInstructionsWidget() // Update menu to reflect new pattern
}

// getBestWorstDescriptions returns best and worst case descriptions for each goal
func (d *TerminalDisplay) getBestWorstDescriptions() (best, worst string) {
	switch d.currentGoal {
	case goal.MinimizeAPICalls:
		best = "All workloads synchronized, maximum batching"
		worst = "Workloads scattered, no batching possible"

	case goal.DistributeLoad:
		best = "Perfect anti-phase, resources evenly distributed"
		worst = "All synchronized, maximum resource contention"

	case goal.ReachConsensus:
		best = "Voting blocs formed, rapid agreement"
		worst = "Every node isolated, no consensus possible"

	case goal.MinimizeLatency:
		best = "Predictable synchronized responses"
		worst = "Alternating phases causing jitter"

	case goal.SaveEnergy:
		best = "Sparse, coordinated activity"
		worst = "All sensors active simultaneously"

	case goal.MaintainRhythm:
		best = "Steady beat, all jobs on schedule"
		worst = "Irregular drift, schedule chaos"

	case goal.RecoverFromFailure:
		best = "Isolated failures, smooth recovery"
		worst = "Clustered failures, cascading impact"

	case goal.AdaptToTraffic:
		best = "Distributed scaling, smooth handling"
		worst = "Synchronized scaling, resource spikes"

	default:
		best = "Optimal synchronization pattern"
		worst = "Chaotic, uncoordinated behavior"
	}

	return best, worst
}

// Initialize sets up the terminal display
func (d *TerminalDisplay) Initialize() error {
	// Create terminal
	t, err := tcell.New()
	if err != nil {
		return fmt.Errorf("failed to create terminal: %w", err)
	}
	d.terminal = t

	// Create widgets
	if err := d.createWidgets(); err != nil {
		return fmt.Errorf("failed to create widgets: %w", err)
	}

	// Create layout
	if err := d.createLayout(); err != nil {
		return fmt.Errorf("failed to create layout: %w", err)
	}

	return nil
}

// createWidgets initializes all display widgets
func (d *TerminalDisplay) createWidgets() error {
	var err error

	// Title text
	d.titleText, err = text.New()
	if err != nil {
		return err
	}

	// Config text (left box)
	d.configText, err = text.New()
	if err != nil {
		return err
	}

	// Description text (right box)
	d.descriptionText, err = text.New()
	if err != nil {
		return err
	}

	// Agents display
	d.agentsText, err = text.New()
	if err != nil {
		return err
	}

	// Metrics display
	d.metricsText, err = text.New()
	if err != nil {
		return err
	}

	// Cost display
	d.costText, err = text.New()
	if err != nil {
		return err
	}

	// Swarm display
	d.swarmText, err = text.New()
	if err != nil {
		return err
	}

	// Swarm visualization display
	d.swarmVisText, err = text.New()
	if err != nil {
		return err
	}

	// Instructions/menu display
	d.instructionsText, err = text.New()
	if err != nil {
		return err
	}
	d.updateInstructionsWidget() // Initialize menu content

	// Target gauge (shows the goal)
	d.targetGauge, err = gauge.New(
		gauge.Height(2), // Reduced height since we have padding
		gauge.Color(cell.ColorYellow),
		gauge.FilledTextColor(cell.ColorBlack),
		gauge.EmptyTextColor(cell.ColorBlack),
	)
	if err != nil {
		return err
	}
	// Set target to initial value (will be updated with actual config value in Update)
	if err := d.targetGauge.Percent(0, gauge.TextLabel("Target")); err != nil {
		return err
	}

	// Coherence gauge (shows current progress)
	d.coherenceGauge, err = gauge.New(
		gauge.Height(2),             // Reduced height since we have padding
		gauge.Color(cell.ColorCyan), // Default color, will be updated dynamically
		gauge.FilledTextColor(cell.ColorBlack),
		gauge.EmptyTextColor(cell.ColorBlack),
	)
	if err != nil {
		return err
	}

	// Coherence chart
	d.coherenceChart, err = linechart.New(
		linechart.YAxisAdaptive(),
	)
	if err != nil {
		return err
	}

	return nil
}

// createLayout builds the terminal layout
func (d *TerminalDisplay) createLayout() error {
	builder := grid.New()
	builder.Add(
		// Title section - 14% with two side-by-side boxes (increased from 12%)
		grid.RowHeightPerc(14,
			// Left box - Configuration (Goal, Pattern, Scale)
			grid.ColWidthPerc(35,
				grid.Widget(d.configText,
					container.Border(linestyle.Round),
					container.BorderTitle(" CONFIGURATION "),
					container.BorderTitleAlignCenter(),
					container.BorderColor(cell.ColorYellow),
				),
			),
			// Right box - Problem/Solution description
			grid.ColWidthPerc(65,
				grid.Widget(d.descriptionText,
					container.Border(linestyle.Round),
					container.BorderTitle(" SCENARIO "),
					container.BorderTitleAlignCenter(),
					container.BorderColor(cell.ColorCyan),
				),
			),
		),

		// Main content - 78% (decreased from 80% to compensate)
		grid.RowHeightPerc(78,
			// Left side - Application/Simulation data
			grid.ColWidthPerc(24,
				grid.RowHeightPerc(52,
					grid.Widget(d.agentsText,
						container.Border(linestyle.Light),
						container.BorderTitle(" WORKLOADS "),
						container.BorderColor(cell.ColorYellow), // App data - Yellow
					),
				),
				grid.RowHeightPerc(48,
					grid.Widget(d.metricsText,
						container.Border(linestyle.Light),
						container.BorderTitle(" METRICS "),
						container.BorderColor(cell.ColorYellow), // App data - Yellow
					),
				),
			),

			// Right side - Visualizations (expanded to 76% from 60%)
			grid.ColWidthPerc(76,
				grid.RowHeightPerc(10,
					grid.Widget(d.targetGauge,
						container.Border(linestyle.Light),
						container.BorderTitle(" TARGET COHERENCE "),
						container.BorderColor(cell.ColorYellow), // App config - Yellow
						container.PaddingTop(1),
					),
				),
				grid.RowHeightPerc(10,
					grid.Widget(d.coherenceGauge,
						container.Border(linestyle.Light),
						container.BorderTitle(" CURRENT COHERENCE "),
						container.BorderColor(cell.ColorCyan), // Emerge data - Cyan
						container.PaddingTop(1),
					),
				),
				grid.RowHeightPerc(40,
					grid.Widget(d.coherenceChart,
						container.Border(linestyle.Light),
						container.BorderTitle(" SYNCHRONIZATION COHERENCE "),
						container.BorderColor(cell.ColorCyan), // Emerge data - Cyan
					),
				),
				grid.RowHeightPerc(40,
					// Split bottom row into three columns
					grid.ColWidthPerc(33,
						grid.Widget(d.costText,
							container.Border(linestyle.Light),
							container.BorderTitle(" ECONOMICS "),
							container.BorderColor(cell.ColorYellow), // App data - Yellow
						),
					),
					grid.ColWidthPerc(33,
						grid.Widget(d.swarmText,
							container.Border(linestyle.Light),
							container.BorderTitle(" SWARM DYNAMICS "),
							container.BorderColor(cell.ColorCyan), // Emerge data - Cyan
						),
					),
					grid.ColWidthPerc(34,
						grid.Widget(d.swarmVisText,
							container.Border(linestyle.Light),
							container.BorderTitle(" SWARM MOVEMENT "),
							container.BorderColor(cell.ColorCyan), // Emerge data - Cyan
						),
					),
				),
			),
		),

		// Instructions - 8% (increased to fit scale controls)
		grid.RowHeightPerc(8,
			grid.Widget(d.instructionsText,
				container.Border(linestyle.Round),
				container.BorderTitle(" MENU "),
				container.BorderColor(cell.ColorYellow),
			),
		),
	)

	gridOpts, err := builder.Build()
	if err != nil {
		return err
	}

	c, err := container.New(d.terminal, gridOpts...)
	if err != nil {
		return err
	}

	d.container = c
	return nil
}

// writeText is a helper for non-critical text widget writes that logs errors but continues
func writeText(w *text.Text, s string, opts ...text.WriteOption) {
	if w == nil {
		return
	}
	if err := w.Write(s, opts...); err != nil {
		// Log the error but continue - UI writes shouldn't crash the demo
		fmt.Printf("UI write error: %v\n", err)
	}
}

// updateInstructionsWidget updates the instructions text widget
func (d *TerminalDisplay) updateInstructionsWidget() {
	if d.instructionsText == nil {
		return
	}

	d.instructionsText.Reset()

	// Goal optimization controls with selected goal bolded
	writeText(d.instructionsText, "  Goals:   ",
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	goals := []struct {
		key  string
		name string
		typ  goal.Type
	}{
		{"[B]", "Batch", goal.MinimizeAPICalls},
		{"[L]", "Load", goal.DistributeLoad},
		{"[C]", "Consensus", goal.ReachConsensus},
		{"[T]", "Latency", goal.MinimizeLatency},
		{"[E]", "Energy", goal.SaveEnergy},
		{"[M]", "Rhythm", goal.MaintainRhythm},
		{"[F]", "Failure", goal.RecoverFromFailure},
		{"[A]", "Traffic", goal.AdaptToTraffic},
	}

	for i, g := range goals {
		writeText(d.instructionsText, " "+g.key+" ",
			text.WriteCellOpts(cell.FgColor(cell.ColorMagenta)))

		// Bold and white if this is the selected goal
		if d.currentGoal == g.typ {
			writeText(d.instructionsText, g.name,
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite), cell.Bold()))
		} else {
			writeText(d.instructionsText, g.name,
				text.WriteCellOpts(cell.FgColor(cell.ColorMagenta)))
		}

		// Add spacing except after last item
		if i < len(goals)-1 {
			writeText(d.instructionsText, " ",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		}
	}

	writeText(d.instructionsText, "\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	// Pattern controls with optimal ones bolded
	writeText(d.instructionsText, "  Patterns:",
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	// Check which patterns are optimal for current goal and scale
	patterns := []struct {
		key  string
		name string
		typ  pattern.Type
	}{
		{"[H]", "High-Freq", pattern.HighFrequency},
		{"[U]", "Burst", pattern.Burst},
		{"[Y]", "Steady", pattern.Steady},
		{"[X]", "Mixed", pattern.Mixed},
		{"[Z]", "Sparse", pattern.Sparse},
	}

	for i, p := range patterns {
		writeText(d.instructionsText, " "+p.key+" ",
			text.WriteCellOpts(cell.FgColor(cell.ColorGreen)))

		// Check if this pattern is optimal for current goal and scale
		agentCount := d.currentScale.DefaultAgentCount()
		qualifier := pattern.GetQualifier(int(d.currentGoal), agentCount, p.typ)

		// Bold and white if optimal, excellent, or good (matches assessment checkmark)
		if qualifier == pattern.QualifierOptimal || qualifier == pattern.QualifierExcellent || qualifier == pattern.QualifierGood {
			writeText(d.instructionsText, p.name,
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite), cell.Bold()))
		} else {
			writeText(d.instructionsText, p.name,
				text.WriteCellOpts(cell.FgColor(cell.ColorGreen)))
		}

		// Add spacing except after last item
		if i < len(patterns)-1 {
			writeText(d.instructionsText, " ",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		}
	}

	// Extra spacing between patterns and scales
	writeText(d.instructionsText, "     ",
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	// Scale controls with optimal ones bolded
	writeText(d.instructionsText, "Scales: ",
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	scales := []struct {
		key   string
		name  string
		size  scale.Size
		count int
	}{
		{"[1]", "Tiny", scale.Tiny, 20},
		{"[2]", "Small", scale.Small, 50},
		{"[3]", "Medium", scale.Medium, 200},
		{"[4]", "Large", scale.Large, 1000},
		{"[5]", "Huge", scale.Huge, 5000},
	}

	for i, s := range scales {
		writeText(d.instructionsText, s.key+" ",
			text.WriteCellOpts(cell.FgColor(cell.ColorYellow)))

		// Check if this scale is recommended for current goal
		if d.currentGoal.IsRecommendedForSize(s.count) {
			// Bold and white if recommended
			writeText(d.instructionsText, s.name,
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite), cell.Bold()))
		} else {
			writeText(d.instructionsText, s.name,
				text.WriteCellOpts(cell.FgColor(cell.ColorYellow)))
		}

		// Add spacing except after last item
		if i < len(scales)-1 {
			writeText(d.instructionsText, "  ",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		}
	}

	writeText(d.instructionsText, "\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	// Add spacing before actions
	writeText(d.instructionsText, "\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	// Action controls
	writeText(d.instructionsText, "  Actions: ",
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	writeText(d.instructionsText, " [D] Disrupt  [R] Reset  [P] Pause  [Q] Quit",
		text.WriteCellOpts(cell.FgColor(cell.ColorCyan)))
}

// Close cleans up the display
func (d *TerminalDisplay) Close() error {
	if d.terminal != nil {
		d.terminal.Close()
	}
	return nil
}

// ShowWelcome displays the welcome screen
func (d *TerminalDisplay) ShowWelcome() {
	// Clear both text widgets
	d.configText.Reset()
	d.descriptionText.Reset()

	// Left box: Configuration (Goal, Pattern, Scale stacked vertically)
	if d.goalName != "" {
		// Goal
		writeText(d.configText, "🚀 Goal:     ",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		writeText(d.configText, d.goalName,
			text.WriteCellOpts(cell.FgColor(cell.ColorMagenta), cell.Bold()))

		// Pattern
		if d.patternName != "" {
			writeText(d.configText, "\n📊 Pattern:  ",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			writeText(d.configText, d.patternName,
				text.WriteCellOpts(cell.FgColor(cell.ColorGreen), cell.Bold()))

			// Show pattern quality
			agentCount := d.currentScale.DefaultAgentCount()
			qualifier := pattern.GetQualifier(int(d.currentGoal), agentCount, d.currentPattern)

			// Show recommendation indicator
			switch qualifier {
			case pattern.QualifierOptimal:
				writeText(d.configText, "  ✓",
					text.WriteCellOpts(cell.FgColor(cell.ColorGreen)))
			case pattern.QualifierExcellent, pattern.QualifierGood:
				writeText(d.configText, "  ✓",
					text.WriteCellOpts(cell.FgColor(cell.ColorGreen)))
			case pattern.QualifierFair:
				writeText(d.configText, "  ⚠",
					text.WriteCellOpts(cell.FgColor(cell.ColorYellow)))
			case pattern.QualifierPoor:
				writeText(d.configText, "  ✗",
					text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
			}
		}

		// Scale
		writeText(d.configText, "\n⚡ Scale:    ",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		writeText(d.configText, d.simulationName,
			text.WriteCellOpts(cell.FgColor(cell.ColorYellow), cell.Bold()))

		// Scale effectiveness indicator
		agentCount := d.currentScale.DefaultAgentCount()
		if d.currentGoal.IsRecommendedForSize(agentCount) {
			writeText(d.configText, "  ✓",
				text.WriteCellOpts(cell.FgColor(cell.ColorGreen)))
		} else {
			writeText(d.configText, "  ✗",
				text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
		}
	}

	// Right box: Configuration Assessment, Best/Worst, and How It Works
	if d.goalName != "" {
		// Configuration Assessment at the top
		agentCount := d.currentScale.DefaultAgentCount()
		qualifier := pattern.GetQualifier(int(d.currentGoal), agentCount, d.currentPattern)

		writeText(d.descriptionText, "🎯 Configuration Assessment:",
			text.WriteCellOpts(cell.FgColor(cell.ColorCyan), cell.Bold()))

		// Pattern assessment
		var patternColor cell.Color
		var patternStatus string
		switch qualifier {
		case pattern.QualifierOptimal:
			patternColor = cell.ColorGreen
			patternStatus = "Optimal"
		case pattern.QualifierExcellent, pattern.QualifierGood:
			patternColor = cell.ColorGreen
			patternStatus = "Good"
		case pattern.QualifierFair:
			patternColor = cell.ColorYellow
			patternStatus = "Fair"
		case pattern.QualifierPoor:
			patternColor = cell.ColorRed
			patternStatus = "Poor"
		default:
			patternColor = cell.ColorGray
			patternStatus = "Unknown"
		}

		writeText(d.descriptionText, "\n   • Pattern: ",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		writeText(d.descriptionText, patternStatus,
			text.WriteCellOpts(cell.FgColor(patternColor)))
		isGood := qualifier == pattern.QualifierOptimal ||
			qualifier == pattern.QualifierExcellent ||
			qualifier == pattern.QualifierGood
		reason := getPatternReason(d.currentGoal, d.currentPattern, isGood)
		writeText(d.descriptionText, reason,
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

		// Scale assessment
		writeText(d.descriptionText, "\n   • Scale: ",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		if d.currentGoal.IsRecommendedForSize(agentCount) {
			writeText(d.descriptionText, "Good",
				text.WriteCellOpts(cell.FgColor(cell.ColorGreen)))

			// Provide specific reason
			reason := ""
			switch d.currentGoal {
			case goal.MinimizeAPICalls:
				reason = " (more agents = bigger batches)"
			case goal.DistributeLoad:
				if agentCount >= 200 {
					reason = " (many workers share the load)"
				} else {
					reason = " (enough workers to distribute)"
				}
			case goal.ReachConsensus:
				if agentCount <= 200 {
					reason = " (fewer agents = faster agreement)"
				} else {
					reason = " (still manageable voting time)"
				}
			case goal.MinimizeLatency:
				reason = " (fewer hops = faster response)"
			case goal.SaveEnergy:
				if agentCount <= 200 {
					reason = " (fewer agents = less power)"
				} else {
					reason = " (coordinated sleep cycles)"
				}
			case goal.MaintainRhythm:
				if agentCount <= 50 {
					reason = " (small groups sync naturally)"
				} else {
					reason = " (frequency locking scales)"
				}
			case goal.RecoverFromFailure:
				reason = " (backup agents ready)"
			case goal.AdaptToTraffic:
				if agentCount >= 200 {
					reason = " (handles traffic spikes)"
				} else {
					reason = " (absorbs normal variance)"
				}
			}
			writeText(d.descriptionText, reason,
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		} else {
			writeText(d.descriptionText, "Poor",
				text.WriteCellOpts(cell.FgColor(cell.ColorRed)))

			// Provide specific reason
			reason := ""
			switch d.currentGoal {
			case goal.MinimizeAPICalls:
				// MinimizeAPICalls works at all scales, so this shouldn't happen
				// but provide a fallback just in case
				reason = " (less batching opportunity)"
			case goal.DistributeLoad:
				// DistributeLoad needs at least 20 agents
				if agentCount < 20 {
					reason = " (not enough workers)"
				} else {
					reason = " (coordination overhead)"
				}
			case goal.ReachConsensus:
				// ReachConsensus optimal range is 50-1000
				switch {
				case agentCount < 50:
					reason = " (too few voices for quorum)"
				case agentCount > 1000:
					reason = " (too many agents slows consensus)"
				default:
					reason = " (group size slows agreement)"
				}
			case goal.MinimizeLatency:
				if agentCount > 200 {
					reason = " (more hops between agents)"
				} else {
					reason = " (coordination adds delay)"
				}
			case goal.SaveEnergy:
				if agentCount > 200 {
					reason = " (too many agents drain power)"
				} else {
					reason = " (energy not optimized)"
				}
			case goal.MaintainRhythm:
				// MaintainRhythm works at all sizes, shouldn't get here
				reason = " (rhythm disrupted)"
			case goal.RecoverFromFailure:
				// RecoverFromFailure needs at least 20 agents
				if agentCount < 20 {
					reason = " (no backup agents)"
				} else {
					reason = " (recovery not optimal)"
				}
			case goal.AdaptToTraffic:
				// AdaptToTraffic optimal range is 20-1000
				switch {
				case agentCount < 20:
					reason = " (can't handle spikes)"
				case agentCount > 1000:
					reason = " (slow to adapt)"
				default:
					reason = " (traffic handling limited)"
				}
			default:
				// Generic fallback - explain the actual problem
				switch {
				case agentCount < 50:
					reason = " (too few agents for coordination)"
				case agentCount > 1000:
					reason = " (coordination overhead dominates)"
				default:
					reason = " (inefficient at this size)"
				}
			}
			writeText(d.descriptionText, reason,
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		}

		// Best and Worst right after assessment (no line break)
		best, worst := d.getBestWorstDescriptions()
		writeText(d.descriptionText, "\n✅ Best:  ",
			text.WriteCellOpts(cell.FgColor(cell.ColorGreen), cell.Bold()))
		writeText(d.descriptionText, best,
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

		writeText(d.descriptionText, "\n❌ Worst: ",
			text.WriteCellOpts(cell.FgColor(cell.ColorRed), cell.Bold()))
		writeText(d.descriptionText, worst,
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

		// HOW IT WORKS section for all goals (with line break)
		writeText(d.descriptionText, "\n\n🎯 HOW IT WORKS:\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorYellow), cell.Bold()))

		switch d.currentGoal {
		case goal.MinimizeAPICalls:
			writeText(d.descriptionText, "1. Workloads generate tasks continuously\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			writeText(d.descriptionText, "2. Tasks queue up (see 'Queue' column)\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			writeText(d.descriptionText, "3. When sync 🟢 = all workloads send together\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			writeText(d.descriptionText, "4. Combined batch gets 50% discount!\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

		case goal.DistributeLoad:
			writeText(d.descriptionText, "1. Server instances compete for resources\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			writeText(d.descriptionText, "2. Emerge spreads them to anti-phase\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			writeText(d.descriptionText, "3. When 🟢 = perfectly distributed load\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			writeText(d.descriptionText, "4. No resource contention, smooth operation!\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

		case goal.ReachConsensus:
			writeText(d.descriptionText, "1. Distributed nodes need to agree\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			writeText(d.descriptionText, "2. Nodes form voting blocs (clusters)\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			writeText(d.descriptionText, "3. When 🟢 = bloc votes together\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			writeText(d.descriptionText, "4. Consensus reached faster!\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

		case goal.MinimizeLatency:
			writeText(d.descriptionText, "1. Components process requests\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			writeText(d.descriptionText, "2. Synchronize for predictable timing\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			writeText(d.descriptionText, "3. When 🟢 = consistent response times\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			writeText(d.descriptionText, "4. Low, predictable latency achieved!\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

		case goal.SaveEnergy:
			writeText(d.descriptionText, "1. IoT sensors transmit data\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			writeText(d.descriptionText, "2. Coordinate sparse activity windows\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			writeText(d.descriptionText, "3. When 🟢 = sensors take turns\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			writeText(d.descriptionText, "4. Battery life extended significantly!\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

		case goal.MaintainRhythm:
			writeText(d.descriptionText, "1. Scheduled jobs drift over time\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			writeText(d.descriptionText, "2. Phase-lock to maintain schedule\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			writeText(d.descriptionText, "3. When 🟢 = steady rhythm locked\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			writeText(d.descriptionText, "4. Jobs execute precisely on time!\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

		case goal.RecoverFromFailure:
			writeText(d.descriptionText, "1. System nodes monitor health\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			writeText(d.descriptionText, "2. Spread failures across time (anti-phase)\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			writeText(d.descriptionText, "3. When 🟢 = failures isolated\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			writeText(d.descriptionText, "4. Smooth recovery, no cascades!\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

		case goal.AdaptToTraffic:
			writeText(d.descriptionText, "1. Auto-scalers monitor traffic\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			writeText(d.descriptionText, "2. Stagger scaling decisions (anti-phase)\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			writeText(d.descriptionText, "3. When 🟢 = distributed scaling\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			writeText(d.descriptionText, "4. Handle bursts without overload!\n",
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		}
	}
}

// updateConfigurationDisplay updates the configuration box with elapsed time
func (d *TerminalDisplay) updateConfigurationDisplay(elapsed time.Duration) {
	d.configText.Reset()

	// Elapsed time at the top
	writeText(d.configText, "⏱️  Elapsed:  ",
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	writeText(d.configText, formatDuration(elapsed),
		text.WriteCellOpts(cell.FgColor(cell.ColorYellow), cell.Bold()))
	writeText(d.configText, "\n")
	writeText(d.configText, strings.Repeat("─", 25)+"\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorGray)))

	// Goal
	writeText(d.configText, "🚀 Goal:     ",
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	writeText(d.configText, d.goalName,
		text.WriteCellOpts(cell.FgColor(cell.ColorMagenta), cell.Bold()))
	writeText(d.configText, "\n")

	// Pattern
	writeText(d.configText, "📊 Pattern:  ",
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	patternColor := cell.ColorCyan
	if d.patternName != "" {
		writeText(d.configText, d.patternName,
			text.WriteCellOpts(cell.FgColor(patternColor), cell.Bold()))
	}
	writeText(d.configText, "\n")

	// Scale
	writeText(d.configText, "📈 Scale:    ",
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	scaleColor := cell.ColorGreen
	if d.simulationName != "" {
		writeText(d.configText, d.simulationName,
			text.WriteCellOpts(cell.FgColor(scaleColor), cell.Bold()))
	}
}

// Update refreshes the display with new simulation data
func (d *TerminalDisplay) Update(snapshot SimulationSnapshot) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Update configuration display with elapsed time
	d.updateConfigurationDisplay(snapshot.ElapsedTime)

	// Update agents display
	d.updateAgentsDisplay(snapshot.Agents)

	// Update metrics
	d.updateMetricsDisplay(snapshot)

	// Update cost display
	d.updateCostDisplay(snapshot)

	// Update swarm display
	d.updateSwarmDisplay(snapshot)

	// Update swarm visualization
	d.updateSwarmVisualization(snapshot)

	// Update menu to reflect current goal/scale/pattern combination
	d.updateInstructionsWidget()

	// Update gauges based on goal type
	d.updateGoalSpecificGauges(snapshot)

	// Update coherence chart
	d.updateCoherenceChart(snapshot.Coherence, snapshot.TargetCoherence, snapshot.Paused)
}

// updateAgentsDisplay updates the agents panel
func (d *TerminalDisplay) updateAgentsDisplay(agents []AgentSnapshot) {
	d.agentsText.Reset()

	// Explanatory indicators (goal-aware)
	switch d.currentGoal {
	case goal.MinimizeAPICalls:
		writeText(d.agentsText, "• 🟢 = Synced (can batch)\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		writeText(d.agentsText, "• 🟡 = Almost synced\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		writeText(d.agentsText, "• 🔴 = Not synced (no batching)\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	case goal.DistributeLoad:
		writeText(d.agentsText, "• 🟢 = Spread out (no conflicts)\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		writeText(d.agentsText, "• 🟡 = Getting distributed\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		writeText(d.agentsText, "• 🔴 = Too synced (contention)\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	case goal.ReachConsensus:
		writeText(d.agentsText, "• 🟢 = In voting bloc\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		writeText(d.agentsText, "• 🟡 = Forming blocs\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		writeText(d.agentsText, "• Goal: Agreement clusters\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	case goal.MinimizeLatency:
		writeText(d.agentsText, "• 🟢 = Predictable timing\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		writeText(d.agentsText, "• 🟡 = Converging\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		writeText(d.agentsText, "• 🔴 = Unpredictable delays\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	case goal.SaveEnergy:
		writeText(d.agentsText, "• 🟢 = Taking turns (saves power)\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		writeText(d.agentsText, "• 🟡 = Partially coordinated\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		writeText(d.agentsText, "• 🔴 = No coordination (drains)\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	case goal.MaintainRhythm:
		writeText(d.agentsText, "• 🟢 = Steady beat locked\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		writeText(d.agentsText, "• 🟡 = Finding rhythm\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		writeText(d.agentsText, "• 🔴 = Drifting schedule\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	case goal.RecoverFromFailure:
		writeText(d.agentsText, "• 🟢 = Staggered (resilient)\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		writeText(d.agentsText, "• 🟡 = Spreading out\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		writeText(d.agentsText, "• 🔴 = Clustered (cascade risk)\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	case goal.AdaptToTraffic:
		writeText(d.agentsText, "• 🟢 = Smooth scaling\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		writeText(d.agentsText, "• 🟡 = Adjusting response\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		writeText(d.agentsText, "• 🔴 = Overreacting together\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	default:
		writeText(d.agentsText, "• Sync = Timing coordination\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		writeText(d.agentsText, "• Queue = Work pending\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		writeText(d.agentsText, "• Goal: Self-organize pattern\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	}
	writeText(d.agentsText, "\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	// Header with more intuitive labels
	writeText(d.agentsText, "Workload      Sync   Queue  Sent\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorCyan), cell.Bold()))
	writeText(d.agentsText, "             (beat) (tasks) (batches)\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	writeText(d.agentsText, strings.Repeat("─", 39)+"\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorGray)))

	// Show first 10 agents
	maxShow := 10
	if len(agents) < maxShow {
		maxShow = len(agents)
	}

	for i := range maxShow {
		a := agents[i]

		// Goal-aware sync indicator
		var syncIndicator string
		var phaseColor cell.Color

		switch d.currentGoal {
		case goal.DistributeLoad, goal.RecoverFromFailure:
			// These goals want ANTI-PHASE (spread out)
			// Red when synchronized, green when distributed
			switch {
			case math.Abs(a.Phase) < 0.5 || math.Abs(a.Phase-2*math.Pi) < 0.5:
				syncIndicator = syncRed // BAD - too synchronized
				phaseColor = cell.ColorRed
			case math.Abs(a.Phase-math.Pi) < 0.5 || math.Abs(a.Phase-math.Pi/2) < 0.5 || math.Abs(a.Phase-3*math.Pi/2) < 0.5:
				syncIndicator = syncGreen // GOOD - distributed
				phaseColor = cell.ColorGreen
			default:
				syncIndicator = syncYellow // Getting there
				phaseColor = cell.ColorYellow
			}

		case goal.SaveEnergy:
			// Wants SPARSE sync (some coordination but not tight)
			// Best when some agents rest while others work
			switch {
			case math.Abs(a.Phase) < 0.3 || math.Abs(a.Phase-2*math.Pi) < 0.3:
				syncIndicator = syncYellow // OK - partially synced
				phaseColor = cell.ColorYellow
			case math.Abs(a.Phase-math.Pi) < 0.5:
				syncIndicator = syncGreen // GOOD - taking turns
				phaseColor = cell.ColorGreen
			default:
				syncIndicator = syncRed // BAD - no coordination
				phaseColor = cell.ColorRed
			}

		case goal.ReachConsensus:
			// Wants CLUSTERS (voting blocs)
			// Green when in a cluster with others
			clustered := math.Abs(a.Phase) < 0.5 || math.Abs(a.Phase-2) < 0.5 || math.Abs(a.Phase-4) < 0.5
			switch {
			case clustered:
				syncIndicator = syncGreen // GOOD - in a voting bloc
				phaseColor = cell.ColorGreen
			default:
				syncIndicator = syncYellow // Forming blocs
				phaseColor = cell.ColorYellow
			}

		case goal.MinimizeAPICalls, goal.MinimizeLatency, goal.MaintainRhythm, goal.AdaptToTraffic:
			// Most goals want IN-PHASE (synchronized)
			// MinimizeAPICalls, MinimizeLatency, MaintainRhythm, AdaptToTraffic
			switch {
			case math.Abs(a.Phase) < 0.5 || math.Abs(a.Phase-2*math.Pi) < 0.5:
				syncIndicator = syncGreen // GOOD - synchronized
				phaseColor = cell.ColorGreen
			case math.Abs(a.Phase) < 1.0:
				syncIndicator = syncYellow // Almost there
				phaseColor = cell.ColorYellow
			default:
				syncIndicator = syncRed // BAD - not synced
				phaseColor = cell.ColorRed
			}

		default:
			// Fallback for any future goals
			syncIndicator = syncYellow
			phaseColor = cell.ColorYellow
		}

		// Write agent info - safely truncate type name if needed
		typeDisplay := a.Type
		if len(typeDisplay) > 11 {
			typeDisplay = typeDisplay[:11]
		}
		writeText(d.agentsText, fmt.Sprintf("%s %-11s ", a.Icon, typeDisplay),
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

		// Sync indicator (more intuitive than phase number)
		writeText(d.agentsText, fmt.Sprintf("%s   ", syncIndicator),
			text.WriteCellOpts(cell.FgColor(phaseColor)))

		// Pending tasks with visual indicator
		taskIndicator := ""
		if a.PendingTasks > 10 {
			taskIndicator = "!"
		}
		writeText(d.agentsText, fmt.Sprintf("%4d%s  ", a.PendingTasks, taskIndicator),
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

		// Batches sent
		writeText(d.agentsText, fmt.Sprintf("%4d\n", a.BatchesSent),
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	}

	if len(agents) > maxShow {
		writeText(d.agentsText, fmt.Sprintf("\n... and %d more\n", len(agents)-maxShow),
			text.WriteCellOpts(cell.FgColor(cell.ColorYellow)))
	}
}

// updateMetricsDisplay updates the metrics panel
func (d *TerminalDisplay) updateMetricsDisplay(snapshot SimulationSnapshot) {
	// Route to goal-specific display
	switch d.currentGoal {
	case goal.MinimizeAPICalls:
		d.updateMetricsForBatching(snapshot)
	case goal.DistributeLoad:
		d.updateMetricsForLoadDistribution(snapshot)
	case goal.ReachConsensus:
		d.updateMetricsForConsensus(snapshot)
	case goal.MinimizeLatency:
		d.updateMetricsForLatency(snapshot)
	case goal.SaveEnergy:
		d.updateMetricsForEnergy(snapshot)
	case goal.MaintainRhythm:
		d.updateMetricsForRhythm(snapshot)
	case goal.RecoverFromFailure:
		d.updateMetricsForFailure(snapshot)
	case goal.AdaptToTraffic:
		d.updateMetricsForTraffic(snapshot)
	default:
		d.updateMetricsForBatching(snapshot) // Default to batching display
	}
}

// Common status display for all goals
func (d *TerminalDisplay) displayCommonStatus(snapshot SimulationSnapshot) {
	// Always show status indicators
	writeText(d.metricsText, "STATUS INDICATORS\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorCyan), cell.Bold()))
	writeText(d.metricsText, strings.Repeat("─", 20)+"\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorGray)))

	// Paused status
	writeText(d.metricsText, "Paused:    ",
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	if snapshot.Paused {
		writeText(d.metricsText, "YES\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorYellow), cell.Bold()))
	} else {
		writeText(d.metricsText, "NO\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorGreen)))
	}

	// Disrupted status
	writeText(d.metricsText, "Disrupted: ",
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	if snapshot.Disrupted {
		writeText(d.metricsText, "YES\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorRed), cell.Bold()))
	} else {
		writeText(d.metricsText, "NO\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorGreen)))
	}

	// Reset status
	writeText(d.metricsText, "Resetting: ",
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	if snapshot.Reset {
		writeText(d.metricsText, "YES\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorCyan), cell.Bold()))
	} else {
		writeText(d.metricsText, "NO\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorGreen)))
	}

	writeText(d.metricsText, "\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
}

// updateMetricsForBatching shows metrics for MinimizeAPICalls goal
func (d *TerminalDisplay) updateMetricsForBatching(snapshot SimulationSnapshot) {
	d.metricsText.Reset()
	d.displayCommonStatus(snapshot)

	// WHAT'S HAPPENING
	writeText(d.metricsText, "WHAT'S HAPPENING\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorYellow), cell.Bold()))
	writeText(d.metricsText, strings.Repeat("─", 20)+"\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorGray)))

	// Synchronization status with explanation
	coherenceColor := cell.ColorRed
	syncStatus := "🔴 Not synced"
	syncExplain := "(sending separately)"
	if snapshot.Coherence >= 0.9 {
		coherenceColor = cell.ColorGreen
		syncStatus = "🟢 In sync!"
		syncExplain = "(sending together)"
	} else if snapshot.Coherence >= 0.7 {
		coherenceColor = cell.ColorYellow
		syncStatus = "🟡 Syncing..."
		syncExplain = "(getting aligned)"
	}
	writeText(d.metricsText, fmt.Sprintf("%s %.0f%%\n", syncStatus, snapshot.Coherence*100),
		text.WriteCellOpts(cell.FgColor(coherenceColor)))
	writeText(d.metricsText, fmt.Sprintf("  %s\n\n", syncExplain),
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	// Current activity
	writeText(d.metricsText, "BATCH ACTIVITY\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorCyan), cell.Bold()))
	writeText(d.metricsText, strings.Repeat("─", 20)+"\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorGray)))

	// Tasks accumulating
	if snapshot.CurrentBatchSize > 0 {
		writeText(d.metricsText, fmt.Sprintf("📦 Collecting: %d tasks\n", snapshot.CurrentBatchSize),
			text.WriteCellOpts(cell.FgColor(cell.ColorYellow)))
		writeText(d.metricsText, "   (waiting for beat)\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	} else {
		writeText(d.metricsText, "⏳ Waiting for tasks...\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	}

	// Batches sent
	writeText(d.metricsText, fmt.Sprintf("✅ Sent: %d batches\n", snapshot.BatchesProcessed),
		text.WriteCellOpts(cell.FgColor(cell.ColorGreen)))

	// Show batch pulse if just sent
	if snapshot.BatchJustSent {
		writeText(d.metricsText, fmt.Sprintf("\n✨ BATCH SENT! (%d items)\n", snapshot.LastBatchSize),
			text.WriteCellOpts(cell.FgColor(cell.ColorYellow), cell.Bold()))
	}
}

// updateMetricsForLoadDistribution shows metrics for DistributeLoad goal
func (d *TerminalDisplay) updateMetricsForLoadDistribution(snapshot SimulationSnapshot) {
	d.metricsText.Reset()
	d.displayCommonStatus(snapshot)

	// LOAD DISTRIBUTION
	writeText(d.metricsText, "LOAD DISTRIBUTION\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorYellow), cell.Bold()))
	writeText(d.metricsText, strings.Repeat("─", 20)+"\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorGray)))

	// Anti-phase level (LOW coherence is good for distribution)
	distributionColor := cell.ColorGreen
	distributionStatus := "BALANCED"
	if snapshot.Coherence > 0.5 {
		distributionColor = cell.ColorRed
		distributionStatus = "IMBALANCED"
	} else if snapshot.Coherence > 0.3 {
		distributionColor = cell.ColorYellow
		distributionStatus = "FAIR"
	}

	writeText(d.metricsText, fmt.Sprintf("Distribution: %s\n", distributionStatus),
		text.WriteCellOpts(cell.FgColor(distributionColor), cell.Bold()))
	writeText(d.metricsText, fmt.Sprintf("Phase Spread: %.1f%%\n", (1-snapshot.Coherence)*100),
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	// Calculate load variance
	loadPerAgent := 100.0 / float64(len(snapshot.Agents))
	writeText(d.metricsText, fmt.Sprintf("Load/Server: %.1f%%\n", loadPerAgent),
		text.WriteCellOpts(cell.FgColor(cell.ColorCyan)))
}

// updateMetricsForConsensus shows metrics for ReachConsensus goal
func (d *TerminalDisplay) updateMetricsForConsensus(snapshot SimulationSnapshot) {
	d.metricsText.Reset()
	d.displayCommonStatus(snapshot)

	// CONSENSUS STATUS
	writeText(d.metricsText, "CONSENSUS STATUS\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorYellow), cell.Bold()))
	writeText(d.metricsText, strings.Repeat("─", 20)+"\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorGray)))

	// Agreement level (HIGH coherence = consensus)
	consensusColor := cell.ColorRed
	consensusStatus := "NO CONSENSUS"
	if snapshot.Coherence >= 0.95 {
		consensusColor = cell.ColorGreen
		consensusStatus = "CONSENSUS!"
	} else if snapshot.Coherence >= 0.8 {
		consensusColor = cell.ColorYellow
		consensusStatus = "CONVERGING"
	}

	writeText(d.metricsText, fmt.Sprintf("Status: %s\n", consensusStatus),
		text.WriteCellOpts(cell.FgColor(consensusColor), cell.Bold()))
	writeText(d.metricsText, fmt.Sprintf("Agreement: %.1f%%\n", snapshot.Coherence*100),
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	// Voting rounds (simulated)
	rounds := int(snapshot.ElapsedTime.Seconds() * 2)
	writeText(d.metricsText, fmt.Sprintf("Voting Rounds: %d\n", rounds),
		text.WriteCellOpts(cell.FgColor(cell.ColorCyan)))
}

// updateMetricsForLatency shows metrics for MinimizeLatency goal
func (d *TerminalDisplay) updateMetricsForLatency(snapshot SimulationSnapshot) {
	d.metricsText.Reset()
	d.displayCommonStatus(snapshot)

	// LATENCY METRICS
	writeText(d.metricsText, "LATENCY METRICS\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorYellow), cell.Bold()))
	writeText(d.metricsText, strings.Repeat("─", 20)+"\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorGray)))

	// Response time predictability (HIGH coherence = predictable)
	latencyColor := cell.ColorRed
	latencyStatus := "UNPREDICTABLE"
	if snapshot.Coherence >= 0.9 {
		latencyColor = cell.ColorGreen
		latencyStatus = "PREDICTABLE"
	} else if snapshot.Coherence >= 0.7 {
		latencyColor = cell.ColorYellow
		latencyStatus = "FAIR"
	}

	writeText(d.metricsText, fmt.Sprintf("Timing: %s\n", latencyStatus),
		text.WriteCellOpts(cell.FgColor(latencyColor), cell.Bold()))

	// Simulated latency metrics
	baseLatency := 10.0 // ms
	jitter := (1 - snapshot.Coherence) * 20.0
	writeText(d.metricsText, fmt.Sprintf("P50 Latency: %.1fms\n", baseLatency),
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	writeText(d.metricsText, fmt.Sprintf("P95 Latency: %.1fms\n", baseLatency+jitter),
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	writeText(d.metricsText, fmt.Sprintf("Jitter: ±%.1fms\n", jitter/2),
		text.WriteCellOpts(cell.FgColor(cell.ColorCyan)))
}

// updateMetricsForEnergy shows metrics for SaveEnergy goal
func (d *TerminalDisplay) updateMetricsForEnergy(snapshot SimulationSnapshot) {
	d.metricsText.Reset()
	d.displayCommonStatus(snapshot)

	// ENERGY METRICS
	writeText(d.metricsText, "ENERGY METRICS\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorYellow), cell.Bold()))
	writeText(d.metricsText, strings.Repeat("─", 20)+"\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorGray)))

	// Count quiet vs active sensors
	quietCount := 0
	for _, agent := range snapshot.Agents {
		if agent.ActivityLevel == activityQuiet {
			quietCount++
		}
	}
	idlePercent := float64(quietCount) / float64(len(snapshot.Agents)) * 100

	energyColor := cell.ColorRed
	energyStatus := "HIGH DRAIN"
	if idlePercent > 70 {
		energyColor = cell.ColorGreen
		energyStatus = "EFFICIENT"
	} else if idlePercent > 50 {
		energyColor = cell.ColorYellow
		energyStatus = "MODERATE"
	}

	writeText(d.metricsText, fmt.Sprintf("Energy Use: %s\n", energyStatus),
		text.WriteCellOpts(cell.FgColor(energyColor), cell.Bold()))
	writeText(d.metricsText, fmt.Sprintf("Sensors Idle: %.0f%%\n", idlePercent),
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	// Simulated battery life
	batteryHours := 24 * (idlePercent/100 + 0.5)
	writeText(d.metricsText, fmt.Sprintf("Est. Battery: %.0fh\n", batteryHours),
		text.WriteCellOpts(cell.FgColor(cell.ColorCyan)))
}

// updateMetricsForRhythm shows metrics for MaintainRhythm goal
func (d *TerminalDisplay) updateMetricsForRhythm(snapshot SimulationSnapshot) {
	d.metricsText.Reset()
	d.displayCommonStatus(snapshot)

	// RHYTHM METRICS
	writeText(d.metricsText, "RHYTHM METRICS\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorYellow), cell.Bold()))
	writeText(d.metricsText, strings.Repeat("─", 20)+"\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorGray)))

	// Schedule adherence (HIGH coherence = on schedule)
	rhythmColor := cell.ColorRed
	rhythmStatus := "OFF SCHEDULE"
	if snapshot.Coherence >= 0.9 {
		rhythmColor = cell.ColorGreen
		rhythmStatus = "ON SCHEDULE"
	} else if snapshot.Coherence >= 0.7 {
		rhythmColor = cell.ColorYellow
		rhythmStatus = "DRIFTING"
	}

	writeText(d.metricsText, fmt.Sprintf("Status: %s\n", rhythmStatus),
		text.WriteCellOpts(cell.FgColor(rhythmColor), cell.Bold()))
	writeText(d.metricsText, fmt.Sprintf("Phase Lock: %.1f%%\n", snapshot.Coherence*100),
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	// Timing drift
	drift := (1 - snapshot.Coherence) * 1000 // ms
	writeText(d.metricsText, fmt.Sprintf("Timing Drift: ±%.0fms\n", drift),
		text.WriteCellOpts(cell.FgColor(cell.ColorCyan)))
}

// updateMetricsForFailure shows metrics for RecoverFromFailure goal
func (d *TerminalDisplay) updateMetricsForFailure(snapshot SimulationSnapshot) {
	d.metricsText.Reset()
	d.displayCommonStatus(snapshot)

	// RESILIENCE METRICS
	writeText(d.metricsText, "RESILIENCE METRICS\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorYellow), cell.Bold()))
	writeText(d.metricsText, strings.Repeat("─", 20)+"\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorGray)))

	// Recovery readiness
	resilienceColor := cell.ColorRed
	resilienceStatus := "VULNERABLE"
	recoveryTime := "N/A"

	switch {
	case snapshot.Disrupted:
		resilienceStatus = "RECOVERING"
		resilienceColor = cell.ColorYellow
		recoveryTime = fmt.Sprintf("%.1fs", time.Since(snapshot.LastBatchTime).Seconds())
	case snapshot.Coherence >= 0.8:
		resilienceColor = cell.ColorGreen
		resilienceStatus = "RESILIENT"
		recoveryTime = "<1s"
	case snapshot.Coherence >= 0.6:
		resilienceColor = cell.ColorYellow
		resilienceStatus = "PARTIAL"
		recoveryTime = "~2s"
	}

	writeText(d.metricsText, fmt.Sprintf("Status: %s\n", resilienceStatus),
		text.WriteCellOpts(cell.FgColor(resilienceColor), cell.Bold()))
	writeText(d.metricsText, fmt.Sprintf("Sync Health: %.1f%%\n", snapshot.Coherence*100),
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	writeText(d.metricsText, fmt.Sprintf("Recovery Time: %s\n", recoveryTime),
		text.WriteCellOpts(cell.FgColor(cell.ColorCyan)))
}

// updateMetricsForTraffic shows metrics for AdaptToTraffic goal
func (d *TerminalDisplay) updateMetricsForTraffic(snapshot SimulationSnapshot) {
	d.metricsText.Reset()
	d.displayCommonStatus(snapshot)

	// TRAFFIC ADAPTATION
	writeText(d.metricsText, "TRAFFIC ADAPTATION\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorYellow), cell.Bold()))
	writeText(d.metricsText, strings.Repeat("─", 20)+"\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorGray)))

	// Count active agents as load indicator
	activeCount := 0
	for _, agent := range snapshot.Agents {
		if agent.ActivityLevel == activityBurst || agent.ActivityLevel == "active" {
			activeCount++
		}
	}
	loadPercent := float64(activeCount) / float64(len(snapshot.Agents)) * 100

	// Adaptation quality
	adaptColor := cell.ColorRed
	adaptStatus := "SLOW"
	if snapshot.Coherence >= 0.8 && loadPercent > 30 {
		adaptColor = cell.ColorGreen
		adaptStatus = "RESPONSIVE"
	} else if snapshot.Coherence >= 0.6 {
		adaptColor = cell.ColorYellow
		adaptStatus = "ADAPTING"
	}

	writeText(d.metricsText, fmt.Sprintf("Adaptation: %s\n", adaptStatus),
		text.WriteCellOpts(cell.FgColor(adaptColor), cell.Bold()))
	writeText(d.metricsText, fmt.Sprintf("Current Load: %.0f%%\n", loadPercent),
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	writeText(d.metricsText, fmt.Sprintf("Scaled Units: %d/%d\n", activeCount, len(snapshot.Agents)),
		text.WriteCellOpts(cell.FgColor(cell.ColorCyan)))
}

// updateCostDisplay updates the economics panel
func (d *TerminalDisplay) updateCostDisplay(snapshot SimulationSnapshot) {
	d.costText.Reset()

	// Cost comparison
	writeText(d.costText, "💰 COST ANALYSIS\n\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorYellow), cell.Bold()))

	writeText(d.costText, "Without Synchronization:\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
	writeText(d.costText, fmt.Sprintf("  $%.2f (individual calls)\n\n", snapshot.CostWithoutSync),
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	writeText(d.costText, "With Synchronization:\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorGreen)))
	writeText(d.costText, fmt.Sprintf("  $%.2f (batched calls)\n\n", snapshot.CostWithSync),
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	// Savings
	savingsColor := cell.ColorGreen
	if snapshot.SavingsPercent < 50 {
		savingsColor = cell.ColorYellow
	}
	if snapshot.SavingsPercent < 20 {
		savingsColor = cell.ColorRed
	}

	writeText(d.costText, "💎 SAVINGS: ",
		text.WriteCellOpts(cell.FgColor(cell.ColorCyan), cell.Bold()))
	writeText(d.costText, fmt.Sprintf("$%.2f (%.1f%%)\n",
		snapshot.Savings, snapshot.SavingsPercent),
		text.WriteCellOpts(cell.FgColor(savingsColor), cell.Bold()))
}

// updateGoalSpecificGauges updates gauges based on the current goal
func (d *TerminalDisplay) updateGoalSpecificGauges(snapshot SimulationSnapshot) {
	switch d.currentGoal {
	case goal.MinimizeAPICalls:
		// Show synchronization level (high = good)
		coherencePercent := int(snapshot.Coherence * 100)
		targetPercent := int(snapshot.TargetCoherence * 100)
		gaugeColor := cell.ColorRed
		if snapshot.Coherence >= snapshot.TargetCoherence {
			gaugeColor = cell.ColorGreen
		}
		if err := d.coherenceGauge.Percent(coherencePercent, gauge.TextLabel("Sync Level"), gauge.Color(gaugeColor)); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: coherence gauge update failed: %v\n", err)
		}
		if err := d.targetGauge.Percent(targetPercent, gauge.TextLabel("Target")); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: target gauge update failed: %v\n", err)
		}

	case goal.DistributeLoad:
		// Show load distribution (low coherence = good, means distributed)
		distributionPercent := int((1 - snapshot.Coherence) * 100) // Invert: high distribution = low coherence
		targetDistribution := int((1 - snapshot.TargetCoherence) * 100)
		gaugeColor := cell.ColorGreen
		if snapshot.Coherence > 0.5 { // Bad if too synchronized
			gaugeColor = cell.ColorRed
		}
		if err := d.coherenceGauge.Percent(distributionPercent, gauge.TextLabel("Distribution"), gauge.Color(gaugeColor)); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: distribution gauge update failed: %v\n", err)
		}
		if err := d.targetGauge.Percent(targetDistribution, gauge.TextLabel("Target Spread")); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: target spread gauge update failed: %v\n", err)
		}

	case goal.ReachConsensus:
		// Show agreement level (high = consensus reached)
		agreementPercent := int(snapshot.Coherence * 100)
		targetPercent := 95 // Consensus needs very high agreement
		gaugeColor := cell.ColorRed
		if agreementPercent >= targetPercent {
			gaugeColor = cell.ColorGreen
		} else if agreementPercent >= 80 {
			gaugeColor = cell.ColorYellow
		}
		if err := d.coherenceGauge.Percent(agreementPercent, gauge.TextLabel("Agreement"), gauge.Color(gaugeColor)); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: agreement gauge update failed: %v\n", err)
		}
		if err := d.targetGauge.Percent(targetPercent, gauge.TextLabel("Consensus")); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: consensus gauge update failed: %v\n", err)
		}

	case goal.MinimizeLatency:
		// Show timing predictability (high coherence = predictable)
		predictabilityPercent := int(snapshot.Coherence * 100)
		targetPercent := int(snapshot.TargetCoherence * 100)
		gaugeColor := cell.ColorRed
		if predictabilityPercent >= targetPercent {
			gaugeColor = cell.ColorGreen
		}
		if err := d.coherenceGauge.Percent(predictabilityPercent,
			gauge.TextLabel("Predictability"), gauge.Color(gaugeColor)); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: predictability gauge update failed: %v\n", err)
		}
		if err := d.targetGauge.Percent(targetPercent, gauge.TextLabel("Target")); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: target gauge update failed: %v\n", err)
		}

	case goal.SaveEnergy:
		// Show idle percentage (more idle = better)
		idleCount := 0
		for _, agent := range snapshot.Agents {
			if agent.ActivityLevel == activityQuiet {
				idleCount++
			}
		}
		idlePercent := int(float64(idleCount) / float64(len(snapshot.Agents)) * 100)
		targetPercent := 70 // Target 70% idle for energy saving
		gaugeColor := cell.ColorRed
		if idlePercent >= targetPercent {
			gaugeColor = cell.ColorGreen
		} else if idlePercent >= 50 {
			gaugeColor = cell.ColorYellow
		}
		if err := d.coherenceGauge.Percent(idlePercent, gauge.TextLabel("Idle %"), gauge.Color(gaugeColor)); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: idle gauge update failed: %v\n", err)
		}
		if err := d.targetGauge.Percent(targetPercent, gauge.TextLabel("Target Idle")); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: target idle gauge update failed: %v\n", err)
		}

	case goal.MaintainRhythm:
		// Show phase lock quality (high = on schedule)
		rhythmPercent := int(snapshot.Coherence * 100)
		targetPercent := int(snapshot.TargetCoherence * 100)
		gaugeColor := cell.ColorRed
		if rhythmPercent >= targetPercent {
			gaugeColor = cell.ColorGreen
		}
		if err := d.coherenceGauge.Percent(rhythmPercent, gauge.TextLabel("Phase Lock"), gauge.Color(gaugeColor)); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: phase lock gauge update failed: %v\n", err)
		}
		if err := d.targetGauge.Percent(targetPercent, gauge.TextLabel("Target")); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: target gauge update failed: %v\n", err)
		}

	case goal.RecoverFromFailure:
		// Show system health
		healthPercent := int(snapshot.Coherence * 100)
		if snapshot.Disrupted {
			healthPercent /= 2 // Show degraded health during disruption
		}
		targetPercent := 80 // Need 80% health for resilience
		gaugeColor := cell.ColorRed
		if healthPercent >= targetPercent && !snapshot.Disrupted {
			gaugeColor = cell.ColorGreen
		} else if healthPercent >= 60 {
			gaugeColor = cell.ColorYellow
		}
		if err := d.coherenceGauge.Percent(healthPercent, gauge.TextLabel("System Health"), gauge.Color(gaugeColor)); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: health gauge update failed: %v\n", err)
		}
		if err := d.targetGauge.Percent(targetPercent, gauge.TextLabel("Healthy")); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: healthy gauge update failed: %v\n", err)
		}

	case goal.AdaptToTraffic:
		// Show capacity utilization
		activeCount := 0
		for _, agent := range snapshot.Agents {
			if agent.ActivityLevel == activityBurst || agent.ActivityLevel == "active" {
				activeCount++
			}
		}
		utilizationPercent := int(float64(activeCount) / float64(len(snapshot.Agents)) * 100)
		targetPercent := 60 // Optimal utilization around 60%
		gaugeColor := cell.ColorYellow
		if utilizationPercent >= 40 && utilizationPercent <= 80 {
			gaugeColor = cell.ColorGreen
		} else if utilizationPercent > 90 {
			gaugeColor = cell.ColorRed
		}
		if err := d.coherenceGauge.Percent(utilizationPercent, gauge.TextLabel("Utilization"), gauge.Color(gaugeColor)); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: utilization gauge update failed: %v\n", err)
		}
		if err := d.targetGauge.Percent(targetPercent, gauge.TextLabel("Optimal")); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: optimal gauge update failed: %v\n", err)
		}

	default:
		// Default to coherence display
		coherencePercent := int(snapshot.Coherence * 100)
		targetPercent := int(snapshot.TargetCoherence * 100)
		gaugeColor := cell.ColorRed
		if snapshot.Coherence >= snapshot.TargetCoherence {
			gaugeColor = cell.ColorGreen
		}
		if err := d.coherenceGauge.Percent(coherencePercent, gauge.TextLabel("Coherence"), gauge.Color(gaugeColor)); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: coherence gauge update failed: %v\n", err)
		}
		if err := d.targetGauge.Percent(targetPercent, gauge.TextLabel("Target")); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: target gauge update failed: %v\n", err)
		}
	}
}

// updateCoherenceChart updates the coherence chart
func (d *TerminalDisplay) updateCoherenceChart(coherence, targetCoherence float64, paused bool) {
	// Don't add new points if paused - just redraw with existing data
	if !paused {
		// Add to history
		d.coherenceHistory = append(d.coherenceHistory, coherence*100)
		d.targetHistory = append(d.targetHistory, targetCoherence*100)

		// Keep only last 50 points
		if len(d.coherenceHistory) > 50 {
			d.coherenceHistory = d.coherenceHistory[1:]
			d.targetHistory = d.targetHistory[1:]
		}
	}

	// Create separate series for above and below target
	// We'll overlap at transition points to maintain continuity
	aboveTarget := make([]float64, len(d.coherenceHistory))
	belowTarget := make([]float64, len(d.coherenceHistory))

	targetValue := targetCoherence * 100
	for i, val := range d.coherenceHistory {
		isAbove := val >= targetValue

		// Always set the current value in the appropriate series
		if isAbove {
			aboveTarget[i] = val
			belowTarget[i] = math.NaN()
		} else {
			belowTarget[i] = val
			aboveTarget[i] = math.NaN()
		}

		// At transition points, also set the value in both series to connect them
		if i > 0 {
			prevVal := d.coherenceHistory[i-1]
			wasAbove := prevVal >= targetValue

			// If we crossed the threshold, duplicate the value at the boundary
			if isAbove != wasAbove {
				// Set both to create connection point
				aboveTarget[i] = val
				belowTarget[i] = val
				// Also update the previous point to connect
				if isAbove {
					// Transitioned from below to above
					belowTarget[i-1] = prevVal
				} else {
					// Transitioned from above to below
					aboveTarget[i-1] = prevVal
				}
			}
		}
	}

	// Draw the target line first (yellow)
	if err := d.coherenceChart.Series("target", d.targetHistory,
		linechart.SeriesCellOpts(cell.FgColor(cell.ColorYellow))); err != nil {
		fmt.Printf("Failed to update target line: %v\n", err)
	}

	// Draw below-target values in red
	if err := d.coherenceChart.Series("below", belowTarget,
		linechart.SeriesCellOpts(cell.FgColor(cell.ColorRed))); err != nil {
		fmt.Printf("Failed to update below-target series: %v\n", err)
	}

	// Draw above-target values in green
	if err := d.coherenceChart.Series("above", aboveTarget,
		linechart.SeriesCellOpts(cell.FgColor(cell.ColorGreen))); err != nil {
		fmt.Printf("Failed to update above-target series: %v\n", err)
	}
}

// calculateMeanField computes the mean field parameters for coherence
func calculateMeanField(agents []AgentSnapshot) (meanPhase, meanMagnitude float64) {
	var sumSin, sumCos float64
	for _, agent := range agents {
		sumSin += math.Sin(agent.Phase)
		sumCos += math.Cos(agent.Phase)
	}
	n := float64(len(agents))
	if n > 0 {
		sumSin /= n
		sumCos /= n
	}
	meanPhase = math.Atan2(sumSin, sumCos)
	meanMagnitude = math.Sqrt(sumSin*sumSin + sumCos*sumCos)
	return
}

// getAgentColor determines the color for an agent based on coherence
func getAgentColor(phase, meanPhase, meanMagnitude float64) cell.Color {
	if meanMagnitude < 0.3 {
		// Low coherence - color by phase group
		phaseGroup := int((phase + math.Pi) * 3 / (2 * math.Pi))
		switch phaseGroup {
		case 0:
			return cell.ColorRed
		case 1:
			return cell.ColorYellow
		case 2:
			return cell.ColorCyan
		default:
			return cell.ColorMagenta
		}
	}

	// Higher coherence - color by alignment with mean field
	deviation := math.Abs(phase - meanPhase)
	if deviation > math.Pi {
		deviation = 2*math.Pi - deviation
	}

	if deviation < math.Pi*(1-meanMagnitude) {
		return cell.ColorGreen
	}
	return cell.ColorRed
}

// updateSwarmDisplay updates the swarm dynamics panel with goal-specific visualization
func (d *TerminalDisplay) updateSwarmDisplay(snapshot SimulationSnapshot) {
	d.swarmText.Reset()

	switch d.currentGoal {
	case goal.MinimizeAPICalls:
		d.displayBatchingWaves(snapshot)
	case goal.DistributeLoad:
		d.displayLoadDistribution(snapshot)
	case goal.ReachConsensus:
		d.displayConsensusProgress(snapshot)
	case goal.MinimizeLatency:
		d.displayLatencyHistogram(snapshot)
	case goal.SaveEnergy:
		d.displayEnergyPattern(snapshot)
	case goal.MaintainRhythm:
		d.displayScheduleAlignment(snapshot)
	case goal.RecoverFromFailure:
		d.displaySystemHealth(snapshot)
	case goal.AdaptToTraffic:
		d.displayTrafficLoad(snapshot)
	default:
		d.displayBatchingWaves(snapshot) // Default to phase waves
	}
}

// displayBatchingWaves shows phase convergence for MinimizeAPICalls
func (d *TerminalDisplay) displayBatchingWaves(snapshot SimulationSnapshot) {
	writeText(d.swarmText, "🌊 PHASE CONVERGENCE\n\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorCyan), cell.Bold()))

	barWidth := 35
	barHeight := 8
	meanPhase, meanMagnitude := calculateMeanField(snapshot.Agents)

	for row := barHeight; row > 0; row-- {
		for i := 0; i < len(snapshot.Agents) && i < barWidth; i++ {
			agent := snapshot.Agents[i]
			normalizedPhase := (math.Sin(agent.Phase) + 1) / 2
			agentHeight := int(normalizedPhase * float64(barHeight))

			if agentHeight >= row {
				color := getAgentColor(agent.Phase, meanPhase, meanMagnitude)
				writeText(d.swarmText, "█",
					text.WriteCellOpts(cell.FgColor(color)))
			} else {
				writeText(d.swarmText, " ")
			}
		}
		writeText(d.swarmText, "\n")
	}

	writeText(d.swarmText, strings.Repeat("─", 35), text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
	writeText(d.swarmText, fmt.Sprintf("\nSync: %.1f%% ", snapshot.Coherence*100),
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	if snapshot.Coherence >= 0.9 {
		writeText(d.swarmText, "✓ BATCHING", text.WriteCellOpts(cell.FgColor(cell.ColorGreen)))
	}
}

// displayLoadDistribution shows anti-phase pattern for DistributeLoad
func (d *TerminalDisplay) displayLoadDistribution(snapshot SimulationSnapshot) {
	writeText(d.swarmText, "⚖️ LOAD DISTRIBUTION\n\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorYellow), cell.Bold()))

	// Show load across different phase regions
	phaseRegions := 4
	loadPerRegion := make([]int, phaseRegions)

	for _, agent := range snapshot.Agents {
		// Map phase to region
		region := int((agent.Phase + math.Pi) / (2 * math.Pi) * float64(phaseRegions))
		if region >= phaseRegions {
			region = phaseRegions - 1
		}
		loadPerRegion[region]++
	}

	// Display as horizontal bars
	maxAgents := len(snapshot.Agents)/phaseRegions + 1
	for i := range phaseRegions {
		writeText(d.swarmText, fmt.Sprintf("Server%d: ", i+1),
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

		barLen := int(float64(loadPerRegion[i]) / float64(maxAgents) * 20)
		color := cell.ColorGreen
		if loadPerRegion[i] > maxAgents {
			color = cell.ColorRed // Overloaded
		}

		for range barLen {
			writeText(d.swarmText, "█", text.WriteCellOpts(cell.FgColor(color)))
		}
		writeText(d.swarmText, fmt.Sprintf(" %d\n", loadPerRegion[i]),
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	}

	// Show distribution quality
	variance := calculateLoadVariance(loadPerRegion)
	status := "BALANCED"
	color := cell.ColorGreen
	if variance > 2.0 {
		status = "IMBALANCED"
		color = cell.ColorRed
	}
	writeText(d.swarmText, fmt.Sprintf("\nStatus: %s\n", status),
		text.WriteCellOpts(cell.FgColor(color), cell.Bold()))
}

// displayConsensusProgress shows voting progress for ReachConsensus
func (d *TerminalDisplay) displayConsensusProgress(snapshot SimulationSnapshot) {
	writeText(d.swarmText, "🗳️ CONSENSUS PROGRESS\n\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorMagenta), cell.Bold()))

	// Group agents by their phase (representing votes)
	voteGroups := make(map[int]int)
	for _, agent := range snapshot.Agents {
		vote := int(agent.Phase * 2) // Quantize to vote options
		voteGroups[vote]++
	}

	// Find majority vote
	maxVotes := 0
	for _, count := range voteGroups {
		if count > maxVotes {
			maxVotes = count
		}
	}

	agreementPercent := float64(maxVotes) / float64(len(snapshot.Agents)) * 100

	// Show voting bar
	writeText(d.swarmText, "Agreement: [", text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	barLen := int(agreementPercent / 5)
	for i := range 20 {
		if i < barLen {
			writeText(d.swarmText, "█", text.WriteCellOpts(cell.FgColor(cell.ColorGreen)))
		} else {
			writeText(d.swarmText, "░", text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
		}
	}
	writeText(d.swarmText, fmt.Sprintf("] %.0f%%\n\n", agreementPercent),
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	// Status
	if agreementPercent >= 95 {
		writeText(d.swarmText, "✓ CONSENSUS REACHED!\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorGreen), cell.Bold()))
	} else {
		writeText(d.swarmText, "⟳ Voting in progress...\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorYellow)))
	}
}

// displayLatencyHistogram shows response time distribution for MinimizeLatency
func (d *TerminalDisplay) displayLatencyHistogram(snapshot SimulationSnapshot) {
	writeText(d.swarmText, "⚡ LATENCY DISTRIBUTION\n\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorCyan), cell.Bold()))

	// Simulate latency based on phase variance
	bins := 5
	latencyBins := make([]int, bins)
	baseLatency := 10.0 // ms

	for _, agent := range snapshot.Agents {
		// Phase variance affects latency
		variance := math.Abs(agent.Phase)
		latency := baseLatency + variance*5
		bin := int(latency / 10)
		if bin >= bins {
			bin = bins - 1
		}
		latencyBins[bin]++
	}

	// Display histogram
	labels := []string{"0-10ms", "10-20ms", "20-30ms", "30-40ms", "40ms+"}
	for i := range bins {
		writeText(d.swarmText, fmt.Sprintf("%-8s: ", labels[i]),
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

		barLen := int(float64(latencyBins[i]) / float64(len(snapshot.Agents)) * 30)
		color := cell.ColorGreen
		if i > 2 {
			color = cell.ColorRed // High latency
		} else if i > 1 {
			color = cell.ColorYellow
		}

		for range barLen {
			writeText(d.swarmText, "█", text.WriteCellOpts(cell.FgColor(color)))
		}
		writeText(d.swarmText, fmt.Sprintf(" %d\n", latencyBins[i]),
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	}
}

// displayEnergyPattern shows activity cycles for SaveEnergy
func (d *TerminalDisplay) displayEnergyPattern(snapshot SimulationSnapshot) {
	writeText(d.swarmText, "🔋 ENERGY PATTERN\n\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorGreen), cell.Bold()))

	// Show sensors in different states
	sleeping := 0
	active := 0
	for _, agent := range snapshot.Agents {
		if agent.ActivityLevel == activityQuiet {
			sleeping++
		} else {
			active++
		}
	}

	totalAgents := len(snapshot.Agents)
	sleepPercent := float64(sleeping) / float64(totalAgents) * 100

	// Visualize as battery level
	writeText(d.swarmText, "Battery: [", text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	batteryBars := int(sleepPercent / 5)
	for i := range 20 {
		if i < batteryBars {
			writeText(d.swarmText, "█", text.WriteCellOpts(cell.FgColor(cell.ColorGreen)))
		} else {
			writeText(d.swarmText, "░", text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
		}
	}
	writeText(d.swarmText, fmt.Sprintf("] %.0f%%\n\n", sleepPercent),
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	// Show sensor states
	writeText(d.swarmText, fmt.Sprintf("💤 Sleeping: %d sensors\n", sleeping),
		text.WriteCellOpts(cell.FgColor(cell.ColorBlue)))
	writeText(d.swarmText, fmt.Sprintf("📡 Active: %d sensors\n", active),
		text.WriteCellOpts(cell.FgColor(cell.ColorYellow)))

	// Efficiency status
	if sleepPercent > 70 {
		writeText(d.swarmText, "\n✓ EFFICIENT\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorGreen), cell.Bold()))
	} else {
		writeText(d.swarmText, "\n⚠ HIGH DRAIN\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
	}
}

// displayScheduleAlignment shows drift from schedule for MaintainRhythm
func (d *TerminalDisplay) displayScheduleAlignment(snapshot SimulationSnapshot) {
	writeText(d.swarmText, "⏰ SCHEDULE ALIGNMENT\n\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorMagenta), cell.Bold()))

	// Show timing drift for each job
	onTime := 0
	early := 0
	late := 0

	for _, agent := range snapshot.Agents {
		drift := agent.Phase // Phase represents timing drift
		switch {
		case math.Abs(drift) < 0.1:
			onTime++
		case drift < 0:
			early++
		default:
			late++
		}
	}

	// Display timing chart
	writeText(d.swarmText, "Early    On-Time    Late\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	// Show distribution
	maxCount := maxInt(maxInt(early, onTime), late)
	scaleRatio := 15.0 / float64(maxCount)

	earlyBars := int(float64(early) * scaleRatio)
	onTimeBars := int(float64(onTime) * scaleRatio)
	lateBars := int(float64(late) * scaleRatio)

	for range earlyBars {
		writeText(d.swarmText, "◀", text.WriteCellOpts(cell.FgColor(cell.ColorYellow)))
	}
	writeText(d.swarmText, " ", text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	for range onTimeBars {
		writeText(d.swarmText, "█", text.WriteCellOpts(cell.FgColor(cell.ColorGreen)))
	}
	writeText(d.swarmText, " ", text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	for range lateBars {
		writeText(d.swarmText, "▶", text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
	}
	writeText(d.swarmText, "\n\n", text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	writeText(d.swarmText, fmt.Sprintf("On Schedule: %d/%d\n", onTime, len(snapshot.Agents)),
		text.WriteCellOpts(cell.FgColor(cell.ColorGreen)))
}

// displaySystemHealth shows node health for RecoverFromFailure
func (d *TerminalDisplay) displaySystemHealth(snapshot SimulationSnapshot) {
	writeText(d.swarmText, "🛡️ SYSTEM HEALTH\n\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorCyan), cell.Bold()))

	// Show primary/replica/standby status
	healthy := 0
	degraded := 0
	failed := 0

	for _, agent := range snapshot.Agents {
		switch {
		case snapshot.Disrupted && math.Abs(agent.Phase) > 2:
			failed++
		case math.Abs(agent.Phase) > 1:
			degraded++
		default:
			healthy++
		}
	}

	// Display health bars
	totalAgents := len(snapshot.Agents)

	writeText(d.swarmText, "Healthy:  ", text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	healthyBars := int(float64(healthy) / float64(totalAgents) * 20)
	for range healthyBars {
		writeText(d.swarmText, "█", text.WriteCellOpts(cell.FgColor(cell.ColorGreen)))
	}
	writeText(d.swarmText, fmt.Sprintf(" %d\n", healthy),
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	if degraded > 0 {
		writeText(d.swarmText, "Degraded: ", text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		degradedBars := int(float64(degraded) / float64(totalAgents) * 20)
		for range degradedBars {
			writeText(d.swarmText, "█", text.WriteCellOpts(cell.FgColor(cell.ColorYellow)))
		}
		writeText(d.swarmText, fmt.Sprintf(" %d\n", degraded),
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	}

	if failed > 0 {
		writeText(d.swarmText, "Failed:   ", text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		failedBars := int(float64(failed) / float64(totalAgents) * 20)
		for range failedBars {
			writeText(d.swarmText, "█", text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
		}
		writeText(d.swarmText, fmt.Sprintf(" %d\n", failed),
			text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	}

	// Recovery status
	if snapshot.Disrupted {
		writeText(d.swarmText, "\n⚠ RECOVERING...\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorYellow), cell.Bold()))
	} else if healthy == totalAgents {
		writeText(d.swarmText, "\n✓ FULLY RESILIENT\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorGreen), cell.Bold()))
	}
}

// displayTrafficLoad shows scaling for AdaptToTraffic
func (d *TerminalDisplay) displayTrafficLoad(snapshot SimulationSnapshot) {
	writeText(d.swarmText, "📈 TRAFFIC LOAD\n\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorYellow), cell.Bold()))

	// Count active units
	idle := 0
	normal := 0
	busy := 0

	for _, agent := range snapshot.Agents {
		switch agent.ActivityLevel {
		case activityQuiet:
			idle++
		case activityBurst:
			busy++
		default:
			normal++
		}
	}

	totalAgents := len(snapshot.Agents)
	utilization := float64(normal+busy) / float64(totalAgents) * 100

	// Show capacity bar
	writeText(d.swarmText, "Capacity: [", text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	capacityBars := int(utilization / 5)
	for i := range 20 {
		color := cell.ColorGreen
		if i >= 16 {
			color = cell.ColorRed // Overload
		} else if i >= 12 {
			color = cell.ColorYellow // High load
		}

		if i < capacityBars {
			writeText(d.swarmText, "█", text.WriteCellOpts(cell.FgColor(color)))
		} else {
			writeText(d.swarmText, "░", text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
		}
	}
	writeText(d.swarmText, fmt.Sprintf("] %.0f%%\n\n", utilization),
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	// Show unit states
	writeText(d.swarmText, fmt.Sprintf("🟢 Idle: %d units\n", idle),
		text.WriteCellOpts(cell.FgColor(cell.ColorGreen)))
	writeText(d.swarmText, fmt.Sprintf("🟡 Normal: %d units\n", normal),
		text.WriteCellOpts(cell.FgColor(cell.ColorYellow)))
	writeText(d.swarmText, fmt.Sprintf("🔴 Busy: %d units\n", busy),
		text.WriteCellOpts(cell.FgColor(cell.ColorRed)))

	// Scaling status
	switch {
	case utilization > 80:
		writeText(d.swarmText, "\n⬆ SCALE UP\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorRed), cell.Bold()))
	case utilization < 30:
		writeText(d.swarmText, "\n⬇ SCALE DOWN\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorBlue), cell.Bold()))
	default:
		writeText(d.swarmText, "\n✓ OPTIMAL\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorGreen), cell.Bold()))
	}

	// Key insights
	writeText(d.swarmText, "\n📊 KEY INSIGHTS:\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorCyan), cell.Bold()))

	switch {
	case snapshot.Coherence >= snapshot.TargetCoherence:
		writeText(d.swarmText, "Agents synchronized!\nBatching effective.",
			text.WriteCellOpts(cell.FgColor(cell.ColorGreen)))
	case snapshot.Coherence >= snapshot.TargetCoherence*0.9:
		writeText(d.swarmText, "Agents learning...\nBatching improving.",
			text.WriteCellOpts(cell.FgColor(cell.ColorYellow)))
	default:
		writeText(d.swarmText, "Agents scattered.\nMostly individual calls.",
			text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
	}
}

// createEmptyGrid creates an empty 2D grid
func createEmptyGrid(width, height int) [][]rune {
	g := make([][]rune, height)
	for i := range g {
		g[i] = make([]rune, width)
		for j := range g[i] {
			g[i][j] = ' '
		}
	}
	return g
}

// getAgentCharacter returns the character to display for an agent based on phase
func getAgentCharacter(phase float64) rune {
	switch {
	case math.Abs(phase) < 0.5:
		return '●'
	case math.Abs(phase) < 1.0:
		return 'o'
	default:
		return '∙'
	}
}

// renderGridChar renders a single character with appropriate color
func renderGridChar(w *text.Text, char rune) {
	switch char {
	case '●':
		writeText(w, string(char),
			text.WriteCellOpts(cell.FgColor(cell.ColorGreen)))
	case 'o':
		writeText(w, string(char),
			text.WriteCellOpts(cell.FgColor(cell.ColorYellow)))
	case '∙':
		writeText(w, string(char),
			text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
	case '⊕':
		writeText(w, string(char),
			text.WriteCellOpts(cell.FgColor(cell.ColorCyan)))
	default:
		writeText(w, " ")
	}
}

// calculateLoadVariance calculates the variance in load across agents
func calculateLoadVariance(loads []int) float64 {
	if len(loads) == 0 {
		return 0
	}

	// Calculate mean load
	var total float64
	for _, load := range loads {
		total += float64(load)
	}
	mean := total / float64(len(loads))

	// Calculate variance
	var variance float64
	for _, load := range loads {
		diff := float64(load) - mean
		variance += diff * diff
	}
	variance /= float64(len(loads))

	return math.Sqrt(variance) // Return standard deviation
}

// maxInt returns the maximum of two integers
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// getPatternReason returns a descriptive reason for why a pattern is good or bad for a goal
func getPatternReason(g goal.Type, p pattern.Type, isGood bool) string {
	// Helper function to get reason based on goal and pattern
	getReason := func(goalType goal.Type, pat pattern.Type, good bool) string {
		type reasonKey struct {
			goal    goal.Type
			pattern pattern.Type
			isGood  bool
		}

		reasons := map[reasonKey]string{
			// MinimizeAPICalls - Good patterns
			{goal.MinimizeAPICalls, pattern.Burst, true}:         " (creates natural batching windows)",
			{goal.MinimizeAPICalls, pattern.HighFrequency, true}: " (continuous flow enables steady batching)",
			{goal.MinimizeAPICalls, pattern.Steady, true}:        " (predictable timing for batch optimization)",
			// MinimizeAPICalls - Bad patterns
			{goal.MinimizeAPICalls, pattern.Sparse, false}:        " (too infrequent for effective batching)",
			{goal.MinimizeAPICalls, pattern.Mixed, false}:         " (unpredictable timing disrupts batching)",
			{goal.MinimizeAPICalls, pattern.HighFrequency, false}: " (too frequent for batching)",
			{goal.MinimizeAPICalls, pattern.Burst, false}:         " (irregular batching)",
			{goal.MinimizeAPICalls, pattern.Steady, false}:        " (suboptimal batch sizes)",

			// DistributeLoad - Good patterns
			{goal.DistributeLoad, pattern.Mixed, true}:  " (varied timing spreads load naturally)",
			{goal.DistributeLoad, pattern.Sparse, true}: " (low frequency prevents overload)",
			{goal.DistributeLoad, pattern.Steady, true}: " (predictable for load balancing)",
			// DistributeLoad - Bad patterns
			{goal.DistributeLoad, pattern.Burst, false}:         " (creates load spikes)",
			{goal.DistributeLoad, pattern.HighFrequency, false}: " (may overload servers)",
			{goal.DistributeLoad, pattern.Steady, false}:        " (predictable load patterns)",
			{goal.DistributeLoad, pattern.Mixed, false}:         " (unpredictable load)",
			{goal.DistributeLoad, pattern.Sparse, false}:        " (uneven distribution)",

			// ReachConsensus - Good patterns
			{goal.ReachConsensus, pattern.Burst, true}:         " (concentrated voting rounds)",
			{goal.ReachConsensus, pattern.HighFrequency, true}: " (rapid consensus iterations)",
			{goal.ReachConsensus, pattern.Steady, true}:        " (regular voting cycles)",
			{goal.ReachConsensus, pattern.Mixed, true}:         " (unpredictable voting timing)",
			{goal.ReachConsensus, pattern.Sparse, true}:        " (slow consensus formation)",
			// ReachConsensus - Bad patterns
			{goal.ReachConsensus, pattern.Sparse, false}:        " (too slow for consensus)",
			{goal.ReachConsensus, pattern.Mixed, false}:         " (unpredictable voting timing)",
			{goal.ReachConsensus, pattern.HighFrequency, false}: " (too rapid for consensus)",
			{goal.ReachConsensus, pattern.Burst, false}:         " (irregular consensus attempts)",
			{goal.ReachConsensus, pattern.Steady, false}:        " (may miss optimal timing)",

			// MinimizeLatency - Good patterns
			{goal.MinimizeLatency, pattern.HighFrequency, true}: " (continuous optimization of response times)",
			{goal.MinimizeLatency, pattern.Steady, true}:        " (predictable latency patterns)",
			{goal.MinimizeLatency, pattern.Burst, true}:         " (latency spikes during bursts)",
			{goal.MinimizeLatency, pattern.Mixed, true}:         " (variable latency)",
			{goal.MinimizeLatency, pattern.Sparse, true}:        " (infrequent optimization opportunities)",
			// MinimizeLatency - Bad patterns
			{goal.MinimizeLatency, pattern.Sparse, false}:        " (infrequent optimization opportunities)",
			{goal.MinimizeLatency, pattern.Burst, false}:         " (creates latency spikes)",
			{goal.MinimizeLatency, pattern.HighFrequency, false}: " (may increase overhead)",
			{goal.MinimizeLatency, pattern.Steady, false}:        " (predictable but not minimal)",
			{goal.MinimizeLatency, pattern.Mixed, false}:         " (unpredictable latency)",

			// SaveEnergy - Good patterns
			{goal.SaveEnergy, pattern.Sparse, true}:        " (minimal activity saves power)",
			{goal.SaveEnergy, pattern.Burst, true}:         " (concentrated work, longer idle periods)",
			{goal.SaveEnergy, pattern.HighFrequency, true}: " (constant activity drains battery)",
			{goal.SaveEnergy, pattern.Steady, true}:        " (predictable power usage)",
			{goal.SaveEnergy, pattern.Mixed, true}:         " (unpredictable energy consumption)",
			// SaveEnergy - Bad patterns
			{goal.SaveEnergy, pattern.HighFrequency, false}: " (constant activity drains battery)",
			{goal.SaveEnergy, pattern.Mixed, false}:         " (unpredictable power usage)",
			{goal.SaveEnergy, pattern.Steady, false}:        " (continuous power draw)",
			{goal.SaveEnergy, pattern.Burst, false}:         " (power spikes)",
			{goal.SaveEnergy, pattern.Sparse, false}:        " (minimal but inefficient)",

			// MaintainRhythm - Good patterns
			{goal.MaintainRhythm, pattern.Steady, true}:        " (perfect for rhythm maintenance)",
			{goal.MaintainRhythm, pattern.HighFrequency, true}: " (frequent sync points)",
			{goal.MaintainRhythm, pattern.Burst, true}:         " (intermittent rhythm disruption)",
			{goal.MaintainRhythm, pattern.Mixed, true}:         " (irregular rhythm)",
			{goal.MaintainRhythm, pattern.Sparse, true}:        " (too infrequent for rhythm)",
			// MaintainRhythm - Bad patterns
			{goal.MaintainRhythm, pattern.Sparse, false}:        " (gaps disrupt rhythm)",
			{goal.MaintainRhythm, pattern.Mixed, false}:         " (irregular timing breaks rhythm)",
			{goal.MaintainRhythm, pattern.Burst, false}:         " (intermittent disruptions)",
			{goal.MaintainRhythm, pattern.HighFrequency, false}: " (too fast for stable rhythm)",
			{goal.MaintainRhythm, pattern.Steady, false}:        " (monotonous rhythm)",

			// RecoverFromFailure - Good patterns
			{goal.RecoverFromFailure, pattern.Burst, true}:         " (rapid recovery attempts)",
			{goal.RecoverFromFailure, pattern.HighFrequency, true}: " (continuous health monitoring)",
			{goal.RecoverFromFailure, pattern.Steady, true}:        " (regular health checks)",
			{goal.RecoverFromFailure, pattern.Mixed, true}:         " (unpredictable recovery timing)",
			{goal.RecoverFromFailure, pattern.Sparse, true}:        " (delayed failure detection)",
			// RecoverFromFailure - Bad patterns
			{goal.RecoverFromFailure, pattern.Sparse, false}:        " (slow failure detection)",
			{goal.RecoverFromFailure, pattern.Mixed, false}:         " (unpredictable recovery timing)",
			{goal.RecoverFromFailure, pattern.Burst, false}:         " (irregular recovery attempts)",
			{goal.RecoverFromFailure, pattern.HighFrequency, false}: " (may overwhelm recovery)",
			{goal.RecoverFromFailure, pattern.Steady, false}:        " (predictable but slow)",

			// AdaptToTraffic - Good patterns
			{goal.AdaptToTraffic, pattern.Mixed, true}:         " (simulates real traffic patterns)",
			{goal.AdaptToTraffic, pattern.Burst, true}:         " (tests surge handling)",
			{goal.AdaptToTraffic, pattern.HighFrequency, true}: " (stress testing traffic)",
			{goal.AdaptToTraffic, pattern.Steady, true}:        " (baseline traffic pattern)",
			{goal.AdaptToTraffic, pattern.Sparse, true}:        " (low traffic simulation)",
			// AdaptToTraffic - Bad patterns
			{goal.AdaptToTraffic, pattern.Sparse, false}:        " (insufficient traffic simulation)",
			{goal.AdaptToTraffic, pattern.Steady, false}:        " (doesn't simulate real traffic variability)",
			{goal.AdaptToTraffic, pattern.HighFrequency, false}: " (unrealistic constant load)",
			{goal.AdaptToTraffic, pattern.Burst, false}:         " (too extreme for baseline)",
			{goal.AdaptToTraffic, pattern.Mixed, false}:         " (too unpredictable)",
		}

		if pat == pattern.Unset {
			return ""
		}

		if reason, ok := reasons[reasonKey{goalType, pat, good}]; ok {
			return reason
		}

		// Default reasons
		if good {
			switch goalType {
			case goal.MinimizeAPICalls:
				return " (supports batch formation)"
			case goal.DistributeLoad:
				return " (helps distribute workload)"
			case goal.ReachConsensus:
				return " (supports voting coordination)"
			case goal.MinimizeLatency:
				return " (enables latency optimization)"
			case goal.SaveEnergy:
				return " (supports energy conservation)"
			case goal.MaintainRhythm:
				return " (enables rhythm coordination)"
			case goal.RecoverFromFailure:
				return " (supports failure detection)"
			case goal.AdaptToTraffic:
				return " (enables traffic adaptation)"
			default:
				return " (compatible with goal)"
			}
		} else {
			switch goalType {
			case goal.MinimizeAPICalls:
				return " (suboptimal for batching)"
			case goal.DistributeLoad:
				return " (uneven load distribution)"
			case goal.ReachConsensus:
				return " (hampers consensus formation)"
			case goal.MinimizeLatency:
				return " (suboptimal for latency)"
			case goal.SaveEnergy:
				return " (inefficient energy use)"
			case goal.MaintainRhythm:
				return " (rhythm disruption likely)"
			case goal.RecoverFromFailure:
				return " (hampers recovery efforts)"
			case goal.AdaptToTraffic:
				return " (poor traffic simulation)"
			default:
				return " (not optimal for this goal)"
			}
		}
	}

	return getReason(g, p, isGood)
}

func (d *TerminalDisplay) updateSwarmVisualization(snapshot SimulationSnapshot) {
	d.swarmVisText.Reset()

	// Create a 2D grid for spatial representation
	width := 25
	height := 13

	// Create empty grid
	swarmGrid := createEmptyGrid(width, height)

	// Place agents on grid based on their synchronization
	centerX := width / 2
	centerY := height / 2

	// Draw center target first
	swarmGrid[centerY][centerX] = '⊕'

	for i, agent := range snapshot.Agents {
		if i >= 20 {
			break
		}

		// Calculate distance from center based on sync
		// Well-synced agents are closer to center
		distance := 5.0
		if snapshot.Coherence > 0.5 {
			distance = 2.0 + 3.0*(1-snapshot.Coherence)
		}

		// Position based on phase and agent index for spreading
		angle := agent.Phase + float64(i)*math.Pi/10
		x := centerX + int(distance*math.Cos(angle))
		y := centerY + int(distance*math.Sin(angle)/2)

		// Ensure within bounds and not overwriting center
		if x >= 0 && x < width && y >= 0 && y < height && (x != centerX || y != centerY) {
			swarmGrid[y][x] = getAgentCharacter(agent.Phase)
		}
	}

	// Render grid - always render all rows
	for y := range height {
		if y < len(swarmGrid) {
			row := swarmGrid[y]
			for _, char := range row {
				renderGridChar(d.swarmVisText, char)
			}
		}
		writeText(d.swarmVisText, "\n")
	}

	// Add legend
	writeText(d.swarmVisText, "\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	writeText(d.swarmVisText, "⊕",
		text.WriteCellOpts(cell.FgColor(cell.ColorCyan)))
	writeText(d.swarmVisText, " Target  ",
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	writeText(d.swarmVisText, "●",
		text.WriteCellOpts(cell.FgColor(cell.ColorGreen)))
	writeText(d.swarmVisText, " Synced\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	writeText(d.swarmVisText, "o",
		text.WriteCellOpts(cell.FgColor(cell.ColorYellow)))
	writeText(d.swarmVisText, " Near    ",
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	writeText(d.swarmVisText, "∙",
		text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
	writeText(d.swarmVisText, " Far",
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
}

// ShowSummary displays the final results
func (d *TerminalDisplay) ShowSummary(stats Statistics) {
	// Clear display
	d.titleText.Reset()
	d.agentsText.Reset()
	d.metricsText.Reset()
	d.costText.Reset()
	d.swarmText.Reset()
	d.swarmVisText.Reset()

	// Show summary in title area
	writeText(d.titleText, "🎉 SIMULATION COMPLETE\n\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorGreen), cell.Bold()))
	writeText(d.titleText, fmt.Sprintf("Total Savings: $%.2f (%.1f%%)\n",
		stats.TotalSavings, stats.SavingsPercent),
		text.WriteCellOpts(cell.FgColor(cell.ColorYellow), cell.Bold()))
	writeText(d.titleText, "Agents successfully learned to batch API calls!",
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

	// Show detailed stats
	writeText(d.metricsText, "📊 FINAL STATISTICS\n\n",
		text.WriteCellOpts(cell.FgColor(cell.ColorCyan), cell.Bold()))
	writeText(d.metricsText, fmt.Sprintf("API Calls (individual): %d\n", stats.TotalAPICalls),
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	writeText(d.metricsText, fmt.Sprintf("API Calls (batched):    %d\n", stats.TotalBatches),
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	writeText(d.metricsText, fmt.Sprintf("Average Batch Size:     %.1f\n", stats.AverageBatchSize),
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	writeText(d.metricsText, fmt.Sprintf("Peak Coherence:         %.1f%%\n", stats.PeakCoherence*100),
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	writeText(d.metricsText, fmt.Sprintf("Time to Converge:       %s\n", formatDuration(stats.TimeToConverge)),
		text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
}

// formatDuration formats a duration for display
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}

// Run runs the terminal display with keyboard handling
func (d *TerminalDisplay) Run(ctx context.Context, controller *KeyboardController) error {
	d.controller = controller

	// Run termdash with keyboard subscriber if controller is provided
	if controller != nil {
		return termdash.Run(ctx, d.terminal, d.container,
			termdash.KeyboardSubscriber(func(k *terminalapi.Keyboard) {
				if d.controller != nil {
					d.controller.ProcessKey(k)
				}
			}),
			termdash.RedrawInterval(100*time.Millisecond),
		)
	}

	// Run without keyboard handling if no controller
	return termdash.Run(ctx, d.terminal, d.container,
		termdash.RedrawInterval(100*time.Millisecond),
	)
}
