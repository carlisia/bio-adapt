package attractor

import (
	"math"
	"testing"
	"time"
)

func TestNewRhythmicPattern(t *testing.T) {
	phases := []float64{0, math.Pi/4, math.Pi/2, 3*math.Pi/4, math.Pi}
	frequencies := []time.Duration{
		100 * time.Millisecond,
		100 * time.Millisecond,
		100 * time.Millisecond,
		100 * time.Millisecond,
		100 * time.Millisecond,
	}

	pattern := NewRhythmicPattern(phases, frequencies)

	if len(pattern.Phases) != len(phases) {
		t.Errorf("Expected %d phases, got %d", len(phases), len(pattern.Phases))
	}

	if len(pattern.Frequencies) != len(frequencies) {
		t.Errorf("Expected %d frequencies, got %d", len(frequencies), len(pattern.Frequencies))
	}

	if pattern.Amplitude <= 0 {
		t.Error("Amplitude should be positive")
	}

	if pattern.Period != 100*time.Millisecond {
		t.Errorf("Expected period 100ms, got %v", pattern.Period)
	}

	if pattern.Confidence != 1.0 {
		t.Errorf("Expected confidence 1.0, got %f", pattern.Confidence)
	}
}

func TestPatternDetectGaps(t *testing.T) {
	// Create pattern with a gap (large phase jump)
	phases := []float64{0, math.Pi/4, math.Pi*1.5, math.Pi*1.75}
	frequencies := []time.Duration{
		100 * time.Millisecond,
		100 * time.Millisecond,
		100 * time.Millisecond,
		100 * time.Millisecond,
	}

	pattern := NewRhythmicPattern(phases, frequencies)
	gaps := pattern.Detect()

	if len(gaps) == 0 {
		t.Error("Should detect at least one gap")
	}

	// The gap should be between indices 1 and 2
	foundGap := false
	for _, gap := range gaps {
		if gap.StartIdx == 1 && gap.EndIdx == 2 {
			foundGap = true
			break
		}
	}

	if !foundGap {
		t.Error("Should detect gap between indices 1 and 2")
	}
}

func TestPatternComplete(t *testing.T) {
	phases := []float64{0, math.Pi/4, math.Pi/2, 3*math.Pi/4, math.Pi}
	frequencies := []time.Duration{
		100 * time.Millisecond,
		100 * time.Millisecond,
		100 * time.Millisecond,
		100 * time.Millisecond,
		100 * time.Millisecond,
	}

	pattern := NewRhythmicPattern(phases, frequencies)

	tests := []struct {
		name     string
		gap      PatternGap
		expected int // expected number of completed values
	}{
		{
			name: "single step gap",
			gap: PatternGap{
				StartIdx: 0,
				EndIdx:   2,
				Duration: 100 * time.Millisecond,
			},
			expected: 1,
		},
		{
			name: "multi step gap",
			gap: PatternGap{
				StartIdx: 1,
				EndIdx:   4,
				Duration: 100 * time.Millisecond,
			},
			expected: 2,
		},
		{
			name: "adjacent indices",
			gap: PatternGap{
				StartIdx: 2,
				EndIdx:   3,
				Duration: 100 * time.Millisecond,
			},
			expected: 0,
		},
		{
			name: "invalid gap",
			gap: PatternGap{
				StartIdx: -1,
				EndIdx:   2,
				Duration: 100 * time.Millisecond,
			},
			expected: -1, // nil result
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			completed := pattern.Complete(tt.gap)
			
			if tt.expected == -1 {
				if completed != nil {
					t.Error("Expected nil for invalid gap")
				}
			} else {
				if len(completed) != tt.expected {
					t.Errorf("Expected %d completed values, got %d", tt.expected, len(completed))
				}
			}
		})
	}
}

func TestPatternSimilarity(t *testing.T) {
	phases1 := []float64{0, math.Pi/4, math.Pi/2, 3*math.Pi/4}
	frequencies1 := []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond}
	pattern1 := NewRhythmicPattern(phases1, frequencies1)

	tests := []struct {
		name     string
		other    *RhythmicPattern
		minSim   float64
		maxSim   float64
	}{
		{
			name:   "identical pattern",
			other:  NewRhythmicPattern(phases1, frequencies1),
			minSim: 0.95,
			maxSim: 1.0,
		},
		{
			name: "similar pattern",
			other: NewRhythmicPattern(
				[]float64{0, math.Pi/4 + 0.1, math.Pi/2, 3*math.Pi/4 - 0.1},
				frequencies1,
			),
			minSim: 0.7,
			maxSim: 1.0,
		},
		{
			name: "different pattern",
			other: NewRhythmicPattern(
				[]float64{math.Pi, math.Pi * 1.25, math.Pi * 1.5, math.Pi * 1.75},
				[]time.Duration{200 * time.Millisecond, 200 * time.Millisecond, 200 * time.Millisecond, 200 * time.Millisecond},
			),
			minSim: 0.0,
			maxSim: 0.5,
		},
		{
			name:   "nil pattern",
			other:  nil,
			minSim: 0.0,
			maxSim: 0.0,
		},
		{
			name:   "empty pattern",
			other:  NewRhythmicPattern([]float64{}, []time.Duration{}),
			minSim: 0.0,
			maxSim: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			similarity := pattern1.Similarity(tt.other)
			if similarity < tt.minSim || similarity > tt.maxSim {
				t.Errorf("Expected similarity in [%f, %f], got %f", tt.minSim, tt.maxSim, similarity)
			}
		})
	}
}

func TestPatternTemplate(t *testing.T) {
	basePattern := RhythmicPattern{
		Phases:      []float64{0, math.Pi/2, math.Pi, 3*math.Pi/2},
		Frequencies: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
		Amplitude:   1.0,
		Period:      100 * time.Millisecond,
		Confidence:  1.0,
	}

	template := &PatternTemplate{
		Name:        "test-template",
		BasePattern: basePattern,
		Variations:  []RhythmicPattern{},
		Tolerance:   0.2,
	}

	tests := []struct {
		name     string
		pattern  *RhythmicPattern
		expected bool
	}{
		{
			name: "exact match",
			pattern: &RhythmicPattern{
				Phases:      []float64{0, math.Pi/2, math.Pi, 3*math.Pi/2},
				Frequencies: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
				Amplitude:   1.0,
				Period:      100 * time.Millisecond,
				Confidence:  1.0,
			},
			expected: true,
		},
		{
			name: "within tolerance",
			pattern: &RhythmicPattern{
				Phases:      []float64{0.1, math.Pi/2 + 0.1, math.Pi, 3*math.Pi/2 - 0.1},
				Frequencies: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
				Amplitude:   0.9,
				Period:      100 * time.Millisecond,
				Confidence:  1.0,
			},
			expected: true,
		},
		{
			name: "outside tolerance",
			pattern: &RhythmicPattern{
				Phases:      []float64{math.Pi/4, 3*math.Pi/4, 5*math.Pi/4, 7*math.Pi/4},
				Frequencies: []time.Duration{200 * time.Millisecond, 200 * time.Millisecond, 200 * time.Millisecond, 200 * time.Millisecond},
				Amplitude:   2.0,
				Period:      200 * time.Millisecond,
				Confidence:  1.0,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := template.Matches(tt.pattern)
			if matches != tt.expected {
				t.Errorf("Expected Matches() = %v, got %v", tt.expected, matches)
			}
		})
	}
}

func TestPatternLibrary(t *testing.T) {
	library := NewPatternLibrary()

	// Add templates
	circadian := &PatternTemplate{
		Name: "circadian",
		BasePattern: RhythmicPattern{
			Phases:      []float64{0, math.Pi/2, math.Pi, 3*math.Pi/2},
			Frequencies: []time.Duration{6 * time.Hour, 6 * time.Hour, 6 * time.Hour, 6 * time.Hour},
			Amplitude:   1.0,
			Period:      24 * time.Hour,
			Confidence:  1.0,
		},
		Tolerance: 0.15,
	}

	ultradian := &PatternTemplate{
		Name: "ultradian",
		BasePattern: RhythmicPattern{
			Phases:      []float64{0, math.Pi},
			Frequencies: []time.Duration{45 * time.Minute, 45 * time.Minute},
			Amplitude:   0.5,
			Period:      90 * time.Minute,
			Confidence:  1.0,
		},
		Tolerance: 0.2,
	}

	library.Add("circadian", circadian)
	library.Add("ultradian", ultradian)

	// Test identification
	testPattern := &RhythmicPattern{
		Phases:      []float64{0, math.Pi/2, math.Pi, 3*math.Pi/2},
		Frequencies: []time.Duration{6 * time.Hour, 6 * time.Hour, 6 * time.Hour, 6 * time.Hour},
		Amplitude:   0.95,
		Period:      24 * time.Hour,
		Confidence:  0.9,
	}

	name, similarity := library.Identify(testPattern)
	
	if name != "circadian" {
		t.Errorf("Expected to identify as 'circadian', got '%s'", name)
	}

	if similarity < 0.8 {
		t.Errorf("Expected high similarity (>0.8), got %f", similarity)
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("calculateAmplitude", func(t *testing.T) {
		// Test with uniform phases (low amplitude)
		uniform := []float64{1.0, 1.0, 1.0, 1.0}
		amp := calculateAmplitude(uniform)
		if amp != 0 {
			t.Errorf("Expected amplitude 0 for uniform phases, got %f", amp)
		}

		// Test with varied phases (higher amplitude)
		varied := []float64{0, math.Pi/2, math.Pi, 3*math.Pi/2}
		amp = calculateAmplitude(varied)
		if amp <= 0 {
			t.Error("Expected positive amplitude for varied phases")
		}

		// Test with empty slice
		empty := []float64{}
		amp = calculateAmplitude(empty)
		if amp != 0 {
			t.Errorf("Expected amplitude 0 for empty phases, got %f", amp)
		}
	})

	t.Run("calculatePeriod", func(t *testing.T) {
		frequencies := []time.Duration{
			100 * time.Millisecond,
			200 * time.Millisecond,
			150 * time.Millisecond,
		}
		period := calculatePeriod(frequencies)
		expected := 150 * time.Millisecond
		if period != expected {
			t.Errorf("Expected period %v, got %v", expected, period)
		}

		// Test with empty slice
		empty := []time.Duration{}
		period = calculatePeriod(empty)
		if period != 0 {
			t.Errorf("Expected period 0 for empty frequencies, got %v", period)
		}
	})

	t.Run("phaseCorrelation", func(t *testing.T) {
		// Test identical phases
		phases1 := []float64{0, math.Pi/2, math.Pi, 3*math.Pi/2}
		phases2 := []float64{0, math.Pi/2, math.Pi, 3*math.Pi/2}
		corr := phaseCorrelation(phases1, phases2)
		if corr < 0.99 {
			t.Errorf("Expected correlation ~1.0 for identical phases, got %f", corr)
		}

		// Test opposite phases
		phases3 := []float64{math.Pi, 3*math.Pi/2, 0, math.Pi/2}
		corr = phaseCorrelation(phases1, phases3)
		if corr > 0.5 {
			t.Errorf("Expected low correlation for opposite phases, got %f", corr)
		}

		// Test empty slices
		empty := []float64{}
		corr = phaseCorrelation(empty, phases1)
		if corr != 0 {
			t.Errorf("Expected correlation 0 for empty phases, got %f", corr)
		}
	})
}