# Paper Mapping

This repository implements the core SimHash idea and applies it to system logs.
It is currently a practical prototype, not a complete reproduction of every
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

Simplified or not implemented yet:

- The current `BandIndex` is an LSH-style bucket index, not the paper's full
  sorted fingerprint-table/permutation scheme.
- There is no disk-backed index.
- There is no large web-scale corpus or benchmark harness yet.

## Engineering Interpretation

The project currently answers:

> Can SimHash make noisy system-log patterns detectable as near-duplicates?

The next implementation work should answer:

> Can we scale candidate generation while preserving measurable recall against
> the brute-force baseline?
