# W07 — Differential Dataflow

> **Arc:** Streaming and Dataflow · **Language:** Rust

## What you'll build
Two programs using the `differential-dataflow` Rust crate: (1) incremental word count — add a document, observe counts update; (2) incremental graph reachability — add an edge, observe which nodes become reachable.

---

## Read
- [ ] [Differential Dataflow](https://www.cidrdb.org/cidr2013/Papers/CIDR13_Paper111.pdf) (McSherry et al., CIDR 2013) — read Sections 1–3. Section 2 defines the data model (collections as functions from time to multisets of changes). Section 3 defines the operators.
- [ ] DD Rust source — skim these two files to connect paper to code: [`src/collection.rs`](https://github.com/TimelyDataflow/differential-dataflow/blob/master/src/collection.rs) (the `Collection` type), [`src/operators/join.rs`](https://github.com/TimelyDataflow/differential-dataflow/blob/master/src/operators/join.rs) (how incremental join works)

**Key question:** What is a "difference" in DD? How does `(key, value, time, diff)` encode both additions and retractions?

---

## Code

Project: `code/dd-examples/` (Rust, Cargo workspace)

**Program 1: `src/word_count.rs`**
- [ ] Read words from a hardcoded vec of "documents" (each doc is a `&str`)
- [ ] Use `flat_map` to split into words, `count()` to count occurrences
- [ ] In the input loop: add a document at time 1, print counts; retract it at time 2, print updated counts; add a different document at time 3
- [ ] Expected output shows only the delta each round, not a full recount

**Program 2: `src/reachability.rs`**
- [ ] Input: a collection of directed edges `(u32, u32)`
- [ ] Compute reachability: use `iterate` to propagate reachability transitively
- [ ] Start from a fixed set of roots (e.g. node 0)
- [ ] In the input loop: add edge (0→1) at t=1, add edge (1→2) at t=2, retract edge (0→1) at t=3; observe which nodes are reachable at each step

**Constraints:** use `differential_dataflow` 0.12+ and `timely` from crates.io. Single-threaded (`timely::execute_directly`). Print changes as they arrive, not final state.

---

## Reflect

**What clicked:**

**What surprised me:**

**What happens inside DD when you retract a record?**

**How arrangements relate to what you built:**

**How this connects to Materialize's query engine:**

**What I'd do differently:**
