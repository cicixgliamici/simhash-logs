package search

import (
	"sort"

	"simhash-logs/internal/simhash"
)

// Block represents a contiguous slice of bits in a 64-bit fingerprint.
type Block struct {
	Offset int // 0-based bit offset in [0, 63]
	Length int // length in bits (1 to 64)
}

// Partition64 partitions the 64 bit positions [0, 63] into p disjoint blocks.
// According to Manku et al. (2007), setting p = k + 1 guarantees by the Pigeonhole
// Principle that if two signatures differ by at most k bits, at least one block
// contains zero bit differences (is 100% identical).
func Partition64(p int) []Block {
	if p <= 0 {
		p = 1
	}
	if p > 64 {
		p = 64
	}

	blocks := make([]Block, p)
	baseLen := 64 / p
	remainder := 64 % p

	currentOffset := 0
	for i := 0; i < p; i++ {
		length := baseLen
		if i < remainder {
			length++
		}
		blocks[i] = Block{
			Offset: currentOffset,
			Length: length,
		}
		currentOffset += length
	}
	return blocks
}

// Permute64 applies a bijective bit permutation that moves the bits of the chosen
// block to the most significant bit positions [64 - b.Length, 63].
// The remaining 64 - b.Length bits are packed contiguously in the lower positions.
func Permute64(sig uint64, b Block) uint64 {
	if b.Length >= 64 {
		return sig
	}

	// 1. Extract block bits
	var blockMask uint64
	if b.Length == 64 {
		blockMask = ^uint64(0)
	} else {
		blockMask = (uint64(1) << b.Length) - 1
	}
	blockBits := (sig >> b.Offset) & blockMask

	// 2. Extract lower bits (bits below b.Offset)
	var lowBits uint64
	if b.Offset > 0 {
		lowBits = sig & ((uint64(1) << b.Offset) - 1)
	}

	// 3. Extract higher bits (bits above b.Offset + b.Length)
	highShift := b.Offset + b.Length
	var highBits uint64
	if highShift < 64 {
		highBits = sig >> highShift
	}

	// 4. Pack higher bits above lower bits
	remaining := (highBits << b.Offset) | lowBits

	// 5. Place block bits in the most significant position
	return (blockBits << (64 - b.Length)) | remaining
}

// MankuEntry stores a permuted 64-bit key and the original record index.
type MankuEntry struct {
	Key uint64
	Idx int
}

// MankuTable represents one sorted table for a specific bit block permutation.
type MankuTable struct {
	Block   Block
	Entries []MankuEntry
}

// MankuIndex implements the sorted permutation table candidate-generation
// algorithm described in Section 3.1 of Manku, Jain, and Das Sarma (2007):
// "Detecting Near-Duplicates for Web Crawling".
//
// Theoretical Guarantee:
// By partitioning 64 bits into p = k + 1 blocks, any two fingerprints differing
// by at most k bits must match identically on at least one block. In the sorted
// table corresponding to that block, all matching fingerprints share an exact
// leading prefix and form a contiguous subarray found via binary search.
// Therefore, Recall is mathematically guaranteed to be 100%.
type MankuIndex struct {
	K      int
	P      int
	Blocks []Block
	Tables []MankuTable
}

// NewMankuIndex constructs a MankuIndex over the provided signatures.
// If tables <= 0, it defaults to k + 1 (the minimum count required for 100% recall).
func NewMankuIndex(sigs []uint64, k, tables int) *MankuIndex {
	if k < 0 {
		k = 0
	}
	if tables <= 0 {
		tables = k + 1
	}
	if tables > 64 {
		tables = 64
	}

	blocks := Partition64(tables)
	mankuTables := make([]MankuTable, len(blocks))

	for t, b := range blocks {
		entries := make([]MankuEntry, len(sigs))
		for i, sig := range sigs {
			entries[i] = MankuEntry{
				Key: Permute64(sig, b),
				Idx: i,
			}
		}

		// Sort by permuted key; tie-break on Idx for deterministic behavior.
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].Key != entries[j].Key {
				return entries[i].Key < entries[j].Key
			}
			return entries[i].Idx < entries[j].Idx
		})

		mankuTables[t] = MankuTable{
			Block:   b,
			Entries: entries,
		}
	}

	return &MankuIndex{
		K:      k,
		P:      len(blocks),
		Blocks: blocks,
		Tables: mankuTables,
	}
}

// Candidates returns all record indices that share an exact block prefix with sig
// in at least one table.
func (mi *MankuIndex) Candidates(sig uint64, self int) []int {
	uniq := make(map[int]struct{})

	for _, table := range mi.Tables {
		b := table.Block
		permSig := Permute64(sig, b)
		prefixLen := b.Length

		// Leading prefix in the permuted key
		prefix := permSig >> (64 - prefixLen)

		// Compute the exact range [minKey, maxKey] of keys sharing this prefix
		minKey := prefix << (64 - prefixLen)
		var maxKey uint64
		if prefixLen >= 64 {
			maxKey = minKey
		} else {
			maxKey = minKey | ((uint64(1) << (64 - prefixLen)) - 1)
		}

		entries := table.Entries
		// Binary search for the first entry matching the prefix
		start := sort.Search(len(entries), func(i int) bool {
			return entries[i].Key >= minKey
		})

		// Binary search for the first entry strictly greater than maxKey
		end := sort.Search(len(entries), func(i int) bool {
			return entries[i].Key > maxKey
		})

		for i := start; i < end; i++ {
			idx := entries[i].Idx
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

// MankuNearDuplicates uses the Manku et al. (2007) sorted permutation index
// to generate candidates with guaranteed 100% recall (when tables >= k + 1),
// followed by exact Hamming distance verification.
func MankuNearDuplicates(sigs []uint64, k, tables int) ([]Pair, int) {
	if len(sigs) == 0 {
		return nil, 0
	}

	idx := NewMankuIndex(sigs, k, tables)
	pairSeen := make(map[[2]int]struct{})
	pairs := make([]Pair, 0)
	comparisons := 0

	for j, sig := range sigs {
		for _, i := range idx.Candidates(sig, j) {
			if i >= j {
				continue
			}

			key := [2]int{i, j}
			if _, ok := pairSeen[key]; ok {
				continue
			}

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

