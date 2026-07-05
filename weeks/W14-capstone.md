# W14 — Capstone

> **Arc:** Distributed ML & Compute · **Language:** Your choice

## What you'll build
One system that combines at least two concepts from this curriculum. It should be something you can explain end-to-end, from the data model to the failure behavior. No scaffolding provided — you design it.

---

## Choose One

### Option A: Replicated Key-Value Store
Combine W01 (LSM storage) + W03 (Raft consensus).

A KV store where writes go through Raft log replication before being applied to an LSM-tree. Reads from the leader. Supports `get`, `put`, `delete`. Handle leader failure: restart a node, verify it catches up via log replay.

**Minimum bar:** 3-node cluster, leader election works, put/get works, one node can crash and rejoin.

---

### Option B: Streaming Pipeline with Exactly-Once
Combine W05 (stream processing) + W13 (snapshots).

A stateful streaming word count that periodically checkpoints using Chandy-Lamport snapshots. On simulated failure: restore from the latest snapshot, replay messages from that point, verify the final count matches a non-failing run.

**Minimum bar:** windowed word count, periodic snapshots, crash-and-recover test passes.

---

### Option C: Incremental Query Engine
Combine W07 (differential dataflow) + W08 (query execution).

A query engine over a changing dataset: supports `SELECT ... WHERE ... JOIN ...` expressed as a DD dataflow graph. Insert/delete rows and observe the query result update incrementally without re-executing from scratch. Benchmark incremental update vs full re-execution.

**Minimum bar:** filter + join working incrementally, measurable speedup over re-execution on updates.

---

## Deliverables

- [ ] Working code in `code/capstone/`
- [ ] `code/capstone/README.md` — explains: what it does, the design decisions you made, what you'd do differently, what breaks at scale
- [ ] `posts/W14-capstone.md` — a technical post (500–1000 words) suitable for a dev blog or GitHub. Explain the system to a fellow engineer who hasn't done this curriculum. This is the artifact you'd share publicly.

---

## Reflect

**Which option you chose and why:**

**The hardest part:**

**What you'd add with another week:**

**What you know now that you didn't at W01:**
