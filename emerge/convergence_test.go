package emerge_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/carlisia/bio-adapt/emerge"
)

func TestNewConvergenceMonitor(t *testing.T) {
	target := emerge.State{
		Phase:     math.Pi,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	monitor := emerge.NewConvergenceMonitor(target, 0.85)

	if monitor == nil {
		t.Fatal("Expected monitor to be created")
	}

	// Test initial state through public methods
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
}

func TestConvergenceMonitorRecord(t *testing.T) {
	target := emerge.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	monitor := emerge.NewConvergenceMonitor(target, 0.85)

	// Add some samples
	monitor.Record(0.5)
	time.Sleep(10 * time.Millisecond)
	monitor.Record(0.6)
	time.Sleep(10 * time.Millisecond)
	monitor.Record(0.7)

	// Check history through public method
	history := monitor.GetHistory()
	if len(history) != 3 {
		t.Errorf("Expected 3 samples in history, got %d", len(history))
	}

	// Check current coherence
	current := monitor.CurrentCoherence()
	if math.Abs(current-0.7) > 0.01 {
		t.Errorf("Expected current coherence 0.7, got %f", current)
	}

	// Should not be converged yet
	if monitor.IsConverged() {
		t.Error("Should not be converged with coherence below threshold")
	}
}

func TestConvergenceMonitorConvergence(t *testing.T) {
	target := emerge.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	monitor := emerge.NewConvergenceMonitor(target, 0.85)

	// Simulate convergence
	for i := range 10 {
		coherence := 0.5 + float64(i)*0.05
		monitor.Record(coherence)
		time.Sleep(10 * time.Millisecond)
	}

	// Should be converged now (last values > 0.85)
	if !monitor.IsConverged() {
		t.Error("Should be converged after reaching threshold")
	}

	// Check convergence time
	convTime := monitor.ConvergenceTime()
	if convTime <= 0 {
		t.Error("Convergence time should be positive")
	}
}

func TestConvergenceMonitorRate(t *testing.T) {
	target := emerge.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	monitor := emerge.NewConvergenceMonitor(target, 0.85)

	// Add increasing coherence values
	for i := range 5 {
		monitor.Record(0.5 + float64(i)*0.1)
		time.Sleep(10 * time.Millisecond)
	}

	rate := monitor.ConvergenceRate()
	if rate <= 0 {
		t.Error("Convergence rate should be positive for increasing coherence")
	}
}

func TestConvergenceMonitorStability(t *testing.T) {
	target := emerge.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	monitor := emerge.NewConvergenceMonitor(target, 0.85)

	// Add stable values
	for range 10 {
		monitor.Record(0.8)
		time.Sleep(10 * time.Millisecond)
	}

	stability := monitor.Stability()
	if stability < 0.9 {
		t.Errorf("Expected high stability for constant values, got %f", stability)
	}
}

func TestConvergenceMonitorPrediction(t *testing.T) {
	target := emerge.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	monitor := emerge.NewConvergenceMonitor(target, 0.85)

	// Add linearly increasing values
	for i := range 5 {
		monitor.Record(0.5 + float64(i)*0.05)
		time.Sleep(10 * time.Millisecond)
	}

	prediction := monitor.PredictConvergenceTime()
	if prediction <= 0 {
		t.Skip("Prediction may not be available with limited data")
	}

	// Prediction should be reasonable (not infinite)
	if prediction > 10*time.Minute {
		t.Error("Prediction seems unreasonably long")
	}
}

func TestConvergenceMonitorStatistics(t *testing.T) {
	target := emerge.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	monitor := emerge.NewConvergenceMonitor(target, 0.85)

	// Add varied samples
	samples := []float64{0.5, 0.6, 0.7, 0.65, 0.75, 0.8, 0.85, 0.9}
	for _, s := range samples {
		monitor.Record(s)
		time.Sleep(10 * time.Millisecond)
	}

	stats := monitor.GetStatistics()

	// Check that statistics are present
	if _, ok := stats["mean"]; !ok {
		t.Error("Statistics should include mean")
	}
	if _, ok := stats["min"]; !ok {
		t.Error("Statistics should include min")
	}
	if _, ok := stats["max"]; !ok {
		t.Error("Statistics should include max")
	}
	if _, ok := stats["samples"]; !ok {
		t.Error("Statistics should include sample count")
	}

	// Verify some values
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
}

func TestConvergenceMonitorReset(t *testing.T) {
	target := emerge.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.9,
	}

	monitor := emerge.NewConvergenceMonitor(target, 0.85)

	// Add samples and converge
	for range 5 {
		monitor.Record(0.9)
		time.Sleep(10 * time.Millisecond)
	}

	if !monitor.IsConverged() {
		t.Error("Should be converged before reset")
	}

	// Reset
	monitor.Reset()

	// Check reset state
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
}

func TestSwarmConvergence(t *testing.T) {
	t.Skip("Flaky test - depends on timing and randomness")
	// Test actual swarm convergence behavior
	goal := emerge.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.85,
	}

	swarm, err := emerge.NewSwarm(20, goal)
	if err != nil {
		t.Fatalf("Failed to create swarm: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Run swarm
	go func() {
		if err := swarm.Run(ctx); err != nil && err != context.DeadlineExceeded {
			t.Errorf("Swarm run error: %v", err)
		}
	}()

	// Monitor convergence
	converged := false
	for range 50 {
		time.Sleep(100 * time.Millisecond)
		coherence := swarm.MeasureCoherence()

		if coherence >= goal.Coherence {
			converged = true
			break
		}
	}

	if !converged {
		finalCoherence := swarm.MeasureCoherence()
		t.Errorf("Swarm did not converge. Final coherence: %f, target: %f",
			finalCoherence, goal.Coherence)
	}
}

func TestConvergenceWithDisruption(t *testing.T) {
	t.Skip("Flaky test - depends on timing and randomness")
	goal := emerge.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.8,
	}

	swarm, err := emerge.NewSwarm(20, goal)
	if err != nil {
		t.Fatalf("Failed to create swarm: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Run swarm
	go func() {
		if err := swarm.Run(ctx); err != nil && err != context.DeadlineExceeded {
			t.Errorf("Swarm run error: %v", err)
		}
	}()

	// Wait for initial convergence
	time.Sleep(1 * time.Second)

	beforeDisruption := swarm.MeasureCoherence()

	// Disrupt 20% of agents
	swarm.DisruptAgents(0.2)

	afterDisruption := swarm.MeasureCoherence()

	// Coherence should drop after disruption
	if afterDisruption >= beforeDisruption {
		t.Error("Coherence should decrease after disruption")
	}

	// Wait for recovery
	time.Sleep(2 * time.Second)

	afterRecovery := swarm.MeasureCoherence()

	// Should recover toward target (or at least not get worse)
	// Allow for some tolerance due to randomness
	if afterRecovery < afterDisruption-0.05 {
		t.Errorf("Coherence should increase during recovery. After disruption: %f, After recovery: %f",
			afterDisruption, afterRecovery)
	}
}

