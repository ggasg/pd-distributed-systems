# W08 — Query Execution

> **Arc:** Streaming and Dataflow · **Language:** Rust

## What you'll build
A vectorized query executor in Rust: columnar filter + hash join + projection over in-memory data. Benchmark it against a row-at-a-time version of the same pipeline.

---

## Read
- [ ] [Volcano — An Extensible and Parallel Query Evaluation System](https://dl.acm.org/doi/10.1109/69.273032) (Graefe, 1994) — read Sections 1–3. This defines the iterator model (the `next()` interface) that every query engine for 20 years was built on.
- [ ] [MonetDB/X100: Hyper-Pipelining Query Execution](https://www.cidrdb.org/cidr2005/papers/P19.pdf) (Boncz et al., CIDR 2005) — read Sections 1–3. This is the argument for vectorized execution and why Volcano is CPU-cache unfriendly.

**Key question:** Why does calling `next()` once per row hurt CPU performance even when the logic is simple? What does processing a batch of 1024 rows at a time fix?

---

## Code

Project: `code/query-exec/` (Rust, Cargo)

Data model: a table of `(id: u32, dept: u32, salary: u32)` rows, stored as three separate `Vec<u32>` (columnar).

**Row-at-a-time executor (baseline):**
- [ ] `row_executor.rs` — struct-of-arrays input; iterator that yields one `(id, dept, salary)` tuple at a time; apply filter `salary > threshold`, then project `(id, salary)`

**Vectorized executor:**
- [ ] `column_filter.rs` — takes `&[u32]` slice, threshold, returns `Vec<bool>` selection mask (branchless: use `map(|v| (v > threshold) as u8)`)
- [ ] `column_project.rs` — applies a selection mask to a column, returning only selected values
- [ ] `hash_join.rs` — build hash table from left side `(key, value)` pairs; probe with right side; emit matching pairs. Both sides as `&[u32]` slices.
- [ ] `bench.rs` — use `std::time::Instant` to benchmark filter + project pipeline on 1M rows, row-at-a-time vs vectorized. Print throughput (rows/sec).

**Expected outcome:** vectorized should be at least 3–5x faster. If not, check if the compiler is auto-vectorizing the row version (add `#[no_vectorize]` or check assembly with `cargo asm`).

---

## Reflect

**What clicked:**

**What surprised me:**

**Benchmark results:**
- Row-at-a-time: __ M rows/sec
- Vectorized: __ M rows/sec
- Speedup: __x

**What does this tell you about how Materialize executes queries?**

**What I'd do differently:**
