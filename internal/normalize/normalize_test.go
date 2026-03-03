package normalize

import (
	"strings"
	"testing"
)

func TestLine_ReplacesPatterns(t *testing.T) {
	in := "2026-02-21T10:01:02Z sshd[12345]: Failed from 192.168.1.20 id=550e8400-e29b-41d4-a716-446655440000 ptr=0xDEADBEEF"
	out := Line(in)

	// Placeholders should appear (case-insensitive check because Line lowercases)
	if !strings.Contains(out, "<ts>") {
		t.Fatalf("expected <TS> placeholder, got: %q", out)
	}
	if !strings.Contains(out, "<ip>") {
		t.Fatalf("expected <IP> placeholder, got: %q", out)
	}
	if !strings.Contains(out, "<uuid>") {
		t.Fatalf("expected <UUID> placeholder, got: %q", out)
	}
	if !strings.Contains(out, "<hex>") {
		t.Fatalf("expected <HEX> placeholder, got: %q", out)
	}

	// Output should be lowercase
	if out != strings.ToLower(out) {
		t.Fatalf("expected lowercase output, got: %q", out)
	}
}

func TestLine_CompressesSpaces(t *testing.T) {
	in := "foo     bar\t\tbaz"
	out := Line(in)
	if out != "foo bar baz" {
		t.Fatalf("expected single-spaced output, got: %q", out)
	}
}