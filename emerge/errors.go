package emerge

import "errors"

// Common error sentinel values for consistent error handling throughout the package.
// These errors can be used with errors.Is() for error checking and provide
// clear, well-defined error conditions.
var (
	// Agent operation errors
	ErrInsufficientEnergy = errors.New("insufficient energy")
	ErrUnknownActionType  = errors.New("unknown action type")
	ErrActionFailed       = errors.New("action execution failed")

	// Swarm configuration errors
	ErrInvalidSwarmSize = errors.New("invalid swarm size")
	ErrInvalidGoalState = errors.New("invalid goal state")
	ErrConfigValidation = errors.New("configuration validation failed")

	// Resource management errors
	ErrResourceExhausted = errors.New("resource exhausted")
	ErrInvalidRequest    = errors.New("invalid resource request")

	// Topology errors
	ErrTopologyBuild      = errors.New("topology build failed")
	ErrInsufficientAgents = errors.New("insufficient agents for topology")

	// Network/gossip errors
	ErrNetworkFailure    = errors.New("network operation failed")
	ErrGossipUnavailable = errors.New("gossip network unavailable")
)
