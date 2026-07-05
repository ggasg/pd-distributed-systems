# Distributed, Data-Intensive Systems — Engineering Curriculum

A self-directed curriculum for software engineers who want pragmatic mastery of distributed and data-intensive systems, with a focus on streaming, ML infrastructure, compute-intensive execution, and production infrastructure. Every week has a specific paper to read, a concrete coding task, and a deliverable.

**Not for:** people who want to pass system design interviews.
**For:** engineers who want to build real distributed systems and understand them from the inside out.

---

## Structure

17 weeks across 4 arcs (plus a W00 setup week). 2h/day, 5 days/week.

| Arc | Weeks | Focus |
|-----|-------|-------|
| Setup | W00 | Local k8s cluster, Prometheus, Grafana, hello-metrics Go service |
| Data Systems Internals | W01–W04 | Storage engines, encoding, MapReduce, vector clocks |
| Streaming and Dataflow | W05–W08 | Stream processing, Naiad, Differential Dataflow, query execution |
| Distributed ML & Compute | W09–W14 | ML pipelines, distributed training, GPU compute, transformers, fault tolerance |
| Infrastructure | W15–W16 | Kubernetes Operators, observability (metrics, tracing, logging) |

## Dates

Set `start_date` in `config.md` — all week dates recalculate automatically in the Obsidian dashboard. To print the full schedule:

```
go run tools/plan-dates.go --start 2026-07-06
```

## Each Week

Every week has:
- **Read** — one or two named papers or chapters, with specific sections called out
- **Code** — a concrete implementation task with named files and a clear deliverable
- **Reflect** — what you built, what surprised you, what you'd do differently

## Language Map

| Weeks | Language | Why |
|-------|----------|-----|
| W00 | Go | Tooling, Docker, k8s deployment |
| W01–W04 | Java 21 | Virtual threads, records — modern concurrency primitives |
| W05–W08 | Scala | FP, algebraic types — natural for dataflow and incremental computation |
| W09–W12 | Python | ML ecosystem, numerical computing, Numba for GPU |
| W13–W14 | Java 21 / Scala / Python | Depends on capstone option |
| W15 | Go | Operators are almost exclusively written in Go |
| W16 | Scala + Go | Instrument existing code; Go sidecar optional |
| W03, W10, W11, W13, W14 | Go (secondary) | Automation tools, coordination services |

---

## Weeks

- [W00 — Infrastructure Setup](weeks/W00-setup.md)
- [W01 — LSM-Trees and Storage Engines](weeks/W01-lsm-storage.md)
- [W02 — Encoding and Wire Formats](weeks/W02-encoding.md)
- [W03 — MapReduce and Its Limits](weeks/W03-mapreduce.md)
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
- [W15 — Kubernetes and Operators](weeks/W15-kubernetes-operators.md)
- [W16 — Observability: Metrics, Tracing, Logging](weeks/W16-observability.md)

---

## How to Use This

1. Clone the repo
2. Open it as an Obsidian vault (`.obsidian/` config is included)
3. Set `start_date` in `config.md`
4. Open `Home.md` as your daily entry point — it auto-detects the current week
5. Code goes in a sibling `code/` directory or your own repo

---

## Prerequisites

- Comfortable with at least one systems language (Java, Go, C++)
- Knows what a hash map and B-tree are
- Has written concurrent code before (threads, async, etc.)
- Familiar with basic probability and algorithms

No PhD required. No ML background required for the early arcs.
