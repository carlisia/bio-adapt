package swarm

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/internal/config"
)

func TestSwarmScalability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping scalability test in short mode")
	}

	tests := []struct {
		name               string
		size               int
		expectedBatching   bool
		expectedWorkerPool bool
		timeout            time.Duration
	}{
		{
			name:               "small_swarm_no_optimization",
			size:               10,
			expectedBatching:   false,
			expectedWorkerPool: false,
			timeout:            5 * time.Second,
		},
		{
			name:               "medium_swarm_basic_optimization",
			size:               100,
			expectedBatching:   false,
			expectedWorkerPool: true, // config.AutoScaleConfig does set MaxConcurrentAgents for 100
			timeout:            10 * time.Second,
		},
		{
			name:               "large_swarm_with_batching",
			size:               500,
			expectedBatching:   true,
			expectedWorkerPool: true,
			timeout:            15 * time.Second,
		},
		{
			name:               "very_large_swarm_optimized",
			size:               1500,
			expectedBatching:   true,
			expectedWorkerPool: true,
			timeout:            20 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Measure memory before
			var memBefore runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&memBefore)

			// Create swarm with auto-scaling config
			config := config.AutoScaleConfig(tt.size)
			swarm, err := New(tt.size, core.State{
				Phase:     0,
				Frequency: 200 * time.Millisecond,
				Coherence: 0.7,
			}, WithConfig(config))

			if err != nil {
				t.Fatalf("Failed to create swarm: %v", err)
			}

			// Verify configuration
			if config.UseBatchProcessing != tt.expectedBatching {
				t.Errorf("Expected batching %v, got %v", tt.expectedBatching, config.UseBatchProcessing)
			}

			if (config.MaxConcurrentAgents > 0) != tt.expectedWorkerPool {
				t.Errorf("Expected worker pool %v, got %v (MaxConcurrentAgents: %d)",
					tt.expectedWorkerPool, config.MaxConcurrentAgents > 0, config.MaxConcurrentAgents)
			}

			// Verify swarm was created with correct size
			if swarm.Size() != tt.size {
				t.Errorf("Expected swarm size %d, got %d", tt.size, swarm.Size())
			}

			// Measure initial coherence
			initialCoherence := swarm.MeasureCoherence()
			if initialCoherence < 0 || initialCoherence > 1 {
				t.Errorf("Invalid initial coherence: %f", initialCoherence)
			}

			// Run swarm with timeout
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			done := make(chan error, 1)
			go func() {
				done <- swarm.Run(ctx)
			}()

			// Let it run for a bit
			time.Sleep(2 * time.Second)

			// Measure coherence after running
			finalCoherence := swarm.MeasureCoherence()
			if finalCoherence < 0 || finalCoherence > 1 {
				t.Errorf("Invalid final coherence: %f", finalCoherence)
			}

			// Cancel and wait for completion
			cancel()
			select {
			case err := <-done:
				if err != nil {
					t.Errorf("Swarm run failed: %v", err)
				}
			case <-time.After(5 * time.Second):
				t.Error("Swarm failed to shutdown within timeout")
			}

			// Measure memory after
			var memAfter runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&memAfter)

			// Check memory usage (this is a rough check)
			var memUsedMB float64
			if memAfter.Alloc >= memBefore.Alloc {
				memUsed := memAfter.Alloc - memBefore.Alloc
				memUsedMB = float64(memUsed) / (1024 * 1024)
			} else {
				// Memory might have been freed by GC, use total allocated
				memUsedMB = float64(memAfter.TotalAlloc) / (1024 * 1024)
			}

			t.Logf("Swarm size: %d, Memory used: %.2f MB, Initial coherence: %.3f, Final coherence: %.3f",
				tt.size, memUsedMB, initialCoherence, finalCoherence)

			// Very rough memory check - should not use excessive memory per agent
			expectedMaxMemoryMB := float64(tt.size) * 0.5          // ~500KB per agent max (more realistic)
			if memUsedMB > expectedMaxMemoryMB && memUsedMB > 10 { // Only warn if > 10MB
				t.Logf("Warning: High memory usage: %.2f MB (expected < %.2f MB)", memUsedMB, expectedMaxMemoryMB)
			}
		})
	}
}

func TestConfigurableLimits(t *testing.T) {
	tests := []struct {
		name          string
		config        config.Swarm
		size          int
		expectError   bool
		errorContains string
	}{
		{
			name: "within_limits",
			config: config.Swarm{
				MaxSwarmSize:             1000,
				MaxConcurrentAgents:      100,
				UseBatchProcessing:       false,
				AgentUpdateInterval:      50 * time.Millisecond,
				MonitoringInterval:       100 * time.Millisecond,
				ConnectionOptimThreshold: 500,
				EnableConnectionOptim:    true,
				// Include other required fields
				ConnectionProbability: 0.5,
				MaxNeighbors:          5,
				MinNeighbors:          2,
				CouplingStrength:      0.5,
				Stubbornness:          0.2,
				InitialEnergy:         100.0,
				BasinStrength:         0.8,
				BasinWidth:            3.14,
				BaseConfidence:        0.6,
				InfluenceDefault:      0.5,
			},
			size:        500,
			expectError: false,
		},
		{
			name: "exceeds_size_limit",
			config: config.Swarm{
				MaxSwarmSize:             100,
				MaxConcurrentAgents:      50,
				UseBatchProcessing:       false,
				AgentUpdateInterval:      50 * time.Millisecond,
				MonitoringInterval:       100 * time.Millisecond,
				ConnectionOptimThreshold: 500,
				EnableConnectionOptim:    false,
				// Include other required fields
				ConnectionProbability: 0.5,
				MaxNeighbors:          5,
				MinNeighbors:          2,
				CouplingStrength:      0.5,
				Stubbornness:          0.2,
				InitialEnergy:         100.0,
				BasinStrength:         0.8,
				BasinWidth:            3.14,
				BaseConfidence:        0.6,
				InfluenceDefault:      0.5,
			},
			size:          500,
			expectError:   true,
			errorContains: "exceeds configured maximum",
		},
		{
			name: "zero_size_limit_unlimited",
			config: config.Swarm{
				MaxSwarmSize:             0, // Unlimited
				MaxConcurrentAgents:      100,
				UseBatchProcessing:       true,
				BatchSize:                50,
				WorkerPoolSize:           10,
				AgentUpdateInterval:      50 * time.Millisecond,
				MonitoringInterval:       100 * time.Millisecond,
				ConnectionOptimThreshold: 500,
				EnableConnectionOptim:    true,
				// Include other required fields
				ConnectionProbability: 0.5,
				MaxNeighbors:          5,
				MinNeighbors:          2,
				CouplingStrength:      0.5,
				Stubbornness:          0.2,
				InitialEnergy:         100.0,
				BasinStrength:         0.8,
				BasinWidth:            3.14,
				BaseConfidence:        0.6,
				InfluenceDefault:      0.5,
			},
			size:        1000,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			swarm, err := New(tt.size, core.State{
				Phase:     0,
				Frequency: 200 * time.Millisecond,
				Coherence: 0.7,
			}, WithConfig(tt.config))

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', but got: %v", tt.errorContains, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify the configuration was applied
			if swarm.config.MaxSwarmSize != tt.config.MaxSwarmSize {
				t.Errorf("MaxSwarmSize not applied correctly: got %d, want %d",
					swarm.config.MaxSwarmSize, tt.config.MaxSwarmSize)
			}

			if swarm.config.UseBatchProcessing != tt.config.UseBatchProcessing {
				t.Errorf("UseBatchProcessing not applied correctly: got %v, want %v",
					swarm.config.UseBatchProcessing, tt.config.UseBatchProcessing)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(substr) <= len(s) && s[len(s)-len(substr):] == substr ||
		len(substr) <= len(s) && s[:len(substr)] == substr ||
		(len(substr) < len(s) && containsMiddle(s, substr))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func BenchmarkSwarmCreationScalability(b *testing.B) {
	sizes := []int{10, 100, 500, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				config := config.AutoScaleConfig(size)
				swarm, err := New(size, core.State{
					Phase:     0,
					Frequency: 200 * time.Millisecond,
					Coherence: 0.7,
				}, WithConfig(config))

				if err != nil {
					b.Fatal(err)
				}

				// Measure initial coherence to ensure swarm is fully initialized
				_ = swarm.MeasureCoherence()
			}
		})
	}
}
