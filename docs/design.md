# Step 1 Design

## Goal

Build a minimal, correct, and reproducible pipeline for near-duplicate detection on system logs using SimHash.

## Pipeline

```text
read lines -> normalize -> tokenize -> simhash64 -> brute-force search -> print matches
