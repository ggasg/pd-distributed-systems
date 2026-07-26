---
week_number: 7
status: not-started
---

# W07: Differential Dataflow and Incremental View Maintenance

> **Arc:** Streaming and Dataflow · **Language:** Java

## What you'll build
Two parts this week, deliberately shorter on the from-scratch build than the rest of the arc. Part 1 (1–2 days): the core Differential Dataflow data model in Java: `(key, value, time, diff)` updates and a `map`/`filter`/`consolidate` `Collection`, applied to incremental word count. Part 2 (remaining days): the same "orders → revenue per region" incremental-view problem built four ways: your own Java benchmark, a real local ClickHouse materialized view, and a real local Spark Structured Streaming stateful aggregation, so the comparison is against systems you actually install and run yourself, not vendor documentation.

---

## Part 1: Differential Dataflow Core (1–2 days)

### Read
- [ ] [Differential Dataflow](https://www.cidrdb.org/cidr2013/Papers/CIDR13_Paper111.pdf) (McSherry et al., CIDR 2013): read Sections 1–2 only this time. Section 2 defines the data model (collections as functions from time to multisets of changes); that's the part this week actually builds. Section 3 (operators) is optional if you want the full picture.
- [ ] Optional: [Large-scale Incremental Processing Using Distributed Transactions and Notifications](https://research.google/pubs/large-scale-incremental-processing-using-distributed-transactions-and-notifications/) (Peng & Dabek, Google, OSDI 2010, the Percolator paper): a completely different mechanism for the same underlying problem, incremental computation instead of batch recompute. Where DD tracks deltas through an explicit `(key, value, time, diff)` model, Percolator gets there by layering distributed transactions and observer-style notifications on top of Bigtable, triggering downstream work when a row changes rather than recomputing anything. Google used it to replace a MapReduce-based web indexing pipeline; worth reading right after the DD paper to see a second, production-scale answer to the same question this week's `Collection` class answers in miniature.

**Key question:** What is a "difference" in DD? How does `(key, value, time, diff)` encode both additions and retractions?

### Code

Project: `code/dd-scratch/` (Java 21, Maven)

- [ ] `Update.java`: `record Update<K, V>(K key, V value, int time, int diff) {}` where `diff` is `+1` (addition) or `-1` (retraction). A generic record here is a straightforward, mature fit, Java's generics have handled parameterized types like this since Java 5, nothing experimental about reaching for one.
- [ ] `Collection.java`: `class Collection<K, V> { private final List<Update<K, V>> updates; ... }` implements:
  - `Collection<K, V> filter(BiPredicate<K, V> p)`: drop updates where the predicate is false, return a new `Collection`
  - `Collection<K, V> consolidate()`: merge updates with the same (key, value, time) via a `HashMap`, sum their diffs, drop zero-diff entries, return a new `Collection`
- [ ] `WordCount.java`: given a `Collection<Integer, String>` (document id to document text): flat-map each document into per-word updates, consolidate, group by key and sum diffs to get current count per word. In a loop: add a document at t=1 (diff=+1), print counts; retract it at t=2 (diff=-1), print updated counts. Only print the delta each round, not the full state.

**Constraints:** JDK standard library only. `filter`/`consolidate` return a new `Collection` rather than mutating the receiver; keep the backing `List` `private final` and never expose it directly, so nothing outside the class can mutate `updates` behind your back.

**Break it, then decide:** if any method on `Collection` ever returns `updates` directly (even just for a test assertion, `assertEquals(expected, coll.getUpdates())`), the caller now holds a live reference to your "immutable" list. Mutate it from outside (`coll.getUpdates().add(...)`) and confirm the `Collection` you thought was safe just changed under you, silently, no exception. `private final` on the field only stops reassignment of the reference; it does nothing to stop the object the reference points to from being mutated. Decide: would you fix this by having any accessor return `List.copyOf(updates)` (a real copy, small ongoing cost) or by simply never writing an accessor that exposes the raw list in the first place (zero cost, but only as safe as every future line of code that touches this class)? Pick one and say why it's the right trade-off for a class this size.

One simplification worth naming, not replicating: `filter`/`consolidate` here compute and return a fully materialized new `Collection` immediately, so this is eager function composition, not a lazy dataflow graph. A real Naiad/DD program builds the operator graph first (map, filter, and consolidate all wired together as nodes) and only pushes data through it once, when the computation actually runs; nothing computes at graph-construction time. This toy version skips that indirection so the DD-specific algebra (the diff-based data model) stays the focus of the week, but it's worth knowing the real thing works differently before you assume this `Collection` API is a faithful model of how Naiad or DD actually execute.

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

Project: `code/dd-scratch/` (same Java/Maven project as Part 1), plus a new `code/dd-scratch/comparisons/` directory for the two external systems.

Data model: a toy `orders` table (`record Order(int orderId, int regionId, int amount) {}` in Java, mirrored as `(order_id, region_id, amount)` rows everywhere else) and a materialized view: total revenue per region. Same shape across all three systems, so the comparison is apples-to-apples.

**1. Java (your own implementation):**
- [ ] `FullRecomputeView.java`: `class FullRecomputeView` holds all orders in a `List<Order>`; on `apply(Order newOrder)`, appends it and recomputes the entire per-region revenue map from scratch by rescanning every order
- [ ] `IncrementalAggregateView.java`: reuses `Update<Integer, Integer>` (regionId, amount) from Part 1: on `apply(Order newOrder)`, updates only the affected region's running total in a `Map<Integer, Long>`, no rescan
- [ ] `MvBenchmark.java`: seed both views with a growing base of orders (1k, 10k, 100k), then apply a stream of new single-order updates to each; time `apply()` for both with `System.nanoTime()`, after a handful of warm-up calls so you're not measuring JIT compilation on top of the algorithm, run each size enough times to get a stable number.

**2. ClickHouse (local server, single machine, no account):**
- [ ] Install: `brew install clickhouse`; run `clickhouse server` in one terminal, `clickhouse client` in another
- [ ] `code/dd-scratch/comparisons/clickhouse_mv.sql`: create an `orders` source table, a `region_revenue` target table (`SummingMergeTree` on `region_id`), and a materialized view that sums `amount` into `region_revenue` on every insert into `orders`. Insert a handful of orders, check `region_revenue`, then insert one more order and confirm, via a plain `SELECT` or the query log, that ClickHouse updated the target table without rescanning `orders`.

**3. Spark Structured Streaming (local mode, single machine):**
- [ ] Install: `pip install pyspark` (you already have Python 3.11 set up; Spark itself runs on the JVM regardless of the Python surface, so you'll also need `brew install openjdk@17` if you don't have a JDK)
- [ ] `code/dd-scratch/comparisons/spark_stateful_agg.py`: a local Structured Streaming job (`master("local[*]")`) that reads order events as small JSON files dropped into a watched directory (`readStream.format("json")`), runs `groupBy("region_id").sum("amount")` with `outputMode("update")`, and writes to the console sink. Drop one new order file mid-run and confirm the console output shows only the affected region's updated row, not every group re-emitted.

**Minimum bar:** for both ClickHouse and Spark, one piece of evidence you observed yourself (a query result, a log line, or the update-mode console output) showing the system updated only the affected slice of state rather than the whole dataset, the same property your Java `IncrementalAggregateView` has and `FullRecomputeView` doesn't.

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

**Connection:** `Collection.consolidate()` in Part 1 and `IncrementalAggregateView.apply()` in Part 2 are both instances of this same operation, just at different scales: one against a synthetic word stream, one against something shaped like a real materialized view.

---

## Reflect

**What clicked:**

**What surprised me:**

**Did `updates` in `Collection` actually stay unmutated through Part 1, or did you catch yourself reaching in and changing it somewhere it shouldn't have?**

**`List.copyOf` on every access, or just never writing an accessor that leaks the raw list, and why (from Break it, then decide above):**

**Your measured numbers (Java):**
- Full recompute latency at 1k / 10k / 100k orders: __ / __ / __
- Incremental latency at 1k / 10k / 100k orders: __ / __ / __

**What you observed in ClickHouse and Spark:**
- ClickHouse: what specifically told you the target table was updated incrementally, not rescanned?
- Spark: what did the console sink show after you dropped the new order file? Did the whole aggregation re-run, or just the affected group?
- Nothing happens when `spark_stateful_agg.py` calls `df.groupBy("region_id").sum("amount")`; the actual streaming job only starts when you call `.start()` on the writer afterward. What does Spark need to see across the whole pipeline before it can decide how to execute it, and what would it lose if `groupBy` ran eagerly the moment you called it?

**Rank your four implementations (Java `IncrementalAggregateView`, ClickHouse MV, Spark stateful aggregation, and pg_ivm if you did the stretch) from closest to furthest from true DD semantics (tracks retractions, recomputes only the affected delta). Where does each one cut a corner, and why might that corner-cutting be the right engineering trade-off for that system's actual use case?**

**What I'd do differently:**
