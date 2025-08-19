package emerge

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestNewSwarm(t *testing.T) {
	tests := []struct {
		name       string
		size       int
		goal       State
		wantErr    bool
		validateFn func(t *testing.T, swarm *Swarm)
	}{
		{
			name: "basic swarm creation",
			size: 10,
			goal: State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			wantErr: false,
			validateFn: func(t *testing.T, swarm *Swarm) {
				if swarm.Size() != 10 {
					t.Errorf("Expected 10 agents, got %d", swarm.Size())
				}
				if swarm.goalState.Phase != 0 {
					t.Error("Goal state not set correctly")
				}
				if swarm.monitor == nil {
					t.Error("Monitor not initialized")
				}
			},
		},
		{
			name: "single agent swarm",
			size: 1,
			goal: State{
				Phase:     math.Pi,
				Frequency: 200 * time.Millisecond,
				Coherence: 0.5,
			},
			wantErr: false,
			validateFn: func(t *testing.T, swarm *Swarm) {
				if swarm.Size() != 1 {
					t.Error("Single agent swarm should have size 1")
				}
			},
		},
		{
			name: "zero size swarm",
			size: 0,
			goal: State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			wantErr: true,
			validateFn: func(t *testing.T, swarm *Swarm) {
				if swarm != nil {
					t.Error("Zero size swarm should not be created")
				}
			},
		},
		{
			name: "negative size swarm",
			size: -5,
			goal: State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			wantErr: true,
			validateFn: func(t *testing.T, swarm *Swarm) {
				if swarm != nil {
					t.Error("Negative size swarm should not be created")
				}
			},
		},
		{
			name: "large swarm",
			size: 1000,
			goal: State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			wantErr: false,
			validateFn: func(t *testing.T, swarm *Swarm) {
				if swarm.Size() != 1000 {
					t.Errorf("Expected 1000 agents, got %d", swarm.Size())
				}
			},
		},
		{
			name: "swarm with zero frequency goal",
			size: 10,
			goal: State{
				Phase:     0,
				Frequency: 0,
				Coherence: 0.9,
			},
			wantErr: true,
			validateFn: func(t *testing.T, swarm *Swarm) {
				if swarm != nil {
					t.Error("Zero frequency should be rejected")
				}
			},
		},
		{
			name: "swarm with negative coherence goal",
			size: 10,
			goal: State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: -0.5,
			},
			wantErr: true,
			validateFn: func(t *testing.T, swarm *Swarm) {
				if swarm != nil {
					t.Error("Negative coherence should be rejected")
				}
			},
		},
		{
			name: "swarm with coherence > 1",
			size: 10,
			goal: State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 1.5,
			},
			wantErr: true,
			validateFn: func(t *testing.T, swarm *Swarm) {
				if swarm != nil {
					t.Error("Coherence > 1 should be rejected")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			swarm, err := NewSwarm(tt.size, tt.goal)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSwarm() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.validateFn(t, swarm)
		})
	}
}

func TestSwarmConnectToNeighbors(t *testing.T) {
	tests := []struct {
		name       string
		size       int
		validateFn func(t *testing.T, swarm *Swarm)
	}{
		{
			name: "small swarm connectivity",
			size: 5,
			validateFn: func(t *testing.T, swarm *Swarm) {
				// Check that at least some connections exist
				totalConnections := 0
				swarm.agents.Range(func(key, value any) bool {
					agent := value.(*Agent)
					agent.neighbors.Range(func(k, v any) bool {
						totalConnections++
						return true
					})
					return true
				})

				if totalConnections == 0 {
					t.Error("No connections established in small swarm")
				}
			},
		},
		{
			name: "medium swarm connectivity",
			size: 20,
			validateFn: func(t *testing.T, swarm *Swarm) {
				// Count total connections
				totalConnections := 0
				agentCount := 0
				swarm.agents.Range(func(key, value any) bool {
					agent := value.(*Agent)
					agentCount++
					agent.neighbors.Range(func(k, v any) bool {
						totalConnections++
						return true
					})
					return true
				})

				// Each connection is counted twice (bidirectional)
				actualConnections := totalConnections / 2

				// With connection probability 0.3, we expect sparse connectivity
				maxPossible := (agentCount * (agentCount - 1)) / 2
				if actualConnections >= maxPossible {
					t.Error("Too many connections (should be sparse)")
				}

				if actualConnections == 0 {
					t.Error("No connections established")
				}
			},
		},
		{
			name: "single agent (no neighbors)",
			size: 1,
			validateFn: func(t *testing.T, swarm *Swarm) {
				connectionCount := 0
				swarm.agents.Range(func(key, value any) bool {
					agent := value.(*Agent)
					agent.neighbors.Range(func(k, v any) bool {
						connectionCount++
						return true
					})
					return true
				})

				if connectionCount != 0 {
					t.Error("Single agent should have no neighbors")
				}
			},
		},
		{
			name: "two agents connectivity",
			size: 2,
			validateFn: func(t *testing.T, swarm *Swarm) {
				// With 2 agents, they might or might not connect (probabilistic)
				connectionCount := 0
				swarm.agents.Range(func(key, value any) bool {
					agent := value.(*Agent)
					agent.neighbors.Range(func(k, v any) bool {
						connectionCount++
						return true
					})
					return true
				})

				// Either 0 or 2 connections (bidirectional)
				if connectionCount != 0 && connectionCount != 2 {
					t.Errorf("Two agents should have 0 or 2 connections, got %d", connectionCount)
				}
			},
		},
		{
			name: "verify bidirectional connections",
			size: 10,
			validateFn: func(t *testing.T, swarm *Swarm) {
				// Verify that if A is neighbor of B, then B is neighbor of A
				violations := 0
				swarm.agents.Range(func(key1, value1 any) bool {
					agent1 := value1.(*Agent)
					agent1.neighbors.Range(func(key2, value2 any) bool {
						agent2 := value2.(*Agent)
						// Check if agent1 is in agent2's neighbors
						found := false
						agent2.neighbors.Range(func(key3, value3 any) bool {
							if key3 == agent1.ID {
								found = true
								return false
							}
							return true
						})
						if !found {
							violations++
						}
						return true
					})
					return true
				})

				if violations > 0 {
					t.Errorf("Found %d non-bidirectional connections", violations)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goal := State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			}

			swarm, err := NewSwarm(tt.size, goal)
			if err != nil {
				t.Fatalf("Failed to create swarm: %v", err)
			}

			tt.validateFn(t, swarm)
		})
	}
}

func TestSwarmMeasureCoherence(t *testing.T) {
	tests := []struct {
		name        string
		size        int
		setupFn     func(swarm *Swarm)
		expectedMin float64
		expectedMax float64
		description string
	}{
		{
			name: "perfectly aligned agents",
			size: 10,
			setupFn: func(swarm *Swarm) {
				swarm.agents.Range(func(key, value any) bool {
					agent := value.(*Agent)
					agent.SetPhase(0)
					return true
				})
			},
			expectedMin: 0.99,
			expectedMax: 1.0,
			description: "All agents at phase 0 should have perfect coherence",
		},
		{
			name: "evenly distributed phases",
			size: 10,
			setupFn: func(swarm *Swarm) {
				i := 0
				swarm.agents.Range(func(key, value any) bool {
					agent := value.(*Agent)
					agent.SetPhase(float64(i) * 2 * math.Pi / 10)
					i++
					return true
				})
			},
			expectedMin: 0.0,
			expectedMax: 0.2,
			description: "Evenly distributed phases should have low coherence",
		},
		{
			name: "opposite phases",
			size: 10,
			setupFn: func(swarm *Swarm) {
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
			},
			expectedMin: 0.0,
			expectedMax: 0.1,
			description: "Half at 0, half at π should have near-zero coherence",
		},
		{
			name: "clustered phases",
			size: 12,
			setupFn: func(swarm *Swarm) {
				i := 0
				swarm.agents.Range(func(key, value any) bool {
					agent := value.(*Agent)
					// Three clusters at 0, 2π/3, 4π/3
					cluster := i / 4
					agent.SetPhase(float64(cluster) * 2 * math.Pi / 3)
					i++
					return true
				})
			},
			expectedMin: 0.0,
			expectedMax: 0.1,
			description: "Three equal clusters should cancel out",
		},
		{
			name: "slight spread around target",
			size: 10,
			setupFn: func(swarm *Swarm) {
				swarm.agents.Range(func(key, value any) bool {
					agent := value.(*Agent)
					// Small random variation around π
					agent.SetPhase(math.Pi + (rand.Float64()-0.5)*0.2)
					return true
				})
			},
			expectedMin: 0.9,
			expectedMax: 1.0,
			description: "Slight spread should maintain high coherence",
		},
		{
			name: "single agent",
			size: 1,
			setupFn: func(swarm *Swarm) {
				swarm.agents.Range(func(key, value any) bool {
					agent := value.(*Agent)
					agent.SetPhase(math.Pi / 2)
					return true
				})
			},
			expectedMin: 1.0,
			expectedMax: 1.0,
			description: "Single agent always has perfect coherence",
		},
		{
			name: "two agents aligned",
			size: 2,
			setupFn: func(swarm *Swarm) {
				swarm.agents.Range(func(key, value any) bool {
					agent := value.(*Agent)
					agent.SetPhase(math.Pi / 4)
					return true
				})
			},
			expectedMin: 1.0,
			expectedMax: 1.0,
			description: "Two aligned agents have perfect coherence",
		},
		{
			name: "two agents perpendicular",
			size: 2,
			setupFn: func(swarm *Swarm) {
				i := 0
				swarm.agents.Range(func(key, value any) bool {
					agent := value.(*Agent)
					if i == 0 {
						agent.SetPhase(0)
					} else {
						agent.SetPhase(math.Pi / 2)
					}
					i++
					return true
				})
			},
			expectedMin: 0.7,
			expectedMax: 0.72,
			description: "Two perpendicular agents have √2/2 coherence",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goal := State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			}

			swarm, err := NewSwarm(tt.size, goal)
			if err != nil {
				t.Fatalf("Failed to create swarm: %v", err)
			}

			tt.setupFn(swarm)

			coherence := swarm.MeasureCoherence()
			if coherence < tt.expectedMin || coherence > tt.expectedMax {
				t.Errorf("%s: coherence = %f, expected [%f, %f]",
					tt.description, coherence, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestSwarmGetAgent(t *testing.T) {
	tests := []struct {
		name       string
		swarmSize  int
		agentID    string
		wantExists bool
		validateFn func(t *testing.T, agent *Agent, exists bool)
	}{
		{
			name:       "get existing agent-0",
			swarmSize:  5,
			agentID:    "agent-0",
			wantExists: true,
			validateFn: func(t *testing.T, agent *Agent, exists bool) {
				if !exists {
					t.Error("Should find agent-0")
					return
				}
				if agent == nil {
					t.Error("Agent should not be nil")
					return
				}
				if agent.ID != "agent-0" {
					t.Errorf("Wrong agent returned: %s", agent.ID)
				}
			},
		},
		{
			name:       "get last agent",
			swarmSize:  5,
			agentID:    "agent-4",
			wantExists: true,
			validateFn: func(t *testing.T, agent *Agent, exists bool) {
				if !exists {
					t.Error("Should find agent-4")
				}
				if agent == nil {
					t.Error("Agent should not be nil")
				}
			},
		},
		{
			name:       "get non-existent agent",
			swarmSize:  5,
			agentID:    "non-existent",
			wantExists: false,
			validateFn: func(t *testing.T, agent *Agent, exists bool) {
				if exists {
					t.Error("Should not find non-existent agent")
				}
				if agent != nil {
					t.Error("Non-existent agent should be nil")
				}
			},
		},
		{
			name:       "get agent with empty ID",
			swarmSize:  5,
			agentID:    "",
			wantExists: false,
			validateFn: func(t *testing.T, agent *Agent, exists bool) {
				if exists {
					t.Error("Should not find agent with empty ID")
				}
			},
		},
		{
			name:       "get agent with out-of-range index",
			swarmSize:  5,
			agentID:    "agent-10",
			wantExists: false,
			validateFn: func(t *testing.T, agent *Agent, exists bool) {
				if exists {
					t.Error("Should not find agent-10 in swarm of size 5")
				}
			},
		},
		{
			name:       "get agent with negative index",
			swarmSize:  5,
			agentID:    "agent--1",
			wantExists: false,
			validateFn: func(t *testing.T, agent *Agent, exists bool) {
				if exists {
					t.Error("Should not find agent with negative index")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goal := State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			}

			swarm, err := NewSwarm(tt.swarmSize, goal)
			if err != nil {
				t.Fatalf("Failed to create swarm: %v", err)
			}

			agent, exists := swarm.GetAgent(tt.agentID)
			if exists != tt.wantExists {
				t.Errorf("GetAgent() exists = %v, want %v", exists, tt.wantExists)
			}
			tt.validateFn(t, agent, exists)
		})
	}
}

func TestSwarmDisruptAgents(t *testing.T) {
	tests := []struct {
		name             string
		size             int
		disruptionFactor float64
		setupFn          func(swarm *Swarm)
		validateFn       func(t *testing.T, swarm *Swarm, initialCoherence, finalCoherence float64)
	}{
		{
			name:             "50% disruption",
			size:             10,
			disruptionFactor: 0.5,
			setupFn: func(swarm *Swarm) {
				swarm.agents.Range(func(key, value any) bool {
					agent := value.(*Agent)
					agent.SetPhase(0)
					return true
				})
			},
			validateFn: func(t *testing.T, swarm *Swarm, initialCoherence, finalCoherence float64) {
				if finalCoherence >= initialCoherence {
					t.Error("Coherence should decrease after disruption")
				}

				// Count disrupted agents
				disrupted := 0
				swarm.agents.Range(func(key, value any) bool {
					agent := value.(*Agent)
					if math.Abs(agent.Phase()) > 0.01 {
						disrupted++
					}
					return true
				})

				// Should be approximately 5 agents (50% of 10)
				if disrupted < 3 || disrupted > 7 {
					t.Errorf("Expected ~5 disrupted agents, got %d", disrupted)
				}
			},
		},
		{
			name:             "0% disruption",
			size:             10,
			disruptionFactor: 0.0,
			setupFn: func(swarm *Swarm) {
				swarm.agents.Range(func(key, value any) bool {
					agent := value.(*Agent)
					agent.SetPhase(math.Pi / 2)
					return true
				})
			},
			validateFn: func(t *testing.T, swarm *Swarm, initialCoherence, finalCoherence float64) {
				if math.Abs(finalCoherence-initialCoherence) > 0.01 {
					t.Error("Coherence should not change with 0% disruption")
				}

				// No agents should be disrupted
				disrupted := 0
				swarm.agents.Range(func(key, value any) bool {
					agent := value.(*Agent)
					if math.Abs(agent.Phase()-math.Pi/2) > 0.01 {
						disrupted++
					}
					return true
				})

				if disrupted != 0 {
					t.Errorf("No agents should be disrupted, but %d were", disrupted)
				}
			},
		},
		{
			name:             "100% disruption",
			size:             10,
			disruptionFactor: 1.0,
			setupFn: func(swarm *Swarm) {
				swarm.agents.Range(func(key, value any) bool {
					agent := value.(*Agent)
					agent.SetPhase(0)
					return true
				})
			},
			validateFn: func(t *testing.T, swarm *Swarm, initialCoherence, finalCoherence float64) {
				if finalCoherence >= initialCoherence {
					t.Error("Coherence should decrease after 100% disruption")
				}

				// All agents should be disrupted
				disrupted := 0
				swarm.agents.Range(func(key, value any) bool {
					agent := value.(*Agent)
					if math.Abs(agent.Phase()) > 0.01 {
						disrupted++
					}
					return true
				})

				if disrupted != 10 {
					t.Errorf("All agents should be disrupted, but only %d were", disrupted)
				}
			},
		},
		{
			name:             "negative disruption factor (invalid)",
			size:             10,
			disruptionFactor: -0.5,
			setupFn: func(swarm *Swarm) {
				swarm.agents.Range(func(key, value any) bool {
					agent := value.(*Agent)
					agent.SetPhase(0)
					return true
				})
			},
			validateFn: func(t *testing.T, swarm *Swarm, initialCoherence, finalCoherence float64) {
				// Negative factor might be treated as 0 or absolute value
				// Just verify no panic occurs
				if swarm == nil {
					t.Error("Swarm should still be valid after negative disruption")
				}
			},
		},
		{
			name:             "disruption factor > 1 (invalid)",
			size:             10,
			disruptionFactor: 1.5,
			setupFn: func(swarm *Swarm) {
				swarm.agents.Range(func(key, value any) bool {
					agent := value.(*Agent)
					agent.SetPhase(0)
					return true
				})
			},
			validateFn: func(t *testing.T, swarm *Swarm, initialCoherence, finalCoherence float64) {
				// Factor > 1 might be clamped to 1
				// All or most agents should be disrupted
				disrupted := 0
				swarm.agents.Range(func(key, value any) bool {
					agent := value.(*Agent)
					if math.Abs(agent.Phase()) > 0.01 {
						disrupted++
					}
					return true
				})

				if disrupted < 8 {
					t.Errorf("Most agents should be disrupted with factor > 1, but only %d were", disrupted)
				}
			},
		},
		{
			name:             "single agent disruption",
			size:             1,
			disruptionFactor: 1.0,
			setupFn: func(swarm *Swarm) {
				swarm.agents.Range(func(key, value any) bool {
					agent := value.(*Agent)
					agent.SetPhase(0)
					return true
				})
			},
			validateFn: func(t *testing.T, swarm *Swarm, initialCoherence, finalCoherence float64) {
				// Single agent should still have coherence 1 after disruption
				if finalCoherence != 1.0 {
					t.Error("Single agent should maintain perfect coherence")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goal := State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			}

			swarm, err := NewSwarm(tt.size, goal)
			if err != nil {
				t.Fatalf("Failed to create swarm: %v", err)
			}

			tt.setupFn(swarm)
			initialCoherence := swarm.MeasureCoherence()

			swarm.DisruptAgents(tt.disruptionFactor)

			finalCoherence := swarm.MeasureCoherence()
			tt.validateFn(t, swarm, initialCoherence, finalCoherence)
		})
	}
}

func TestSwarmMonitor(t *testing.T) {
	tests := []struct {
		name       string
		setupFn    func(monitor *Monitor)
		validateFn func(t *testing.T, monitor *Monitor)
	}{
		{
			name: "basic monitoring",
			setupFn: func(monitor *Monitor) {
				monitor.RecordSample(0.5)
				monitor.RecordSample(0.6)
				monitor.RecordSample(0.7)
			},
			validateFn: func(t *testing.T, monitor *Monitor) {
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
			},
		},
		{
			name: "empty monitor",
			setupFn: func(monitor *Monitor) {
				// Don't add any samples
			},
			validateFn: func(t *testing.T, monitor *Monitor) {
				latest := monitor.GetLatest()
				if latest != 0 {
					t.Errorf("Empty monitor should have latest 0, got %f", latest)
				}

				avg := monitor.GetAverage()
				if avg != 0 {
					t.Errorf("Empty monitor should have average 0, got %f", avg)
				}

				history := monitor.GetHistory()
				if len(history) != 0 {
					t.Errorf("Empty monitor should have empty history, got %d", len(history))
				}
			},
		},
		{
			name: "single sample",
			setupFn: func(monitor *Monitor) {
				monitor.RecordSample(0.42)
			},
			validateFn: func(t *testing.T, monitor *Monitor) {
				latest := monitor.GetLatest()
				if latest != 0.42 {
					t.Errorf("Expected latest 0.42, got %f", latest)
				}

				avg := monitor.GetAverage()
				if avg != 0.42 {
					t.Errorf("Single sample average should be 0.42, got %f", avg)
				}
			},
		},
		{
			name: "many samples",
			setupFn: func(monitor *Monitor) {
				for i := 0; i < 100; i++ {
					monitor.RecordSample(float64(i) / 100)
				}
			},
			validateFn: func(t *testing.T, monitor *Monitor) {
				latest := monitor.GetLatest()
				if latest != 0.99 {
					t.Errorf("Expected latest 0.99, got %f", latest)
				}

				history := monitor.GetHistory()
				if len(history) != 100 {
					t.Errorf("Expected 100 samples in history, got %d", len(history))
				}

				// Average should be close to 0.495
				avg := monitor.GetAverage()
				if math.Abs(avg-0.495) > 0.01 {
					t.Errorf("Expected average ~0.495, got %f", avg)
				}
			},
		},
		{
			name: "negative values",
			setupFn: func(monitor *Monitor) {
				monitor.RecordSample(-0.5)
				monitor.RecordSample(0.5)
			},
			validateFn: func(t *testing.T, monitor *Monitor) {
				avg := monitor.GetAverage()
				if avg != 0 {
					t.Errorf("Average of -0.5 and 0.5 should be 0, got %f", avg)
				}
			},
		},
		{
			name: "values > 1",
			setupFn: func(monitor *Monitor) {
				monitor.RecordSample(1.5)
				monitor.RecordSample(2.5)
			},
			validateFn: func(t *testing.T, monitor *Monitor) {
				avg := monitor.GetAverage()
				if avg != 2.0 {
					t.Errorf("Average of 1.5 and 2.5 should be 2.0, got %f", avg)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goal := State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			}

			swarm, err := NewSwarm(10, goal)
			if err != nil {
				t.Fatalf("Failed to create swarm: %v", err)
			}

			monitor := swarm.GetMonitor()
			if monitor == nil {
				t.Fatal("Monitor should not be nil")
			}

			tt.setupFn(monitor)
			tt.validateFn(t, monitor)
		})
	}
}

func TestSwarmConvergence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping convergence test in short mode")
	}

	tests := []struct {
		name       string
		size       int
		goal       State
		setupFn    func(swarm *Swarm)
		timeout    time.Duration
		validateFn func(t *testing.T, swarm *Swarm, initialCoherence, finalCoherence float64)
	}{
		{
			name: "small swarm convergence",
			size: 5,
			goal: State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.7,
			},
			setupFn: func(swarm *Swarm) {
				i := 0
				swarm.agents.Range(func(key, value any) bool {
					agent := value.(*Agent)
					agent.SetPhase(float64(i) * 2 * math.Pi / 5)
					agent.SetLocalGoal(0)
					agent.SetStubbornness(0.01)
					agent.SetInfluence(0.8)
					i++
					return true
				})
			},
			timeout: 2 * time.Second,
			validateFn: func(t *testing.T, swarm *Swarm, initialCoherence, finalCoherence float64) {
				// Allow for some randomness
				if finalCoherence < initialCoherence-0.1 {
					t.Errorf("Coherence decreased from %f to %f", initialCoherence, finalCoherence)
				}
			},
		},
		{
			name: "medium swarm convergence",
			size: 20,
			goal: State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			setupFn: func(swarm *Swarm) {
				swarm.agents.Range(func(key, value any) bool {
					agent := value.(*Agent)
					agent.SetPhase(rand.Float64() * 2 * math.Pi)
					agent.SetLocalGoal(0)
					agent.SetStubbornness(0.05)
					agent.SetInfluence(0.7)
					return true
				})
			},
			timeout: 2 * time.Second,
			validateFn: func(t *testing.T, swarm *Swarm, initialCoherence, finalCoherence float64) {
				// Check for any improvement or at least no significant degradation
				if finalCoherence < initialCoherence-0.15 {
					t.Errorf("Coherence decreased significantly from %f to %f", initialCoherence, finalCoherence)
				}
			},
		},
		{
			name: "aligned initial conditions",
			size: 10,
			goal: State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			setupFn: func(swarm *Swarm) {
				swarm.agents.Range(func(key, value any) bool {
					agent := value.(*Agent)
					// Start close to goal
					agent.SetPhase(math.Pi + (rand.Float64()-0.5)*0.2)
					agent.SetLocalGoal(math.Pi)
					agent.SetStubbornness(0.02)
					agent.SetInfluence(0.8)
					return true
				})
			},
			timeout: 1 * time.Second,
			validateFn: func(t *testing.T, swarm *Swarm, initialCoherence, finalCoherence float64) {
				// Should maintain or improve high coherence
				if finalCoherence < 0.8 {
					t.Errorf("Failed to maintain high coherence, got %f", finalCoherence)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			swarm, err := NewSwarm(tt.size, tt.goal)
			if err != nil {
				t.Fatalf("Failed to create swarm: %v", err)
			}

			tt.setupFn(swarm)
			initialCoherence := swarm.MeasureCoherence()

			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			// Run swarm
			errChan := make(chan error, 1)
			go func() {
				if err := swarm.Run(ctx); err != nil {
					errChan <- err
				}
			}()

			// Wait for convergence or timeout
			time.Sleep(tt.timeout - 100*time.Millisecond)

			finalCoherence := swarm.MeasureCoherence()
			tt.validateFn(t, swarm, initialCoherence, finalCoherence)

			// Check monitor recorded samples
			if swarm.GetMonitor() != nil {
				history := swarm.GetMonitor().GetHistory()
				if len(history) < 5 {
					t.Errorf("Monitor should have recorded samples, got %d", len(history))
				}
			}
		})
	}
}

func TestSwarmConcurrency(t *testing.T) {
	goal := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	swarm, err := NewSwarm(50, goal)
	if err != nil {
		t.Fatalf("Failed to create swarm: %v", err)
	}

	// Concurrent operations
	var wg sync.WaitGroup

	// Measure coherence concurrently
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			_ = swarm.MeasureCoherence()
		}()
	}

	// Get agents concurrently
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer wg.Done()
			_, _ = swarm.GetAgent(fmt.Sprintf("agent-%d", id))
		}(i)
	}

	// Disrupt agents concurrently
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func() {
			defer wg.Done()
			swarm.DisruptAgents(0.1)
		}()
	}

	// Access monitor concurrently
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			monitor := swarm.GetMonitor()
			if monitor != nil {
				monitor.RecordSample(rand.Float64())
				_ = monitor.GetAverage()
			}
		}()
	}

	wg.Wait()

	// If we get here without panic, concurrent access is safe
	if swarm.Size() != 50 {
		t.Error("Swarm size should remain consistent after concurrent operations")
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
			swarm, err := NewSwarm(size, goal)
			if err != nil {
				b.Fatalf("Failed to create swarm: %v", err)
			}

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
				_, err := NewSwarm(size, goal)
				if err != nil {
					b.Fatalf("Failed to create swarm: %v", err)
				}
			}
		})
	}
}

func BenchmarkSwarmDisruption(b *testing.B) {
	goal := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	sizes := []int{10, 50, 100}
	factors := []float64{0.1, 0.5, 1.0}

	for _, size := range sizes {
		for _, factor := range factors {
			b.Run(fmt.Sprintf("size-%d-factor-%.1f", size, factor), func(b *testing.B) {
				swarm, err := NewSwarm(size, goal)
				if err != nil {
					b.Fatalf("Failed to create swarm: %v", err)
				}

				// Set all to same phase for consistent benchmarks
				swarm.agents.Range(func(key, value any) bool {
					agent := value.(*Agent)
					agent.SetPhase(0)
					return true
				})

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					swarm.DisruptAgents(factor)
					// Reset for next iteration
					swarm.agents.Range(func(key, value any) bool {
						agent := value.(*Agent)
						agent.SetPhase(0)
						return true
					})
				}
			})
		}
	}
}
