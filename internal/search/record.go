package search

// Record is the canonical representation of an input line after preprocessing.
type Record struct {
	Raw        string
	Normalized string
	Tokens     []string
	Sig        uint64
}
