package emerge

import (
	"fmt"
	"time"
)

// TestAgent creates an agent optimized for testing with predictable behavior.
// All randomization is disabled.
func TestAgent(id string) *Agent {
	return NewAgentFromConfig(id, TestAgentConfig())
}

// TestAgentWithMocks creates an agent with mock dependencies for unit testing.
func TestAgentWithMocks(id string, dm DecisionMaker, rm ResourceManager) *Agent {
	config := TestAgentConfig()
	config.DecisionMaker = dm
	config.ResourceManager = rm
	return NewAgentFromConfig(id, config)
}

// TestSwarm creates a predictable swarm for testing.
// Uses a fully connected topology and deterministic agent behavior.
func TestSwarm(size int, goal State) (*Swarm, error) {
	// Use small swarm config for fast convergence in tests
	config := SmallSwarmConfig()
	config.ConnectionProbability = 1.0 // Fully connected for predictability
	config.Stubbornness = 0.01         // Very low for deterministic behavior

	return NewSwarm(size, goal,
		WithConfig(config),
		WithAgentBuilder(func(id string, swarmSize int, config SwarmConfig) *Agent {
			cfg := TestAgentConfig()
			cfg.SwarmSize = swarmSize
			return NewAgentFromConfig(id, cfg)
		}),
		WithTopology(FullyConnectedTopology),
	)
}

// BenchmarkSwarm creates a swarm optimized for benchmarking.
// Uses minimal monitoring and fixed parameters.
func BenchmarkSwarm(size int) (*Swarm, error) {
	goal := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.8,
	}

	config := DefaultConfig()

	return NewSwarm(size, goal,
		WithConfig(config),
		WithMonitor(nil), // Disable monitoring for benchmarks
		WithConvergenceMonitor(nil),
	)
}

// FullyConnectedTopology creates a fully connected network topology.
// Every agent is connected to every other agent.
func FullyConnectedTopology(s *Swarm) error {
	var agents []*Agent
	s.agents.Range(func(key, value any) bool {
		agents = append(agents, value.(*Agent))
		return true
	})

	for i, agent := range agents {
		for j, neighbor := range agents {
			if i != j {
				agent.neighbors.Store(neighbor.ID, neighbor)
			}
		}
	}

	return nil
}

// RingTopology creates a ring network topology.
// Each agent is connected to its immediate neighbors in a circle.
func RingTopology(s *Swarm) error {
	var agents []*Agent
	s.agents.Range(func(key, value any) bool {
		agents = append(agents, value.(*Agent))
		return true
	})

	n := len(agents)
	if n < 2 {
		return fmt.Errorf("ring topology requires at least 2 agents, got %d", n)
	}

	for i, agent := range agents {
		// Connect to previous neighbor
		prev := agents[(i-1+n)%n]
		agent.neighbors.Store(prev.ID, prev)

		// Connect to next neighbor
		next := agents[(i+1)%n]
		agent.neighbors.Store(next.ID, next)
	}

	return nil
}

// StarTopology creates a star network topology.
// One central agent is connected to all others.
func StarTopology(s *Swarm) error {
	var agents []*Agent
	s.agents.Range(func(key, value any) bool {
		agents = append(agents, value.(*Agent))
		return true
	})

	if len(agents) < 2 {
		return fmt.Errorf("star topology requires at least 2 agents, got %d", len(agents))
	}

	hub := agents[0]

	for i, agent := range agents {
		if i == 0 {
			// Hub connects to everyone
			for j, neighbor := range agents {
				if j != 0 {
					hub.neighbors.Store(neighbor.ID, neighbor)
				}
			}
		} else {
			// Everyone else connects only to hub
			agent.neighbors.Store(hub.ID, hub)
		}
	}

	return nil
}

// MockDecisionMaker is a test double for DecisionMaker interface.
type MockDecisionMaker struct {
	DecisionFunc func(State, []Action) (Action, float64)
	CallCount    int
}

func (m *MockDecisionMaker) Decide(current State, options []Action) (Action, float64) {
	m.CallCount++
	if m.DecisionFunc != nil {
		return m.DecisionFunc(current, options)
	}
	// Default: always choose first option with high confidence
	if len(options) > 0 {
		return options[0], 1.0
	}
	return Action{Type: "maintain"}, 1.0
}

// MockResourceManager is a test double for ResourceManager interface.
type MockResourceManager struct {
	AvailableFunc func() float64
	RequestFunc   func(float64) float64
	ReleaseFunc   func(float64)
	Energy        float64
}

func (m *MockResourceManager) Available() float64 {
	if m.AvailableFunc != nil {
		return m.AvailableFunc()
	}
	return m.Energy
}

func (m *MockResourceManager) Request(amount float64) float64 {
	if m.RequestFunc != nil {
		return m.RequestFunc(amount)
	}
	if m.Energy >= amount {
		m.Energy -= amount
		return amount
	}
	return 0
}

func (m *MockResourceManager) Release(amount float64) {
	if m.ReleaseFunc != nil {
		m.ReleaseFunc(amount)
		return
	}
	m.Energy += amount
}

// MockSyncStrategy is a test double for SyncStrategy interface.
type MockSyncStrategy struct {
	ProposeFunc func(State, State, Context) (Action, float64)
	CallCount   int
}

func (m *MockSyncStrategy) Propose(current, target State, ctx Context) (Action, float64) {
	m.CallCount++
	if m.ProposeFunc != nil {
		return m.ProposeFunc(current, target, ctx)
	}
	// Default: simple phase adjustment
	return Action{
		Type:  "adjust_phase",
		Value: target.Phase - current.Phase,
		Cost:  1.0,
	}, 0.9
}

// MockGoalManager is a test double for GoalManager interface.
type MockGoalManager struct {
	BlendFunc func(State, State, float64) State
}

func (m *MockGoalManager) Blend(local, global State, weight float64) State {
	if m.BlendFunc != nil {
		return m.BlendFunc(local, global, weight)
	}
	// Default: simple weighted average
	return State{
		Phase:     local.Phase*(1-weight) + global.Phase*weight,
		Frequency: local.Frequency,
		Coherence: local.Coherence*(1-weight) + global.Coherence*weight,
	}
}

