package emerge_test

import (
	"math"
	"testing"
	"time"

	"github.com/carlisia/bio-adapt/emerge"
)

func TestNewAgent(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		validateFn func(t *testing.T, agent *emerge.Agent)
	}{
		{
			name: "basic agent creation",
			id:   "test-agent",
			validateFn: func(t *testing.T, agent *emerge.Agent) {
				if agent.ID != "test-agent" {
					t.Errorf("ID = %s, want test-agent", agent.ID)
				}
			},
		},
		{
			name: "empty ID",
			id:   "",
			validateFn: func(t *testing.T, agent *emerge.Agent) {
				if agent.ID != "" {
					t.Errorf("ID = %s, want empty string", agent.ID)
				}
			},
		},
		{
			name: "special characters in ID",
			id:   "agent-123!@#$%",
			validateFn: func(t *testing.T, agent *emerge.Agent) {
				if agent.ID != "agent-123!@#$%" {
					t.Errorf("ID = %s, want agent-123!@#$%%", agent.ID)
				}
			},
		},
		{
			name: "very long ID",
			id:   "this-is-a-very-long-agent-id-that-exceeds-normal-length-expectations-and-should-still-work-properly",
			validateFn: func(t *testing.T, agent *emerge.Agent) {
				if agent.ID != "this-is-a-very-long-agent-id-that-exceeds-normal-length-expectations-and-should-still-work-properly" {
					t.Error("Long ID not preserved")
				}
			},
		},
		{
			name: "initial values in range",
			id:   "range-test",
			validateFn: func(t *testing.T, agent *emerge.Agent) {
				phase := agent.Phase()
				if phase < 0 || phase > 2*math.Pi {
					t.Errorf("Phase = %f, want in [0, 2π]", phase)
				}

				energy := agent.Energy()
				if energy != 100.0 {
					t.Errorf("Energy = %f, want 100.0", energy)
				}

				localGoal := agent.LocalGoal()
				if localGoal < 0 || localGoal > 2*math.Pi {
					t.Errorf("LocalGoal = %f, want in [0, 2π]", localGoal)
				}

				influence := agent.Influence()
				if influence < 0.3 || influence > 0.7 {
					t.Errorf("Influence = %f, want in [0.3, 0.7]", influence)
				}

				stubbornness := agent.Stubbornness()
				if stubbornness < 0 || stubbornness > 0.3 {
					t.Errorf("Stubbornness = %f, want in [0, 0.3]", stubbornness)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := emerge.NewAgent(tt.id)
			if agent == nil {
				t.Fatal("NewAgent returned nil")
			}
			tt.validateFn(t, agent)
		})
	}
}

func TestAgentSettersAndGetters(t *testing.T) {
	tests := []struct {
		name    string
		setupFn func(agent *emerge.Agent)
		checkFn func(t *testing.T, agent *emerge.Agent)
	}{
		{
			name: "set phase to π",
			setupFn: func(agent *emerge.Agent) {
				agent.SetPhase(math.Pi)
			},
			checkFn: func(t *testing.T, agent *emerge.Agent) {
				if math.Abs(agent.Phase()-math.Pi) > 0.01 {
					t.Errorf("Phase = %f, want %f", agent.Phase(), math.Pi)
				}
			},
		},
		{
			name: "set phase to 0",
			setupFn: func(agent *emerge.Agent) {
				agent.SetPhase(0)
			},
			checkFn: func(t *testing.T, agent *emerge.Agent) {
				if agent.Phase() != 0 {
					t.Errorf("Phase = %f, want 0", agent.Phase())
				}
			},
		},
		{
			name: "set phase to 2π",
			setupFn: func(agent *emerge.Agent) {
				agent.SetPhase(2 * math.Pi)
			},
			checkFn: func(t *testing.T, agent *emerge.Agent) {
				// Should wrap to 0
				if math.Abs(agent.Phase()) > 0.01 && math.Abs(agent.Phase()-2*math.Pi) > 0.01 {
					t.Errorf("Phase = %f, want 0 or 2π", agent.Phase())
				}
			},
		},
		{
			name: "set phase to negative (wrapping)",
			setupFn: func(agent *emerge.Agent) {
				agent.SetPhase(-math.Pi)
			},
			checkFn: func(t *testing.T, agent *emerge.Agent) {
				// Should wrap to π
				expected := math.Pi
				if math.Abs(agent.Phase()-expected) > 0.01 {
					t.Errorf("Phase = %f, want %f (wrapped)", agent.Phase(), expected)
				}
			},
		},
		{
			name: "set local goal to π/2",
			setupFn: func(agent *emerge.Agent) {
				agent.SetLocalGoal(math.Pi / 2)
			},
			checkFn: func(t *testing.T, agent *emerge.Agent) {
				if math.Abs(agent.LocalGoal()-math.Pi/2) > 0.01 {
					t.Errorf("LocalGoal = %f, want %f", agent.LocalGoal(), math.Pi/2)
				}
			},
		},
		{
			name: "set influence",
			setupFn: func(agent *emerge.Agent) {
				agent.SetInfluence(0.75)
			},
			checkFn: func(t *testing.T, agent *emerge.Agent) {
				if math.Abs(agent.Influence()-0.75) > 0.01 {
					t.Errorf("Influence = %f, want 0.75", agent.Influence())
				}
			},
		},
		{
			name: "set influence out of range (clamping)",
			setupFn: func(agent *emerge.Agent) {
				agent.SetInfluence(1.5)
			},
			checkFn: func(t *testing.T, agent *emerge.Agent) {
				if agent.Influence() > 1.0 {
					t.Errorf("Influence = %f, should be clamped to <= 1.0", agent.Influence())
				}
			},
		},
		{
			name: "set stubbornness",
			setupFn: func(agent *emerge.Agent) {
				agent.SetStubbornness(0.2)
			},
			checkFn: func(t *testing.T, agent *emerge.Agent) {
				if math.Abs(agent.Stubbornness()-0.2) > 0.01 {
					t.Errorf("Stubbornness = %f, want 0.2", agent.Stubbornness())
				}
			},
		},
		{
			name: "set negative stubbornness (clamping)",
			setupFn: func(agent *emerge.Agent) {
				agent.SetStubbornness(-0.5)
			},
			checkFn: func(t *testing.T, agent *emerge.Agent) {
				if agent.Stubbornness() < 0 {
					t.Errorf("Stubbornness = %f, should be clamped to >= 0", agent.Stubbornness())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := emerge.NewAgent("test")
			tt.setupFn(agent)
			tt.checkFn(t, agent)
		})
	}
}

func TestAgentProposeAdjustment(t *testing.T) {
	tests := []struct {
		name       string
		setupFn    func() *emerge.Agent
		globalGoal emerge.State
		validateFn func(t *testing.T, action emerge.Action, accepted bool)
	}{
		{
			name: "agent far from goal",
			setupFn: func() *emerge.Agent {
				agent := emerge.NewAgent("test")
				agent.SetPhase(0)
				agent.SetLocalGoal(math.Pi)
				return agent
			},
			globalGoal: emerge.State{
				Phase:     math.Pi / 2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			validateFn: func(t *testing.T, action emerge.Action, accepted bool) {
				if accepted {
					if action.Type == "" {
						t.Error("Accepted action should have a type")
					}
					if action.Cost < 0 {
						t.Error("Action cost should be non-negative")
					}
				}
			},
		},
		{
			name: "agent at goal",
			setupFn: func() *emerge.Agent {
				agent := emerge.NewAgent("test")
				agent.SetPhase(math.Pi)
				agent.SetLocalGoal(math.Pi)
				return agent
			},
			globalGoal: emerge.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			validateFn: func(t *testing.T, action emerge.Action, accepted bool) {
				if accepted && action.Type == "adjust_phase" {
					// Should maintain position more likely
					if math.Abs(action.Value) > 0.1 {
						t.Error("Agent at goal should propose small adjustments")
					}
				}
			},
		},
		{
			name: "stubborn agent",
			setupFn: func() *emerge.Agent {
				return emerge.NewAgent("stubborn",
					emerge.WithStubbornness(0.99),
					emerge.WithPhase(0),
				)
			},
			globalGoal: emerge.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			validateFn: func(t *testing.T, action emerge.Action, accepted bool) {
				// Stubborn agents accept fewer proposals
				// This is probabilistic, so we just check valid structure
				if accepted {
					if action.Type == "" {
						t.Error("Action should have a type")
					}
				}
			},
		},
		{
			name: "low energy agent",
			setupFn: func() *emerge.Agent {
				agent := emerge.NewAgent("test")
				// Deplete energy
				for range 20 {
					_, _, _ = agent.ApplyAction(emerge.Action{
						Type:  "adjust_phase",
						Value: 0.1,
						Cost:  5.0,
					})
				}
				return agent
			},
			globalGoal: emerge.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			validateFn: func(t *testing.T, action emerge.Action, accepted bool) {
				if accepted {
					// Low energy should prefer low-cost actions
					if action.Cost > 5.0 {
						t.Error("Low energy agent should prefer low-cost actions")
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := tt.setupFn()
			action, accepted := agent.ProposeAdjustment(tt.globalGoal)
			tt.validateFn(t, action, accepted)
		})
	}
}

func TestAgentApplyAction(t *testing.T) {
	tests := []struct {
		name        string
		setupFn     func() *emerge.Agent
		action      emerge.Action
		wantSuccess bool
		wantCost    float64
		wantErr     bool
		validateFn  func(t *testing.T, agent *emerge.Agent, beforePhase float64)
	}{
		{
			name: "adjust phase positive",
			setupFn: func() *emerge.Agent {
				agent := emerge.NewAgent("test")
				agent.SetPhase(0)
				return agent
			},
			action: emerge.Action{
				Type:  "adjust_phase",
				Value: 0.5,
				Cost:  1.0,
			},
			wantSuccess: true,
			wantCost:    1.0,
			wantErr:     false,
			validateFn: func(t *testing.T, agent *emerge.Agent, beforePhase float64) {
				expectedPhase := math.Mod(beforePhase+0.5, 2*math.Pi)
				if math.Abs(agent.Phase()-expectedPhase) > 0.01 {
					t.Errorf("Phase = %f, want %f", agent.Phase(), expectedPhase)
				}
			},
		},
		{
			name: "adjust phase negative",
			setupFn: func() *emerge.Agent {
				agent := emerge.NewAgent("test")
				agent.SetPhase(math.Pi)
				return agent
			},
			action: emerge.Action{
				Type:  "adjust_phase",
				Value: -0.5,
				Cost:  1.0,
			},
			wantSuccess: true,
			wantCost:    1.0,
			wantErr:     false,
			validateFn: func(t *testing.T, agent *emerge.Agent, beforePhase float64) {
				expectedPhase := beforePhase - 0.5
				if expectedPhase < 0 {
					expectedPhase += 2 * math.Pi
				}
				if math.Abs(agent.Phase()-expectedPhase) > 0.01 {
					t.Errorf("Phase = %f, want %f", agent.Phase(), expectedPhase)
				}
			},
		},
		{
			name: "adjust phase large value (wrapping)",
			setupFn: func() *emerge.Agent {
				agent := emerge.NewAgent("test")
				agent.SetPhase(math.Pi)
				return agent
			},
			action: emerge.Action{
				Type:  "adjust_phase",
				Value: 3 * math.Pi,
				Cost:  2.0,
			},
			wantSuccess: true,
			wantCost:    2.0,
			validateFn: func(t *testing.T, agent *emerge.Agent, beforePhase float64) {
				// Should wrap around
				expectedPhase := math.Mod(beforePhase+3*math.Pi, 2*math.Pi)
				if math.Abs(agent.Phase()-expectedPhase) > 0.01 {
					t.Errorf("Phase = %f, want %f", agent.Phase(), expectedPhase)
				}
			},
		},
		{
			name: "maintain action",
			setupFn: func() *emerge.Agent {
				agent := emerge.NewAgent("test")
				agent.SetPhase(1.5)
				return agent
			},
			action: emerge.Action{
				Type: "maintain",
				Cost: 0.1,
			},
			wantSuccess: true,
			wantCost:    0.1,
			wantErr:     false,
			validateFn: func(t *testing.T, agent *emerge.Agent, beforePhase float64) {
				if agent.Phase() != beforePhase {
					t.Errorf("Phase = %f, want %f (unchanged)", agent.Phase(), beforePhase)
				}
			},
		},
		{
			name: "invalid action type",
			setupFn: func() *emerge.Agent {
				return emerge.NewAgent("test")
			},
			action: emerge.Action{
				Type: "invalid_action",
				Cost: 1.0,
			},
			wantSuccess: false,
			wantCost:    0,
			wantErr:     true,
			validateFn: func(t *testing.T, agent *emerge.Agent, beforePhase float64) {
				if agent.Phase() != beforePhase {
					t.Error("Invalid action should not change phase")
				}
			},
		},
		{
			name: "empty action type",
			setupFn: func() *emerge.Agent {
				return emerge.NewAgent("test")
			},
			action: emerge.Action{
				Type: "",
				Cost: 1.0,
			},
			wantSuccess: false,
			wantCost:    0,
			wantErr:     true,
			validateFn: func(t *testing.T, agent *emerge.Agent, beforePhase float64) {
				if agent.Phase() != beforePhase {
					t.Error("Empty action should not change phase")
				}
			},
		},
		{
			name: "zero cost action",
			setupFn: func() *emerge.Agent {
				return emerge.NewAgent("test")
			},
			action: emerge.Action{
				Type: "maintain",
				Cost: 0,
			},
			wantSuccess: true,
			wantCost:    0,
			wantErr:     false,
			validateFn: func(t *testing.T, agent *emerge.Agent, beforePhase float64) {
				if agent.Energy() != 100.0 {
					t.Error("Zero cost action should not consume energy")
				}
			},
		},
		{
			name: "negative cost action (invalid)",
			setupFn: func() *emerge.Agent {
				return emerge.NewAgent("test")
			},
			action: emerge.Action{
				Type:  "adjust_phase",
				Value: 0.1,
				Cost:  -5.0,
			},
			wantSuccess: true,
			wantCost:    -5.0, // Negative cost might add energy
			wantErr:     false,
			validateFn: func(t *testing.T, agent *emerge.Agent, beforePhase float64) {
				// Implementation specific - negative cost might add energy
				if agent.Energy() < 100.0 {
					t.Error("Negative cost should not reduce energy")
				}
			},
		},
		{
			name: "insufficient energy",
			setupFn: func() *emerge.Agent {
				agent := emerge.NewAgent("test")
				// Deplete energy
				for range 19 {
					_, _, _ = agent.ApplyAction(emerge.Action{
						Type:  "adjust_phase",
						Value: 0.01,
						Cost:  5.0,
					})
				}
				return agent
			},
			action: emerge.Action{
				Type:  "adjust_phase",
				Value: 1.0,
				Cost:  50.0, // More than remaining energy
			},
			wantSuccess: false,
			wantCost:    0,
			wantErr:     true,
			validateFn: func(t *testing.T, agent *emerge.Agent, beforePhase float64) {
				if agent.Phase() != beforePhase {
					t.Error("Action with insufficient energy should not change phase")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := tt.setupFn()
			beforePhase := agent.Phase()

			success, cost, err := agent.ApplyAction(tt.action)

			if success != tt.wantSuccess {
				t.Errorf("ApplyAction() success = %v, want %v", success, tt.wantSuccess)
			}
			if cost != tt.wantCost {
				t.Errorf("ApplyAction() cost = %f, want %f", cost, tt.wantCost)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyAction() error = %v, wantErr %v", err, tt.wantErr)
			}

			tt.validateFn(t, agent, beforePhase)
		})
	}
}

func TestAgentEnergyManagement(t *testing.T) {
	tests := []struct {
		name        string
		actions     []emerge.Action
		wantEnergy  func(float64) bool
		description string
	}{
		{
			name: "single action energy consumption",
			actions: []emerge.Action{
				{Type: "adjust_phase", Value: 1.0, Cost: 10.0},
			},
			wantEnergy: func(e float64) bool {
				return e == 90.0
			},
			description: "should have 90 energy after 10 cost",
		},
		{
			name: "multiple actions energy consumption",
			actions: []emerge.Action{
				{Type: "adjust_phase", Value: 0.5, Cost: 5.0},
				{Type: "maintain", Cost: 2.0},
				{Type: "adjust_phase", Value: -0.3, Cost: 3.0},
			},
			wantEnergy: func(e float64) bool {
				return e == 90.0 // 100 - 5 - 2 - 3
			},
			description: "should have 90 energy after total 10 cost",
		},
		{
			name: "energy depletion",
			actions: []emerge.Action{
				{Type: "adjust_phase", Value: 1.0, Cost: 30.0},
				{Type: "adjust_phase", Value: 1.0, Cost: 30.0},
				{Type: "adjust_phase", Value: 1.0, Cost: 30.0},
				{Type: "adjust_phase", Value: 1.0, Cost: 30.0}, // This should fail
			},
			wantEnergy: func(e float64) bool {
				return e == 10.0 // 100 - 30 - 30 - 30
			},
			description: "should have 10 energy, last action should fail",
		},
		{
			name: "zero cost actions",
			actions: []emerge.Action{
				{Type: "maintain", Cost: 0},
				{Type: "maintain", Cost: 0},
				{Type: "maintain", Cost: 0},
			},
			wantEnergy: func(e float64) bool {
				return e == 100.0
			},
			description: "should maintain full energy with zero cost actions",
		},
		{
			name: "mixed cost actions",
			actions: []emerge.Action{
				{Type: "adjust_phase", Value: 0.1, Cost: 15.5},
				{Type: "maintain", Cost: 0},
				{Type: "adjust_phase", Value: 0.2, Cost: 24.5},
			},
			wantEnergy: func(e float64) bool {
				return e == 60.0 // 100 - 15.5 - 24.5
			},
			description: "should handle fractional costs correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := emerge.NewAgent("test")

			for _, action := range tt.actions {
				_, _, _ = agent.ApplyAction(action)
			}

			energy := agent.Energy()
			if !tt.wantEnergy(energy) {
				t.Errorf("Energy = %f, %s", energy, tt.description)
			}
		})
	}
}

func TestAgentWithOptions(t *testing.T) {
	tests := []struct {
		name     string
		options  []emerge.AgentOption
		validate func(t *testing.T, agent *emerge.Agent)
	}{
		{
			name: "with phase",
			options: []emerge.AgentOption{
				emerge.WithPhase(1.5),
			},
			validate: func(t *testing.T, agent *emerge.Agent) {
				if agent.Phase() != 1.5 {
					t.Errorf("Phase = %f, want 1.5", agent.Phase())
				}
			},
		},
		{
			name: "with stubbornness",
			options: []emerge.AgentOption{
				emerge.WithStubbornness(0.05),
			},
			validate: func(t *testing.T, agent *emerge.Agent) {
				if agent.Stubbornness() != 0.05 {
					t.Errorf("Stubbornness = %f, want 0.05", agent.Stubbornness())
				}
			},
		},
		{
			name: "with influence",
			options: []emerge.AgentOption{
				emerge.WithInfluence(0.9),
			},
			validate: func(t *testing.T, agent *emerge.Agent) {
				if agent.Influence() != 0.9 {
					t.Errorf("Influence = %f, want 0.9", agent.Influence())
				}
			},
		},
		{
			name: "multiple options",
			options: []emerge.AgentOption{
				emerge.WithPhase(math.Pi),
				emerge.WithStubbornness(0.1),
				emerge.WithInfluence(0.6),
			},
			validate: func(t *testing.T, agent *emerge.Agent) {
				if math.Abs(agent.Phase()-math.Pi) > 0.01 {
					t.Errorf("Phase = %f, want π", agent.Phase())
				}
				if agent.Stubbornness() != 0.1 {
					t.Errorf("Stubbornness = %f, want 0.1", agent.Stubbornness())
				}
				if agent.Influence() != 0.6 {
					t.Errorf("Influence = %f, want 0.6", agent.Influence())
				}
			},
		},
		{
			name: "options with clamping",
			options: []emerge.AgentOption{
				emerge.WithStubbornness(-0.5), // Should clamp to 0
				emerge.WithInfluence(1.5),     // Should clamp to 1
			},
			validate: func(t *testing.T, agent *emerge.Agent) {
				if agent.Stubbornness() < 0 {
					t.Errorf("Stubbornness = %f, should be >= 0", agent.Stubbornness())
				}
				if agent.Influence() > 1.0 {
					t.Errorf("Influence = %f, should be <= 1.0", agent.Influence())
				}
			},
		},
		{
			name:    "no options (defaults)",
			options: []emerge.AgentOption{},
			validate: func(t *testing.T, agent *emerge.Agent) {
				if agent.Energy() != 100.0 {
					t.Errorf("Energy = %f, want 100.0 (default)", agent.Energy())
				}
				// Check other defaults are in expected ranges
				if agent.Influence() < 0.3 || agent.Influence() > 0.7 {
					t.Errorf("Influence = %f, want in default range [0.3, 0.7]", agent.Influence())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := emerge.NewAgent("test", tt.options...)
			tt.validate(t, agent)
		})
	}
}

func TestAgentFromConfig(t *testing.T) {
	tests := []struct {
		name     string
		configFn func() emerge.AgentConfig
		validate func(t *testing.T, agent *emerge.Agent, config emerge.AgentConfig)
	}{
		{
			name:     "default config",
			configFn: emerge.DefaultAgentConfig,
			validate: func(t *testing.T, agent *emerge.Agent, config emerge.AgentConfig) {
				if agent.Energy() != 100.0 {
					t.Errorf("Energy = %f, want 100.0", agent.Energy())
				}
				if config.RandomizePhase {
					t.Error("Default config should not randomize phase")
				}
			},
		},
		{
			name:     "test config",
			configFn: emerge.TestAgentConfig,
			validate: func(t *testing.T, agent *emerge.Agent, config emerge.AgentConfig) {
				if agent.Stubbornness() != 0.01 {
					t.Errorf("Stubbornness = %f, want 0.01", agent.Stubbornness())
				}
				if agent.Influence() != 0.8 {
					t.Errorf("Influence = %f, want 0.8", agent.Influence())
				}
			},
		},
		{
			name:     "randomized config",
			configFn: emerge.RandomizedAgentConfig,
			validate: func(t *testing.T, agent *emerge.Agent, config emerge.AgentConfig) {
				if !config.RandomizePhase {
					t.Error("Randomized config should randomize phase")
				}
				phase := agent.Phase()
				if phase < 0 || phase > 2*math.Pi {
					t.Errorf("Phase = %f, want in [0, 2π]", phase)
				}
			},
		},
		{
			name: "custom config with extreme values",
			configFn: func() emerge.AgentConfig {
				return emerge.AgentConfig{
					Phase:          3 * math.Pi, // Should wrap
					InitialEnergy:  200.0,       // Above normal
					Stubbornness:   2.0,         // Should clamp
					Influence:      -0.5,        // Should clamp
					RandomizePhase: false,
				}
			},
			validate: func(t *testing.T, agent *emerge.Agent, config emerge.AgentConfig) {
				// Phase should wrap
				phase := agent.Phase()
				if phase < 0 || phase > 2*math.Pi {
					t.Errorf("Phase = %f, should be wrapped to [0, 2π]", phase)
				}
				// Stubbornness should be clamped
				if agent.Stubbornness() < 0 || agent.Stubbornness() > 1 {
					t.Errorf("Stubbornness = %f, should be clamped to [0, 1]", agent.Stubbornness())
				}
			},
		},
		{
			name: "config with zero values",
			configFn: func() emerge.AgentConfig {
				return emerge.AgentConfig{
					Phase:          0,
					InitialEnergy:  0,
					Stubbornness:   0,
					Influence:      0,
					RandomizePhase: false,
				}
			},
			validate: func(t *testing.T, agent *emerge.Agent, config emerge.AgentConfig) {
				if agent.Phase() != 0 {
					t.Errorf("Phase = %f, want 0", agent.Phase())
				}
				// Zero values should be replaced with defaults
				if agent.Energy() != 100.0 {
					t.Errorf("Energy = %f, want 100.0 (default for 0)", agent.Energy())
				}
				if agent.Stubbornness() != 0.2 {
					t.Errorf("Stubbornness = %f, want 0.2 (default for 0)", agent.Stubbornness())
				}
				if agent.Influence() != 0.5 {
					t.Errorf("Influence = %f, want 0.5 (default for 0)", agent.Influence())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.configFn()
			agent := emerge.NewAgentFromConfig("test", config)
			if agent == nil {
				t.Fatal("NewAgentFromConfig returned nil")
			}
			tt.validate(t, agent, config)
		})
	}
}

func TestAgentConcurrency(t *testing.T) {
	agent := emerge.NewAgent("concurrent-test")

	// Test concurrent reads
	done := make(chan bool, 10)
	for range 10 {
		go func() {
			_ = agent.Phase()
			_ = agent.Energy()
			_ = agent.LocalGoal()
			_ = agent.Influence()
			_ = agent.Stubbornness()
			done <- true
		}()
	}

	for range 10 {
		<-done
	}

	// Test concurrent writes
	for i := range 10 {
		go func(val float64) {
			agent.SetPhase(val)
			agent.SetLocalGoal(val)
			agent.SetInfluence(val / 10)
			agent.SetStubbornness(val / 20)
			done <- true
		}(float64(i))
	}

	for range 10 {
		<-done
	}

	// Test concurrent actions
	for range 10 {
		go func() {
			action := emerge.Action{
				Type:  "adjust_phase",
				Value: 0.01,
				Cost:  0.1,
			}
			_, _, _ = agent.ApplyAction(action)
			done <- true
		}()
	}

	for range 10 {
		<-done
	}

	// Energy should be reduced after concurrent actions
	if agent.Energy() >= 100.0 {
		t.Error("Energy should be reduced after concurrent actions")
	}
}

func TestAgentNeighborManagement(t *testing.T) {
	tests := []struct {
		name     string
		setupFn  func() (*emerge.Agent, *emerge.Agent)
		validate func(t *testing.T, agent1, agent2 *emerge.Agent)
	}{
		{
			name: "add neighbor",
			setupFn: func() (*emerge.Agent, *emerge.Agent) {
				agent1 := emerge.NewAgent("agent1")
				agent2 := emerge.NewAgent("agent2")
				return agent1, agent2
			},
			validate: func(t *testing.T, agent1, agent2 *emerge.Agent) {
				// Access neighbors through public method
				neighbors := agent1.Neighbors()
				if neighbors == nil {
					t.Error("Neighbors() should not return nil")
				}

				// Store neighbor
				neighbors.Store(agent2.ID, agent2)

				// Verify neighbor was added
				val, exists := neighbors.Load(agent2.ID)
				if !exists {
					t.Error("Neighbor should exist after adding")
				}
				if val.(*emerge.Agent).ID != agent2.ID {
					t.Error("Wrong neighbor stored")
				}
			},
		},
		{
			name: "multiple neighbors",
			setupFn: func() (*emerge.Agent, *emerge.Agent) {
				agent1 := emerge.NewAgent("agent1")
				agent2 := emerge.NewAgent("agent2")
				agent3 := emerge.NewAgent("agent3")

				neighbors := agent1.Neighbors()
				neighbors.Store(agent2.ID, agent2)
				neighbors.Store(agent3.ID, agent3)

				return agent1, agent2
			},
			validate: func(t *testing.T, agent1, agent2 *emerge.Agent) {
				count := 0
				agent1.Neighbors().Range(func(key, value any) bool {
					count++
					return true
				})

				if count != 2 {
					t.Errorf("Expected 2 neighbors, got %d", count)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent1, agent2 := tt.setupFn()
			tt.validate(t, agent1, agent2)
		})
	}
}
