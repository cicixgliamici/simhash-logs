package search

import (
	"reflect"
	"testing"
)

func TestBandIndexCandidates_DeduplicatesAcrossBands(t *testing.T) {
	bi := NewBandIndex(4)
	sig := uint64(0xABCD00000000ABCD)

	bi.Add(sig, 3)
	bi.Add(sig, 3)
	bi.Add(sig, 7)

	got := bi.Candidates(sig)
	want := []int{3, 7}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected candidates: got=%v want=%v", got, want)
	}
}

func TestBandKey_LastBandUsesRemainderBits(t *testing.T) {
	bi := NewBandIndex(6) // 10 bits per band + 14-bit remainder in last band.
	a := uint64(0)
	b := uint64(1) << 63

	keyA := bi.bandKey(a, bi.Bands-1)
	keyB := bi.bandKey(b, bi.Bands-1)
	if keyA == keyB {
		t.Fatalf("expected different last-band keys, got keyA=%d keyB=%d", keyA, keyB)
	}
}
