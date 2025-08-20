package agent

import (
	"math"
	"sync"

	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/decision"
	"github.com/carlisia/bio-adapt/emerge/goal"
	"github.com/carlisia/bio-adapt/emerge/strategy"
	"github.com/carlisia/bio-adapt/internal/config"
	"github.com/carlisia/bio-adapt/internal/resource"
)

// OptimizedAgent uses pre-allocated neighbor storage for better performance.
// This is used automatically for large swarms (>100 agents).
type OptimizedAgent struct {
	*Agent // Embed standard agent for most functionality

	// Optimized neighbor storage
	optimizedNeighbors *NeighborStorage
	useOptimized       bool
}

// NewOptimized creates an agent with optimized neighbor storage.
func NewOptimized(id string, maxNeighbors int, opts ...Option) *OptimizedAgent {
	// Create base agent
	base := New(id, opts...)

	// Create optimized storage
	oa := &OptimizedAgent{
		Agent:              base,
		optimizedNeighbors: NewNeighborStorage(maxNeighbors),
		useOptimized:       true,
	}

	return oa
}

// NewFromConfig creates an optimized agent from configuration.
func NewOptimizedFromConfig(id string, cfg config.Agent) (*OptimizedAgent, error) {
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

	return NewOptimized(id, maxNeighbors, opts...), nil
}

// Neighbors returns a compatibility wrapper for the optimized storage.
func (oa *OptimizedAgent) Neighbors() *sync.Map {
	if !oa.useOptimized {
		return oa.Agent.Neighbors()
	}

	// Create a sync.Map wrapper for compatibility
	// This is only used during connection establishment
	wrapper := &sync.Map{}
	oa.optimizedNeighbors.Range(func(id string, agent *Agent) bool {
		wrapper.Store(id, agent)
		return true
	})
	return wrapper
}

// ConnectTo establishes a bidirectional connection.
func (oa *OptimizedAgent) ConnectTo(other *Agent) {
	if other == nil || other.ID == oa.ID {
		return
	}

	if oa.useOptimized {
		// Use optimized storage
		oa.optimizedNeighbors.Store(other.ID, other)

		// Handle other side of connection
		// Since we can't type assert from *Agent to *OptimizedAgent (it embeds, not implements),
		// we just use the standard Neighbors() method which OptimizedAgent overrides
		other.Neighbors().Store(oa.ID, oa.Agent)
	} else {
		// Fall back to standard implementation
		oa.Agent.ConnectTo(other)
	}
}

// DisconnectFrom removes a connection.
func (oa *OptimizedAgent) DisconnectFrom(other *Agent) {
	if other == nil {
		return
	}

	if oa.useOptimized {
		oa.optimizedNeighbors.Delete(other.ID)

		// Handle other side
		other.Neighbors().Delete(oa.ID)
	} else {
		oa.Agent.DisconnectFrom(other)
	}
}

// IsConnectedTo checks if connected to another agent.
func (oa *OptimizedAgent) IsConnectedTo(other *Agent) bool {
	if other == nil {
		return false
	}

	if oa.useOptimized {
		_, exists := oa.optimizedNeighbors.Load(other.ID)
		return exists
	}
	return oa.Agent.IsConnectedTo(other)
}

// NeighborCount returns the number of connected neighbors.
func (oa *OptimizedAgent) NeighborCount() int {
	if oa.useOptimized {
		return oa.optimizedNeighbors.Count()
	}
	return oa.Agent.NeighborCount()
}

// UpdateContext updates the agent's perception efficiently.
func (oa *OptimizedAgent) UpdateContext() {
	if !oa.useOptimized {
		oa.Agent.UpdateContext()
		return
	}

	// Optimized context update using direct slice access
	neighbors := oa.optimizedNeighbors.All()
	if len(neighbors) == 0 {
		oa.context.Store(core.Context{
			Neighbors:      0,
			Density:        0,
			LocalCoherence: 0,
			Stability:      oa.calculateStability(),
		})
		return
	}

	myPhase := oa.phase.Load()
	sumCos := 0.0
	sumSin := 0.0

	// Direct slice iteration is more cache-friendly
	for _, neighbor := range neighbors {
		diff := neighbor.Phase() - myPhase
		sumCos += math.Cos(diff)
		sumSin += math.Sin(diff)
	}

	neighborCount := len(neighbors)
	localCoherence := math.Sqrt(sumCos*sumCos+sumSin*sumSin) / float64(neighborCount)

	// Kuramoto coupling
	if neighborCount > 0 {
		avgSin := sumSin / float64(neighborCount)
		avgCos := sumCos / float64(neighborCount)
		targetPhaseShift := math.Atan2(avgSin, avgCos)

		couplingStrength := 0.5 + 0.5*localCoherence
		effectiveShift := targetPhaseShift * couplingStrength

		oa.localGoal.Store(core.WrapPhase(myPhase + effectiveShift))
	}

	// Calculate density
	maxNeighbors := oa.assumedMaxNeighbors
	if maxNeighbors == 0 {
		maxNeighbors = oa.swarmSize - 1
	}
	density := float64(neighborCount) / float64(maxNeighbors)

	// Store context
	oa.context.Store(core.Context{
		Neighbors:      neighborCount,
		Density:        density,
		LocalCoherence: localCoherence,
		Stability:      oa.calculateStability(),
	})
}

// calculateLocalCoherence efficiently calculates coherence with neighbors.
func (oa *OptimizedAgent) calculateLocalCoherence() float64 {
	if !oa.useOptimized {
		return oa.Agent.calculateLocalCoherence()
	}

	neighbors := oa.optimizedNeighbors.All()
	if len(neighbors) == 0 {
		return 0
	}

	myPhase := oa.phase.Load()
	sumCos := 0.0
	sumSin := 0.0

	for _, neighbor := range neighbors {
		diff := neighbor.Phase() - myPhase
		sumCos += math.Cos(diff)
		sumSin += math.Sin(diff)
	}

	return math.Sqrt(sumCos*sumCos+sumSin*sumSin) / float64(len(neighbors))
}
