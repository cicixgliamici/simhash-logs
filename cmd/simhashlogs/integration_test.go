package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildCLI(t *testing.T) string {
	exe := filepath.Join(t.TempDir(), "simhashlogs.exe")
	// Compile the CLI. Test runs in cmd/simhashlogs, so "." is the package.
	cmd := exec.Command("go", "build", "-o", exe, ".")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build CLI: %v\n%s", err, out)
	}
	return exe
}

func TestCLIExampleAuthFailuresRunsAndReturnsStructuredOutput(t *testing.T) {
	exe := buildCLI(t)
	// From cmd/simhashlogs, the examples are in ../../examples
	examplePath := filepath.Join("..", "..", "examples", "auth_failures.log")

	// If run from root via some IDE configs, adjust path:
	if _, err := os.Stat(examplePath); os.IsNotExist(err) {
		examplePath = filepath.Join("examples", "auth_failures.log")
	}

	cmd := exec.Command(exe, "dedup", "-input", examplePath, "-k", "6", "-max", "2000", "-json")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput:\n%s", err, string(out))
	}

	output := string(out)
	if strings.TrimSpace(output) == "" {
		t.Fatal("expected non-empty CLI output")
	}
	if !strings.Contains(output, "distance") {
		t.Fatalf("expected output to contain json match output; got:\n%s", output)
	}
}

func TestCLIExampleAuthFailuresWithLSHRuns(t *testing.T) {
	exe := buildCLI(t)
	examplePath := filepath.Join("..", "..", "examples", "auth_failures.log")
	if _, err := os.Stat(examplePath); os.IsNotExist(err) {
		examplePath = filepath.Join("examples", "auth_failures.log")
	}

	cmd := exec.Command(exe, "dedup", "-input", examplePath, "-k", "6", "-max", "2000", "-use-lsh", "-json")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput:\n%s", err, string(out))
	}

	output := string(out)
	if strings.TrimSpace(output) == "" {
		t.Fatal("expected non-empty CLI output")
	}
}
