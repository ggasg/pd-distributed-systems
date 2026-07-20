---
week_number: 2
status: not-started
---

# W02: Encoding and Wire Formats

> **Arc:** Data Systems Internals · **Language:** Java

## What you'll build
Varint encoding/decoding from scratch + a row vs columnar layout benchmark over 1M integer records. Numbers tell the story.

---

## Read
- [ ] DDIA Ch.4: focus on Thrift/Protobuf encoding, schema evolution, and why forward/backward compatibility matters
- [ ] [Protocol Buffers encoding spec](https://protobuf.dev/programming-guides/encoding/): read the varint and field encoding sections; this is short (~15 min)
- [ ] [Apache Arrow columnar format overview](https://arrow.apache.org/docs/format/Columnar.html): read through "Physical Memory Layout" section

**Key question:** If you have 1M rows each with 10 integer columns, is it faster to read column 3 from a row layout or columnar layout, and why?

---

## Code

Project: `code/encoding/` (Java 21, Maven)

- [ ] `Varint.java`: implement protobuf-style variable-length integer encoding: `static byte[] encodeVarint(long value)`, `static long[] decodeVarint(byte[] buf, int offset)` (return a two-element array: decoded value and new offset, or define a small `record DecodedVarint(long value, int newOffset) {}` if you'd rather have named fields than a bare array). Handle sign extension for negative numbers (zigzag encoding: `(n << 1) ^ (n >> 63)`, identical bit math to every other language, Java's `long` is signed 64-bit just like Go's `int64`)
- [ ] `RowStore.java`: store 1M records of 10 `int` columns in row-major layout (one contiguous `int[]` array, stride 10; Java's `int` is already a fixed 32-bit signed integer, no separate `int32` type to reach for the way Go needed one). Implement `int[] readColumn(int col)`
- [ ] `ColumnStore.java`: store the same data in columnar layout (one `int[]` per column). Implement `int[] readColumn(int col)`
- [ ] `Benchmark.java`: use `System.nanoTime()` to measure: (1) full column scan in row store vs column store; (2) point lookup by row index in both layouts. Print results. Alternatively, use [JMH](https://github.com/openjdk/jmh) (the standard Java microbenchmark harness) instead of hand-rolled timing; either is fine, but if you go the JMH route note that it already handles JIT warm-up iterations for you, the same reason Go's `testing.B` was offered as an alternative in other languages' versions of this exercise, hand-rolled timing on a cold JIT will understate steady-state performance if you're not careful to warm up first

**Expected outcome:** column scan should be ~5–10x faster in columnar layout. If it's not, investigate why (cache effects, whether you accidentally allocated inside the hot loop, or whether the JIT hasn't warmed up yet if you're using hand-rolled timing).

---

## 🐍 Python DSA Review (optional)

**Bit manipulation + byte packing**: implement varint in Python before doing it in Java. The bit ops are the same; Python makes them easy to inspect.

```python
# varint.py
def encode_varint(n: int) -> bytes:
    out = []
    while n > 0x7F:
        out.append((n & 0x7F) | 0x80)  # low 7 bits + continuation bit
        n >>= 7
    out.append(n)
    return bytes(out)

def decode_varint(data: bytes, pos: int = 0) -> tuple[int, int]:
    result, shift = 0, 0
    while True:
        b = data[pos]; pos += 1
        result |= (b & 0x7F) << shift
        if not (b & 0x80):  # no continuation bit, done
            return result, pos

# Verify round-trip
for n in [0, 1, 127, 128, 300, 16383, 2**21 - 1]:
    encoded = encode_varint(n)
    decoded, _ = decode_varint(encoded)
    assert decoded == n, f"Failed for {n}"
    print(f"{n:>8} → {list(encoded)} ({len(encoded)} bytes)")
```

**Connection:** this IS the week's core topic, but implementing it in Python first lets you verify the bit logic interactively before writing the Java version, where a wrong shift or mask is a silent correctness bug rather than a crash.

---

## Reflect

**What clicked:**

**What surprised me:**

**Benchmark results:**
- Row store column scan: __ ms
- Column store column scan: __ ms
- Speedup: __x

**How this connects to a system you've worked with:**

**What I'd do differently:**
