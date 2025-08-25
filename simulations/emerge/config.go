package main

import (
	"time"

	"github.com/carlisia/bio-adapt/emerge/goal"
	"github.com/carlisia/bio-adapt/emerge/scale"
	"github.com/carlisia/bio-adapt/simulations/emerge/simulation/pattern"
)

// Config holds runtime configuration for the minimize_api_calls simulation.
type Config struct {
	GoalType       goal.Type     // Optimization goal
	Scale          scale.Size    // Swarm size (Tiny, Small, Medium, Large, Huge)
	Pattern        pattern.Type  // Request pattern (HighFrequency, Burst, Steady, Mixed, Sparse)
	UpdateInterval time.Duration // Display update interval
	Timeout        time.Duration // Maximum runtime (0 for no timeout)
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		GoalType:       goal.MinimizeAPICalls,
		Scale:          scale.Tiny,
		Pattern:        pattern.Unset, // Will auto-select based on goal
		UpdateInterval: 100 * time.Millisecond,
		Timeout:        5 * time.Minute,
	}
}

// Goal returns the goal for this simulation.
func (c Config) Goal() goal.Type {
	return c.GoalType
}

// DisplayName returns a display name based on the scale.
func (c Config) DisplayName() string {
	switch c.Scale {
	case scale.Tiny:
		return "TINY SWARM (20 workloads)"
	case scale.Small:
		return "SMALL SWARM (50 workloads)"
	case scale.Medium:
		return "MEDIUM SWARM (200 workloads)"
	case scale.Large:
		return "LARGE SWARM (1000 workloads)"
	case scale.Huge:
		return "HUGE SWARM (2000 workloads)"
	default:
		return "CUSTOM SWARM"
	}
}

// ParseScale converts a string to a scale.Size.
func ParseScale(s string) (scale.Size, bool) {
	switch s {
	case "tiny":
		return scale.Tiny, true
	case "small":
		return scale.Small, true
	case "medium":
		return scale.Medium, true
	case "large":
		return scale.Large, true
	case "huge":
		return scale.Huge, true
	default:
		return scale.Tiny, false
	}
}
