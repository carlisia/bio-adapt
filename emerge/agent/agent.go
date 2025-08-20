package agent

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/decision"
	"github.com/carlisia/bio-adapt/emerge/goal"
	"github.com/carlisia/bio-adapt/emerge/strategy"
	"github.com/carlisia/bio-adapt/internal/resource"
	"go.uber.org/atomic"
)

// Agent represents an autonomous entity with local goals and decision-making
// capability. Agents make their own decisions about synchronization while
// contributing to emergent global patterns.
type Agent struct {
	// ID is the unique identifier for this agent
	ID string

	// Private fields - accessed through methods
	localGoal    atomic.Float64  // Preferred phase
	phase        atomic.Float64  // Current phase [0, 2π]
	frequency    atomic.Duration // Current frequency
	energy       atomic.Float64  // Available energy
	influence    atomic.Float64  // Local vs global weight [0, 1]
	stubbornness atomic.Float64  // Resistance to change [0, 1]

	neighbors sync.Map     // map[string]*Agent
	context   atomic.Value // stores Context

	// Configuration
	swarmSize           int
	assumedMaxNeighbors int

	// Components (interfaces for extensibility)
	decider     core.DecisionMaker
	goalManager goal.Manager
	resources   core.ResourceManager
	strategy    core.SyncStrategy
}

// Option configures an Agent
type Option func(*Agent)

// New creates an agent with the provided options.
func New(id string, opts ...Option) *Agent {
	a := &Agent{
		ID: id,
	}

	// Apply defaults
	defaults := []Option{
		WithDecisionMaker(&decision.SimpleDecisionMaker{}),
		WithGoalManager(&goal.WeightedManager{}),
		WithResourceManager(resource.NewTokenManager(100)),
		WithStrategy(&strategy.PhaseNudge{Rate: 0.3}),
		WithRandomPhase(),
		WithRandomLocalGoal(),
		WithRandomFrequency(),
		WithEnergy(100.0),
		WithInfluence(0.5),
		WithStubbornness(0.2),
	}

	for _, opt := range defaults {
		opt(a)
	}

	// Apply user options (can override defaults)
	for _, opt := range opts {
		opt(a)
	}

	// Initialize context
	a.context.Store(core.Context{})

	return a
}

// ============= Public Getters (no "Get" prefix) =============

// Phase returns the agent's current phase
func (a *Agent) Phase() float64 {
	return a.phase.Load()
}

// Frequency returns the agent's oscillation frequency
func (a *Agent) Frequency() time.Duration {
	return a.frequency.Load()
}

// Energy returns the agent's available energy
func (a *Agent) Energy() float64 {
	return a.energy.Load()
}

// Influence returns the agent's influence weight
func (a *Agent) Influence() float64 {
	return a.influence.Load()
}

// Stubbornness returns the agent's resistance to change
func (a *Agent) Stubbornness() float64 {
	return a.stubbornness.Load()
}

// LocalGoal returns the agent's individual target phase
func (a *Agent) LocalGoal() float64 {
	return a.localGoal.Load()
}

// NeighborCount returns the number of connected neighbors
func (a *Agent) NeighborCount() int {
	count := 0
	a.neighbors.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

// ============= Public Setters =============

// SetPhase updates the agent's phase
func (a *Agent) SetPhase(phase float64) {
	a.phase.Store(core.WrapPhase(phase))
}

// SetFrequency updates the agent's frequency
func (a *Agent) SetFrequency(freq time.Duration) {
	a.frequency.Store(freq)
}

// SetEnergy updates the agent's energy
func (a *Agent) SetEnergy(energy float64) {
	a.energy.Store(math.Max(0, energy))
}

// SetInfluence updates the agent's influence weight
func (a *Agent) SetInfluence(influence float64) {
	// Clamp to [0, 1]
	if influence < 0 {
		influence = 0
	} else if influence > 1 {
		influence = 1
	}
	a.influence.Store(influence)
}

// SetStubbornness updates the agent's resistance
func (a *Agent) SetStubbornness(stubbornness float64) {
	// Clamp to [0, 1]
	if stubbornness < 0 {
		stubbornness = 0
	} else if stubbornness > 1 {
		stubbornness = 1
	}
	a.stubbornness.Store(stubbornness)
}

// SetLocalGoal updates the agent's individual target
func (a *Agent) SetLocalGoal(goal float64) {
	a.localGoal.Store(core.WrapPhase(goal))
}

// Neighbors returns the agent's neighbors map for iteration
func (a *Agent) Neighbors() *sync.Map {
	return &a.neighbors
}

// SetDecisionMaker replaces the decision-making component
func (a *Agent) SetDecisionMaker(dm core.DecisionMaker) {
	a.decider = dm
}

// ============= Public Methods =============

// ConnectTo establishes a bidirectional connection with another agent
func (a *Agent) ConnectTo(other *Agent) {
	if other != nil && other.ID != a.ID {
		a.neighbors.Store(other.ID, other)
		other.neighbors.Store(a.ID, a)
	}
}

// DisconnectFrom removes a connection
func (a *Agent) DisconnectFrom(other *Agent) {
	if other != nil {
		a.neighbors.Delete(other.ID)
		other.neighbors.Delete(a.ID)
	}
}

// IsConnectedTo checks if connected to another agent
func (a *Agent) IsConnectedTo(other *Agent) bool {
	if other == nil {
		return false
	}
	_, exists := a.neighbors.Load(other.ID)
	return exists
}

// ProposeAdjustment evaluates and potentially accepts an adjustment
func (a *Agent) ProposeAdjustment(globalGoal core.State) (core.Action, bool) {
	// Stubborn agents resist change
	if a.rejectsDueToStubbornness() {
		return core.Action{Type: "maintain"}, false
	}

	// Get blended goal
	localState := core.State{
		Phase:     a.localGoal.Load(),
		Frequency: a.frequency.Load(),
		Coherence: 0, // Individual agent has no coherence
	}
	blendedGoal := a.goalManager.Blend(localState, globalGoal, a.influence.Load())

	// Generate proposal using context
	currentState := core.State{
		Phase:     a.phase.Load(),
		Frequency: a.frequency.Load(),
		Coherence: a.calculateLocalCoherence(),
	}

	// Create a simple context for the strategy
	ctx := core.Context{
		LocalCoherence: a.calculateLocalCoherence(),
		Stability:      a.calculateStability(),
		Density:        float64(a.NeighborCount()) / float64(a.swarmSize),
		Neighbors:      a.NeighborCount(),
	}

	proposal, confidence := a.strategy.Propose(currentState, blendedGoal, ctx)

	// Make decision
	options := []core.Action{
		proposal,
		{Type: "maintain", Cost: 0.1, Benefit: ctx.Stability},
	}

	chosen, acceptance := a.decider.Decide(currentState, options)

	// Check energy
	if chosen.Cost > a.energy.Load() {
		return core.Action{Type: "maintain"}, false
	}

	// Accept based on confidence
	if rand.Float64() < confidence*acceptance {
		return chosen, true
	}

	return core.Action{Type: "maintain"}, false
}

// ApplyAction executes an action and returns success status, energy cost, and any error.
// The error provides detailed context about why the action failed.
func (a *Agent) ApplyAction(action core.Action) (bool, float64, error) {
	energyCost := action.Cost
	available := a.energy.Load()

	// Check energy
	if energyCost > available {
		return false, 0, fmt.Errorf("%w: required %.2f, available %.2f",
			core.ErrInsufficientEnergy, energyCost, available)
	}

	// Apply based on action type
	switch action.Type {
	case "adjust_phase":
		a.adjustPhase(action.Value)
	case "maintain":
		// No change needed
	case "phase_nudge", "frequency_lock", "energy_save", "pulse":
		// Strategy-specific actions
		a.adjustPhase(action.Value)
	default:
		return false, 0, fmt.Errorf("%w: %s", core.ErrUnknownActionType, action.Type)
	}

	// Consume energy
	a.consumeEnergy(energyCost)

	return true, energyCost, nil
}

// UpdateContext updates the agent's perception of its environment
func (a *Agent) UpdateContext() {
	neighbors := 0
	sumCos := 0.0
	sumSin := 0.0
	myPhase := a.phase.Load()

	a.neighbors.Range(func(key, value any) bool {
		neighbor := value.(*Agent)
		neighbors++

		diff := neighbor.Phase() - myPhase
		sumCos += math.Cos(diff)
		sumSin += math.Sin(diff)

		return true
	})

	// Calculate local coherence
	localCoherence := 0.0
	if neighbors > 0 {
		localCoherence = math.Sqrt(sumCos*sumCos+sumSin*sumSin) / float64(neighbors)
	}

	// Calculate density
	maxNeighbors := a.assumedMaxNeighbors
	if maxNeighbors == 0 {
		maxNeighbors = a.swarmSize - 1
	}
	density := float64(neighbors) / float64(maxNeighbors)

	// Store context
	a.context.Store(core.Context{
		Neighbors:      neighbors,
		Density:        density,
		LocalCoherence: localCoherence,
		Stability:      a.calculateStability(),
	})
}

// ============= Option Functions =============

// WithPhase sets initial phase
func WithPhase(phase float64) Option {
	return func(a *Agent) {
		a.phase.Store(core.WrapPhase(phase))
	}
}

// WithRandomPhase sets random initial phase
func WithRandomPhase() Option {
	return func(a *Agent) {
		a.phase.Store(rand.Float64() * 2 * math.Pi)
	}
}

// WithFrequency sets oscillation frequency
func WithFrequency(freq time.Duration) Option {
	return func(a *Agent) {
		a.frequency.Store(freq)
	}
}

// WithRandomFrequency sets random frequency
func WithRandomFrequency() Option {
	return func(a *Agent) {
		baseFreq := 100 * time.Millisecond
		variation := time.Duration(rand.Float64()*50) * time.Millisecond
		a.frequency.Store(baseFreq + variation)
	}
}

// WithLocalGoal sets the agent's individual target
func WithLocalGoal(goal float64) Option {
	return func(a *Agent) {
		a.localGoal.Store(core.WrapPhase(goal))
	}
}

// WithRandomLocalGoal sets random local goal
func WithRandomLocalGoal() Option {
	return func(a *Agent) {
		a.localGoal.Store(rand.Float64() * 2 * math.Pi)
	}
}

// WithEnergy sets initial energy
func WithEnergy(energy float64) Option {
	return func(a *Agent) {
		a.energy.Store(math.Max(0, energy))
	}
}

// WithInfluence sets influence weight
func WithInfluence(influence float64) Option {
	return func(a *Agent) {
		a.SetInfluence(influence)
	}
}

// WithStubbornness sets resistance to change
func WithStubbornness(stubbornness float64) Option {
	return func(a *Agent) {
		a.SetStubbornness(stubbornness)
	}
}

// WithDecisionMaker sets decision-making component
func WithDecisionMaker(dm core.DecisionMaker) Option {
	return func(a *Agent) {
		a.decider = dm
	}
}

// WithGoalManager sets goal blending component
func WithGoalManager(gm goal.Manager) Option {
	return func(a *Agent) {
		a.goalManager = gm
	}
}

// WithResourceManager sets resource management component
func WithResourceManager(rm core.ResourceManager) Option {
	return func(a *Agent) {
		a.resources = rm
		a.energy.Store(rm.Available())
	}
}

// WithStrategy sets synchronization strategy
func WithStrategy(s core.SyncStrategy) Option {
	return func(a *Agent) {
		a.strategy = s
	}
}

// WithSwarmInfo sets swarm configuration
func WithSwarmInfo(swarmSize, assumedMaxNeighbors int) Option {
	return func(a *Agent) {
		a.swarmSize = swarmSize
		a.assumedMaxNeighbors = assumedMaxNeighbors
	}
}

// ============= Private Methods =============

func (a *Agent) rejectsDueToStubbornness() bool {
	return rand.Float64() < a.stubbornness.Load()
}

func (a *Agent) adjustPhase(delta float64) {
	newPhase := a.phase.Load() + delta
	a.phase.Store(core.WrapPhase(newPhase))
}

func (a *Agent) consumeEnergy(amount float64) {
	current := a.energy.Load()
	a.energy.Store(math.Max(0, current-amount))

	if a.resources != nil {
		a.resources.Request(amount)
	}
}

func (a *Agent) calculateStability() float64 {
	// Simple stability metric based on recent phase changes
	// In a full implementation, this would track history
	return 0.5 // Placeholder
}

func (a *Agent) calculateLocalCoherence() float64 {
	neighbors := 0
	sumCos := 0.0
	sumSin := 0.0
	myPhase := a.phase.Load()

	a.neighbors.Range(func(key, value any) bool {
		neighbor := value.(*Agent)
		neighbors++

		diff := neighbor.Phase() - myPhase
		sumCos += math.Cos(diff)
		sumSin += math.Sin(diff)

		return true
	})

	if neighbors == 0 {
		return 0
	}

	return math.Sqrt(sumCos*sumCos+sumSin*sumSin) / float64(neighbors)
}
