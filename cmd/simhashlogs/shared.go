package main

import (
	"bufio"
	"io"
	"os"

	"simhash-logs/internal/normalize"
	"simhash-logs/internal/search"
	"simhash-logs/internal/simhash"
	"simhash-logs/internal/tokenize"
)

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
