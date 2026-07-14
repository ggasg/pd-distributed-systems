---
week_number: 2
status: not-started
---

# W02: Encoding and Wire Formats

> **Arc:** Data Systems Internals · **Language:** Go

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

Project: `code/encoding/` (Go, module)

- [ ] `varint.go`: implement protobuf-style variable-length integer encoding: `EncodeVarint(value int64) []byte`, `DecodeVarint(buf []byte, offset int) (int64, int)` (returns decoded value and new offset). Handle sign extension for negative numbers (zigzag encoding: `(n << 1) ^ (n >> 63)`)
- [ ] `row_store.go`: store 1M records of `[10]int32` in row-major layout (one contiguous `[]int32` slice, stride 10). Implement `ReadColumn(col int) []int32`
- [ ] `column_store.go`: store the same data in columnar layout (one `[]int32` per column). Implement `ReadColumn(col int) []int32`
- [ ] `benchmark.go`, exposed as a `cmd/benchmark` binary: use `time.Now()`/`time.Since()` to measure: (1) full column scan in row store vs column store; (2) point lookup by row index in both layouts. Print results. Alternatively, write this as a proper Go benchmark using `testing.B` (`go test -bench=.`) instead of hand-rolled timing; either is fine, but if you go the `testing.B` route note that Go's benchmark framework already handles warm-up iterations for you

**Expected outcome:** column scan should be ~5–10x faster in columnar layout. If it's not, investigate why (cache effects, whether you accidentally allocated inside the hot loop).

---

## 🐍 Python DSA Review (optional)

**Bit manipulation + byte packing**: implement varint in Python before doing it in Go. The bit ops are the same; Python makes them easy to inspect.

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
        shift += 7

# Verify round-trip
for n in [0, 1, 127, 128, 300, 16383, 2**21 - 1]:
    encoded = encode_varint(n)
    decoded, _ = decode_varint(encoded)
    assert decoded == n, f"Failed for {n}"
    print(f"{n:>8} → {list(encoded)} ({len(encoded)} bytes)")
```

**Connection:** this IS the week's core topic, but implementing it in Python first lets you verify the bit logic interactively before writing the Go version, where a wrong shift or mask is a silent correctness bug rather than a crash.

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
