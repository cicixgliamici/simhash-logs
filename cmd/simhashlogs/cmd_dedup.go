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
	bandsFlag := fs.Int("bands", 0, "Number of LSH bands (0 = auto, default auto is k+1)")
	quietStats := fs.Bool("quiet-stats", false, "Disable stats output on stderr")

	if err := fs.Parse(args); err != nil {
		return 2
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

	if *useLSH && *k < 64 {
		bands := *bandsFlag
		if bands == 0 {
			bands = *k + 1
		}
		pairs, comparisons = search.LSHNearDuplicates(sigs, *k, bands)
	} else {
		pairs = search.BruteNearDuplicates(sigs, *k)
		n := len(sigs)
		comparisons = n * (n - 1) / 2
	}
	searchElapsed := time.Since(searchStart)

	if *limit > 0 && len(pairs) > *limit {
		pairs = pairs[:*limit]
	}

	if !*quietStats {
		mode := "brute"
		bands := 0
		if *useLSH && *k < 64 {
			mode = "lsh"
			if *bandsFlag == 0 {
				bands = *k + 1
			} else {
				bands = *bandsFlag
			}
		}
		fmt.Fprintf(stderr,
			"stats mode=%s bands=%d records=%d comparisons=%d matches=%d prep_ms=%d search_ms=%d\n",
			mode, bands, len(records), comparisons, len(pairs),
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
