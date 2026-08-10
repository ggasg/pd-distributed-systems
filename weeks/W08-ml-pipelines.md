---
week_number: 8
status: not-started
---

# W08: ML Data Pipelines and Table Formats

> **Arc:** Distributed ML Systems · **Language:** Python
> **Budget:** about 10 hours. The Minimum bar is what a bad week looks like, not the target.

## What you'll build
A versioned feature pipeline in Python: raw events to features to versioned Parquet snapshot. Query historical feature snapshots with DuckDB. No ML model, just the data plumbing that makes models reliable. Then, in Part 2, you replace your hand-rolled versioning with a real open table format and open up its transaction log to see how the production answer differs from yours.

**Scenario:** a model trained against v1 features starts making noticeably worse predictions after a routine pipeline change, and nobody can say why without being able to answer "what exactly changed between v1 and v2, for which users?" `FeatureStore.diff()` is that question made answerable instead of a Slack message asking whoever last touched the pipeline.

---

## Read
- [ ] [Hidden Technical Debt in Machine Learning Systems](https://proceedings.neurips.cc/paper_files/paper/2015/file/86df7dcfd896fcaf2674f757a2463eba-Paper.pdf) (Sculley et al., NeurIPS 2015): 9 pages, read all of it. The CACE principle and "glue code" sections are most relevant.
- [ ] [Delta Lake: High-Performance ACID Table Storage over Cloud Object Stores](https://www.vldb.org/pvldb/vol13/p3411-armbrust.pdf) (Armbrust et al., VLDB 2020): read Sections 1–4. Understand why versioning and ACID matter for ML pipelines, not just OLTP.
- [ ] **DDIA Chapter 8** (2nd ed.), Transactions, read the atomicity and isolation sections. The Delta Lake paper's title claims *ACID* table storage over object stores, and this chapter is what makes each of those four letters mean something specific rather than a marketing adjective. Read it before Part 2 and then ask, of the transaction log you open by hand, which of the four Delta actually gives you and at what isolation level.
- [ ] **DDIA Chapter 5** (2nd ed.), Encoding and Evolution. Your feature pipeline writes Parquet, a schema-carrying binary format, and then somebody adds a column. The chapter's backward-versus-forward compatibility distinction is precisely the question "can a model trained on v1 features still read v2, and can a v2 reader still read v1," which is the thing `FeatureStore.diff()` exists to make answerable. Read it before Part 2, where Delta's schema evolution turns it from a property you hope for into a setting somebody chose.

**Depth: read Hidden Technical Debt, Sections 1 to 4 of the Delta Lake paper, and DDIA Ch.5.** No study reading: this unit's mechanism lives in the transaction log you open by hand, not in a paper. The Iceberg spec is a skim.

**Key question:** The paper says "changing anything changes everything" (CACE). Give a concrete example from an ML pipeline where this would cause a silent, hard-to-debug failure.

---

## Part 1: Versioning It Yourself

### Code

Project: `code/feature-pipeline/` (Python 3.12+, `uv` or `pip`)

Dependencies: `pandas`, `pyarrow`, `duckdb`. No ML libraries. Part 2 adds `deltalake`.

Scenario: raw user activity events to session features to versioned feature store.

**Data:** the shared commerce dataset from W05, this time the `events` table at `--scale 1 --skew 0 --seed 42`, giving 1,000,000 events over 200,000 customers. A million rows rather than a handful, because the versioned diff, the compaction exercise, and W16's training input all need a table with enough rows that file counts and scan times are real numbers rather than noise. Schema in [code/README.md](../code/README.md).

- [ ] `generate_events.py`: emit `events` as `(event_id, customer_id, event_type, value, event_ts)` to `data/raw/events.parquet`. Reuse W05's generator rather than writing a second one.
- [ ] `feature_engineering.py`: read raw events, compute per-customer features: `(customer_id, session_count, avg_value, max_value, last_seen_ts)`. Write versioned output to `data/features/v{N}/features.parquet` where N is a monotonically increasing version number (read current version from `data/features/latest.txt`).
- [ ] `feature_store.py`: class `FeatureStore` with:
  - `write(version: int, df: pd.DataFrame)`: writes Parquet + updates `latest.txt`
  - `read(version: int) -> pd.DataFrame`: reads a specific version
  - `latest() -> pd.DataFrame`: reads current version
  - `diff(v1: int, v2: int) -> pd.DataFrame`: returns rows that changed between versions (join on `customer_id`, compare feature values)
- [ ] `query.py`: use DuckDB to run SQL against the versioned Parquet files: (1) find top 10 customers by `avg_value` in the latest version; (2) find customers whose `session_count` changed between v1 and v2
- [ ] `pipeline.py`: end-to-end script: generate, engineer features (v1), modify 10% of raw events, re-engineer (v2), print diff

**Constraints:** no MLflow, no DVC, no feature store library. Implement versioning yourself.

**Your call:** `pipeline.py` modifies 10% of raw events between v1 and v2, but real pipelines also gain and lose customers entirely between versions, someone stops generating events, someone new shows up. Decide what `diff(v1, v2)` should do with a `customer_id` present in one version but not the other: silently skip it (the join just won't match, so this is what happens if you don't handle it explicitly), report it as a row with the old values and nulls for the new ones, or exclude it with a separate count logged ("N users present in v1 dropped from v2"). Pick one and implement it deliberately, then note in Reflect which choice you made and what a downstream model-monitoring job would need from `diff()` that your choice doesn't give it.

---

## Part 2: What a Table Format Actually Is

You just built versioning by hand: a directory per version and a `latest.txt` pointer. That works, and it is worth having built, because it makes the next part legible. An open table format like Delta Lake or Apache Iceberg is the same idea done properly, and the difference is entirely in the metadata layer. Instead of a text file naming the current version, there is an ordered log of commits, each one listing exactly which data files were added and which were removed. Every guarantee people advertise about these formats, ACID commits, time travel, concurrent writers, schema evolution, falls out of that one design.

Both Delta Lake and Apache Iceberg are open, independently governed formats, and both are implemented against object storage by more than one engine. That independence is the interesting part: the format is a contract, and any engine that can read the log can read the table. "How does the transaction log work" is a question worth being able to answer from having opened one.

**Two terms before you start.**

An **incremental read** consumes only the commits added since some version you already processed, rather than rescanning the whole table. The ordered commit log is what makes it possible: each commit names exactly which files were added and removed, so a reader that remembers where it stopped can ask for the difference directly. Your `latest.txt` cannot answer that question, which is the concrete thing Part 2 is showing you. This is the same idea as the incremental processing defined in [W04](W04-streaming.md), moved from a stream of events to a log of table commits, and it is the mechanism underneath every "process only new data" pipeline you will meet in production. The naive alternative, rescanning the table and diffing, is what your Part 1 `diff()` does, and it costs a full read of both versions every time.

**Time travel** is the same log read the other direction: reconstructing the set of files that were live as of an earlier version or timestamp, which is what `DeltaTable(path, version=v)` does. Incremental reads move forward from a known position, time travel jumps backward to one. Both are consequences of keeping an ordered log of immutable files rather than mutating a table in place, and vacuum is what destroys the second one.

### Read
- [ ] The Delta Lake paper is already required above. Reread Section 3 specifically, now that you've built versioning yourself, and note what your `latest.txt` cannot do that an ordered commit log can.
- [ ] [Apache Iceberg Table Spec](https://iceberg.apache.org/spec/): read the "Overview" and "Table Metadata" sections only. You want the three-level shape: metadata file, manifest list, manifest files pointing at data files. Iceberg and Delta solve the same problem with visibly different structures, and seeing both stops you from mistaking one implementation for the concept.

**Key question:** Both formats keep immutable data files and express updates as metadata commits. Why is that easier to make correct on object storage than editing files in place would be?

### Code

Same project. New dependency: `deltalake` (`pip install deltalake`), which is delta-rs, an independent Rust implementation of the Delta format with a Python binding. No JVM, no Spark cluster, and you will not write any Rust. It is used here because it is the shortest path to a real transaction log on a laptop, not because the format matters more than Iceberg's; the Iceberg spec in the reading is there so you see the same idea structured differently.

- [ ] `delta_store.py`: rewrite `FeatureStore` against Delta. `write()` becomes `write_deltalake(path, df, mode="overwrite")`, `read(version)` becomes `DeltaTable(path, version=v).to_pandas()`. Confirm your existing `query.py` still works by pointing DuckDB at the table's data files. The point of the rewrite is how little application code it takes once the format handles versioning.
- [ ] `inspect_log.py`: after writing three versions, open `_delta_log/` and read the JSON commit files directly with the standard library. Print, for each commit, which files were added and which were removed. Then call `DeltaTable(path).history()` and confirm it is telling you the same story the raw files told you. Do the raw read first. The API is easy to trust without understanding, and the files are the thing that is actually true.
- [ ] Optional: `small_files.py`: append 200 tiny batches, one at a time, the way a streaming job writing every few seconds would. Count the files on disk and time a full table scan. Then run `DeltaTable(path).optimize.compact()`, count files and time the scan again. Report both numbers.

**Break it, then decide:**
- [ ] Optional, and only if you built `small_files.py`: the small-file problem is one of the most common real complaints about lakehouse tables, and the cause is unglamorous: every commit writes at least one new file, and query planning cost scales with file count regardless of how little data each file holds. Confirm the scan time actually improved after compaction, and note how much of the original slowness was metadata rather than data.
- [ ] Now run `DeltaTable(path).vacuum(retention_hours=0, dry_run=True)` and read what it proposes to delete. Those are the files compaction orphaned, still on disk, still referenced by older versions of the table. Deleting them reclaims storage and permanently breaks time travel to those versions.
- [ ] **Your call:** you own a feature table that a model-retraining job reads and an auditor occasionally queries months later. Compaction is clearly worth running. Vacuum is the question: an aggressive retention window keeps storage costs flat but destroys your ability to reproduce a training run from six units ago, and a long one preserves reproducibility while accumulating files nobody reads. Pick a retention window, write it down as a number with a justification, and say which of the two people above you would have to go apologize to if you got it wrong in each direction.

---

## Reflect


**Prediction versus measurement.** Fill the predictions in *before* you run anything, and do not edit them afterwards. The gap is where calibration comes from.

| Quantity | Predicted | Measured | Which term I got wrong |
|----------|-----------|----------|------------------------|
| | | | |

Copy anything worth carrying into [MEASUREMENTS.md](../MEASUREMENTS.md).

**What clicked:**

**What surprised me:**

**What would break at scale that works fine here?**

**How did you handle users present in one version but not the other in `diff()`, and what would a real monitoring job still be missing from your choice?**

**What does the Delta commit log give you that `latest.txt` could not, and what did the raw `_delta_log/` JSON show that `history()` did not?**

**File count and scan time before and after compaction, plus the retention window you chose and why:**

**How a streaming system could replace the batch feature pipeline:**

**What I'd do differently:**

---

## Review and articulate

Two steps that exist because self-study has no examiner. Do them at the end of every unit, before marking it done.

- [ ] **Adversarial review.** Hand over three things separately: the number you predicted, the number you measured, and the conclusion you drew. Then ask for the strongest case that the conclusion is *not* supported by the measurement. Do not ask whether you are right; ask what would falsify this. An assistant asked to check your work will tend to find support for your framing, so the prompt has to be adversarial by construction or the exercise is theatre.
- [ ] **Ninety seconds, out loud, timed.** Explain this unit's finding as you would to someone in an interview or a design review: what you measured, what surprised you, and what decision it would change. Articulation under time pressure is a separate skill from understanding, and it is the one that gets tested. If you cannot do it in ninety seconds you do not have the finding yet, you have notes.
