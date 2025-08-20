package swarm

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/carlisia/bio-adapt/emerge/agent"
	"github.com/carlisia/bio-adapt/emerge/completion"
	"github.com/carlisia/bio-adapt/emerge/convergence"
	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/strategy"
)

// GoalDirectedSync achieves synchronization through adaptive strategies
type GoalDirectedSync struct {
	swarm              *Swarm
	targetPattern      *core.RhythmicPattern
	completionEngine   *completion.Engine
	convergenceMonitor *convergence.Monitor
	strategies         []core.SyncStrategy
	currentStrategy    core.SyncStrategy
	strategyPerf       map[string]*StrategyPerformance
	mu                 sync.RWMutex
}

// StrategyPerformance tracks how well a strategy works
type StrategyPerformance struct {
	Name           string
	Attempts       int
	Successes      int
	TotalReward    float64
	LastUsed       time.Time
	AvgConvergence float64
}

// NewGoalDirectedSync creates a goal-directed synchronization system
func NewGoalDirectedSync(s *Swarm) *GoalDirectedSync {
	gds := &GoalDirectedSync{
		swarm:              s,
		completionEngine:   completion.NewEngine(),
		convergenceMonitor: convergence.NewMonitor(10),
		strategies: []core.SyncStrategy{
			&strategy.PhaseNudge{Rate: 0.3},
			&strategy.PhaseNudge{Rate: 0.7}, // Aggressive version
			&strategy.FrequencyLock{SyncRate: 0.5},
			&strategy.EnergyAware{Threshold: 20},
			strategy.NewPulse(100*time.Millisecond, 0.8),
		},
		strategyPerf: make(map[string]*StrategyPerformance),
	}

	// Initialize performance tracking
	for _, strat := range gds.strategies {
		gds.strategyPerf[strat.Name()] = &StrategyPerformance{
			Name: strat.Name(),
		}
	}

	// Load pattern templates
	gds.completionEngine.LoadDefaultTemplates()

	// Set initial strategy
	gds.currentStrategy = gds.strategies[0]

	return gds
}

// AchieveSynchronization runs goal-directed synchronization loop
func (gds *GoalDirectedSync) AchieveSynchronization(ctx context.Context, target *core.RhythmicPattern) error {
	gds.targetPattern = target
	gds.convergenceMonitor.SetTarget(target)

	// Goal-directed loop
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	iterationCount := 0
	maxIterations := 200 // Safety limit

	for iterationCount < maxIterations {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			iterationCount++

			// Step 1: Measure current pattern
			currentPattern := gds.measureSystemPattern()
			coherence := gds.swarm.MeasureCoherence()

			// Step 2: Record convergence
			gds.convergenceMonitor.RecordSample(currentPattern, coherence)

			// Step 3: Check if we've achieved the goal
			if gds.isPatternAchieved(currentPattern) {
				return nil // Success!
			}

			// Step 4: Check if we should switch strategy
			if gds.convergenceMonitor.ShouldSwitchStrategy() {
				gds.switchStrategy()
			}

			// Step 5: Identify pattern gaps
			gaps := core.IdentifyGaps(currentPattern, target)

			// Step 6: Complete pattern using bioelectric memory
			completion := gds.completionEngine.CompletePattern(currentPattern, gaps)

			// Step 7: Apply adjustments through current strategy
			gds.applyPatternCompletion(completion)

			// Step 8: Add noise if stuck to escape local minima
			if gds.convergenceMonitor.IsStuck() {
				gds.addStochasticResonance()
			}
		}
	}

	return fmt.Errorf("failed to achieve synchronization after %d iterations", maxIterations)
}

// measureSystemPattern calculates the current system-wide pattern
func (gds *GoalDirectedSync) measureSystemPattern() *core.RhythmicPattern {
	agents := gds.swarm.Agents()
	if len(agents) == 0 {
		return &core.RhythmicPattern{}
	}

	// Calculate average phase using circular mean
	sumSin := 0.0
	sumCos := 0.0
	totalFreq := time.Duration(0)
	count := 0

	for _, agent := range agents {
		phase := agent.Phase()
		sumSin += math.Sin(phase)
		sumCos += math.Cos(phase)
		totalFreq += agent.Frequency()
		count++
	}

	avgPhase := math.Atan2(sumSin/float64(count), sumCos/float64(count))
	avgFreq := totalFreq / time.Duration(count)
	coherence := gds.swarm.MeasureCoherence()

	return &core.RhythmicPattern{
		Phase:     avgPhase,
		Frequency: avgFreq,
		Amplitude: 1.0,
		Coherence: coherence,
		Stability: 0.5,
	}
}

// isPatternAchieved checks if we've reached the target
func (gds *GoalDirectedSync) isPatternAchieved(current *core.RhythmicPattern) bool {
	distance := core.PatternDistance(current, gds.targetPattern)

	// Also check coherence threshold
	coherenceAchieved := current.Coherence >= gds.targetPattern.Coherence-0.05
	distanceAchieved := distance < 0.1

	return coherenceAchieved && distanceAchieved
}

// applyPatternCompletion applies the completed pattern to agents
func (gds *GoalDirectedSync) applyPatternCompletion(completion *completion.CompletedPattern) {
	agents := gds.swarm.Agents()

	// Get adjustments from completion
	phaseAdjustment := completion.GetPhaseAdjustment()
	freqAdjustment := completion.GetFrequencyAdjustment()

	// Apply more aggressive adjustments for stronger convergence
	adjustmentScale := 0.5 // Increase from subtle adjustments

	// Apply to all agents with some variation
	for _, agent := range agents {
		// Add small random variation to avoid perfect sync
		variation := (rand.Float64() - 0.5) * 0.1

		// Always adjust phase toward target (removed threshold check)
		currentPhase := agent.Phase()
		// Move toward target phase more strongly
		targetPhase := gds.targetPattern.Phase
		phaseDiff := core.WrapPhase(targetPhase - currentPhase)
		// Use combination of completion adjustment and direct pull to target
		effectiveAdjustment := phaseAdjustment*0.3 + phaseDiff*adjustmentScale
		newPhase := core.WrapPhase(currentPhase + effectiveAdjustment*(1+variation))
		agent.SetPhase(newPhase)

		// Adjust frequency if needed
		if freqAdjustment != 0 {
			currentFreq := agent.Frequency()
			newFreq := currentFreq + freqAdjustment
			if newFreq > 0 {
				agent.SetFrequency(newFreq)
			}
		}
	}
}

// switchStrategy selects a new strategy based on performance
func (gds *GoalDirectedSync) switchStrategy() {
	gds.mu.Lock()
	defer gds.mu.Unlock()

	// Record failure for current strategy
	if gds.currentStrategy != nil {
		perf := gds.strategyPerf[gds.currentStrategy.Name()]
		perf.Attempts++
		perf.LastUsed = time.Now()
	}

	// Select strategy with best success rate
	var bestStrategy core.SyncStrategy
	bestScore := -1.0

	for _, strat := range gds.strategies {
		perf := gds.strategyPerf[strat.Name()]

		// Calculate score with exploration bonus
		score := 0.0
		if perf.Attempts > 0 {
			score = float64(perf.Successes) / float64(perf.Attempts)
		}

		// Add exploration bonus for less-used strategies
		timeSinceUsed := time.Since(perf.LastUsed).Seconds()
		explorationBonus := math.Min(timeSinceUsed/60.0, 0.3) // Up to 30% bonus
		score += explorationBonus

		// Add randomness for exploration
		score += rand.Float64() * 0.1

		if score > bestScore {
			bestScore = score
			bestStrategy = strat
		}
	}

	gds.currentStrategy = bestStrategy
}

// addStochasticResonance adds noise to help escape local minima
func (gds *GoalDirectedSync) addStochasticResonance() {
	agents := gds.swarm.Agents()

	// Add small random perturbations to 10% of agents
	perturbCount := len(agents) / 10
	if perturbCount < 1 {
		perturbCount = 1
	}

	// Randomly select agents to perturb
	for i := 0; i < perturbCount; i++ {
		// Pick random agent
		var targetAgent *agent.Agent
		idx := rand.Intn(len(agents))
		count := 0
		for _, a := range agents {
			if count == idx {
				targetAgent = a
				break
			}
			count++
		}

		if targetAgent != nil {
			// Add phase noise
			noise := (rand.Float64() - 0.5) * 0.5 // ±0.25 radians
			currentPhase := targetAgent.Phase()
			targetAgent.SetPhase(core.WrapPhase(currentPhase + noise))
		}
	}
}

// RecordSuccess records that current strategy succeeded
func (gds *GoalDirectedSync) RecordSuccess(reward float64) {
	gds.mu.Lock()
	defer gds.mu.Unlock()

	if gds.currentStrategy != nil {
		perf := gds.strategyPerf[gds.currentStrategy.Name()]
		perf.Successes++
		perf.TotalReward += reward
		perf.AvgConvergence = gds.convergenceMonitor.GetConvergenceRate()
	}
}
