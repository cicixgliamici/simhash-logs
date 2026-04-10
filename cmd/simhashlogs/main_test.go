package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"simhash-logs/internal/search"
)

func TestRun_EndToEndTextOutput(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.log")

	content := strings.Join([]string{
		"2026-02-21T10:01:02Z sshd[12345]: Failed password for invalid user admin from 192.168.1.20 port 55221 ssh2",
		"2026-02-21T10:01:05Z sshd[12346]: Failed password for invalid user admin from 192.168.1.21 port 55222 ssh2",
		"2026-02-21T10:02:10Z kernel: eth0 link up at 1000Mbps",
	}, "\n")

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp log: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{
		"-input", path,
		"-k", "6",
		"-max", "100",
		"-print-raw",
	}, strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr=%q", code, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "match (dist=") {
		t.Fatalf("expected at least one match, got output: %q", out)
	}
	if !strings.Contains(out, "A(raw):") || !strings.Contains(out, "B(raw):") {
		t.Fatalf("expected raw lines in output, got: %q", out)
	}
}

func TestRun_EndToEndJSONOutput(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.log")

	content := strings.Join([]string{
		"2026-02-21T10:03:00Z nginx[9981]: upstream timed out while connecting to 10.0.0.5",
		"2026-02-21T10:03:01Z nginx[9982]: upstream timed out while connecting to 10.0.0.6",
		"2026-02-21T10:05:00Z unrelated message",
	}, "\n")

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp log: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{
		"-input", path,
		"-k", "6",
		"-max", "100",
		"-json",
		"-print-raw",
	}, strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr=%q", code, stderr.String())
	}

	var matches []matchOutput
	if err := json.Unmarshal(stdout.Bytes(), &matches); err != nil {
		t.Fatalf("invalid json output: %v\noutput=%s", err, stdout.String())
	}

	if len(matches) == 0 {
		t.Fatalf("expected at least one match in JSON output, got 0")
	}

	first := matches[0]
	if first.Distance < 0 {
		t.Fatalf("unexpected negative distance: %+v", first)
	}
	if first.NormalizedA == "" || first.NormalizedB == "" {
		t.Fatalf("expected normalized fields, got: %+v", first)
	}
	if first.RawA == "" || first.RawB == "" {
		t.Fatalf("expected raw fields with -print-raw, got: %+v", first)
	}
}

func TestReadLines_RespectsMax(t *testing.T) {
	input := "a\nb\nc\nd\n"
	lines, err := readLines("", 2, strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}

	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if lines[0] != "a" || lines[1] != "b" {
		t.Fatalf("unexpected lines: %v", lines)
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

	got := lshNearDuplicates(sigs, k, k+1)
	want := search.BruteNearDuplicates(sigs, k)

	if len(got) != len(want) {
		t.Fatalf("pair count mismatch: got=%d want=%d; got=%v; want=%v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("pair mismatch at %d: got=%+v want=%+v", i, got[i], want[i])
		}
	}
}
