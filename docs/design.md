# Design Notes

## Goal

Build a minimal, correct, and reproducible near-duplicate detector for system
logs using SimHash. The design favors clarity first: every optimized path should
be comparable against a simple brute-force baseline.

## Current Pipeline

```text
read lines -> normalize -> tokenize -> simhash64 -> search -> print matches
```

The CLI exposes two commands:

- `dedup` runs the matching pipeline and prints near-duplicate pairs.
- `eval` compares LSH-style candidate generation against brute-force results.

## Components

### `cmd/simhashlogs`

The CLI is intentionally thin. It parses flags, reads log lines from stdin or a
file, builds records, and chooses the search mode.

Important flags:

- `-input` reads from a file instead of stdin.
- `-k` sets the maximum Hamming distance.
- `-max` caps the number of input lines.
- `-json` switches output to structured JSON.
- `-use-lsh` enables candidate generation before exact verification.
- `-bands` controls the number of LSH bands; `0` means auto.
- `eval -k-values` and `eval -bands-values` run parameter sweeps.

### `internal/normalize`

Normalization replaces common high-variance fields with placeholders:

- timestamps -> `<TS>`
- IPv4 addresses -> `<IP>`
- UUIDs -> `<UUID>`
- hex values -> `<HEX>`
- long numbers -> `<NUM>`

The normalized line is lowercased and whitespace is collapsed.

### `internal/tokenize`

Tokenization splits normalized lines on non-alphanumeric separators while keeping
placeholder tokens such as `<ip>` and `<num>`.

### `internal/simhash`

SimHash64 hashes each token, updates a 64-dimensional accumulator, and emits a
64-bit signature. Repeated tokens contribute repeatedly, so token frequency
affects the final fingerprint.

### `internal/search`

The search package contains:

- `BruteNearDuplicates`, the exact `O(N^2)` baseline.
- `BandIndex`, an in-memory LSH-style candidate index.
- `LSHNearDuplicates`, which retrieves candidates and then verifies exact
  Hamming distance.

## Why Brute Force Still Matters

The brute-force path is the ground truth for small and medium datasets. It keeps
the project reviewable and makes it possible to measure recall for faster
candidate-generation strategies.

## Current Limitations

- The current LSH index is a practical banding prototype, not yet a faithful
  implementation of the sorted fingerprint-table strategy from Manku et al.
- Normalization rules are hard-coded.
- Evaluation is useful but still small; it needs larger datasets and parameter
  sweeps before performance claims are strong.
- There is no persistence, streaming ingestion, metrics export, or production
  deployment story yet.

## Useful Commands

```bash
go test ./...
go run ./cmd/simhashlogs dedup -input examples/sample.log -k 6 -max 2000 -print-raw
go run ./cmd/simhashlogs dedup -input examples/auth_failures.log -k 6 -max 2000 -use-lsh -json
go run ./cmd/simhashlogs eval -input examples/auth_failures.log -k 6 -max 2000
go run ./cmd/simhashlogs eval -input examples/auth_failures.log -k-values 3,6,9 -bands-values 0,5,8 -csv
```
