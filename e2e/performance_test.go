//nolint:paralleltest,intrange // E2E tests shouldn't run in parallel, benchmark loops need explicit indexing
package e2e_test

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/swarm"
)

// TestOptimizationImpact verifies that optimizations are working
func TestOptimizationImpact(t *testing.T) {
	tests := []struct {
		name      string
		swarmSize int
		maxTime   time.Duration
	}{
		{
			name:      "small_optimized",
			swarmSize: 50,
			maxTime:   2 * time.Second,
		},
		{
			name:      "medium_optimized",
			swarmSize: 200,
			maxTime:   5 * time.Second,
		},
		{
			name:      "large_optimized",
			swarmSize: 1000,
			maxTime:   10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goal := core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.85,
			}

			// Create swarm (will use optimized agents for size > 100)
			start := time.Now()
			s, err := swarm.New(tt.swarmSize, goal)
			createTime := time.Since(start)
			require.NoError(t, err)

			// Log creation performance
			t.Logf("%s: Created %d agents in %v (%.2f µs/agent)",
				tt.name, tt.swarmSize, createTime,
				float64(createTime.Microseconds())/float64(tt.swarmSize))

			// Test convergence
			ctx, cancel := context.WithTimeout(context.Background(), tt.maxTime)
			defer cancel()

			convergeStart := time.Now()
			done := make(chan struct{})

			// Monitor convergence
			go func() {
				ticker := time.NewTicker(100 * time.Millisecond)
				defer ticker.Stop()
				for {
					select {
					case <-ticker.C:
						if s.MeasureCoherence() >= 0.85 {
							close(done)
							return
						}
					case <-ctx.Done():
						return
					}
				}
			}()

			// Run swarm
			go s.Run(ctx)

			select {
			case <-done:
				convergeTime := time.Since(convergeStart)
				t.Logf("%s: Converged in %v (%.2f ms/agent)",
					tt.name, convergeTime,
					float64(convergeTime.Milliseconds())/float64(tt.swarmSize))

				// Performance assertions
				assert.Less(t, createTime, time.Duration(tt.swarmSize)*time.Millisecond,
					"Creation should be fast")
				assert.Less(t, convergeTime, tt.maxTime,
					"Should converge within time limit")
			case <-ctx.Done():
				t.Errorf("%s: Failed to converge in %v", tt.name, tt.maxTime)
			}
		})
	}
}

// BenchmarkOptimizedSwarm benchmarks the optimized implementation
func BenchmarkOptimizedSwarm(b *testing.B) {
	sizes := []int{10, 100, 500}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			goal := core.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.80,
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				s, _ := swarm.New(size, goal)
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				s.Run(ctx)
				cancel()
			}
		})
	}
}
