---
week_number: 8
status: not-started
---

# W08: Query Execution

> **Arc:** Streaming and Dataflow · **Language:** Rust

## What you'll build
A vectorized query executor in Rust: columnar filter + hash join + projection over in-memory data. Benchmark it against a row-at-a-time version of the same pipeline and measure the speedup.

---

## Read
- [ ] [Volcano, An Extensible and Parallel Query Evaluation System](https://dl.acm.org/doi/10.1109/69.273032) (Graefe, 1994): read Sections 1–3. This defines the iterator model (the `next()` interface) that every query engine for 20 years was built on.
- [ ] [MonetDB/X100: Hyper-Pipelining Query Execution](https://www.cidrdb.org/cidr2005/papers/P19.pdf) (Boncz et al., CIDR 2005): read Sections 1–3. This is the argument for vectorized execution and why Volcano is CPU-cache unfriendly.

**Key question:** Why does calling `next()` once per row hurt CPU performance even when the logic is simple? What does processing a batch of 1024 rows at a time fix?

---

## Code

Project: `code/query-exec/` (Rust, cargo)

Data model: a table of 1M rows with columns `id: Vec<i32>`, `dept: Vec<i32>`, `salary: Vec<i32>` stored separately (columnar).

**Row-at-a-time executor (baseline):**

- [ ] `row_executor.rs`: `struct Row { id: i32, dept: i32, salary: i32 }`; an iterator that zips the three columns (`id.iter().zip(dept.iter()).zip(salary.iter())`); apply `.filter(|r| r.salary > threshold)`, then `.map(|r| (r.id, r.salary))`; collect results

**Vectorized executor:**

- [ ] `column_filter.rs`: `fn filter(col: &[i32], threshold: i32) -> Vec<bool>`, branchless: `col.iter().map(|&v| v > threshold).collect()`
- [ ] `column_project.rs`: `fn project(col: &[i32], mask: &[bool]) -> Vec<i32>`, collect values where mask is true
- [ ] `hash_join.rs`: `fn join(left_key: &[i32], left_val: &[i32], right_key: &[i32], right_val: &[i32]) -> Vec<(i32, i32)>`, build a `HashMap<i32, i32>` from the left side, probe with the right side, emit matching pairs
- [ ] `src/bin/benchmark.rs`: generate 1M random rows; time the filter + project pipeline 10 times each using `std::time::Instant` (3 warm-up runs first, so caches and branch prediction are primed); print mean throughput in M rows/sec for both executors. Run with `cargo run --release --bin benchmark` — **the `--release` flag is not optional here**, an unoptimized debug build will make the comparison meaningless

**Expected outcome:** vectorized should be 3–8x faster. If the gap is smaller, the first thing to check is whether you actually ran `--release` — that's the single most common reason a Rust benchmark looks flat.

---

## 🐍 Python DSA Review (optional)

**Hash join + binary search on sorted arrays**: the two algorithms your Rust `hash_join.rs` and `column_filter.rs` implement. Python makes the probe/build logic easy to inspect.

```python
from collections import defaultdict
from bisect import bisect_left

# hash_join.py: classic hash join, build phase + probe phase
def hash_join(left: list[dict], right: list[dict], key: str) -> list[dict]:
    # Build: index left side by join key
    ht: dict = defaultdict(list)
    for row in left:
        ht[row[key]].append(row)
    # Probe: for each right row, look up in hash table
    result = []
    for row in right:
        for match in ht.get(row[key], []):
            result.append(match | row)  # merge dicts (Python 3.9+)
    return result

left  = [{"id": 1, "name": "alice"}, {"id": 2, "name": "bob"}]
right = [{"id": 1, "score": 95},     {"id": 1, "score": 88}]
joined = hash_join(left, right, "id")
assert len(joined) == 2 and all(r["name"] == "alice" for r in joined)

# binary_filter.py: binary search on a sorted column (like column_filter.rs)
def filter_sorted_col(col: list[int], predicate_min: int, predicate_max: int) -> list[int]:
    lo = bisect_left(col, predicate_min)
    hi = bisect_left(col, predicate_max + 1)
    return col[lo:hi]  # slice is O(k) not O(n)

col = sorted([5, 2, 8, 1, 9, 3, 7, 4, 6])
assert filter_sorted_col(col, 3, 7) == [3, 4, 5, 6, 7]
```

**Connection:** `hash_join.rs` is the build+probe pattern in Rust over `&[i32]` columns. `column_filter.rs` can use a `bisect_left`-equivalent (binary search on a sorted column, `slice::binary_search`) for range filters, 3–8x faster than scanning.

---

## Reflect

**What clicked:**

**What surprised me:**

**Benchmark results:**
- Row-at-a-time: __ M rows/sec
- Vectorized: __ M rows/sec
- Speedup: __x

**What does this tell you about how query execution works in a system you know?**

**Where did mutation buy you speed?** Your vectorized executor builds and fills `Vec`s directly for performance. That's in tension with the "never mutate, always produce a new `Collection`" style the borrow checker forced on you in W05 and W07. Point to one specific place in `column_filter.rs` or `hash_join.rs` where you used `&mut` or filled a pre-allocated `Vec` in place, and estimate what it would've cost you in speed to avoid that mutation and chain allocations instead.

**What I'd do differently:**
