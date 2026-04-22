package simhash

import (
	"encoding/binary"
	"hash/fnv"
	"math/bits"
)

// SimHash64 computes a 64-bit SimHash signature from tokens.
// Weighting: token frequency (each token occurrence contributes +1 or -1).
func SimHash64(tokens []string) uint64 {
	if len(tokens) == 0 {
		return 0
	}

	// 64-d accumulator
	var acc [64]int

	for _, tok := range tokens {
		h := hash64(tok)

		// For each bit: add +1 if bit is 1 else -1
		for i := 0; i < 64; i++ {
			if (h>>uint(i))&1 == 1 {
				acc[i] += 1
			} else {
				acc[i] -= 1
			}
		}
	}

	// Build final signature: bit i is 1 if acc[i] > 0
	var sig uint64
	for i := 0; i < 64; i++ {
		if acc[i] > 0 {
			sig |= (1 << uint(i))
		}
	}
	return sig
}

// Stable 64-bit hash for a token.
// We use FNV-1a 64-bit (fast, stable, built-in).
func hash64(s string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	sum := h.Sum(nil)
	return binary.LittleEndian.Uint64(sum)
}

// HammingDistance64 counts differing bits between two signatures.
func HammingDistance64(a, b uint64) int {
	return bits.OnesCount64(a ^ b)
}