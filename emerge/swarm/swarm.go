package swarm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/carlisia/bio-adapt/emerge/agent"
	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/monitoring"
	"github.com/carlisia/bio-adapt/internal/config"
	"github.com/carlisia/bio-adapt/internal/emerge"
	"github.com/carlisia/bio-adapt/internal/random"
)

// Swarm represents a collection of autonomous agents achieving
// synchronization through emergent behavior - no central control.
type Swarm struct {
	agents    sync.Map // map[string]*agent.Agent
	goalState core.State
	config    config.Swarm
	size      int

	// Monitoring (read-only, doesn't control) - private implementation details
	monitor     *monitoring.Monitor
	basin       *emerge.AttractorBasin
	convergence *monitoring.ConvergenceMonitor

	// Track initialization state
	connectionsEstablished bool

	// Goal-directed synchronization
	goalDirectedSync *GoalDirectedSync
}

// Option configures a Swarm.
type Option func(*Swarm) error

// New creates a swarm with the provided options.
func New(size int, goal core.State, opts ...Option) (*Swarm, error) {
	if size <= 0 {
		return nil, fmt.Errorf("%w: got %d", ErrInvalidSwarmSize, size)
	}

	// Validate goal state using centralized validation
	if err := goal.Validate(); err != nil {
		return nil, fmt.Errorf("invalid goal state: %w", err)
	}

	// Start with auto-scaled config as default
	cfg := config.AutoScaleConfig(size)
	if err := cfg.NormalizeAndValidate(size); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	s := &Swarm{
		goalState:   goal,
		config:      cfg,
		size:        size,
		monitor:     monitoring.New(),
		basin:       emerge.NewAttractorBasin(goal, cfg.BasinStrength, cfg.BasinWidth),
		convergence: monitoring.NewConvergence(goal, goal.Coherence),
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, fmt.Errorf("failed to apply swarm option: %w", err)
		}
	}

	// Re-validate config after options are applied
	if err := s.config.NormalizeAndValidate(size); err != nil {
		return nil, fmt.Errorf("config validation failed after options: %w", err)
	}

	// Check configurable size limit after options are applied
	if s.config.MaxSwarmSize > 0 && size > s.config.MaxSwarmSize {
		return nil, fmt.Errorf("swarm size %d exceeds configured maximum %d", size, s.config.MaxSwarmSize)
	}

	// Create agents if not already created by options
	agentCount := 0
	s.agents.Range(func(_, _ any) bool {
		agentCount++
		return false // Just need to check if any exist
	})
	if agentCount == 0 {
		if err := s.createDefaultAgents(); err != nil {
			return nil, fmt.Errorf("failed to create agents: %w", err)
		}
	}

	// Establish connections if not already done
	if !s.connectionsEstablished {
		s.establishConnections()
	}

	// Initialize goal-directed synchronization
	s.goalDirectedSync = NewGoalDirectedSync(s)

	return s, nil
}

// NewSwarmFromConfig creates a swarm using a configuration struct.
// This provides an alternative to the functional options pattern.
func NewSwarmFromConfig(size int, goal core.State, cfg config.Swarm) (*Swarm, error) {
	if size <= 0 {
		return nil, fmt.Errorf("%w: got %d", ErrInvalidSwarmSize, size)
	}

	// Validate goal state using centralized validation
	if err := goal.Validate(); err != nil {
		return nil, fmt.Errorf("invalid goal state: %w", err)
	}

	// Validate and normalize config using centralized validation
	if err := cfg.NormalizeAndValidate(size); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	s := &Swarm{
		goalState:   goal,
		config:      cfg,
		size:        size,
		monitor:     monitoring.New(),
		basin:       emerge.NewAttractorBasin(goal, cfg.BasinStrength, cfg.BasinWidth),
		convergence: monitoring.NewConvergence(goal, goal.Coherence),
	}

	// Create agents
	if err := s.createDefaultAgents(); err != nil {
		return nil, fmt.Errorf("failed to create agents: %w", err)
	}

	// Establish connections
	s.establishConnections()

	// Initialize goal-directed synchronization
	s.goalDirectedSync = NewGoalDirectedSync(s)

	return s, nil
}

// createDefaultAgents creates agents with config-based settings.
func (s *Swarm) createDefaultAgents() error {
	agentConfig := config.AgentFromSwarm(s.config)
	agentConfig.SwarmSize = s.size

	for i := range s.size {
		a, err := agent.NewFromConfig(fmt.Sprintf("agent-%d", i), agentConfig)
		if err != nil {
			return fmt.Errorf("failed to create agent %d: %w", i, err)
		}
		s.agents.Store(a.ID, a)
	}
	return nil
}

// WithConfig sets a custom configuration.
func WithConfig(cfg config.Swarm) Option {
	return func(s *Swarm) error {
		if err := cfg.NormalizeAndValidate(s.size); err != nil {
			return fmt.Errorf("config validation failed: %w", err)
		}
		s.config = cfg
		// Update basin and convergence with new config
		s.basin = emerge.NewAttractorBasin(s.goalState, cfg.BasinStrength, cfg.BasinWidth)
		return nil
	}
}

// WithMonitor sets a custom monitor.
func WithMonitor(monitor *monitoring.Monitor) Option {
	return func(s *Swarm) error {
		s.monitor = monitor
		return nil
	}
}

// WithAgentBuilder uses a custom function to create agents.
func WithAgentBuilder(builder func(id string, swarmSize int, cfg config.Swarm) *agent.Agent) Option {
	return func(s *Swarm) error {
		for i := range s.size {
			a := builder(fmt.Sprintf("agent-%d", i), s.size, s.config)
			s.agents.Store(a.ID, a)
		}
		return nil
	}
}

// WithTopology uses a custom topology builder.
func WithTopology(builder func(*Swarm) error) Option {
	return func(s *Swarm) error {
		if err := builder(s); err != nil {
			return fmt.Errorf("topology build failed: %w", err)
		}
		s.connectionsEstablished = true
		return nil
	}
}

// establishConnections creates network topology based on configuration.
func (s *Swarm) establishConnections() {
	agents := s.collectAgents()

	if s.config.EnableConnectionOptim && len(agents) > s.config.ConnectionOptimThreshold {
		s.establishMinimalConnections(agents)
	} else {
		s.establishProbabilisticConnections(agents)
	}
}

// collectAgents gathers all agents from the swarm
func (s *Swarm) collectAgents() []*agent.Agent {
	var agents []*agent.Agent
	s.agents.Range(func(_, value any) bool {
		if a, ok := value.(*agent.Agent); ok {
			agents = append(agents, a)
		}
		return true
	})
	return agents
}

// establishMinimalConnections creates minimal random connections for large swarms
func (s *Swarm) establishMinimalConnections(agents []*agent.Agent) {
	for _, a := range agents {
		for i := 0; i < s.config.MinNeighbors && i < len(agents)-1; i++ {
			idx := random.Intn(len(agents))
			if agents[idx].ID != a.ID {
				a.Neighbors().Store(agents[idx].ID, agents[idx])
				agents[idx].Neighbors().Store(a.ID, a)
			}
		}
	}
}

// establishProbabilisticConnections creates connections based on probability for smaller swarms
func (s *Swarm) establishProbabilisticConnections(agents []*agent.Agent) {
	for i, a := range agents {
		connected := s.connectToNeighbors(a, agents, i)
		s.ensureMinimumConnectivity(a, agents, connected)
	}
}

// connectToNeighbors connects an agent to neighbors based on probability
func (s *Swarm) connectToNeighbors(a *agent.Agent, agents []*agent.Agent, agentIndex int) int {
	connected := 0

	for j, neighbor := range agents {
		if j == agentIndex {
			continue
		}

		if connected >= s.config.MaxNeighbors {
			break
		}

		if _, exists := a.Neighbors().Load(neighbor.ID); exists {
			connected++
			continue
		}

		if random.Float64() < s.config.ConnectionProbability {
			a.Neighbors().Store(neighbor.ID, neighbor)
			neighbor.Neighbors().Store(a.ID, a)
			connected++
		}
	}

	return connected
}

// ensureMinimumConnectivity forces connections to meet minimum neighbor requirements
func (s *Swarm) ensureMinimumConnectivity(a *agent.Agent, agents []*agent.Agent, connected int) {
	if connected < s.config.MinNeighbors && len(agents) > s.config.MinNeighbors {
		for connected < s.config.MinNeighbors {
			idx := random.Intn(len(agents))
			neighbor := agents[idx]

			if neighbor.ID == a.ID {
				continue
			}

			if _, exists := a.Neighbors().Load(neighbor.ID); !exists {
				a.Neighbors().Store(neighbor.ID, neighbor)
				neighbor.Neighbors().Store(a.ID, a)
				connected++
			}
		}
	}
}

// Run starts the swarm and achieves synchronization using bioelectric pattern completion.
// Uses adaptive strategies and attractor basin dynamics to ensure convergence.
func (s *Swarm) Run(ctx context.Context) error {
	// Create target pattern from goal state
	targetPattern := &core.RhythmicPattern{
		Phase:     s.goalState.Phase,
		Frequency: s.goalState.Frequency,
		Coherence: s.goalState.Coherence,
		Amplitude: 1.0,
		Stability: 0.9,
	}

	// Use goal-directed synchronization
	return s.goalDirectedSync.AchieveSynchronization(ctx, targetPattern)
}

// MeasureCoherence calculates global synchronization level.
// This is for monitoring only - agents don't have access to this.
func (s *Swarm) MeasureCoherence() float64 {
	var phases []float64

	s.agents.Range(func(_, value any) bool {
		if a, ok := value.(*agent.Agent); ok {
			phases = append(phases, a.Phase())
		}
		return true
	})

	return core.MeasureCoherence(phases)
}

// Agents returns all agents in the swarm.
func (s *Swarm) Agents() map[string]*agent.Agent {
	agents := make(map[string]*agent.Agent)
	s.agents.Range(func(key, value interface{}) bool {
		if k, ok := key.(string); ok {
			if a, ok := value.(*agent.Agent); ok {
				agents[k] = a
			}
		}
		return true
	})
	return agents
}

// Agent retrieves an agent by ID.
func (s *Swarm) Agent(id string) (*agent.Agent, bool) {
	value, exists := s.agents.Load(id)
	if !exists {
		return nil, false
	}
	if a, ok := value.(*agent.Agent); ok {
		return a, true
	}
	return nil, false
}

// Size returns the number of agents in the swarm.
func (s *Swarm) Size() int {
	return s.size
}

// Config returns the swarm configuration.
func (s *Swarm) Config() config.Swarm {
	return s.config
}

// IsConverged returns whether the swarm has reached convergence.
func (s *Swarm) IsConverged() bool {
	return s.convergence.IsConverged()
}

// ConvergenceTime returns how long it took to converge, or 0 if not converged.
func (s *Swarm) ConvergenceTime() time.Duration {
	return s.convergence.ConvergenceTime()
}

// CurrentCoherence returns the most recent coherence measurement.
func (s *Swarm) CurrentCoherence() float64 {
	return s.convergence.CurrentCoherence()
}

// DisruptAgents randomly disrupts a percentage of agents.
func (s *Swarm) DisruptAgents(percentage float64) {
	targetCount := int(float64(s.Size()) * percentage)
	disrupted := 0

	s.agents.Range(func(_, value any) bool {
		if disrupted >= targetCount {
			return false
		}

		if a, ok := value.(*agent.Agent); ok {
			a.SetPhase(random.Phase())
			disrupted++
		}
		return true
	})
}

// ForEachAgent applies a function to each agent in the swarm.
// This provides controlled access to agents without exposing the internal map.
func (s *Swarm) ForEachAgent(fn func(*agent.Agent) bool) {
	s.agents.Range(func(_, value any) bool {
		if agent, ok := value.(*agent.Agent); ok {
			return fn(agent)
		}
		return true
	})
}

// AutoScaleConfig returns a configuration that automatically scales based on swarm size.
func AutoScaleConfig(swarmSize int) config.Swarm {
	return config.AutoScaleConfig(swarmSize)
}

// ConfigForBatching returns configuration optimized for request batching scenarios.
func ConfigForBatching(workloadCount int, batchWindow time.Duration) config.Swarm {
	return config.ForBatching(workloadCount, batchWindow)
}
