# Distributed, Data-Intensive Systems: Engineering Curriculum

A self-directed curriculum for software engineers who want pragmatic mastery of distributed and data-intensive systems, with a focus on streaming, query planning, ML infrastructure, compute-intensive execution, and production observability. Every week has a specific paper to read, a concrete coding task, and a deliverable.

This isn't for people who want to pass system design interviews. It's for engineers who want to build real distributed systems and understand them from the inside out.

---

## Quick Start

1. Clone the repo and open it as an Obsidian vault
2. Follow [SETUP.md](SETUP.md) to install Go, C++, Scala, Python, Docker, and Obsidian plugins
3. Skim [RESOURCES.md](RESOURCES.md): most readings are free links, but a few (DDIA, the most-used book) are worth buying before you hit the week that needs them
4. Set `start_date` in [config.md](config.md), and all week dates recalculate automatically
5. Open [Home.md](Home.md) as your daily entry point
6. Start at W00 (infrastructure setup), then W01

---

## Structure

21 weeks across 4 arcs (plus a W00 setup week), and an optional W21 grand capstone. 2h/day, 5 days/week: roughly 4.8 months for the core curriculum, 5.1 with the optional capstone.

| Arc | Weeks | Focus | Language |
|-----|-------|-------|----------|
| Setup | W00 | Local k8s, Prometheus, Grafana | Go |
| Data Systems Internals | W01–W04 | Storage engines, encoding, MapReduce, causality | Go |
| Streaming, Dataflow, and Query Planning | W05–W10 | Stream processing, Naiad, Differential Dataflow, query execution, rule-based query planning, aggregation algebra | C++ (W05–W08) / Scala (W09–W10) |
| Distributed ML & Compute | W11–W18 | ML pipelines, PySpark vs. Scala Spark performance, distributed training, actor model (Ray), GPU compute, transformers, fault tolerance | Python / Scala + Python (W12) / Go (W17) |
| Infrastructure | W19–W20 | Kubernetes Operators, observability (Prometheus, OTel, Grafana) | Go / C++ |
| Capstone (optional) | W21 | Distributed training + serving platform, fully observed (synthesizes W11, W13, W16, W17, W19, W20) | Go / Python |

---

## What You'll Be Able to Do After Each Arc

**After Arc 1 (W01–W04):**
- Explain why LSM-trees beat B-trees for write-heavy workloads and when they don't
- Implement varint encoding and measure column vs. row scan performance
- Write a MapReduce framework with goroutines; explain why iterative algorithms are slow on it
- Implement vector clocks; reason about causal consistency and concurrent events

**After Arc 2 (W05–W10):**
- Build a streaming windowed aggregator with watermarks; explain what "late data" means
- Implement Naiad's timestamp and progress-tracking model from the paper
- Build the core of a Differential Dataflow engine from scratch (incremental word count), then build and benchmark an incremental materialized view against a full-recompute baseline, the same trade-off ClickHouse and Spark Structured Streaming make in production
- Benchmark vectorized vs. row-at-a-time query execution; explain the 3–8x gap
- Build a toy rule-based query optimizer in Scala (case classes + pattern matching + a `transform` combinator), the same technique Spark's real Catalyst optimizer uses
- Implement a `Semigroup`/`Monoid` typeclass hierarchy from scratch and explain why associativity, not commutativity, is what makes a distributed reduction safe to compute as a tree instead of strictly left-to-right

**After Arc 3 (W11–W18):**
- Design and implement a versioned ML feature store with Parquet + DuckDB
- Measure, not guess, where PySpark's performance actually diverges from Scala Spark, and explain the mechanism (JVM boundary crossings in row-at-a-time UDFs) rather than just the folklore
- Implement ring-allreduce over raw TCP sockets; explain the bandwidth math
- Build a stateful actor system with Ray; explain why actors (not stateless tasks) are the right abstraction for coordinating training workers
- Write a tiled CUDA matmul with Numba; read a roofline chart
- Implement multi-head attention and KV cache from scratch in NumPy
- Implement Chandy-Lamport distributed snapshots; explain what "consistent cut" means
- Complete a capstone that combines at least two arcs

**After Arc 4 (W19–W20):**
- Write a Kubernetes Operator in Go with a custom CRD and reconcile loop
- Instrument a distributed system with Prometheus metrics and OpenTelemetry traces
- Build a Grafana dashboard from scratch; explain the four golden signals

---

## Each Week

Every week has:
- **Read**: one or two named papers or chapters, with specific sections called out
- **Code**: a concrete implementation task with named files and a clear deliverable
- **🐍 Python DSA Review**: optional, a short Python warmup of the underlying algorithm (W01–W08, W13, W14, W17)
- **Reflect**: what you built, what surprised you, what you'd do differently

---

## Language Map

| Weeks | Language | Why |
|-------|----------|-----|
| W00 | Go | Service + k8s deployment; Prometheus metrics |
| W01–W04 | Go | Storage engines and coordination logic; goroutines/channels for the concurrent parts, plain structs and interfaces for the data structures |
| W05–W08 | C++ | This is the substrate of the systems the curriculum's actual target (distributed model training, compute-intensive AI workflows) runs on: PyTorch's `c10d`/ATen, NCCL, gRPC's core, and DuckDB's execution engine are all C++. The dataflow papers this arc is built around (Naiad, Differential Dataflow) have no maintained C++ reference implementation the way they do a Rust one (`timely-dataflow`/`differential-dataflow`), so this arc leans on the papers directly and points at adjacent production C++ codebases (PyTorch's autograd engine, DuckDB) instead of a source-level companion |
| W09–W10 | Scala | Spark itself is Scala, and its query optimizer (Catalyst) and its "abstract algebra for big data" aggregation story (Algebird) are both genuinely built the way these two weeks have you build toy versions: case classes, pattern matching, typeclasses. Low ramp cost given prior production Spark/Scala experience; this is a formalization of existing intuition, not a fresh language investment |
| W11 | Python | ML ecosystem, numerical computing |
| W12 | Scala + Python (PySpark) | Direct continuation of W09–W10: runs real Spark, in both languages, on the identical job, to measure where the host language actually costs you (UDFs) and where it doesn't (the DataFrame API, same Catalyst plan either way) |
| W13–W16 | Python | ML ecosystem, numerical computing, Ray for distributed actors, Numba for GPU |
| W17 | Go | Native channels are FIFO by construction, a direct fit for Chandy-Lamport's marker protocol |
| W18 | Go / C++ / Python | Depends on capstone option |
| W19 | Go | Operators are almost exclusively written in Go |
| W20 | C++ + Go | Instrument existing C++ code (the W07 DD engine) with `prometheus-cpp` and `opentelemetry-cpp`; Go log-aggregator built and wired in as a sidecar on the W19 operator |
| W03, W13, W15 | Go (secondary) | Automation tools, coordination services |

---

## Repository Layout

```
.
├── Home.md               # Daily entry point, open this in Obsidian
├── config.md             # Set start_date here
├── README.md             # This file
├── SETUP.md              # Environment setup (Go, C++, Scala, Python, Docker, Obsidian)
├── RESOURCES.md          # All papers and books, by week, with free links
├── CONTEXT.md            # Session context for AI-assisted study sessions
├── weeks/                # One .md file per week (W00–W20)
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
- [W09: Rule-Based Query Planning in Scala](weeks/W09-query-planning.md)
- [W10: Aggregation Algebra: Monoids and Semigroups](weeks/W10-aggregation-algebra.md)
- [W11: ML Data Pipelines](weeks/W11-ml-pipelines.md)
- [W12: PySpark vs. Scala Spark: Where the JVM Boundary Costs You](weeks/W12-spark-lang-bench.md)
- [W13: Distributed Training](weeks/W13-distributed-training.md)
- [W14: The Actor Model and Ray](weeks/W14-actor-model-ray.md)
- [W15: GPU Memory and Compute](weeks/W15-gpu-compute.md)
- [W16: Attention and KV Cache](weeks/W16-attention.md)
- [W17: Fault Tolerance and Snapshots](weeks/W17-fault-tolerance.md)
- [W18: Capstone](weeks/W18-capstone.md)
- [W19: Kubernetes Operators](weeks/W19-kubernetes-operators.md)
- [W20: Observability: Metrics, Tracing, Logging](weeks/W20-observability.md)
- [W21: Grand Capstone: Distributed Training & Serving Platform (optional)](weeks/W21-capstone-platform.md)

---

## Adapting This Curriculum

**Only 1h/day?** Focus on Read + Reflect each week; treat Code as optional. Prioritize W01, W03, W05, W07, W14, W16; those give the most conceptual leverage.

**Skip the infrastructure arc?** W00, W19, W20 are independent. You can complete W01–W18 without touching Kubernetes, and come back to Arc 4 when it's relevant to your work.

**Add your own week?** Copy `Templates/week-template.md`, set `week_number` in frontmatter, and it appears in the Home.md dashboard automatically.

**Tracking progress separately from curriculum edits?** Keep `main` for curriculum changes and a separate `progress` branch for checked-off tasks and Reflect answers. See the Branch Workflow section in [CONTEXT.md](CONTEXT.md) for how to merge updates between them.

**Different languages?** The algorithms are language-agnostic. This curriculum is built around four languages, each for a specific reason: Go as the one genuinely new language to gain fluency in; C++ as a deliberate refresh into modern idioms (smart pointers, move semantics, RAII, templates) rather than a cold start; Scala for a short, focused module where the real production system (Spark) is itself written in Scala, making it worth the two-week investment even given prior Scala familiarity; and Python for the ML-native arc. Substitutions: the Go weeks could be Java if you'd rather stay on a GC'd/OOP-familiar language; the C++ weeks could be Rust if you'd rather trade manual memory management for a borrow-checked model (Rust is in fact the closer conceptual fit for W05–W08, since ownership and algebraic enums map naturally onto dataflow and incremental computation; it was the original choice here, dropped in favor of C++ for tighter alignment with this curriculum's actual target of distributed training and compute-intensive AI workflows, not because it's the wrong tool for the topic); the Python weeks could be Julia. The language choices are justified in the Language Map above, but they're not sacred.

---

## Prerequisites

- Comfortable programming in at least one language, in any paradigm. This curriculum is explicitly meant to be your on-ramp into Go, and a refresh back into modern C++ and Scala, not something that assumes you already know them
- Knows what a hash map and B-tree are
- Has written concurrent code before (threads, async, actors, etc.) in whatever language you already know
- Familiar with basic algorithms (sorting, BFS, binary search)

No PhD required. No ML background required for the early arcs.

**New to Go? Rusty on C++? Already know Scala?** Go is the genuinely new language here: see its ramp notes in [SETUP.md](SETUP.md) before W00, and don't try to learn it and LSM-trees simultaneously on day one. C++ is different: if you learned it years ago (school, an earlier job) and haven't touched it since, W05 is a refresh into modern idioms, not a cold start; budget real time regardless, both for the idioms that changed and for CMake, which has no equivalent to Cargo's zero-config build experience. Scala (W09–W10) is the lowest-ramp of the three if you already have production Spark/Scala experience: these two weeks are meant to formalize FP intuition you likely already use, not teach the language from zero. See the language-specific sections of [SETUP.md](SETUP.md) for what to review before each.

---

## Licensing

This curriculum's written content (everything under `weeks/`, `Templates/`, `posts/`, and this documentation) is licensed under [CC BY 4.0](LICENSE): free to use, adapt, and build on, including for your own trainings, courses, or videos, provided you credit Gaston Guitart as the original author. See [CITATION.cff](CITATION.cff) for the exact attribution format.

Code under `code/` and `tools/` is licensed separately under the [MIT License](LICENSE-MIT), permissive with no attribution requirement beyond keeping the copyright notice intact.
