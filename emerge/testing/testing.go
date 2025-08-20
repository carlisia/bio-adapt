package testing

import (
	"errors"
	"fmt"
	"time"

	"github.com/carlisia/bio-adapt/emerge/agent"
	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/goal"
	"github.com/carlisia/bio-adapt/emerge/strategy"
	"github.com/carlisia/bio-adapt/emerge/swarm"
	"github.com/carlisia/bio-adapt/internal/config"
)

var (
	// ErrInsufficientAgents indicates not enough agents for topology
	ErrInsufficientAgents = errors.New("insufficient agents")
)

// TestAgent creates an agent optimized for testing with predictable behavior.
// All randomization is disabled.
func TestAgent(id string) *agent.Agent {
	a, _ := agent.NewFromConfig(id, config.TestAgent())
	return a
}

// TestAgentWithMocks creates an agent with mock dependencies for unit testing.
func TestAgentWithMocks(id string, dm core.DecisionMaker, rm core.ResourceManager) *agent.Agent {
	return agent.New(id,
		agent.WithDecisionMaker(dm),
		agent.WithResourceManager(rm),
		agent.WithGoalManager(&goal.WeightedManager{}),
		agent.WithStrategy(&strategy.PhaseNudge{Rate: 0.5}),
	)
}

// TestSwarm creates a predictable swarm for testing.
// Uses a fully connected topology and deterministic agent behavior.
func TestSwarm(size int, g core.State) (*swarm.Swarm, error) {
	// Use small swarm config for fast convergence in tests
	swarmCfg := config.SmallSwarmConfig()
	swarmCfg.ConnectionProbability = 1.0 // Fully connected for predictability
	swarmCfg.Stubbornness = 0.01         // Very low for deterministic behavior

	return swarm.New(size, g,
		swarm.WithConfig(swarmCfg),
		swarm.WithAgentBuilder(func(id string, swarmSize int, sc config.Swarm) *agent.Agent {
			cfg := config.TestAgent()
			cfg.SwarmSize = swarmSize
			a, _ := agent.NewFromConfig(id, cfg)
			return a
		}),
		swarm.WithTopology(FullyConnectedTopology),
	)
}

// BenchmarkSwarm creates a swarm optimized for benchmarking.
// Uses minimal monitoring and fixed parameters.
func BenchmarkSwarm(size int) (*swarm.Swarm, error) {
	goal := core.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.8,
	}

	swarmCfg := config.DefaultConfig()

	return swarm.New(size, goal,
		swarm.WithConfig(swarmCfg),
		swarm.WithMonitor(nil), // Disable monitoring for benchmarks
	)
}

// FullyConnectedTopology creates a fully connected network topology.
// Every agent is connected to every other agent.
func FullyConnectedTopology(s *swarm.Swarm) error {
	agents := s.Agents()

	for _, a1 := range agents {
		for _, a2 := range agents {
			if a1.ID != a2.ID {
				a1.Neighbors().Store(a2.ID, a2)
			}
		}
	}

	return nil
}

// RingTopology creates a ring network topology.
// Each agent is connected to its immediate neighbors in a circle.
func RingTopology(s *swarm.Swarm) error {
	agentMap := s.Agents()

	// Convert map to slice for ordering
	var agents []*agent.Agent
	for _, a := range agentMap {
		agents = append(agents, a)
	}

	n := len(agents)
	if n < 2 {
		return fmt.Errorf("%w for ring topology: got %d, need at least 2", ErrInsufficientAgents, n)
	}

	for i, a := range agents {
		// Connect to previous neighbor
		prev := agents[(i-1+n)%n]
		a.Neighbors().Store(prev.ID, prev)

		// Connect to next neighbor
		next := agents[(i+1)%n]
		a.Neighbors().Store(next.ID, next)
	}

	return nil
}

// StarTopology creates a star network topology.
// One central agent is connected to all others.
func StarTopology(s *swarm.Swarm) error {
	agentMap := s.Agents()

	// Convert map to slice
	var agents []*agent.Agent
	for _, a := range agentMap {
		agents = append(agents, a)
	}

	if len(agents) < 2 {
		return fmt.Errorf("%w for star topology: got %d, need at least 2", ErrInsufficientAgents, len(agents))
	}

	hub := agents[0]

	for i, a := range agents {
		if i == 0 {
			// Hub connects to everyone
			for j, neighbor := range agents {
				if j != 0 {
					hub.Neighbors().Store(neighbor.ID, neighbor)
				}
			}
		} else {
			// Everyone else connects only to hub
			a.Neighbors().Store(hub.ID, hub)
		}
	}

	return nil
}

// MockDecisionMaker is a test double for DecisionMaker interface.
type MockDecisionMaker struct {
	DecisionFunc func(core.State, []core.Action) (core.Action, float64)
	CallCount    int
}

func (m *MockDecisionMaker) Decide(current core.State, options []core.Action) (core.Action, float64) {
	m.CallCount++
	if m.DecisionFunc != nil {
		return m.DecisionFunc(current, options)
	}
	// Default: always choose first option with high confidence
	if len(options) > 0 {
		return options[0], 1.0
	}
	return core.Action{Type: "maintain"}, 1.0
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
	ProposeFunc func(core.State, core.State, core.Context) (core.Action, float64)
	CallCount   int
}

func (m *MockSyncStrategy) Propose(current, target core.State, ctx core.Context) (core.Action, float64) {
	m.CallCount++
	if m.ProposeFunc != nil {
		return m.ProposeFunc(current, target, ctx)
	}
	// Default: simple phase adjustment
	return core.Action{
		Type:  "adjust_phase",
		Value: target.Phase - current.Phase,
		Cost:  1.0,
	}, 0.9
}

func (m *MockSyncStrategy) Name() string {
	return "mock"
}

// MockGoalManager is a test double for GoalManager interface.
type MockGoalManager struct {
	BlendFunc func(core.State, core.State, float64) core.State
}

func (m *MockGoalManager) Blend(local, global core.State, weight float64) core.State {
	if m.BlendFunc != nil {
		return m.BlendFunc(local, global, weight)
	}
	// Default: simple weighted average
	return core.State{
		Phase:     local.Phase*(1-weight) + global.Phase*weight,
		Frequency: local.Frequency,
		Coherence: local.Coherence*(1-weight) + global.Coherence*weight,
	}
}
