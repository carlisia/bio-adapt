package emerge

import (
	"math"
	"math/rand"
	"time"
)

// AgentConfig holds all configuration for creating an Agent.
// Zero values are replaced with sensible defaults.
type AgentConfig struct {
	// Core components (nil values get defaults)
	DecisionMaker   DecisionMaker
	GoalManager     GoalManager
	ResourceManager ResourceManager
	Strategy        SyncStrategy

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

// DefaultAgentConfig returns a configuration with sensible defaults.
// All values are deterministic (no randomization) for predictability.
func DefaultAgentConfig() AgentConfig {
	return AgentConfig{
		DecisionMaker:   &SimpleDecisionMaker{},
		GoalManager:     &WeightedGoalManager{},
		ResourceManager: NewTokenResourceManager(100),
		Strategy:        NewPhaseNudgeStrategy(0.3),

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

// RandomizedAgentConfig returns a configuration with randomized initial conditions.
// This is useful for creating diverse swarms.
func RandomizedAgentConfig() AgentConfig {
	config := DefaultAgentConfig()
	config.RandomizePhase = true
	config.RandomizeLocalGoal = true
	config.RandomizeFrequency = true
	return config
}

// TestAgentConfig returns a configuration suitable for testing.
// All randomization is disabled and values are predictable.
func TestAgentConfig() AgentConfig {
	return AgentConfig{
		DecisionMaker:   &SimpleDecisionMaker{},
		GoalManager:     &WeightedGoalManager{},
		ResourceManager: NewTokenResourceManager(100),
		Strategy:        NewPhaseNudgeStrategy(0.5), // Higher for faster convergence in tests

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

// NewAgentFromConfig creates an agent from a configuration.
// Zero values in the config are replaced with defaults.
func NewAgentFromConfig(id string, cfg AgentConfig) *Agent {
	// Apply defaults for zero values
	if cfg.DecisionMaker == nil {
		cfg.DecisionMaker = &SimpleDecisionMaker{}
	}
	if cfg.GoalManager == nil {
		cfg.GoalManager = &WeightedGoalManager{}
	}
	if cfg.ResourceManager == nil {
		energy := cfg.InitialEnergy
		if energy == 0 {
			energy = 100.0
		}
		cfg.ResourceManager = NewTokenResourceManager(energy)
	}
	if cfg.Strategy == nil {
		cfg.Strategy = NewPhaseNudgeStrategy(0.3)
	}

	// Build options from config
	opts := []AgentOption{
		WithDecisionMaker(cfg.DecisionMaker),
		WithGoalManager(cfg.GoalManager),
		WithResourceManager(cfg.ResourceManager),
		WithStrategy(cfg.Strategy),
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

	return NewAgent(id, opts...)
}

// AgentConfigFromSwarmConfig creates an agent config based on swarm configuration.
// This ensures agents are configured consistently with swarm parameters.
func AgentConfigFromSwarmConfig(sc SwarmConfig) AgentConfig {
	return AgentConfig{
		DecisionMaker:       &SimpleDecisionMaker{},
		GoalManager:         &WeightedGoalManager{},
		ResourceManager:     NewTokenResourceManager(sc.InitialEnergy),
		Strategy:            NewPhaseNudgeStrategy(sc.CouplingStrength),
		InitialEnergy:       sc.InitialEnergy,
		Influence:           sc.InfluenceDefault,
		Stubbornness:        sc.Stubbornness,
		AssumedMaxNeighbors: sc.AssumedMaxNeighbors,
		RandomizePhase:      true,
		RandomizeLocalGoal:  true,
		RandomizeFrequency:  true,
	}
}
