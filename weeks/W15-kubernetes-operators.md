---
week_number: 15
status: not-started
---

# W15: Kubernetes Operators

> **Arc:** Infrastructure · **Language:** Go

## What you'll build
A Kubernetes Operator in Go that manages a custom `DistributedJob` resource. When a `DistributedJob` is created, the operator creates worker Pods and a coordinator Service. When it's deleted, GC cleans up everything via owner references. This is the pattern behind Kafka operators, Spark operators, Flink operators, and every managed ML training job on k8s.

**Prerequisite:** W00 stack (kind cluster + monitoring) must be running.

---

## Read
- [ ] [Kubernetes Operators](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/): k8s docs. Read "Motivation" and "Deploying operators". (~10 min)
- [ ] [controller-runtime pkg docs](https://pkg.go.dev/sigs.k8s.io/controller-runtime): focus on `Reconciler` interface and `ctrl.Manager`. (~20 min)
- [ ] [Kubebuilder Book](https://book.kubebuilder.io/), Chapters 1–3: read for concepts, not the `kubebuilder generate` commands. Understand what the reconcile loop does and why it's level-triggered, not edge-triggered. (~45 min)

**Key question:** What does "level-triggered" mean for a Kubernetes controller, and why is it safer than edge-triggered for distributed systems correctness?

---

## Code

Project: `code/operator/` (Go, `sigs.k8s.io/controller-runtime`)

Write it from scratch, no `kubebuilder generate`. Every file is small and intentional.

- [ ] `go.mod`: module `github.com/you/pd-operator`, dependencies: `sigs.k8s.io/controller-runtime`, `k8s.io/api`, `k8s.io/apimachinery`, `k8s.io/client-go`
- [ ] `api/v1/types.go`: define CRD structs:
  ```go
  type DistributedJobSpec struct {
      Workers int32  `json:"workers"`
      Image   string `json:"image"`
      Command string `json:"command"`
  }
  type DistributedJobStatus struct {
      Phase        string `json:"phase"` // Pending | Running | Complete
      ReadyWorkers int32  `json:"readyWorkers"`
  }
  type DistributedJob struct {
      metav1.TypeMeta   `json:",inline"`
      metav1.ObjectMeta `json:"metadata,omitempty"`
      Spec   DistributedJobSpec   `json:"spec,omitempty"`
      Status DistributedJobStatus `json:"status,omitempty"`
  }
  ```
- [ ] `api/v1/register.go`: register the type with the scheme (`SchemeBuilder.Register`)
- [ ] `config/crd.yaml`: hand-write the `CustomResourceDefinition` YAML:
  - group: `pd.systems`, version: `v1`, kind: `DistributedJob`, scope: `Namespaced`
  - Include `spec.versions[].schema.openAPIV3Schema` for basic field validation
- [ ] `controllers/reconciler.go`: implement `Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error)`:
  1. Fetch `DistributedJob` by name, return if not found (deleted)
  2. List existing worker Pods labelled `job-name=<name>`
  3. If len(pods) < spec.Workers: create the missing Pods (set owner reference to DistributedJob)
  4. Count Ready pods, update `status.ReadyWorkers`
  5. If readyWorkers == spec.Workers: set `status.Phase = "Running"`; else `"Pending"`
  6. Patch status subresource
- [ ] `main.go`: set up `ctrl.Manager`, register scheme, start `DistributedJobReconciler` with `ctrl.SetupWithManager`
- [ ] `config/sample.yaml`: a `DistributedJob` with `workers: 3`, image `busybox:latest`, command `sleep 30`
- [ ] Deploy and test:
  ```bash
  kubectl apply -f config/crd.yaml
  go run main.go   # runs controller locally, talks to kind cluster
  # in another terminal:
  kubectl apply -f config/sample.yaml
  kubectl get pods -l job-name=my-job          # should see 3 pods
  kubectl get distributedjob my-job -o yaml    # check status.phase
  kubectl delete distributedjob my-job         # pods should GC via owner refs
  ```

**Minimum bar:** create event creates 3 Pods; delete event GCs the Pods; status reflects ready count.

---

## Reflect

**What "reconcile loop" means and why idempotency is not optional:**

**How owner references enable garbage collection without the controller doing explicit cleanup:**

**How this maps to real systems (Kafka operator, Flink operator, Kubeflow training operator):**

**What you'd add to make this production-ready (finalizers, leader election, webhook validation, metrics):**
