---
week_number: 6
status: not-started
---

# W06: Query Execution

> **Arc:** Data Movement and Execution · **Language:** Java (DuckDB JDBC)
> **Budget:** about 5 hours. Hit the Minimum bar first; everything past it is optional.

## What you'll build
Not a query engine. One naive row-at-a-time pipeline you write, the same query run by a real vectorized engine, and an honest measurement of the gap between them.

**Why not build the vectorized executor?** An earlier version of this unit had you write a columnar filter, a hash join, and a projection in Go and benchmark them against a row-at-a-time version of the same thing. The problem is that both sides of that benchmark were yours, which means the result was a measurement of two programs you wrote in an afternoon, not of the technique. You will never ship a query executor. What you will do, repeatedly, is reason about why an engine is fast, predict where it stops being fast, and read an operator-level profile to find out. DuckDB gives you all three, and its execution engine is the production version of the thing the earlier build approximated.

**Scenario:** the 3 to 8x figure quoted for vectorized execution is measured against one workload shape. Take it to production unexamined and the first colleague who runs a highly selective filter, or a join where one side does not fit in memory, will find the exact place it stops holding. The point of this unit is to find those places yourself, on purpose, before they find you.

---

## Read
- [ ] **DDIA Chapter 3** (2nd ed.), Data Models and Query Languages. Specifically the declarative-versus-imperative argument. This is the chapter that explains why you are able to hand DuckDB a `SELECT` and let it decide how to execute, and why that separation is what makes vectorization possible at all: a declarative query does not specify a row-at-a-time loop, so the engine is free not to run one. It is also the precondition for W07, where the same freedom becomes a decision about network traffic. Previously uncited anywhere in this curriculum, which was a real gap given that Arc 2 now runs on SQL end to end.
- [ ] [Volcano, An Extensible and Parallel Query Evaluation System](https://dl.acm.org/doi/10.1109/69.273032) (Graefe, 1994): read Sections 1 to 3. This defines the iterator model (the `next()` interface) that every query engine for twenty years was built on, and it is the model your row-at-a-time baseline below is an instance of.
- [ ] [MonetDB/X100: Hyper-Pipelining Query Execution](https://www.cidrdb.org/cidr2005/papers/P19.pdf) (Boncz et al., CIDR 2005): read Sections 1 to 3. This is the argument for vectorized execution and why Volcano is CPU-cache unfriendly. It is also the argument your measurement is about to either confirm or fail to reproduce.
- [ ] Optional: [Dremel: Interactive Analysis of Web-Scale Datasets](https://research.google/pubs/dremel-interactive-analysis-of-web-scale-datasets/) (Melnik et al., VLDB 2010): read Sections 1 to 3. Same columnar instinct, one level up: Dremel shreds nested records into columns, the ancestor of Parquet's on-disk format, and spreads the aggregation across a serving tree. Read it for what changes when "vectorize the scan" becomes "vectorize the scan, then fan the aggregation across a cluster."
- [ ] Optional: **Burns, *Designing Distributed Systems*, 2nd ed., Chapter 8** (Scatter/Gather). Pairs directly with the Dremel paper above: Dremel describes a multi-level serving tree, Burns gives the same shape as a reusable pattern and asks the question Dremel does not, which is "Choosing the Right Number of Leaves." That is the same question as choosing a partition count, which W05 and W07 both make you answer with a number.
- [ ] Optional: [DuckDB execution engine source](https://github.com/duckdb/duckdb/tree/main/src/execution): skim `PhysicalFilter` and how DuckDB batches rows into `DataChunk`s. This is no longer a foreign artifact you are told to admire; it is the implementation of the thing you are measuring.

**Depth: study Sections 1 to 3 of MonetDB/X100.** Volcano is a read. Dremel and the DuckDB source are skims and both optional.

**Key question:** Why does calling `next()` once per row hurt CPU performance even when the logic inside is trivial? Name at least two distinct costs, and predict which one dominates. You are about to see a per-operator profile that will tell you whether you were right.

---

## Code

Project: `code/query-exec/` (Java 21, Maven, DuckDB JDBC)

DuckDB rather than Spark for this one unit, deliberately. Spark's per-query overhead is large enough to swamp the effect being measured at this data size, and DuckDB is a single-node vectorized engine, which is exactly and only the thing this unit is about. You already depend on it in W08.

**Data:** 10,000,000 rows with columns `id`, `dept`, `salary`, written once to Parquet. Same file feeds both sides of the comparison, so neither side gets to blame the input.

**Baseline, which you write:**

- [ ] `RowAtATime.java`: read the Parquet file, materialise rows as objects, and run filter (`salary > threshold`), then projection, then a hash join against a second table, one row at a time through a `next()`-style iterator. Write it the obvious way. This is the Volcano model and it is not a straw man; it is how most query engines worked for two decades and how most hand-written data processing code still works.

**The real engine, which you drive:**

- [ ] `DuckDbRun.java`: the same query as SQL over the same Parquet file, via the DuckDB JDBC driver.
- [ ] **Set `SET threads=1` before measuring.** This matters more than anything else in the unit. DuckDB parallelises by default, so without this you are measuring core count and calling it vectorization. You want the execution model isolated, and you can turn threads back on afterwards to see what parallelism adds on top, which is a separate and also interesting number.
- [ ] `EXPLAIN ANALYZE` the query and read the per-operator timing. This is the artifact the unit is really after: you can see which operator ate the time, and the answer is frequently not the one you would have guessed from reading the SQL.

**Minimum bar:** both pipelines produce the same result on the same data, you have the single-threaded ratio between them, and you can explain the gap in terms of memory layout and batch size rather than instruction count. Plus one `EXPLAIN ANALYZE` output you can read out loud, operator by operator.

**Break it, then decide:**

- [ ] **Selectivity sweep.** Run three thresholds: one where under 1 percent of rows pass, one near 50 percent, one where nearly everything passes. Plot or tabulate the ratio at each. The gap is not constant, and the shape of how it changes tells you what the engine is actually spending its time on. Predict the shape before you run it.
- [ ] **Make the join side exceed memory.** Grow the build side of the join until it does not fit. Your `RowAtATime.java` will die with an `OutOfMemoryError`, because a naive hash join builds the whole table before probing. DuckDB will not: it spills to disk and finishes slower. Measure how much slower. This is the single most important difference between a toy engine and a real one, and it is worth having felt rather than read about.
- [ ] **Your call:** given the number you just measured for the spill, would you rather an engine that fails fast when a join will not fit, so you find out immediately and go fix the query, or one that silently degrades to disk and finishes eventually? Both are defensible and real engines differ on this. Say which you would want as a platform operator, then say whether your answer changes if the person running the query is an analyst rather than you.
- [ ] Turn threads back on and re-measure. Report vectorization and parallelism as two separate numbers rather than one combined one. Being able to say which of the two bought you what is the difference between understanding a benchmark and quoting it.

---

## Reflect
<!-- Fill in at the end of the unit -->

**What clicked:**

**What surprised me:**

**Single-threaded results:**
- Row-at-a-time: __ M rows/sec
- DuckDB, `threads=1`: __ M rows/sec
- Ratio: __x
- DuckDB, threads on: __ M rows/sec, so parallelism added __x on top

**Which operator dominated `EXPLAIN ANALYZE`, and whether it was the one you predicted in the Key question:**

**How the ratio changed across the three selectivities, and what that says about where the time goes:**

**What happened when the build side exceeded memory, and what the spill cost:**

**Fail fast or degrade to disk, and whether your answer changes for an analyst rather than an operator:**

**What I'd do differently:**
