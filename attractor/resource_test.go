package attractor

import (
	"sync"
	"testing"
)

func TestTokenResourceManager(t *testing.T) {
	rm := NewTokenResourceManager(100)

	// Check initial state
	if rm.Available() != 100 {
		t.Errorf("Expected initial tokens 100, got %f", rm.Available())
	}

	// Test requesting resources
	allocated := rm.Request(30)
	if allocated != 30 {
		t.Errorf("Expected to allocate 30, got %f", allocated)
	}
	if rm.Available() != 70 {
		t.Errorf("Expected 70 remaining, got %f", rm.Available())
	}

	// Test requesting more than available
	allocated = rm.Request(100)
	if allocated != 70 {
		t.Errorf("Expected to allocate only 70 available, got %f", allocated)
	}
	if rm.Available() != 0 {
		t.Errorf("Expected 0 remaining, got %f", rm.Available())
	}

	// Test releasing resources
	rm.Release(50)
	if rm.Available() != 50 {
		t.Errorf("Expected 50 after release, got %f", rm.Available())
	}

	// Test releasing more than max (should cap at max)
	rm.Release(100)
	if rm.Available() != 100 {
		t.Errorf("Expected to cap at max 100, got %f", rm.Available())
	}
}

func TestTokenResourceManagerConcurrency(t *testing.T) {
	rm := NewTokenResourceManager(1000)

	var wg sync.WaitGroup
	numGoroutines := 100
	requestsPerGoroutine := 10

	// Track total requested and released
	var totalRequested, totalReleased float64
	var mu sync.Mutex

	for range numGoroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for range requestsPerGoroutine {
				// Request random amount
				requested := rm.Request(5)
				mu.Lock()
				totalRequested += requested
				mu.Unlock()

				// Do some work...

				// Release half back
				released := requested / 2
				rm.Release(released)
				mu.Lock()
				totalReleased += released
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Final available should be initial - requested + released
	expected := 1000 - totalRequested + totalReleased
	actual := rm.Available()

	// Allow small tolerance for floating point
	if abs(expected-actual) > 0.01 {
		t.Errorf("Concurrent operations inconsistent: expected %f, got %f", expected, actual)
	}
}

func TestTokenResourceManagerEdgeCases(t *testing.T) {
	rm := NewTokenResourceManager(100)

	// Test requesting 0
	allocated := rm.Request(0)
	if allocated != 0 {
		t.Errorf("Requesting 0 should allocate 0, got %f", allocated)
	}
	if rm.Available() != 100 {
		t.Error("Requesting 0 should not change available resources")
	}

	// Test releasing 0
	rm.Release(0)
	if rm.Available() != 100 {
		t.Error("Releasing 0 should not change available resources")
	}

	// Test negative request (should treat as 0)
	allocated = rm.Request(-10)
	if allocated != 0 {
		t.Errorf("Negative request should allocate 0, got %f", allocated)
	}

	// Deplete all resources
	rm.Request(100)

	// Test requesting when empty
	allocated = rm.Request(10)
	if allocated != 0 {
		t.Errorf("Should allocate 0 when empty, got %f", allocated)
	}
}

func TestTokenResourceManagerMaxTokens(t *testing.T) {
	rm := NewTokenResourceManager(50)

	// Deplete resources
	rm.Request(50)

	// Try to release more than max
	rm.Release(100)

	// Should cap at maxTokens
	if rm.Available() != 50 {
		t.Errorf("Should cap at maxTokens (50), got %f", rm.Available())
	}
}

func BenchmarkTokenResourceManagerRequest(b *testing.B) {
	rm := NewTokenResourceManager(1000000) // Large pool to avoid depletion

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rm.Request(10)
	}
}

func BenchmarkTokenResourceManagerRelease(b *testing.B) {
	rm := NewTokenResourceManager(1000000)
	rm.Request(500000) // Deplete half

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rm.Release(10)
	}
}

func BenchmarkTokenResourceManagerConcurrent(b *testing.B) {
	rm := NewTokenResourceManager(1000000)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			allocated := rm.Request(10)
			rm.Release(allocated)
		}
	})
}

// Helper function for absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
