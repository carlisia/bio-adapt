// Package pattern defines request pattern types for workload simulation.
package pattern

// Type represents a request pattern type.
type Type int

const (
	// Unset represents an uninitialized pattern (used to detect when to auto-select)
	Unset Type = iota

	// HighFrequency represents continuous stream of requests (>10/sec per workload)
	HighFrequency

	// Burst represents sudden spikes of activity followed by quiet periods
	Burst

	// Steady represents consistent, predictable request rate
	Steady

	// Mixed represents combination of patterns from different workloads
	Mixed

	// Sparse represents infrequent, irregular requests (<1/sec per workload)
	Sparse
)

// String returns the human-readable name of the pattern.
func (p Type) String() string {
	switch p {
	case Unset:
		return "Auto"
	case HighFrequency:
		return "High-Frequency"
	case Burst:
		return "Burst"
	case Steady:
		return "Steady"
	case Mixed:
		return "Mixed"
	case Sparse:
		return "Sparse"
	default:
		return "Unknown"
	}
}

// ShortKey returns the keyboard shortcut for this pattern.
func (p Type) ShortKey() string {
	switch p {
	case Unset:
		return ""
	case HighFrequency:
		return "h"
	case Burst:
		return "u"
	case Steady:
		return "y"
	case Mixed:
		return "x"
	case Sparse:
		return "z"
	default:
		return ""
	}
}

// Description returns a brief description of the pattern.
func (p Type) Description() string {
	switch p {
	case Unset:
		return "Auto-selected based on goal"
	case HighFrequency:
		return "Continuous high-rate requests"
	case Burst:
		return "Spikes followed by quiet periods"
	case Steady:
		return "Consistent predictable rate"
	case Mixed:
		return "Combination of patterns"
	case Sparse:
		return "Infrequent irregular requests"
	default:
		return ""
	}
}

// GetCoherenceModifier returns a coherence modifier based on goal, size, and pattern combination.
// Returns a value between 0.5 (poor combination) and 1.5 (excellent combination).
func GetCoherenceModifier(goalType int, agentCount int, p Type) float64 {
	// Base modifier
	modifier := 1.0

	// Pattern-specific modifiers based on goal type
	// Note: goalType is passed as int to avoid circular dependency
	switch goalType {
	case 0: // MinimizeAPICalls
		switch p {
		case HighFrequency, Burst:
			modifier = 1.3 // Excellent for batching
		case Mixed:
			modifier = 1.2 // Good adaptation
		case Steady:
			modifier = 1.0 // OK
		case Sparse:
			modifier = 0.7 // Poor - few calls to batch
		case Unset:
			modifier = 1.0 // Default
		}
	case 1: // DistributeLoad
		switch p {
		case HighFrequency, Steady:
			modifier = 1.3 // Excellent for distribution
		case Mixed:
			modifier = 1.1 // Good
		case Burst:
			modifier = 0.9 // OK - can flatten spikes
		case Sparse:
			modifier = 0.6 // Poor - already distributed
		case Unset:
			modifier = 1.0 // Default
		}
	case 2: // ReachConsensus
		switch p {
		case Steady:
			modifier = 1.3 // Excellent - regular voting
		case HighFrequency:
			modifier = 1.0 // OK
		case Mixed:
			modifier = 0.9 // OK - handles variance
		case Burst, Sparse:
			modifier = 0.6 // Poor - inconsistent participation
		case Unset:
			modifier = 1.0 // Default
		}
	case 3: // MinimizeLatency
		switch p {
		case HighFrequency:
			modifier = 1.3 // Excellent - quick response
		case Steady:
			modifier = 1.2 // Good - predictable latency
		case Burst:
			modifier = 0.8 // Fair - latency spikes
		case Mixed:
			modifier = 0.9 // OK - variable latency
		case Sparse:
			modifier = 0.6 // Poor - slow response
		case Unset:
			modifier = 1.0 // Default
		}
	case 4: // SaveEnergy
		switch p {
		case Sparse:
			modifier = 1.3 // Excellent - minimal activity
		case Burst:
			modifier = 1.2 // Good - concentrated work periods
		case Steady:
			modifier = 0.9 // OK - constant drain
		case Mixed:
			modifier = 0.8 // Fair - unpredictable
		case HighFrequency:
			modifier = 0.6 // Poor - high energy use
		case Unset:
			modifier = 1.0 // Default
		}
	case 5: // MaintainRhythm
		switch p {
		case Steady:
			modifier = 1.3 // Excellent - perfect rhythm
		case HighFrequency:
			modifier = 1.1 // Good - regular sync
		case Mixed:
			modifier = 0.9 // OK - some disruption
		case Burst:
			modifier = 0.8 // Fair - rhythm breaks
		case Sparse:
			modifier = 0.7 // Poor - gaps disrupt rhythm
		case Unset:
			modifier = 1.0 // Default
		}
	case 6: // RecoverFromFailure
		switch p {
		case Mixed:
			modifier = 1.3 // Excellent - handles variability
		case HighFrequency:
			modifier = 1.2 // Good - quick detection
		case Burst:
			modifier = 1.1 // Good - rapid response
		case Steady:
			modifier = 1.0 // OK - predictable
		case Sparse:
			modifier = 0.6 // Poor - slow detection
		case Unset:
			modifier = 1.0 // Default
		}
	case 7: // AdaptToTraffic
		switch p {
		case Burst:
			modifier = 1.3 // Excellent - simulates surges
		case Mixed:
			modifier = 1.2 // Good - variable traffic
		case HighFrequency:
			modifier = 1.0 // OK - high load
		case Steady:
			modifier = 0.8 // Fair - no variation
		case Sparse:
			modifier = 0.6 // Poor - insufficient load
		case Unset:
			modifier = 1.0 // Default
		}
	}

	// Size adjustments
	if agentCount >= 1000 && p == HighFrequency {
		modifier *= 0.9 // Harder to coordinate at large scale
	}
	if agentCount <= 50 && p == Sparse {
		modifier *= 0.8 // Not enough activity for small swarms
	}

	return modifier
}

// Qualifier constants for pattern evaluation
const (
	QualifierOptimal   = "optimal"
	QualifierExcellent = "excellent"
	QualifierGood      = "good"
	QualifierFair      = "fair"
	QualifierPoor      = "poor"
)

// Modifier thresholds for qualifiers
const (
	ThresholdOptimal   = 1.3
	ThresholdExcellent = 1.2
	ThresholdGood      = 1.0
	ThresholdFair      = 0.8
)

// GetQualifier returns a quality qualifier for a pattern based on goal and size.
func GetQualifier(goalType int, agentCount int, p Type) string {
	modifier := GetCoherenceModifier(goalType, agentCount, p)

	switch {
	case modifier >= ThresholdOptimal:
		return QualifierOptimal
	case modifier >= ThresholdExcellent:
		return QualifierExcellent
	case modifier >= ThresholdGood:
		return QualifierGood
	case modifier >= ThresholdFair:
		return QualifierFair
	default:
		return QualifierPoor
	}
}

// IsRecommended returns whether this pattern is recommended for a given goal and size.
func IsRecommended(goalType int, agentCount int, p Type) bool {
	return GetCoherenceModifier(goalType, agentCount, p) >= ThresholdExcellent
}

// BestPatternForGoal returns the optimal request pattern for a given goal type.
func BestPatternForGoal(goalType int) Type {
	switch goalType {
	case 0: // MinimizeAPICalls
		return HighFrequency // Best for batching many requests
	case 1: // DistributeLoad
		return Steady // Best for even distribution
	case 2: // ReachConsensus
		return Steady // Best for regular participation
	case 3: // MinimizeLatency
		return HighFrequency // Best for quick response
	case 4: // SaveEnergy
		return Sparse // Best for minimal activity
	case 5: // MaintainRhythm
		return Steady // Best for perfect rhythm
	case 6: // RecoverFromFailure
		return Mixed // Best for handling variability
	case 7: // AdaptToTraffic
		return Burst // Best for simulating surges
	default:
		return Steady // Safe default
	}
}
