package core

import (
	"math"
	"time"
)

// RhythmicPattern represents a complete synchronization pattern (bioelectric morphology).
type RhythmicPattern struct {
	Phase     float64
	Frequency time.Duration
	Amplitude float64
	Coherence float64
	Waveform  []float64 // Discrete waveform representation
	Stability float64   // Attractor strength
}

// PatternGap identifies missing pieces in synchronization.
type PatternGap struct {
	Type     string // "phase", "frequency", "coherence", "waveform"
	Current  float64
	Target   float64
	Severity float64 // 0-1, how critical this gap is
}

// PatternDistance calculates how far current pattern is from target.
func PatternDistance(current, target *RhythmicPattern) float64 {
	if current == nil || target == nil {
		return math.MaxFloat64
	}

	// Phase difference (wrapped)
	phaseDiff := math.Abs(PhaseDifference(target.Phase, current.Phase))
	phaseDistance := phaseDiff / math.Pi // Normalize to 0-1

	// Frequency difference
	freqDistance := 0.0
	if target.Frequency > 0 {
		freqDiff := math.Abs(float64(current.Frequency-target.Frequency)) / float64(target.Frequency)
		freqDistance = math.Min(freqDiff, 1.0)
	}

	// Coherence difference
	coherenceDistance := math.Abs(target.Coherence - current.Coherence)

	// Weighted distance
	return 0.4*phaseDistance + 0.3*freqDistance + 0.3*coherenceDistance
}

// IdentifyGaps finds what's missing from current pattern.
func IdentifyGaps(current, target *RhythmicPattern) []PatternGap {
	var gaps []PatternGap

	// Phase gap
	phaseDiff := math.Abs(PhaseDifference(target.Phase, current.Phase))
	if phaseDiff > 0.1 {
		gaps = append(gaps, PatternGap{
			Type:     "phase",
			Current:  current.Phase,
			Target:   target.Phase,
			Severity: phaseDiff / math.Pi,
		})
	}

	// Frequency gap
	if target.Frequency > 0 && current.Frequency > 0 {
		freqRatio := float64(current.Frequency) / float64(target.Frequency)
		if math.Abs(freqRatio-1.0) > 0.05 {
			gaps = append(gaps, PatternGap{
				Type:     "frequency",
				Current:  float64(current.Frequency),
				Target:   float64(target.Frequency),
				Severity: math.Abs(freqRatio - 1.0),
			})
		}
	}

	// Coherence gap
	if current.Coherence < target.Coherence-0.05 {
		gaps = append(gaps, PatternGap{
			Type:     "coherence",
			Current:  current.Coherence,
			Target:   target.Coherence,
			Severity: target.Coherence - current.Coherence,
		})
	}

	return gaps
}

// GenerateSineWave creates a sine waveform.
func GenerateSineWave(points int) []float64 {
	wave := make([]float64, points)
	for i := range points {
		phase := float64(i) * 2 * math.Pi / float64(points)
		wave[i] = math.Sin(phase)
	}
	return wave
}
