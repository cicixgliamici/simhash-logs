package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

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

	if err := fs.Parse(args); err != nil {
		return 2
	}

	lines, err := readLines(*inputPath, *maxLines, stdin)
	fmt.Fprintf(stderr, "read %d lines\n", len(lines))
	if err != nil {
		fmt.Fprintf(stderr, "read error: %v\n", err)
		return 1
	}

	sigs := make([]uint64, 0, len(lines))
	normed := make([]string, 0, len(lines))

	for _, line := range lines {
		n := normalize.Line(line)
		toks := tokenize.Simple(n)
		s := simhash.SimHash64(toks)

		normed = append(normed, n)
		sigs = append(sigs, s)
	}

	pairs := search.BruteNearDuplicates(sigs, *k)

	if *jsonOut {
		out := make([]matchOutput, 0, len(pairs))
		for _, p := range pairs {
			item := matchOutput{
				Distance:    p.Distance,
				NormalizedA: normed[p.I],
				NormalizedB: normed[p.J],
			}
			if *printRaw {
				item.RawA = lines[p.I]
				item.RawB = lines[p.J]
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
			fmt.Fprintf(stdout, "  A(raw): %s\n", lines[p.I])
			fmt.Fprintf(stdout, "  B(raw): %s\n", lines[p.J])
		}
		fmt.Fprintf(stdout, "  A: %s\n", normed[p.I])
		fmt.Fprintf(stdout, "  B: %s\n", normed[p.J])
		fmt.Fprintln(stdout)
	}

	return 0
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
	}
	return out, scanner.Err()
}
