package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"simhash-logs/internal/normalize"
	"simhash-logs/internal/search"
	"simhash-logs/internal/simhash"
	"simhash-logs/internal/tokenize"
)

const (
	weightingUniform = "uniform"
	weightingIDF     = "idf"
	weightingTFIDF   = "tfidf"
)

func validateWeighting(w string) error {
	switch w {
	case weightingUniform, weightingIDF, weightingTFIDF, "":
		return nil
	default:
		return fmt.Errorf("invalid weighting %q: must be uniform, idf, or tfidf", w)
	}
}

// buildRecords computes signatures using the historical uniform weighting baseline.
func buildRecords(lines []string) []search.Record {
	return buildRecordsWithWeighting(lines, weightingUniform)
}

// buildRecordsWithWeighting computes signatures with the selected token weighting scheme.
// When idf or tfidf is chosen, an IDFModel is learned over the tokenized lines,
// significantly downweighting ubiquitous boilerplate tokens in the fingerprints.
func buildRecordsWithWeighting(lines []string, weighting string) []search.Record {
	records := make([]search.Record, len(lines))
	tokenList := make([][]string, len(lines))

	for i, line := range lines {
		normalized := normalize.Line(line)
		tokens := tokenize.Simple(normalized)
		tokenList[i] = tokens
		records[i] = search.Record{
			Raw:        line,
			Normalized: normalized,
			Tokens:     tokens,
		}
	}

	if weighting == weightingIDF || weighting == weightingTFIDF {
		idfModel := simhash.NewIDFModel(tokenList)
		for i := range records {
			records[i].Sig = idfModel.Fingerprint(records[i].Tokens)
		}
	} else {
		for i := range records {
			records[i].Sig = simhash.SimHash64(records[i].Tokens)
		}
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
