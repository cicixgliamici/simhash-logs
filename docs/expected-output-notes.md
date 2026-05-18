# Expected Output Notes - SimHash Log Demo

This note explains what a reviewer should expect when running the example log
file through the CLI. It focuses on behavior rather than exact formatting.

## Example Commands

Brute-force baseline:

```bash
go run ./cmd/simhashlogs dedup -input examples/auth_failures.log -k 6 -max 2000 -json
```

LSH-style candidate generation:

```bash
go run ./cmd/simhashlogs dedup -input examples/auth_failures.log -k 6 -max 2000 -use-lsh -json
```

Evaluation against brute force:

```bash
go run ./cmd/simhashlogs eval -input examples/auth_failures.log -k 6 -max 2000
```

CSV parameter sweep:

```bash
go run ./cmd/simhashlogs eval -input examples/auth_failures.log -k-values 3,6,9 -bands-values 0,5,8 -csv
```

## Expected Match Patterns

### 1. SSH authentication failures cluster

Several failed SSH login lines should appear as close matches. They differ mainly
in timestamp, process ID, IP address, and port number, so normalization makes
them identical or nearly identical before fingerprinting.

### 2. Kernel link-up messages are very close

Repeated kernel link messages should become exact or near-exact duplicates after
normalization. This is the simplest sanity check in the sample.

### 3. Nginx 500 errors on the same endpoint match closely

The repeated `/api/orders` errors should appear as a near-duplicate pair because
the endpoint and error shape are preserved while the request IDs are normalized.

### 4. Nginx errors on different endpoints are weaker matches

The `/api/profile` error should look more similar to the other nginx errors than
to SSH or kernel lines, but weaker than the two `/api/orders` lines.

### 5. The sudo authentication failure is mostly isolated

The sudo line is security-related, but its token structure differs enough from
the SSH failures that it should not collapse into the main SSH cluster.

## Current Example Evaluation

On the current sample with `k=6`, `eval` reports the LSH-style search matching
the brute-force result set while doing fewer exact comparisons:

```text
Records:           12
Distance (k):      6
LSH Bands:         7
Brute matches:     17
Brute comparisons: 66
LSH matches:       17
LSH comparisons:   22
Recall:            100.00%
```

This is a small sanity check, not a benchmark claim. The new `-k-values` and
`-bands-values` sweep flags make it easier to collect comparison rows, but larger
datasets are still needed before making performance claims.
