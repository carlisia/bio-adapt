//nolint:intrange // benchmark loops need explicit indexing
package agent

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/carlisia/bio-adapt/emerge/core"
)

// BenchmarkNeighborStorage compares sync.Map vs optimized storage.
func BenchmarkNeighborStorage(b *testing.B) {
	sizes := []int{10, 20, 50}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("syncmap_%d", size), func(b *testing.B) {
			// Create agents with sync.Map
			agents := make([]*Agent, size)
			for i := 0; i < size; i++ {
				agents[i] = New(fmt.Sprintf("agent-%d", i))
			}

			// Connect all agents
			for i := 0; i < size; i++ {
				for j := i + 1; j < size && j < i+6; j++ { // Connect to ~5 neighbors
					agents[i].ConnectTo(agents[j].ID, agents[j])
				}
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				// Simulate typical neighbor access patterns
				for _, a := range agents {
					// Count neighbors
					_ = a.NeighborCount()

					// Calculate local coherence
					_ = a.calculateLocalCoherence()

					// Update context
					a.UpdateContext()
				}
			}
		})

		b.Run(fmt.Sprintf("optimized_%d", size), func(b *testing.B) {
			// Create optimized agents
			agents := make([]*Agent, size)
			for i := 0; i < size; i++ {
				agents[i] = New(fmt.Sprintf("agent-%d", i))
			}

			// Connect all agents
			for i := 0; i < size; i++ {
				for j := i + 1; j < size && j < i+6; j++ {
					agents[i].ConnectTo(agents[j].ID, agents[j])
				}
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				// Simulate typical neighbor access patterns
				for _, a := range agents {
					// Count neighbors
					_ = a.NeighborCount()

					// Calculate local coherence
					_ = a.calculateLocalCoherence()

					// Update context
					a.UpdateContext()
				}
			}
		})
	}
}

// BenchmarkNeighborIteration tests iteration performance.
func BenchmarkNeighborIteration(b *testing.B) {
	numNeighbors := 20

	b.Run("syncmap", func(b *testing.B) {
		a := New("test")
		// Add neighbors
		for i := 0; i < numNeighbors; i++ {
			neighbor := New(fmt.Sprintf("neighbor-%d", i))
			a.Neighbors().Store(neighbor.ID, neighbor)
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			sum := 0.0
			a.Neighbors().Range(func(_, value any) bool {
				if neighbor, ok := value.(*Agent); ok {
					sum += neighbor.Phase()
				}
				return true
			})
		}
	})

	b.Run("optimized", func(b *testing.B) {
		a := New("test")
		// Add neighbors
		for i := 0; i < numNeighbors; i++ {
			neighbor := New(fmt.Sprintf("neighbor-%d", i))
			a.optimizedNeighbors.Store(neighbor.ID, neighbor)
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			sum := 0.0
			neighbors := a.optimizedNeighbors.All()
			for _, neighbor := range neighbors {
				sum += neighbor.Phase()
			}
		}
	})
}

// BenchmarkCoherenceCalculation tests coherence calculation performance.
func BenchmarkCoherenceCalculation(b *testing.B) {
	numNeighbors := 20

	b.Run("standard", func(b *testing.B) {
		a := New("test")
		for i := 0; i < numNeighbors; i++ {
			neighbor := New(fmt.Sprintf("neighbor-%d", i),
				WithPhase(float64(i)*0.1))
			a.Neighbors().Store(neighbor.ID, neighbor)
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = a.calculateLocalCoherence()
		}
	})

	b.Run("optimized", func(b *testing.B) {
		a := New("test")
		for i := 0; i < numNeighbors; i++ {
			neighbor := New(fmt.Sprintf("neighbor-%d", i),
				WithPhase(float64(i)*0.1))
			a.optimizedNeighbors.Store(neighbor.ID, neighbor)
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = a.calculateLocalCoherence()
		}
	})
}

// BenchmarkConcurrentNeighborAccess tests concurrent access patterns.
func BenchmarkConcurrentNeighborAccess(b *testing.B) {
	numAgents := 100
	numNeighborsPerAgent := 10

	b.Run("syncmap", func(b *testing.B) {
		agents := make([]*Agent, numAgents)
		for i := 0; i < numAgents; i++ {
			agents[i] = New(fmt.Sprintf("agent-%d", i))
		}

		// Connect agents
		for i := 0; i < numAgents; i++ {
			for j := 0; j < numNeighborsPerAgent; j++ {
				neighborIdx := (i + j + 1) % numAgents
				agents[i].ConnectTo(agents[neighborIdx].ID, agents[neighborIdx])
			}
		}

		b.ResetTimer()
		b.ReportAllocs()

		var wg sync.WaitGroup
		for i := 0; i < b.N; i++ {
			wg.Add(numAgents)
			for _, a := range agents {
				go func(agent *Agent) {
					defer wg.Done()
					agent.UpdateContext()
					agent.ProposeAdjustment(core.State{
						Phase:     0,
						Frequency: 100 * time.Millisecond,
						Coherence: 0.8,
					})
				}(a)
			}
			wg.Wait()
		}
	})

	b.Run("optimized", func(b *testing.B) {
		agents := make([]*Agent, numAgents)
		for i := 0; i < numAgents; i++ {
			agents[i] = New(fmt.Sprintf("agent-%d", i))
		}

		// Connect agents
		for i := 0; i < numAgents; i++ {
			for j := 0; j < numNeighborsPerAgent; j++ {
				neighborIdx := (i + j + 1) % numAgents
				agents[i].ConnectTo(agents[neighborIdx].ID, agents[neighborIdx])
			}
		}

		b.ResetTimer()
		b.ReportAllocs()

		var wg sync.WaitGroup
		for i := 0; i < b.N; i++ {
			wg.Add(numAgents)
			for _, a := range agents {
				go func(agent *Agent) {
					defer wg.Done()
					agent.UpdateContext()
					agent.ProposeAdjustment(core.State{
						Phase:     0,
						Frequency: 100 * time.Millisecond,
						Coherence: 0.8,
					})
				}(a)
			}
			wg.Wait()
		}
	})
}
