package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"time"
	
	"simhash-logs/internal/normalize"
	"simhash-logs/internal/search"
	"simhash-logs/internal/simhash"
	"simhash-logs/internal/tokenize"
)

type matchOutput struct {
	Distance    int    `json:"distance"`
	RawA        string `json:"raw_a,omitempty"`
	RawB        string `json:"raw_b,omitempty"`
	NormalizedA string `json:"normalized_a"`
	NormalizedB string `json:"normalized_b"`
}

func main() {
	code := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	os.Exit(code)
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("simhashlogs", flag.ContinueOnError)
	fs.SetOutput(stderr)

	inputPath := fs.String("input", "", "Path to a log file (default: stdin)")
	k := fs.Int("k", 3, "Max Hamming distance threshold for near-duplicates")
	maxLines := fs.Int("max", 5000, "Max number of lines to read (keeps brute-force manageable)")
	printRaw := fs.Bool("print-raw", false, "Print raw lines alongside normalized lines")
	jsonOut := fs.Bool("json", false, "Print matches as JSON")
	useLSH := fs.Bool("use-lsh", false, "Use LSH candidate generation before exact Hamming verification")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	lines, err := readLines(*inputPath, *maxLines, stdin)
	fmt.Fprintf(stderr, "read %d lines\n", len(lines))
	if err != nil {
		fmt.Fprintf(stderr, "read error: %v\n", err)
		return 1
	}

	srecords := buildRecords(lines)
	pairs := search.BruteNearDuplicates(records, *k)
	if *useLSH && *k < 64 {
		bands := *k + 1
		pairs = lshNearDuplicates(sigs, *k, bands)
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
		fmt.Fprintf(stdout, "match (dist=%d)\n", p.Distance)
		if *printRaw {
			fmt.Fprintf(stdout, "  A(raw): %s\n", records[p.I].Raw)
			fmt.Fprintf(stdout, "  B(raw): %s\n", records[p.J].Raw)
		}
		fmt.Fprintf(stdout, "  A: %s\n", records[p.I].Normalized)
		fmt.Fprintf(stdout, "  B: %s\n", records[p.J].Normalized)
		fmt.Fprintln(stdout)
	}

	return 0
}

func buildRecords(lines []string) []search.Record {
	records := make([]search.Record, 0, len(lines))

	for _, line := range lines {
		normalized := normalize.Line(line)
		tokens := tokenize.Simple(normalized)
		sig := simhash.SimHash64(tokens)

		records = append(records, search.Record{
			Raw:        line,
			Normalized: normalized,
			Tokens:     tokens,
			Sig:        sig,
		})
	}

	return records
}

func lshNearDuplicates(sigs []uint64, k, bands int) []search.Pair {
	if len(sigs) == 0 {
		return nil
	}

	idx := search.NewBandIndex(bands)
	pairSeen := make(map[[2]int]struct{})
	pairs := make([]search.Pair, 0)

	for j, sig := range sigs {
		for _, i := range idx.Candidates(sig) {
			if i >= j {
				continue
			}

			key := [2]int{i, j}
			if _, ok := pairSeen[key]; ok {
				continue
			}

			d := simhash.HammingDistance64(sigs[i], sig)
			pairSeen[key] = struct{}{}
			if d <= k {
				pairs = append(pairs, search.Pair{I: i, J: j, Distance: d})
			}
		}

		idx.Add(sig, j)
	}

	sort.Slice(pairs, func(a, b int) bool {
		if pairs[a].I != pairs[b].I {
			return pairs[a].I < pairs[b].I
		}
		return pairs[a].J < pairs[b].J
	})

	return pairs
}

func readLines(path string, max int, stdin io.Reader) ([]string, error) {
	var r io.Reader

	if path == "" {
		r = stdin
	} else {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		r = f
	}

	scanner := bufio.NewScanner(r)

	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	out := make([]string, 0, max)
	for scanner.Scan() {
		out = append(out, scanner.Text())
		if len(out) >= max {
			break
		}
