---
week_number: 11
status: not-started
---

# W11: ML Data Pipelines

> **Arc:** Distributed ML & Compute · **Language:** Python

## What you'll build
A versioned feature pipeline in Python: raw events to features to versioned Parquet snapshot. Query historical feature snapshots with DuckDB. No ML model, just the data plumbing that makes models reliable.

---

## Read
- [ ] [Hidden Technical Debt in Machine Learning Systems](https://proceedings.neurips.cc/paper_files/paper/2015/file/86df7dcfd896fcaf2674f757a2463eba-Paper.pdf) (Sculley et al., NeurIPS 2015): 9 pages, read all of it. The CACE principle and "glue code" sections are most relevant.
- [ ] [Delta Lake: High-Performance ACID Table Storage over Cloud Object Stores](https://www.vldb.org/pvldb/vol13/p3411-armbrust.pdf) (Armbrust et al., VLDB 2020): read Sections 1–4. Understand why versioning and ACID matter for ML pipelines, not just OLTP.

**Key question:** The paper says "changing anything changes everything" (CACE). Give a concrete example from an ML pipeline where this would cause a silent, hard-to-debug failure.

---

## Code

Project: `code/feature-pipeline/` (Python 3.11+, `uv` or `pip`)

Dependencies: `pandas`, `pyarrow`, `duckdb`. No ML libraries.

Scenario: raw user activity events to session features to versioned feature store.

- [ ] `generate_events.py`: generate 10k synthetic events as `(user_id: int, event_type: str, timestamp: int, value: float)`. Write to `data/raw/events.parquet`.
- [ ] `feature_engineering.py`: read raw events, compute per-user features: `(user_id, session_count, avg_value, max_value, last_seen_timestamp)`. Write versioned output to `data/features/v{N}/features.parquet` where N is a monotonically increasing version number (read current version from `data/features/latest.txt`).
- [ ] `feature_store.py`: class `FeatureStore` with:
  - `write(version: int, df: pd.DataFrame)`: writes Parquet + updates `latest.txt`
  - `read(version: int) -> pd.DataFrame`: reads a specific version
  - `latest() -> pd.DataFrame`: reads current version
  - `diff(v1: int, v2: int) -> pd.DataFrame`: returns rows that changed between versions (join on `user_id`, compare feature values)
- [ ] `query.py`: use DuckDB to run SQL against the versioned Parquet files: (1) find top 10 users by `avg_value` in the latest version; (2) find users whose `session_count` changed between v1 and v2
- [ ] `pipeline.py`: end-to-end script: generate, engineer features (v1), modify 10% of raw events, re-engineer (v2), print diff

**Constraints:** no MLflow, no DVC, no feature store library. Implement versioning yourself.

---

## Optional: Memory-Bound Processing (evidence-based)

`events.parquet` at 10k rows fits in memory without thinking about it. Real event tables don't: mixed-type JSON fields (a `tags` column that's sometimes a list, sometimes a string, sometimes null, depending on which client emitted the event) and row counts in the millions are exactly the shape of data that makes naive pandas fall over. Worth seeing the failure and the fix firsthand instead of taking it on faith.

No new dependencies: everything below uses `pandas`, `pyarrow`, and the standard library you already have for this week. No Dask, no Polars, no new file format; the point is that `pandas` + `pyarrow` + Parquet, used correctly, already get you most of the way there.

- [ ] `generate_events_large.py`: same schema as `generate_events.py`, but 2M rows (bump higher if your machine has headroom and you want a starker gap) and one deliberately dirty column, `tags`: for each row, randomly emit a JSON-serializable list of strings, a single string, or `null` (roughly equal thirds). Write to `data/raw/events_large.parquet`.
- [ ] `memory_naive.py`: load `events_large.parquet` in one call (`pd.read_parquet`), then normalize `tags` to a consistent string representation column-wide (`.astype(str)`). Wrap the load-and-normalize step in `tracemalloc.start()` / `tracemalloc.get_traced_memory()` and print peak memory in MB.
- [ ] `memory_chunked.py`: process the same file in real chunks, using `pyarrow.parquet.ParquetFile(path).iter_batches(batch_size=100_000)` directly, converting each batch to a small pandas DataFrame only for that batch, normalizing `tags`, and writing each cleaned batch out incrementally with `pyarrow.parquet.ParquetWriter` (one `write_table` call per batch). At no point should the full 2M-row dataset exist as one in-memory object; that's the actual difference between this and a loop that slices an already-fully-loaded DataFrame, which isn't real chunking, just decorative iteration. Measure peak memory the same way as `memory_naive.py` and compare.
- [ ] `memory_columnar.py`: read the same file two ways and compare `.memory_usage(deep=True).sum()` for the `tags` column: once as plain pandas (`pd.read_parquet(path)`, `object` dtype, one Python string object per cell) and once with `pd.read_parquet(path, dtype_backend="pyarrow")` (pandas 2.x's PyArrow-backed columnar dtypes, no per-cell Python object overhead; `pip install pandas` today gets you 2.x by default). This is W02's row-vs-column-store memory argument again, one level up: DataFrame columns instead of hand-rolled Go structs.

**Minimum bar:** three real numbers you measured, not estimated: peak memory naive vs. chunked (MB), and `tags` column memory footprint object-dtype vs. `dtype_backend="pyarrow"` (MB). `memory_chunked.py`'s peak should be substantially lower than `memory_naive.py`'s; if it isn't, the first thing to check is whether the full DataFrame got materialized somewhere before chunking kicked in. That's the most common way this exercise goes wrong, and it's the same mistake the inspiration for this exercise (a published data engineering article) actually made in its own "chunked" solution.

**Reflect (this section only):**
- Peak memory, naive: __ MB. Chunked: __ MB.
- `tags` column memory, object dtype: __ MB. PyArrow-backed dtype: __ MB.
- What did W02's row-vs-column-store benchmark predict about this result, and did it hold?
- Where would the naive-chunking mistake (iterating over an already-fully-loaded DataFrame instead of never materializing the whole thing) be easy to make by accident in `feature_engineering.py` above, at real scale?

---

## Reflect

**What clicked:**

**What surprised me:**

**What would break at scale that works fine here?**

**How a streaming system could replace the batch feature pipeline:**

**What I'd do differently:**
