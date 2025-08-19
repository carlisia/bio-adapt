package emerge

import (
	"sync"

	"github.com/gammazero/deque"
)

// Monitor tracks convergence without influencing it.
type Monitor struct {
	history *deque.Deque[float64]
	mu      sync.RWMutex
}

// NewMonitor creates a new monitor for tracking coherence history.
func NewMonitor() *Monitor {
	return &Monitor{
		history: deque.New[float64](100),
	}
}

// RecordSample adds a coherence sample to the history.
func (m *Monitor) RecordSample(coherence float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.history.Len() >= 100 {
		m.history.PopFront()
	}
	m.history.PushBack(coherence)
}

// GetHistory returns the coherence history as a slice.
func (m *Monitor) GetHistory() []float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]float64, m.history.Len())
	for i := range m.history.Len() {
		result[i] = m.history.At(i)
	}
	return result
}

// GetLatest returns the most recent coherence value.
func (m *Monitor) GetLatest() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.history.Len() == 0 {
		return 0
	}
	return m.history.Back()
}

// GetAverage returns the average coherence over the history.
func (m *Monitor) GetAverage() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.history.Len() == 0 {
		return 0
	}

	sum := 0.0
	for i := range m.history.Len() {
		sum += m.history.At(i)
	}
	return sum / float64(m.history.Len())
}
