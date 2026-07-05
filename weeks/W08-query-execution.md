---
week_number: 8
status: not-started
---

# W08 — Query Execution

> **Arc:** Streaming and Dataflow · **Language:** Scala

## What you'll build
A vectorized query executor in Scala: columnar filter + hash join + projection over in-memory data. Benchmark it against a row-at-a-time version of the same pipeline and measure the speedup.

---

## Read
- [ ] [Volcano — An Extensible and Parallel Query Evaluation System](https://dl.acm.org/doi/10.1109/69.273032) (Graefe, 1994) — read Sections 1–3. This defines the iterator model (the `next()` interface) that every query engine for 20 years was built on.
- [ ] [MonetDB/X100: Hyper-Pipelining Query Execution](https://www.cidrdb.org/cidr2005/papers/P19.pdf) (Boncz et al., CIDR 2005) — read Sections 1–3. This is the argument for vectorized execution and why Volcano is CPU-cache unfriendly.

**Key question:** Why does calling `next()` once per row hurt CPU performance even when the logic is simple? What does processing a batch of 1024 rows at a time fix?

---

## Code

Project: `code/query-exec/` (Scala 3, sbt)

Data model: a table of 1M rows with columns `id: Array[Int]`, `dept: Array[Int]`, `salary: Array[Int]` stored separately (columnar).

**Row-at-a-time executor (baseline):**

- [ ] `RowExecutor.scala` — case class `Row(id: Int, dept: Int, salary: Int)`; `Iterator[Row]` that zips the three arrays; apply `filter(_.salary > threshold)`, then `map(r => (r.id, r.salary))`; collect results

**Vectorized executor:**

- [ ] `ColumnFilter.scala` — `def filter(col: Array[Int], threshold: Int): Array[Boolean]` — branchless: `col.map(v => v > threshold)`
- [ ] `ColumnProject.scala` — `def project(col: Array[Int], mask: Array[Boolean]): Array[Int]` — collect values where mask is true
- [ ] `HashJoin.scala` — `def join(leftKey: Array[Int], leftVal: Array[Int], rightKey: Array[Int], rightVal: Array[Int]): Array[(Int, Int)]` — build `HashMap[Int, Int]` from left side, probe with right side, emit matching pairs
- [ ] `Benchmark.scala` — generate 1M random rows; time filter + project pipeline 10 times each (warm up JVM first with 3 dry runs); print mean throughput in M rows/sec for both executors

**Expected outcome:** vectorized should be 3–8x faster. If the gap is smaller, check whether the JIT is auto-vectorizing the iterator version (run with `-XX:+PrintCompilation` to inspect).

---

## 🐍 Python DSA Review (optional)

**Hash join + binary search on sorted arrays** — the two algorithms your Scala `HashJoin.scala` and `ColumnFilter.scala` implement. Python makes the probe/build logic easy to inspect.

```python
from collections import defaultdict
from bisect import bisect_left

# hash_join.py — classic hash join: build phase + probe phase
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

# binary_filter.py — binary search on a sorted column (like ColumnFilter.scala)
def filter_sorted_col(col: list[int], predicate_min: int, predicate_max: int) -> list[int]:
    lo = bisect_left(col, predicate_min)
    hi = bisect_left(col, predicate_max + 1)
    return col[lo:hi]  # slice is O(k) not O(n)

col = sorted([5, 2, 8, 1, 9, 3, 7, 4, 6])
assert filter_sorted_col(col, 3, 7) == [3, 4, 5, 6, 7]
```

**Connection:** `HashJoin.scala` is the build+probe pattern in Scala with `Array[Int]` columns. `ColumnFilter.scala` can use `bisect_left` equivalent for range filters on sorted columns — 3–8x faster than scanning.

---

## Reflect

**What clicked:**

**What surprised me:**

**Benchmark results:**
- Row-at-a-time: __ M rows/sec
- Vectorized: __ M rows/sec
- Speedup: __x

**What does this tell you about how query execution works in your current role?**

**What I'd do differently:**
