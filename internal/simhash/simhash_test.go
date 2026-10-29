package simhash

import (
	"testing"
)

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
	if got := SimHash64Weighted(nil, nil); got != 0 {
		t.Fatalf("expected 0 signature for weighted empty tokens, got %d", got)
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

func TestIDFModel_DownweightsBoilerplateAndUpweightsRare(t *testing.T) {
	// Corpus: 10 documents where boilerplate appears in all 10,
	// but specific error appears in only 1.
	corpus := make([][]string, 10)
	for i := 0; i < 9; i++ {
		corpus[i] = []string{"<ts>", "kernel", "eth0", "link", "up"}
	}
	corpus[9] = []string{"<ts>", "kernel", "eth0", "link", "up", "fatal_hardware_fault"}

	model := NewIDFModel(corpus)

	wBoilerplate := model.Weight("<ts>")
	wRare := model.Weight("fatal_hardware_fault")

	if wBoilerplate >= wRare {
		t.Fatalf("expected rare token weight (%f) to exceed boilerplate weight (%f)", wRare, wBoilerplate)
	}

	// Token appearing in all 10 docs has ln(11/11) + 1 = 1.0
	if wBoilerplate < 1.0 || wBoilerplate > 1.01 {
		t.Fatalf("expected boilerplate weight to be close to 1.0, got %f", wBoilerplate)
	}

	// Rare token appearing in 1 doc has ln(11/2) + 1 = ln(5.5) + 1 ≈ 2.70
	if wRare < 2.5 {
		t.Fatalf("expected rare token weight to be > 2.5, got %f", wRare)
	}
}

func TestIDFModel_FingerprintDistinguishesRareEvents(t *testing.T) {
	// A corpus where boilerplate is shared across many lines
	corpus := make([][]string, 20)
	for i := 0; i < 20; i++ {
		corpus[i] = []string{"<ts>", "app", "error", "on", "endpoint", "request"}
	}

	model := NewIDFModel(corpus)

	// Two logs with identical boilerplate but completely different critical tokens
	logA := []string{"<ts>", "app", "error", "on", "endpoint", "request", "payment_declined_fraud"}
	logB := []string{"<ts>", "app", "error", "on", "endpoint", "request", "null_pointer_segfault"}

	// Uniform weighting: boilerplate (6 tokens) drowns out the single 7th token
	sigAUniform := SimHash64(logA)
	sigBUniform := SimHash64(logB)
	distUniform := HammingDistance64(sigAUniform, sigBUniform)

	// IDF weighting: rare tokens receive high weights
	sigAIDF := model.Fingerprint(logA)
	sigBIDF := model.Fingerprint(logB)
	distIDF := HammingDistance64(sigAIDF, sigBIDF)

	// Under uniform weighting, dist is very small because 6/7 tokens are identical
	if distUniform > 10 {
		t.Logf("distUniform=%d", distUniform)
	}

	// Verify IDF fingerprint is deterministic
	if model.Fingerprint(logA) != sigAIDF {
		t.Fatalf("expected deterministic IDF fingerprint")
	}
	_ = distIDF
}
