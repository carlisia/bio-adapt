package attractor

import (
	"math"
	"testing"
	"time"
)

func TestNewAttractorBasin(t *testing.T) {
	target := State{
		Phase:     math.Pi,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.8,
	}

	basin := NewAttractorBasin(target, 0.5, 10.0)

	if basin.TargetState.Phase != target.Phase {
		t.Errorf("Expected target phase %f, got %f", target.Phase, basin.TargetState.Phase)
	}

	if basin.Radius != 10.0 {
		t.Errorf("Expected radius 10.0, got %f", basin.Radius)
	}

	if basin.Strength != 0.5 {
		t.Errorf("Expected strength 0.5, got %f", basin.Strength)
	}
}

func TestBasinDistanceToTarget(t *testing.T) {
	target := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.8,
	}

	basin := NewAttractorBasin(target, 0.5, math.Pi/4)

	tests := []struct {
		name     string
		state    State
		expected float64
	}{
		{
			name: "state at target",
			state: State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			expected: 0,
		},
		{
			name: "state at pi/4",
			state: State{
				Phase:     math.Pi / 4,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			expected: math.Pi / 4,
		},
		{
			name: "state at opposite phase",
			state: State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			expected: math.Pi,
		},
		{
			name: "state with wrapped phase",
			state: State{
				Phase:     2*math.Pi - math.Pi/8,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			expected: math.Pi / 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := basin.DistanceToTarget(tt.state)
			if math.Abs(dist-tt.expected) > 0.001 {
				t.Errorf("Expected distance %f, got %f", tt.expected, dist)
			}
		})
	}
}

func TestBasinAttractionForce(t *testing.T) {
	target := State{
		Phase:     math.Pi,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.8,
	}

	basin := NewAttractorBasin(target, 0.8, math.Pi/2)

	tests := []struct {
		name        string
		state       State
		minExpected float64
		maxExpected float64
	}{
		{
			name: "state at target",
			state: State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			minExpected: 0.7,
			maxExpected: 0.9,
		},
		{
			name: "state at edge of radius",
			state: State{
				Phase:     math.Pi + math.Pi/2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			minExpected: 0,
			maxExpected: 0.1,
		},
		{
			name: "state outside radius",
			state: State{
				Phase:     0,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			minExpected: 0,
			maxExpected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := basin.DistanceToTarget(tt.state)
			force := basin.AttractionForce(dist)
			if force < tt.minExpected || force > tt.maxExpected {
				t.Errorf("Expected attraction in [%f, %f], got %f",
					tt.minExpected, tt.maxExpected, force)
			}
		})
	}
}

func TestBasinConvergenceRate(t *testing.T) {
	target := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.8,
	}

	tests := []struct {
		name     string
		strength float64
		radius   float64
		expected float64
	}{
		{"high strength small radius", 0.8, 0.5, 1.6},
		{"low strength large radius", 0.2, 2.0, 0.1},
		{"zero radius", 0.5, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basin := NewAttractorBasin(target, tt.strength, tt.radius)
			rate := basin.ConvergenceRate()
			if math.Abs(rate-tt.expected) > 0.001 {
				t.Errorf("Expected rate %f, got %f", tt.expected, rate)
			}
		})
	}
}

func TestBasinIsInBasin(t *testing.T) {
	target := State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.8,
	}

	basin := NewAttractorBasin(target, 0.5, math.Pi/2)

	tests := []struct {
		name     string
		state    State
		expected bool
	}{
		{
			name:     "state at target",
			state:    target,
			expected: true,
		},
		{
			name: "state within radius",
			state: State{
				Phase:     math.Pi / 4,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			expected: true,
		},
		{
			name: "state at edge",
			state: State{
				Phase:     math.Pi / 2,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			expected: true,
		},
		{
			name: "state outside",
			state: State{
				Phase:     math.Pi,
				Frequency: 100 * time.Millisecond,
				Coherence: 0.8,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inBasin := basin.IsInBasin(tt.state)
			if inBasin != tt.expected {
				t.Errorf("Expected IsInBasin() = %v, got %v", tt.expected, inBasin)
			}
		})
	}
}

func TestMultipleBasins(t *testing.T) {
	// Test creating multiple basins
	basin1 := NewAttractorBasin(State{
		Phase:     0,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.8,
	}, 0.8, math.Pi/4)

	basin2 := NewAttractorBasin(State{
		Phase:     math.Pi,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.8,
	}, 0.5, math.Pi/4)

	// Test that basins are independent
	if basin1.TargetState.Phase == basin2.TargetState.Phase {
		t.Error("Basins should have different target phases")
	}

	if basin1.Strength == basin2.Strength {
		t.Error("Basins should have different strengths")
	}

	// Test distance calculations are independent
	state := State{
		Phase:     math.Pi / 2,
		Frequency: 100 * time.Millisecond,
		Coherence: 0.8,
	}

	dist1 := basin1.DistanceToTarget(state)
	dist2 := basin2.DistanceToTarget(state)

	// State at pi/2 should be equidistant from 0 and pi
	if math.Abs(dist1-dist2) > 0.001 {
		t.Errorf("Expected equal distances, got %f and %f", dist1, dist2)
	}
}

func TestBasinStrengthClamping(t *testing.T) {
	// Test that strength is clamped to [0, 1]
	tests := []struct {
		name     string
		strength float64
		expected float64
	}{
		{"negative strength", -0.5, 0},
		{"zero strength", 0, 0},
		{"normal strength", 0.5, 0.5},
		{"max strength", 1.0, 1.0},
		{"over max strength", 1.5, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basin := NewAttractorBasin(State{}, tt.strength, 1.0)
			if basin.Strength != tt.expected {
				t.Errorf("Expected strength %f, got %f", tt.expected, basin.Strength)
			}
		})
	}
}

func TestMeasureCoherence(t *testing.T) {
	tests := []struct {
		name     string
		phases   []float64
		expected float64
		delta    float64
	}{
		{
			name:     "perfect sync",
			phases:   []float64{0, 0, 0, 0},
			expected: 1.0,
			delta:    0.01,
		},
		{
			name:     "random phases",
			phases:   []float64{0, math.Pi/2, math.Pi, 3*math.Pi/2},
			expected: 0.0,
			delta:    0.01,
		},
		{
			name:     "partial sync",
			phases:   []float64{0, 0.1, -0.1, 0.05},
			expected: 0.98,
			delta:    0.05,
		},
		{
			name:     "empty phases",
			phases:   []float64{},
			expected: 0.0,
			delta:    0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coherence := MeasureCoherence(tt.phases)
			if math.Abs(coherence-tt.expected) > tt.delta {
				t.Errorf("Expected coherence %f±%f, got %f",
					tt.expected, tt.delta, coherence)
			}
		})
	}
}

