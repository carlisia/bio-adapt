package goal

import "github.com/carlisia/bio-adapt/emerge/core"

// Manager blends local and global objectives hierarchically.
// This creates multi-scale goal structure where individual goals
// contribute to but don't override collective goals.
type Manager interface {
	// Blend combines local and global goals based on context.
	// Weight parameter determines local (0) vs global (1) influence.
	Blend(local, global core.State, weight float64) core.State
}
