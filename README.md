# Distributed, Data-Intensive Systems: Engineering Curriculum

A self-directed curriculum for software engineers who want pragmatic mastery of distributed and data-intensive systems, with a focus on storage and query internals, partitioning and the shuffle, distributed model training and inference, and running all of it on Kubernetes with real observability. Every unit has a specific paper or chapter to read, a concrete task, and a deliverable.

Roughly half the units build a mechanism from primitives; the other half drive a real production engine (Spark, Flink, DuckDB, Kubeflow) until it fails in an instructive way. The rule is measure or operate a real system where one exists, and reimplement only where the mechanism itself is the lesson.

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

**Budgeted at about 10 hours per unit.** At that pace the core is roughly four months. The Minimum bar in each unit is what a bad week looks like rather than the target, so a week that only clears the bar is still a week that counts.

Running this alongside a full-time job works at roughly half that pace: hit the Minimum bar, treat everything past it as optional, and expect a unit to span two calendar weeks. The units are written so that the next one still works.

| Arc | Units | Focus | Language |
|-----|-------|-------|----------|
| Setup | W00 | Local k8s, Prometheus, Grafana | Go |
| Storage, Batch, and Failure | W01–W03 | What a write costs, the price of materializing between iterations, clocks and failure detection | Go (W01, W03) / PySpark (W02) |
| Data Movement and Execution | W04–W07 | Windows and backpressure, partitioning and the shuffle, vectorized execution, and physical planning: where a distributed engine decides to move data | Java where you author or where Flink requires it (W04, W05 Part 1) / Python for Spark and DuckDB (W05 Part 2, W06, W07) |
| Distributed ML Systems | W08–W13 | ML pipelines and table formats, distributed training, tensor/pipeline parallelism, actor model (Ray), attention and cache-aware routing, fault tolerance | Python (W08–W12) / Java (W13) |
| Infrastructure | W14–W15 | Kubernetes Operators (Kubeflow Trainer, Spark Operator), gang scheduling, observability (Prometheus, OTel, Grafana) | YAML + Python (W14) / Java + Go (W15) |
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
- Package your own PySpark job into an image and submit it as a `SparkApplication`, then debug the dependency-packaging failure that every team hits on their first PySpark-on-Kubernetes deploy
- Explain why gang scheduling exists by deadlocking two training jobs on partial placement, and say what a queueing layer needs to know that the default scheduler does not
- Instrument a distributed system with Prometheus metrics and OpenTelemetry traces
- Build a Grafana dashboard from scratch; explain the four golden signals

---

## Each Unit

Every unit has:
- **Read**: one or two named papers or chapters, with specific sections called out
- **Code**: a concrete implementation task with named files and a clear deliverable
- **Rehearse it in Python first**: optional, 20 minutes, W03 only. Writing the algorithm in Python before writing it in Go tells you whether a failure is the algorithm or the syntax
- **Reflect**: a prediction-versus-measurement table filled in before you run anything, then what surprised you and what you'd do differently
- **Review and articulate**: an adversarial review of your own conclusion, and a timed ninety-second explanation of the finding. These exist because self-study has no examiner, and confident wrongness is its characteristic failure

---

## Language Map

| Units | Language | Why |
|-------|----------|-----|
| W00 | Go | Service + k8s deployment; Prometheus metrics. `net/http` (standard library) keeps this framework-free, and this small a service is the gentlest possible first exposure to Go before W01 and W03 lean on it for real |
| W01, W03 | Go | Storage measurement and coordination logic. MIT's 6.824/6.5840 distributed systems course builds this material in Go, the field's own canonical choice, not an arbitrary one. Goroutines and channels are Go's signature idiom for W03's message-passing simulation; W01 is deliberately the gentlest Go exercise in the plan: file I/O, a loop, and a stopwatch |
| W02 | Python (PySpark) | No framework to build, so no reason to pick a language for expressiveness. The evidence is a stage DAG, not code you wrote, so the driver should be the lightest surface available. PySpark is also the one Spark is actually driven with |
| W04 | Java (Flink) / Python (Spark) | Part 1 is Java because Flink 2.0 ships no other API that exposes the knobs this unit turns. The Spark comparison switches to PySpark, since the point is what the two engines do, not what two APIs look like. Part 2 is a hand-built Java exercise, because comparing block, drop, and spill against one overload is something no framework will do for you: each implements exactly one policy and hides it |
| W05 | Java (Part 1) / Python (Part 2) | Part 1 authors the shuffle in Java, where sealed interfaces give `Partitioner` compiler-enforced exhaustiveness from the same idiom W04 and W13 use. Part 2 drives Spark in PySpark, because finding a straggler in the Spark UI is a diagnosis skill and the driver language is nearly invisible while you do it |
| W06 | Python (DuckDB, NumPy) | Three implementations of one query: a generator pipeline (which *is* the Volcano model, lazy and pull-based, given to you as a language feature), hand-rolled NumPy vectorization, and DuckDB. The middle term is what makes it honest, separating "vectorized beats row-at-a-time" from "C beats Python." DuckDB rather than Spark because Spark's per-query overhead would swamp the effect at this data size |
| W07 | Python (PySpark) and SQL | Spark's own planner rather than a toy one. A planner you wrote agrees with you by construction and cannot surprise you; Catalyst has a cost model you did not write and will make decisions you did not expect. Almost the whole unit is `spark.sql(...)` and reading what comes back, which is precisely why the driver should be the lightest surface going |
| W08–W12 | Python | ML ecosystem and numerical computing, plus Ray for distributed actors. W10 deliberately adds no new dependency; it imports W09's own `ring_allreduce` as its communication layer |
| W13 | Java | Chandy-Lamport's `Message` type is exactly the shape a sealed interface and exhaustive pattern-matching `switch` were built for (`DataMessage`/`Marker`, compiler-enforced coverage), a real improvement over a language without sum types, not just a language-consistency choice. `LinkedBlockingQueue` substitutes cleanly for FIFO channels |
| W14 | YAML + a little Python; reads Go | You operate two real operators, Kubeflow Trainer and Kubeflow's Spark Operator, both implemented in Go, rather than author one yourself: install them, deploy a `TrainJob`/`SparkApplication`, break and debug each, then read (not write) a slice of each one's real reconciler. The optional Spark half has you package and submit your own PySpark script, where the real lesson is that a Python job ships an interpreter environment rather than one self-contained JAR, which is the most common operational pain in running PySpark on Kubernetes. Trainer is the vendor-neutral choice deliberately: its `TrainJob` API unified the older framework-specific CRDs and runs the same way on any cluster. By this point you've written Go in three other units, enough that this reading is not a cold start |
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
├── MEASUREMENTS.md       # Running log of numbers you measured yourself, with predictions
├── weeks/                # One .md file per unit (W00–W16, where W16 is the optional capstone project)
├── code/                 # Your implementations, see code/README.md
├── tools/                # Automation: plan-dates.go, grad_server, bench_runner, log-aggregator (Go)
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

**Under 3 hours some weeks?** That will happen, and the plan expects it. Do the Read and the Reflect, hit the Minimum bar if you can, skip the rest without guilt. If you have to skip units entirely, the load-bearing ones are W03, W05, W09, and W12: the shuffle and the allreduce are the two data-movement patterns nearly everything else is built from, and W03's failure detection is the idea five later units keep returning to.

**Skip the infrastructure arc?** W00, W14, and W15 are independent. You can complete W01 through W13 without touching Kubernetes and come back to them when it's relevant to your work.

**Add your own unit?** Copy `Templates/week-template.md`, set `week_number` in frontmatter, and it appears in the Home.md dashboard automatically. Keep the `week` naming in filenames and frontmatter: Home.md's queries depend on it.

**Tracking progress separately from curriculum edits?** Keep `main` for curriculum changes and a separate `progress` branch for checked-off tasks and Reflect answers, merging `main` into `progress` but never the reverse. See [Branch Workflow](#branch-workflow) for the merge procedure and how to handle conflicts.

**Why these languages?** One rule decides all of them: **each system is driven in the language that system is actually written and used in**, and the hand-built units pick whichever language makes the mechanism clearest.

That gives **Python** the largest share: it is what Spark is driven with (W02, W05 Part 2, W07, W14), what DuckDB is driven with (W06, W08), and what the entire ML arc runs on (W08 to W12, W16). **Java** covers two situations: W04 Part 1, where Flink ships no alternative, and the units where you author a mechanism and sealed interfaces plus record patterns do real work (W04 Part 2, W05 Part 1, W13, W15). **Go** carries W00, W01, W03, and the secondary tooling, with a footprint large enough to make W14's operator reading legible rather than token.

Substitutions: if you don't have a Java background the way this plan assumes, the hand-built Java units could run in Go instead (the two are close in scope for these exercises), though the engine-driving units would then need PySpark and PyFlink; the Python units could be Julia. The language choices are justified in the Language Map above, but they're not sacred.

---

## Branch Workflow

`main` holds the curriculum. `progress` holds your ticks, your Reflect answers, and your `Current State` updates. `main` merges into `progress`, never the reverse.

### Before any merge

```bash
git add -A && git commit -m "progress: W00"
git push origin progress
git fetch origin
git merge origin/main
```

Committing first is what makes everything below safe, because git never discards committed work. Pushing means the branch survives the machine as well.

### When it stops

Conflicts happen on lines both branches changed. In practice that means step lines you have ticked, because a curriculum revision rewords the same line your `x` sits on. Answers written on the blank line below a question rarely conflict.

```bash
git diff --name-only --diff-filter=U     # which files
```

A conflict looks like this, where `HEAD` is your `progress` branch and the lower half is the incoming curriculum:

```
<<<<<<< HEAD
- [x] `main.go`: `http.HandleFunc` for two routes, plus two metric objects
=======
- [ ] In `main.go`, declare two package-level variables, created exactly once:
>>>>>>> origin/main
```

**Path A, resolve in place.** Edit the file: keep the incoming wording, put your `x` back, keep your answer lines. Then:

```bash
git add weeks/W00-setup.md
git commit
```

During a merge, `--ours` means `progress` and `--theirs` means `main`:

```bash
git checkout --theirs weeks/W00-setup.md   # take main's whole file, drop your ticks in it
git checkout --ours   weeks/W00-setup.md   # keep yours, drop the curriculum update
git checkout -m       weeks/W00-setup.md   # restore the conflict markers if you mangled the file
git merge --abort                          # back out entirely, nothing changed
```

**Path B, take the new file and re-apply your marks.** For a heavily rewritten unit, resolving hunk by hunk is worse than starting from the new text. Take `main`'s version wholesale, then diff it against your pre-merge copy:

```bash
git checkout --theirs weeks/W00-setup.md
git add weeks/W00-setup.md
git commit

git show ORIG_HEAD:weeks/W00-setup.md > ~/W00-setup.mine.md
diff ~/W00-setup.mine.md weeks/W00-setup.md
```

`ORIG_HEAD` is your `progress` commit from immediately before the merge, so there is no need to copy the file beforehand. Re-tick and re-paste your answers from the diff, commit again, then delete the scratch copy.

Keep that copy outside the repo. A `weeks/W00-setup.old.md` inside the vault carries `week_number` frontmatter and would appear as a second W00 row in the Home.md dashboard.

Path B discards your ticks and answers in that one file and you re-apply them from the diff. That is the cost of taking a cleanly rewritten unit.

### If the merge goes wrong after it has committed

```bash
git reset --hard ORIG_HEAD
```

`git reflog` finds anything else.

---

## Prerequisites

- Comfortable programming in at least one language, in any paradigm. This curriculum is explicitly meant to be a refresh back into Java for someone with prior exposure to it, plus a gentle, genuinely new introduction to Go, not something that assumes you're starting Java from zero
- Knows what a hash map and B-tree are
- Has written concurrent code before (threads, async, actors, etc.) in whatever language you already know
- Familiar with basic algorithms (sorting, BFS, binary search)

No PhD required. No ML background required for the early arcs.

**Already know Java? New to Go?** Against a production Java background, the Java units (W04, W05 Part 1, W13, W15) are near-zero syntax review: mostly naming patterns you already use (records, sealed interfaces, pattern matching). Go (W00–W03 and secondary tooling) is kept deliberately narrow: `net/http` and goroutines-plus-channels cover nearly everything it's used for, no framework, no generics-heavy code. Budget ramp time for Go before W00: [A Tour of Go](https://go.dev/tour/) (~1 hour for the Basics and Methods/Interfaces sections) covers everything the early units need. See the language-specific sections of [SETUP.md](SETUP.md) for what to review before each.

---

## Licensing

This curriculum's written content (everything under `weeks/`, `Templates/`, and this documentation) is licensed under [CC BY 4.0](LICENSE): free to use, adapt, and build on, including for your own trainings, courses, or videos, provided you credit Gaston Guitart as the original author. See [CITATION.cff](CITATION.cff) for the exact attribution format.

Code under `code/` and `tools/` is licensed separately under the [MIT License](LICENSE-MIT), permissive with no attribution requirement beyond keeping the copyright notice intact.
