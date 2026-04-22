package main

import (
	"flag"
	"fmt"
	"io"
	"time"

	"simhash-logs/internal/search"
)

func runEval(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("eval", flag.ContinueOnError)
	fs.SetOutput(stderr)

	inputPath := fs.String("input", "", "Path to a log file (default: stdin)")
	k := fs.Int("k", 3, "Max Hamming distance threshold for near-duplicates")
	maxLines := fs.Int("max", 5000, "Max number of lines to read")
	bandsFlag := fs.Int("bands", 0, "Number of LSH bands (0 = auto, default auto is k+1)")
	csvOut := fs.Bool("csv", false, "Output results in CSV format")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	bands := *bandsFlag
	if bands == 0 {
		bands = *k + 1
	}

	lines, err := readLines(*inputPath, *maxLines, stdin)
	if err != nil {
		fmt.Fprintf(stderr, "read error: %v\n", err)
		return 1
	}

	records := buildRecords(lines)
	sigs := make([]uint64, len(records))
	for i := range records {
		sigs[i] = records[i].Sig
	}

	// Run Brute Force as Ground Truth
	bruteStart := time.Now()
	brutePairs := search.BruteNearDuplicates(sigs, *k)
	bruteElapsed := time.Since(bruteStart)
	
	// Create a set of true pairs for O(1) lookup
	truePairs := make(map[[2]int]struct{})
	for _, p := range brutePairs {
		truePairs[[2]int{p.I, p.J}] = struct{}{}
	}

	// Run LSH
	lshStart := time.Now()
	lshPairs, lshComps := search.LSHNearDuplicates(sigs, *k, bands)
	lshElapsed := time.Since(lshStart)

	// Calculate recall
	truePositives := 0
	for _, p := range lshPairs {
		if _, ok := truePairs[[2]int{p.I, p.J}]; ok {
			truePositives++
		}
	}

	totalActual := len(truePairs)
	recall := 0.0
	if totalActual > 0 {
		recall = float64(truePositives) / float64(totalActual) * 100.0
	}

	n := len(sigs)
	bruteComps := n * (n - 1) / 2
	compReduction := 0.0
	if bruteComps > 0 {
		compReduction = float64(bruteComps - lshComps) / float64(bruteComps) * 100.0
	}

	if *csvOut {
		fmt.Fprintf(stdout, "k,bands,records,brute_ms,lsh_ms,brute_comps,lsh_comps,total_actual,true_positives,recall_pct,comp_reduction_pct\n")
		fmt.Fprintf(stdout, "%d,%d,%d,%d,%d,%d,%d,%d,%d,%.2f,%.2f\n",
			*k, bands, len(records), bruteElapsed.Milliseconds(), lshElapsed.Milliseconds(),
			bruteComps, lshComps, totalActual, truePositives, recall, compReduction)
	} else {
		fmt.Fprintf(stdout, "Evaluation Results:\n")
		fmt.Fprintf(stdout, "-------------------\n")
		fmt.Fprintf(stdout, "Records:           %d\n", len(records))
		fmt.Fprintf(stdout, "Distance (k):      %d\n", *k)
		fmt.Fprintf(stdout, "LSH Bands:         %d\n", bands)
		fmt.Fprintf(stdout, "\nBrute Force Ground Truth:\n")
		fmt.Fprintf(stdout, "  Matches Found:   %d\n", totalActual)
		fmt.Fprintf(stdout, "  Comparisons:     %d\n", bruteComps)
		fmt.Fprintf(stdout, "  Time:            %v\n", bruteElapsed)
		fmt.Fprintf(stdout, "\nLSH Approach:\n")
		fmt.Fprintf(stdout, "  Matches Found:   %d (True Positives: %d)\n", len(lshPairs), truePositives)
		fmt.Fprintf(stdout, "  Recall:          %.2f%%\n", recall)
		fmt.Fprintf(stdout, "  Comparisons:     %d (%.2f%% reduction)\n", lshComps, compReduction)
		fmt.Fprintf(stdout, "  Time:            %v\n", lshElapsed)
	}

	return 0
}
