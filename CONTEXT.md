# Claude Session Context

> Paste this at the start of any new Cowork session to restore context.
> This file is the single source of truth — no Notion.

---

## Goal

Become a pragmatic master of Distributed, Data-Intensive Systems. Focus areas:
- ML Model Training infrastructure
- Compute-Intensive Distributed Execution
- AI Tasks / LLM inference systems

Working at Materialize (streaming SQL on Differential Dataflow). 2h/day, 5 days/week.

---

## Repo

`/Users/gaston/dev/pd-distributed-systems` — Obsidian vault + GitHub repo.

Vault structure:
- `weeks/W01-*.md` through `W14-*.md` — one file per week (task checklist + notes)
- `README.md` — public-facing GitHub description

No Notion. No separate task tracker. Everything lives here.

---

## Curriculum

14 weeks. 3 arcs.

### Arc 1: Data Systems Internals (W01–W04) — Java 21

| Week | Topic | Key Paper | Deliverable |
|------|-------|-----------|-------------|
| W01 | LSM-Trees + Storage Engines | DDIA Ch.3; LevelDB source | `lsm/` Java project |
| W02 | Encoding + Wire Formats | DDIA Ch.4; Protobuf encoding spec | `encoding/` benchmark |
| W03 | Raft Consensus | Ongaro & Ousterhout (2014) | `raft/` Java — leader election + log replication |
| W04 | Clocks, Causality, Time | Lamport (1978); DDIA Ch.8; Spanner TrueTime | `clocks/` Java — vector clocks + causal delivery |

### Arc 2: Streaming and Dataflow (W05–W08) — Scala / Rust

| Week | Topic | Key Paper | Deliverable |
|------|-------|-----------|-------------|
| W05 | Stream Processing Primitives | Dataflow Model (Akidau et al., 2015); DDIA Ch.11 | `streaming/` Scala — windowed aggregation |
| W06 | Naiad + Timely Dataflow | Naiad paper (Murray et al., SOSP 2013) | `timely-toy/` Scala — timestamps + notifications |
| W07 | Differential Dataflow | DD paper (McSherry, 2013) | `dd-examples/` Rust — incremental word count + reachability |
| W08 | Query Execution | Volcano (Graefe, 1994); MonetDB/X100 (Boncz, 2005) | `query-exec/` Rust — vectorized filter + hash join |

### Arc 3: Distributed ML & Compute (W09–W14) — Python / CUDA

| Week | Topic | Key Paper | Deliverable |
|------|-------|-----------|-------------|
| W09 | ML Data Pipelines | Hidden Tech Debt (Sculley, 2015); Delta Lake (2020) | `feature-pipeline/` Python — versioned features with Parquet + DuckDB |
| W10 | Distributed Training | Horovod (2018); PyTorch DDP source | `distributed-training/` Python — manual allreduce, 2-worker MLP |
| W11 | GPU Memory + Compute | CUDA Programming Guide Ch.1-3; Roofline (2009) | `gpu-gemm/` — naive vs tiled CUDA matmul + roofline analysis |
| W12 | Attention + KV Cache | Attention Is All You Need (2017); FlashAttention (2022); PagedAttention (2023) | `attention/` Python — MHA forward pass + KV cache, NumPy only |
| W13 | Fault Tolerance | Chandy-Lamport (1985); Flink ABS (2015) | `snapshot/` Java — Chandy-Lamport in simulated 3-node system |
| W14 | Capstone | — | Choose one: replicated KV, streaming pipeline, or query engine |

---

## Current State

> Update this section at the end of each week.

- [ ] W01 — not started
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

---

## Cowork Setup on a New Machine

1. Clone repo: `git clone <repo-url> /Users/gaston/dev/pd-distributed-systems`
2. Open Cowork → Select folder → `/Users/gaston/dev/pd-distributed-systems`
3. Create project "Professional Development" with instructions:
   > *"Be direct and concise with practical, action-oriented suggestions. My objective is to become a pragmatic master of Distributed, Data-Intensive Systems — not a theory know-it-all. I am a software engineer."*
4. Paste contents of this file to restore context
