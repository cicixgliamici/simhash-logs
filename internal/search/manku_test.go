package search

import (
	"math/bits"
	"math/rand"
	"reflect"
	"testing"
)

func TestPartition64_ValidPartitions(t *testing.T) {
	testCases := []int{1, 2, 3, 4, 5, 7, 8, 16, 32, 64}

	for _, p := range testCases {
		blocks := Partition64(p)
		if len(blocks) != p {
			t.Fatalf("expected %d blocks, got %d", p, len(blocks))
		}

		expectedOffset := 0
		totalBits := 0
		for i, b := range blocks {
			if b.Offset != expectedOffset {
				t.Fatalf("p=%d block=%d: expected offset %d, got %d", p, i, expectedOffset, b.Offset)
			}
			if b.Length <= 0 {
				t.Fatalf("p=%d block=%d: length must be > 0, got %d", p, i, b.Length)
			}
			expectedOffset += b.Length
			totalBits += b.Length
		}

		if totalBits != 64 {
			t.Fatalf("p=%d: expected sum of lengths 64, got %d", p, totalBits)
		}
	}
}

func TestPermute64_PreservesHammingWeightAndPositionsBlockAtMSB(t *testing.T) {
	blocks := Partition64(4) // 4 blocks of 16 bits each

	for bIdx, b := range blocks {
		// Test various 64-bit patterns
		patterns := []uint64{
			0x0000000000000000,
			0xFFFFFFFFFFFFFFFF,
			0x123456789ABCDEF0,
			0xAAAAAAAAAAAAAAAA,
			0x5555555555555555,
		}

		for _, sig := range patterns {
			permuted := Permute64(sig, b)

			// 1. Popcount must be preserved (bijection property)
			if bits.OnesCount64(sig) != bits.OnesCount64(permuted) {
				t.Fatalf("block %d: Permute64 changed Hamming weight for %#x: original=%d permuted=%d",
					bIdx, sig, bits.OnesCount64(sig), bits.OnesCount64(permuted))
			}

			// 2. The block bits in original signature must equal the top b.Length bits of permuted key
			expectedBlockBits := (sig >> b.Offset) & ((uint64(1) << b.Length) - 1)
			gotMSBBits := permuted >> (64 - b.Length)

			if expectedBlockBits != gotMSBBits {
				t.Fatalf("block %d: MSB mismatch for %#x: expected %#x, got %#x",
					bIdx, sig, expectedBlockBits, gotMSBBits)
			}
		}
	}
}

func TestMankuNearDuplicates_100PercentRecallAgainstBrute(t *testing.T) {
	// Generate a reproducible set of pseudo-random signatures with known clusters
	rng := rand.New(rand.NewSource(42))
	baseSigs := []uint64{
		0x1111222233334444,
		0xAAAA5555AAAA5555,
		0x0F0F0F0F0F0F0F0F,
		0xF0F0F0F0F0F0F0F0,
	}

	sigs := make([]uint64, 0, 100)
	for _, base := range baseSigs {
		sigs = append(sigs, base)
		// Add mutations with Hamming distance 1, 2, 3
		for i := 0; i < 15; i++ {
			mut := base
			// Flip 1 to 3 random bits
			numFlips := 1 + rng.Intn(3)
			for f := 0; f < numFlips; f++ {
				bit := rng.Intn(64)
				mut ^= (uint64(1) << bit)
			}
			sigs = append(sigs, mut)
		}
	}

	for _, k := range []int{1, 2, 3} {
		want := BruteNearDuplicates(sigs, k)
		// With default tables = k + 1, recall MUST be 100%
		got, comparisons := MankuNearDuplicates(sigs, k, k+1)

		if len(got) != len(want) {
			t.Fatalf("k=%d: match count mismatch: got=%d, want=%d", k, len(got), len(want))
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("k=%d: pairs do not match brute force ground truth", k)
		}

		totalPairs := len(sigs) * (len(sigs) - 1) / 2
		if comparisons >= totalPairs {
			t.Fatalf("k=%d: expected comparison reduction: got %d comps out of %d total pairs",
				k, comparisons, totalPairs)
		}
	}
}

func TestMankuNearDuplicates_EdgeCases(t *testing.T) {
	// 1. Empty slice
	got, comps := MankuNearDuplicates(nil, 3, 4)
	if got != nil || comps != 0 {
		t.Fatalf("expected nil and 0 comps for empty input")
	}

	// 2. Single item
	got, comps = MankuNearDuplicates([]uint64{42}, 3, 4)
	if len(got) != 0 || comps != 0 {
		t.Fatalf("expected 0 pairs for single item")
	}

	// 3. Exact duplicates (k=0)
	exactSigs := []uint64{100, 200, 100, 300, 100}
	got, _ = MankuNearDuplicates(exactSigs, 0, 1)
	want := BruteNearDuplicates(exactSigs, 0)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("k=0 exact duplicates mismatch: got=%v want=%v", got, want)
	}
}

