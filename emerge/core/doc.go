// Package core implements distributed agent synchronization through
// goal-directed coordination. Unlike traditional consensus algorithms, this approach
// achieves synchronization through autonomous agents with local goals that contribute
// to emergent global behavior - no central coordinator required.
//
// Key system properties:
//   - Agents have genuine autonomy and can refuse adjustments
//   - Local goals blend with global goals hierarchically
//   - Coordination emerges from gossip, not central control
//   - Context-sensitive strategy selection
//   - Energy-based resource management
//
// Performance characteristics:
//   - ~800ms convergence for 1000 agents (vs 500ms centralized)
//   - O(log N * log N) convergence with gossip protocol
//   - Probabilistic but robust convergence
//   - Graceful degradation under agent failures
//
// This implementation prioritizes emergent behavior and code simplicity
// over raw performance. The modular design allows upgrading to more sophisticated
// adaptive behaviors without architectural changes.
package core
