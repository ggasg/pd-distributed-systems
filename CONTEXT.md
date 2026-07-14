# Claude Session Context

> Paste this at the start of any new Cowork session to restore context.
> This file is the single source of truth. No Notion.

---

## Goal

Become a pragmatic master of Distributed, Data-Intensive Systems. Focus areas:
- ML Model Training infrastructure
- Compute-Intensive Distributed Execution
- AI Tasks / LLM inference systems

2h/day, 5 days/week.

---

## Repo

Two folders, same GitHub repo, checked out as separate `git worktree`s so neither tool has to switch branches:
- `~/dev/pd-distributed-systems` (branch `main`): the Cowork/CLI folder. Curriculum authoring happens here.
- `~/dev/pd-distributed-systems-progress` (branch `progress`): the Obsidian vault. Running the plan happens here.

See Branch Workflow below for how these two stay in sync.

Vault structure:
- `config.md`: set `start_date` here; all week dates recalculate in Home.md
- `Home.md`: daily entry point; DataviewJS auto-detects current week and shows schedule
- `weeks/W00-*.md` through `W19-*.md`: one file per week; each has Read, Code, optional Python DSA Review, and Reflect sections
- `weeks/W20-*.md`: optional grand capstone; synthesizes W11, W12, W15, W16, W18, W19 into one distributed training and serving platform
- `tools/plan-dates.go`: `go run tools/plan-dates.go --start 2026-07-13` prints full schedule
- `SETUP.md`: full environment setup (Go, C++, Scala, Python, Docker, Obsidian)
- `RESOURCES.md`: all papers and books by week with free links
- `Templates/week-template.md`: blank week file for adding custom weeks
- `posts/TEMPLATE.md`: structured blog post format for weekly write-ups
- `code/README.md`: expected directory tree and build commands for all weeks
- `README.md`: public-facing GitHub description

No Notion. No separate task tracker. Everything lives here.

---

## Curriculum

20 core weeks (W00 pre-week + W01–W19) across 4 arcs, plus an optional W20 grand capstone.

### W00: Infrastructure Setup (pre-week)
| Week | Topic | Deliverable |
|------|-------|-------------|
| W00 | Local k8s (kind) + Prometheus + Grafana | `code/hello-metrics/` Go service with Prometheus metrics, deployed to kind |

W00 also carries the curriculum's first reading: **DDIA Chapter 1**, read before anything technical starts — vocabulary the rest of the curriculum assumes rather than material this week's build needs directly.

### Arc 1: Data Systems Internals (W01–W04), Go

| Week | Topic | Key Paper | Deliverable |
|------|-------|-----------|-------------|
| W01 | LSM-Trees + Storage Engines | DDIA Ch.3; LevelDB source | `lsm/` Go: MemTable, SSTable, LSMTree |
| W02 | Encoding + Wire Formats | DDIA Ch.4; Protobuf encoding spec | `encoding/` Go: varint, row vs column store benchmark |
| W03 | MapReduce and Its Limits | DDIA Ch.10; Dean & Ghemawat (2004); Zaharia et al. (2012) | `mapreduce/` Go: MR framework (goroutines), word count, iterative PageRank; HTTP coordinator |
| W04 | Clocks, Causality, Time | Lamport (1978); DDIA Ch.8 | `clocks/` Go: vector clocks + causal delivery over channels |

### Arc 2: Streaming, Dataflow, and Query Planning (W05–W10), C++ (W05–W08) / Scala (W09–W10)

| Week | Topic | Key Paper | Deliverable |
|------|-------|-----------|-------------|
| W05 | Stream Processing Primitives | Dataflow Model (Akidau et al., 2015); DDIA Ch.11 | `streaming/` C++: windowed aggregation, immutability enforced by design discipline (encapsulation + `const`), not the compiler |
| W06 | Naiad + Timely Dataflow | Naiad paper (Murray et al., SOSP 2013) | `timely-toy/` C++: Timestamp, Pointstamp, ProgressTracker — reference reading swapped from the Rust `timely-dataflow` crate to PyTorch's autograd engine (`torch/csrc/autograd/engine.cpp`) and Ray's `CoreWorker` (`core_worker.cc`/`task_manager.cc`), both real dependency-counted schedulers |
| W07 | Differential Dataflow + Incremental View Maintenance | DD paper (McSherry, 2013), Sections 1–2 | `dd-scratch/` C++: trimmed DD core (Update, Collection, WordCount, 1–2 days) + incremental materialized view vs. full-recompute benchmark, tested hands-on against a locally-installed ClickHouse server and local-mode Spark Structured Streaming — both real, both OSS, both run on your own machine, no vendor docs taken on faith |
| W08 | Query Execution | Volcano (1994); MonetDB/X100 (2005) | `query-exec/` C++: RowExecutor, ColumnFilter, HashJoin, Benchmark (3–8x speedup) — benchmarked conceptually against DuckDB's vectorized execution engine (C++, already in the stack via W11) |
| W09 | Rule-Based Query Planning | Spark SQL/Catalyst paper (Armbrust et al., SIGMOD 2015) | `query-planner/` Scala: toy Catalyst-style optimizer — `LogicalPlan`/`Expr` ADTs, pattern-matching rewrite rules, `transformDown` combinator; reads real Catalyst source in the same language it's written in |
| W10 | Aggregation Algebra: Monoids and Semigroups | Algebird source; "Of Algebirds, Monoids, Monads..." (Noll) | `agg-algebra/` Scala: `Semigroup`/`Monoid` typeclass from scratch, associative-vs-non-associative average, reconnects to W07's `consolidate()` |

**Note on the C++ swap (2026-07-13):** Arc 2 was originally Rust, chosen because `timely-dataflow`/`differential-dataflow` are themselves actively maintained Rust crates. Gaston found Rust's ramp (the borrow checker specifically) too costly relative to the payoff and asked for the next-best alternative. C++ won over staying with Rust or reverting to Scala because it points directly at the curriculum's actual target — distributed model training and compute-intensive AI workflows are implemented in C++ at the systems level (PyTorch's `c10d`/ATen, NCCL, gRPC) — and because Gaston already has prior C++ exposure (his first language), lowering the activation energy relative to a brand-new ownership model. The honest cost: the "read the real reference implementation" rationale that justified Rust for this arc doesn't transfer — no C++ project continues the Naiad/timely-dataflow/differential-dataflow lineage. W06–W08 substitute adjacent, still-real C++ reference material (PyTorch's autograd engine, Ray's CoreWorker, DuckDB's execution engine) instead. See [[pd-curriculum-language-stack]] for the full history.

**Note on the Scala addition, W09–W10 (2026-07-13):** added as a direct answer to "how do FP operators relate to MPP systems" — not a hypothetical, but a real, checked connection: Spark itself is Scala, and Catalyst (its optimizer) is genuinely built from case classes and pattern-matching rewrite rules; `reduceByKey`-style operators are only safe to parallelize because the combining function is required to be a semigroup, the exact idea Twitter's Algebird formalizes. Placed at the end of Arc 2 (not scattered into existing weeks) because it's the natural capstone to the arc's query-processing throughline — W05–W08 build the individual operators, W09–W10 build what arranges and combines them the way a real MPP engine does. Low ramp cost given Gaston's existing production Spark/Scala background; this closes the "re-introduce Scala for FP-heavy material" item that was open since the Rust→C++ swap. Inserting these two weeks pushed W09–W18 to W11–W20; every cross-reference in weeks/, RESOURCES.md, code/README.md, and this file was updated accordingly. Total core curriculum grew from 18 to 20 weeks (~4.6 months at 2h/day, 5 days/week), which Gaston confirmed he's happy with (target: up to 5 months). See [[pd-curriculum-language-stack]] for the full history.

### Arc 3: Distributed ML & Compute (W11–W17), Python / Go secondary

| Week | Topic | Key Paper | Deliverable |
|------|-------|-----------|-------------|
| W11 | ML Data Pipelines | Hidden Tech Debt (2015); Delta Lake (2020) | `feature-pipeline/` Python: versioned features, Parquet + DuckDB |
| W12 | Distributed Training | Horovod (2018); PyTorch DDP source | `distributed-training/` Python: ring-allreduce via sockets, 2-worker MLP; Go gradient server |
| W13 | The Actor Model and Ray | Hewitt, Bishop, Steiger (1973); Ray (Moritz et al., OSDI 2018) | `actor-training/` Python: Ray actors (TrainerWorker + ParameterServer), PyTorch CNN on MNIST, benchmarked against W12 |
| W14 | GPU Memory + Compute | CUDA Guide Ch.1-3; Roofline (2009) | `gpu-gemm/` Python/Numba: naive vs tiled CUDA matmul + roofline; Go bench runner |
| W15 | Attention + KV Cache | Attention Is All You Need (2017); FlashAttention (2022); PagedAttention (2023) | `attention/` Python: MHA forward pass + KV cache, NumPy only |
| W16 | Fault Tolerance | Chandy-Lamport (1985); Flink ABS (2015); DDIA Ch.9 (optional) | `snapshot/` Go: Chandy-Lamport 3-node simulation over native channels |
| W17 | Capstone | DDIA Ch.5 (Option A only) | Go: replicated KV store / C++: streaming pipeline / C++: incremental query engine / Python: GPU-accelerated ring-allreduce (W12+W14, the option that stays inside this arc) |

### Arc 4: Infrastructure (W18–W19), Go / C++

| Week | Topic | Deliverable |
|------|-------|-------------|
| W18 | Kubernetes Operators | `code/operator/` Go: custom DistributedJob CRD + reconciler |
| W19 | Observability: Metrics, Tracing, Logging | Instrument W07 DD engine with Prometheus + OTel; Grafana dashboard |

### Optional: W20, Grand Capstone (stretch week)

| Week | Topic | Deliverable |
|------|-------|-------------|
| W20 | Distributed training + serving platform | `code/capstone-platform/`: training across worker Pods (W12+W15), Chandy-Lamport-style checkpoint/restore (W16), operator-managed recovery (W18), fully observed in Grafana (W19) |

Optional companion read: **DDIA Chapter 12**, The Future of Data Systems — Kleppmann's own synthesis chapter, read in the week that's doing the same thing to the curriculum's own pieces.

---

## Date Configuration

Edit `start_date` in `config.md` to reschedule. Current config (`2026-07-13`):
- W00: Jul 6 to Jul 12
- W01: Jul 13 to Jul 19
- W05: Aug 10 to Aug 16
- W09: Sep 7 to Sep 13
- W11: Sep 21 to Sep 27
- W17: Nov 2 to Nov 8
- W18: Nov 9 to Nov 15
- W19: Nov 16 to Nov 22
- W20 (optional): Nov 23 to Nov 29

`go run tools/plan-dates.go --start 2026-07-13` prints the full table.

---

## Current State

> Update this section at the end of each week.

- [ ] W00, not started
- [ ] W01
- [ ] W02
- [ ] W03
- [ ] W04
- [ ] W05
- [ ] W06
- [ ] W07
- [ ] W08
- [ ] W09
- [ ] W10
- [ ] W11
- [ ] W12
- [ ] W13
- [ ] W14
- [ ] W15
- [ ] W16
- [ ] W17
- [ ] W18
- [ ] W19
- [ ] W20 (optional)

---

## Branch Workflow

Two long-lived branches, each permanently checked out in its own folder via `git worktree` (not `git checkout`), so Obsidian and Cowork/CLI never touch each other's branch:
- `main`, in `~/dev/pd-distributed-systems`: curriculum authoring. Structural edits to weeks, resources, and setup docs (what happens in a Cowork session when refining the plan).
- `progress`, in `~/dev/pd-distributed-systems-progress`: actually running the plan. Checked-off tasks, filled-in Reflect answers, `Current State` updates. Never merged back into `main`, it only receives.

**One-time setup** (run once, from the `main` folder):

```bash
cd ~/dev/pd-distributed-systems
git worktree add ../pd-distributed-systems-progress progress
```

Then point Obsidian at `~/dev/pd-distributed-systems-progress` (File → Open folder as vault) and leave it there. Cowork/CLI keeps using `~/dev/pd-distributed-systems`. Neither tool needs `git checkout` again.

**To pull curriculum updates from `main` into `progress`** without losing tracked progress, run from the `progress` folder:

```bash
cd ~/dev/pd-distributed-systems-progress
git pull origin progress
git fetch origin main
git merge origin/main
git push origin progress
```

This merges cleanly as long as answers always go on the blank line *below* each Reflect question, never on the question line itself. That keeps curriculum edits (the question line) and progress edits (the answer line) on different lines, which is what avoids merge conflicts in Markdown.

---

## Cowork Setup on a New Machine

1. Clone repo: `git clone <repo-url> ~/dev/pd-distributed-systems` (checks out `main`)
2. Set up the `progress` worktree: `cd ~/dev/pd-distributed-systems && git worktree add ../pd-distributed-systems-progress progress`. See Branch Workflow above.
3. Open Cowork → Select folder → `~/dev/pd-distributed-systems` (the `main` folder; Cowork/CLI never touches `progress` directly)
4. Open Obsidian → vault → `~/dev/pd-distributed-systems-progress`
5. Create project "Professional Development" with instructions:
   > *"Be direct and concise with practical, action-oriented suggestions. My objective is to become a pragmatic master of Distributed, Data-Intensive Systems, not a theory know-it-all. I am a software engineer."*
6. Paste contents of this file to restore context
