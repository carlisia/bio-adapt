// batch_manager.go handles batch processing and API call simulation.
// This component manages how synchronized agent batches are processed.

package simulation

import (
	"context"
	"sync"
	"time"
)

const (
	// CostPerCall is the cost per individual API call ($0.03)
	CostPerCall = 0.03
	// BatchDiscount is the discount applied for batched calls (50%)
	BatchDiscount = 0.5
	// MaxBatchSize is the maximum items per API batch
	MaxBatchSize = 100

	// BatchWindow is the collection window for batches
	BatchWindow = 50 * time.Millisecond
)

// BatchManager handles batch collection and processing
type BatchManager struct {
	// Batch queue
	queue chan AgentBatch

	// Current batch being collected
	current   []Task
	currentMu sync.Mutex

	// Statistics
	totalBatches int
	mu           sync.RWMutex

	// Batch pulse tracking
	lastBatchTime time.Time
	lastBatchSize int
}

// AgentBatch represents tasks from a single agent
type AgentBatch struct {
	AgentID   string
	Tasks     []Task
	Timestamp time.Time
}

// NewBatchManager creates a new batch manager
func NewBatchManager() *BatchManager {
	return &BatchManager{
		queue:   make(chan AgentBatch, 100),
		current: make([]Task, 0),
	}
}

// SubmitBatch submits a batch from an agent
func (bm *BatchManager) SubmitBatch(agentID string, tasks []Task) {
	batch := AgentBatch{
		AgentID:   agentID,
		Tasks:     tasks,
		Timestamp: time.Now(),
	}

	select {
	case bm.queue <- batch:
		// Batch queued
	default:
		// Queue full, drop batch (shouldn't happen in normal operation)
	}
}

// ProcessBatches processes batches from the queue
func (bm *BatchManager) ProcessBatches(ctx context.Context, metrics *MetricsCollector) {
	for {
		select {
		case <-ctx.Done():
			return
		case batch := <-bm.queue:
			// Start collecting batches that arrive close together
			collected := bm.collectSynchronizedBatches(batch)

			// Process the collected batch
			bm.processCollectedBatch(collected, metrics)
		}
	}
}

// collectSynchronizedBatches collects batches arriving within the window
func (bm *BatchManager) collectSynchronizedBatches(initial AgentBatch) []Task {
	collected := make([]Task, 0)
	collected = append(collected, initial.Tasks...)

	// Set a timer for the batch window
	timer := time.NewTimer(BatchWindow)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			// Window closed
			return collected
		case additional := <-bm.queue:
			// Add to collected batch
			collected = append(collected, additional.Tasks...)
		default:
			// No more batches immediately available
			if len(collected) >= MaxBatchSize {
				// Batch is full
				return collected
			}
		}
	}
}

// processCollectedBatch processes a collected batch
func (bm *BatchManager) processCollectedBatch(tasks []Task, metrics *MetricsCollector) {
	if len(tasks) == 0 {
		return
	}

	// Split into API-sized batches if needed
	batches := bm.splitIntoAPIBatches(tasks)

	for _, batch := range batches {
		// Calculate costs
		individualCost := float64(len(batch)) * CostPerCall
		batchedCost := CostPerCall * (1 - BatchDiscount) // Single call with discount
		savings := individualCost - batchedCost

		// Update metrics
		metrics.RecordBatch(len(batch), individualCost, batchedCost, savings)

		// Update local stats and track for pulse animation
		bm.mu.Lock()
		bm.totalBatches++
		bm.lastBatchTime = time.Now()
		bm.lastBatchSize = len(batch)
		bm.mu.Unlock()

		// Simulate API call processing time
		time.Sleep(10 * time.Millisecond)
	}
}

// splitIntoAPIBatches splits tasks into API-compliant batch sizes
func (*BatchManager) splitIntoAPIBatches(tasks []Task) [][]Task {
	var batches [][]Task

	for i := 0; i < len(tasks); i += MaxBatchSize {
		end := i + MaxBatchSize
		if end > len(tasks) {
			end = len(tasks)
		}
		batches = append(batches, tasks[i:end])
	}

	return batches
}

// PendingCount returns number of pending tasks
func (bm *BatchManager) PendingCount() int {
	bm.currentMu.Lock()
	defer bm.currentMu.Unlock()
	return len(bm.current)
}

// CurrentBatchSize returns size of current batch
func (bm *BatchManager) CurrentBatchSize() int {
	// Estimate based on queue depth
	return len(bm.queue) * 10 // Rough estimate
}

// Reset resets the batch manager
func (bm *BatchManager) Reset() {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.totalBatches = 0
	bm.lastBatchTime = time.Time{}
	bm.lastBatchSize = 0

	// Clear queue
	for len(bm.queue) > 0 {
		<-bm.queue
	}

	bm.currentMu.Lock()
	bm.current = bm.current[:0]
	bm.currentMu.Unlock()
}

// LastBatchInfo returns information about the last batch sent
func (bm *BatchManager) LastBatchInfo() (time.Time, int) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	return bm.lastBatchTime, bm.lastBatchSize
}
