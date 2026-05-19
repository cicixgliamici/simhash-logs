package main

import (
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"simhash-logs/internal/search"
)

type evalResult struct {
	K            int
	Index        string
	Bands        int
	Tables       int
	Window       int
	Records      int
	BruteElapsed time.Duration
	IndexElapsed time.Duration
	BruteComps   int
	IndexComps   int
	TotalActual  int
	IndexMatches int
	TruePositive int
	RecallPct    float64
	ReductionPct float64
}

func runEval(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("eval", flag.ContinueOnError)
	fs.SetOutput(stderr)

	inputPath := fs.String("input", "", "Path to a log file (default: stdin)")
	k := fs.Int("k", 3, "Max Hamming distance threshold for near-duplicates")
	kValuesFlag := fs.String("k-values", "", "Comma-separated k values for evaluation sweep (overrides -k)")
	maxLines := fs.Int("max", 5000, "Max number of lines to read")
	indexValuesFlag := fs.String("index-values", "lsh,paper", "Comma-separated candidate indexes to evaluate: lsh,paper")
	bandsFlag := fs.Int("bands", 0, "Number of LSH bands (0 = auto, default auto is k+1)")
	bandsValuesFlag := fs.String("bands-values", "", "Comma-separated band counts for evaluation sweep (0 = auto per k)")
	tablesFlag := fs.Int("tables", 0, "Number of paper-style permutation tables (0 = auto, default auto is k+1)")
	windowFlag := fs.Int("window", 0, "Paper-style sorted-table scan window per table (0 = auto)")
	csvOut := fs.Bool("csv", false, "Output results in CSV format")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	kValues, err := evalIntValues(*kValuesFlag, *k, "k")
	if err != nil {
		fmt.Fprintf(stderr, "invalid -k-values: %v\n", err)
		return 2
	}
	indexValues, err := evalIndexValues(*indexValuesFlag)
	if err != nil {
		fmt.Fprintf(stderr, "invalid -index-values: %v\n", err)
		return 2
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

	bandsValues, err := evalIntValues(*bandsValuesFlag, *bandsFlag, "bands")
	if err != nil {
		fmt.Fprintf(stderr, "invalid -bands-values: %v\n", err)
		return 2
	}

	if *tablesFlag < 0 {
		fmt.Fprintf(stderr, "invalid -tables: tables must be >= 0\n")
		return 2
	}
	if *windowFlag < 0 {
		fmt.Fprintf(stderr, "invalid -window: window must be >= 0\n")
		return 2
	}

	results := make([]evalResult, 0, len(kValues)*len(bandsValues)*len(indexValues))
	for _, kv := range kValues {
		for _, index := range indexValues {
			if index == indexLSH {
				for _, bandsValue := range bandsValues {
					bands := bandsValue
					if bands == 0 {
						bands = defaultTables(kv)
					}
					results = append(results, evaluateIndex(sigs, kv, index, bands, 0, 0))
				}
				continue
			}

			tables := *tablesFlag
			if tables == 0 {
				tables = defaultTables(kv)
			}
			window := *windowFlag
			if window == 0 {
				window = defaultWindow(kv)
			}
			results = append(results, evaluateIndex(sigs, kv, index, 0, tables, window))
		}
	}

	if *csvOut {
		fmt.Fprintf(stdout, "index,k,bands,tables,window,records,brute_ms,index_ms,brute_comps,index_comps,total_actual,index_matches,true_positives,recall_pct,comp_reduction_pct\n")
		for _, r := range results {
			fmt.Fprintf(stdout, "%s,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%.2f,%.2f\n",
				r.Index, r.K, r.Bands, r.Tables, r.Window, r.Records,
				r.BruteElapsed.Milliseconds(), r.IndexElapsed.Milliseconds(),
				r.BruteComps, r.IndexComps, r.TotalActual, r.IndexMatches, r.TruePositive,
				r.RecallPct, r.ReductionPct)
		}
		return 0
	}

	for i, r := range results {
		if i > 0 {
			fmt.Fprintln(stdout)
		}
		fmt.Fprintf(stdout, "Evaluation Results:\n")
		fmt.Fprintf(stdout, "-------------------\n")
		fmt.Fprintf(stdout, "Records:           %d\n", r.Records)
		fmt.Fprintf(stdout, "Distance (k):      %d\n", r.K)
		fmt.Fprintf(stdout, "Index:             %s\n", r.Index)
		if r.Index == indexLSH {
			fmt.Fprintf(stdout, "LSH Bands:         %d\n", r.Bands)
		} else if r.Index == indexPaper {
			fmt.Fprintf(stdout, "Paper Tables:      %d\n", r.Tables)
			fmt.Fprintf(stdout, "Paper Window:      %d\n", r.Window)
		}
		fmt.Fprintf(stdout, "\nBrute Force Ground Truth:\n")
		fmt.Fprintf(stdout, "  Matches Found:   %d\n", r.TotalActual)
		fmt.Fprintf(stdout, "  Comparisons:     %d\n", r.BruteComps)
		fmt.Fprintf(stdout, "  Time:            %v\n", r.BruteElapsed)
		fmt.Fprintf(stdout, "\n%s Approach:\n", r.Index)
		fmt.Fprintf(stdout, "  Matches Found:   %d (True Positives: %d)\n", r.IndexMatches, r.TruePositive)
		fmt.Fprintf(stdout, "  Recall:          %.2f%%\n", r.RecallPct)
		fmt.Fprintf(stdout, "  Comparisons:     %d (%.2f%% reduction)\n", r.IndexComps, r.ReductionPct)
		fmt.Fprintf(stdout, "  Time:            %v\n", r.IndexElapsed)
	}

	return 0
}

func evaluateLSH(sigs []uint64, k, bands int) evalResult {
	return evaluateIndex(sigs, k, indexLSH, bands, 0, 0)
}

func evaluateIndex(sigs []uint64, k int, index string, bands, tables, window int) evalResult {
	// Brute force is the local ground truth for evaluating candidate-generation
	// recall on reviewable datasets.
	bruteStart := time.Now()
	brutePairs := search.BruteNearDuplicates(sigs, k)
	bruteElapsed := time.Since(bruteStart)

	// Store true pairs for O(1) recall checks.
	truePairs := make(map[[2]int]struct{})
	for _, p := range brutePairs {
		truePairs[[2]int{p.I, p.J}] = struct{}{}
	}

	indexStart := time.Now()
	var indexPairs []search.Pair
	indexComps := 0
	switch index {
	case indexPaper:
		indexPairs, indexComps = search.PaperNearDuplicates(sigs, k, tables, window)
	case indexLSH:
		indexPairs, indexComps = search.LSHNearDuplicates(sigs, k, bands)
	}
	indexElapsed := time.Since(indexStart)

	truePositives := 0
	for _, p := range indexPairs {
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
		compReduction = float64(bruteComps-indexComps) / float64(bruteComps) * 100.0
	}

	return evalResult{
		K:            k,
		Index:        index,
		Bands:        bands,
		Tables:       tables,
		Window:       window,
		Records:      len(sigs),
		BruteElapsed: bruteElapsed,
		IndexElapsed: indexElapsed,
		BruteComps:   bruteComps,
		IndexComps:   indexComps,
		TotalActual:  totalActual,
		IndexMatches: len(indexPairs),
		TruePositive: truePositives,
		RecallPct:    recall,
		ReductionPct: compReduction,
	}
}

func evalIndexValues(raw string) ([]string, error) {
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			return nil, fmt.Errorf("index list contains an empty value")
		}
		if err := validateIndexMode(value); err != nil {
			return nil, err
		}
		if value == indexBrute {
			return nil, fmt.Errorf("eval indexes must be candidate indexes: lsh or paper")
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}
	return values, nil
}

func evalIntValues(raw string, fallback int, name string) ([]int, error) {
	if strings.TrimSpace(raw) == "" {
		if err := validateEvalIntValue(fallback, name); err != nil {
			return nil, err
		}
		return []int{fallback}, nil
	}

	parts := strings.Split(raw, ",")
	values := make([]int, 0, len(parts))
	seen := make(map[int]struct{}, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, fmt.Errorf("%s list contains an empty value", name)
		}

		value, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("%q is not an integer", part)
		}
		if err := validateEvalIntValue(value, name); err != nil {
			return nil, err
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}

	return values, nil
}

func validateEvalIntValue(value int, name string) error {
	if value < 0 {
		return fmt.Errorf("%s values must be >= 0", name)
	}
	if name == "k" && value > 64 {
		return fmt.Errorf("k values must be <= 64")
	}
	if name == "bands" && value > 64 {
		return fmt.Errorf("bands values must be <= 64")
	}
	return nil
}
