---
week_number: 17
status: not-started
---

# W17: Capstone

> **Arc:** Distributed ML & Compute · **Language:** Go, C++, or Python, your choice

## What you'll build
One system that combines at least two concepts from this curriculum. It should be something you can explain end-to-end, from the data model to the failure behavior. No scaffolding provided; you design it.

---

## Choose One

### Option A: Distributed KV Store (Go)
Combine W01 (LSM storage) + W04 (clocks/ordering).

A 3-node key-value store in Go where writes are replicated via a simple primary-backup protocol (not Raft, keep it tractable). The primary assigns a logical timestamp to each write using a Lamport clock before forwarding to backups. Supports `get`, `put`, `delete`. Test: kill the primary, promote a backup, verify reads are consistent.

**Why Go:** goroutines + channels make the node communication natural; Go's stdlib HTTP makes the client API trivial. This is the kind of tool engineers actually write in Go.

**Minimum bar:** 3-node cluster, primary-backup replication works, one node can fail and the system continues.

---

### Option B: Streaming Pipeline with Exactly-Once (C++)
Combine W05 (stream processing) + W16 (snapshots).

A stateful streaming word count in C++ that periodically checkpoints using Chandy-Lamport snapshots. On simulated failure: restore from the latest snapshot, replay messages from that point, verify the final word count matches a non-failing run.

**Minimum bar:** windowed word count, periodic snapshots, crash-and-recover test passes.

---

### Option C: Incremental Query Engine (C++)
Combine W07 (differential dataflow) + W08 (query execution).

Extend your W07 DD engine with vectorized operator execution from W08: `filter` and `join` operate on batches of updates rather than one at a time. Benchmark: insert 10k rows, run a filter+join query, then update 100 rows and measure incremental re-evaluation vs full re-execution.

**Minimum bar:** filter + join working incrementally, measurable speedup over full re-execution on updates.

---

## Deliverables

- [ ] Working code in `code/capstone/`
- [ ] `code/capstone/README.md`: explains what it does, the design decisions, what you'd do differently, what breaks at scale
- [ ] `posts/W17-capstone.md`: a technical post (500–1000 words) for a dev blog or GitHub. Explain the system to a fellow engineer who hasn't done this curriculum.

---

## Reflect

**Which option you chose and why:**

**The hardest part:**

**What you'd add with another week:**

**What you know now that you didn't at W01:**
