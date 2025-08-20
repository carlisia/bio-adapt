// Package main demonstrates energy management in agent synchronization.
// This example demonstrates how agents manage their energy resources
// and how energy constraints affect synchronization behavior.
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/carlisia/bio-adapt/emerge/agent"
	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/swarm"
	"github.com/carlisia/bio-adapt/internal/random"
)

// energyStats holds energy statistics for the swarm.
type energyStats struct {
	totalEnergy     float64
	minEnergy       float64
	maxEnergy       float64
	avgEnergy       float64
	exhaustedAgents int
}

// setupEnergySwarm creates and configures a swarm with energy profiles.
func setupEnergySwarm() (*swarm.Swarm, core.State) {
	target := core.State{
		Phase:     0,
		Frequency: 150 * time.Millisecond,
		Coherence: 0.85,
	}

	swarmSize := 30
	s, err := swarm.New(swarmSize, target)
	if err != nil {
		fmt.Printf("Error creating swarm: %v\n", err)
		panic(err)
	}

	fmt.Println("Configuring agents with different energy profiles:")
	configureEnergyProfiles(s)

	return s, target
}

// configureEnergyProfiles sets up agents with different energy levels.
func configureEnergyProfiles(s *swarm.Swarm) {
	agentCount := 0
	for _, a := range s.Agents() {
		configureAgent(a, agentCount)
		agentCount++
	}
}

// configureAgent configures a single agent based on its position.
func configureAgent(a *agent.Agent, index int) {
	var energy float64
	var influence float64
	var profile string

	switch {
	case index < 10:
		energy = 100
		influence = 0.8
		profile = "High energy (100), high influence"
	case index < 20:
		energy = 50
		influence = 0.5
		profile = "Medium energy (50), medium influence"
	default:
		energy = 20
		influence = 0.3
		profile = "Low energy (20), low influence"
	}

	a.SetEnergy(energy)
	a.SetInfluence(influence)

	idDisplay := a.ID
	if len(idDisplay) > 8 {
		idDisplay = idDisplay[:8]
	}
	fmt.Printf("  Agent %s: %s\n", idDisplay, profile)
}

// runEnergyMonitoring runs the synchronization with energy monitoring.
func runEnergyMonitoring(s *swarm.Swarm, target core.State) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	fmt.Println("\nStarting energy-aware synchronization...")

	// Start swarm
	errChan := startSwarmAsync(ctx, s)

	// Monitor progress
	monitorProgress(ctx, s, target, errChan, cancel)
}

// startSwarmAsync starts the swarm in a goroutine.
func startSwarmAsync(ctx context.Context, s *swarm.Swarm) chan error {
	errChan := make(chan error, 1)
	go func() {
		if err := s.Run(ctx); err != nil {
			errChan <- err
		}
	}()
	return errChan
}

// monitorProgress monitors the swarm progress.
func monitorProgress(ctx context.Context, s *swarm.Swarm, target core.State, errChan chan error, _ context.CancelFunc) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	iteration := 0
	startTime := time.Now()
	swarmSize := len(s.Agents())

	for {
		select {
		case err := <-errChan:
			fmt.Printf("\nError in swarm: %v\n", err)
			return
		case <-ticker.C:
			iteration++

			if !processIteration(s, target, iteration, startTime, swarmSize) {
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

// processIteration processes a single monitoring iteration.
func processIteration(s *swarm.Swarm, target core.State, iteration int, startTime time.Time, swarmSize int) bool {
	// Calculate energy statistics
	stats := calculateEnergyStats(s)
	coherence := s.MeasureCoherence()

	// Print iteration summary
	printIterationSummary(iteration, startTime, coherence, stats, swarmSize)

	// Simulate energy replenishment for some agents
	if iteration%3 == 0 {
		replenishEnergy(s, 0.2) // Replenish 20% of agents
		fmt.Println("  [Energy replenished for some agents]")
	}

	// Check termination conditions
	if coherence >= target.Coherence {
		fmt.Printf("\n✓ Target coherence reached!\n")
		return false
	}

	if iteration >= 15 {
		fmt.Println("\nMaximum iterations reached")
		return false
	}

	if stats.exhaustedAgents > swarmSize/2 {
		fmt.Println("\nToo many exhausted agents - system struggling")
	}

	return true
}

// calculateEnergyStats calculates energy statistics for the swarm.
func calculateEnergyStats(s *swarm.Swarm) energyStats {
	stats := energyStats{minEnergy: 100}
	agentCount := 0

	for _, a := range s.Agents() {
		energy := a.Energy()
		stats.totalEnergy += energy

		if energy < stats.minEnergy {
			stats.minEnergy = energy
		}
		if energy > stats.maxEnergy {
			stats.maxEnergy = energy
		}
		if energy < 10 {
			stats.exhaustedAgents++
		}
		agentCount++
	}

	if agentCount > 0 {
		stats.avgEnergy = stats.totalEnergy / float64(agentCount)
	}

	return stats
}

// printIterationSummary prints the summary for an iteration.
func printIterationSummary(iteration int, startTime time.Time, coherence float64, stats energyStats, swarmSize int) {
	fmt.Printf("\n--- Iteration %d (%.1fs) ---\n",
		iteration, time.Since(startTime).Seconds())
	fmt.Printf("  Coherence:        %.3f\n", coherence)
	fmt.Printf("  Avg Energy:       %.1f\n", stats.avgEnergy)
	fmt.Printf("  Energy Range:     %.1f - %.1f\n", stats.minEnergy, stats.maxEnergy)
	fmt.Printf("  Exhausted Agents: %d/%d\n", stats.exhaustedAgents, swarmSize)
}

// printEnergyAnalysis prints the final energy analysis.
func printEnergyAnalysis(s *swarm.Swarm, target core.State) {
	fmt.Println("\n=== Energy Analysis ===")

	// Group agents by remaining energy
	counts := countAgentsByEnergy(s)

	fmt.Printf("High energy (≥70):    %d agents\n", counts["high"])
	fmt.Printf("Medium energy (30-70): %d agents\n", counts["medium"])
	fmt.Printf("Low energy (10-30):    %d agents\n", counts["low"])
	fmt.Printf("Exhausted (<10):       %d agents\n", counts["exhausted"])

	finalCoherence := s.MeasureCoherence()
	fmt.Printf("\nFinal coherence: %.3f\n", finalCoherence)
	fmt.Printf("Target achieved: %v\n", finalCoherence >= target.Coherence)
}

// countAgentsByEnergy counts agents by energy level.
func countAgentsByEnergy(s *swarm.Swarm) map[string]int {
	counts := map[string]int{
		"high":      0,
		"medium":    0,
		"low":       0,
		"exhausted": 0,
	}

	for _, a := range s.Agents() {
		energy := a.Energy()
		switch {
		case energy >= 70:
			counts["high"]++
		case energy >= 30:
			counts["medium"]++
		case energy >= 10:
			counts["low"]++
		default:
			counts["exhausted"]++
		}
	}

	return counts
}

func main() {
	fmt.Println("=== Energy-Constrained Synchronization Example ===")
	fmt.Println()

	// Setup swarm with energy profiles
	s, target := setupEnergySwarm()

	fmt.Printf("\nInitial coherence: %.3f\n", s.MeasureCoherence())

	// Run synchronization with energy monitoring
	runEnergyMonitoring(s, target)

	// Final analysis
	printEnergyAnalysis(s, target)

	// Demonstrate energy-based decision making
	fmt.Println("\n=== Energy-Based Decision Example ===")
	demonstrateEnergyDecisions()
}

// replenishEnergy simulates energy recovery for a fraction of agents.
func replenishEnergy(s *swarm.Swarm, fraction float64) {
	count := 0
	for _, a := range s.Agents() {
		if random.Float64() < fraction {
			currentEnergy := a.Energy()
			// Replenish up to 30 energy units
			a.SetEnergy(currentEnergy + 30)
			if a.Energy() > 100 {
				a.SetEnergy(100)
			}
			count++
		}
	}
}

// demonstrateEnergyDecisions shows how agents make decisions based on energy.
func demonstrateEnergyDecisions() {
	// Define outcome constants
	const (
		outcomeFailed       = "failed"
		outcomeUnsuccessful = "unsuccessful"
		outcomeSuccessful   = "successful"
		outcomeRejected     = "rejected"
	)

	// Create two agents with different energy levels
	richAgent := agent.New("energy-rich")
	richAgent.SetEnergy(90)

	poorAgent := agent.New("energy-poor")
	poorAgent.SetEnergy(15)

	target := core.State{
		Phase:     1.0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.8,
	}

	fmt.Println("\nTwo agents face the same adjustment opportunity:")

	fmt.Printf("Rich agent (90 energy): ")

	// Rich agent can afford expensive adjustments
	var richOutcome string
	action1, accepted1 := richAgent.ProposeAdjustment(target)
	if accepted1 {
		fmt.Printf("Accepts adjustment (cost: %.1f)\n", action1.Cost)
		success, energyCost, err := richAgent.ApplyAction(action1)
		switch {
		case err != nil:
			// Action failed - could be insufficient energy or invalid action
			fmt.Printf("  Action failed: %v\n", err)
			richOutcome = outcomeFailed
		case !success:
			// Action was valid but unsuccessful
			fmt.Printf("  Action unsuccessful (cost: %.1f)\n", energyCost)
			richOutcome = outcomeUnsuccessful
		default:
			fmt.Printf("  Action successful (cost: %.1f)\n", energyCost)
			richOutcome = outcomeSuccessful
		}
	} else {
		fmt.Println("Rejects adjustment")
		richOutcome = outcomeRejected
	}

	fmt.Printf("Poor agent (15 energy): ")

	// Poor agent must be conservative
	var poorOutcome string
	action2, accepted2 := poorAgent.ProposeAdjustment(target)
	if accepted2 {
		fmt.Printf("Accepts adjustment (cost: %.1f)\n", action2.Cost)
		success, energyCost, err := poorAgent.ApplyAction(action2)
		switch {
		case err != nil:
			// Action failed - this demonstrates energy constraints
			fmt.Printf("  Action failed: %v\n", err)
			poorOutcome = outcomeFailed
		case !success:
			// Action was valid but unsuccessful
			fmt.Printf("  Action unsuccessful (cost: %.1f)\n", energyCost)
			poorOutcome = outcomeUnsuccessful
		default:
			fmt.Printf("  Action successful (cost: %.1f)\n", energyCost)
			poorOutcome = outcomeSuccessful
		}
	} else {
		fmt.Println("Rejects adjustment (too expensive)")
		poorOutcome = outcomeRejected
	}

	fmt.Printf("\nFinal energy levels:\n")
	fmt.Printf("  Rich agent: %.1f (outcome: %s)\n", richAgent.Energy(), richOutcome)
	fmt.Printf("  Poor agent: %.1f (outcome: %s)\n", poorAgent.Energy(), poorOutcome)

	// Demonstrate the key point
	fmt.Println("\nKey insight: Energy constraints shape agent behavior")
	if richOutcome == outcomeSuccessful && (poorOutcome == outcomeFailed || poorOutcome == outcomeRejected) {
		fmt.Println("✓ Rich agents can afford actions that poor agents cannot")
	}
}
