package display

import (
	"fmt"
	"strings"
)

// DrawProgressBar draws a progress bar with current/target values
func DrawProgressBar(current, target float64, width int) {
	progress := min(current/target, 1.0)
	filled := int(progress * float64(width))

	// Progress indicator
	indicator := getProgressIndicator(progress)

	fmt.Print(indicator)
	fmt.Print(" [")
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	fmt.Print(bar)
	fmt.Print("]")
}

// DrawMiniBar draws a compact progress bar
func DrawMiniBar(progress float64, width int) {
	filled := int(progress * float64(width))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	fmt.Printf("[%s]", bar)
}

// getProgressIndicator returns appropriate progress indicator
func getProgressIndicator(progress float64) string {
	if UseEmoji() {
		if progress < 0.3 {
			return "🔴"
		} else if progress < 0.7 {
			return "🟡"
		} else {
			return "🟢"
		}
	} else {
		if progress < 0.3 {
			return "[RED]"
		} else if progress < 0.7 {
			return "[YLW]"
		} else {
			return "[GRN]"
		}
	}
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
