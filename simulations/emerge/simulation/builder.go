// Package simulation - builder.go creates simulations from configurations.
// This file bridges the emerge client with the simulation implementation.
package simulation

import (
	"fmt"

	emerge "github.com/carlisia/bio-adapt/client/emerge"
	"github.com/carlisia/bio-adapt/emerge/goal"
	"github.com/carlisia/bio-adapt/emerge/scale"
	"github.com/carlisia/bio-adapt/simulations/emerge/simulation/pattern"
)

// BuildConfig is the configuration passed from main to create a simulation
type BuildConfig struct {
	Goal            goal.Type
	Scale           scale.Size
	Pattern         pattern.Type
	TargetCoherence float64
}

// New creates a simulation from a configuration.
// This is the main entry point for creating simulations.
func New(config BuildConfig) (*Simulation, error) {
	// Create emerge client based on configuration
	emergeClient, err := createEmergeClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create emerge client: %w", err)
	}

	// Create simulation with the emerge client and config
	sim := &Simulation{
		config:       convertToSimConfig(config),
		emergeClient: emergeClient,
		pattern:      config.Pattern,
		goalType:     config.Goal,
	}

	// Initialize simulation components
	sim.initialize()

	return sim, nil
}

// createEmergeClient creates an emerge client from configuration.
// This isolates emerge framework setup from simulation logic.
func createEmergeClient(config BuildConfig) (*emerge.Client, error) {
	// Use convenience methods when possible for clarity
	if config.Goal == goal.MinimizeAPICalls &&
		config.TargetCoherence == config.Scale.DefaultTargetCoherence() {
		return emerge.MinimizeAPICalls(config.Scale)
	}

	// Use custom builder for non-default configurations
	return emerge.New().
		WithGoal(config.Goal).
		WithScale(config.Scale).
		WithTargetCoherence(config.TargetCoherence).
		Build()
}

// convertToSimConfig converts the build config to the internal Config type.
// This maintains compatibility with existing simulation code.
func convertToSimConfig(config BuildConfig) Config {
	return Config{
		NumAgents:       config.Scale.DefaultAgentCount(),
		TargetCoherence: config.TargetCoherence,
	}
}

// initialize sets up the simulation components.
// This is now a method on Simulation for better encapsulation.
func (s *Simulation) initialize() {
	// Get actual agent count from scale
	actualAgentCount := s.config.NumAgents

	// Get the swarm agents and create workloads
	swarmAgents := s.emergeClient.Agents()
	workloads := make([]*Workload, 0, actualAgentCount)
	i := 0
	for id, emergeAgent := range swarmAgents {
		if i >= actualAgentCount {
			break
		}
		// Create a workload that wraps the emerge agent with pattern-specific behavior
		workloads = append(workloads, NewWorkloadWithEmerge(id, i, emergeAgent, s, s.pattern, s.goalType))
		i++
	}
	s.workloads = workloads

	// Create simulation-specific components
	s.batch = NewBatchManager()
	s.metrics = NewMetricsCollector()
}
