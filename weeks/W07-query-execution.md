---
week_number: 7
status: not-started
---

# W07: Query Execution

> **Arc:** Streaming and Dataflow · **Language:** Go
> **Budget:** about 5 hours. Hit the Minimum bar first; everything past it is optional.

## What you'll build
A vectorized query executor in Go: columnar filter + hash join + projection over in-memory data. Benchmark it against a row-at-a-time version of the same pipeline and measure the speedup.

**Scenario:** the 3-8x speedup below is measured against one workload shape. Ship this benchmark's conclusion to production unexamined and the first coworker who runs a highly selective filter, or a join where one side doesn't fit in memory, will find the exact place it stops holding.

**Note on why Go, specifically, for this one week:** the rest of this arc is Java, but this week's benchmark is memory-layout-sensitive in a way the others aren't, and Go gives you two concrete things Java can't here. First, Go compiles ahead of time to native code; there's no JIT to warm up, so a naive, hand-timed benchmark (which is this week's style, no microbenchmark harness) measures real steady-state performance from the first call, instead of risking measuring JIT compilation overhead the way an unwarmed Java benchmark would. Second, and more important for a columnar engine specifically: Go structs are real value types in slices, `[]Row` is genuinely contiguous memory. Java has no true value types (records are still heap-allocated objects), so an array of row structs in Java is an array of pointers to scattered allocations, exactly the pointer-chasing a columnar query engine exists to avoid. The whole point of this week is that vectorized, columnar execution beats naive row-at-a-time processing because of memory layout; Go gets you closer to that lesson honestly than Java would.

---

## Read
- [ ] [Volcano, An Extensible and Parallel Query Evaluation System](https://dl.acm.org/doi/10.1109/69.273032) (Graefe, 1994): read Sections 1–3. This defines the iterator model (the `next()` interface) that every query engine for 20 years was built on.
- [ ] [MonetDB/X100: Hyper-Pipelining Query Execution](https://www.cidrdb.org/cidr2005/papers/P19.pdf) (Boncz et al., CIDR 2005): read Sections 1–3. This is the argument for vectorized execution and why Volcano is CPU-cache unfriendly.
- [ ] Optional: [Dremel: Interactive Analysis of Web-Scale Datasets](https://research.google/pubs/dremel-interactive-analysis-of-web-scale-datasets/) (Melnik et al., Google, VLDB 2010): read Sections 1–3. Same columnar-storage instinct as MonetDB/X100, but scaled a level up: Dremel shreds nested records into columns (the ancestor of Parquet's on-disk format) and spreads the aggregation itself across a multi-level serving tree of thousands of machines. Read it for what changes when "vectorize the scan" becomes "vectorize the scan, then fan the aggregation out across a cluster."
- [ ] Optional: [DuckDB execution engine source](https://github.com/duckdb/duckdb/tree/main/src/execution): optional but worth it: a real, actively maintained vectorized query engine in C++ (a different language than this week's build, the lesson is the technique, not the syntax), and one you already depend on via W09's feature store. Skim `PhysicalFilter` and how DuckDB batches rows into `DataChunk`s; that's the production version of what you're building this week.
- [ ] Optional, context only (a free public blog post, not something you install or test against): [Announcing Photon](https://www.databricks.com/blog/2021/06/17/announcing-photon-public-preview-the-next-generation-query-engine-on-the-databricks-lakehouse-platform.html). Photon is "written from the ground up in C++" specifically to replace the JVM-based Spark execution engine for exactly this reason: columnar batches, tight vectorized loops, SIMD, none of it playing well with a garbage collector or a heap of boxed objects. Your actual hands-on comparison this week is DuckDB (and optionally ClickHouse) below; this is just confirmation the same technique is load-bearing in production, regardless of which non-GC'd language a given engine picks.
- [ ] Optional: [ClickHouse execution pipeline source](https://github.com/ClickHouse/ClickHouse/tree/master/src/Processors): optional: ClickHouse is C++ end to end; skim `IProcessor` and how the pull-based pipeline batches rows into `Chunk`s. A second real reference point alongside DuckDB and Photon, from a different target company with a different pipeline design.

**Depth: study Sections 1 to 3 of MonetDB/X100.** It contains the argument your benchmark is about to either confirm or fail to reproduce, which makes it worth real attention. Volcano is a read. Dremel, DuckDB, Photon, and ClickHouse are all skims and all optional.

**Key question:** Why does calling `next()` once per row hurt CPU performance even when the logic is simple? What does processing a batch of 1024 rows at a time fix?

---

## Code

Project: `code/query-exec/` (Go modules)

Data model: a table of 1M rows with columns `[]int32` for `id`, `dept`, `salary`, stored separately (columnar).

**Row-at-a-time executor (baseline):**

- [ ] `row_executor.go`: `type Row struct { ID, Dept, Salary int32 }`; build rows by iterating the three slices in lockstep (a plain indexed `for` loop is clearest here); apply a filter predicate (`salary > threshold`), then project `(id, salary)`; collect results into a `[]Row` or a `[]struct{ ID, Salary int32 }`, either way a slice of actual structs, not pointers to them

**Vectorized executor:**

- [ ] `column_filter.go`: `func Filter(col []int32, threshold int32) []bool`, branchless: build the mask with a tight `for` loop, `mask[i] = col[i] > threshold`
- [ ] `column_project.go`: `func Project(col []int32, mask []bool) []int32`, collect values where the mask is true
- [ ] `hash_join.go`: `func Join(leftKey, leftVal, rightKey, rightVal []int32) []struct{ Left, Right int32 }`, build a `map[int32]int32` from the left side, probe with the right side, emit matching pairs
- [ ] `benchmark_test.go`: use Go's `testing.B` (`go test -bench=. -benchmem`) to time the filter + project pipeline for both executors over 1M random rows; `testing.B` runs enough iterations to get a stable number and reports allocations per operation, worth checking that neither executor is allocating inside its hot loop.

**Expected outcome:** vectorized should be 3–8x faster. If the gap is smaller, check `-benchmem`'s allocation count first, an unexpected allocation inside the loop (for example, `append` triggering a slice reallocation because the destination wasn't pre-sized with `make([]int32, 0, n)`) is the most common reason a Go benchmark like this looks flatter than expected.

**Minimum bar:** the vectorized pipeline beats the row-at-a-time one on the same data, you have the measured speedup, and you can explain the gap in terms of memory layout rather than instruction count. The optional source reading is genuinely optional.

**Break it, then decide:**
- [ ] Re-run the filter+project benchmark twice more: once with a threshold so selective that under 1% of rows pass, once with a threshold near the middle so roughly 50% pass. Compare `-benchmem`'s bytes-per-op and allocs-per-op across all three selectivities, not just the wall-clock number. If `Project` always allocates its output slice sized to the full input length regardless of how many rows actually pass the mask, that's real wasted memory at high selectivity, invisible if you only ever benchmarked one selectivity and called it done.
- [ ] `hash_join.go` builds its hash table from the entire left side before probing. That's fine at 1M rows in memory; it stops being fine the moment the left side is bigger than RAM. Would you keep hash join and accept that limit, or switch to a sort-merge join (sort both sides, then merge in one linear pass, no full in-memory table required, but now you're paying for two sorts)? There's a real, workload-dependent answer; give yours and say what property of your data would have to change to flip it.

---

## Rehearse it in Python first (optional, 20 minutes)

**Hash join + binary search on sorted arrays**: the two algorithms your Go `hash_join.go` and `column_filter.go` implement. Python makes the probe/build logic easy to inspect.

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

# binary_filter.py: binary search on a sorted column (like column_filter.go)
def filter_sorted_col(col: list[int], predicate_min: int, predicate_max: int) -> list[int]:
    lo = bisect_left(col, predicate_min)
    hi = bisect_left(col, predicate_max + 1)
    return col[lo:hi]  # slice is O(k) not O(n)

col = sorted([5, 2, 8, 1, 9, 3, 7, 4, 6])
assert filter_sorted_col(col, 3, 7) == [3, 4, 5, 6, 7]
```

**Connection:** `hash_join.go` is the build+probe pattern in Go over `[]int32` columns. `column_filter.go` can use `sort.Search` (Go's `bisect_left` equivalent) for range filters over a sorted column, 3–8x faster than scanning.

---

## Reflect

**What clicked:**

**What surprised me:**

**Benchmark results:**
- Row-at-a-time: __ M rows/sec
- Vectorized: __ M rows/sec
- Speedup: __x

**What does this tell you about how query execution works in a system you know?**

**Allocation behavior across selectivities, and hash join vs. sort-merge for a left side bigger than memory (from Break it, then decide above):**

**Where did a value-type slice buy you speed?** Your vectorized executor builds and fills `[]int32` slices directly, real contiguous memory, not a slice of pointers to scattered structs. Notice that Go never forces immutability on you the way W04 and W06's Java code does by construction (`filter`/`consolidate` returning new values), so this trade-off was always available and always invisible until you looked for it. Point to one specific place in `column_filter.go` or `hash_join.go` where you wrote into a pre-sized slice by index instead of building a fresh one and estimate what it would've cost you in speed to instead allocate a new slice per step and chain functional-style transforms, the way `filter`/`consolidate` in W06 do by construction.

**What I'd do differently:**
