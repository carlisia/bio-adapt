// Package main demonstrates custom decision-making strategies for agents.
// This example shows how to implement custom decision-making strategies
// for agents, including risk-averse, aggressive, and adaptive strategies.
package main

import (
	"context"
	"fmt"
	"math"
	"sync/atomic"
	"time"

	"github.com/carlisia/bio-adapt/emerge/agent"
	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/swarm"
	"github.com/carlisia/bio-adapt/internal/random"
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

func (r *RiskAverseDecisionMaker) Decide(_ core.State, options []core.Action) (core.Action, float64) {
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

func (a *AggressiveDecisionMaker) Decide(_ core.State, options []core.Action) (core.Action, float64) {
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

func (a *AdaptiveDecisionMaker) Decide(_ core.State, options []core.Action) (core.Action, float64) {
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

// stats tracks decision statistics for an agent.
type stats struct {
	proposed     int
	accepted     int
	totalCost    float64
	totalBenefit float64
}

// setupAgents creates and configures agents with different strategies.
func setupAgents() []*agent.Agent {
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

	// Set initial energy and phase
	for _, a := range agents {
		a.SetEnergy(100)
		a.SetPhase(random.Phase())
	}

	// Connect agents in a ring topology
	for i := range agents {
		next := (i + 1) % len(agents)
		agents[i].Neighbors().Store(agents[next].ID, agents[next])
		agents[next].Neighbors().Store(agents[i].ID, agents[i])
	}

	return agents
}

// printStrategies prints the description of agent strategies.
func printStrategies() {
	fmt.Println("Agent strategies:")
	fmt.Println("1. Risk-Averse: Avoids high-cost actions")
	fmt.Println("2. Aggressive:  Takes risks for higher rewards")
	fmt.Println("3. Adaptive:    Learns from past decisions")
	fmt.Println("4. Default:     Standard cost-benefit analysis")
	fmt.Println()
}

// runSimulation runs the decision-making simulation.
func runSimulation(agents []*agent.Agent, target core.State, cycles int) []stats {
	agentStats := make([]stats, len(agents))

	for cycle := range cycles {
		processCycle(agents, target, agentStats, cycle)

		// Occasionally replenish energy
		if cycle%20 == 0 {
			replenishEnergy(agents)
		}
	}

	return agentStats
}

// processCycle processes a single simulation cycle.
func processCycle(agents []*agent.Agent, target core.State, agentStats []stats, cycle int) {
	for i, a := range agents {
		// Update context periodically
		if random.Float64() < 0.3 {
			a.UpdateContext()
		}

		// Propose adjustment
		action, accepted := a.ProposeAdjustment(target)
		agentStats[i].proposed++

		if accepted {
			processAcceptedAction(a, &agentStats[i], action, cycle)
		}
	}
}

// processAcceptedAction processes an accepted action.
func processAcceptedAction(a *agent.Agent, s *stats, action core.Action, cycle int) {
	success, energyCost, err := a.ApplyAction(action)

	if err != nil {
		// Always handle errors, log periodically
		if cycle%25 == 0 {
			fmt.Printf("Agent %s action failed: %v\n", a.ID, err)
		}
		return // Don't update stats for failed actions
	}

	if !success {
		if cycle%25 == 0 {
			fmt.Printf("Agent %s action unsuccessful (cost: %.1f)\n", a.ID, energyCost)
		}
		return // Don't update stats for unsuccessful actions
	}

	// Only update stats for successful actions
	s.accepted++
	s.totalCost += action.Cost
	s.totalBenefit += action.Benefit
}

// replenishEnergy replenishes agent energy.
func replenishEnergy(agents []*agent.Agent) {
	for _, a := range agents {
		newEnergy := math.Min(a.Energy()+20, 100)
		a.SetEnergy(newEnergy)
	}
}

// displayResults displays the simulation results.
func displayResults(agentStats []stats, agents []*agent.Agent, target core.State) {
	fmt.Println("\n=== Decision Strategy Performance ===")
	strategies := []string{"Risk-Averse", "Aggressive", "Adaptive", "Default"}

	for i, name := range strategies {
		displayStrategyStats(name, agentStats[i], agents[i])
	}

	// Calculate and display final coherence
	coherence := calculateCoherence(agents)
	fmt.Printf("\nFinal group coherence: %.3f\n", coherence)
	fmt.Printf("Target coherence: %.3f\n", target.Coherence)
}

// displayStrategyStats displays statistics for a single strategy.
func displayStrategyStats(name string, s stats, a *agent.Agent) {
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
	fmt.Printf("  Final Energy: %.1f\n", a.Energy())
	fmt.Printf("  Final Phase: %.3f\n", a.Phase())
}

// calculateCoherence calculates the group coherence.
func calculateCoherence(agents []*agent.Agent) float64 {
	var sumCos, sumSin float64
	for _, a := range agents {
		phase := a.Phase()
		sumCos += math.Cos(phase)
		sumSin += math.Sin(phase)
	}
	return math.Sqrt(sumCos*sumCos+sumSin*sumSin) / float64(len(agents))
}

func main() {
	fmt.Println("=== Custom Decision Maker Strategies Example ===")
	fmt.Println()

	// Setup agents with different strategies
	agents := setupAgents()
	target := core.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	printStrategies()

	// Simulate decision-making over time
	fmt.Println("Simulating 100 decision cycles...")
	agentStats := runSimulation(agents, target, 100)

	// Display results
	displayResults(agentStats, agents, target)

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

	s, err := swarm.New(12, target)
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
	for _, a := range s.Agents() {
		switch i % 4 {
		case 0:
			a.SetDecisionMaker(NewRiskAverseDecisionMaker(5.0))
			strategyCount["risk-averse"]++
		case 1:
			a.SetDecisionMaker(NewAggressiveDecisionMaker(0.7))
			strategyCount["aggressive"]++
		case 2:
			a.SetDecisionMaker(NewAdaptiveDecisionMaker(0.15))
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

	initialCoherence := s.MeasureCoherence()
	fmt.Printf("\nInitial coherence: %.3f\n", initialCoherence)

	errChan := make(chan error, 1)
	go func() {
		if err := s.Run(ctx); err != nil {
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

	finalCoherence := s.MeasureCoherence()
	fmt.Printf("Final coherence: %.3f\n", finalCoherence)
	fmt.Printf("Improvement: %.1f%%\n",
		(finalCoherence-initialCoherence)/initialCoherence*100)
}
