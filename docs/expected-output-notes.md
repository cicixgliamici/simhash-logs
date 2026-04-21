# Expected Output Notes — SimHash Log Demo

This note explains what a reviewer should expect, conceptually, when running the example log file through the CLI.

It is intentionally descriptive rather than tied to one exact output format, because the repository may evolve while keeping the same core behavior.

---

## Example command

```bash
cat examples/auth-failures.log | go run ./cmd/simhashlogs -k 6 -max 2000 -json
````

If LSH-style candidate generation is enabled:

```bash
cat examples/auth-failures.log | go run ./cmd/simhashlogs -k 6 -max 2000 -use-lsh -json
```

---

## What a reviewer should expect conceptually

A good run on the example input should show behavior of the following kind:

### 1. SSH authentication failures should cluster

Several lines describing failed SSH logins for an invalid user should appear as close matches or near-duplicates.

They differ mainly in:

* timestamp
* process ID
* IP address
* port number

So after normalization and SimHash fingerprinting, they should remain close in Hamming space.

---

### 2. Kernel link-up messages should be very close or identical

The repeated kernel messages should likely become exact duplicates or near-exact duplicates after normalization.

This is the simplest sanity-check cluster in the example.

---

### 3. Nginx 500 errors on the same endpoint should be related

The repeated `/api/orders` errors should appear as a near-duplicate pair because they preserve the same structural event pattern while changing only high-variance request identifiers.

---

### 4. The `/api/profile` error should be related, but less strongly

The `/api/profile` error should still look more similar to the other nginx 500 errors than to the SSH or kernel messages, but it should generally be a weaker match than the two `/api/orders` lines are to each other.

---

### 5. The sudo authentication failure should remain relatively isolated

This line should not collapse into the SSH failure cluster even though both concern authentication. Its surface structure and token composition are sufficiently different that it should appear as an outlier or much weaker match.

---

## What this demonstrates

The example is useful because it highlights three different cases:

* **strong near-duplicates**
  repeated SSH failures, repeated kernel lines, repeated nginx errors on the same endpoint

* **moderate similarity**
  nginx errors on different endpoints

* **clear outliers**
  an unrelated authentication-related line with different structure

This is exactly the kind of distribution one wants in a practical near-duplicate detection system.

---

## Good follow-up improvement

A strong future addition would be to include one real sample output block from the CLI, annotated line by line to explain why the matches make sense.

