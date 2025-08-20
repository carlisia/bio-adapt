package display

import (
	"fmt"
)

// resultOpts holds options for result display.
type resultOpts struct {
	batchReduction *batchReductionInfo
}

// batchReductionInfo holds batch reduction metrics.
type batchReductionInfo struct {
	totalAgents  int
	finalBatches int
}

// ResultOpt is an option for PrintResultsSummary.
type ResultOpt func(*resultOpts)

// PrintResultsSummary displays final results in a formatted box.
func PrintResultsSummary(initial, final, target, improvement float64, opts ...ResultOpt) {
	options := &resultOpts{}
	for _, opt := range opts {
		opt(options)
	}

	fmt.Println("\n╔════════════════════════════════════════════════╗")
	fmt.Println("║               FINAL RESULTS                   ║")
	fmt.Println("╠════════════════════════════════════════════════╣")
	fmt.Printf("║ Initial Coherence:     %6.1f%%                ║\n", initial*100)
	fmt.Printf("║ Final Coherence:       %6.1f%%                ║\n", final*100)
	fmt.Printf("║ Target:                %6.1f%%                ║\n", target*100)
	fmt.Printf("║ Improvement:           %6.1f%%                ║\n", improvement*100)

	if options.batchReduction != nil {
		reduction := float64(options.batchReduction.totalAgents-options.batchReduction.finalBatches) /
			float64(options.batchReduction.totalAgents) * 100
		fmt.Println("╠════════════════════════════════════════════════╣")
		fmt.Printf("║ API Calls:       %3d → %3d (%.0f%% reduction)   ║\n",
			options.batchReduction.totalAgents,
			options.batchReduction.finalBatches,
			reduction)
	}

	fmt.Println("╠════════════════════════════════════════════════╣")

	// Success/failure indicator
	if final >= target {
		if UseEmoji() {
			fmt.Println("║         ✅ TARGET REACHED! ✅                 ║")
		} else {
			fmt.Println("║         [OK] TARGET REACHED!                  ║")
		}
	} else {
		gap := (target - final) * 100
		if UseEmoji() {
			fmt.Printf("║    ⚠️  Missed target by %.1f%% ⚠️              ║\n", gap)
		} else {
			fmt.Printf("║    [!!] Missed target by %.1f%%                ║\n", gap)
		}
	}

	fmt.Println("╚════════════════════════════════════════════════╝")
}

// WithBatchReduction adds batch reduction info to results.
func WithBatchReduction(totalAgents, finalBatches int) ResultOpt {
	return func(opts *resultOpts) {
		opts.batchReduction = &batchReductionInfo{
			totalAgents:  totalAgents,
			finalBatches: finalBatches,
		}
	}
}
