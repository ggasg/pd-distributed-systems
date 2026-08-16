---
week_number: 7
status: not-started
---

# W07: Query Planning: Choosing Where Data Moves

> **Arc:** Data Movement and Execution · **Language:** Python (PySpark) and SQL
> **Budget:** about 10 hours. The Minimum bar is what a bad week looks like, not the target.

## What you'll build

Not a planner. You are going to make Spark's own planner change its mind, on purpose, four times, and read the evidence each time.

A logical `Join(customers, orders)` says nothing about data movement. It is a statement about the answer, not about the machines. Turning it into something executable means choosing a *strategy*, and the strategies differ almost entirely in what they move:

- **Broadcast hash join**: ship the small side in full to every worker, leave the large side exactly where it is. No shuffle at all. Requires the small side to fit in each worker's memory.
- **Shuffle hash join**: partition both sides by the join key so matching rows land together. That is the W05 shuffle, and it moves everything.
- **Sort-merge join**: partition both sides *and* sort them. More movement still, but it survives inputs too large to build a hash table over.

Choosing between these is the highest-leverage decision a distributed query engine makes, and it is made from an *estimate* of how big each side is. Get it right and a job runs in two minutes. Get it wrong and you either move terabytes you did not need to, or you try to broadcast something that does not fit and the executors die.

A planner you wrote agrees with you by construction. Spark's has a cost model you did not write and will make decisions you did not expect, which is the only way to see the failure modes.

**Scenario:** a report that ran in ninety seconds for a year now takes four hours, and nobody deployed anything. A dimension table crossed a size threshold, the planner quietly stopped choosing a broadcast, and every run since has been shuffling both sides of a join across the cluster. Nothing errored. Nothing was logged. This is the most common performance regression in a data platform, and by the end of this unit you will have caused it on purpose and then found it from the plan alone.

---

## Read

- [ ] **DDIA Chapter 3** (2nd ed.) if you skipped it in W06. The declarative-versus-imperative distinction is the entire precondition for this unit: a planner only gets to choose a join strategy because you did not tell it which one to use. Everything below is that freedom being exercised, and occasionally misused.
- [ ] [Spark SQL: Relational Data Processing in Spark](https://people.csail.mit.edu/matei/papers/2015/sigmod_spark_sql.pdf) (Armbrust et al., SIGMOD 2015): **the physical planning section** is the part that matters here. Catalyst applies rule-based rewrites first, then uses cost to select physical operators, and the join strategy is what it spends that cost model on. You are about to watch this exact process run.
- [ ] [Spark: Adaptive Query Execution](https://spark.apache.org/docs/latest/sql-performance-tuning.html#adaptive-query-execution): short docs page. Read the three features it lists, particularly dynamically switching join strategies. That feature exists because the estimate this unit is built on turns out to be unreliable at scale, and Spark eventually stopped trusting it before execution.

**Depth: study the Catalyst paper's physical-planning section.** It is a few pages and it names every component you are about to see in a plan printout. The AQE page is a short read.

**Key question, on paper, before you run anything:** A broadcast join moves the small table to every worker. A shuffle join moves both tables across the network once. For a 10 GB fact table and a 100 MB dimension table across 20 workers, roughly how many bytes does each strategy move? Work it out first and keep the number, because Spark is about to show you its own estimate and you want to have committed to yours before you see it.

---

## Code

Project: `code/query-plans/` (Python 3.12, PySpark 4.1.0, local mode)

**Data:** the shared commerce dataset again, at `--scale 1 --skew 0 --seed 42`. The lopsided sizes are the point: `orders` at 10,000,000 rows, `customers` at 200,000, `regions` at 50. `orders` carries `customer_id` and `customers` carries `region_id`, so a three-table join is natural. Uniform keys here, since this unit is about the planner's size estimates rather than about skew.

### Step 1: `fixtures.py`

- [ ] Generate the three tables and register them as tables, so the catalog can hold statistics for them.
- [ ] Run `ANALYZE TABLE <t> COMPUTE STATISTICS FOR ALL COLUMNS` on all three. Without this, Spark falls back to file-size estimates and half this unit's behaviour becomes noise rather than signal. The last item under Break it returns to this step.

### Step 2: the experiments

Each one is the same shape: change one thing, re-explain, diff the plan.

- [ ] **1. Read a plan at all.** Join all three tables and print `EXPLAIN FORMATTED`. Find the join nodes and name each one's strategy. Find every `Exchange` and say what it is moving and why. Then print `EXPLAIN COST` and compare Spark's `sizeInBytes` and `rowCount` estimates against what you know the real tables to be. Write down where its estimate is wrong; it will be wrong somewhere, usually after a filter.
- [ ] **2. The silent regression.** Grow `customers` past `spark.sql.autoBroadcastJoinThreshold` (10 MB by default) by regenerating it at 1,000,000 rows and re-running `ANALYZE`. Change nothing else. Note that the threshold is compared against Spark's estimated `sizeInBytes`, not against the Parquet file on disk, and compression makes those two numbers differ by several times. Read the estimate out of `EXPLAIN COST` before and after rather than inferring it from `ls`; if the plan did not change, the estimate has not crossed the threshold yet and the row count is what you adjust. Re-explain and diff: a `BroadcastHashJoin` became a `SortMergeJoin`, and two `Exchange` nodes appeared that were not there before. Now notice what did *not* happen. No error, no warning, no log line, no config change, no deploy. The only signal in production is that the job got slower. Write down the metric you would put on a dashboard to catch this, and be specific about where it would come from.
- [ ] **3. The expensive lie.** Force the broadcast that Spark just declined, with a `/*+ BROADCAST(customers) */` hint. The hint overrides the cost model, because hints always do. Watch what happens to driver memory as it collects the whole side to broadcast it. Locally you will see it get slow or die; on a twenty-worker cluster every executor materialises that table simultaneously and they die together, with an out-of-memory error that names the join but not the reason. This is one of the most common ways a Spark job fails in production, and the cause is almost always upstream: statistics that are stale, absent, or overridden by a hint somebody added during an incident two years ago and never removed.
- [ ] **4. Spark's third answer.** Turn AQE off (`spark.sql.adaptive.enabled=false`), run the query, and look at the plan. Turn it back on, run again, and look at the plan *after* completion: it reports `AdaptiveSparkPlan isFinalPlan=true` and the strategy may differ from the one chosen up front, because Spark re-decided using row counts it *observed* rather than ones it estimated. Note also `spark.sql.adaptive.autoBroadcastJoinThreshold`, a separate knob, which exists precisely because a post-shuffle decision can afford to be braver than a pre-execution one.

**Minimum bar:** for the three-table join, you can point at a plan printout and say which strategy each join uses, where every `Exchange` is, and what it is moving. Plus experiment 2 done and diffed, with your dashboard metric written down. Experiments 3 and 4 are worth doing and are not the bar.

---

## Break it, then decide

- [ ] **Your call:** you cannot make the estimate reliable, so you are choosing how to be wrong. Lower `autoBroadcastJoinThreshold` and you rarely broadcast, giving up a large speedup on every query that would have qualified, in exchange for never blowing up an executor. Raise it and you keep the speedup and accept occasional job failures. Pick a value, defend it, and name the workload that would make you change it.
- [ ] Now the harder half. AQE does not tune that trade-off, it declines to make the decision until it has real numbers. Say precisely what AQE cannot fix, given that it only re-decides at shuffle boundaries: which of the four experiments above would AQE have saved you from, and which would it not? Experiment 3 is the interesting case.
- [ ] Go back to the `ANALYZE TABLE` step in Step 1. You ran it manually. In a production warehouse, who runs it, how often, and what happens to every plan in the system on the day that job silently stops working? This is the actual mechanism behind experiment 3, and it is an operational question rather than a query-engine one.

---

## Reflect

<!-- Fill in at the end of the unit -->

**Prediction versus measurement.** Fill the predictions in *before* you run anything, and do not edit them afterwards. The gap is where calibration comes from.

| Quantity | Predicted | Measured | Which term I got wrong |
|----------|-----------|----------|------------------------|
| | | | |

Copy anything worth carrying into [MEASUREMENTS.md](../MEASUREMENTS.md).

**What clicked:**

**What surprised me:**

**Your bytes-moved answer for the 10 GB, 100 MB, 20-worker case, and what `EXPLAIN COST` estimated:**

**The two plans either side of the silent regression, and the metric you would monitor to catch it:**

**Where you would set the broadcast threshold, and which failure that accepts:**

**Which of the four experiments AQE would have saved you from, and which it would not:**

**A logical `Join` says nothing about data movement. Name every decision a planner has to make before that join can run on a cluster.**

**What I'd do differently:**
