package swarm

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/carlisia/bio-adapt/emerge/agent"
	"github.com/carlisia/bio-adapt/emerge/core"
)

// BenchmarkOriginalSwarm tests the original sync.Map based implementation.
//
//nolint:intrange // b.N is not a constant
func BenchmarkOriginalSwarm(b *testing.B) {
	sizes := []int{100, 500, 1000, 2000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				s, err := New(size, core.State{
					Phase:     0,
					Frequency: 100 * time.Millisecond,
					Coherence: 0.7,
				})
				if err != nil {
					b.Fatal(err)
				}

				// Measure coherence calculation
				for j := 0; j < 10; j++ {
					_ = s.MeasureCoherence()
				}
			}
		})
	}
}

// BenchmarkLargeSwarm tests the automatically optimized implementation for large swarms.
//
//nolint:intrange // b.N is not a constant
func BenchmarkLargeSwarm(b *testing.B) {
	sizes := []int{200, 500, 1000, 2000} // All above OptimizedSwarmThreshold

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				s, err := New(size, core.State{
					Phase:     0,
					Frequency: 100 * time.Millisecond,
					Coherence: 0.7,
				})
				if err != nil {
					b.Fatal(err)
				}

				// Measure coherence calculation
				for j := 0; j < 10; j++ {
					_ = s.MeasureCoherence()
				}

				// Clean up worker pool if it exists
				if s.workerPool != nil {
					s.workerPool.Stop()
				}
			}
		})
	}
}

// BenchmarkCoherenceMeasurement compares coherence calculation performance.
//
//nolint:intrange // b.N is not a constant
func BenchmarkCoherenceMeasurement(b *testing.B) {
	sizes := []int{100, 1000, 5000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("original_%d", size), func(b *testing.B) {
			s, _ := New(size, core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.7,
			})

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = s.MeasureCoherence()
			}
		})

		b.Run(fmt.Sprintf("optimized_%d", size), func(b *testing.B) {
			s, _ := New(size+OptimizedSwarmThreshold, core.State{ // Force optimized path
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.7,
			})
			defer func() {
				if s.workerPool != nil {
					s.workerPool.Stop()
				}
			}()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = s.MeasureCoherence()
			}
		})
	}
}

// BenchmarkConcurrentUpdates tests concurrent update performance.
//
//nolint:intrange // b.N is not a constant
func BenchmarkConcurrentUpdates(b *testing.B) {
	updateFunc := func(a *agent.Agent) {
		// Simulate some work
		phase := a.Phase()
		a.SetPhase(phase + 0.01)
	}

	sizes := []int{1000, 5000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("original_%d", size), func(b *testing.B) {
			s, _ := New(size, core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.7,
			})

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				// Original returns a map copy which creates overhead
				agents := s.Agents()
				for _, a := range agents {
					updateFunc(a)
				}
			}
		})

		b.Run(fmt.Sprintf("optimized_%d", size), func(b *testing.B) {
			s, _ := New(size+OptimizedSwarmThreshold, core.State{ // Force optimized path
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.7,
			})
			defer func() {
				if s.workerPool != nil {
					s.workerPool.Stop()
				}
			}()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				agents := s.Agents()
				for _, a := range agents {
					updateFunc(a)
				}
			}
		})
	}
}

// BenchmarkMemoryUsage compares memory usage patterns.
//
//nolint:intrange // b.N is not a constant
func BenchmarkMemoryUsage(b *testing.B) {
	b.Run("original_1000_agents", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			s, _ := New(1000, core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.7,
			})

			// Run briefly to simulate usage
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			_ = s.Run(ctx)
			cancel()
		}
	})

	b.Run("optimized_1000_agents", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			s, _ := New(1100, core.State{ // Force optimized path
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.7,
			})

			// Simulate equivalent work
			for j := 0; j < 10; j++ {
				agents := s.Agents()
				for _, a := range agents {
					phase := a.Phase()
					a.SetPhase(phase + 0.01)
				}
			}

			if s.workerPool != nil {
				s.workerPool.Stop()
			}
		}
	})
}
