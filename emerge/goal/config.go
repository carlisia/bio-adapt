// Package goal defines business objectives for swarm synchronization.
// Goals represent WHAT you want to achieve with the swarm.
package goal

// No imports needed

// Type represents a business objective that the swarm should achieve.
type Type int

const (
	// MinimizeAPICalls reduces external API calls through batching.
	// Achieves high synchronization to coordinate request timing.
	MinimizeAPICalls Type = iota

	// DistributeLoad spreads work evenly across agents.
	// Uses anti-phase synchronization to avoid resource contention.
	DistributeLoad

	// ReachConsensus ensures all agents agree on a common state.
	// Requires both high coherence and phase convergence.
	ReachConsensus

	// MinimizeLatency optimizes for real-time responsiveness.
	// Balances synchronization with quick adaptation.
	MinimizeLatency

	// SaveEnergy reduces resource consumption.
	// Uses minimal synchronization updates and gentle strategies.
	SaveEnergy

	// MaintainRhythm keeps periodic tasks synchronized.
	// Focuses on frequency locking over phase alignment.
	MaintainRhythm

	// RecoverFromFailure prioritizes self-healing and resilience.
	// Uses robust strategies that handle agent failures.
	RecoverFromFailure

	// AdaptToTraffic optimizes for dynamic traffic pattern changes.
	// Quickly responds to shifting local goals and load patterns.
	AdaptToTraffic
)

// String returns a human-friendly description of what this goal optimizes for.
func (g Type) String() string {
	switch g {
	case MinimizeAPICalls:
		return "Batch API Calls"
	case DistributeLoad:
		return "Load Distribution"
	case ReachConsensus:
		return "Consensus Building"
	case MinimizeLatency:
		return "Low Latency"
	case SaveEnergy:
		return "Energy Saving"
	case MaintainRhythm:
		return "Periodic Tasks"
	case RecoverFromFailure:
		return "Fault Recovery"
	case AdaptToTraffic:
		return "Traffic Adaptation"
	default:
		return "Custom Goal"
	}
}

// ShortKey returns a single character key for keyboard control.
func (g Type) ShortKey() string {
	switch g {
	case MinimizeAPICalls:
		return "B" // Batch
	case DistributeLoad:
		return "L" // Load
	case ReachConsensus:
		return "C" // Consensus
	case MinimizeLatency:
		return "T" // laTency (avoid conflict with L)
	case SaveEnergy:
		return "E" // Energy
	case MaintainRhythm:
		return "R" // Rhythm
	case RecoverFromFailure:
		return "F" // Failure
	case AdaptToTraffic:
		return "A" // Adapt
	default:
		return "?"
	}
}

// IsRecommendedForSize returns whether this goal works well with the given swarm size.
func (g Type) IsRecommendedForSize(agentCount int) bool {
	switch g {
	case MinimizeAPICalls:
		// Works well at all sizes - batching benefits increase with size
		return true
	case DistributeLoad:
		// Needs enough agents to distribute load effectively (20+)
		return agentCount >= 20
	case ReachConsensus:
		// 50-1000 agents is optimal - too hard at huge sizes
		return agentCount >= 50 && agentCount <= 1000
	case MinimizeLatency:
		// Better at smaller sizes for quick response (≤200)
		return agentCount <= 200
	case SaveEnergy:
		// Energy savings harder to coordinate at large sizes (≤200)
		return agentCount <= 200
	case MaintainRhythm:
		// Works at all sizes
		return true
	case RecoverFromFailure:
		// Better with redundancy, needs at least 20 agents
		return agentCount >= 20
	case AdaptToTraffic:
		// Needs enough agents to handle varying traffic (20-1000)
		return agentCount >= 20 && agentCount <= 1000
	default:
		return true
	}
}
