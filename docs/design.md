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
- `eval` compares candidate-generation indexes against brute-force results.

## Components

### `cmd/simhashlogs`

The CLI is intentionally thin. It parses flags, reads log lines from stdin or a
file, builds records, and chooses the search mode.

Important flags:

- `-input` reads from a file instead of stdin.
- `-k` sets the maximum Hamming distance.
- `-max` caps the number of input lines.
- `-json` switches output to structured JSON.
- `-use-lsh` enables LSH candidate generation before exact verification.
- `-index` selects `brute`, `lsh`, `paper` (or `paper-v1`), or `manku` search for `dedup`.
- `-bands` controls the number of LSH bands; `0` means auto.
- `-tables` tunes the number of tables for `manku` or `paper`.
- `-window` tunes the scan window for the historical v1 `paper` index.
- `eval -k-values`, `eval -index-values`, and `eval -bands-values` run parameter sweeps.

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

Provides SimHash fingerprinting and token weighting:

- `SimHash64`: Historical baseline using integer accumulators with uniform token frequency.
- `SimHash64Weighted`: Arbitrary weighted SimHash directly implementing Charikar (2002).
- `IDFModel`: Information-theoretic model learning smoothed IDF weights $\ln\left(\frac{N+1}{\text{DF}+1}\right) + 1.0$ across the corpus. It attenuates boilerplate noise and concentrates bit variance on rare semantic tokens.
- Controlled via the `-weighting uniform|idf` CLI flag.

### `internal/search`

The search package contains:

- `BruteNearDuplicates`, the exact `O(N^2)` baseline.
- `BandIndex`, an in-memory LSH-style candidate index.
- `LSHNearDuplicates`, which retrieves candidates and then verifies exact
  Hamming distance.
- `MankuIndex` and `MankuNearDuplicates`, the paper-faithful implementation of
  Manku et al. (2007) Section 3.1: Pigeonhole Principle block partitioning into $p = k + 1$
  disjoint blocks, sorted permutation tables, binary range search for contiguous prefix matches,
  and guaranteed 100% recall.
- `PermutationIndex` and `PaperNearDuplicates`, the historical v1 heuristic prototype
  preserved as an evolutionary baseline to demonstrate the transition from heuristic window
  scans to combinatorial prefix ranges.

## Why Brute Force Still Matters

The brute-force path is the ground truth for small and medium datasets. It keeps
the project reviewable and makes it possible to measure recall for faster
candidate-generation strategies.

## Current Limitations

- The paper-style index is intentionally small and in-memory; it demonstrates
  sorted permutation tables but is not disk-backed or web-scale.
- Normalization rules are hard-coded.
- Evaluation is reproducible on included samples, but still needs external
  corpora and plots before performance claims are strong.
- There is no persistence, streaming ingestion, metrics export, or production
  deployment story yet.

## Useful Commands

```bash
go test ./...
go run ./cmd/simhashlogs dedup -input examples/sample.log -k 6 -max 2000 -print-raw
go run ./cmd/simhashlogs dedup -input examples/auth_failures.log -k 6 -max 2000 -use-lsh -json
go run ./cmd/simhashlogs dedup -input examples/auth_failures.log -k 6 -max 2000 -index paper -json
go run ./cmd/simhashlogs eval -input examples/auth_failures.log -k 6 -max 2000
go run ./cmd/simhashlogs eval -input examples/auth_failures.log -k-values 3,6,9 -bands-values 0,5,8 -csv
go run ./cmd/simhashlogs eval -input examples/synthetic_benchmark.log -k-values 3,6 -bands-values 0,8 -csv
```
