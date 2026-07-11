---
week_number: 17
status: not-started
---

# W17: Observability: Metrics, Tracing, Logging

> **Arc:** Infrastructure · **Language:** Rust + Go

## What you'll build
Instrument your W07 Differential Dataflow engine (Rust) with Prometheus metrics, OpenTelemetry traces, and structured JSON logs. Deploy it to the kind cluster (W00 stack). Build a Grafana dashboard with four panels that show operator behavior in real time.

**Prerequisites:** W00 (kind + Prometheus + Grafana), W07 (DD engine).

---

## Read
- [ ] [Prometheus data model + metric types](https://prometheus.io/docs/concepts/data_model/): Counter vs Gauge vs Histogram vs Summary. Know when to use each. When to NOT use a Summary. (~20 min)
- [ ] [OpenTelemetry concepts](https://opentelemetry.io/docs/concepts/): read "Signals > Traces" and "Signals > Metrics". Understand what a Span is, what attributes are for, how traces differ from metrics. (~25 min)
- [ ] [Google SRE Book, Chapter 6: Monitoring Distributed Systems](https://sre.google/sre-book/monitoring-distributed-systems/): free online. The four golden signals. (~25 min)

**Key question:** Why are histograms better than averages for latency? What does p99 tell you that avg hides?

---

## Code

**Part 1: Instrument the DD Engine (Rust)**

Add observability to `code/dd-scratch/` (your W07 Differential Dataflow engine).

- [ ] Add dependencies:
  ```bash
  cd code/dd-scratch
  cargo add prometheus tracing tracing-subscriber tracing-opentelemetry
  cargo add opentelemetry opentelemetry_sdk opentelemetry-otlp
  ```
  `tracing` is the idiomatic choice here: the same `tracing::info!`/`tracing::span!` macros give you both structured logs and spans, and `tracing-opentelemetry` bridges those spans into OTel directly — you're not hand-writing separate metrics and tracing code paths the way the original design implied.
- [ ] `metrics.rs`: define and register the following against a `prometheus::Registry`:
  - `updates_processed_total`: `IntCounter`, incremented once per `Update` processed
  - `consolidation_duration_seconds`: `Histogram` (buckets: 0.1ms, 1ms, 5ms, 10ms, 50ms), time spent in `consolidate()`
  - `batch_size`: `Histogram` (buckets: 1, 10, 100, 1000, 10000), updates per batch
  - `active_keys`: `IntGauge`, current distinct key count in the collection
  - Start an HTTP server on port 9091 exposing `/metrics` (a small handler using `tiny_http` or `hyper` that calls `prometheus::TextEncoder::encode` on scrape)
- [ ] `tracing_setup.rs`: initialize a `tracing_subscriber::Registry` with the `tracing-opentelemetry` layer; annotate `map`, `filter`, and `consolidate` in `collection.rs` with `#[tracing::instrument]`, recording input batch size and output batch size as span fields
- [ ] `logging.rs`: add a JSON-formatting layer to the `tracing_subscriber` registry so every `tracing::info!` call writes a JSON line to stdout:
  ```json
  {"level":"INFO","ts":"2026-10-19T10:00:00Z","op":"consolidate","input":1000,"output":42,"duration_ms":3}
  ```
  Replace any `println!` in the DD engine with `tracing::info!(...)`.
- [ ] Run `word_count.rs` with 10k document updates. Verify `/metrics` at `localhost:9091`.

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
