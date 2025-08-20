// Monitoring and Metrics Example
// This example demonstrates comprehensive monitoring of swarm behavior,
// including real-time metrics, performance analysis, and visualization data.

package main

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/monitoring"
	"github.com/carlisia/bio-adapt/emerge/swarm"
)

// MetricsCollector collects detailed metrics from the swarm
type MetricsCollector struct {
	mu               sync.RWMutex
	coherenceHistory []float64
	energyHistory    []float64
	phaseVariance    []float64
	decisionCounts   map[string]int
	convergenceTime  time.Duration
	startTime        time.Time
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		coherenceHistory: make([]float64, 0, 1000),
		energyHistory:    make([]float64, 0, 1000),
		phaseVariance:    make([]float64, 0, 1000),
		decisionCounts:   make(map[string]int),
		startTime:        time.Now(),
	}
}

func (m *MetricsCollector) RecordCoherence(value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.coherenceHistory = append(m.coherenceHistory, value)
}

func (m *MetricsCollector) RecordEnergy(value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.energyHistory = append(m.energyHistory, value)
}

func (m *MetricsCollector) RecordPhaseVariance(value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.phaseVariance = append(m.phaseVariance, value)
}

func (m *MetricsCollector) IncrementDecision(decision string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.decisionCounts[decision]++
}

func (m *MetricsCollector) SetConvergenceTime(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.convergenceTime = duration
}

func (m *MetricsCollector) GetStatistics() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]any)

	// Coherence statistics
	if len(m.coherenceHistory) > 0 {
		stats["coherence_mean"] = calculateMean(m.coherenceHistory)
		stats["coherence_std"] = calculateStdDev(m.coherenceHistory)
		stats["coherence_min"] = calculateMin(m.coherenceHistory)
		stats["coherence_max"] = calculateMax(m.coherenceHistory)
		stats["coherence_final"] = m.coherenceHistory[len(m.coherenceHistory)-1]
	}

	// Energy statistics
	if len(m.energyHistory) > 0 {
		stats["energy_mean"] = calculateMean(m.energyHistory)
		stats["energy_std"] = calculateStdDev(m.energyHistory)
		stats["energy_final"] = m.energyHistory[len(m.energyHistory)-1]
	}

	// Phase variance
	if len(m.phaseVariance) > 0 {
		stats["phase_variance_mean"] = calculateMean(m.phaseVariance)
		stats["phase_variance_final"] = m.phaseVariance[len(m.phaseVariance)-1]
	}

	// Decision statistics
	stats["decision_counts"] = m.decisionCounts
	stats["convergence_time"] = m.convergenceTime.Seconds()
	stats["total_runtime"] = time.Since(m.startTime).Seconds()

	return stats
}

func main() {
	fmt.Println("=== Comprehensive Monitoring and Metrics Example ===")
	fmt.Println()

	// Configuration
	swarmSize := 100
	target := core.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	// Create swarm and monitoring infrastructure
	swarm, err := swarm.New(swarmSize, target)
	if err != nil {
		fmt.Printf("Error creating swarm: %v\n", err)
		return
	}
	monitor := monitoring.New()
	metrics := NewMetricsCollector()

	fmt.Printf("Monitoring swarm of %d agents\n", swarmSize)
	fmt.Printf("Target coherence: %.2f\n\n", target.Coherence)

	// Start swarm
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		if err := swarm.Run(ctx); err != nil {
			errChan <- err
		}
	}()

	// Real-time monitoring
	fmt.Println("=== Real-Time Metrics ===")
	fmt.Println("Time    Coherence  Energy  Variance  Converging")
	fmt.Println("------------------------------------------------")

	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	converged := false
	convergenceStart := time.Now()
	iteration := 0

	for !converged && iteration < 60 {
		select {
		case err := <-errChan:
			fmt.Printf("\nError in swarm: %v\n", err)
			goto analysis
		case <-ticker.C:
			iteration++

			// Collect metrics
			coherence := swarm.MeasureCoherence()
			monitor.RecordSample(coherence)
			metrics.RecordCoherence(coherence)

			// Calculate average energy
			totalEnergy := 0.0
			agentCount := 0
			var phases []float64

			for _, agent := range swarm.Agents() {
				totalEnergy += agent.Energy()
				phases = append(phases, agent.Phase())
				agentCount++
			}

			avgEnergy := totalEnergy / float64(agentCount)
			metrics.RecordEnergy(avgEnergy)

			// Calculate phase variance
			variance := calculateVariance(phases)
			metrics.RecordPhaseVariance(variance)

			// Determine if converging
			isConverging := false
			history := monitor.History()
			if len(history) >= 5 {
				// Check if coherence is improving
				recent := history[len(history)-5:]
				isConverging = isIncreasing(recent)
			}

			// Display metrics
			elapsed := time.Since(convergenceStart).Seconds()
			convergingStr := "No"
			if isConverging {
				convergingStr = "Yes"
			}

			fmt.Printf("%.1fs    %.3f      %.1f    %.3f      %s\n",
				elapsed, coherence, avgEnergy, variance, convergingStr)

			// Check convergence
			if coherence >= target.Coherence && !converged {
				converged = true
				metrics.SetConvergenceTime(time.Since(convergenceStart))
				fmt.Printf("\n✓ Converged to target coherence in %.2fs\n",
					time.Since(convergenceStart).Seconds())
			}

			// Track decision types (sample a few agents)
			sampleDecisions(swarm, metrics)

		case <-ctx.Done():
			goto analysis
		}
	}

analysis:
	// Detailed analysis
	fmt.Println("\n=== Performance Analysis ===")
	fmt.Println("---------------------------")

	stats := metrics.GetStatistics()

	// Coherence analysis
	fmt.Println("\nCoherence Metrics:")
	fmt.Printf("  Mean:     %.3f\n", stats["coherence_mean"])
	fmt.Printf("  Std Dev:  %.3f\n", stats["coherence_std"])
	fmt.Printf("  Min:      %.3f\n", stats["coherence_min"])
	fmt.Printf("  Max:      %.3f\n", stats["coherence_max"])
	fmt.Printf("  Final:    %.3f\n", stats["coherence_final"])

	// Energy analysis
	fmt.Println("\nEnergy Metrics:")
	fmt.Printf("  Mean:     %.1f\n", stats["energy_mean"])
	fmt.Printf("  Std Dev:  %.1f\n", stats["energy_std"])
	fmt.Printf("  Final:    %.1f\n", stats["energy_final"])

	// Phase analysis
	fmt.Println("\nPhase Metrics:")
	fmt.Printf("  Variance Mean:  %.3f\n", stats["phase_variance_mean"])
	fmt.Printf("  Variance Final: %.3f\n", stats["phase_variance_final"])

	// Timing analysis
	fmt.Println("\nTiming Metrics:")
	if converged {
		fmt.Printf("  Convergence Time: %.2fs\n", stats["convergence_time"])
	} else {
		fmt.Println("  Convergence Time: Not converged")
	}
	fmt.Printf("  Total Runtime:    %.2fs\n", stats["total_runtime"])

	// Generate visualization data
	fmt.Println("\n=== Visualization Data Export ===")
	exportVisualizationData(monitor, metrics)

	// Agent-level analysis
	fmt.Println("\n=== Agent-Level Analysis ===")
	analyzeAgentBehavior(swarm)

	// Network analysis
	fmt.Println("\n=== Network Topology Analysis ===")
	analyzeNetworkTopology(swarm)
}

// Helper functions for statistics
func calculateMean(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

func calculateStdDev(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	mean := calculateMean(data)
	sumSq := 0.0
	for _, v := range data {
		diff := v - mean
		sumSq += diff * diff
	}
	return math.Sqrt(sumSq / float64(len(data)))
}

func calculateVariance(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	mean := calculateMean(data)
	sumSq := 0.0
	for _, v := range data {
		diff := v - mean
		sumSq += diff * diff
	}
	return sumSq / float64(len(data))
}

func calculateMin(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	min := data[0]
	for _, v := range data {
		if v < min {
			min = v
		}
	}
	return min
}

func calculateMax(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	max := data[0]
	for _, v := range data {
		if v > max {
			max = v
		}
	}
	return max
}

func isIncreasing(data []float64) bool {
	if len(data) < 2 {
		return false
	}
	increases := 0
	for i := 1; i < len(data); i++ {
		if data[i] > data[i-1] {
			increases++
		}
	}
	return increases > len(data)/2
}

func sampleDecisions(swarm *swarm.Swarm, metrics *MetricsCollector) {
	// Sample a few agents to track decision patterns
	sampled := 0
	for _, agent := range swarm.Agents() {
		if sampled >= 5 {
			break
		}

		// Simulate decision tracking
		if agent.Energy() < 20 {
			metrics.IncrementDecision("energy_conserve")
		} else if agent.Influence() > 0.7 {
			metrics.IncrementDecision("high_influence")
		} else {
			metrics.IncrementDecision("normal")
		}

		sampled++
	}
}

func exportVisualizationData(monitor *monitoring.Monitor, _ *MetricsCollector) {
	// Export coherence time series
	history := monitor.History()
	fmt.Println("\nCoherence Time Series (for plotting):")
	fmt.Println("Index,Coherence")
	for i, c := range history {
		if i%5 == 0 { // Sample every 5th point for brevity
			fmt.Printf("%d,%.3f\n", i, c)
		}
	}

	fmt.Printf("\n[Full dataset contains %d points]\n", len(history))
}

func analyzeAgentBehavior(swarm *swarm.Swarm) {
	type agentStats struct {
		id           string
		phase        float64
		energy       float64
		influence    float64
		stubbornness float64
		neighbors    int
	}

	var agents []agentStats

	for _, agent := range swarm.Agents() {
		neighborCount := 0
		agent.Neighbors().Range(func(k, v any) bool {
			neighborCount++
			return true
		})

		agents = append(agents, agentStats{
			id:           agent.ID[:8],
			phase:        agent.Phase(),
			energy:       agent.Energy(),
			influence:    agent.Influence(),
			stubbornness: agent.Stubbornness(),
			neighbors:    neighborCount,
		})
	}

	// Sort by energy to find outliers
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].energy > agents[j].energy
	})

	fmt.Println("\nTop 5 Highest Energy Agents:")
	maxShow := min(5, len(agents))
	for i := range maxShow {
		a := agents[i]
		fmt.Printf("  %s: Energy=%.1f, Phase=%.2f, Neighbors=%d\n",
			a.id, a.energy, a.phase, a.neighbors)
	}

	fmt.Println("\nTop 5 Lowest Energy Agents:")
	for i := len(agents) - 5; i < len(agents) && i >= 0; i++ {
		a := agents[i]
		fmt.Printf("  %s: Energy=%.1f, Phase=%.2f, Neighbors=%d\n",
			a.id, a.energy, a.phase, a.neighbors)
	}

	// Find most stubborn agents
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].stubbornness > agents[j].stubbornness
	})

	fmt.Println("\nMost Stubborn Agents:")
	maxStubborn := min(3, len(agents))
	for i := range maxStubborn {
		a := agents[i]
		fmt.Printf("  %s: Stubbornness=%.2f, Phase=%.2f\n",
			a.id, a.stubbornness, a.phase)
	}
}

func analyzeNetworkTopology(swarm *swarm.Swarm) {
	connectionCounts := make(map[int]int)
	totalConnections := 0
	maxConnections := 0
	minConnections := 1000

	for _, agent := range swarm.Agents() {
		connections := 0
		agent.Neighbors().Range(func(k, v any) bool {
			connections++
			return true
		})

		connectionCounts[connections]++
		totalConnections += connections

		if connections > maxConnections {
			maxConnections = connections
		}
		if connections < minConnections {
			minConnections = connections
		}
	}

	agentCount := 0
	for _, count := range connectionCounts {
		agentCount += count
	}

	avgConnections := float64(totalConnections) / float64(agentCount)

	fmt.Printf("\nNetwork Statistics:\n")
	fmt.Printf("  Average connections: %.1f\n", avgConnections)
	fmt.Printf("  Min connections:     %d\n", minConnections)
	fmt.Printf("  Max connections:     %d\n", maxConnections)

	fmt.Println("\nConnection Distribution:")
	for connections, count := range connectionCounts {
		if count > 0 {
			fmt.Printf("  %d connections: %d agents\n", connections, count)
		}
	}
}
