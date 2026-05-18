package main

import (
	"encoding/csv"
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

func TestCLIEvalSweepCSVEndToEnd(t *testing.T) {
	exe := buildCLI(t)
	examplePath := filepath.Join("..", "..", "examples", "auth_failures.log")
	if _, err := os.Stat(examplePath); os.IsNotExist(err) {
		examplePath = filepath.Join("examples", "auth_failures.log")
	}

	cmd := exec.Command(exe,
		"eval",
		"-input", examplePath,
		"-k-values", "3,6",
		"-bands-values", "0,5",
		"-max", "2000",
		"-csv",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput:\n%s", err, string(out))
	}

	rows, err := csv.NewReader(strings.NewReader(string(out))).ReadAll()
	if err != nil {
		t.Fatalf("invalid CSV output: %v\noutput:\n%s", err, string(out))
	}

	if len(rows) != 5 {
		t.Fatalf("expected header plus 4 rows, got %d rows:\n%s", len(rows), string(out))
	}

	wantHeader := []string{
		"k",
		"bands",
		"records",
		"brute_ms",
		"lsh_ms",
		"brute_comps",
		"lsh_comps",
		"total_actual",
		"true_positives",
		"recall_pct",
		"comp_reduction_pct",
	}
	if strings.Join(rows[0], ",") != strings.Join(wantHeader, ",") {
		t.Fatalf("unexpected header: got=%v want=%v", rows[0], wantHeader)
	}

	wantPairs := [][2]string{
		{"3", "4"},
		{"3", "5"},
		{"6", "7"},
		{"6", "5"},
	}
	for i, want := range wantPairs {
		row := rows[i+1]
		if row[0] != want[0] || row[1] != want[1] {
			t.Fatalf("unexpected k/bands at row %d: got=(%s,%s) want=(%s,%s)",
				i+1, row[0], row[1], want[0], want[1])
		}
		if row[2] != "12" {
			t.Fatalf("expected 12 records at row %d, got %s", i+1, row[2])
		}
	}
}
