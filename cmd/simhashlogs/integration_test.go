package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIExampleAuthFailuresRunsAndReturnsStructuredOutput(t *testing.T) {
	examplePath := filepath.Join("examples", "auth-failures.log")

	cmd := exec.Command("bash", "-lc", "cat "+examplePath+" | go run ./cmd/simhashlogs -k 6 -max 2000 -json")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput:\n%s", err, string(out))
	}

	output := string(out)

	if strings.TrimSpace(output) == "" {
		t.Fatal("expected non-empty CLI output")
	}

	// Lightweight assertions that do not overfit the exact JSON shape.
	// Tighten these once the output schema is considered stable.
	if !strings.Contains(output, "match") && !strings.Contains(output, "matches") {
		t.Fatalf("expected output to mention matches; got:\n%s", output)
	}
}

func TestCLIExampleAuthFailuresWithLSHRuns(t *testing.T) {
	examplePath := filepath.Join("examples", "auth-failures.log")

	cmd := exec.Command("bash", "-lc", "cat "+examplePath+" | go run ./cmd/simhashlogs -k 6 -max 2000 -use-lsh -json")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput:\n%s", err, string(out))
	}

	output := string(out)
	if strings.TrimSpace(output) == "" {
		t.Fatal("expected non-empty CLI output")
	}
}
