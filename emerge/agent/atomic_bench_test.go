package agent

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/carlisia/bio-adapt/emerge/core"
)

// BenchmarkAtomicOperations compares standard vs grouped atomic operations.
func BenchmarkAtomicOperations(b *testing.B) {
	b.Run("standard_atomics", func(b *testing.B) {
		a := New("test")
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			// Simulate typical access pattern
			_ = a.Phase()
			_ = a.Energy()
			_ = a.LocalGoal()
			a.SetPhase(float64(i) * 0.01)
			a.SetEnergy(100 - float64(i)*0.1)
		}
	})

	b.Run("grouped_atomics", func(b *testing.B) {
		a := New("test")
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			// Simulate typical access pattern
			state := a.state.Load() // Single atomic read
			_ = state.Phase
			_ = state.Energy
			_ = state.LocalGoal
			a.state.Update(func(s *StateData) {
				s.Phase = float64(i) * 0.01
				s.Energy = 100 - float64(i)*0.1
			}) // Single atomic write
		}
	})

	b.Run("grouped_atomics_v2", func(b *testing.B) {
		state := NewAtomicState()
		state.Store(StateData{
			Phase:     0,
			Energy:    100,
			LocalGoal: 0,
			Frequency: 100 * time.Millisecond,
		})
		
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			// Simulate typical access pattern
			s := state.Load() // Single atomic read
			_ = s.Phase
			_ = s.Energy
			_ = s.LocalGoal
			state.Update(func(s *StateData) {
				s.Phase = float64(i) * 0.01
				s.Energy = 100 - float64(i)*0.1
			}) // Single atomic write
		}
	})
}

// BenchmarkConcurrentAccess tests concurrent field access.
func BenchmarkConcurrentAccess(b *testing.B) {
	numGoroutines := 100

	b.Run("standard_concurrent", func(b *testing.B) {
		a := New("test")
		b.ResetTimer()
		b.ReportAllocs()

		var wg sync.WaitGroup
		for i := 0; i < b.N; i++ {
			wg.Add(numGoroutines)
			for g := 0; g < numGoroutines; g++ {
				go func(id int) {
					defer wg.Done()
					// Multiple atomic operations
					phase := a.Phase()
					energy := a.Energy()
					a.SetPhase(phase + 0.01)
					a.SetEnergy(energy - 0.1)
					a.SetLocalGoal(phase)
				}(g)
			}
			wg.Wait()
		}
	})

	b.Run("grouped_concurrent", func(b *testing.B) {
		a := New("test")
		b.ResetTimer()
		b.ReportAllocs()

		var wg sync.WaitGroup
		for i := 0; i < b.N; i++ {
			wg.Add(numGoroutines)
			for g := 0; g < numGoroutines; g++ {
				go func(id int) {
					defer wg.Done()
					// Single atomic operation for all fields
					a.state.Update(func(s *StateData) {
						s.Phase = s.Phase + 0.01
						s.Energy = s.Energy - 0.1
						s.LocalGoal = s.Phase
					})
				}(g)
			}
			wg.Wait()
		}
	})
}

// BenchmarkUpdateContext tests context update performance.
func BenchmarkUpdateContext(b *testing.B) {
	numNeighbors := 20

	b.Run("standard_context", func(b *testing.B) {
		a := New("test")
		// Add neighbors
		for i := 0; i < numNeighbors; i++ {
			neighbor := New(fmt.Sprintf("neighbor-%d", i))
			a.Neighbors().Store(neighbor.ID, neighbor)
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			a.UpdateContext()
		}
	})

	b.Run("optimized_context", func(b *testing.B) {
		a := New("test")
		// Add neighbors
		for i := 0; i < numNeighbors; i++ {
			neighbor := New(fmt.Sprintf("neighbor-%d", i))
			a.optimizedNeighbors.Store(neighbor.ID, neighbor)
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			a.UpdateContext()
		}
	})
}

// BenchmarkProposeAdjustment tests decision-making performance.
func BenchmarkProposeAdjustment(b *testing.B) {
	goal := core.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.8,
	}

	b.Run("standard_propose", func(b *testing.B) {
		a := New("test")
		// Add some neighbors
		for i := 0; i < 10; i++ {
			neighbor := New(fmt.Sprintf("neighbor-%d", i))
			a.Neighbors().Store(neighbor.ID, neighbor)
		}
		a.UpdateContext()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, _ = a.ProposeAdjustment(goal)
		}
	})

	b.Run("optimized_propose", func(b *testing.B) {
		a := New("test")
		// Add some neighbors
		for i := 0; i < 10; i++ {
			neighbor := New(fmt.Sprintf("neighbor-%d", i))
			a.optimizedNeighbors.Store(neighbor.ID, neighbor)
		}
		a.UpdateContext()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, _ = a.ProposeAdjustment(goal)
		}
	})
}

// BenchmarkMemoryLayout tests memory access patterns.
func BenchmarkMemoryLayout(b *testing.B) {
	numAgents := 1000

	b.Run("standard_memory", func(b *testing.B) {
		agents := make([]*Agent, numAgents)
		for i := 0; i < numAgents; i++ {
			agents[i] = New(fmt.Sprintf("agent-%d", i))
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			// Simulate swarm-wide operation
			totalEnergy := 0.0
			for _, a := range agents {
				totalEnergy += a.Energy()
				a.SetPhase(a.Phase() + 0.001)
			}
		}
	})

	b.Run("grouped_memory", func(b *testing.B) {
		agents := make([]*Agent, numAgents)
		for i := 0; i < numAgents; i++ {
			agents[i] = New(fmt.Sprintf("agent-%d", i))
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			// Simulate swarm-wide operation with grouped atomics
			totalEnergy := 0.0
			for _, a := range agents {
				state := a.state.Load() // One atomic read
				totalEnergy += state.Energy
				a.state.Update(func(s *StateData) {
					s.Phase = s.Phase + 0.001
				}) // One atomic write
			}
		}
	})
}