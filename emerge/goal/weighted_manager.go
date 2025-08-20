package goal

import (
	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/util"
)

// WeightedManager provides weighted blending of local and global goals.
type WeightedManager struct{}

// Blend combines local and global goals based on weight.
func (w *WeightedManager) Blend(local, global core.State, weight float64) core.State {
	// Clamp weight to [0, 1]
	if weight < 0 {
		weight = 0
	} else if weight > 1 {
		weight = 1
	}

	// Weight determines influence: 0 = fully local, 1 = fully global
	localInfluence := 1.0 - weight
	globalInfluence := weight

	// Blend phases with proper circular interpolation
	// Use PhaseDifference to get the shortest angular path from local to global
	phaseDiff := util.PhaseDifference(global.Phase, local.Phase)
	blendedPhase := util.WrapPhase(local.Phase + phaseDiff*globalInfluence)

	return core.State{
		Phase:     blendedPhase,
		Frequency: local.Frequency,
		Coherence: local.Coherence*localInfluence + global.Coherence*globalInfluence,
	}
}
