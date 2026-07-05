---
week_number: 9
status: not-started
---

# W09 — ML Data Pipelines

> **Arc:** Distributed ML & Compute · **Language:** Python

## What you'll build
A versioned feature pipeline in Python: raw events → features → versioned Parquet snapshot. Query historical feature snapshots with DuckDB. No ML model — just the data plumbing that makes models reliable.

---

## Read
- [ ] [Hidden Technical Debt in Machine Learning Systems](https://proceedings.neurips.cc/paper_files/paper/2015/file/86df7dcfd896fcaf2674f757a2463eba-Paper.pdf) (Sculley et al., NeurIPS 2015) — 9 pages, read all of it. The CACE principle and "glue code" sections are most relevant.
- [ ] [Delta Lake: High-Performance ACID Table Storage over Cloud Object Stores](https://www.vldb.org/pvldb/vol13/p3411-armbrust.pdf) (Armbrust et al., VLDB 2020) — read Sections 1–4. Understand why versioning and ACID matter for ML pipelines, not just OLTP.

**Key question:** The paper says "changing anything changes everything" (CACE). Give a concrete example from an ML pipeline where this would cause a silent, hard-to-debug failure.

---

## Code

Project: `code/feature-pipeline/` (Python 3.11+, `uv` or `pip`)

Dependencies: `pandas`, `pyarrow`, `duckdb` — no ML libraries.

Scenario: raw user activity events → session features → versioned feature store.

- [ ] `generate_events.py` — generate 10k synthetic events as `(user_id: int, event_type: str, timestamp: int, value: float)`. Write to `data/raw/events.parquet`.
- [ ] `feature_engineering.py` — read raw events, compute per-user features: `(user_id, session_count, avg_value, max_value, last_seen_timestamp)`. Write versioned output to `data/features/v{N}/features.parquet` where N is a monotonically increasing version number (read current version from `data/features/latest.txt`).
- [ ] `feature_store.py` — class `FeatureStore` with:
  - `write(version: int, df: pd.DataFrame)` — writes Parquet + updates `latest.txt`
  - `read(version: int) -> pd.DataFrame` — reads a specific version
  - `latest() -> pd.DataFrame` — reads current version
  - `diff(v1: int, v2: int) -> pd.DataFrame` — returns rows that changed between versions (join on `user_id`, compare feature values)
- [ ] `query.py` — use DuckDB to run SQL against the versioned Parquet files: (1) find top 10 users by `avg_value` in the latest version; (2) find users whose `session_count` changed between v1 and v2
- [ ] `pipeline.py` — end-to-end script: generate → engineer features (v1) → modify 10% of raw events → re-engineer (v2) → print diff

**Constraints:** no MLflow, no DVC, no feature store library. Implement versioning yourself.

---

## Reflect

**What clicked:**

**What surprised me:**

**What would break at scale that works fine here?**

**How a streaming system could replace the batch feature pipeline:**

**What I'd do differently:**
