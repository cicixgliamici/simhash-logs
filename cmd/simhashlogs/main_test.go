package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"simhash-logs/internal/normalize"
	"strings"
	"testing"
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

	code := runDedup([]string{
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

	code := runDedup([]string{
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

func TestRun_EvalSubcommand(t *testing.T) {
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

	code := runEval([]string{
		"-input", path,
		"-k", "6",
		"-max", "100",
		"-bands", "5",
	}, strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr=%q", code, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "Evaluation Results:") {
		t.Fatalf("expected Evaluation Results in output, got: %q", out)
	}
	if !strings.Contains(out, "Brute Force Ground Truth:") || !strings.Contains(out, "LSH Approach:") {
		t.Fatalf("expected evaluation details in output, got: %q", out)
	}
}

func TestRun_JSONOutputSortedAndLimited(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.log")

	lines := []string{
		"2026-02-21T10:03:00Z sshd[1001]: Failed password for invalid user admin from 10.0.0.1 port 55001 ssh2",
		"2026-02-21T10:03:01Z sshd[1002]: Failed password for invalid user admin from 10.0.0.2 port 55002 ssh2",
		"2026-02-21T10:03:02Z sshd[1003]: Failed password for invalid user root from 10.0.0.3 port 55003 ssh2",
		"2026-02-21T10:03:03Z kernel: eth0 link up at 1000Mbps",
	}
	content := strings.Join(lines, "\n")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp log: %v", err)
	}

	normIndex := make(map[string]int, len(lines))
	for i, line := range lines {
		normIndex[normalize.Line(line)] = i
	}

	var allStdout bytes.Buffer
	var allStderr bytes.Buffer
	allCode := runDedup([]string{
		"-input", path,
		"-k", "64",
		"-max", "100",
		"-json",
	}, strings.NewReader(""), &allStdout, &allStderr)
	if allCode != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr=%q", allCode, allStderr.String())
	}

	var allMatches []matchOutput
	if err := json.Unmarshal(allStdout.Bytes(), &allMatches); err != nil {
		t.Fatalf("invalid json output: %v\noutput=%s", err, allStdout.String())
	}
	if len(allMatches) < 4 {
		t.Fatalf("expected multiple matches, got %d", len(allMatches))
	}

	for i := 1; i < len(allMatches); i++ {
		prev := allMatches[i-1]
		cur := allMatches[i]

		prevI, ok := normIndex[prev.NormalizedA]
		if !ok {
			t.Fatalf("normalized_a not found in source lines: %q", prev.NormalizedA)
		}
		prevJ, ok := normIndex[prev.NormalizedB]
		if !ok {
			t.Fatalf("normalized_b not found in source lines: %q", prev.NormalizedB)
		}
		curI, ok := normIndex[cur.NormalizedA]
		if !ok {
			t.Fatalf("normalized_a not found in source lines: %q", cur.NormalizedA)
		}
		curJ, ok := normIndex[cur.NormalizedB]
		if !ok {
			t.Fatalf("normalized_b not found in source lines: %q", cur.NormalizedB)
		}

		isOrdered := prev.Distance < cur.Distance ||
			(prev.Distance == cur.Distance && (prevI < curI ||
				(prevI == curI && prevJ <= curJ)))
		if !isOrdered {
			t.Fatalf("matches not sorted at %d: prev=%+v (i=%d,j=%d), cur=%+v (i=%d,j=%d)",
				i, prev, prevI, prevJ, cur, curI, curJ)
		}
	}

	var limitedStdout bytes.Buffer
	var limitedStderr bytes.Buffer
	limit := 3
	limitedCode := runDedup([]string{
		"-input", path,
		"-k", "64",
		"-max", "100",
		"-json",
		"-limit", "3",
	}, strings.NewReader(""), &limitedStdout, &limitedStderr)
	if limitedCode != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr=%q", limitedCode, limitedStderr.String())
	}

	var limitedMatches []matchOutput
	if err := json.Unmarshal(limitedStdout.Bytes(), &limitedMatches); err != nil {
		t.Fatalf("invalid limited json output: %v\noutput=%s", err, limitedStdout.String())
	}
	if len(limitedMatches) != limit {
		t.Fatalf("expected %d limited matches, got %d", limit, len(limitedMatches))
	}
	for i := 0; i < limit; i++ {
		if limitedMatches[i] != allMatches[i] {
			t.Fatalf("limited output differs from sorted prefix at %d: got=%+v want=%+v", i, limitedMatches[i], allMatches[i])
		}
	}
}

func TestRun_LSHWithCustomBands(t *testing.T) {
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
	code := runDedup([]string{
		"-input", path,
		"-k", "6",
		"-max", "100",
		"-json",
		"-use-lsh",
		"-bands", "4",
	}, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr=%q", code, stderr.String())
	}

	if !strings.Contains(stderr.String(), "mode=lsh") || !strings.Contains(stderr.String(), "bands=4") {
		t.Fatalf("expected lsh stats with custom bands, got stderr: %q", stderr.String())
	}

	var matches []matchOutput
	if err := json.Unmarshal(stdout.Bytes(), &matches); err != nil {
		t.Fatalf("invalid json output: %v\noutput=%s", err, stdout.String())
	}
	if len(matches) == 0 {
		t.Fatalf("expected at least one match with lsh, got 0")
	}
}
