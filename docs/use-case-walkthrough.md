# Use-Case Walkthrough - Authentication and Error Logs

This walkthrough shows how the repository detects repeated or near-duplicate log
patterns even when noisy fields differ.

The example is intentionally small, but it reflects realistic observability and
security scenarios:

- password spraying and brute-force attempts
- repeated application failures with changing request identifiers
- repeated kernel or network messages during incidents
- log-stream deduplication for triage

## Example Input

The file `examples/auth_failures.log` contains a short mixed log stream with:

- repeated SSH authentication failures from different IPs and ports
- repeated nginx 500 errors with different request IDs
- repeated kernel link-up messages
- one unrelated sudo authentication failure

## Why Exact Matching Is Not Enough

These two lines represent the same operational pattern, but exact string
comparison treats them as different events:

```text
2026-02-21T10:01:02Z sshd[12345]: Failed password for invalid user admin from 192.168.1.20 port 55221 ssh2
2026-02-21T10:01:05Z sshd[12346]: Failed password for invalid user admin from 192.168.1.21 port 55222 ssh2
```

They differ in timestamp, process ID, IP, and port. The useful structure is the
failed SSH login for the same user class.

## What Normalization Achieves

The repository first replaces common noisy fields with placeholders. Conceptually,
the SSH lines above become:

```text
<TS> sshd[<NUM>]: failed password for invalid user admin from <IP> port <NUM> ssh2
```

Likewise, nginx lines with different request IDs normalize to:

```text
<TS> nginx[<NUM>]: <NUM> error on get /api/orders request_id=<UUID>
```

This preserves the event shape while reducing incidental variance.

## Tokenization and Fingerprinting

After normalization, each line is tokenized and converted into a 64-bit SimHash
fingerprint. Similar normalized token streams should produce fingerprints with
small Hamming distance.

This lets the repository separate:

- exact or near-exact repeats
- related events with small structural changes
- unrelated outliers

## Expected Behavior

On `examples/auth_failures.log`, a reviewer should see:

- SSH authentication failures grouped closely.
- Kernel link-up messages matched almost exactly.
- The two `/api/orders` nginx failures matched closely.
- The `/api/profile` nginx failure related but weaker than the `/api/orders`
  pair.
- The sudo authentication failure mostly isolated.

## Suggested CLI Usage

Run the example with the brute-force baseline:

```bash
go run ./cmd/simhashlogs dedup -input examples/auth_failures.log -k 6 -max 2000 -json
```

Run the same example with LSH-style candidate generation:

```bash
go run ./cmd/simhashlogs dedup -input examples/auth_failures.log -k 6 -max 2000 -use-lsh -json
```

Run the paper-style sorted permutation index:

```bash
go run ./cmd/simhashlogs dedup -input examples/auth_failures.log -k 6 -max 2000 -index paper -json
```

Compare candidate indexes against brute force:

```bash
go run ./cmd/simhashlogs eval -input examples/auth_failures.log -k 6 -max 2000
```

Collect CSV rows from the reproducible synthetic benchmark:

```bash
go run ./cmd/simhashlogs eval -input examples/synthetic_benchmark.log -k-values 3,6 -bands-values 0,8 -csv
```

## What This Demonstrates

The repository is not only a SimHash implementation in the abstract. It is a
small engineering artifact for observability and security workflows where noisy
log streams need to be reduced into a smaller set of recurring behavior patterns.
