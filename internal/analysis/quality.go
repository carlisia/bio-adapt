package analysis

import (
	"os"
)

// DescribeSyncQuality returns a human-readable description of synchronization quality.
func DescribeSyncQuality(coherence float64, mode string) string {
	useEmoji := os.Getenv("EMOJI") == "1"

	switch mode {
	case "sync":
		return describeSyncMode(coherence, useEmoji)
	case "batch":
		return describeBatchMode(coherence, useEmoji)
	default:
		return describeSyncMode(coherence, useEmoji)
	}
}

func describeSyncMode(coherence float64, useEmoji bool) string {
	switch {
	case coherence < 0.2:
		if useEmoji {
			return "(🌪️  Chaos - no coordination)"
		}
		return "(Chaos - no coordination)"
	case coherence < 0.4:
		if useEmoji {
			return "(🌊 Groups forming - multiple rhythms)"
		}
		return "(Groups forming - multiple rhythms)"
	case coherence < 0.6:
		if useEmoji {
			return "(⚡ Partial coordination - groups merging)"
		}
		return "(Partial coordination - groups merging)"
	case coherence < 0.8:
		if useEmoji {
			return "(🎵 Good sync - single dominant rhythm)"
		}
		return "(Good sync - single dominant rhythm)"
	default:
		if useEmoji {
			return "(✨ Excellent sync - unified rhythm)"
		}
		return "(Excellent sync - unified rhythm)"
	}
}

func describeBatchMode(coherence float64, useEmoji bool) string {
	switch {
	case coherence < 0.2:
		if useEmoji {
			return "(🌪️  Chaos - no batching)"
		}
		return "(Chaos - no batching)"
	case coherence < 0.4:
		if useEmoji {
			return "(🌊 Weak batching emerging)"
		}
		return "(Weak batching emerging)"
	case coherence < 0.6:
		if useEmoji {
			return "(⚡ Moderate batching)"
		}
		return "(Moderate batching)"
	case coherence < 0.8:
		if useEmoji {
			return "(📦 Good batching)"
		}
		return "(Good batching)"
	default:
		if useEmoji {
			return "(🚀 Excellent batching!)"
		}
		return "(Excellent batching!)"
	}
}

// InterpretCoherenceLevel provides a simple coherence level description.
func InterpretCoherenceLevel(coherence float64) string {
	switch {
	case coherence < 0.2:
		return "Very Low"
	case coherence < 0.4:
		return "Low"
	case coherence < 0.6:
		return "Moderate"
	case coherence < 0.8:
		return "Good"
	default:
		return "Excellent"
	}
}
