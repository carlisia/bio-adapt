package agent_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/carlisia/bio-adapt/emerge/agent"
	"github.com/carlisia/bio-adapt/emerge/core"
)

func TestErrorSentinelValues(t *testing.T) {
	t.Run("insufficient energy error", func(t *testing.T) {
		a := agent.New("test")
		a.SetEnergy(1.0)
		_, _, err := a.ApplyAction(core.Action{
			Type: "adjust_phase",
			Cost: 5.0,
		})

		require.Error(t, err, "should return error when agent lacks energy")
		assert.ErrorIs(t, err, core.ErrInsufficientEnergy, "should return ErrInsufficientEnergy")
		assert.NotEqual(t, err.Error(), core.ErrInsufficientEnergy.Error(), "error should have additional context")
	})

	t.Run("unknown action type error", func(t *testing.T) {
		a := agent.New("test")
		_, _, err := a.ApplyAction(core.Action{
			Type: "unknown_action",
			Cost: 1.0,
		})

		require.Error(t, err, "should return error for invalid action")
		assert.ErrorIs(t, err, core.ErrUnknownActionType, "should return ErrUnknownActionType")
		assert.NotEqual(t, err.Error(), core.ErrUnknownActionType.Error(), "error should have additional context")
	})
}

func TestErrorSentinelValuesUnwrapping(t *testing.T) {
	// Test that wrapped errors can be checked with errors.Is()
	a := agent.New("test")
	a.SetEnergy(1.0)
	_, _, err := a.ApplyAction(core.Action{
		Type: "adjust_phase",
		Cost: 10.0,
	})

	require.Error(t, err, "expected error for insufficient energy")
	assert.ErrorIs(t, err, core.ErrInsufficientEnergy, "errors.Is should find ErrInsufficientEnergy in wrapped error")

	// Test that the error message contains context
	errorMsg := err.Error()
	assert.Contains(t, errorMsg, "insufficient energy", "error message should contain 'insufficient energy'")
	assert.Contains(t, errorMsg, "required", "error message should contain 'required'")
	assert.Contains(t, errorMsg, "available", "error message should contain 'available'")
}
