package swarm

import "time"

// Recovery thresholds - Constants for disruption detection and recovery
const (
	// Minimum viable coherence levels based on target coherence
	MinViableCoherenceHigh    = 0.5 // For targets >= 0.9
	MinViableCoherenceMedium  = 0.4 // For targets >= 0.7
	MinViableCoherenceLow     = 0.3 // For targets >= 0.5
	MinViableCoherenceVeryLow = 0.2 // For targets < 0.5

	// Target margin ratios - how close to target is "good enough"
	TargetMarginRatioHigh    = 0.98 // Within 98% of target for high coherence
	TargetMarginRatioMedium  = 0.95 // Within 95% of target for medium coherence
	TargetMarginRatioLow     = 0.92 // Within 92% of target for low coherence
	TargetMarginRatioVeryLow = 0.85 // Within 85% of target for very low coherence

	// Small drop ratios - when to start investigating
	SmallDropRatioHigh    = 0.03 // 3% drop for high coherence systems
	SmallDropRatioMedium  = 0.05 // 5% drop for medium coherence
	SmallDropRatioLow     = 0.08 // 8% drop for low coherence
	SmallDropRatioVeryLow = 0.10 // 10% drop for very low coherence

	// Large drop ratios - when to immediately resync
	LargeDropRatioHigh    = 0.08 // 8% drop is major for high coherence
	LargeDropRatioMedium  = 0.12 // 12% drop is major for medium coherence
	LargeDropRatioLow     = 0.15 // 15% drop is major for low coherence
	LargeDropRatioVeryLow = 0.20 // 20% drop is major for very low coherence

	// Timing constants
	DefaultCheckInterval  = 100 * time.Millisecond
	DefaultStuckThreshold = 30 // Number of checks without improvement

	// Coherence level boundaries for threshold selection
	HighCoherenceThreshold   = 0.9
	MediumCoherenceThreshold = 0.7
	LowCoherenceThreshold    = 0.5

	// Peak coherence decay rate
	PeakCoherenceDecayRate = 0.9995 // Slowly forget old peaks to adapt

	// Minimum time between resync attempts
	MinResyncInterval = 500 * time.Millisecond

	// Improvement detection threshold
	ImprovementThreshold = 0.001 // Minimum change to count as improvement
)

// Goal-directed synchronization constants
const (
	// Default convergence tolerances by swarm size
	DefaultToleranceSmall  = 0.015 // For swarms < 10
	DefaultToleranceMedium = 0.008 // For swarms < 50
	DefaultToleranceLarge  = 0.005 // For swarms >= 50

	// Default convergence parameters
	DefaultPhaseConvergenceGoal     = 0.85 // How close phases need to converge (0-1)
	DefaultPatternDistanceThreshold = 0.1  // How close pattern needs to match target
	DefaultBaseAdjustmentScale      = 0.65 // Base aggressiveness of adjustments

	// Default coherence thresholds for behaviors
	DefaultPhaseVariance     = 0.5  // Triggers special phase alignment mode
	DefaultHighCoherence     = 0.9  // High coherence threshold
	DefaultVeryHighCoherence = 0.92 // Very high coherence threshold
	DefaultModerateCoherence = 0.85 // Moderate coherence threshold

	// Default variation parameters
	DefaultVariationMin          = 0.15 // Minimum variation to prevent perfect sync
	DefaultVariationMax          = 0.30 // Maximum variation to prevent perfect sync
	DefaultCoherenceFactor       = 0.2  // Scale variation with coherence level
	DefaultRandomWalkMagnitude   = 0.1  // Random perturbations for diversity
	DefaultPerturbationMagnitude = 0.08 // Perturbation strength
	DefaultPerturbationChance    = 0.15 // Probability of perturbation

	// Default strategy parameters
	DefaultMaxIterationsFactor   = 200.0                  // Convergence time control
	DefaultUpdateInterval        = 100 * time.Millisecond // Update frequency
	DefaultExplorationBonusMax   = 0.3                    // Maximum exploration bonus
	DefaultExplorationTimeWindow = 60 * time.Second       // Exploration time window
	DefaultRandomExploration     = 0.1                    // Random exploration chance

	// Default resonance parameters
	DefaultNoiseMagnitude = 0.5 // Stochastic resonance noise (±0.25 radians)
	DefaultAffectedAgents = 0.1 // Fraction of swarm affected by resonance
	DefaultActivationRate = 0.1 // Chance of resonance when stuck

	// Scale adjustment factors
	TinyScaleUpdateInterval  = 20 * time.Millisecond
	SmallScaleUpdateInterval = 50 * time.Millisecond
	LargeScaleUpdateInterval = 150 * time.Millisecond
	HugeScaleUpdateInterval  = 200 * time.Millisecond

	TinyScaleTolerance  = 0.020
	SmallScaleTolerance = 0.015

	// Trait tuning multipliers
	StabilityAdjustmentScale = 0.6 // Gentler adjustments for stability
	StabilityVariationMin    = 1.2 // Increase minimum variation
	StabilityVariationMax    = 1.3 // Increase maximum variation
	StabilityPerturbation    = 0.7 // Reduce perturbations
	StabilityUpdateInterval  = 1.5 // Slower updates
	StabilityNoise           = 0.6 // Less noise

	SpeedAdjustmentScale     = 1.3 // More aggressive for speed
	SpeedToleranceMultiplier = 1.5 // Looser tolerances
	SpeedUpdateInterval      = 0.5 // Faster updates
	SpeedIterationFactor     = 0.7 // Fewer iterations

	EfficiencyUpdateInterval     = 2.0 // Slower updates for efficiency
	EfficiencyPerturbationChance = 0.5 // Less perturbation
	EfficiencyAffectedAgents     = 0.5 // Fewer agents in resonance
	EfficiencyExplorationBonus   = 0.7 // Less exploration

	ThroughputAdjustmentScale   = 1.2 // More aggressive
	ThroughputUpdateInterval    = 0.7 // Faster updates
	ThroughputRandomExploration = 1.5 // More exploration

	ResilienceVariationMin   = 1.3 // More variation
	ResilienceVariationMax   = 1.4 // More variation
	ResilienceNoise          = 1.5 // More noise
	ResilienceAffectedAgents = 1.5 // More agents affected
	ResilienceActivationRate = 1.8 // More frequent activation

	PrecisionToleranceMultiplier = 0.5  // Tighter tolerances
	PrecisionPatternDistance     = 0.5  // Tighter pattern matching
	PrecisionPhaseConvergence    = 1.1  // Higher convergence goal
	PrecisionMaxConvergence      = 0.98 // Maximum convergence goal

	// Scale size multipliers
	LargeScaleToleranceSmall  = 1.5
	LargeScaleToleranceMedium = 1.3
	LargeScaleToleranceLarge  = 1.2
	LargeScaleVariationMin    = 1.2
	LargeScaleVariationMax    = 1.3
	LargeScaleIterationFactor = 1.5

	HugeScaleToleranceSmall    = 2.0
	HugeScaleToleranceMedium   = 1.8
	HugeScaleToleranceLarge    = 1.5
	HugeScaleVariationMin      = 1.5
	HugeScaleVariationMax      = 1.8
	HugeScaleIterationFactor   = 2.0
	HugeScaleMaxAffectedAgents = 0.05 // Maximum affected agents for huge swarms
)
