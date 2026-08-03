---
week_number: 0
status: not-started
---

# W00: Infrastructure Setup

> **Budget:** about 5 hours, and it is fine if it takes two sittings. This is the only unit that is mostly installation rather than thinking.

> **Pre-week:** Complete before W01 begins · **Language:** Go + shell

## What you'll build
A local Kubernetes cluster (kind) with a working observability stack (Prometheus + Grafana). By end of week, you can deploy any of your weekly code artifacts to kind and see their metrics in Grafana. This stack is your running lab. You'll return to it in W14 and W15.

**Scenario:** you've just inherited `hello-metrics` from someone who left the team, and the only handoff note is "it's fine, I think." Nobody, including you, currently has evidence for that claim. Standing up the stack below is what turns "I think it's fine" into something you can actually check.

---

## Read

**Depth: skim everything here.** These are reference docs you will come back to, not material to absorb up front. Twenty minutes each, tops.

> **DDIA Chapters 1 and 2 are no longer part of this week.** They used to be, and that was a mistake: they are the least mechanism-dense chapters in the book, nothing in W00 or W01 depends on them, and putting roughly 70 pages of careful reading in front of an infrastructure install guaranteed the first week overran before anything was built. Chapter 1 is now pre-curriculum orientation you read at your own pace, outside any week's budget (see [RESOURCES.md](../RESOURCES.md)). Chapter 2 moved to W15, where you will have a running system to measure and its definitions of reliability and tail latency will have something to attach to.

---

## Install

- [ ] Install Docker Desktop (or Podman Desktop)
- [ ] `brew install kind kubectl helm go`
- [ ] Create cluster: `kind create cluster --name pd-systems`
- [ ] Verify: `kubectl cluster-info --context kind-pd-systems`

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

Project: `code/hello-metrics/` (Go modules)

A minimal Go HTTP service that exposes Prometheus metrics, deployed to kind. This is the pattern every small service you build from here on can follow, this one and every secondary tool in later weeks (W02's job coordinator, W09's gradient server, W14's bench runner, W15's log-aggregator sidecar).

- [ ] `go.mod`: `go mod init hello-metrics`, then `go get github.com/prometheus/client_golang/prometheus` and `go get github.com/prometheus/client_golang/prometheus/promhttp`, the standard Go Prometheus client. No web framework: `net/http`, the standard library's own HTTP server, is enough for two routes.
- [ ] `main.go`: `http.HandleFunc` for two routes, plus two metric objects shared by both handlers.

  **Setup: two shared metrics**
  Before starting the server, create two objects once, at package or `main()` scope, using `prometheus.NewCounter` and `prometheus.NewHistogram`, then register both with `prometheus.MustRegister`:
  - a Counter named `request_count_total`
  - a Histogram named `request_duration_seconds`, with bucket boundaries spanning roughly 5ms to 500ms (seven boundaries: `0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5`, passed to `prometheus.HistogramOpts.Buckets`)

  Keep both as package-level variables, created exactly once, and reuse those same objects in every handler. Never construct a new Counter or Histogram inside a handler; a fresh object on every request would reset to zero each time and nothing would ever accumulate. It's a one-line mistake with a silent failure mode: a `/metrics` endpoint that always reads zero, no error, no panic, just a counter that never counts.

  **`GET /`**
  - Response: `{"status":"ok"}`
  - Format: JSON, written by hand with `fmt.Fprintf` or `w.Write`, a two-key object doesn't need `encoding/json`.
  - Side effect: increments the shared Counter (`counter.Inc()`), and observes its own response time into the shared Histogram (`time.Now()` when the request comes in, `histogram.Observe(time.Since(start).Seconds())` right before responding, or wrap the whole handler body in `defer` to guarantee the observation runs even on an early return).

  **`GET /metrics`**
  - Response: the current value of both metrics created in Setup, `request_count_total` and `request_duration_seconds`, rendered in Prometheus's plain-text exposition format, not JSON. Example output for those two metrics:
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
  - Don't write this text by hand: `promhttp.Handler()` already renders the whole registry in this exact format. Register it directly as your route's handler: `http.Handle("/metrics", promhttp.Handler())`. This is the one route where you're not writing the handler body yourself, the library owns the format because the format is a contract Prometheus's scraper depends on, not something worth re-deriving.

**The full pipeline:** your Go service registers and updates the two metrics → they render as text at `/metrics` → the `ServiceMonitor` (below) tells Prometheus to scrape that text every 15s and store it as a time series → Grafana panels (later in this week) query Prometheus (never your Go service, never `/metrics` directly) to draw graphs. Nothing "goes into" Grafana; it only reads what Prometheus already collected.
- [ ] `Dockerfile`: multi-stage build:
  ```dockerfile
  FROM golang:1.26 AS builder
  WORKDIR /app
  COPY . .
  RUN CGO_ENABLED=0 go build -o hello-metrics .

  FROM gcr.io/distroless/static
  COPY --from=builder /app/hello-metrics /hello-metrics
  EXPOSE 8080
  ENTRYPOINT ["/hello-metrics"]
  ```
  A statically linked Go binary needs nothing at runtime, not even a C library, which is why `distroless/static` works as the final stage: no shell, no package manager, just the binary. This is the smallest, simplest deployment image in the whole curriculum.
- [ ] `k8s/deployment.yaml`: `Deployment` (1 replica, image `hello-metrics:latest`, imagePullPolicy: Never) + `Service` (ClusterIP, port 8080)
- [ ] `k8s/service-monitor.yaml`: `ServiceMonitor` resource (so Prometheus scrapes `/metrics` every 15s)
- [ ] Build and deploy:
  ```bash
  docker build -t hello-metrics:latest .
  kind load docker-image hello-metrics:latest --name pd-systems
  kubectl apply -f k8s/
  ```
- [ ] Verify: port-forward the service, send 20 requests with `curl`, query `request_count_total` in Prometheus, and see the counter. Also query `histogram_quantile(0.95, rate(request_duration_seconds_bucket[1m]))`; confirm it returns a real number, not an empty result. An empty result means the Histogram's observe call is never actually being reached.
- [ ] In Grafana, create two panels: `rate(request_count_total[1m])` and `histogram_quantile(0.95, rate(request_duration_seconds_bucket[1m]))`. Save the dashboard.

**Minimum bar:** the kind cluster is up, `hello-metrics` is deployed to it, Prometheus is scraping its `/metrics`, and one Grafana panel shows a number that moves when you hit the service. The rest of this week is setup you can finish later; that loop is what every subsequent week assumes.

**Break it, then decide:**
- [ ] Point `k8s/service-monitor.yaml` at the wrong port on purpose (`8081` instead of `8080`), reapply, and check Prometheus's Targets page at `localhost:9090/targets`. Confirm it shows the target as `DOWN` with a connection-refused error, not just silently missing from your dashboard. That's the actual failure mode a misconfigured ServiceMonitor produces in production: nothing crashes, the panel just quietly has nothing to show. Fix the port and confirm the target goes healthy again.
- [ ] The histogram buckets given above (5ms-500ms) assume a fast, hot-path endpoint. If `hello-metrics` were instead a batch endpoint you expected to occasionally take 2-3 seconds, would you add buckets at the high end, switch to fewer and wider exponential buckets, or leave the existing ones and let the `+Inf` bucket absorb the outliers? Pick one, change `HistogramOpts.Buckets` to match your answer, and be ready to defend the choice in Reflect below.

---

## Reflect

**What a ServiceMonitor is and why it exists instead of editing prometheus.yaml directly:**

**What Helm does that raw kubectl apply doesn't:**

**What metrics you'd expose on a real distributed system component (you'll instrument one for real in W15):**
