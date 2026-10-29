package search

import (
	"sort"

	"simhash-logs/internal/simhash"
)

// [HISTORICAL PROTOTYPE / HEURISTIC v1]
//
// PermutationIndex represents the initial v1 prototype of a sorted table index.
// It explores the sorted-table concept by applying deterministic bit rotations
// and scanning a numerical window around sort.Search insertion points.
//
// Theoretical Note & Known Limitations:
// In integer sort order, numerical proximity does not correspond to Hamming proximity
// (e.g., 0b00001111 and 0b00010000 are adjacent numerically with diff=1, but differ by 5 bits,
// whereas 0 and 2^63 differ by only 1 bit in Hamming distance but are separated by 2^63 integers).
// As a result, scanning an arbitrary [pos - window, pos + window] interval is an empirical heuristic
// that lacks theoretical recall guarantees.
//
// SUPERSEDED BY:
// MankuIndex in manku.go, which implements the true combinatorial algorithm from Section 3.1
// of Manku, Jain, and Das Sarma (2007) based on the Pigeonhole Principle (partitioning 64 bits
// into p = k + 1 blocks, sorting by permuted prefix, and retrieving exact prefix matches via
// binary range search). MankuIndex provides 100% guaranteed recall with no heuristic windows.
//
// This v1 implementation is intentionally preserved as a benchmark baseline and historical artifact
// to showcase the progression from an initial heuristic prototype to a mathematically rigorous design.

type permutationEntry struct {
	// Key is the fingerprint after applying one deterministic table-specific
	// permutation. Sorting by this key makes nearby keys cheap to scan.
	Key uint64
	// Idx points back to the original input slice. We keep it in the table so
	// sorted order never loses the caller's record identity.
	Idx int
}

// PermutationIndex stores multiple sorted views of the same fingerprints.
// [Deprecated / Historical v1]: Use MankuIndex in manku.go for guaranteed recall.
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

// PaperNearDuplicates uses the historical v1 heuristic sorted permutation-table
// index to generate candidates, followed by exact Hamming verification.
// [Historical / v1 Heuristic]: Superseded by MankuNearDuplicates in manku.go.
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
