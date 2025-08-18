package biofield

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"
)

func TestNewSwarm(t *testing.T) {
	goal := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	swarm := NewSwarm(10, goal)

	if swarm.Size() != 10 {
		t.Errorf("Expected 10 agents, got %d", swarm.Size())
	}

	if swarm.goalState.Phase != goal.Phase {
		t.Error("Goal state not set correctly")
	}

	if swarm.monitor == nil {
		t.Error("Monitor not initialized")
	}

	// Check that agents have neighbors
	hasNeighbors := false
	swarm.agents.Range(func(key, value any) bool {
		agent := value.(*Agent)
		count := 0
		agent.neighbors.Range(func(k, v any) bool {
			count++
			return true
		})
		if count > 0 {
			hasNeighbors = true
			return false
		}
		return true
	})

	if !hasNeighbors {
		t.Error("Agents should have neighbors")
	}
}

func TestSwarmConnectToNeighbors(t *testing.T) {
	goal := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	swarm := NewSwarm(20, goal)

	// Count total connections
	totalConnections := 0
	swarm.agents.Range(func(key, value any) bool {
		agent := value.(*Agent)
		agent.neighbors.Range(func(k, v any) bool {
			totalConnections++
			return true
		})
		return true
	})

	// Each connection is counted twice (bidirectional)
	actualConnections := totalConnections / 2

	// With connection probability 0.3 and 5 attempts per agent,
	// we expect some but not all possible connections
	if actualConnections == 0 {
		t.Error("No connections established")
	}

	maxPossible := (20 * 19) / 2 // Complete graph
	if actualConnections >= maxPossible {
		t.Error("Too many connections (should be sparse)")
	}
}

func TestSwarmMeasureCoherence(t *testing.T) {
	goal := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	swarm := NewSwarm(10, goal)

	// Set all agents to same phase (perfect coherence)
	swarm.agents.Range(func(key, value any) bool {
		agent := value.(*Agent)
		agent.SetPhase(0)
		return true
	})

	coherence := swarm.MeasureCoherence()
	if math.Abs(coherence-1.0) > 0.01 {
		t.Errorf("Expected coherence 1.0 for aligned agents, got %f", coherence)
	}

	// Set agents to random phases (low coherence)
	swarm.agents.Range(func(key, value any) bool {
		agent := value.(*Agent)
		agent.SetPhase(rand.Float64() * 2 * math.Pi)
		return true
	})

	coherence = swarm.MeasureCoherence()
	if coherence > 0.5 {
		t.Errorf("Expected low coherence for random phases, got %f", coherence)
	}

	// Set half to 0, half to π (zero coherence)
	i := 0
	swarm.agents.Range(func(key, value any) bool {
		agent := value.(*Agent)
		if i < 5 {
			agent.SetPhase(0)
		} else {
			agent.SetPhase(math.Pi)
		}
		i++
		return true
	})

	coherence = swarm.MeasureCoherence()
	if math.Abs(coherence) > 0.1 {
		t.Errorf("Expected near-zero coherence for opposite phases, got %f", coherence)
	}
}

func TestSwarmGetAgent(t *testing.T) {
	goal := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	swarm := NewSwarm(5, goal)

	// Test getting existing agent
	agent, exists := swarm.GetAgent("agent-0")
	if !exists {
		t.Error("Should find agent-0")
	}
	if agent == nil {
		t.Error("Agent should not be nil")
	}
	if agent.ID != "agent-0" {
		t.Errorf("Wrong agent returned: %s", agent.ID)
	}

	// Test getting non-existent agent
	agent, exists = swarm.GetAgent("non-existent")
	if exists {
		t.Error("Should not find non-existent agent")
	}
	if agent != nil {
		t.Error("Non-existent agent should be nil")
	}
}

func TestSwarmDisruptAgents(t *testing.T) {
	goal := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	swarm := NewSwarm(10, goal)

	// Set all agents to same phase
	swarm.agents.Range(func(key, value any) bool {
		agent := value.(*Agent)
		agent.SetPhase(0)
		return true
	})

	initialCoherence := swarm.MeasureCoherence()

	// Disrupt 50% of agents
	swarm.DisruptAgents(0.5)

	finalCoherence := swarm.MeasureCoherence()

	if finalCoherence >= initialCoherence {
		t.Error("Coherence should decrease after disruption")
	}

	// Count disrupted agents (phase != 0)
	disrupted := 0
	swarm.agents.Range(func(key, value any) bool {
		agent := value.(*Agent)
		if math.Abs(agent.GetPhase()) > 0.01 {
			disrupted++
		}
		return true
	})

	// Should be approximately 5 agents (50% of 10)
	if disrupted < 3 || disrupted > 7 {
		t.Errorf("Expected ~5 disrupted agents, got %d", disrupted)
	}
}

func TestSwarmMonitor(t *testing.T) {
	goal := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	swarm := NewSwarm(10, goal)
	monitor := swarm.GetMonitor()

	if monitor == nil {
		t.Fatal("Monitor should not be nil")
	}

	// Record some samples
	monitor.RecordSample(0.5)
	monitor.RecordSample(0.6)
	monitor.RecordSample(0.7)

	latest := monitor.GetLatest()
	if latest != 0.7 {
		t.Errorf("Expected latest 0.7, got %f", latest)
	}

	avg := monitor.GetAverage()
	expectedAvg := (0.5 + 0.6 + 0.7) / 3
	if math.Abs(avg-expectedAvg) > 0.01 {
		t.Errorf("Expected average %f, got %f", expectedAvg, avg)
	}

	history := monitor.GetHistory()
	if len(history) != 3 {
		t.Errorf("Expected 3 samples in history, got %d", len(history))
	}
}

func TestSwarmConvergence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping convergence test in short mode")
	}

	goal := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.8,
	}

	swarm := NewSwarm(20, goal)

	// Random initial phases
	swarm.agents.Range(func(key, value any) bool {
		agent := value.(*Agent)
		agent.SetPhase(rand.Float64() * 2 * math.Pi)
		return true
	})

	initialCoherence := swarm.MeasureCoherence()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Run swarm
	go swarm.Run(ctx)

	// Wait for some convergence
	time.Sleep(1 * time.Second)

	finalCoherence := swarm.MeasureCoherence()

	if finalCoherence <= initialCoherence {
		t.Errorf("Expected coherence to improve from %f to > %f, got %f",
			initialCoherence, initialCoherence, finalCoherence)
	}

	// Check monitor recorded samples
	history := swarm.GetMonitor().GetHistory()
	if len(history) < 5 {
		t.Error("Monitor should have recorded samples during convergence")
	}
}

func BenchmarkSwarmMeasureCoherence(b *testing.B) {
	goal := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	sizes := []int{10, 50, 100, 500}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size-%d", size), func(b *testing.B) {
			swarm := NewSwarm(size, goal)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				swarm.MeasureCoherence()
			}
		})
	}
}

func BenchmarkSwarmCreation(b *testing.B) {
	goal := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	sizes := []int{10, 50, 100}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size-%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				NewSwarm(size, goal)
			}
		})
	}
}