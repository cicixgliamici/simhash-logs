package simhash

import "testing"

func TestHammingDistance64(t *testing.T) {
	var a uint64 = 0b1010
	var b uint64 = 0b1000
	if d := HammingDistance64(a, b); d != 1 {
		t.Fatalf("expected 1, got %d", d)
	}
}

func TestSimHash64_Deterministic(t *testing.T) {
	toks := []string{"sshd", "failed", "from", "<ip>"}
	a := SimHash64(toks)
	b := SimHash64(toks)
	if a != b {
		t.Fatalf("expected deterministic hash, got %d vs %d", a, b)
	}
}

func TestSimHash64_Empty(t *testing.T) {
	if got := SimHash64(nil); got != 0 {
		t.Fatalf("expected 0 signature for empty tokens, got %d", got)
	}
}

func TestSimHash64_IdenticalTokensDistanceZero(t *testing.T) {
	a := []string{"sshd", "failed", "password", "user", "admin", "from", "<ip>"}
	b := []string{"sshd", "failed", "password", "user", "admin", "from", "<ip>"}

	ha := SimHash64(a)
	hb := SimHash64(b)

	if d := HammingDistance64(ha, hb); d != 0 {
		t.Fatalf("expected 0 distance, got %d", d)
	}
}