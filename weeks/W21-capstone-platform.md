---
week_number: 21
status: not-started
---

# W21: Grand Capstone: Distributed Training & Serving Platform

> **Arc:** Optional capstone · **Language:** Go, Python, C++, reuses W11, W13, W16, W17, W19, W20
> **Status:** Optional / stretch week. Not required to finish the core curriculum (W00–W20).

**Prerequisite:** W11 (feature pipeline), W13 (distributed training), W16 (attention + KV cache), W17 (Chandy-Lamport snapshots), W19 (Kubernetes Operators), and W20 (Observability) all completed.

## What you'll build

A small end-to-end distributed ML platform on your kind cluster: a versioned feature pipeline feeds a tiny attention-based model that trains across multiple worker Pods via ring-allreduce, checkpoints itself so a killed worker resumes instead of retraining from scratch, is orchestrated by a Kubernetes Operator extended from W18, and is served afterward through a KV-cached inference endpoint, all instrumented with the Prometheus/Grafana stack from W19.

**Why this week exists:** W18 combines two arcs. This one chains six weeks into a single working system, end to end, on your own cluster: data in, model trained, failure survived, model served, everything observed. It's the closest thing in this curriculum to what you'd actually build on the job.

---

## Read

No new required reading tied to this week's build. If any of these feel rusty, skim your own notes before starting:

- W13: ring-allreduce, and why it's bandwidth-efficient
- W16: KV cache, and why it turns O(N²) generation into O(N)
- W17: Chandy-Lamport, and what a *consistent* recorded state actually means
- W19: the reconcile loop, and why it's level-triggered
- W20: the four golden signals

One genuinely new read, if you want it: **DDIA Chapter 12**, The Future of Data Systems. It's the book's own synthesis chapter (unbundling databases into composable derived-data systems, correctness as a property of the whole pipeline rather than any one component), and this is the week where that stops being abstract and becomes the actual shape of what you built: a feature store, a trainer, a checkpoint coordinator, an operator, and a server, each correct on its own, wired into one pipeline where correctness is a property of the whole. Fitting bookend to a curriculum that leaned on this book from W00 onward.

**Key question:** If a worker dies mid-training and the operator restarts it, what exactly needs to be true about the checkpoint for the resumed training to be *correct*, not just "the process didn't crash"?

---

## Code

Project: `code/capstone-platform/` (Go for the operator and serving layer, Python for training, reusing your W11/W13/W16 code directly)

**Part 1: Data (reuse W11)**

- [ ] Reuse `feature_pipeline/` from W11 (or a trimmed copy) to generate a versioned training dataset. No changes needed; this is the input to Part 2.

**Part 2: Distributed training (combine W13 and W16)**

- [ ] `train_worker.py`: combine W13's `ring_allreduce.py` with a small attention model from W16's `attention.py` (instead of the MLP from W13). Each worker trains on a shard of the feature data and exchanges gradients via ring-allreduce.
- [ ] Every N steps, each worker writes a checkpoint (model weights + optimizer state + current step number) to a shared path: `checkpoints/worker-{rank}/step-{n}.npz`

**Part 3: Fault tolerance (extend W17)**

- [ ] `checkpoint_coordinator.py`: adapt the Chandy-Lamport idea from W17: when a checkpoint is triggered, the coordinator waits until all workers have paused at the *same* training step (not mid-allreduce) before recording state. This is what makes the recorded checkpoint a genuinely consistent cut across workers, rather than N independent snapshots that disagree with each other.
- [ ] Kill a worker process mid-training (`kill -9`). Verify: the operator (Part 4) restarts it, it loads the last consistent checkpoint, and training resumes without corrupting the other workers' state.

**Part 4: Orchestration (extend W19)**

- [ ] Extend your `DistributedJob` CRD (or add a new `TrainingJob` CRD) with a `checkpointPath` field
- [ ] Update `reconciler.go`: on detecting a crashed worker Pod (not Ready), recreate it with an env var pointing at the last known-good checkpoint path instead of always starting fresh
- [ ] `status.Phase` gains a new value, `Recovering`, set while a restarted worker is loading its checkpoint

**Part 5: Serving + observability (extend W16 and W20)**

- [ ] `serve.py`: a small HTTP service (Flask/FastAPI is fine here, this isn't the distributed part) that loads the trained model and serves `/generate` using the KV cache from W16
- [ ] Instrument the serving endpoint the same way you instrumented the DD engine in W20 (Python's `prometheus_client` this time, not the C++ `prometheus-cpp` library): `requests_total` (Counter), `generation_latency_seconds` (Histogram), `active_kv_cache_entries` (Gauge)
- [ ] Deploy everything to kind: training Job, operator, serving Deployment + ServiceMonitor
- [ ] Grafana dashboard, 4 panels: training throughput (updates/sec across workers), checkpoint/restore events over time, inference p99 latency, active KV cache size

**Minimum bar:** training runs across 2+ worker Pods; killing one mid-training triggers a real recovery (not a restart from step 0); the trained model is served with KV-cached generation; training, recovery, and serving all show up on one Grafana dashboard.

---

## Deliverables

- [ ] Working code in `code/capstone-platform/`
- [ ] `code/capstone-platform/README.md`: design doc: what it does, where the consistent-checkpoint logic lives, what you'd change for a real multi-node (not single-machine) deployment
- [ ] `posts/W21-capstone-platform.md`: a technical post (500–1000 words): walk a fellow engineer through the data → train → fail → recover → serve → observe loop

---

## Reflect

**What was the hardest integration point? Where did two weeks' code not fit together as cleanly as you expected?**

**What does "consistent checkpoint" mean in your system, concretely? Not the textbook definition, the thing you actually had to guarantee.**

**If you scaled this from 2 workers on one kind cluster to 50 workers on real hardware, what's the first thing that breaks?**

**What you know now, having built this, that you didn't know after W19:**
