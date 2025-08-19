package attractor

import (
	"math"
	"testing"
	"time"
)

func TestNewAgent(t *testing.T) {
	agent := NewAgent("test-agent")

	if agent.ID != "test-agent" {
		t.Errorf("Expected ID 'test-agent', got %s", agent.ID)
	}

	// Check initial values are in expected ranges
	phase := agent.GetPhase()
	if phase < 0 || phase > 2*math.Pi {
		t.Errorf("Phase out of range: %f", phase)
	}

	energy := agent.GetEnergy()
	if energy != 100.0 {
		t.Errorf("Expected initial energy 100, got %f", energy)
	}

	localGoal := agent.LocalGoal.Load()
	if localGoal < 0 || localGoal > 2*math.Pi {
		t.Errorf("Local goal out of range: %f", localGoal)
	}

	influence := agent.influence.Load()
	if influence < 0.3 || influence > 0.7 {
		t.Errorf("Influence out of range: %f", influence)
	}

	stubbornness := agent.stubbornness.Load()
	if stubbornness < 0 || stubbornness > 0.3 {
		t.Errorf("Stubbornness out of range: %f", stubbornness)
	}
}

func TestAgentUpdateContext(t *testing.T) {
	agent1 := NewAgent("agent1")
	agent2 := NewAgent("agent2")
	agent3 := NewAgent("agent3")

	// Connect agents as neighbors
	agent1.neighbors.Store(agent2.ID, agent2)
	agent1.neighbors.Store(agent3.ID, agent3)

	// Set known phases
	agent2.SetPhase(0)
	agent3.SetPhase(0)
	agent1.SetPhase(0)

	agent1.UpdateContext()
	ctx := agent1.context.Load().(Context)

	// With 2 neighbors out of assumed 20 max
	expectedDensity := 2.0 / 20.0
	if math.Abs(ctx.Density-expectedDensity) > 0.01 {
		t.Errorf("Expected density %f, got %f", expectedDensity, ctx.Density)
	}

	// All phases are the same, so stability should be high
	if ctx.Stability < 0.9 {
		t.Errorf("Expected high stability, got %f", ctx.Stability)
	}

	// All phases aligned, coherence should be 1
	if math.Abs(ctx.LocalCoherence-1.0) > 0.01 {
		t.Errorf("Expected coherence 1.0, got %f", ctx.LocalCoherence)
	}
}

func TestAgentCalculateLocalCoherence(t *testing.T) {
	agent1 := NewAgent("agent1")
	agent2 := NewAgent("agent2")
	agent3 := NewAgent("agent3")

	// Test case 1: All aligned (coherence = 1)
	agent1.SetPhase(0)
	agent2.SetPhase(0)
	agent3.SetPhase(0)

	agent1.neighbors.Store(agent2.ID, agent2)
	agent1.neighbors.Store(agent3.ID, agent3)

	coherence := agent1.calculateLocalCoherence()
	if math.Abs(coherence-1.0) > 0.01 {
		t.Errorf("Expected coherence 1.0 for aligned agents, got %f", coherence)
	}

	// Test case 2: Opposite phases (coherence = 0)
	agent2.SetPhase(0)
	agent3.SetPhase(math.Pi)

	coherence = agent1.calculateLocalCoherence()
	if math.Abs(coherence) > 0.01 {
		t.Errorf("Expected coherence 0.0 for opposite phases, got %f", coherence)
	}

	// Test case 3: No neighbors
	agent4 := NewAgent("agent4")
	coherence = agent4.calculateLocalCoherence()
	if coherence != 0 {
		t.Errorf("Expected coherence 0 for no neighbors, got %f", coherence)
	}
}

func TestAgentProposeAdjustment(t *testing.T) {
	agent := NewAgent("test")
	agent.SetPhase(0)
	agent.LocalGoal.Store(math.Pi)

	globalGoal := State{
		Phase:     math.Pi / 2,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	// Agent should propose some adjustment
	action, accepted := agent.ProposeAdjustment(globalGoal)

	// Not all proposals are accepted due to autonomy
	if accepted {
		if action.Type == "" {
			t.Error("Accepted action should have a type")
		}
		if action.Cost < 0 {
			t.Error("Action cost should be non-negative")
		}
	}
}

func TestAgentApplyAction(t *testing.T) {
	agent := NewAgent("test")
	initialPhase := agent.GetPhase()

	// Test adjust_phase action
	adjustAction := Action{
		Type:  "adjust_phase",
		Value: 0.5,
		Cost:  1.0,
	}

	success, cost := agent.ApplyAction(adjustAction)
	if !success {
		t.Error("adjust_phase action should succeed")
	}
	if cost != adjustAction.Cost {
		t.Errorf("Expected cost %f, got %f", adjustAction.Cost, cost)
	}

	newPhase := agent.GetPhase()
	expectedPhase := math.Mod(initialPhase+0.5, 2*math.Pi)
	if math.Abs(newPhase-expectedPhase) > 0.01 {
		t.Errorf("Phase not adjusted correctly: expected %f, got %f", expectedPhase, newPhase)
	}

	// Test maintain action
	maintainAction := Action{
		Type: "maintain",
		Cost: 0.1,
	}

	phaseBeforeMaintain := agent.GetPhase()
	success, cost = agent.ApplyAction(maintainAction)
	if !success {
		t.Error("maintain action should succeed")
	}
	if cost != maintainAction.Cost {
		t.Errorf("Expected cost %f, got %f", maintainAction.Cost, cost)
	}
	if agent.GetPhase() != phaseBeforeMaintain {
		t.Error("maintain action should not change phase")
	}

	// Test invalid action
	invalidAction := Action{
		Type: "invalid",
	}
	success, cost = agent.ApplyAction(invalidAction)
	if success {
		t.Error("invalid action should fail")
	}
	if cost != 0 {
		t.Error("failed action should have zero cost")
	}
}

func TestAgentEnergyManagement(t *testing.T) {
	agent := NewAgent("test")

	// Agent starts with full energy
	if agent.GetEnergy() != 100.0 {
		t.Errorf("Expected initial energy 100, got %f", agent.GetEnergy())
	}

	// Create expensive action
	expensiveAction := Action{
		Type:    "adjust_phase",
		Value:   1.0,
		Cost:    150.0, // More than available
		Benefit: 2.0,
	}

	// Set up agent for testing
	agent.stubbornness.Store(0) // Not stubborn
	globalGoal := State{
		Phase:     math.Pi,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	// Force expensive action consideration
	agent.decider = &testDecisionMaker{chosenAction: expensiveAction}

	// Should reject due to insufficient energy
	_, accepted := agent.ProposeAdjustment(globalGoal)
	if accepted {
		t.Error("Should reject action due to insufficient energy")
	}
}

// testDecisionMaker for testing specific decisions
type testDecisionMaker struct {
	chosenAction Action
}

func (t *testDecisionMaker) Decide(state State, options []Action) (Action, float64) {
	return t.chosenAction, 1.0
}

func TestAgentStubbornness(t *testing.T) {
	// Use functional options for clean setup
	agent := NewAgent("stubborn",
		WithStubbornness(1.0), // Always stubborn
		WithPhase(0),
	)

	globalGoal := State{
		Phase:     math.Pi,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	// Stubborn agent should refuse adjustments
	for range 10 {
		action, accepted := agent.ProposeAdjustment(globalGoal)
		if accepted && action.Type != "maintain" {
			t.Error("Stubborn agent should not accept adjustments")
		}
	}
}

// TestAgentWithOptions demonstrates idiomatic dependency injection
func TestAgentWithOptions(t *testing.T) {
	t.Run("custom decision maker", func(t *testing.T) {
		mockDM := &MockDecisionMaker{
			DecisionFunc: func(s State, opts []Action) (Action, float64) {
				return Action{Type: "custom"}, 1.0
			},
		}

		agent := NewAgent("test",
			WithDecisionMaker(mockDM),
			WithPhase(1.5),
			WithStubbornness(0.05),
		)

		if agent.decider != mockDM {
			t.Error("expected mock decision maker")
		}
		if agent.GetPhase() != 1.5 {
			t.Errorf("expected phase 1.5, got %f", agent.GetPhase())
		}
	})

	t.Run("custom resource manager", func(t *testing.T) {
		mockRM := &MockResourceManager{Energy: 200}

		agent := NewAgent("test",
			WithResourceManager(mockRM),
		)

		if agent.GetEnergy() != 200 {
			t.Errorf("expected energy 200, got %f", agent.GetEnergy())
		}
	})
}

// TestAgentFromConfig demonstrates config-based creation
func TestAgentFromConfig(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		config := DefaultAgentConfig()
		agent := NewAgentFromConfig("test", config)

		if agent.GetEnergy() != 100.0 {
			t.Errorf("expected energy 100, got %f", agent.GetEnergy())
		}
		// Default config should not randomize
		if config.RandomizePhase {
			t.Error("default config should not randomize phase")
		}
	})

	t.Run("test config", func(t *testing.T) {
		config := TestAgentConfig()
		agent := NewAgentFromConfig("test", config)

		// Test config should have predictable values
		if agent.GetStubbornness() != 0.01 {
			t.Errorf("expected stubbornness 0.01, got %f", agent.GetStubbornness())
		}
		if agent.GetInfluence() != 0.8 {
			t.Errorf("expected influence 0.8, got %f", agent.GetInfluence())
		}
	})

	t.Run("randomized config", func(t *testing.T) {
		config := RandomizedAgentConfig()
		agent := NewAgentFromConfig("test", config)

		// Should have randomized initial conditions
		if !config.RandomizePhase {
			t.Error("randomized config should randomize phase")
		}
		// Phase should be in valid range
		phase := agent.GetPhase()
		if phase < 0 || phase > 2*math.Pi {
			t.Errorf("phase out of range: %f", phase)
		}
	})
}

// TestTestingHelpers demonstrates the testing utilities
func TestTestingHelpers(t *testing.T) {
	t.Run("TestAgent helper", func(t *testing.T) {
		agent := TestAgent("helper-test")

		// Should have predictable test values
		if agent.GetStubbornness() != 0.01 {
			t.Errorf("expected test stubbornness 0.01, got %f", agent.GetStubbornness())
		}
	})

	t.Run("TestAgentWithMocks helper", func(t *testing.T) {
		mockDM := &MockDecisionMaker{}
		mockRM := &MockResourceManager{Energy: 150}

		agent := TestAgentWithMocks("mock-test", mockDM, mockRM)

		if agent.decider != mockDM {
			t.Error("expected mock decision maker")
		}
		if agent.GetEnergy() != 150 {
			t.Errorf("expected energy 150, got %f", agent.GetEnergy())
		}
	})
}
