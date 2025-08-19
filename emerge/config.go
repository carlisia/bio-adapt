package emerge

import (
	"math"
	"time"
)

// SwarmConfig holds all configurable parameters for swarm behavior.
// This allows fine-tuning for different scales and use cases.
type SwarmConfig struct {
	// Network topology parameters
	ConnectionProbability float64 // Probability of connecting two agents (0-1)
	MaxNeighbors          int     // Maximum neighbors per agent
	MinNeighbors          int     // Minimum neighbors to ensure connectivity

	// Agent dynamics
	CouplingStrength float64 // How strongly agents influence each other (0-1)
	Stubbornness     float64 // Agent resistance to change (0-1)
	InitialEnergy    float64 // Starting energy for each agent

	// Density calculation
	AssumedMaxNeighbors int // For density calculations (0 = use actual swarm size)

	// Attractor basin parameters
	BasinStrength float64 // Strength of the attractor (0-1)
	BasinWidth    float64 // Width of the basin in radians

	// Decision making
	BaseConfidence   float64 // Base confidence level for decisions (0-1)
	InfluenceDefault float64 // Default influence level (0-1)

	// Auto-scaling
	AutoScale bool // Automatically adjust parameters based on swarm size
}

// DefaultConfig returns the default configuration for large swarms (100+ agents)
func DefaultConfig() SwarmConfig {
	return SwarmConfig{
		ConnectionProbability: 0.3,
		MaxNeighbors:          5,
		MinNeighbors:          2,
		CouplingStrength:      0.5,
		Stubbornness:          0.2,
		InitialEnergy:         100.0,
		AssumedMaxNeighbors:   20,
		BasinStrength:         0.8,
		BasinWidth:            math.Pi,
		BaseConfidence:        0.6,
		InfluenceDefault:      0.5,
		AutoScale:             false,
	}
}

// SmallSwarmConfig returns optimized configuration for small swarms (< 20 agents)
func SmallSwarmConfig() SwarmConfig {
	return SwarmConfig{
		ConnectionProbability: 0.9,  // Almost fully connected
		MaxNeighbors:          10,   // Can connect to most agents
		MinNeighbors:          2,    // Ensure some connectivity
		CouplingStrength:      0.8,  // Strong coupling for faster sync
		Stubbornness:          0.05, // Low resistance to change
		InitialEnergy:         100.0,
		AssumedMaxNeighbors:   0,           // Use actual swarm size
		BasinStrength:         0.9,         // Strong attractor
		BasinWidth:            2 * math.Pi, // Wider basin
		BaseConfidence:        0.8,         // Higher confidence
		InfluenceDefault:      0.7,         // Higher influence
		AutoScale:             false,
	}
}

// MediumSwarmConfig returns configuration for medium swarms (20-100 agents)
func MediumSwarmConfig() SwarmConfig {
	return SwarmConfig{
		ConnectionProbability: 0.5,
		MaxNeighbors:          8,
		MinNeighbors:          3,
		CouplingStrength:      0.6,
		Stubbornness:          0.1,
		InitialEnergy:         100.0,
		AssumedMaxNeighbors:   0, // Use actual swarm size
		BasinStrength:         0.85,
		BasinWidth:            1.5 * math.Pi,
		BaseConfidence:        0.7,
		InfluenceDefault:      0.6,
		AutoScale:             false,
	}
}

// AutoScaleConfig returns a configuration that automatically scales based on swarm size
func AutoScaleConfig(swarmSize int) SwarmConfig {
	config := SwarmConfig{
		AutoScale:     true,
		InitialEnergy: 100.0,
	}

	// Scale parameters based on swarm size
	if swarmSize < 10 {
		// Very small swarm - need strong connectivity and coupling
		config.ConnectionProbability = 1.0 // Fully connected
		config.MaxNeighbors = swarmSize - 1
		config.MinNeighbors = swarmSize - 1
		config.CouplingStrength = 0.9
		config.Stubbornness = 0.01
		config.BasinStrength = 0.95
		config.BasinWidth = 2.5 * math.Pi
		config.BaseConfidence = 0.9
		config.InfluenceDefault = 0.8
	} else if swarmSize < 20 {
		// Small swarm
		config = SmallSwarmConfig()
		config.MaxNeighbors = minInt(10, swarmSize-1)
		config.AutoScale = true
	} else if swarmSize < 100 {
		// Medium swarm
		config = MediumSwarmConfig()
		config.MaxNeighbors = minInt(15, swarmSize/5)
		config.AutoScale = true
	} else {
		// Large swarm
		config = DefaultConfig()
		config.MaxNeighbors = minInt(20, swarmSize/10)
		config.AutoScale = true
	}

	// Always use actual swarm size for density when auto-scaling
	config.AssumedMaxNeighbors = 0

	// Scale coupling inversely with size for stability
	if swarmSize > 100 {
		config.CouplingStrength *= 100.0 / float64(swarmSize)
	}

	return config
}

// ConfigForBatching returns configuration optimized for request batching scenarios
func ConfigForBatching(workloadCount int, batchWindow time.Duration) SwarmConfig {
	// For batching, we need very strong synchronization
	// Start with a small swarm config as baseline since batching needs tight coupling
	var config SwarmConfig
	if workloadCount <= 30 {
		// For small to medium batching scenarios, use very strong coupling
		config = SwarmConfig{
			ConnectionProbability: 0.9, // Almost fully connected
			MaxNeighbors:          minInt(workloadCount-1, 15),
			MinNeighbors:          minInt(workloadCount/2, 8),
			CouplingStrength:      0.85,
			Stubbornness:          0.02, // Very low resistance
			InitialEnergy:         100.0,
			AssumedMaxNeighbors:   0,
			BasinStrength:         0.95,
			BasinWidth:            2.5 * math.Pi,
			BaseConfidence:        0.9,
			InfluenceDefault:      0.85,
			AutoScale:             false,
		}
	} else {
		// For larger batching scenarios, still need strong but slightly relaxed
		config = AutoScaleConfig(workloadCount)
		config.ConnectionProbability = minFloat(0.75, config.ConnectionProbability*1.5)
		config.CouplingStrength = minFloat(0.8, config.CouplingStrength*1.3)
		config.Stubbornness = config.Stubbornness * 0.2
		config.BaseConfidence = 0.85
		config.InfluenceDefault = 0.8
		config.BasinStrength = 0.9
		config.BasinWidth = 2.0 * math.Pi
	}

	// Adjust based on batch window size
	if batchWindow < 100*time.Millisecond {
		// Fast batching - need very quick convergence
		config.CouplingStrength = minFloat(0.95, config.CouplingStrength*1.2)
		config.Stubbornness = config.Stubbornness * 0.5
	} else if batchWindow > 500*time.Millisecond {
		// Slower batching - can be slightly more gradual
		config.CouplingStrength = config.CouplingStrength * 0.95
	}

	return config
}

// Validate checks if the configuration is valid and adjusts if needed
func (c *SwarmConfig) Validate(swarmSize int) {
	// Ensure connection probability is valid
	if c.ConnectionProbability < 0 {
		c.ConnectionProbability = 0
	} else if c.ConnectionProbability > 1 {
		c.ConnectionProbability = 1
	}

	// Ensure max neighbors doesn't exceed swarm size
	if c.MaxNeighbors > swarmSize-1 {
		c.MaxNeighbors = swarmSize - 1
	}
	if c.MaxNeighbors < 1 {
		c.MaxNeighbors = 1
	}

	// Ensure min neighbors is reasonable
	if c.MinNeighbors > c.MaxNeighbors {
		c.MinNeighbors = c.MaxNeighbors
	}
	if c.MinNeighbors < 0 {
		c.MinNeighbors = 0
	}

	// Validate other parameters
	c.CouplingStrength = clamp(c.CouplingStrength, 0, 1)
	c.Stubbornness = clamp(c.Stubbornness, 0, 1)
	c.BasinStrength = clamp(c.BasinStrength, 0, 1)
	c.BaseConfidence = clamp(c.BaseConfidence, 0, 1)
	c.InfluenceDefault = clamp(c.InfluenceDefault, 0, 1)

	if c.InitialEnergy <= 0 {
		c.InitialEnergy = 100.0
	}

	if c.BasinWidth <= 0 {
		c.BasinWidth = math.Pi
	}
}

// clamp returns value clamped between min and max
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// minFloat returns the minimum of two floats
func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
