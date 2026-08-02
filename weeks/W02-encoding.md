---
week_number: 2
status: not-started
---

# W02: Encoding and Wire Formats

> **Arc:** Data Systems Internals · **Language:** Go

## What you'll build
Varint encoding/decoding from scratch + a row vs columnar layout benchmark over 1M integer records. Numbers tell the story.

**Scenario:** two bugs live in this space that never show up as a crash, only as a wrong number or a slow query nobody can explain. Both get made visible on purpose below instead of staying theoretical.

---

## Read
- [ ] DDIA Ch.5 (2nd ed.): focus on Protobuf encoding, schema evolution, and why forward/backward compatibility matters
- [ ] [Protocol Buffers encoding spec](https://protobuf.dev/programming-guides/encoding/): read the varint and field encoding sections; this is short (~15 min)
- [ ] [Apache Arrow columnar format overview](https://arrow.apache.org/docs/format/Columnar.html): read through "Physical Memory Layout" section

**Key question:** If you have 1M rows each with 10 integer columns, is it faster to read column 3 from a row layout or columnar layout, and why?

---

## Code

Project: `code/encoding/` (Go modules)

- [ ] `varint.go`: implement protobuf-style variable-length integer encoding: `func EncodeVarint(value int64) []byte`, `func DecodeVarint(buf []byte, offset int) (value int64, newOffset int)`. Handle sign extension for negative numbers (zigzag encoding: `(n << 1) ^ (n >> 63)`, identical bit math regardless of language, Go's `int64` is signed 64-bit just like every other mainstream language's)
- [ ] `row_store.go`: store 1M records of 10 `int32` columns in row-major layout (one contiguous `[]int32` slice, stride 10, exactly the contiguous memory layout you want here, a Go slice of a fixed-width type is real backing-array storage, not a slice of pointers). Implement `func (rs *RowStore) ReadColumn(col int) []int32`
- [ ] `column_store.go`: store the same data in columnar layout (one `[]int32` per column). Implement `func (cs *ColumnStore) ReadColumn(col int) []int32`
- [ ] `benchmark_test.go`: use Go's built-in `testing.B` (`go test -bench=.`) to measure: (1) full column scan in row store vs column store; (2) point lookup by row index in both layouts. `testing.B` already handles the concerns hand-rolled timing has to get right by hand, it runs the benchmark body enough times to get a stable measurement and reports allocations per operation if you pass `-benchmem`, worth using instead of `time.Now()`/`time.Since()` by hand for this one.

**Expected outcome:** column scan should be ~5–10x faster in columnar layout. If it's not, investigate why (cache effects, or whether you accidentally allocated inside the hot loop, `-benchmem` will show you directly if that's what happened).

**Minimum bar:** varint round-trips correctly including negative numbers, and you have one measured number comparing a row-layout scan against a columnar scan of the same column. One number you measured beats three you reasoned about.

**Break it, then decide:**
- [ ] Temporarily bypass the zigzag transform: encode `-1` by feeding the raw `int64` value straight into your unsigned-varint loop instead of `(n << 1) ^ (n >> 63)` first. Count the bytes it takes. It should balloon to something like 10 bytes, because two's-complement `-1` is all 1-bits, and your continuation-bit loop has no way to know most of those bits are sign extension rather than real magnitude. Put the zigzag transform back and confirm `-1` drops to 1 byte. This is the actual, historically real bug zigzag encoding exists to prevent, not a formula worth memorizing without seeing what it protects against.
- [ ] Your benchmark's workload reads one column across all 1M rows and never reads a full row. Now imagine a second, equally realistic workload shows up: a point lookup that fetches one full record by `id`, the shape a request-serving path needs, not an analytics query. Would you keep the columnar layout and eat the cost for that second workload, add the row layout back for it specifically, or maintain both and route each query to whichever layout fits? There's no single right answer, but you should be able to say which one you'd pick for a system you were actually operating, and why. This is the real reason production systems keep separate OLTP (row) and OLAP (column) storage paths instead of picking one.

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

**Row, column, or both for the second workload, and why:**

**What I'd do differently:**
