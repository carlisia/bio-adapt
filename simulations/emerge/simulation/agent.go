// agent.go implements workloads that use emerge for synchronization.
// Each agent generates tasks and coordinates API calls with other agents.
//
// ARCHITECTURE:
// AIAgent (this file) wraps an emerge.Agent (from the framework).
// The emerge agent handles synchronization physics (phase, frequency, coupling).
// The workload adds application logic (simulating API calls).
//
// Think of it like a dancer (workload) following a metronome (emerge agent):
// - The metronome provides the beat (phase synchronization)
// - The dancer performs actions on the beat (making API calls)
// - When all dancers sync to the same beat, they can coordinate moves (batch API calls)

package simulation

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/carlisia/bio-adapt/emerge/agent"
	"github.com/carlisia/bio-adapt/emerge/goal"
	"github.com/carlisia/bio-adapt/simulations/emerge/simulation/pattern"
)

// Activity level constants
const (
	activitySteady = "steady"
	activityActive = "active"
	activityQuiet  = "quiet"
	activityBurst  = "burst"
)

// secureRandFloat32 returns a cryptographically secure random float32
func secureRandFloat32() float32 {
	var b [4]byte
	if _, err := cryptorand.Read(b[:]); err != nil {
		// Fall back to a default value on error
		return 0.5
	}
	return float32(binary.LittleEndian.Uint32(b[:])) / float32(1<<32)
}

// secureRandIntn returns a cryptographically secure random int
func secureRandIntn(n int) int {
	if n <= 0 {
		return 0
	}
	var b [8]byte
	if _, err := cryptorand.Read(b[:]); err != nil {
		// Fall back to a default value on error
		return n / 2
	}
	// Convert n to uint64 for safe arithmetic
	nUint := uint64(n)
	if nUint > math.MaxInt32 {
		nUint = math.MaxInt32
	}
	// Safe conversion: result is guaranteed to be < n after modulo
	result := binary.LittleEndian.Uint64(b[:]) % nUint
	// Since result < nUint and nUint <= MaxInt32, this conversion is safe
	// Additional explicit check for gosec
	if result > math.MaxInt32 {
		return int(math.MaxInt32)
	}
	return int(result)
}

// Workload represents a specific type of work that makes API calls.
// It wraps an emerge.Agent to inherit synchronization behavior.
//
// Key concept: The emerge agent tells us WHEN to act (via phase).
// The workload decides WHAT to do (generate tasks, make API calls).
type Workload struct {
	id          string
	agentType   AgentType
	emergeAgent *agent.Agent // The underlying emerge agent for synchronization
	pattern     pattern.Type // Request pattern for this agent

	// Task management - simulates work the workload needs to do
	pendingTasks []Task
	taskCounter  int
	mu           sync.Mutex

	// Pattern-specific state
	lastRequestTime time.Time
	burstActive     bool
	burstEndTime    time.Time

	// Metrics
	tasksGenerated int
	batchesSent    int

	// Batch callback - where we send our API calls when synchronized
	batchManager *BatchManager

	// Reference to check pause state
	simulation *Simulation
}

// AgentType defines the type of a workload
type AgentType struct {
	Name  string
	Icon  string
	Color string
}

// Goal-specific workload sets - each goal has workloads that make its optimization obvious
var goalWorkloads = map[goal.Type][]AgentType{
	// MinimizeAPICalls: Batch processing workloads (optimal: Steady pattern, Medium+ scale)
	goal.MinimizeAPICalls: {
		{"DataETL", "📊", "cyan"},        // Extract-Transform-Load pipeline
		{"LogProcessor", "📝", "green"},  // Log analysis and processing
		{"MLTraining", "🤖", "yellow"},   // Machine learning model training
		{"ReportGen", "📈", "magenta"},   // Batch report generation
		{"ImageProc", "🖼️", "blue"},     // Bulk image processing
		{"VideoTrans", "🎥", "white"},    // Video transcoding jobs
		{"DataValidator", "✅", "red"},   // Bulk data validation
		{"FileConverter", "📄", "green"}, // Batch file conversions
	},
	// DistributeLoad: Server workloads (optimal: Burst pattern, Large scale for anti-phase)
	goal.DistributeLoad: {
		{"WebServer1", "🌐", "cyan"},      // Web server instance 1
		{"WebServer2", "🌐", "green"},     // Web server instance 2
		{"APIGateway1", "🔌", "yellow"},   // API gateway instance 1
		{"APIGateway2", "🔌", "magenta"},  // API gateway instance 2
		{"LoadBalancer1", "⚖️", "blue"},  // Load balancer instance 1
		{"LoadBalancer2", "⚖️", "white"}, // Load balancer instance 2
		{"CacheServer1", "💾", "red"},     // Cache server instance 1
		{"CacheServer2", "💾", "green"},   // Cache server instance 2
	},
	// ReachConsensus: Distributed consensus workloads (optimal: Steady pattern, Small scale)
	goal.ReachConsensus: {
		{"Leader", "👑", "cyan"},       // Current consensus leader
		{"Candidate1", "🎯", "green"},  // Candidate for leadership
		{"Candidate2", "🎯", "yellow"}, // Candidate for leadership
		{"Follower1", "📡", "magenta"}, // Follower replica 1
		{"Follower2", "📡", "blue"},    // Follower replica 2
		{"Follower3", "📡", "white"},   // Follower replica 3
		{"DataNode1", "💾", "red"},     // Data node replica 1
		{"DataNode2", "💾", "green"},   // Data node replica 2
	},
	// MinimizeLatency: Real-time workloads (optimal: HighFrequency pattern, Tiny scale)
	goal.MinimizeLatency: {
		{"GameServer", "🎮", "cyan"},    // Real-time game server
		{"StreamProc", "📡", "green"},   // Stream processor
		{"RTAnalytics", "⚡", "yellow"}, // Real-time analytics
		{"LiveChat", "💬", "magenta"},   // Live chat system
		{"Ticker", "📊", "blue"},        // Stock ticker
		{"Monitor", "📈", "white"},      // System monitor
		{"Alerter", "🚨", "red"},        // Alert system
		{"Tracker", "🎯", "green"},      // Event tracker
	},
	// SaveEnergy: IoT/sensor workloads (optimal: Sparse pattern, Large scale)
	goal.SaveEnergy: {
		{"TempSensor", "🌡️", "cyan"},     // Temperature sensor
		{"HumiditySensor", "💧", "green"}, // Humidity sensor
		{"LightSensor", "💡", "yellow"},   // Light sensor
		{"MotionSensor", "🏃", "magenta"}, // Motion sensor
		{"SmokeSensor", "🔥", "blue"},     // Smoke detector
		{"DoorSensor", "🚪", "white"},     // Door sensor
		{"PowerMeter", "⚡", "red"},       // Power meter
		{"WaterMeter", "💧", "green"},     // Water meter
	},
	// MaintainRhythm: Scheduled workloads (optimal: Steady pattern, Medium scale)
	goal.MaintainRhythm: {
		{"CronJob1", "⏰", "cyan"},   // Scheduled job 1
		{"CronJob2", "⏰", "green"},  // Scheduled job 2
		{"Backup1", "💾", "yellow"},  // Backup task 1
		{"Backup2", "💾", "magenta"}, // Backup task 2
		{"Cleanup1", "🧹", "blue"},   // Cleanup task 1
		{"Cleanup2", "🧹", "white"},  // Cleanup task 2
		{"Reporter1", "📊", "red"},   // Report generator 1
		{"Reporter2", "📊", "green"}, // Report generator 2
	},
	// RecoverFromFailure: Resilient workloads (optimal: Mixed pattern, Medium scale)
	goal.RecoverFromFailure: {
		{"Primary1", "🔵", "cyan"},    // Primary node 1
		{"Primary2", "🔵", "green"},   // Primary node 2
		{"Replica1", "🟢", "yellow"},  // Replica node 1
		{"Replica2", "🟢", "magenta"}, // Replica node 2
		{"Standby1", "🟡", "blue"},    // Standby node 1
		{"Standby2", "🟡", "white"},   // Standby node 2
		{"Monitor1", "👁️", "red"},    // Health monitor 1
		{"Monitor2", "👁️", "green"},  // Health monitor 2
	},
	// AdaptToTraffic: Dynamic workloads (optimal: Burst pattern, Medium scale)
	goal.AdaptToTraffic: {
		{"AutoScaler1", "📏", "cyan"},  // Auto-scaler 1
		{"AutoScaler2", "📏", "green"}, // Auto-scaler 2
		{"CDN1", "🌍", "yellow"},       // CDN node 1
		{"CDN2", "🌍", "magenta"},      // CDN node 2
		{"Queue1", "📬", "blue"},       // Message queue 1
		{"Queue2", "📬", "white"},      // Message queue 2
		{"Worker1", "⚙️", "red"},      // Worker node 1
		{"Worker2", "⚙️", "green"},    // Worker node 2
	},
}

// getWorkloadsForGoal returns the appropriate workload set for a given goal
func getWorkloadsForGoal(goalType goal.Type) []AgentType {
	if roles, ok := goalWorkloads[goalType]; ok {
		return roles
	}
	// Default to MinimizeAPICalls workloads if goal not found
	return goalWorkloads[goal.MinimizeAPICalls]
}

// NewWorkload creates a new workload (typically not used - see NewWorkloadWithEmerge)
func NewWorkload(id string, typeIndex int, goalType goal.Type) *Workload {
	types := getWorkloadsForGoal(goalType)
	agentType := types[typeIndex%len(types)]

	return &Workload{
		id:           id,
		agentType:    agentType,
		emergeAgent:  agent.New(id),
		pendingTasks: make([]Task, 0),
	}
}

// NewWorkloadWithEmerge creates a workload that wraps an existing emerge agent.
// This is the primary constructor used by the builder.
//
// The emerge agent (from the swarm) provides synchronization.
// The workload adds task processing behavior on top.
func NewWorkloadWithEmerge(
	id string,
	typeIndex int,
	emergeAgent *agent.Agent,
	sim *Simulation,
	requestPattern pattern.Type,
	goalType goal.Type,
) *Workload {
	types := getWorkloadsForGoal(goalType)
	agentType := types[typeIndex%len(types)]

	return &Workload{
		id:           id,
		agentType:    agentType,
		emergeAgent:  emergeAgent, // Wrap the emerge agent
		pattern:      requestPattern,
		pendingTasks: make([]Task, 0),
		simulation:   sim,
	}
}

// Start begins the agent's operation
func (w *Workload) Start(ctx context.Context, batchManager *BatchManager) {
	w.batchManager = batchManager

	// Start task generation
	go w.generateTasks(ctx)

	// Start batch monitoring
	go w.monitorBatching(ctx)
}

// generateTasks simulates the agent creating work based on the request pattern
func (w *Workload) generateTasks(ctx context.Context) {
	// Base timing depends on pattern
	var interval time.Duration

	switch w.pattern {
	case pattern.Unset:
		// Should not happen - pattern should be set before agent starts
		// Fall back to steady
		interval = 500 * time.Millisecond
	case pattern.HighFrequency:
		// Very frequent requests (>10/sec)
		interval = 50*time.Millisecond + time.Duration(secureRandIntn(50))*time.Millisecond
	case pattern.Burst:
		// Will be handled specially in the loop
		interval = 100 * time.Millisecond
	case pattern.Steady:
		// Consistent, predictable rate
		interval = 500 * time.Millisecond
	case pattern.Mixed:
		// Will vary in the loop
		interval = 300 * time.Millisecond
	case pattern.Sparse:
		// Infrequent requests (<1/sec)
		interval = 2*time.Second + time.Duration(secureRandIntn(2000))*time.Millisecond
	default:
		// Default to steady
		interval = 500 * time.Millisecond
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Skip task generation when paused
			if w.simulation != nil {
				w.simulation.mu.RLock()
				paused := w.simulation.paused
				w.simulation.mu.RUnlock()
				if paused {
					continue
				}
			}

			// Pattern-specific task generation logic
			shouldGenerate := w.shouldGenerateTask()
			if !shouldGenerate {
				continue
			}

			// Generate multiple tasks for burst/high-frequency patterns
			numTasks := w.getTaskBatchSize()
			for range numTasks {
				task := w.createTask()
				w.mu.Lock()
				w.pendingTasks = append(w.pendingTasks, task)
				w.tasksGenerated++
				w.mu.Unlock()
			}

			// Update interval for mixed pattern
			if w.pattern == pattern.Mixed {
				// Randomly vary the interval
				ticker.Reset(time.Duration(200+secureRandIntn(800)) * time.Millisecond)
			}
		}
	}
}

// shouldGenerateTask determines if a task should be generated based on pattern
func (w *Workload) shouldGenerateTask() bool {
	now := time.Now()

	switch w.pattern {
	case pattern.Burst:
		// Burst pattern: periods of high activity followed by quiet
		if !w.burstActive {
			// Start a new burst with 20% probability
			if secureRandFloat32() < 0.2 {
				w.burstActive = true
				w.burstEndTime = now.Add(time.Duration(500+secureRandIntn(1500)) * time.Millisecond)
				return true
			}
			return false
		}
		// Check if burst should end
		if now.After(w.burstEndTime) {
			w.burstActive = false
			return false
		}
		return true

	case pattern.Sparse:
		// Sparse pattern: random gaps between requests
		return secureRandFloat32() < 0.3 // Only 30% chance of generating

	case pattern.HighFrequency:
		// High frequency: always generate
		return true

	case pattern.Steady:
		// Steady: regular generation
		return true

	case pattern.Mixed:
		// Mixed: varied probability
		return secureRandFloat32() < 0.7

	case pattern.Unset:
		// Default behavior
		return true

	default:
		// Always generate for other patterns
		return true
	}
}

// getTaskBatchSize returns how many tasks to generate at once based on pattern
func (w *Workload) getTaskBatchSize() int {
	switch w.pattern {
	case pattern.HighFrequency:
		// Generate multiple tasks at once
		return 2 + secureRandIntn(3) // 2-4 tasks
	case pattern.Burst:
		if w.burstActive {
			return 3 + secureRandIntn(5) // 3-7 tasks during burst
		}
		return 1
	case pattern.Mixed:
		// Sometimes generate multiple, sometimes single
		if secureRandFloat32() < 0.3 {
			return 2 + secureRandIntn(2) // 2-3 tasks
		}
		return 1
	case pattern.Steady:
		// Steady: consistent single tasks
		return 1
	case pattern.Sparse:
		// Sparse: single tasks when they occur
		return 1
	case pattern.Unset:
		// Default single task
		return 1
	default:
		return 1
	}
}

// createTask creates a new task
func (w *Workload) createTask() Task {
	w.taskCounter++

	// Goal-specific task types
	taskTypes := map[string][]string{
		// MinimizeAPICalls workloads
		"DataETL":       {"extract_records", "transform_dataset", "load_warehouse"},
		"LogProcessor":  {"parse_logs", "aggregate_metrics", "detect_anomalies"},
		"MLTraining":    {"train_model", "evaluate_model", "tune_hyperparams"},
		"ReportGen":     {"generate_daily", "generate_weekly", "generate_monthly"},
		"ImageProc":     {"resize_batch", "optimize_batch", "convert_format"},
		"VideoTrans":    {"transcode_1080p", "transcode_720p", "extract_audio"},
		"DataValidator": {"validate_schema", "check_integrity", "verify_rules"},
		"FileConverter": {"convert_pdf", "convert_csv", "convert_json"},

		// DistributeLoad workloads
		"WebServer1":    {"handle_request", "serve_static", "process_api"},
		"WebServer2":    {"handle_request", "serve_static", "process_api"},
		"APIGateway1":   {"route_request", "validate_auth", "rate_limit"},
		"APIGateway2":   {"route_request", "validate_auth", "rate_limit"},
		"LoadBalancer1": {"distribute_load", "health_check", "route_traffic"},
		"LoadBalancer2": {"distribute_load", "health_check", "route_traffic"},
		"CacheServer1":  {"get_cache", "set_cache", "invalidate_cache"},
		"CacheServer2":  {"get_cache", "set_cache", "invalidate_cache"},

		// ReachConsensus workloads (distributed systems)
		"Leader":     {"send_heartbeat", "replicate_log", "commit_entries"},
		"Candidate1": {"request_votes", "increment_term", "check_majority"},
		"Candidate2": {"request_votes", "increment_term", "check_majority"},
		"Follower1":  {"append_entries", "vote_response", "apply_commits"},
		"Follower2":  {"append_entries", "vote_response", "apply_commits"},
		"Follower3":  {"append_entries", "vote_response", "apply_commits"},
		"DataNode1":  {"sync_data", "verify_checksum", "acknowledge_write"},
		"DataNode2":  {"sync_data", "verify_checksum", "acknowledge_write"},

		// MinimizeLatency workloads
		"GameServer":  {"update_state", "process_input", "send_update"},
		"StreamProc":  {"process_stream", "buffer_data", "emit_event"},
		"RTAnalytics": {"analyze_realtime", "compute_metric", "alert_threshold"},
		"LiveChat":    {"send_message", "receive_message", "update_presence"},
		"Ticker":      {"update_price", "broadcast_tick", "calculate_delta"},
		"Monitor":     {"check_status", "log_metric", "raise_alert"},
		"Alerter":     {"check_condition", "send_alert", "escalate_issue"},
		"Tracker":     {"track_event", "update_counter", "emit_metric"},

		// SaveEnergy workloads
		"TempSensor":     {"read_temperature", "calibrate", "send_reading"},
		"HumiditySensor": {"read_humidity", "calibrate", "send_reading"},
		"LightSensor":    {"read_luminance", "calibrate", "send_reading"},
		"MotionSensor":   {"detect_motion", "calibrate", "send_alert"},
		"SmokeSensor":    {"detect_smoke", "test_alarm", "send_alert"},
		"DoorSensor":     {"check_status", "log_event", "send_notification"},
		"PowerMeter":     {"read_consumption", "calculate_cost", "send_report"},
		"WaterMeter":     {"read_flow", "detect_leak", "send_report"},

		// MaintainRhythm workloads
		"CronJob1":  {"execute_task", "check_schedule", "log_completion"},
		"CronJob2":  {"execute_task", "check_schedule", "log_completion"},
		"Backup1":   {"backup_data", "verify_backup", "rotate_old"},
		"Backup2":   {"backup_data", "verify_backup", "rotate_old"},
		"Cleanup1":  {"clean_temp", "remove_logs", "free_space"},
		"Cleanup2":  {"clean_temp", "remove_logs", "free_space"},
		"Reporter1": {"generate_report", "email_report", "archive_report"},
		"Reporter2": {"generate_report", "email_report", "archive_report"},

		// RecoverFromFailure workloads
		"Primary1": {"process_request", "replicate_state", "heartbeat"},
		"Primary2": {"process_request", "replicate_state", "heartbeat"},
		"Replica1": {"sync_state", "standby_ready", "promote_primary"},
		"Replica2": {"sync_state", "standby_ready", "promote_primary"},
		"Standby1": {"monitor_health", "backup_state", "failover_ready"},
		"Standby2": {"monitor_health", "backup_state", "failover_ready"},
		"Monitor1": {"check_health", "detect_failure", "initiate_recovery"},
		"Monitor2": {"check_health", "detect_failure", "initiate_recovery"},

		// AdaptToTraffic workloads
		"AutoScaler1": {"monitor_load", "scale_up", "scale_down"},
		"AutoScaler2": {"monitor_load", "scale_up", "scale_down"},
		"CDN1":        {"cache_content", "serve_cached", "invalidate_stale"},
		"CDN2":        {"cache_content", "serve_cached", "invalidate_stale"},
		"Queue1":      {"enqueue_message", "dequeue_message", "check_depth"},
		"Queue2":      {"enqueue_message", "dequeue_message", "check_depth"},
		"Worker1":     {"process_job", "report_progress", "complete_task"},
		"Worker2":     {"process_job", "report_progress", "complete_task"},
	}

	tasks, exists := taskTypes[w.agentType.Name]
	if !exists {
		tasks = []string{"process_request"}
	}

	taskType := tasks[w.taskCounter%len(tasks)]

	return Task{
		ID:      fmt.Sprintf("%s-task-%d", w.id, w.taskCounter),
		AgentID: w.id,
		Type:    taskType,
		Payload: fmt.Sprintf("%s: %s #%d", w.agentType.Name, taskType, w.taskCounter),
	}
}

// monitorBatching checks when to send batches
func (w *Workload) monitorBatching(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Skip batch monitoring when paused
			if w.simulation != nil {
				w.simulation.mu.RLock()
				paused := w.simulation.paused
				w.simulation.mu.RUnlock()
				if paused {
					continue
				}
			}

			if w.shouldBatch() {
				w.sendBatch()
			}
		}
	}
}

// shouldBatch determines if agent should send its batch
func (w *Workload) shouldBatch() bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.pendingTasks) == 0 {
		return false
	}

	// Check synchronization phase
	phase := w.emergeAgent.Phase()
	// Batch when phase is near 0 (modulo 2π)
	return phase < 0.1 || phase > (2*math.Pi-0.1)
}

// sendBatch sends pending tasks as a batch
func (w *Workload) sendBatch() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.pendingTasks) == 0 {
		return
	}

	// Copy tasks
	batch := make([]Task, len(w.pendingTasks))
	copy(batch, w.pendingTasks)

	// Clear pending
	w.pendingTasks = w.pendingTasks[:0]

	// Update metrics
	w.batchesSent++

	// Send to batch manager
	if w.batchManager != nil {
		w.batchManager.SubmitBatch(w.id, batch)
	}
}

// Reset resets the workload counters
func (w *Workload) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Reset counters
	w.pendingTasks = w.pendingTasks[:0] // Clear but keep capacity
	w.taskCounter = 0
	w.tasksGenerated = 0
	w.batchesSent = 0

	// Reset pattern state
	w.lastRequestTime = time.Time{}
	w.burstActive = false
	w.burstEndTime = time.Time{}
}

// Snapshot returns agent's current state
func (w *Workload) Snapshot() AgentSnapshot {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Determine activity level based on pattern and state
	activityLevel := activitySteady
	inBurstMode := false

	switch w.pattern {
	case pattern.Burst:
		if w.burstActive {
			activityLevel = activityBurst
			inBurstMode = true
		} else {
			activityLevel = activityQuiet
		}
	case pattern.HighFrequency:
		activityLevel = activityActive
	case pattern.Sparse:
		if len(w.pendingTasks) > 0 {
			activityLevel = activityActive
		} else {
			activityLevel = activityQuiet
		}
	case pattern.Mixed:
		// Check recent activity
		if time.Since(w.lastRequestTime) < 200*time.Millisecond {
			activityLevel = activityActive
		} else if time.Since(w.lastRequestTime) > 1*time.Second {
			activityLevel = activityQuiet
		}
	case pattern.Steady:
		// Steady pattern has consistent activity
		activityLevel = activitySteady
	case pattern.Unset:
		// Default to steady
		activityLevel = activitySteady
	default:
		activityLevel = activitySteady
	}

	return AgentSnapshot{
		ID:            w.id,
		Type:          w.agentType.Name,
		Icon:          w.agentType.Icon,
		Phase:         w.emergeAgent.Phase(),
		PendingTasks:  len(w.pendingTasks),
		BatchesSent:   w.batchesSent,
		InBurstMode:   inBurstMode,
		ActivityLevel: activityLevel,
	}
}

// EmergeAgent returns the underlying emerge agent
func (w *Workload) EmergeAgent() *agent.Agent {
	return w.emergeAgent
}
