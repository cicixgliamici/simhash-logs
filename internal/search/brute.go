package search

import "simhash-logs/internal/simhash"

type Pair struct {
	I, J     int
	Distance int
}

// BruteNearDuplicates compares all pairs (O(N^2)).
// Good for Step 1 correctness on small N.
func BruteNearDuplicates(records []Record, k int) []Pair {
	n := len(records)
	out := make([]Pair, 0)

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			d := simhash.HammingDistance64(records[i].Sig, records[j].Sig)
			if d <= k {
				out = append(out, Pair{I: i, J: j, Distance: d})
			}
		}
	}
	return out
}
