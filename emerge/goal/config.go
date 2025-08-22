// Package goal defines business objectives for swarm synchronization.
// Goals represent WHAT you want to achieve with the swarm.
package goal

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

// String returns the string representation of the goal.
func (g Type) String() string {
	switch g {
	case MinimizeAPICalls:
		return "minimize-api-calls"
	case DistributeLoad:
		return "distribute-load"
	case ReachConsensus:
		return "reach-consensus"
	case MinimizeLatency:
		return "minimize-latency"
	case SaveEnergy:
		return "save-energy"
	case MaintainRhythm:
		return "maintain-rhythm"
	case RecoverFromFailure:
		return "recover-from-failure"
	case AdaptToTraffic:
		return "adapt-to-traffic"
	default:
		return "unknown"
	}
}
