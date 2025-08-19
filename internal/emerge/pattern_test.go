package emerge

import (
	"math"
	"testing"
	"time"
)

func TestNewRhythmicPattern(t *testing.T) {
	tests := []struct {
		name           string
		phases         []float64
		frequencies    []time.Duration
		wantNumPhases  int
		wantNumFreqs   int
		wantAmplitudes []float64
	}{
		{
			name:           "standard pattern",
			phases:         []float64{0, math.Pi/2, math.Pi},
			frequencies:    []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 150 * time.Millisecond},
			wantNumPhases:  3,
			wantNumFreqs:   3,
			wantAmplitudes: []float64{1.0, 1.0, 1.0},
		},
		{
			name:           "single element",
			phases:         []float64{math.Pi},
			frequencies:    []time.Duration{50 * time.Millisecond},
			wantNumPhases:  1,
			wantNumFreqs:   1,
			wantAmplitudes: []float64{1.0},
		},
		{
			name:           "empty pattern",
			phases:         []float64{},
			frequencies:    []time.Duration{},
			wantNumPhases:  0,
			wantNumFreqs:   0,
			wantAmplitudes: []float64{},
		},
		{
			name:           "mismatched lengths",
			phases:         []float64{0, math.Pi},
			frequencies:    []time.Duration{100 * time.Millisecond},
			wantNumPhases:  2,
			wantNumFreqs:   1,
			wantAmplitudes: []float64{1.0, 1.0},
		},
		{
			name:           "many elements",
			phases:         []float64{0, 0.5, 1.0, 1.5, 2.0, 2.5, 3.0},
			frequencies:    []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 30 * time.Millisecond, 40 * time.Millisecond, 50 * time.Millisecond, 60 * time.Millisecond, 70 * time.Millisecond},
			wantNumPhases:  7,
			wantNumFreqs:   7,
			wantAmplitudes: []float64{1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := NewRhythmicPattern(tt.phases, tt.frequencies)

			if pattern == nil {
				t.Fatal("Expected pattern to be created")
			}

			if len(pattern.phases) != tt.wantNumPhases {
				t.Errorf("phases length = %d, want %d", len(pattern.phases), tt.wantNumPhases)
			}

			if len(pattern.frequencies) != tt.wantNumFreqs {
				t.Errorf("frequencies length = %d, want %d", len(pattern.frequencies), tt.wantNumFreqs)
			}

			if len(pattern.amplitudes) != len(tt.wantAmplitudes) {
				t.Errorf("amplitudes length = %d, want %d", len(pattern.amplitudes), len(tt.wantAmplitudes))
			}

			for i, wantAmp := range tt.wantAmplitudes {
				if i < len(pattern.amplitudes) && pattern.amplitudes[i] != wantAmp {
					t.Errorf("amplitude[%d] = %f, want %f", i, pattern.amplitudes[i], wantAmp)
				}
			}

			if len(pattern.timestamps) != tt.wantNumPhases {
				t.Errorf("timestamps length = %d, want %d", len(pattern.timestamps), tt.wantNumPhases)
			}
		})
	}
}

func TestRhythmicPatternSimilarity(t *testing.T) {
	tests := []struct {
		name      string
		pattern1  *RhythmicPattern
		pattern2  *RhythmicPattern
		wantMin   float64
		wantMax   float64
		description string
	}{
		{
			name: "identical patterns",
			pattern1: NewRhythmicPattern(
				[]float64{0, math.Pi/2, math.Pi},
				[]time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
			),
			pattern2: NewRhythmicPattern(
				[]float64{0, math.Pi/2, math.Pi},
				[]time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
			),
			wantMin:   0.95,
			wantMax:   1.0,
			description: "identical should have high similarity",
		},
		{
			name: "different phases same frequencies",
			pattern1: NewRhythmicPattern(
				[]float64{0, 0, 0},
				[]time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
			),
			pattern2: NewRhythmicPattern(
				[]float64{math.Pi, math.Pi, math.Pi},
				[]time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
			),
			wantMin:   0.85,
			wantMax:   1.0,
			description: "constant phase offset with same frequencies",
		},
		{
			name: "same phases different frequencies",
			pattern1: NewRhythmicPattern(
				[]float64{0, 0, 0},
				[]time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
			),
			pattern2: NewRhythmicPattern(
				[]float64{0, 0, 0},
				[]time.Duration{200 * time.Millisecond, 200 * time.Millisecond, 200 * time.Millisecond},
			),
			wantMin:   0.8,
			wantMax:   0.9,
			description: "same phases with different frequencies",
		},
		{
			name: "completely different",
			pattern1: NewRhythmicPattern(
				[]float64{0, math.Pi/4, math.Pi/2},
				[]time.Duration{50 * time.Millisecond, 75 * time.Millisecond, 100 * time.Millisecond},
			),
			pattern2: NewRhythmicPattern(
				[]float64{math.Pi, 3*math.Pi/2, 2*math.Pi},
				[]time.Duration{200 * time.Millisecond, 300 * time.Millisecond, 400 * time.Millisecond},
			),
			wantMin:   0.0,
			wantMax:   0.7,
			description: "different phases and frequencies",
		},
		{
			name:      "nil pattern",
			pattern1:  NewRhythmicPattern([]float64{0}, []time.Duration{100 * time.Millisecond}),
			pattern2:  nil,
			wantMin:   0,
			wantMax:   0,
			description: "nil pattern should give 0 similarity",
		},
		{
			name:     "empty pattern",
			pattern1: NewRhythmicPattern([]float64{0}, []time.Duration{100 * time.Millisecond}),
			pattern2: &RhythmicPattern{
				phases:      []float64{},
				frequencies: []time.Duration{},
			},
			wantMin:   0,
			wantMax:   0,
			description: "empty pattern should give 0 similarity",
		},
		{
			name: "different lengths",
			pattern1: NewRhythmicPattern(
				[]float64{0, math.Pi/2},
				[]time.Duration{100 * time.Millisecond, 100 * time.Millisecond},
			),
			pattern2: NewRhythmicPattern(
				[]float64{0, math.Pi/2, math.Pi, 3*math.Pi/2},
				[]time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
			),
			wantMin:   0.95,
			wantMax:   1.0,
			description: "should compare only common elements",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			similarity := tt.pattern1.Similarity(tt.pattern2)
			if similarity < tt.wantMin || similarity > tt.wantMax {
				t.Errorf("Similarity() = %f, want in [%f, %f] for %s",
					similarity, tt.wantMin, tt.wantMax, tt.description)
			}
		})
	}
}

func TestPatternDetectGaps(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		pattern   *RhythmicPattern
		threshold time.Duration
		wantGaps  []PatternGap
	}{
		{
			name: "single gap",
			pattern: &RhythmicPattern{
				phases: []float64{0, math.Pi/2, math.Pi, 3*math.Pi/2},
				timestamps: []time.Time{
					now,
					now.Add(100 * time.Millisecond),
					now.Add(300 * time.Millisecond), // 200ms gap
					now.Add(400 * time.Millisecond),
				},
			},
			threshold: 150 * time.Millisecond,
			wantGaps: []PatternGap{
				{StartIdx: 1, EndIdx: 2, Duration: 200 * time.Millisecond},
			},
		},
		{
			name: "multiple gaps",
			pattern: &RhythmicPattern{
				phases: []float64{0, 0.5, 1.0, 1.5, 2.0},
				timestamps: []time.Time{
					now,
					now.Add(100 * time.Millisecond),
					now.Add(350 * time.Millisecond), // 250ms gap
					now.Add(450 * time.Millisecond),
					now.Add(700 * time.Millisecond), // 250ms gap
				},
			},
			threshold: 200 * time.Millisecond,
			wantGaps: []PatternGap{
				{StartIdx: 1, EndIdx: 2, Duration: 250 * time.Millisecond},
				{StartIdx: 3, EndIdx: 4, Duration: 250 * time.Millisecond},
			},
		},
		{
			name: "no gaps",
			pattern: &RhythmicPattern{
				phases: []float64{0, 0.5, 1.0},
				timestamps: []time.Time{
					now,
					now.Add(50 * time.Millisecond),
					now.Add(100 * time.Millisecond),
				},
			},
			threshold: 100 * time.Millisecond,
			wantGaps:  []PatternGap{},
		},
		{
			name: "all gaps",
			pattern: &RhythmicPattern{
				phases: []float64{0, 0.5, 1.0},
				timestamps: []time.Time{
					now,
					now.Add(500 * time.Millisecond),
					now.Add(1000 * time.Millisecond),
				},
			},
			threshold: 400 * time.Millisecond,
			wantGaps: []PatternGap{
				{StartIdx: 0, EndIdx: 1, Duration: 500 * time.Millisecond},
				{StartIdx: 1, EndIdx: 2, Duration: 500 * time.Millisecond},
			},
		},
		{
			name: "empty pattern",
			pattern: &RhythmicPattern{
				phases:     []float64{},
				timestamps: []time.Time{},
			},
			threshold: 100 * time.Millisecond,
			wantGaps:  []PatternGap{},
		},
		{
			name: "single timestamp",
			pattern: &RhythmicPattern{
				phases:     []float64{0},
				timestamps: []time.Time{now},
			},
			threshold: 100 * time.Millisecond,
			wantGaps:  []PatternGap{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gaps := tt.pattern.DetectGaps(tt.threshold)

			if len(gaps) != len(tt.wantGaps) {
				t.Fatalf("got %d gaps, want %d", len(gaps), len(tt.wantGaps))
			}

			for i, wantGap := range tt.wantGaps {
				if gaps[i].StartIdx != wantGap.StartIdx {
					t.Errorf("gap[%d].StartIdx = %d, want %d", i, gaps[i].StartIdx, wantGap.StartIdx)
				}
				if gaps[i].EndIdx != wantGap.EndIdx {
					t.Errorf("gap[%d].EndIdx = %d, want %d", i, gaps[i].EndIdx, wantGap.EndIdx)
				}
				// Allow 1ms tolerance for duration due to time precision
				diff := gaps[i].Duration - wantGap.Duration
				if diff < -time.Millisecond || diff > time.Millisecond {
					t.Errorf("gap[%d].Duration = %v, want %v", i, gaps[i].Duration, wantGap.Duration)
				}
			}
		})
	}
}

func TestPhaseCorrelation(t *testing.T) {
	tests := []struct {
		name      string
		phases1   []float64
		phases2   []float64
		wantMin   float64
		wantMax   float64
		description string
	}{
		{
			name:      "identical phases",
			phases1:   []float64{0, math.Pi/2, math.Pi},
			phases2:   []float64{0, math.Pi/2, math.Pi},
			wantMin:   0.99,
			wantMax:   1.0,
			description: "identical phases should have maximum correlation",
		},
		{
			name:      "constant offset",
			phases1:   []float64{0, 0, 0},
			phases2:   []float64{math.Pi, math.Pi, math.Pi},
			wantMin:   0.9,
			wantMax:   1.0,
			description: "constant phase offset should have high correlation",
		},
		{
			name:      "opposite phases",
			phases1:   []float64{0, math.Pi/2, math.Pi, 3*math.Pi/2},
			phases2:   []float64{math.Pi, 3*math.Pi/2, 0, math.Pi/2},
			wantMin:   0.9,
			wantMax:   1.0,
			description: "phase-shifted pattern",
		},
		{
			name:      "constant offset 1.1 radians",
			phases1:   []float64{0.1, 2.3, 4.5},
			phases2:   []float64{1.2, 3.4, 5.6},
			wantMin:   0.99,
			wantMax:   1.01, // Allow for floating point precision
			description: "constant offset of ~1.1 radians",
		},
		{
			name:      "empty arrays",
			phases1:   []float64{},
			phases2:   []float64{},
			wantMin:   0,
			wantMax:   0,
			description: "empty arrays should have 0 correlation",
		},
		{
			name:      "one empty array",
			phases1:   []float64{0, math.Pi},
			phases2:   []float64{},
			wantMin:   0,
			wantMax:   0,
			description: "one empty array should give 0",
		},
		{
			name:      "different lengths",
			phases1:   []float64{0, math.Pi/2},
			phases2:   []float64{0, math.Pi/2, math.Pi},
			wantMin:   0.99,
			wantMax:   1.0,
			description: "should compare only common length",
		},
		{
			name:      "single element",
			phases1:   []float64{math.Pi/4},
			phases2:   []float64{math.Pi/4},
			wantMin:   0.99,
			wantMax:   1.0,
			description: "single matching element",
		},
		{
			name:      "alternating phases",
			phases1:   []float64{0, math.Pi, 0, math.Pi},
			phases2:   []float64{math.Pi, 0, math.Pi, 0},
			wantMin:   0.9,
			wantMax:   1.0,
			description: "alternating but consistent pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			correlation := phaseCorrelation(tt.phases1, tt.phases2)
			if correlation < tt.wantMin || correlation > tt.wantMax {
				t.Errorf("phaseCorrelation() = %f, want in [%f, %f] for %s",
					correlation, tt.wantMin, tt.wantMax, tt.description)
			}
		})
	}
}

func TestPatternSimilarityWeighting(t *testing.T) {
	tests := []struct {
		name            string
		pattern1        *RhythmicPattern
		pattern2        *RhythmicPattern
		expectedSim     float64
		tolerance       float64
		description     string
	}{
		{
			name: "70-30 weighting test",
			pattern1: NewRhythmicPattern(
				[]float64{0, 0, 0},
				[]time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
			),
			pattern2: NewRhythmicPattern(
				[]float64{0, 0, 0},
				[]time.Duration{200 * time.Millisecond, 200 * time.Millisecond, 200 * time.Millisecond},
			),
			expectedSim: 0.85,
			tolerance:   0.05,
			description: "same phases (weight 0.7), different frequencies",
		},
		{
			name: "different phases same frequencies",
			pattern1: NewRhythmicPattern(
				[]float64{0, math.Pi/2, math.Pi},
				[]time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
			),
			pattern2: NewRhythmicPattern(
				[]float64{math.Pi/4, 3*math.Pi/4, 5*math.Pi/4},
				[]time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
			),
			expectedSim: 1.0,  // Constant offset phases have high correlation
			tolerance:   0.05,
			description: "constant phase offset, same frequencies",
		},
		{
			name: "both different",
			pattern1: NewRhythmicPattern(
				[]float64{0, 0.5, 1.0},
				[]time.Duration{50 * time.Millisecond, 75 * time.Millisecond, 100 * time.Millisecond},
			),
			pattern2: NewRhythmicPattern(
				[]float64{2.0, 2.5, 3.0},
				[]time.Duration{150 * time.Millisecond, 175 * time.Millisecond, 200 * time.Millisecond},
			),
			expectedSim: 0.8,  // Constant offset phases still have high correlation
			tolerance:   0.1,
			description: "constant phase offset, different frequencies",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			similarity := tt.pattern1.Similarity(tt.pattern2)
			diff := math.Abs(similarity - tt.expectedSim)
			if diff > tt.tolerance {
				t.Errorf("Similarity() = %f, want %f ± %f for %s",
					similarity, tt.expectedSim, tt.tolerance, tt.description)
			}
		})
	}
}

func TestRhythmicPatternConcurrency(t *testing.T) {
	pattern1 := NewRhythmicPattern(
		[]float64{0, math.Pi/2, math.Pi},
		[]time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
	)
	pattern2 := NewRhythmicPattern(
		[]float64{math.Pi/4, 3*math.Pi/4, 5*math.Pi/4},
		[]time.Duration{150 * time.Millisecond, 150 * time.Millisecond, 150 * time.Millisecond},
	)

	// Test concurrent similarity calculations
	done := make(chan float64, 10)
	for i := 0; i < 10; i++ {
		go func() {
			sim := pattern1.Similarity(pattern2)
			done <- sim
		}()
	}

	// All calculations should give the same result
	firstResult := <-done
	for i := 0; i < 9; i++ {
		result := <-done
		if math.Abs(result-firstResult) > 0.001 {
			t.Errorf("Concurrent calculations gave different results: %f vs %f", firstResult, result)
		}
	}

	// Test concurrent gap detection
	gapsDone := make(chan int, 10)
	for i := 0; i < 10; i++ {
		go func() {
			gaps := pattern1.DetectGaps(50 * time.Millisecond)
			gapsDone <- len(gaps)
		}()
	}

	// All detections should find the same number of gaps
	firstGaps := <-gapsDone
	for i := 0; i < 9; i++ {
		numGaps := <-gapsDone
		if numGaps != firstGaps {
			t.Errorf("Concurrent gap detection gave different results: %d vs %d", firstGaps, numGaps)
		}
	}
}

func TestRhythmicPatternDurationEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		phases      []float64
		frequencies []time.Duration
		description string
	}{
		{
			name:        "zero duration",
			phases:      []float64{0, math.Pi},
			frequencies: []time.Duration{0, 0},
			description: "handle zero durations",
		},
		{
			name:        "maximum duration",
			phases:      []float64{0, math.Pi},
			frequencies: []time.Duration{time.Duration(math.MaxInt64), time.Duration(math.MaxInt64)},
			description: "handle maximum duration values",
		},
		{
			name:        "negative duration",
			phases:      []float64{0, math.Pi},
			frequencies: []time.Duration{-100 * time.Millisecond, -200 * time.Millisecond},
			description: "handle negative durations",
		},
		{
			name:        "mixed durations",
			phases:      []float64{0, math.Pi/2, math.Pi},
			frequencies: []time.Duration{0, time.Duration(math.MaxInt64), -100 * time.Millisecond},
			description: "handle mixed edge case durations",
		},
		{
			name:        "nanosecond precision",
			phases:      []float64{0, 0.1, 0.2},
			frequencies: []time.Duration{1 * time.Nanosecond, 2 * time.Nanosecond, 3 * time.Nanosecond},
			description: "handle nanosecond precision",
		},
		{
			name:        "very large hours",
			phases:      []float64{0, math.Pi},
			frequencies: []time.Duration{1000000 * time.Hour, 2000000 * time.Hour},
			description: "handle very large durations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			pattern := NewRhythmicPattern(tt.phases, tt.frequencies)
			
			if pattern == nil {
				t.Fatalf("%s: pattern should be created", tt.description)
			}

			// Test with another pattern for similarity
			otherPattern := NewRhythmicPattern(
				[]float64{0, math.Pi},
				[]time.Duration{100 * time.Millisecond, 100 * time.Millisecond},
			)
			
			// Should handle edge cases without panic
			_ = pattern.Similarity(otherPattern)
			
			// Gap detection with various thresholds
			_ = pattern.DetectGaps(0)
			_ = pattern.DetectGaps(time.Duration(math.MaxInt64))
			_ = pattern.DetectGaps(-100 * time.Millisecond)
		})
	}
}

func TestRhythmicPatternNaNInfPhases(t *testing.T) {
	tests := []struct {
		name        string
		phases1     []float64
		phases2     []float64
		frequencies []time.Duration
		description string
	}{
		{
			name:        "NaN phases",
			phases1:     []float64{math.NaN(), math.NaN()},
			phases2:     []float64{0, math.Pi},
			frequencies: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond},
			description: "handle NaN in phases",
		},
		{
			name:        "Inf phases",
			phases1:     []float64{math.Inf(1), math.Inf(-1)},
			phases2:     []float64{0, math.Pi},
			frequencies: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond},
			description: "handle Inf in phases",
		},
		{
			name:        "mixed special values",
			phases1:     []float64{math.NaN(), 0, math.Inf(1), math.Pi, math.Inf(-1)},
			phases2:     []float64{0, 0, 0, 0, 0},
			frequencies: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
			description: "handle mixed NaN and Inf",
		},
		{
			name:        "both patterns with NaN",
			phases1:     []float64{math.NaN(), math.NaN()},
			phases2:     []float64{math.NaN(), math.NaN()},
			frequencies: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond},
			description: "NaN in both patterns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s: unexpected panic: %v", tt.description, r)
				}
			}()

			pattern1 := NewRhythmicPattern(tt.phases1, tt.frequencies)
			pattern2 := NewRhythmicPattern(tt.phases2, tt.frequencies)
			
			// Should handle NaN/Inf without panic
			similarity := pattern1.Similarity(pattern2)
			
			// Similarity with NaN should typically be NaN or 0
			if !math.IsNaN(similarity) && similarity < 0 {
				t.Errorf("%s: similarity should be non-negative or NaN, got %f", tt.description, similarity)
			}
			
			// Test phase correlation directly
			correlation := phaseCorrelation(tt.phases1, tt.phases2)
			_ = correlation // Just ensure no panic
			
			// Gap detection should still work
			_ = pattern1.DetectGaps(100 * time.Millisecond)
		})
	}
}

func TestPhaseCorrelationNegativeValues(t *testing.T) {
	tests := []struct {
		name        string
		phases1     []float64
		phases2     []float64
		wantMin     float64
		wantMax     float64
		description string
	}{
		{
			name:        "all negative phases",
			phases1:     []float64{-math.Pi, -math.Pi/2, -math.Pi/4},
			phases2:     []float64{-math.Pi, -math.Pi/2, -math.Pi/4},
			wantMin:     0.99,
			wantMax:     1.0,
			description: "identical negative phases",
		},
		{
			name:        "negative to positive",
			phases1:     []float64{-math.Pi/2, -math.Pi/4, 0},
			phases2:     []float64{math.Pi/2, 3*math.Pi/4, math.Pi},
			wantMin:     0.9,
			wantMax:     1.0,
			description: "negative to positive with consistent offset",
		},
		{
			name:        "large negative values",
			phases1:     []float64{-10*math.Pi, -9*math.Pi, -8*math.Pi},
			phases2:     []float64{-10*math.Pi + 0.1, -9*math.Pi + 0.1, -8*math.Pi + 0.1},
			wantMin:     0.97,
			wantMax:     1.01,  // Allow slight overshoot due to floating point
			description: "large negative with small offset",
		},
		{
			name:        "negative wrapped values",
			phases1:     []float64{-3*math.Pi/2, -math.Pi, -math.Pi/2},
			phases2:     []float64{math.Pi/2, math.Pi, 3*math.Pi/2},
			wantMin:     0.9,
			wantMax:     1.0,
			description: "negative values that wrap to same as positive",
		},
		{
			name:        "crossing zero correlation",
			phases1:     []float64{-0.5, -0.25, 0, 0.25, 0.5},
			phases2:     []float64{-0.4, -0.15, 0.1, 0.35, 0.6},
			wantMin:     0.98,
			wantMax:     1.0,
			description: "phases crossing zero with small offset",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			correlation := phaseCorrelation(tt.phases1, tt.phases2)
			if correlation < tt.wantMin || correlation > tt.wantMax {
				t.Errorf("%s: phaseCorrelation() = %f, want in [%f, %f]",
					tt.description, correlation, tt.wantMin, tt.wantMax)
			}
			
			// Correlation should always be in [0, 1] range (with small tolerance for floating point)
			if correlation < -0.0001 || correlation > 1.0001 {
				t.Errorf("%s: correlation %f is outside [0, 1] range",
					tt.description, correlation)
			}
		})
	}
}

func TestRhythmicPatternNegativePhases(t *testing.T) {
	tests := []struct {
		name        string
		phases      []float64
		frequencies []time.Duration
		description string
	}{
		{
			name:        "all negative phases",
			phases:      []float64{-math.Pi, -math.Pi/2, -math.Pi/4, -math.Pi/8},
			frequencies: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
			description: "pattern with all negative phases",
		},
		{
			name:        "large negative phases",
			phases:      []float64{-100*math.Pi, -99*math.Pi, -98*math.Pi},
			frequencies: []time.Duration{50 * time.Millisecond, 50 * time.Millisecond, 50 * time.Millisecond},
			description: "extremely large negative phases",
		},
		{
			name:        "mixed negative positive",
			phases:      []float64{-math.Pi, -math.Pi/2, 0, math.Pi/2, math.Pi},
			frequencies: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
			description: "mixed negative and positive phases",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := NewRhythmicPattern(tt.phases, tt.frequencies)
			
			if pattern == nil {
				t.Fatalf("%s: pattern should be created", tt.description)
			}
			
			// Test similarity with negative phases
			otherPattern := NewRhythmicPattern(
				[]float64{0, math.Pi/2, math.Pi},
				[]time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
			)
			
			similarity := pattern.Similarity(otherPattern)
			
			// Similarity should always be in [0, 1] range
			if similarity < 0 || similarity > 1 {
				t.Errorf("%s: similarity %f is outside [0, 1] range",
					tt.description, similarity)
			}
			
			// Test self-similarity with negative phases
			selfSimilarity := pattern.Similarity(pattern)
			if selfSimilarity < 0.99 {
				t.Errorf("%s: self-similarity = %f, should be ~1.0",
					tt.description, selfSimilarity)
			}
		})
	}
}

func TestPatternGapDetectionEdgeCases(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		pattern   *RhythmicPattern
		threshold time.Duration
		description string
	}{
		{
			name: "zero threshold",
			pattern: &RhythmicPattern{
				phases: []float64{0, 0.5, 1.0},
				timestamps: []time.Time{
					now,
					now.Add(1 * time.Nanosecond),
					now.Add(2 * time.Nanosecond),
				},
			},
			threshold: 0,
			description: "zero threshold should find all gaps",
		},
		{
			name: "negative threshold",
			pattern: &RhythmicPattern{
				phases: []float64{0, 0.5, 1.0},
				timestamps: []time.Time{
					now,
					now.Add(100 * time.Millisecond),
					now.Add(200 * time.Millisecond),
				},
			},
			threshold: -100 * time.Millisecond,
			description: "negative threshold behavior",
		},
		{
			name: "maximum duration threshold",
			pattern: &RhythmicPattern{
				phases: []float64{0, 0.5},
				timestamps: []time.Time{
					now,
					now.Add(100 * time.Millisecond),
				},
			},
			threshold: time.Duration(math.MaxInt64),
			description: "maximum threshold should find no gaps",
		},
		{
			name: "timestamps out of order",
			pattern: &RhythmicPattern{
				phases: []float64{0, 0.5, 1.0},
				timestamps: []time.Time{
					now.Add(200 * time.Millisecond),
					now.Add(100 * time.Millisecond),
					now.Add(300 * time.Millisecond),
				},
			},
			threshold: 50 * time.Millisecond,
			description: "handle out-of-order timestamps",
		},
		{
			name: "identical timestamps",
			pattern: &RhythmicPattern{
				phases: []float64{0, 0.5, 1.0},
				timestamps: []time.Time{
					now,
					now,
					now,
				},
			},
			threshold: 0,
			description: "identical timestamps",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s: unexpected panic: %v", tt.description, r)
				}
			}()

			gaps := tt.pattern.DetectGaps(tt.threshold)
			_ = gaps // Just ensure no panic
			
			// Verify gaps structure is valid
			for i, gap := range gaps {
				if gap.StartIdx < 0 || gap.EndIdx < 0 {
					t.Errorf("%s: gap[%d] has negative indices: start=%d, end=%d", 
						tt.description, i, gap.StartIdx, gap.EndIdx)
				}
			}
		})
	}
}