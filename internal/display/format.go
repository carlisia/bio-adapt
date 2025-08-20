package display

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// Banner prints a prominent banner title.
func Banner(title string) {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Printf("║ %-58s ║\n", title)
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

// Section prints a section header.
func Section(title string) {
	if UseEmoji() {
		fmt.Printf("📊 %s\n", strings.ToUpper(title))
	} else {
		fmt.Printf("=== %s ===\n", strings.ToUpper(title))
	}
}

// Bullet prints bulleted lines.
func Bullet(lines ...string) {
	for _, line := range lines {
		fmt.Printf("• %s\n", line)
	}
}

// UseEmoji returns true unless emoji display is explicitly disabled via EMOJI=0 environment variable.
func UseEmoji() bool {
	return os.Getenv("EMOJI") != "0"
}

// NewTable creates a new table writer with consistent styling.
func NewTable() table.Writer {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	// Set consistent style
	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = false
	t.Style().Options.DrawBorder = false
	t.Style().Format.Header = text.FormatDefault

	// Limit width to 80 columns
	t.SetAllowedRowLength(80)

	return t
}
