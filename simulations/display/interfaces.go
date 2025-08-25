package display

import (
	"context"
	"time"
)

// Display defines the interface for visualization
type Display interface {
	// Initialize sets up the display
	Initialize() error

	// Close cleans up display resources
	Close() error

	// ShowWelcome displays the welcome screen
	ShowWelcome()

	// Update refreshes the display with new data
	Update(snapshot SimulationSnapshot)

	// ShowSummary displays final results
	ShowSummary(stats Statistics)

	// Run runs the display (for interactive displays)
	Run(ctx context.Context, controller *KeyboardController) error
}

// Controller defines the interface for user input
type Controller interface {
	// Events returns a channel of control events
	Events() <-chan ControlEvent
}

// SimulationSnapshot represents a point-in-time view of the simulation
type SimulationSnapshot struct {
	// Time information
	Timestamp   time.Time
	ElapsedTime time.Duration

	// Agent states
	Agents []AgentSnapshot

	// Synchronization metrics
	Coherence       float64
	TargetCoherence float64

	// Batch metrics
	PendingTasks     int
	CurrentBatchSize int
	BatchesProcessed int

	// Cost metrics
	CostWithoutSync float64
	CostWithSync    float64
	Savings         float64
	SavingsPercent  float64

	// Status
	Paused    bool
	Disrupted bool
	Reset     bool

	// Batch pulse animation
	BatchJustSent bool
	LastBatchTime time.Time
	LastBatchSize int
}

// AgentSnapshot represents the state of a single agent
type AgentSnapshot struct {
	ID            string
	Type          string // The type of agent (e.g., DataETL, Leader, etc.)
	Icon          string
	Phase         float64
	PendingTasks  int
	BatchesSent   int
	InBurstMode   bool   // True if agent is in burst mode
	ActivityLevel string // "burst", "quiet", "steady", "active"
}

// Statistics holds final simulation statistics
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

// ControlEvent represents a user control action
type ControlEvent struct {
	Type EventType
	Data interface{}
}

// EventType represents the type of control event
type EventType int

// Control event types
const (
	EventQuit       EventType = iota // EventQuit signals to quit the application
	EventReset                       // EventReset signals to reset the simulation
	EventDisrupt                     // EventDisrupt signals to disrupt the simulation
	EventPause                       // EventPause signals to pause the simulation
	EventResume                      // EventResume signals to resume the simulation
	ScaleTiny                        // ScaleTiny signals to switch to tiny scale
	ScaleSmall                       // ScaleSmall signals to switch to small scale
	ScaleMedium                      // ScaleMedium signals to switch to medium scale
	ScaleLarge                       // ScaleLarge signals to switch to large scale
	ScaleHuge                        // ScaleHuge signals to switch to huge scale
	GoalBatch                        // GoalBatch signals to optimize for API batching
	GoalLoad                         // GoalLoad signals to optimize for load distribution
	GoalConsensus                    // GoalConsensus signals to optimize for consensus
	GoalLatency                      // GoalLatency signals to optimize for low latency
	GoalEnergy                       // GoalEnergy signals to optimize for energy saving
	GoalRhythm                       // GoalRhythm signals to optimize for periodic tasks
	GoalFailure                      // GoalFailure signals to optimize for fault recovery
	GoalTraffic                      // GoalTraffic signals to optimize for traffic adaptation
	PatternHighFreq                  // PatternHighFreq signals to use high-frequency pattern
	PatternBurst                     // PatternBurst signals to use burst pattern
	PatternSteady                    // PatternSteady signals to use steady pattern
	PatternMixed                     // PatternMixed signals to use mixed pattern
	PatternSparse                    // PatternSparse signals to use sparse pattern
)
