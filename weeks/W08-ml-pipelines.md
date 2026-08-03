---
week_number: 8
status: not-started
---

# W08: ML Data Pipelines and Table Formats

> **Arc:** Distributed ML & Compute · **Language:** Python
> **Budget:** about 5 hours. Hit the Minimum bar first; everything past it is optional.

## What you'll build
A versioned feature pipeline in Python: raw events to features to versioned Parquet snapshot. Query historical feature snapshots with DuckDB. No ML model, just the data plumbing that makes models reliable. Then, in Part 2, you replace your hand-rolled versioning with a real open table format and open up its transaction log to see how the production answer differs from yours.

**Scenario:** a model trained against v1 features starts making noticeably worse predictions after a routine pipeline change, and nobody can say why without being able to answer "what exactly changed between v1 and v2, for which users?" `FeatureStore.diff()` is that question made answerable instead of a Slack message asking whoever last touched the pipeline.

---

## Read
- [ ] [Hidden Technical Debt in Machine Learning Systems](https://proceedings.neurips.cc/paper_files/paper/2015/file/86df7dcfd896fcaf2674f757a2463eba-Paper.pdf) (Sculley et al., NeurIPS 2015): 9 pages, read all of it. The CACE principle and "glue code" sections are most relevant.
- [ ] [Delta Lake: High-Performance ACID Table Storage over Cloud Object Stores](https://www.vldb.org/pvldb/vol13/p3411-armbrust.pdf) (Armbrust et al., VLDB 2020): read Sections 1–4. Understand why versioning and ACID matter for ML pipelines, not just OLTP.

**Depth: read Hidden Technical Debt and Sections 1 to 4 of the Delta Lake paper.** No study reading: this week's mechanism lives in the transaction log you open by hand, not in a paper. The Iceberg spec is a skim, and only the two sections named.

**Key question:** The paper says "changing anything changes everything" (CACE). Give a concrete example from an ML pipeline where this would cause a silent, hard-to-debug failure.

---

## Part 1: Versioning It Yourself

### Code

Project: `code/feature-pipeline/` (Python 3.13+, `uv` or `pip`)

Dependencies: `pandas`, `pyarrow`, `duckdb`. No ML libraries. Part 2 adds `deltalake`.

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

**Your call:** `pipeline.py` modifies 10% of raw events between v1 and v2, but real pipelines also gain and lose users entirely between versions, someone stops generating events, someone new shows up. Decide what `diff(v1, v2)` should do with a `user_id` present in one version but not the other: silently skip it (the join just won't match, so this is what happens if you don't handle it explicitly), report it as a row with the old values and nulls for the new ones, or exclude it with a separate count logged ("N users present in v1 dropped from v2"). Pick one and implement it deliberately, then note in Reflect which choice you made and what a downstream model-monitoring job would need from `diff()` that your choice doesn't give it.

---

## Part 2: What a Table Format Actually Is

You just built versioning by hand: a directory per version and a `latest.txt` pointer. That works, and it is worth having built, because it makes the next part legible. An open table format like Delta Lake or Apache Iceberg is the same idea done properly, and the difference is entirely in the metadata layer. Instead of a text file naming the current version, there is an ordered log of commits, each one listing exactly which data files were added and which were removed. Every guarantee people advertise about these formats, ACID commits, time travel, concurrent writers, schema evolution, falls out of that one design.

This matters commercially, not just intellectually: Delta Lake is the storage layer under Databricks, Iceberg is what Snowflake and nearly everyone else has standardized on, and "how does the transaction log work" is a question you should be able to answer from having looked at one.

### Read
- [ ] The Delta Lake paper is already required above. Reread Section 3 specifically, now that you've built versioning yourself, and note what your `latest.txt` cannot do that an ordered commit log can.
- [ ] [Apache Iceberg Table Spec](https://iceberg.apache.org/spec/): read the "Overview" and "Table Metadata" sections only. You want the three-level shape: metadata file, manifest list, manifest files pointing at data files. Iceberg and Delta solve the same problem with visibly different structures, and seeing both stops you from mistaking one implementation for the concept.

**Key question:** Both formats keep immutable data files and express updates as metadata commits. Why is that easier to make correct on object storage than editing files in place would be?

### Code

Same project. New dependency: `deltalake` (`pip install deltalake`). This is delta-rs, a native Rust implementation with a Python binding, so there is no JVM and no Spark cluster involved. You will not write any Rust.

- [ ] `delta_store.py`: rewrite `FeatureStore` against Delta. `write()` becomes `write_deltalake(path, df, mode="overwrite")`, `read(version)` becomes `DeltaTable(path, version=v).to_pandas()`. Confirm your existing `query.py` still works by pointing DuckDB at the table's data files. The point of the rewrite is how little application code it takes once the format handles versioning.
- [ ] `inspect_log.py`: after writing three versions, open `_delta_log/` and read the JSON commit files directly with the standard library. Print, for each commit, which files were added and which were removed. Then call `DeltaTable(path).history()` and confirm it is telling you the same story the raw files told you. Do the raw read first. The API is easy to trust without understanding, and the files are the thing that is actually true.
- [ ] Optional: `small_files.py`: append 200 tiny batches, one at a time, the way a streaming job writing every few seconds would. Count the files on disk and time a full table scan. Then run `DeltaTable(path).optimize.compact()`, count files and time the scan again. Report both numbers.

**Break it, then decide:**
- [ ] Optional, and only if you built `small_files.py`: the small-file problem is one of the most common real complaints about lakehouse tables, and the cause is unglamorous: every commit writes at least one new file, and query planning cost scales with file count regardless of how little data each file holds. Confirm the scan time actually improved after compaction, and note how much of the original slowness was metadata rather than data.
- [ ] Now run `DeltaTable(path).vacuum(retention_hours=0, dry_run=True)` and read what it proposes to delete. Those are the files compaction orphaned, still on disk, still referenced by older versions of the table. Deleting them reclaims storage and permanently breaks time travel to those versions.
- [ ] **Your call:** you own a feature table that a model-retraining job reads and an auditor occasionally queries months later. Compaction is clearly worth running. Vacuum is the question: an aggressive retention window keeps storage costs flat but destroys your ability to reproduce a training run from six weeks ago, and a long one preserves reproducibility while accumulating files nobody reads. Pick a retention window, write it down as a number with a justification, and say which of the two people above you would have to go apologize to if you got it wrong in each direction.

---

## Reflect

**What clicked:**

**What surprised me:**

**What would break at scale that works fine here?**

**How did you handle users present in one version but not the other in `diff()`, and what would a real monitoring job still be missing from your choice?**

**What does the Delta commit log give you that `latest.txt` could not, and what did the raw `_delta_log/` JSON show that `history()` did not?**

**File count and scan time before and after compaction, plus the retention window you chose and why:**

**How a streaming system could replace the batch feature pipeline:**

**What I'd do differently:**
