package search

import "sort"

// BandIndex stores signatures in multiple band-specific hash buckets.
type BandIndex struct {
	Bands       int
	BitsPerBand int
	Buckets     []map[uint64][]int
}

func NewBandIndex(bands int) *BandIndex {
	if bands <= 0 {
		bands = 1
	}
	if bands > 64 {
		bands = 64
	}

	bitsPerBand := 64 / bands
	if bitsPerBand <= 0 {
		bitsPerBand = 1
	}

	buckets := make([]map[uint64][]int, bands)
	for i := range buckets {
		buckets[i] = make(map[uint64][]int)
	}

	return &BandIndex{
		Bands:       bands,
		BitsPerBand: bitsPerBand,
		Buckets:     buckets,
	}
}

func (bi *BandIndex) Add(sig uint64, idx int) {
	for b := 0; b < bi.Bands; b++ {
		key := bi.bandKey(sig, b)
		bi.Buckets[b][key] = append(bi.Buckets[b][key], idx)
	}
}

func (bi *BandIndex) Candidates(sig uint64) []int {
	uniq := make(map[int]struct{})
	for b := 0; b < bi.Bands; b++ {
		key := bi.bandKey(sig, b)
		for _, idx := range bi.Buckets[b][key] {
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

func (bi *BandIndex) bandKey(sig uint64, band int) uint64 {
	shift := band * bi.BitsPerBand
	if shift >= 64 {
		return 0
	}

	bits := bi.BitsPerBand
	if band == bi.Bands-1 {
		bits = 64 - shift
	}
	if bits >= 64 {
		return sig
	}

	mask := uint64((uint64(1) << bits) - 1)
	return (sig >> shift) & mask
}
