package biofield

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"go.uber.org/atomic"
)

// Agent represents an autonomous entity with local goals and decision-making
// capability. Unlike traditional passive agents, these have genuine agency -
// they can refuse adjustments and pursue local objectives while contributing
// to global patterns.
//
// This implements several biological properties:
//   - Local goals (individual phase preference)
//   - Energy constraints (limited resources for changes)
//   - Social influence (neighborhood awareness)
//   - Autonomous decisions (not just following orders)
type Agent struct {
	ID string

	// Local goal - what this agent individually wants to achieve
	// This creates multi-scale goal structure: individual + collective
	LocalGoal atomic.Float64 // Preferred phase

	// Current state
	phase     atomic.Float64  // Current phase [0, 2π]
	frequency atomic.Duration // Current frequency

	// Autonomy parameters
	energy       atomic.Float64 // Available energy for adjustments
	influence    atomic.Float64 // Weight of local vs global goals [0, 1]
	stubbornness atomic.Float64 // Resistance to change [0, 1]

	// Social network
	neighbors sync.Map // map[string]*Agent - local connections

	// Context awareness
	context atomic.Value // stores Context

	// Decision-making components (interfaces for future upgrade)
	decider     DecisionMaker
	goalManager GoalManager
	resources   ResourceManager
}

// NewAgent creates an autonomous agent with random initial state and goals.
// Each agent starts with:
//   - Random phase (creates initial disorder)
//   - Random local goal (creates tension with global goal)
//   - Full energy (depletes through actions)
//   - Moderate influence (balances local/global)
func NewAgent(id string) *Agent {
	a := &Agent{
		ID:          id,
		decider:     &SimpleDecisionMaker{},
		goalManager: &WeightedGoalManager{},
		resources:   NewTokenResourceManager(100),
	}

	// Random initialization for diversity
	a.phase.Store(rand.Float64() * 2 * math.Pi)
	a.LocalGoal.Store(rand.Float64() * 2 * math.Pi)
	a.frequency.Store(time.Duration(50+rand.Intn(100)) * time.Millisecond)
	a.energy.Store(100.0)
	a.influence.Store(0.3 + rand.Float64()*0.4) // 0.3-0.7 range
	a.stubbornness.Store(rand.Float64() * 0.3)  // 0-0.3 range

	// Initialize context
	a.context.Store(Context{})

	return a
}

// UpdateContext recalculates environmental context from local observations.
// This enables context-sensitive behavior without global knowledge.
func (a *Agent) UpdateContext() {
	var (
		neighborCount int
		phaseSum      float64
		phaseVarSum   float64
	)

	// Analyze local neighborhood
	a.neighbors.Range(func(key, value any) bool {
		neighbor := value.(*Agent)
		neighborCount++
		phase := neighbor.phase.Load()
		phaseSum += phase
		phaseVarSum += math.Pow(phase-a.phase.Load(), 2)
		return true
	})

	density := float64(neighborCount) / 20.0 // Assume 20 max neighbors
	stability := 1.0 / (1.0 + phaseVarSum)  // Inverse variance

	// Calculate local coherence (Kuramoto order parameter)
	localCoherence := a.calculateLocalCoherence()

	ctx := Context{
		Density:        math.Min(density, 1.0),
		Stability:      stability,
		Progress:       0.5, // Would track improvement over time
		LocalCoherence: localCoherence,
	}

	a.context.Store(ctx)
}

// calculateLocalCoherence measures synchronization with neighbors.
// Uses Kuramoto order parameter: R = |Σ e^(iφ)| / N
func (a *Agent) calculateLocalCoherence() float64 {
	var sumCos, sumSin float64
	var count int

	a.neighbors.Range(func(key, value any) bool {
		neighbor := value.(*Agent)
		phase := neighbor.phase.Load()
		sumCos += math.Cos(phase)
		sumSin += math.Sin(phase)
		count++
		return true
	})

	if count == 0 {
		return 0
	}

	return math.Sqrt(sumCos*sumCos+sumSin*sumSin) / float64(count)
}

// ProposeAdjustment suggests a phase change based on local and global goals.
// This implements hierarchical goal blending - agents pursue weighted
// combination of individual preferences and collective objectives.
//
// The agent won't blindly follow external commands but negotiates
// based on energy, stubbornness, and goal alignment.
func (a *Agent) ProposeAdjustment(globalGoal State) (Action, bool) {
	// Update environmental awareness
	a.UpdateContext()
	ctx := a.context.Load().(Context)

	// Blend local and global goals based on influence
	localState := State{
		Phase:     a.LocalGoal.Load(),
		Frequency: a.frequency.Load(),
		Coherence: ctx.LocalCoherence,
	}

	influence := a.influence.Load()
	blendedGoal := a.goalManager.Blend(localState, globalGoal, influence)

	// Generate possible actions
	options := a.generateActions(blendedGoal)

	// Make autonomous decision
	chosen, confidence := a.decider.Decide(localState, options)

	// Check energy availability
	available := a.resources.Request(chosen.Cost)
	if available < chosen.Cost*0.8 { // Need at least 80% of required energy
		a.resources.Release(available) // Return unused energy
		return Action{Type: "maintain"}, false
	}

	// Consider stubbornness (resistance to change)
	if rand.Float64() < a.stubbornness.Load() {
		a.resources.Release(available)
		return Action{Type: "maintain"}, false
	}

	// Accept adjustment with probability based on confidence
	if rand.Float64() < confidence {
		return chosen, true
	}

	a.resources.Release(available)
	return Action{Type: "maintain"}, false
}

// generateActions creates possible actions for current situation.
// Actions have costs proportional to change magnitude.
func (a *Agent) generateActions(goal State) []Action {
	current := a.phase.Load()
	target := goal.Phase

	// Calculate phase difference
	diff := target - current
	for diff > math.Pi {
		diff -= 2 * math.Pi
	}
	for diff < -math.Pi {
		diff += 2 * math.Pi
	}

	return []Action{
		{
			Type:    "adjust_phase",
			Value:   diff * 0.1, // Small adjustment
			Cost:    math.Abs(diff) * 2.0,
			Benefit: 1.0 - math.Abs(diff)/math.Pi,
		},
		{
			Type:    "adjust_phase",
			Value:   diff * 0.3, // Medium adjustment
			Cost:    math.Abs(diff) * 5.0,
			Benefit: 1.5 - math.Abs(diff)/math.Pi,
		},
		{
			Type:    "maintain",
			Value:   0,
			Cost:    0.1, // Small cost to maintain
			Benefit: a.context.Load().(Context).Stability,
		},
	}
}

// ApplyAction executes a chosen action, consuming resources.
// Returns success status and energy consumed.
func (a *Agent) ApplyAction(action Action) (bool, float64) {
	switch action.Type {
	case "adjust_phase":
		newPhase := math.Mod(a.phase.Load()+action.Value, 2*math.Pi)
		if newPhase < 0 {
			newPhase += 2 * math.Pi
		}
		a.phase.Store(newPhase)
		return true, action.Cost

	case "maintain":
		// Do nothing but consume small energy
		return true, action.Cost

	default:
		return false, 0
	}
}

// GetPhase returns the current phase of the agent.
func (a *Agent) GetPhase() float64 {
	return a.phase.Load()
}

// SetPhase sets the agent's phase directly (for testing).
func (a *Agent) SetPhase(phase float64) {
	a.phase.Store(phase)
}

// GetEnergy returns the current energy level.
func (a *Agent) GetEnergy() float64 {
	return a.resources.Available()
}

// StateUpdate carries synchronization information between agents.
type StateUpdate struct {
	FromID    string
	Phase     float64
	Frequency time.Duration
	Energy    float64
}