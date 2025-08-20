package strategy

import (
	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/decision"
	"github.com/carlisia/bio-adapt/emerge/goal"
	"github.com/carlisia/bio-adapt/internal/resource"
)

// NewSimpleDecisionMaker creates a simple decision maker.
func NewSimpleDecisionMaker() core.DecisionMaker {
	return &decision.SimpleDecisionMaker{}
}

// NewWeightedGoalManager creates a weighted goal manager.
func NewWeightedGoalManager() goal.Manager {
	return &goal.WeightedManager{}
}

// NewTokenResourceManager creates a token-based resource manager.
func NewTokenResourceManager(maxTokens float64) core.ResourceManager {
	return resource.NewTokenManager(maxTokens)
}

// NewAdaptive creates an adaptive synchronization strategy.
func NewAdaptive(strategies []core.SyncStrategy) core.SyncStrategy {
	return newAdaptiveStrategy(strategies)
}
