# Step 1 Design

## Goal

Build a minimal, correct, and reproducible pipeline for near-duplicate detection on system logs using SimHash.

## Pipeline

```text
read lines -> normalize -> tokenize -> simhash64 -> brute-force search -> print matches
```

## Components

### `cmd/simhashlogs`

CLI entrypoint.

Responsibilities:

* read log lines from file or stdin
* cap input size with `-max`
* run the full pipeline
* print near-duplicate pairs within threshold `-k`

### `internal/normalize`

Replaces high-variance fields with placeholders:

* timestamps -> `<TS>`
* IPv4 -> `<IP>`
* UUIDs -> `<UUID>`
* hex values -> `<HEX>`
* long numbers -> `<NUM>`

Also lowercases and compresses whitespace.

### `internal/tokenize`

Splits normalized lines into tokens while preserving placeholders like `<ip>` and `<num>`.

### `internal/simhash`

Computes a 64-bit SimHash signature from tokens.
Also provides Hamming distance for pair comparison.

### `internal/search`

Implements the Step 1 baseline:

* compare all pairs
* keep those with distance <= k

This is intentionally `O(N^2)` to make correctness easy to validate before introducing indexing.

## Why brute force first

Step 1 is a correctness baseline.
Before building LSH buckets or streaming ingestion, we want:

* deterministic behavior
* simple tests
* transparent matching logic
* easy comparison for future indexed implementations

## Known limitations

* brute-force comparison does not scale
* normalization is intentionally simple
* no persistence
* no streaming ingestion
* no metrics yet

## Expected next step

Step 2 will introduce bucket-based candidate generation so we can avoid comparing every pair.

## `Makefile`

```make
.PHONY: test run-sample run-auth fmt

test:
	go test ./...

run-sample:
	go run ./cmd/simhashlogs -input examples/sample.log -k 6 -max 2000 -print-raw

run-auth:
	go run ./cmd/simhashlogs -input examples/auth_failures.log -k 6 -max 2000 -print-raw

fmt:
	go fmt ./...
```
