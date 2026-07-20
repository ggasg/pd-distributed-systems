---
week_number: 19
status: not-started
---

# W19: Kubernetes Operators

> **Arc:** Infrastructure · **Language:** Go

## What you'll build
A Kubernetes Operator in Go that manages a custom `DistributedJob` resource. When a `DistributedJob` is created, the operator creates worker Pods and a coordinator Service. When it's deleted, GC cleans up everything via owner references. This is the pattern behind Kafka operators, Spark operators, Flink operators, and every managed ML training job on k8s.

The Pod builder also supports an optional sidecar container (`spec.sidecarImage`), the same mechanism Kubeflow's training operator uses to attach log/metric shippers to each worker Pod. You wire the field in now; W20 builds the actual sidecar image and plugs it in.

**Prerequisite:** W00 stack (kind cluster + monitoring) must be running.

---

## Before you start: Go Warm-Up (recommended if goroutines and channels are new or rusty)

20–30 minutes, separate from the operator build below. This is the one week in the entire curriculum that asks you to write Go: everything from W01 through W17 has been Java. `controller-runtime` below is real, idiomatic Go, so rather than meet goroutines and channels for the first time inside an actual reconciler, this drill gets that syntax and the one genuinely new mechanic (channels, not the sequential parts, which will feel familiar from any C-family language) out of the way on a problem simple enough that the channel mechanics are the only thing you're thinking about.

Build a tiny worker pool: N goroutines pull ints off an input channel, double them, and send the result to an output channel.

```go
func doubler(in <-chan int, out chan<- int, wg *sync.WaitGroup) {
    defer wg.Done()
    for n := range in {
        out <- n * 2
    }
}

func main() {
    in, out := make(chan int), make(chan int)
    var wg sync.WaitGroup

    for i := 0; i < 3; i++ { // 3 workers
        wg.Add(1)
        go doubler(in, out, &wg)
    }
    go func() { wg.Wait(); close(out) }() // close AFTER all workers finish

    go func() {
        for i := 1; i <= 10; i++ {
            in <- i
        }
        close(in) // tells workers' `range in` to stop
    }()

    sum := 0
    for n := range out {
        sum += n
    }
    fmt.Println(sum) // 110: doubled 1..10
}
```

Run it, then break it on purpose once: move `close(out)` outside the `wg.Wait()` goroutine and call it directly after the loop that starts workers. Watch it panic ("send on closed channel") or deadlock, depending on timing. That failure mode, closing a channel before everyone writing to it is done, has no equivalent in the Java concurrency you've been using so far (an `ExecutorService` handles this lifecycle for you); it's a genuinely new footgun, worth seeing in eleven lines before you meet it, less legibly, inside `reconciler.go` below.

If this all felt obvious, skip straight to the Read section below; you don't need it. If `close(in)` / `range in` / the two-goroutine close pattern felt new, this was worth the 20 minutes.

---

## Read
- [ ] **Burns, *Designing Distributed Systems*, 2nd ed., Chapter 2** (Important Distributed System Concepts): read the "Idempotency" and "Orchestration and Kubernetes" sections specifically. Idempotency is the one property this week's Reflect section already asks you to justify ("why idempotency is not optional"); Burns gives you the vocabulary and failure examples before you're asked to explain it in your own words, rather than after.
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
      Workers      int32  `json:"workers"`
      Image        string `json:"image"`
      Command      string `json:"command"`
      SidecarImage string `json:"sidecarImage,omitempty"` // optional; wired up in W20
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
  - Fields: `workers` (integer), `image` (string), `command` (string), `sidecarImage` (string, optional)
- [ ] `controllers/reconciler.go`: implement `Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error)`:
  1. Fetch `DistributedJob` by name, return if not found (deleted)
  2. List existing worker Pods labelled `job-name=<name>`
  3. If len(pods) < spec.Workers: create the missing Pods (set owner reference to DistributedJob). Build `pod.Spec.Containers` as a slice: always append the main container (`spec.Image`, `spec.Command`, name `"main"`); if `spec.SidecarImage != ""`, append a second container (name `"sidecar"`) running that image. Both containers share the Pod's network namespace and lifecycle: the main container reaches the sidecar at `localhost:<port>`, and the sidecar terminates when the Pod does.
  4. Count Ready pods, update `status.ReadyWorkers`
  5. If readyWorkers == spec.Workers: set `status.Phase = "Running"`; else `"Pending"`
  6. Patch status subresource
- [ ] `main.go`: set up `ctrl.Manager`, register scheme, start `DistributedJobReconciler` with `ctrl.SetupWithManager`
- [ ] `config/sample.yaml`: a `DistributedJob` with `workers: 3`, image `busybox:latest`, command `sleep 30`. Leave `sidecarImage` unset for now; W20 builds the image and sets it.
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

**Why the sidecar container goes in the same Pod as the main container instead of a separate Pod or Deployment (think: network namespace, lifecycle, scheduling):**

**What you'd add to make this production-ready (finalizers, leader election, webhook validation, metrics):**
