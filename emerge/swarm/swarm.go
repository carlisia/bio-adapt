package swarm

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/carlisia/bio-adapt/emerge/agent"
	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/monitoring"
	"github.com/carlisia/bio-adapt/internal/config"
	"github.com/carlisia/bio-adapt/internal/emerge"
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
}

// Option configures a Swarm
type Option func(*Swarm) error

// NewSwarm creates a swarm with the provided options.
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
	s.agents.Range(func(key, value any) bool {
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

	return s, nil
}

// createDefaultAgents creates agents with config-based settings
func (s *Swarm) createDefaultAgents() error {
	agentConfig := config.AgentFromSwarm(s.config)
	agentConfig.SwarmSize = s.size

	for i := range s.size {
		agent, err := agent.NewFromConfig(fmt.Sprintf("agent-%d", i), agentConfig)
		if err != nil {
			return fmt.Errorf("failed to create agent %d: %w", i, err)
		}
		s.agents.Store(agent.ID, agent)
	}
	return nil
}

// WithConfig sets a custom configuration
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

// WithMonitor sets a custom monitor
func WithMonitor(monitor *monitoring.Monitor) Option {
	return func(s *Swarm) error {
		s.monitor = monitor
		return nil
	}
}

// WithAgentBuilder uses a custom function to create agents
func WithAgentBuilder(builder func(id string, swarmSize int, cfg config.Swarm) *agent.Agent) Option {
	return func(s *Swarm) error {
		for i := range s.size {
			agent := builder(fmt.Sprintf("agent-%d", i), s.size, s.config)
			s.agents.Store(agent.ID, agent)
		}
		return nil
	}
}

// WithTopology uses a custom topology builder
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
	// Collect all agents first
	var agents []*agent.Agent
	s.agents.Range(func(key, value any) bool {
		agents = append(agents, value.(*agent.Agent))
		return true
	})

	// Use configurable threshold for connection optimization
	if s.config.EnableConnectionOptim && len(agents) > s.config.ConnectionOptimThreshold {
		// Just establish minimal random connections for large swarms
		for _, agent := range agents {
			for i := 0; i < s.config.MinNeighbors && i < len(agents)-1; i++ {
				idx := rand.Intn(len(agents))
				if agents[idx].ID != agent.ID {
					agent.Neighbors().Store(agents[idx].ID, agents[idx])
					agents[idx].Neighbors().Store(agent.ID, agent)
				}
			}
		}
		return
	}

	// Connect agents based on configuration for smaller swarms
	for i, agent := range agents {
		connected := 0

		// Try to connect to other agents
		for j, neighbor := range agents {
			if i == j {
				continue
			}

			// Check if we've reached max neighbors
			if connected >= s.config.MaxNeighbors {
				break
			}

			// Check if already connected
			if _, exists := agent.Neighbors().Load(neighbor.ID); exists {
				connected++
				continue
			}

			// Connect based on probability
			if rand.Float64() < s.config.ConnectionProbability {
				agent.Neighbors().Store(neighbor.ID, neighbor)
				neighbor.Neighbors().Store(agent.ID, agent)
				connected++
			}
		}

		// Ensure minimum connectivity
		if connected < s.config.MinNeighbors && len(agents) > s.config.MinNeighbors {
			// Force connect to random neighbors
			for connected < s.config.MinNeighbors {
				idx := rand.Intn(len(agents))
				neighbor := agents[idx]

				if neighbor.ID == agent.ID {
					continue
				}

				if _, exists := agent.Neighbors().Load(neighbor.ID); !exists {
					agent.Neighbors().Store(neighbor.ID, neighbor)
					neighbor.Neighbors().Store(agent.ID, agent)
					connected++
				}
			}
		}
	}
}

// Run starts all agents autonomously - no central orchestration.
// Synchronization emerges from local interactions and gossip.
func (s *Swarm) Run(ctx context.Context) error {
	if s.config.UseBatchProcessing {
		if err := s.runWithBatchProcessing(ctx); err != nil {
			return fmt.Errorf("batch processing failed: %w", err)
		}
		return nil
	}
	if err := s.runWithDirectProcessing(ctx); err != nil {
		return fmt.Errorf("direct processing failed: %w", err)
	}
	return nil
}

// runWithDirectProcessing runs each agent in its own goroutine (original behavior)
func (s *Swarm) runWithDirectProcessing(ctx context.Context) error {
	var wg sync.WaitGroup

	// Collect all agents
	var agents []*agent.Agent
	s.agents.Range(func(key, value any) bool {
		agents = append(agents, value.(*agent.Agent))
		return true
	})

	// Apply concurrency limit if configured
	concurrencyLimit := s.config.MaxConcurrentAgents
	if concurrencyLimit > 0 && len(agents) > concurrencyLimit {
		// Use worker pool with limited goroutines
		if err := s.runWithWorkerPool(ctx, agents); err != nil {
			return fmt.Errorf("worker pool execution failed: %w", err)
		}
		return nil
	}

	// Start each agent independently
	for _, a := range agents {
		wg.Add(1)
		go func(ag *agent.Agent) {
			defer wg.Done()
			s.runAgentLoop(ctx, ag)
		}(a)
	}

	// Monitor convergence (observation only, no control)
	go s.monitorConvergence(ctx)

	wg.Wait()
	return nil
}

// runWithWorkerPool runs agents using a limited worker pool
func (s *Swarm) runWithWorkerPool(ctx context.Context, agents []*agent.Agent) error {
	workerCount := s.config.WorkerPoolSize
	if workerCount == 0 {
		workerCount = s.config.MaxConcurrentAgents
	}

	agentChan := make(chan *agent.Agent, len(agents))
	var wg sync.WaitGroup

	// Start workers
	for range workerCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for agent := range agentChan {
				select {
				case <-ctx.Done():
					return
				default:
					s.runAgentCycle(agent)
				}
			}
		}()
	}

	// Monitor convergence
	go s.monitorConvergence(ctx)

	// Continuously feed agents to workers
	go func() {
		defer close(agentChan)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				for _, agent := range agents {
					select {
					case agentChan <- agent:
					case <-ctx.Done():
						return
					}
				}
				// Small delay between cycles
				time.Sleep(s.config.AgentUpdateInterval)
			}
		}
	}()

	wg.Wait()
	return nil
}

// runWithBatchProcessing processes agents in batches
func (s *Swarm) runWithBatchProcessing(ctx context.Context) error {
	// Collect all agents
	var agents []*agent.Agent
	s.agents.Range(func(key, value any) bool {
		agents = append(agents, value.(*agent.Agent))
		return true
	})

	batchSize := s.config.BatchSize
	if batchSize == 0 {
		batchSize = len(agents) / 10 // Default to 10 batches
		if batchSize < 1 {
			batchSize = 1
		}
	}

	// Monitor convergence
	go s.monitorConvergence(ctx)

	// Process agents in batches continuously
	ticker := time.NewTicker(s.config.AgentUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			s.processBatch(agents, batchSize)
		}
	}
}

// processBatch processes a batch of agents concurrently
func (s *Swarm) processBatch(agents []*agent.Agent, batchSize int) {
	var wg sync.WaitGroup

	for i := 0; i < len(agents); i += batchSize {
		end := i + batchSize
		if end > len(agents) {
			end = len(agents)
		}

		batch := agents[i:end]
		wg.Add(1)
		go func(agentBatch []*agent.Agent) {
			defer wg.Done()
			for _, agent := range agentBatch {
				s.runAgentCycle(agent)
			}
		}(batch)
	}

	wg.Wait()
}

// runAgentLoop runs a single agent continuously
func (s *Swarm) runAgentLoop(ctx context.Context, agent *agent.Agent) {
	ticker := time.NewTicker(s.config.AgentUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runAgentCycle(agent)
		}
	}
}

// runAgentCycle runs a single cycle of agent processing
func (s *Swarm) runAgentCycle(agent *agent.Agent) {
	// Update context from local observations
	agent.UpdateContext()

	// Make autonomous decision
	action, accepted := agent.ProposeAdjustment(s.goalState)

	if accepted {
		success, energyCost, err := agent.ApplyAction(action)
		if !success {
			// Action failed - log the detailed error for debugging while maintaining autonomous behavior
			// This is expected behavior in autonomous systems where agents can fail due to resource constraints
			if err != nil {
				// In a production system, this could be logged to monitoring systems
				// For now, we silently handle the failure as part of autonomous behavior
				_ = err // err contains detailed context like "insufficient energy: required 5.20, available 3.40"
			}
			return
		}
		// Successfully applied action with energy cost
		// The energy is already deducted from the agent's resource manager
		// We could use energyCost for monitoring/metrics here if needed
		_ = energyCost // Intentionally unused - energy tracking is internal
	}
}

// monitorConvergence observes emergent behavior without controlling it.
func (s *Swarm) monitorConvergence(ctx context.Context) {
	ticker := time.NewTicker(s.config.MonitoringInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			coherence := s.MeasureCoherence()
			s.monitor.RecordSample(coherence)
			s.convergence.Record(coherence)
		}
	}
}

// MeasureCoherence calculates global synchronization level.
// This is for monitoring only - agents don't have access to this.
func (s *Swarm) MeasureCoherence() float64 {
	var phases []float64

	s.agents.Range(func(key, value any) bool {
		agent := value.(*agent.Agent)
		phases = append(phases, agent.Phase())
		return true
	})

	return core.MeasureCoherence(phases)
}

// Agents returns all agents in the swarm.
func (s *Swarm) Agents() map[string]*agent.Agent {
	agents := make(map[string]*agent.Agent)
	s.agents.Range(func(key, value interface{}) bool {
		agents[key.(string)] = value.(*agent.Agent)
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
	return value.(*agent.Agent), true
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

	s.agents.Range(func(key, value any) bool {
		if disrupted >= targetCount {
			return false
		}

		agent := value.(*agent.Agent)
		agent.SetPhase(rand.Float64() * 2 * math.Pi)
		disrupted++
		return true
	})
}

// ForEachAgent applies a function to each agent in the swarm.
// This provides controlled access to agents without exposing the internal map.
func (s *Swarm) ForEachAgent(fn func(*agent.Agent) bool) {
	s.agents.Range(func(key, value any) bool {
		agent := value.(*agent.Agent)
		return fn(agent)
	})
}

// AutoScaleConfig returns a configuration that automatically scales based on swarm size.
func AutoScaleConfig(swarmSize int) config.Swarm {
	return config.AutoScaleConfig(swarmSize)
}

// ConfigForBatching returns configuration optimized for request batching scenarios.
func ConfigForBatching(workloadCount int, batchWindow time.Duration) config.Swarm {
	return config.ConfigForBatching(workloadCount, batchWindow)
}
