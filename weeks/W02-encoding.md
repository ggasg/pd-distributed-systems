---
week_number: 2
status: not-started
---

# W02 — Encoding and Wire Formats

> **Arc:** Data Systems Internals · **Language:** Java 21

## What you'll build
Varint encoding/decoding from scratch + a row vs columnar layout benchmark over 1M integer records. Numbers tell the story.

---

## Read
- [ ] DDIA Ch.4 — focus on Thrift/Protobuf encoding, schema evolution, and why forward/backward compatibility matters
- [ ] [Protocol Buffers encoding spec](https://protobuf.dev/programming-guides/encoding/) — read the varint and field encoding sections; this is short (~15 min)
- [ ] [Apache Arrow columnar format overview](https://arrow.apache.org/docs/format/Columnar.html) — read through "Physical Memory Layout" section

**Key question:** If you have 1M rows each with 10 integer columns, is it faster to read column 3 from a row layout or columnar layout, and why?

---

## Code

Project: `code/encoding/` (Java 21)

- [ ] `Varint.java` — implement protobuf-style variable-length integer encoding: `encode(long value) -> byte[]`, `decode(byte[] buf, int offset) -> long`. Handle sign extension for negative numbers (zigzag encoding).
- [ ] `RowStore.java` — store 1M records of `int[10]` in row-major layout (one contiguous byte array). Implement `readColumn(int col) -> int[]`.
- [ ] `ColumnStore.java` — store the same data in columnar layout (one array per column). Implement `readColumn(int col) -> int[]`.
- [ ] `Benchmark.java` — use `System.nanoTime()` to measure: (1) full column scan in row store vs column store; (2) point lookup by row index in both layouts. Print results.

**Expected outcome:** column scan should be ~5–10x faster in columnar layout. If it's not, investigate why (cache effects, JIT warmup).

---

## Reflect

**What clicked:**

**What surprised me:**

**Benchmark results:**
- Row store column scan: __ ms
- Column store column scan: __ ms
- Speedup: __x

**How this connects to my current role:**

**What I'd do differently:**
