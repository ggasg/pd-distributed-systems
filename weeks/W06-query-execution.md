---
week_number: 6
status: not-started
---

# W06: Query Execution

> **Arc:** Data Movement and Execution · **Language:** Python (DuckDB, NumPy)
> **Budget:** about 10 hours. The Minimum bar is what a bad week looks like, not the target.

## What you'll build

Not a query engine. One naive row-at-a-time pipeline you write, the same query run by a real vectorized engine, and an honest measurement of the gap between them.

One side of the benchmark is a real engine on purpose. If you write both sides, the result measures two programs you wrote in an afternoon rather than the technique.

**Scenario:** the 3 to 8x figure quoted for vectorized execution is measured against one workload shape. Take it to production unexamined and the first colleague who runs a highly selective filter, or a join where one side does not fit in memory, will find the exact place it stops holding. The point of this unit is to find those places yourself, on purpose, before they find you.

---

## Read

- [ ] **DDIA Chapter 3** (2nd ed.), Data Models and Query Languages. Specifically the declarative-versus-imperative argument. This is the chapter that explains why you are able to hand DuckDB a `SELECT` and let it decide how to execute, and why that separation is what makes vectorization possible at all: a declarative query does not specify a row-at-a-time loop, so the engine is free not to run one. It is also the precondition for W07, where the same freedom becomes a decision about network traffic.
- [ ] [Volcano, An Extensible and Parallel Query Evaluation System](https://dl.acm.org/doi/10.1109/69.273032) (Graefe, 1994): read Sections 1 to 3. This defines the iterator model (the `next()` interface) that every query engine for twenty years was built on, and it is the model your row-at-a-time baseline below is an instance of.
- [ ] [MonetDB/X100: Hyper-Pipelining Query Execution](https://www.cidrdb.org/cidr2005/papers/P19.pdf) (Boncz et al., CIDR 2005): read Sections 1 to 3. This is the argument for vectorized execution and why Volcano is CPU-cache unfriendly. It is also the argument your measurement is about to either confirm or fail to reproduce.
- [ ] Optional: [Dremel: Interactive Analysis of Web-Scale Datasets](https://research.google/pubs/dremel-interactive-analysis-of-web-scale-datasets/) (Melnik et al., VLDB 2010): read Sections 1 to 3. Same columnar instinct, one level up: Dremel shreds nested records into columns, the ancestor of Parquet's on-disk format, and spreads the aggregation across a serving tree. Read it for what changes when "vectorize the scan" becomes "vectorize the scan, then fan the aggregation across a cluster."
- [ ] Optional: **Burns, *Designing Distributed Systems*, 2nd ed., Chapter 8** (Scatter/Gather). Pairs directly with the Dremel paper above: Dremel describes a multi-level serving tree, Burns gives the same shape as a reusable pattern and asks the question Dremel does not, which is "Choosing the Right Number of Leaves." That is the same question as choosing a partition count, which W05 and W07 both make you answer with a number.
- [ ] Optional: [DuckDB execution engine source](https://github.com/duckdb/duckdb/tree/main/src/execution): skim `PhysicalFilter` and how DuckDB batches rows into `DataChunk`s. This is the implementation of the thing you are measuring.

**Depth: study Sections 1 to 3 of MonetDB/X100.** Volcano is a read. Dremel and the DuckDB source are skims and both optional.

**Key question:** Why does calling `next()` once per row hurt CPU performance even when the logic inside is trivial? Name at least two distinct costs, and predict which one dominates. You are about to see a per-operator profile that will tell you whether you were right.

---

## Code

Project: `code/query-exec/` (Python 3.12, `duckdb`, `numpy`, `pyarrow`)

DuckDB rather than Spark here: Spark's per-query overhead would swamp the effect being measured at this data size, and DuckDB is a single-node vectorized engine, which is what this unit is about.

**Data:** the shared commerce dataset from W05, regenerated here at `--scale 1 --skew 0 --seed 42`: 10,000,000 `orders` joined against 200,000 `customers`. Uniform keys, because skew is W05's subject and would only add noise to a measurement about CPU. The query is a filter on `orders.amount`, a projection, and a hash join to `customers` on `customer_id`. Same files feed all three implementations, so none of them gets to blame the input.

**The volume has a job to do.** Ten million rows is chosen so the row-at-a-time version takes tens of seconds rather than tens of milliseconds: at a few hundred milliseconds, process startup and Parquet decode dominate and the three-way comparison measures the wrong thing. Time the row-at-a-time run first. If it finishes in under about ten seconds, raise `--scale` until it does not, and note the scale you settled on, since your ratios are only comparable to themselves.

The same query, implemented three ways. Predict the ordering and the size of both gaps before you measure.

### Step 1: `row_at_a_time.py`

- [ ] Read the Parquet file and run filter (`amount > threshold`), then projection, then a hash join against `customers`, **one row at a time through a Python generator pipeline**. Write it the obvious way: `(r for r in rows if r.amount > t)` chained into the next stage.

A Python generator is lazy, pull-based, and yields one element at a time, with each stage requesting the next from the stage below it. That is the Volcano iterator model Graefe described in 1994, handed to you as a language feature. The architecture MonetDB/X100 argues against is the one you reach for by default.

### Step 2: `vectorized_numpy.py`

- [ ] The same query with NumPy arrays. Filter becomes a boolean mask over a column, projection becomes a masked take, and the join becomes a dictionary probe over arrays. No per-row Python at all.
- [ ] This middle term is what makes the comparison honest: it separates "vectorized beats row-at-a-time" from "C beats Python."

### Step 3: `duckdb_run.py`

- [ ] The same query as SQL over the same Parquet file, through DuckDB's Python API.

### Step 4: measure it honestly

- [ ] **Run `duckdb.sql("SET threads=1")` before measuring.** This matters more than anything else in the unit. DuckDB parallelises by default, so without this you are measuring core count and calling it vectorization. You want the execution model isolated, and you can turn threads back on afterwards to see what parallelism adds on top, which is a separate and also interesting number.
- [ ] `EXPLAIN ANALYZE` the query (via `duckdb.sql`) and read the per-operator timing. This is the artifact the unit is really after: you can see which operator ate the time, and the answer is frequently not the one you would have guessed from reading the SQL.

**Minimum bar:** all three implementations produce the same result on the same data, you have the single-threaded ratios between them, and you can explain **both** gaps: generator to NumPy, and NumPy to DuckDB. They have different causes and saying so is the unit. Plus one `EXPLAIN ANALYZE` output you can read out loud, operator by operator.

---

## Break it, then decide

- [ ] **Selectivity sweep.** Run three thresholds: one where under 1 percent of rows pass, one near 50 percent, one where nearly everything passes. Plot or tabulate the ratio at each. The gap is not constant, and the shape of how it changes tells you what the engine is actually spending its time on. Predict the shape before you run it.
- [ ] **Make the join side exceed memory.** Grow the build side of the join until it does not fit. Both of your implementations will die, the generator one with a `MemoryError` and the NumPy one when the array allocation fails, because a naive hash join builds the whole table before probing. DuckDB will not: it spills to disk and finishes slower. Measure how much slower. This is the single most important difference between a toy engine and a real one, and it is worth having felt rather than read about.
- [ ] **Your call:** given the number you just measured for the spill, would you rather an engine that fails fast when a join will not fit, so you find out immediately and go fix the query, or one that silently degrades to disk and finishes eventually? Both are defensible and real engines differ on this. Say which you would want as a platform operator, then say whether your answer changes if the person running the query is an analyst rather than you.
- [ ] **Your call:** given the two gaps you measured, where is your own line between "write the loop," "reach for NumPy or pandas," and "push it into DuckDB"? Answer in rows or in bytes, not in adjectives. Then say what would move the line: a different selectivity, a join that does not fit, or a transformation NumPy cannot express.
- [ ] Turn threads back on and re-measure. Report vectorization and parallelism as separate numbers, since a single combined figure cannot tell you which one bought you what.

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
