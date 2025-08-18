package biofield

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
	monitor *Monitor
}

// NewSwarm creates a distributed swarm with autonomous agents.
func NewSwarm(size int, goal State) *Swarm {
	s := &Swarm{
		goalState: goal,
		monitor:   NewMonitor(),
	}

	// Create autonomous agents
	for i := range size {
		agent := NewAgent(fmt.Sprintf("agent-%d", i))
		s.agents.Store(agent.ID, agent)

		// Establish random neighbor connections (small-world network)
		s.connectToNeighbors(agent, 5)
	}

	return s
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
						a.ApplyAction(action)
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

			if coherence > s.goalState.Coherence {
				fmt.Printf("Achieved target coherence: %.3f\n", coherence)
			}
		}
	}
}

// MeasureCoherence calculates global synchronization level.
// This is for monitoring only - agents don't have access to this.
func (s *Swarm) MeasureCoherence() float64 {
	var sumCos, sumSin float64
	var count int

	s.agents.Range(func(key, value any) bool {
		agent := value.(*Agent)
		phase := agent.GetPhase()
		sumCos += math.Cos(phase)
		sumSin += math.Sin(phase)
		count++
		return true
	})

	if count == 0 {
		return 0
	}

	return math.Sqrt(sumCos*sumCos+sumSin*sumSin) / float64(count)
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
