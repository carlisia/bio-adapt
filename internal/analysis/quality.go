package analysis

import (
	"os"
)

// DescribeSyncQuality returns a human-readable description of synchronization quality
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
	if coherence < 0.2 {
		if useEmoji {
			return "(🌪️  Chaos - no coordination)"
		}
		return "(Chaos - no coordination)"
	} else if coherence < 0.4 {
		if useEmoji {
			return "(🌊 Groups forming - multiple rhythms)"
		}
		return "(Groups forming - multiple rhythms)"
	} else if coherence < 0.6 {
		if useEmoji {
			return "(⚡ Partial coordination - groups merging)"
		}
		return "(Partial coordination - groups merging)"
	} else if coherence < 0.8 {
		if useEmoji {
			return "(🎵 Good sync - single dominant rhythm)"
		}
		return "(Good sync - single dominant rhythm)"
	} else {
		if useEmoji {
			return "(✨ Excellent sync - unified rhythm)"
		}
		return "(Excellent sync - unified rhythm)"
	}
}

func describeBatchMode(coherence float64, useEmoji bool) string {
	if coherence < 0.2 {
		if useEmoji {
			return "(🌪️  Chaos - no batching)"
		}
		return "(Chaos - no batching)"
	} else if coherence < 0.4 {
		if useEmoji {
			return "(🌊 Weak batching emerging)"
		}
		return "(Weak batching emerging)"
	} else if coherence < 0.6 {
		if useEmoji {
			return "(⚡ Moderate batching)"
		}
		return "(Moderate batching)"
	} else if coherence < 0.8 {
		if useEmoji {
			return "(📦 Good batching)"
		}
		return "(Good batching)"
	} else {
		if useEmoji {
			return "(🚀 Excellent batching!)"
		}
		return "(Excellent batching!)"
	}
}

// InterpretCoherenceLevel provides a simple coherence level description
func InterpretCoherenceLevel(coherence float64) string {
	if coherence < 0.2 {
		return "Very Low"
	} else if coherence < 0.4 {
		return "Low"
	} else if coherence < 0.6 {
		return "Moderate"
	} else if coherence < 0.8 {
		return "Good"
	} else {
		return "Excellent"
	}
}
