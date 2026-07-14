# Distributed, Data-Intensive Systems: Engineering Curriculum

A self-directed curriculum for software engineers who want pragmatic mastery of distributed and data-intensive systems, with a focus on streaming, ML infrastructure, compute-intensive execution, and production observability. Every week has a specific paper to read, a concrete coding task, and a deliverable.

This isn't for people who want to pass system design interviews. It's for engineers who want to build real distributed systems and understand them from the inside out.

---

## Quick Start

1. Clone the repo and open it as an Obsidian vault
2. Follow [SETUP.md](SETUP.md) to install Go, C++, Python, Docker, and Obsidian plugins
3. Skim [RESOURCES.md](RESOURCES.md): most readings are free links, but a few (DDIA, the most-used book) are worth buying before you hit the week that needs them
4. Set `start_date` in [config.md](config.md), and all week dates recalculate automatically
5. Open [Home.md](Home.md) as your daily entry point
6. Start at W00 (infrastructure setup), then W01

---

## Structure

18 weeks across 4 arcs (plus a W00 setup week), and an optional W18 grand capstone. 2h/day, 5 days/week.

| Arc | Weeks | Focus | Language |
|-----|-------|-------|----------|
| Setup | W00 | Local k8s, Prometheus, Grafana | Go |
| Data Systems Internals | W01–W04 | Storage engines, encoding, MapReduce, causality | Go |
| Streaming and Dataflow | W05–W08 | Stream processing, Naiad, Differential Dataflow, query execution | C++ |
| Distributed ML & Compute | W09–W15 | ML pipelines, distributed training, actor model (Ray), GPU compute, transformers, fault tolerance | Python / Go (W14) |
| Infrastructure | W16–W17 | Kubernetes Operators, observability (Prometheus, OTel, Grafana) | Go / C++ |
| Capstone (optional) | W18 | Distributed training + serving platform, fully observed (synthesizes W09, W10, W13, W14, W16, W17) | Go / Python |

---

## What You'll Be Able to Do After Each Arc

**After Arc 1 (W01–W04):**
- Explain why LSM-trees beat B-trees for write-heavy workloads and when they don't
- Implement varint encoding and measure column vs. row scan performance
- Write a MapReduce framework with goroutines; explain why iterative algorithms are slow on it
- Implement vector clocks; reason about causal consistency and concurrent events

**After Arc 2 (W05–W08):**
- Build a streaming windowed aggregator with watermarks; explain what "late data" means
- Implement Naiad's timestamp and progress-tracking model from the paper
- Build the core of a Differential Dataflow engine from scratch (incremental word count), then build and benchmark an incremental materialized view against a full-recompute baseline — the same trade-off Snowflake, Databricks, and ClickHouse make in production
- Benchmark vectorized vs. row-at-a-time query execution; explain the 3–8x gap

**After Arc 3 (W09–W15):**
- Design and implement a versioned ML feature store with Parquet + DuckDB
- Implement ring-allreduce over raw TCP sockets; explain the bandwidth math
- Build a stateful actor system with Ray; explain why actors (not stateless tasks) are the right abstraction for coordinating training workers
- Write a tiled CUDA matmul with Numba; read a roofline chart
- Implement multi-head attention and KV cache from scratch in NumPy
- Implement Chandy-Lamport distributed snapshots; explain what "consistent cut" means
- Complete a capstone that combines at least two arcs

**After Arc 4 (W15–W16):**
- Write a Kubernetes Operator in Go with a custom CRD and reconcile loop
- Instrument a distributed system with Prometheus metrics and OpenTelemetry traces
- Build a Grafana dashboard from scratch; explain the four golden signals

---

## Each Week

Every week has:
- **Read**: one or two named papers or chapters, with specific sections called out
- **Code**: a concrete implementation task with named files and a clear deliverable
- **🐍 Python DSA Review**: optional, a short Python warmup of the underlying algorithm (W01–W08, W10, W13)
- **Reflect**: what you built, what surprised you, what you'd do differently

---

## Language Map

| Weeks | Language | Why |
|-------|----------|-----|
| W00 | Go | Service + k8s deployment; Prometheus metrics |
| W01–W04 | Go | Storage engines and coordination logic; goroutines/channels for the concurrent parts, plain structs and interfaces for the data structures |
| W05–W08 | C++ | This is the substrate of the systems the curriculum's actual target (distributed model training, compute-intensive AI workflows) runs on — PyTorch's `c10d`/ATen, NCCL, gRPC's core, and DuckDB's execution engine are all C++. The dataflow papers this arc is built around (Naiad, Differential Dataflow) have no maintained C++ reference implementation the way they do a Rust one (`timely-dataflow`/`differential-dataflow`), so this arc leans on the papers directly and points at adjacent production C++ codebases (PyTorch's autograd engine, DuckDB) instead of a source-level companion |
| W09–W13 | Python | ML ecosystem, numerical computing, Ray for distributed actors, Numba for GPU |
| W14 | Go | Native channels are FIFO by construction, a direct fit for Chandy-Lamport's marker protocol |
| W15 | Go / C++ / Python | Depends on capstone option |
| W16 | Go | Operators are almost exclusively written in Go |
| W17 | C++ + Go | Instrument existing C++ code (the W07 DD engine) with `prometheus-cpp` and `opentelemetry-cpp`; Go log-aggregator built and wired in as a sidecar on the W16 operator |
| W03, W10, W12, W15 | Go (secondary) | Automation tools, coordination services |

---

## Repository Layout

```
.
├── Home.md               # Daily entry point, open this in Obsidian
├── config.md             # Set start_date here
├── README.md             # This file
├── SETUP.md              # Environment setup (Go, C++, Python, Docker, Obsidian)
├── RESOURCES.md          # All papers and books, by week, with free links
├── CONTEXT.md            # Session context for AI-assisted study sessions
├── weeks/                # One .md file per week (W00–W17)
├── code/                 # Your implementations, see code/README.md
├── posts/                # Weekly blog posts, see posts/TEMPLATE.md
├── tools/                # Go automation tools (plan-dates, job_coordinator, grad_server)
└── Templates/            # week-template.md for adding custom weeks
```

---

## Weeks

- [W00: Infrastructure Setup](weeks/W00-setup.md)
- [W01: LSM-Trees and Storage Engines](weeks/W01-lsm-storage.md)
- [W02: Encoding and Wire Formats](weeks/W02-encoding.md)
- [W03: MapReduce and Its Limits](weeks/W03-mapreduce.md)
- [W04: Clocks, Causality, and Time](weeks/W04-clocks.md)
- [W05: Stream Processing Primitives](weeks/W05-streaming.md)
- [W06: Naiad and Timely Dataflow](weeks/W06-naiad.md)
- [W07: Differential Dataflow](weeks/W07-differential-dataflow.md)
- [W08: Query Execution](weeks/W08-query-execution.md)
- [W09: ML Data Pipelines](weeks/W09-ml-pipelines.md)
- [W10: Distributed Training](weeks/W10-distributed-training.md)
- [W11: The Actor Model and Ray](weeks/W11-actor-model-ray.md)
- [W12: GPU Memory and Compute](weeks/W12-gpu-compute.md)
- [W13: Attention and KV Cache](weeks/W13-attention.md)
- [W14: Fault Tolerance and Snapshots](weeks/W14-fault-tolerance.md)
- [W15: Capstone](weeks/W15-capstone.md)
- [W16: Kubernetes and Operators](weeks/W16-kubernetes-operators.md)
- [W17: Observability: Metrics, Tracing, Logging](weeks/W17-observability.md)
- [W18: Grand Capstone: Distributed Training & Serving Platform (optional)](weeks/W18-capstone-platform.md)

---

## Adapting This Curriculum

**Only 1h/day?** Focus on Read + Reflect each week; treat Code as optional. Prioritize W01, W03, W05, W07, W11, W13; those give the most conceptual leverage.

**Skip the infrastructure arc?** W00, W16, W17 are independent. You can complete W01–W15 without touching Kubernetes, and come back to Arc 4 when it's relevant to your work.

**Add your own week?** Copy `Templates/week-template.md`, set `week_number` in frontmatter, and it appears in the Home.md dashboard automatically.

**Tracking progress separately from curriculum edits?** Keep `main` for curriculum changes and a separate `progress` branch for checked-off tasks and Reflect answers. See the Branch Workflow section in [CONTEXT.md](CONTEXT.md) for how to merge updates between them.

**Different languages?** The algorithms are language-agnostic. This curriculum is built around Go as the one genuinely new language to gain fluency in, C++ as a deliberate refresh into modern idioms (smart pointers, move semantics, RAII, templates) rather than a cold start, and Python for the ML-native arc — but the Go weeks could be Java if you'd rather stay on a GC'd/OOP-familiar language; the C++ weeks could be Rust, Scala, or Java if you'd rather trade manual memory management for a borrow-checked or GC'd model (Rust in particular is the closer conceptual fit for W05–W08, since ownership and algebraic enums map naturally onto dataflow and incremental computation — it was the original choice here, dropped in favor of C++ for tighter alignment with this curriculum's actual target of distributed training and compute-intensive AI workflows, not because it's the wrong tool for the topic); the Python weeks could be Julia. The language choices are justified in the Language Map above, but they're not sacred.

---

## Prerequisites

- Comfortable programming in at least one language, in any paradigm — this curriculum is explicitly meant to be your on-ramp into Go, and a refresh back into modern C++, not something that assumes you already know them
- Knows what a hash map and B-tree are
- Has written concurrent code before (threads, async, actors, etc.) in whatever language you already know
- Familiar with basic algorithms (sorting, BFS, binary search)

No PhD required. No ML background required for the early arcs.

**New to Go? Rusty on C++?** Go is the genuinely new language here — see its ramp notes in [SETUP.md](SETUP.md) before W00, and don't try to learn it and LSM-trees simultaneously on day one. C++ is different: if you learned it years ago (school, an earlier job) and haven't touched it since, W05 is a refresh into modern idioms, not a cold start — budget real time regardless, both for the idioms that changed and for CMake, which has no equivalent to Cargo's zero-config build experience. See the C++ section of [SETUP.md](SETUP.md) for what specifically to review before W05.
