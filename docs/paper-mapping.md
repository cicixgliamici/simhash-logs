# Paper Mapping

This repository implements the core SimHash idea and applies it to system logs.
It is currently a practical prototype, not a web-scale reproduction of every
algorithmic detail in the reference papers.

## Charikar 2002

Implemented:

- Compact binary fingerprints for similarity estimation.
- 64-bit signatures built from token hashes.
- Similarity search by Hamming distance between fingerprints.
- Deterministic implementation suitable for tests and examples.

Simplified:

- Token hashing uses built-in FNV-1a rather than explicitly sampled random
  hyperplanes.
- Input objects are normalized log token streams rather than arbitrary weighted
  vectors.
- Token frequency is used as a simple weight.

## Manku, Jain, and Das Sarma 2007

Implemented:

- Near-duplicate detection over SimHash fingerprints.
- Candidate generation before exact Hamming verification.
- Evaluation against brute-force ground truth.
- A small sorted fingerprint-table/permutation index:
  - each table stores fingerprints by a deterministic permuted key;
  - lookup scans a bounded neighborhood around the query key;
  - every candidate is still verified with exact Hamming distance.

Simplified or not implemented yet:

- `BandIndex` is an engineering LSH baseline, not part of the paper-faithful
  path.
- `PermutationIndex` captures the paper's sorted-table idea, but uses simple
  deterministic bit rotations and configurable scan windows rather than the
  full production-scale table construction and tuning from the paper.
- There is no disk-backed index.
- There is no large web-scale corpus or benchmark harness yet; the included
  synthetic benchmark is for reproducible local sanity checks.

## Engineering Interpretation

The project currently answers:

> Can SimHash make noisy system-log patterns detectable as near-duplicates?

The Step 2 implementation now also answers:

> Can we scale candidate generation while preserving measurable recall against
> the brute-force baseline?
