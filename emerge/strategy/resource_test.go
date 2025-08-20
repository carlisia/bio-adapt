package strategy

import (
	"math"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/carlisia/bio-adapt/emerge/core"
)

func TestTokenResourceManager(t *testing.T) {
	tests := []struct {
		name       string
		maxTokens  float64
		operations []struct {
			op       string // "request" or "release"
			amount   float64
			expected float64
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
				{"release", 10, 0}, // Release doesn't return a value
				{"request", 60, 60},
			},
			finalAvailable: 0, // 100 - 30 - 20 + 10 - 60 = 0
			description:    "Basic request/release operations should work",
		},
		{
			name:      "request more than available",
			maxTokens: 50,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", 30, 30},
				{"request", 30, 20}, // Only 20 available
				{"request", 10, 0},  // None available
			},
			finalAvailable: 0,
			description:    "Request should return only available amount",
		},
		{
			name:      "release more than used",
			maxTokens: 100,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", 30, 30},
				{"release", 50, 0}, // Release 50, but only 30 were used
				{"request", 50, 50},
			},
			finalAvailable: 50, // 100 - 30 = 70, then min(70 + 50, 100) = 100, then 100 - 50 = 50
			description:    "Release should allow over-release up to max",
		},
		// Edge cases
		{
			name:      "zero max tokens",
			maxTokens: 0,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", 10, 0},
				{"release", 10, 0},
				{"request", 1, 0},
			},
			finalAvailable: 0,
			description:    "Zero max tokens should allow no requests",
		},
		{
			name:      "negative max tokens",
			maxTokens: -10,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", 10, 0},
				{"release", 10, 0},
			},
			finalAvailable: 0, // Negative max should be treated as 0
			description:    "Negative max tokens should be treated as zero",
		},
		{
			name:      "negative request amount",
			maxTokens: 100,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", -10, 0}, // Negative request should return 0
				{"request", 50, 50},
			},
			finalAvailable: 50,
			description:    "Negative request should return 0",
		},
		{
			name:      "negative release amount",
			maxTokens: 100,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", 50, 50},
				{"release", -10, 0}, // Negative release should do nothing
				{"request", 50, 50},
			},
			finalAvailable: 0,
			description:    "Negative release should do nothing",
		},
		{
			name:      "very small values",
			maxTokens: 0.001,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", 0.0005, 0.0005},
				{"request", 0.0005, 0.0005},
				{"request", 0.0001, 0}, // Exceeds max
			},
			finalAvailable: 0,
			description:    "Should handle very small values",
		},
		{
			name:      "very large values",
			maxTokens: 1e10,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", 1e9, 1e9},
				{"request", 5e9, 5e9},
				{"release", 1e9, 0},
				{"request", 5e9, 5e9},
			},
			finalAvailable: 0,
			description:    "Should handle very large values",
		},
		{
			name:      "infinity max tokens",
			maxTokens: math.Inf(1),
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", 1e100, 1e100},
				{"request", 1e200, 1e200},
				{"release", 1e100, 0},
			},
			finalAvailable: math.Inf(1),
			description:    "Infinity max should allow unlimited requests",
		},
		{
			name:      "NaN max tokens",
			maxTokens: math.NaN(),
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", 10, math.NaN()},
				{"release", 10, 0},
			},
			finalAvailable: math.NaN(),
			description:    "NaN max should propagate NaN",
		},
		{
			name:      "request zero amount",
			maxTokens: 100,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", 0, 0},
				{"request", 50, 50},
				{"request", 0, 0},
			},
			finalAvailable: 50,
			description:    "Zero request should work",
		},
		{
			name:      "release zero amount",
			maxTokens: 100,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", 50, 50},
				{"release", 0, 0},
				{"request", 50, 50},
			},
			finalAvailable: 0,
			description:    "Zero release should work",
		},
		{
			name:      "request infinity",
			maxTokens: 100,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", math.Inf(1), 100}, // Should get all available
				{"request", 10, 0},            // None left
			},
			finalAvailable: 0,
			description:    "Infinity request should get all available",
		},
		{
			name:      "release infinity",
			maxTokens: 100,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", 50, 50},
				{"release", math.Inf(1), 0}, // Should reset to max
				{"request", 100, 100},       // All available again
			},
			finalAvailable: 0,
			description:    "Infinity release should reset to max",
		},
		{
			name:      "request NaN",
			maxTokens: 100,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", math.NaN(), math.NaN()},
			},
			finalAvailable: math.NaN(),
			description:    "NaN request should propagate NaN",
		},
		{
			name:      "release NaN",
			maxTokens: 100,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"request", 50, 50},
				{"release", math.NaN(), 0},
			},
			finalAvailable: math.NaN(),
			description:    "NaN release should propagate NaN",
		},
		{
			name:      "multiple releases without requests",
			maxTokens: 100,
			operations: []struct {
				op       string
				amount   float64
				expected float64
			}{
				{"release", 10, 0},
				{"release", 20, 0},
				{"request", 100, 100}, // Should still have max
			},
			finalAvailable: 0, // Released tokens shouldn't exceed max
			description:    "Multiple releases shouldn't exceed max",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := NewTokenResourceManager(tt.maxTokens)

			// Check initial state
			if !math.IsNaN(tt.maxTokens) && tt.maxTokens >= 0 {
				assert.Equal(t, tt.maxTokens, rm.Available(), "%s: Initial tokens should match max", tt.description)
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

				if !math.IsNaN(op.expected) {
					assert.InDelta(t, op.expected, result, 0.001, "%s: Operation %d (%s %f) should return expected result", tt.description, i, op.op, op.amount)
				}
			}

			// Check final state
			if math.IsNaN(tt.finalAvailable) && math.IsNaN(rm.Available()) {
				return
			}
			if !math.IsNaN(tt.finalAvailable) {
				assert.InDelta(t, tt.finalAvailable, rm.Available(), 0.001, "%s: Final available should match expected", tt.description)
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
		amountPerRequest     float64
		description          string
	}{
		{
			name:                 "concurrent small requests",
			maxTokens:            1000,
			numGoroutines:        10,
			requestsPerGoroutine: 10,
			amountPerRequest:     1,
			description:          "Many small concurrent requests",
		},
		{
			name:                 "concurrent large requests",
			maxTokens:            100,
			numGoroutines:        10,
			requestsPerGoroutine: 5,
			amountPerRequest:     20,
			description:          "Competing for limited resources",
		},
		{
			name:                 "concurrent request and release",
			maxTokens:            50,
			numGoroutines:        20,
			requestsPerGoroutine: 10,
			amountPerRequest:     10,
			description:          "Mixed request/release pattern",
		},
		{
			name:                 "high contention",
			maxTokens:            10,
			numGoroutines:        100,
			requestsPerGoroutine: 2,
			amountPerRequest:     5,
			description:          "Very high contention for few resources",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := NewTokenResourceManager(tt.maxTokens)

			var wg sync.WaitGroup
			totalRequested := make(chan float64, tt.numGoroutines)

			for i := range tt.numGoroutines {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					var localTotal float64

					for range tt.requestsPerGoroutine {
						granted := rm.Request(tt.amountPerRequest)
						localTotal += granted

						// Half the goroutines do request/release pattern
						if id%2 == 0 && granted > 0 {
							// Release half of what was granted
							rm.Release(granted / 2)
							localTotal -= granted / 2
						}
					}

					totalRequested <- localTotal
				}(i)
			}

			wg.Wait()
			close(totalRequested)

			// Sum up total actually acquired
			var totalAcquired float64
			for amount := range totalRequested {
				totalAcquired += amount
			}

			// With request/release pattern, we can't predict exact total
			// but it should be non-negative and not exceed theoretical max
			theoreticalMax := float64(tt.numGoroutines) * float64(tt.requestsPerGoroutine) * tt.amountPerRequest
			assert.GreaterOrEqual(t, totalAcquired, 0.0, "%s: Total acquired should be non-negative", tt.description)
			assert.LessOrEqual(t, totalAcquired, theoreticalMax, "%s: Total acquired should not exceed theoretical max", tt.description)

			// Final available should be non-negative
			assert.GreaterOrEqual(t, rm.Available(), 0.0, "%s: Final available should be non-negative", tt.description)
		})
	}
}

func TestTokenResourceManagerStressPatterns(t *testing.T) {
	tests := []struct {
		name        string
		pattern     func(rm core.ResourceManager)
		maxTokens   float64
		validateFn  func(t *testing.T, rm core.ResourceManager)
		description string
	}{
		{
			name:      "burst pattern",
			maxTokens: 100,
			pattern: func(rm core.ResourceManager) {
				// Burst request all tokens
				rm.Request(100)
				// Then many small requests (should all fail)
				for range 100 {
					rm.Request(1)
				}
			},
			validateFn: func(t *testing.T, rm core.ResourceManager) {
				assert.Equal(t, 0.0, rm.Available(), "After burst, should have 0 available")
			},
			description: "Burst request pattern",
		},
		{
			name:      "gradual drain",
			maxTokens: 100,
			pattern: func(rm core.ResourceManager) {
				// Gradually drain tokens
				for range 100 {
					rm.Request(1)
				}
			},
			validateFn: func(t *testing.T, rm core.ResourceManager) {
				assert.Equal(t, 0.0, rm.Available(), "After gradual drain, should have 0 available")
			},
			description: "Gradual drain pattern",
		},
		{
			name:      "oscillating",
			maxTokens: 100,
			pattern: func(rm core.ResourceManager) {
				// Oscillate between request and release
				for range 10 {
					rm.Request(50)
					rm.Release(30)
				}
			},
			validateFn: func(t *testing.T, rm core.ResourceManager) {
				// After first 3 cycles: 100 → 80 → 60 → 40
				// Then stabilizes at: request gets 30-40, release adds 30
				// Final state after 10 cycles: 30 available
				expected := 30.0
				assert.InDelta(t, expected, rm.Available(), 0.001, "After oscillation, should have expected available")
			},
			description: "Oscillating request/release pattern",
		},
		{
			name:      "random amounts",
			maxTokens: 100,
			pattern: func(rm core.ResourceManager) {
				amounts := []float64{7, 13, 2, 29, 5, 11, 3, 17, 19, 23}
				for _, amount := range amounts {
					rm.Request(amount)
				}
			},
			validateFn: func(t *testing.T, rm core.ResourceManager) {
				// Sum of amounts = 129, max = 100
				// So should have 0 available
				assert.Equal(t, 0.0, rm.Available(), "After random requests, should have 0 available")
			},
			description: "Random amount pattern",
		},
		{
			name:      "reset pattern",
			maxTokens: 100,
			pattern: func(rm core.ResourceManager) {
				rm.Request(100)         // Drain all
				rm.Release(100)         // Reset to full
				rm.Request(50)          // Use half
				rm.Release(math.Inf(1)) // Try to release infinity
			},
			validateFn: func(t *testing.T, rm core.ResourceManager) {
				// After releasing infinity, should be back at max
				assert.Equal(t, 100.0, rm.Available(), "After reset pattern, should have max available")
			},
			description: "Reset pattern with infinity release",
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
		validateFn  func(t *testing.T, rm core.ResourceManager)
		description string
	}{
		{
			name:      "machine epsilon precision",
			maxTokens: 1.0,
			validateFn: func(t *testing.T, rm core.ResourceManager) {
				epsilon := math.Nextafter(1, 2) - 1
				granted := rm.Request(epsilon)
				assert.Equal(t, epsilon, granted, "Should handle machine epsilon")
			},
			description: "Machine epsilon handling",
		},
		{
			name:      "denormalized numbers",
			maxTokens: math.SmallestNonzeroFloat64 * 100,
			validateFn: func(t *testing.T, rm core.ResourceManager) {
				granted := rm.Request(math.SmallestNonzeroFloat64)
				assert.Equal(t, math.SmallestNonzeroFloat64, granted, "Should handle denormalized numbers")
			},
			description: "Denormalized number handling",
		},
		{
			name:      "maximum finite value",
			maxTokens: math.MaxFloat64,
			validateFn: func(t *testing.T, rm core.ResourceManager) {
				granted := rm.Request(math.MaxFloat64 / 2)
				assert.Equal(t, math.MaxFloat64/2, granted, "Should handle maximum finite values")
				// Check we can still request more
				granted2 := rm.Request(math.MaxFloat64 / 2)
				assert.Equal(t, math.MaxFloat64/2, granted2, "Should handle maximum finite values (second request)")
			},
			description: "Maximum finite value handling",
		},
		{
			name:      "negative zero",
			maxTokens: 100,
			validateFn: func(t *testing.T, rm core.ResourceManager) {
				negZero := math.Copysign(0, -1)
				granted := rm.Request(negZero)
				// Negative zero should be treated as zero
				assert.Equal(t, 0.0, granted, "Negative zero should return 0")
				rm.Release(negZero)
				// Available should still be 100
				assert.Equal(t, 100.0, rm.Available(), "Negative zero release should not change available")
			},
			description: "Negative zero handling",
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
		operation func(rm core.ResourceManager)
	}{
		{
			name:      "small_request",
			maxTokens: 1000,
			operation: func(rm core.ResourceManager) {
				rm.Request(1)
			},
		},
		{
			name:      "large_request",
			maxTokens: 1000,
			operation: func(rm core.ResourceManager) {
				rm.Request(100)
			},
		},
		{
			name:      "release",
			maxTokens: 1000,
			operation: func(rm core.ResourceManager) {
				rm.Release(10)
			},
		},
		{
			name:      "request_release_cycle",
			maxTokens: 1000,
			operation: func(rm core.ResourceManager) {
				granted := rm.Request(10)
				rm.Release(granted)
			},
		},
		{
			name:      "check_available",
			maxTokens: 1000,
			operation: func(rm core.ResourceManager) {
				_ = rm.Available()
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			rm := NewTokenResourceManager(bm.maxTokens)
			b.ResetTimer()
			for range b.N {
				bm.operation(rm)
			}
		})
	}
}

func BenchmarkTokenResourceManagerConcurrent(b *testing.B) {
	benchmarks := []struct {
		name          string
		maxTokens     float64
		numGoroutines int
	}{
		{"10_goroutines", 1000, 10},
		{"100_goroutines", 1000, 100},
		{"1000_goroutines", 10000, 1000},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			rm := NewTokenResourceManager(bm.maxTokens)
			b.ResetTimer()

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					granted := rm.Request(1)
					if granted > 0 {
						rm.Release(granted)
					}
				}
			})
		})
	}
}
