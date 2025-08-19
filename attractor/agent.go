package attractor

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

	// Swarm configuration
	swarmSize           int // Total size of swarm (for density calculations)
	assumedMaxNeighbors int // Assumed max neighbors (0 = use swarm size)

	// Context awareness
	context atomic.Value // stores Context

	// Decision-making components (interfaces for future upgrade)
	decider     DecisionMaker
	goalManager GoalManager
	resources   ResourceManager
	strategy    SyncStrategy // Synchronization strategy
}

// AgentOption configures an Agent
type AgentOption func(*Agent)

// NewAgent creates an agent with the provided options.
// If no options are provided, sensible defaults are used.
func NewAgent(id string, opts ...AgentOption) *Agent {
	a := &Agent{
		ID: id,
	}

	// Apply defaults
	defaults := []AgentOption{
		WithDecisionMaker(&SimpleDecisionMaker{}),
		WithGoalManager(&WeightedGoalManager{}),
		WithResourceManager(NewTokenResourceManager(100)),
		WithStrategy(NewPhaseNudgeStrategy(0.3)),
		WithRandomPhase(),
		WithRandomLocalGoal(),
		WithRandomFrequency(),
		WithEnergy(100.0),
		WithInfluence(0.5),
		WithStubbornness(0.2),
	}

	// Apply defaults first
	for _, opt := range defaults {
		opt(a)
	}

	// Apply user options (can override defaults)
	for _, opt := range opts {
		opt(a)
	}

	// Initialize context
	a.context.Store(Context{})

	return a
}

// WithDecisionMaker sets the agent's decision maker
func WithDecisionMaker(dm DecisionMaker) AgentOption {
	return func(a *Agent) {
		a.decider = dm
	}
}

// WithGoalManager sets the agent's goal manager
func WithGoalManager(gm GoalManager) AgentOption {
	return func(a *Agent) {
		a.goalManager = gm
	}
}

// WithResourceManager sets the agent's resource manager
func WithResourceManager(rm ResourceManager) AgentOption {
	return func(a *Agent) {
		a.resources = rm
	}
}

// WithStrategy sets the agent's synchronization strategy
func WithStrategy(s SyncStrategy) AgentOption {
	return func(a *Agent) {
		a.strategy = s
	}
}

// WithPhase sets the agent's initial phase
func WithPhase(phase float64) AgentOption {
	return func(a *Agent) {
		a.phase.Store(phase)
	}
}

// WithRandomPhase sets a random initial phase
func WithRandomPhase() AgentOption {
	return func(a *Agent) {
		a.phase.Store(rand.Float64() * 2 * math.Pi)
	}
}

// WithLocalGoal sets the agent's local goal
func WithLocalGoal(goal float64) AgentOption {
	return func(a *Agent) {
		a.LocalGoal.Store(goal)
	}
}

// WithRandomLocalGoal sets a random local goal
func WithRandomLocalGoal() AgentOption {
	return func(a *Agent) {
		a.LocalGoal.Store(rand.Float64() * 2 * math.Pi)
	}
}

// WithFrequency sets the agent's frequency
func WithFrequency(freq time.Duration) AgentOption {
	return func(a *Agent) {
		a.frequency.Store(freq)
	}
}

// WithRandomFrequency sets a random frequency between 50-150ms
func WithRandomFrequency() AgentOption {
	return func(a *Agent) {
		a.frequency.Store(time.Duration(50+rand.Intn(100)) * time.Millisecond)
	}
}

// WithEnergy sets the agent's initial energy
func WithEnergy(energy float64) AgentOption {
	return func(a *Agent) {
		a.energy.Store(energy)
	}
}

// WithInfluence sets the agent's influence parameter
func WithInfluence(influence float64) AgentOption {
	return func(a *Agent) {
		a.influence.Store(influence)
	}
}

// WithStubbornness sets the agent's stubbornness parameter
func WithStubbornness(stubbornness float64) AgentOption {
	return func(a *Agent) {
		a.stubbornness.Store(stubbornness)
	}
}

// WithSwarmInfo sets the swarm size information for density calculations
func WithSwarmInfo(swarmSize, assumedMaxNeighbors int) AgentOption {
	return func(a *Agent) {
		a.swarmSize = swarmSize
		a.assumedMaxNeighbors = assumedMaxNeighbors
	}
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

	// Calculate density based on actual swarm size or assumed max
	maxNeighbors := a.assumedMaxNeighbors
	if maxNeighbors == 0 && a.swarmSize > 0 {
		// Use actual swarm size minus self
		maxNeighbors = a.swarmSize - 1
	} else if maxNeighbors == 0 {
		// Fallback to old default if not configured
		maxNeighbors = 20
	}

	density := float64(neighborCount) / float64(maxNeighbors)
	stability := 1.0 / (1.0 + phaseVarSum) // Inverse variance

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

	// Current state of the agent
	currentState := State{
		Phase:     a.phase.Load(),
		Frequency: a.frequency.Load(),
		Coherence: ctx.LocalCoherence,
	}

	// Local goal (what the agent individually wants)
	localGoal := State{
		Phase:     a.LocalGoal.Load(),
		Frequency: a.frequency.Load(),
		Coherence: ctx.LocalCoherence,
	}

	// Blend local and global goals based on influence
	influence := a.influence.Load()
	blendedGoal := a.goalManager.Blend(localGoal, globalGoal, influence)

	// Use strategy to generate action from current state toward blended goal
	proposedAction, strategyConfidence := a.strategy.Propose(currentState, blendedGoal, ctx)

	// Generate additional options for decision maker
	// Include maintain option as alternative
	maintainAction := Action{
		Type:    "maintain",
		Value:   0,
		Cost:    0.1,
		Benefit: ctx.Stability,
	}

	options := []Action{proposedAction, maintainAction}

	// Make autonomous decision
	chosen, confidence := a.decider.Decide(currentState, options)
	// Use maximum confidence rather than product to avoid being too conservative
	if strategyConfidence > confidence {
		confidence = strategyConfidence
	}

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

// SetSyncStrategy sets the agent's synchronization strategy.
func (a *Agent) SetSyncStrategy(strategy SyncStrategy) {
	a.strategy = strategy
}

// GetSyncStrategy returns the agent's current sync strategy.
func (a *Agent) GetSyncStrategy() SyncStrategy {
	return a.strategy
}

// ApplyAction executes a chosen action, consuming resources.
// Returns success status and energy consumed.
func (a *Agent) ApplyAction(action Action) (bool, float64) {
	switch action.Type {
	case "adjust_phase", "phase_nudge", "frequency_lock", "energy_save", "pulse":
		// All these involve phase adjustment
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

// SetEnergy sets the agent's energy level (for testing).
func (a *Agent) SetEnergy(energy float64) {
	if rm, ok := a.resources.(*TokenResourceManager); ok {
		// Reset to new energy level
		rm.tokens.Store(energy)
	}
}

// GetInfluence returns the agent's influence parameter.
func (a *Agent) GetInfluence() float64 {
	return a.influence.Load()
}

// SetInfluence sets the agent's influence parameter.
func (a *Agent) SetInfluence(influence float64) {
	a.influence.Store(influence)
}

// GetStubbornness returns the agent's stubbornness parameter.
func (a *Agent) GetStubbornness() float64 {
	return a.stubbornness.Load()
}

// SetStubbornness sets the agent's stubbornness parameter.
func (a *Agent) SetStubbornness(stubbornness float64) {
	a.stubbornness.Store(stubbornness)
}

// Neighbors returns the agent's neighbors map for direct access.
func (a *Agent) Neighbors() *sync.Map {
	return &a.neighbors
}

// DecisionMaker returns the agent's decision maker.
func (a *Agent) DecisionMaker() DecisionMaker {
	return a.decider
}

// SetDecisionMaker sets a custom decision maker for the agent.
func (a *Agent) SetDecisionMaker(dm DecisionMaker) {
	a.decider = dm
}

// StateUpdate carries synchronization information between agents.
type StateUpdate struct {
	FromID    string
	Phase     float64
	Frequency time.Duration
	Energy    float64
}
