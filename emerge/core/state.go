package core

import (
	"fmt"
	"math"
	"time"

	"github.com/carlisia/bio-adapt/internal/config"
)

// State represents a system configuration that agents work toward.
// In biological terms, this is like a morphological target.
type State struct {
	Phase     float64       // Target phase in radians [0, 2π]
	Frequency time.Duration // Target oscillation period
	Coherence float64       // Target synchronization level [0, 1]
}

// Validate checks if the State is valid.
func (s *State) Validate() error {
	var errors config.ValidationErrors

	// Validate frequency
	if s.Frequency <= 0 {
		errors = append(errors, config.ValidationError{
			Field: "Frequency", Value: s.Frequency, Message: "must be positive",
		})
	}

	// Validate coherence
	if math.IsNaN(s.Coherence) {
		errors = append(errors, config.ValidationError{
			Field: "Coherence", Value: s.Coherence, Message: "cannot be NaN",
		})
	} else if s.Coherence < 0 || s.Coherence > 1 {
		errors = append(errors, config.ValidationError{
			Field: "Coherence", Value: s.Coherence, Message: "must be between 0 and 1",
		})
	}

	// Phase can be any value (it will be wrapped), but check for special values
	if math.IsNaN(s.Phase) {
		errors = append(errors, config.ValidationError{
			Field: "Phase", Value: s.Phase, Message: "cannot be NaN",
		})
	}

	if math.IsInf(s.Phase, 0) {
		errors = append(errors, config.ValidationError{
			Field: "Phase", Value: s.Phase, Message: "cannot be infinite",
		})
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// NormalizeAndValidate performs validation and normalization of the State.
func (s *State) NormalizeAndValidate() error {
	// First validate
	if err := s.Validate(); err != nil {
		// Try to auto-correct some issues
		s.normalize()

		// Validate again after normalization
		if err := s.Validate(); err != nil {
			return fmt.Errorf("state validation failed even after normalization: %w", err)
		}
	} else {
		// Even if valid, apply normalization for consistency
		s.normalize()
	}

	return nil
}

// normalize applies auto-corrections to the State.
func (s *State) normalize() {
	// Wrap phase to [0, 2π]
	if !math.IsNaN(s.Phase) && !math.IsInf(s.Phase, 0) {
		for s.Phase < 0 {
			s.Phase += 2 * math.Pi
		}
		for s.Phase >= 2*math.Pi {
			s.Phase -= 2 * math.Pi
		}
	}

	// Clamp coherence to [0, 1]
	if !math.IsNaN(s.Coherence) {
		if s.Coherence < 0 {
			s.Coherence = 0
		} else if s.Coherence > 1 {
			s.Coherence = 1
		}
	}

	// Fix frequency if needed
	if s.Frequency <= 0 {
		s.Frequency = 100 * time.Millisecond // Default frequency
	}
}

// StateUpdate represents a change in agent state for gossip protocol.
type StateUpdate struct {
	AgentID   string
	FromID    string // ID of the agent who sent the update
	Phase     float64
	Frequency time.Duration
	Energy    float64 // Energy level of the agent
	Timestamp time.Time
}
