package search

import "testing"

func TestBruteNearDuplicates_Empty(t *testing.T) {
	got := BruteNearDuplicates(nil, 3)
	if len(got) != 0 {
		t.Fatalf("expected no matches for empty input, got=%v", got)
	}
}

func TestBruteNearDuplicates_SingleSignature(t *testing.T) {
	got := BruteNearDuplicates([]uint64{0b1010}, 0)
	if len(got) != 0 {
		t.Fatalf("expected no matches for single signature, got=%v", got)
	}
}

func TestBruteNearDuplicates_ExactMatchAtKZero(t *testing.T) {
	sigs := []uint64{0b101010, 0b101010}
	got := BruteNearDuplicates(sigs, 0)

	if len(got) != 1 {
		t.Fatalf("expected 1 match, got=%d", len(got))
	}
	if got[0].I != 0 || got[0].J != 1 || got[0].Distance != 0 {
		t.Fatalf("unexpected pair: %+v", got[0])
	}
}

func TestBruteNearDuplicates_WithinThreshold(t *testing.T) {
	// Distance between 0b1010 and 0b1000 is 1.
	sigs := []uint64{0b1010, 0b1000, 0b0000}
	got := BruteNearDuplicates(sigs, 1)

	if len(got) != 1 {
		t.Fatalf("expected 1 match, got=%d (%v)", len(got), got)
	}
	if got[0].I != 0 || got[0].J != 1 || got[0].Distance != 1 {
		t.Fatalf("unexpected pair: %+v", got[0])
	}
}

func TestBruteNearDuplicates_BelowThresholdExcluded(t *testing.T) {
	// Distance between 0b1010 and 0b1000 is 1, so k=0 should exclude it.
	sigs := []uint64{0b1010, 0b1000}
	got := BruteNearDuplicates(sigs, 0)

	if len(got) != 0 {
		t.Fatalf("expected no matches, got=%v", got)
	}
}
