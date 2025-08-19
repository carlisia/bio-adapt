package emerge

import (
	"math"
	"testing"
)

func TestNewAttractorBasin(t *testing.T) {
	tests := []struct {
		name            string
		targetPhase     float64
		targetCoherence float64
		strength        float64
		radius          float64
		wantStrength    float64 // After clamping
	}{
		{
			name:            "standard basin",
			targetPhase:     0,
			targetCoherence: 0.8,
			strength:        0.5,
			radius:          math.Pi / 4,
			wantStrength:    0.5,
		},
		{
			name:            "negative strength clamped",
			targetPhase:     math.Pi,
			targetCoherence: 0.9,
			strength:        -0.5,
			radius:          math.Pi / 2,
			wantStrength:    0,
		},
		{
			name:            "excessive strength clamped",
			targetPhase:     math.Pi / 2,
			targetCoherence: 0.7,
			strength:        1.5,
			radius:          math.Pi,
			wantStrength:    1,
		},
		{
			name:            "zero radius",
			targetPhase:     0,
			targetCoherence: 0.85,
			strength:        0.3,
			radius:          0,
			wantStrength:    0.3,
		},
		{
			name:            "full circle radius",
			targetPhase:     -math.Pi,
			targetCoherence: 0.95,
			strength:        0.9,
			radius:          2 * math.Pi,
			wantStrength:    0.9,
		},
		{
			name:            "boundary strength values",
			targetPhase:     0,
			targetCoherence: 0.5,
			strength:        0.0,
			radius:          math.Pi / 3,
			wantStrength:    0.0,
		},
		{
			name:            "high precision values",
			targetPhase:     1.234567,
			targetCoherence: 0.876543,
			strength:        0.654321,
			radius:          0.987654,
			wantStrength:    0.654321,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basin := NewAttractorBasin(tt.targetPhase, tt.targetCoherence, tt.strength, tt.radius)

			if basin == nil {
				t.Fatal("Expected basin to be created")
			}

			if basin.targetPhase != tt.targetPhase {
				t.Errorf("targetPhase = %f, want %f", basin.targetPhase, tt.targetPhase)
			}

			if basin.targetCoherence != tt.targetCoherence {
				t.Errorf("targetCoherence = %f, want %f", basin.targetCoherence, tt.targetCoherence)
			}

			if basin.strength != tt.wantStrength {
				t.Errorf("strength = %f, want %f", basin.strength, tt.wantStrength)
			}

			if basin.radius != tt.radius {
				t.Errorf("radius = %f, want %f", basin.radius, tt.radius)
			}
		})
	}
}

func TestAttractorBasinAttractionForce(t *testing.T) {
	tests := []struct {
		name         string
		basin        *AttractorBasin
		currentPhase float64
		wantMin      float64
		wantMax      float64
		description  string
	}{
		{
			name:         "at target center",
			basin:        NewAttractorBasin(math.Pi, 0.9, 0.5, math.Pi/2),
			currentPhase: math.Pi,
			wantMin:      0.45,
			wantMax:      0.5,
			description:  "maximum force at target",
		},
		{
			name:         "halfway to edge",
			basin:        NewAttractorBasin(math.Pi, 0.9, 0.5, math.Pi/2),
			currentPhase: math.Pi + math.Pi/4,
			wantMin:      0.2,
			wantMax:      0.3,
			description:  "decreasing force away from target",
		},
		{
			name:         "at radius edge",
			basin:        NewAttractorBasin(math.Pi, 0.9, 0.5, math.Pi/2),
			currentPhase: math.Pi + math.Pi/2,
			wantMin:      0,
			wantMax:      0.05,
			description:  "minimal force at edge",
		},
		{
			name:         "outside radius",
			basin:        NewAttractorBasin(math.Pi, 0.9, 0.5, math.Pi/2),
			currentPhase: 0,
			wantMin:      0,
			wantMax:      0,
			description:  "no force outside basin",
		},
		{
			name:         "zero radius basin",
			basin:        NewAttractorBasin(0, 0.8, 0.7, 0),
			currentPhase: 0.1,
			wantMin:      0,
			wantMax:      0,
			description:  "no force with zero radius",
		},
		{
			name:         "maximum strength basin",
			basin:        NewAttractorBasin(0, 0.95, 1.0, math.Pi/4),
			currentPhase: 0,
			wantMin:      0.9,
			wantMax:      1.0,
			description:  "maximum possible force",
		},
		{
			name:         "wrapped phase difference",
			basin:        NewAttractorBasin(0, 0.8, 0.5, math.Pi/2),
			currentPhase: 2*math.Pi - 0.1,
			wantMin:      0.35,
			wantMax:      0.5,
			description:  "handle phase wrapping correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			force := tt.basin.AttractionForce(tt.currentPhase)
			if force < tt.wantMin || force > tt.wantMax {
				t.Errorf("AttractionForce() = %f, want in [%f, %f] for %s",
					force, tt.wantMin, tt.wantMax, tt.description)
			}
		})
	}
}

func TestAttractorBasinPhaseDistance(t *testing.T) {
	tests := []struct {
		name         string
		basin        *AttractorBasin
		currentPhase float64
		wantDist     float64
		tolerance    float64
	}{
		{
			name:         "at target",
			basin:        NewAttractorBasin(0, 0.8, 0.5, math.Pi),
			currentPhase: 0,
			wantDist:     0,
			tolerance:    0.001,
		},
		{
			name:         "quarter circle clockwise",
			basin:        NewAttractorBasin(0, 0.8, 0.5, math.Pi),
			currentPhase: math.Pi / 2,
			wantDist:     math.Pi / 2,
			tolerance:    0.01,
		},
		{
			name:         "quarter circle counter-clockwise",
			basin:        NewAttractorBasin(0, 0.8, 0.5, math.Pi),
			currentPhase: -math.Pi / 2,
			wantDist:     math.Pi / 2,
			tolerance:    0.01,
		},
		{
			name:         "opposite side",
			basin:        NewAttractorBasin(0, 0.8, 0.5, math.Pi),
			currentPhase: math.Pi,
			wantDist:     math.Pi,
			tolerance:    0.01,
		},
		{
			name:         "wrapped positive",
			basin:        NewAttractorBasin(0, 0.8, 0.5, math.Pi),
			currentPhase: 3 * math.Pi / 2,
			wantDist:     math.Pi / 2,
			tolerance:    0.01,
		},
		{
			name:         "wrapped negative",
			basin:        NewAttractorBasin(math.Pi, 0.8, 0.5, math.Pi),
			currentPhase: -math.Pi,
			wantDist:     0,
			tolerance:    0.01,
		},
		{
			name:         "small angle",
			basin:        NewAttractorBasin(0.1, 0.8, 0.5, math.Pi),
			currentPhase: 0.2,
			wantDist:     0.1,
			tolerance:    0.001,
		},
		{
			name:         "near 2π boundary",
			basin:        NewAttractorBasin(2*math.Pi - 0.1, 0.8, 0.5, math.Pi),
			currentPhase: 0.1,
			wantDist:     0.2,
			tolerance:    0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := tt.basin.PhaseDistance(tt.currentPhase)
			if math.Abs(dist-tt.wantDist) > tt.tolerance {
				t.Errorf("PhaseDistance() = %f, want %f ± %f",
					dist, tt.wantDist, tt.tolerance)
			}
		})
	}
}

func TestAttractorBasinIsInBasin(t *testing.T) {
	tests := []struct {
		name    string
		basin   *AttractorBasin
		phases  []struct {
			phase   float64
			inBasin bool
		}
	}{
		{
			name:  "small radius basin",
			basin: NewAttractorBasin(math.Pi/2, 0.8, 0.5, math.Pi/8),
			phases: []struct {
				phase   float64
				inBasin bool
			}{
				{math.Pi / 2, true},              // At target
				{math.Pi/2 + math.Pi/16, true},    // Within radius
				{math.Pi/2 + math.Pi/8, true},     // At boundary
				{math.Pi/2 + math.Pi/4, false},    // Outside
				{math.Pi, false},                  // Far outside
			},
		},
		{
			name:  "large radius basin",
			basin: NewAttractorBasin(0, 0.9, 0.7, math.Pi),
			phases: []struct {
				phase   float64
				inBasin bool
			}{
				{0, true},                // At target
				{math.Pi / 2, true},       // Within radius
				{math.Pi - 0.01, true},    // Near boundary
				{math.Pi + 0.01, true},    // Just outside but within floating point tolerance
				{3 * math.Pi / 2, true},   // Wrapped, within
			},
		},
		{
			name:  "zero radius basin",
			basin: NewAttractorBasin(math.Pi, 0.85, 0.6, 0),
			phases: []struct {
				phase   float64
				inBasin bool
			}{
				{math.Pi, true},         // At target (distance 0 <= 0)
				{math.Pi + 0.001, false}, // Very close
				{0, false},              // Opposite
			},
		},
		{
			name:  "full circle basin",
			basin: NewAttractorBasin(0, 0.95, 0.9, 2*math.Pi),
			phases: []struct {
				phase   float64
				inBasin bool
			}{
				{0, true},           // At target
				{math.Pi / 2, true},  // Quarter
				{math.Pi, true},      // Opposite
				{3 * math.Pi / 2, true}, // Three quarters
				{-math.Pi, true},     // Wrapped negative
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, p := range tt.phases {
				inBasin := tt.basin.IsInBasin(p.phase)
				if inBasin != p.inBasin {
					t.Errorf("phase[%d]=%f: IsInBasin() = %v, want %v",
						i, p.phase, inBasin, p.inBasin)
				}
			}
		})
	}
}

func TestAttractorBasinConvergenceRate(t *testing.T) {
	tests := []struct {
		name      string
		basin     *AttractorBasin
		phase     float64
		wantMin   float64
		wantMax   float64
		description string
	}{
		{
			name:      "at target with high strength",
			basin:     NewAttractorBasin(0, 0.9, 0.8, math.Pi/4),
			phase:     0,
			wantMin:   0.7,
			wantMax:   0.8,
			description: "maximum convergence at target",
		},
		{
			name:      "within basin",
			basin:     NewAttractorBasin(0, 0.9, 0.8, math.Pi/2),
			phase:     math.Pi / 4,
			wantMin:   0.3,
			wantMax:   0.5,
			description: "partial convergence within basin",
		},
		{
			name:      "outside basin",
			basin:     NewAttractorBasin(0, 0.9, 0.8, math.Pi/4),
			phase:     math.Pi,
			wantMin:   0,
			wantMax:   0,
			description: "no convergence outside basin",
		},
		{
			name:      "zero radius",
			basin:     NewAttractorBasin(0, 0.9, 0.8, 0),
			phase:     0,
			wantMin:   0,
			wantMax:   0,
			description: "no convergence with zero radius",
		},
		{
			name:      "zero strength",
			basin:     NewAttractorBasin(0, 0.9, 0, math.Pi/2),
			phase:     0,
			wantMin:   0,
			wantMax:   0,
			description: "no convergence with zero strength",
		},
		{
			name:      "at basin edge",
			basin:     NewAttractorBasin(math.Pi, 0.85, 0.6, math.Pi/3),
			phase:     math.Pi + math.Pi/3,
			wantMin:   0,
			wantMax:   0.1,
			description: "minimal convergence at edge",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate := tt.basin.ConvergenceRate(tt.phase)
			if rate < tt.wantMin || rate > tt.wantMax {
				t.Errorf("ConvergenceRate() = %f, want in [%f, %f] for %s",
					rate, tt.wantMin, tt.wantMax, tt.description)
			}
		})
	}
}

func TestAttractorBasinEdgeCases(t *testing.T) {
	t.Run("negative coherence", func(t *testing.T) {
		basin := NewAttractorBasin(0, -0.5, 0.5, math.Pi)
		if basin.targetCoherence != -0.5 {
			t.Errorf("Should accept negative coherence, got %f", basin.targetCoherence)
		}
	})

	t.Run("very large radius", func(t *testing.T) {
		basin := NewAttractorBasin(0, 0.8, 0.5, 100*math.Pi)
		force := basin.AttractionForce(math.Pi)
		if force <= 0 {
			t.Error("Should have non-zero force even with large radius")
		}
	})

	t.Run("negative radius", func(t *testing.T) {
		basin := NewAttractorBasin(0, 0.8, 0.5, -math.Pi)
		// Negative radius is stored as-is, IsInBasin compares distance <= radius
		// Since distance is always >= 0 and radius is negative, this will always be false
		inBasin := basin.IsInBasin(0) // Even at target
		if inBasin {
			t.Error("Negative radius means nothing is in basin")
		}
	})
}