package swarm

import (
	"fmt"
	"time"

	"github.com/carlisia/bio-adapt/emerge/goal"
	"github.com/carlisia/bio-adapt/emerge/scale"
	"github.com/carlisia/bio-adapt/emerge/trait"
)

// Config holds all configuration for goal-directed synchronization.
type Config struct {
	Convergence ConvergenceConfig
	Thresholds  ThresholdConfig
	Variation   VariationConfig
	Strategy    StrategyConfig
	Resonance   ResonanceConfig
}

// ConvergenceConfig controls how the swarm converges to targets.
type ConvergenceConfig struct {
	// Size-based tolerances for achieving target coherence
	ToleranceSmall  float64 // Default: 0.015 (swarms < 10)
	ToleranceMedium float64 // Default: 0.008 (swarms < 50)
	ToleranceLarge  float64 // Default: 0.005 (swarms >= 50)

	// How close phases need to converge (0-1)
	PhaseConvergenceGoal float64 // Default: 0.85

	// How close pattern needs to match target
	PatternDistanceThreshold float64 // Default: 0.1

	// Base aggressiveness of adjustments
	BaseAdjustmentScale float64 // Default: 0.65
}

// ThresholdConfig defines trigger points for different behaviors.
type ThresholdConfig struct {
	// Triggers special phase alignment mode
	PhaseVariance float64 // Default: 0.5

	// Different coherence levels trigger different behaviors
	HighCoherence     float64 // Default: 0.9
	VeryHighCoherence float64 // Default: 0.92
	ModerateCoherence float64 // Default: 0.85
}

// VariationConfig controls randomness and diversity.
type VariationConfig struct {
	// Prevents perfect synchronization (min, max)
	BaseRange [2]float64 // Default: [0.15, 0.30]

	// Scale variation with coherence level
	CoherenceFactor float64 // Default: 0.2

	// Random perturbations for diversity
	RandomWalkMagnitude   float64 // Default: 0.1
	PerturbationMagnitude float64 // Default: 0.08
	PerturbationChance    float64 // Default: 0.15
}

// StrategyConfig controls strategy selection and timing.
type StrategyConfig struct {
	// Convergence time control
	MaxIterationsFactor float64       // Default: 200
	UpdateInterval      time.Duration // Default: 100ms

	// Strategy selection parameters
	ExplorationBonusMax   float64       // Default: 0.3
	ExplorationTimeWindow time.Duration // Default: 60s
	RandomExploration     float64       // Default: 0.1
}

// ResonanceConfig controls stochastic resonance for escaping local minima.
type ResonanceConfig struct {
	// Stochastic resonance to escape local minima
	NoiseMagnitude float64 // Default: 0.5 (±0.25 radians)
	AffectedAgents float64 // Default: 0.1 (10% of swarm)
	ActivationRate float64 // Default: 0.1 (10% chance when stuck)
}

// For returns a configuration optimized for a specific goal.
func For(g goal.Type) *Config {
	switch g {
	case goal.MinimizeAPICalls:
		return &Config{
			Convergence: ConvergenceConfig{
				ToleranceSmall:           0.01,
				ToleranceMedium:          0.005,
				ToleranceLarge:           0.003,
				PhaseConvergenceGoal:     0.90,
				PatternDistanceThreshold: 0.05,
				BaseAdjustmentScale:      0.75,
			},
			Thresholds: ThresholdConfig{
				PhaseVariance:     0.3,
				HighCoherence:     0.85,
				VeryHighCoherence: 0.90,
				ModerateCoherence: 0.80,
			},
			Variation: VariationConfig{
				BaseRange:             [2]float64{0.05, 0.15},
				CoherenceFactor:       0.15,
				RandomWalkMagnitude:   0.05,
				PerturbationMagnitude: 0.04,
				PerturbationChance:    0.10,
			},
			Strategy: StrategyConfig{
				MaxIterationsFactor:   150,
				UpdateInterval:        200 * time.Millisecond,
				ExplorationBonusMax:   0.2,
				ExplorationTimeWindow: 45 * time.Second,
				RandomExploration:     0.05,
			},
			Resonance: ResonanceConfig{
				NoiseMagnitude: 0.3,
				AffectedAgents: 0.05,
				ActivationRate: 0.05,
			},
		}

	case goal.DistributeLoad:
		return &Config{
			Convergence: ConvergenceConfig{
				ToleranceSmall:           0.025,
				ToleranceMedium:          0.015,
				ToleranceLarge:           0.010,
				PhaseConvergenceGoal:     0.30, // Anti-phase
				PatternDistanceThreshold: 0.2,
				BaseAdjustmentScale:      0.55,
			},
			Thresholds: ThresholdConfig{
				PhaseVariance:     0.8, // High variance is good
				HighCoherence:     0.4, // Lower targets
				VeryHighCoherence: 0.5,
				ModerateCoherence: 0.3,
			},
			Variation: VariationConfig{
				BaseRange:             [2]float64{0.25, 0.45},
				CoherenceFactor:       0.3,
				RandomWalkMagnitude:   0.2,
				PerturbationMagnitude: 0.15,
				PerturbationChance:    0.25,
			},
			Strategy: StrategyConfig{
				MaxIterationsFactor:   100,
				UpdateInterval:        50 * time.Millisecond,
				ExplorationBonusMax:   0.4,
				ExplorationTimeWindow: 30 * time.Second,
				RandomExploration:     0.15,
			},
			Resonance: ResonanceConfig{
				NoiseMagnitude: 0.6,
				AffectedAgents: 0.15,
				ActivationRate: 0.15,
			},
		}

	case goal.ReachConsensus:
		return &Config{
			Convergence: ConvergenceConfig{
				ToleranceSmall:           0.008,
				ToleranceMedium:          0.004,
				ToleranceLarge:           0.002,
				PhaseConvergenceGoal:     0.95,
				PatternDistanceThreshold: 0.03,
				BaseAdjustmentScale:      0.80,
			},
			Thresholds: ThresholdConfig{
				PhaseVariance:     0.2,
				HighCoherence:     0.90,
				VeryHighCoherence: 0.95,
				ModerateCoherence: 0.85,
			},
			Variation: VariationConfig{
				BaseRange:             [2]float64{0.03, 0.10},
				CoherenceFactor:       0.10,
				RandomWalkMagnitude:   0.03,
				PerturbationMagnitude: 0.02,
				PerturbationChance:    0.05,
			},
			Strategy: StrategyConfig{
				MaxIterationsFactor:   250,
				UpdateInterval:        100 * time.Millisecond,
				ExplorationBonusMax:   0.25,
				ExplorationTimeWindow: 60 * time.Second,
				RandomExploration:     0.08,
			},
			Resonance: ResonanceConfig{
				NoiseMagnitude: 0.2,
				AffectedAgents: 0.03,
				ActivationRate: 0.03,
			},
		}

	case goal.MinimizeLatency:
		return &Config{
			Convergence: ConvergenceConfig{
				ToleranceSmall:           0.020,
				ToleranceMedium:          0.012,
				ToleranceLarge:           0.008,
				PhaseConvergenceGoal:     0.75,
				PatternDistanceThreshold: 0.15,
				BaseAdjustmentScale:      0.85,
			},
			Thresholds: ThresholdConfig{
				PhaseVariance:     0.4,
				HighCoherence:     0.80,
				VeryHighCoherence: 0.85,
				ModerateCoherence: 0.75,
			},
			Variation: VariationConfig{
				BaseRange:             [2]float64{0.10, 0.20},
				CoherenceFactor:       0.15,
				RandomWalkMagnitude:   0.08,
				PerturbationMagnitude: 0.06,
				PerturbationChance:    0.12,
			},
			Strategy: StrategyConfig{
				MaxIterationsFactor:   80,
				UpdateInterval:        25 * time.Millisecond,
				ExplorationBonusMax:   0.35,
				ExplorationTimeWindow: 20 * time.Second,
				RandomExploration:     0.12,
			},
			Resonance: ResonanceConfig{
				NoiseMagnitude: 0.4,
				AffectedAgents: 0.08,
				ActivationRate: 0.10,
			},
		}

	case goal.SaveEnergy:
		return &Config{
			Convergence: ConvergenceConfig{
				ToleranceSmall:           0.030,
				ToleranceMedium:          0.020,
				ToleranceLarge:           0.015,
				PhaseConvergenceGoal:     0.70,
				PatternDistanceThreshold: 0.20,
				BaseAdjustmentScale:      0.40,
			},
			Thresholds: ThresholdConfig{
				PhaseVariance:     0.6,
				HighCoherence:     0.75,
				VeryHighCoherence: 0.80,
				ModerateCoherence: 0.70,
			},
			Variation: VariationConfig{
				BaseRange:             [2]float64{0.08, 0.18},
				CoherenceFactor:       0.12,
				RandomWalkMagnitude:   0.06,
				PerturbationMagnitude: 0.05,
				PerturbationChance:    0.08,
			},
			Strategy: StrategyConfig{
				MaxIterationsFactor:   300,
				UpdateInterval:        250 * time.Millisecond,
				ExplorationBonusMax:   0.15,
				ExplorationTimeWindow: 90 * time.Second,
				RandomExploration:     0.05,
			},
			Resonance: ResonanceConfig{
				NoiseMagnitude: 0.25,
				AffectedAgents: 0.05,
				ActivationRate: 0.05,
			},
		}

	case goal.MaintainRhythm:
		return &Config{
			Convergence: ConvergenceConfig{
				ToleranceSmall:           0.012,
				ToleranceMedium:          0.006,
				ToleranceLarge:           0.003,
				PhaseConvergenceGoal:     0.80,
				PatternDistanceThreshold: 0.08,
				BaseAdjustmentScale:      0.70,
			},
			Thresholds: ThresholdConfig{
				PhaseVariance:     0.4,
				HighCoherence:     0.85,
				VeryHighCoherence: 0.90,
				ModerateCoherence: 0.80,
			},
			Variation: VariationConfig{
				BaseRange:             [2]float64{0.08, 0.16},
				CoherenceFactor:       0.15,
				RandomWalkMagnitude:   0.06,
				PerturbationMagnitude: 0.05,
				PerturbationChance:    0.10,
			},
			Strategy: StrategyConfig{
				MaxIterationsFactor:   180,
				UpdateInterval:        100 * time.Millisecond,
				ExplorationBonusMax:   0.25,
				ExplorationTimeWindow: 50 * time.Second,
				RandomExploration:     0.08,
			},
			Resonance: ResonanceConfig{
				NoiseMagnitude: 0.3,
				AffectedAgents: 0.06,
				ActivationRate: 0.08,
			},
		}

	case goal.RecoverFromFailure:
		return &Config{
			Convergence: ConvergenceConfig{
				ToleranceSmall:           0.025,
				ToleranceMedium:          0.015,
				ToleranceLarge:           0.010,
				PhaseConvergenceGoal:     0.65,
				PatternDistanceThreshold: 0.18,
				BaseAdjustmentScale:      0.60,
			},
			Thresholds: ThresholdConfig{
				PhaseVariance:     0.7,
				HighCoherence:     0.70,
				VeryHighCoherence: 0.75,
				ModerateCoherence: 0.65,
			},
			Variation: VariationConfig{
				BaseRange:             [2]float64{0.20, 0.40},
				CoherenceFactor:       0.25,
				RandomWalkMagnitude:   0.15,
				PerturbationMagnitude: 0.12,
				PerturbationChance:    0.20,
			},
			Strategy: StrategyConfig{
				MaxIterationsFactor:   350,
				UpdateInterval:        75 * time.Millisecond,
				ExplorationBonusMax:   0.40,
				ExplorationTimeWindow: 30 * time.Second,
				RandomExploration:     0.15,
			},
			Resonance: ResonanceConfig{
				NoiseMagnitude: 0.7,
				AffectedAgents: 0.15,
				ActivationRate: 0.20,
			},
		}

	case goal.AdaptToTraffic:
		return &Config{
			Convergence: ConvergenceConfig{
				ToleranceSmall:           0.040, // Looser tolerance for quick adaptation
				ToleranceMedium:          0.030,
				ToleranceLarge:           0.020,
				PhaseConvergenceGoal:     0.50, // Lower goal allows more flexibility
				PatternDistanceThreshold: 0.25, // Accept larger pattern differences
				BaseAdjustmentScale:      0.45, // Moderate adjustments
			},
			Thresholds: ThresholdConfig{
				PhaseVariance:     0.9,  // High variance is OK for traffic routing
				HighCoherence:     0.60, // Lower coherence thresholds
				VeryHighCoherence: 0.70,
				ModerateCoherence: 0.50,
			},
			Variation: VariationConfig{
				BaseRange:             [2]float64{0.30, 0.50}, // High variation for responsiveness
				CoherenceFactor:       0.35,
				RandomWalkMagnitude:   0.25, // More random walk allows quick changes
				PerturbationMagnitude: 0.20,
				PerturbationChance:    0.30, // Frequent perturbations
			},
			Strategy: StrategyConfig{
				MaxIterationsFactor:   100,                   // Shorter iterations for quick response
				UpdateInterval:        50 * time.Millisecond, // Fast updates
				ExplorationBonusMax:   0.50,
				ExplorationTimeWindow: 20 * time.Second,
				RandomExploration:     0.25, // More exploration
			},
			Resonance: ResonanceConfig{
				NoiseMagnitude: 0.8,  // High noise for quick escape from patterns
				AffectedAgents: 0.20, // More agents affected
				ActivationRate: 0.25, // Frequent activation
			},
		}

	default:
		// Default balanced configuration
		return defaultConfig()
	}
}

// TuneFor adjusts the configuration for a specific optimization target.
func (c *Config) TuneFor(t trait.Target) *Config {
	switch t {
	case trait.Stability:
		// Make adjustments gentler and more predictable
		c.Convergence.BaseAdjustmentScale *= 0.6
		c.Variation.BaseRange[0] *= 1.2
		c.Variation.BaseRange[1] *= 1.3
		c.Variation.PerturbationMagnitude *= 0.7
		c.Strategy.UpdateInterval = time.Duration(float64(c.Strategy.UpdateInterval) * 1.5)
		c.Resonance.NoiseMagnitude *= 0.6

	case trait.Speed:
		// Make convergence faster
		c.Convergence.BaseAdjustmentScale = minValue(1.0, c.Convergence.BaseAdjustmentScale*1.3)
		c.Convergence.ToleranceSmall *= 1.5
		c.Convergence.ToleranceMedium *= 1.5
		c.Convergence.ToleranceLarge *= 1.5
		c.Strategy.UpdateInterval = time.Duration(float64(c.Strategy.UpdateInterval) * 0.5)
		c.Strategy.MaxIterationsFactor *= 0.7

	case trait.Efficiency:
		// Reduce resource usage
		c.Strategy.UpdateInterval = time.Duration(float64(c.Strategy.UpdateInterval) * 2.0)
		c.Variation.PerturbationChance *= 0.5
		c.Resonance.AffectedAgents *= 0.5
		c.Strategy.ExplorationBonusMax *= 0.7

	case trait.Throughput:
		// Optimize for maximum processing
		c.Convergence.BaseAdjustmentScale *= 1.2
		c.Strategy.UpdateInterval = time.Duration(float64(c.Strategy.UpdateInterval) * 0.7)
		c.Strategy.RandomExploration *= 1.5

	case trait.Resilience:
		// Add redundancy and fault tolerance
		c.Variation.BaseRange[0] *= 1.3
		c.Variation.BaseRange[1] *= 1.4
		c.Resonance.NoiseMagnitude *= 1.5
		c.Resonance.AffectedAgents *= 1.5
		c.Resonance.ActivationRate *= 1.8

	case trait.Precision:
		// Tighten tolerances for accuracy
		c.Convergence.ToleranceSmall *= 0.5
		c.Convergence.ToleranceMedium *= 0.5
		c.Convergence.ToleranceLarge *= 0.5
		c.Convergence.PatternDistanceThreshold *= 0.5
		c.Convergence.PhaseConvergenceGoal = minValue(0.98, c.Convergence.PhaseConvergenceGoal*1.1)
	}

	return c
}

// With adjusts the configuration for a specific swarm scale.
func (c *Config) With(s scale.Size) *Config {
	switch s {
	case scale.Tiny:
		// Very small swarms need tight coupling
		c.Convergence.ToleranceSmall = 0.020
		c.Convergence.ToleranceMedium = 0.020
		c.Convergence.ToleranceLarge = 0.020
		c.Variation.BaseRange = [2]float64{0.05, 0.10}
		c.Strategy.UpdateInterval = 20 * time.Millisecond

	case scale.Small:
		// Small swarms can converge quickly
		c.Convergence.ToleranceSmall = 0.015
		c.Convergence.ToleranceMedium = 0.015
		c.Convergence.ToleranceLarge = 0.015
		c.Variation.BaseRange[0] *= 0.8
		c.Variation.BaseRange[1] *= 0.9
		c.Strategy.UpdateInterval = time.Duration(minValue(50, int(c.Strategy.UpdateInterval/time.Millisecond))) * time.Millisecond

	case scale.Medium:
		// Medium swarms use default tolerances
		// (no changes needed)

	case scale.Large:
		// Large swarms need more tolerance
		c.Convergence.ToleranceSmall *= 1.5
		c.Convergence.ToleranceMedium *= 1.3
		c.Convergence.ToleranceLarge *= 1.2
		c.Variation.BaseRange[0] *= 1.2
		c.Variation.BaseRange[1] *= 1.3
		c.Strategy.UpdateInterval = time.Duration(maxValue(150, int(c.Strategy.UpdateInterval/time.Millisecond))) * time.Millisecond
		c.Strategy.MaxIterationsFactor *= 1.5

	case scale.Huge:
		// Huge swarms need significant adjustments
		c.Convergence.ToleranceSmall *= 2.0
		c.Convergence.ToleranceMedium *= 1.8
		c.Convergence.ToleranceLarge *= 1.5
		c.Variation.BaseRange[0] *= 1.5
		c.Variation.BaseRange[1] *= 1.8
		c.Strategy.UpdateInterval = time.Duration(maxValue(200, int(c.Strategy.UpdateInterval/time.Millisecond))) * time.Millisecond
		c.Strategy.MaxIterationsFactor *= 2.0
		c.Resonance.AffectedAgents = minValue(0.05, c.Resonance.AffectedAgents)
	}

	return c
}

// Validate ensures all configuration parameters are within valid ranges.
func (c *Config) Validate() error {
	if err := c.validateConvergence(); err != nil {
		return err
	}
	if err := c.validateThresholds(); err != nil {
		return err
	}
	if err := c.validateVariation(); err != nil {
		return err
	}
	if err := c.validateStrategy(); err != nil {
		return err
	}
	return c.validateResonance()
}

func (c *Config) validateConvergence() error {
	if c.Convergence.BaseAdjustmentScale < 0 || c.Convergence.BaseAdjustmentScale > 1 {
		return fmt.Errorf("BaseAdjustmentScale must be between 0 and 1, got %f", c.Convergence.BaseAdjustmentScale)
	}
	if c.Convergence.PhaseConvergenceGoal < 0 || c.Convergence.PhaseConvergenceGoal > 1 {
		return fmt.Errorf("PhaseConvergenceGoal must be between 0 and 1, got %f", c.Convergence.PhaseConvergenceGoal)
	}
	return nil
}

func (c *Config) validateThresholds() error {
	if c.Thresholds.ModerateCoherence > c.Thresholds.HighCoherence {
		return fmt.Errorf("ModerateCoherence (%f) must be <= HighCoherence (%f)",
			c.Thresholds.ModerateCoherence, c.Thresholds.HighCoherence)
	}
	if c.Thresholds.HighCoherence > c.Thresholds.VeryHighCoherence {
		return fmt.Errorf("HighCoherence (%f) must be <= VeryHighCoherence (%f)",
			c.Thresholds.HighCoherence, c.Thresholds.VeryHighCoherence)
	}
	return nil
}

func (c *Config) validateVariation() error {
	if c.Variation.BaseRange[0] < 0 || c.Variation.BaseRange[1] < 0 {
		return fmt.Errorf("variation ranges must be positive")
	}
	if c.Variation.BaseRange[0] > c.Variation.BaseRange[1] {
		return fmt.Errorf("variation min (%f) must be <= max (%f)",
			c.Variation.BaseRange[0], c.Variation.BaseRange[1])
	}
	if c.Variation.PerturbationChance < 0 || c.Variation.PerturbationChance > 1 {
		return fmt.Errorf("PerturbationChance must be between 0 and 1, got %f", c.Variation.PerturbationChance)
	}
	return nil
}

func (c *Config) validateStrategy() error {
	if c.Strategy.UpdateInterval <= 0 {
		return fmt.Errorf("UpdateInterval must be positive, got %v", c.Strategy.UpdateInterval)
	}
	if c.Strategy.ExplorationTimeWindow <= 0 {
		return fmt.Errorf("ExplorationTimeWindow must be positive, got %v", c.Strategy.ExplorationTimeWindow)
	}
	return nil
}

func (c *Config) validateResonance() error {
	if c.Resonance.AffectedAgents < 0 || c.Resonance.AffectedAgents > 1 {
		return fmt.Errorf("AffectedAgents must be between 0 and 1, got %f", c.Resonance.AffectedAgents)
	}
	if c.Resonance.ActivationRate < 0 || c.Resonance.ActivationRate > 1 {
		return fmt.Errorf("ActivationRate must be between 0 and 1, got %f", c.Resonance.ActivationRate)
	}
	return nil
}

// defaultConfig returns a balanced default configuration.
func defaultConfig() *Config {
	return &Config{
		Convergence: ConvergenceConfig{
			ToleranceSmall:           0.015,
			ToleranceMedium:          0.008,
			ToleranceLarge:           0.005,
			PhaseConvergenceGoal:     0.85,
			PatternDistanceThreshold: 0.1,
			BaseAdjustmentScale:      0.65,
		},
		Thresholds: ThresholdConfig{
			PhaseVariance:     0.5,
			HighCoherence:     0.9,
			VeryHighCoherence: 0.92,
			ModerateCoherence: 0.85,
		},
		Variation: VariationConfig{
			BaseRange:             [2]float64{0.15, 0.30},
			CoherenceFactor:       0.2,
			RandomWalkMagnitude:   0.1,
			PerturbationMagnitude: 0.08,
			PerturbationChance:    0.15,
		},
		Strategy: StrategyConfig{
			MaxIterationsFactor:   200,
			UpdateInterval:        100 * time.Millisecond,
			ExplorationBonusMax:   0.3,
			ExplorationTimeWindow: 60 * time.Second,
			RandomExploration:     0.1,
		},
		Resonance: ResonanceConfig{
			NoiseMagnitude: 0.5,
			AffectedAgents: 0.1,
			ActivationRate: 0.1,
		},
	}
}

// minValue returns the minimum of two values.
func minValue[T ~int | ~float64](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// maxValue returns the maximum of two values.
func maxValue[T ~int | ~float64](a, b T) T {
	if a > b {
		return a
	}
	return b
}
