package emerge

import (
	"math"
	"testing"
	"time"
)

func TestNewConvergenceMonitor(t *testing.T) {
	tests := []struct {
		name            string
		targetCoherence float64
		wantNil         bool
	}{
		{
			name:            "normal target",
			targetCoherence: 0.85,
			wantNil:         false,
		},
		{
			name:            "low target",
			targetCoherence: 0.1,
			wantNil:         false,
		},
		{
			name:            "high target",
			targetCoherence: 0.99,
			wantNil:         false,
		},
		{
			name:            "zero target",
			targetCoherence: 0.0,
			wantNil:         false,
		},
		{
			name:            "negative target",
			targetCoherence: -0.5,
			wantNil:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := NewConvergenceMonitor(tt.targetCoherence)

			if (monitor == nil) != tt.wantNil {
				t.Errorf("NewConvergenceMonitor() = nil: %v, want %v", monitor == nil, tt.wantNil)
				return
			}

			if monitor != nil {
				if monitor.targetCoherence != tt.targetCoherence {
					t.Errorf("targetCoherence = %f, want %f", monitor.targetCoherence, tt.targetCoherence)
				}

				if len(monitor.history) != 0 {
					t.Error("History should be empty initially")
				}

				if monitor.convergedAt != nil {
					t.Error("Should not be converged initially")
				}
			}
		})
	}
}

func TestConvergenceMonitorRecord(t *testing.T) {
	tests := []struct {
		name            string
		targetCoherence float64
		recordings      []float64
		wantHistoryLen  int
		wantConverged   bool
	}{
		{
			name:            "basic recording",
			targetCoherence: 0.8,
			recordings:      []float64{0.5, 0.6, 0.7},
			wantHistoryLen:  3,
			wantConverged:   false,
		},
		{
			name:            "achieve convergence",
			targetCoherence: 0.8,
			recordings:      []float64{0.5, 0.6, 0.85, 0.85, 0.85, 0.85, 0.85},
			wantHistoryLen:  7,
			wantConverged:   true,
		},
		{
			name:            "fluctuating values",
			targetCoherence: 0.8,
			recordings:      []float64{0.3, 0.9, 0.2, 0.85, 0.4, 0.82},
			wantHistoryLen:  6,
			wantConverged:   false,
		},
		{
			name:            "immediate convergence",
			targetCoherence: 0.5,
			recordings:      []float64{0.9, 0.9, 0.9, 0.9, 0.9},
			wantHistoryLen:  5,
			wantConverged:   true,
		},
		{
			name:            "single value",
			targetCoherence: 0.8,
			recordings:      []float64{0.9},
			wantHistoryLen:  1,
			wantConverged:   false, // Need 5 stable iterations
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := NewConvergenceMonitor(tt.targetCoherence)

			for _, val := range tt.recordings {
				monitor.Record(val)
			}

			if len(monitor.history) != tt.wantHistoryLen {
				t.Errorf("history length = %d, want %d", len(monitor.history), tt.wantHistoryLen)
			}

			if monitor.IsConverged() != tt.wantConverged {
				t.Errorf("IsConverged() = %v, want %v", monitor.IsConverged(), tt.wantConverged)
			}

			if tt.wantConverged && monitor.convergedAt == nil {
				t.Error("Converged time should be set when converged")
			}
		})
	}
}

func TestConvergenceMonitorRate(t *testing.T) {
	tests := []struct {
		name        string
		recordings  []float64
		wantPositive bool
		description string
	}{
		{
			name:        "increasing coherence",
			recordings:  []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8},
			wantPositive: true,
			description: "linearly increasing",
		},
		{
			name:        "decreasing coherence",
			recordings:  []float64{0.9, 0.8, 0.7, 0.6, 0.5, 0.4, 0.3, 0.2},
			wantPositive: false,
			description: "linearly decreasing",
		},
		{
			name:        "constant coherence",
			recordings:  []float64{0.5, 0.5, 0.5, 0.5, 0.5},
			wantPositive: false, // Rate should be ~0, which is <= 0
			description: "constant values",
		},
		{
			name:        "exponential increase",
			recordings:  []float64{0.1, 0.11, 0.121, 0.1331, 0.14641, 0.161051},
			wantPositive: true,
			description: "exponential growth",
		},
		{
			name:        "oscillating",
			recordings:  []float64{0.3, 0.7, 0.3, 0.7, 0.3, 0.7},
			wantPositive: true, // Linear regression on oscillating can have slight positive trend
			description: "oscillating pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := NewConvergenceMonitor(0.9)

			for _, val := range tt.recordings {
				monitor.Record(val)
				time.Sleep(time.Millisecond) // Ensure timestamps differ
			}

			rate := monitor.Rate()

			if tt.wantPositive && rate <= 0 {
				t.Errorf("Rate() = %f, want positive for %s", rate, tt.description)
			}
			if !tt.wantPositive && rate > 0 {
				t.Errorf("Rate() = %f, want non-positive for %s", rate, tt.description)
			}
		})
	}
}

func TestConvergenceMonitorStatistics(t *testing.T) {
	tests := []struct {
		name         string
		samples      []float64
		wantMean     float64
		wantMin      float64
		wantMax      float64
		wantSamples  float64
		tolerance    float64
	}{
		{
			name:        "basic samples",
			samples:     []float64{0.5, 0.6, 0.7, 0.65, 0.75},
			wantMean:    0.64,
			wantMin:     0.5,
			wantMax:     0.75,
			wantSamples: 5,
			tolerance:   0.01,
		},
		{
			name:        "single sample",
			samples:     []float64{0.42},
			wantMean:    0.42,
			wantMin:     0.42,
			wantMax:     0.42,
			wantSamples: 1,
			tolerance:   0.001,
		},
		{
			name:        "identical samples",
			samples:     []float64{0.8, 0.8, 0.8, 0.8},
			wantMean:    0.8,
			wantMin:     0.8,
			wantMax:     0.8,
			wantSamples: 4,
			tolerance:   0.001,
		},
		{
			name:        "wide range",
			samples:     []float64{0.1, 0.2, 0.3, 0.8, 0.9, 1.0},
			wantMean:    0.55,
			wantMin:     0.1,
			wantMax:     1.0,
			wantSamples: 6,
			tolerance:   0.01,
		},
		{
			name:        "negative and positive",
			samples:     []float64{-0.5, -0.2, 0.0, 0.3, 0.6},
			wantMean:    0.04,
			wantMin:     -0.5,
			wantMax:     0.6,
			wantSamples: 5,
			tolerance:   0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := NewConvergenceMonitor(0.8)

			for _, s := range tt.samples {
				monitor.Record(s)
			}

			stats := monitor.Statistics()

			// Check samples count
			if samples := stats["samples"]; samples != tt.wantSamples {
				t.Errorf("samples = %f, want %f", samples, tt.wantSamples)
			}

			// Check mean
			if mean := stats["mean"]; math.Abs(mean-tt.wantMean) > tt.tolerance {
				t.Errorf("mean = %f, want %f ± %f", mean, tt.wantMean, tt.tolerance)
			}

			// Check min
			if min := stats["min"]; math.Abs(min-tt.wantMin) > tt.tolerance {
				t.Errorf("min = %f, want %f ± %f", min, tt.wantMin, tt.tolerance)
			}

			// Check max
			if max := stats["max"]; math.Abs(max-tt.wantMax) > tt.tolerance {
				t.Errorf("max = %f, want %f ± %f", max, tt.wantMax, tt.tolerance)
			}
		})
	}
}

func TestConvergenceMonitorReset(t *testing.T) {
	tests := []struct {
		name              string
		targetCoherence   float64
		preResetSamples   []float64
		postResetSamples  []float64
		checkConvergence  bool
	}{
		{
			name:             "reset after convergence",
			targetCoherence:  0.8,
			preResetSamples:  []float64{0.85, 0.85, 0.85, 0.85, 0.85},
			postResetSamples: []float64{0.5, 0.6},
			checkConvergence: true,
		},
		{
			name:             "reset without convergence",
			targetCoherence:  0.9,
			preResetSamples:  []float64{0.5, 0.6, 0.7},
			postResetSamples: []float64{0.95, 0.95},
			checkConvergence: false,
		},
		{
			name:             "reset empty monitor",
			targetCoherence:  0.8,
			preResetSamples:  []float64{},
			postResetSamples: []float64{0.8},
			checkConvergence: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := NewConvergenceMonitor(tt.targetCoherence)

			// Add pre-reset samples
			for _, val := range tt.preResetSamples {
				monitor.Record(val)
			}

			preResetConverged := monitor.IsConverged()
			if tt.checkConvergence && len(tt.preResetSamples) >= 5 {
				allAboveThreshold := true
				for _, v := range tt.preResetSamples {
					if v < tt.targetCoherence {
						allAboveThreshold = false
						break
					}
				}
				if allAboveThreshold && !preResetConverged {
					t.Error("Should be converged before reset with all values above threshold")
				}
			}

			// Reset
			monitor.Reset()

			// Check reset state
			if monitor.IsConverged() {
				t.Error("Should not be converged after reset")
			}

			if len(monitor.history) != 0 {
				t.Error("History should be empty after reset")
			}

			if monitor.stableIterations != 0 {
				t.Error("Stable iterations should be reset")
			}

			if monitor.convergedAt != nil {
				t.Error("Converged time should be nil after reset")
			}

			// Add post-reset samples
			for _, val := range tt.postResetSamples {
				monitor.Record(val)
			}

			if len(monitor.history) != len(tt.postResetSamples) {
				t.Errorf("Post-reset history length = %d, want %d", 
					len(monitor.history), len(tt.postResetSamples))
			}
		})
	}
}

func TestConvergenceMonitorConcurrency(t *testing.T) {
	monitor := NewConvergenceMonitor(0.8)
	
	// Test concurrent recording
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(val float64) {
			monitor.Record(val)
			done <- true
		}(float64(i) / 10.0)
	}
	
	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Should have recorded all values
	if len(monitor.history) != 10 {
		t.Errorf("Expected 10 recordings, got %d", len(monitor.history))
	}
	
	// Test concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			_ = monitor.IsConverged()
			_ = monitor.Rate()
			_ = monitor.Statistics()
			done <- true
		}()
	}
	
	// Wait for all reads
	for i := 0; i < 10; i++ {
		<-done
	}
}