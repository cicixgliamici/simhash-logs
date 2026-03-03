package normalize

import (
	"regexp"
	"strings"
)

// NOTE: This is intentionally simple for Step 1.
// In later steps, we’ll make normalization configurable (YAML/JSON rules),
// and we’ll add domain-specific patterns (paths, emails, MACs, ports, etc.).

var (
	reISOTime  = regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})?\b`)
	reSyslogTS = regexp.MustCompile(`\b(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s+\d{1,2}\s+\d{2}:\d{2}:\d{2}\b`)
	reIPv4     = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	reUUID     = regexp.MustCompile(`\b[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}\b`)
	reHex      = regexp.MustCompile(`\b0x[0-9a-fA-F]+\b`)
	reLongNum  = regexp.MustCompile(`\b\d{4,}\b`) // 4+ digits as "often an ID"
	reMultiSpc = regexp.MustCompile(`\s+`)
)

func Line(s string) string {
	s = strings.TrimSpace(s)

	// Order matters: replace specific patterns first.
	s = reISOTime.ReplaceAllString(s, "<TS>")
	s = reSyslogTS.ReplaceAllString(s, "<TS>")
	s = reIPv4.ReplaceAllString(s, "<IP>")
	s = reUUID.ReplaceAllString(s, "<UUID>")
	s = reHex.ReplaceAllString(s, "<HEX>")
	s = reLongNum.ReplaceAllString(s, "<NUM>")

	// Normalize whitespace + lowercase
	s = strings.ToLower(s)
	s = reMultiSpc.ReplaceAllString(s, " ")
	return s
}