---
week_number: 16
status: not-started
---

# W16: Grand Capstone: Distributed Training & Serving Platform

> **Arc:** Optional capstone · **Language:** Python (+ Helm/YAML), reuses W08, W09, W12, W13, W15; deploys via Kubeflow Trainer from W14
> **Budget:** this one is a project, not a 5-hour unit. Expect 20 to 30 hours spread over as long as it takes, and treat each Part as its own sitting. It is optional for exactly this reason.

> **Status:** Optional stretch project, not a unit. Not required to finish the core curriculum (W00 to W15).

**Prerequisite:** W08 (feature pipeline), W09 (distributed training), W12 (attention + KV cache), W13 (Chandy-Lamport snapshots), W14 (Kubernetes Operators), and W15 (Observability) all completed.

## What you'll build

A small end-to-end distributed ML platform on your kind cluster: a versioned feature pipeline feeds a tiny attention-based model that trains across multiple worker Pods via ring-allreduce, checkpoints itself so a killed worker resumes instead of retraining from scratch, runs as a `TrainJob` under the Kubeflow Trainer operator you installed in W14 (you're operating a real operator here, not extending one you wrote), lands in a model registry that records which data version produced it, and is served afterward from a specific registered version through a KV-cached inference endpoint, all instrumented with the Prometheus/Grafana stack from W15.

**Why this unit exists:** it chains six units into a single working system, end to end, on your own cluster: data in, model trained, failure survived, model served, everything observed. It's the closest thing in this curriculum to what you'd actually build on the job, and it's the only place the pieces get tested against each other rather than in isolation.

---

## Read

No new required reading tied to this unit's build. If any of these feel rusty, skim your own notes before starting:

- W09: ring-allreduce, and why it's bandwidth-efficient
- W12: KV cache, and why it turns O(N²) generation into O(N)
- W13: Chandy-Lamport, and what a *consistent* recorded state actually means
- W14: the reconcile loop, and why it's level-triggered
- W15: the four golden signals

One genuinely new read, if you want it: **DDIA Chapter 13** (2nd ed.), A Philosophy of Streaming Systems (renamed from "The Future of Data Systems"). It's the book's own synthesis chapter (unbundling databases into composable derived-data systems, correctness as a property of the whole pipeline rather than any one component), and this is the unit where that stops being abstract and becomes the actual shape of what you built: a feature store, a trainer, a checkpoint coordinator, an operator, and a server, each correct on its own, wired into one pipeline where correctness is a property of the whole. Fitting bookend to a curriculum that leaned on this book throughout.

**Depth: skim your own notes, study nothing new.** This unit is synthesis. If a concept feels shaky, the fix is to reread what you wrote in that unit's Reflect section, not to reread the paper.

**Key question:** If a worker dies mid-training and the operator restarts it, what exactly needs to be true about the checkpoint for the resumed training to be *correct*, not just "the process didn't crash"?

**Second key question:** Six months from now someone asks why a prediction from last March looked wrong. What would you need to have recorded, at training time, to answer that at all?

---

## Code

Project: `code/capstone-platform/` (Python for training, coordination, and serving; Helm/YAML for deploying as a `TrainJob` via the Kubeflow Trainer operator from W14; reusing your W08/W09/W12 code directly)

**Part 1: Data (reuse W08)**

- [ ] Reuse `feature_pipeline/` from W08 (or a trimmed copy) to generate a versioned training dataset. No changes needed; this is the input to Part 2.

**Part 2: Distributed training (combine W09 and W12)**

- [ ] `train_worker.py`: combine W09's `ring_allreduce.py` with a small attention model from W12's `attention.py` (instead of the MLP from W09). Each worker trains on a shard of the feature data and exchanges gradients via ring-allreduce.
- [ ] Every N steps, each worker writes a checkpoint (model weights + optimizer state + current step number) to a shared path: `checkpoints/worker-{rank}/step-{n}.npz`

**Part 3: Fault tolerance (extend W13)**

- [ ] `checkpoint_coordinator.py`: adapt the Chandy-Lamport idea from W13: when a checkpoint is triggered, the coordinator waits until all workers have paused at the *same* training step (not mid-allreduce) before recording state. This is what makes the recorded checkpoint a genuinely consistent cut across workers, rather than N independent snapshots that disagree with each other.
- [ ] Kill a worker process mid-training (`kill -9`). Verify: the operator (Part 4) restarts it, it loads the last consistent checkpoint, and training resumes without corrupting the other workers' state.

**Your call:** `checkpoint_coordinator.py` waits until every worker has paused at the same step before recording a checkpoint. Now kill a worker permanently (don't let Part 4's operator restart it, or kill it before starting Part 4) right before a checkpoint is due. The coordinator is now waiting on a step-pause signal from a worker that will never send one, forever. Decide: should the coordinator time out and checkpoint the workers that did pause, accepting that the resulting snapshot is missing one worker's state (and deciding what that even means for a consistent restart), or should a missing worker block checkpointing entirely, on the reasoning that a checkpoint without every worker's state isn't a real consistent cut at all? This is the same "wait for all N or proceed with a quorum" trade-off from W09 and W11, applied to a real consistency mechanism instead of a gradient exchange; whichever you pick, implement a timeout in `checkpoint_coordinator.py` and note what you'd tell a teammate about what the resulting checkpoint does and doesn't guarantee.

**Part 4: Orchestration (operate Kubeflow Trainer from W14)**

- [ ] Deploy your Part 2 training workers as a `TrainJob` instead of raw Pods or `multiprocessing.Process`: `spec.trainer.numNodes` set to your worker count, each node running `train_worker.py`. Reuse the Trainer install from W14 if it's still on your cluster. If you also stood up a queueing layer in W14's optional Part 4, submit through it so the job is admitted as a gang rather than piecemeal; if not, just make sure your cluster has room for every node Pod before you submit, which is the manual version of the same guarantee.
- [ ] `kill -9` a worker process inside its container, or `kubectl delete pod` on one node Pod directly, the same test you ran in W14 Part 1. The operator's own reconcile loop recreates the Pod; you write zero orchestration code for this. That's the actual point of operating a real operator here instead of hand-rolling one: the "restart a dead worker" mechanism already exists and already works.
- [ ] The recovery logic that matters is entirely at the application layer, not the operator layer, and W14 Part 1 already showed you why: the restarted Pod comes back with no memory of anything. On startup, `train_worker.py` checks `checkpoints/worker-{rank}/` for the latest checkpoint written by `checkpoint_coordinator.py` (Part 3) and resumes from it instead of starting at step 0. The operator's job is only "keep N node Pods running"; your code's job is "come back correctly when one of them restarts." Keeping those two responsibilities separate, rather than teaching the operator about your checkpoint format, is itself the lesson: it's the same division of labor a real managed training platform uses.
- [ ] Track restart events yourself for the Part 6 dashboard: increment a `worker_restarts_total` Prometheus counter from inside `train_worker.py` the moment it detects it's resuming from a checkpoint rather than starting cold, rather than trying to read Pod restart counts back out of Kubernetes.

**Part 5: The handoff (MLflow model registry)**

Parts 1 through 4 produce a trained model. Part 6 serves one. Right now nothing connects them except a file path, and that is a real hole rather than a detail: you cannot say which data version a served model was trained on, you cannot roll back to the previous one, and "what is in production" is answerable only by looking at a disk.

This is W08's lesson again, one level up. Replacing a `latest.txt` pointer with a commit log is what an open table format did for your feature data. A model registry does the same thing for models.

- [ ] Deploy MLflow to your kind cluster: a small `Deployment` running `mlflow server` plus a `Service` so Pods can reach it by name. Ephemeral storage is fine here; this is a capstone exercise, not a system of record. Point `MLFLOW_TRACKING_URI` at that Service from both the training and serving Pods.
- [ ] From `train_worker.py`, log the run and register the finished model. Log it from **rank 0 only**. Every worker calling `mlflow.log_model` gives you N near-identical registered versions per training run, which is not a subtle failure but is an easy one to ship, since each worker is running correct code in isolation.
- [ ] Tag the registered version with the W08 feature-store version it trained on. That single tag is what makes the whole thing worth doing: it turns "which model is this?" into a question with an answer that reaches all the way back to the data.
- [ ] Have `serve.py` (Part 6) load from the registry, `models:/<name>/<version>`, rather than from a path on disk.

**Your call:** you can point serving at a pinned version number or at an alias like `@champion` that you move between versions. Pinning means the deployment spec fully describes what is running, and changing models requires a deploy. An alias means you can promote a new model without touching the deployment, and it also means nobody can tell you what is serving by reading the YAML. Pick one, implement it, then register a second model version and perform a rollback under your chosen scheme. Write down how many steps the rollback took and who would have needed to be involved.

**Part 6: Serving + observability (extend W12 and W15)**

- [ ] `serve.py`: a small HTTP service (Flask/FastAPI is fine here, this isn't the distributed part) that loads the registered model version from Part 5 and serves `/generate` using the KV cache from W12
- [ ] Instrument the serving endpoint the same way you instrumented the DD engine in W15 (Python's `prometheus_client` this time, not the Java Prometheus client library): `requests_total` (Counter), `generation_latency_seconds` (Histogram), `active_kv_cache_entries` (Gauge)
- [ ] Deploy everything to kind: training Job, operator, serving Deployment + ServiceMonitor
- [ ] Grafana dashboard, 4 panels: training throughput (updates/sec across workers), checkpoint/restore events over time, inference p99 latency, active KV cache size

**Minimum bar:** training runs across 2+ worker Pods; killing one mid-training triggers a real recovery (not a restart from step 0); the trained model is registered with the feature version it trained on and served from a specific registered version, not a file path; training, recovery, and serving all show up on one Grafana dashboard.

---

## Deliverables

- [ ] Working code in `code/capstone-platform/`
- [ ] `code/capstone-platform/README.md`: design doc: what it does, where the consistent-checkpoint logic lives, how a served prediction traces back to a data version, and what you'd change for a real multi-node (not single-machine) deployment

---

## Reflect

**What was the hardest integration point? Where did two units' code not fit together as cleanly as you expected?**

**What does "consistent checkpoint" mean in your system, concretely? Not the textbook definition, the thing you actually had to guarantee.**

**Timeout-and-proceed or block-forever-without-quorum for `checkpoint_coordinator.py`, and what does the resulting checkpoint actually guarantee (from Your call above)?**

**Pinned version or moving alias for serving, how many steps your rollback took, and who else would have had to be involved:**

**If you scaled this from 2 workers on one kind cluster to 50 workers on real hardware, what's the first thing that breaks?**

**What you know now, having built this, that you didn't know after W14:**
