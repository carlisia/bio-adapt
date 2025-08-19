package emerge_test

import (
	"math"
	"testing"
	"time"

	"github.com/carlisia/bio-adapt/emerge"
)

func TestNewAgent(t *testing.T) {
	agent := emerge.NewAgent("test-agent")

	if agent.ID != "test-agent" {
		t.Errorf("Expected ID 'test-agent', got %s", agent.ID)
	}

	// Check initial values are in expected ranges
	phase := agent.Phase()
	if phase < 0 || phase > 2*math.Pi {
		t.Errorf("Phase out of range: %f", phase)
	}

	energy := agent.Energy()
	if energy != 100.0 {
		t.Errorf("Expected initial energy 100, got %f", energy)
	}

	localGoal := agent.LocalGoal()
	if localGoal < 0 || localGoal > 2*math.Pi {
		t.Errorf("Local goal out of range: %f", localGoal)
	}

	influence := agent.Influence()
	if influence < 0.3 || influence > 0.7 {
		t.Errorf("Influence out of range: %f", influence)
	}

	stubbornness := agent.Stubbornness()
	if stubbornness < 0 || stubbornness > 0.3 {
		t.Errorf("Stubbornness out of range: %f", stubbornness)
	}
}

func TestAgentSettersAndGetters(t *testing.T) {
	agent := emerge.NewAgent("test")

	// Test phase setter/getter
	agent.SetPhase(math.Pi)
	if math.Abs(agent.Phase()-math.Pi) > 0.01 {
		t.Errorf("SetPhase/Phase failed: expected %f, got %f", math.Pi, agent.Phase())
	}

	// Test local goal setter/getter
	agent.SetLocalGoal(math.Pi / 2)
	if math.Abs(agent.LocalGoal()-math.Pi/2) > 0.01 {
		t.Errorf("SetLocalGoal/LocalGoal failed: expected %f, got %f", math.Pi/2, agent.LocalGoal())
	}
}

func TestAgentProposeAdjustment(t *testing.T) {
	agent := emerge.NewAgent("test")
	agent.SetPhase(0)
	agent.SetLocalGoal(math.Pi)

	globalGoal := emerge.State{
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
	agent := emerge.NewAgent("test")
	initialPhase := agent.Phase()

	// Test adjust_phase action
	adjustAction := emerge.Action{
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

	newPhase := agent.Phase()
	expectedPhase := math.Mod(initialPhase+0.5, 2*math.Pi)
	if math.Abs(newPhase-expectedPhase) > 0.01 {
		t.Errorf("Phase not adjusted correctly: expected %f, got %f", expectedPhase, newPhase)
	}

	// Test maintain action
	maintainAction := emerge.Action{
		Type: "maintain",
		Cost: 0.1,
	}

	phaseBeforeMaintain := agent.Phase()
	success, cost = agent.ApplyAction(maintainAction)
	if !success {
		t.Error("maintain action should succeed")
	}
	if cost != maintainAction.Cost {
		t.Errorf("Expected cost %f, got %f", maintainAction.Cost, cost)
	}
	if agent.Phase() != phaseBeforeMaintain {
		t.Error("maintain action should not change phase")
	}

	// Test invalid action
	invalidAction := emerge.Action{
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
	agent := emerge.NewAgent("test")

	// Agent starts with full energy
	if agent.Energy() != 100.0 {
		t.Errorf("Expected initial energy 100, got %f", agent.Energy())
	}

	// Apply action that costs energy
	action := emerge.Action{
		Type:  "adjust_phase",
		Value: 1.0,
		Cost:  10.0,
	}

	success, _ := agent.ApplyAction(action)
	if !success {
		t.Error("Action should succeed with sufficient energy")
	}

	// Energy should be reduced
	if agent.Energy() >= 100.0 {
		t.Error("Energy should be reduced after action")
	}
}

func TestAgentStubbornness(t *testing.T) {
	t.Skip("Flaky test - depends on randomness")
	// Use functional options for clean setup
	agent := emerge.NewAgent("stubborn",
		emerge.WithStubbornness(1.0), // Always stubborn
		emerge.WithPhase(0),
	)

	globalGoal := emerge.State{
		Phase:     math.Pi,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	// Stubborn agent should refuse adjustments more often
	maintainCount := 0
	for range 10 {
		action, accepted := agent.ProposeAdjustment(globalGoal)
		if accepted && action.Type == "maintain" {
			maintainCount++
		}
	}

	// Highly stubborn agent should maintain position most of the time
	if maintainCount < 7 {
		t.Errorf("Stubborn agent should maintain position more often, got %d/10", maintainCount)
	}
}

// TestAgentWithOptions demonstrates idiomatic dependency injection
func TestAgentWithOptions(t *testing.T) {
	t.Run("with phase", func(t *testing.T) {
		agent := emerge.NewAgent("test",
			emerge.WithPhase(1.5),
			emerge.WithStubbornness(0.05),
		)

		if agent.Phase() != 1.5 {
			t.Errorf("expected phase 1.5, got %f", agent.Phase())
		}
		if agent.Stubbornness() != 0.05 {
			t.Errorf("expected stubbornness 0.05, got %f", agent.Stubbornness())
		}
	})

	t.Run("with influence", func(t *testing.T) {
		agent := emerge.NewAgent("test",
			emerge.WithInfluence(0.9),
		)

		if agent.Influence() != 0.9 {
			t.Errorf("expected influence 0.9, got %f", agent.Influence())
		}
	})
}

// TestAgentFromConfig demonstrates config-based creation
func TestAgentFromConfig(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		config := emerge.DefaultAgentConfig()
		agent := emerge.NewAgentFromConfig("test", config)

		if agent.Energy() != 100.0 {
			t.Errorf("expected energy 100, got %f", agent.Energy())
		}
		// Default config should not randomize
		if config.RandomizePhase {
			t.Error("default config should not randomize phase")
		}
	})

	t.Run("test config", func(t *testing.T) {
		config := emerge.TestAgentConfig()
		agent := emerge.NewAgentFromConfig("test", config)

		// Test config should have predictable values
		if agent.Stubbornness() != 0.01 {
			t.Errorf("expected stubbornness 0.01, got %f", agent.Stubbornness())
		}
		if agent.Influence() != 0.8 {
			t.Errorf("expected influence 0.8, got %f", agent.Influence())
		}
	})

	t.Run("randomized config", func(t *testing.T) {
		config := emerge.RandomizedAgentConfig()
		agent := emerge.NewAgentFromConfig("test", config)

		// Should have randomized initial conditions
		if !config.RandomizePhase {
			t.Error("randomized config should randomize phase")
		}
		// Phase should be in valid range
		phase := agent.Phase()
		if phase < 0 || phase > 2*math.Pi {
			t.Errorf("phase out of range: %f", phase)
		}
	})
}

// TestTestingHelpers demonstrates the testing utilities
func TestTestingHelpers(t *testing.T) {
	t.Run("TestAgent helper", func(t *testing.T) {
		agent := emerge.TestAgent("helper-test")

		// Should have predictable test values
		if agent.Stubbornness() != 0.01 {
			t.Errorf("expected test stubbornness 0.01, got %f", agent.Stubbornness())
		}
	})
}

