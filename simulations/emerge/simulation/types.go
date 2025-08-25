// Package simulation contains the core simulation logic for the minimize_api_calls demo.
package simulation

import (
	"time"
)

// Config holds internal simulation configuration
type Config struct {
	NumAgents       int
	TargetCoherence float64
}

// Snapshot represents a point-in-time view of the simulation
type Snapshot struct {
	Timestamp        time.Time
	ElapsedTime      time.Duration
	Agents           []AgentSnapshot
	Coherence        float64
	TargetCoherence  float64
	PendingTasks     int
	CurrentBatchSize int
	BatchesProcessed int
	CostWithoutSync  float64
	CostWithSync     float64
	Savings          float64
	SavingsPercent   float64
	Paused           bool
	Disrupted        bool
	Reset            bool
	BatchJustSent    bool      // True if a batch was sent in the last update
	LastBatchTime    time.Time // When the last batch was sent
	LastBatchSize    int       // Size of the last batch sent
}

// AgentSnapshot represents an agent's state
type AgentSnapshot struct {
	ID            string
	Type          string // The type of agent (e.g., DataETL, Leader, etc.)
	Icon          string
	Phase         float64
	PendingTasks  int
	BatchesSent   int
	InBurstMode   bool   // True if agent is in burst mode (for burst pattern)
	ActivityLevel string // "burst", "quiet", "steady", "active"
}

// Statistics holds final simulation results
type Statistics struct {
	TotalAPICalls    int
	TotalBatches     int
	AverageBatchSize float64
	CostWithoutSync  float64
	CostWithSync     float64
	TotalSavings     float64
	SavingsPercent   float64
	FinalCoherence   float64
	PeakCoherence    float64
	TimeToConverge   time.Duration
}

// Task represents a single API task
type Task struct {
	ID      string
	AgentID string
	Type    string
	Payload interface{}
}

// APIBatch represents a collection of API calls to be processed together
type APIBatch struct {
	ID        string
	Tasks     []Task
	Timestamp time.Time
}

// BatchMetrics contains metrics about batch processing
type BatchMetrics struct {
	TotalBatches     int
	TotalTasks       int
	AverageBatchSize float64
	MaxBatchSize     int
	ProcessingTime   time.Duration
}
