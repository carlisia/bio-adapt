// simulation.go implements the Simulator interface for the API minimization demo.
// This handles the core simulation logic, agent coordination, and statistics.
//
// CONCEPTUAL OVERVIEW:
// This simulation demonstrates how emergent synchronization can optimize API costs.
//
// Components:
// 1. Emerge swarm (from framework) - Provides phase synchronization between agents
// 2. Workloads (this package) - Simulate different types of work making API calls
// 3. Batch manager - Collects synchronized API calls into efficient batches
// 4. Metrics collector - Tracks coherence improvements from batching
//
// How it works:
// - Without synchronization: Each workload makes API calls independently (expensive)
// - With synchronization: Workloads coordinate timing via emerge, enabling batch API calls (efficient)
// - The emerge framework handles the physics of synchronization
// - This simulation adds the application domain (workloads, API calls, batching)

package simulation

import (
	"context"
	"errors"
	"sync"
	"time"

	emerge "github.com/carlisia/bio-adapt/client/emerge"
	"github.com/carlisia/bio-adapt/emerge/goal"
	"github.com/carlisia/bio-adapt/simulations/emerge/simulation/pattern"
)

// Simulation implements the Simulator interface
type Simulation struct {
	config       Config
	pattern      pattern.Type
	goalType     goal.Type // Goal type for workload selection
	workloads    []*Workload
	emergeClient *emerge.Client
	batch        *BatchManager
	metrics      *MetricsCollector

	mu              sync.RWMutex
	running         bool
	paused          bool
	pausedCoherence float64            // Cached coherence when paused
	pausedPhases    map[string]float64 // Cached agent phases when paused
	disrupted       bool
	disruptTime     time.Time
	reset           bool
	resetTime       time.Time
	startTime       time.Time
	parentCtx       context.Context // Parent context from Start
	swarmCtx        context.Context
	swarmCancel     context.CancelFunc
}

// Start begins the simulation
func (s *Simulation) Start(ctx context.Context) error {
	s.mu.Lock()
	s.running = true
	s.startTime = time.Now()
	s.parentCtx = ctx
	// Create swarm context that we can cancel on pause
	s.swarmCtx, s.swarmCancel = context.WithCancel(ctx)
	s.mu.Unlock()

	// Start swarm synchronization
	go func() {
		if err := s.emergeClient.Start(s.swarmCtx); err != nil && !errors.Is(err, context.Canceled) {
			// Log error but don't crash
			_ = err
		}
	}()

	// Start agents
	for _, workload := range s.workloads {
		workload.Start(ctx, s.batch)
	}

	// Start batch processing
	go s.batch.ProcessBatches(ctx, s.metrics)

	// Wait for context
	<-ctx.Done()

	s.mu.Lock()
	s.running = false
	s.mu.Unlock()

	return nil
}

// Reset resets the simulation
func (s *Simulation) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Reset metrics
	s.metrics.Reset()

	// Reset batch manager
	s.batch.Reset()

	// Reset workload counters
	for _, workload := range s.workloads {
		workload.Reset()
	}

	// Reset emerge agents' phases to WORST possible state for each goal
	// This creates maximum dramatic effect for demos - showing recovery from chaos
	agents := s.emergeClient.Swarm().Agents()
	agentCount := len(agents)
	i := 0

	switch s.goalType {
	case goal.MinimizeAPICalls:
		// WORST: Maximally spread out (anti-phase) when we want sync
		// This prevents any batching - every agent sends alone
		phaseStep := (2 * 3.14159) / float64(agentCount)
		for _, agent := range agents {
			agent.SetPhase(float64(i) * phaseStep)
			i++
		}

	case goal.DistributeLoad:
		// WORST: All synchronized at 0 when we want anti-phase
		// This causes maximum resource contention
		for _, agent := range agents {
			agent.SetPhase(0)
		}

	case goal.ReachConsensus:
		// WORST: Every agent at a unique phase (no agreement possible)
		// Maximum disagreement - no voting blocs
		phaseStep := (2 * 3.14159) / float64(agentCount)
		for _, agent := range agents {
			agent.SetPhase(float64(i) * phaseStep)
			i++
		}

	case goal.MinimizeLatency:
		// WORST: Alternating phases causing unpredictable delays
		for _, agent := range agents {
			// Alternating between 0 and π for maximum jitter
			if i%2 == 0 {
				agent.SetPhase(0)
			} else {
				agent.SetPhase(3.14159)
			}
			i++
		}

	case goal.SaveEnergy:
		// WORST: All agents active at once (draining battery)
		for _, agent := range agents {
			agent.SetPhase(0)
		}

	case goal.MaintainRhythm:
		// WORST: Irregular drifting phases (no steady beat)
		// Jobs at random phases = execution times all over the place
		for _, agent := range agents {
			// Irregular pattern simulating schedule drift
			phase := float64(i) * 1.7 // Non-divisible pattern
			agent.SetPhase(phase)
			i++
		}

	case goal.RecoverFromFailure:
		// WORST: Clustered phases (cascading failures)
		// Half at 0, half at π = failure waves instead of isolated incidents
		for _, agent := range agents {
			if i < agentCount/2 {
				agent.SetPhase(0)
			} else {
				agent.SetPhase(3.14159)
			}
			i++
		}

	case goal.AdaptToTraffic:
		// WORST: All synchronized (traffic spikes overwhelm system)
		// Everyone scales at once = resource contention during bursts
		for _, agent := range agents {
			agent.SetPhase(0)
		}

	default:
		// Default worst: Evenly distributed (hard to sync)
		phaseStep := (2 * 3.14159) / float64(agentCount)
		for _, agent := range agents {
			agent.SetPhase(float64(i) * phaseStep)
			i++
		}
	}

	// Reset start time
	s.startTime = time.Now()

	// Mark as reset for visual feedback
	s.reset = true
	s.resetTime = time.Now()

	// Clear reset flag after delay
	go func() {
		time.Sleep(2 * time.Second)
		s.mu.Lock()
		s.reset = false
		s.mu.Unlock()
	}()
}

// Disrupt introduces disruption
func (s *Simulation) Disrupt() {
	s.mu.Lock()
	// Mark as disrupted for visual feedback
	s.disrupted = true
	s.disruptTime = time.Now()
	s.mu.Unlock()

	// Disrupt synchronization by changing some agents' phases
	agents := s.emergeClient.Swarm().Agents()
	disrupted := 0
	maxDisrupt := len(agents) / 5 // 20% of agents

	for _, a := range agents {
		if disrupted >= maxDisrupt {
			break
		}
		currentPhase := a.Phase()
		a.SetPhase(currentPhase + 3.14159) // Add π
		disrupted++
	}

	// Monitor recovery - clear disruption flag when coherence recovers above target
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		// Safety timeout after 30 seconds
		timeout := time.After(30 * time.Second)

		for {
			select {
			case <-ticker.C:
				s.mu.RLock()
				if !s.disrupted {
					// Already cleared (maybe by reset or another disruption)
					s.mu.RUnlock()
					return
				}
				target := s.config.TargetCoherence
				s.mu.RUnlock()

				coherence := s.emergeClient.Swarm().MeasureCoherence()
				if coherence >= target {
					// Recovered above target
					s.mu.Lock()
					s.disrupted = false
					s.mu.Unlock()
					return
				}
			case <-timeout:
				// Safety timeout - clear disruption flag after 30 seconds
				s.mu.Lock()
				s.disrupted = false
				s.mu.Unlock()
				return
			}
		}
	}()
}

// Pause pauses the simulation
func (s *Simulation) Pause() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.paused = true
	// Cache the current coherence and phases so everything freezes
	s.pausedCoherence = s.emergeClient.Swarm().MeasureCoherence()
	s.pausedPhases = make(map[string]float64)
	for id, agent := range s.emergeClient.Swarm().Agents() {
		s.pausedPhases[id] = agent.Phase()
	}
	// Cancel swarm context to stop oscillations
	if s.swarmCancel != nil {
		s.swarmCancel()
	}
}

// Resume resumes the simulation
func (s *Simulation) Resume() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.paused = false
	// Restart swarm with a new context
	if s.parentCtx != nil {
		s.swarmCtx, s.swarmCancel = context.WithCancel(s.parentCtx)
		go func() {
			if err := s.emergeClient.Start(s.swarmCtx); err != nil && !errors.Is(err, context.Canceled) {
				// Log error but don't crash
				_ = err
			}
		}()
	}
}

// Snapshot returns current state
func (s *Simulation) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get agent snapshots and calculate total pending tasks
	agentSnapshots := make([]AgentSnapshot, len(s.workloads))
	totalPendingTasks := 0
	for i, agent := range s.workloads {
		snapshot := agent.Snapshot()
		// Use cached phase if paused
		if s.paused && s.pausedPhases != nil {
			if phase, ok := s.pausedPhases[snapshot.ID]; ok {
				snapshot.Phase = phase
			}
		}
		agentSnapshots[i] = snapshot
		totalPendingTasks += snapshot.PendingTasks
	}

	// Get metrics
	metrics := s.metrics.Current()

	// Calculate coherence (use cached value if paused)
	var coherence float64
	if s.paused {
		coherence = s.pausedCoherence
	} else {
		coherence = s.emergeClient.Swarm().MeasureCoherence()
	}

	// Get current batch size from queue depth
	currentBatchSize := len(s.batch.queue)

	// Get last batch info for pulse animation
	lastBatchTime, lastBatchSize := s.batch.LastBatchInfo()
	batchJustSent := time.Since(lastBatchTime) < 500*time.Millisecond // Show pulse for 500ms

	return Snapshot{
		Timestamp:        time.Now(),
		ElapsedTime:      time.Since(s.startTime),
		Agents:           agentSnapshots,
		Coherence:        coherence,
		TargetCoherence:  s.config.TargetCoherence,
		PendingTasks:     totalPendingTasks,
		CurrentBatchSize: currentBatchSize,
		BatchesProcessed: metrics.TotalBatches,
		CostWithoutSync:  metrics.CostWithoutSync,
		CostWithSync:     metrics.CostWithSync,
		Savings:          metrics.TotalSavings,
		SavingsPercent:   metrics.SavingsPercent,
		Paused:           s.paused,
		Disrupted:        s.disrupted,
		Reset:            s.reset,
		BatchJustSent:    batchJustSent,
		LastBatchTime:    lastBatchTime,
		LastBatchSize:    lastBatchSize,
	}
}

// Statistics returns final statistics
func (s *Simulation) Statistics() Statistics {
	metrics := s.metrics.Final()

	return Statistics{
		TotalAPICalls:    metrics.TotalAPICalls,
		TotalBatches:     metrics.TotalBatches,
		AverageBatchSize: metrics.AverageBatchSize,
		CostWithoutSync:  metrics.CostWithoutSync,
		CostWithSync:     metrics.CostWithSync,
		TotalSavings:     metrics.TotalSavings,
		SavingsPercent:   metrics.SavingsPercent,
		FinalCoherence:   s.emergeClient.Swarm().MeasureCoherence(),
		PeakCoherence:    metrics.PeakCoherence,
		TimeToConverge:   metrics.TimeToConverge,
	}
}
