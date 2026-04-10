# simhash-logs - Near-Duplicate Detection for System Logs (SimHash)

This repository implements a practical, engineering-focused reproduction of the core ideas behind SimHash as introduced in:

**Moses Charikar (2002)** - *Similarity Estimation Techniques from Rounding Algorithms* (STOC 2002).  
The paper shows how randomized rounding / random hyperplanes can be used to build compact binary fingerprints that preserve similarity (e.g., cosine similarity), enabling efficient near-duplicate detection via Hamming distance.

The goal of this project is to translate the paper’s key concept into a production-oriented prototype for **system log analytics**, with a focus on **observability** and **cybersecurity** use cases.

---

## Motivation

System logs often contain high-variance fields (timestamps, IPs, ports, request IDs, hex pointers) that make exact matching ineffective.  
By **normalizing** such fields and computing **SimHash fingerprints**, we can detect repeated patterns and near-duplicate events even when some details differ—useful for:

- authentication failure storms (brute force / password spraying),
- repeated error patterns during incidents,
- recurring kernel/network messages across hosts,
- noisy alert deduplication and log clustering.

---

## Project Plan (3 Steps)

### Step 1 — Minimal Correct Implementation (MVP)
**Objective:** implement the full SimHash pipeline and validate correctness on small datasets.

**Deliverables**
- Log ingestion (file or stdin)
- Normalization of common noisy fields:
  - timestamps, IPv4, UUIDs, long numbers, hex addresses → placeholders (e.g., `<TS>`, `<IP>`, `<UUID>`, `<NUM>`, `<HEX>`)
- Tokenization (lowercased tokens; placeholders preserved)
- SimHash **64-bit** fingerprint computation
- Hamming distance comparison (popcount)
- Brute-force near-duplicate search (O(N²)) for correctness
- Unit tests for core components + a small example dataset

> Step 1 is intentionally brute-force to keep the implementation transparent and verifiable.

---

### Step 2 — Efficient Indexing (LSH Buckets)
**Objective:** scale beyond small datasets by replacing brute-force comparisons with candidate generation.

**Deliverables**
- Banding / bucket-based indexing on SimHash fingerprints (LSH-style)
- Candidate retrieval via buckets, followed by exact Hamming verification
- CLI commands to build an index and query it
- Evaluation harness:
  - precision/recall vs distance threshold
  - indexing throughput, query latency, memory usage
- Reproducible benchmarks (CSV + plots)

---

### Step 3 — Production-Oriented Prototype (Observability + Security)
**Objective:** make the system usable in realistic pipelines (streaming logs, persistence, metrics).

**Deliverables**
- Incremental (streaming) ingestion and indexing
- Persistent storage for fingerprints and buckets (Go-friendly KV/DB)
- Time-windowed aggregation (e.g., per host/service in 30s windows)
- Metrics export (e.g., Prometheus) to support monitoring workflows
- Security/observability use-case demos:
  - log storm deduplication
  - near-duplicate clustering for suspicious activity detection
  - incident pattern surfacing across hosts

---

## Current Status 

The project is currently between **Step 1 completed** and **Step 2 in progress**.

### Implemented
- ✅ End-to-end pipeline (ingestion → normalize → tokenize → SimHash64 → matching).
- ✅ Brute-force baseline for correctness (`O(N²)`).
- ✅ Unit tests for CLI and core packages, all passing.
- ✅ In-memory LSH-like candidate generation (`BandIndex`) behind `-use-lsh`.
- ✅ Deterministic ordering of match output and optional `-limit`.
- ✅ Basic runtime stats (`records`, `comparisons`, `matches`, `prep_ms`, `search_ms`) with `-quiet-stats` opt-out.

---

## Repository Structure (high level)

- `cmd/` — CLI entrypoint(s)
- `internal/normalize/` — normalization rules for log lines
- `internal/tokenize/` — tokenization utilities
- `internal/simhash/` — SimHash + Hamming distance
- `internal/search/` — brute-force search (Step 1) and later indexing (Step 2)
- `examples/` — small example log files
- `docs/` — design notes, figures, experiment reports
- `internal/search/` — brute-force search (Step 1) and indexing building blocks (Step 2)

---

## Quick Start

Run tests:

```bash
go test ./...
```

Run the CLI on stdin (works without `examples/` files):

```bash
cat <<'EOF_LOG' | go run ./cmd/simhashlogs -k 6 -max 2000 -json
2026-02-21T10:01:02Z sshd[12345]: Failed password for invalid user admin from 192.168.1.20 port 55221 ssh2
2026-02-21T10:01:05Z sshd[12346]: Failed password for invalid user admin from 192.168.1.21 port 55222 ssh2
2026-02-21T10:02:10Z kernel: eth0 link up at 1000Mbps
EOF_LOG
```

Try LSH candidate generation + stats:

```bash
cat logs.txt | go run ./cmd/simhashlogs -k 6 -max 5000 -use-lsh -json
```
