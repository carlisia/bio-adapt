package agent_test

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/carlisia/bio-adapt/emerge/agent"
	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/internal/config"
)

func TestNew(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		id         string
		validateFn func(t *testing.T, agent *agent.Agent)
	}{
		{
			name: "basic agent creation",
			id:   "test-agent",
			validateFn: func(t *testing.T, agent *agent.Agent) {
				t.Helper()
				assert.Equal(t, "test-agent", agent.ID, "ID should match expected value")
			},
		},
		{
			name: "empty ID",
			id:   "",
			validateFn: func(t *testing.T, agent *agent.Agent) {
				t.Helper()
				assert.Equal(t, "", agent.ID, "ID should be empty string")
			},
		},
		{
			name: "special characters in ID",
			id:   "agent-123!@#$%",
			validateFn: func(t *testing.T, agent *agent.Agent) {
				t.Helper()
				assert.Equal(t, "agent-123!@#$%", agent.ID, "ID should preserve special characters")
			},
		},
		{
			name: "very long ID",
			id:   "this-is-a-very-long-agent-id-that-exceeds-normal-length-expectations-and-should-still-work-properly",
			validateFn: func(t *testing.T, agent *agent.Agent) {
				t.Helper()
				assert.Equal(t, "this-is-a-very-long-agent-id-that-exceeds-normal-length-expectations-and-should-still-work-properly", agent.ID, "Long ID should be preserved")
			},
		},
		{
			name: "initial values in range",
			id:   "range-test",
			validateFn: func(t *testing.T, agent *agent.Agent) {
				t.Helper()
				phase := agent.Phase()
				assert.GreaterOrEqual(t, phase, 0.0, "Phase should be >= 0")
				assert.LessOrEqual(t, phase, 2*math.Pi, "Phase should be <= 2π")

				energy := agent.Energy()
				assert.Equal(t, 100.0, energy, "Energy should be 100.0")

				localGoal := agent.LocalGoal()
				assert.GreaterOrEqual(t, localGoal, 0.0, "LocalGoal should be >= 0")
				assert.LessOrEqual(t, localGoal, 2*math.Pi, "LocalGoal should be <= 2π")

				influence := agent.Influence()
				assert.GreaterOrEqual(t, influence, 0.1, "Influence should be >= 0.1")
				assert.LessOrEqual(t, influence, 0.2, "Influence should be <= 0.2")

				stubbornness := agent.Stubbornness()
				assert.GreaterOrEqual(t, stubbornness, 0.0, "Stubbornness should be >= 0")
				assert.LessOrEqual(t, stubbornness, 0.3, "Stubbornness should be <= 0.3")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			agent := agent.New(tt.id)
			require.NotNil(t, agent, "New should not return nil")
			tt.validateFn(t, agent)
		})
	}
}

func TestAgentSettersAndGetters(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		setupFn func(agent *agent.Agent)
		checkFn func(t *testing.T, agent *agent.Agent)
	}{
		{
			name: "set phase to π",
			setupFn: func(agent *agent.Agent) {
				agent.SetPhase(math.Pi)
			},
			checkFn: func(t *testing.T, agent *agent.Agent) {
				t.Helper()
				assert.InDelta(t, math.Pi, agent.Phase(), 0.01, "Phase should be π")
			},
		},
		{
			name: "set phase to 0",
			setupFn: func(agent *agent.Agent) {
				agent.SetPhase(0)
			},
			checkFn: func(t *testing.T, agent *agent.Agent) {
				t.Helper()
				assert.Equal(t, 0.0, agent.Phase(), "Phase should be 0")
			},
		},
		{
			name: "set phase to 2π",
			setupFn: func(agent *agent.Agent) {
				agent.SetPhase(2 * math.Pi)
			},
			checkFn: func(t *testing.T, agent *agent.Agent) {
				t.Helper()
				// Should wrap to 0
				phase := agent.Phase()
				assert.True(t, math.Abs(phase) <= 0.01 || math.Abs(phase-2*math.Pi) <= 0.01, "Phase should wrap to 0 or 2π, got %f", phase)
			},
		},
		{
			name: "set phase to negative (wrapping)",
			setupFn: func(agent *agent.Agent) {
				agent.SetPhase(-math.Pi)
			},
			checkFn: func(t *testing.T, agent *agent.Agent) {
				t.Helper()
				// Should wrap to π
				assert.InDelta(t, math.Pi, agent.Phase(), 0.01, "Phase should wrap to π")
			},
		},
		{
			name: "set local goal to π/2",
			setupFn: func(agent *agent.Agent) {
				agent.SetLocalGoal(math.Pi / 2)
			},
			checkFn: func(t *testing.T, agent *agent.Agent) {
				t.Helper()
				assert.InDelta(t, math.Pi/2, agent.LocalGoal(), 0.01, "LocalGoal should be π/2")
			},
		},
		{
			name: "set influence",
			setupFn: func(agent *agent.Agent) {
				agent.SetInfluence(0.75)
			},
			checkFn: func(t *testing.T, agent *agent.Agent) {
				t.Helper()
				assert.InDelta(t, 0.75, agent.Influence(), 0.01, "Influence should be 0.75")
			},
		},
		{
			name: "set influence out of range (clamping)",
			setupFn: func(agent *agent.Agent) {
				agent.SetInfluence(1.5)
			},
			checkFn: func(t *testing.T, agent *agent.Agent) {
				t.Helper()
				assert.LessOrEqual(t, agent.Influence(), 1.0, "Influence should be clamped to <= 1.0")
			},
		},
		{
			name: "set stubbornness",
			setupFn: func(agent *agent.Agent) {
				agent.SetStubbornness(0.2)
			},
			checkFn: func(t *testing.T, agent *agent.Agent) {
				t.Helper()
				assert.InDelta(t, 0.2, agent.Stubbornness(), 0.01, "Stubbornness should be 0.2")
			},
		},
		{
			name: "set negative stubbornness (clamping)",
			setupFn: func(agent *agent.Agent) {
				agent.SetStubbornness(-0.5)
			},
			checkFn: func(t *testing.T, agent *agent.Agent) {
				t.Helper()
				assert.GreaterOrEqual(t, agent.Stubbornness(), 0.0, "Stubbornness should be clamped to >= 0")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			agent := agent.New("test")
			tt.setupFn(agent)
			tt.checkFn(t, agent)
		})
	}
}

func TestAgentProposeAdjustment(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		setupFn    func() *agent.Agent
		globalGoal core.State
		validateFn func(t *testing.T, action core.Action, accepted bool)
	}{
		{
			name: "agent far from goal",
			setupFn: func() *agent.Agent {
				agent := agent.New("test")
				agent.SetPhase(0)
				agent.SetLocalGoal(math.Pi)
				return agent
			},
			globalGoal: core.State{
				Phase:     math.Pi / 2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			validateFn: func(t *testing.T, action core.Action, accepted bool) {
				t.Helper()
				if accepted {
					assert.NotEmpty(t, action.Type, "Accepted action should have a type")
					assert.GreaterOrEqual(t, action.Cost, 0.0, "Action cost should be non-negative")
				}
			},
		},
		{
			name: "agent at goal",
			setupFn: func() *agent.Agent {
				agent := agent.New("test")
				agent.SetPhase(math.Pi)
				agent.SetLocalGoal(math.Pi)
				return agent
			},
			globalGoal: core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			validateFn: func(t *testing.T, action core.Action, accepted bool) {
				t.Helper()
				if accepted && action.Type == "adjust_phase" {
					// Should maintain position more likely
					assert.LessOrEqual(t, math.Abs(action.Value), 0.1, "Agent at goal should propose small adjustments")
				}
			},
		},
		{
			name: "stubborn agent",
			setupFn: func() *agent.Agent {
				return agent.New("stubborn",
					agent.WithStubbornness(0.99),
					agent.WithPhase(0),
				)
			},
			globalGoal: core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			validateFn: func(t *testing.T, action core.Action, accepted bool) {
				t.Helper()
				// Stubborn agents accept fewer proposals
				// This is probabilistic, so we just check valid structure
				if accepted {
					assert.NotEmpty(t, action.Type, "Action should have a type")
				}
			},
		},
		{
			name: "low energy agent",
			setupFn: func() *agent.Agent {
				agent := agent.New("test")
				// Deplete energy
				for range 20 {
					_, _, _ = agent.ApplyAction(core.Action{
						Type:  "adjust_phase",
						Value: 0.1,
						Cost:  5.0,
					})
				}
				return agent
			},
			globalGoal: core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			validateFn: func(t *testing.T, action core.Action, accepted bool) {
				t.Helper()
				if accepted {
					// Low energy should prefer low-cost actions
					assert.LessOrEqual(t, action.Cost, 5.0, "Low energy agent should prefer low-cost actions")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			agent := tt.setupFn()
			action, accepted := agent.ProposeAdjustment(tt.globalGoal)
			tt.validateFn(t, action, accepted)
		})
	}
}

func TestAgentApplyAction(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		setupFn     func() *agent.Agent
		action      core.Action
		wantSuccess bool
		wantCost    float64
		wantErr     bool
		validateFn  func(t *testing.T, agent *agent.Agent, beforePhase float64)
	}{
		{
			name: "adjust phase positive",
			setupFn: func() *agent.Agent {
				agent := agent.New("test")
				agent.SetPhase(0)
				return agent
			},
			action: core.Action{
				Type:  "adjust_phase",
				Value: 0.5,
				Cost:  1.0,
			},
			wantSuccess: true,
			wantCost:    1.0,
			wantErr:     false,
			validateFn: func(t *testing.T, agent *agent.Agent, beforePhase float64) {
				t.Helper()
				expectedPhase := math.Mod(beforePhase+0.5, 2*math.Pi)
				assert.InDelta(t, expectedPhase, agent.Phase(), 0.01, "Phase should be updated correctly")
			},
		},
		{
			name: "adjust phase negative",
			setupFn: func() *agent.Agent {
				agent := agent.New("test")
				agent.SetPhase(math.Pi)
				return agent
			},
			action: core.Action{
				Type:  "adjust_phase",
				Value: -0.5,
				Cost:  1.0,
			},
			wantSuccess: true,
			wantCost:    1.0,
			wantErr:     false,
			validateFn: func(t *testing.T, agent *agent.Agent, beforePhase float64) {
				t.Helper()
				expectedPhase := beforePhase - 0.5
				if expectedPhase < 0 {
					expectedPhase += 2 * math.Pi
				}
				assert.InDelta(t, expectedPhase, agent.Phase(), 0.01, "Phase should be updated with wrapping")
			},
		},
		{
			name: "adjust phase large value (wrapping)",
			setupFn: func() *agent.Agent {
				agent := agent.New("test")
				agent.SetPhase(math.Pi)
				return agent
			},
			action: core.Action{
				Type:  "adjust_phase",
				Value: 3 * math.Pi,
				Cost:  2.0,
			},
			wantSuccess: true,
			wantCost:    2.0,
			validateFn: func(t *testing.T, agent *agent.Agent, beforePhase float64) {
				t.Helper()
				// Should wrap around
				expectedPhase := math.Mod(beforePhase+3*math.Pi, 2*math.Pi)
				assert.InDelta(t, expectedPhase, agent.Phase(), 0.01, "Phase should wrap around correctly")
			},
		},
		{
			name: "maintain action",
			setupFn: func() *agent.Agent {
				agent := agent.New("test")
				agent.SetPhase(1.5)
				return agent
			},
			action: core.Action{
				Type: "maintain",
				Cost: 0.1,
			},
			wantSuccess: true,
			wantCost:    0.1,
			wantErr:     false,
			validateFn: func(t *testing.T, agent *agent.Agent, beforePhase float64) {
				t.Helper()
				assert.Equal(t, beforePhase, agent.Phase(), "Phase should remain unchanged for maintain action")
			},
		},
		{
			name: "invalid action type",
			setupFn: func() *agent.Agent {
				return agent.New("test")
			},
			action: core.Action{
				Type: "invalid_action",
				Cost: 1.0,
			},
			wantSuccess: false,
			wantCost:    0,
			wantErr:     true,
			validateFn: func(t *testing.T, agent *agent.Agent, beforePhase float64) {
				t.Helper()
				assert.Equal(t, beforePhase, agent.Phase(), "Invalid action should not change phase")
			},
		},
		{
			name: "empty action type",
			setupFn: func() *agent.Agent {
				return agent.New("test")
			},
			action: core.Action{
				Type: "",
				Cost: 1.0,
			},
			wantSuccess: false,
			wantCost:    0,
			wantErr:     true,
			validateFn: func(t *testing.T, agent *agent.Agent, beforePhase float64) {
				t.Helper()
				assert.Equal(t, beforePhase, agent.Phase(), "Empty action should not change phase")
			},
		},
		{
			name: "zero cost action",
			setupFn: func() *agent.Agent {
				return agent.New("test")
			},
			action: core.Action{
				Type: "maintain",
				Cost: 0,
			},
			wantSuccess: true,
			wantCost:    0,
			wantErr:     false,
			validateFn: func(t *testing.T, agent *agent.Agent, _ float64) {
				t.Helper()
				assert.Equal(t, 100.0, agent.Energy(), "Zero cost action should not consume energy")
			},
		},
		{
			name: "negative cost action (invalid)",
			setupFn: func() *agent.Agent {
				return agent.New("test")
			},
			action: core.Action{
				Type:  "adjust_phase",
				Value: 0.1,
				Cost:  -5.0,
			},
			wantSuccess: true,
			wantCost:    -5.0, // Negative cost might add energy
			wantErr:     false,
			validateFn: func(t *testing.T, agent *agent.Agent, _ float64) {
				t.Helper()
				// Implementation specific - negative cost might add energy
				assert.GreaterOrEqual(t, agent.Energy(), 100.0, "Negative cost should not reduce energy")
			},
		},
		{
			name: "insufficient energy",
			setupFn: func() *agent.Agent {
				agent := agent.New("test")
				// Deplete energy
				for range 19 {
					_, _, _ = agent.ApplyAction(core.Action{
						Type:  "adjust_phase",
						Value: 0.01,
						Cost:  5.0,
					})
				}
				return agent
			},
			action: core.Action{
				Type:  "adjust_phase",
				Value: 1.0,
				Cost:  50.0, // More than remaining energy
			},
			wantSuccess: false,
			wantCost:    0,
			wantErr:     true,
			validateFn: func(t *testing.T, agent *agent.Agent, beforePhase float64) {
				t.Helper()
				assert.Equal(t, beforePhase, agent.Phase(), "Action with insufficient energy should not change phase")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			agent := tt.setupFn()
			beforePhase := agent.Phase()

			success, cost, err := agent.ApplyAction(tt.action)

			assert.Equal(t, tt.wantSuccess, success, "ApplyAction() success should match expected")
			assert.Equal(t, tt.wantCost, cost, "ApplyAction() cost should match expected")
			if tt.wantErr {
				assert.Error(t, err, "ApplyAction() should return error")
			} else {
				assert.NoError(t, err, "ApplyAction() should not return error")
			}

			tt.validateFn(t, agent, beforePhase)
		})
	}
}

func TestAgentEnergyManagement(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		actions     []core.Action
		wantEnergy  func(float64) bool
		description string
	}{
		{
			name: "single action energy consumption",
			actions: []core.Action{
				{Type: "adjust_phase", Value: 1.0, Cost: 10.0},
			},
			wantEnergy: func(e float64) bool {
				return e == 90.0
			},
			description: "should have 90 energy after 10 cost",
		},
		{
			name: "multiple actions energy consumption",
			actions: []core.Action{
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
			actions: []core.Action{
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
			actions: []core.Action{
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
			actions: []core.Action{
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
			t.Parallel()
			agent := agent.New("test")

			for _, action := range tt.actions {
				_, _, _ = agent.ApplyAction(action)
			}

			energy := agent.Energy()
			assert.True(t, tt.wantEnergy(energy), "Energy = %f, %s", energy, tt.description)
		})
	}
}

func TestAgentWithOptions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		options  []agent.Option
		validate func(t *testing.T, agent *agent.Agent)
	}{
		{
			name: "with phase",
			options: []agent.Option{
				agent.WithPhase(1.5),
			},
			validate: func(t *testing.T, agent *agent.Agent) {
				t.Helper()
				assert.Equal(t, 1.5, agent.Phase(), "Phase should be 1.5")
			},
		},
		{
			name: "with stubbornness",
			options: []agent.Option{
				agent.WithStubbornness(0.05),
			},
			validate: func(t *testing.T, agent *agent.Agent) {
				t.Helper()
				assert.Equal(t, 0.05, agent.Stubbornness(), "Stubbornness should be 0.05")
			},
		},
		{
			name: "with influence",
			options: []agent.Option{
				agent.WithInfluence(0.9),
			},
			validate: func(t *testing.T, agent *agent.Agent) {
				t.Helper()
				assert.Equal(t, 0.9, agent.Influence(), "Influence should be 0.9")
			},
		},
		{
			name: "multiple options",
			options: []agent.Option{
				agent.WithPhase(math.Pi),
				agent.WithStubbornness(0.1),
				agent.WithInfluence(0.6),
			},
			validate: func(t *testing.T, agent *agent.Agent) {
				t.Helper()
				assert.InDelta(t, math.Pi, agent.Phase(), 0.01, "Phase should be π")
				assert.Equal(t, 0.1, agent.Stubbornness(), "Stubbornness should be 0.1")
				assert.Equal(t, 0.6, agent.Influence(), "Influence should be 0.6")
			},
		},
		{
			name: "options with clamping",
			options: []agent.Option{
				agent.WithStubbornness(-0.5), // Should clamp to 0
				agent.WithInfluence(1.5),     // Should clamp to 1
			},
			validate: func(t *testing.T, agent *agent.Agent) {
				t.Helper()
				assert.GreaterOrEqual(t, agent.Stubbornness(), 0.0, "Stubbornness should be >= 0")
				assert.LessOrEqual(t, agent.Influence(), 1.0, "Influence should be <= 1.0")
			},
		},
		{
			name:    "no options (defaults)",
			options: []agent.Option{},
			validate: func(t *testing.T, agent *agent.Agent) {
				t.Helper()
				assert.Equal(t, 100.0, agent.Energy(), "Energy should be 100.0 (default)")
				// Check other defaults are in expected ranges
				assert.GreaterOrEqual(t, agent.Influence(), 0.1, "Influence should be >= 0.1")
				assert.LessOrEqual(t, agent.Influence(), 0.2, "Influence should be <= 0.2")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			agent := agent.New("test", tt.options...)
			tt.validate(t, agent)
		})
	}
}

func TestAgentFromConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		configFn func() config.Agent
		validate func(t *testing.T, agent *agent.Agent, _ config.Agent)
	}{
		{
			name:     "default config",
			configFn: config.DefaultAgent,
			validate: func(t *testing.T, agent *agent.Agent, _ config.Agent) {
				t.Helper()
				assert.Equal(t, 100.0, agent.Energy(), "Energy should be 100.0")
				// Note: Can't check cfg.RandomizePhase without the parameter
			},
		},
		{
			name:     "test config",
			configFn: config.TestAgent,
			validate: func(t *testing.T, agent *agent.Agent, _ config.Agent) {
				t.Helper()
				assert.Equal(t, 0.01, agent.Stubbornness(), "Stubbornness should be 0.01")
				assert.Equal(t, 0.8, agent.Influence(), "Influence should be 0.8")
			},
		},
		{
			name:     "randomized config",
			configFn: config.RandomizedAgent,
			validate: func(t *testing.T, agent *agent.Agent, _ config.Agent) {
				t.Helper()
				// Note: Can't check cfg.RandomizePhase without the parameter
				phase := agent.Phase()
				assert.GreaterOrEqual(t, phase, 0.0, "Phase should be >= 0")
				assert.LessOrEqual(t, phase, 2*math.Pi, "Phase should be <= 2π")
			},
		},
		{
			name: "custom config with extreme values",
			configFn: func() config.Agent {
				return config.Agent{
					Phase:          3 * math.Pi, // Should wrap
					InitialEnergy:  200.0,       // Above normal
					Stubbornness:   2.0,         // Should clamp
					Influence:      -0.5,        // Should clamp
					RandomizePhase: false,
				}
			},
			validate: func(t *testing.T, agent *agent.Agent, _ config.Agent) {
				t.Helper()
				// Phase should wrap
				phase := agent.Phase()
				assert.GreaterOrEqual(t, phase, 0.0, "Phase should be wrapped to >= 0")
				assert.LessOrEqual(t, phase, 2*math.Pi, "Phase should be wrapped to <= 2π")
				// Stubbornness should be clamped
				assert.GreaterOrEqual(t, agent.Stubbornness(), 0.0, "Stubbornness should be clamped to >= 0")
				assert.LessOrEqual(t, agent.Stubbornness(), 1.0, "Stubbornness should be clamped to <= 1")
			},
		},
		{
			name: "config with zero values",
			configFn: func() config.Agent {
				return config.Agent{
					Phase:          0,
					InitialEnergy:  0,
					Stubbornness:   0,
					Influence:      0,
					RandomizePhase: false,
				}
			},
			validate: func(t *testing.T, agent *agent.Agent, _ config.Agent) {
				t.Helper()
				assert.Equal(t, 0.0, agent.Phase(), "Phase should be 0")
				// Zero values should be replaced with defaults
				assert.Equal(t, 100.0, agent.Energy(), "Energy should be 100.0 (default for 0)")
				assert.Equal(t, 0.2, agent.Stubbornness(), "Stubbornness should be 0.2 (default for 0)")
				assert.Equal(t, 0.5, agent.Influence(), "Influence should be 0.5 (default for 0)")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			config := tt.configFn()
			agent, err := agent.NewFromConfig("test", config)
			require.NoError(t, err, "NewFromConfig should not fail")
			require.NotNil(t, agent, "NewFromConfig should not return nil")
			tt.validate(t, agent, config)
		})
	}
}

func TestAgentConcurrency(t *testing.T) {
	t.Parallel()
	a := agent.New("concurrent-test")

	// Test concurrent reads
	done := make(chan bool, 10)
	for range 10 {
		go func() {
			_ = a.Phase()
			_ = a.Energy()
			_ = a.LocalGoal()
			_ = a.Influence()
			_ = a.Stubbornness()
			done <- true
		}()
	}

	for range 10 {
		<-done
	}

	// Test concurrent writes
	for i := range 10 {
		go func(val float64) {
			a.SetPhase(val)
			a.SetLocalGoal(val)
			a.SetInfluence(val / 10)
			a.SetStubbornness(val / 20)
			done <- true
		}(float64(i))
	}

	for range 10 {
		<-done
	}

	// Test concurrent actions
	for range 10 {
		go func() {
			action := core.Action{
				Type:  "adjust_phase",
				Value: 0.01,
				Cost:  0.1,
			}
			_, _, _ = a.ApplyAction(action)
			done <- true
		}()
	}

	for range 10 {
		<-done
	}

	// Energy should be reduced after concurrent actions
	assert.Less(t, a.Energy(), 100.0, "Energy should be reduced after concurrent actions")
}

func TestAgentNeighborManagement(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		setupFn  func() (*agent.Agent, *agent.Agent)
		validate func(t *testing.T, agent1, _ *agent.Agent)
	}{
		{
			name: "add neighbor",
			setupFn: func() (*agent.Agent, *agent.Agent) {
				agent1 := agent.New("agent1")
				agent2 := agent.New("agent2")
				return agent1, agent2
			},
			validate: func(t *testing.T, agent1, agent2 *agent.Agent) {
				t.Helper()
				// Access neighbors through public method
				neighbors := agent1.Neighbors()
				assert.NotNil(t, neighbors, "Neighbors() should not return nil")

				// Store neighbor
				neighbors.Store(agent2.ID, agent2)

				// Verify neighbor was added
				val, exists := neighbors.Load(agent2.ID)
				assert.True(t, exists, "Neighbor should exist after adding")
				assert.Equal(t, agent2.ID, val.(*agent.Agent).ID, "Correct neighbor should be stored")
			},
		},
		{
			name: "multiple neighbors",
			setupFn: func() (*agent.Agent, *agent.Agent) {
				agent1 := agent.New("agent1")
				agent2 := agent.New("agent2")
				agent3 := agent.New("agent3")

				neighbors := agent1.Neighbors()
				neighbors.Store(agent2.ID, agent2)
				neighbors.Store(agent3.ID, agent3)

				return agent1, agent2
			},
			validate: func(t *testing.T, agent1, _ *agent.Agent) {
				t.Helper()
				count := 0
				agent1.Neighbors().Range(func(_, _ any) bool {
					count++
					return true
				})

				assert.Equal(t, 2, count, "Should have exactly 2 neighbors")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			agent1, agent2 := tt.setupFn()
			tt.validate(t, agent1, agent2)
		})
	}
}
