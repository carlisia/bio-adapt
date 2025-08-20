package swarm

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

			require.NoError(t, err, "Failed to create swarm")

			// Verify configuration
			assert.Equal(t, tt.expectedBatching, config.UseBatchProcessing, "Expected batching %v", tt.expectedBatching)
			assert.Equal(t, tt.expectedWorkerPool, config.MaxConcurrentAgents > 0, "Expected worker pool %v (MaxConcurrentAgents: %d)", tt.expectedWorkerPool, config.MaxConcurrentAgents)

			// Verify swarm was created with correct size
			assert.Equal(t, tt.size, swarm.Size(), "Expected swarm size %d", tt.size)

			// Measure initial coherence
			initialCoherence := swarm.MeasureCoherence()
			assert.GreaterOrEqual(t, initialCoherence, 0.0, "Initial coherence should be >= 0")
			assert.LessOrEqual(t, initialCoherence, 1.0, "Initial coherence should be <= 1")

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
			assert.GreaterOrEqual(t, finalCoherence, 0.0, "Final coherence should be >= 0")
			assert.LessOrEqual(t, finalCoherence, 1.0, "Final coherence should be <= 1")

			// Cancel and wait for completion
			cancel()
			select {
			case err := <-done:
				// Context cancellation is expected and not an error
				if err != nil {
					assert.True(t, err == context.Canceled || err == context.DeadlineExceeded, "Swarm run failed: %v", err)
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
				require.Error(t, err, "Expected error but got none")
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains, "Expected error to contain '%s'", tt.errorContains)
				}
				return
			}

			assert.NoError(t, err, "Unexpected error")

			// Verify the configuration was applied
			assert.Equal(t, tt.config.MaxSwarmSize, swarm.config.MaxSwarmSize, "MaxSwarmSize not applied correctly")
			assert.Equal(t, tt.config.UseBatchProcessing, swarm.config.UseBatchProcessing, "UseBatchProcessing not applied correctly")
		})
	}
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

				require.NoError(b, err, "Failed to create swarm in benchmark")

				// Measure initial coherence to ensure swarm is fully initialized
				_ = swarm.MeasureCoherence()
			}
		})
	}
}
