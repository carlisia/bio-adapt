// metrics.go handles metrics collection and statistics.
// This component tracks performance metrics throughout the simulation.

package simulation

import (
	"sync"
	"time"
)

// MetricsCollector collects and aggregates simulation metrics
type MetricsCollector struct {
	mu sync.RWMutex

	// Counts
	totalAPICalls int
	totalBatches  int
	totalTasks    int

	// Costs
	costWithoutSync float64
	costWithSync    float64

	// Coherence tracking
	peakCoherence float64
	convergedAt   time.Time
	startTime     time.Time

	// Batch size tracking
	batchSizes   []int
	largestBatch int
}

// MetricsSnapshot represents metrics at a point in time
type MetricsSnapshot struct {
	TotalAPICalls    int
	TotalBatches     int
	AverageBatchSize float64
	CostWithoutSync  float64
	CostWithSync     float64
	TotalSavings     float64
	SavingsPercent   float64
	PeakCoherence    float64
	TimeToConverge   time.Duration
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		startTime:  time.Now(),
		batchSizes: make([]int, 0),
	}
}

// RecordBatch records a processed batch
func (mc *MetricsCollector) RecordBatch(size int, individualCost, batchedCost, _ float64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Update counts
	mc.totalAPICalls += size // Would have been this many individual calls
	mc.totalBatches++        // But we made one batch call
	mc.totalTasks += size

	// Update costs
	mc.costWithoutSync += individualCost
	mc.costWithSync += batchedCost

	// Track batch sizes
	mc.batchSizes = append(mc.batchSizes, size)
	if size > mc.largestBatch {
		mc.largestBatch = size
	}
}

// UpdateCoherence updates coherence metrics
func (mc *MetricsCollector) UpdateCoherence(coherence float64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if coherence > mc.peakCoherence {
		mc.peakCoherence = coherence
	}

	// Mark convergence time (first time reaching 70% coherence)
	if coherence >= 0.7 && mc.convergedAt.IsZero() {
		mc.convergedAt = time.Now()
	}
}

// Current returns current metrics
func (mc *MetricsCollector) Current() MetricsSnapshot {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	snapshot := MetricsSnapshot{
		TotalAPICalls:   mc.totalAPICalls,
		TotalBatches:    mc.totalBatches,
		CostWithoutSync: mc.costWithoutSync,
		CostWithSync:    mc.costWithSync,
		TotalSavings:    mc.costWithoutSync - mc.costWithSync,
		PeakCoherence:   mc.peakCoherence,
	}

	// Calculate average batch size
	if mc.totalBatches > 0 {
		snapshot.AverageBatchSize = float64(mc.totalTasks) / float64(mc.totalBatches)
	}

	// Calculate savings percentage
	if mc.costWithoutSync > 0 {
		snapshot.SavingsPercent = (snapshot.TotalSavings / mc.costWithoutSync) * 100
	}

	// Calculate time to converge
	if !mc.convergedAt.IsZero() {
		snapshot.TimeToConverge = mc.convergedAt.Sub(mc.startTime)
	}

	return snapshot
}

// Final returns final metrics
func (mc *MetricsCollector) Final() MetricsSnapshot {
	return mc.Current()
}

// Reset resets all metrics
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.totalAPICalls = 0
	mc.totalBatches = 0
	mc.totalTasks = 0
	mc.costWithoutSync = 0
	mc.costWithSync = 0
	mc.peakCoherence = 0
	mc.convergedAt = time.Time{}
	mc.startTime = time.Now()
	mc.batchSizes = mc.batchSizes[:0]
	mc.largestBatch = 0
}

// BatchSizeDistribution returns distribution of batch sizes
func (mc *MetricsCollector) BatchSizeDistribution() map[string]int {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	distribution := map[string]int{
		"1-10":   0,
		"11-25":  0,
		"26-50":  0,
		"51-75":  0,
		"76-100": 0,
	}

	for _, size := range mc.batchSizes {
		switch {
		case size <= 10:
			distribution["1-10"]++
		case size <= 25:
			distribution["11-25"]++
		case size <= 50:
			distribution["26-50"]++
		case size <= 75:
			distribution["51-75"]++
		default:
			distribution["76-100"]++
		}
	}

	return distribution
}
