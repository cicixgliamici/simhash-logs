package search

import (
	"sort"

	"simhash-logs/internal/simhash"
)

type Pair struct {
	I, J     int
	Distance int
}

// BruteNearDuplicates compares all pairs (O(N^2)).
// Good for Step 1 correctness on small N.
func BruteNearDuplicates(sigs []uint64, k int) []Pair {
	n := len(sigs)
	out := make([]Pair, 0)

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			d := simhash.HammingDistance64(sigs[i], sigs[j])
			if d <= k {
				out = append(out, Pair{I: i, J: j, Distance: d})
			}
		}
	}

	sort.Slice(out, func(a, b int) bool {
		if out[a].Distance != out[b].Distance {
			return out[a].Distance < out[b].Distance
		}
		if out[a].I != out[b].I {
			return out[a].I < out[b].I
		}
		return out[a].J < out[b].J
	})

	return out
}
