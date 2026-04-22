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

func TestLSHNearDuplicates_MatchesBruteForKLessThan64(t *testing.T) {
	sigs := []uint64{
		0,
		1, // dist(0,1)=1
		3, // dist(1,3)=1
		0xFFFF0000FFFF0000,
		0xFFFF0000FFFF0001,
	}
	k := 2

	got, _ := LSHNearDuplicates(sigs, k, k+1)
	want := BruteNearDuplicates(sigs, k)

	if len(got) != len(want) {
		t.Fatalf("pair count mismatch: got=%d want=%d; got=%v; want=%v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("pair mismatch at %d: got=%+v want=%+v", i, got[i], want[i])
		}
	}
}
