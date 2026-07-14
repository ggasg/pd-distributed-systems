---
week_number: 0
status: not-started
---

# W00: Infrastructure Setup

> **Pre-week:** Complete before W01 begins · **Language:** Go + shell

## What you'll build
A local Kubernetes cluster (kind) with a working observability stack (Prometheus + Grafana). By end of week, you can deploy any of your weekly code artifacts to kind and see their metrics in Grafana. This stack is your running lab. You'll return to it in W18 and W19.

---

## Install

- [ ] Install Docker Desktop (or Podman Desktop)
- [ ] `brew install kind kubectl helm`
- [ ] Create cluster: `kind create cluster --name pd-systems`
- [ ] Verify: `kubectl cluster-info --context kind-pd-systems`

---

## Go Warm-Up (recommended if goroutines and channels are new or rusty)

20–30 minutes, separate from the observability build below. Go's sequential syntax (structs, slices, `if err != nil`) tends to feel familiar fast if you know any C-family language; goroutines and channels are the part that's actually new, and W03 (MapReduce) throws you into them for real, mid-task, with a warning that a small mistake there costs "an hour of confusing debugging instead of teaching you anything." This drill gets that mistake out of the way now, on a problem simple enough that the channel mechanics are the only thing you're thinking about.

Build a tiny worker pool: N goroutines pull ints off an input channel, double them, and send the result to an output channel — the exact fan-out/fan-in shape W03's `runner.go` uses, isolated from any MapReduce logic.

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

Run it, then break it on purpose once: move `close(out)` outside the `wg.Wait()` goroutine and call it directly after the loop that starts workers. Watch it panic ("send on closed channel") or deadlock, depending on timing. That failure mode — closing a channel before everyone writing to it is done — is the one W03 calls out by name; seeing it here, in eleven lines, is cheaper than seeing it there, in the middle of a shuffle phase.

If this all felt obvious, skip straight to the observability stack below — you don't need it. If `close(in)` / `range in` / the two-goroutine close pattern felt new, this was worth the 20 minutes; W01–W04 build on it being second nature.

---

## Deploy the Observability Stack

- [ ] Add Helm repo and install kube-prometheus-stack:
  ```bash
  helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
  helm repo update
  helm install monitoring prometheus-community/kube-prometheus-stack \
    --namespace monitoring --create-namespace \
    --set grafana.adminPassword=admin
  ```
- [ ] Wait for pods: `kubectl get pods -n monitoring -w`
- [ ] Port-forward Grafana: `kubectl port-forward svc/monitoring-grafana 3000:80 -n monitoring`
- [ ] Verify Grafana at `localhost:3000` (admin / admin). Check that the Prometheus datasource is connected.
- [ ] Port-forward Prometheus: `kubectl port-forward svc/monitoring-kube-prometheus-prometheus 9090 -n monitoring`
- [ ] Verify Prometheus at `localhost:9090` by running a query: `up`

---

## Code

Project: `code/hello-metrics/` (Go)

A minimal Go HTTP service that exposes Prometheus metrics, deployed to kind. This is the pattern every service you build from here on can follow.

- [ ] `go.mod`: module `hello-metrics`. Run `go get github.com/prometheus/client_golang` to add the one external dependency this project needs, before writing `main.go`.
- [ ] `main.go`: Go HTTP server with two routes, plus two metric objects shared by both routes.

  **Setup: two shared metrics**
  Before either route runs, create two objects once, at startup, using `github.com/prometheus/client_golang/prometheus/promauto`:
  - a Counter named `request_count_total`, via `promauto.NewCounter(prometheus.CounterOpts{Name: "request_count_total"})`
  - a Histogram named `request_duration_seconds`, via `promauto.NewHistogram(prometheus.HistogramOpts{Name: "request_duration_seconds", Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5}})`, seven boundaries spanning 5ms to 500ms

  `promauto` registers each metric with Prometheus's default registry the instant you create it, so there's no separate registration call to make. Store the two returned objects as package-level variables (or fields on a struct) and reuse those same objects on every request. Never call `promauto.NewCounter` or `promauto.NewHistogram` inside a handler function; a fresh object on every request would reset to zero each time and nothing would ever accumulate.

  **`GET /`**
  - Response: `{"status":"ok"}`
  - Format: JSON, written with the standard library. The Counter and Histogram from Setup aren't involved in building this response body.
  - Side effect: increments the shared Counter, and observes its own response time into the shared Histogram (start a timer when the request comes in, call `.Observe(duration.Seconds())` right before responding). This is why the program needs `github.com/prometheus/client_golang/prometheus` even though the response above is plain JSON: `.Inc()` and `.Observe()` are methods on that package's types.

  **`GET /metrics`**
  - Response: the current value of both metrics created in Setup, meaning `request_count_total` and `request_duration_seconds`
  - Format: Prometheus's plain-text exposition format, not JSON. Example output for those two metrics:
    ```
    # TYPE request_count_total counter
    request_count_total 42
    # TYPE request_duration_seconds histogram
    request_duration_seconds_bucket{le="0.005"} 10
    request_duration_seconds_bucket{le="0.01"} 18
    request_duration_seconds_bucket{le="0.025"} 25
    request_duration_seconds_bucket{le="0.05"} 30
    request_duration_seconds_bucket{le="0.1"} 35
    request_duration_seconds_bucket{le="0.25"} 40
    request_duration_seconds_bucket{le="0.5"} 42
    request_duration_seconds_bucket{le="+Inf"} 42
    request_duration_seconds_sum 3.1
    request_duration_seconds_count 42
    ```
    Notice the Histogram alone takes ten lines: one per bucket boundary from Setup (seven), plus `+Inf`, plus `_sum` and `_count`. There is no single field or single JSON value that represents it.
  - Package: `github.com/prometheus/client_golang/prometheus/promhttp`. Call `promhttp.Handler()` and mount it at `/metrics`; it reads the registry from Setup and writes this text for you. You never construct the response by hand.

**The full pipeline:** your Go code registers and updates the two metrics → they render as text at `/metrics` → the `ServiceMonitor` (below) tells Prometheus to scrape that text every 15s and store it as a time series → Grafana panels (later in this week) query Prometheus (never your Go service, never `/metrics` directly) to draw graphs. Nothing "goes into" Grafana; it only reads what Prometheus already collected.
- [ ] `Dockerfile`: multi-stage build:
  ```dockerfile
  FROM golang:1.22-alpine AS builder
  WORKDIR /app
  COPY . .
  RUN go build -o hello-metrics .

  FROM alpine:latest
  COPY --from=builder /app/hello-metrics /hello-metrics
  EXPOSE 8080
  CMD ["/hello-metrics"]
  ```
- [ ] `k8s/deployment.yaml`: `Deployment` (1 replica, image `hello-metrics:latest`, imagePullPolicy: Never) + `Service` (ClusterIP, port 8080)
- [ ] `k8s/service-monitor.yaml`: `ServiceMonitor` resource (so Prometheus scrapes `/metrics` every 15s)
- [ ] Build and deploy:
  ```bash
  docker build -t hello-metrics:latest .
  kind load docker-image hello-metrics:latest --name pd-systems
  kubectl apply -f k8s/
  ```
- [ ] Verify: port-forward the service, send 20 requests with `curl`, query `request_count_total` in Prometheus, and see the counter. Also query `histogram_quantile(0.95, rate(request_duration_seconds_bucket[1m]))`; confirm it returns a real number, not an empty result. An empty result means `.Observe()` is never actually being called.
- [ ] In Grafana, create two panels: `rate(request_count_total[1m])` and `histogram_quantile(0.95, rate(request_duration_seconds_bucket[1m]))`. Save the dashboard.

---

## Reflect

**What a ServiceMonitor is and why it exists instead of editing prometheus.yaml directly:**

**What Helm does that raw kubectl apply doesn't:**

**What metrics you'd expose on a real distributed system component (you'll instrument one for real in W19):**
