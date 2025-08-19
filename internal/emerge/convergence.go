package emerge

import (
	"math"
	"sync"
	"time"
)

// ConvergenceMonitor tracks convergence behavior internally.
// This is not exposed to users - they interact through Swarm's public methods.
type ConvergenceMonitor struct {
	targetCoherence  float64
	history          []float64
	timestamps       []time.Time
	mu               sync.RWMutex
	convergedAt      *time.Time
	stableIterations int
}

// NewConvergenceMonitor creates a new convergence monitor.
func NewConvergenceMonitor(targetCoherence float64) *ConvergenceMonitor {
	return &ConvergenceMonitor{
		targetCoherence: targetCoherence,
		history:         make([]float64, 0, 100),
		timestamps:      make([]time.Time, 0, 100),
	}
}

// Record adds a coherence measurement.
func (cm *ConvergenceMonitor) Record(coherence float64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.history = append(cm.history, coherence)
	cm.timestamps = append(cm.timestamps, time.Now())

	// Check for convergence
	if coherence >= cm.targetCoherence {
		cm.stableIterations++
		if cm.stableIterations >= 5 && cm.convergedAt == nil {
			now := time.Now()
			cm.convergedAt = &now
		}
	} else {
		cm.stableIterations = 0
		cm.convergedAt = nil
	}
}

// IsConverged returns whether the system has converged.
func (cm *ConvergenceMonitor) IsConverged() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.convergedAt != nil
}

// Rate calculates the convergence rate.
func (cm *ConvergenceMonitor) Rate() float64 {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if len(cm.history) < 2 {
		return 0
	}

	// Simple linear regression for rate
	n := float64(len(cm.history))
	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0

	for i, coherence := range cm.history {
		x := float64(i)
		sumX += x
		sumY += coherence
		sumXY += x * coherence
		sumX2 += x * x
	}

	denominator := n*sumX2 - sumX*sumX
	if math.Abs(denominator) < 1e-10 {
		return 0
	}

	return (n*sumXY - sumX*sumY) / denominator
}

// Statistics returns convergence statistics.
func (cm *ConvergenceMonitor) Statistics() map[string]float64 {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if len(cm.history) == 0 {
		return map[string]float64{"samples": 0}
	}

	sum := 0.0
	min := cm.history[0]
	max := cm.history[0]

	for _, v := range cm.history {
		sum += v
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	mean := sum / float64(len(cm.history))

	// Calculate variance
	variance := 0.0
	for _, v := range cm.history {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(cm.history))

	return map[string]float64{
		"samples":  float64(len(cm.history)),
		"mean":     mean,
		"min":      min,
		"max":      max,
		"variance": variance,
		"stddev":   math.Sqrt(variance),
		"rate":     cm.Rate(),
	}
}

// Reset clears the monitor state.
func (cm *ConvergenceMonitor) Reset() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.history = cm.history[:0]
	cm.timestamps = cm.timestamps[:0]
	cm.convergedAt = nil
	cm.stableIterations = 0
}

