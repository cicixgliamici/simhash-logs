# Use-Case Walkthrough — Near-Duplicate Detection for Authentication and Error Logs

This walkthrough shows how the repository can be used to detect repeated or near-duplicate log patterns even when noisy fields differ.

The example is intentionally small, but it reflects realistic observability and security scenarios such as:
- password-spraying or brute-force attempts
- repeated application failures with changing request identifiers
- log-stream deduplication during incidents

---

## Example input

The file `examples/auth-failures.log` contains a short mixed log stream with:
- repeated SSH authentication failures from different IP addresses and ports
- repeated nginx 500 errors with different request IDs
- repeated kernel link-up messages
- one unrelated sudo authentication failure

---

## Why exact matching is not enough

If we compare raw lines directly, many of the most interesting events appear different because they contain high-variance fields such as:
- timestamps
- IPv4 addresses
- ports
- request IDs / UUIDs
- process identifiers

For example, these two lines should be understood as essentially the same pattern:

```text
2026-02-21T10:01:02Z sshd[12345]: Failed password for invalid user admin from 192.168.1.20 port 55221 ssh2
2026-02-21T10:01:05Z sshd[12346]: Failed password for invalid user admin from 192.168.1.21 port 55222 ssh2
````

But exact string comparison would treat them as different events.

---

## What normalization achieves

The repository first normalizes noisy fields into placeholders.

Conceptually, lines like the two above become something close to:

```text
<TS> sshd[<NUM>]: Failed password for invalid user admin from <IP> port <NUM> ssh2
```

Likewise, nginx lines with different request IDs can normalize to something like:

```text
<TS> nginx[<NUM>]: <NUM> error on GET /api/orders request_id=<UUID>
```

This makes the structural similarity much easier to capture.

---

## Tokenization and fingerprinting

After normalization, the line is tokenized and converted into a SimHash fingerprint.

The key idea is that similar normalized token sets should produce fingerprints with small Hamming distance.

That allows the repository to detect near-duplicates efficiently:

* exact duplicates are easy
* near-duplicates remain close in fingerprint space
* unrelated lines remain farther apart

---

## Expected intuition on this example

On the sample file, we would expect the system to recover clusters such as:

### Cluster A — SSH authentication failures

These lines should group closely because they differ mostly in timestamp, IP, port, and PID.

### Cluster B — Kernel link-up messages

These should likely match almost exactly after normalization.

### Cluster C — Nginx 500 errors on the same endpoint

The two `/api/orders` failures should appear closely related.

### Outlier — Sudo authentication failure

This line should remain relatively distinct from the SSH and nginx clusters.

---

## Why this matters in practice

This type of matching is useful in real pipelines because it supports:

* **incident deduplication**
  Avoid showing analysts hundreds of nearly identical lines.

* **attack pattern surfacing**
  Detect repeated authentication failures that vary in source fields.

* **noise reduction**
  Group recurring operational messages into stable patterns.

* **faster triage**
  Move from “many lines” to “a few behavior classes”.

---

## Suggested CLI usage

A reviewer can try the example with something like:

```bash
go test ./...
cat examples/auth-failures.log | go run ./cmd/simhashlogs -k 6 -max 2000 -json
```

And, if LSH-style candidate generation is available:

```bash
cat examples/auth-failures.log | go run ./cmd/simhashlogs -k 6 -max 2000 -use-lsh -json
```

---

## What this walkthrough demonstrates

This small example is meant to show that the repository is not only an implementation of SimHash in the abstract, but also a concrete engineering artifact for:

* observability workflows
* security-oriented log analysis
* near-duplicate clustering under noisy field variation

---

## Natural future extensions

Good next additions to this walkthrough would be:

* a short “before vs after normalization” table
* an example of match output with Hamming distances
* a mini benchmark comparing brute-force mode and LSH-based candidate generation
