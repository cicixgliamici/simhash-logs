package search

import (
	"reflect"
	"testing"
)

func TestNewPermutationIndex_ClampsTableCountAndSortsDeterministically(t *testing.T) {
	sigs := []uint64{3, 1, 1, 2}

	oneTable := NewPermutationIndex(sigs, 0)
	if len(oneTable.Tables) != 1 {
		t.Fatalf("expected non-positive table count to clamp to 1, got %d", len(oneTable.Tables))
	}

	manyTables := NewPermutationIndex(sigs, 128)
	if len(manyTables.Tables) != 64 {
		t.Fatalf("expected table count to clamp to 64, got %d", len(manyTables.Tables))
	}

	got := oneTable.Tables[0]
	want := []permutationEntry{
		{Key: 1, Idx: 1},
		{Key: 1, Idx: 2},
		{Key: 2, Idx: 3},
		{Key: 3, Idx: 0},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected sorted table:\ngot:  %+v\nwant: %+v", got, want)
	}
}

func TestPermutationIndexCandidates_DeduplicatesAcrossTables(t *testing.T) {
	sigs := []uint64{0, 1, 2, 3}
	idx := NewPermutationIndex(sigs, 4)

	got := idx.Candidates(1, 1, 3)
	seen := make(map[int]struct{}, len(got))
	for _, candidate := range got {
		if candidate == 1 {
			t.Fatalf("candidate list included self: %v", got)
		}
		if _, ok := seen[candidate]; ok {
			t.Fatalf("candidate list contains duplicate %d: %v", candidate, got)
		}
		seen[candidate] = struct{}{}
	}
}

func TestPermutationIndexCandidates_ClampsWindowAndReturnsSortedCandidates(t *testing.T) {
	sigs := []uint64{10, 20, 30, 40}
	idx := NewPermutationIndex(sigs, 1)

	got := idx.Candidates(25, -1, 0)
	want := []int{1, 2, 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected candidates for clamped window: got=%v want=%v", got, want)
	}
}

func TestPermutationIndexCandidates_HandlesInsertionPointAtEdges(t *testing.T) {
	sigs := []uint64{10, 20, 30}
	idx := NewPermutationIndex(sigs, 1)

	low := idx.Candidates(1, -1, 2)
	if !reflect.DeepEqual(low, []int{0, 1, 2}) {
		t.Fatalf("unexpected low-edge candidates: %v", low)
	}

	high := idx.Candidates(99, -1, 2)
	if !reflect.DeepEqual(high, []int{1, 2}) {
		t.Fatalf("unexpected high-edge candidates: %v", high)
	}
}

func TestPaperNearDuplicates_VerifiesCandidatesAndSortsPairs(t *testing.T) {
	sigs := []uint64{
		0,
		1, // dist(0,1)=1
		3, // dist(1,3)=1
		7, // dist(3,7)=1
	}

	got, comparisons := PaperNearDuplicates(sigs, 1, 4, 4)
	want := BruteNearDuplicates(sigs, 1)

	if comparisons == 0 {
		t.Fatal("expected exact comparisons")
	}
	if len(got) != len(want) {
		t.Fatalf("pair count mismatch: got=%d want=%d; got=%v want=%v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("pair mismatch at %d: got=%+v want=%+v", i, got[i], want[i])
		}
	}
}

func TestPaperNearDuplicates_FiltersCandidatesOutsideThreshold(t *testing.T) {
	sigs := []uint64{
		0b0000,
		0b0001, // within k=1 of first signature
		0b1111, // candidate with a large Hamming distance
	}

	got, comparisons := PaperNearDuplicates(sigs, 1, 1, 3)
	want := []Pair{{I: 0, J: 1, Distance: 1}}

	if comparisons != 3 {
		t.Fatalf("expected all three unique pairs to be verified once, got %d", comparisons)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected verified pairs: got=%v want=%v", got, want)
	}
}

func TestPaperNearDuplicates_EmptyInput(t *testing.T) {
	got, comparisons := PaperNearDuplicates(nil, 3, 4, 4)
	if got != nil || comparisons != 0 {
		t.Fatalf("expected nil pairs and zero comparisons, got pairs=%v comparisons=%d", got, comparisons)
	}
}

func TestPermutedKey_UsesStableRotations(t *testing.T) {
	sig := uint64(0x8000000000000001)

	if got := permutedKey(sig, 0); got != sig {
		t.Fatalf("table zero should preserve the signature: got=%#x want=%#x", got, sig)
	}
	if got, want := permutedKey(1, 1), uint64(1)<<11; got != want {
		t.Fatalf("unexpected table-one rotation: got=%#x want=%#x", got, want)
	}
	if got, want := permutedKey(sig, 64), sig; got != want {
		t.Fatalf("table 64 should cycle back to the original rotation: got=%#x want=%#x", got, want)
	}
}
