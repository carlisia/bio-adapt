package emerge

import (
	"math"
	"time"
)

// RhythmicPattern represents temporal patterns in agent behavior.
// Internal use only - for pattern detection and analysis.
type RhythmicPattern struct {
	phases      []float64
	frequencies []time.Duration
	amplitudes  []float64
	timestamps  []time.Time
}

// NewRhythmicPattern creates a new pattern tracker.
func NewRhythmicPattern(phases []float64, frequencies []time.Duration) *RhythmicPattern {
	amplitudes := make([]float64, len(phases))
	timestamps := make([]time.Time, len(phases))

	for i := range phases {
		amplitudes[i] = 1.0
		timestamps[i] = time.Now()
	}

	return &RhythmicPattern{
		phases:      phases,
		frequencies: frequencies,
		amplitudes:  amplitudes,
		timestamps:  timestamps,
	}
}

// Similarity calculates pattern similarity.
func (p *RhythmicPattern) Similarity(other *RhythmicPattern) float64 {
	if p == nil || other == nil || len(p.phases) == 0 || len(other.phases) == 0 {
		return 0
	}

	// Compare phase distributions
	phaseSim := phaseCorrelation(p.phases, other.phases)

	// Compare frequencies
	freqSim := 0.0
	minLen := min(len(p.frequencies), len(other.frequencies))

	for i := range minLen {
		diff := math.Abs(float64(p.frequencies[i] - other.frequencies[i]))
		maxFreq := math.Max(float64(p.frequencies[i]), float64(other.frequencies[i]))
		if maxFreq > 0 {
			freqSim += 1 - diff/maxFreq
		}
	}

	if minLen > 0 {
		freqSim /= float64(minLen)
	}

	return 0.7*phaseSim + 0.3*freqSim
}

// PatternGap identifies gaps in patterns.
type PatternGap struct {
	StartIdx int
	EndIdx   int
	Duration time.Duration
}

// DetectGaps finds gaps in the pattern.
func (p *RhythmicPattern) DetectGaps(threshold time.Duration) []PatternGap {
	var gaps []PatternGap

	for i := 1; i < len(p.timestamps); i++ {
		duration := p.timestamps[i].Sub(p.timestamps[i-1])
		if duration > threshold {
			gaps = append(gaps, PatternGap{
				StartIdx: i - 1,
				EndIdx:   i,
				Duration: duration,
			})
		}
	}

	return gaps
}

// phaseCorrelation is a helper for comparing phase arrays
func phaseCorrelation(phases1, phases2 []float64) float64 {
	if len(phases1) == 0 || len(phases2) == 0 {
		return 0
	}

	n := min(len(phases1), len(phases2))

	sumCos := 0.0
	sumSin := 0.0

	for i := range n {
		diff := phases1[i] - phases2[i]
		sumCos += math.Cos(diff)
		sumSin += math.Sin(diff)
	}

	return math.Sqrt(sumCos*sumCos+sumSin*sumSin) / float64(n)
}
