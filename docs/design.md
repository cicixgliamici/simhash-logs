# Design (Step 1)

## Pipeline
```mermaid
flowchart LR
  A[Raw log line] --> B[Normalize\n<TS>, <IP>, <UUID>, <NUM>, <HEX>]
  B --> C[Tokenize]
  C --> D[SimHash64]
  D --> E[Brute force\npairwise compare]
  E --> F[Near-duplicate pairs\nHamming <= k]