# Distributed, Data-Intensive Systems — Engineering Curriculum

A self-directed curriculum for software engineers who want pragmatic mastery of distributed and data-intensive systems, with a focus on streaming, ML infrastructure, compute-intensive execution, and production observability. Every week has a specific paper to read, a concrete coding task, and a deliverable.

**Not for:** people who want to pass system design interviews.
**For:** engineers who want to build real distributed systems and understand them from the inside out.

---

## Quick Start

1. Clone the repo and open it as an Obsidian vault
2. Follow [SETUP.md](SETUP.md) to install Java, Scala, Python, Go, Docker, and Obsidian plugins
3. Set `start_date` in [config.md](config.md) — all week dates recalculate automatically
4. Open [Home.md](Home.md) as your daily entry point
5. Start at W00 (infrastructure setup), then W01

---

## Structure

17 weeks across 4 arcs (plus a W00 setup week), and an optional W17 grand capstone. 2h/day, 5 days/week.

| Arc | Weeks | Focus | Language |
|-----|-------|-------|----------|
| Setup | W00 | Local k8s, Prometheus, Grafana | Go |
| Data Systems Internals | W01–W04 | Storage engines, encoding, MapReduce, causality | Java 21 |
| Streaming and Dataflow | W05–W08 | Stream processing, Naiad, Differential Dataflow, query execution | Scala 2.13 |
| Distributed ML & Compute | W09–W14 | ML pipelines, distributed training, GPU compute, transformers, fault tolerance | Python / Go secondary |
| Infrastructure | W15–W16 | Kubernetes Operators, observability (Prometheus, OTel, Grafana) | Go / Scala |
| Capstone (optional) | W17 | Distributed training + serving platform, fully observed — synthesizes W09, W10, W12, W13, W15, W16 | Go / Python |

---

## What You'll Be Able to Do After Each Arc

**After Arc 1 (W01–W04):**
- Explain why LSM-trees beat B-trees for write-heavy workloads and when they don't
- Implement varint encoding and measure column vs. row scan performance
- Write a MapReduce framework with virtual threads; explain why iterative algorithms are slow on it
- Implement vector clocks; reason about causal consistency and concurrent events

**After Arc 2 (W05–W08):**
- Build a streaming windowed aggregator with watermarks; explain what "late data" means
- Implement Naiad's timestamp and progress-tracking model from the paper
- Build a Differential Dataflow engine from scratch: incremental word count + reachability
- Benchmark vectorized vs. row-at-a-time query execution; explain the 3–8x gap

**After Arc 3 (W09–W14):**
- Design and implement a versioned ML feature store with Parquet + DuckDB
- Implement ring-allreduce over raw TCP sockets; explain the bandwidth math
- Write a tiled CUDA matmul with Numba; read a roofline chart
- Implement multi-head attention and KV cache from scratch in NumPy
- Implement Chandy-Lamport distributed snapshots; explain what "consistent cut" means
- Complete a capstone that combines at least two arcs

**After Arc 4 (W15–W16):**
- Write a Kubernetes Operator in Go with a custom CRD and reconcile loop
- Instrument a distributed system with Prometheus metrics + OpenTelemetry traces
- Build a Grafana dashboard from scratch; explain the four golden signals

---

## Each Week

Every week has:
- **Read** — one or two named papers or chapters, with specific sections called out
- **Code** — a concrete implementation task with named files and a clear deliverable
- **🐍 Python DSA Review** — optional; a short Python warmup of the underlying algorithm (W01–W08, W10, W13)
- **Reflect** — what you built, what surprised you, what you'd do differently

---

## Language Map

| Weeks | Language | Why |
|-------|----------|-----|
| W00 | Go | Service + k8s deployment; Prometheus metrics |
| W01–W04 | Java 21 | Virtual threads, records — modern concurrency primitives |
| W05–W08 | Scala 2.13 | FP, algebraic types — natural for dataflow and incremental computation |
| W09–W12 | Python | ML ecosystem, numerical computing, Numba for GPU |
| W13–W14 | Java 21 / Scala / Python | Depends on capstone option |
| W15 | Go | Operators are almost exclusively written in Go |
| W16 | Scala + Go | Instrument existing Scala code; Go sidecar optional |
| W03, W10, W11, W13, W14 | Go (secondary) | Automation tools, coordination services |

---

## Repository Layout

```
.
├── Home.md               # Daily entry point — open this in Obsidian
├── config.md             # Set start_date here
├── README.md             # This file
├── SETUP.md              # Environment setup (Java, Scala, Python, Go, Docker, Obsidian)
├── RESOURCES.md          # All papers and books, by week, with free links
├── CONTEXT.md            # Session context for AI-assisted study sessions
├── weeks/                # One .md file per week (W00–W16)
├── code/                 # Your implementations — see code/README.md
├── posts/                # Weekly blog posts — see posts/TEMPLATE.md
├── tools/                # Go automation tools (plan-dates, job_coordinator, grad_server)
└── Templates/            # week-template.md for adding custom weeks
```

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
- [W17 — Grand Capstone: Distributed Training & Serving Platform (optional)](weeks/W17-capstone-platform.md)

---

## Adapting This Curriculum

**Only 1h/day?** Focus on Read + Reflect each week; treat Code as optional. Prioritize W01, W03, W05, W07, W12 — those give the most conceptual leverage.

**Skip the infrastructure arc?** W00, W15, W16 are independent — you can complete W01–W14 without touching Kubernetes. Come back to Arc 4 when it's relevant to your work.

**Add your own week?** Copy `Templates/week-template.md`, set `week_number` in frontmatter, and it appears in the Home.md dashboard automatically.

**Different languages?** The algorithms are language-agnostic. The Java weeks could be Go; the Scala weeks could be Haskell or OCaml; the Python weeks could be Julia. The language choices are justified in the Language Map above, but they're not sacred.

---

## Prerequisites

- Comfortable with at least one systems language (Java, Go, C++, Rust)
- Knows what a hash map and B-tree are
- Has written concurrent code before (threads, async, actors, etc.)
- Familiar with basic algorithms (sorting, BFS, binary search)

No PhD required. No ML background required for the early arcs.
