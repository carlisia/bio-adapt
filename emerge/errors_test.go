package emerge

import (
	"errors"
	"testing"
)

func TestErrorSentinelValues(t *testing.T) {
	tests := []struct {
		name          string
		setupFn       func() error
		expectedError error
		description   string
	}{
		{
			name: "insufficient energy error",
			setupFn: func() error {
				agent := NewAgent("test")
				agent.SetEnergy(1.0)
				_, _, err := agent.ApplyAction(Action{
					Type: "adjust_phase",
					Cost: 5.0,
				})
				return err
			},
			expectedError: ErrInsufficientEnergy,
			description:   "should return ErrInsufficientEnergy when agent lacks energy",
		},
		{
			name: "unknown action type error",
			setupFn: func() error {
				agent := NewAgent("test")
				_, _, err := agent.ApplyAction(Action{
					Type: "unknown_action",
					Cost: 1.0,
				})
				return err
			},
			expectedError: ErrUnknownActionType,
			description:   "should return ErrUnknownActionType for invalid action",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setupFn()

			if err == nil {
				t.Errorf("%s: expected error, got nil", tt.description)
				return
			}

			if !errors.Is(err, tt.expectedError) {
				t.Errorf("%s: expected error to be %v, got %v", tt.description, tt.expectedError, err)
			}

			// Verify the error has meaningful context
			if err.Error() == tt.expectedError.Error() {
				t.Errorf("%s: error should have additional context beyond sentinel value", tt.description)
			}
		})
	}
}

func TestErrorSentinelValuesUnwrapping(t *testing.T) {
	// Test that wrapped errors can be checked with errors.Is()
	agent := NewAgent("test")
	agent.SetEnergy(1.0)
	_, _, err := agent.ApplyAction(Action{
		Type: "adjust_phase",
		Cost: 10.0,
	})

	if err == nil {
		t.Fatal("expected error for insufficient energy")
	}

	// Test that errors.Is works with wrapped errors
	if !errors.Is(err, ErrInsufficientEnergy) {
		t.Errorf("errors.Is should find ErrInsufficientEnergy in wrapped error")
	}

	// Test that the error message contains context
	expectedSubstrings := []string{"insufficient energy", "required", "available"}
	for _, substr := range expectedSubstrings {
		if !containsString(err.Error(), substr) {
			t.Errorf("error message should contain '%s', got: %s", substr, err.Error())
		}
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsStringHelper(s, substr))))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
