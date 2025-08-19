package attractor

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

	// Monitoring (read-only, doesn't control)
	monitor     *Monitor
	basin       *AttractorBasin
	convergence *ConvergenceMonitor
}

// NewSwarm creates a distributed swarm with autonomous agents.
func NewSwarm(size int, goal State) (*Swarm, error) {
	if size <= 0 {
		return nil, fmt.Errorf("swarm size must be positive, got %d", size)
	}
	if goal.Frequency <= 0 {
		return nil, fmt.Errorf("goal frequency must be positive, got %v", goal.Frequency)
	}
	if goal.Coherence < 0 || goal.Coherence > 1 {
		return nil, fmt.Errorf("goal coherence must be in [0, 1], got %f", goal.Coherence)
	}

	s := &Swarm{
		goalState:   goal,
		monitor:     NewMonitor(),
		basin:       NewAttractorBasin(goal, 0.8, math.Pi),
		convergence: NewConvergenceMonitor(goal, goal.Coherence),
	}

	// Create autonomous agents
	for i := range size {
		agent := NewAgent(fmt.Sprintf("agent-%d", i))
		s.agents.Store(agent.ID, agent)

		// Establish random neighbor connections (small-world network)
		s.connectToNeighbors(agent, 5)
	}

	return s, nil
}

// connectToNeighbors creates local connections for emergent behavior.
func (s *Swarm) connectToNeighbors(agent *Agent, count int) {
	connected := 0

	s.agents.Range(func(key, value any) bool {
		if connected >= count {
			return false
		}

		neighbor := value.(*Agent)
		if neighbor.ID != agent.ID && rand.Float64() < 0.3 {
			agent.neighbors.Store(neighbor.ID, neighbor)
			neighbor.neighbors.Store(agent.ID, agent)
			connected++
		}

		return true
	})
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
		phases = append(phases, agent.GetPhase())
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
	count := 0
	s.agents.Range(func(key, value any) bool {
		count++
		return true
	})
	return count
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
