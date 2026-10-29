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
- **Token Weighting Modes**:
  - *Historical Uniform/Frequency Weighting*: Each token occurrence contributes $\pm 1$ in integer accumulators.
  - *Information-Theoretic TF-IDF Weighting (`IDFModel`)*: Learned from corpus document frequencies. Downweights ubiquitous boilerplate tokens (e.g. `<ts>`, `sshd`, `<ip>`) and upweights rare discriminative error tokens, implementing Charikar's arbitrary weighted vector model $\vec{v} = \sum w_i \vec{r}_i$.

Simplified:

- Token hashing uses built-in FNV-1a rather than explicitly sampled random
  hyperplanes.
- Input objects are normalized log token streams rather than arbitrary continuous
  vectors.

## Manku, Jain, and Das Sarma 2007

Implemented:

- Near-duplicate detection over SimHash fingerprints.
- Candidate generation before exact Hamming verification.
- Evaluation against brute-force ground truth.
- **Historical v1 Prototype (`PermutationIndex` / `-index paper`)**:
  - Each table applied a deterministic bit rotation.
  - Candidate lookup scanned a bounded numerical window around insertion points.
  - Preserved as an empirical baseline and evolutionary record showing why numerical
    window sorting over 64-bit integers lacks theoretical distance guarantees.
- **Paper-Rigorous v2 Implementation (`MankuIndex` / `-index manku`)**:
  - Implements Section 3.1 of Manku et al. (2007) based on the **Pigeonhole Principle**.
  - Partitions 64 bits into $p = k + 1$ disjoint blocks.
  - Constructs $p$ sorted tables where each table permutes one block to the most significant bit positions (prefix).
  - Uses binary range search (`[minKey, maxKey]`) to retrieve the exact contiguous slice of fingerprints sharing the identical block prefix in $O(\log N)$ time.
  - **Theoretical Guarantee**: For any pair with Hamming distance $\le k$, at least one block has zero differences, guaranteeing **100% Recall** with no heuristic windows and minimal exact comparisons.

Simplified or not implemented yet:

- `BandIndex` is an engineering LSH baseline, not part of the paper-faithful
  path.
- The sorted permutation tables are in-memory; there is no disk-backed storage or delta-compression yet.
- There is no large web-scale corpus or benchmark harness yet; the included
  synthetic benchmark is for reproducible local sanity checks.

## Engineering Interpretation

The project currently answers:

> Can SimHash make noisy system-log patterns detectable as near-duplicates?

The Step 2 implementation now also answers:

> Can we scale candidate generation while preserving measurable recall against
> the brute-force baseline?
