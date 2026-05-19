package search

import (
	"sort"

	"simhash-logs/internal/simhash"
)

type permutationEntry struct {
	// Key is the fingerprint after applying one deterministic table-specific
	// permutation. Sorting by this key makes nearby keys cheap to scan.
	Key uint64
	// Idx points back to the original input slice. We keep it in the table so
	// sorted order never loses the caller's record identity.
	Idx int
}

// PermutationIndex stores multiple sorted views of the same fingerprints.
// It follows the sorted-table idea from Manku et al.: each table applies a
// deterministic bit permutation, sorts fingerprints by that permuted key, and
// candidate lookup scans a small neighborhood around the query key.
type PermutationIndex struct {
	Tables [][]permutationEntry
}

func NewPermutationIndex(sigs []uint64, tables int) *PermutationIndex {
	// A non-positive value is treated as the smallest useful index. This keeps
	// callers simple: CLI "auto" values can resolve to zero before reaching
	// this package without causing an empty, unusable index.
	if tables <= 0 {
		tables = 1
	}
	// There are only 64 bit positions in the fingerprint. More than 64 rotated
	// views would repeat shifts and add memory/CPU cost without new ordering
	// information.
	if tables > 64 {
		tables = 64
	}

	out := &PermutationIndex{
		Tables: make([][]permutationEntry, tables),
	}
	for table := 0; table < tables; table++ {
		entries := make([]permutationEntry, len(sigs))
		for i, sig := range sigs {
			// Each table is a different sorted view over the same signatures.
			// The implementation uses deterministic rotations instead of random
			// permutations so tests, examples, and benchmark CSVs are stable.
			entries[i] = permutationEntry{
				Key: permutedKey(sig, table),
				Idx: i,
			}
		}
		// Tie-break on Idx to make candidate order deterministic when multiple
		// records share the same permuted key.
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].Key != entries[j].Key {
				return entries[i].Key < entries[j].Key
			}
			return entries[i].Idx < entries[j].Idx
		})
		out.Tables[table] = entries
	}
	return out
}

func (pi *PermutationIndex) Candidates(sig uint64, self, window int) []int {
	// The window is the approximation knob. A wider window improves recall by
	// checking more sorted neighbors, while a narrower window reduces exact
	// Hamming comparisons. Clamp to one so every table can contribute at least
	// the immediate insertion-point neighborhood.
	if window <= 0 {
		window = 1
	}

	uniq := make(map[int]struct{})
	for table, entries := range pi.Tables {
		key := permutedKey(sig, table)
		// Find where this query would appear in the table, then scan a bounded
		// neighborhood around that position. This mirrors the sorted-table
		// candidate-generation idea: only records close in at least one permuted
		// ordering become exact-verification candidates.
		pos := sort.Search(len(entries), func(i int) bool {
			return entries[i].Key >= key
		})

		start := pos - window
		if start < 0 {
			start = 0
		}
		end := pos + window + 1
		if end > len(entries) {
			end = len(entries)
		}

		for i := start; i < end; i++ {
			idx := entries[i].Idx
			// PaperNearDuplicates queries signatures already present in the
			// index, so the query's own row must not become its candidate.
			if idx == self {
				continue
			}
			uniq[idx] = struct{}{}
		}
	}

	out := make([]int, 0, len(uniq))
	for idx := range uniq {
		out = append(out, idx)
	}
	sort.Ints(out)
	return out
}

// PaperNearDuplicates uses a sorted permutation-table index to generate
// candidates, followed by exact Hamming verification.
func PaperNearDuplicates(sigs []uint64, k, tables, window int) ([]Pair, int) {
	if len(sigs) == 0 {
		return nil, 0
	}

	idx := NewPermutationIndex(sigs, tables)
	pairSeen := make(map[[2]int]struct{})
	pairs := make([]Pair, 0)
	comparisons := 0

	for j, sig := range sigs {
		for _, i := range idx.Candidates(sig, j, window) {
			// Emit each unordered pair once. The index is built up front, but the
			// search keeps the same pair orientation as BruteNearDuplicates.
			if i >= j {
				continue
			}

			key := [2]int{i, j}
			// The same pair can be found in several permutation tables. Count and
			// verify it once so comparison metrics describe unique exact checks.
			if _, ok := pairSeen[key]; ok {
				continue
			}

			// Candidate generation is approximate; this exact Hamming check is
			// the correctness gate that prevents false positives in the output.
			comparisons++
			d := simhash.HammingDistance64(sigs[i], sig)
			pairSeen[key] = struct{}{}
			if d <= k {
				pairs = append(pairs, Pair{I: i, J: j, Distance: d})
			}
		}
	}

	sortPairs(pairs)
	return pairs, comparisons
}

func permutedKey(sig uint64, table int) uint64 {
	if table == 0 {
		return sig
	}
	// Use a stride that is coprime with 64 so consecutive tables walk through
	// different bit alignments before eventually cycling. This is a compact,
	// deterministic substitute for the richer permutations described in the
	// paper, and it keeps this prototype easy to review.
	shift := uint((table * 11) % 64)
	return (sig << shift) | (sig >> (64 - shift))
}
