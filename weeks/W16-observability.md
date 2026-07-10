---
week_number: 16
status: not-started
---

# W16: Observability: Metrics, Tracing, Logging

> **Arc:** Infrastructure · **Language:** Scala + Go

## What you'll build
Instrument your W07 Differential Dataflow engine with Prometheus metrics, OpenTelemetry traces, and structured JSON logs. Deploy it to the kind cluster (W00 stack). Build a Grafana dashboard with four panels that show operator behavior in real time.

**Prerequisites:** W00 (kind + Prometheus + Grafana), W07 (DD engine).

---

## Read
- [ ] [Prometheus data model + metric types](https://prometheus.io/docs/concepts/data_model/): Counter vs Gauge vs Histogram vs Summary. Know when to use each. When to NOT use a Summary. (~20 min)
- [ ] [OpenTelemetry concepts](https://opentelemetry.io/docs/concepts/): read "Signals > Traces" and "Signals > Metrics". Understand what a Span is, what attributes are for, how traces differ from metrics. (~25 min)
- [ ] [Google SRE Book, Chapter 6: Monitoring Distributed Systems](https://sre.google/sre-book/monitoring-distributed-systems/): free online. The four golden signals. (~25 min)

**Key question:** Why are histograms better than averages for latency? What does p99 tell you that avg hides?

---

## Code

**Part 1: Instrument the DD Engine (Scala)**

Add observability to `code/dd-scratch/` (your W07 Differential Dataflow engine).

- [ ] Add to `build.sbt`:
  ```scala
  "io.prometheus" % "simpleclient" % "0.16.0",
  "io.prometheus" % "simpleclient_httpserver" % "0.16.0",
  "io.opentelemetry" % "opentelemetry-api" % "1.36.0",
  "io.opentelemetry" % "opentelemetry-sdk" % "1.36.0"
  ```
- [ ] `metrics/DDMetrics.scala`: define and register:
  - `updates_processed_total`: Counter, incremented once per `Update` processed
  - `consolidation_duration_seconds`: Histogram (buckets: 0.1ms, 1ms, 5ms, 10ms, 50ms), time in `consolidate()`
  - `batch_size`: Histogram (buckets: 1, 10, 100, 1000, 10000), updates per batch
  - `active_keys`: Gauge, current distinct key count in collection
  - Start `HTTPServer` on port 9091 to expose `/metrics`
- [ ] `tracing/DDTracer.scala`: add OTel spans around `map`, `filter`, and `consolidate` in `Collection.scala`. Each span records: operation name, input batch size, output batch size as attributes.
- [ ] `logging/Log.scala`: a small structured logger that writes JSON lines to stdout:
  ```json
  {"level":"INFO","ts":"2026-10-19T10:00:00Z","op":"consolidate","input":1000,"output":42,"duration_ms":3}
  ```
  Replace any `println` in the DD engine with `Log.info(...)`.
- [ ] Run `WordCount.scala` with 10k document updates. Verify `/metrics` at `localhost:9091`.

**Part 2: Grafana Dashboard**

- [ ] Containerize the instrumented DD engine (`Dockerfile`, k8s `Deployment` + `ServiceMonitor`)
- [ ] Deploy to kind: `kind load docker-image dd-engine:latest --name pd-systems && kubectl apply -f k8s/`
- [ ] In Grafana, create a dashboard with 4 panels:
  - `rate(updates_processed_total[1m])`: update throughput (graph)
  - `histogram_quantile(0.99, rate(consolidation_duration_seconds_bucket[5m]))`: p99 consolidation latency (stat)
  - `batch_size` bucket heatmap: shows batch size distribution over time
  - `active_keys`: gauge value over time (graph)
- [ ] Export the dashboard as `config/grafana-dashboard.json` (Grafana → Share → Export)

**Part 3: Go log aggregator (optional stretch)**

- [ ] `tools/log-aggregator/main.go`: HTTP server that accepts structured log lines via `POST /log` (body: JSON) and serves `GET /logs` (last 100 lines, newest first, JSON array). Acts as a sidecar container in your k8s Pod. Use a ring buffer protected by a `sync.RWMutex`. ~80 lines.

---

## Reflect

**What the four golden signals are and which ones your DD engine was "blind" to before this week:**

**What tracing reveals that metrics alone can't (think: which operator is slow for which specific inputs):**

**How you'd extend this instrumentation to W10's distributed training setup:**

**What I'd do differently:**
