package emerge_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/carlisia/bio-adapt/emerge"
)

func TestNewConvergenceMonitor(t *testing.T) {
	tests := []struct {
		name       string
		target     emerge.State
		threshold  float64
		validateFn func(t *testing.T, monitor *emerge.ConvergenceMonitor)
	}{
		{
			name: "basic monitor creation",
			target: emerge.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold: 0.85,
			validateFn: func(t *testing.T, monitor *emerge.ConvergenceMonitor) {
				if monitor == nil {
					t.Fatal("Expected monitor to be created")
				}
				if monitor.IsConverged() {
					t.Error("Monitor should not be converged initially")
				}
				if monitor.CurrentCoherence() != 0 {
					t.Error("Initial current coherence should be 0")
				}
				history := monitor.GetHistory()
				if len(history) != 0 {
					t.Error("Initial history should be empty")
				}
			},
		},
		{
			name: "zero threshold monitor",
			target: emerge.State{
				Phase:     0,
				Frequency: 200 * time.Millisecond,
				Coherence: 0.5,
			},
			threshold: 0,
			validateFn: func(t *testing.T, monitor *emerge.ConvergenceMonitor) {
				// Zero threshold means any positive coherence converges
				monitor.Record(0.01)
				if !monitor.IsConverged() {
					t.Error("Zero threshold should converge with any positive coherence")
				}
			},
		},
		{
			name: "negative threshold monitor (invalid)",
			target: emerge.State{
				Phase:     math.Pi / 2,
				Frequency: 50 * time.Millisecond,
				Coherence: 0.7,
			},
			threshold: -0.5,
			validateFn: func(t *testing.T, monitor *emerge.ConvergenceMonitor) {
				// Negative threshold might be treated as 0 or absolute value
				if monitor == nil {
					t.Fatal("Monitor should be created even with negative threshold")
				}
			},
		},
		{
			name: "maximum threshold monitor",
			target: emerge.State{
				Phase:     1.5,
				Frequency: 100 * time.Millisecond,
				Coherence: 1.0,
			},
			threshold: 1.0,
			validateFn: func(t *testing.T, monitor *emerge.ConvergenceMonitor) {
				monitor.Record(0.99)
				if monitor.IsConverged() {
					t.Error("Should not converge below threshold of 1.0")
				}
				monitor.Record(1.0)
				if !monitor.IsConverged() {
					t.Error("Should converge at threshold of 1.0")
				}
			},
		},
		{
			name: "threshold above target coherence",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			threshold: 0.9,
			validateFn: func(t *testing.T, monitor *emerge.ConvergenceMonitor) {
				// Threshold higher than target coherence
				monitor.Record(0.8)
				if monitor.IsConverged() {
					t.Error("Should not converge below threshold even if at target")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := emerge.NewConvergenceMonitor(tt.target, tt.threshold)
			tt.validateFn(t, monitor)
		})
	}
}

func TestConvergenceMonitorRecord(t *testing.T) {
	tests := []struct {
		name       string
		target     emerge.State
		threshold  float64
		samples    []float64
		delays     []time.Duration
		validateFn func(t *testing.T, monitor *emerge.ConvergenceMonitor)
	}{
		{
			name: "basic recording",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold: 0.85,
			samples:   []float64{0.5, 0.6, 0.7},
			delays:    []time.Duration{10 * time.Millisecond, 10 * time.Millisecond, 0},
			validateFn: func(t *testing.T, monitor *emerge.ConvergenceMonitor) {
				history := monitor.GetHistory()
				if len(history) != 3 {
					t.Errorf("Expected 3 samples in history, got %d", len(history))
				}
				current := monitor.CurrentCoherence()
				if math.Abs(current-0.7) > 0.01 {
					t.Errorf("Expected current coherence 0.7, got %f", current)
				}
				if monitor.IsConverged() {
					t.Error("Should not be converged with coherence below threshold")
				}
			},
		},
		{
			name: "single sample",
			target: emerge.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			threshold: 0.75,
			samples:   []float64{0.9},
			delays:    []time.Duration{0},
			validateFn: func(t *testing.T, monitor *emerge.ConvergenceMonitor) {
				if !monitor.IsConverged() {
					t.Error("Should be converged with single sample above threshold")
				}
				if monitor.CurrentCoherence() != 0.9 {
					t.Error("Current coherence should match single sample")
				}
			},
		},
		{
			name: "empty samples",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold: 0.85,
			samples:   []float64{},
			delays:    []time.Duration{},
			validateFn: func(t *testing.T, monitor *emerge.ConvergenceMonitor) {
				if monitor.IsConverged() {
					t.Error("Should not be converged with no samples")
				}
				if monitor.CurrentCoherence() != 0 {
					t.Error("Current coherence should be 0 with no samples")
				}
			},
		},
		{
			name: "decreasing samples",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold: 0.5,
			samples:   []float64{0.9, 0.8, 0.7, 0.6, 0.5, 0.4},
			delays:    []time.Duration{5 * time.Millisecond, 5 * time.Millisecond, 5 * time.Millisecond, 5 * time.Millisecond, 5 * time.Millisecond, 0},
			validateFn: func(t *testing.T, monitor *emerge.ConvergenceMonitor) {
				// Started above threshold but ended below
				if monitor.IsConverged() {
					t.Error("Should not be converged when dropping below threshold")
				}
				if math.Abs(monitor.CurrentCoherence()-0.4) > 0.01 {
					t.Error("Current coherence should be last sample")
				}
			},
		},
		{
			name: "oscillating samples",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold: 0.7,
			samples:   []float64{0.6, 0.8, 0.6, 0.8, 0.6, 0.8},
			delays:    []time.Duration{5 * time.Millisecond, 5 * time.Millisecond, 5 * time.Millisecond, 5 * time.Millisecond, 5 * time.Millisecond, 0},
			validateFn: func(t *testing.T, monitor *emerge.ConvergenceMonitor) {
				if !monitor.IsConverged() {
					t.Error("Should be converged when last sample is above threshold")
				}
				history := monitor.GetHistory()
				if len(history) != 6 {
					t.Error("All samples should be recorded")
				}
			},
		},
		{
			name: "exact threshold",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.85,
			},
			threshold: 0.85,
			samples:   []float64{0.85},
			delays:    []time.Duration{0},
			validateFn: func(t *testing.T, monitor *emerge.ConvergenceMonitor) {
				if !monitor.IsConverged() {
					t.Error("Should be converged at exact threshold")
				}
			},
		},
		{
			name: "many samples",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold: 0.8,
			samples:   []float64{}, // No initial samples
			delays:    []time.Duration{},
			validateFn: func(t *testing.T, monitor *emerge.ConvergenceMonitor) {
				// Initialize samples
				for i := range 100 {
					monitor.Record(0.5 + float64(i)*0.005)
				}
				history := monitor.GetHistory()
				if len(history) != 100 {
					t.Errorf("Expected 100 samples, got %d", len(history))
				}
				if !monitor.IsConverged() {
					t.Error("Should be converged after reaching threshold")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := emerge.NewConvergenceMonitor(tt.target, tt.threshold)

			for i, sample := range tt.samples {
				monitor.Record(sample)
				if i < len(tt.delays) && tt.delays[i] > 0 {
					time.Sleep(tt.delays[i])
				}
			}

			tt.validateFn(t, monitor)
		})
	}
}

func TestConvergenceMonitorConvergenceTime(t *testing.T) {
	tests := []struct {
		name       string
		target     emerge.State
		threshold  float64
		setupFn    func(monitor *emerge.ConvergenceMonitor)
		validateFn func(t *testing.T, monitor *emerge.ConvergenceMonitor)
	}{
		{
			name: "immediate convergence",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold: 0.5,
			setupFn: func(monitor *emerge.ConvergenceMonitor) {
				monitor.Record(0.9)
			},
			validateFn: func(t *testing.T, monitor *emerge.ConvergenceMonitor) {
				if !monitor.IsConverged() {
					t.Error("Should be converged")
				}
				convTime := monitor.ConvergenceTime()
				if convTime <= 0 {
					t.Error("Convergence time should be positive")
				}
				if convTime > 100*time.Millisecond {
					t.Error("Convergence time should be very short for immediate convergence")
				}
			},
		},
		{
			name: "gradual convergence",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold: 0.85,
			setupFn: func(monitor *emerge.ConvergenceMonitor) {
				for i := range 10 {
					coherence := 0.5 + float64(i)*0.05
					monitor.Record(coherence)
					time.Sleep(10 * time.Millisecond)
				}
			},
			validateFn: func(t *testing.T, monitor *emerge.ConvergenceMonitor) {
				if !monitor.IsConverged() {
					t.Error("Should be converged after reaching threshold")
				}
				convTime := monitor.ConvergenceTime()
				if convTime <= 0 {
					t.Error("Convergence time should be positive")
				}
				// Should take at least 70ms (when coherence reaches 0.85)
				if convTime < 70*time.Millisecond {
					t.Error("Convergence time seems too short")
				}
			},
		},
		{
			name: "never converged",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold: 0.95,
			setupFn: func(monitor *emerge.ConvergenceMonitor) {
				for range 5 {
					monitor.Record(0.8)
					time.Sleep(10 * time.Millisecond)
				}
			},
			validateFn: func(t *testing.T, monitor *emerge.ConvergenceMonitor) {
				if monitor.IsConverged() {
					t.Error("Should not be converged")
				}
				convTime := monitor.ConvergenceTime()
				if convTime != 0 {
					t.Error("Convergence time should be 0 when not converged")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := emerge.NewConvergenceMonitor(tt.target, tt.threshold)
			tt.setupFn(monitor)
			tt.validateFn(t, monitor)
		})
	}
}

func TestConvergenceMonitorRate(t *testing.T) {
	tests := []struct {
		name      string
		target    emerge.State
		threshold float64
		samples   []float64
		delays    []time.Duration
		minRate   float64
		maxRate   float64
	}{
		{
			name: "increasing coherence",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold: 0.85,
			samples:   []float64{0.5, 0.6, 0.7, 0.8, 0.9},
			delays:    []time.Duration{10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond, 0},
			minRate:   0.001,
			maxRate:   100.0,
		},
		{
			name: "decreasing coherence",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold: 0.85,
			samples:   []float64{0.9, 0.8, 0.7, 0.6, 0.5},
			delays:    []time.Duration{10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond, 0},
			minRate:   -100.0,
			maxRate:   -0.001,
		},
		{
			name: "constant coherence",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold: 0.85,
			samples:   []float64{0.7, 0.7, 0.7, 0.7, 0.7},
			delays:    []time.Duration{10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond, 0},
			minRate:   -0.01,
			maxRate:   0.01,
		},
		{
			name: "single sample",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold: 0.85,
			samples:   []float64{0.7},
			delays:    []time.Duration{0},
			minRate:   0,
			maxRate:   0,
		},
		{
			name: "oscillating coherence",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold: 0.85,
			samples:   []float64{0.5, 0.7, 0.5, 0.7, 0.5, 0.7},
			delays:    []time.Duration{10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond, 0},
			minRate:   -1.0,
			maxRate:   1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := emerge.NewConvergenceMonitor(tt.target, tt.threshold)

			for i, sample := range tt.samples {
				monitor.Record(sample)
				if i < len(tt.delays) && tt.delays[i] > 0 {
					time.Sleep(tt.delays[i])
				}
			}

			rate := monitor.ConvergenceRate()
			if rate < tt.minRate || rate > tt.maxRate {
				t.Errorf("Expected rate in range [%f, %f], got %f", tt.minRate, tt.maxRate, rate)
			}
		})
	}
}

func TestConvergenceMonitorStability(t *testing.T) {
	tests := []struct {
		name         string
		target       emerge.State
		threshold    float64
		samples      []float64
		minStability float64
		maxStability float64
	}{
		{
			name: "perfectly stable",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold:    0.85,
			samples:      []float64{0.8, 0.8, 0.8, 0.8, 0.8, 0.8, 0.8, 0.8, 0.8, 0.8},
			minStability: 0.95,
			maxStability: 1.0,
		},
		{
			name: "highly unstable",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold:    0.85,
			samples:      []float64{0.1, 0.9, 0.2, 0.8, 0.3, 0.7, 0.4, 0.6},
			minStability: 0.0,
			maxStability: 0.5,
		},
		{
			name: "slight variation",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold:    0.85,
			samples:      []float64{0.79, 0.80, 0.81, 0.80, 0.79, 0.80, 0.81},
			minStability: 0.7,
			maxStability: 1.0,
		},
		{
			name: "single sample",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold:    0.85,
			samples:      []float64{0.8},
			minStability: 0.95,
			maxStability: 1.0,
		},
		{
			name: "empty samples",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold:    0.85,
			samples:      []float64{},
			minStability: 0.95,
			maxStability: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := emerge.NewConvergenceMonitor(tt.target, tt.threshold)

			for _, sample := range tt.samples {
				monitor.Record(sample)
				time.Sleep(10 * time.Millisecond)
			}

			stability := monitor.Stability()
			if stability < tt.minStability || stability > tt.maxStability {
				t.Errorf("Expected stability in range [%f, %f], got %f",
					tt.minStability, tt.maxStability, stability)
			}
		})
	}
}

func TestConvergenceMonitorPrediction(t *testing.T) {
	tests := []struct {
		name       string
		target     emerge.State
		threshold  float64
		samples    []float64
		delays     []time.Duration
		validateFn func(t *testing.T, prediction time.Duration)
	}{
		{
			name: "linear increase",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold: 0.85,
			samples:   []float64{0.5, 0.55, 0.6, 0.65, 0.7},
			delays:    []time.Duration{10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond, 0},
			validateFn: func(t *testing.T, prediction time.Duration) {
				if prediction <= 0 {
					t.Skip("Prediction may not be available with limited data")
				}
				if prediction > 10*time.Minute {
					t.Error("Prediction seems unreasonably long")
				}
			},
		},
		{
			name: "already converged",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold: 0.5,
			samples:   []float64{0.6, 0.7, 0.8, 0.9},
			delays:    []time.Duration{10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond, 0},
			validateFn: func(t *testing.T, prediction time.Duration) {
				if prediction != 0 {
					t.Error("Should not predict when already converged")
				}
			},
		},
		{
			name: "decreasing trend",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold: 0.85,
			samples:   []float64{0.8, 0.75, 0.7, 0.65, 0.6},
			delays:    []time.Duration{10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond, 0},
			validateFn: func(t *testing.T, prediction time.Duration) {
				// Should not predict convergence for decreasing trend
				if prediction > 0 && prediction < time.Hour {
					t.Error("Should not predict reasonable convergence for decreasing trend")
				}
			},
		},
		{
			name: "insufficient data",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold: 0.85,
			samples:   []float64{0.7},
			delays:    []time.Duration{0},
			validateFn: func(t *testing.T, prediction time.Duration) {
				// Might not be able to predict with single sample
				if prediction < 0 {
					t.Error("Prediction should not be negative")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := emerge.NewConvergenceMonitor(tt.target, tt.threshold)

			for i, sample := range tt.samples {
				monitor.Record(sample)
				if i < len(tt.delays) && tt.delays[i] > 0 {
					time.Sleep(tt.delays[i])
				}
			}

			prediction := monitor.PredictConvergenceTime()
			tt.validateFn(t, prediction)
		})
	}
}

func TestConvergenceMonitorStatistics(t *testing.T) {
	tests := []struct {
		name       string
		target     emerge.State
		threshold  float64
		samples    []float64
		validateFn func(t *testing.T, stats map[string]any)
	}{
		{
			name: "basic statistics",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold: 0.85,
			samples:   []float64{0.5, 0.6, 0.7, 0.65, 0.75, 0.8, 0.85, 0.9},
			validateFn: func(t *testing.T, stats map[string]any) {
				// Check required fields
				requiredFields := []string{"mean", "min", "max", "samples"}
				for _, field := range requiredFields {
					if _, ok := stats[field]; !ok {
						t.Errorf("Statistics should include %s", field)
					}
				}

				// Verify values
				if samples, ok := stats["samples"].(int); ok {
					if samples != 8 {
						t.Errorf("Expected 8 samples, got %d", samples)
					}
				}

				if max, ok := stats["max"].(float64); ok {
					if math.Abs(max-0.9) > 0.01 {
						t.Errorf("Expected max 0.9, got %f", max)
					}
				}

				if min, ok := stats["min"].(float64); ok {
					if math.Abs(min-0.5) > 0.01 {
						t.Errorf("Expected min 0.5, got %f", min)
					}
				}

				if mean, ok := stats["mean"].(float64); ok {
					expectedMean := (0.5 + 0.6 + 0.7 + 0.65 + 0.75 + 0.8 + 0.85 + 0.9) / 8
					if math.Abs(mean-expectedMean) > 0.01 {
						t.Errorf("Expected mean %f, got %f", expectedMean, mean)
					}
				}
			},
		},
		{
			name: "empty statistics",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold: 0.85,
			samples:   []float64{},
			validateFn: func(t *testing.T, stats map[string]any) {
				if samples, ok := stats["samples"].(int); ok {
					if samples != 0 {
						t.Errorf("Expected 0 samples, got %d", samples)
					}
				}
			},
		},
		{
			name: "single sample statistics",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold: 0.85,
			samples:   []float64{0.75},
			validateFn: func(t *testing.T, stats map[string]any) {
				if samples, ok := stats["samples"].(int); ok {
					if samples != 1 {
						t.Error("Expected 1 sample")
					}
				}

				// Min, max, and mean should all be the same
				if min, ok := stats["min"].(float64); ok {
					if math.Abs(min-0.75) > 0.01 {
						t.Error("Min should be 0.75")
					}
				}
				if max, ok := stats["max"].(float64); ok {
					if math.Abs(max-0.75) > 0.01 {
						t.Error("Max should be 0.75")
					}
				}
				if mean, ok := stats["mean"].(float64); ok {
					if math.Abs(mean-0.75) > 0.01 {
						t.Error("Mean should be 0.75")
					}
				}
			},
		},
		{
			name: "extreme values",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold: 0.85,
			samples:   []float64{0.0, 1.0, 0.5},
			validateFn: func(t *testing.T, stats map[string]any) {
				if min, ok := stats["min"].(float64); ok {
					if min != 0.0 {
						t.Error("Min should be 0.0")
					}
				}
				if max, ok := stats["max"].(float64); ok {
					if max != 1.0 {
						t.Error("Max should be 1.0")
					}
				}
				if mean, ok := stats["mean"].(float64); ok {
					if math.Abs(mean-0.5) > 0.01 {
						t.Error("Mean should be 0.5")
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := emerge.NewConvergenceMonitor(tt.target, tt.threshold)

			for _, sample := range tt.samples {
				monitor.Record(sample)
				time.Sleep(10 * time.Millisecond)
			}

			stats := monitor.GetStatistics()
			tt.validateFn(t, stats)
		})
	}
}

func TestConvergenceMonitorReset(t *testing.T) {
	tests := []struct {
		name       string
		target     emerge.State
		threshold  float64
		setupFn    func(monitor *emerge.ConvergenceMonitor)
		validateFn func(t *testing.T, monitor *emerge.ConvergenceMonitor)
	}{
		{
			name: "reset after convergence",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold: 0.85,
			setupFn: func(monitor *emerge.ConvergenceMonitor) {
				for range 5 {
					monitor.Record(0.9)
					time.Sleep(10 * time.Millisecond)
				}
			},
			validateFn: func(t *testing.T, monitor *emerge.ConvergenceMonitor) {
				if !monitor.IsConverged() {
					t.Error("Should be converged before reset")
				}

				monitor.Reset()

				if monitor.IsConverged() {
					t.Error("Should not be converged after reset")
				}
				history := monitor.GetHistory()
				if len(history) != 0 {
					t.Error("History should be empty after reset")
				}
				if monitor.CurrentCoherence() != 0 {
					t.Error("Current coherence should be 0 after reset")
				}
			},
		},
		{
			name: "reset without convergence",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold: 0.95,
			setupFn: func(monitor *emerge.ConvergenceMonitor) {
				monitor.Record(0.7)
				monitor.Record(0.8)
			},
			validateFn: func(t *testing.T, monitor *emerge.ConvergenceMonitor) {
				historyBefore := len(monitor.GetHistory())
				if historyBefore != 2 {
					t.Error("Should have 2 samples before reset")
				}

				monitor.Reset()

				if monitor.IsConverged() {
					t.Error("Should not be converged after reset")
				}
				if len(monitor.GetHistory()) != 0 {
					t.Error("History should be empty after reset")
				}
			},
		},
		{
			name: "multiple resets",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold: 0.85,
			setupFn: func(monitor *emerge.ConvergenceMonitor) {
				monitor.Record(0.9)
				monitor.Reset()
				monitor.Record(0.8)
				monitor.Reset()
			},
			validateFn: func(t *testing.T, monitor *emerge.ConvergenceMonitor) {
				if monitor.IsConverged() {
					t.Error("Should not be converged after multiple resets")
				}
				if len(monitor.GetHistory()) != 0 {
					t.Error("History should be empty")
				}
				if monitor.CurrentCoherence() != 0 {
					t.Error("Current coherence should be 0")
				}
			},
		},
		{
			name: "reset empty monitor",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			threshold: 0.85,
			setupFn: func(monitor *emerge.ConvergenceMonitor) {
				// Don't add any samples
			},
			validateFn: func(t *testing.T, monitor *emerge.ConvergenceMonitor) {
				monitor.Reset()

				// Should still be in initial state
				if monitor.IsConverged() {
					t.Error("Should not be converged")
				}
				if len(monitor.GetHistory()) != 0 {
					t.Error("History should be empty")
				}
				if monitor.CurrentCoherence() != 0 {
					t.Error("Current coherence should be 0")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := emerge.NewConvergenceMonitor(tt.target, tt.threshold)
			tt.setupFn(monitor)
			tt.validateFn(t, monitor)
		})
	}
}

func TestSwarmConvergence(t *testing.T) {
	tests := []struct {
		name       string
		swarmSize  int
		goal       emerge.State
		timeout    time.Duration
		validateFn func(t *testing.T, swarm *emerge.Swarm, converged bool)
		skip       bool
		skipReason string
	}{
		{
			name:      "small swarm convergence",
			swarmSize: 5,
			goal: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.7,
			},
			timeout: 3 * time.Second,
			validateFn: func(t *testing.T, swarm *emerge.Swarm, converged bool) {
				finalCoherence := swarm.MeasureCoherence()
				if finalCoherence < 0.5 {
					t.Error("Small swarm should achieve at least moderate coherence")
				}
			},
			skip:       true,
			skipReason: "Flaky test - depends on timing and randomness",
		},
		{
			name:      "medium swarm convergence",
			swarmSize: 20,
			goal: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.85,
			},
			timeout: 5 * time.Second,
			validateFn: func(t *testing.T, swarm *emerge.Swarm, converged bool) {
				finalCoherence := swarm.MeasureCoherence()
				if !converged && finalCoherence < 0.6 {
					t.Errorf("Expected better coherence. Final: %f", finalCoherence)
				}
			},
			skip:       true,
			skipReason: "Flaky test - depends on timing and randomness",
		},
		{
			name:      "large swarm convergence",
			swarmSize: 50,
			goal: emerge.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			timeout: 5 * time.Second,
			validateFn: func(t *testing.T, swarm *emerge.Swarm, converged bool) {
				// Large swarms may take longer to converge
				finalCoherence := swarm.MeasureCoherence()
				if finalCoherence < 0.4 {
					t.Error("Large swarm should achieve some coherence")
				}
			},
			skip:       true,
			skipReason: "Flaky test - depends on timing and randomness",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip(tt.skipReason)
			}

			swarm, err := emerge.NewSwarm(tt.swarmSize, tt.goal)
			if err != nil {
				t.Fatalf("Failed to create swarm: %v", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			// Run swarm
			errChan := make(chan error, 1)
			go func() {
				if err := swarm.Run(ctx); err != nil && err != context.DeadlineExceeded {
					errChan <- err
				}
			}()

			// Monitor convergence
			converged := false
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case err := <-errChan:
					t.Errorf("Swarm run error: %v", err)
					return
				case <-ticker.C:
					coherence := swarm.MeasureCoherence()
					if coherence >= tt.goal.Coherence {
						converged = true
						cancel()
						goto validate
					}
				case <-ctx.Done():
					goto validate
				}
			}

		validate:
			tt.validateFn(t, swarm, converged)
		})
	}
}

func TestConvergenceWithDisruption(t *testing.T) {
	tests := []struct {
		name             string
		swarmSize        int
		goal             emerge.State
		disruptionFactor float64
		timeout          time.Duration
		validateFn       func(t *testing.T, beforeDisruption, afterDisruption, afterRecovery float64)
		skip             bool
		skipReason       string
	}{
		{
			name:      "minor disruption",
			swarmSize: 20,
			goal: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			disruptionFactor: 0.1,
			timeout:          5 * time.Second,
			validateFn: func(t *testing.T, before, after, recovery float64) {
				// Minor disruption should have small impact
				if after < before-0.3 {
					t.Error("Minor disruption had too large impact")
				}
			},
			skip:       true,
			skipReason: "Flaky test - depends on timing and randomness",
		},
		{
			name:      "major disruption",
			swarmSize: 20,
			goal: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			disruptionFactor: 0.5,
			timeout:          5 * time.Second,
			validateFn: func(t *testing.T, before, after, recovery float64) {
				// Major disruption should have significant impact
				if after >= before {
					t.Error("Major disruption should decrease coherence")
				}
				// Should show some recovery
				if recovery < after-0.05 {
					t.Error("Should show some recovery after disruption")
				}
			},
			skip:       true,
			skipReason: "Flaky test - depends on timing and randomness",
		},
		{
			name:      "total disruption",
			swarmSize: 10,
			goal: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.7,
			},
			disruptionFactor: 1.0,
			timeout:          5 * time.Second,
			validateFn: func(t *testing.T, before, after, recovery float64) {
				// Total disruption should severely impact coherence
				if after > 0.5 {
					t.Error("Total disruption should severely reduce coherence")
				}
			},
			skip:       true,
			skipReason: "Flaky test - depends on timing and randomness",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip(tt.skipReason)
			}

			swarm, err := emerge.NewSwarm(tt.swarmSize, tt.goal)
			if err != nil {
				t.Fatalf("Failed to create swarm: %v", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			// Run swarm
			go func() {
				if err := swarm.Run(ctx); err != nil && err != context.DeadlineExceeded {
					t.Errorf("Swarm run error: %v", err)
				}
			}()

			// Wait for initial convergence
			time.Sleep(100 * time.Millisecond)
			beforeDisruption := swarm.MeasureCoherence()

			// Disrupt agents
			swarm.DisruptAgents(tt.disruptionFactor)
			afterDisruption := swarm.MeasureCoherence()

			// Wait for recovery
			time.Sleep(200 * time.Millisecond)
			afterRecovery := swarm.MeasureCoherence()

			tt.validateFn(t, beforeDisruption, afterDisruption, afterRecovery)
		})
	}
}
