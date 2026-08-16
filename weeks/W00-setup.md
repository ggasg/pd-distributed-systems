---
week_number: 0
status: not-started
---

# W00: Infrastructure Setup

> **Budget:** about 7 hours, split roughly 2 hours reading and 5 hours installing. Below a full unit because it is mostly installation rather than thinking, and it is fine if it takes two sittings.

> **Pre-unit:** Complete before W01 begins · **Language:** Go + shell

## What you'll build
A local Kubernetes cluster (kind) with a working observability stack (Prometheus + Grafana), plus one small Go service, `hello-metrics`, running inside it and being scraped. By the end of this unit you can deploy any later code artifact to kind and see its metrics in Grafana. This stack is your running lab. You'll return to it in W14 and W15.

**Scenario:** you've just inherited `hello-metrics` from someone who left the team, and the only handoff note is "it's fine, I think." Nobody, including you, currently has evidence for that claim. Standing up the stack below is what turns "I think it's fine" into something you can actually check.

---

## Read

**Depth: skim everything here, without exception.** This unit is a vocabulary ramp-up, not a study session. Read for the terms and the shape of the arguments, then move on to the install. Do not work through the code examples, do not chase the footnoted papers, and do not stop to reconcile anything against systems you already know. Every idea in these chapters gets built or measured later, and that is when it will stick. Budget about 2 hours for all of it.

- [ ] **DDIA Chapter 1** (2nd ed.), Trade-Offs in Data Systems Architecture. About 45 minutes at skim depth. This gives you the two axes the whole curriculum is arranged along: operational versus analytical, and distributed versus single-node. They are what make W01's write path and W06's columnar executor read as a deliberate contrast rather than two unrelated builds.
- [ ] **DDIA Chapter 2** (2nd ed.), Defining Nonfunctional Requirements. About 45 minutes at skim depth. This is where reliability, scalability, maintainability, response-time percentiles, and tail latency get defined precisely rather than used as adjectives. You are skimming for the definitions only. W15 sends you back to the percentile and tail-latency sections at study depth, once you have a system emitting numbers that the definitions can attach to.
- [ ] Optional: **Burns, *Designing Distributed Systems*, 2nd ed., Chapter 1** (Introduction). Twenty minutes, same job as DDIA Ch.1 and worth reading in the same sitting. It argues that distributed systems are assembled from a small number of recurring patterns, which is the premise the rest of that book cashes out across W02 through W15.
- [ ] [kind docs](https://kind.sigs.k8s.io/) and [kube-prometheus-stack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack): reference material for the install below. Twenty minutes each, tops. You will come back to both rather than absorb them now.

**You should be able to define these afterwards, in one sentence each:** operational versus analytical system, data warehouse, reliability, fault versus failure, scalability, load parameter, response time versus latency, percentile, p99, tail latency, and maintainability. If any of them is still vague, that is fine at this stage; W15 is where the last four have to be exact.

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
  The release name `monitoring` matters beyond this command. Every resource below is named after it, and the `ServiceMonitor` you write later has to carry it as a label. If you pick a different name, substitute it consistently everywhere.
- [ ] Wait for pods: `kubectl get pods -n monitoring -w`
- [ ] Port-forward Grafana: `kubectl port-forward svc/monitoring-grafana 3000:80 -n monitoring`
- [ ] Verify Grafana at `localhost:3000` (admin / admin). Check that the Prometheus datasource is connected.
- [ ] Port-forward Prometheus: `kubectl port-forward svc/monitoring-kube-prometheus-prometheus 9090 -n monitoring`
- [ ] Verify Prometheus at `localhost:9090` by running a query: `up`

---

## Code

Project: `code/hello-metrics/` (Go modules)

A minimal Go HTTP service that exposes Prometheus metrics, deployed to kind. Later units reuse this pattern for their own small services.

**The service has exactly two endpoints:**

| Endpoint | Purpose | Who calls it |
|----------|---------|--------------|
| `GET /` | The application itself. Returns `{"status":"ok"}` and updates the two metrics. | You, with `curl`, to generate traffic |
| `GET /metrics` | Exposes the current value of every registered metric as plain text. | Prometheus, every 15s |

Work through the steps below in order. Each one builds on the previous.

### Step 1: `go.mod`

- [ ] `go mod init hello-metrics`, then `go get github.com/prometheus/client_golang/prometheus` and `go get github.com/prometheus/client_golang/prometheus/promhttp`, the standard Go Prometheus client. No web framework: `net/http`, the standard library's own HTTP server, is enough for two routes.

### Step 2: create and register the two metrics

- [ ] In `main.go`, declare two package-level variables, created exactly once:
  - a Counter named `request_count_total`, via `prometheus.NewCounter`
  - a Histogram named `request_duration_seconds`, via `prometheus.NewHistogram`, with seven bucket boundaries spanning roughly 5ms to 500ms (`0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5`, passed as `prometheus.HistogramOpts.Buckets`)
- [ ] Register both with `prometheus.MustRegister`, in an `init()` function or at the top of `main()`.

Registration is what makes a metric visible at `/metrics`. An unregistered metric still counts correctly in memory and is simply never exposed.

Never construct a Counter or Histogram inside a handler. A fresh object on every request starts at zero, so nothing accumulates, and the failure is silent: no error, no panic, just a `/metrics` endpoint that always reads zero.

### Step 3: the `GET /` handler

- [ ] Response body: `{"status":"ok"}`, written by hand with `fmt.Fprintf` or `w.Write`. A two-key object doesn't need `encoding/json`.
- [ ] The handler does four things, in this order:

  1. `start := time.Now()` as the first line.
  2. `defer func() { requestDuration.Observe(time.Since(start).Seconds()) }()` as the second line. Using `defer` here means the observation still happens if the handler returns early.
  3. `requestCount.Inc()`.
  4. Write the response body.

  `Observe` takes a `float64`. Use `.Seconds()`, not `.Milliseconds()`: [Prometheus naming conventions](https://prometheus.io/docs/practices/naming/#base-units) call for base units, which is what the `_seconds` suffix declares.

### Step 4: the `GET /metrics` handler

- [ ] `http.Handle("/metrics", promhttp.Handler())`. No handler body of your own.

From the [`promhttp` godoc](https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/promhttp#Handler):

> Handler returns an http.Handler for the prometheus.DefaultGatherer, using default HandlerOpts [...] This function is meant to cover the bulk of basic use cases. If you are doing anything that requires more customization (including using a non-default Gatherer, different instrumentation, and non-default HandlerOpts), use the HandlerFor function.

`prometheus.MustRegister` in Step 2 registers into the default registry that `prometheus.DefaultGatherer` reads, so the default case applies here.

**Response**

- Content-Type: `text/plain`, in the [Prometheus text-based exposition format](https://prometheus.io/docs/instrumenting/exposition_formats/)
- Body, filtered to your two metrics:

```
# HELP request_count_total Total number of requests served by GET /.
# TYPE request_count_total counter
request_count_total 42
# HELP request_duration_seconds Wall-clock time spent in the GET / handler.
# TYPE request_duration_seconds histogram
request_duration_seconds_bucket{le="0.005"} 40
request_duration_seconds_bucket{le="0.01"} 41
request_duration_seconds_bucket{le="0.025"} 42
request_duration_seconds_bucket{le="0.05"} 42
request_duration_seconds_bucket{le="0.1"} 42
request_duration_seconds_bucket{le="0.25"} 42
request_duration_seconds_bucket{le="0.5"} 42
request_duration_seconds_bucket{le="+Inf"} 42
request_duration_seconds_sum 0.21
request_duration_seconds_count 42
```

The full response is longer: the default registry also carries the Go runtime and process collectors, so `go_*` and `process_*` series appear alongside yours.

On the ten lines the histogram occupies, from the [Prometheus metric types documentation](https://prometheus.io/docs/concepts/metric_types/#histogram):

> A classic histogram with a base metric name of `<basename>` results in the following time series:
> - cumulative counters for the observation buckets, exposed as `<basename>_bucket{le="<upper inclusive bound>"}`
> - the **total sum** of all observed values, exposed as `<basename>_sum`
> - the **count** of events that have been observed, exposed as `<basename>_count`

### Step 5: `main()`

- [ ] Register both routes and start the server on `:8080` with `http.ListenAndServe`.

### Step 6: `Dockerfile`

- [ ] Multi-stage build:
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

### Step 7: Kubernetes manifests

**The full pipeline, before you write these:** your Go service registers and updates the two metrics, they render as text at `/metrics`, the `ServiceMonitor` tells Prometheus to scrape that text every 15s and store it as a time series, and Grafana panels query Prometheus (never your Go service, never `/metrics` directly) to draw graphs. Nothing "goes into" Grafana; it only reads what Prometheus already collected.

- [ ] `k8s/deployment.yaml`: a `Deployment` (1 replica, image `hello-metrics:latest`, `imagePullPolicy: Never` so kubelet uses the image you side-loaded into kind rather than trying to pull it) plus a `Service`. The Service needs two things the `ServiceMonitor` depends on:
  ```yaml
  apiVersion: v1
  kind: Service
  metadata:
    name: hello-metrics
    labels:
      app: hello-metrics      # the ServiceMonitor selects on this
  spec:
    type: ClusterIP
    selector:
      app: hello-metrics      # this one selects pods, a different job
    ports:
      - name: http            # the ServiceMonitor references this name
        port: 8080
        targetPort: 8080
  ```
  An unnamed port is the most common reason a `ServiceMonitor` produces no target at all.

- [ ] `k8s/service-monitor.yaml`:
  ```yaml
  apiVersion: monitoring.coreos.com/v1
  kind: ServiceMonitor
  metadata:
    name: hello-metrics
    namespace: default
    labels:
      release: monitoring     # must match your Helm release name
  spec:
    selector:
      matchLabels:
        app: hello-metrics    # matches the Service's labels, not the pods'
    namespaceSelector:
      matchNames:
        - default
    endpoints:
      - port: http            # the Service port's *name*, not its number
        path: /metrics
        interval: 15s
  ```

  Each of the following fails silently when wrong, with no error message anywhere:

  - **`labels.release: monitoring`.** kube-prometheus-stack configures its Prometheus to pick up only ServiceMonitors labelled with the Helm release name. Without this label your resource is created successfully, sits in the cluster, and is ignored. Confirm what your Prometheus is actually selecting with:
    ```bash
    kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.serviceMonitorSelector}'
    ```
  - **`spec.selector` matches the Service, not the Deployment.** A `ServiceMonitor` selects Services; the Service selects pods. Pointing the ServiceMonitor at pod labels is a common miss, and it only shows up as a missing target.
  - **`endpoints[].port` is the port name.** Numbers go in `targetPort` if you need them.

### Step 8: build and deploy

- [ ] ```bash
  docker build -t hello-metrics:latest .
  kind load docker-image hello-metrics:latest --name pd-systems
  kubectl apply -f k8s/
  ```

---

## Verify

- [ ] Port-forward the service in one terminal:
  ```bash
  kubectl port-forward svc/hello-metrics 8080:8080
  ```
- [ ] In a second terminal, send 20 requests sequentially. Concurrency is not the point here; you are checking that the counter moves, not measuring throughput.
  ```bash
  for i in $(seq 1 20); do curl -s localhost:8080/ > /dev/null; done
  ```
- [ ] Confirm the target is healthy at `localhost:9090/targets` before querying anything. A missing or `DOWN` target explains every empty result below, and checking it first saves you from debugging a query that was never the problem.
- [ ] Query `request_count_total` in Prometheus. It should read 20, possibly after up to 15 seconds of waiting for the next scrape.
- [ ] Now generate continuous traffic for about two minutes, leaving it running:
  ```bash
  while true; do curl -s localhost:8080/ > /dev/null; sleep 0.2; done
  ```
  `rate()` over a 1m window needs at least two scrapes inside that window to return anything. Twenty requests fired in one second land in a single scrape, so `rate(request_count_total[1m])` will be empty or flat even though the counter is clearly moving. This is not a bug in your setup and it is worth seeing once.
- [ ] With traffic still flowing, query `histogram_quantile(0.95, rate(request_duration_seconds_bucket[1m]))`. Confirm it returns a real number rather than an empty result. An empty result means the `Observe` call is never reached.

  Expect roughly 0.005 or below. A handler that writes fifteen bytes finishes in microseconds, so every observation falls in the first bucket and `histogram_quantile` interpolates inside it. The number you get is a property of your bucket boundaries, not a measurement of your handler. Real latency numbers arrive in W15, when the buckets are chosen against a workload that actually spends time.
- [ ] In Grafana, create two panels: `rate(request_count_total[1m])` and `histogram_quantile(0.95, rate(request_duration_seconds_bucket[1m]))`. Save the dashboard.

**Minimum bar:** the kind cluster is up, `hello-metrics` is deployed to it, Prometheus is scraping its `/metrics`, and one Grafana panel shows a number that moves when you hit the service. The rest of this unit is setup you can finish later; that loop is what every subsequent unit assumes.

---

## Break it, then decide

- [ ] **Break the scrape, and watch the Targets page.** In `k8s/service-monitor.yaml`, replace `port: http` with `targetPort: 8081`, reapply, and watch `localhost:9090/targets`. The target appears and goes `DOWN` with a connection-refused error, because Prometheus resolved an endpoint and found nothing listening on that port. Now revert that and instead break the *name*, `port: httpx`. The target does not go `DOWN`; it disappears from the page entirely, because no endpoint was ever resolved. Restore the working config and confirm the target is healthy again.

  Both are real production failure modes and neither crashes anything. Note which page you would have to be looking at to catch each one, because your Grafana panel looks identical in both cases: empty.
- [ ] **Bucket boundaries.** The buckets above (5ms to 500ms) assume a fast, hot-path endpoint. If `hello-metrics` were instead a batch endpoint you expected to occasionally take 2 to 3 seconds, would you add buckets at the high end, switch to fewer and wider exponential buckets, or leave the existing ones and let the `+Inf` bucket absorb the outliers? Pick one, change `HistogramOpts.Buckets` to match, and be ready to defend the choice below. Note what each option costs: every boundary is an extra time series, stored for every label combination.

---

## Reflect

The prediction versus measurement table starts in W01; this unit installs rather than measures. Record anything from the verification steps that surprised you in [MEASUREMENTS.md](../MEASUREMENTS.md).

**What a ServiceMonitor is and why it exists instead of editing prometheus.yaml directly:**

**What Helm does that raw kubectl apply doesn't:**

**Why the `/metrics` route needs no handler code of your own, and what would change if you used a custom registry:**

**What metrics you'd expose on a real distributed system component (you'll instrument one for real in W15):**
