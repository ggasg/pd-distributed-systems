# Distributed, Data-Intensive Systems: Engineering Curriculum

A self-directed curriculum for software engineers who want pragmatic mastery of distributed and data-intensive systems, with a focus on storage and query internals, partitioning and incremental computation, distributed model training, GPU-bound compute, and running all of it on Kubernetes with real observability. Every unit has a specific paper to read, a concrete coding task, and a deliverable.

This isn't for people who want to pass system design interviews. It's for engineers who want to build real distributed systems and understand them from the inside out.

---

## Quick Start

1. Clone the repo and open it as an Obsidian vault
2. Follow [SETUP.md](SETUP.md) to install Java, Go, Python, Docker, and Obsidian plugins
3. Skim [RESOURCES.md](RESOURCES.md): most readings are free links, but a few (DDIA, the most-used book) are worth buying before you hit the unit that needs them
4. Set `start_date` in [config.md](config.md); dates recalculate automatically, and they are a running order rather than deadlines
5. Open [Home.md](Home.md) as your daily entry point
6. Start at W00 (infrastructure setup), then W01

---

## Structure

15 units across 4 arcs, plus a W00 setup unit (16 in total, W00 through W15), and an optional W16 grand capstone project.

**Budgeted at 5 hours per unit**, which is one hour a weekday or two evening sessions. That is deliberately modest: this is designed to run alongside a full-time job and whatever else you are studying, not to be your main commitment. At a steady 5 hours a week the core is about 4 months; at 3 hours a week, closer to 7. Neither is falling behind.

| Arc | Units | Focus | Language |
|-----|-------|-------|----------|
| Setup | W00 | Local k8s, Prometheus, Grafana | Go |
| Storage, Batch, and Failure | W01–W03 | What a write costs, the price of materializing between iterations, clocks and failure detection | Go (W01, W03) / Java + Spark (W02) |
| Data Movement and Execution | W04–W07 | Windows and backpressure, partitioning and the shuffle, vectorized execution, and physical planning: where a distributed engine decides to move data | Java throughout, driving Flink (W04), Spark (W05, W07), and DuckDB (W06) |
| Distributed ML Systems | W08–W13 | ML pipelines and table formats, distributed training, tensor/pipeline parallelism, actor model (Ray), attention and cache-aware routing, fault tolerance | Python (W08–W12) / Java (W13) |
| Infrastructure | W14–W15 | Kubernetes Operators (Kubeflow Trainer, Spark Operator), gang scheduling, observability (Prometheus, OTel, Grafana) | YAML + Java (W14) / Java + Go (W15) |
| Capstone (optional) | W16 | Distributed training + serving platform, fully observed (synthesizes W08, W09, W12, W13, W15; deploys as a TrainJob via W14) | Python |

---

## What You'll Be Able to Do After Each Arc

**After Arc 1, Storage, Batch, and Failure (W01–W03):**
- Measure, rather than assert, how much faster appending to a log is than updating in place, then measure what that speed costs you on reads and explain why compaction has to exist
- Predict the disk I/O cost of an iterative job from graph size alone, then check yourself against a real Spark stage DAG and find out which term you got wrong
- Read a stage DAG well enough to say, in thirty seconds, whether a slow job is re-reading its input every iteration, and defend `cache` against `checkpoint` at 100 iterations
- Implement vector clocks; reason about causal consistency and concurrent events
- Build a heartbeat failure detector, then make it declare a healthy node dead on purpose, and defend a timeout knowing it can only trade false suspicions against slow detection
- State precisely what at-most-once, at-least-once, and effectively-once mean, and why the third is never a delivery guarantee

**After Arc 2, Data Movement and Execution (W04–W07):**
- Drive Flink until it silently drops a valid event, find the metric that counted it, and choose between allowed lateness and a side output knowing exactly who pays for each
- Say what Spark Structured Streaming can and cannot reproduce of Flink's late-data behaviour, from having run both against the same input
- Use Little's Law to predict queue growth before running anything, then implement block, drop, and spill against the same overload and measure what each one costs, including how wrong it makes the answer
- Build a working shuffle (partitioned map-side spill, reduce-side fetch) and explain why M times R files exist
- Diagnose skew the way you would at work: find the straggler in the Spark UI's task duration distribution, distinguish "one task got more data" from "one task was on a slow machine," and fix it
- Measure a real vectorized engine against a row-at-a-time baseline, single-threaded so the number means something, and report vectorization and parallelism as two separate figures rather than one
- Read `EXPLAIN FORMATTED` and name every join strategy and every `Exchange` in a plan
- Cause the silent regression on the real thing: grow one table past `autoBroadcastJoinThreshold`, watch a broadcast quietly become a shuffle with no error and no log line, and name the metric that would have caught it
- Say precisely what Adaptive Query Execution fixes and what it cannot, rather than treating it as a switch that makes planning problems go away

**After Arc 3, Distributed ML Systems (W08–W13):**
- Design and implement a versioned ML feature store with Parquet + DuckDB
- Read a real Delta transaction log by hand, create and then fix a small-file problem, and defend a vacuum retention window against both a retraining job and an auditor
- Implement ring-allreduce over raw TCP sockets, measure bytes on the wire against a naive baseline, and explain why an allreduce is a reduce-scatter plus an all-gather
- Split a model rather than the data: tensor-parallel a single matmul, pipeline-parallel a stack of layers, shard optimizer state ZeRO-style, and measure the pipeline bubble against its theoretical value
- Build a stateful actor system with Ray; explain why actors (not stateless tasks) are the right abstraction for coordinating training workers
- Build a KV cache against a given multi-head attention implementation in NumPy; explain and fix a cross-request cache-bleed bug
- Put a router in front of two replicas and measure how much prefill work round-robin balancing throws away, then trade cache locality against load balance and defend where you set the line
- Implement Chandy-Lamport distributed snapshots; explain what "consistent cut" means

**After Arc 4, Infrastructure (W14–W15):**
- Deploy, break, and debug two real Kubernetes operators (Kubeflow Trainer, Kubeflow's Spark Operator); explain a reconcile loop by reading one, not just defining it
- Package your own Spark job into an image and submit it as a `SparkApplication`, then debug the class-not-found failure that every team hits on their first Spark-on-Kubernetes deploy
- Explain why gang scheduling exists by deadlocking two training jobs on partial placement, and say what a queueing layer needs to know that the default scheduler does not
- Instrument a distributed system with Prometheus metrics and OpenTelemetry traces
- Build a Grafana dashboard from scratch; explain the four golden signals

---

## Each Unit

Every unit has:
- **Read**: one or two named papers or chapters, with specific sections called out
- **Code**: a concrete implementation task with named files and a clear deliverable
- **Rehearse it in Python first**: optional, 20 minutes, and now only in W03. Writing the algorithm in Python before writing it in Go tells you whether a failure is the algorithm or the syntax. W02 and W06 used to carry one too, and lost it along with their Go builds when both units moved to measuring a real engine instead
- **Reflect**: what you built, what surprised you, what you'd do differently

---

## Language Map

| Units | Language | Why |
|-------|----------|-----|
| W00 | Go | Service + k8s deployment; Prometheus metrics. `net/http` (standard library) keeps this framework-free, and this small a service is the gentlest possible first exposure to Go before W01 and W03 lean on it for real |
| W01, W03 | Go | Storage measurement and coordination logic. MIT's 6.824/6.5840 distributed systems course builds this material in Go, the field's own canonical choice, not an arbitrary one. Goroutines and channels are Go's signature idiom for W03's message-passing simulation; W01 is deliberately the gentlest Go exercise in the plan: file I/O, a loop, and a stopwatch |
| W02 | Java (Spark) | No framework to build, so no reason to pick a language for expressiveness. Spark's Java API drives the engine, the arithmetic happens on paper, and the evidence is in a stage DAG rather than in code you wrote |
| W04 | Java (Flink, then Spark) | Part 1 configures Flink's watermark machinery, which exposes out-of-orderness bounds, allowed lateness, and late-data side outputs as three separate visible knobs, then asks the same question of Spark Structured Streaming and compares the answers. Part 2 stays a hand-built Java exercise, because comparing block, drop, and spill against one overload is something no framework will do for you: each implements exactly one policy and hides it |
| W05 | Java | The one build in Arc 2 that survives whole. Part 1 writes the shuffle, where sealed interfaces give `Partitioner` compiler-enforced exhaustiveness from the same idiom W04 and W13 use. Part 2 reproduces the skew in real Spark, because finding a straggler in the Spark UI is the diagnosis skill and no build teaches it |
| W06 | Java (DuckDB over JDBC) | You write the row-at-a-time baseline and DuckDB provides the vectorized side, so only one half of the comparison is yours. DuckDB rather than Spark here specifically: Spark's per-query overhead would swamp the effect at this data size, and DuckDB is a single-node vectorized engine, which is exactly and only what this unit is about |
| W07 | Java (Spark) | Spark's own planner rather than a toy one. A planner you wrote agrees with you by construction and cannot surprise you; Catalyst has a cost model you did not write and will make decisions you did not expect. Reading `EXPLAIN` is also the version of this skill you use at work, roughly weekly, forever |
| W08–W12 | Python | ML ecosystem and numerical computing, plus Ray for distributed actors. W10 deliberately adds no new dependency; it imports W09's own `ring_allreduce` as its communication layer |
| W13 | Java | Chandy-Lamport's `Message` type is exactly the shape a sealed interface and exhaustive pattern-matching `switch` were built for (`DataMessage`/`Marker`, compiler-enforced coverage), a real improvement over a language without sum types, not just a language-consistency choice. `LinkedBlockingQueue` substitutes cleanly for FIFO channels |
| W14 | YAML + a little Java; reads Go | You operate two real operators, Kubeflow Trainer and Kubeflow's Spark Operator, both implemented in Go, rather than author one yourself: install them, deploy a `TrainJob`/`SparkApplication`, break and debug each, then read (not write) a slice of each one's real reconciler. The optional Spark half has you package and submit your own Java JAR, reusing the Maven setup from W02, W05, and W07. Trainer is the vendor-neutral choice deliberately: its `TrainJob` API unified the older framework-specific CRDs and runs the same way on any cluster. By this point you've written Go in three other units, enough that this reading is not a cold start |
| W15 | Java + Go | Instrument the W05 shuffle (Java) with the Prometheus Java client and the OpenTelemetry Java SDK; Go log-aggregator built and wired in as a sidecar on the W14 `TrainJob`'s node Pods, the language cloud-native sidecars are overwhelmingly written in for real. The sidecar's language is independent of the workload it's attached to |
| W09 | Go (secondary) | The gradient server is a small automation tool using `net/http` (standard library), no framework |

---

## Repository Layout

```
.
├── Home.md               # Daily entry point, open this in Obsidian
├── config.md             # Set start_date here
├── README.md             # This file
├── SETUP.md              # Environment setup (Java, Go, Python, Docker, Obsidian)
├── RESOURCES.md          # All papers and books, by unit, with free links
├── CONTEXT.md            # Session context for AI-assisted study sessions
├── weeks/                # One .md file per unit (W00–W16, where W16 is the optional capstone project)
├── code/                 # Your implementations, see code/README.md
├── tools/                # Automation tools: plan-dates.go (unrelated to the curriculum's language choices); grad_server, bench_runner, log-aggregator (Go)
└── Templates/            # week-template.md for adding custom units
```

---

## Units

- [W00: Infrastructure Setup](weeks/W00-setup.md)
- [W01: Storage Engines and the Cost of a Write](weeks/W01-storage-engines.md)
- [W02: MapReduce and Its Limits](weeks/W02-mapreduce.md)
- [W03: Clocks, Causality, Time, and Unreliable Networks](weeks/W03-clocks.md)
- [W04: Stream Processing Primitives](weeks/W04-streaming.md)
- [W05: Partitioning and the Shuffle](weeks/W05-shuffle.md)
- [W06: Query Execution](weeks/W06-query-execution.md)
- [W07: Query Planning: Choosing Where Data Moves](weeks/W07-query-planning.md)
- [W08: ML Data Pipelines and Table Formats](weeks/W08-ml-pipelines.md)
- [W09: Distributed Training](weeks/W09-distributed-training.md)
- [W10: Beyond Data Parallelism](weeks/W10-parallelism-strategies.md)
- [W11: The Actor Model and Ray](weeks/W11-actor-model-ray.md)
- [W12: Attention, KV Cache, and Cache-Aware Routing](weeks/W12-attention.md)
- [W13: Fault Tolerance and Snapshots](weeks/W13-fault-tolerance.md)
- [W14: Operating Kubernetes Operators](weeks/W14-kubernetes-operators.md)
- [W15: Observability: Metrics, Tracing, Logging](weeks/W15-observability.md)
- [W16: Grand Capstone: Distributed Training & Serving Platform (optional)](weeks/W16-capstone-platform.md)

**Every unit has a Minimum bar.** It names the smallest thing that counts as having done it, and everything past it is explicitly optional. When a unit runs long, drop from the bottom and hit the bar rather than half-finishing the whole thing; the bar is chosen so the next unit still works.

**Under 3 hours some weeks?** (Calendar weeks, this time.) That will happen, and the plan expects it. Do the Read and the Reflect, hit the Minimum bar if you can, skip the rest without guilt. If you have to skip units entirely, the load-bearing ones are W03, W05, W09, and W12: the shuffle and the allreduce are the two data-movement patterns nearly everything else is built from, and W03's failure detection is the idea five later units keep returning to.

**Skip the infrastructure arc?** W00, W14, and W15 are independent. You can complete W01 through W13 without touching Kubernetes and come back to them when it's relevant to your work.

**Add your own unit?** Copy `Templates/week-template.md`, set `week_number` in frontmatter, and it appears in the Home.md dashboard automatically. The `week` naming in filenames and frontmatter is legacy and kept only because Home.md's queries depend on it.

**Tracking progress separately from curriculum edits?** Keep `main` for curriculum changes and a separate `progress` branch for checked-off tasks and Reflect answers. See the Branch Workflow section in [CONTEXT.md](CONTEXT.md) for how to merge updates between them.

**Different languages?** The algorithms are language-agnostic. This curriculum is built around three languages you write, weighted toward depth in what you already know plus exactly one deliberately introduced new component, rather than breadth for its own sake. **Go** carries W00, W01, W03, and the secondary automation tools: net new, scoped as a gentle introduction, and with a footprint real enough to make W14's operator reading legible rather than token. **Java** carries all of Arc 2 plus W13 and W15: near-zero ramp cost against a production Java background, and it is the single driver language for every engine this curriculum operates (Spark, Flink, DuckDB over JDBC), which is what keeps four consecutive units on one stack. **Python** covers the ML-native arc.

Scala was cut deliberately. It previously bought one unit (W07's toy planner) and one packaging exercise, in exchange for a fourth language and a build tool used nowhere else. W07 now reads Spark's real planner instead of imitating it, which removed the only argument for keeping Scala. Deeper FP mastery is a separate, dedicated plan, not this curriculum's job.

Substitutions: if you don't have a Java background the way this plan assumes, the hand-built Java units could run in Go instead (the two are close in scope for these exercises), though the engine-driving units would then need PySpark and PyFlink; the Python units could be Julia. The language choices are justified in the Language Map above, but they're not sacred.

---

## Prerequisites

- Comfortable programming in at least one language, in any paradigm. This curriculum is explicitly meant to be a refresh back into Java for someone with prior exposure to it, plus a gentle, genuinely new introduction to Go, not something that assumes you're starting Java from zero
- Knows what a hash map and B-tree are
- Has written concurrent code before (threads, async, actors, etc.) in whatever language you already know
- Familiar with basic algorithms (sorting, BFS, binary search)

No PhD required. No ML background required for the early arcs.

**Already know Java? New to Go?** Java (W02, W04–W07, W13, W15) is the lowest-ramp language in the curriculum against a production Java background: near-zero syntax review, closer to formalizing patterns (records, sealed interfaces, pattern matching) you may not have named explicitly than learning anything new. In the engine-driving units it barely shows at all, since most of the work is a SQL string and a plan you read. Go (W00–W03 and secondary tooling) is this curriculum's one deliberately introduced new component, kept gentle by design: `net/http` and goroutines-plus-channels cover nearly everything it's used for, no framework, no generics-heavy code. Budget real ramp time for Go specifically before W00: [A Tour of Go](https://go.dev/tour/) (~1 hour for the Basics and Methods/Interfaces sections) covers everything the early units need. See the language-specific sections of [SETUP.md](SETUP.md) for what to review before each.

---

## Licensing

This curriculum's written content (everything under `weeks/`, `Templates/`, and this documentation) is licensed under [CC BY 4.0](LICENSE): free to use, adapt, and build on, including for your own trainings, courses, or videos, provided you credit Gaston Guitart as the original author. See [CITATION.cff](CITATION.cff) for the exact attribution format.

Code under `code/` and `tools/` is licensed separately under the [MIT License](LICENSE-MIT), permissive with no attribution requirement beyond keeping the copyright notice intact.
