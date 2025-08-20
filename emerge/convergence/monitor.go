// Package convergence provides monitoring and analysis of swarm convergence behavior.
package convergence

import (
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/carlisia/bio-adapt/emerge/core"
)

// Monitor tracks convergence toward target pattern.
type Monitor struct {
	history    []Sample
	historyMu  sync.RWMutex
	windowSize int
	target     *core.RhythmicPattern
	converging atomic.Bool
	stuckCount atomic.Int32
}

// Sample represents a convergence measurement.
type Sample struct {
	Timestamp    time.Time
	Distance     float64 // Distance to target
	Coherence    float64
	Velocity     float64 // Rate of change
	Acceleration float64 // Change in rate
}

// NewMonitor creates a convergence monitor.
func NewMonitor(windowSize int) *Monitor {
	if windowSize <= 0 {
		windowSize = 10
	}
	return &Monitor{
		history:    make([]Sample, 0, windowSize*2),
		windowSize: windowSize,
	}
}

// SetTarget sets the target pattern to converge toward.
func (m *Monitor) SetTarget(target *core.RhythmicPattern) {
	m.target = target
}

// RecordSample records a new convergence sample.
func (m *Monitor) RecordSample(current *core.RhythmicPattern, coherence float64) {
	m.historyMu.Lock()
	defer m.historyMu.Unlock()

	now := time.Now()
	distance := core.PatternDistance(current, m.target)

	// Calculate velocity and acceleration if we have history
	velocity := 0.0
	acceleration := 0.0

	if len(m.history) > 0 {
		lastSample := m.history[len(m.history)-1]
		dt := now.Sub(lastSample.Timestamp).Seconds()
		if dt > 0 {
			velocity = (distance - lastSample.Distance) / dt

			if len(m.history) > 1 {
				acceleration = (velocity - lastSample.Velocity) / dt
			}
		}
	}

	sample := Sample{
		Timestamp:    now,
		Distance:     distance,
		Coherence:    coherence,
		Velocity:     velocity,
		Acceleration: acceleration,
	}

	m.history = append(m.history, sample)

	// Keep history bounded
	if len(m.history) > m.windowSize*2 {
		m.history = m.history[len(m.history)-m.windowSize:]
	}

	// Update convergence status
	m.updateConvergenceStatus()
}

// IsConverging returns true if system is converging toward target.
func (m *Monitor) IsConverging() bool {
	return m.converging.Load()
}

// IsStuck returns true if system is stuck (not making progress).
func (m *Monitor) IsStuck() bool {
	return m.stuckCount.Load() > 5
}

// GetProgress returns convergence progress (0-1).
func (m *Monitor) GetProgress() float64 {
	m.historyMu.RLock()
	defer m.historyMu.RUnlock()

	if len(m.history) == 0 {
		return 0
	}

	// Get initial and current distance
	initialDistance := 1.0 // Assume max distance initially
	if len(m.history) > 0 {
		initialDistance = m.history[0].Distance
	}

	currentDistance := m.history[len(m.history)-1].Distance

	// Calculate progress
	if initialDistance <= 0 {
		return 1.0
	}

	progress := 1.0 - (currentDistance / initialDistance)
	return math.Max(0, math.Min(1, progress))
}

// GetConvergenceRate returns the average convergence rate.
func (m *Monitor) GetConvergenceRate() float64 {
	m.historyMu.RLock()
	defer m.historyMu.RUnlock()

	if len(m.history) < 2 {
		return 0
	}

	// Calculate average velocity over recent samples
	recentSamples := m.getRecentSamples()
	if len(recentSamples) == 0 {
		return 0
	}

	totalVelocity := 0.0
	count := 0
	for _, sample := range recentSamples {
		if sample.Velocity != 0 {
			totalVelocity += sample.Velocity
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return totalVelocity / float64(count)
}

// ShouldSwitchStrategy returns true if current strategy isn't working.
func (m *Monitor) ShouldSwitchStrategy() bool {
	m.historyMu.RLock()
	defer m.historyMu.RUnlock()

	if len(m.history) < 3 {
		return false
	}

	// Check if we're stuck or diverging
	recentSamples := m.getRecentSamples()
	if len(recentSamples) < 3 {
		return false
	}

	// Count samples with no progress or negative progress
	stuckCount := 0
	for _, sample := range recentSamples {
		if sample.Velocity >= -0.001 { // Not improving or getting worse
			stuckCount++
		}
	}

	// Switch if more than 70% of recent samples show no progress
	return float64(stuckCount) > float64(len(recentSamples))*0.7
}

// updateConvergenceStatus updates internal convergence state.
func (m *Monitor) updateConvergenceStatus() {
	if len(m.history) < 3 {
		m.converging.Store(false)
		return
	}

	recentSamples := m.getRecentSamples()
	if len(recentSamples) < 3 {
		m.converging.Store(false)
		return
	}

	// Check trend in distances
	distances := make([]float64, len(recentSamples))
	for i, sample := range recentSamples {
		distances[i] = sample.Distance
	}

	// Simple linear regression to find trend
	trend := calculateTrend(distances)

	// Negative trend means converging
	converging := trend < -0.001

	// Check if stuck
	if math.Abs(trend) < 0.001 {
		m.stuckCount.Add(1)
	} else {
		m.stuckCount.Store(0)
	}

	m.converging.Store(converging)
}

// getRecentSamples returns the most recent samples within window.
func (m *Monitor) getRecentSamples() []Sample {
	start := len(m.history) - m.windowSize
	if start < 0 {
		start = 0
	}
	return m.history[start:]
}

// calculateTrend calculates linear trend in data.
func calculateTrend(data []float64) float64 {
	if len(data) < 2 {
		return 0
	}

	// Simple linear regression
	n := float64(len(data))
	sumX := 0.0
	sumY := 0.0
	sumXY := 0.0
	sumX2 := 0.0

	for i, y := range data {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Calculate slope
	denominator := n*sumX2 - sumX*sumX
	if math.Abs(denominator) < 1e-10 {
		return 0
	}

	slope := (n*sumXY - sumX*sumY) / denominator
	return slope
}

// Reset clears the convergence history.
func (m *Monitor) Reset() {
	m.historyMu.Lock()
	defer m.historyMu.Unlock()
	m.history = m.history[:0]
	m.stuckCount.Store(0)
	m.converging.Store(false)
}
