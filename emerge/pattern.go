package emerge

import (
	"math"
	"time"
)

// RhythmicPattern represents a temporal pattern of phases and frequencies.
// This captures repeating biological patterns like heartbeats, circadian rhythms,
// or neural oscillations.
type RhythmicPattern struct {
	Phases      []float64       // Phase values at each time point
	Frequencies []time.Duration // Frequency at each time point
	Amplitude   float64         // Pattern strength/amplitude
	Period      time.Duration   // Overall pattern period
	Confidence  float64         // Confidence in pattern detection [0, 1]
}

// NewRhythmicPattern creates a new rhythmic pattern.
func NewRhythmicPattern(phases []float64, frequencies []time.Duration) *RhythmicPattern {
	amplitude := calculateAmplitude(phases)
	period := calculatePeriod(frequencies)

	return &RhythmicPattern{
		Phases:      phases,
		Frequencies: frequencies,
		Amplitude:   amplitude,
		Period:      period,
		Confidence:  1.0, // Default full confidence
	}
}

// PatternGap represents a missing segment in a pattern.
// These gaps need to be filled through pattern completion.
type PatternGap struct {
	StartIdx  int           // Start index in pattern
	EndIdx    int           // End index in pattern
	Duration  time.Duration // Time duration of gap
	Predicted []float64     // Predicted values for the gap
}

// Detect finds gaps in the pattern that need completion.
func (p *RhythmicPattern) Detect() []PatternGap {
	gaps := make([]PatternGap, 0)

	// Simple gap detection: look for sudden phase jumps
	for i := 1; i < len(p.Phases); i++ {
		diff := math.Abs(PhaseDifference(p.Phases[i], p.Phases[i-1]))

		// If phase jump is too large, likely a gap
		if diff > math.Pi/2 {
			gap := PatternGap{
				StartIdx: i - 1,
				EndIdx:   i,
				Duration: p.Frequencies[i-1],
			}
			gaps = append(gaps, gap)
		}
	}

	return gaps
}

// Complete fills in missing parts of a pattern.
func (p *RhythmicPattern) Complete(gap PatternGap) []float64 {
	if gap.StartIdx < 0 || gap.EndIdx >= len(p.Phases) || gap.StartIdx >= gap.EndIdx {
		return nil
	}

	// Simple linear interpolation for now
	startPhase := p.Phases[gap.StartIdx]
	endPhase := p.Phases[gap.EndIdx]

	steps := gap.EndIdx - gap.StartIdx - 1
	if steps <= 0 {
		return []float64{}
	}

	completed := make([]float64, steps)
	diff := PhaseDifference(endPhase, startPhase)

	for i := range steps {
		fraction := float64(i+1) / float64(steps+1)
		completed[i] = WrapPhase(startPhase + diff*fraction)
	}

	return completed
}

// Similarity calculates how similar two patterns are.
// Returns a value between 0 (completely different) and 1 (identical).
func (p *RhythmicPattern) Similarity(other *RhythmicPattern) float64 {
	if other == nil || len(p.Phases) == 0 || len(other.Phases) == 0 {
		return 0
	}

	// Compare amplitudes
	ampDiff := math.Abs(p.Amplitude - other.Amplitude)
	ampSim := 1.0 - math.Min(ampDiff/math.Max(p.Amplitude, other.Amplitude), 1.0)

	// Compare periods
	periodDiff := math.Abs(float64(p.Period - other.Period))
	maxPeriod := max(p.Period, other.Period)
	periodSim := 1.0 - math.Min(periodDiff/float64(maxPeriod), 1.0)

	// Check for massively different scales before normalization
	scaleFactor := 1.0
	if len(p.Phases) == len(other.Phases) && len(p.Phases) > 0 {
		var maxP, maxO float64
		for i := range p.Phases {
			maxP = math.Max(maxP, math.Abs(p.Phases[i]))
			maxO = math.Max(maxO, math.Abs(other.Phases[i]))
		}
		if maxP > 0 && maxO > 0 {
			ratio := math.Max(maxP/maxO, maxO/maxP)
			if ratio > 100 { // If scales differ by more than 100x
				scaleFactor = 1.0 / math.Min(ratio/100, 10) // Reduce similarity significantly
			}
		}
	}

	// Compare phase patterns (using correlation)
	phaseSim := phaseCorrelation(p.Phases, other.Phases) * scaleFactor

	// Weighted average
	return (ampSim*0.3 + periodSim*0.3 + phaseSim*0.4)
}

// PatternTemplate represents a known pattern archetype.
// These are like "morphogenetic templates" that guide development.
type PatternTemplate struct {
	Name        string
	BasePattern RhythmicPattern
	Variations  []RhythmicPattern // Acceptable variations
	Tolerance   float64           // How much deviation is acceptable
}

// Matches checks if a pattern matches this template.
func (t *PatternTemplate) Matches(pattern *RhythmicPattern) bool {
	// Check against base pattern
	similarity := t.BasePattern.Similarity(pattern)
	if similarity >= (1.0 - t.Tolerance) {
		return true
	}

	// Check against variations
	for _, variation := range t.Variations {
		if variation.Similarity(pattern) >= (1.0 - t.Tolerance) {
			return true
		}
	}

	return false
}

// PatternLibrary stores known pattern templates.
type PatternLibrary struct {
	templates map[string]*PatternTemplate
}

// NewPatternLibrary creates a new pattern library.
func NewPatternLibrary() *PatternLibrary {
	return &PatternLibrary{
		templates: make(map[string]*PatternTemplate),
	}
}

// Add registers a new pattern template.
func (l *PatternLibrary) Add(name string, template *PatternTemplate) {
	l.templates[name] = template
}

// Identify attempts to identify which template a pattern matches.
func (l *PatternLibrary) Identify(pattern *RhythmicPattern) (string, float64) {
	var bestMatch string
	var bestSimilarity float64

	for name, template := range l.templates {
		similarity := template.BasePattern.Similarity(pattern)
		if similarity > bestSimilarity {
			bestSimilarity = similarity
			bestMatch = name
		}
	}

	return bestMatch, bestSimilarity
}

// Helper functions

func calculateAmplitude(phases []float64) float64 {
	if len(phases) < 2 {
		return 0
	}

	// Normalize phases to [0, 2π] for circular data
	normalizedPhases := make([]float64, len(phases))
	for i, p := range phases {
		normalizedPhases[i] = WrapPhase(p)
	}

	// Calculate variance as measure of amplitude
	mean := 0.0
	for _, p := range normalizedPhases {
		mean += p
	}
	mean /= float64(len(normalizedPhases))

	variance := 0.0
	for _, p := range normalizedPhases {
		diff := p - mean
		variance += diff * diff
	}
	variance /= float64(len(normalizedPhases))

	return math.Sqrt(variance)
}

func calculatePeriod(frequencies []time.Duration) time.Duration {
	if len(frequencies) == 0 {
		return 0
	}

	// Average frequency as period estimate
	var total time.Duration
	for _, f := range frequencies {
		total += f
	}

	return total / time.Duration(len(frequencies))
}

func phaseCorrelation(phases1, phases2 []float64) float64 {
	// Return 0 if arrays have different lengths
	if len(phases1) != len(phases2) {
		return 0
	}

	if len(phases1) == 0 {
		return 0
	}

	// Calculate correlation coefficient
	var sum float64
	for i := range len(phases1) {
		diff := math.Abs(PhaseDifference(phases1[i], phases2[i]))
		sum += 1.0 - (diff / math.Pi) // Normalize to [0, 1]
	}

	return sum / float64(len(phases1))
}
