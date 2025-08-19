package emerge

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// Swarm represents a collection of autonomous agents achieving
// synchronization through emergent behavior - no central control.
type Swarm struct {
	agents    sync.Map // map[string]*Agent
	goalState State
	config    SwarmConfig
	size      int

	// Monitoring (read-only, doesn't control)
	monitor     *Monitor
	basin       *AttractorBasin
	convergence *ConvergenceMonitor

	// Track initialization state
	connectionsEstablished bool
}

// SwarmOption configures a Swarm
type SwarmOption func(*Swarm) error

// NewSwarm creates a swarm with the provided options.
func NewSwarm(size int, goal State, opts ...SwarmOption) (*Swarm, error) {
	if size <= 0 {
		return nil, fmt.Errorf("%w: got %d", ErrInvalidSwarmSize, size)
	}

	// Validate goal state using centralized validation
	if err := goal.Validate(); err != nil {
		return nil, fmt.Errorf("invalid goal state: %w", err)
	}

	// Start with auto-scaled config as default
	config := AutoScaleConfig(size)
	if err := config.NormalizeAndValidate(size); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	s := &Swarm{
		goalState:   goal,
		config:      config,
		size:        size,
		monitor:     NewMonitor(),
		basin:       NewAttractorBasin(goal, config.BasinStrength, config.BasinWidth),
		convergence: NewConvergenceMonitor(goal, goal.Coherence),
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
		s.createDefaultAgents()
	}

	// Establish connections if not already done
	if !s.connectionsEstablished {
		s.establishConnections()
	}

	return s, nil
}

// NewSwarmFromConfig creates a swarm using a configuration struct.
// This provides an alternative to the functional options pattern.
func NewSwarmFromConfig(size int, goal State, config SwarmConfig) (*Swarm, error) {
	if size <= 0 {
		return nil, fmt.Errorf("%w: got %d", ErrInvalidSwarmSize, size)
	}

	// Validate goal state using centralized validation
	if err := goal.Validate(); err != nil {
		return nil, fmt.Errorf("invalid goal state: %w", err)
	}

	// Validate and normalize config using centralized validation
	if err := config.NormalizeAndValidate(size); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	s := &Swarm{
		goalState:   goal,
		config:      config,
		size:        size,
		monitor:     NewMonitor(),
		basin:       NewAttractorBasin(goal, config.BasinStrength, config.BasinWidth),
		convergence: NewConvergenceMonitor(goal, goal.Coherence),
	}

	// Create agents
	s.createDefaultAgents()

	// Establish connections
	s.establishConnections()

	return s, nil
}

// createDefaultAgents creates agents with config-based settings
func (s *Swarm) createDefaultAgents() {
	agentConfig := AgentConfigFromSwarmConfig(s.config)
	agentConfig.SwarmSize = s.size

	for i := range s.size {
		agent := NewAgentFromConfig(fmt.Sprintf("agent-%d", i), agentConfig)
		s.agents.Store(agent.ID, agent)
	}
}

// WithConfig sets a custom configuration
func WithConfig(config SwarmConfig) SwarmOption {
	return func(s *Swarm) error {
		if err := config.NormalizeAndValidate(s.size); err != nil {
			return fmt.Errorf("config validation failed: %w", err)
		}
		s.config = config
		// Update basin and convergence with new config
		s.basin = NewAttractorBasin(s.goalState, config.BasinStrength, config.BasinWidth)
		return nil
	}
}

// WithMonitor sets a custom monitor
func WithMonitor(monitor *Monitor) SwarmOption {
	return func(s *Swarm) error {
		s.monitor = monitor
		return nil
	}
}

// WithConvergenceMonitor sets a custom convergence monitor
func WithConvergenceMonitor(cm *ConvergenceMonitor) SwarmOption {
	return func(s *Swarm) error {
		s.convergence = cm
		return nil
	}
}

// WithAgentBuilder uses a custom function to create agents
func WithAgentBuilder(builder func(id string, swarmSize int, config SwarmConfig) *Agent) SwarmOption {
	return func(s *Swarm) error {
		for i := range s.size {
			agent := builder(fmt.Sprintf("agent-%d", i), s.size, s.config)
			s.agents.Store(agent.ID, agent)
		}
		return nil
	}
}

// WithTopology uses a custom topology builder
func WithTopology(builder func(*Swarm) error) SwarmOption {
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
	var agents []*Agent
	s.agents.Range(func(key, value any) bool {
		agents = append(agents, value.(*Agent))
		return true
	})

	// Use configurable threshold for connection optimization
	if s.config.EnableConnectionOptim && len(agents) > s.config.ConnectionOptimThreshold {
		// Just establish minimal random connections for large swarms
		for _, agent := range agents {
			for i := 0; i < s.config.MinNeighbors && i < len(agents)-1; i++ {
				idx := rand.Intn(len(agents))
				if agents[idx].ID != agent.ID {
					agent.neighbors.Store(agents[idx].ID, agents[idx])
					agents[idx].neighbors.Store(agent.ID, agent)
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
			if _, exists := agent.neighbors.Load(neighbor.ID); exists {
				connected++
				continue
			}

			// Connect based on probability
			if rand.Float64() < s.config.ConnectionProbability {
				agent.neighbors.Store(neighbor.ID, neighbor)
				neighbor.neighbors.Store(agent.ID, agent)
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

				if _, exists := agent.neighbors.Load(neighbor.ID); !exists {
					agent.neighbors.Store(neighbor.ID, neighbor)
					neighbor.neighbors.Store(agent.ID, agent)
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
	var agents []*Agent
	s.agents.Range(func(key, value any) bool {
		agents = append(agents, value.(*Agent))
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
	for _, agent := range agents {
		wg.Add(1)
		go func(a *Agent) {
			defer wg.Done()
			s.runAgentLoop(ctx, a)
		}(agent)
	}

	// Monitor convergence (observation only, no control)
	go s.monitorConvergence(ctx)

	wg.Wait()
	return nil
}

// runWithWorkerPool runs agents using a limited worker pool
func (s *Swarm) runWithWorkerPool(ctx context.Context, agents []*Agent) error {
	workerCount := s.config.WorkerPoolSize
	if workerCount == 0 {
		workerCount = s.config.MaxConcurrentAgents
	}

	agentChan := make(chan *Agent, len(agents))
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
	var agents []*Agent
	s.agents.Range(func(key, value any) bool {
		agents = append(agents, value.(*Agent))
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
func (s *Swarm) processBatch(agents []*Agent, batchSize int) {
	var wg sync.WaitGroup

	for i := 0; i < len(agents); i += batchSize {
		end := i + batchSize
		if end > len(agents) {
			end = len(agents)
		}

		batch := agents[i:end]
		wg.Add(1)
		go func(agentBatch []*Agent) {
			defer wg.Done()
			for _, agent := range agentBatch {
				s.runAgentCycle(agent)
			}
		}(batch)
	}

	wg.Wait()
}

// runAgentLoop runs a single agent continuously
func (s *Swarm) runAgentLoop(ctx context.Context, agent *Agent) {
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
func (s *Swarm) runAgentCycle(agent *Agent) {
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
		agent := value.(*Agent)
		phases = append(phases, agent.Phase())
		return true
	})

	return MeasureCoherence(phases)
}

// GetAgent retrieves an agent by ID.
func (s *Swarm) GetAgent(id string) (*Agent, bool) {
	value, exists := s.agents.Load(id)
	if !exists {
		return nil, false
	}
	return value.(*Agent), true
}

// Size returns the number of agents in the swarm.
func (s *Swarm) Size() int {
	return s.size
}

// Config returns the swarm configuration.
func (s *Swarm) Config() SwarmConfig {
	return s.config
}

// GetMonitor returns the swarm's monitor.
func (s *Swarm) GetMonitor() *Monitor {
	return s.monitor
}

// DisruptAgents randomly disrupts a percentage of agents.
func (s *Swarm) DisruptAgents(percentage float64) {
	targetCount := int(float64(s.Size()) * percentage)
	disrupted := 0

	s.agents.Range(func(key, value any) bool {
		if disrupted >= targetCount {
			return false
		}

		agent := value.(*Agent)
		agent.SetPhase(rand.Float64() * 2 * math.Pi)
		disrupted++
		return true
	})
}

// Agents returns the internal agents map for direct access.
func (s *Swarm) Agents() *sync.Map {
	return &s.agents
}

// GetBasin returns the swarm's attractor basin.
func (s *Swarm) GetBasin() *AttractorBasin {
	return s.basin
}

// GetConvergenceMonitor returns the convergence monitor.
func (s *Swarm) GetConvergenceMonitor() *ConvergenceMonitor {
	return s.convergence
}
