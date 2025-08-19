package emerge_test

import (
	"math"
	"testing"
	"time"

	"github.com/carlisia/bio-adapt/emerge"
)

func TestNewAttractorBasin(t *testing.T) {
	tests := []struct {
		name         string
		target       emerge.State
		strength     float64
		radius       float64
		validateFn   func(t *testing.T, basin *emerge.AttractorBasin)
	}{
		{
			name: "basic basin creation",
			target: emerge.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			strength: 0.5,
			radius:   10.0,
			validateFn: func(t *testing.T, basin *emerge.AttractorBasin) {
				if basin == nil {
					t.Fatal("Expected basin to be created")
				}
			},
		},
		{
			name: "zero strength basin",
			target: emerge.State{
				Phase:     0,
				Frequency: 200 * time.Millisecond,
				Coherence: 0.9,
			},
			strength: 0,
			radius:   1.0,
			validateFn: func(t *testing.T, basin *emerge.AttractorBasin) {
				// Zero strength basin should have no attraction force
				force := basin.AttractionForce(emerge.State{
					Phase:     0.1,
					Frequency: 200 * time.Millisecond,
					Coherence: 0.9,
				})
				if force != 0 {
					t.Errorf("Zero strength basin should have no force, got %f", force)
				}
			},
		},
		{
			name: "negative strength basin (invalid)",
			target: emerge.State{
				Phase:     math.Pi / 2,
				Frequency: 50 * time.Millisecond,
				Coherence: 0.7,
			},
			strength: -0.5,
			radius:   2.0,
			validateFn: func(t *testing.T, basin *emerge.AttractorBasin) {
				// Implementation specific - negative strength might be clamped or allowed
				if basin == nil {
					t.Fatal("Basin should be created even with negative strength")
				}
			},
		},
		{
			name: "zero radius basin",
			target: emerge.State{
				Phase:     1.5,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.85,
			},
			strength: 0.8,
			radius:   0,
			validateFn: func(t *testing.T, basin *emerge.AttractorBasin) {
				// Zero radius means only the exact target is in basin
				target := emerge.State{
					Phase:     1.5,
					Frequency: 100 * time.Millisecond,
					Coherence: 0.85,
				}
				if !basin.IsInBasin(target) {
					t.Error("Target should be in zero-radius basin")
				}
				// Even slightly off target should be outside
				offTarget := emerge.State{
					Phase:     1.50001,
					Frequency: 100 * time.Millisecond,
					Coherence: 0.85,
				}
				if basin.IsInBasin(offTarget) {
					t.Error("Off-target should not be in zero-radius basin")
				}
			},
		},
		{
			name: "negative radius basin (invalid)",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.5,
			},
			strength: 0.5,
			radius:   -1.0,
			validateFn: func(t *testing.T, basin *emerge.AttractorBasin) {
				// Negative radius might be treated as 0 or absolute value
				if basin == nil {
					t.Fatal("Basin should be created even with negative radius")
				}
			},
		},
		{
			name: "very large radius basin",
			target: emerge.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.99,
			},
			strength: 1.0,
			radius:   100 * math.Pi,
			validateFn: func(t *testing.T, basin *emerge.AttractorBasin) {
				// Everything should be in a very large basin
				anyState := emerge.State{
					Phase:     math.Pi * 0.5,
					Frequency: 100 * time.Millisecond,
					Coherence: 0.99,
				}
				if !basin.IsInBasin(anyState) {
					t.Error("Any state should be in very large basin")
				}
			},
		},
		{
			name: "maximum strength basin",
			target: emerge.State{
				Phase:     2.0,
				Frequency: 150 * time.Millisecond,
				Coherence: 1.0,
			},
			strength: 1.0,
			radius:   math.Pi / 2,
			validateFn: func(t *testing.T, basin *emerge.AttractorBasin) {
				force := basin.AttractionForce(emerge.State{
					Phase:     2.0,
					Frequency: 150 * time.Millisecond,
					Coherence: 1.0,
				})
				if force < 0.9 || force > 1.0 {
					t.Errorf("Max strength at target should give max force, got %f", force)
				}
			},
		},
		{
			name: "phase at boundary conditions",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			strength: 0.5,
			radius:   math.Pi,
			validateFn: func(t *testing.T, basin *emerge.AttractorBasin) {
				// Test phase 0 target
				dist := basin.DistanceToTarget(emerge.State{
					Phase:     2 * math.Pi,
					Frequency: 100 * time.Millisecond,
					Coherence: 0.8,
				})
				if dist > 0.01 {
					t.Errorf("Phase 2π should be same as 0, distance = %f", dist)
				}
			},
		},
		{
			name: "target state validation",
			target: emerge.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			strength: 0.5,
			radius:   10.0,
			validateFn: func(t *testing.T, basin *emerge.AttractorBasin) {
				// Target should always be in its own basin
				target := emerge.State{
					Phase:     math.Pi,
					Frequency: 100 * time.Millisecond,
					Coherence: 0.8,
				}
				if !basin.IsInBasin(target) {
					t.Error("Target state should be in its own basin")
				}
				// Distance to target should be 0
				if dist := basin.DistanceToTarget(target); dist != 0 {
					t.Errorf("Distance from target to itself should be 0, got %f", dist)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basin := emerge.NewAttractorBasin(tt.target, tt.strength, tt.radius)
			tt.validateFn(t, basin)
		})
	}
}

func TestBasinDistanceToTarget(t *testing.T) {
	tests := []struct {
		name     string
		target   emerge.State
		state    emerge.State
		expected float64
		tolerance float64
	}{
		// Happy path cases
		{
			name: "state at target",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			state: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			expected: 0,
			tolerance: 0.01,
		},
		{
			name: "state at π/4 from 0",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			state: emerge.State{
				Phase:     math.Pi / 4,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			expected: math.Pi / 4,
			tolerance: 0.01,
		},
		{
			name: "state at π from 0",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			state: emerge.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			expected: math.Pi,
			tolerance: 0.01,
		},
		// Edge cases with phase wrapping
		{
			name: "state at 3π/2 from 0 (shortest path)",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			state: emerge.State{
				Phase:     3 * math.Pi / 2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			expected: math.Pi / 2,
			tolerance: 0.01,
		},
		{
			name: "state at 2π from 0 (wraps to 0)",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			state: emerge.State{
				Phase:     2 * math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			expected: 0,
			tolerance: 0.01,
		},
		// Negative phase values
		{
			name: "negative phase to positive target",
			target: emerge.State{
				Phase:     math.Pi / 2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			state: emerge.State{
				Phase:     -math.Pi / 2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			expected: math.Pi,
			tolerance: 0.01,
		},
		// Very small differences
		{
			name: "very small phase difference",
			target: emerge.State{
				Phase:     1.0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			state: emerge.State{
				Phase:     1.0001,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			expected: 0.0001,
			tolerance: 0.00001,
		},
		// Large phase values
		{
			name: "large phase values",
			target: emerge.State{
				Phase:     100 * math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			state: emerge.State{
				Phase:     100.5 * math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			expected: math.Pi / 2,
			tolerance: 0.01,
		},
		// Target at π
		{
			name: "target at π, state at 0",
			target: emerge.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			state: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			expected: math.Pi,
			tolerance: 0.01,
		},
		{
			name: "target at π, state at 2π",
			target: emerge.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			state: emerge.State{
				Phase:     2 * math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			expected: math.Pi,
			tolerance: 0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basin := emerge.NewAttractorBasin(tt.target, 0.5, math.Pi/4)
			dist := basin.DistanceToTarget(tt.state)
			if math.Abs(dist-tt.expected) > tt.tolerance {
				t.Errorf("Expected distance %f±%f, got %f", tt.expected, tt.tolerance, dist)
			}
		})
	}
}

func TestBasinIsInBasin(t *testing.T) {
	tests := []struct {
		name     string
		target   emerge.State
		radius   float64
		strength float64
		state    emerge.State
		inBasin  bool
	}{
		// Happy path cases
		{
			name: "state at target",
			target: emerge.State{
				Phase:     math.Pi / 2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			radius:   math.Pi / 4,
			strength: 0.5,
			state: emerge.State{
				Phase:     math.Pi / 2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			inBasin: true,
		},
		{
			name: "state within radius",
			target: emerge.State{
				Phase:     math.Pi / 2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			radius:   math.Pi / 4,
			strength: 0.5,
			state: emerge.State{
				Phase:     math.Pi/2 + math.Pi/8,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			inBasin: true,
		},
		{
			name: "state at radius boundary",
			target: emerge.State{
				Phase:     math.Pi / 2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			radius:   math.Pi / 4,
			strength: 0.5,
			state: emerge.State{
				Phase:     math.Pi/2 + math.Pi/4,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			inBasin: true,
		},
		{
			name: "state outside radius",
			target: emerge.State{
				Phase:     math.Pi / 2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			radius:   math.Pi / 4,
			strength: 0.5,
			state: emerge.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			inBasin: false,
		},
		// Edge cases
		{
			name: "zero radius basin - at target",
			target: emerge.State{
				Phase:     1.0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			radius:   0,
			strength: 0.5,
			state: emerge.State{
				Phase:     1.0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			inBasin: true,
		},
		{
			name: "zero radius basin - near target",
			target: emerge.State{
				Phase:     1.0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			radius:   0,
			strength: 0.5,
			state: emerge.State{
				Phase:     1.00001,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			inBasin: false,
		},
		{
			name: "very large radius - everything in basin",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			radius:   10 * math.Pi,
			strength: 0.5,
			state: emerge.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			inBasin: true,
		},
		{
			name: "negative phase values",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			radius:   math.Pi / 4,
			strength: 0.5,
			state: emerge.State{
				Phase:     -math.Pi / 8,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			inBasin: true,
		},
		{
			name: "phase wrapping at 2π boundary",
			target: emerge.State{
				Phase:     2*math.Pi - 0.1,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			radius:   0.2,
			strength: 0.5,
			state: emerge.State{
				Phase:     0.05,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			inBasin: true,
		},
		{
			name: "exact radius boundary - floating point",
			target: emerge.State{
				Phase:     1.0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			radius:   0.5,
			strength: 0.5,
			state: emerge.State{
				Phase:     1.5,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			inBasin: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basin := emerge.NewAttractorBasin(tt.target, tt.strength, tt.radius)
			inBasin := basin.IsInBasin(tt.state)
			if inBasin != tt.inBasin {
				t.Errorf("Expected IsInBasin=%v, got %v", tt.inBasin, inBasin)
			}
		})
	}
}

func TestBasinAttractionForce(t *testing.T) {
	tests := []struct {
		name     string
		target   emerge.State
		strength float64
		radius   float64
		state    emerge.State
		minForce float64
		maxForce float64
	}{
		// Happy path cases
		{
			name: "state at target",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			strength: 0.5,
			radius:   math.Pi / 2,
			state: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			minForce: 0.45,
			maxForce: 0.5,
		},
		{
			name: "state at quarter radius",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			strength: 0.5,
			radius:   math.Pi / 2,
			state: emerge.State{
				Phase:     math.Pi / 8,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			minForce: 0.3,
			maxForce: 0.45,
		},
		{
			name: "state at radius boundary",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			strength: 0.5,
			radius:   math.Pi / 2,
			state: emerge.State{
				Phase:     math.Pi / 2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			minForce: 0,
			maxForce: 0.1,
		},
		{
			name: "state outside radius",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			strength: 0.5,
			radius:   math.Pi / 2,
			state: emerge.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			minForce: 0,
			maxForce: 0,
		},
		// Edge cases
		{
			name: "zero strength basin",
			target: emerge.State{
				Phase:     math.Pi / 4,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			strength: 0,
			radius:   math.Pi,
			state: emerge.State{
				Phase:     math.Pi / 2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			minForce: 0,
			maxForce: 0,
		},
		{
			name: "maximum strength basin at target",
			target: emerge.State{
				Phase:     1.5,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			strength: 1.0,
			radius:   math.Pi / 3,
			state: emerge.State{
				Phase:     1.5,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			minForce: 0.95,
			maxForce: 1.0,
		},
		{
			name: "very small radius basin",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			strength: 0.7,
			radius:   0.01,
			state: emerge.State{
				Phase:     0.005,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			minForce: 0.3,
			maxForce: 0.7,
		},
		{
			name: "negative phase with positive target",
			target: emerge.State{
				Phase:     math.Pi / 2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			strength: 0.6,
			radius:   math.Pi,
			state: emerge.State{
				Phase:     -math.Pi / 4,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			minForce: 0.3,
			maxForce: 0.5,
		},
		{
			name: "phase wrapping around 2π",
			target: emerge.State{
				Phase:     2*math.Pi - 0.1,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			strength: 0.8,
			radius:   0.3,
			state: emerge.State{
				Phase:     0.1,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			minForce: 0.2,
			maxForce: 0.6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basin := emerge.NewAttractorBasin(tt.target, tt.strength, tt.radius)
			force := basin.AttractionForce(tt.state)
			if force < tt.minForce || force > tt.maxForce {
				t.Errorf("Expected force in range [%f, %f], got %f",
					tt.minForce, tt.maxForce, force)
			}
		})
	}
}

func TestBasinConvergenceRate(t *testing.T) {
	tests := []struct {
		name     string
		target   emerge.State
		strength float64
		radius   float64
		state    emerge.State
		minRate  float64
		maxRate  float64
	}{
		// Happy path cases
		{
			name: "state at target",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			strength: 0.8,
			radius:   math.Pi / 4,
			state: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			minRate: 0.7,
			maxRate: 0.8,
		},
		{
			name: "state within basin",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			strength: 0.8,
			radius:   math.Pi / 4,
			state: emerge.State{
				Phase:     math.Pi / 8,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			minRate: 0.3,
			maxRate: 0.6,
		},
		{
			name: "state outside basin",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			strength: 0.8,
			radius:   math.Pi / 4,
			state: emerge.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			minRate: 0,
			maxRate: 0,
		},
		// Edge cases
		{
			name: "zero strength basin",
			target: emerge.State{
				Phase:     math.Pi / 2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			strength: 0,
			radius:   math.Pi,
			state: emerge.State{
				Phase:     math.Pi / 2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			minRate: 0,
			maxRate: 0,
		},
		{
			name: "maximum strength basin",
			target: emerge.State{
				Phase:     1.0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			strength: 1.0,
			radius:   math.Pi / 2,
			state: emerge.State{
				Phase:     1.1,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.9,
			},
			minRate: 0.8,
			maxRate: 1.0,
		},
		{
			name: "state at radius boundary",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			strength: 0.5,
			radius:   1.0,
			state: emerge.State{
				Phase:     1.0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			minRate: 0,
			maxRate: 0.1,
		},
		{
			name: "very small radius basin",
			target: emerge.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			strength: 0.6,
			radius:   0.01,
			state: emerge.State{
				Phase:     math.Pi + 0.005,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			minRate: 0.2,
			maxRate: 0.5,
		},
		{
			name: "negative radius treated as positive",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			strength: 0.7,
			radius:   -math.Pi / 4,
			state: emerge.State{
				Phase:     0.1,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			minRate: 0,
			maxRate: 0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basin := emerge.NewAttractorBasin(tt.target, tt.strength, tt.radius)
			rate := basin.ConvergenceRate(tt.state)
			if rate < tt.minRate || rate > tt.maxRate {
				t.Errorf("Expected rate in range [%f, %f], got %f",
					tt.minRate, tt.maxRate, rate)
			}
		})
	}
}

func TestBasinOptimalAdjustment(t *testing.T) {
	tests := []struct {
		name         string
		target       emerge.State
		strength     float64
		radius       float64
		current      emerge.State
		expectSign   float64 // Expected sign of adjustment (-1, 0, or 1)
		minMagnitude float64 // Minimum expected absolute value
		maxMagnitude float64 // Maximum expected absolute value
	}{
		// Happy path cases
		{
			name: "state at target",
			target: emerge.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			strength:     1.0,
			radius:       math.Pi,
			current:      emerge.State{Phase: math.Pi, Frequency: 100 * time.Millisecond, Coherence: 0.8},
			expectSign:   0,
			minMagnitude: 0,
			maxMagnitude: 0.01,
		},
		{
			name: "state before target",
			target: emerge.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			strength:     1.0,
			radius:       math.Pi,
			current:      emerge.State{Phase: math.Pi / 2, Frequency: 100 * time.Millisecond, Coherence: 0.8},
			expectSign:   1,
			minMagnitude: 0.1,
			maxMagnitude: 2.0,
		},
		{
			name: "state after target",
			target: emerge.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			strength:     1.0,
			radius:       math.Pi,
			current:      emerge.State{Phase: 3 * math.Pi / 2, Frequency: 100 * time.Millisecond, Coherence: 0.8},
			expectSign:   -1,
			minMagnitude: 0.1,
			maxMagnitude: 2.0,
		},
		// Edge cases
		{
			name: "state outside basin",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			strength:     0.5,
			radius:       math.Pi / 4,
			current:      emerge.State{Phase: math.Pi, Frequency: 100 * time.Millisecond, Coherence: 0.8},
			expectSign:   -1,
			minMagnitude: 0,
			maxMagnitude: 0.1,
		},
		{
			name: "zero strength basin",
			target: emerge.State{
				Phase:     1.0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			strength:     0,
			radius:       1.0,
			current:      emerge.State{Phase: 2.0, Frequency: 100 * time.Millisecond, Coherence: 0.8},
			expectSign:   -1,
			minMagnitude: 0,
			maxMagnitude: 0.01,
		},
		{
			name: "phase wrapping - shortest path forward",
			target: emerge.State{
				Phase:     0.1,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			strength:     0.8,
			radius:       math.Pi,
			current:      emerge.State{Phase: 2*math.Pi - 0.1, Frequency: 100 * time.Millisecond, Coherence: 0.8},
			expectSign:   1,
			minMagnitude: 0.1,
			maxMagnitude: 0.5,
		},
		{
			name: "phase wrapping - shortest path backward",
			target: emerge.State{
				Phase:     2*math.Pi - 0.1,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			strength:     0.8,
			radius:       math.Pi,
			current:      emerge.State{Phase: 0.1, Frequency: 100 * time.Millisecond, Coherence: 0.8},
			expectSign:   -1,
			minMagnitude: 0.1,
			maxMagnitude: 0.5,
		},
		{
			name: "very small distance",
			target: emerge.State{
				Phase:     1.0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			strength:     0.5,
			radius:       1.0,
			current:      emerge.State{Phase: 1.0001, Frequency: 100 * time.Millisecond, Coherence: 0.8},
			expectSign:   -1,
			minMagnitude: 0,
			maxMagnitude: 0.01,
		},
		{
			name: "maximum strength with large distance",
			target: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			strength:     1.0,
			radius:       2 * math.Pi,
			current:      emerge.State{Phase: math.Pi, Frequency: 100 * time.Millisecond, Coherence: 0.8},
			expectSign:   -1,
			minMagnitude: 0.5,
			maxMagnitude: 3.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basin := emerge.NewAttractorBasin(tt.target, tt.strength, tt.radius)
			adjustment := basin.OptimalAdjustment(tt.current)

			// Check sign
			if tt.expectSign == 0 {
				if math.Abs(adjustment) > tt.maxMagnitude {
					t.Errorf("Expected near-zero adjustment (<=%f), got %f", tt.maxMagnitude, adjustment)
				}
			} else if tt.expectSign > 0 {
				if adjustment <= 0 {
					t.Errorf("Expected positive adjustment, got %f", adjustment)
				}
			} else {
				if adjustment >= 0 {
					t.Errorf("Expected negative adjustment, got %f", adjustment)
				}
			}

			// Check magnitude
			abs := math.Abs(adjustment)
			if abs < tt.minMagnitude || abs > tt.maxMagnitude {
				t.Errorf("Expected magnitude in range [%f, %f], got %f",
					tt.minMagnitude, tt.maxMagnitude, abs)
			}
		})
	}
}

func TestPhaseUtilityFunctions(t *testing.T) {
	t.Run("WrapPhase", func(t *testing.T) {
		tests := []struct {
			name     string
			input    float64
			expected float64
		}{
			// Happy path cases
			{name: "zero phase", input: 0, expected: 0},
			{name: "π phase", input: math.Pi, expected: math.Pi},
			{name: "2π wraps to 0", input: 2 * math.Pi, expected: 0},
			{name: "3π wraps to π", input: 3 * math.Pi, expected: math.Pi},
			{name: "π/2 phase", input: math.Pi / 2, expected: math.Pi / 2},
			{name: "3π/2 phase", input: 3 * math.Pi / 2, expected: 3 * math.Pi / 2},
			
			// Negative values
			{name: "negative π wraps to π", input: -math.Pi, expected: math.Pi},
			{name: "negative 2π wraps to 0", input: -2 * math.Pi, expected: 0},
			{name: "negative π/2", input: -math.Pi / 2, expected: 3 * math.Pi / 2},
			{name: "negative 3π/2", input: -3 * math.Pi / 2, expected: math.Pi / 2},
			
			// Large values
			{name: "4π wraps to 0", input: 4 * math.Pi, expected: 0},
			{name: "5π wraps to π", input: 5 * math.Pi, expected: math.Pi},
			{name: "10π wraps to 0", input: 10 * math.Pi, expected: 0},
			{name: "100π wraps to 0", input: 100 * math.Pi, expected: 0},
			{name: "101π wraps to π", input: 101 * math.Pi, expected: math.Pi},
			
			// Very small values
			{name: "0.001 radians", input: 0.001, expected: 0.001},
			{name: "negative 0.001", input: -0.001, expected: 2*math.Pi - 0.001},
			
			// Non-π multiples
			{name: "1.5 radians", input: 1.5, expected: 1.5},
			{name: "2.7 radians", input: 2.7, expected: 2.7},
			{name: "7.5 radians", input: 7.5, expected: 7.5 - 2*math.Pi},
			{name: "negative 7.5", input: -7.5, expected: 2*math.Pi - 7.5 + 2*math.Pi},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := emerge.WrapPhase(tt.input)
				if math.Abs(result-tt.expected) > 0.0001 {
					t.Errorf("WrapPhase(%f) = %f, expected %f",
						tt.input, result, tt.expected)
				}
			})
		}
	})

	t.Run("PhaseDifference", func(t *testing.T) {
		tests := []struct {
			name     string
			phase1   float64
			phase2   float64
			expected float64
		}{
			// Happy path cases
			{name: "identical phases", phase1: 0, phase2: 0, expected: 0},
			{name: "π to 0", phase1: math.Pi, phase2: 0, expected: math.Pi},
			{name: "0 to π", phase1: 0, phase2: math.Pi, expected: -math.Pi},
			{name: "π/2 to π/4", phase1: math.Pi / 2, phase2: math.Pi / 4, expected: math.Pi / 4},
			{name: "π/4 to π/2", phase1: math.Pi / 4, phase2: math.Pi / 2, expected: -math.Pi / 4},
			
			// Wrapping cases (shortest path)
			{name: "0.1 to 2π-0.1", phase1: 0.1, phase2: 2*math.Pi - 0.1, expected: 0.2},
			{name: "2π-0.1 to 0.1", phase1: 2*math.Pi - 0.1, phase2: 0.1, expected: -0.2},
			{name: "0 to 3π/2", phase1: 0, phase2: 3 * math.Pi / 2, expected: -math.Pi / 2},
			{name: "3π/2 to 0", phase1: 3 * math.Pi / 2, phase2: 0, expected: math.Pi / 2},
			
			// Large phase values
			{name: "10π to 10.5π", phase1: 10 * math.Pi, phase2: 10.5 * math.Pi, expected: -math.Pi / 2},
			{name: "100.1π to 100π", phase1: 100.1 * math.Pi, phase2: 100 * math.Pi, expected: 0.1 * math.Pi},
			
			// Negative phase values
			{name: "-π/2 to π/2", phase1: -math.Pi / 2, phase2: math.Pi / 2, expected: -math.Pi},
			{name: "π/2 to -π/2", phase1: math.Pi / 2, phase2: -math.Pi / 2, expected: math.Pi},
			{name: "-π to π", phase1: -math.Pi, phase2: math.Pi, expected: 0},
			
			// Edge cases
			{name: "π to -π (same angle)", phase1: math.Pi, phase2: -math.Pi, expected: 0},
			{name: "very small difference", phase1: 1.0, phase2: 1.0001, expected: -0.0001},
			{name: "near 2π boundary", phase1: 2*math.Pi - 0.01, phase2: 0.01, expected: -0.02},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := emerge.PhaseDifference(tt.phase1, tt.phase2)
				if math.Abs(result-tt.expected) > 0.0001 {
					t.Errorf("PhaseDifference(%f, %f) = %f, expected %f",
						tt.phase1, tt.phase2, result, tt.expected)
				}
			})
		}
	})

	t.Run("MeasureCoherence", func(t *testing.T) {
		tests := []struct {
			name      string
			phases    []float64
			expected  float64
			tolerance float64
		}{
			// Happy path cases
			{
				name:      "all aligned at 0",
				phases:    []float64{0, 0, 0, 0},
				expected:  1.0,
				tolerance: 0.01,
			},
			{
				name:      "all aligned at π",
				phases:    []float64{math.Pi, math.Pi, math.Pi},
				expected:  1.0,
				tolerance: 0.01,
			},
			{
				name:      "all aligned at π/2",
				phases:    []float64{math.Pi / 2, math.Pi / 2, math.Pi / 2, math.Pi / 2},
				expected:  1.0,
				tolerance: 0.01,
			},
			{
				name:      "opposite pairs cancel",
				phases:    []float64{0, math.Pi, 0, math.Pi},
				expected:  0.0,
				tolerance: 0.01,
			},
			{
				name:      "quadrant distribution",
				phases:    []float64{0, math.Pi / 2, math.Pi, 3 * math.Pi / 2},
				expected:  0.0,
				tolerance: 0.01,
			},
			
			// Edge cases
			{
				name:      "empty array",
				phases:    []float64{},
				expected:  0.0,
				tolerance: 0.01,
			},
			{
				name:      "single phase",
				phases:    []float64{math.Pi / 2},
				expected:  1.0,
				tolerance: 0.01,
			},
			{
				name:      "two phases same",
				phases:    []float64{1.0, 1.0},
				expected:  1.0,
				tolerance: 0.01,
			},
			{
				name:      "two phases opposite",
				phases:    []float64{0, math.Pi},
				expected:  0.0,
				tolerance: 0.01,
			},
			{
				name:      "two phases perpendicular",
				phases:    []float64{0, math.Pi / 2},
				expected:  math.Sqrt(2) / 2,
				tolerance: 0.01,
			},
			
			// Many agents
			{
				name:      "100 aligned agents",
				phases:    make([]float64, 100), // All zeros
				expected:  1.0,
				tolerance: 0.01,
			},
			{
				name:      "slight spread",
				phases:    []float64{-0.1, -0.05, 0, 0.05, 0.1},
				expected:  0.99, // Very high but not perfect
				tolerance: 0.02,
			},
			{
				name:      "moderate spread",
				phases:    []float64{-0.5, -0.25, 0, 0.25, 0.5},
				expected:  0.95, // Still high coherence
				tolerance: 0.05,
			},
			{
				name:      "uniform distribution",
				phases:    uniformPhases(8), // 8 evenly spaced phases
				expected:  0.0,
				tolerance: 0.01,
			},
			
			// Numerical edge cases
			{
				name:      "very small phases",
				phases:    []float64{0.0001, 0.0002, 0.0003},
				expected:  1.0,
				tolerance: 0.01,
			},
			{
				name:      "phases near 2π",
				phases:    []float64{2*math.Pi - 0.01, 2*math.Pi - 0.02, 0.01},
				expected:  0.99,
				tolerance: 0.02,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := emerge.MeasureCoherence(tt.phases)
				if math.Abs(result-tt.expected) > tt.tolerance {
					t.Errorf("Expected coherence %f±%f, got %f",
						tt.expected, tt.tolerance, result)
				}
				// Coherence should always be between 0 and 1
				if result < -0.01 || result > 1.01 {
					t.Errorf("Coherence %f is outside valid range [0, 1]", result)
				}
			})
		}
	})
}

// Helper function to generate uniformly distributed phases
func uniformPhases(n int) []float64 {
	phases := make([]float64, n)
	for i := range n {
		phases[i] = float64(i) * 2 * math.Pi / float64(n)
	}
	return phases
}

