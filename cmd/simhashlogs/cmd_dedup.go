package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"time"

	"simhash-logs/internal/search"
)

type matchOutput struct {
	Distance    int    `json:"distance"`
	RawA        string `json:"raw_a,omitempty"`
	RawB        string `json:"raw_b,omitempty"`
	NormalizedA string `json:"normalized_a"`
	NormalizedB string `json:"normalized_b"`
}

func runDedup(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("dedup", flag.ContinueOnError)
	fs.SetOutput(stderr)

	inputPath := fs.String("input", "", "Path to a log file (default: stdin)")
	k := fs.Int("k", 3, "Max Hamming distance threshold for near-duplicates")
	maxLines := fs.Int("max", 5000, "Max number of lines to read")
	limit := fs.Int("limit", 0, "Max number of matches to print (0 = no limit)")
	printRaw := fs.Bool("print-raw", false, "Print raw lines alongside normalized lines")
	jsonOut := fs.Bool("json", false, "Print matches as JSON")
	useLSH := fs.Bool("use-lsh", false, "Use LSH candidate generation before exact Hamming verification")
	indexMode := fs.String("index", indexBrute, "Candidate index to use: brute, lsh, or paper")
	bandsFlag := fs.Int("bands", 0, "Number of LSH bands (0 = auto, default auto is k+1)")
	tablesFlag := fs.Int("tables", 0, "Number of paper-style permutation tables (0 = auto, default auto is k+1)")
	windowFlag := fs.Int("window", 0, "Paper-style sorted-table scan window per table (0 = auto)")
	quietStats := fs.Bool("quiet-stats", false, "Disable stats output on stderr")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	indexExplicit := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "index" {
			indexExplicit = true
		}
	})
	mode := *indexMode
	if *useLSH && !indexExplicit {
		mode = indexLSH
	}
	if err := validateIndexMode(mode); err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 2
	}
	if *bandsFlag < 0 {
		fmt.Fprintf(stderr, "invalid -bands: bands must be >= 0\n")
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
	if *k >= 64 {
		mode = indexBrute
	}

	lines, err := readLines(*inputPath, *maxLines, stdin)
	if err != nil {
		fmt.Fprintf(stderr, "read error: %v\n", err)
		return 1
	}

	prepStart := time.Now()
	records := buildRecords(lines)
	sigs := make([]uint64, len(records))
	for i := range records {
		sigs[i] = records[i].Sig
	}
	prepElapsed := time.Since(prepStart)

	searchStart := time.Now()
	var pairs []search.Pair
	comparisons := 0

	// Optimized indexes are candidate generators only; every candidate is still
	// verified with exact Hamming distance before it becomes an output match.
	if mode == indexLSH {
		bands := *bandsFlag
		if bands == 0 {
			bands = defaultTables(*k)
		}
		pairs, comparisons = search.LSHNearDuplicates(sigs, *k, bands)
	} else if mode == indexPaper {
		tables := *tablesFlag
		if tables == 0 {
			tables = defaultTables(*k)
		}
		window := *windowFlag
		if window == 0 {
			window = defaultWindow(*k)
		}
		pairs, comparisons = search.PaperNearDuplicates(sigs, *k, tables, window)
	} else {
		pairs = search.BruteNearDuplicates(sigs, *k)
		n := len(sigs)
		comparisons = n * (n - 1) / 2
	}
	searchElapsed := time.Since(searchStart)

	if *limit > 0 && len(pairs) > *limit {
		pairs = pairs[:*limit]
	}

	// Stats go to stderr so stdout remains machine-readable when -json is set.
	if !*quietStats {
		bands := 0
		tables := 0
		window := 0
		if mode == indexLSH {
			if *bandsFlag == 0 {
				bands = defaultTables(*k)
			} else {
				bands = *bandsFlag
			}
		} else if mode == indexPaper {
			if *tablesFlag == 0 {
				tables = defaultTables(*k)
			} else {
				tables = *tablesFlag
			}
			if *windowFlag == 0 {
				window = defaultWindow(*k)
			} else {
				window = *windowFlag
			}
		}
		fmt.Fprintf(stderr,
			"stats mode=%s bands=%d tables=%d window=%d records=%d comparisons=%d matches=%d prep_ms=%d search_ms=%d\n",
			mode, bands, tables, window, len(records), comparisons, len(pairs),
			prepElapsed.Milliseconds(), searchElapsed.Milliseconds(),
		)
	}

	if *jsonOut {
		out := make([]matchOutput, 0, len(pairs))
		for _, p := range pairs {
			item := matchOutput{
				Distance:    p.Distance,
				NormalizedA: records[p.I].Normalized,
				NormalizedB: records[p.J].Normalized,
			}
			if *printRaw {
				item.RawA = records[p.I].Raw
				item.RawB = records[p.J].Raw
			}
			out = append(out, item)
		}

		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(out); err != nil {
			fmt.Fprintf(stderr, "json encode error: %v\n", err)
			return 1
		}
		return 0
	}

	for _, p := range pairs {
		if *printRaw {
			fmt.Fprintf(stdout, "match (dist=%d)\nA(raw): %s\nB(raw): %s\nA(norm): %s\nB(norm): %s\n\n",
				p.Distance, records[p.I].Raw, records[p.J].Raw,
				records[p.I].Normalized, records[p.J].Normalized)
		} else {
			fmt.Fprintf(stdout, "match (dist=%d): %s || %s\n",
				p.Distance, records[p.I].Normalized, records[p.J].Normalized)
		}
	}

	return 0
}
