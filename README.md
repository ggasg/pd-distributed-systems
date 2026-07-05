# Distributed, Data-Intensive Systems — 14-Week Engineering Plan

A self-directed curriculum for software engineers who want pragmatic mastery of distributed and data-intensive systems, with a focus on streaming, ML infrastructure, and compute-intensive execution. Every week has a specific paper to read, a concrete coding task, and a deliverable.

**Not for:** people who want to pass system design interviews.
**For:** engineers who want to build real distributed systems and understand them from the inside out.

---

## Structure

14 weeks across 3 arcs. 2h/day, 5 days/week.

| Arc | Weeks | Focus |
|-----|-------|-------|
| Data Systems Internals | W01–W04 | Storage engines, encoding, consensus, clocks |
| Streaming and Dataflow | W05–W08 | Stream processing, Naiad, Differential Dataflow, query execution |
| Distributed ML & Compute | W09–W14 | ML pipelines, distributed training, GPU compute, transformers, fault tolerance |

## Each Week

Every week has:
- **Read** — one or two named papers or chapters, with specific sections called out
- **Code** — a concrete implementation task with named files and a clear deliverable
- **Reflect** — what you built, what surprised you, what you'd do differently

## Language Map

| Weeks | Language | Why |
|-------|----------|-----|
| W01–W04 | Java 21 | Virtual threads, records, StructuredTaskScope — modern concurrency primitives |
| W05–W06 | Scala | Algebraic data types, functional composition — natural fit for dataflow |
| W07–W08 | Rust | DD crate is in Rust; vectorized execution benefits from Rust's control |
| W09–W14 | Python / CUDA | ML ecosystem; GPU kernels |

---

## Weeks

- [W01 — LSM-Trees and Storage Engines](weeks/W01-lsm-storage.md)
- [W02 — Encoding and Wire Formats](weeks/W02-encoding.md)
- [W03 — Raft Consensus](weeks/W03-raft.md)
- [W04 — Clocks, Causality, and Time](weeks/W04-clocks.md)
- [W05 — Stream Processing Primitives](weeks/W05-streaming.md)
- [W06 — Naiad and Timely Dataflow](weeks/W06-naiad.md)
- [W07 — Differential Dataflow](weeks/W07-differential-dataflow.md)
- [W08 — Query Execution](weeks/W08-query-execution.md)
- [W09 — ML Data Pipelines](weeks/W09-ml-pipelines.md)
- [W10 — Distributed Training](weeks/W10-distributed-training.md)
- [W11 — GPU Memory and Compute](weeks/W11-gpu-compute.md)
- [W12 — Attention and KV Cache](weeks/W12-attention.md)
- [W13 — Fault Tolerance and Snapshots](weeks/W13-fault-tolerance.md)
- [W14 — Capstone](weeks/W14-capstone.md)

---

## How to Use This

1. Clone the repo
2. Open it as an Obsidian vault (`.obsidian/` config is included)
3. Start at W01, work through the checklist
4. Fill in the Reflect section at the end of each week
5. Code goes in a sibling `code/` directory or your own repo

---

## Prerequisites

- Comfortable with at least one systems language (Java, Go, Rust, C++)
- Knows what a hash map and B-tree are
- Has written concurrent code before (threads, async, etc.)
- Familiar with basic probability and algorithms

No PhD required. No ML background required for the early arcs.
