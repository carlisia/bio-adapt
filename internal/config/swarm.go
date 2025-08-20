package config

import (
	"fmt"
	"math"
	"time"
)

// Swarm holds all configurable parameters for swarm behavior.
// This allows fine-tuning for different scales and use cases.
type Swarm struct {
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

	// Scalability and Concurrency
	MaxSwarmSize             int           // Maximum allowed swarm size (0 = unlimited)
	MaxConcurrentAgents      int           // Maximum concurrent goroutines for agents (0 = no limit)
	UseBatchProcessing       bool          // Enable batch processing for large swarms
	BatchSize                int           // Number of agents per batch (0 = auto-calculate)
	WorkerPoolSize           int           // Number of worker goroutines (0 = auto-calculate)
	AgentUpdateInterval      time.Duration // Interval between agent updates
	MonitoringInterval       time.Duration // Interval between monitoring samples
	ConnectionOptimThreshold int           // Swarm size above which to use optimized connections
	EnableConnectionOptim    bool          // Use simplified connection establishment for large swarms

	// Auto-scaling
	AutoScale bool // Automatically adjust parameters based on swarm size
}

// DefaultConfig returns the default configuration for large swarms (100+ agents).
func DefaultConfig() Swarm {
	return Swarm{
		ConnectionProbability:    0.3,
		MaxNeighbors:             5,
		MinNeighbors:             2,
		CouplingStrength:         0.5,
		Stubbornness:             0.2,
		InitialEnergy:            100.0,
		AssumedMaxNeighbors:      20,
		BasinStrength:            0.8,
		BasinWidth:               math.Pi,
		BaseConfidence:           0.6,
		InfluenceDefault:         0.2,                    // Low global influence for Kuramoto coupling
		MaxSwarmSize:             1000000,                // 1M agent limit
		MaxConcurrentAgents:      1000,                   // Limit concurrent goroutines
		UseBatchProcessing:       true,                   // Enable batching for large swarms
		BatchSize:                0,                      // Auto-calculate
		WorkerPoolSize:           0,                      // Auto-calculate
		AgentUpdateInterval:      50 * time.Millisecond,  // Default update rate
		MonitoringInterval:       100 * time.Millisecond, // Default monitoring rate
		ConnectionOptimThreshold: 50000,                  // Optimize connections above 50K
		EnableConnectionOptim:    true,                   // Use optimized connections
		AutoScale:                false,
	}
}

// SmallSwarmConfig returns optimized configuration for small swarms (< 20 agents).
func SmallSwarmConfig() Swarm {
	return Swarm{
		ConnectionProbability:    0.9,  // Almost fully connected
		MaxNeighbors:             10,   // Can connect to most agents
		MinNeighbors:             2,    // Ensure some connectivity
		CouplingStrength:         0.8,  // Strong coupling for faster sync
		Stubbornness:             0.05, // Low resistance to change
		InitialEnergy:            100.0,
		AssumedMaxNeighbors:      0,           // Use actual swarm size
		BasinStrength:            0.9,         // Strong attractor
		BasinWidth:               2 * math.Pi, // Wider basin
		BaseConfidence:           0.8,         // Higher confidence
		InfluenceDefault:         0.1,         // Low global influence for Kuramoto coupling
		MaxSwarmSize:             100,         // Small swarm limit
		MaxConcurrentAgents:      0,           // No limit for small swarms
		UseBatchProcessing:       false,       // No batching needed
		BatchSize:                0,
		WorkerPoolSize:           0,
		AgentUpdateInterval:      25 * time.Millisecond, // Faster updates for small swarms
		MonitoringInterval:       50 * time.Millisecond, // More frequent monitoring
		ConnectionOptimThreshold: 1000,                  // No optimization for small swarms
		EnableConnectionOptim:    false,                 // Full connections for small swarms
		AutoScale:                false,
	}
}

// MediumSwarmConfig returns configuration for medium swarms (20-100 agents).
func MediumSwarmConfig() Swarm {
	return Swarm{
		ConnectionProbability:    0.5,
		MaxNeighbors:             8,
		MinNeighbors:             3,
		CouplingStrength:         0.6,
		Stubbornness:             0.1,
		InitialEnergy:            100.0,
		AssumedMaxNeighbors:      0, // Use actual swarm size
		BasinStrength:            0.85,
		BasinWidth:               1.5 * math.Pi,
		BaseConfidence:           0.7,
		InfluenceDefault:         0.2,   // Low global influence for Kuramoto coupling
		MaxSwarmSize:             1000,  // Medium swarm limit
		MaxConcurrentAgents:      100,   // Moderate concurrency limit
		UseBatchProcessing:       false, // No batching for medium swarms
		BatchSize:                0,
		WorkerPoolSize:           0,
		AgentUpdateInterval:      40 * time.Millisecond, // Balanced update rate
		MonitoringInterval:       75 * time.Millisecond, // Balanced monitoring rate
		ConnectionOptimThreshold: 5000,                  // Optimize at 5K+ agents
		EnableConnectionOptim:    false,                 // Full connections for medium swarms
		AutoScale:                false,
	}
}

// AutoScaleConfig returns a configuration that automatically scales based on swarm size.
func AutoScaleConfig(swarmSize int) Swarm {
	config := Swarm{
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
		config.InfluenceDefault = 0.1 // Low global influence for Kuramoto coupling
		// Concurrency settings for very small swarms
		config.MaxSwarmSize = 50
		config.MaxConcurrentAgents = 0 // No limit
		config.UseBatchProcessing = false
		config.AgentUpdateInterval = 20 * time.Millisecond
		config.MonitoringInterval = 25 * time.Millisecond
		config.ConnectionOptimThreshold = 1000
		config.EnableConnectionOptim = false
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
	} else if swarmSize < 1000 {
		// Large swarm
		config = DefaultConfig()
		config.MaxNeighbors = minInt(20, swarmSize/10)
		config.AutoScale = true
		// For proper Kuramoto synchronization
		config.CouplingStrength = 0.7 // Strong coupling for synchronization
		config.InfluenceDefault = 0.2 // Favor local (neighbor) goals over global
		config.Stubbornness = 0.1     // Less rejection for faster convergence
		// Scale concurrency for large swarms
		config.MaxConcurrentAgents = minInt(500, swarmSize/2)
		config.UseBatchProcessing = swarmSize > 200
		config.BatchSize = maxInt(10, swarmSize/50)
		config.WorkerPoolSize = minInt(50, swarmSize/20)
	} else {
		// Very large swarm (1000+)
		config = DefaultConfig()
		config.MaxNeighbors = minInt(10, swarmSize/100) // Fewer neighbors for very large swarms
		config.AutoScale = true
		// Heavy optimization for very large swarms
		config.MaxConcurrentAgents = 1000 // Fixed limit
		config.UseBatchProcessing = true
		config.BatchSize = maxInt(50, swarmSize/100)
		config.WorkerPoolSize = 100                         // Fixed worker pool
		config.AgentUpdateInterval = 100 * time.Millisecond // Slower updates
		config.MonitoringInterval = 250 * time.Millisecond  // Less frequent monitoring
		config.ConnectionOptimThreshold = 1000              // Always use optimized connections
		config.EnableConnectionOptim = true
	}

	// Always use actual swarm size for density when auto-scaling
	config.AssumedMaxNeighbors = 0

	// Scale coupling inversely with size for stability
	if swarmSize > 100 {
		config.CouplingStrength *= 100.0 / float64(swarmSize)
	}

	return config
}

// ConfigForBatching returns configuration optimized for request batching scenarios.
func ConfigForBatching(workloadCount int, batchWindow time.Duration) Swarm {
	// For batching, we need very strong synchronization
	// Start with a small swarm config as baseline since batching needs tight coupling
	var config Swarm
	if workloadCount <= 30 {
		// For small to medium batching scenarios, use very strong coupling
		config = Swarm{
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

// Validate performs comprehensive validation and returns errors instead of modifying the config.
func (c *Swarm) Validate(swarmSize int) error {
	var errors ValidationErrors

	// Validate swarm size constraints
	if swarmSize <= 0 {
		errors = append(errors, ValidationError{
			Field: "swarmSize", Value: swarmSize, Message: "must be positive",
		})
	}

	if c.MaxSwarmSize > 0 && swarmSize > c.MaxSwarmSize {
		errors = append(errors, ValidationError{
			Field: "MaxSwarmSize", Value: fmt.Sprintf("swarmSize=%d, MaxSwarmSize=%d", swarmSize, c.MaxSwarmSize),
			Message: "swarm size exceeds configured maximum",
		})
	}

	// Validate probability parameters (must be in [0,1])
	if c.ConnectionProbability < 0 || c.ConnectionProbability > 1 {
		errors = append(errors, ValidationError{
			Field: "ConnectionProbability", Value: c.ConnectionProbability, Message: "must be between 0 and 1",
		})
	}

	if c.CouplingStrength < 0 || c.CouplingStrength > 1 {
		errors = append(errors, ValidationError{
			Field: "CouplingStrength", Value: c.CouplingStrength, Message: "must be between 0 and 1",
		})
	}

	if c.Stubbornness < 0 || c.Stubbornness > 1 {
		errors = append(errors, ValidationError{
			Field: "Stubbornness", Value: c.Stubbornness, Message: "must be between 0 and 1",
		})
	}

	if c.BasinStrength < 0 || c.BasinStrength > 1 {
		errors = append(errors, ValidationError{
			Field: "BasinStrength", Value: c.BasinStrength, Message: "must be between 0 and 1",
		})
	}

	if c.BaseConfidence < 0 || c.BaseConfidence > 1 {
		errors = append(errors, ValidationError{
			Field: "BaseConfidence", Value: c.BaseConfidence, Message: "must be between 0 and 1",
		})
	}

	if c.InfluenceDefault < 0 || c.InfluenceDefault > 1 {
		errors = append(errors, ValidationError{
			Field: "InfluenceDefault", Value: c.InfluenceDefault, Message: "must be between 0 and 1",
		})
	}

	// Validate neighbor constraints
	maxPossibleNeighbors := swarmSize - 1
	if swarmSize > 0 && maxPossibleNeighbors <= 0 {
		// Single agent swarm - MaxNeighbors must be 0
		if c.MaxNeighbors != 0 {
			errors = append(errors, ValidationError{
				Field: "MaxNeighbors", Value: c.MaxNeighbors, Message: "must be 0 for single agent swarm",
			})
		}
	} else {
		// Multi-agent swarm - MaxNeighbors must be at least 1
		if c.MaxNeighbors < 1 {
			errors = append(errors, ValidationError{
				Field: "MaxNeighbors", Value: c.MaxNeighbors, Message: "must be at least 1",
			})
		}
	}

	if swarmSize > 0 && c.MaxNeighbors > swarmSize-1 {
		errors = append(errors, ValidationError{
			Field: "MaxNeighbors", Value: fmt.Sprintf("MaxNeighbors=%d, swarmSize=%d", c.MaxNeighbors, swarmSize),
			Message: "cannot exceed swarm size - 1",
		})
	}

	if c.MinNeighbors < 0 {
		errors = append(errors, ValidationError{
			Field: "MinNeighbors", Value: c.MinNeighbors, Message: "cannot be negative",
		})
	}

	if c.MinNeighbors > c.MaxNeighbors {
		errors = append(errors, ValidationError{
			Field: "MinNeighbors", Value: fmt.Sprintf("MinNeighbors=%d, MaxNeighbors=%d", c.MinNeighbors, c.MaxNeighbors),
			Message: "cannot exceed MaxNeighbors",
		})
	}

	// Validate energy and time parameters
	if c.InitialEnergy <= 0 {
		errors = append(errors, ValidationError{
			Field: "InitialEnergy", Value: c.InitialEnergy, Message: "must be positive",
		})
	}

	if c.BasinWidth <= 0 {
		errors = append(errors, ValidationError{
			Field: "BasinWidth", Value: c.BasinWidth, Message: "must be positive",
		})
	}

	if c.AgentUpdateInterval <= 0 {
		errors = append(errors, ValidationError{
			Field: "AgentUpdateInterval", Value: c.AgentUpdateInterval, Message: "must be positive",
		})
	}

	if c.MonitoringInterval <= 0 {
		errors = append(errors, ValidationError{
			Field: "MonitoringInterval", Value: c.MonitoringInterval, Message: "must be positive",
		})
	}

	// Validate concurrency parameters
	if c.MaxConcurrentAgents < 0 {
		errors = append(errors, ValidationError{
			Field: "MaxConcurrentAgents", Value: c.MaxConcurrentAgents, Message: "cannot be negative",
		})
	}

	if c.BatchSize < 0 {
		errors = append(errors, ValidationError{
			Field: "BatchSize", Value: c.BatchSize, Message: "cannot be negative",
		})
	}

	if c.WorkerPoolSize < 0 {
		errors = append(errors, ValidationError{
			Field: "WorkerPoolSize", Value: c.WorkerPoolSize, Message: "cannot be negative",
		})
	}

	if c.ConnectionOptimThreshold < 0 {
		errors = append(errors, ValidationError{
			Field: "ConnectionOptimThreshold", Value: c.ConnectionOptimThreshold, Message: "cannot be negative",
		})
	}

	// Return errors if any
	if len(errors) > 0 {
		return errors
	}

	return nil
}

// NormalizeAndValidate performs validation and auto-corrects values where possible.
func (c *Swarm) NormalizeAndValidate(swarmSize int) error {
	// First validate and get errors
	if err := c.Validate(swarmSize); err != nil {
		// Try to auto-correct correctable issues
		c.normalize(swarmSize)

		// Validate again after normalization
		if err := c.Validate(swarmSize); err != nil {
			return fmt.Errorf("configuration validation failed even after normalization: %w", err)
		}
	} else {
		// Even if valid, apply normalization for consistency
		c.normalize(swarmSize)
	}

	return nil
}

// normalize applies auto-corrections and auto-calculations.
func (c *Swarm) normalize(swarmSize int) {
	// Clamp probability values to [0,1]
	c.ConnectionProbability = clamp(c.ConnectionProbability, 0, 1)
	c.CouplingStrength = clamp(c.CouplingStrength, 0, 1)
	c.Stubbornness = clamp(c.Stubbornness, 0, 1)
	c.BasinStrength = clamp(c.BasinStrength, 0, 1)
	c.BaseConfidence = clamp(c.BaseConfidence, 0, 1)
	c.InfluenceDefault = clamp(c.InfluenceDefault, 0, 1)

	// Fix neighbor constraints
	if swarmSize > 0 {
		maxPossibleNeighbors := swarmSize - 1
		if maxPossibleNeighbors <= 0 {
			// Single agent swarm - no neighbors possible
			c.MaxNeighbors = 0
			c.MinNeighbors = 0
		} else {
			if c.MaxNeighbors > maxPossibleNeighbors {
				c.MaxNeighbors = maxPossibleNeighbors
			}
			if c.MaxNeighbors < 1 {
				c.MaxNeighbors = 1
			}
		}
	} else {
		// If swarmSize is invalid, set minimal neighbors
		if c.MaxNeighbors < 1 {
			c.MaxNeighbors = 1
		}
	}
	if c.MinNeighbors > c.MaxNeighbors {
		c.MinNeighbors = c.MaxNeighbors
	}
	if c.MinNeighbors < 0 {
		c.MinNeighbors = 0
	}

	// Fix energy and time parameters
	if c.InitialEnergy <= 0 {
		c.InitialEnergy = 100.0
	}
	if c.BasinWidth <= 0 {
		c.BasinWidth = math.Pi
	}
	if c.AgentUpdateInterval <= 0 {
		c.AgentUpdateInterval = 50 * time.Millisecond
	}
	if c.MonitoringInterval <= 0 {
		c.MonitoringInterval = 100 * time.Millisecond
	}

	// Fix concurrency parameters
	if c.MaxConcurrentAgents < 0 {
		c.MaxConcurrentAgents = 0
	}
	if c.BatchSize < 0 {
		c.BatchSize = 0
	}
	if c.WorkerPoolSize < 0 {
		c.WorkerPoolSize = 0
	}
	if c.ConnectionOptimThreshold < 0 {
		c.ConnectionOptimThreshold = 0
	}

	// Auto-calculate batch size if needed
	if c.UseBatchProcessing && c.BatchSize == 0 {
		c.BatchSize = maxInt(10, swarmSize/50)
	}

	// Auto-calculate worker pool size if needed
	if c.UseBatchProcessing && c.WorkerPoolSize == 0 {
		c.WorkerPoolSize = maxInt(4, minInt(100, swarmSize/20))
	}
}

// clamp returns value clamped between min and max.
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// minInt returns the minimum of two integers.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// minFloat returns the minimum of two floats.
func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// maxInt returns the maximum of two integers.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
