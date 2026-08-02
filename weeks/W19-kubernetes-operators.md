---
week_number: 19
status: not-started
---

# W19: Operating Kubernetes Operators: Kubeflow Trainer and Spark Operator

> **Arc:** Infrastructure · **Language:** Helm/YAML; you'll read Go, not write it

## What you'll build

Not build, operate. Deploy two real, production-grade Kubernetes operators to your kind cluster: Kubeflow Trainer (the vendor-neutral way to run distributed training jobs on Kubernetes, and a PyTorch Ecosystem project) and Kubeflow's Spark Operator (what Databricks- and Cloudera-adjacent Spark-on-Kubernetes deployments run on). Create a `TrainJob` and a `SparkApplication`, watch each one reconcile, break each on purpose, and debug it the way you'd debug someone else's operator in production. Then read, don't write, each operator's real reconcile loop: the production version of the level-triggered control loop this week is actually about.

Part 3 goes one layer deeper: both operators only work because Kubernetes' own control plane has a consistent view of cluster state, and that consistency comes entirely from etcd running Raft underneath it. You'll stand up a local etcd cluster, kill its leader on purpose, and watch consensus recover in real time. Part 4 goes one layer up instead, to the question of who gets scheduled when there isn't enough hardware for everybody.

**Why not hand-write a custom operator from scratch, the way this week used to?** Writing your own CRD types and a `controller-runtime` reconciler is the skill set of someone *building* a new operator, which is a narrow platform-infrastructure specialization. It's not what a Staff Data Platform Engineer, a Field/Customer/Professional Services Engineer, or a Developer Advocate does day to day; those roles *operate* operators someone else wrote: install via Helm, configure a CR, read logs, debug a stuck reconcile loop. This week teaches that instead. You've already written Go by now, W00 to W04 and W08 all use it, so the reconciler source below reads as familiar syntax applied to a much bigger, real codebase, not a cold start; what this week specifically avoids isn't Go itself, it's the much larger, narrower skill of authoring a `controller-runtime` operator from scratch.

**Why Kubeflow Trainer rather than a framework-specific operator?** Because `TrainJob` is the most portable distributed-training abstraction available. Trainer v2 collapsed the old framework-specific CRDs (`PyTorchJob`, `MPIJob`, `JAXJob`, `XGBoostJob`) into one `TrainJob` plus a pluggable runtime, it's governed by Kubeflow rather than by any single vendor, and it runs the same way on a managed cloud offering as it does on the kind cluster on your laptop. Anything you learn here transfers regardless of which company you end up at, which is exactly what a vendor-specific operator cannot promise.

**A pointer back to W16.** The router you wrote in W16 Part 2, the one that sends a request to the replica already holding its KV cache, is a control-plane component in production, not part of the model server. It runs as a Kubernetes service, it's written in Go, and it needs exactly the kind of per-replica state that the operators below spend their lives tracking. Keep that in mind while reading the reconcilers: the thing you built by hand for two in-process replicas is a small version of what this layer does for a cluster.

**Scenario:** this is what your first week on call for a platform team actually looks like: nothing you're debugging is code you wrote, and the job is reading logs, forming a hypothesis, and checking it, not fixing a bug in your own mental model of a system you built yourself.

**Prerequisite:** W00 stack (kind cluster + monitoring) must be running.

---

## Read
- [ ] **Burns, *Designing Distributed Systems*, 2nd ed., Chapter 2** (Important Distributed System Concepts): read the "Idempotency" and "Orchestration and Kubernetes" sections. Idempotency is the property this week's Reflect section asks you to observe directly: both operators re-run their reconcile logic constantly, and neither breaks anything by doing so.
- [ ] [Kubernetes Operators](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/): k8s docs. Read "Motivation" and "Deploying operators". (~10 min)
- [ ] [Kubeflow Trainer overview](https://www.kubeflow.org/docs/components/trainer/overview/): read the architecture section and the description of `TrainJob`, `TrainingRuntime`, and `ClusterTrainingRuntime`. The split is worth understanding before you deploy anything: a runtime is a reusable template describing *how* a kind of training job runs, and a `TrainJob` is a specific request that points at one. (~20 min)
- [ ] [Kubeflow Spark Operator: quick start guide](https://kubeflow.github.io/spark-operator/docs/quick-start-guide.html): read through the `SparkApplication` example. Note the driver/executor split, one long-lived driver Pod coordinating a set of executors, because that internal shape is what you'll be comparing against `TrainJob` later. (~15 min)

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
- [ ] **Break it, on purpose:** `kubectl delete pod <one-of-the-node-pods>` directly, without touching the `TrainJob`. Watch `kubectl get pods -w`. The Pod comes back, because the controller sees a gap between desired and actual and closes it. Now notice the part that matters more: the restarted Pod starts from scratch, with no memory of the training progress the old one had. The operator restarted your *process*; it did nothing about your *state*. That division of labor is the single most important idea in this week and it's what W21 builds on directly.
- [ ] Skim the real reconciler: browse [`kubeflow/trainer`](https://github.com/kubeflow/trainer) and search for `TrainJobReconciler` (the package path has moved between releases; look under `pkg/controller/` or `internal/controller/` depending on the tag you're viewing). Find the top-level `Reconcile` function. You don't need to read the whole file. Compare its shape, fetch object, compute desired state, diff against actual, patch the difference, against what you predicted for the Key Question.

**Part 2: Spark Operator**

- [ ] Install Kubeflow's Spark Operator via Helm:
  ```bash
  helm repo add spark-operator https://kubeflow.github.io/spark-operator
  helm repo update
  helm install spark-operator spark-operator/spark-operator --namespace spark-operator --create-namespace
  kubectl get pods -n spark-operator   # controller and webhook pods should be Running
  ```
- [ ] `code/operator/config/spark-pi.yaml`: a `SparkApplication` CR running the operator's built-in SparkPi example (`apiVersion: sparkoperator.k8s.io/v1beta2`; copy the example from the quick-start guide you read above rather than hand-writing the full spec, it's long and none of it is new to you after W07's Spark work).
- [ ] Apply it, watch it run to completion:
  ```bash
  kubectl apply -f code/operator/config/spark-pi.yaml -n spark-operator
  kubectl get sparkapplication spark-pi -n spark-operator -w   # State: SUBMITTED -> RUNNING -> COMPLETED
  kubectl logs <driver-pod-name> -n spark-operator   # should print an estimate of Pi
  ```
  Treat SparkPi as a smoke test rather than an exercise. Getting it green first means that when your own job fails in a minute, you already know the operator, the cluster, and the RBAC are fine, so the problem is yours. That sequencing is a habit worth having, not a formality.

**Part 2b: Submit your own Scala job**

SparkPi ships inside the Spark image, which is exactly why it always works and teaches you nothing about deployment. Everything that actually goes wrong when a team moves a Spark job onto Kubernetes happens in the gap between "my JAR compiles" and "the driver Pod can find my main class." That gap is this exercise.

Keep the job itself boring on purpose. Twenty lines of aggregation is plenty, because none of the difficulty is in the logic.

- [ ] `code/spark-k8s-job/`: a minimal sbt project. `build.sbt` targets Scala 2.13 (matching what Spark itself is built against, same reasoning as W09 and W10) and declares `libraryDependencies += "org.apache.spark" %% "spark-sql" % "<version>" % "provided"`. The `provided` matters: Spark is already inside the image, and bundling a second copy is the most common way a first submission fails with a confusing class-loading error. `Main.scala` creates a `SparkSession`, builds a small DataFrame inline (no external data, nothing to mount), does a `groupBy().agg()`, and prints the result.
- [ ] Build a thin JAR with `sbt package`. You do not need `sbt-assembly` here, and reaching for it is the usual overcorrection: an assembly JAR exists to bundle dependencies, and with `provided` you have none to bundle.
- [ ] `code/spark-k8s-job/Dockerfile`: `FROM apache/spark:<matching-version>` and `COPY` your JAR to `/opt/spark/examples/jars/`. Then the same two commands you already know from W00 and W20:
  ```bash
  docker build -t pd-spark-job:latest code/spark-k8s-job
  kind load docker-image pd-spark-job:latest --name pd-systems
  ```
  Matching the image's Spark version to the one your `build.sbt` compiled against is not optional, and a mismatch here produces a runtime error that reads like a code bug.
- [ ] `code/operator/config/spark-job.yaml`: a `SparkApplication` pointing at your image, with `spec.mainClass` set to your fully qualified class name and `spec.mainApplicationFile` set to `local:///opt/spark/examples/jars/<your-jar>.jar`. The `local://` scheme means "already inside the image," as opposed to a path the driver would have to download at submit time.
- [ ] Apply it and confirm it reaches `COMPLETED`, with your aggregation in the driver logs.
- [ ] **Break it, on purpose:** change `mainClass` to something that doesn't exist and reapply. The job fails, and the interesting part is where the explanation lives. `kubectl get sparkapplication` shows you a terminal state and nothing useful about why; `kubectl describe` gets you closer; the actual `ClassNotFoundException` is only in the driver Pod's logs. Walk all three and note the order you'd check them next time. This is the single most common Spark-on-Kubernetes failure and the debugging path is not obvious the first time.
- [ ] **Break it, on purpose:** apply a second `SparkApplication` with a deliberately wrong image tag (`image: apache/spark:not-a-real-tag`). Watch it go to `State: FAILED` or get stuck in `ImagePullBackOff`. Find the reason two ways: `kubectl describe sparkapplication` (check `status` and `Events`) and `kubectl logs` on the spark-operator controller Pod itself. This is the actual debugging workflow for a production Spark-on-k8s failure, not a simulation of one.
- [ ] **Break it again, differently:** delete the Spark *driver* Pod of a running application. Compare what happens to what happened when you deleted a `TrainJob` node Pod in Part 1. The two operators make genuinely different promises here, and finding out which one is which by doing it is worth more than reading either project's documentation on the subject.
- [ ] Skim the reconciler: browse [`kubeflow/spark-operator`](https://github.com/kubeflow/spark-operator) for the `SparkApplication` reconcile logic (the exact package path has moved between releases; look under `internal/controller/` or `controllers/` depending on which tag you're viewing, and search for `SparkApplicationReconciler`). You're looking for the same shape you just found in Trainer's: fetch, diff, patch. You don't need to understand the submission-runner mechanics.

**Comparing the two:**

- [ ] In your notes: both `TrainJob` and `SparkApplication` are job-shaped resources, they run to completion and record a terminal state, unlike a CRD describing a cluster you keep standing. But their internal topology differs sharply. A `SparkApplication` has one privileged driver coordinating interchangeable executors, so losing the driver ends the application. A `TrainJob` has a set of peer nodes with no single coordinator, so losing one is recoverable in principle, provided the application code can resume. Write three or four sentences on what that difference implies for how you'd design a job to survive a node failure in each system, and connect it back to what you observed in the two delete tests above.

**Part 3: etcd and the Raft Consensus Underneath It All**

Neither Part 1 nor Part 2 would work if the Kubernetes API server itself couldn't agree with the rest of the control plane on what's true. That agreement is etcd's job, and etcd stays consistent across replicas using Raft. This curriculum doesn't implement Raft anywhere (W18's Option A deliberately picked primary-backup replication over it, "keep it tractable"), but the mechanism is worth watching directly, not just reading about, since it's the actual foundation the first two parts of this week rest on. If you haven't already read the Raft paper from W17 (Ongaro & Ousterhout, "In Search of an Understandable Consensus Algorithm"), do that first; Section 5 describes exactly the leader-election mechanism you're about to watch happen.

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
- [ ] Watch a new leader get elected, usually within about a second (Raft's default election timeout), with a strictly higher `TERM` number than before. That term increment is the same idea as W04's Lamport clocks and W17's Chandy-Lamport markers: a monotonically increasing counter used to establish a total order on events, here applied to "who's allowed to lead" instead of to messages or snapshots.
- [ ] Restart the killed member with the identical command you started it with (`--initial-cluster-state new` still works since its data directory already has state; it rejoins as a follower and catches up via replicated log entries). Confirm with `endpoint status --cluster` that it's back and no longer shows `true` under `IS LEADER`, since the member that won the election during its absence keeps the role.
- [ ] Read a slice of the real implementation, not a diagram: [etcd-io/raft](https://github.com/etcd-io/raft), the standalone Raft library that also runs inside Kubernetes' own vendored copy, CockroachDB, and TiKV. Open `raft.go` and find `becomeLeader` and `campaign`; you don't need the whole file, just enough to confirm the shape: a deterministic state machine that takes a `Message` (a timer tick or a peer RPC) as input and emits `{Messages, LogEntries, NextState}` as output. The same "explicit state transition, not implicit control flow" idea W17's sealed-interface `Message` and exhaustive `switch` gave you a small taste of, at production scale.

**Part 4: Gang Scheduling with Kueue**

Everything so far assumed your cluster had room. Real clusters don't, and the way Kubernetes handles that by default is actively wrong for training jobs.

Here's why. The default scheduler places Pods one at a time, independently. That's correct for a web service, where three of five replicas running is three-fifths of a working service. It's useless for a distributed training job, where three of five nodes running is zero working job: the three that started sit there holding expensive GPUs, blocked forever on a collective operation waiting for the two nodes that never got scheduled. Run two such jobs at once on a cluster that fits only one, and each can end up holding half the hardware and waiting on the other, neither able to finish or release. That is a genuine deadlock, and it is caused entirely by admitting jobs piecemeal.

**Gang scheduling** is the fix: admit all of a job's Pods or none of them. Kueue is the Kubernetes-native implementation, and Kubeflow Trainer integrates with it directly.

- [ ] Install Kueue by following the [installation guide](https://kueue.sigs.k8s.io/docs/installation/), pinning the current release. Then set up the minimum object graph it needs, which is three resources and worth understanding rather than copying blindly: a `ResourceFlavor` (a description of a class of hardware), a `ClusterQueue` (a quota pool over that flavor, where you'll set a deliberately small CPU and memory budget), and a `LocalQueue` in your namespace pointing at the `ClusterQueue`. The [quickstart for batch users](https://kueue.sigs.k8s.io/docs/tasks/run/jobs/) has a working example of all three.
- [ ] Set the `ClusterQueue` quota so it fits exactly one of your Part 1 `TrainJob`s and no more. Submit two of them, a few seconds apart, each labelled with `kueue.x-k8s.io/queue-name` pointing at your `LocalQueue`.
- [ ] Watch what happens: `kubectl get workloads` and `kubectl get pods`. The first job should be admitted in full and start running. The second should sit in a suspended, unadmitted state with no Pods created at all, and then start on its own once the first finishes. Nothing is half-running at any point. Confirm with `kubectl describe workload <name>` and read the admission events.
- [ ] **Break it, on purpose:** submit the same two jobs *without* the queue label, so Kueue never sees them and the default scheduler handles them directly. Now both get partially admitted, each holding a share of the quota, and neither can complete. Watch the Pods sit in `Pending` while their already-running siblings idle. Leave it in that state long enough to see that nothing resolves it, because nothing will: this is the deadlock described above, and it is not a bug in anything, it's just what happens when a system that assumes independent Pods meets a workload that isn't.
- [ ] **Your call:** your cluster runs both training jobs and Spark jobs, and there is not enough hardware for everything at peak. You have two obvious policies available. Strict priority means the training queue always preempts Spark, which keeps expensive accelerators busy but means an analytics job can be evicted repeatedly and effectively never finish. Fair sharing means each queue gets a guaranteed slice, which bounds everyone's worst case but leaves accelerators idle when the training queue is empty and the Spark queue is backed up. Configure one of them in your `ClusterQueue` (Kueue supports both through cohorts and borrowing), and write down which team is going to complain first under your choice, and what you'd say to them.

**Minimum bar:** both a `TrainJob` and a `SparkApplication` reach a healthy running state on your kind cluster, and one of those `SparkApplication`s runs a Scala JAR you compiled and packaged into an image yourself, not the built-in example; you've triggered and diagnosed one real failure in each using `kubectl describe` and `logs`, not by reading about what the failure would look like; you've killed a 3-member etcd cluster's leader specifically, confirmed via `endpoint status` that a new leader was elected with a higher `TERM`, and rejoined the killed member as a follower; and you've watched two jobs queue cleanly under Kueue and then deadlock without it.

---

## Reflect

**What "level-triggered" means for a Kubernetes controller, demonstrated concretely by the Pod-delete test in Part 1, not just defined:**

**Where you saw idempotency in practice this week (a reconcile pass that ran and changed nothing, because nothing needed to change):**

**What happened when you deleted a `TrainJob` node Pod versus a `SparkApplication` driver Pod, and what that says about where each system expects fault tolerance to live:**

**How owner references and garbage collection worked when you deleted each CR entirely, versus what you had to clean up by hand:**

**Where the `ClassNotFoundException` actually turned up, and the order you'd check `get`, `describe`, and driver logs next time:**

**What surprised you about reading a real reconciler's source after only ever reading about the pattern in the abstract:**

**What the `TERM` number increasing after you killed the etcd leader actually tells you, and why the member you killed couldn't just declare itself leader again the moment it rejoined:**

**What the un-queued jobs actually looked like when they deadlocked, and how long it took you to be sure nothing was going to resolve it:**

**Which scheduling policy you configured in Kueue, and which team complains first:**
