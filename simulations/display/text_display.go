// text_display.go provides a simple text output for non-TTY environments.
// This is useful for testing and running in CI/CD pipelines.

package display

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// TextDisplay implements Display interface with simple text output
type TextDisplay struct {
	lastUpdate     time.Time
	updateInterval time.Duration
	simulationName string
}

// NewTextDisplay creates a text-only display
func NewTextDisplay() Display {
	return &TextDisplay{
		updateInterval: 1 * time.Second,      // Only update once per second in text mode
		simulationName: "MINIMIZE API CALLS", // Default name
	}
}

// SetSimulationName sets the simulation name for display
func (d *TextDisplay) SetSimulationName(name string) {
	d.simulationName = name
}

// Initialize does nothing for text display
func (*TextDisplay) Initialize() error {
	return nil
}

// Close does nothing for text display
func (*TextDisplay) Close() error {
	return nil
}

// ShowWelcome prints welcome message
func (d *TextDisplay) ShowWelcome() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("   EMERGE: %s DEMO\n", d.simulationName)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()
	fmt.Println("Async workloads learning to batch their API calls...")
	fmt.Println("No central coordinator - emergent synchronization!")
	fmt.Println()
}

// Update prints current status (rate-limited for text mode)
func (d *TextDisplay) Update(snapshot SimulationSnapshot) {
	// Rate limit updates to avoid flooding the console
	if time.Since(d.lastUpdate) < d.updateInterval {
		return
	}
	d.lastUpdate = time.Now()

	// Use a simple newline-based output for better compatibility
	fmt.Printf("[%s] Coherence: %.1f%% | Savings: $%.2f (%.1f%%) | Batches: %d\n",
		formatTime(snapshot.ElapsedTime),
		snapshot.Coherence*100,
		snapshot.Savings,
		snapshot.SavingsPercent,
		snapshot.BatchesProcessed,
	)
}

// ShowSummary prints final results
func (*TextDisplay) ShowSummary(stats Statistics) {
	fmt.Println()
	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("   SIMULATION COMPLETE")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Total API Calls (without sync): %d\n", stats.TotalAPICalls)
	fmt.Printf("Total Batches (with sync):       %d\n", stats.TotalBatches)
	fmt.Printf("Average Batch Size:              %.1f\n", stats.AverageBatchSize)
	fmt.Printf("Total Savings:                   $%.2f (%.1f%%)\n",
		stats.TotalSavings, stats.SavingsPercent)
	fmt.Printf("Peak Coherence:                  %.1f%%\n", stats.PeakCoherence*100)
	fmt.Println()
	fmt.Println("✓ Agents successfully learned to batch API calls!")
	fmt.Println()
}

// Run just waits for context cancellation
func (*TextDisplay) Run(ctx context.Context, _ *KeyboardController) error {
	<-ctx.Done()
	return nil
}

func formatTime(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%02.0fs", d.Seconds())
	}
	return fmt.Sprintf("%02.0fm", d.Minutes())
}
