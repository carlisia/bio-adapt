package config

import (
	"math"
	"time"
)

// Agent holds all configuration for creating an Agent.
// Zero values are replaced with sensible defaults.
// This is a pure data structure with no interface dependencies.
type Agent struct {
	// Component selection (string identifiers)
	DecisionMakerType   string // e.g., "simple", "adaptive"
	GoalManagerType     string // e.g., "weighted"
	ResourceManagerType string // e.g., "token"
	StrategyType        string // e.g., "phase_nudge", "frequency_lock", "adaptive"

	// Component parameters
	MaxTokens        float64 // For token resource manager
	StrategyRate     float64 // For phase nudge and other strategies
	AdaptiveStrength float64 // For adaptive strategy

	// Initial state
	Phase         float64       // Initial phase (0 = random)
	LocalGoal     float64       // Local goal phase (0 = random)
	Frequency     time.Duration // Oscillation frequency (0 = random 50-150ms)
	InitialEnergy float64       // Starting energy (0 = 100.0)

	// Behavior parameters
	Influence    float64 // Weight of local vs global goals [0,1] (0 = 0.5)
	Stubbornness float64 // Resistance to change [0,1] (0 = 0.2)

	// Swarm context
	SwarmSize           int // Total swarm size for density calculations
	AssumedMaxNeighbors int // For density calculations (0 = use swarm size)

	// Randomization flags
	RandomizePhase     bool // If true, ignore Phase value and randomize
	RandomizeLocalGoal bool // If true, ignore LocalGoal value and randomize
	RandomizeFrequency bool // If true, ignore Frequency value and randomize
}

// DefaultAgent returns a configuration with sensible defaults.
// All values are deterministic (no randomization) for predictability.
func DefaultAgent() Agent {
	return Agent{
		DecisionMakerType:   "simple",
		GoalManagerType:     "weighted",
		ResourceManagerType: "token",
		StrategyType:        "phase_nudge",
		MaxTokens:           100,
		StrategyRate:        0.3,

		Phase:         0,                      // Will be at phase 0
		LocalGoal:     math.Pi,                // Opposite of typical target (0)
		Frequency:     100 * time.Millisecond, // Middle of range
		InitialEnergy: 100.0,

		Influence:    0.5,
		Stubbornness: 0.2,

		RandomizePhase:     false,
		RandomizeLocalGoal: false,
		RandomizeFrequency: false,
	}
}

// RandomizedAgent returns a configuration with randomized initial conditions.
// This is useful for creating diverse swarms.
func RandomizedAgent() Agent {
	config := DefaultAgent()
	config.RandomizePhase = true
	config.RandomizeLocalGoal = true
	config.RandomizeFrequency = true
	return config
}

// TestAgent returns a configuration suitable for testing.
// All randomization is disabled and values are predictable.
func TestAgent() Agent {
	return Agent{
		DecisionMakerType:   "simple",
		GoalManagerType:     "weighted",
		ResourceManagerType: "token",
		StrategyType:        "phase_nudge",
		MaxTokens:           100,
		StrategyRate:        0.5, // Higher for faster convergence in tests

		Phase:         0,
		LocalGoal:     0,                     // Same as global goal for predictability
		Frequency:     50 * time.Millisecond, // Fast for quick tests
		InitialEnergy: 100.0,

		Influence:    0.8,  // High influence for faster convergence
		Stubbornness: 0.01, // Very low for predictable behavior

		RandomizePhase:     false,
		RandomizeLocalGoal: false,
		RandomizeFrequency: false,
	}
}

// AgentFromSwarm creates an agent config based on swarm configuration.
// This ensures agents are configured consistently with swarm parameters.
func AgentFromSwarm(sc Swarm) Agent {
	return Agent{
		DecisionMakerType:   "simple",
		GoalManagerType:     "weighted",
		ResourceManagerType: "token",
		StrategyType:        "phase_nudge",
		MaxTokens:           sc.InitialEnergy,
		StrategyRate:        sc.CouplingStrength,
		InitialEnergy:       sc.InitialEnergy,
		Influence:           sc.InfluenceDefault,
		Stubbornness:        sc.Stubbornness,
		AssumedMaxNeighbors: sc.AssumedMaxNeighbors,
		RandomizePhase:      true,
		RandomizeLocalGoal:  true,
		RandomizeFrequency:  true,
	}
}
