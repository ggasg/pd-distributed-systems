---
week_number: 14
status: not-started
---

# W14: Operating Kubernetes Operators: Kubeflow Trainer and Spark Operator

> **Arc:** Infrastructure · **Language:** Helm/YAML; you'll read Go, not write it
> **Budget:** about 10 hours. The Minimum bar is what a bad week looks like, not the target.

## What you'll build

Not build, operate. Deploy two real, production-grade Kubernetes operators to your kind cluster: Kubeflow Trainer (the vendor-neutral way to run distributed training jobs on Kubernetes, and a PyTorch Ecosystem project) and Kubeflow's Spark Operator (the standard way to run Apache Spark on Kubernetes). Create a `TrainJob` and a `SparkApplication`, watch each one reconcile, break each on purpose, and debug it the way you'd debug someone else's operator in production. Then read, don't write, each operator's real reconcile loop: the production version of the level-triggered control loop this unit is actually about.

Part 4 goes one layer up, to the question of who gets scheduled when there isn't enough hardware for everybody. Part 3 is an optional stretch that goes one layer down instead, to the etcd cluster running Raft underneath the control plane both operators depend on; skip it on a normal pass.

**Why operate rather than write an operator?** Writing your own CRD types and a `controller-runtime` reconciler is the skill set of someone *building* a new operator, which is a narrow platform-infrastructure specialization. It's not what a Staff Data Platform Engineer, a Field/Customer/Professional Services Engineer, or a Developer Advocate does day to day; those roles *operate* operators someone else wrote: install via Helm, configure a CR, read logs, debug a stuck reconcile loop. This unit teaches that instead. You've already written Go by now, W00 through W03 all use it, so the reconciler source below reads as familiar syntax applied to a much bigger, real codebase, not a cold start; what this unit specifically avoids isn't Go itself, it's the much larger, narrower skill of authoring a `controller-runtime` operator from scratch.

**Why Kubeflow Trainer rather than a framework-specific operator?** Because `TrainJob` is the most portable distributed-training abstraction available. Trainer v2 collapsed the old framework-specific CRDs (`PyTorchJob`, `MPIJob`, `JAXJob`, `XGBoostJob`) into one `TrainJob` plus a pluggable runtime, it's governed by Kubeflow rather than by any single vendor, and it runs the same way on a managed cloud offering as it does on the kind cluster on your laptop. Anything you learn here transfers regardless of which company you end up at, which is exactly what a vendor-specific operator cannot promise.

**A pointer back to W12.** The router you wrote in W12 Part 2, the one that sends a request to the replica already holding its KV cache, is a control-plane component in production, not part of the model server. It runs as a Kubernetes service, it's written in Go, and it needs exactly the kind of per-replica state that the operators below spend their lives tracking. Keep that in mind while reading the reconcilers: the thing you built by hand for two in-process replicas is a small version of what this layer does for a cluster.

**Scenario:** this is what your first week on call for a platform team actually looks like: nothing you're debugging is code you wrote, and the job is reading logs, forming a hypothesis, and checking it, not fixing a bug in your own mental model of a system you built yourself.

**Prerequisite:** W00 stack (kind cluster + monitoring) must be running.

---

## Read
- [ ] **Burns, *Designing Distributed Systems*, 2nd ed., Chapter 2** (Important Distributed System Concepts): read the "Idempotency" and "Orchestration and Kubernetes" sections. Idempotency is the property this unit's Reflect section asks you to observe directly: both operators re-run their reconcile logic constantly, and neither breaks anything by doing so.
- [ ] **Burns, *Designing Distributed Systems*, 2nd ed., Chapter 10** (Ownership Election): read "Determining If You Even Need Leader Election" and "The Basics of Leader Election." Both operators you are about to install run leader election so that only one replica reconciles at a time, and Part 3's etcd work is the same chapter's hands-on. The first section is the useful one: most people reach for leader election before establishing that they need it.
- [ ] [Kubernetes Operators](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/): k8s docs. Read "Motivation" and "Deploying operators". (~10 min)
- [ ] [Kubeflow Trainer overview](https://www.kubeflow.org/docs/components/trainer/overview/): read the architecture section and the description of `TrainJob`, `TrainingRuntime`, and `ClusterTrainingRuntime`. The split is worth understanding before you deploy anything: a runtime is a reusable template describing *how* a kind of training job runs, and a `TrainJob` is a specific request that points at one. (~20 min)
- [ ] [Kubeflow Spark Operator: quick start guide](https://kubeflow.github.io/spark-operator/docs/quick-start-guide.html): read through the `SparkApplication` example. Note the driver/executor split, one long-lived driver Pod coordinating a set of executors, because that internal shape is what you'll be comparing against `TrainJob` later. (~15 min)

**Depth: skim everything.** All four readings are reference documentation for systems you are about to operate, and you will learn more in ten minutes of `kubectl describe` than in an hour of docs. Come back to them when something breaks, which is the only time docs are worth reading closely.

**Key question:** Both controllers are level-triggered reconcilers: they don't react to the delete-a-Pod event directly, they notice "actual state doesn't match desired state" on their next pass and correct it. Predict what happens if you `kubectl delete` one worker Pod of a running `TrainJob` instead of deleting the `TrainJob` itself. Then test it in Part 1 below and see if you were right.

---

## Code

**Part 1: Kubeflow Trainer**

- [ ] Install the Trainer control plane and the default runtimes. Trainer's install is versioned, and the manifests move between releases, so pin a real version rather than copying a tag from here: check the [releases page](https://github.com/kubeflow/trainer/releases) for the current `v2.x.y`, then follow the [installation guide](https://www.kubeflow.org/docs/components/trainer/operator-guides/installation/). It is a two-step install, the manager first and the runtimes second, and skipping the second step is the most common way this goes wrong. The symptom is a `TrainJob` that sits there doing nothing because the runtime it references does not exist.
  ```bash
  # after installing, confirm both the CRDs and the runtimes are present
  kubectl get crds | grep trainer.kubeflow.org
  kubectl get clustertrainingruntimes
  kubectl get pods -n kubeflow-system   # the trainer controller should be Running
  ```
- [ ] `code/operator/config/train-job.yaml`: a small `TrainJob` referencing the built-in `torch-distributed` runtime, with `numNodes: 2` and a trivial training script (a few steps of a tiny model, or even a script that just initializes the process group and prints its rank, which is enough to prove distributed setup worked). Confirm the API group with `kubectl api-resources | grep trainjob` before writing the file, since the exact `apiVersion` depends on the release you installed.
- [ ] Apply it, watch it converge:
  ```bash
  kubectl apply -f code/operator/config/train-job.yaml
  kubectl get trainjob -w
  kubectl get pods -l trainer.kubeflow.org/trainjob-name=<your-job-name>
  kubectl logs <one-of-the-node-pods>   # each node should report a distinct rank
  ```
- [ ] **Break it, on purpose:** `kubectl delete pod <one-of-the-node-pods>` directly, without touching the `TrainJob`. Watch `kubectl get pods -w`. The Pod comes back, because the controller sees a gap between desired and actual and closes it. Now notice the part that matters more: the restarted Pod starts from scratch, with no memory of the training progress the old one had. The operator restarted your *process*; it did nothing about your *state*. That division of labor is the single most important idea in this unit and it's what W16 builds on directly.
- [ ] Skim the real reconciler: browse [`kubeflow/trainer`](https://github.com/kubeflow/trainer) and search for `TrainJobReconciler` (the package path has moved between releases; look under `pkg/controller/` or `internal/controller/` depending on the tag you're viewing). Find the top-level `Reconcile` function. You don't need to read the whole file. Compare its shape, fetch object, compute desired state, diff against actual, patch the difference, against what you predicted for the Key Question.

**Part 2: Spark Operator**

- [ ] Install Kubeflow's Spark Operator via Helm:
  ```bash
  helm repo add spark-operator https://kubeflow.github.io/spark-operator
  helm repo update
  helm install spark-operator spark-operator/spark-operator --namespace spark-operator --create-namespace
  kubectl get pods -n spark-operator   # controller and webhook pods should be Running
  ```
- [ ] `code/operator/config/spark-pi.yaml`: a `SparkApplication` CR running the operator's built-in SparkPi example (`apiVersion: sparkoperator.k8s.io/v1beta2`; copy the example from the quick-start guide you read above rather than hand-writing the full spec, it's long and none of it is new to you after the Spark work in W02, W05, and W07).
- [ ] Apply it, watch it run to completion:
  ```bash
  kubectl apply -f code/operator/config/spark-pi.yaml -n spark-operator
  kubectl get sparkapplication spark-pi -n spark-operator -w   # State: SUBMITTED -> RUNNING -> COMPLETED
  kubectl logs <driver-pod-name> -n spark-operator   # should print an estimate of Pi
  ```
  Treat SparkPi as a smoke test rather than an exercise. Getting it green first means that when your own job fails in a minute, you already know the operator, the cluster, and the RBAC are fine, so the problem is yours. That sequencing is a habit worth having, not a formality.

**Part 2b (optional, stretch): Submit your own job**

> Worth doing, and not on top of Parts 1 and 2 in one sitting. Come back to it; the script-to-image-to-`SparkApplication` path is the piece a data platform role actually exercises.

SparkPi ships inside the Spark image, which is exactly why it always works and teaches you nothing about deployment. Everything that actually goes wrong when a team moves a Spark job onto Kubernetes happens in the gap between "my JAR compiles" and "the driver Pod can find my main class." That gap is this exercise.

Keep the job itself boring on purpose. Twenty lines of aggregation is plenty, because none of the difficulty is in the logic.

- [ ] `code/spark-k8s-job/main.py`: the same PySpark you have been writing since W02. Create a `SparkSession`, build a small DataFrame inline (no external data, nothing to mount), do a `groupBy().agg()`, and print the result. Twenty lines.
- [ ] `code/spark-k8s-job/Dockerfile`: `FROM apache/spark:<matching-version>` and `COPY` your script to `/opt/spark/work-dir/`. **This is where the real lesson lives.** A JVM job ships one self-contained JAR; a Python job ships a script plus an interpreter environment, and if your script imports anything beyond the standard library you now have a dependency-packaging problem that the image must solve. Add one third-party import on purpose and find out what it takes to make it available inside the driver and executor Pods. This is the single most common operational pain in running PySpark on Kubernetes, and it does not exist on the JVM side. Then the same two commands you already know from W00 and W15:
  ```bash
  docker build -t pd-spark-job:latest code/spark-k8s-job
  kind load docker-image pd-spark-job:latest --name pd-systems
  ```
  Matching the image's Spark version to the one you developed against is not optional, and a mismatch here produces a runtime error that reads like a code bug.
- [ ] `code/operator/config/spark-job.yaml`: a `SparkApplication` pointing at your image, with `spec.type: Python`, `spec.pythonVersion: "3"`, and `spec.mainApplicationFile` set to `local:///opt/spark/work-dir/main.py`. The `local://` scheme means "already inside the image," as opposed to a path the driver would have to download at submit time.
- [ ] Apply it and confirm it reaches `COMPLETED`, with your aggregation in the driver logs.
- [ ] **Break it, on purpose:** point `mainApplicationFile` at a path that does not exist in the image, or import a module you never installed, and reapply. The job fails, and the interesting part is where the explanation lives. `kubectl get sparkapplication` shows you a terminal state and nothing useful about why; `kubectl describe` gets you closer; the actual `ModuleNotFoundError` or file-not-found is only in the driver Pod's logs. Walk all three and note the order you would check them next time. This is the single most common Spark-on-Kubernetes failure and the debugging path is not obvious the first time.
- [ ] **Break it, on purpose:** apply a second `SparkApplication` with a deliberately wrong image tag (`image: apache/spark:not-a-real-tag`). Watch it go to `State: FAILED` or get stuck in `ImagePullBackOff`. Find the reason two ways: `kubectl describe sparkapplication` (check `status` and `Events`) and `kubectl logs` on the spark-operator controller Pod itself. This is the actual debugging workflow for a production Spark-on-k8s failure, not a simulation of one.
- [ ] **Break it again, differently:** delete the Spark *driver* Pod of a running application. Compare what happens to what happened when you deleted a `TrainJob` node Pod in Part 1. The two operators make genuinely different promises here, and finding out which one is which by doing it is worth more than reading either project's documentation on the subject.
- [ ] Skim the reconciler: browse [`kubeflow/spark-operator`](https://github.com/kubeflow/spark-operator) for the `SparkApplication` reconcile logic (the exact package path has moved between releases; look under `internal/controller/` or `controllers/` depending on which tag you're viewing, and search for `SparkApplicationReconciler`). You're looking for the same shape you just found in Trainer's: fetch, diff, patch. You don't need to understand the submission-runner mechanics.

**Comparing the two:**

- [ ] In your notes: both `TrainJob` and `SparkApplication` are job-shaped resources, they run to completion and record a terminal state, unlike a CRD describing a cluster you keep standing. But their internal topology differs sharply. A `SparkApplication` has one privileged driver coordinating interchangeable executors, so losing the driver ends the application. A `TrainJob` has a set of peer nodes with no single coordinator, so losing one is recoverable in principle, provided the application code can resume. Write three or four sentences on what that difference implies for how you'd design a job to survive a node failure in each system, and connect it back to what you observed in the two delete tests above.

**Part 3 (optional, stretch): etcd and the Raft Consensus Underneath It All**

> Skip this on a normal pass and come back to it when you have spare time. It is genuinely worth doing and nothing else in the curriculum depends on it, which is exactly why it is the part to drop when the unit is full. The concept is already carried by W13's Raft paper; what you lose by skipping is watching the algorithm run, not understanding what it does.

Neither Part 1 nor Part 2 would work if the Kubernetes API server itself couldn't agree with the rest of the control plane on what's true. That agreement is etcd's job, and etcd stays consistent across replicas using Raft. This curriculum doesn't implement Raft anywhere, a deliberate scope call, but the mechanism is worth watching directly, not just reading about, since it's the actual foundation the first two parts of this unit rest on. If you haven't already read the Raft paper from W13 (Ongaro & Ousterhout, "In Search of an Understandable Consensus Algorithm"), do that first; Section 5 describes exactly the leader-election mechanism you're about to watch happen.

- [ ] Install etcd locally: `brew install etcd` (or download a release binary from [etcd-io/etcd releases](https://github.com/etcd-io/etcd/releases) if you're not on macOS). This gives you the `etcd` server and `etcdctl` client, no cluster build required.
- [ ] Start a 3-member local cluster: run each of the following in its own terminal (or backgrounded with `&`, logs redirected to a file per member). This is the same single-machine, distinct-ports bootstrap etcd's own docs use for local testing, just run by hand instead of via their `Procfile`/`goreman` wrapper:
  ```bash
  # shared across all three
  TOKEN=etcd-cluster-1
  CLUSTER="infra1=http://localhost:2380,infra2=http://localhost:22380,infra3=http://localhost:32380"

  # terminal 1
  etcd --name infra1 --data-dir /tmp/etcd-infra1 \
    --listen-client-urls http://localhost:2379 --advertise-client-urls http://localhost:2379 \
    --listen-peer-urls http://localhost:2380 --initial-advertise-peer-urls http://localhost:2380 \
    --initial-cluster-token $TOKEN --initial-cluster $CLUSTER --initial-cluster-state new

  # terminal 2 (name infra2, data-dir /tmp/etcd-infra2, client 22379, peer 22380)
  # terminal 3 (name infra3, data-dir /tmp/etcd-infra3, client 32379, peer 32380)
  ```
- [ ] Confirm all three joined: `etcdctl --endpoints=localhost:2379,localhost:22379,localhost:32379 member list -w table`.
- [ ] Find the current leader: `etcdctl --endpoints=localhost:2379,localhost:22379,localhost:32379 endpoint status --cluster -w table`. Exactly one row shows `true` under `IS LEADER`; note its `TERM` number and which member (`infra1`/`infra2`/`infra3`) it is.
- [ ] **Kill the leader specifically**, not just any member: Ctrl-C the terminal running that member's `etcd` process. Immediately re-run `endpoint status --cluster` against the two remaining endpoints.
- [ ] Watch a new leader get elected, usually within about a second (Raft's default election timeout), with a strictly higher `TERM` number than before. That term increment is the same idea as W03's Lamport clocks and W13's Chandy-Lamport markers: a monotonically increasing counter used to establish a total order on events, here applied to "who's allowed to lead" instead of to messages or snapshots.
- [ ] Restart the killed member with the identical command you started it with (`--initial-cluster-state new` still works since its data directory already has state; it rejoins as a follower and catches up via replicated log entries). Confirm with `endpoint status --cluster` that it's back and no longer shows `true` under `IS LEADER`, since the member that won the election during its absence keeps the role.
- [ ] Read a slice of the real implementation, not a diagram: [etcd-io/raft](https://github.com/etcd-io/raft), the standalone Raft library that also runs inside Kubernetes' own vendored copy, CockroachDB, and TiKV. Open `raft.go` and find `becomeLeader` and `campaign`; you don't need the whole file, just enough to confirm the shape: a deterministic state machine that takes a `Message` (a timer tick or a peer RPC) as input and emits `{Messages, LogEntries, NextState}` as output. The same "explicit state transition, not implicit control flow" idea W13's sealed-interface `Message` and exhaustive `switch` gave you a small taste of, at production scale.

**Part 4 (optional, stretch): Why Gang Scheduling Exists**

Everything so far assumed your cluster had room. Real clusters don't, and the default behaviour is actively wrong for training jobs.

Kubernetes places Pods one at a time, independently. That's correct for a web service, where three of five replicas running is three-fifths of a working service. It's useless for a distributed training job, where three of five nodes running is *zero* working job: the three that started sit holding expensive hardware, blocked forever on a collective operation waiting for two nodes that never got scheduled. Run two such jobs on a cluster that fits one, and each can end up holding half the hardware and waiting on the other. Neither finishes and neither releases.

**Gang scheduling** is the fix, and it is an old idea rather than a Kubernetes one: Ousterhout named it in 1982, and it is why Slurm, Borg, and every ML scheduler since have some version of all-or-nothing admission. A set of processes that must communicate has to be scheduled together, or none of them progress.

You can see the failure without installing anything, which is the point of doing it this way:

- [ ] Submit two of your Part 1 `TrainJob`s at once, each requesting enough CPU and memory that the two together exceed what your kind cluster can provide. Watch `kubectl get pods`. You should get a partial placement: some Pods `Running`, some `Pending`, and no job able to finish. Leave it long enough to be sure nothing resolves it, because nothing will. This is not a bug in anything. It is what happens when a scheduler that assumes independent Pods meets a workload where they are not.
- [ ] **Your call (written):** the production answer is a queueing layer that admits a job's Pods all together or not at all. On Kubernetes that is Kueue or Volcano; on an HPC cluster it is Slurm; inside Google it was Borg. Read [Kueue's overview](https://kueue.sigs.k8s.io/docs/overview/) for fifteen minutes, then write down what it needs to know that the default scheduler does not, and why that information cannot live in the Pod spec. Installing Kueue and configuring its `ResourceFlavor`, `ClusterQueue`, and `LocalQueue` is a genuinely useful afternoon, but it teaches you Kueue's API rather than the idea, and the idea is what you just watched fail.

**Minimum bar:** a `TrainJob` and a `SparkApplication` both reach a healthy running state on your kind cluster, and you've triggered and diagnosed one real failure in each using `kubectl describe` and `logs` rather than reading about what it would look like. That is Parts 1 and 2. Parts 2b, 3, and 4 are all stretch; pick whichever one interests you most if a second sitting appears.

---

## Reflect


**Prediction versus measurement.** Fill the predictions in *before* you run anything, and do not edit them afterwards. The gap is where calibration comes from.

| Quantity | Predicted | Measured | Which term I got wrong |
|----------|-----------|----------|------------------------|
| | | | |

Copy anything worth carrying into [MEASUREMENTS.md](../MEASUREMENTS.md).

**What "level-triggered" means for a Kubernetes controller, demonstrated concretely by the Pod-delete test in Part 1, not just defined:**

**Where you saw idempotency in practice this unit (a reconcile pass that ran and changed nothing, because nothing needed to change):**

**What happened when you deleted a `TrainJob` node Pod versus a `SparkApplication` driver Pod, and what that says about where each system expects fault tolerance to live:**

**How owner references and garbage collection worked when you deleted each CR entirely, versus what you had to clean up by hand:**

**Where the `ClassNotFoundException` actually turned up, and the order you'd check `get`, `describe`, and driver logs next time:**

**What surprised you about reading a real reconciler's source after only ever reading about the pattern in the abstract:**

**(Part 3 only) What the `TERM` number increasing after you killed the etcd leader actually tells you, and why the member you killed couldn't just declare itself leader again the moment it rejoined:**

**What the un-queued jobs actually looked like when they deadlocked, and how long it took you to be sure nothing was going to resolve it:**

**(Part 4 only) What a queueing layer needs to know that the default scheduler does not, and why that information cannot live in the Pod spec:**

---

## Review and articulate

Two steps that exist because self-study has no examiner. Do them at the end of every unit, before marking it done.

- [ ] **Adversarial review.** Hand over three things separately: the number you predicted, the number you measured, and the conclusion you drew. Then ask for the strongest case that the conclusion is *not* supported by the measurement. Do not ask whether you are right; ask what would falsify this. An assistant asked to check your work will tend to find support for your framing, so the prompt has to be adversarial by construction or the exercise is theatre.
- [ ] **Ninety seconds, out loud, timed.** Explain this unit's finding as you would to someone in an interview or a design review: what you measured, what surprised you, and what decision it would change. Articulation under time pressure is a separate skill from understanding, and it is the one that gets tested. If you cannot do it in ninety seconds you do not have the finding yet, you have notes.
