---
week_number: 8
status: not-started
---

# W08: Query Execution

> **Arc:** Streaming and Dataflow · **Language:** C++

## What you'll build
A vectorized query executor in C++: columnar filter + hash join + projection over in-memory data. Benchmark it against a row-at-a-time version of the same pipeline and measure the speedup.

---

## Read
- [ ] [Volcano, An Extensible and Parallel Query Evaluation System](https://dl.acm.org/doi/10.1109/69.273032) (Graefe, 1994): read Sections 1–3. This defines the iterator model (the `next()` interface) that every query engine for 20 years was built on.
- [ ] [MonetDB/X100: Hyper-Pipelining Query Execution](https://www.cidrdb.org/cidr2005/papers/P19.pdf) (Boncz et al., CIDR 2005): read Sections 1–3. This is the argument for vectorized execution and why Volcano is CPU-cache unfriendly.
- [ ] [DuckDB execution engine source](https://github.com/duckdb/duckdb/tree/main/src/execution): optional but worth it: a real, actively maintained vectorized query engine in C++, and one you already depend on via W11's feature store. Skim `PhysicalFilter` and how DuckDB batches rows into `DataChunk`s; that's the production version of what you're building this week.
- [ ] Optional, context only (a free public blog post, not something you install or test against): [Announcing Photon](https://www.databricks.com/blog/2021/06/17/announcing-photon-public-preview-the-next-generation-query-engine-on-the-databricks-lakehouse-platform.html). Photon is "written from the ground up in C++" specifically to replace the JVM-based Spark execution engine for exactly this reason: columnar batches, tight vectorized loops, SIMD. Your actual hands-on comparison this week is DuckDB (and optionally ClickHouse) below; this is just confirmation the same technique is load-bearing in production.
- [ ] [ClickHouse execution pipeline source](https://github.com/ClickHouse/ClickHouse/tree/master/src/Processors): optional: ClickHouse is C++ end to end; skim `IProcessor` and how the pull-based pipeline batches rows into `Chunk`s. A second real reference point alongside DuckDB and Photon, from a different target company with a different pipeline design.

**Key question:** Why does calling `next()` once per row hurt CPU performance even when the logic is simple? What does processing a batch of 1024 rows at a time fix?

---

## Code

Project: `code/query-exec/` (C++, CMake)

Data model: a table of 1M rows with columns `std::vector<int32_t> id, dept, salary` stored separately (columnar).

**Row-at-a-time executor (baseline):**

- [ ] `include/query_exec/row_executor.hpp`: `struct Row { int32_t id; int32_t dept; int32_t salary; };`; build rows by iterating the three vectors in lockstep (a plain indexed `for` loop is clearest here); apply a filter predicate (`salary > threshold`), then project `(id, salary)`; collect results into a `std::vector<std::pair<int32_t, int32_t>>`

**Vectorized executor:**

- [ ] `include/query_exec/column_filter.hpp` + `src/column_filter.cpp`: `std::vector<bool> filter(const std::vector<int32_t>& col, int32_t threshold)`, branchless: build the mask with `std::transform` or a tight `for` loop, `mask[i] = col[i] > threshold`
- [ ] `include/query_exec/column_project.hpp` + `src/column_project.cpp`: `std::vector<int32_t> project(const std::vector<int32_t>& col, const std::vector<bool>& mask)`, collect values where the mask is true
- [ ] `include/query_exec/hash_join.hpp` + `src/hash_join.cpp`: `std::vector<std::pair<int32_t, int32_t>> join(const std::vector<int32_t>& left_key, const std::vector<int32_t>& left_val, const std::vector<int32_t>& right_key, const std::vector<int32_t>& right_val)`, build a `std::unordered_map<int32_t, int32_t>` from the left side, probe with the right side, emit matching pairs
- [ ] `benchmark/benchmark.cpp`: generate 1M random rows; time the filter + project pipeline 10 times each using `std::chrono::high_resolution_clock` (3 warm-up runs first, so caches and branch prediction are primed); print mean throughput in M rows/sec for both executors. Build and run in Release mode:
  ```bash
  cmake -S . -B build -DCMAKE_BUILD_TYPE=Release
  cmake --build build
  ./build/benchmark
  ```
  **The Release build type is not optional here**, same reason `cargo run --release` mattered in the original plan. A default CMake build has no `-O2`/`-O3` and minimal inlining; an unoptimized debug build will make the comparison meaningless.

**Expected outcome:** vectorized should be 3–8x faster. If the gap is smaller, the first thing to check is whether you actually configured `-DCMAKE_BUILD_TYPE=Release`; that's the single most common reason a C++ benchmark looks flat.

---

## 🐍 Python DSA Review (optional)

**Hash join + binary search on sorted arrays**: the two algorithms your C++ `hash_join.cpp` and `column_filter.cpp` implement. Python makes the probe/build logic easy to inspect.

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

# binary_filter.py: binary search on a sorted column (like column_filter.cpp)
def filter_sorted_col(col: list[int], predicate_min: int, predicate_max: int) -> list[int]:
    lo = bisect_left(col, predicate_min)
    hi = bisect_left(col, predicate_max + 1)
    return col[lo:hi]  # slice is O(k) not O(n)

col = sorted([5, 2, 8, 1, 9, 3, 7, 4, 6])
assert filter_sorted_col(col, 3, 7) == [3, 4, 5, 6, 7]
```

**Connection:** `hash_join.cpp` is the build+probe pattern in C++ over `std::vector<int32_t>` columns. `column_filter.cpp` can use `std::lower_bound`/`std::upper_bound` (C++'s `bisect_left` equivalent) for range filters over a sorted column, 3–8x faster than scanning.

---

## Reflect

**What clicked:**

**What surprised me:**

**Benchmark results:**
- Row-at-a-time: __ M rows/sec
- Vectorized: __ M rows/sec
- Speedup: __x

**What does this tell you about how query execution works in a system you know?**

**Where did mutation buy you speed?** Your vectorized executor builds and fills `std::vector`s directly for performance. Notice this is actually less of a tension in C++ than it would have been in Rust: nothing here forced immutability on you the way the borrow checker did in W05 and W07, so this trade-off was always available and always invisible until you looked for it. Point to one specific place in `column_filter.cpp` or `hash_join.cpp` where you mutated a vector in place (`push_back` into a pre-sized buffer, writing through an index, etc.) and estimate what it would've cost you in speed to instead allocate a fresh vector per step and chain functional-style transforms, the way `map`/`filter` in W07 do by construction.

**What I'd do differently:**
