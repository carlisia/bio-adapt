package emerge_test

import (
	"math"
	"testing"
	"time"

	"github.com/carlisia/bio-adapt/emerge"
)

func TestNewAttractorBasin(t *testing.T) {
	target := emerge.State{
		Phase:     math.Pi,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.8,
	}

	basin := emerge.NewAttractorBasin(target, 0.5, 10.0)

	// Test basic functionality - we can't access private fields
	// so we test behavior instead
	if basin == nil {
		t.Fatal("Expected basin to be created")
	}

	// Test that a state at the target is in the basin
	if !basin.IsInBasin(target) {
		t.Error("Target state should be in its own basin")
	}

	// Distance to target should be 0
	if dist := basin.DistanceToTarget(target); dist != 0 {
		t.Errorf("Distance from target to itself should be 0, got %f", dist)
	}
}

func TestBasinDistanceToTarget(t *testing.T) {
	target := emerge.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.8,
	}

	basin := emerge.NewAttractorBasin(target, 0.5, math.Pi/4)

	tests := []struct {
		name     string
		state    emerge.State
		expected float64
	}{
		{
			name: "state at target",
			state: emerge.State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			expected: 0,
		},
		{
			name: "state at π/4",
			state: emerge.State{
				Phase:     math.Pi / 4,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			expected: math.Pi / 4,
		},
		{
			name: "state at π",
			state: emerge.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			expected: math.Pi,
		},
		{
			name: "state at 3π/2 (wraps to π/2)",
			state: emerge.State{
				Phase:     3 * math.Pi / 2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			expected: math.Pi / 2,
		},
		{
			name: "state at 2π (wraps to 0)",
			state: emerge.State{
				Phase:     2 * math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := basin.DistanceToTarget(tt.state)
			if math.Abs(dist-tt.expected) > 0.01 {
				t.Errorf("Expected distance %f, got %f", tt.expected, dist)
			}
		})
	}
}

func TestBasinIsInBasin(t *testing.T) {
	target := emerge.State{
		Phase:     math.Pi / 2,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.8,
	}

	// Basin with radius π/4
	basin := emerge.NewAttractorBasin(target, 0.5, math.Pi/4)

	tests := []struct {
		name    string
		state   emerge.State
		inBasin bool
	}{
		{
			name:    "state at target",
			state:   target,
			inBasin: true,
		},
		{
			name: "state within radius",
			state: emerge.State{
				Phase:     math.Pi/2 + math.Pi/8, // π/2 + π/8 = 5π/8
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			inBasin: true,
		},
		{
			name: "state at radius boundary",
			state: emerge.State{
				Phase:     math.Pi/2 + math.Pi/4, // π/2 + π/4 = 3π/4
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			inBasin: true,
		},
		{
			name: "state outside radius",
			state: emerge.State{
				Phase:     math.Pi, // π is π/2 away from π/2
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			inBasin: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inBasin := basin.IsInBasin(tt.state)
			if inBasin != tt.inBasin {
				t.Errorf("Expected IsInBasin=%v, got %v", tt.inBasin, inBasin)
			}
		})
	}
}

func TestBasinAttractionForce(t *testing.T) {
	target := emerge.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.8,
	}

	basin := emerge.NewAttractorBasin(target, 0.5, math.Pi/2)

	tests := []struct {
		name     string
		state    emerge.State
		minForce float64
		maxForce float64
	}{
		{
			name:     "state at target",
			state:    target,
			minForce: 0.45, // Should be close to max strength (0.5)
			maxForce: 0.5,
		},
		{
			name: "state at quarter radius",
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
			state: emerge.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			minForce: 0,
			maxForce: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			force := basin.AttractionForce(tt.state)
			if force < tt.minForce || force > tt.maxForce {
				t.Errorf("Expected force in range [%f, %f], got %f",
					tt.minForce, tt.maxForce, force)
			}
		})
	}
}

func TestBasinConvergenceRate(t *testing.T) {
	target := emerge.State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.8,
	}

	basin := emerge.NewAttractorBasin(target, 0.8, math.Pi/4)

	tests := []struct {
		name    string
		state   emerge.State
		minRate float64
		maxRate float64
	}{
		{
			name:    "state at target",
			state:   target,
			minRate: 0.7,
			maxRate: 0.8,
		},
		{
			name: "state within basin",
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
			state: emerge.State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			minRate: 0,
			maxRate: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate := basin.ConvergenceRate(tt.state)
			if rate < tt.minRate || rate > tt.maxRate {
				t.Errorf("Expected rate in range [%f, %f], got %f",
					tt.minRate, tt.maxRate, rate)
			}
		})
	}
}

func TestBasinOptimalAdjustment(t *testing.T) {
	target := emerge.State{
		Phase:     math.Pi,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.8,
	}

	basin := emerge.NewAttractorBasin(target, 1.0, math.Pi)

	tests := []struct {
		name       string
		current    emerge.State
		expectSign float64 // Expected sign of adjustment (-1, 0, or 1)
	}{
		{
			name:       "state at target",
			current:    target,
			expectSign: 0,
		},
		{
			name: "state before target",
			current: emerge.State{
				Phase:     math.Pi / 2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			expectSign: 1, // Should adjust forward
		},
		{
			name: "state after target",
			current: emerge.State{
				Phase:     3 * math.Pi / 2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			expectSign: -1, // Should adjust backward
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adjustment := basin.OptimalAdjustment(tt.current)

			if tt.expectSign == 0 {
				if math.Abs(adjustment) > 0.01 {
					t.Errorf("Expected near-zero adjustment, got %f", adjustment)
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
		})
	}
}

func TestPhaseUtilityFunctions(t *testing.T) {
	t.Run("WrapPhase", func(t *testing.T) {
		tests := []struct {
			input    float64
			expected float64
		}{
			{0, 0},
			{math.Pi, math.Pi},
			{2 * math.Pi, 0},
			{3 * math.Pi, math.Pi},
			{-math.Pi, math.Pi},
			{-2 * math.Pi, 0},
		}

		for _, tt := range tests {
			result := emerge.WrapPhase(tt.input)
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("WrapPhase(%f) = %f, expected %f",
					tt.input, result, tt.expected)
			}
		}
	})

	t.Run("PhaseDifference", func(t *testing.T) {
		tests := []struct {
			phase1   float64
			phase2   float64
			expected float64
		}{
			{0, 0, 0},
			{math.Pi, 0, math.Pi},
			{0, math.Pi, -math.Pi},
			{0.1, 2*math.Pi - 0.1, 0.2},
			{2*math.Pi - 0.1, 0.1, -0.2},
		}

		for _, tt := range tests {
			result := emerge.PhaseDifference(tt.phase1, tt.phase2)
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("PhaseDifference(%f, %f) = %f, expected %f",
					tt.phase1, tt.phase2, result, tt.expected)
			}
		}
	})

	t.Run("MeasureCoherence", func(t *testing.T) {
		tests := []struct {
			name     string
			phases   []float64
			expected float64
		}{
			{
				name:     "all aligned",
				phases:   []float64{0, 0, 0, 0},
				expected: 1.0,
			},
			{
				name:     "opposite pairs",
				phases:   []float64{0, math.Pi, 0, math.Pi},
				expected: 0.0,
			},
			{
				name:     "empty",
				phases:   []float64{},
				expected: 0.0,
			},
			{
				name:     "single phase",
				phases:   []float64{math.Pi / 2},
				expected: 1.0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := emerge.MeasureCoherence(tt.phases)
				if math.Abs(result-tt.expected) > 0.01 {
					t.Errorf("Expected coherence %f, got %f",
						tt.expected, result)
				}
			})
		}
	})
}

