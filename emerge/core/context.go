package core

// Context represents environmental awareness for decision-making.
// Agents sense their local environment and adapt strategies accordingly.
type Context struct {
	Neighbors      int     // Number of connected neighbors
	Density        float64 // Neighbor count / max neighbors
	Stability      float64 // Inverse of recent phase variance
	Progress       float64 // Convergence rate
	LocalCoherence float64 // Synchronization with neighbors
}

// Action represents a possible state change an agent can make.
// Actions have costs and expected benefits for decision-making.
type Action struct {
	Type    string  // "adjust_phase", "change_freq", "maintain"
	Value   float64 // Magnitude of change
	Cost    float64 // Energy required
	Benefit float64 // Expected improvement
}
