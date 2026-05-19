package search

import "sort"

func sortPairs(pairs []Pair) {
	sort.Slice(pairs, func(a, b int) bool {
		if pairs[a].Distance != pairs[b].Distance {
			return pairs[a].Distance < pairs[b].Distance
		}
		if pairs[a].I != pairs[b].I {
			return pairs[a].I < pairs[b].I
		}
		return pairs[a].J < pairs[b].J
	})
}
