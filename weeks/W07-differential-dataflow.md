---
week_number: 7
status: not-started
---

# W07: Differential Dataflow and Incremental View Maintenance

> **Arc:** Streaming and Dataflow · **Language:** C++

## What you'll build
Two parts this week, deliberately shorter on the from-scratch build than the rest of the arc. Part 1 (1–2 days): the core Differential Dataflow data model in C++: `(key, value, time, diff)` updates and a `map`/`filter`/`consolidate` `Collection`, applied to incremental word count. Part 2 (remaining days): the same "orders → revenue per region" incremental-view problem built four ways: your own C++ benchmark, a real local ClickHouse materialized view, and a real local Spark Structured Streaming stateful aggregation, so the comparison is against systems you actually install and run yourself, not vendor documentation.

---

## Part 1: Differential Dataflow Core (1–2 days)

### Read
- [ ] [Differential Dataflow](https://www.cidrdb.org/cidr2013/Papers/CIDR13_Paper111.pdf) (McSherry et al., CIDR 2013): read Sections 1–2 only this time. Section 2 defines the data model (collections as functions from time to multisets of changes); that's the part this week actually builds. Section 3 (operators) is optional if you want the full picture.

**Key question:** What is a "difference" in DD? How does `(key, value, time, diff)` encode both additions and retractions?

### Code

Project: `code/dd-scratch/` (C++, CMake, header-only where templated)

- [ ] `include/dd_scratch/update.hpp`: `template <typename K, typename V> struct Update { K key; V value; int32_t time; int32_t diff; };` where `diff` is `+1` (addition) or `-1` (retraction)
- [ ] `include/dd_scratch/collection.hpp`: `template <typename K, typename V> class Collection { std::vector<Update<K, V>> updates_; public: ... };` implements:
  - `Collection<K, V> filter(std::function<bool(const K&, const V&)> p) const`: drop updates where the predicate is false
  - `Collection<K, V> consolidate() const`: merge updates with the same (key, value, time) via an `unordered_map`, sum their diffs, drop zero-diff entries
- [ ] `src/word_count.cpp`: given a `Collection<int32_t, std::string>` (document id to document text): flat-map each document into per-word updates, consolidate, group by key and sum diffs to get current count per word. In a loop: add a document at t=1 (diff=+1), print counts; retract it at t=2 (diff=-1), print updated counts. Only print the delta each round, not the full state.

**Constraints:** zero external dependencies beyond the standard library. `filter`/`consolidate` are `const` methods that return a new `Collection` rather than mutating `*this`. Nothing in C++ enforces that the way the borrow checker would have; treat it as a promise you're keeping by discipline, and notice if you break it.

---

## Part 2: Incremental Materialized Views, Tested Against Real Local Systems (remaining days)

The "compute only the delta" idea from Part 1 is the same idea behind every production incremental-refresh system; they just implement it very differently from each other and from DD. Rather than reading vendor documentation and taking that on faith, this part has you install two real open-source systems locally and watch each one handle the same tiny dataset, so the comparison is evidence you gathered yourself, not marketing copy.

**Note on scope:** Snowflake Dynamic Tables is arguably the closest production system to true DD-style incremental view maintenance, but it's closed-source SaaS with no self-hosted option. There's no way to install it locally and watch it work, so it's excluded from the required exercise below for the same reason ClickHouse would have been excluded if it weren't locally installable. If you want the context anyway: [Snowflake Dynamic Tables docs](https://docs.snowflake.com/en/user-guide/dynamic-tables-about), optional, not required.

### Read
- [ ] [ClickHouse Materialized Views](https://clickhouse.com/docs/en/guides/developer/cascading-materialized-views): a ClickHouse materialized view is an insert trigger: it transforms rows as they land and writes the result to a target table. It does not track retractions or recompute deltas against existing state the way DD does. Read this before the exercise below so you know what you're about to watch happen.
- [ ] [Spark Structured Streaming: stateful operations](https://spark.apache.org/docs/latest/structured-streaming-programming-guide.html#arbitrary-stateful-operations): read the sections on stateful aggregation and the state store. Spark keeps per-key aggregation state between micro-batches and updates it incrementally as new data arrives, rather than rescanning history each time.
- [ ] Optional, stretch: [pg_ivm](https://github.com/sraoss/pg_ivm): a real, actively maintained PostgreSQL extension implementing true incremental view maintenance, closer in spirit to DD than either system below (it tracks base-table changes and recomputes only the affected view rows, retractions included). Left optional because it requires building a Postgres extension from source via PGXS (`make && make install` against your local `pg_config`) rather than a single-binary or `pip install`; real, but more setup friction than the two required systems.

**Key question:** Of ClickHouse and Spark's stateful aggregation, which one can retract a previously emitted result if an earlier input turns out to be wrong or late, and which can only append forward? What does that distinction cost or buy each of them?

### Code

Project: `code/dd-scratch/` (same C++/CMake project as Part 1), plus a new `code/dd-scratch/comparisons/` directory for the two external systems.

Data model: a toy `orders` table (`struct Order { int32_t order_id; int32_t region_id; int32_t amount; };` in C++, mirrored as `(order_id, region_id, amount)` rows everywhere else) and a materialized view: total revenue per region. Same shape across all three systems, so the comparison is apples-to-apples.

**1. C++ (your own implementation):**
- [ ] `include/dd_scratch/full_recompute_view.hpp` + `src/full_recompute_view.cpp`: `class FullRecomputeView` holds all orders in a `std::vector<Order>`; on `apply(const Order& new_order)`, appends it and recomputes the entire per-region revenue map from scratch by rescanning every order
- [ ] `include/dd_scratch/materialized_view.hpp` + `src/materialized_view.cpp`: `class IncrementalAggregateView` reuses `Update<int32_t, int32_t>` (region_id, amount) from Part 1: on `apply(const Order& new_order)`, updates only the affected region's running total in a `std::unordered_map<int32_t, int64_t>`, no rescan
- [ ] `benchmark/mv_benchmark.cpp`: seed both views with a growing base of orders (1k, 10k, 100k), then apply a stream of new single-order updates to each; time `apply()` for both. Release build required, same as W08.

**2. ClickHouse (local server, single machine, no account):**
- [ ] Install: `brew install clickhouse`; run `clickhouse server` in one terminal, `clickhouse client` in another
- [ ] `code/dd-scratch/comparisons/clickhouse_mv.sql`: create an `orders` source table, a `region_revenue` target table (`SummingMergeTree` on `region_id`), and a materialized view that sums `amount` into `region_revenue` on every insert into `orders`. Insert a handful of orders, check `region_revenue`, then insert one more order and confirm, via a plain `SELECT` or the query log, that ClickHouse updated the target table without rescanning `orders`.

**3. Spark Structured Streaming (local mode, single machine):**
- [ ] Install: `pip install pyspark` (you already have Python 3.11 set up; Spark itself runs on the JVM regardless of the Python surface, so you'll also need `brew install openjdk@17` if you don't have a JDK)
- [ ] `code/dd-scratch/comparisons/spark_stateful_agg.py`: a local Structured Streaming job (`master("local[*]")`) that reads order events as small JSON files dropped into a watched directory (`readStream.format("json")`), runs `groupBy("region_id").sum("amount")` with `outputMode("update")`, and writes to the console sink. Drop one new order file mid-run and confirm the console output shows only the affected region's updated row, not every group re-emitted.

**Minimum bar:** for both ClickHouse and Spark, one piece of evidence you observed yourself (a query result, a log line, or the update-mode console output) showing the system updated only the affected slice of state rather than the whole dataset, the same property your C++ `IncrementalAggregateView` has and `FullRecomputeView` doesn't.

---

## 🐍 Python DSA Review (optional)

**defaultdict as a multiset + sorted consolidation**: a DD `Collection` is a map from keys to signed integer multiplicities. Consolidation collapses duplicates and drops zeros; the same operation `IncrementalAggregateView` does per-region in Part 2.

```python
from collections import defaultdict

# consolidate.py: core operation in every DD collection
def consolidate(updates: list[tuple]) -> dict:
    """
    updates: list of (key, diff) where diff is +1 (add) or -1 (retract)
    Returns: dict of key to net_diff, with zeros removed
    """
    counts: dict = defaultdict(int)
    for key, diff in updates:
        counts[key] += diff
    return {k: v for k, v in counts.items() if v != 0}

# Test: add 3 "apple", retract 2, net +1
updates = [("apple", 1), ("apple", 1), ("apple", 1), ("apple", -1), ("apple", -1),
           ("banana", 1), ("banana", -1)]  # banana nets to 0, dropped
result = consolidate(updates)
assert result == {"apple": 1}
```

**Connection:** `Collection::consolidate()` in Part 1 and `IncrementalAggregateView::apply()` in Part 2 are both instances of this same operation, just at different scales: one against a synthetic word stream, one against something shaped like a real materialized view.

---

## Reflect

**What clicked:**

**What surprised me:**

**Did `const` actually hold in Part 1, or did you catch yourself mutating `updates_` somewhere it shouldn't have?**

**Your measured numbers (C++):**
- Full recompute latency at 1k / 10k / 100k orders: __ / __ / __
- Incremental latency at 1k / 10k / 100k orders: __ / __ / __

**What you observed in ClickHouse and Spark:**
- ClickHouse: what specifically told you the target table was updated incrementally, not rescanned?
- Spark: what did the console sink show after you dropped the new order file? Did the whole aggregation re-run, or just the affected group?

**Rank your four implementations (C++ `IncrementalAggregateView`, ClickHouse MV, Spark stateful aggregation, and pg_ivm if you did the stretch) from closest to furthest from true DD semantics (tracks retractions, recomputes only the affected delta). Where does each one cut a corner, and why might that corner-cutting be the right engineering trade-off for that system's actual use case?**

**What I'd do differently:**
