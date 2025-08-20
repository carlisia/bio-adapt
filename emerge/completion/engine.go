package completion

import (
	"math"
	"sync"
	"time"

	"github.com/carlisia/bio-adapt/emerge/core"
)

// Engine fills in missing parts of synchronization patterns
type Engine struct {
	templates      map[string]*core.RhythmicPattern
	templatesMu    sync.RWMutex
	matchThreshold float64
}

// NewEngine creates a pattern completion engine
func NewEngine() *Engine {
	return &Engine{
		templates:      make(map[string]*core.RhythmicPattern),
		matchThreshold: 0.7,
	}
}

// CompletePattern fills gaps in the current pattern
func (e *Engine) CompletePattern(current *core.RhythmicPattern, gaps []core.PatternGap) *CompletedPattern {
	// Find best matching template
	template := e.findBestTemplate(current)

	completed := &CompletedPattern{
		Base:   current,
		Filled: make(map[string]interface{}),
	}

	// If no template, interpolate toward target
	if template == nil {
		return e.interpolatePattern(current, gaps)
	}

	// Use template to fill gaps
	for _, gap := range gaps {
		switch gap.Type {
		case "phase":
			// Gradually adjust phase toward template
			completed.Filled["phase"] = e.blendPhase(current.Phase, template.Phase, gap.Severity)

		case "frequency":
			// Blend frequency toward template
			completed.Filled["frequency"] = e.blendFrequency(current.Frequency, template.Frequency, gap.Severity)

		case "coherence":
			// Use template's coherence strategy
			completed.Filled["coherence_boost"] = template.Coherence * gap.Severity

		case "waveform":
			// Morph waveform toward template
			if len(template.Waveform) > 0 {
				completed.Filled["waveform"] = e.morphWaveform(current.Waveform, template.Waveform, gap.Severity)
			}
		}
	}

	return completed
}

// CompletedPattern represents a pattern with gaps filled
type CompletedPattern struct {
	Base     *core.RhythmicPattern
	Template *core.RhythmicPattern
	Filled   map[string]interface{}
}

// GetPhaseAdjustment returns the phase adjustment to apply
func (cp *CompletedPattern) GetPhaseAdjustment() float64 {
	if phase, ok := cp.Filled["phase"].(float64); ok {
		return core.PhaseDifference(phase, cp.Base.Phase)
	}
	return 0
}

// GetFrequencyAdjustment returns the frequency adjustment
func (cp *CompletedPattern) GetFrequencyAdjustment() time.Duration {
	if freq, ok := cp.Filled["frequency"].(time.Duration); ok {
		return freq - cp.Base.Frequency
	}
	return 0
}

// findBestTemplate finds the closest matching pattern template
func (e *Engine) findBestTemplate(current *core.RhythmicPattern) *core.RhythmicPattern {
	e.templatesMu.RLock()
	defer e.templatesMu.RUnlock()

	var bestTemplate *core.RhythmicPattern
	bestDistance := math.MaxFloat64

	for _, template := range e.templates {
		distance := core.PatternDistance(current, template)
		if distance < bestDistance && distance < e.matchThreshold {
			bestDistance = distance
			bestTemplate = template
		}
	}

	return bestTemplate
}

// interpolatePattern creates interpolated values toward target
func (e *Engine) interpolatePattern(current *core.RhythmicPattern, gaps []core.PatternGap) *CompletedPattern {
	completed := &CompletedPattern{
		Base:   current,
		Filled: make(map[string]interface{}),
	}

	for _, gap := range gaps {
		// Linear interpolation weighted by severity
		weight := math.Min(gap.Severity, 0.5) // Don't jump more than 50% in one step

		switch gap.Type {
		case "phase":
			diff := core.PhaseDifference(gap.Target, gap.Current)
			completed.Filled["phase"] = core.WrapPhase(gap.Current + diff*weight)

		case "frequency":
			diff := gap.Target - gap.Current
			completed.Filled["frequency"] = time.Duration(gap.Current + diff*weight)

		case "coherence":
			diff := gap.Target - gap.Current
			completed.Filled["coherence_boost"] = diff * weight
		}
	}

	return completed
}

// blendPhase blends two phases with given weight
func (e *Engine) blendPhase(current, template float64, weight float64) float64 {
	diff := core.PhaseDifference(template, current)
	return core.WrapPhase(current + diff*weight)
}

// blendFrequency blends two frequencies
func (e *Engine) blendFrequency(current, template time.Duration, weight float64) time.Duration {
	diff := float64(template - current)
	return current + time.Duration(diff*weight)
}

// morphWaveform morphs one waveform toward another
func (e *Engine) morphWaveform(current, template []float64, weight float64) []float64 {
	if len(current) == 0 {
		return template
	}
	if len(template) == 0 {
		return current
	}

	// Resample if different lengths
	maxLen := len(current)
	if len(template) > maxLen {
		maxLen = len(template)
	}

	morphed := make([]float64, maxLen)
	for i := range morphed {
		// Sample from both waveforms
		currentVal := current[i*len(current)/maxLen]
		templateVal := template[i*len(template)/maxLen]

		// Weighted blend
		morphed[i] = currentVal*(1-weight) + templateVal*weight
	}

	return morphed
}

// AddTemplate adds a learned pattern template
func (e *Engine) AddTemplate(name string, pattern *core.RhythmicPattern) {
	e.templatesMu.Lock()
	defer e.templatesMu.Unlock()
	e.templates[name] = pattern
}

// LoadDefaultTemplates loads standard synchronization patterns
func (e *Engine) LoadDefaultTemplates() {
	// Perfect sync template
	e.AddTemplate("perfect_sync", &core.RhythmicPattern{
		Phase:     0,
		Frequency: 250 * time.Millisecond,
		Amplitude: 1.0,
		Coherence: 1.0,
		Waveform:  core.GenerateSineWave(100),
		Stability: 1.0,
	})

	// Partial sync template
	e.AddTemplate("partial_sync", &core.RhythmicPattern{
		Phase:     0,
		Frequency: 250 * time.Millisecond,
		Amplitude: 0.8,
		Coherence: 0.65,
		Waveform:  core.GenerateSineWave(100),
		Stability: 0.7,
	})

	// Recovery template (for disrupted systems)
	e.AddTemplate("recovery", &core.RhythmicPattern{
		Phase:     0,
		Frequency: 200 * time.Millisecond,
		Amplitude: 0.6,
		Coherence: 0.5,
		Waveform:  core.GenerateSineWave(100),
		Stability: 0.5,
	})
}
