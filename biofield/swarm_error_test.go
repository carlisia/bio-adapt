package biofield

import (
	"testing"
	"time"
)

func TestNewSwarmErrors(t *testing.T) {
	tests := []struct {
		name      string
		size      int
		goal      State
		wantError bool
		errorMsg  string
	}{
		{
			name: "negative size",
			size: -1,
			goal: State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			wantError: true,
			errorMsg:  "swarm size must be positive",
		},
		{
			name: "zero size",
			size: 0,
			goal: State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			wantError: true,
			errorMsg:  "swarm size must be positive",
		},
		{
			name: "negative frequency",
			size: 10,
			goal: State{
				Phase:     0,
				Frequency: -100 * time.Millisecond,
				Coherence: 0.9,
			},
			wantError: true,
			errorMsg:  "goal frequency must be positive",
		},
		{
			name: "zero frequency",
			size: 10,
			goal: State{
				Phase:     0,
				Frequency: 0,
				Coherence: 0.9,
			},
			wantError: true,
			errorMsg:  "goal frequency must be positive",
		},
		{
			name: "negative coherence",
			size: 10,
			goal: State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: -0.1,
			},
			wantError: true,
			errorMsg:  "goal coherence must be in [0, 1]",
		},
		{
			name: "coherence > 1",
			size: 10,
			goal: State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 1.1,
			},
			wantError: true,
			errorMsg:  "goal coherence must be in [0, 1]",
		},
		{
			name: "valid parameters",
			size: 10,
			goal: State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			swarm, err := NewSwarm(tt.size, tt.goal)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.errorMsg, err)
				}
				if swarm != nil {
					t.Errorf("Expected nil swarm on error, got %v", swarm)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if swarm == nil {
					t.Errorf("Expected valid swarm, got nil")
				} else if swarm.Size() != tt.size {
					t.Errorf("Expected swarm size %d, got %d", tt.size, swarm.Size())
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && contains(s[1:], substr)
}

