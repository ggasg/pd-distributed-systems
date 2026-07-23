---
week_number: 19
status: not-started
---

# W19: Operating Kubernetes Operators: KubeRay and Spark Operator

> **Arc:** Infrastructure · **Language:** Helm/YAML; you'll read Go, not write it

## What you'll build

Not build, operate. Deploy two real, production-grade Kubernetes operators to your kind cluster: KubeRay (what Ray, and Anyscale's platform, runs on) and Kubeflow's Spark Operator (what Databricks- and Cloudera-adjacent Spark-on-Kubernetes deployments run on). Create a `RayCluster` and a `SparkApplication`, watch each one reconcile, break each on purpose, and debug it the way you'd debug someone else's operator in production. Then read, don't write, each operator's real reconcile loop: the production version of the level-triggered control loop this week is actually about.

**Why not hand-write a custom operator from scratch, the way this week used to?** Writing your own CRD types and a `controller-runtime` reconciler is the skill set of someone *building* a new operator, which is a narrow platform-infrastructure specialization. It's not what a Staff Data Platform Engineer, a Field/Customer/Professional Services Engineer, or a Developer Advocate does day to day; those roles *operate* operators someone else wrote: install via Helm, configure a CR, read logs, debug a stuck reconcile loop. This week now teaches that instead, using two operators that are directly relevant to the companies this curriculum's job search targets. You've already written Go by now, W00–W04 and W08 all use it, so the reconciler source below reads as familiar syntax applied to a much bigger, real codebase, not a cold start; what this week specifically avoids isn't Go itself, it's the much larger, narrower skill of authoring a `controller-runtime` operator from scratch.

**Prerequisite:** W00 stack (kind cluster + monitoring) must be running.

---

## Read
- [ ] **Burns, *Designing Distributed Systems*, 2nd ed., Chapter 2** (Important Distributed System Concepts): read the "Idempotency" and "Orchestration and Kubernetes" sections. Idempotency is the property this week's Reflect section asks you to observe directly: KubeRay and Spark Operator both re-run their reconcile logic constantly, and neither breaks anything by doing so.
- [ ] [Kubernetes Operators](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/): k8s docs. Read "Motivation" and "Deploying operators". (~10 min)
- [ ] [KubeRay architecture overview](https://ray-project.github.io/kuberay/): read "Introduction" and skim the CRD list (`RayCluster`, `RayJob`, `RayService`). Understand what each of the three is for before you deploy one. (~15 min)
- [ ] [Kubeflow Spark Operator: quick start guide](https://kubeflow.github.io/spark-operator/docs/quick-start-guide.html): read through the `SparkApplication` example. Note that a `SparkApplication` describes a finite, completing job, not a standing cluster, that's the core design difference from `RayCluster` you'll be comparing later. (~15 min)

**Key question:** Both controllers are level-triggered reconcilers: they don't react to the delete-a-Pod event directly, they notice "actual state doesn't match desired state" on their next pass and correct it. Predict what happens if you `kubectl delete` a `RayCluster` worker Pod directly instead of deleting the `RayCluster` itself. Then test it in Part 1 below and see if you were right.

---

## Code

**Part 1: KubeRay**

- [ ] Install the KubeRay operator and its CRDs via Helm:
  ```bash
  helm repo add kuberay https://ray-project.github.io/kuberay-helm/
  helm repo update
  helm install kuberay-operator kuberay/kuberay-operator --version 1.6.0
  kubectl get pods   # kuberay-operator-... should be Running
  ```
- [ ] `code/operator/config/ray-cluster.yaml`: a small `RayCluster` CR, `apiVersion: ray.io/v1`, one head + two workers, minimal resource requests (`cpu: "500m"`, `memory: "1Gi"` is enough for kind):
  ```yaml
  apiVersion: ray.io/v1
  kind: RayCluster
  metadata:
    name: pd-cluster
  spec:
    rayVersion: "2.9.0"
    headGroupSpec:
      rayStartParams: {}
      template:
        spec:
          containers:
            - name: ray-head
              image: rayproject/ray:2.9.0
              resources:
                requests: { cpu: "500m", memory: "1Gi" }
    workerGroupSpecs:
      - groupName: small-group
        replicas: 2
        rayStartParams: {}
        template:
          spec:
            containers:
              - name: ray-worker
                image: rayproject/ray:2.9.0
                resources:
                  requests: { cpu: "500m", memory: "1Gi" }
  ```
- [ ] Apply it, watch it converge:
  ```bash
  kubectl apply -f code/operator/config/ray-cluster.yaml
  kubectl get raycluster pd-cluster -w
  kubectl get pods -l ray.io/cluster=pd-cluster   # expect 1 head + 2 workers
  ```
- [ ] **Break it, on purpose:** `kubectl delete pod <one-of-the-worker-pods>` directly, without touching the `RayCluster`. Watch `kubectl get pods -w`; the operator should notice the mismatch between `spec.workerGroupSpecs[0].replicas: 2` and actual worker count, and recreate the Pod without you doing anything. This is the answer to the Key Question above, and it's the entire value of a reconcile loop demonstrated in one command.
- [ ] Skim the real reconciler: [`raycluster_controller.go`](https://github.com/ray-project/kuberay/blob/master/ray-operator/controllers/ray/raycluster_controller.go) in the `ray-project/kuberay` repo. Find the top-level `Reconcile` function; you don't need to read the whole file. Compare its shape (fetch object, compute desired state, diff against actual, patch the difference) against what you predicted for the Key Question.

**Part 2: Spark Operator**

- [ ] Install Kubeflow's Spark Operator via Helm:
  ```bash
  helm repo add spark-operator https://kubeflow.github.io/spark-operator
  helm repo update
  helm install spark-operator spark-operator/spark-operator --namespace spark-operator --create-namespace
  kubectl get pods -n spark-operator   # controller and webhook pods should be Running
  ```
- [ ] `code/operator/config/spark-pi.yaml`: a `SparkApplication` CR running the operator's built-in SparkPi example (`apiVersion: sparkoperator.k8s.io/v1beta2`; copy the example from the quick-start guide you read above rather than hand-writing the full spec, it's long and none of it is new to you after W12).
- [ ] Apply it, watch it run to completion:
  ```bash
  kubectl apply -f code/operator/config/spark-pi.yaml -n spark-operator
  kubectl get sparkapplication spark-pi -n spark-operator -w   # State: SUBMITTED -> RUNNING -> COMPLETED
  kubectl logs <driver-pod-name> -n spark-operator   # should print an estimate of Pi
  ```
- [ ] **Break it, on purpose:** apply a second `SparkApplication` with a deliberately wrong image tag (`image: apache/spark:not-a-real-tag`). Watch it go to `State: FAILED` or get stuck in `ImagePullBackOff`. Find the reason two ways: `kubectl describe sparkapplication` (check `status` and `Events`) and `kubectl logs` on the spark-operator controller Pod itself. This is the actual debugging workflow for a production Spark-on-k8s failure, not a simulation of one.
- [ ] Skim the reconciler: browse [`kubeflow/spark-operator`](https://github.com/kubeflow/spark-operator) for the `SparkApplication` reconcile logic (the exact package path has moved between releases; look under `internal/controller/` or `controllers/` depending on which tag you're viewing, and search for `SparkApplicationReconciler`). You're looking for the same shape as KubeRay's: fetch, diff, patch. You don't need to understand the submission-runner mechanics, just confirm the control loop pattern is the same one you just watched KubeRay run.

**Comparing the two:**

- [ ] In your notes: `RayCluster` is a long-running resource (the head/worker Pods stay up until you delete the CR); `SparkApplication` is a job-shaped resource (its own Pods terminate on completion, and the CR's `status` records a terminal state). Both use the same reconcile-loop mechanics underneath. Write two or three sentences on why the CRD shape differs even though the control-loop shape doesn't, think about what "desired state" means for a cluster you keep running versus a job you run once.

**Minimum bar:** both `RayCluster` and `SparkApplication` reach a healthy running state on your kind cluster; you've triggered and diagnosed one real failure in each (the deleted worker Pod for KubeRay, the bad image tag for Spark Operator) using `kubectl describe`/`logs`, not by reading about what the failure would look like.

---

## Reflect

**What "level-triggered" means for a Kubernetes controller, demonstrated concretely by the worker-Pod-delete test in Part 1, not just defined:**

**Where you saw idempotency in practice this week (a reconcile pass that ran and changed nothing, because nothing needed to change):**

**How owner references and garbage collection worked when you deleted each CR entirely (`kubectl delete raycluster pd-cluster` / `kubectl delete sparkapplication spark-pi`), versus what you had to clean up by hand:**

**Which of your target companies' platforms you're most likely to actually operate one of these two on, and what that changes about how deep you'd go here if this were your job starting Monday:**

**What surprised you about reading a real reconciler's source after only ever reading about the pattern in the abstract:**
