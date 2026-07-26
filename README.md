# Distributed, Data-Intensive Systems: Engineering Curriculum

A self-directed curriculum for software engineers who want pragmatic mastery of distributed and data-intensive systems, with a focus on streaming, query planning, ML infrastructure, compute-intensive execution, and production observability. Every week has a specific paper to read, a concrete coding task, and a deliverable.

This isn't for people who want to pass system design interviews. It's for engineers who want to build real distributed systems and understand them from the inside out.

---

## Quick Start

1. Clone the repo and open it as an Obsidian vault
2. Follow [SETUP.md](SETUP.md) to install Java, Go, Scala, Python, Docker, and Obsidian plugins
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
| Streaming, Dataflow, and Query Planning | W05–W10 | Stream processing, Naiad, Differential Dataflow, query execution, rule-based query planning, aggregation algebra | Java (W05–W07) / Go (W08) / Scala (W09–W10) |
| Distributed ML & Compute | W11–W18 | ML pipelines, PySpark vs. Scala Spark performance, distributed training, actor model (Ray), GPU compute, transformers, fault tolerance | Python / Scala + Python (W12) / Java (W17) |
| Infrastructure | W19–W20 | Kubernetes Operators (KubeRay, Spark Operator), observability (Prometheus, OTel, Grafana) | Go (W19) / Java + Go (W20) |
| Capstone (optional) | W21 | Distributed training + serving platform, fully observed (synthesizes W11, W13, W16, W17, W20; deploys to KubeRay from W19) | Python |

---

## What You'll Be Able to Do After Each Arc

**After Arc 1 (W01–W04):**
- Explain why LSM-trees beat B-trees for write-heavy workloads and when they don't
- Implement varint encoding and measure column vs. row scan performance
- Write a MapReduce framework with goroutines and channels; explain why iterative algorithms are slow on it
- Implement vector clocks; reason about causal consistency and concurrent events

**After Arc 2 (W05–W10):**
- Build a streaming windowed aggregator with watermarks; explain what "late data" means
- Implement Naiad's timestamp and progress-tracking model from the paper
- Build the core of a Differential Dataflow engine from scratch (incremental word count), then build and benchmark an incremental materialized view against a full-recompute baseline, the same trade-off ClickHouse and Spark Structured Streaming make in production
- Benchmark vectorized vs. row-at-a-time query execution; explain the 3–8x gap
- Build a toy rule-based query optimizer in Scala (case classes + pattern matching + a `transform` combinator), the same technique Spark's real Catalyst optimizer uses
- Reason about a `Semigroup`/`Monoid` typeclass hierarchy and implement against it, explaining why associativity, not commutativity, is what makes a distributed reduction safe to compute as a tree instead of strictly left-to-right

**After Arc 3 (W11–W18):**
- Design and implement a versioned ML feature store with Parquet + DuckDB
- Measure, not guess, where PySpark's performance actually diverges from Scala Spark, and explain the mechanism (JVM boundary crossings in row-at-a-time UDFs) rather than just the folklore
- Implement ring-allreduce over raw TCP sockets; explain the bandwidth math
- Build a stateful actor system with Ray; explain why actors (not stateless tasks) are the right abstraction for coordinating training workers
- Write a cache-blocked GEMM kernel in C, measure what compiler auto-vectorization buys you, and read a roofline chart
- Build a KV cache against a given multi-head attention implementation in NumPy; explain and fix a cross-request cache-bleed bug
- Implement Chandy-Lamport distributed snapshots; explain what "consistent cut" means
- Complete a capstone that combines at least two arcs

**After Arc 4 (W19–W20):**
- Deploy, break, and debug two real Kubernetes operators (KubeRay, Kubeflow's Spark Operator); explain a reconcile loop by reading one, not just defining it
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
| W00 | Go | Service + k8s deployment; Prometheus metrics. `net/http` (standard library) keeps this framework-free, and this small a service is the gentlest possible first exposure to Go before W01 leans on it for real |
| W01–W04 | Go | Storage engines and coordination logic. MIT's 6.824/6.5840 distributed systems course builds this exact material (MapReduce, then Raft) in Go, the field's own canonical choice, not an arbitrary one. Goroutines and channels are Go's signature idiom for W04's message-passing simulation; BadgerDB (a real, pure-Go LSM store) gives W01 a genuine same-language reference implementation to read |
| W05–W07 | Java | Prior production Java background keeps ramp cost near zero, and modern Java's sealed interfaces plus record patterns give W05's `StreamItem` the same compiler-enforced sum-type treatment W17 already uses, one idiom reused rather than a new one introduced. The dataflow papers this arc is built around (Naiad, Differential Dataflow) have no maintained JVM reference implementation; their actively maintained implementations live in a different systems language, so this arc leans on the papers directly and points at adjacent production codebases (PyTorch's autograd engine, Ray's CoreWorker, ClickHouse, Spark Structured Streaming) instead of a source-level companion |
| W08 | Go | The one week in this arc where memory layout is the actual subject, not incidental. Go compiles ahead of time, no JIT to warm up before a benchmark means something, and Go structs are real value types in slices, genuinely contiguous memory, the same property a columnar query engine depends on. Java's lack of true value types would work against the week's own point here |
| W09–W10 | Scala | Spark itself is Scala, and its query optimizer (Catalyst) and its "abstract algebra for big data" aggregation story (Algebird) are both genuinely built the way these two weeks have you build toy versions: case classes, pattern matching, typeclasses. Low ramp cost given prior production Spark/Scala experience; kept deliberately gentle, deeper FP mastery is a separate, dedicated Scala-and-Haskell plan, not this curriculum's job |
| W11 | Python | ML ecosystem, numerical computing |
| W12 | Scala + Python (PySpark) | Direct continuation of W09–W10: runs real Spark, in both languages, on the identical job, to measure where the host language actually costs you (UDFs) and where it doesn't (the DataFrame API, same Catalyst plan either way) |
| W13–W16 | Python | ML ecosystem, numerical computing, Ray for distributed actors, Numba for GPU |
| W17 | Java | Chandy-Lamport's `Message` type is exactly the shape a sealed interface and exhaustive pattern-matching `switch` were built for (`DataMessage`/`Marker`, compiler-enforced coverage), a real improvement over a language without sum types, not just a language-consistency choice. `LinkedBlockingQueue` substitutes cleanly for FIFO channels |
| W18 | Go / Java / Python | Depends on capstone option; Option A (KV store) is Go to match what it's extending, W01 and W04; Options B and C are Java to match W05/W07/W17 |
| W19 | Helm/YAML; reads Go | You operate two real operators, KubeRay and Kubeflow's Spark Operator, both implemented in Go, rather than author one yourself: install via Helm, deploy a `RayCluster`/`SparkApplication`, break and debug each, then read (not write) a slice of each one's real reconciler. By this point you've already written Go in five other weeks, so this reading is no longer a cold start the way it would be otherwise |
| W20 | Java + Go | Instrument the W07 DD engine (Java) with the Prometheus Java client and the OpenTelemetry Java SDK; Go log-aggregator built and wired in as a sidecar on the KubeRay cluster from W19, the language cloud-native sidecars are overwhelmingly written in for real. The sidecar's language is independent of the cluster it's attached to |
| W03, W13, W15 | Go (secondary) | Automation tools, coordination services, all using `net/http` (standard library), no framework |

---

## Repository Layout

```
.
├── Home.md               # Daily entry point, open this in Obsidian
├── config.md             # Set start_date here
├── README.md             # This file
├── SETUP.md              # Environment setup (Java, Go, Scala, Python, Docker, Obsidian)
├── RESOURCES.md          # All papers and books, by week, with free links
├── CONTEXT.md            # Session context for AI-assisted study sessions
├── weeks/                # One .md file per week (W00–W20)
├── code/                 # Your implementations, see code/README.md
├── tools/                # Automation tools: plan-dates.go (unrelated to the curriculum's language choices); job_coordinator, grad_server, bench_runner, log-aggregator (Go)
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

**Different languages?** The algorithms are language-agnostic. This curriculum is built around four languages you write, weighted toward depth in what you already know plus exactly one deliberately introduced new component, rather than breadth for its own sake: Go carries most of Arc 1, W08, W18's KV-store option, and every secondary automation tool (net new, but scoped as a gentle introduction, and its footprint is real rather than token, MIT's own 6.824 distributed systems course builds this exact material in Go first); Java carries the rest of Arc 2 and one week of Arc 3 (near-zero ramp cost against a production Java background, and modern Java's records, sealed interfaces, and pattern matching give the ADT-and-exhaustiveness story this curriculum leans on without a new language to learn); Scala is a short, focused, deliberately gentle module where the real production system (Spark) is itself written in Scala, making it worth the investment even given prior Scala familiarity, deeper FP mastery is intentionally left to a separate plan; Python covers the ML-native arc. Substitutions: if you don't have a Java background the way this plan assumes, the Java weeks could run in Go instead (the two are close in scope for these exercises) or stay Java with more ramp time budgeted; the Python weeks could be Julia. The language choices are justified in the Language Map above, but they're not sacred.

---

## Prerequisites

- Comfortable programming in at least one language, in any paradigm. This curriculum is explicitly meant to be a refresh back into Java and Scala for someone with prior exposure to both, plus a gentle, genuinely new introduction to Go, not something that assumes you're starting Java or Scala from zero
- Knows what a hash map and B-tree are
- Has written concurrent code before (threads, async, actors, etc.) in whatever language you already know
- Familiar with basic algorithms (sorting, BFS, binary search)

No PhD required. No ML background required for the early arcs.

**Already know Java? New to Go?** Java (W05–W07, W17) is the lowest-ramp language in the curriculum against a production Java background: near-zero syntax review, closer to formalizing patterns (records, sealed interfaces, pattern matching) you may not have named explicitly than learning anything new. Go (W00–W04, W08, secondary tooling) is this curriculum's one deliberately introduced new component, kept gentle by design: `net/http` and goroutines-plus-channels cover nearly everything it's used for, no framework, no generics-heavy code. Scala (W09–W10, W12) is similarly low-ramp if you already have production Spark/Scala experience, and is intentionally kept shallow here too, deeper FP mastery is a separate plan, not this curriculum's job. Budget real ramp time for Go specifically before W00: [A Tour of Go](https://go.dev/tour/) (~1 hour for the Basics and Methods/Interfaces sections) covers everything the early weeks need. See the language-specific sections of [SETUP.md](SETUP.md) for what to review before each.

---

## Licensing

This curriculum's written content (everything under `weeks/`, `Templates/`, and this documentation) is licensed under [CC BY 4.0](LICENSE): free to use, adapt, and build on, including for your own trainings, courses, or videos, provided you credit Gaston Guitart as the original author. See [CITATION.cff](CITATION.cff) for the exact attribution format.

Code under `code/` and `tools/` is licensed separately under the [MIT License](LICENSE-MIT), permissive with no attribution requirement beyond keeping the copyright notice intact.
