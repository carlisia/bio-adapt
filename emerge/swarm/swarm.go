package swarm

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/carlisia/bio-adapt/emerge/agent"
	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/monitoring"
	"github.com/carlisia/bio-adapt/internal/config"
	"github.com/carlisia/bio-adapt/internal/emerge"
	"github.com/carlisia/bio-adapt/internal/random"
)

// Implementation selection thresholds
const (
	// OptimizedSwarmThreshold is the size above which we use the optimized implementation
	OptimizedSwarmThreshold = 100

	// LargeSwarmThreshold is the size above which we enable additional optimizations
	LargeSwarmThreshold = 1000
)

// Swarm represents a collection of autonomous agents achieving
// synchronization through emergent behavior - no central control.
type Swarm struct {
	// Storage - automatically selected based on size
	agents      sync.Map       // map[string]*agent.Agent - used for small swarms
	agentSlice  []*agent.Agent // Direct access by index - used for large swarms
	agentIndex  map[string]int // ID to index mapping - used for large swarms
	agentsMutex sync.RWMutex   // Protects slice access
	optimized   bool           // Whether using optimized storage

	goalState core.State
	config    config.Swarm
	size      int

	// Monitoring (read-only, doesn't control) - private implementation details
	monitor     *monitoring.Monitor
	basin       *emerge.AttractorBasin
	convergence *monitoring.ConvergenceMonitor

	// Track initialization state
	connectionsEstablished bool

	// Goal-directed synchronization
	goalDirectedSync *GoalDirectedSync
	goalConfig       *Config // Configuration for goal-directed sync

	// Performance optimization for large swarms
	workerPool *WorkerPool // Goroutine pool for concurrent updates

	// Recovery configuration for continuous operation
	recoveryConfig RecoveryConfig
}

// Option configures a Swarm.
type Option func(*Swarm) error

// New creates a swarm with automatic storage optimization based on size.
// For smaller swarms (≤100 agents), uses sync.Map for simplicity.
// For larger swarms (>100 agents), uses slice-based storage with better cache locality.
//
//nolint:gocyclo // Initialization requires multiple validation steps
func New(size int, goal core.State, opts ...Option) (*Swarm, error) {
	if size <= 0 {
		return nil, fmt.Errorf("%w: got %d", ErrInvalidSwarmSize, size)
	}

	// Validate goal state using centralized validation
	if err := goal.Validate(); err != nil {
		return nil, fmt.Errorf("invalid goal state: %w", err)
	}

	// Start with auto-scaled config as default
	cfg := config.AutoScaleConfig(size)
	if err := cfg.NormalizeAndValidate(size); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// Validate coherence target and adjust if impossible
	limits := GetCoherenceLimits(size)
	if goal.Coherence > limits.Practical {
		// Adjust to practical limit with small buffer
		goal.Coherence = limits.Practical
	}

	// Initialize swarm with optimized storage for large sizes
	s := &Swarm{
		goalState:      goal,
		config:         cfg,
		size:           size,
		monitor:        monitoring.New(),
		basin:          emerge.NewAttractorBasin(goal, cfg.BasinStrength, cfg.BasinWidth),
		convergence:    monitoring.NewConvergence(goal, goal.Coherence),
		optimized:      size > OptimizedSwarmThreshold,
		recoveryConfig: DefaultRecoveryConfig(goal.Coherence),
	}

	// Initialize optimized storage for large swarms
	if s.optimized {
		s.agentSlice = make([]*agent.Agent, 0, size)
		s.agentIndex = make(map[string]int, size)
		// Initialize worker pool for concurrent operations
		numWorkers := getOptimalWorkerCount(size)
		s.workerPool = NewWorkerPool(numWorkers)
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, fmt.Errorf("failed to apply swarm option: %w", err)
		}
	}

	// Re-validate config after options are applied
	if err := s.config.NormalizeAndValidate(size); err != nil {
		return nil, fmt.Errorf("config validation failed after options: %w", err)
	}

	// Check configurable size limit after options are applied
	if s.config.MaxSwarmSize > 0 && size > s.config.MaxSwarmSize {
		return nil, fmt.Errorf("swarm size %d exceeds configured maximum %d", size, s.config.MaxSwarmSize)
	}

	// Create agents if not already created by options
	if s.optimized {
		// Use optimized creation for large swarms
		if len(s.agentSlice) == 0 {
			if err := s.createOptimizedAgents(); err != nil {
				return nil, fmt.Errorf("failed to create agents: %w", err)
			}
		}
	} else {
		// Use standard creation for small swarms
		agentCount := 0
		s.agents.Range(func(_, _ any) bool {
			agentCount++
			return false // Just need to check if any exist
		})
		if agentCount == 0 {
			if err := s.createDefaultAgents(); err != nil {
				return nil, fmt.Errorf("failed to create agents: %w", err)
			}
		}
	}

	// Establish connections if not already done
	if !s.connectionsEstablished {
		s.establishConnections()
	}

	// Initialize goal-directed synchronization
	if s.goalConfig != nil {
		s.goalDirectedSync = NewGoalDirectedSyncWithConfig(s, s.goalConfig)
	} else {
		s.goalDirectedSync = NewGoalDirectedSync(s)
	}

	return s, nil
}

// NewSwarmFromConfig creates a swarm using a configuration struct.
// This provides an alternative to the functional options pattern.
func NewSwarmFromConfig(size int, goal core.State, cfg config.Swarm) (*Swarm, error) {
	if size <= 0 {
		return nil, fmt.Errorf("%w: got %d", ErrInvalidSwarmSize, size)
	}

	// Validate goal state using centralized validation
	if err := goal.Validate(); err != nil {
		return nil, fmt.Errorf("invalid goal state: %w", err)
	}

	// Validate and normalize config using centralized validation
	if err := cfg.NormalizeAndValidate(size); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	s := &Swarm{
		goalState:   goal,
		config:      cfg,
		size:        size,
		monitor:     monitoring.New(),
		basin:       emerge.NewAttractorBasin(goal, cfg.BasinStrength, cfg.BasinWidth),
		convergence: monitoring.NewConvergence(goal, goal.Coherence),
	}

	// Create agents
	if err := s.createDefaultAgents(); err != nil {
		return nil, fmt.Errorf("failed to create agents: %w", err)
	}

	// Establish connections
	s.establishConnections()

	// Initialize goal-directed synchronization
	if s.goalConfig != nil {
		s.goalDirectedSync = NewGoalDirectedSyncWithConfig(s, s.goalConfig)
	} else {
		s.goalDirectedSync = NewGoalDirectedSync(s)
	}

	return s, nil
}

// createDefaultAgents creates agents with config-based settings for standard storage.
func (s *Swarm) createDefaultAgents() error {
	agentConfig := config.AgentFromSwarm(s.config)
	agentConfig.SwarmSize = s.size

	for i := range s.size {
		a, err := agent.NewFromConfig(fmt.Sprintf("agent-%d", i), agentConfig)
		if err != nil {
			return fmt.Errorf("failed to create agent %d: %w", i, err)
		}
		s.agents.Store(a.ID, a)
	}
	return nil
}

// createOptimizedAgents creates agents efficiently for large swarms.
func (s *Swarm) createOptimizedAgents() error {
	agentConfig := config.AgentFromSwarm(s.config)
	agentConfig.SwarmSize = s.size

	// Pre-allocate the slice
	s.agentSlice = make([]*agent.Agent, s.size)

	// Batch create optimized agents for better performance
	batchSize := 100
	for i := 0; i < s.size; i += batchSize {
		end := minInt(i+batchSize, s.size)

		for j := i; j < end; j++ {
			id := fmt.Sprintf("agent-%d", j)

			// Create agent with pre-allocated neighbor storage
			a, err := agent.NewOptimizedFromConfig(id, agentConfig)
			if err != nil {
				return fmt.Errorf("failed to create agent %d: %w", j, err)
			}

			// Store the Agent directly
			s.agentSlice[j] = a
			s.agentIndex[id] = j
		}
	}

	return nil
}

// WithConfig sets a custom configuration.
func WithConfig(cfg config.Swarm) Option {
	return func(s *Swarm) error {
		if err := cfg.NormalizeAndValidate(s.size); err != nil {
			return fmt.Errorf("config validation failed: %w", err)
		}
		s.config = cfg
		// Update basin and convergence with new config
		s.basin = emerge.NewAttractorBasin(s.goalState, cfg.BasinStrength, cfg.BasinWidth)
		return nil
	}
}

// WithGoalConfig sets a custom configuration for goal-directed synchronization.
func WithGoalConfig(cfg *Config) Option {
	return func(s *Swarm) error {
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("goal config validation failed: %w", err)
		}
		s.goalConfig = cfg
		return nil
	}
}

// WithMonitor sets a custom monitor.
func WithMonitor(monitor *monitoring.Monitor) Option {
	return func(s *Swarm) error {
		s.monitor = monitor
		return nil
	}
}

// WithAgentBuilder uses a custom function to create agents.
func WithAgentBuilder(builder func(id string, swarmSize int, cfg config.Swarm) *agent.Agent) Option {
	return func(s *Swarm) error {
		for i := range s.size {
			a := builder(fmt.Sprintf("agent-%d", i), s.size, s.config)
			s.agents.Store(a.ID, a)
		}
		return nil
	}
}

// WithTopology uses a custom topology builder.
func WithTopology(builder func(*Swarm) error) Option {
	return func(s *Swarm) error {
		if err := builder(s); err != nil {
			return fmt.Errorf("topology build failed: %w", err)
		}
		s.connectionsEstablished = true
		return nil
	}
}

// WithRecoveryConfig sets a custom recovery configuration for continuous operation.
// This controls how the swarm detects and recovers from disruptions.
func WithRecoveryConfig(cfg RecoveryConfig) Option {
	return func(s *Swarm) error {
		s.recoveryConfig = cfg
		return nil
	}
}

// establishConnections creates network topology based on configuration.
func (s *Swarm) establishConnections() {
	agents := s.collectAgents()

	if s.config.EnableConnectionOptim && len(agents) > s.config.ConnectionOptimThreshold {
		s.establishMinimalConnections(agents)
	} else {
		s.establishProbabilisticConnections(agents)
	}
}

// collectAgents gathers all agents from the swarm
func (s *Swarm) collectAgents() []*agent.Agent {
	if s.optimized {
		// Return slice directly for optimized storage
		s.agentsMutex.RLock()
		defer s.agentsMutex.RUnlock()
		result := make([]*agent.Agent, len(s.agentSlice))
		copy(result, s.agentSlice)
		return result
	}

	// Standard path
	var agents []*agent.Agent
	s.agents.Range(func(_, value any) bool {
		if a, ok := value.(*agent.Agent); ok {
			agents = append(agents, a)
		}
		return true
	})
	return agents
}

// establishMinimalConnections creates minimal random connections for large swarms
func (s *Swarm) establishMinimalConnections(agents []*agent.Agent) {
	// For optimized agents, use ConnectTo method which handles optimized storage
	for _, a := range agents {
		connected := 0
		attempts := 0
		maxAttempts := len(agents) * 2

		for connected < s.config.MinNeighbors && connected < len(agents)-1 && attempts < maxAttempts {
			idx := random.Intn(len(agents))
			neighbor := agents[idx]

			if neighbor.ID != a.ID {
				// Check if already connected
				if _, exists := a.Neighbors().Load(neighbor.ID); !exists {
					a.Neighbors().Store(neighbor.ID, neighbor)
					neighbor.Neighbors().Store(a.ID, a)
					connected++
				}
			}
			attempts++
		}
	}
}

// establishProbabilisticConnections creates connections based on probability for smaller swarms
func (s *Swarm) establishProbabilisticConnections(agents []*agent.Agent) {
	for i, a := range agents {
		connected := s.connectToNeighbors(a, agents, i)
		s.ensureMinimumConnectivity(a, agents, connected)
	}
}

// connectToNeighbors connects an agent to neighbors based on probability
func (s *Swarm) connectToNeighbors(a *agent.Agent, agents []*agent.Agent, agentIndex int) int {
	connected := 0

	for j, neighbor := range agents {
		if j == agentIndex {
			continue
		}

		if connected >= s.config.MaxNeighbors {
			break
		}

		if _, exists := a.Neighbors().Load(neighbor.ID); exists {
			connected++
			continue
		}

		if random.Float64() < s.config.ConnectionProbability {
			a.Neighbors().Store(neighbor.ID, neighbor)
			neighbor.Neighbors().Store(a.ID, a)
			connected++
		}
	}

	return connected
}

// ensureMinimumConnectivity forces connections to meet minimum neighbor requirements
func (s *Swarm) ensureMinimumConnectivity(a *agent.Agent, agents []*agent.Agent, connected int) {
	if connected < s.config.MinNeighbors && len(agents) > s.config.MinNeighbors {
		for connected < s.config.MinNeighbors {
			idx := random.Intn(len(agents))
			neighbor := agents[idx]

			if neighbor.ID == a.ID {
				continue
			}

			if _, exists := a.Neighbors().Load(neighbor.ID); !exists {
				a.Neighbors().Store(neighbor.ID, neighbor)
				neighbor.Neighbors().Store(a.ID, a)
				connected++
			}
		}
	}
}

// Run starts the swarm and achieves synchronization using goal-directed pattern completion.
// Uses adaptive strategies and convergence dynamics to ensure goal achievement.
// This method exits once the target coherence is achieved or the context is canceled.
//
// Use Run() when you need:
//   - One-time synchronization (batch processing, initialization)
//   - Simple convergence without recovery (tests, benchmarks)
//   - To minimize resource usage (no continuous monitoring)
//
// Use RunContinuous() when you need:
//   - Long-running operation with automatic recovery from disruptions
//   - Continuous monitoring and maintenance of synchronization
//   - Production systems that must maintain coherence over time
func (s *Swarm) Run(ctx context.Context) error {
	// Create target pattern from goal state
	targetPattern := &core.TargetPattern{
		Phase:     s.goalState.Phase,
		Frequency: s.goalState.Frequency,
		Coherence: s.goalState.Coherence,
		Amplitude: 1.0,
		Stability: 0.9,
	}

	// Use goal-directed synchronization
	return s.goalDirectedSync.AchieveSynchronization(ctx, targetPattern)
}

// MeasureCoherence calculates global synchronization level.
// This is for monitoring only - agents don't have access to this.
func (s *Swarm) MeasureCoherence() float64 {
	if s.optimized {
		// Optimized path for large swarms - better cache locality
		s.agentsMutex.RLock()
		phases := make([]float64, len(s.agentSlice))
		for i, a := range s.agentSlice {
			phases[i] = a.Phase()
		}
		s.agentsMutex.RUnlock()
		return core.MeasureCoherence(phases)
	}

	// Standard path for small swarms
	var phases []float64
	s.agents.Range(func(_, value any) bool {
		if a, ok := value.(*agent.Agent); ok {
			phases = append(phases, a.Phase())
		}
		return true
	})
	return core.MeasureCoherence(phases)
}

// Agents returns all agents in the swarm.
func (s *Swarm) Agents() map[string]*agent.Agent {
	if s.optimized {
		// Optimized path - convert slice to map
		s.agentsMutex.RLock()
		agents := make(map[string]*agent.Agent, len(s.agentSlice))
		for _, a := range s.agentSlice {
			agents[a.ID] = a
		}
		s.agentsMutex.RUnlock()
		return agents
	}

	// Standard path
	agents := make(map[string]*agent.Agent)
	s.agents.Range(func(key, value any) bool {
		if k, ok := key.(string); ok {
			if a, ok := value.(*agent.Agent); ok {
				agents[k] = a
			}
		}
		return true
	})
	return agents
}

// Agent retrieves an agent by ID.
func (s *Swarm) Agent(id string) (*agent.Agent, bool) {
	if s.optimized {
		// Optimized path - O(1) lookup
		s.agentsMutex.RLock()
		defer s.agentsMutex.RUnlock()
		if idx, ok := s.agentIndex[id]; ok {
			return s.agentSlice[idx], true
		}
		return nil, false
	}

	// Standard path
	value, exists := s.agents.Load(id)
	if !exists {
		return nil, false
	}
	if a, ok := value.(*agent.Agent); ok {
		return a, true
	}
	return nil, false
}

// Size returns the number of agents in the swarm.
func (s *Swarm) Size() int {
	return s.size
}

// Config returns the swarm configuration.
func (s *Swarm) Config() config.Swarm {
	return s.config
}

// IsConverged returns whether the swarm has reached convergence.
func (s *Swarm) IsConverged() bool {
	return s.convergence.IsConverged()
}

// ConvergenceTime returns how long it took to converge, or 0 if not converged.
func (s *Swarm) ConvergenceTime() time.Duration {
	return s.convergence.ConvergenceTime()
}

// CurrentCoherence returns the most recent coherence measurement.
func (s *Swarm) CurrentCoherence() float64 {
	return s.convergence.CurrentCoherence()
}

// DisruptAgents randomly disrupts a percentage of agents.
func (s *Swarm) DisruptAgents(percentage float64) {
	// Clamp percentage to valid range [0, 1]
	if percentage < 0 {
		percentage = 0
	} else if percentage > 1 {
		percentage = 1
	}

	targetCount := int(float64(s.Size()) * percentage)
	disrupted := 0

	s.agents.Range(func(_, value any) bool {
		if disrupted >= targetCount {
			return false
		}

		if a, ok := value.(*agent.Agent); ok {
			a.SetPhase(random.Phase())
			disrupted++
		}
		return true
	})

	// Important: After disruption, the goal-directed sync may have already
	// completed and returned from its AchieveSynchronization loop.
	// We need to ensure it continues working toward the goal.

	// Reset convergence tracking to clear stale state
	if s.convergence != nil {
		s.convergence.Reset()
	}

	// For goal-directed sync, we need to ensure it doesn't think it has
	// already achieved the goal. The problem is that once AchieveSynchronization
	// returns, the Run loop exits and no more updates happen.
	//
	// Since we can't restart a completed goroutine, the best approach is to
	// prevent the goal-directed sync from exiting prematurely by ensuring
	// it properly detects the disruption.
	if s.goalDirectedSync != nil && s.goalDirectedSync.convergenceMonitor != nil {
		// Reset the convergence monitor so it knows we're not converged anymore
		s.goalDirectedSync.convergenceMonitor.Reset()

		// HACK: Set a flag or state that forces the goal-directed sync to
		// re-evaluate convergence. Since we can't directly do this without
		// modifying the goal-directed sync structure, we'll take a different
		// approach in the actual fix.
	}
}

// ForEachAgent applies a function to each agent in the swarm.
// This provides controlled access to agents without exposing the internal map.
func (s *Swarm) ForEachAgent(fn func(*agent.Agent) bool) {
	s.agents.Range(func(_, value any) bool {
		if agent, ok := value.(*agent.Agent); ok {
			return fn(agent)
		}
		return true
	})
}

// AutoScaleConfig returns a configuration that automatically scales based on swarm size.
func AutoScaleConfig(swarmSize int) config.Swarm {
	return config.AutoScaleConfig(swarmSize)
}

// ConfigForBatching returns configuration optimized for request batching scenarios.
func ConfigForBatching(workloadCount int, batchWindow time.Duration) config.Swarm {
	return config.ForBatching(workloadCount, batchWindow)
}

// WorkerPool manages a fixed number of goroutines for agent updates.
// Used internally for optimized swarms to reduce goroutine creation overhead.
type WorkerPool struct {
	workers   int
	workQueue chan func()
	quit      chan struct{}
}

// NewWorkerPool creates a new worker pool.
func NewWorkerPool(workers int) *WorkerPool {
	wp := &WorkerPool{
		workers:   workers,
		workQueue: make(chan func(), workers*2),
		quit:      make(chan struct{}),
	}

	// Start workers
	for range workers {
		go wp.worker()
	}

	return wp
}

// worker processes tasks from the queue.
func (wp *WorkerPool) worker() {
	for {
		select {
		case work := <-wp.workQueue:
			work()
		case <-wp.quit:
			return
		}
	}
}

// Submit adds work to the queue.
func (wp *WorkerPool) Submit(work func()) {
	wp.workQueue <- work
}

// Stop shuts down the worker pool.
func (wp *WorkerPool) Stop() {
	close(wp.quit)
}

// getOptimalWorkerCount determines the best number of workers based on swarm size.
func getOptimalWorkerCount(swarmSize int) int {
	// Base it on CPU count and swarm size
	numCPU := runtime.NumCPU()

	switch {
	case swarmSize < 100:
		return numCPU
	case swarmSize < 1000:
		return numCPU * 2
	default:
		// For very large swarms, cap at 4x CPU count
		return minInt(numCPU*4, 32)
	}
}

// minInt returns the minimum of two integers.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
