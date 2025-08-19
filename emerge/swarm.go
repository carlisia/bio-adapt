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
		return nil, fmt.Errorf("swarm size must be positive, got %d", size)
	}
	if size > 1000000 { // Reasonable limit to prevent infinite loops
		return nil, fmt.Errorf("swarm size too large, got %d (max 1,000,000)", size)
	}
	if goal.Frequency <= 0 {
		return nil, fmt.Errorf("goal frequency must be positive, got %v", goal.Frequency)
	}
	if math.IsNaN(goal.Coherence) || goal.Coherence < 0 || goal.Coherence > 1 {
		return nil, fmt.Errorf("goal coherence must be in [0, 1], got %f", goal.Coherence)
	}

	// Start with auto-scaled config as default
	config := AutoScaleConfig(size)
	config.Validate(size)

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
			return nil, err
		}
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
		config.Validate(s.size)
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
			return fmt.Errorf("topology builder failed: %w", err)
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

	// For very large swarms, skip full connection establishment to avoid O(n²) performance
	if len(agents) > 50000 {
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
	var wg sync.WaitGroup

	// Start each agent independently
	s.agents.Range(func(key, value any) bool {
		agent := value.(*Agent)

		wg.Add(1)
		go func(a *Agent) {
			defer wg.Done()

			// Each agent runs autonomously
			for {
				select {
				case <-ctx.Done():
					return
				default:
					// Update context from local observations
					a.UpdateContext()

					// Make autonomous decision
					action, accepted := a.ProposeAdjustment(s.goalState)

					if accepted {
						success, energyCost := a.ApplyAction(action)
						if !success {
							// Action failed - agent may be out of energy or action type unknown
							// This is expected behavior in autonomous systems where agents
							// can fail to execute actions due to resource constraints
							continue
						}
						// Successfully applied action with energy cost
						// The energy is already deducted from the agent's resource manager
						// We could use energyCost for monitoring/metrics here if needed
						_ = energyCost // Intentionally unused - energy tracking is internal
					}

					// Small delay to prevent CPU spinning
					time.Sleep(50 * time.Millisecond)
				}
			}
		}(agent)

		return true
	})

	// Monitor convergence (observation only, no control)
	go s.monitorConvergence(ctx)

	wg.Wait()
	return nil
}

// monitorConvergence observes emergent behavior without controlling it.
func (s *Swarm) monitorConvergence(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)
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
