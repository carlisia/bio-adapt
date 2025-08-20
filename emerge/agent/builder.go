package agent

import (
	"fmt"
	"time"

	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/decision"
	"github.com/carlisia/bio-adapt/emerge/goal"
	"github.com/carlisia/bio-adapt/emerge/strategy"
	"github.com/carlisia/bio-adapt/internal/config"
	"github.com/carlisia/bio-adapt/internal/random"
	"github.com/carlisia/bio-adapt/internal/resource"
)

// NewFromConfig creates an agent from a configuration.
// This function converts the pure data config into actual component instances.
func NewFromConfig(id string, cfg config.Agent) (*Agent, error) {
	// Create core components
	dm, err := createDecisionMaker(cfg)
	if err != nil {
		return nil, err
	}

	gm, err := createGoalManager(cfg)
	if err != nil {
		return nil, err
	}

	rm, err := createResourceManager(cfg)
	if err != nil {
		return nil, err
	}

	sync, err := createStrategy(cfg)
	if err != nil {
		return nil, err
	}

	// Build options from config
	opts := []Option{
		WithDecisionMaker(dm),
		WithGoalManager(gm),
		WithResourceManager(rm),
		WithStrategy(sync),
	}

	// Add configuration-specific options
	opts = append(opts, createPhaseOptions(cfg)...)
	opts = append(opts, createGoalOptions(cfg)...)
	opts = append(opts, createFrequencyOptions(cfg)...)
	opts = append(opts, createParameterOptions(cfg)...)

	return New(id, opts...), nil
}

// createDecisionMaker creates a decision maker based on configuration
func createDecisionMaker(cfg config.Agent) (core.DecisionMaker, error) {
	switch cfg.DecisionMakerType {
	case "", "simple":
		return &decision.SimpleDecisionMaker{}, nil
	case "adaptive":
		// Future: could add adaptive decision maker
		return &decision.SimpleDecisionMaker{}, nil
	default:
		return nil, fmt.Errorf("unknown decision maker type: %s", cfg.DecisionMakerType)
	}
}

// createGoalManager creates a goal manager based on configuration
func createGoalManager(cfg config.Agent) (goal.Manager, error) {
	switch cfg.GoalManagerType {
	case "", "weighted":
		return &goal.WeightedManager{}, nil
	default:
		return nil, fmt.Errorf("unknown goal manager type: %s", cfg.GoalManagerType)
	}
}

// createResourceManager creates a resource manager based on configuration
func createResourceManager(cfg config.Agent) (core.ResourceManager, error) {
	switch cfg.ResourceManagerType {
	case "", "token":
		maxTokens := cfg.MaxTokens
		if maxTokens == 0 {
			maxTokens = cfg.InitialEnergy
			if maxTokens == 0 {
				maxTokens = 100.0
			}
		}
		return resource.NewTokenManager(maxTokens), nil
	default:
		return nil, fmt.Errorf("unknown resource manager type: %s", cfg.ResourceManagerType)
	}
}

// createStrategy creates a synchronization strategy based on configuration
func createStrategy(cfg config.Agent) (core.SyncStrategy, error) {
	switch cfg.StrategyType {
	case "", "phase_nudge":
		rate := cfg.StrategyRate
		if rate == 0 {
			rate = 0.7 // Increased default for better convergence
		}
		return &strategy.PhaseNudge{Rate: rate}, nil
	case "frequency_lock":
		rate := cfg.StrategyRate
		if rate == 0 {
			rate = 0.5
		}
		return &strategy.FrequencyLock{SyncRate: rate}, nil
	case "energy_aware":
		threshold := cfg.StrategyRate
		if threshold == 0 {
			threshold = 20.0
		}
		return &strategy.EnergyAware{Threshold: threshold}, nil
	case "adaptive":
		baseStrategies := []core.SyncStrategy{
			&strategy.PhaseNudge{Rate: 0.3},
			&strategy.FrequencyLock{SyncRate: 0.5},
		}
		return strategy.NewAdaptive(baseStrategies), nil
	case "pulse":
		return strategy.NewPulse(100*time.Millisecond, 0.8), nil
	default:
		return nil, fmt.Errorf("unknown strategy type: %s", cfg.StrategyType)
	}
}

// createPhaseOptions creates phase-related options
func createPhaseOptions(cfg config.Agent) []Option {
	switch {
	case cfg.RandomizePhase:
		return []Option{WithRandomPhase()}
	case cfg.Phase != 0:
		return []Option{WithPhase(cfg.Phase)}
	default:
		return []Option{WithPhase(0)}
	}
}

// createGoalOptions creates local goal-related options
func createGoalOptions(cfg config.Agent) []Option {
	switch {
	case cfg.RandomizeLocalGoal:
		return []Option{WithRandomLocalGoal()}
	case cfg.LocalGoal != 0:
		return []Option{WithLocalGoal(cfg.LocalGoal)}
	default:
		return []Option{WithLocalGoal(random.Phase())}
	}
}

// createFrequencyOptions creates frequency-related options
func createFrequencyOptions(cfg config.Agent) []Option {
	switch {
	case cfg.RandomizeFrequency:
		return []Option{WithRandomFrequency()}
	case cfg.Frequency != 0:
		return []Option{WithFrequency(cfg.Frequency)}
	default:
		return []Option{WithFrequency(100 * time.Millisecond)}
	}
}

// createParameterOptions creates parameter-related options
func createParameterOptions(cfg config.Agent) []Option {
	var opts []Option

	// Handle energy
	energy := cfg.InitialEnergy
	if energy == 0 {
		energy = 100.0
	}
	opts = append(opts, WithEnergy(energy))

	// Handle influence
	influence := cfg.Influence
	if influence == 0 {
		influence = 0.5
	}
	opts = append(opts, WithInfluence(influence))

	// Handle stubbornness
	stubbornness := cfg.Stubbornness
	if stubbornness == 0 {
		stubbornness = 0.2
	}
	opts = append(opts, WithStubbornness(stubbornness))

	// Handle swarm info
	if cfg.SwarmSize > 0 {
		opts = append(opts, WithSwarmInfo(cfg.SwarmSize, cfg.AssumedMaxNeighbors))
	}

	return opts
}

// TestAgentConfig returns a configuration suitable for testing.
// All randomization is disabled and values are predictable.
func TestAgentConfig() config.Agent {
	return config.TestAgent()
}
