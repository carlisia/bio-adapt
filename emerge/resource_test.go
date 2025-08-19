package emerge

import (
	"math"
	"sync"
	"testing"
)

func TestTokenResourceManager(t *testing.T) {
	tests := []struct {
		name       string
		maxTokens  float64
		operations []struct {
			op       string // "request" or "release"
			amount   float64
			expected float64 // expected result of operation
		}
		finalAvailable float64
		description    string
	}{
		// Happy path cases
		{
			name:      "basic request and release",
			maxTokens: 100,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", 30, 30},
				{"request", 20, 20},
				{"release", 25, 25},
				{"request", 40, 40},
			},
			finalAvailable: 35,
			description:    "Basic request and release operations",
		},
		{
			name:      "request more than available",
			maxTokens: 100,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", 60, 60},
				{"request", 50, 40}, // Only 40 left
			},
			finalAvailable: 0,
			description:    "Should allocate only what's available",
		},
		{
			name:      "release more than max",
			maxTokens: 50,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", 50, 50},
				{"release", 100, 100}, // Try to release 100
			},
			finalAvailable: 50, // Should cap at max
			description:    "Should cap at maxTokens when releasing",
		},
		{
			name:      "sequential deplete and restore",
			maxTokens: 100,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", 100, 100},
				{"request", 10, 0},
				{"release", 50, 50},
				{"request", 30, 30},
				{"release", 80, 80},
			},
			finalAvailable: 100,
			description:    "Full depletion and restoration",
		},
		// Edge cases
		{
			name:      "zero operations",
			maxTokens: 100,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", 0, 0},
				{"release", 0, 0},
				{"request", 0, 0},
			},
			finalAvailable: 100,
			description:    "Zero operations should not change state",
		},
		{
			name:      "negative request",
			maxTokens: 100,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", -10, 0},
				{"request", -100, 0},
			},
			finalAvailable: 100,
			description:    "Negative requests should be treated as 0",
		},
		{
			name:      "negative release",
			maxTokens: 100,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", 50, 50},
				{"release", -10, -10}, // Implementation specific
			},
			finalAvailable: 50, // Depends on implementation
			description:    "Negative release behavior",
		},
		{
			name:      "empty pool requests",
			maxTokens: 100,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", 100, 100},
				{"request", 1, 0},
				{"request", 10, 0},
				{"request", 100, 0},
			},
			finalAvailable: 0,
			description:    "Requests from empty pool should return 0",
		},
		{
			name:      "fractional values",
			maxTokens: 100,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", 10.5, 10.5},
				{"request", 20.7, 20.7},
				{"release", 5.3, 5.3},
			},
			finalAvailable: 74.1,
			description:    "Should handle fractional values",
		},
		{
			name:      "very small values",
			maxTokens: 100,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", 0.001, 0.001},
				{"request", 0.002, 0.002},
				{"release", 0.001, 0.001},
			},
			finalAvailable: 99.998,
			description:    "Should handle very small values",
		},
		{
			name:      "very large values",
			maxTokens: 1000000,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", 999999, 999999},
				{"request", 10, 1},
				{"release", 500000, 500000},
			},
			finalAvailable: 500000,
			description:    "Should handle large token pools",
		},
		{
			name:      "zero max tokens",
			maxTokens: 0,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", 10, 0},
				{"release", 10, 10},
			},
			finalAvailable: 0, // Should cap at max (0)
			description:    "Zero max tokens edge case",
		},
		{
			name:      "negative max tokens",
			maxTokens: -100,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", 10, 0}, // Can't request from negative pool
			},
			finalAvailable: -100,
			description:    "Negative max tokens edge case",
		},
		{
			name:      "infinity values",
			maxTokens: 100,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", math.Inf(1), 100}, // Request infinity gets all
				{"release", math.Inf(1), math.Inf(1)},
			},
			finalAvailable: 100, // Should cap at max
			description:    "Infinity value handling",
		},
		{
			name:      "NaN values",
			maxTokens: 100,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", math.NaN(), 0}, // NaN likely treated as 0 or error
				{"release", math.NaN(), 0},
			},
			finalAvailable: 100,
			description:    "NaN value handling",
		},
		{
			name:      "mixed invalid and valid operations",
			maxTokens: 100,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", 50, 50},
				{"request", math.Inf(1), 50}, // Request infinity gets remaining
				{"release", math.NaN(), 0},   // NaN release ignored
				{"request", -10, 0},          // Negative request ignored
				{"release", 25, 25},          // Valid release
			},
			finalAvailable: 25,
			description:    "Mixed invalid and valid operations",
		},
		{
			name:      "stress test edge values",
			maxTokens: 1,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", 0.000001, 0.000001},     // Very small request
				{"release", 0.000001, 0.000001},     // Very small release
				{"request", 1000000000, 0.999999},   // Huge request on tiny pool
				{"release", 1000000000, 1000000000}, // Huge release (should cap)
			},
			finalAvailable: 1,
			description:    "Stress test with edge values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := NewTokenResourceManager(tt.maxTokens)

			// Check initial state
			if math.IsNaN(tt.maxTokens) || tt.maxTokens < 0 {
				// Skip initial check for invalid max tokens
			} else if rm.Available() != tt.maxTokens {
				t.Errorf("%s: Initial tokens = %f, expected %f",
					tt.description, rm.Available(), tt.maxTokens)
			}

			// Execute operations
			for i, op := range tt.operations {
				var result float64
				switch op.op {
				case "request":
					result = rm.Request(op.amount)
				case "release":
					rm.Release(op.amount)
					// For release, we don't check the return value in original impl
					continue
				}

				// Skip NaN comparisons
				if math.IsNaN(op.expected) && math.IsNaN(result) {
					continue
				}

				if math.Abs(result-op.expected) > 0.001 {
					t.Errorf("%s: Operation %d (%s %f) = %f, expected %f",
						tt.description, i, op.op, op.amount, result, op.expected)
				}
			}

			// Check final state
			if math.IsNaN(tt.finalAvailable) && math.IsNaN(rm.Available()) {
				return
			}
			if math.Abs(rm.Available()-tt.finalAvailable) > 0.001 {
				t.Errorf("%s: Final available = %f, expected %f",
					tt.description, rm.Available(), tt.finalAvailable)
			}
		})
	}
}

func TestTokenResourceManagerConcurrency(t *testing.T) {
	tests := []struct {
		name                 string
		maxTokens            float64
		numGoroutines        int
		requestsPerGoroutine int
		requestAmount        float64
		releaseRatio         float64 // What fraction to release
		description          string
	}{
		{
			name:                 "basic concurrent operations",
			maxTokens:            1000,
			numGoroutines:        100,
			requestsPerGoroutine: 10,
			requestAmount:        5,
			releaseRatio:         0.5,
			description:          "Basic concurrent request/release",
		},
		{
			name:                 "high contention",
			maxTokens:            100,
			numGoroutines:        1000,
			requestsPerGoroutine: 5,
			requestAmount:        10,
			releaseRatio:         0.8,
			description:          "High contention with many goroutines",
		},
		{
			name:                 "full depletion scenario",
			maxTokens:            100,
			numGoroutines:        10,
			requestsPerGoroutine: 20,
			requestAmount:        20,
			releaseRatio:         0.1,
			description:          "Scenario causing frequent depletion",
		},
		{
			name:                 "balanced operations",
			maxTokens:            1000,
			numGoroutines:        50,
			requestsPerGoroutine: 50,
			requestAmount:        10,
			releaseRatio:         1.0,
			description:          "Balanced request and release",
		},
		{
			name:                 "small pool high demand",
			maxTokens:            10,
			numGoroutines:        100,
			requestsPerGoroutine: 10,
			requestAmount:        5,
			releaseRatio:         0.9,
			description:          "Small pool with high demand",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := NewTokenResourceManager(tt.maxTokens)

			var wg sync.WaitGroup
			var totalRequested, totalReleased float64
			var mu sync.Mutex

			for i := 0; i < tt.numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					for j := 0; j < tt.requestsPerGoroutine; j++ {
						// Request
						requested := rm.Request(tt.requestAmount)
						mu.Lock()
						totalRequested += requested
						mu.Unlock()

						// Release portion
						released := requested * tt.releaseRatio
						rm.Release(released)
						mu.Lock()
						totalReleased += released
						mu.Unlock()
					}
				}()
			}

			wg.Wait()

			// Verify consistency
			expected := tt.maxTokens - totalRequested + totalReleased
			actual := rm.Available()

			// Allow small tolerance for floating point
			if math.Abs(expected-actual) > 0.01 {
				t.Errorf("%s: Expected %f, got %f (requested: %f, released: %f)",
					tt.description, expected, actual, totalRequested, totalReleased)
			}

			// Verify no tokens were created or destroyed
			if actual > tt.maxTokens {
				t.Errorf("%s: Available (%f) exceeds max (%f)",
					tt.description, actual, tt.maxTokens)
			}
		})
	}
}

func TestTokenResourceManagerStressPatterns(t *testing.T) {
	tests := []struct {
		name        string
		pattern     func(rm ResourceManager)
		maxTokens   float64
		validateFn  func(t *testing.T, rm ResourceManager)
		description string
	}{
		{
			name: "rapid request-release cycle",
			pattern: func(rm ResourceManager) {
				for range 1000 {
					allocated := rm.Request(1)
					rm.Release(allocated)
				}
			},
			maxTokens: 100,
			validateFn: func(t *testing.T, rm ResourceManager) {
				if rm.Available() != 100 {
					t.Errorf("Should return to initial state, got %f", rm.Available())
				}
			},
			description: "Rapid cycling should maintain consistency",
		},
		{
			name: "pyramid pattern",
			pattern: func(rm ResourceManager) {
				// Gradually increase requests
				for i := 1; i <= 10; i++ {
					rm.Request(float64(i))
				}
				// Then release all at once
				rm.Release(55) // Sum of 1..10
			},
			maxTokens: 100,
			validateFn: func(t *testing.T, rm ResourceManager) {
				if rm.Available() != 100 {
					t.Errorf("Pyramid pattern failed, available: %f", rm.Available())
				}
			},
			description: "Pyramid request pattern",
		},
		{
			name: "sawtooth pattern",
			pattern: func(rm ResourceManager) {
				for range 5 {
					rm.Request(100) // Deplete
					rm.Release(100) // Restore
				}
			},
			maxTokens: 100,
			validateFn: func(t *testing.T, rm ResourceManager) {
				if rm.Available() != 100 {
					t.Errorf("Sawtooth pattern failed, available: %f", rm.Available())
				}
			},
			description: "Sawtooth depletion/restoration",
		},
		{
			name: "fibonacci requests",
			pattern: func(rm ResourceManager) {
				fib := []float64{1, 1, 2, 3, 5, 8, 13, 21}
				for _, n := range fib {
					rm.Request(n)
				}
			},
			maxTokens: 100,
			validateFn: func(t *testing.T, rm ResourceManager) {
				// Sum of fibonacci numbers above is 54
				if rm.Available() != 46 {
					t.Errorf("Fibonacci pattern failed, available: %f", rm.Available())
				}
			},
			description: "Fibonacci sequence requests",
		},
		{
			name: "alternating small large",
			pattern: func(rm ResourceManager) {
				for i := range 10 {
					if i%2 == 0 {
						rm.Request(1)
					} else {
						rm.Request(10)
					}
				}
			},
			maxTokens: 100,
			validateFn: func(t *testing.T, rm ResourceManager) {
				// 5*1 + 5*10 = 55
				if rm.Available() != 45 {
					t.Errorf("Alternating pattern failed, available: %f", rm.Available())
				}
			},
			description: "Alternating small and large requests",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := NewTokenResourceManager(tt.maxTokens)
			tt.pattern(rm)
			tt.validateFn(t, rm)
		})
	}
}

func TestTokenResourceManagerBoundaryConditions(t *testing.T) {
	tests := []struct {
		name        string
		maxTokens   float64
		validateFn  func(t *testing.T, rm ResourceManager)
		description string
	}{
		{
			name:      "minimum possible value",
			maxTokens: math.SmallestNonzeroFloat64,
			validateFn: func(t *testing.T, rm ResourceManager) {
				allocated := rm.Request(1)
				if allocated != math.SmallestNonzeroFloat64 {
					t.Errorf("Should allocate tiny amount, got %f", allocated)
				}
			},
			description: "Handle smallest float64 value",
		},
		{
			name:      "maximum possible value",
			maxTokens: math.MaxFloat64,
			validateFn: func(t *testing.T, rm ResourceManager) {
				allocated := rm.Request(math.MaxFloat64)
				if allocated != math.MaxFloat64 {
					t.Errorf("Should handle max float64, got %f", allocated)
				}
			},
			description: "Handle maximum float64 value",
		},
		{
			name:      "precision edge case",
			maxTokens: 1.0,
			validateFn: func(t *testing.T, rm ResourceManager) {
				// Request in small increments
				for range 10 {
					rm.Request(0.1)
				}
				// Due to floating point precision, might not be exactly 0
				if rm.Available() > 0.0001 {
					t.Errorf("Precision issue: %f remaining", rm.Available())
				}
			},
			description: "Floating point precision handling",
		},
		{
			name:      "alternating infinity",
			maxTokens: 100,
			validateFn: func(t *testing.T, rm ResourceManager) {
				rm.Request(math.Inf(1))
				if rm.Available() != 0 {
					t.Errorf("Infinity request should deplete, got %f", rm.Available())
				}
				rm.Release(math.Inf(1))
				if rm.Available() != 100 {
					t.Errorf("Infinity release should restore to max, got %f", rm.Available())
				}
			},
			description: "Infinity request and release",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := NewTokenResourceManager(tt.maxTokens)
			tt.validateFn(t, rm)
		})
	}
}

func BenchmarkTokenResourceManager(b *testing.B) {
	benchmarks := []struct {
		name      string
		maxTokens float64
		operation func(rm ResourceManager)
	}{
		{
			name:      "small_pool_request",
			maxTokens: 100,
			operation: func(rm ResourceManager) {
				rm.Request(10)
			},
		},
		{
			name:      "large_pool_request",
			maxTokens: 1000000,
			operation: func(rm ResourceManager) {
				rm.Request(10)
			},
		},
		{
			name:      "request_release_cycle",
			maxTokens: 1000,
			operation: func(rm ResourceManager) {
				allocated := rm.Request(10)
				rm.Release(allocated)
			},
		},
		{
			name:      "depleted_pool_request",
			maxTokens: 10,
			operation: func(rm ResourceManager) {
				rm.Request(100) // Will get only what's available
			},
		},
		{
			name:      "fractional_operations",
			maxTokens: 1000,
			operation: func(rm ResourceManager) {
				rm.Request(0.1)
				rm.Release(0.05)
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			rm := NewTokenResourceManager(bm.maxTokens)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bm.operation(rm)
			}
		})
	}
}

func BenchmarkTokenResourceManagerConcurrent(b *testing.B) {
	benchmarks := []struct {
		name      string
		maxTokens float64
		request   float64
		release   float64
	}{
		{"balanced", 1000000, 10, 10},
		{"net_depletion", 1000000, 10, 5},
		{"net_accumulation", 1000000, 5, 10},
		{"high_contention", 100, 50, 50},
		{"low_contention", 1000000, 1, 1},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			rm := NewTokenResourceManager(bm.maxTokens)

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					allocated := rm.Request(bm.request)
					if allocated > 0 {
						rm.Release(bm.release)
					}
				}
			})
		})
	}
}
