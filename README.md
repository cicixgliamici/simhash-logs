# simhash-logs - Near-Duplicate Detection for System Logs

This repository is an engineering-oriented implementation of SimHash for noisy
system logs. It is based on the core idea from:

**Moses Charikar (2002)** - *Similarity Estimation Techniques from Rounding Algorithms*

The project also uses the near-duplicate detection setting from:

**Manku, Jain, and Das Sarma (2007)** - *Detecting Near-Duplicates for Web Crawling*

The current goal is not to be a full production log platform yet. It is a
small, reviewable prototype that turns raw log lines into normalized token
streams, computes 64-bit SimHash fingerprints, and finds near-duplicate pairs
with brute force, an in-memory LSH-style candidate index, or an experimental
paper-style sorted permutation index.

## Motivation

System logs often contain high-variance fields such as timestamps, IPs, ports,
request IDs, process IDs, UUIDs, and hex values. Exact string matching treats
many repeated events as unrelated because those fields change from line to line.

This repository reduces that incidental variance before fingerprinting, so
similar operational events remain close in Hamming space. The main use cases are:

- authentication failure storms and password spraying
- repeated application or infrastructure errors during incidents
- recurring kernel/network messages across hosts
- noisy alert deduplication and log clustering

## Project Status

The project is currently at **Step 1 complete** and **Step 2 complete as a
reviewable candidate-generation prototype**.

Implemented:

- End-to-end `dedup` pipeline: read logs, normalize, tokenize, fingerprint, match.
- 64-bit SimHash with token-frequency weighting.
- Brute-force `O(N^2)` matching as the correctness baseline.
- In-memory LSH-style band index behind `dedup -use-lsh` or `dedup -index lsh`.
- Experimental sorted permutation-table index behind `dedup -index paper`.
- `eval` command comparing candidate indexes against brute-force ground truth.
- Evaluation sweeps over multiple `k`, index, and parameter values with CSV output.
- JSON and text output, deterministic match ordering, optional `-limit`.
- Reproducible synthetic benchmark log sample.
- Unit and integration tests for the CLI and core packages.

Not implemented yet:

- Persistent index storage.
- Streaming ingestion.
- Configurable normalization rules.
- Larger external benchmark datasets and plots.
- Metrics export or production observability integrations.

## Roadmap

### Step 1 - Minimal Correct Implementation

Build a transparent correctness baseline:

- file/stdin ingestion
- normalization of common noisy fields into placeholders
- tokenization with placeholders preserved
- SimHash64 fingerprinting
- Hamming distance comparison
- brute-force near-duplicate search
- tests and small example datasets

Status: **complete**.

### Step 2 - Efficient Candidate Generation

Reduce the number of exact Hamming comparisons:

- LSH-style bucket candidate generation
- exact verification after candidate retrieval
- recall and comparison-count evaluation against brute force
- benchmark-friendly CSV output
- larger datasets and parameter sweeps
- experimental sorted fingerprint-table/permutation index inspired by Manku et al.

Status: **complete for this prototype**. `BandIndex` remains the practical LSH
baseline, while `PermutationIndex` provides a small paper-style sorted-table
implementation for comparison. Both are still in-memory and verified against
brute force.

### Step 3 - Production-Oriented Prototype

Make the system usable in realistic log pipelines:

- incremental ingestion
- persistent fingerprint and bucket storage
- time-windowed grouping
- metrics export
- observability and security demos

Status: **not started**.

## Repository Structure

- `cmd/simhashlogs/` - CLI commands: `dedup` and `eval`
- `internal/normalize/` - log normalization rules
- `internal/tokenize/` - tokenization utilities
- `internal/simhash/` - SimHash and Hamming distance
- `internal/search/` - brute-force search, LSH-style indexing, and sorted permutation indexing
- `examples/` - small sample log files
- `docs/` - design notes, paper mapping, walkthroughs, and expected outputs

## Quick Start

Run tests:

```bash
go test ./...
```

Run near-duplicate detection on stdin:

```bash
cat <<'EOF_LOG' | go run ./cmd/simhashlogs dedup -k 6 -max 2000 -json
2026-02-21T10:01:02Z sshd[12345]: Failed password for invalid user admin from 192.168.1.20 port 55221 ssh2
2026-02-21T10:01:05Z sshd[12346]: Failed password for invalid user admin from 192.168.1.21 port 55222 ssh2
2026-02-21T10:02:10Z kernel: eth0 link up at 1000Mbps
EOF_LOG
```

Run the included example:

```bash
go run ./cmd/simhashlogs dedup -input examples/auth_failures.log -k 6 -max 2000 -json
```

Try LSH-style candidate generation:

```bash
go run ./cmd/simhashlogs dedup -input examples/auth_failures.log -k 6 -max 2000 -use-lsh -json
```

Try the paper-style sorted permutation index:

```bash
go run ./cmd/simhashlogs dedup -input examples/auth_failures.log -k 6 -max 2000 -index paper -json
```

Evaluate candidate indexes against brute-force ground truth:

```bash
go run ./cmd/simhashlogs eval -input examples/auth_failures.log -k 6 -max 2000
```

Run a small parameter sweep as CSV:

```bash
go run ./cmd/simhashlogs eval -input examples/auth_failures.log -k-values 3,6,9 -bands-values 0,5,8 -csv
```

Run the synthetic benchmark sample:

```bash
go run ./cmd/simhashlogs eval -input examples/synthetic_benchmark.log -k-values 3,6 -bands-values 0,8 -csv
```
