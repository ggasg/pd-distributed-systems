---
week_number: 18
status: not-started
---

# W18: Capstone

> **Arc:** Distributed ML & Compute · **Language:** Go, Java, or Python, your choice

## What you'll build
One system that combines at least two concepts from this curriculum. It should be something you can explain end-to-end, from the data model to the failure behavior. No scaffolding provided; you design it.

**Scenario:** every option below asks you to break something on purpose, kill a primary, crash a worker, corrupt a query's incremental state, and prove your design survives it, not just that it works when nothing goes wrong. That's deliberate: a design you can only defend on the happy path isn't actually a design yet.

---

## Choose One

### Option A: Distributed KV Store (Go)
Combine W01 (LSM storage) + W04 (clocks/ordering).

A 3-node key-value store in Go where writes are replicated via a simple primary-backup protocol (not Raft, keep it tractable; you get real Raft in W17's reading and W19's live etcd cluster instead, without paying this exercise's time budget for a full leader-election-plus-log-replication implementation). The primary assigns a logical timestamp to each write using a Lamport clock before forwarding to backups. Supports `get`, `put`, `delete`. Test: kill the primary, promote a backup, verify reads are consistent.

**Why Go:** it's a direct extension of the actual code from W01 and W04, not a rewrite in a third language; goroutines and channels make the node communication natural, and `net/http` (standard library) makes the client API trivial without reaching for a framework.

**If you pick this option, read first: DDIA Chapter 6** (2nd ed., Replication), specifically "Single-Leader Replication." Primary-backup *is* the leader-based replication Ch. 6 describes; the chapter names the failure modes worth designing around before you write `Promote()` (replication lag, what happens to in-flight writes when the leader dies mid-forward) rather than discovering them by hand.

**Optional companion: DDIA Chapter 7** (2nd ed., Sharding, renamed from "Partitioning"). Your 3-node store only replicates, it doesn't shard the keyspace, but Ch. 7 is the other half of the scaling story Ch. 6 started: replication makes each copy of the data more available, sharding is what lets the dataset grow past what one node holds. Worth reading for the concept even though this exercise doesn't implement it; the "What you'd add with another week" Reflect question is a natural place to sketch how you'd combine the two.

**Minimum bar:** 3-node cluster, primary-backup replication works, one node can fail and the system continues.

---

### Option B: Streaming Pipeline with Exactly-Once (Java)
Combine W05 (stream processing) + W17 (snapshots).

A stateful streaming word count in Java that periodically checkpoints using Chandy-Lamport snapshots. On simulated failure: restore from the latest snapshot, replay messages from that point, verify the final word count matches a non-failing run.

**Why Java:** it's a direct extension of the actual code from W05 and W17, both already Java; the `TumblingWindowAggregator` and the sealed-interface `Message`/snapshot machinery from W17 plug together without a rewrite.

**Minimum bar:** windowed word count, periodic snapshots, crash-and-recover test passes.

---

### Option C: Incremental Query Engine (Java)
Combine W07 (differential dataflow) + W08 (query execution).

Extend your W07 DD engine with vectorized, batch-at-a-time operator execution: `filter` and `join` operate on batches of updates rather than one at a time, the same technique W08 applies, reimplemented here in Java against `Collection<K, V>` rather than imported directly from W08's code, since W08 is a separate Go module. Benchmark: insert 10k rows, run a filter+join query, then update 100 rows and measure incremental re-evaluation vs full re-execution.

**Why this is still a fair combination even though W07 and W08 are different languages:** the thing being combined is the *technique* (batch-at-a-time columnar processing applied to incrementally-maintained collections), not literal shared code. You already know what batch-vectorized `filter` and `join` look like from W08; this option asks you to bring that technique into W07's Java `Collection` API, which is a more realistic version of "combine two concepts" than gluing two language runtimes together with an RPC boundary would be.

**Minimum bar:** filter + join working incrementally, measurable speedup over full re-execution on updates.

---

### Option D: GPU-Accelerated Distributed Training (Python)
Combine W13 (ring-allreduce) + W15 (GPU-accelerated GEMM).

The only option that stays inside this arc rather than reaching back into Arc 1/Arc 2, worth choosing if distributed training and compute-intensive AI workflows specifically are what you're optimizing this curriculum for. Take W13's 2-worker ring-allreduce training loop and replace the MLP's NumPy matrix multiplies with your W15 tiled CUDA GEMM kernel, so gradient exchange still happens over real TCP sockets between workers, but the compute inside each worker is GPU-accelerated instead of CPU NumPy. Benchmark per-epoch wall time, W13's CPU-only baseline vs. this GPU-accelerated version, and break down where time actually goes: compute or network.

**No GPU?** Use W15's cache-blocked/AVX2 C kernel via `ctypes` instead of CUDA: same comparison, CPU-baseline vs. optimized-kernel, without requiring hardware you may not have.

**Minimum bar:** the 2-worker ring-allreduce loop runs end-to-end with the GPU (or SIMD C) kernel doing the matmuls, converges to comparable accuracy to W13, and your writeup names where wall-clock time goes at each worker count.

---

## Deliverables

- [ ] Working code in `code/capstone/`
- [ ] `code/capstone/README.md`: explains what it does, the design decisions, what you'd do differently, what breaks at scale

---

## Reflect

**Which option you chose and why:**

**The hardest part:**

**What you'd add with another week:**

**What you know now that you didn't at W01:**
