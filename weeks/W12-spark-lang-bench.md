---
week_number: 12
status: not-started
---

# W12: PySpark vs. Scala Spark: Where the JVM Boundary Costs You

> **Arc:** Distributed ML & Compute · **Language:** Scala + Python (PySpark)

## What you'll build
The same aggregation job, run four ways: Scala Spark using only the DataFrame API, PySpark using only the DataFrame API, a Scala Spark job with a native UDF, and a PySpark job with a plain Python UDF doing the identical row-level transform. You'll measure wall-clock time and peak driver memory for each, at increasing row counts, and get a real, measured answer to a question usually settled by taking a vendor blog post on faith: does the host language actually matter for Spark performance, and if so, exactly where.

This closes out the Scala thread from W09 and W10. Those two weeks built toy versions of Catalyst and the aggregation algebra Spark's own `reduceByKey` relies on; this week runs real Spark and watches the thing you read about in W09's SIGMOD paper actually execute.

---

## Read
- [ ] [Introducing Pandas UDF for PySpark](https://www.databricks.com/blog/2017/10/30/introducing-vectorized-udfs-for-pyspark.html) (Databricks engineering blog, 2017): explains exactly why a row-at-a-time Python UDF is slow (every row round-trips through py4j serialization between the JVM and a separate Python process) and what Arrow-based vectorization fixes. You'll measure this yourself below rather than take the post's numbers on faith, but it names the mechanism clearly before you go looking for it.
- [ ] [Apache Spark docs: Pandas UDFs (a.k.a. Vectorized UDFs)](https://spark.apache.org/docs/latest/api/python/user_guide/sql/arrow_pandas.html): the current, versioned reference for `pandas_udf`, if you attempt the optional stretch arm below.
- [ ] Recall from W09: Catalyst compiles a logical plan into a physical plan, and whole-stage code generation turns that physical plan into actual JVM bytecode. A plain DataFrame operation (`groupBy`, `sum`, `filter`) never leaves the JVM regardless of which language built the plan; a UDF is different; it's a black box to Catalyst, so Spark has no choice but to call out to it for every row.

**Key question:** Scala and PySpark DataFrame code both compile down to the same Catalyst physical plan and the same generated JVM bytecode. So where exactly does a UDF re-introduce a language-dependent cost that plain DataFrame operations don't have, and why can't Catalyst optimize that cost away the way it optimizes everything else?

---

## Code

Project: `code/spark-lang-bench/` (`scala/` subproject: sbt, Scala 2.13; `python/` subproject: PySpark)

Pin the exact same Spark release in both subprojects, for example `3.5.x`: `libraryDependencies += "org.apache.spark" %% "spark-sql" % "3.5.1"` in `scala/build.sbt`, `pyspark==3.5.1` in `python/requirements.txt`. If the versions drift, you're no longer comparing languages, you're comparing Spark releases.

**Shared data (generate once, read from both languages):**

- [ ] `generate_orders.py`: same `(order_id: int64, region_id: int32, amount: float64)` shape as W07's orders table, `region_id` uniformly in `[0, 10)`. Generate three Parquet files at `data/orders_100k.parquet`, `data/orders_1m.parquet`, `data/orders_5m.parquet` (scale down if 5M rows is uncomfortably slow for the UDF benchmark on your machine; a starker gap at a smaller size is a fine substitute for a slow one at a larger size). Both the Scala and Python jobs below read these same files, so the input is identical across all four arms.

**Benchmark 1: DataFrame API only, no UDF (Scala vs. Python)**

- [ ] `scala/src/main/scala/DataFrameBenchmark.scala`: read an orders Parquet file, run `df.groupBy("region_id").sum("amount")`, call `.collect()` to force execution (Spark's DataFrame API is lazy, nothing runs until an action does; recall W07's Reflect question about this), time the whole thing with `System.nanoTime()`, print milliseconds.
- [ ] `python/dataframe_benchmark.py`: the identical job in PySpark: `df.groupBy("region_id").sum("amount").collect()`, timed with `time.perf_counter()`.

**Benchmark 2: same transform, as a UDF (Scala native vs. Python row-at-a-time)**

Transform: `adjust_amount(amount, region_id) = round(amount * (1.0 + (region_id % 5) * 0.01), 2)`. Deterministic, cheap, and trivially expressible as a built-in Spark SQL expression; that's deliberate. The point of this benchmark isn't that you need a UDF for this specific transform (you don't, and in real Spark work you should reach for built-in functions over UDFs whenever you can, precisely because of what you're about to measure); the point is to force every row through the language boundary on purpose so the cost is visible.

- [ ] `scala/src/main/scala/UdfBenchmark.scala`: register `adjustAmount` as a native Scala UDF (`udf((amount: Double, regionId: Int) => ...)`), apply with `withColumn`, then `groupBy("region_id").sum("adjusted")`, `.collect()`, timed the same way as Benchmark 1.
- [ ] `python/udf_benchmark.py`: register the same transform as a plain PySpark `udf()` (row-at-a-time, not `pandas_udf`), apply with `withColumn`, same groupBy-sum-collect, timed the same way.

**Peak memory, both benchmarks:** don't use `tracemalloc` here; for PySpark, the actual DataFrame data lives in the JVM's memory, not Python's, so an in-process Python memory tracer would only see driver-side Python object overhead and silently miss the thing you actually care about. Instead wrap each run externally and read the OS's own number: `/usr/bin/time -l python3 dataframe_benchmark.py <path>` on macOS (`/usr/bin/time -v` on Linux), and record "maximum resident set size" for both the Scala (run via `sbt run`, or `java -jar` against an assembled jar) and Python processes.

**Optional stretch: close the gap with a Pandas UDF**

- [ ] `python/pandas_udf_benchmark.py`: the same `adjust_amount` transform, this time as a `pandas_udf` (Arrow-vectorized: Spark hands Python whole columns as batches instead of one row at a time). Time it the same way, and place it alongside the other three results.

**Minimum bar:** four real, measured wall-clock numbers at your largest comfortable row count, Scala DataFrame, Python DataFrame, Scala UDF, Python UDF, plus peak RSS for at least the two UDF runs. The two DataFrame-only numbers should land close to each other; the two UDF numbers should not.

---

## Reflect

**What clicked:**

**What surprised me:**

**Your measured numbers, largest row count you ran (N = __):**
- Scala DataFrame-only: __ ms
- Python DataFrame-only: __ ms
- Scala UDF: __ ms
- Python UDF: __ ms
- Python UDF peak RSS: __ MB. Scala UDF peak RSS: __ MB.
- (If you did the stretch) Pandas UDF: __ ms. How much of the Scala-vs-Python UDF gap did vectorization close?

**Did the two DataFrame-only numbers land close together? If they didn't, what's your best guess for why, given that both should compile to the same physical plan?**

**Concretely, what does a value cross when Spark calls a Python UDF that it doesn't cross for a native Scala UDF or a built-in DataFrame operation? Name the actual mechanism, not just "it's slower."**

**Given what you measured, when would you actually reach for Scala Spark over PySpark on a real job, and when is that decision not worth the switching cost?**

**What I'd do differently:**
