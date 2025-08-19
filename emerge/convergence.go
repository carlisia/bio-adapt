package emerge

import (
	"math"
	"sync"
	"time"
)

// ConvergenceMonitor tracks system convergence toward attractor basins.
// This extracts and enhances the monitoring logic from monitor.go.
type ConvergenceMonitor struct {
	history       []float64   // Coherence history
	timestamps    []time.Time // When measurements were taken
	targetState   State       // Target we're converging to
	convergenceAt float64     // Coherence level considered "converged"

	// Statistics
	startTime     time.Time
	convergedTime *time.Time
	maxCoherence  float64
	minCoherence  float64

	mu sync.RWMutex
}

// NewConvergenceMonitor creates a monitor for tracking convergence.
func NewConvergenceMonitor(target State, convergenceThreshold float64) *ConvergenceMonitor {
	return &ConvergenceMonitor{
		history:       make([]float64, 0, 1000),
		timestamps:    make([]time.Time, 0, 1000),
		targetState:   target,
		convergenceAt: convergenceThreshold,
		startTime:     time.Now(),
		maxCoherence:  0,
		minCoherence:  1,
	}
}

// Record adds a new coherence measurement.
func (m *ConvergenceMonitor) Record(coherence float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	m.history = append(m.history, coherence)
	m.timestamps = append(m.timestamps, now)

	// Update statistics
	if coherence > m.maxCoherence {
		m.maxCoherence = coherence
	}
	if coherence < m.minCoherence {
		m.minCoherence = coherence
	}

	// Check for convergence
	if coherence >= m.convergenceAt && m.convergedTime == nil {
		m.convergedTime = &now
	} else if coherence < m.convergenceAt && m.convergedTime != nil {
		// Reset convergence if we drop below threshold
		m.convergedTime = nil
	}
}

// IsConverged returns whether the system has converged.
func (m *ConvergenceMonitor) IsConverged() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.convergedTime != nil
}

// ConvergenceTime returns how long it took to converge.
// Returns 0 if not yet converged.
func (m *ConvergenceMonitor) ConvergenceTime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.convergedTime == nil {
		return 0
	}
	return m.convergedTime.Sub(m.startTime)
}

// CurrentCoherence returns the most recent coherence value.
func (m *ConvergenceMonitor) CurrentCoherence() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.history) == 0 {
		return 0
	}
	return m.history[len(m.history)-1]
}

// ConvergenceRate calculates the rate of convergence.
// Positive values indicate convergence, negative indicate divergence.
func (m *ConvergenceMonitor) ConvergenceRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.history) < 2 {
		return 0
	}

	// Calculate rate over last 10 samples or available history
	samples := 10
	if len(m.history) < samples {
		samples = len(m.history)
	}

	// Linear regression to find slope
	startIdx := len(m.history) - samples
	endIdx := len(m.history)

	var sumX, sumY, sumXY, sumX2 float64
	for i := startIdx; i < endIdx; i++ {
		x := float64(i - startIdx)
		y := m.history[i]
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	n := float64(samples)
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)

	return slope
}

// Stability returns how stable the convergence is.
// Lower values indicate more stable convergence.
func (m *ConvergenceMonitor) Stability() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.history) < 2 {
		return 1.0
	}

	// Calculate variance over recent history
	samples := 20
	if len(m.history) < samples {
		samples = len(m.history)
	}

	startIdx := len(m.history) - samples

	// Calculate mean
	var sum float64
	for i := startIdx; i < len(m.history); i++ {
		sum += m.history[i]
	}
	mean := sum / float64(samples)

	// Calculate variance
	var variance float64
	for i := startIdx; i < len(m.history); i++ {
		diff := m.history[i] - mean
		variance += diff * diff
	}
	variance /= float64(samples)

	// Convert to stability score (1 = stable, 0 = unstable)
	// Use exponential decay for smooth transition
	stdDev := math.Sqrt(variance)
	return math.Exp(-stdDev * 10) // High stability for low variance
}

// PredictConvergenceTime estimates time to convergence based on current rate.
// Returns 0 if already converged or diverging.
func (m *ConvergenceMonitor) PredictConvergenceTime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.convergedTime != nil {
		return 0 // Already converged
	}

	rate := m.ConvergenceRate()
	if rate <= 0 {
		return 0 // Not converging
	}

	current := m.CurrentCoherence()
	remaining := m.convergenceAt - current

	if remaining <= 0 {
		return 0 // Already at target
	}

	// Estimate based on current rate
	// This is simplified - real systems would use more sophisticated prediction
	timeToConverge := remaining / rate

	// Assume each sample is about 100ms apart (configurable)
	samplePeriod := 100 * time.Millisecond

	return time.Duration(timeToConverge * float64(samplePeriod))
}

// GetHistory returns the full coherence history.
func (m *ConvergenceMonitor) GetHistory() []float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]float64, len(m.history))
	copy(result, m.history)
	return result
}

// GetStatistics returns comprehensive convergence statistics.
func (m *ConvergenceMonitor) GetStatistics() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]any)

	stats["samples"] = len(m.history)
	stats["current_coherence"] = m.CurrentCoherence()
	stats["max"] = m.maxCoherence
	stats["min"] = m.minCoherence
	stats["converged"] = m.convergedTime != nil
	stats["target_coherence"] = m.convergenceAt

	if m.convergedTime != nil {
		stats["convergence_time"] = m.convergedTime.Sub(m.startTime).Seconds()
	}

	if len(m.history) > 0 {
		// Calculate mean
		var sum float64
		for _, v := range m.history {
			sum += v
		}
		mean := sum / float64(len(m.history))
		stats["mean"] = mean

		// Calculate std dev
		var variance float64
		for _, v := range m.history {
			diff := v - mean
			variance += diff * diff
		}
		variance /= float64(len(m.history))
		stats["std_dev"] = math.Sqrt(variance)
	}

	stats["convergence_rate"] = m.ConvergenceRate()
	stats["stability"] = m.Stability()

	prediction := m.PredictConvergenceTime()
	if prediction > 0 {
		stats["predicted_convergence_time"] = prediction.Seconds()
	}

	return stats
}

// Reset clears the monitor for a new convergence run.
func (m *ConvergenceMonitor) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.history = m.history[:0]
	m.timestamps = m.timestamps[:0]
	m.startTime = time.Now()
	m.convergedTime = nil
	m.maxCoherence = 0
	m.minCoherence = 1
}
