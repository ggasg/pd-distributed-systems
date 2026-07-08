# Claude Session Context

> Paste this at the start of any new Cowork session to restore context.
> This file is the single source of truth — no Notion.

---

## Goal

Become a pragmatic master of Distributed, Data-Intensive Systems. Focus areas:
- ML Model Training infrastructure
- Compute-Intensive Distributed Execution
- AI Tasks / LLM inference systems

2h/day, 5 days/week.

---

## Repo

`~/dev/pd-distributed-systems` — Obsidian vault + GitHub repo.

Vault structure:
- `config.md` — set `start_date` here; all week dates recalculate in Home.md
- `Home.md` — daily entry point; DataviewJS auto-detects current week and shows schedule
- `weeks/W00-*.md` through `W16-*.md` — one file per week; each has Read, Code, optional Python DSA Review, and Reflect sections
- `weeks/W17-*.md` — optional grand capstone; synthesizes W09, W10, W12, W13, W15, W16 into one distributed training + serving platform
- `tools/plan-dates.go` — `go run tools/plan-dates.go --start 2026-07-13` prints full schedule
- `SETUP.md` — full environment setup (Java, Scala, Python, Go, Docker, Obsidian)
- `RESOURCES.md` — all papers and books by week with free links
- `Templates/week-template.md` — blank week file for adding custom weeks
- `posts/TEMPLATE.md` — structured blog post format for weekly write-ups
- `code/README.md` — expected directory tree and build commands for all weeks
- `README.md` — public-facing GitHub description

No Notion. No separate task tracker. Everything lives here.

---

## Curriculum

17 core weeks (W00 pre-week + W01–W16) across 4 arcs, plus an optional W17 grand capstone.

### W00: Infrastructure Setup (pre-week)
| Week | Topic | Deliverable |
|------|-------|-------------|
| W00 | Local k8s (kind) + Prometheus + Grafana | `code/hello-metrics/` Go service with Prometheus metrics, deployed to kind |

### Arc 1: Data Systems Internals (W01–W04) — Java 21

| Week | Topic | Key Paper | Deliverable |
|------|-------|-----------|-------------|
| W01 | LSM-Trees + Storage Engines | DDIA Ch.3; LevelDB source | `lsm/` Java — MemTable, SSTable, LSMTree |
| W02 | Encoding + Wire Formats | DDIA Ch.4; Protobuf encoding spec | `encoding/` Java — varint, row vs column store benchmark |
| W03 | MapReduce and Its Limits | Dean & Ghemawat (2004); Zaharia et al. (2012) | `mapreduce/` Java — MR framework, word count, iterative PageRank; Go HTTP coordinator |
| W04 | Clocks, Causality, Time | Lamport (1978); DDIA Ch.8 | `clocks/` Java — vector clocks + causal delivery |

### Arc 2: Streaming and Dataflow (W05–W08) — Scala

| Week | Topic | Key Paper | Deliverable |
|------|-------|-----------|-------------|
| W05 | Stream Processing Primitives | Dataflow Model (Akidau et al., 2015); DDIA Ch.11 | `streaming/` Scala — windowed aggregation, purely functional |
| W06 | Naiad + Timely Dataflow | Naiad paper (Murray et al., SOSP 2013) | `timely-toy/` Scala — Timestamp, Pointstamp, ProgressTracker |
| W07 | Differential Dataflow | DD paper (McSherry, 2013) | `dd-scratch/` Scala — DD engine from scratch: Update, Collection, WordCount, Reachability |
| W08 | Query Execution | Volcano (1994); MonetDB/X100 (2005) | `query-exec/` Scala — RowExecutor, ColumnFilter, HashJoin, Benchmark (3–8x speedup) |

### Arc 3: Distributed ML & Compute (W09–W14) — Python / Go secondary

| Week | Topic | Key Paper | Deliverable |
|------|-------|-----------|-------------|
| W09 | ML Data Pipelines | Hidden Tech Debt (2015); Delta Lake (2020) | `feature-pipeline/` Python — versioned features, Parquet + DuckDB |
| W10 | Distributed Training | Horovod (2018); PyTorch DDP source | `distributed-training/` Python — ring-allreduce via sockets, 2-worker MLP; Go gradient server |
| W11 | GPU Memory + Compute | CUDA Guide Ch.1-3; Roofline (2009) | `gpu-gemm/` Python/Numba — naive vs tiled CUDA matmul + roofline; Go bench runner |
| W12 | Attention + KV Cache | Attention Is All You Need (2017); FlashAttention (2022); PagedAttention (2023) | `attention/` Python — MHA forward pass + KV cache, NumPy only |
| W13 | Fault Tolerance | Chandy-Lamport (1985); Flink ABS (2015) | `snapshot/` Java — Chandy-Lamport 3-node simulation; Go rewrite optional |
| W14 | Capstone | — | Go: replicated KV store / Scala: streaming pipeline / Scala: incremental query engine |

### Arc 4: Infrastructure (W15–W16) — Go / Scala

| Week | Topic | Deliverable |
|------|-------|-------------|
| W15 | Kubernetes Operators | `code/operator/` Go — custom DistributedJob CRD + reconciler |
| W16 | Observability: Metrics, Tracing, Logging | Instrument W07 DD engine with Prometheus + OTel; Grafana dashboard |

### Optional: W17 — Grand Capstone (stretch week)

| Week | Topic | Deliverable |
|------|-------|-------------|
| W17 | Distributed training + serving platform | `code/capstone-platform/` — training across worker Pods (W10+W12), Chandy-Lamport-style checkpoint/restore (W13), operator-managed recovery (W15), fully observed in Grafana (W16) |

---

## Date Configuration

Edit `start_date` in `config.md` to reschedule. Current config (`2026-07-13`):
- W00: Jul 6 – Jul 12
- W01: Jul 13 – Jul 19
- W14: Oct 12 – Oct 18
- W15: Oct 19 – Oct 25
- W16: Oct 26 – Nov 1
- W17 (optional): Nov 2 – Nov 8

`go run tools/plan-dates.go --start 2026-07-13` prints the full table.

---

## Current State

> Update this section at the end of each week.

- [ ] W00 — not started
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
- [ ] W17 (optional)

---

## Branch Workflow

Two long-lived branches:
- `main` — curriculum authoring. Structural edits to weeks, resources, and setup docs (what happens in a Cowork session when refining the plan).
- `progress` — actually running the plan. Checked-off tasks, filled-in Reflect answers, `Current State` updates. Never merged back into `main` — it only receives.

To pull curriculum updates from `main` into `progress` without losing tracked progress:

```bash
git checkout progress
git pull origin progress
git fetch origin main
git merge origin/main
git push origin progress
```

This merges cleanly as long as answers always go on the blank line *below* each Reflect question — never on the question line itself. That keeps curriculum edits (the question line) and progress edits (the answer line) on different lines, which is what avoids merge conflicts in Markdown.

---

## Cowork Setup on a New Machine

1. Clone repo: `git clone <repo-url> ~/dev/pd-distributed-systems`
2. Open Cowork → Select folder → `~/dev/pd-distributed-systems`
3. Create project "Professional Development" with instructions:
   > *"Be direct and concise with practical, action-oriented suggestions. My objective is to become a pragmatic master of Distributed, Data-Intensive Systems — not a theory know-it-all. I am a software engineer."*
4. Paste contents of this file to restore context
