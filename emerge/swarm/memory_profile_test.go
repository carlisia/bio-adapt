package swarm

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/carlisia/bio-adapt/emerge/agent"
	"github.com/carlisia/bio-adapt/emerge/core"
)

//nolint:paralleltest // Memory profiling should run sequentially
func TestMemoryProfile(t *testing.T) {
	t.Skip("Run with -run TestMemoryProfile to profile memory")

	sizes := []int{10, 50, 100, 500, 1000}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("size_%d", size), func(t *testing.T) {
			// Force GC before measurement
			runtime.GC()
			runtime.GC()

			var m1, m2 runtime.MemStats
			runtime.ReadMemStats(&m1)

			// Create swarm
			swarm, err := New(size, core.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.7,
			})
			if err != nil {
				t.Fatal(err)
			}

			// Run briefly
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			_ = swarm.Run(ctx)

			runtime.ReadMemStats(&m2)

			allocated := m2.Alloc - m1.Alloc
			heapObjects := m2.HeapObjects - m1.HeapObjects

			t.Logf("Size %d: Allocated: %.2f MB, Heap Objects: %d",
				size, float64(allocated)/(1024*1024), heapObjects)

			// Check for reasonable memory usage
			expectedMaxMB := float64(size) * 0.1 // 0.1 MB per agent
			actualMB := float64(allocated) / (1024 * 1024)

			if actualMB > expectedMaxMB {
				t.Logf("Warning: High memory usage: %.2f MB (expected < %.2f MB)",
					actualMB, expectedMaxMB)
			}
		})
	}
}

//nolint:intrange // b.N is not a constant
func BenchmarkAgentMemory(b *testing.B) {
	for i := 0; i < b.N; i++ {
		a := agent.New("test")
		a.SetPhase(1.0)
		a.SetEnergy(100)
		_ = a.Phase()
	}
}
