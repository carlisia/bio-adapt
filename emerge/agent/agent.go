package agent

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/decision"
	"github.com/carlisia/bio-adapt/emerge/goal"
	"github.com/carlisia/bio-adapt/emerge/strategy"
	"github.com/carlisia/bio-adapt/internal/config"
	"github.com/carlisia/bio-adapt/internal/random"
	"github.com/carlisia/bio-adapt/internal/resource"
)

// Option configures an Agent.
type Option func(*Agent)

// Agent represents an autonomous entity with local goals and decision-making
// capability. It uses grouped atomic fields for efficient concurrent access
// and optimized neighbor storage for better cache locality.
type Agent struct {
	// ID is the unique identifier for this agent
	ID string

	// Optimization 1: Grouped atomic fields to reduce cache line bouncing
	state    *AtomicState    // Phase, Energy, LocalGoal, Frequency
	behavior *AtomicBehavior // Influence, Stubbornness
	
	// Context is updated less frequently
	context atomic.Value // stores Context

	// Optimization 2: Fixed-size neighbor storage for better cache locality
	optimizedNeighbors *NeighborStorage
	neighbors          sync.Map // Fallback for compatibility
	useOptimizedNeighbors bool

	// Configuration (read-only after creation)
	swarmSize           int
	assumedMaxNeighbors int

	// Components (interfaces for extensibility)
	decider     core.DecisionMaker
	goalManager goal.Manager
	resources   core.ResourceManager
	strategy    core.SyncStrategy
}

// New creates a new agent with the given ID.
func New(id string, opts ...Option) *Agent {
	a := &Agent{
		ID:                    id,
		state:                 NewAtomicState(),
		behavior:              NewAtomicBehavior(),
		optimizedNeighbors:    NewNeighborStorage(20), // Default for small-world networks
		useOptimizedNeighbors: true,
	}

	// Initialize with defaults
	a.state.Store(StateData{
		Phase:     random.Phase(),
		Energy:    100.0,
		LocalGoal: random.Phase(),
		Frequency: 100*time.Millisecond + time.Duration(random.Float64()*50)*time.Millisecond,
	})

	a.behavior.Store(BehaviorData{
		Influence:    0.1,
		Stubbornness: 0.2,
	})

	// Default components
	a.decider = &decision.SimpleDecisionMaker{}
	a.goalManager = &goal.WeightedManager{}
	a.resources = resource.NewTokenManager(100)
	a.strategy = &strategy.PhaseNudge{Rate: 0.3}

	// Apply options
	for _, opt := range opts {
		opt(a)
	}

	// Initialize context
	a.context.Store(core.Context{})

	return a
}

// NewOptimizedFromConfig creates an optimized agent from configuration.
func NewOptimizedFromConfig(id string, cfg config.Agent) (*Agent, error) {
	// Determine max neighbors based on config
	maxNeighbors := cfg.AssumedMaxNeighbors
	if maxNeighbors == 0 {
		// Default for small-world networks
		if cfg.SwarmSize > 100 {
			maxNeighbors = 20 // Small-world topology
		} else {
			maxNeighbors = cfg.SwarmSize - 1 // Fully connected for small swarms
		}
	}

	// Create options from config
	opts := []Option{
		WithPhase(cfg.Phase),
		WithFrequency(cfg.Frequency),
		WithEnergy(cfg.InitialEnergy),
		WithInfluence(cfg.Influence),
		WithStubbornness(cfg.Stubbornness),
		WithSwarmInfo(cfg.SwarmSize, maxNeighbors),
		WithLocalGoal(cfg.LocalGoal),
	}

	// Handle randomization
	if cfg.RandomizePhase {
		opts = append(opts, WithRandomPhase())
	}
	if cfg.RandomizeLocalGoal {
		opts = append(opts, WithRandomLocalGoal())
	}
	if cfg.RandomizeFrequency {
		opts = append(opts, WithRandomFrequency())
	}

	// Add decision maker based on type
	switch cfg.DecisionMakerType {
	case "simple":
		opts = append(opts, WithDecisionMaker(&decision.SimpleDecisionMaker{}))
	}

	// Add goal manager based on type
	switch cfg.GoalManagerType {
	case "weighted":
		opts = append(opts, WithGoalManager(&goal.WeightedManager{}))
	}

	// Add resource manager based on type
	if cfg.ResourceManagerType == "token" && cfg.MaxTokens > 0 {
		opts = append(opts, WithResourceManager(resource.NewTokenManager(cfg.MaxTokens)))
	}

	// Add strategy based on type
	switch cfg.StrategyType {
	case "phase_nudge":
		opts = append(opts, WithStrategy(&strategy.PhaseNudge{Rate: cfg.StrategyRate}))
	case "frequency_lock":
		opts = append(opts, WithStrategy(&strategy.FrequencyLock{SyncRate: 0.5}))
	case "energy_aware":
		opts = append(opts, WithStrategy(&strategy.EnergyAware{Threshold: 10}))
	default:
		opts = append(opts, WithStrategy(&strategy.PhaseNudge{Rate: cfg.StrategyRate}))
	}

	return New(id, opts...), nil
}


// Phase returns the agent's current phase.
func (a *Agent) Phase() float64 {
	return a.state.Load().Phase
}

// SetPhase sets the agent's phase.
func (a *Agent) SetPhase(phase float64) {
	a.state.Update(func(s *StateData) {
		s.Phase = core.WrapPhase(phase)
	})
}

// Frequency returns the agent's oscillation frequency.
func (a *Agent) Frequency() time.Duration {
	return a.state.Load().Frequency
}

// SetFrequency updates the agent's frequency.
func (a *Agent) SetFrequency(freq time.Duration) {
	a.state.Update(func(s *StateData) {
		s.Frequency = freq
	})
}

// Energy returns the agent's available energy.
func (a *Agent) Energy() float64 {
	return a.state.Load().Energy
}

// SetEnergy updates the agent's energy.
func (a *Agent) SetEnergy(energy float64) {
	a.state.Update(func(s *StateData) {
		s.Energy = math.Max(0, energy)
	})
}

// LocalGoal returns the agent's individual target phase.
func (a *Agent) LocalGoal() float64 {
	return a.state.Load().LocalGoal
}

// SetLocalGoal updates the agent's individual target.
func (a *Agent) SetLocalGoal(g float64) {
	a.state.Update(func(s *StateData) {
		s.LocalGoal = core.WrapPhase(g)
	})
}

// Influence returns the agent's influence weight.
func (a *Agent) Influence() float64 {
	return a.behavior.Load().Influence
}

// SetInfluence updates the agent's influence weight.
func (a *Agent) SetInfluence(influence float64) {
	if influence < 0 {
		influence = 0
	} else if influence > 1 {
		influence = 1
	}
	a.behavior.Update(func(b *BehaviorData) {
		b.Influence = influence
	})
}

// Stubbornness returns the agent's resistance to change.
func (a *Agent) Stubbornness() float64 {
	return a.behavior.Load().Stubbornness
}

// SetStubbornness updates the agent's resistance.
func (a *Agent) SetStubbornness(stubbornness float64) {
	if stubbornness < 0 {
		stubbornness = 0
	} else if stubbornness > 1 {
		stubbornness = 1
	}
	a.behavior.Update(func(b *BehaviorData) {
		b.Stubbornness = stubbornness
	})
}

// SetDecisionMaker replaces the decision-making component.
func (a *Agent) SetDecisionMaker(dm core.DecisionMaker) {
	a.decider = dm
}

// NeighborCount returns the number of connected neighbors.
func (a *Agent) NeighborCount() int {
	// Check both storage methods for compatibility
	optimizedCount := 0
	if a.useOptimizedNeighbors && a.optimizedNeighbors != nil {
		optimizedCount = a.optimizedNeighbors.Count()
	}
	
	mapCount := 0
	a.neighbors.Range(func(_, _ any) bool {
		mapCount++
		return true
	})
	
	// Return the max to handle both cases
	if optimizedCount > mapCount {
		return optimizedCount
	}
	return mapCount
}

// Neighbors returns the agent's neighbors map for compatibility.
func (a *Agent) Neighbors() *sync.Map {
	// Always return the standard neighbors map
	// This allows tests to directly manipulate neighbors
	return &a.neighbors
}

// ConnectTo establishes a connection to another optimized agent.
func (a *Agent) ConnectTo(otherID string, other *Agent) {
	if other == nil || otherID == a.ID {
		return
	}

	if a.useOptimizedNeighbors {
		a.optimizedNeighbors.Store(otherID, other)
	} else {
		a.neighbors.Store(otherID, other)
	}
}

// DisconnectFrom removes a connection.
func (a *Agent) DisconnectFrom(otherID string) {
	if a.useOptimizedNeighbors {
		a.optimizedNeighbors.Delete(otherID)
	} else {
		a.neighbors.Delete(otherID)
	}
}

// IsConnectedTo checks if connected to another agent.
func (a *Agent) IsConnectedTo(otherID string) bool {
	if a.useOptimizedNeighbors {
		_, exists := a.optimizedNeighbors.Load(otherID)
		return exists
	}
	_, exists := a.neighbors.Load(otherID)
	return exists
}

// UpdateContext updates the agent's perception efficiently with both optimizations.
func (a *Agent) UpdateContext() {
	// Get neighbors efficiently
	var neighborList []*Agent
	if a.useOptimizedNeighbors {
		neighborList = a.optimizedNeighbors.All()
	} else {
		a.neighbors.Range(func(_, value any) bool {
			if neighbor, ok := value.(*Agent); ok {
				neighborList = append(neighborList, neighbor)
			}
			return true
		})
	}

	if len(neighborList) == 0 {
		a.context.Store(core.Context{
			Neighbors:      0,
			Density:        0,
			LocalCoherence: 0,
			Stability:      0.5,
		})
		return
	}

	// Load state once (single atomic operation)
	state := a.state.Load()
	myPhase := state.Phase

	sumCos := 0.0
	sumSin := 0.0

	// Process neighbors
	for _, neighbor := range neighborList {
		diff := neighbor.Phase() - myPhase
		sumCos += math.Cos(diff)
		sumSin += math.Sin(diff)
	}

	neighborCount := len(neighborList)
	localCoherence := math.Sqrt(sumCos*sumCos+sumSin*sumSin) / float64(neighborCount)

	// Kuramoto coupling - update local goal
	if neighborCount > 0 {
		avgSin := sumSin / float64(neighborCount)
		avgCos := sumCos / float64(neighborCount)
		targetPhaseShift := math.Atan2(avgSin, avgCos)

		couplingStrength := 0.5 + 0.5*localCoherence
		effectiveShift := targetPhaseShift * couplingStrength

		// Update local goal (single atomic operation for state update)
		a.state.Update(func(s *StateData) {
			s.LocalGoal = core.WrapPhase(myPhase + effectiveShift)
		})
	}

	// Calculate density
	maxNeighbors := a.assumedMaxNeighbors
	if maxNeighbors == 0 {
		maxNeighbors = a.swarmSize - 1
	}
	density := float64(neighborCount) / float64(maxNeighbors)

	// Store context
	a.context.Store(core.Context{
		Neighbors:      neighborCount,
		Density:        density,
		LocalCoherence: localCoherence,
		Stability:      0.5, // Placeholder
	})
}

// ProposeAdjustment evaluates and potentially accepts an adjustment.
func (a *Agent) ProposeAdjustment(globalGoal core.State) (core.Action, bool) {
	behavior := a.behavior.Load()
	
	// Stubborn agents resist change
	if random.Float64() < behavior.Stubbornness {
		return core.Action{Type: "maintain"}, false
	}

	state := a.state.Load()
	
	// Use pure local goal for Kuramoto synchronization
	blendedGoal := core.State{
		Phase:     state.LocalGoal,
		Frequency: state.Frequency,
		Coherence: globalGoal.Coherence,
	}

	// Generate proposal using context
	currentState := core.State{
		Phase:     state.Phase,
		Frequency: state.Frequency,
		Coherence: a.calculateLocalCoherence(),
	}

	// Get context
	ctx := a.context.Load().(core.Context)

	proposal, confidence := a.strategy.Propose(currentState, blendedGoal, ctx)

	// Make decision
	options := []core.Action{
		proposal,
		{Type: "maintain", Cost: 0.1, Benefit: ctx.Stability},
	}

	chosen, acceptance := a.decider.Decide(currentState, options)

	// Check energy
	if chosen.Cost > state.Energy {
		return core.Action{Type: "maintain"}, false
	}

	// Accept based on confidence
	if random.Float64() < confidence*acceptance {
		return chosen, true
	}

	return core.Action{Type: "maintain"}, false
}

// ApplyAction executes an action with optimized state updates.
func (a *Agent) ApplyAction(action core.Action) (bool, float64, error) {
	state := a.state.Load()
	energyCost := action.Cost
	
	if energyCost > state.Energy {
		return false, 0, fmt.Errorf("%w: required %.2f, available %.2f",
			core.ErrInsufficientEnergy, energyCost, state.Energy)
	}

	// Apply action and update energy in a single atomic operation
	success := false
	switch action.Type {
	case "adjust_phase", "phase_nudge", "frequency_lock", "energy_save", "pulse":
		a.state.Update(func(s *StateData) {
			s.Phase = core.WrapPhase(s.Phase + action.Value)
			s.Energy = math.Max(0, s.Energy-energyCost)
		})
		success = true
	case "maintain":
		a.state.Update(func(s *StateData) {
			s.Energy = math.Max(0, s.Energy-energyCost)
		})
		success = true
	default:
		return false, 0, fmt.Errorf("%w: %s", core.ErrUnknownActionType, action.Type)
	}

	if success && a.resources != nil {
		a.resources.Request(energyCost)
	}

	return success, energyCost, nil
}

// calculateLocalCoherence efficiently calculates coherence.
func (a *Agent) calculateLocalCoherence() float64 {
	var neighborList []*Agent
	if a.useOptimizedNeighbors {
		neighborList = a.optimizedNeighbors.All()
	} else {
		a.neighbors.Range(func(_, value any) bool {
			if neighbor, ok := value.(*Agent); ok {
				neighborList = append(neighborList, neighbor)
			}
			return true
		})
	}

	if len(neighborList) == 0 {
		return 0
	}

	myPhase := a.state.Load().Phase
	sumCos := 0.0
	sumSin := 0.0

	for _, neighbor := range neighborList {
		diff := neighbor.Phase() - myPhase
		sumCos += math.Cos(diff)
		sumSin += math.Sin(diff)
	}

	return math.Sqrt(sumCos*sumCos+sumSin*sumSin) / float64(len(neighborList))
}

// ============= Option Functions =============

// WithPhase sets initial phase.
func WithPhase(phase float64) Option {
	return func(a *Agent) {
		a.state.Update(func(s *StateData) {
			s.Phase = core.WrapPhase(phase)
		})
	}
}

// WithRandomPhase sets random initial phase.
func WithRandomPhase() Option {
	return func(a *Agent) {
		a.state.Update(func(s *StateData) {
			s.Phase = random.Phase()
		})
	}
}

// WithFrequency sets oscillation frequency.
func WithFrequency(freq time.Duration) Option {
	return func(a *Agent) {
		a.state.Update(func(s *StateData) {
			s.Frequency = freq
		})
	}
}

// WithRandomFrequency sets random frequency.
func WithRandomFrequency() Option {
	return func(a *Agent) {
		baseFreq := 100 * time.Millisecond
		variation := time.Duration(random.Float64()*50) * time.Millisecond
		a.state.Update(func(s *StateData) {
			s.Frequency = baseFreq + variation
		})
	}
}

// WithLocalGoal sets the agent's individual target.
func WithLocalGoal(g float64) Option {
	return func(a *Agent) {
		a.state.Update(func(s *StateData) {
			s.LocalGoal = core.WrapPhase(g)
		})
	}
}

// WithRandomLocalGoal sets random local goal.
func WithRandomLocalGoal() Option {
	return func(a *Agent) {
		a.state.Update(func(s *StateData) {
			s.LocalGoal = random.Phase()
		})
	}
}

// WithEnergy sets initial energy.
func WithEnergy(energy float64) Option {
	return func(a *Agent) {
		a.state.Update(func(s *StateData) {
			s.Energy = math.Max(0, energy)
		})
	}
}

// WithInfluence sets influence weight.
func WithInfluence(influence float64) Option {
	return func(a *Agent) {
		a.SetInfluence(influence)
	}
}

// WithStubbornness sets resistance to change.
func WithStubbornness(stubbornness float64) Option {
	return func(a *Agent) {
		a.SetStubbornness(stubbornness)
	}
}

// WithDecisionMaker sets decision-making component.
func WithDecisionMaker(dm core.DecisionMaker) Option {
	return func(a *Agent) {
		a.decider = dm
	}
}

// WithGoalManager sets goal blending component.
func WithGoalManager(gm goal.Manager) Option {
	return func(a *Agent) {
		a.goalManager = gm
	}
}

// WithResourceManager sets resource management component.
func WithResourceManager(rm core.ResourceManager) Option {
	return func(a *Agent) {
		a.resources = rm
		a.state.Update(func(s *StateData) {
			s.Energy = rm.Available()
		})
	}
}

// WithStrategy sets synchronization strategy.
func WithStrategy(s core.SyncStrategy) Option {
	return func(a *Agent) {
		a.strategy = s
	}
}

// WithSwarmInfo sets swarm configuration.
func WithSwarmInfo(swarmSize, assumedMaxNeighbors int) Option {
	return func(a *Agent) {
		a.swarmSize = swarmSize
		a.assumedMaxNeighbors = assumedMaxNeighbors
	}
}

