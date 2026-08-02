---
week_number: 7
status: not-started
---

# W07: Differential Dataflow and Incremental View Maintenance

> **Arc:** Streaming and Dataflow · **Language:** Java

## What you'll build
Two parts this week, and the weight sits firmly on the second. Part 1 (1 day): read and exercise a provided Differential Dataflow core in Java, `(key, value, time, diff)` updates and a `filter`/`consolidate` `Collection`, so you understand the diff-based data model well enough to use it. Part 2 (the rest of the week): the same "orders to revenue per region" incremental-view problem built three ways, your own Java benchmark, a real local ClickHouse materialized view, and a real local Spark Structured Streaming stateful aggregation, so the comparison is against systems you actually install and run yourself, not vendor documentation.

The balance is deliberate. Incremental view maintenance is a technique you will meet again in ClickHouse, Flink, dbt, and every streaming feature store, and being able to tell which system genuinely recomputes deltas and which only appends forward is the durable skill here. Reimplementing Differential Dataflow's specific algebra from scratch is a narrower payoff, so Part 1 hands you the code and spends its day making sure you can read it.

---

## Part 1: Differential Dataflow Core (1 day)

### Read
- [ ] [Differential Dataflow](https://www.cidrdb.org/cidr2013/Papers/CIDR13_Paper111.pdf) (McSherry et al., CIDR 2013): read Sections 1–2 only this time. Section 2 defines the data model (collections as functions from time to multisets of changes); that's the part this week actually builds. Section 3 (operators) is optional if you want the full picture.
- [ ] Optional: [Large-scale Incremental Processing Using Distributed Transactions and Notifications](https://research.google/pubs/large-scale-incremental-processing-using-distributed-transactions-and-notifications/) (Peng & Dabek, Google, OSDI 2010, the Percolator paper): a completely different mechanism for the same underlying problem, incremental computation instead of batch recompute. Where DD tracks deltas through an explicit `(key, value, time, diff)` model, Percolator gets there by layering distributed transactions and observer-style notifications on top of Bigtable, triggering downstream work when a row changes rather than recomputing anything. Google used it to replace a MapReduce-based web indexing pipeline; worth reading right after the DD paper to see a second, production-scale answer to the same question this week's `Collection` class answers in miniature.

**Key question:** What is a "difference" in DD? How does `(key, value, time, diff)` encode both additions and retractions?

### Code

Project: `code/dd-scratch/` (Java 21, Maven)

**Given, not built:** `Update.java` and `Collection.java` are both provided as starter files, already implemented and tested.

- `Update.java`: `record Update<K, V>(K key, V value, int time, int diff) {}` where `diff` is `+1` (an addition) or `-1` (a retraction). That single signed integer is the whole trick: a retraction is not a delete, it is a negative-weighted addition, which means merging changes is just arithmetic.
- `Collection.java`: `class Collection<K, V>` wrapping a `private final List<Update<K, V>>`, with `filter(BiPredicate<K, V>)` and `consolidate()`. `consolidate()` merges updates sharing the same (key, value, time) via a `HashMap`, sums their diffs, and drops anything that nets to zero. Both methods return a new `Collection` rather than mutating the receiver.

Read both files once before writing anything. The thing to look for is why `consolidate()` can drop a zero-diff entry entirely: an add and a retract of the same row cancel exactly, so the row simply stops existing in the output with no special-case code anywhere.

- [ ] `WordCount.java`: this is your build for Part 1, and it is small on purpose. Given a `Collection<Integer, String>` (document id to document text), flat-map each document into per-word updates, consolidate, then group by key and sum diffs to get the current count per word. In a loop: add a document at t=1 with diff=+1 and print counts, then retract it at t=2 with diff=-1 and print the updated counts. Print only the delta each round, not the full state, because printing the full state would quietly hide whether your incremental path is doing anything at all.

**Constraints:** JDK standard library only. Do not modify the two provided files.

**Break it, then decide:** add a method to `Collection` that returns `updates` directly, the way you would if a test wanted to assert on it (`assertEquals(expected, coll.getUpdates())`). The caller now holds a live reference to your supposedly immutable list. Mutate it from outside with `coll.getUpdates().add(...)` and confirm the `Collection` you thought was safe just changed underneath you, silently, with no exception thrown. `private final` on the field only stops the reference from being reassigned; it does nothing to stop the list it points at from being mutated. This surprises most people once and then never again. Decide how you'd fix it: return `List.copyOf(updates)` from any accessor, which is a real copy with a small ongoing cost, or never write an accessor that exposes the raw list at all, which is free but only as safe as every future line of code someone adds to this class. Pick one and say why it's right for a class this size.

One simplification worth naming, not replicating: `filter` and `consolidate` here compute and return a fully materialized new `Collection` immediately, so this is eager function composition, not a lazy dataflow graph. A real Differential Dataflow program builds the operator graph first, with map, filter, and consolidate all wired together as nodes, and only pushes data through it once, when the computation actually runs. Nothing computes at graph-construction time. The provided version skips that indirection so the diff-based data model stays the focus, but it is worth knowing the real thing works differently before you assume this `Collection` API is a faithful model of how DD actually executes.

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
- [ ] Install: `pip install "pyspark==4.2.0"` (check for a newer 4.x). Spark runs on the JVM regardless of the Python surface, but the Java 21 you installed for this arc already satisfies it: Spark 4 supports Java 17, 21, and 25, so there is no second JDK to install.
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
