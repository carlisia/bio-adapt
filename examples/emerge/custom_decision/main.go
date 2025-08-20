// Custom Decision Maker Example
// This example shows how to implement custom decision-making strategies
// for agents, including risk-averse, aggressive, and adaptive strategies.

package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/carlisia/bio-adapt/emerge/agent"
	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/swarm"
)

// RiskAverseDecisionMaker avoids high-cost actions.
type RiskAverseDecisionMaker struct {
	maxAcceptableCost float64
	history           []float64
}

func NewRiskAverseDecisionMaker(maxCost float64) *RiskAverseDecisionMaker {
	return &RiskAverseDecisionMaker{
		maxAcceptableCost: maxCost,
		history:           make([]float64, 0, 100),
	}
}

func (r *RiskAverseDecisionMaker) Decide(state core.State, options []core.Action) (core.Action, float64) {
	if len(options) == 0 {
		return core.Action{Type: "maintain"}, 0.5
	}

	bestAction := options[0]
	bestScore := -math.MaxFloat64

	for _, action := range options {
		// Reject if cost exceeds threshold
		if action.Cost > r.maxAcceptableCost {
			continue
		}

		// Calculate expected value with risk penalty
		expectedValue := action.Benefit - action.Cost
		riskPenalty := action.Cost * 0.5 // Extra penalty for risk
		score := expectedValue - riskPenalty

		// Track history
		r.history = append(r.history, score)
		if len(r.history) > 100 {
			r.history = r.history[1:]
		}

		if score > bestScore && score > 0 {
			bestScore = score
			bestAction = action
		}
	}

	// Confidence based on score
	confidence := math.Min(math.Max(bestScore/2.0, 0), 1.0)
	return bestAction, confidence
}

// AggressiveDecisionMaker takes more risks for higher rewards.
type AggressiveDecisionMaker struct {
	aggressiveness float64
	successCount   atomic.Int32
	totalCount     atomic.Int32
}

func NewAggressiveDecisionMaker(aggressiveness float64) *AggressiveDecisionMaker {
	return &AggressiveDecisionMaker{
		aggressiveness: math.Min(1.0, math.Max(0.0, aggressiveness)),
	}
}

func (a *AggressiveDecisionMaker) Decide(state core.State, options []core.Action) (core.Action, float64) {
	if len(options) == 0 {
		return core.Action{Type: "maintain"}, 0.5
	}

	a.totalCount.Add(1)

	bestAction := options[0]
	bestScore := -math.MaxFloat64

	for _, action := range options {
		// Aggressive agents focus more on benefit than cost
		weightedBenefit := action.Benefit * (1 + a.aggressiveness)
		weightedCost := action.Cost * (1 - a.aggressiveness*0.5)

		score := weightedBenefit - weightedCost

		// More likely to accept risky actions
		threshold := -a.aggressiveness * 2

		if score > threshold && score > bestScore {
			bestScore = score
			bestAction = action
			if score > 0 {
				a.successCount.Add(1)
			}
		}
	}

	// Higher confidence for aggressive strategies
	confidence := math.Min(math.Max((bestScore+2)/4.0, 0), 1.0)
	return bestAction, confidence
}

func (a *AggressiveDecisionMaker) GetSuccessRate() float64 {
	total := a.totalCount.Load()
	if total == 0 {
		return 0
	}
	return float64(a.successCount.Load()) / float64(total)
}

// AdaptiveDecisionMaker learns from past decisions.
type AdaptiveDecisionMaker struct {
	learningRate     float64
	costThreshold    atomic.Value // float64
	benefitThreshold atomic.Value // float64
	recentOutcomes   []float64
}

func NewAdaptiveDecisionMaker(learningRate float64) *AdaptiveDecisionMaker {
	dm := &AdaptiveDecisionMaker{
		learningRate:   learningRate,
		recentOutcomes: make([]float64, 0, 50),
	}
	dm.costThreshold.Store(10.0)
	dm.benefitThreshold.Store(5.0)
	return dm
}

func (a *AdaptiveDecisionMaker) Decide(state core.State, options []core.Action) (core.Action, float64) {
	if len(options) == 0 {
		return core.Action{Type: "maintain"}, 0.5
	}

	var costThresh, benefitThresh float64
	if val := a.costThreshold.Load(); val != nil {
		if ct, ok := val.(float64); ok {
			costThresh = ct
		}
	}
	if val := a.benefitThreshold.Load(); val != nil {
		if bt, ok := val.(float64); ok {
			benefitThresh = bt
		}
	}

	bestAction := options[0]
	bestScore := -math.MaxFloat64

	for _, action := range options {
		// Calculate score based on current thresholds
		score := action.Benefit - action.Cost

		// Record outcome
		a.recentOutcomes = append(a.recentOutcomes, score)
		if len(a.recentOutcomes) > 50 {
			a.recentOutcomes = a.recentOutcomes[1:]
		}

		// Make decision based on adapted thresholds
		if action.Cost <= costThresh && action.Benefit >= benefitThresh && score > bestScore {
			bestScore = score
			bestAction = action
		}
	}

	// Adapt thresholds based on recent performance
	if len(a.recentOutcomes) >= 10 {
		avgOutcome := a.calculateAverage()

		if avgOutcome < 0 {
			// Recent decisions have been poor, be more conservative
			a.costThreshold.Store(costThresh * (1 - a.learningRate))
			a.benefitThreshold.Store(benefitThresh * (1 + a.learningRate))
		} else {
			// Recent decisions have been good, can be more aggressive
			a.costThreshold.Store(costThresh * (1 + a.learningRate*0.5))
			a.benefitThreshold.Store(benefitThresh * (1 - a.learningRate*0.5))
		}
	}

	// Adaptive confidence based on average outcomes
	avgOutcome := a.calculateAverage()
	confidence := math.Min(math.Max((avgOutcome+2)/4.0, 0), 1.0)
	return bestAction, confidence
}

func (a *AdaptiveDecisionMaker) calculateAverage() float64 {
	sum := 0.0
	for _, outcome := range a.recentOutcomes {
		sum += outcome
	}
	return sum / float64(len(a.recentOutcomes))
}

func main() {
	fmt.Println("=== Custom Decision Maker Strategies Example ===")
	fmt.Println()

	// Create agents with different decision strategies
	agents := []*agent.Agent{
		agent.New("risk-averse"),
		agent.New("aggressive"),
		agent.New("adaptive"),
		agent.New("default"),
	}

	// Set custom decision makers
	agents[0].SetDecisionMaker(NewRiskAverseDecisionMaker(5.0))
	agents[1].SetDecisionMaker(NewAggressiveDecisionMaker(0.8))
	agents[2].SetDecisionMaker(NewAdaptiveDecisionMaker(0.1))
	// agents[3] uses the default SimpleDecisionMaker

	// Set initial energy for all agents
	for _, agent := range agents {
		agent.SetEnergy(100)
		agent.SetPhase(rand.Float64() * 2 * math.Pi)
	}

	// Connect agents in a ring topology
	for i := range agents {
		next := (i + 1) % len(agents)
		agents[i].Neighbors().Store(agents[next].ID, agents[next])
		agents[next].Neighbors().Store(agents[i].ID, agents[i])
	}

	// Define target state
	target := core.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	fmt.Println("Agent strategies:")
	fmt.Println("1. Risk-Averse: Avoids high-cost actions")
	fmt.Println("2. Aggressive:  Takes risks for higher rewards")
	fmt.Println("3. Adaptive:    Learns from past decisions")
	fmt.Println("4. Default:     Standard cost-benefit analysis")
	fmt.Println()

	// Track decision statistics
	type stats struct {
		proposed     int
		accepted     int
		totalCost    float64
		totalBenefit float64
	}

	agentStats := make([]stats, len(agents))

	// Simulate decision-making over time
	fmt.Println("Simulating 100 decision cycles...")

	for cycle := range 100 {
		for i, agent := range agents {
			// Update context periodically
			if rand.Float64() < 0.3 {
				agent.UpdateContext()
			}

			// Propose adjustment
			action, accepted := agent.ProposeAdjustment(target)
			agentStats[i].proposed++

			if accepted {
				agentStats[i].accepted++
				agentStats[i].totalCost += action.Cost
				agentStats[i].totalBenefit += action.Benefit
				success, energyCost, err := agent.ApplyAction(action)
				if err != nil {
					// Action failed due to insufficient energy or invalid action type
					// This is expected behavior in energy-constrained systems - continue simulation
					if cycle%25 == 0 { // Only log occasionally to avoid spam
						fmt.Printf("Agent %s action failed: %v\n", agent.ID, err)
					}
				} else if !success {
					// Action was valid but unsuccessful for other reasons - continue
					if cycle%25 == 0 {
						fmt.Printf("Agent %s action unsuccessful (cost: %.1f)\n", agent.ID, energyCost)
					}
				}
			}
		}

		// Occasionally replenish energy
		if cycle%20 == 0 {
			for _, agent := range agents {
				agent.SetEnergy(agent.Energy() + 20)
				if agent.Energy() > 100 {
					agent.SetEnergy(100)
				}
			}
		}
	}

	// Display results
	fmt.Println("\n=== Decision Strategy Performance ===")
	strategies := []string{"Risk-Averse", "Aggressive", "Adaptive", "Default"}

	for i, name := range strategies {
		s := agentStats[i]
		acceptRate := float64(s.accepted) / float64(s.proposed) * 100
		avgCost := s.totalCost / float64(max(s.accepted, 1))
		avgBenefit := s.totalBenefit / float64(max(s.accepted, 1))
		netValue := s.totalBenefit - s.totalCost

		fmt.Printf("\n%s Strategy:\n", name)
		fmt.Printf("  Decisions: %d proposed, %d accepted (%.1f%%)\n",
			s.proposed, s.accepted, acceptRate)
		fmt.Printf("  Avg Cost per action: %.2f\n", avgCost)
		fmt.Printf("  Avg Benefit per action: %.2f\n", avgBenefit)
		fmt.Printf("  Net Value: %.2f\n", netValue)
		fmt.Printf("  Final Energy: %.1f\n", agents[i].Energy())
		fmt.Printf("  Final Phase: %.3f\n", agents[i].Phase())
	}

	// Special reporting for aggressive strategy
	// Note: Can't type assert to our custom type through the interface
	// Would need to track this separately in production code

	// Calculate final coherence
	var sumCos, sumSin float64
	for _, agent := range agents {
		phase := agent.Phase()
		sumCos += math.Cos(phase)
		sumSin += math.Sin(phase)
	}
	coherence := math.Sqrt(sumCos*sumCos+sumSin*sumSin) / float64(len(agents))

	fmt.Printf("\nFinal group coherence: %.3f\n", coherence)
	fmt.Printf("Target coherence: %.3f\n", target.Coherence)

	// Demonstrate real-time strategy comparison
	fmt.Println("\n=== Real-Time Strategy Comparison ===")
	demonstrateRealTimeComparison()
}

func demonstrateRealTimeComparison() {
	// Create a mini-swarm with different strategies
	target := core.State{
		Phase:     0,
		Frequency: 50 * time.Millisecond,
		Coherence: 0.8,
	}

	swarm, err := swarm.New(12, target)
	if err != nil {
		fmt.Printf("Error creating swarm: %v\n", err)
		return
	}

	// Assign different strategies to agents
	strategyCount := map[string]int{
		"risk-averse": 0,
		"aggressive":  0,
		"adaptive":    0,
		"default":     0,
	}

	i := 0
	for _, agent := range swarm.Agents() {
		switch i % 4 {
		case 0:
			agent.SetDecisionMaker(NewRiskAverseDecisionMaker(5.0))
			strategyCount["risk-averse"]++
		case 1:
			agent.SetDecisionMaker(NewAggressiveDecisionMaker(0.7))
			strategyCount["aggressive"]++
		case 2:
			agent.SetDecisionMaker(NewAdaptiveDecisionMaker(0.15))
			strategyCount["adaptive"]++
		default:
			strategyCount["default"]++
		}

		i++
	}

	fmt.Println("Mixed-strategy swarm composition:")
	for strategy, count := range strategyCount {
		fmt.Printf("  %s: %d agents\n", strategy, count)
	}

	// Run brief synchronization
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	initialCoherence := swarm.MeasureCoherence()
	fmt.Printf("\nInitial coherence: %.3f\n", initialCoherence)

	errChan := make(chan error, 1)
	go func() {
		if err := swarm.Run(ctx); err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		fmt.Printf("Error in swarm: %v\n", err)
		return
	case <-time.After(3 * time.Second):
		// Continue after timeout
	}

	finalCoherence := swarm.MeasureCoherence()
	fmt.Printf("Final coherence: %.3f\n", finalCoherence)
	fmt.Printf("Improvement: %.1f%%\n",
		(finalCoherence-initialCoherence)/initialCoherence*100)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
