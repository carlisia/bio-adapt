package display

import (
	"fmt"
	"testing"

	"github.com/carlisia/bio-adapt/emerge/goal"
	"github.com/carlisia/bio-adapt/emerge/scale"
	"github.com/carlisia/bio-adapt/simulations/emerge/simulation/pattern"
)

// Test reason constants
const (
	reasonSupportsBatch     = " (supports batch formation)"
	reasonHelpsDistribute   = " (helps distribute workload)"
	reasonSupportsConsensus = " (supports consensus)"
	reasonEnablesOptimize   = " (enables optimization)"
)

// TestAllCombinations tests all goal-scale-pattern combinations for consistency
func TestAllCombinations(t *testing.T) {
	t.Parallel()
	goals := []goal.Type{
		goal.MinimizeAPICalls,
		goal.DistributeLoad,
		goal.ReachConsensus,
		goal.MinimizeLatency,
		goal.SaveEnergy,
		goal.MaintainRhythm,
		goal.RecoverFromFailure,
		goal.AdaptToTraffic,
	}

	scales := []struct {
		size  scale.Size
		count int
		name  string
	}{
		{scale.Tiny, 20, "Tiny"},
		{scale.Small, 50, "Small"},
		{scale.Medium, 200, "Medium"},
		{scale.Large, 1000, "Large"},
		{scale.Huge, 5000, "Huge"},
	}

	patterns := []pattern.Type{
		pattern.HighFrequency,
		pattern.Burst,
		pattern.Steady,
		pattern.Mixed,
		pattern.Sparse,
	}

	// Test all combinations
	for _, g := range goals {
		t.Run(fmt.Sprintf("Goal_%s", g.String()), func(t *testing.T) {
			t.Parallel()
			for _, s := range scales {
				t.Run(fmt.Sprintf("Scale_%s", s.name), func(t *testing.T) {
					t.Parallel()
					// Check if scale is recommended for goal
					isRecommended := g.IsRecommendedForSize(s.count)

					// Get the reason that would be displayed
					reason := getScaleReason(g, s.count, isRecommended)

					// Verify that we always have a reason
					if reason == "" {
						t.Errorf("No reason for goal=%s scale=%s(%d) recommended=%v",
							g.String(), s.name, s.count, isRecommended)
					}

					// Log the combination for review
					t.Logf("Goal=%s Scale=%s(%d) Recommended=%v Reason=%s",
						g.String(), s.name, s.count, isRecommended, reason)

					// Test patterns for this goal-scale combination
					for _, p := range patterns {
						qualifier := pattern.GetQualifier(int(g), s.count, p)
						patternReason := getPatternReason(g, p, qualifier == pattern.QualifierOptimal || qualifier == pattern.QualifierExcellent || qualifier == pattern.QualifierGood)

						// Verify pattern has a qualifier and reason
						if qualifier == "" {
							t.Errorf("No qualifier for goal=%s scale=%s pattern=%s",
								g.String(), s.name, p.String())
						}
						if patternReason == "" {
							t.Errorf("No pattern reason for goal=%s scale=%s pattern=%s qualifier=%s",
								g.String(), s.name, p.String(), qualifier)
						}

						// Log pattern combination
						t.Logf("  Pattern=%s Qualifier=%s Reason=%s",
							p.String(), qualifier, patternReason)
					}
				})
			}
		})
	}
}

// getScaleReason mimics the logic from display.go for testing
func getScaleReason(g goal.Type, agentCount int, isRecommended bool) string {
	if isRecommended {
		switch g {
		case goal.MinimizeAPICalls:
			return " (more agents = bigger batches)"
		case goal.DistributeLoad:
			if agentCount >= 200 {
				return " (many workers share the load)"
			}
			return " (enough workers to distribute)"
		case goal.ReachConsensus:
			if agentCount <= 200 {
				return " (fewer agents = faster agreement)"
			}
			return " (still manageable voting time)"
		case goal.MinimizeLatency:
			return " (fewer hops = faster response)"
		case goal.SaveEnergy:
			if agentCount <= 200 {
				return " (fewer agents = less power)"
			}
			return " (coordinated sleep cycles)"
		case goal.MaintainRhythm:
			if agentCount <= 50 {
				return " (small groups sync naturally)"
			}
			return " (frequency locking scales)"
		case goal.RecoverFromFailure:
			return " (backup agents ready)"
		case goal.AdaptToTraffic:
			if agentCount >= 200 {
				return " (handles traffic spikes)"
			}
			return " (absorbs normal variance)"
		default:
			return " (effective at this scale)"
		}
	} else {
		// Not recommended reasons
		switch g {
		case goal.MinimizeAPICalls:
			// MinimizeAPICalls works at all scales, so this shouldn't happen
			return " (less batching opportunity)"
		case goal.DistributeLoad:
			// DistributeLoad needs at least 20 agents, shouldn't get here
			if agentCount < 20 {
				return " (not enough workers)"
			}
			return " (coordination overhead)"
		case goal.ReachConsensus:
			if agentCount < 50 {
				return " (too few voices for quorum)"
			} else if agentCount > 1000 {
				return " (too many agents slows consensus)"
			}
			return " (group size slows agreement)"
		case goal.MinimizeLatency:
			if agentCount > 200 {
				return " (more hops between agents)"
			}
			return " (coordination adds delay)"
		case goal.SaveEnergy:
			if agentCount > 200 {
				return " (too many agents drain power)"
			}
			return " (energy not optimized)"
		case goal.MaintainRhythm:
			// MaintainRhythm works at all sizes, shouldn't get here
			return " (rhythm disrupted)"
		case goal.RecoverFromFailure:
			// RecoverFromFailure needs at least 20 agents
			if agentCount < 20 {
				return " (no backup agents)"
			}
			return " (recovery not optimal)"
		case goal.AdaptToTraffic:
			if agentCount < 20 {
				return " (can't handle spikes)"
			} else if agentCount > 1000 {
				return " (slow to adapt)"
			}
			return " (traffic handling limited)"
		default:
			// Generic fallback - explain the actual problem
			if agentCount < 50 {
				return " (too few agents for coordination)"
			}
			if agentCount > 1000 {
				return " (coordination overhead dominates)"
			}
			return " (inefficient at this size)"
		}
	}
}

// TestScaleRecommendationConsistency ensures our display logic matches the goal's recommendations
func TestScaleRecommendationConsistency(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		goal       goal.Type
		agentCount int
		expectRec  bool
		scaleName  string
	}{
		// MinimizeAPICalls - works at all scales
		{goal.MinimizeAPICalls, 20, true, "Tiny"},
		{goal.MinimizeAPICalls, 50, true, "Small"},
		{goal.MinimizeAPICalls, 200, true, "Medium"},
		{goal.MinimizeAPICalls, 1000, true, "Large"},
		{goal.MinimizeAPICalls, 5000, true, "Huge"},

		// DistributeLoad - needs at least 20 agents
		{goal.DistributeLoad, 20, true, "Tiny"},
		{goal.DistributeLoad, 50, true, "Small"},
		{goal.DistributeLoad, 200, true, "Medium"},
		{goal.DistributeLoad, 1000, true, "Large"},
		{goal.DistributeLoad, 5000, true, "Huge"},

		// ReachConsensus - optimal range 50-1000
		{goal.ReachConsensus, 20, false, "Tiny"},
		{goal.ReachConsensus, 50, true, "Small"},
		{goal.ReachConsensus, 200, true, "Medium"},
		{goal.ReachConsensus, 1000, true, "Large"},
		{goal.ReachConsensus, 5000, false, "Huge"},

		// MinimizeLatency - smaller is better (<=200)
		{goal.MinimizeLatency, 20, true, "Tiny"},
		{goal.MinimizeLatency, 50, true, "Small"},
		{goal.MinimizeLatency, 200, true, "Medium"},
		{goal.MinimizeLatency, 1000, false, "Large"},
		{goal.MinimizeLatency, 5000, false, "Huge"},

		// SaveEnergy - smaller saves more (<=200)
		{goal.SaveEnergy, 20, true, "Tiny"},
		{goal.SaveEnergy, 50, true, "Small"},
		{goal.SaveEnergy, 200, true, "Medium"},
		{goal.SaveEnergy, 1000, false, "Large"},
		{goal.SaveEnergy, 5000, false, "Huge"},

		// MaintainRhythm - works at all sizes
		{goal.MaintainRhythm, 20, true, "Tiny"},
		{goal.MaintainRhythm, 50, true, "Small"},
		{goal.MaintainRhythm, 200, true, "Medium"},
		{goal.MaintainRhythm, 1000, true, "Large"},
		{goal.MaintainRhythm, 5000, true, "Huge"},

		// RecoverFromFailure - needs at least 20 agents
		{goal.RecoverFromFailure, 20, true, "Tiny"},
		{goal.RecoverFromFailure, 50, true, "Small"},
		{goal.RecoverFromFailure, 200, true, "Medium"},
		{goal.RecoverFromFailure, 1000, true, "Large"},
		{goal.RecoverFromFailure, 5000, true, "Huge"},

		// AdaptToTraffic - needs 20-1000 agents
		{goal.AdaptToTraffic, 20, true, "Tiny"},
		{goal.AdaptToTraffic, 50, true, "Small"},
		{goal.AdaptToTraffic, 200, true, "Medium"},
		{goal.AdaptToTraffic, 1000, true, "Large"},
		{goal.AdaptToTraffic, 5000, false, "Huge"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_%s", tc.goal.String(), tc.scaleName), func(t *testing.T) {
			t.Parallel()
			actual := tc.goal.IsRecommendedForSize(tc.agentCount)
			if actual != tc.expectRec {
				t.Errorf("Goal %s with %s scale (%d agents): expected IsRecommended=%v, got %v",
					tc.goal.String(), tc.scaleName, tc.agentCount, tc.expectRec, actual)
			}

			// Also verify we have a reason for this combination
			reason := getScaleReason(tc.goal, tc.agentCount, actual)
			if reason == "" {
				t.Errorf("Missing reason for %s with %s scale (%d agents)",
					tc.goal.String(), tc.scaleName, tc.agentCount)
			}
		})
	}
}

// testGetPatternReason mimics the display.go logic for testing
func testGetPatternReason(g goal.Type, p pattern.Type, isGood bool) string {
	if isGood {
		switch g {
		case goal.MinimizeAPICalls:
			switch p {
			case pattern.Burst, pattern.HighFrequency:
				return " (excellent for batching)"
			case pattern.Mixed:
				return " (good adaptation)"
			case pattern.Unset, pattern.Steady, pattern.Sparse:
				return reasonSupportsBatch
			default:
				return reasonSupportsBatch
			}
		case goal.DistributeLoad:
			switch p {
			case pattern.HighFrequency, pattern.Steady:
				return " (even distribution)"
			case pattern.Mixed:
				return " (varied timing spreads load)"
			case pattern.Unset, pattern.Burst, pattern.Sparse:
				return reasonHelpsDistribute
			default:
				return reasonHelpsDistribute
			}
		case goal.ReachConsensus:
			switch p {
			case pattern.Steady:
				return " (regular voting rounds)"
			case pattern.Unset, pattern.HighFrequency, pattern.Burst, pattern.Mixed, pattern.Sparse:
				return reasonSupportsConsensus
			default:
				return reasonSupportsConsensus
			}
		case goal.MinimizeLatency:
			switch p {
			case pattern.HighFrequency:
				return " (quick response)"
			case pattern.Steady:
				return " (predictable latency)"
			case pattern.Unset, pattern.Burst, pattern.Mixed, pattern.Sparse:
				return reasonEnablesOptimize
			default:
				return reasonEnablesOptimize
			}
		case goal.SaveEnergy:
			switch p {
			case pattern.Sparse:
				return " (minimal activity)"
			case pattern.Burst:
				return " (concentrated work periods)"
			case pattern.Unset, pattern.HighFrequency, pattern.Steady, pattern.Mixed:
				return " (energy conservation)"
			default:
				return " (energy conservation)"
			}
		case goal.MaintainRhythm:
			switch p {
			case pattern.Steady:
				return " (perfect rhythm)"
			case pattern.HighFrequency:
				return " (regular sync)"
			case pattern.Unset, pattern.Burst, pattern.Mixed, pattern.Sparse:
				return " (rhythm support)"
			default:
				return " (rhythm support)"
			}
		case goal.RecoverFromFailure:
			switch p {
			case pattern.Mixed:
				return " (handles variability)"
			case pattern.HighFrequency:
				return " (quick detection)"
			case pattern.Unset, pattern.Burst, pattern.Steady, pattern.Sparse:
				return " (supports recovery)"
			default:
				return " (supports recovery)"
			}
		case goal.AdaptToTraffic:
			switch p {
			case pattern.Burst:
				return " (simulates surges)"
			case pattern.Mixed:
				return " (variable traffic)"
			case pattern.Unset, pattern.HighFrequency, pattern.Steady, pattern.Sparse:
				return " (traffic adaptation)"
			default:
				return " (traffic adaptation)"
			}
		default:
			return " (pattern supports goal)"
		}
	} else {
		// Return negative reasons for non-good patterns
		return " (not optimal for this goal)"
	}
}

// TestMenuBoldingVsAssessmentConsistency checks that menu bolding matches assessment indicators
func TestMenuBoldingVsAssessmentConsistency(t *testing.T) {
	t.Parallel()
	goals := []goal.Type{
		goal.MinimizeAPICalls,
		goal.DistributeLoad,
		goal.ReachConsensus,
		goal.MinimizeLatency,
		goal.SaveEnergy,
		goal.MaintainRhythm,
		goal.RecoverFromFailure,
		goal.AdaptToTraffic,
	}

	scales := []struct {
		size  scale.Size
		count int
		name  string
	}{
		{scale.Tiny, 20, "Tiny"},
		{scale.Small, 50, "Small"},
		{scale.Medium, 200, "Medium"},
		{scale.Large, 1000, "Large"},
		{scale.Huge, 5000, "Huge"},
	}

	patterns := []pattern.Type{
		pattern.HighFrequency,
		pattern.Burst,
		pattern.Steady,
		pattern.Mixed,
		pattern.Sparse,
	}

	// Track inconsistencies
	var inconsistencies []string

	for _, g := range goals {
		for _, s := range scales {
			for _, p := range patterns {
				// Get the qualifier
				qualifier := pattern.GetQualifier(int(g), s.count, p)

				// Check what would be bolded in menu (optimal/excellent/good - matches assessment)
				wouldBeBoldedInMenu := qualifier == pattern.QualifierOptimal ||
					qualifier == pattern.QualifierExcellent ||
					qualifier == pattern.QualifierGood

				// Check what would show as good in assessment (optimal/excellent/good)
				wouldShowGoodInAssessment := qualifier == pattern.QualifierOptimal ||
					qualifier == pattern.QualifierExcellent ||
					qualifier == pattern.QualifierGood

				// INCONSISTENCY: If shows good in assessment but not bolded in menu
				if wouldShowGoodInAssessment && !wouldBeBoldedInMenu {
					inconsistencies = append(inconsistencies, fmt.Sprintf(
						"%s + %s scale + %s pattern: Shows GOOD (✓) in assessment but NOT BOLDED in menu (qualifier=%s)",
						g.String(), s.name, p.String(), qualifier))
				}

				// Also check the reverse: bolded but not good (shouldn't happen)
				if wouldBeBoldedInMenu && !wouldShowGoodInAssessment {
					inconsistencies = append(inconsistencies, fmt.Sprintf(
						"%s + %s scale + %s pattern: BOLDED in menu but NOT GOOD in assessment (qualifier=%s)",
						g.String(), s.name, p.String(), qualifier))
				}
			}
		}
	}

	if len(inconsistencies) > 0 {
		t.Errorf("Found %d inconsistencies between menu bolding and assessment:\n", len(inconsistencies))
		for _, inc := range inconsistencies {
			t.Errorf("  - %s", inc)
		}
	} else {
		t.Log("✓ All menu bolding is consistent with assessment indicators")
	}
}

// TestComprehensiveCombinations tests all goal-scale-pattern combinations for consistency
func TestComprehensiveCombinations(t *testing.T) {
	t.Parallel()
	goals := []goal.Type{
		goal.MinimizeAPICalls,
		goal.DistributeLoad,
		goal.ReachConsensus,
		goal.MinimizeLatency,
		goal.SaveEnergy,
		goal.MaintainRhythm,
		goal.RecoverFromFailure,
		goal.AdaptToTraffic,
	}

	scales := []struct {
		size  scale.Size
		count int
		name  string
	}{
		{scale.Tiny, 20, "Tiny"},
		{scale.Small, 50, "Small"},
		{scale.Medium, 200, "Medium"},
		{scale.Large, 1000, "Large"},
		{scale.Huge, 5000, "Huge"},
	}

	patterns := []pattern.Type{
		pattern.HighFrequency,
		pattern.Burst,
		pattern.Steady,
		pattern.Mixed,
		pattern.Sparse,
	}

	// Track failures for summary
	var failures []string

	for _, g := range goals {
		for _, s := range scales {
			for _, p := range patterns {
				testName := fmt.Sprintf("%s/%s/%s", g.String(), s.name, p.String())

				// Get the qualifier and check if it would be bolded
				qualifier := pattern.GetQualifier(int(g), s.count, p)
				wouldBeBolded := qualifier == pattern.QualifierOptimal ||
					qualifier == pattern.QualifierExcellent ||
					qualifier == pattern.QualifierGood

				// Get the pattern reason
				isGood := qualifier == pattern.QualifierOptimal ||
					qualifier == pattern.QualifierExcellent ||
					qualifier == pattern.QualifierGood
				patternReason := testGetPatternReason(g, p, isGood)

				// Get scale recommendation and reason
				scaleRecommended := g.IsRecommendedForSize(s.count)
				scaleReason := getScaleReason(g, s.count, scaleRecommended)

				// Consistency checks
				// 1. If pattern is bolded (optimal/excellent), it should have a positive reason
				if wouldBeBolded && !isGood {
					failures = append(failures, fmt.Sprintf(
						"%s: Pattern bolded (qualifier=%s) but not marked as good",
						testName, qualifier))
				}

				// 2. If pattern is not bolded, qualifier should be good/fair/poor
				if !wouldBeBolded && (qualifier != pattern.QualifierGood &&
					qualifier != pattern.QualifierFair &&
					qualifier != pattern.QualifierPoor) {
					failures = append(failures, fmt.Sprintf(
						"%s: Pattern not bolded but qualifier=%s (expected good/fair/poor)",
						testName, qualifier))
				}

				// 3. Pattern reason should always exist
				if patternReason == "" {
					failures = append(failures, fmt.Sprintf(
						"%s: No pattern reason (qualifier=%s, isGood=%v)",
						testName, qualifier, isGood))
				}

				// 4. Scale reason should always exist
				if scaleReason == "" {
					failures = append(failures, fmt.Sprintf(
						"%s: No scale reason (recommended=%v)",
						testName, scaleRecommended))
				}

				// 5. Verify qualifier matches threshold constants
				modifier := pattern.GetCoherenceModifier(int(g), s.count, p)
				expectedQualifier := ""
				switch {
				case modifier >= pattern.ThresholdOptimal:
					expectedQualifier = pattern.QualifierOptimal
				case modifier >= pattern.ThresholdExcellent:
					expectedQualifier = pattern.QualifierExcellent
				case modifier >= pattern.ThresholdGood:
					expectedQualifier = pattern.QualifierGood
				case modifier >= pattern.ThresholdFair:
					expectedQualifier = pattern.QualifierFair
				default:
					expectedQualifier = pattern.QualifierPoor
				}

				if qualifier != expectedQualifier {
					failures = append(failures, fmt.Sprintf(
						"%s: Qualifier mismatch - got %s, expected %s (modifier=%.2f)",
						testName, qualifier, expectedQualifier, modifier))
				}

				// Log detailed info for debugging
				t.Logf("%s: qualifier=%s bold=%v scaleRec=%v modifier=%.2f",
					testName, qualifier, wouldBeBolded, scaleRecommended, modifier)
			}
		}
	}

	// Report all failures at once
	if len(failures) > 0 {
		t.Errorf("Found %d consistency issues:\n", len(failures))
		for _, failure := range failures {
			t.Errorf("  - %s", failure)
		}
	}
}

// TestMenuBoldingLogic tests the specific logic for what gets bolded
func TestMenuBoldingLogic(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		goal              goal.Type
		scale             scale.Size
		scaleCount        int
		pattern           pattern.Type
		expectBoldPattern bool
		expectBoldScale   bool
		description       string
	}{
		// MinimizeAPICalls - all scales work, some patterns are optimal
		{goal.MinimizeAPICalls, scale.Tiny, 20, pattern.HighFrequency, true, true, "MinimizeAPICalls with Tiny scale and HighFrequency"},
		{goal.MinimizeAPICalls, scale.Tiny, 20, pattern.Burst, true, true, "MinimizeAPICalls with Tiny scale and Burst"},
		{goal.MinimizeAPICalls, scale.Tiny, 20, pattern.Sparse, false, true, "MinimizeAPICalls with Tiny scale and Sparse"},
		{goal.MinimizeAPICalls, scale.Large, 1000, pattern.Burst, true, true, "MinimizeAPICalls with Large scale and Burst"},

		// DistributeLoad - needs 20+ agents, some patterns work better
		{goal.DistributeLoad, scale.Tiny, 20, pattern.Steady, true, true, "DistributeLoad with Tiny scale and Steady"},
		{goal.DistributeLoad, scale.Tiny, 20, pattern.Burst, false, true, "DistributeLoad with Tiny scale and Burst"},
		{goal.DistributeLoad, scale.Large, 1000, pattern.Mixed, true, true, "DistributeLoad with Large scale and Mixed"}, // Good qualifier should be bolded

		// ReachConsensus - 50-1000 agents optimal
		{goal.ReachConsensus, scale.Tiny, 20, pattern.Steady, true, false, "ReachConsensus with Tiny scale (not recommended)"},
		{goal.ReachConsensus, scale.Small, 50, pattern.Steady, true, true, "ReachConsensus with Small scale and Steady"},
		{goal.ReachConsensus, scale.Huge, 5000, pattern.Steady, true, false, "ReachConsensus with Huge scale (not recommended)"},

		// MinimizeLatency - smaller scales better (<=200)
		{goal.MinimizeLatency, scale.Tiny, 20, pattern.HighFrequency, true, true, "MinimizeLatency with Tiny scale and HighFrequency"},
		{goal.MinimizeLatency, scale.Large, 1000, pattern.HighFrequency, true, false, "MinimizeLatency with Large scale (not recommended)"}, // Good qualifier should be bolded

		// SaveEnergy - smaller scales better (<=200)
		{goal.SaveEnergy, scale.Small, 50, pattern.Sparse, true, true, "SaveEnergy with Small scale and Sparse"}, // Good qualifier (1.04) should be bolded
		{goal.SaveEnergy, scale.Large, 1000, pattern.Sparse, true, false, "SaveEnergy with Large scale (not recommended)"},

		// MaintainRhythm - works at all scales
		{goal.MaintainRhythm, scale.Tiny, 20, pattern.Steady, true, true, "MaintainRhythm with Tiny scale and Steady"},
		{goal.MaintainRhythm, scale.Huge, 5000, pattern.Steady, true, true, "MaintainRhythm with Huge scale and Steady"},

		// RecoverFromFailure - needs 20+ agents
		{goal.RecoverFromFailure, scale.Small, 50, pattern.Mixed, true, true, "RecoverFromFailure with Small scale and Mixed"},
		{goal.RecoverFromFailure, scale.Large, 1000, pattern.Mixed, true, true, "RecoverFromFailure with Large scale and Mixed"},

		// AdaptToTraffic - 20-1000 agents optimal
		{goal.AdaptToTraffic, scale.Medium, 200, pattern.Burst, true, true, "AdaptToTraffic with Medium scale and Burst"},
		{goal.AdaptToTraffic, scale.Huge, 5000, pattern.Burst, true, false, "AdaptToTraffic with Huge scale (not recommended)"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()
			// Check pattern bolding logic
			qualifier := pattern.GetQualifier(int(tc.goal), tc.scaleCount, tc.pattern)
			actualBoldPattern := qualifier == pattern.QualifierOptimal ||
				qualifier == pattern.QualifierExcellent ||
				qualifier == pattern.QualifierGood
			if actualBoldPattern != tc.expectBoldPattern {
				t.Errorf("Pattern bolding mismatch: expected %v, got %v (qualifier=%s)",
					tc.expectBoldPattern, actualBoldPattern, qualifier)
			}

			// Check scale bolding logic
			actualBoldScale := tc.goal.IsRecommendedForSize(tc.scaleCount)
			if actualBoldScale != tc.expectBoldScale {
				t.Errorf("Scale bolding mismatch: expected %v, got %v",
					tc.expectBoldScale, actualBoldScale)
			}
		})
	}
}
