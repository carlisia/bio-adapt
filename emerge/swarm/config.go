package swarm

import (
	"fmt"
	"time"

	"github.com/carlisia/bio-adapt/emerge/goal"
	"github.com/carlisia/bio-adapt/emerge/trait"
)

// Config holds all configuration for goal-directed synchronization.
type Config struct {
	Convergence ConvergenceConfig
	Thresholds  ThresholdConfig
	Variation   VariationConfig
	Strategy    StrategyConfig
	Resonance   ResonanceConfig

	// Agent count override (0 means use default from swarm size)
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
				ToleranceSmall:           MinAPIToleranceSmall,
				ToleranceMedium:          MinAPIToleranceMedium,
				ToleranceLarge:           MinAPIToleranceLarge,
				PhaseConvergenceGoal:     MinAPIPhaseConvergenceGoal,
				PatternDistanceThreshold: MinAPIPatternDistanceThreshold,
				BaseAdjustmentScale:      MinAPIBaseAdjustmentScale,
			},
			Thresholds: ThresholdConfig{
				PhaseVariance:     MinAPIPhaseVariance,
				HighCoherence:     MinAPIHighCoherence,
				VeryHighCoherence: MinAPIVeryHighCoherence,
				ModerateCoherence: MinAPIModerateCoherence,
			},
			Variation: VariationConfig{
				BaseRange:             [2]float64{MinAPIVariationMin, MinAPIVariationMax},
				CoherenceFactor:       MinAPICoherenceFactor,
				RandomWalkMagnitude:   MinAPIRandomWalkMagnitude,
				PerturbationMagnitude: MinAPIPerturbationMagnitude,
				PerturbationChance:    MinAPIPerturbationChance,
			},
			Strategy: StrategyConfig{
				MaxIterationsFactor:   MinAPIMaxIterationsFactor,
				UpdateInterval:        MinAPIUpdateInterval,
				ExplorationBonusMax:   MinAPIExplorationBonusMax,
				ExplorationTimeWindow: MinAPIExplorationTimeWindow,
				RandomExploration:     MinAPIRandomExploration,
			},
			Resonance: ResonanceConfig{
				NoiseMagnitude: MinAPINoiseMagnitude,
				AffectedAgents: MinAPIAffectedAgents,
				ActivationRate: MinAPIActivationRate,
			},
		}

	case goal.DistributeLoad:
		return &Config{
			Convergence: ConvergenceConfig{
				ToleranceSmall:           DistLoadToleranceSmall,
				ToleranceMedium:          DistLoadToleranceMedium,
				ToleranceLarge:           DistLoadToleranceLarge,
				PhaseConvergenceGoal:     DistLoadPhaseConvergenceGoal,
				PatternDistanceThreshold: DistLoadPatternDistanceThreshold,
				BaseAdjustmentScale:      DistLoadBaseAdjustmentScale,
			},
			Thresholds: ThresholdConfig{
				PhaseVariance:     DistLoadPhaseVariance,
				HighCoherence:     DistLoadHighCoherence,
				VeryHighCoherence: DistLoadVeryHighCoherence,
				ModerateCoherence: DistLoadModerateCoherence,
			},
			Variation: VariationConfig{
				BaseRange:             [2]float64{DistLoadVariationMin, DistLoadVariationMax},
				CoherenceFactor:       DistLoadCoherenceFactor,
				RandomWalkMagnitude:   DistLoadRandomWalkMagnitude,
				PerturbationMagnitude: DistLoadPerturbationMagnitude,
				PerturbationChance:    DistLoadPerturbationChance,
			},
			Strategy: StrategyConfig{
				MaxIterationsFactor:   DistLoadMaxIterationsFactor,
				UpdateInterval:        DistLoadUpdateInterval,
				ExplorationBonusMax:   DistLoadExplorationBonusMax,
				ExplorationTimeWindow: DistLoadExplorationTimeWindow,
				RandomExploration:     DistLoadRandomExploration,
			},
			Resonance: ResonanceConfig{
				NoiseMagnitude: DistLoadNoiseMagnitude,
				AffectedAgents: DistLoadAffectedAgents,
				ActivationRate: DistLoadActivationRate,
			},
		}

	case goal.ReachConsensus:
		return &Config{
			Convergence: ConvergenceConfig{
				ToleranceSmall:           ConsensusToleranceSmall,
				ToleranceMedium:          ConsensusToleranceMedium,
				ToleranceLarge:           ConsensusToleranceLarge,
				PhaseConvergenceGoal:     ConsensusPhaseConvergenceGoal,
				PatternDistanceThreshold: ConsensusPatternDistanceThreshold,
				BaseAdjustmentScale:      ConsensusBaseAdjustmentScale,
			},
			Thresholds: ThresholdConfig{
				PhaseVariance:     ConsensusPhaseVariance,
				HighCoherence:     ConsensusHighCoherence,
				VeryHighCoherence: ConsensusVeryHighCoherence,
				ModerateCoherence: ConsensusModerateCoherence,
			},
			Variation: VariationConfig{
				BaseRange:             [2]float64{ConsensusVariationMin, ConsensusVariationMax},
				CoherenceFactor:       ConsensusCoherenceFactor,
				RandomWalkMagnitude:   ConsensusRandomWalkMagnitude,
				PerturbationMagnitude: ConsensusPerturbationMagnitude,
				PerturbationChance:    ConsensusPerturbationChance,
			},
			Strategy: StrategyConfig{
				MaxIterationsFactor:   ConsensusMaxIterationsFactor,
				UpdateInterval:        ConsensusUpdateInterval,
				ExplorationBonusMax:   ConsensusExplorationBonusMax,
				ExplorationTimeWindow: ConsensusExplorationTimeWindow,
				RandomExploration:     ConsensusRandomExploration,
			},
			Resonance: ResonanceConfig{
				NoiseMagnitude: ConsensusNoiseMagnitude,
				AffectedAgents: ConsensusAffectedAgents,
				ActivationRate: ConsensusActivationRate,
			},
		}

	case goal.MinimizeLatency:
		return &Config{
			Convergence: ConvergenceConfig{
				ToleranceSmall:           LatencyToleranceSmall,
				ToleranceMedium:          LatencyToleranceMedium,
				ToleranceLarge:           LatencyToleranceLarge,
				PhaseConvergenceGoal:     LatencyPhaseConvergenceGoal,
				PatternDistanceThreshold: LatencyPatternDistanceThreshold,
				BaseAdjustmentScale:      LatencyBaseAdjustmentScale,
			},
			Thresholds: ThresholdConfig{
				PhaseVariance:     LatencyPhaseVariance,
				HighCoherence:     LatencyHighCoherence,
				VeryHighCoherence: LatencyVeryHighCoherence,
				ModerateCoherence: LatencyModerateCoherence,
			},
			Variation: VariationConfig{
				BaseRange:             [2]float64{LatencyVariationMin, LatencyVariationMax},
				CoherenceFactor:       LatencyCoherenceFactor,
				RandomWalkMagnitude:   LatencyRandomWalkMagnitude,
				PerturbationMagnitude: LatencyPerturbationMagnitude,
				PerturbationChance:    LatencyPerturbationChance,
			},
			Strategy: StrategyConfig{
				MaxIterationsFactor:   LatencyMaxIterationsFactor,
				UpdateInterval:        LatencyUpdateInterval,
				ExplorationBonusMax:   LatencyExplorationBonusMax,
				ExplorationTimeWindow: LatencyExplorationTimeWindow,
				RandomExploration:     LatencyRandomExploration,
			},
			Resonance: ResonanceConfig{
				NoiseMagnitude: LatencyNoiseMagnitude,
				AffectedAgents: LatencyAffectedAgents,
				ActivationRate: LatencyActivationRate,
			},
		}

	case goal.SaveEnergy:
		return &Config{
			Convergence: ConvergenceConfig{
				ToleranceSmall:           EnergyToleranceSmall,
				ToleranceMedium:          EnergyToleranceMedium,
				ToleranceLarge:           EnergyToleranceLarge,
				PhaseConvergenceGoal:     EnergyPhaseConvergenceGoal,
				PatternDistanceThreshold: EnergyPatternDistanceThreshold,
				BaseAdjustmentScale:      EnergyBaseAdjustmentScale,
			},
			Thresholds: ThresholdConfig{
				PhaseVariance:     EnergyPhaseVariance,
				HighCoherence:     EnergyHighCoherence,
				VeryHighCoherence: EnergyVeryHighCoherence,
				ModerateCoherence: EnergyModerateCoherence,
			},
			Variation: VariationConfig{
				BaseRange:             [2]float64{EnergyVariationMin, EnergyVariationMax},
				CoherenceFactor:       EnergyCoherenceFactor,
				RandomWalkMagnitude:   EnergyRandomWalkMagnitude,
				PerturbationMagnitude: EnergyPerturbationMagnitude,
				PerturbationChance:    EnergyPerturbationChance,
			},
			Strategy: StrategyConfig{
				MaxIterationsFactor:   EnergyMaxIterationsFactor,
				UpdateInterval:        EnergyUpdateInterval,
				ExplorationBonusMax:   EnergyExplorationBonusMax,
				ExplorationTimeWindow: EnergyExplorationTimeWindow,
				RandomExploration:     EnergyRandomExploration,
			},
			Resonance: ResonanceConfig{
				NoiseMagnitude: EnergyNoiseMagnitude,
				AffectedAgents: EnergyAffectedAgents,
				ActivationRate: EnergyActivationRate,
			},
		}

	case goal.MaintainRhythm:
		return &Config{
			Convergence: ConvergenceConfig{
				ToleranceSmall:           RhythmToleranceSmall,
				ToleranceMedium:          RhythmToleranceMedium,
				ToleranceLarge:           RhythmToleranceLarge,
				PhaseConvergenceGoal:     RhythmPhaseConvergenceGoal,
				PatternDistanceThreshold: RhythmPatternDistanceThreshold,
				BaseAdjustmentScale:      RhythmBaseAdjustmentScale,
			},
			Thresholds: ThresholdConfig{
				PhaseVariance:     RhythmPhaseVariance,
				HighCoherence:     RhythmHighCoherence,
				VeryHighCoherence: RhythmVeryHighCoherence,
				ModerateCoherence: RhythmModerateCoherence,
			},
			Variation: VariationConfig{
				BaseRange:             [2]float64{RhythmVariationMin, RhythmVariationMax},
				CoherenceFactor:       RhythmCoherenceFactor,
				RandomWalkMagnitude:   RhythmRandomWalkMagnitude,
				PerturbationMagnitude: RhythmPerturbationMagnitude,
				PerturbationChance:    RhythmPerturbationChance,
			},
			Strategy: StrategyConfig{
				MaxIterationsFactor:   RhythmMaxIterationsFactor,
				UpdateInterval:        RhythmUpdateInterval,
				ExplorationBonusMax:   RhythmExplorationBonusMax,
				ExplorationTimeWindow: RhythmExplorationTimeWindow,
				RandomExploration:     RhythmRandomExploration,
			},
			Resonance: ResonanceConfig{
				NoiseMagnitude: RhythmNoiseMagnitude,
				AffectedAgents: RhythmAffectedAgents,
				ActivationRate: RhythmActivationRate,
			},
		}

	case goal.RecoverFromFailure:
		return &Config{
			Convergence: ConvergenceConfig{
				ToleranceSmall:           RecoveryToleranceSmall,
				ToleranceMedium:          RecoveryToleranceMedium,
				ToleranceLarge:           RecoveryToleranceLarge,
				PhaseConvergenceGoal:     RecoveryPhaseConvergenceGoal,
				PatternDistanceThreshold: RecoveryPatternDistanceThreshold,
				BaseAdjustmentScale:      RecoveryBaseAdjustmentScale,
			},
			Thresholds: ThresholdConfig{
				PhaseVariance:     RecoveryPhaseVariance,
				HighCoherence:     RecoveryHighCoherence,
				VeryHighCoherence: RecoveryVeryHighCoherence,
				ModerateCoherence: RecoveryModerateCoherence,
			},
			Variation: VariationConfig{
				BaseRange:             [2]float64{RecoveryVariationMin, RecoveryVariationMax},
				CoherenceFactor:       RecoveryCoherenceFactor,
				RandomWalkMagnitude:   RecoveryRandomWalkMagnitude,
				PerturbationMagnitude: RecoveryPerturbationMagnitude,
				PerturbationChance:    RecoveryPerturbationChance,
			},
			Strategy: StrategyConfig{
				MaxIterationsFactor:   RecoveryMaxIterationsFactor,
				UpdateInterval:        RecoveryUpdateInterval,
				ExplorationBonusMax:   RecoveryExplorationBonusMax,
				ExplorationTimeWindow: RecoveryExplorationTimeWindow,
				RandomExploration:     RecoveryRandomExploration,
			},
			Resonance: ResonanceConfig{
				NoiseMagnitude: RecoveryNoiseMagnitude,
				AffectedAgents: RecoveryAffectedAgents,
				ActivationRate: RecoveryActivationRate,
			},
		}

	case goal.AdaptToTraffic:
		return &Config{
			Convergence: ConvergenceConfig{
				ToleranceSmall:           TrafficToleranceSmall,
				ToleranceMedium:          TrafficToleranceMedium,
				ToleranceLarge:           TrafficToleranceLarge,
				PhaseConvergenceGoal:     TrafficPhaseConvergenceGoal,
				PatternDistanceThreshold: TrafficPatternDistanceThreshold,
				BaseAdjustmentScale:      TrafficBaseAdjustmentScale,
			},
			Thresholds: ThresholdConfig{
				PhaseVariance:     TrafficPhaseVariance,
				HighCoherence:     TrafficHighCoherence,
				VeryHighCoherence: TrafficVeryHighCoherence,
				ModerateCoherence: TrafficModerateCoherence,
			},
			Variation: VariationConfig{
				BaseRange:             [2]float64{TrafficVariationMin, TrafficVariationMax},
				CoherenceFactor:       TrafficCoherenceFactor,
				RandomWalkMagnitude:   TrafficRandomWalkMagnitude,
				PerturbationMagnitude: TrafficPerturbationMagnitude,
				PerturbationChance:    TrafficPerturbationChance,
			},
			Strategy: StrategyConfig{
				MaxIterationsFactor:   TrafficMaxIterationsFactor,
				UpdateInterval:        TrafficUpdateInterval,
				ExplorationBonusMax:   TrafficExplorationBonusMax,
				ExplorationTimeWindow: TrafficExplorationTimeWindow,
				RandomExploration:     TrafficRandomExploration,
			},
			Resonance: ResonanceConfig{
				NoiseMagnitude: TrafficNoiseMagnitude,
				AffectedAgents: TrafficAffectedAgents,
				ActivationRate: TrafficActivationRate,
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

// WithSize adjusts the configuration for a specific swarm size.
func (c *Config) WithSize(agentCount int) *Config {
	switch {
	case agentCount <= 20:
		// Very small swarms need tight coupling
		c.Convergence.ToleranceSmall = 0.020
		c.Convergence.ToleranceMedium = 0.020
		c.Convergence.ToleranceLarge = 0.020
		c.Variation.BaseRange = [2]float64{0.05, 0.10}
		c.Strategy.UpdateInterval = 20 * time.Millisecond

	case agentCount <= 50:
		// Small swarms can converge quickly
		c.Convergence.ToleranceSmall = 0.015
		c.Convergence.ToleranceMedium = 0.015
		c.Convergence.ToleranceLarge = 0.015
		c.Variation.BaseRange[0] *= 0.8
		c.Variation.BaseRange[1] *= 0.9
		c.Strategy.UpdateInterval = time.Duration(minValue(50, int(c.Strategy.UpdateInterval/time.Millisecond))) * time.Millisecond

	case agentCount <= 200:
		// Medium swarms use default tolerances
		// (no changes needed)

	case agentCount <= 1000:
		// Large swarms need more tolerance
		c.Convergence.ToleranceSmall *= 1.5
		c.Convergence.ToleranceMedium *= 1.3
		c.Convergence.ToleranceLarge *= 1.2
		c.Variation.BaseRange[0] *= 1.2
		c.Variation.BaseRange[1] *= 1.3
		c.Strategy.UpdateInterval = time.Duration(maxValue(150, int(c.Strategy.UpdateInterval/time.Millisecond))) * time.Millisecond
		c.Strategy.MaxIterationsFactor *= 1.5

	default:
		// Huge swarms (2000+) need significant adjustments
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
			ToleranceSmall:           DefaultToleranceSmall,
			ToleranceMedium:          DefaultToleranceMedium,
			ToleranceLarge:           DefaultToleranceLarge,
			PhaseConvergenceGoal:     DefaultPhaseConvergenceGoal,
			PatternDistanceThreshold: DefaultPatternDistanceThreshold,
			BaseAdjustmentScale:      DefaultBaseAdjustmentScale,
		},
		Thresholds: ThresholdConfig{
			PhaseVariance:     DefaultPhaseVariance,
			HighCoherence:     DefaultHighCoherence,
			VeryHighCoherence: DefaultVeryHighCoherence,
			ModerateCoherence: DefaultModerateCoherence,
		},
		Variation: VariationConfig{
			BaseRange:             [2]float64{DefaultVariationMin, DefaultVariationMax},
			CoherenceFactor:       DefaultCoherenceFactor,
			RandomWalkMagnitude:   DefaultRandomWalkMagnitude,
			PerturbationMagnitude: DefaultPerturbationMagnitude,
			PerturbationChance:    DefaultPerturbationChance,
		},
		Strategy: StrategyConfig{
			MaxIterationsFactor:   DefaultMaxIterationsFactor,
			UpdateInterval:        DefaultUpdateInterval,
			ExplorationBonusMax:   DefaultExplorationBonusMax,
			ExplorationTimeWindow: DefaultExplorationTimeWindow,
			RandomExploration:     DefaultRandomExploration,
		},
		Resonance: ResonanceConfig{
			NoiseMagnitude: DefaultNoiseMagnitude,
			AffectedAgents: DefaultAffectedAgents,
			ActivationRate: DefaultActivationRate,
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
