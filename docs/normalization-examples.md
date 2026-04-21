# Normalization Examples — SimHash for Logs

This document gives a few concrete before/after normalization examples to make the repository easier to review quickly.

The goal is to show how noisy operational fields are reduced to placeholders so that structural similarity becomes easier to detect.

---

## Example 1 — SSH authentication failure

### Raw line

```text
2026-02-21T10:01:02Z sshd[12345]: Failed password for invalid user admin from 192.168.1.20 port 55221 ssh2
````

### Normalized intuition

```text
<TS> sshd[<NUM>]: Failed password for invalid user admin from <IP> port <NUM> ssh2
```

### Why this helps

The important structure is preserved:

* authentication failure
* invalid user `admin`
* SSH context

The high-variance fields are abstracted away:

* timestamp
* process identifier
* IP address
* port number

---

## Example 2 — Repeated SSH failure with different source values

### Raw line

```text
2026-02-21T10:01:29Z sshd[12350]: Failed password for invalid user admin from 192.168.1.25 port 55226 ssh2
```

### Normalized intuition

```text
<TS> sshd[<NUM>]: Failed password for invalid user admin from <IP> port <NUM> ssh2
```

### Why this helps

After normalization, this line becomes structurally almost identical to the previous SSH example. That is exactly the kind of relationship the repository is meant to detect as a near-duplicate.

---

## Example 3 — Nginx application error

### Raw line

```text
2026-02-21T10:02:55Z nginx[987]: 500 error on GET /api/orders request_id=6f0f3e12-91d6-4c0b-b6a8-7feee5c7e201
```

### Normalized intuition

```text
<TS> nginx[<NUM>]: <NUM> error on GET /api/orders request_id=<UUID>
```

### Why this helps

The endpoint and error pattern remain visible, while the request ID and process identifier stop dominating the string representation.

---

## Example 4 — Similar application error on another endpoint

### Raw line

```text
2026-02-21T10:03:11Z nginx[989]: 500 error on GET /api/profile request_id=41ac1bca-7790-4902-8dae-b6ce49f7d22f
```

### Normalized intuition

```text
<TS> nginx[<NUM>]: <NUM> error on GET /api/profile request_id=<UUID>
```

### Why this helps

This line remains similar to the previous nginx error, but not identical:

* same service style
* same HTTP method
* same error pattern
* different endpoint

This is the kind of case where Hamming distance becomes more informative than exact matching.

---

## Example 5 — Kernel link message

### Raw line

```text
2026-02-21T10:02:10Z kernel: eth0 link up at 1000Mbps
```

### Normalized intuition

```text
<TS> kernel: eth0 link up at <NUM>Mbps
```

### Why this helps

Repeated operational messages often differ only in timing or small numeric values. Normalization helps group them into stable event patterns.

---

## Summary

These examples illustrate the repository’s core engineering idea:

> preserve semantic structure, reduce incidental variance.

That is what makes downstream tokenization, SimHash fingerprinting, and Hamming-distance comparison useful for noisy log streams.
