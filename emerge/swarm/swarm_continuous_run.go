package swarm

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/carlisia/bio-adapt/emerge/core"
)

// RecoveryConfig defines thresholds for disruption detection and recovery.
// These thresholds determine when the swarm needs to resynchronize and what
// level of coherence is considered acceptable.
type RecoveryConfig struct {
	// MinimumViableCoherence is the absolute floor below which the system
	// is considered non-functional (e.g., 0.3 for 30% minimum coherence).
	// Any coherence below this triggers immediate resynchronization.
	MinimumViableCoherence float64

	// TargetMarginRatio defines how close to target we need to be to consider
	// the system "synchronized" (e.g., 0.95 means within 95% of target).
	// If current coherence < target * TargetMarginRatio, resync is needed.
	TargetMarginRatio float64

	// SmallDropRatio is the fraction of current coherence that constitutes
	// a small disruption (e.g., 0.05 means 5% drop from peak).
	SmallDropRatio float64

	// LargeDropRatio is the fraction that constitutes a major disruption
	// (e.g., 0.15 means 15% drop from peak triggers immediate resync).
	LargeDropRatio float64

	// StuckThreshold is the number of consecutive checks without improvement
	// before forcing a resynchronization attempt.
	StuckThreshold int

	// CheckInterval is how often to check coherence and decide on resync.
	CheckInterval time.Duration
}

// DefaultRecoveryConfig returns a RecoveryConfig with sensible defaults
// based on the target coherence level. Higher targets get tighter thresholds.
func DefaultRecoveryConfig(targetCoherence float64) RecoveryConfig {
	cfg := RecoveryConfig{
		CheckInterval:  DefaultCheckInterval,
		StuckThreshold: DefaultStuckThreshold,
	}

	// Scale thresholds based on target coherence
	switch {
	case targetCoherence >= HighCoherenceThreshold:
		// High coherence: tight control
		cfg.MinimumViableCoherence = MinViableCoherenceHigh
		cfg.TargetMarginRatio = TargetMarginRatioHigh
		cfg.SmallDropRatio = SmallDropRatioHigh
		cfg.LargeDropRatio = LargeDropRatioHigh
	case targetCoherence >= MediumCoherenceThreshold:
		// Medium coherence: balanced control
		cfg.MinimumViableCoherence = MinViableCoherenceMedium
		cfg.TargetMarginRatio = TargetMarginRatioMedium
		cfg.SmallDropRatio = SmallDropRatioMedium
		cfg.LargeDropRatio = LargeDropRatioMedium
	case targetCoherence >= LowCoherenceThreshold:
		// Low-medium coherence: looser control
		cfg.MinimumViableCoherence = MinViableCoherenceLow
		cfg.TargetMarginRatio = TargetMarginRatioLow
		cfg.SmallDropRatio = SmallDropRatioLow
		cfg.LargeDropRatio = LargeDropRatioLow
	default:
		// Low coherence: very loose control
		cfg.MinimumViableCoherence = MinViableCoherenceVeryLow
		cfg.TargetMarginRatio = TargetMarginRatioVeryLow
		cfg.SmallDropRatio = SmallDropRatioVeryLow
		cfg.LargeDropRatio = LargeDropRatioVeryLow
	}

	return cfg
}

// monitorState tracks the state of continuous monitoring
type monitorState struct {
	lastCoherence float64
	peakCoherence float64 // Best coherence achieved recently
	stableCount   int     // Consecutive measurements without improvement
	syncActive    bool    // Is synchronization currently running
	lastSyncTime  time.Time
}

// needsResync determines if synchronization should be restarted based on
// the recovery configuration thresholds.
func (s *Swarm) needsResync(state *monitorState, currentCoherence float64) bool {
	cfg := s.recoveryConfig
	target := s.goalState.Coherence

	// Condition 1: Below minimum viable coherence (system non-functional)
	if currentCoherence < cfg.MinimumViableCoherence {
		return true
	}

	// Condition 2: Below acceptable margin of target
	acceptableMin := target * cfg.TargetMarginRatio
	if currentCoherence < acceptableMin {
		return true
	}

	// Condition 3: Significant drop from peak (disruption detected)
	dropFromPeak := state.peakCoherence - currentCoherence
	if dropFromPeak > state.peakCoherence*cfg.SmallDropRatio {
		// The larger the drop, the more likely it's a disruption
		if dropFromPeak > state.peakCoherence*cfg.LargeDropRatio {
			return true // Definite disruption
		} else if currentCoherence < target {
			// Small drop - only sync if we're below target
			return true
		}
	}

	// Condition 4: Stuck at suboptimal level
	improvement := currentCoherence - state.lastCoherence
	if math.Abs(improvement) < ImprovementThreshold {
		state.stableCount++
		if state.stableCount > cfg.StuckThreshold && currentCoherence < acceptableMin {
			state.stableCount = 0
			return true
		}
	} else if improvement > ImprovementThreshold {
		// Reset stable count on improvement
		state.stableCount = 0
	}

	// Condition 5: Rapid degradation (even if small)
	if currentCoherence < state.lastCoherence-state.lastCoherence*cfg.SmallDropRatio {
		return true
	}

	return false
}

// RunContinuous starts the swarm and continuously maintains synchronization,
// recovering from disruptions automatically. Unlike Run(), this method doesn't
// exit after achieving synchronization but continues monitoring and recovering.
//
// This method is designed to handle:
//   - Any size disruption (from single agents to total system failure)
//   - Gradual degradation over time
//   - Network partitions and rejoins
//   - Dynamic system changes
//
// Recovery aggressiveness scales with disruption severity:
//   - Small drops (3-8%): Monitored, resync if below target
//   - Large drops (8-20%): Immediate resynchronization
//   - Massive drops (>20%): Emergency resynchronization
//   - System failure (below minimum viable): Continuous recovery attempts
//
// The recovery behavior is controlled by RecoveryConfig which can be customized
// via WithRecoveryConfig() option. By default, it uses thresholds based on the
// target coherence level (see DefaultRecoveryConfig).
//
// This method only exits when the context is canceled. For one-time
// synchronization without continuous monitoring, use Run() instead.
func (s *Swarm) RunContinuous(ctx context.Context) error {
	// Initialize recovery config if not already set
	if s.recoveryConfig.CheckInterval == 0 {
		s.recoveryConfig = DefaultRecoveryConfig(s.goalState.Coherence)
	}

	// Create target pattern from goal state
	targetPattern := &core.TargetPattern{
		Phase:     s.goalState.Phase,
		Frequency: s.goalState.Frequency,
		Coherence: s.goalState.Coherence,
		Amplitude: 1.0,
		Stability: 0.9,
	}

	// Helper function to start synchronization
	startSync := func(ctx context.Context) (<-chan error, context.CancelFunc) {
		syncCtx, cancel := context.WithCancel(ctx)
		done := make(chan error, 1)

		// Reset convergence monitors for fresh start
		if s.convergence != nil {
			s.convergence.Reset()
		}
		if s.goalDirectedSync != nil && s.goalDirectedSync.convergenceMonitor != nil {
			s.goalDirectedSync.convergenceMonitor.Reset()
		}

		go func() {
			done <- s.goalDirectedSync.AchieveSynchronization(syncCtx, targetPattern)
		}()

		return done, cancel
	}

	// Start initial synchronization
	syncDone, syncCancel := startSync(ctx)
	defer syncCancel()

	// Monitoring state
	ticker := time.NewTicker(s.recoveryConfig.CheckInterval)
	defer ticker.Stop()

	state := &monitorState{
		lastCoherence: s.MeasureCoherence(),
		peakCoherence: 0.0,
		syncActive:    true,
		lastSyncTime:  time.Now(),
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case err := <-syncDone:
			state.syncActive = false

			if err != nil && !errors.Is(err, context.Canceled) {
				// Real error occurred
				return err
			}
			// Synchronization completed successfully, continue monitoring

		case <-ticker.C:
			currentCoherence := s.MeasureCoherence()

			// Update peak coherence with slow decay
			if currentCoherence > state.peakCoherence {
				state.peakCoherence = currentCoherence
			} else {
				// Slowly forget old peak to adapt to system changes
				state.peakCoherence *= PeakCoherenceDecayRate
			}

			// Determine if synchronization is needed
			shouldSync := s.needsResync(state, currentCoherence)

			// Start synchronization if needed and not already running
			if shouldSync && !state.syncActive {
				// Avoid too frequent restarts
				if time.Since(state.lastSyncTime) > MinResyncInterval {
					syncCancel() // Cancel any lingering sync
					syncDone, syncCancel = startSync(ctx)
					state.syncActive = true
					state.lastSyncTime = time.Now()
					state.peakCoherence = currentCoherence // Reset peak after restart
				}
			}

			state.lastCoherence = currentCoherence
		}
	}
}
