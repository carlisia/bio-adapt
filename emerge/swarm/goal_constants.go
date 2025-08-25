package swarm

import "time"

// Goal-specific configuration constants for different optimization goals.
// These constants define how the swarm behaves for each specific goal type.

// MinimizeAPICalls goal constants - Optimize for minimal external API usage
const (
	MinAPIToleranceSmall           = 0.01
	MinAPIToleranceMedium          = 0.005
	MinAPIToleranceLarge           = 0.003
	MinAPIPhaseConvergenceGoal     = 0.90
	MinAPIPatternDistanceThreshold = 0.05
	MinAPIBaseAdjustmentScale      = 0.75

	MinAPIPhaseVariance     = 0.3
	MinAPIHighCoherence     = 0.85
	MinAPIVeryHighCoherence = 0.90
	MinAPIModerateCoherence = 0.80

	MinAPIVariationMin          = 0.05
	MinAPIVariationMax          = 0.15
	MinAPICoherenceFactor       = 0.15
	MinAPIRandomWalkMagnitude   = 0.05
	MinAPIPerturbationMagnitude = 0.04
	MinAPIPerturbationChance    = 0.10

	MinAPIMaxIterationsFactor   = 150.0
	MinAPIUpdateInterval        = 200 * time.Millisecond
	MinAPIExplorationBonusMax   = 0.2
	MinAPIExplorationTimeWindow = 45 * time.Second
	MinAPIRandomExploration     = 0.05

	MinAPINoiseMagnitude = 0.3
	MinAPIAffectedAgents = 0.05
	MinAPIActivationRate = 0.05
)

// DistributeLoad goal constants - Optimize for load distribution
const (
	DistLoadToleranceSmall           = 0.025
	DistLoadToleranceMedium          = 0.015
	DistLoadToleranceLarge           = 0.010
	DistLoadPhaseConvergenceGoal     = 0.30 // Anti-phase for distribution
	DistLoadPatternDistanceThreshold = 0.2
	DistLoadBaseAdjustmentScale      = 0.55

	DistLoadPhaseVariance     = 0.8 // High variance is good for distribution
	DistLoadHighCoherence     = 0.4 // Lower coherence targets
	DistLoadVeryHighCoherence = 0.5
	DistLoadModerateCoherence = 0.3

	DistLoadVariationMin          = 0.25
	DistLoadVariationMax          = 0.45
	DistLoadCoherenceFactor       = 0.3
	DistLoadRandomWalkMagnitude   = 0.2
	DistLoadPerturbationMagnitude = 0.15
	DistLoadPerturbationChance    = 0.25

	DistLoadMaxIterationsFactor   = 100.0
	DistLoadUpdateInterval        = 50 * time.Millisecond
	DistLoadExplorationBonusMax   = 0.4
	DistLoadExplorationTimeWindow = 30 * time.Second
	DistLoadRandomExploration     = 0.15

	DistLoadNoiseMagnitude = 0.6
	DistLoadAffectedAgents = 0.15
	DistLoadActivationRate = 0.15
)

// ReachConsensus goal constants - Optimize for consensus formation
const (
	ConsensusToleranceSmall           = 0.008
	ConsensusToleranceMedium          = 0.004
	ConsensusToleranceLarge           = 0.002
	ConsensusPhaseConvergenceGoal     = 0.95
	ConsensusPatternDistanceThreshold = 0.03
	ConsensusBaseAdjustmentScale      = 0.80

	ConsensusPhaseVariance     = 0.2
	ConsensusHighCoherence     = 0.90
	ConsensusVeryHighCoherence = 0.95
	ConsensusModerateCoherence = 0.85

	ConsensusVariationMin          = 0.03
	ConsensusVariationMax          = 0.10
	ConsensusCoherenceFactor       = 0.10
	ConsensusRandomWalkMagnitude   = 0.03
	ConsensusPerturbationMagnitude = 0.02
	ConsensusPerturbationChance    = 0.05

	ConsensusMaxIterationsFactor   = 250.0
	ConsensusUpdateInterval        = 100 * time.Millisecond
	ConsensusExplorationBonusMax   = 0.25
	ConsensusExplorationTimeWindow = 60 * time.Second
	ConsensusRandomExploration     = 0.08

	ConsensusNoiseMagnitude = 0.2
	ConsensusAffectedAgents = 0.03
	ConsensusActivationRate = 0.03
)

// MinimizeLatency goal constants - Optimize for low latency
const (
	LatencyToleranceSmall           = 0.020
	LatencyToleranceMedium          = 0.012
	LatencyToleranceLarge           = 0.008
	LatencyPhaseConvergenceGoal     = 0.75
	LatencyPatternDistanceThreshold = 0.15
	LatencyBaseAdjustmentScale      = 0.85

	LatencyPhaseVariance     = 0.4
	LatencyHighCoherence     = 0.80
	LatencyVeryHighCoherence = 0.85
	LatencyModerateCoherence = 0.75

	LatencyVariationMin          = 0.10
	LatencyVariationMax          = 0.20
	LatencyCoherenceFactor       = 0.15
	LatencyRandomWalkMagnitude   = 0.08
	LatencyPerturbationMagnitude = 0.06
	LatencyPerturbationChance    = 0.12

	LatencyMaxIterationsFactor   = 80.0
	LatencyUpdateInterval        = 25 * time.Millisecond
	LatencyExplorationBonusMax   = 0.35
	LatencyExplorationTimeWindow = 20 * time.Second
	LatencyRandomExploration     = 0.12

	LatencyNoiseMagnitude = 0.4
	LatencyAffectedAgents = 0.08
	LatencyActivationRate = 0.10
)

// SaveEnergy goal constants - Optimize for energy efficiency
const (
	EnergyToleranceSmall           = 0.030
	EnergyToleranceMedium          = 0.020
	EnergyToleranceLarge           = 0.015
	EnergyPhaseConvergenceGoal     = 0.70
	EnergyPatternDistanceThreshold = 0.20
	EnergyBaseAdjustmentScale      = 0.40

	EnergyPhaseVariance     = 0.6
	EnergyHighCoherence     = 0.75
	EnergyVeryHighCoherence = 0.80
	EnergyModerateCoherence = 0.70

	EnergyVariationMin          = 0.08
	EnergyVariationMax          = 0.18
	EnergyCoherenceFactor       = 0.12
	EnergyRandomWalkMagnitude   = 0.06
	EnergyPerturbationMagnitude = 0.05
	EnergyPerturbationChance    = 0.08

	EnergyMaxIterationsFactor   = 300.0
	EnergyUpdateInterval        = 250 * time.Millisecond
	EnergyExplorationBonusMax   = 0.15
	EnergyExplorationTimeWindow = 90 * time.Second
	EnergyRandomExploration     = 0.05

	EnergyNoiseMagnitude = 0.25
	EnergyAffectedAgents = 0.05
	EnergyActivationRate = 0.05
)

// MaintainRhythm goal constants - Optimize for rhythmic synchronization
const (
	RhythmToleranceSmall           = 0.012
	RhythmToleranceMedium          = 0.006
	RhythmToleranceLarge           = 0.003
	RhythmPhaseConvergenceGoal     = 0.80
	RhythmPatternDistanceThreshold = 0.08
	RhythmBaseAdjustmentScale      = 0.70

	RhythmPhaseVariance     = 0.4
	RhythmHighCoherence     = 0.85
	RhythmVeryHighCoherence = 0.90
	RhythmModerateCoherence = 0.80

	RhythmVariationMin          = 0.08
	RhythmVariationMax          = 0.16
	RhythmCoherenceFactor       = 0.15
	RhythmRandomWalkMagnitude   = 0.06
	RhythmPerturbationMagnitude = 0.05
	RhythmPerturbationChance    = 0.10

	RhythmMaxIterationsFactor   = 180.0
	RhythmUpdateInterval        = 100 * time.Millisecond
	RhythmExplorationBonusMax   = 0.25
	RhythmExplorationTimeWindow = 50 * time.Second
	RhythmRandomExploration     = 0.08

	RhythmNoiseMagnitude = 0.3
	RhythmAffectedAgents = 0.06
	RhythmActivationRate = 0.08
)

// RecoverFromFailure goal constants - Optimize for failure recovery
const (
	RecoveryToleranceSmall           = 0.025
	RecoveryToleranceMedium          = 0.015
	RecoveryToleranceLarge           = 0.010
	RecoveryPhaseConvergenceGoal     = 0.65
	RecoveryPatternDistanceThreshold = 0.18
	RecoveryBaseAdjustmentScale      = 0.60

	RecoveryPhaseVariance     = 0.7
	RecoveryHighCoherence     = 0.70
	RecoveryVeryHighCoherence = 0.75
	RecoveryModerateCoherence = 0.65

	RecoveryVariationMin          = 0.20
	RecoveryVariationMax          = 0.40
	RecoveryCoherenceFactor       = 0.25
	RecoveryRandomWalkMagnitude   = 0.15
	RecoveryPerturbationMagnitude = 0.12
	RecoveryPerturbationChance    = 0.20

	RecoveryMaxIterationsFactor   = 350.0
	RecoveryUpdateInterval        = 75 * time.Millisecond
	RecoveryExplorationBonusMax   = 0.40
	RecoveryExplorationTimeWindow = 30 * time.Second
	RecoveryRandomExploration     = 0.15

	RecoveryNoiseMagnitude = 0.7
	RecoveryAffectedAgents = 0.15
	RecoveryActivationRate = 0.20
)

// AdaptToTraffic goal constants - Optimize for traffic adaptation
const (
	TrafficToleranceSmall           = 0.040
	TrafficToleranceMedium          = 0.030
	TrafficToleranceLarge           = 0.020
	TrafficPhaseConvergenceGoal     = 0.50
	TrafficPatternDistanceThreshold = 0.25
	TrafficBaseAdjustmentScale      = 0.45

	TrafficPhaseVariance     = 0.9
	TrafficHighCoherence     = 0.60
	TrafficVeryHighCoherence = 0.70
	TrafficModerateCoherence = 0.50

	TrafficVariationMin          = 0.30
	TrafficVariationMax          = 0.50
	TrafficCoherenceFactor       = 0.35
	TrafficRandomWalkMagnitude   = 0.25
	TrafficPerturbationMagnitude = 0.20
	TrafficPerturbationChance    = 0.30

	TrafficMaxIterationsFactor   = 100.0
	TrafficUpdateInterval        = 50 * time.Millisecond
	TrafficExplorationBonusMax   = 0.50
	TrafficExplorationTimeWindow = 20 * time.Second
	TrafficRandomExploration     = 0.25

	TrafficNoiseMagnitude = 0.8
	TrafficAffectedAgents = 0.20
	TrafficActivationRate = 0.25
)
