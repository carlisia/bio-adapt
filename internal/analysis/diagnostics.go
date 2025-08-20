package analysis

import (
	"fmt"
	"math"
	"os"
)

// RunStats holds statistics about a simulation run.
type RunStats struct {
	Initial    float64
	Final      float64
	Iter       int
	MaxIter    int
	StuckCount int
}

// Diagnose analyzes run statistics and provides diagnostics.
func Diagnose(s RunStats, target float64, mode string) (summary string, suggestions []string) {
	useEmoji := os.Getenv("EMOJI") == "1"

	// Check if target was reached
	if s.Final >= target {
		if useEmoji {
			summary = fmt.Sprintf("✅ Success! Reached %.1f%% coherence", s.Final*100)
		} else {
			summary = fmt.Sprintf("[OK] Success! Reached %.1f%% coherence", s.Final*100)
		}
		return summary, nil
	}

	// Analyze failure mode
	gap := math.Abs(target - s.Final)
	improvement := s.Final - s.Initial

	// Determine primary issue
	switch {
	case s.StuckCount > s.MaxIter/3:
		// Stuck in local minimum
		if useEmoji {
			summary = fmt.Sprintf("⚠️  Stuck at %.1f%% (target: %.1f%%)", s.Final*100, target*100)
		} else {
			summary = fmt.Sprintf("[!!] Stuck at %.1f%% (target: %.1f%%)", s.Final*100, target*100)
		}

		suggestions = []string{
			"Increase swarm size (more agents = more pathways)",
			"Reduce target coherence (try 0.6-0.7)",
			"Adjust frequency to 200-300ms",
		}

		if mode == "batch" {
			suggestions = append(suggestions,
				"Check network topology (ensure sufficient connections)",
				"Consider partial sync (0.65-0.75) for load spreading")
		}
	case improvement < 0.1:
		// No significant improvement
		if useEmoji {
			summary = fmt.Sprintf("⚠️  Minimal progress: %.1f%% → %.1f%%", s.Initial*100, s.Final*100)
		} else {
			summary = fmt.Sprintf("[!!] Minimal progress: %.1f%% → %.1f%%", s.Initial*100, s.Final*100)
		}

		suggestions = []string{
			"Check agent connectivity (isolated agents?)",
			"Increase influence strength",
			"Reduce stubbornness parameter",
		}
	case gap < 0.1:
		// Close but not quite
		if useEmoji {
			summary = fmt.Sprintf("🔶 Close! Reached %.1f%% (target: %.1f%%)", s.Final*100, target*100)
		} else {
			summary = fmt.Sprintf("[~] Close! Reached %.1f%% (target: %.1f%%)", s.Final*100, target*100)
		}

		suggestions = []string{
			"Increase iterations (more time to converge)",
			"Fine-tune coupling strength",
		}
	default:
		// General convergence issue
		if useEmoji {
			summary = fmt.Sprintf("❌ Convergence failed: %.1f%% (target: %.1f%%)", s.Final*100, target*100)
		} else {
			summary = fmt.Sprintf("[X] Convergence failed: %.1f%% (target: %.1f%%)", s.Final*100, target*100)
		}

		suggestions = []string{
			"Start with smaller swarm (5-10 agents)",
			"Lower target to 0.5-0.6",
			"Increase frequency to 300-500ms",
			"Check for competing attractors",
		}
	}

	return summary, suggestions
}

// AnalyzeConvergenceRate determines if convergence is healthy.
func AnalyzeConvergenceRate(iterations int, maxIterations int) string {
	ratio := float64(iterations) / float64(maxIterations)

	switch {
	case ratio < 0.3:
		return "Fast convergence"
	case ratio < 0.6:
		return "Normal convergence"
	case ratio < 0.9:
		return "Slow convergence"
	default:
		return "Very slow convergence"
	}
}
