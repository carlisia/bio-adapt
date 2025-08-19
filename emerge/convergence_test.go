package emerge

import (
	"context"
	"math"
	"sync"
	"testing"
	"time"
)

func TestNewConvergenceMonitor(t *testing.T) {
	target := State{
		Phase:     math.Pi,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	monitor := NewConvergenceMonitor(target, 0.85)

	if monitor.targetState.Phase != target.Phase {
		t.Errorf("Expected target phase %f, got %f", target.Phase, monitor.targetState.Phase)
	}

	if monitor.convergenceAt != 0.85 {
		t.Errorf("Expected convergence threshold 0.85, got %f", monitor.convergenceAt)
	}

	if monitor.maxCoherence != 0 {
		t.Error("Initial max coherence should be 0")
	}

	if monitor.minCoherence != 1 {
		t.Error("Initial min coherence should be 1")
	}
}

func TestConvergenceMonitorRecord(t *testing.T) {
	target := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	monitor := NewConvergenceMonitor(target, 0.85)

	// Add some samples
	monitor.Record(0.5)
	time.Sleep(10 * time.Millisecond)
	monitor.Record(0.6)
	time.Sleep(10 * time.Millisecond)
	monitor.Record(0.7)

	if len(monitor.history) != 3 {
		t.Errorf("Expected 3 samples in history, got %d", len(monitor.history))
	}

	if monitor.maxCoherence != 0.7 {
		t.Errorf("Expected max coherence 0.7, got %f", monitor.maxCoherence)
	}

	if monitor.minCoherence != 0.5 {
		t.Errorf("Expected min coherence 0.5, got %f", monitor.minCoherence)
	}
}

func TestConvergenceMonitorIsConverged(t *testing.T) {
	target := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	tests := []struct {
		name      string
		samples   []float64
		threshold float64
		expected  bool
	}{
		{
			name:      "converged above threshold",
			samples:   []float64{0.85, 0.87, 0.89, 0.91, 0.92},
			threshold: 0.85,
			expected:  true,
		},
		{
			name:      "not converged below threshold",
			samples:   []float64{0.3, 0.4, 0.5, 0.6, 0.7},
			threshold: 0.85,
			expected:  false,
		},
		{
			name:      "converged at exact threshold",
			samples:   []float64{0.83, 0.84, 0.85, 0.85, 0.85},
			threshold: 0.85,
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := NewConvergenceMonitor(target, tt.threshold)

			for _, sample := range tt.samples {
				monitor.Record(sample)
				time.Sleep(10 * time.Millisecond)
			}

			result := monitor.IsConverged()
			if result != tt.expected {
				t.Errorf("Expected IsConverged() = %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestConvergenceMonitorRate(t *testing.T) {
	target := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	monitor := NewConvergenceMonitor(target, 0.85)

	// Simulate improving convergence
	values := []float64{0.3, 0.5, 0.6, 0.7, 0.75, 0.78}
	for _, v := range values {
		monitor.Record(v)
		time.Sleep(10 * time.Millisecond)
	}

	rate := monitor.ConvergenceRate()
	if rate <= 0 {
		t.Error("Convergence rate should be positive for improving values")
	}
}

func TestConvergenceMonitorGetStatistics(t *testing.T) {
	target := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	monitor := NewConvergenceMonitor(target, 0.85)

	// Add samples and mark convergence
	values := []float64{0.5, 0.6, 0.7, 0.8, 0.86, 0.88}
	for _, v := range values {
		monitor.Record(v)
		time.Sleep(10 * time.Millisecond)
	}

	stats := monitor.GetStatistics()

	if minCoherence, ok := stats["min_coherence"].(float64); !ok || minCoherence != 0.5 {
		t.Errorf("Expected min coherence 0.5, got %v", stats["min_coherence"])
	}

	if maxCoherence, ok := stats["max_coherence"].(float64); !ok || maxCoherence != 0.88 {
		t.Errorf("Expected max coherence 0.88, got %v", stats["max_coherence"])
	}

	if currentCoherence, ok := stats["current_coherence"].(float64); !ok || currentCoherence != 0.88 {
		t.Errorf("Expected current coherence 0.88, got %v", stats["current_coherence"])
	}

	if samples, ok := stats["samples"].(int); !ok || samples != 6 {
		t.Errorf("Expected 6 samples, got %v", stats["samples"])
	}
}

func TestConvergencePrediction(t *testing.T) {
	target := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	monitor := NewConvergenceMonitor(target, 0.85)

	// Add samples simulating convergence
	samples := []float64{0.3, 0.4, 0.5, 0.6, 0.7, 0.75, 0.8}
	for _, s := range samples {
		monitor.Record(s)
		time.Sleep(10 * time.Millisecond)
	}

	// Test convergence rate
	rate := monitor.ConvergenceRate()
	if rate <= 0 {
		t.Error("Should have positive convergence rate")
	}

	// Test prediction
	prediction := monitor.PredictConvergenceTime()
	if prediction <= 0 {
		t.Error("Should predict positive time to convergence")
	}

	// Test stability
	stability := monitor.Stability()
	if stability < 0 || stability > 1 {
		t.Errorf("Stability should be in [0, 1], got %f", stability)
	}
}

func TestConvergenceMonitorReset(t *testing.T) {
	target := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	monitor := NewConvergenceMonitor(target, 0.85)

	// Add some samples
	monitor.Record(0.5)
	monitor.Record(0.6)
	monitor.Record(0.7)

	history := monitor.GetHistory()
	if len(history) != 3 {
		t.Errorf("Expected 3 samples before reset, got %d", len(history))
	}

	monitor.Reset()

	history = monitor.GetHistory()
	if len(history) != 0 {
		t.Errorf("Expected 0 samples after reset, got %d", len(history))
	}

	if monitor.IsConverged() {
		t.Error("Should not be converged after reset")
	}
}

func TestConvergenceIntegration(t *testing.T) {
	// Test the full convergence monitoring system
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	target := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	monitor := NewConvergenceMonitor(target, 0.85)

	var wg sync.WaitGroup
	converged := false

	// Simulate convergence monitoring
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(20 * time.Millisecond)
		defer ticker.Stop()

		coherence := 0.3
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Simulate improving coherence
				coherence += 0.05
				if coherence > 0.9 {
					coherence = 0.9
				}

				monitor.Record(coherence)

				if monitor.IsConverged() && !converged {
					converged = true
					return
				}
			}
		}
	}()

	wg.Wait()

	if !converged {
		t.Error("System should have detected convergence")
	}

	stats := monitor.GetStatistics()
	if maxCoherence, ok := stats["max_coherence"].(float64); !ok || maxCoherence < 0.85 {
		t.Error("Should have reached convergence threshold")
	}
}

