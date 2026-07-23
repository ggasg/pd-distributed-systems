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

Project: `code/encoding/` (Go modules)

- [ ] `varint.go`: implement protobuf-style variable-length integer encoding: `func EncodeVarint(value int64) []byte`, `func DecodeVarint(buf []byte, offset int) (value int64, newOffset int)`. Handle sign extension for negative numbers (zigzag encoding: `(n << 1) ^ (n >> 63)`, identical bit math regardless of language, Go's `int64` is signed 64-bit just like every other mainstream language's)
- [ ] `row_store.go`: store 1M records of 10 `int32` columns in row-major layout (one contiguous `[]int32` slice, stride 10, exactly the contiguous memory layout you want here, a Go slice of a fixed-width type is real backing-array storage, not a slice of pointers). Implement `func (rs *RowStore) ReadColumn(col int) []int32`
- [ ] `column_store.go`: store the same data in columnar layout (one `[]int32` per column). Implement `func (cs *ColumnStore) ReadColumn(col int) []int32`
- [ ] `benchmark_test.go`: use Go's built-in `testing.B` (`go test -bench=.`) to measure: (1) full column scan in row store vs column store; (2) point lookup by row index in both layouts. `testing.B` already handles the concerns hand-rolled timing has to get right by hand, it runs the benchmark body enough times to get a stable measurement and reports allocations per operation if you pass `-benchmem`, worth using instead of `time.Now()`/`time.Since()` by hand for this one.

**Expected outcome:** column scan should be ~5–10x faster in columnar layout. If it's not, investigate why (cache effects, or whether you accidentally allocated inside the hot loop, `-benchmem` will show you directly if that's what happened).

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

**What I'd do differently:**
