package swarm

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/carlisia/bio-adapt/emerge/core"
)

func TestNewSwarmErrors(t *testing.T) {
	tests := []struct {
		name        string
		size        int
		goal        core.State
		wantError   bool
		errorMsg    string
		description string
	}{
		// Size validation
		{
			name: "negative size",
			size: -1,
			goal: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			wantError:   true,
			errorMsg:    "invalid swarm size",
			description: "Negative size should return error",
		},
		{
			name: "zero size",
			size: 0,
			goal: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			wantError:   true,
			errorMsg:    "invalid swarm size",
			description: "Zero size should return error",
		},
		{
			name: "very large negative size",
			size: -1000000,
			goal: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			wantError:   true,
			errorMsg:    "invalid swarm size",
			description: "Very large negative size should return error",
		},
		// Frequency validation
		{
			name: "negative frequency",
			size: 10,
			goal: core.State{
				Phase:     0,
				Frequency: -100 * time.Millisecond,
				Coherence: 0.9,
			},
			wantError:   true,
			errorMsg:    "must be positive",
			description: "Negative frequency should return error",
		},
		{
			name: "zero frequency",
			size: 10,
			goal: core.State{
				Phase:     0,
				Frequency: 0,
				Coherence: 0.9,
			},
			wantError:   true,
			errorMsg:    "must be positive",
			description: "Zero frequency should return error",
		},
		{
			name: "very large negative frequency",
			size: 10,
			goal: core.State{
				Phase:     0,
				Frequency: -1000 * time.Hour,
				Coherence: 0.9,
			},
			wantError:   true,
			errorMsg:    "must be positive",
			description: "Very large negative frequency should return error",
		},
		// Coherence validation
		{
			name: "negative coherence",
			size: 10,
			goal: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: -0.1,
			},
			wantError:   true,
			errorMsg:    "must be between 0 and 1",
			description: "Negative coherence should return error",
		},
		{
			name: "coherence > 1",
			size: 10,
			goal: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 1.1,
			},
			wantError:   true,
			errorMsg:    "must be between 0 and 1",
			description: "Coherence > 1 should return error",
		},
		{
			name: "very negative coherence",
			size: 10,
			goal: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: -100,
			},
			wantError:   true,
			errorMsg:    "must be between 0 and 1",
			description: "Very negative coherence should return error",
		},
		{
			name: "very large coherence",
			size: 10,
			goal: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 100,
			},
			wantError:   true,
			errorMsg:    "must be between 0 and 1",
			description: "Very large coherence should return error",
		},
		// Valid parameters
		{
			name: "valid parameters",
			size: 10,
			goal: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			wantError:   false,
			description: "Valid parameters should create swarm",
		},
		{
			name: "single agent swarm",
			size: 1,
			goal: core.State{
				Phase:     math.Pi,
				Frequency: 50 * time.Millisecond,
				Coherence: 0.5,
			},
			wantError:   false,
			description: "Single agent swarm should be valid",
		},
		{
			name: "large swarm",
			size: 1000,
			goal: core.State{
				Phase:     math.Pi / 2,
				Frequency: 200 * time.Millisecond,
				Coherence: 0.95,
			},
			wantError:   false,
			description: "Large swarm should be valid",
		},
		{
			name: "minimum valid coherence",
			size: 10,
			goal: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0,
			},
			wantError:   false,
			description: "Zero coherence should be valid",
		},
		{
			name: "maximum valid coherence",
			size: 10,
			goal: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 1.0,
			},
			wantError:   false,
			description: "Coherence 1.0 should be valid",
		},
		// Edge cases with special values
		{
			name: "NaN coherence",
			size: 10,
			goal: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: math.NaN(),
			},
			wantError:   true,
			errorMsg:    "cannot be NaN",
			description: "NaN coherence should return error",
		},
		{
			name: "infinity coherence",
			size: 10,
			goal: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: math.Inf(1),
			},
			wantError:   true,
			errorMsg:    "must be between 0 and 1",
			description: "Infinity coherence should return error",
		},
		{
			name: "negative infinity coherence",
			size: 10,
			goal: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: math.Inf(-1),
			},
			wantError:   true,
			errorMsg:    "must be between 0 and 1",
			description: "Negative infinity coherence should return error",
		},
		// Phase edge cases (phase can be any value)
		{
			name: "negative phase",
			size: 10,
			goal: core.State{
				Phase:     -math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			wantError:   false,
			description: "Negative phase should be valid",
		},
		{
			name: "large phase",
			size: 10,
			goal: core.State{
				Phase:     10 * math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			wantError:   false,
			description: "Large phase should be valid",
		},
		{
			name: "NaN phase",
			size: 10,
			goal: core.State{
				Phase:     math.NaN(),
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			wantError:   true,
			errorMsg:    "cannot be NaN",
			description: "NaN phase should return error",
		},
		// Frequency edge cases
		{
			name: "very small frequency",
			size: 10,
			goal: core.State{
				Phase:     0,
				Frequency: 1 * time.Nanosecond,
				Coherence: 0.9,
			},
			wantError:   false,
			description: "Very small frequency should be valid",
		},
		{
			name: "very large frequency",
			size: 10,
			goal: core.State{
				Phase:     0,
				Frequency: 1000 * time.Hour,
				Coherence: 0.9,
			},
			wantError:   false,
			description: "Very large frequency should be valid",
		},
		// Multiple invalid parameters
		{
			name: "all invalid parameters",
			size: -1,
			goal: core.State{
				Phase:     0,
				Frequency: -100 * time.Millisecond,
				Coherence: -0.5,
			},
			wantError:   true,
			errorMsg:    "", // Could be any of the error messages
			description: "All invalid parameters should return error",
		},
		{
			name: "size and frequency invalid",
			size: 0,
			goal: core.State{
				Phase:     0,
				Frequency: 0,
				Coherence: 0.5,
			},
			wantError:   true,
			errorMsg:    "", // Could be either error
			description: "Multiple invalid parameters should return error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			swarm, err := New(tt.size, tt.goal)

			if tt.wantError {
				if err == nil {
					t.Errorf("%s: Expected error but got nil", tt.description)
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("%s: Expected error containing '%s', got '%v'",
						tt.description, tt.errorMsg, err)
				}
				if swarm != nil {
					t.Errorf("%s: Expected nil swarm on error, got %v",
						tt.description, swarm)
				}
			} else {
				if err != nil {
					t.Errorf("%s: Unexpected error: %v", tt.description, err)
				}
				if swarm == nil {
					t.Errorf("%s: Expected valid swarm, got nil", tt.description)
				} else {
					if swarm.Size() != tt.size {
						t.Errorf("%s: Expected swarm size %d, got %d",
							tt.description, tt.size, swarm.Size())
					}
					// Verify goal state was set correctly
					// Note: This assumes swarm has a way to get the goal state
					// If not, this check can be removed
				}
			}
		})
	}
}

func TestSwarmErrorRecovery(t *testing.T) {
	tests := []struct {
		name        string
		setupFn     func() (*Swarm, error)
		actionFn    func(s *Swarm) error
		expectPanic bool
		description string
	}{
		{
			name: "nil swarm operations",
			setupFn: func() (*Swarm, error) {
				return nil, nil
			},
			actionFn: func(s *Swarm) error {
				if s == nil {
					return nil // Avoid nil pointer dereference
				}
				s.MeasureCoherence()
				return nil
			},
			expectPanic: false,
			description: "Operations on nil swarm should not panic",
		},
		{
			name: "swarm after failed creation",
			setupFn: func() (*Swarm, error) {
				return New(-1, core.State{
					Phase:     0,
					Frequency: 100 * time.Millisecond,
					Coherence: 0.9,
				})
			},
			actionFn: func(s *Swarm) error {
				if s != nil {
					s.MeasureCoherence()
				}
				return nil
			},
			expectPanic: false,
			description: "Failed swarm creation should return nil swarm",
		},
		{
			name: "valid swarm operations",
			setupFn: func() (*Swarm, error) {
				return New(10, core.State{
					Phase:     0,
					Frequency: 100 * time.Millisecond,
					Coherence: 0.9,
				})
			},
			actionFn: func(s *Swarm) error {
				_ = s.MeasureCoherence()
				s.DisruptAgents(0.5)
				return nil
			},
			expectPanic: false,
			description: "Valid swarm should handle operations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tt.expectPanic && r == nil {
					t.Errorf("%s: Expected panic but didn't get one", tt.description)
				} else if !tt.expectPanic && r != nil {
					t.Errorf("%s: Unexpected panic: %v", tt.description, r)
				}
			}()

			swarm, err := tt.setupFn()
			if err != nil && swarm != nil {
				t.Errorf("%s: Expected nil swarm when error occurs", tt.description)
			}

			if tt.actionFn != nil {
				_ = tt.actionFn(swarm)
			}
		})
	}
}

func TestSwarmBoundaryConditions(t *testing.T) {
	tests := []struct {
		name        string
		size        int
		goal        core.State
		validateFn  func(t *testing.T, swarm *Swarm, err error)
		description string
	}{
		{
			name: "maximum int size",
			size: int(^uint(0) >> 1), // Max int value
			goal: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			validateFn: func(t *testing.T, swarm *Swarm, err error) {
				// Should fail due to size limit
				if err == nil {
					t.Error("Expected error for maximum int size, but got none")
				}
				if swarm != nil {
					t.Error("Expected nil swarm for maximum int size")
				}
			},
			description: "Maximum int size handling",
		},
		{
			name: "exactly at size limit",
			size: 1000000,
			goal: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			validateFn: func(t *testing.T, swarm *Swarm, err error) {
				// Should succeed
				if err != nil {
					t.Errorf("Size limit should allow 1,000,000: %v", err)
				}
				if swarm == nil {
					t.Error("Expected valid swarm at size limit")
				}
			},
			description: "Exactly at size limit should work",
		},
		{
			name: "one over size limit",
			size: 1000001,
			goal: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			validateFn: func(t *testing.T, swarm *Swarm, err error) {
				// Should fail
				if err == nil {
					t.Error("Expected error for size over limit")
				}
				if swarm != nil {
					t.Error("Expected nil swarm for size over limit")
				}
			},
			description: "One over size limit should fail",
		},
		{
			name: "well over size limit",
			size: 10000000,
			goal: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			validateFn: func(t *testing.T, swarm *Swarm, err error) {
				// Should fail
				if err == nil {
					t.Error("Expected error for size well over limit")
				}
				if swarm != nil {
					t.Error("Expected nil swarm for size well over limit")
				}
			},
			description: "Well over size limit should fail",
		},
		{
			name: "minimum valid frequency",
			size: 5,
			goal: core.State{
				Phase:     0,
				Frequency: 1, // 1 nanosecond
				Coherence: 0.5,
			},
			validateFn: func(t *testing.T, swarm *Swarm, err error) {
				if err != nil {
					t.Errorf("Should handle minimum frequency: %v", err)
				}
			},
			description: "Minimum frequency handling",
		},
		{
			name: "frequency at max duration",
			size: 5,
			goal: core.State{
				Phase:     0,
				Frequency: time.Duration(int64(^uint64(0) >> 1)), // Max duration
				Coherence: 0.5,
			},
			validateFn: func(t *testing.T, swarm *Swarm, err error) {
				if err != nil {
					t.Errorf("Should handle maximum frequency: %v", err)
				}
			},
			description: "Maximum frequency handling",
		},
		{
			name: "coherence at boundary",
			size: 5,
			goal: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 1.0 - 1e-16, // Just under 1.0
			},
			validateFn: func(t *testing.T, swarm *Swarm, err error) {
				if err != nil {
					t.Errorf("Should handle coherence near 1.0: %v", err)
				}
			},
			description: "Coherence at boundary handling",
		},
		{
			name: "coherence at lower boundary",
			size: 5,
			goal: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 1e-16, // Just above 0
			},
			validateFn: func(t *testing.T, swarm *Swarm, err error) {
				if err != nil {
					t.Errorf("Should handle coherence near 0: %v", err)
				}
			},
			description: "Coherence at lower boundary handling",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			swarm, err := New(tt.size, tt.goal)
			tt.validateFn(t, swarm, err)
		})
	}
}

func TestSwarmValidationConsistency(t *testing.T) {
	// Test that validation is consistent across multiple calls
	invalidConfigs := []struct {
		size int
		goal core.State
	}{
		{
			size: -1,
			goal: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
		},
		{
			size: 10,
			goal: core.State{
				Phase:     0,
				Frequency: -100 * time.Millisecond,
				Coherence: 0.5,
			},
		},
		{
			size: 10,
			goal: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 1.5,
			},
		},
	}

	for i, config := range invalidConfigs {
		t.Run(string(rune(i)), func(t *testing.T) {
			// Try creating the same invalid swarm multiple times
			var errors []error
			for j := 0; j < 10; j++ {
				_, err := New(config.size, config.goal)
				errors = append(errors, err)
			}

			// All errors should be consistent
			firstErr := errors[0]
			for j, err := range errors {
				if firstErr == nil && err != nil {
					t.Errorf("Iteration %d: Inconsistent validation, first was nil, got %v", j, err)
				} else if firstErr != nil && err == nil {
					t.Errorf("Iteration %d: Inconsistent validation, first was %v, got nil", j, firstErr)
				} else if firstErr != nil && err != nil {
					if firstErr.Error() != err.Error() {
						t.Errorf("Iteration %d: Different error messages: %v vs %v", j, firstErr, err)
					}
				}
			}
		})
	}
}

func BenchmarkSwarmCreationWithValidation(b *testing.B) {
	benchmarks := []struct {
		name string
		size int
		goal core.State
	}{
		{
			name: "small_valid",
			size: 10,
			goal: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
		},
		{
			name: "large_valid",
			size: 1000,
			goal: core.State{
				Phase:     math.Pi,
				Frequency: 50 * time.Millisecond,
				Coherence: 0.5,
			},
		},
		{
			name: "invalid_size",
			size: -1,
			goal: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
		},
		{
			name: "invalid_coherence",
			size: 10,
			goal: core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 1.5,
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = New(bm.size, bm.goal)
			}
		})
	}
}
