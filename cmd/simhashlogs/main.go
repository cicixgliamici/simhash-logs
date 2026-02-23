package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"simhash-logs/internal/normalize"
	"simhash-logs/internal/search"
	"simhash-logs/internal/simhash"
	"simhash-logs/internal/tokenize"
)

func main() {
	var (
		inputPath = flag.String("input", "", "Path to a log file (default: stdin)")
		k         = flag.Int("k", 3, "Max Hamming distance threshold for near-duplicates")
		maxLines  = flag.Int("max", 5000, "Max number of lines to read (keeps brute-force manageable)")
		printRaw  = flag.Bool("print-raw", false, "Print raw lines alongside normalized lines")
	)
	flag.Parse()

	lines, err := readLines(*inputPath, *maxLines)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read error: %v\n", err)
		os.Exit(1)
	}

	// Normalize + tokenize + simhash
	sigs := make([]uint64, 0, len(lines))
	normed := make([]string, 0, len(lines))

	for _, line := range lines {
		n := normalize.Line(line)
		toks := tokenize.Simple(n)
		s := simhash.SimHash64(toks)

		normed = append(normed, n)
		sigs = append(sigs, s)
	}

	// Brute-force near duplicates
	pairs := search.BruteNearDuplicates(sigs, *k)

	// Print results
	for _, p := range pairs {
		fmt.Printf("match (dist=%d)\n", p.Distance)
		if *printRaw {
			fmt.Printf("  A(raw): %s\n", lines[p.I])
			fmt.Printf("  B(raw): %s\n", lines[p.J])
		}
		fmt.Printf("  A: %s\n", normed[p.I])
		fmt.Printf("  B: %s\n", normed[p.J])
		fmt.Println()
	}
}

func readLines(path string, max int) ([]string, error) {
	var scanner *bufio.Scanner

	if path == "" {
		scanner = bufio.NewScanner(os.Stdin)
	} else {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		scanner = bufio.NewScanner(f)
	}

	// Increase buffer for long log lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	out := make([]string, 0, max)
	for scanner.Scan() {
		out = append(out, scanner.Text())
		if len(out) >= max {
			break
		}
	}
	return out, scanner.Err()
}
