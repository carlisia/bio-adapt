package display

import (
	"fmt"
	"math"
	"strings"
)

// BinPhases distributes phases into bins for visualization.
func BinPhases(phases []float64, bins int) []int {
	counts := make([]int, bins)
	for _, phase := range phases {
		// Normalize phase to [0, 2π]
		normalizedPhase := phase
		for normalizedPhase < 0 {
			normalizedPhase += 2 * math.Pi
		}
		for normalizedPhase >= 2*math.Pi {
			normalizedPhase -= 2 * math.Pi
		}

		// Calculate bin index
		binIndex := int(normalizedPhase / (2 * math.Pi) * float64(bins))
		if binIndex >= bins {
			binIndex = bins - 1
		}
		counts[binIndex]++
	}
	return counts
}

// PrintClockBins visualizes phase distribution as clock positions.
func PrintClockBins(bins []int) {
	if len(bins) != 12 {
		return
	}

	clockEmojis := []string{"🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕗", "🕘", "🕙", "🕚", "🕛"}
	clockText := []string{"12:00", "1:00", "2:00", "3:00", "4:00", "5:00", "6:00", "7:00", "8:00", "9:00", "10:00", "11:00"}

	fmt.Println("Agent Phase Distribution (clock positions):")
	for i, count := range bins {
		if UseEmoji() {
			fmt.Printf("  %s ", clockEmojis[i])
		} else {
			fmt.Printf("  %5s ", clockText[i])
		}

		// Draw mini histogram
		stars := count
		if stars > 20 {
			stars = 20 // Cap for display
		}
		fmt.Printf("%-20s %d agents\n", strings.Repeat("*", stars), count)
	}
}

// PrintTimeline visualizes phase distribution as a timeline with batch windows.
func PrintTimeline(bins []int, msPerWindow int, batchThreshold int) {
	fmt.Printf("Request Timeline (%dms windows):\n", msPerWindow)

	maxCount := 0
	for _, count := range bins {
		if count > maxCount {
			maxCount = count
		}
	}

	// Scale factor for visualization
	scale := 20.0 / float64(maxCount)
	if maxCount == 0 {
		scale = 1
	}

	for i, count := range bins {
		// Time window label
		startMs := i * msPerWindow
		endMs := (i + 1) * msPerWindow
		fmt.Printf("%4d-%4dms: ", startMs, endMs)

		// Visual bar
		barLength := int(float64(count) * scale)
		if barLength > 20 {
			barLength = 20
		}

		bar := strings.Repeat("█", barLength)
		fmt.Printf("%-20s", bar)

		// Count and batch indicator
		if count >= batchThreshold {
			if UseEmoji() {
				fmt.Printf(" %2d reqs 📦 BATCH!\n", count)
			} else {
				fmt.Printf(" %2d reqs [BATCH]\n", count)
			}
		} else {
			fmt.Printf(" %2d reqs\n", count)
		}
	}
}
