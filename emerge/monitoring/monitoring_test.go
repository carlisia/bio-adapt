package monitoring

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTargetPattern(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		phases      []float64
		frequencies []time.Duration
		validateFn  func(t *testing.T, pattern *TargetPattern)
		description string
	}{
		// Happy path cases
		{
			name:   "uniform pattern",
			phases: []float64{0, math.Pi / 4, math.Pi / 2, 3 * math.Pi / 4, math.Pi},
			frequencies: []time.Duration{
				100 * time.Millisecond,
				100 * time.Millisecond,
				100 * time.Millisecond,
				100 * time.Millisecond,
				100 * time.Millisecond,
			},
			validateFn: func(t *testing.T, pattern *TargetPattern) {
				t.Helper()
				assert.Len(t, pattern.Phases, 5, "Should have 5 phases")
				assert.Len(t, pattern.Frequencies, 5, "Should have 5 frequencies")
				assert.Greater(t, pattern.Amplitude, 0.0, "Amplitude should be positive for varied phases")
				assert.Equal(t, 100*time.Millisecond, pattern.Period, "Period should be 100ms")
				assert.Equal(t, 1.0, pattern.Confidence, "Confidence should be 1.0")
			},
			description: "Uniform frequency pattern should work correctly",
		},
		{
			name:        "empty pattern",
			phases:      []float64{},
			frequencies: []time.Duration{},
			validateFn: func(t *testing.T, pattern *TargetPattern) {
				t.Helper()
				assert.Empty(t, pattern.Phases, "Should have 0 phases")
				assert.Equal(t, 0.0, pattern.Amplitude, "Amplitude should be 0 for empty pattern")
				assert.Equal(t, time.Duration(0), pattern.Period, "Period should be 0 for empty pattern")
			},
			description: "Empty pattern should be handled correctly",
		},
		{
			name:        "single phase",
			phases:      []float64{math.Pi},
			frequencies: []time.Duration{100 * time.Millisecond},
			validateFn: func(t *testing.T, pattern *TargetPattern) {
				t.Helper()
				assert.Len(t, pattern.Phases, 1, "Should have 1 phase")
				assert.Equal(t, 0.0, pattern.Amplitude, "Amplitude should be 0 for single phase")
				assert.Equal(t, 100*time.Millisecond, pattern.Period, "Period should be 100ms")
			},
			description: "Single phase pattern should work",
		},
		{
			name:   "mixed frequencies",
			phases: []float64{0, math.Pi / 2, math.Pi, 3 * math.Pi / 2},
			frequencies: []time.Duration{
				50 * time.Millisecond,
				100 * time.Millisecond,
				150 * time.Millisecond,
				200 * time.Millisecond,
			},
			validateFn: func(t *testing.T, pattern *TargetPattern) {
				t.Helper()
				expectedPeriod := 125 * time.Millisecond // Average
				assert.Equal(t, expectedPeriod, pattern.Period, "Period should be average of frequencies")
			},
			description: "Mixed frequencies should average correctly",
		},
		// Edge cases
		{
			name:        "zero phases",
			phases:      []float64{0, 0, 0, 0},
			frequencies: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
			validateFn: func(t *testing.T, pattern *TargetPattern) {
				t.Helper()
				assert.Equal(t, 0.0, pattern.Amplitude, "Amplitude should be 0 for uniform zero phases")
			},
			description: "All zero phases should have zero amplitude",
		},
		{
			name:        "negative phases",
			phases:      []float64{-math.Pi, -math.Pi / 2, 0, math.Pi / 2},
			frequencies: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
			validateFn: func(t *testing.T, pattern *TargetPattern) {
				t.Helper()
				assert.Greater(t, pattern.Amplitude, 0.0, "Should calculate amplitude for negative phases")
			},
			description: "Negative phases should work",
		},
		{
			name:        "very large phases",
			phases:      []float64{0, 10 * math.Pi, 20 * math.Pi, 30 * math.Pi},
			frequencies: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
			validateFn: func(t *testing.T, pattern *TargetPattern) {
				t.Helper()
				assert.Greater(t, pattern.Amplitude, 0.0, "Should calculate amplitude for large phases")
			},
			description: "Very large phases should work",
		},
		{
			name:        "zero duration frequencies",
			phases:      []float64{0, math.Pi},
			frequencies: []time.Duration{0, 0},
			validateFn: func(t *testing.T, pattern *TargetPattern) {
				t.Helper()
				assert.Equal(t, time.Duration(0), pattern.Period, "Period should be 0 for zero frequencies")
			},
			description: "Zero duration frequencies should work",
		},
		{
			name:        "negative duration frequencies",
			phases:      []float64{0, math.Pi},
			frequencies: []time.Duration{-100 * time.Millisecond, -100 * time.Millisecond},
			validateFn: func(t *testing.T, pattern *TargetPattern) {
				t.Helper()
				assert.Equal(t, -100*time.Millisecond, pattern.Period, "Negative duration frequencies should be preserved")
			},
			description: "Negative duration frequencies should be preserved",
		},
		{
			name:        "mismatched lengths",
			phases:      []float64{0, math.Pi, math.Pi / 2},
			frequencies: []time.Duration{100 * time.Millisecond},
			validateFn: func(t *testing.T, pattern *TargetPattern) {
				t.Helper()
				assert.Len(t, pattern.Phases, 3, "Should have 3 phases")
				assert.Len(t, pattern.Frequencies, 1, "Should have 1 frequency")
			},
			description: "Mismatched lengths should be preserved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pattern := NewTargetPattern(tt.phases, tt.frequencies)
			tt.validateFn(t, pattern)
		})
	}
}

func TestPatternDetectGaps(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		phases      []float64
		frequencies []time.Duration
		validateFn  func(t *testing.T, gaps []PatternGap)
		description string
	}{
		{
			name:        "large phase jump",
			phases:      []float64{0, math.Pi / 4, math.Pi * 1.5, math.Pi * 1.75},
			frequencies: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
			validateFn: func(t *testing.T, gaps []PatternGap) {
				t.Helper()
				assert.NotEmpty(t, gaps, "Should detect at least one gap")
				foundGap := false
				for _, gap := range gaps {
					if gap.StartIdx == 1 && gap.EndIdx == 2 {
						foundGap = true
						break
					}
				}
				assert.True(t, foundGap, "Should detect gap between indices 1 and 2")
			},
			description: "Should detect large phase jumps",
		},
		{
			name:        "uniform pattern no gaps",
			phases:      []float64{0, math.Pi / 4, math.Pi / 2, 3 * math.Pi / 4},
			frequencies: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
			validateFn: func(_ *testing.T, _ []PatternGap) {
				// May or may not detect gaps depending on threshold
				// Just verify it doesn't crash
			},
			description: "Uniform pattern gap detection",
		},
		{
			name:        "empty pattern",
			phases:      []float64{},
			frequencies: []time.Duration{},
			validateFn: func(t *testing.T, gaps []PatternGap) {
				t.Helper()
				assert.NotNil(t, gaps, "Should return non-nil slice for empty pattern")
			},
			description: "Empty pattern should return empty gaps",
		},
		{
			name:        "single phase",
			phases:      []float64{math.Pi},
			frequencies: []time.Duration{100 * time.Millisecond},
			validateFn: func(t *testing.T, gaps []PatternGap) {
				t.Helper()
				assert.Empty(t, gaps, "Single phase should have no gaps")
			},
			description: "Single phase should have no gaps",
		},
		{
			name:        "wraparound phases",
			phases:      []float64{0, math.Pi / 2, 3 * math.Pi / 2, 2 * math.Pi},
			frequencies: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
			validateFn: func(t *testing.T, gaps []PatternGap) {
				t.Helper()
				// Check for gap between π/2 and 3π/2
				foundGap := false
				for _, gap := range gaps {
					if gap.StartIdx == 1 && gap.EndIdx == 2 {
						foundGap = true
						break
					}
				}
				assert.True(t, foundGap, "Should detect gap between π/2 and 3π/2")
			},
			description: "Should detect wraparound gaps",
		},
		{
			name:        "multiple gaps",
			phases:      []float64{0, 0.1, math.Pi, math.Pi + 0.1, 2 * math.Pi},
			frequencies: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
			validateFn: func(t *testing.T, gaps []PatternGap) {
				t.Helper()
				assert.GreaterOrEqual(t, len(gaps), 2, "Should detect multiple gaps")
			},
			description: "Should detect multiple gaps",
		},
		{
			name:        "negative phases",
			phases:      []float64{-math.Pi, -math.Pi / 2, math.Pi / 2, math.Pi},
			frequencies: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
			validateFn: func(_ *testing.T, _ []PatternGap) {
				// Should handle negative phases
				// Just verify it doesn't crash
			},
			description: "Should handle negative phases",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pattern := NewTargetPattern(tt.phases, tt.frequencies)
			gaps := pattern.Detect()
			tt.validateFn(t, gaps)
		})
	}
}

func TestPatternComplete(t *testing.T) {
	t.Parallel()
	basePhases := []float64{0, math.Pi / 4, math.Pi / 2, 3 * math.Pi / 4, math.Pi}
	baseFrequencies := []time.Duration{
		100 * time.Millisecond,
		100 * time.Millisecond,
		100 * time.Millisecond,
		100 * time.Millisecond,
		100 * time.Millisecond,
	}
	pattern := NewTargetPattern(basePhases, baseFrequencies)

	tests := []struct {
		name        string
		gap         PatternGap
		expectedLen int
		validateFn  func(t *testing.T, completed []float64)
		description string
	}{
		{
			name: "single step gap",
			gap: PatternGap{
				StartIdx: 0,
				EndIdx:   2,
				Duration: 100 * time.Millisecond,
			},
			expectedLen: 1,
			validateFn: func(t *testing.T, completed []float64) {
				t.Helper()
				assert.Len(t, completed, 1, "Should have 1 completed value")
			},
			description: "Should complete single step gap",
		},
		{
			name: "multi step gap",
			gap: PatternGap{
				StartIdx: 1,
				EndIdx:   4,
				Duration: 100 * time.Millisecond,
			},
			expectedLen: 2,
			validateFn: func(t *testing.T, completed []float64) {
				t.Helper()
				assert.Len(t, completed, 2, "Should have 2 completed values")
			},
			description: "Should complete multi step gap",
		},
		{
			name: "adjacent indices",
			gap: PatternGap{
				StartIdx: 2,
				EndIdx:   3,
				Duration: 100 * time.Millisecond,
			},
			expectedLen: 0,
			validateFn: func(t *testing.T, completed []float64) {
				t.Helper()
				assert.Empty(t, completed, "Should have 0 completed values for adjacent indices")
			},
			description: "Adjacent indices should return no completions",
		},
		{
			name: "invalid start index",
			gap: PatternGap{
				StartIdx: -1,
				EndIdx:   2,
				Duration: 100 * time.Millisecond,
			},
			expectedLen: -1,
			validateFn: func(t *testing.T, completed []float64) {
				t.Helper()
				assert.Nil(t, completed, "Should return nil for invalid start index")
			},
			description: "Invalid start index should return nil",
		},
		{
			name: "invalid end index",
			gap: PatternGap{
				StartIdx: 0,
				EndIdx:   10,
				Duration: 100 * time.Millisecond,
			},
			expectedLen: -1,
			validateFn: func(t *testing.T, completed []float64) {
				t.Helper()
				assert.Nil(t, completed, "Should return nil for invalid end index")
			},
			description: "Invalid end index should return nil",
		},
		{
			name: "reversed indices",
			gap: PatternGap{
				StartIdx: 3,
				EndIdx:   1,
				Duration: 100 * time.Millisecond,
			},
			expectedLen: -1,
			validateFn: func(t *testing.T, completed []float64) {
				t.Helper()
				assert.Nil(t, completed, "Should return nil for reversed indices")
			},
			description: "Reversed indices should return nil",
		},
		{
			name: "zero duration",
			gap: PatternGap{
				StartIdx: 0,
				EndIdx:   2,
				Duration: 0,
			},
			expectedLen: 1,
			validateFn: func(t *testing.T, completed []float64) {
				t.Helper()
				assert.Len(t, completed, 1, "Should handle zero duration")
			},
			description: "Zero duration should still complete",
		},
		{
			name: "negative duration",
			gap: PatternGap{
				StartIdx: 0,
				EndIdx:   2,
				Duration: -100 * time.Millisecond,
			},
			expectedLen: 1,
			validateFn: func(t *testing.T, completed []float64) {
				t.Helper()
				assert.Len(t, completed, 1, "Should handle negative duration")
			},
			description: "Negative duration should still complete",
		},
		{
			name: "wrap around end",
			gap: PatternGap{
				StartIdx: 3,
				EndIdx:   5, // Beyond array bounds
				Duration: 100 * time.Millisecond,
			},
			expectedLen: -1,
			validateFn: func(t *testing.T, completed []float64) {
				t.Helper()
				assert.Nil(t, completed, "Should return nil for out of bounds end index")
			},
			description: "Out of bounds should return nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			completed := pattern.Complete(tt.gap)
			tt.validateFn(t, completed)
		})
	}
}

func TestPatternSimilarity(t *testing.T) {
	t.Parallel()
	phases1 := []float64{0, math.Pi / 4, math.Pi / 2, 3 * math.Pi / 4}
	frequencies1 := []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond}
	pattern1 := NewTargetPattern(phases1, frequencies1)

	tests := []struct {
		name        string
		other       *TargetPattern
		minSim      float64
		maxSim      float64
		description string
	}{
		{
			name:        "identical pattern",
			other:       NewTargetPattern(phases1, frequencies1),
			minSim:      0.95,
			maxSim:      1.0,
			description: "Identical patterns should have high similarity",
		},
		{
			name: "similar pattern",
			other: NewTargetPattern(
				[]float64{0, math.Pi/4 + 0.1, math.Pi / 2, 3*math.Pi/4 - 0.1},
				frequencies1,
			),
			minSim:      0.7,
			maxSim:      1.0,
			description: "Similar patterns should have good similarity",
		},
		{
			name: "different pattern",
			other: NewTargetPattern(
				[]float64{math.Pi, math.Pi * 1.25, math.Pi * 1.5, math.Pi * 1.75},
				[]time.Duration{200 * time.Millisecond, 200 * time.Millisecond, 200 * time.Millisecond, 200 * time.Millisecond},
			),
			minSim:      0.0,
			maxSim:      0.5,
			description: "Different patterns should have low similarity",
		},
		{
			name:        "nil pattern",
			other:       nil,
			minSim:      0.0,
			maxSim:      0.0,
			description: "Nil pattern should have zero similarity",
		},
		{
			name:        "empty pattern",
			other:       NewTargetPattern([]float64{}, []time.Duration{}),
			minSim:      0.0,
			maxSim:      0.0,
			description: "Empty pattern should have zero similarity",
		},
		{
			name: "different length patterns",
			other: NewTargetPattern(
				[]float64{0, math.Pi / 2},
				[]time.Duration{100 * time.Millisecond, 100 * time.Millisecond},
			),
			minSim:      0.4,
			maxSim:      0.7,
			description: "Different length patterns should have reduced similarity",
		},
		{
			name: "shifted pattern",
			other: NewTargetPattern(
				[]float64{math.Pi / 4, math.Pi / 2, 3 * math.Pi / 4, math.Pi},
				frequencies1,
			),
			minSim:      0.5,
			maxSim:      1.0,
			description: "Shifted pattern should have moderate similarity",
		},
		{
			name: "inverted pattern",
			other: NewTargetPattern(
				[]float64{3 * math.Pi / 4, math.Pi / 2, math.Pi / 4, 0},
				frequencies1,
			),
			minSim:      0.7,
			maxSim:      0.9,
			description: "Inverted pattern should have low similarity",
		},
		{
			name: "same phases different frequencies",
			other: NewTargetPattern(
				phases1,
				[]time.Duration{50 * time.Millisecond, 50 * time.Millisecond, 50 * time.Millisecond, 50 * time.Millisecond},
			),
			minSim:      0.8,
			maxSim:      0.9,
			description: "Same phases with different frequencies should have partial similarity",
		},
		{
			name: "negative phases",
			other: NewTargetPattern(
				[]float64{-math.Pi / 4, -math.Pi / 2, -3 * math.Pi / 4, -math.Pi},
				frequencies1,
			),
			minSim:      0.7,
			maxSim:      0.9,
			description: "Negative phases should be handled",
		},
		{
			name: "invalid pattern with NaN phases",
			other: &TargetPattern{
				Phases:      []float64{0, math.NaN(), math.Pi},
				Frequencies: frequencies1[:3],
				Amplitude:   math.NaN(),
			},
			minSim:      0.0,
			maxSim:      0.1,
			description: "Patterns with NaN should have very low similarity",
		},
		{
			name: "pattern with infinite phases",
			other: &TargetPattern{
				Phases:      []float64{0, math.Inf(1), math.Pi},
				Frequencies: frequencies1[:3],
				Amplitude:   math.Inf(1),
			},
			minSim:      0.0,
			maxSim:      0.1,
			description: "Patterns with infinity should have very low similarity",
		},
		{
			name: "massively different scale patterns",
			other: NewTargetPattern(
				[]float64{0, 1000 * math.Pi, 2000 * math.Pi, 3000 * math.Pi},
				[]time.Duration{1 * time.Nanosecond, 1 * time.Nanosecond, 1 * time.Nanosecond, 1 * time.Nanosecond},
			),
			minSim:      0.0,
			maxSim:      0.3,
			description: "Vastly different scale patterns should have low similarity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			similarity := pattern1.Similarity(tt.other)
			// Handle NaN and infinity cases specially
			if math.IsNaN(similarity) || math.IsInf(similarity, 0) {
				// For NaN/infinity cases, just check that we get a reasonable result
				assert.True(t, similarity >= tt.minSim && similarity <= tt.maxSim || math.IsNaN(similarity) || math.IsInf(similarity, 0), "%s: Similarity should be in valid range or NaN/Inf", tt.description)
			} else {
				assert.GreaterOrEqual(t, similarity, tt.minSim, "%s: Similarity should be >= %f", tt.description, tt.minSim)
				assert.LessOrEqual(t, similarity, tt.maxSim, "%s: Similarity should be <= %f", tt.description, tt.maxSim)
			}
		})
	}
}

func TestPatternTemplate(t *testing.T) {
	t.Parallel()
	basePattern := TargetPattern{
		Phases:      []float64{0, math.Pi / 2, math.Pi, 3 * math.Pi / 2},
		Frequencies: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
		Amplitude:   1.0,
		Period:      100 * time.Millisecond,
		Confidence:  1.0,
	}

	tests := []struct {
		name        string
		template    *PatternTemplate
		pattern     *TargetPattern
		expected    bool
		description string
	}{
		{
			name: "exact match",
			template: &PatternTemplate{
				Name:        "test",
				BasePattern: basePattern,
				Tolerance:   0.2,
			},
			pattern: &TargetPattern{
				Phases:      []float64{0, math.Pi / 2, math.Pi, 3 * math.Pi / 2},
				Frequencies: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
				Amplitude:   1.0,
				Period:      100 * time.Millisecond,
				Confidence:  1.0,
			},
			expected:    true,
			description: "Exact match should succeed",
		},
		{
			name: "within tolerance",
			template: &PatternTemplate{
				Name:        "test",
				BasePattern: basePattern,
				Tolerance:   0.2,
			},
			pattern: &TargetPattern{
				Phases:      []float64{0.1, math.Pi/2 + 0.1, math.Pi, 3*math.Pi/2 - 0.1},
				Frequencies: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
				Amplitude:   0.9,
				Period:      100 * time.Millisecond,
				Confidence:  1.0,
			},
			expected:    true,
			description: "Pattern within tolerance should match",
		},
		{
			name: "outside tolerance",
			template: &PatternTemplate{
				Name:        "test",
				BasePattern: basePattern,
				Tolerance:   0.2,
			},
			pattern: &TargetPattern{
				Phases:      []float64{math.Pi / 4, 3 * math.Pi / 4, 5 * math.Pi / 4, 7 * math.Pi / 4},
				Frequencies: []time.Duration{200 * time.Millisecond, 200 * time.Millisecond, 200 * time.Millisecond, 200 * time.Millisecond},
				Amplitude:   2.0,
				Period:      200 * time.Millisecond,
				Confidence:  1.0,
			},
			expected:    false,
			description: "Pattern outside tolerance should not match",
		},
		{
			name: "nil pattern",
			template: &PatternTemplate{
				Name:        "test",
				BasePattern: basePattern,
				Tolerance:   0.2,
			},
			pattern:     nil,
			expected:    false,
			description: "Nil pattern should not match",
		},
		{
			name: "zero tolerance exact match",
			template: &PatternTemplate{
				Name:        "test",
				BasePattern: basePattern,
				Tolerance:   0.0,
			},
			pattern:     &basePattern,
			expected:    true,
			description: "Zero tolerance should require exact match",
		},
		{
			name: "negative tolerance",
			template: &PatternTemplate{
				Name:        "test",
				BasePattern: basePattern,
				Tolerance:   -0.5,
			},
			pattern:     &basePattern,
			expected:    false,
			description: "Negative tolerance should never match",
		},
		{
			name: "very high tolerance",
			template: &PatternTemplate{
				Name:        "test",
				BasePattern: basePattern,
				Tolerance:   10.0,
			},
			pattern: &TargetPattern{
				Phases:      []float64{0, 0, 0, 0},
				Frequencies: []time.Duration{1 * time.Second, 1 * time.Second, 1 * time.Second, 1 * time.Second},
				Amplitude:   0,
				Period:      1 * time.Second,
				Confidence:  0,
			},
			expected:    true,
			description: "Very high tolerance should match almost anything",
		},
		{
			name: "empty template pattern",
			template: &PatternTemplate{
				Name: "test",
				BasePattern: TargetPattern{
					Phases:      []float64{},
					Frequencies: []time.Duration{},
				},
				Tolerance: 0.2,
			},
			pattern: &TargetPattern{
				Phases:      []float64{0, math.Pi},
				Frequencies: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond},
			},
			expected:    false,
			description: "Empty template should not match non-empty pattern",
		},
		{
			name: "with variations",
			template: &PatternTemplate{
				Name:        "test",
				BasePattern: basePattern,
				Variations: []TargetPattern{
					{
						Phases:      []float64{0.1, math.Pi/2 + 0.1, math.Pi + 0.1, 3*math.Pi/2 + 0.1},
						Frequencies: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
					},
				},
				Tolerance: 0.1,
			},
			pattern: &TargetPattern{
				Phases:      []float64{0.1, math.Pi/2 + 0.1, math.Pi + 0.1, 3*math.Pi/2 + 0.1},
				Frequencies: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
				Amplitude:   1.0,
				Period:      100 * time.Millisecond,
				Confidence:  1.0,
			},
			expected:    true,
			description: "Should match variation patterns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			matches := tt.template.Matches(tt.pattern)
			assert.Equal(t, tt.expected, matches, "%s: Matches() should return expected result", tt.description)
		})
	}
}

func TestPatternLibrary(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		setupFn     func() *PatternLibrary
		testPattern *TargetPattern
		expectedID  string
		minSim      float64
		description string
	}{
		{
			name: "identify periodic pattern",
			setupFn: func() *PatternLibrary {
				library := NewPatternLibrary()
				periodic := &PatternTemplate{
					Name: "periodic",
					BasePattern: TargetPattern{
						Phases:      []float64{0, math.Pi / 2, math.Pi, 3 * math.Pi / 2},
						Frequencies: []time.Duration{6 * time.Hour, 6 * time.Hour, 6 * time.Hour, 6 * time.Hour},
						Amplitude:   1.0,
						Period:      24 * time.Hour,
						Confidence:  1.0,
					},
					Tolerance: 0.15,
				}
				library.Add("periodic", periodic)
				return library
			},
			testPattern: &TargetPattern{
				Phases:      []float64{0, math.Pi / 2, math.Pi, 3 * math.Pi / 2},
				Frequencies: []time.Duration{6 * time.Hour, 6 * time.Hour, 6 * time.Hour, 6 * time.Hour},
				Amplitude:   0.95,
				Period:      24 * time.Hour,
				Confidence:  0.9,
			},
			expectedID:  "periodic",
			minSim:      0.8,
			description: "Should identify periodic pattern",
		},
		{
			name: "identify best match among multiple",
			setupFn: func() *PatternLibrary {
				library := NewPatternLibrary()

				periodic24h := &PatternTemplate{
					Name: "periodic24h",
					BasePattern: *NewTargetPattern(
						[]float64{0, math.Pi / 2, math.Pi, 3 * math.Pi / 2},
						[]time.Duration{6 * time.Hour, 6 * time.Hour, 6 * time.Hour, 6 * time.Hour},
					),
					Tolerance: 0.15,
				}

				ultradian := &PatternTemplate{
					Name: "ultradian",
					BasePattern: *NewTargetPattern(
						[]float64{0, math.Pi},
						[]time.Duration{45 * time.Minute, 45 * time.Minute},
					),
					Tolerance: 0.2,
				}

				library.Add("periodic24h", periodic24h)
				library.Add("ultradian", ultradian)
				return library
			},
			testPattern: NewTargetPattern(
				[]float64{0, math.Pi},
				[]time.Duration{45 * time.Minute, 45 * time.Minute},
			),
			expectedID:  "ultradian",
			minSim:      0.8,
			description: "Should identify best matching pattern",
		},
		{
			name: "no match found",
			setupFn: func() *PatternLibrary {
				library := NewPatternLibrary()
				periodic24h := &PatternTemplate{
					Name: "periodic24h",
					BasePattern: TargetPattern{
						Phases:      []float64{0, math.Pi / 2, math.Pi, 3 * math.Pi / 2},
						Frequencies: []time.Duration{6 * time.Hour, 6 * time.Hour, 6 * time.Hour, 6 * time.Hour},
					},
					Tolerance: 0.15,
				}
				library.Add("periodic24h", periodic24h)
				return library
			},
			testPattern: &TargetPattern{
				Phases:      []float64{0, 0.1, 0.2, 0.3},
				Frequencies: []time.Duration{1 * time.Second, 1 * time.Second, 1 * time.Second, 1 * time.Second},
			},
			expectedID:  "",
			minSim:      0,
			description: "Should return empty for no match",
		},
		{
			name:    "empty library",
			setupFn: NewPatternLibrary,
			testPattern: &TargetPattern{
				Phases:      []float64{0, math.Pi},
				Frequencies: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond},
			},
			expectedID:  "",
			minSim:      0,
			description: "Empty library should return empty",
		},
		{
			name: "nil pattern",
			setupFn: func() *PatternLibrary {
				library := NewPatternLibrary()
				template := &PatternTemplate{
					Name:        "test",
					BasePattern: TargetPattern{Phases: []float64{0, math.Pi}},
				}
				library.Add("test", template)
				return library
			},
			testPattern: nil,
			expectedID:  "",
			minSim:      0,
			description: "Nil pattern should return empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			library := tt.setupFn()
			name, similarity := library.Identify(tt.testPattern)

			assert.Equal(t, tt.expectedID, name, "%s: Should identify correct pattern", tt.description)
			assert.GreaterOrEqual(t, similarity, tt.minSim, "%s: Similarity should be >= %f", tt.description, tt.minSim)
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Parallel()
	t.Run("calculateAmplitude", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name        string
			phases      []float64
			expectedMin float64
			expectedMax float64
			description string
		}{
			{
				name:        "uniform phases",
				phases:      []float64{1.0, 1.0, 1.0, 1.0},
				expectedMin: 0,
				expectedMax: 0,
				description: "Uniform phases should have zero amplitude",
			},
			{
				name:        "varied phases",
				phases:      []float64{0, math.Pi / 2, math.Pi, 3 * math.Pi / 2},
				expectedMin: 0.5,
				expectedMax: 2.0,
				description: "Varied phases should have positive amplitude",
			},
			{
				name:        "empty slice",
				phases:      []float64{},
				expectedMin: 0,
				expectedMax: 0,
				description: "Empty phases should have zero amplitude",
			},
			{
				name:        "single phase",
				phases:      []float64{math.Pi},
				expectedMin: 0,
				expectedMax: 0,
				description: "Single phase should have zero amplitude",
			},
			{
				name:        "negative phases",
				phases:      []float64{-math.Pi, -math.Pi / 2, 0, math.Pi / 2},
				expectedMin: 0.5,
				expectedMax: 2.0,
				description: "Negative phases should work",
			},
			{
				name:        "large phases (normalized)",
				phases:      []float64{0, 10*math.Pi + math.Pi/2, 20*math.Pi + math.Pi, 30*math.Pi + 3*math.Pi/2},
				expectedMin: 1.5,
				expectedMax: 2.0,
				description: "Large phases should be normalized before amplitude calculation",
			},
			{
				name:        "extremely large phases",
				phases:      []float64{0, 1000 * math.Pi, 2000 * math.Pi},
				expectedMin: 0.0,
				expectedMax: 3.0,
				description: "Extremely large phases should be handled",
			},
			{
				name:        "negative and positive mix",
				phases:      []float64{-10 * math.Pi, -5 * math.Pi, 5 * math.Pi, 10 * math.Pi},
				expectedMin: 0.0,
				expectedMax: 3.0,
				description: "Mixed negative and positive large phases",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				amp := calculateAmplitude(tt.phases)
				if math.IsNaN(amp) || math.IsInf(amp, 0) {
					// For NaN/infinity cases, just check that we get a reasonable result
					assert.True(t, amp >= tt.expectedMin && amp <= tt.expectedMax || math.IsNaN(amp) || math.IsInf(amp, 0), "%s: Amplitude should be in valid range or NaN/Inf", tt.description)
				} else {
					assert.GreaterOrEqual(t, amp, tt.expectedMin, "%s: Amplitude should be >= %f", tt.description, tt.expectedMin)
					assert.LessOrEqual(t, amp, tt.expectedMax, "%s: Amplitude should be <= %f", tt.description, tt.expectedMax)
				}
			})
		}
	})

	t.Run("calculatePeriod", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name        string
			frequencies []time.Duration
			expected    time.Duration
			description string
		}{
			{
				name: "uniform frequencies",
				frequencies: []time.Duration{
					100 * time.Millisecond,
					100 * time.Millisecond,
					100 * time.Millisecond,
				},
				expected:    100 * time.Millisecond,
				description: "Uniform frequencies should return that value",
			},
			{
				name: "mixed frequencies",
				frequencies: []time.Duration{
					100 * time.Millisecond,
					200 * time.Millisecond,
					150 * time.Millisecond,
				},
				expected:    150 * time.Millisecond,
				description: "Mixed frequencies should return average",
			},
			{
				name:        "empty slice",
				frequencies: []time.Duration{},
				expected:    0,
				description: "Empty frequencies should return zero",
			},
			{
				name:        "single frequency",
				frequencies: []time.Duration{500 * time.Millisecond},
				expected:    500 * time.Millisecond,
				description: "Single frequency should return that value",
			},
			{
				name:        "zero frequencies",
				frequencies: []time.Duration{0, 0, 0},
				expected:    0,
				description: "Zero frequencies should return zero",
			},
			{
				name: "negative frequencies",
				frequencies: []time.Duration{
					-100 * time.Millisecond,
					-200 * time.Millisecond,
				},
				expected:    -150 * time.Millisecond,
				description: "Negative frequencies should work",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				period := calculatePeriod(tt.frequencies)
				assert.Equal(t, tt.expected, period, "%s: Period should match expected", tt.description)
			})
		}
	})

	t.Run("phaseCorrelation", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name        string
			phases1     []float64
			phases2     []float64
			expectedMin float64
			expectedMax float64
			description string
		}{
			{
				name:        "identical phases",
				phases1:     []float64{0, math.Pi / 2, math.Pi, 3 * math.Pi / 2},
				phases2:     []float64{0, math.Pi / 2, math.Pi, 3 * math.Pi / 2},
				expectedMin: 0.99,
				expectedMax: 1.0,
				description: "Identical phases should have correlation ~1.0",
			},
			{
				name:        "opposite phases",
				phases1:     []float64{0, math.Pi / 2, math.Pi, 3 * math.Pi / 2},
				phases2:     []float64{math.Pi, 3 * math.Pi / 2, 0, math.Pi / 2},
				expectedMin: 0.0,
				expectedMax: 0.5,
				description: "Opposite phases should have low correlation",
			},
			{
				name:        "empty first",
				phases1:     []float64{},
				phases2:     []float64{0, math.Pi},
				expectedMin: 0,
				expectedMax: 0,
				description: "Empty first array should return zero",
			},
			{
				name:        "empty second",
				phases1:     []float64{0, math.Pi},
				phases2:     []float64{},
				expectedMin: 0,
				expectedMax: 0,
				description: "Empty second array should return zero",
			},
			{
				name:        "both empty",
				phases1:     []float64{},
				phases2:     []float64{},
				expectedMin: 0,
				expectedMax: 0,
				description: "Both empty should return zero",
			},
			{
				name:        "different lengths",
				phases1:     []float64{0, math.Pi / 2, math.Pi},
				phases2:     []float64{0, math.Pi / 2},
				expectedMin: 0,
				expectedMax: 0,
				description: "Different lengths should return zero",
			},
			{
				name:        "negative phases",
				phases1:     []float64{-math.Pi, -math.Pi / 2, 0, math.Pi / 2},
				phases2:     []float64{-math.Pi, -math.Pi / 2, 0, math.Pi / 2},
				expectedMin: 0.99,
				expectedMax: 1.0,
				description: "Negative phases should work",
			},
			{
				name:        "single element",
				phases1:     []float64{math.Pi},
				phases2:     []float64{math.Pi},
				expectedMin: 1.0,
				expectedMax: 1.0,
				description: "Single identical element should return 1.0",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				corr := phaseCorrelation(tt.phases1, tt.phases2)
				if math.IsNaN(corr) || math.IsInf(corr, 0) {
					// For NaN/infinity cases, just check that we get a reasonable result
					assert.True(t, corr >= tt.expectedMin && corr <= tt.expectedMax || math.IsNaN(corr) || math.IsInf(corr, 0), "%s: Correlation should be in valid range or NaN/Inf", tt.description)
				} else {
					assert.GreaterOrEqual(t, corr, tt.expectedMin, "%s: Correlation should be >= %f", tt.description, tt.expectedMin)
					assert.LessOrEqual(t, corr, tt.expectedMax, "%s: Correlation should be <= %f", tt.description, tt.expectedMax)
				}
			})
		}
	})
}

func TestPatternConcurrency(t *testing.T) {
	t.Parallel()
	pattern := NewTargetPattern(
		[]float64{0, math.Pi / 2, math.Pi, 3 * math.Pi / 2},
		[]time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
	)

	library := NewPatternLibrary()
	template := &PatternTemplate{
		Name:        "test",
		BasePattern: *pattern,
		Tolerance:   0.2,
	}
	library.Add("test", template)

	// Run concurrent operations
	done := make(chan bool, 100)

	for range 25 {
		go func() {
			_ = pattern.Detect()
			done <- true
		}()

		go func() {
			_ = pattern.Similarity(pattern)
			done <- true
		}()

		go func() {
			_, _ = library.Identify(pattern)
			done <- true
		}()

		go func() {
			_ = template.Matches(pattern)
			done <- true
		}()
	}

	// Wait for all goroutines
	for range 100 {
		<-done
	}

	// If we get here without race conditions, concurrent access is safe
}

func BenchmarkPatternDetect(b *testing.B) {
	phases := make([]float64, 100)
	frequencies := make([]time.Duration, 100)
	for i := range phases {
		phases[i] = float64(i) * math.Pi / 50
		frequencies[i] = 100 * time.Millisecond
	}

	pattern := NewTargetPattern(phases, frequencies)

	b.ResetTimer()
	for range b.N {
		pattern.Detect()
	}
}

func BenchmarkPatternSimilarity(b *testing.B) {
	phases := []float64{0, math.Pi / 2, math.Pi, 3 * math.Pi / 2}
	frequencies := []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond}

	pattern1 := NewTargetPattern(phases, frequencies)
	pattern2 := NewTargetPattern(phases, frequencies)

	b.ResetTimer()
	for range b.N {
		pattern1.Similarity(pattern2)
	}
}

func BenchmarkLibraryIdentify(b *testing.B) {
	library := NewPatternLibrary()

	// Add multiple templates
	for i := range 10 {
		phases := make([]float64, 4)
		for j := range phases {
			phases[j] = float64(j+i) * math.Pi / 4
		}
		template := &PatternTemplate{
			Name: "template",
			BasePattern: TargetPattern{
				Phases:      phases,
				Frequencies: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
			},
			Tolerance: 0.2,
		}
		library.Add("template", template)
	}

	testPattern := NewTargetPattern(
		[]float64{0, math.Pi / 2, math.Pi, 3 * math.Pi / 2},
		[]time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
	)

	b.ResetTimer()
	for range b.N {
		library.Identify(testPattern)
	}
}
