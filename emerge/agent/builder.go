package agent

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/decision"
	"github.com/carlisia/bio-adapt/emerge/goal"
	"github.com/carlisia/bio-adapt/emerge/strategy"
	"github.com/carlisia/bio-adapt/internal/config"
	"github.com/carlisia/bio-adapt/internal/resource"
)

// NewFromConfig creates an agent from a configuration.
// This function converts the pure data config into actual component instances.
func NewFromConfig(id string, cfg config.Agent) (*Agent, error) {
	// Create decision maker based on type
	var dm core.DecisionMaker
	switch cfg.DecisionMakerType {
	case "", "simple":
		dm = &decision.SimpleDecisionMaker{}
	case "adaptive":
		// Future: could add adaptive decision maker
		dm = &decision.SimpleDecisionMaker{}
	default:
		return nil, fmt.Errorf("unknown decision maker type: %s", cfg.DecisionMakerType)
	}

	// Create goal manager based on type
	var gm goal.Manager
	switch cfg.GoalManagerType {
	case "", "weighted":
		gm = &goal.WeightedManager{}
	default:
		return nil, fmt.Errorf("unknown goal manager type: %s", cfg.GoalManagerType)
	}

	// Create resource manager based on type
	var rm core.ResourceManager
	switch cfg.ResourceManagerType {
	case "", "token":
		maxTokens := cfg.MaxTokens
		if maxTokens == 0 {
			maxTokens = cfg.InitialEnergy
			if maxTokens == 0 {
				maxTokens = 100.0
			}
		}
		rm = resource.NewTokenManager(maxTokens)
	default:
		return nil, fmt.Errorf("unknown resource manager type: %s", cfg.ResourceManagerType)
	}

	// Create strategy based on type
	var sync core.SyncStrategy
	switch cfg.StrategyType {
	case "", "phase_nudge":
		rate := cfg.StrategyRate
		if rate == 0 {
			rate = 0.7 // Increased default for better convergence
		}
		sync = &strategy.PhaseNudge{Rate: rate}
	case "frequency_lock":
		rate := cfg.StrategyRate
		if rate == 0 {
			rate = 0.5
		}
		sync = &strategy.FrequencyLock{SyncRate: rate}
	case "energy_aware":
		threshold := cfg.StrategyRate
		if threshold == 0 {
			threshold = 20.0
		}
		sync = &strategy.EnergyAware{Threshold: threshold}
	case "adaptive":
		// Create base strategies for adaptive
		baseStrategies := []core.SyncStrategy{
			&strategy.PhaseNudge{Rate: 0.3},
			&strategy.FrequencyLock{SyncRate: 0.5},
		}
		sync = strategy.NewAdaptive(baseStrategies)
	case "pulse":
		sync = strategy.NewPulse(100*time.Millisecond, 0.8)
	default:
		return nil, fmt.Errorf("unknown strategy type: %s", cfg.StrategyType)
	}

	// Build options from config
	opts := []Option{
		WithDecisionMaker(dm),
		WithGoalManager(gm),
		WithResourceManager(rm),
		WithStrategy(sync),
	}

	// Handle phase
	if cfg.RandomizePhase {
		opts = append(opts, WithRandomPhase())
	} else if cfg.Phase != 0 {
		opts = append(opts, WithPhase(cfg.Phase))
	} else {
		opts = append(opts, WithPhase(0))
	}

	// Handle local goal
	if cfg.RandomizeLocalGoal {
		opts = append(opts, WithRandomLocalGoal())
	} else if cfg.LocalGoal != 0 {
		opts = append(opts, WithLocalGoal(cfg.LocalGoal))
	} else {
		opts = append(opts, WithLocalGoal(rand.Float64()*2*math.Pi))
	}

	// Handle frequency
	if cfg.RandomizeFrequency {
		opts = append(opts, WithRandomFrequency())
	} else if cfg.Frequency != 0 {
		opts = append(opts, WithFrequency(cfg.Frequency))
	} else {
		opts = append(opts, WithFrequency(100*time.Millisecond))
	}

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

	return New(id, opts...), nil
}

// TestAgentConfig returns a configuration suitable for testing.
// All randomization is disabled and values are predictable.
func TestAgentConfig() config.Agent {
	return config.TestAgent()
}
