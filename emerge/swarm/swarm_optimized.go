package swarm

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"time"

	"github.com/carlisia/bio-adapt/emerge/agent"
	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/internal/config"
	"github.com/carlisia/bio-adapt/internal/random"
)

// OptimizedSwarm is a performance-optimized version for 1000+ agents.
// It uses slice-based storage and batch operations to reduce allocations.
type OptimizedSwarm struct {
	// Slice-based storage for better cache locality and faster iteration
	agents      []*agent.Agent // Direct access by index
	agentIndex  map[string]int // ID to index mapping
	agentsMutex sync.RWMutex   // Protects agents slice

	// Configuration
	goalState core.State
	config    config.Swarm
	size      int

	// Optimization: pre-allocated buffers for batch operations
	updateBuffer []agentUpdate // Reusable buffer for updates
	workerPool   *WorkerPool   // Goroutine pool for concurrent updates

	// Track initialization
	connectionsEstablished bool

	// Goal-directed synchronization
	goalDirectedSync *GoalDirectedSync
}

// agentUpdate holds pending updates for batch processing
type agentUpdate struct {
	index int
	phase float64
	freq  time.Duration
}

// WorkerPool manages a fixed number of goroutines for agent updates
type WorkerPool struct {
	workers   int
	workQueue chan func()
	quit      chan struct{}
}

// NewOptimized creates an optimized swarm for high agent counts.
func NewOptimized(size int, goalState core.State) (*OptimizedSwarm, error) {
	if size <= 0 {
		return nil, fmt.Errorf("%w: got %d", ErrInvalidSwarmSize, size)
	}

	// Validate goal state using centralized validation
	if err := goalState.Validate(); err != nil {
		return nil, fmt.Errorf("invalid goal state: %w", err)
	}

	// Pre-allocate all memory upfront
	s := &OptimizedSwarm{
		agents:       make([]*agent.Agent, 0, size),
		agentIndex:   make(map[string]int, size),
		goalState:    goalState,
		config:       config.AutoScaleConfig(size),
		size:         size,
		updateBuffer: make([]agentUpdate, 0, size),
	}

	// Initialize worker pool with reasonable goroutine count
	// For 1000+ agents, use NumCPU * 2 workers
	numWorkers := getOptimalWorkerCount(size)
	s.workerPool = NewWorkerPool(numWorkers)

	// Create agents efficiently
	s.createAgents()

	// Establish connections
	s.establishConnections()

	// Initialize goal-directed sync
	s.goalDirectedSync = NewGoalDirectedSync(&Swarm{
		goalState: goalState,
		config:    s.config,
		size:      size,
	})

	return s, nil
}

// createAgents efficiently creates all agents with pre-allocation.
func (s *OptimizedSwarm) createAgents() {
	// Pre-allocate the slice
	s.agents = make([]*agent.Agent, s.size)

	// Batch create agents
	batchSize := 100
	for i := 0; i < s.size; i += batchSize {
		end := minInt(i+batchSize, s.size)

		// Create batch of agents
		for j := i; j < end; j++ {
			id := fmt.Sprintf("agent-%d", j)
			a := agent.New(id)

			// Set initial state
			a.SetPhase(random.Phase())
			a.SetFrequency(100 * time.Millisecond)
			a.SetEnergy(100)

			s.agents[j] = a
			s.agentIndex[id] = j
		}
	}
}

// establishConnections creates the network topology efficiently.
func (s *OptimizedSwarm) establishConnections() {
	if s.connectionsEstablished {
		return
	}

	// For large swarms, use small-world topology for efficiency
	if s.size > 100 {
		s.establishSmallWorldConnections()
	} else {
		s.establishMeshConnections()
	}

	s.connectionsEstablished = true
}

// establishSmallWorldConnections creates efficient small-world topology.
func (s *OptimizedSwarm) establishSmallWorldConnections() {
	// Each agent connects to k nearest neighbors + random long-range connections
	k := minInt(6, s.size/10) // Local connections
	p := 0.1                  // Rewiring probability

	for i, a := range s.agents {
		// Connect to k/2 neighbors on each side (ring topology)
		for j := 1; j <= k/2; j++ {
			// Right neighbor
			rightIdx := (i + j) % s.size
			a.Neighbors().Store(s.agents[rightIdx].ID, s.agents[rightIdx])

			// Left neighbor
			leftIdx := (i - j + s.size) % s.size
			a.Neighbors().Store(s.agents[leftIdx].ID, s.agents[leftIdx])
		}

		// Rewire some connections randomly (small-world property)
		if random.Float64() < p {
			// Add a random long-range connection
			randomIdx := random.Intn(s.size)
			if randomIdx != i {
				a.Neighbors().Store(s.agents[randomIdx].ID, s.agents[randomIdx])
			}
		}
	}
}

// establishMeshConnections creates full mesh for small swarms.
func (s *OptimizedSwarm) establishMeshConnections() {
	for i, a := range s.agents {
		for j, neighbor := range s.agents {
			if i != j {
				a.Neighbors().Store(neighbor.ID, neighbor)
			}
		}
	}
}

// GetAgents returns all agents (efficient slice access).
func (s *OptimizedSwarm) GetAgents() []*agent.Agent {
	s.agentsMutex.RLock()
	defer s.agentsMutex.RUnlock()

	// Return copy of slice to prevent external modifications
	result := make([]*agent.Agent, len(s.agents))
	copy(result, s.agents)
	return result
}

// GetAgent returns a specific agent by ID.
func (s *OptimizedSwarm) GetAgent(id string) (*agent.Agent, bool) {
	s.agentsMutex.RLock()
	defer s.agentsMutex.RUnlock()

	if idx, ok := s.agentIndex[id]; ok {
		return s.agents[idx], true
	}
	return nil, false
}

// BatchUpdatePhases updates multiple agent phases efficiently.
func (s *OptimizedSwarm) BatchUpdatePhases(updates []agentUpdate) {
	s.agentsMutex.Lock()
	defer s.agentsMutex.Unlock()

	// Apply all updates in a single pass
	for _, update := range updates {
		if update.index >= 0 && update.index < len(s.agents) {
			s.agents[update.index].SetPhase(update.phase)
			if update.freq > 0 {
				s.agents[update.index].SetFrequency(update.freq)
			}
		}
	}
}

// MeasureCoherence calculates coherence efficiently.
func (s *OptimizedSwarm) MeasureCoherence() float64 {
	s.agentsMutex.RLock()
	defer s.agentsMutex.RUnlock()

	if len(s.agents) == 0 {
		return 0
	}

	// Use Kuramoto order parameter
	var sumCos, sumSin float64

	// Process in batches to improve cache locality
	batchSize := 64 // Typical cache line size
	for i := 0; i < len(s.agents); i += batchSize {
		end := min(i+batchSize, len(s.agents))
		for j := i; j < end; j++ {
			phase := s.agents[j].Phase()
			sumCos += math.Cos(phase)
			sumSin += math.Sin(phase)
		}
	}

	n := float64(len(s.agents))
	r := math.Sqrt(sumCos*sumCos+sumSin*sumSin) / n
	return r
}

// UpdateConcurrent performs concurrent updates using worker pool.
func (s *OptimizedSwarm) UpdateConcurrent(updateFunc func(*agent.Agent)) {
	s.agentsMutex.RLock()
	agents := s.agents
	s.agentsMutex.RUnlock()

	// Submit work to pool in batches
	batchSize := len(agents) / s.workerPool.workers
	if batchSize < 10 {
		batchSize = 10
	}

	var wg sync.WaitGroup
	for i := 0; i < len(agents); i += batchSize {
		end := min(i+batchSize, len(agents))
		batch := agents[i:end]

		wg.Add(1)
		s.workerPool.Submit(func() {
			defer wg.Done()
			for _, a := range batch {
				updateFunc(a)
			}
		})
	}

	wg.Wait()
}

// Cleanup releases resources.
func (s *OptimizedSwarm) Cleanup() {
	if s.workerPool != nil {
		s.workerPool.Stop()
	}
}

// NewWorkerPool creates a new worker pool.
func NewWorkerPool(workers int) *WorkerPool {
	wp := &WorkerPool{
		workers:   workers,
		workQueue: make(chan func(), workers*2),
		quit:      make(chan struct{}),
	}

	// Start workers
	for i := 0; i < workers; i++ { //nolint:intrange // i is not used in the loop body
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

// getOptimalWorkerCount determines the best number of workers.
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
