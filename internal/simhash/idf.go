package simhash

import (
	"math"
)

// IDFModel estimates Inverse Document Frequency (IDF) weights from a corpus of tokenized logs.
//
// In system logs, ubiquitous boilerplate tokens (such as timestamps, IP placeholders, or common
// daemon names) carry very little discriminative entropy, while rare tokens (such as specific
// error codes, exception classes, or query paths) carry high semantic information.
//
// By weighting each token by its IDF in Charikar's SimHash accumulation vector,
// rare tokens gain higher voting influence on the 64-bit signature, significantly reducing
// false positive collisions between unrelated logs that share identical boilerplate structure.
type IDFModel struct {
	DocCount int
	DocFreq  map[string]int
	Weights  map[string]float64
}

// NewIDFModel builds and trains an IDFModel from a corpus of tokenized documents.
func NewIDFModel(corpus [][]string) *IDFModel {
	m := &IDFModel{
		DocCount: len(corpus),
		DocFreq:  make(map[string]int),
		Weights:  make(map[string]float64),
	}

	for _, doc := range corpus {
		seen := make(map[string]struct{}, len(doc))
		for _, tok := range doc {
			if _, ok := seen[tok]; !ok {
				seen[tok] = struct{}{}
				m.DocFreq[tok]++
			}
		}
	}

	// Compute smoothed IDF: ln((N + 1) / (DF + 1)) + 1.0
	// This ensures:
	// 1. Minimum weight is >= 1.0 (for tokens appearing in all documents).
	// 2. Weights scale logarithmically with rarity.
	// 3. Division by zero is avoided.
	nPlus1 := float64(m.DocCount + 1)
	for tok, df := range m.DocFreq {
		idf := math.Log(nPlus1/float64(df+1)) + 1.0
		m.Weights[tok] = idf
	}

	return m
}

// Weight returns the learned IDF weight for the given token.
// If the token was not seen during training, it is assigned the maximum IDF weight:
// ln(DocCount + 1) + 1.0 (treating unseen tokens as maximally rare).
func (m *IDFModel) Weight(token string) float64 {
	if w, ok := m.Weights[token]; ok {
		return w
	}
	// Fallback for unseen token: assume DF = 0
	return math.Log(float64(m.DocCount+1)) + 1.0
}

// Fingerprint computes the TF-IDF weighted SimHash64 signature for a token list.
// Each token occurrence adds or subtracts its IDF weight from the 64-dimensional accumulator,
// which naturally combines Term Frequency (TF) with Inverse Document Frequency (IDF).
func (m *IDFModel) Fingerprint(tokens []string) uint64 {
	if len(tokens) == 0 {
		return 0
	}

	var acc [64]float64
	for _, tok := range tokens {
		w := m.Weight(tok)
		h := hash64(tok)
		for i := 0; i < 64; i++ {
			if (h>>uint(i))&1 == 1 {
				acc[i] += w
			} else {
				acc[i] -= w
			}
		}
	}

	var sig uint64
	for i := 0; i < 64; i++ {
		if acc[i] > 0 {
			sig |= (1 << uint(i))
		}
	}
	return sig
}

