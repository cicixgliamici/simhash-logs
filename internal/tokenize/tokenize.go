package tokenize

import (
	"regexp"
	"strings"
)

var reSplit = regexp.MustCompile(`[^a-z0-9<>]+`)

// Simple splits on non-alphanumeric chars, keeps placeholders like <ip>.
func Simple(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := reSplit.Split(s, -1)

	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}