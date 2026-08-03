---
week_number: 7
status: not-started
---

# W07: Query Planning: Choosing Where Data Moves

> **Arc:** Data Movement and Execution · **Language:** Scala
> **Budget:** about 5 hours. Hit the Minimum bar first; everything past it is optional.

## What you'll build
A physical planner: the component that takes a logical query and decides, for every join in it, **how much data crosses the network and in which direction**.

Here is the framing, because it would be easy to mistake this unit for compiler work. A logical `Join(customers, orders)` says nothing about data movement. It is a statement about the answer, not about the machines. Turning it into something executable means choosing a *strategy*, and the strategies differ almost entirely in what they move:

- **Broadcast hash join**: ship the small side in full to every worker, leave the large side exactly where it is. No shuffle at all. Requires the small side to fit in each worker's memory.
- **Shuffle hash join**: partition both sides by the join key so matching rows land together. That is the W05 shuffle, and it moves everything.
- **Sort-merge join**: partition both sides *and* sort them. More movement still, but it survives inputs too large to build a hash table over.

Choosing between these is the highest-leverage decision a distributed query engine makes, and it is made from an *estimate* of how big each side is. Get it right and a job runs in two minutes. Get it wrong and you either move terabytes you did not need to, or you try to broadcast something that does not fit and the executors die.

**Scenario:** a report that ran in ninety seconds for a year now takes four hours, and nobody deployed anything. A dimension table crossed a size threshold, the planner quietly stopped choosing a broadcast, and every run since has been shuffling both sides of a join across the cluster. Nothing errored. Nothing was logged. This is the most common performance regression in a data platform, and by the end of this unit you will have caused it on purpose.

**Note on why Scala:** Spark is written in Scala and Catalyst genuinely works this way, case classes for plan nodes and pattern matching to rewrite them. Reading real Catalyst source while writing a small version in the same language is something no other unit here offers. The Scala stays gentle: case classes, pattern matching, and one recursive function.

---

## Read
- [ ] [Spark SQL: Relational Data Processing in Spark](https://people.csail.mit.edu/matei/papers/2015/sigmod_spark_sql.pdf) (Armbrust et al., SIGMOD 2015): **the physical planning section** is the part that matters here. Catalyst applies rule-based rewrites first, then uses cost to select physical operators, and the join strategy is what it spends that cost model on.
- [ ] [Spark: Adaptive Query Execution](https://spark.apache.org/docs/latest/sql-performance-tuning.html#adaptive-query-execution): short docs page. Read the three features it lists, particularly dynamically switching join strategies. That feature exists because the estimate this unit is built on turns out to be unreliable at scale, and Spark eventually stopped trusting it before execution.

**Depth: study the Catalyst paper's physical-planning section.** It is a few pages and it is exactly what you are about to build. The AQE page is a short read. Skim anything else you open.

**Key question:** A broadcast join moves the small table to every worker. A shuffle join moves both tables across the network once. For a 10 GB fact table and a 100 MB dimension table across 20 workers, roughly how many bytes does each strategy move? Work it out before you code; the gap is larger than most people expect, and it does not go the way you might guess.

---

## Code

Project: `code/query-planner/` (Scala 2.13, sbt)

**Given, not built:** `LogicalPlan.scala` is provided: `sealed trait LogicalPlan` with `Scan(table, columns)`, `Filter(predicate, child)`, `Project(columns, child)`, and `Join(left, right, key)`. Plain case classes, no methods. Defining the tree type is not the lesson and you would spend an hour on it.

- [ ] `Statistics.scala`: `case class TableStats(rowCount: Long, avgRowBytes: Int)` plus a `Catalog = Map[String, TableStats]`. Write one for three tables with deliberately lopsided sizes: `orders` at 10,000,000 rows, `customers` at 200,000, `regions` at 50. `estimatedBytes` is `rowCount * avgRowBytes`, which is crude, and is exactly what a real planner does before it has anything better.
- [ ] `Strategy.scala`: `sealed trait JoinStrategy` permitting `BroadcastHash`, `ShuffleHash`, and `SortMerge`. Give each a `bytesMoved(left: Long, right: Long, workers: Int): Long`. Broadcast moves `smallSide * workers`. Shuffle moves `left + right`. Sort-merge moves the same as shuffle, with a comment noting it also pays a sort. Those three formulas are the entire cost model and they are enough to make every decision in this unit.
- [ ] `Planner.scala`: `def plan(logical: LogicalPlan, catalog: Catalog, broadcastThresholdBytes: Long, workers: Int): PhysicalPlan`. Walk the tree; for each `Join`, estimate both sides and choose `BroadcastHash` if the smaller side is under the threshold, `ShuffleHash` otherwise. Push each `Filter` below its `Join` *before* estimating, because a filter that removes 90 percent of a table changes which strategy wins. That ordering is the one genuine correctness constraint in this unit, and it is also why real optimizers run rules before costing.
- [ ] `Explain.scala`: print the physical plan the way `df.explain()` does, with an `Exchange` node wherever data moves and the estimated bytes on it. That printout is the deliverable. You should be able to read it and say, without running anything, where the network is about to be used and how hard.

**Minimum bar:** for a three-table join with one 50-row dimension table, your planner broadcasts the tiny one, shuffles the two large ones, and `Explain` prints a plan with `Exchange` nodes carrying byte estimates. You can point at the plan and say where the network gets used.

**Break it, then decide:**
- [ ] **The silent regression.** Change `customers` from 200,000 rows to 2,000,000. One number, nothing else. Re-plan and diff the two printouts: a broadcast became a shuffle, and estimated bytes moved jumped by whatever your model says. Now notice what did *not* happen. No error, no warning, no log line. In production this is a deploy-free, config-free change in behaviour caused only by data growth, and the sole signal is that the job got slower. Write down the metric you would put on a dashboard to catch it.
- [ ] **The expensive lie.** Tell the catalog `customers` is 50 MB when it is really 5 GB. The planner cheerfully chooses a broadcast. On a real cluster all 20 workers then try to materialise 5 GB each, and the executors die with out-of-memory errors that name the join but not the reason. This is one of the most common ways a Spark job fails in production, and the cause is always upstream: statistics that are stale, or absent.
- [ ] **Your call:** you cannot make the estimate reliable, so you choose how to be wrong. Lower the broadcast threshold and you rarely broadcast, giving up a large speedup on every query that would have qualified, in exchange for never running out of memory. Keep it high and you keep the speedup and accept occasional job failures. Implement one, then say what Spark's third answer is, given what you read about AQE, and why declining to commit before execution is different in kind from either of your two options rather than just a better tuning of them.

### Optional stretch: watch a real planner change its mind

Fifteen minutes, observation only, using the PySpark you already have.

- [ ] Write a small PySpark join between a large DataFrame and a small one and call `df.explain("formatted")`. Find the `BroadcastHashJoin` in the plan. Now set `spark.sql.autoBroadcastJoinThreshold` to `-1`, disabling broadcasts entirely, and explain again: the same query is a `SortMergeJoin` with `Exchange` nodes in it. You have just watched your own planner's decision made by a production one.
- [ ] With AQE enabled, run the query and explain again *after* it finishes. The plan reports itself as adaptive, and the strategy shown may differ from the one chosen up front, because Spark re-decided using row counts it observed rather than ones it estimated.

---

## Reflect
<!-- Fill in at the end of the unit -->

**What clicked:**

**What surprised me:**

**Your bytes-moved answer for the 10 GB, 100 MB, 20-worker case before coding, and what your planner actually computed:**

**The two plans either side of the silent regression, and the metric you would monitor to catch it:**

**Where you would set the broadcast threshold, which failure that accepts, and how AQE avoids the choice rather than tuning it:**

**A logical `Join` says nothing about data movement. Name every decision a planner has to make before that join can run on a cluster.**

**What I'd do differently:**
