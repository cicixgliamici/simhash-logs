# Normalization Examples

This document shows how noisy operational fields are reduced to placeholders
before tokenization and SimHash fingerprinting.

The goal is to preserve event structure while removing values that commonly
change from one occurrence to the next.

## SSH Authentication Failure

Raw:

```text
2026-02-21T10:01:02Z sshd[12345]: Failed password for invalid user admin from 192.168.1.20 port 55221 ssh2
```

Normalized:

```text
<ts> sshd[<num>]: failed password for invalid user admin from <ip> port <num> ssh2
```

The important structure is preserved: SSH context, failed password, invalid
user, and username. Timestamp, process ID, IP, and port are abstracted away.

## Repeated SSH Failure With Different Source Values

Raw:

```text
2026-02-21T10:01:29Z sshd[12350]: Failed password for invalid user admin from 192.168.1.25 port 55226 ssh2
```

Normalized:

```text
<ts> sshd[<num>]: failed password for invalid user admin from <ip> port <num> ssh2
```

This becomes structurally identical to the previous SSH example, which is why
the pair should have a very small Hamming distance.

## Nginx Application Error

Raw:

```text
2026-02-21T10:02:55Z nginx[987]: 500 error on GET /api/orders request_id=6f0f3e12-91d6-4c0b-b6a8-7feee5c7e201
```

Normalized:

```text
<ts> nginx[987]: 500 error on get /api/orders request_id=<uuid>
```

The endpoint and error pattern remain visible, while the request ID stops
dominating similarity.

## Similar Application Error on Another Endpoint

Raw:

```text
2026-02-21T10:03:11Z nginx[989]: 500 error on GET /api/profile request_id=41ac1bca-7790-4902-8dae-b6ce49f7d22f
```

Normalized:

```text
<ts> nginx[989]: 500 error on get /api/profile request_id=<uuid>
```

This should remain related to the `/api/orders` errors, but less strongly than
two errors on the same endpoint.

## Kernel Link Message

Raw:

```text
2026-02-21T10:02:10Z kernel: eth0 link up at 1000Mbps
```

Normalized:

```text
<ts> kernel: eth0 link up at 1000mbps
```

Repeated operational messages often differ only by time or small numeric values.
Normalization helps group them into stable event patterns.
