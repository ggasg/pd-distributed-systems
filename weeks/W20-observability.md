---
week_number: 20
status: not-started
---

# W20: Observability: Metrics, Tracing, Logging

> **Arc:** Infrastructure · **Language:** C++ + Java

## What you'll build
Instrument your W07 Differential Dataflow engine (C++) with Prometheus metrics, OpenTelemetry traces, and structured JSON logs. Deploy it to the kind cluster (W00 stack). Build a Grafana dashboard with four panels that show operator behavior in real time.

**Prerequisites:** W00 (kind + Prometheus + Grafana), W07 (DD engine).

---

## Read
- [ ] **Burns, *Designing Distributed Systems*, 2nd ed., Chapter 3** (The Sidecar Pattern): read before Part 3 below. You're about to build a log-aggregator sidecar wired into the W19 operator's `DistributedJob` Pod without ever naming what you're doing; this chapter names it, and walks through the same "modular container with its own small API, composed alongside a main container it knows nothing about" design your `LogAggregator.java` follows.
- [ ] [Prometheus data model + metric types](https://prometheus.io/docs/concepts/data_model/): Counter vs Gauge vs Histogram vs Summary. Know when to use each. When to NOT use a Summary. (~20 min)
- [ ] [OpenTelemetry concepts](https://opentelemetry.io/docs/concepts/): read "Signals > Traces" and "Signals > Metrics". Understand what a Span is, what attributes are for, how traces differ from metrics. (~25 min)
- [ ] [Google SRE Book, Chapter 6: Monitoring Distributed Systems](https://sre.google/sre-book/monitoring-distributed-systems/): free online. The four golden signals. (~25 min)

**Key question:** Why are histograms better than averages for latency? What does p99 tell you that avg hides?

---

## Code

**Part 1: Instrument the DD Engine (C++)**

Add observability to `code/dd-scratch/` (your W07 Differential Dataflow engine).

- [ ] Add dependencies via CMake (`FetchContent` or vcpkg, your call; vcpkg is what's recommended in SETUP.md):
  - [prometheus-cpp](https://github.com/jupp0r/prometheus-cpp): the standard C++ Prometheus client, provides `Registry`, `Counter`, `Histogram`, `Gauge`, and an `Exposer` HTTP server for `/metrics`
  - [opentelemetry-cpp](https://opentelemetry.io/docs/languages/cpp/): official OTel SDK for C++
  - [nlohmann/json](https://github.com/nlohmann/json): header-only, for the structured log lines
- [ ] `include/dd_scratch/metrics.hpp` + `src/metrics.cpp`: define and register the following against a `prometheus::Registry`:
  - `updates_processed_total`: `Counter`, incremented once per `Update` processed
  - `consolidation_duration_seconds`: `Histogram` (buckets: 0.1ms, 1ms, 5ms, 10ms, 50ms), time spent in `consolidate()`
  - `batch_size`: `Histogram` (buckets: 1, 10, 100, 1000, 10000), updates per batch
  - `active_keys`: `Gauge`, current distinct key count in the collection
  - Start a `prometheus::Exposer` on port 9091 serving `/metrics`
- [ ] `include/dd_scratch/tracing_setup.hpp` + `src/tracing_setup.cpp`: initialize an OpenTelemetry `TracerProvider` with an OTLP exporter. Unlike Rust's `#[tracing::instrument]` attribute macro, C++ has no equivalent sugar. Write a small RAII `ScopedSpan` class instead: it starts a span in its constructor and ends it in its destructor, so wrapping a function body in `ScopedSpan span("consolidate");` gets you the same "span closes when the function returns" guarantee the macro gave you in Rust, just spelled out explicitly. Wrap `map`, `filter`, and `consolidate` in `collection.hpp` this way, recording input batch size and output batch size as span attributes.
- [ ] `include/dd_scratch/logging.hpp` + `src/logging.cpp`: a small helper that builds a `nlohmann::json` object per log event and writes it as a single line to stdout:
  ```json
  {"level":"INFO","ts":"2026-10-19T10:00:00Z","op":"consolidate","input":1000,"output":42,"duration_ms":3}
  ```
  Replace any `std::cout <<` in the DD engine with calls through this helper.
- [ ] Run `word_count.cpp` with 10k document updates. Verify `/metrics` at `localhost:9091`.

**Part 2: Grafana Dashboard**

- [ ] Containerize the instrumented DD engine (`Dockerfile`: a multi-stage build compiling with CMake in a builder stage, then copying the binary into a slim runtime image; k8s `Deployment` + `ServiceMonitor`)
- [ ] Deploy to kind: `kind load docker-image dd-engine:latest --name pd-systems && kubectl apply -f k8s/`
- [ ] In Grafana, create a dashboard with 4 panels:
  - `rate(updates_processed_total[1m])`: update throughput (graph)
  - `histogram_quantile(0.99, rate(consolidation_duration_seconds_bucket[5m]))`: p99 consolidation latency (stat)
  - `batch_size` bucket heatmap: shows batch size distribution over time
  - `active_keys`: gauge value over time (graph)
- [ ] Export the dashboard as `config/grafana-dashboard.json` (Grafana → Share → Export)

**Part 3: Java log aggregator, wired into the W19 operator**

- [ ] `tools/log-aggregator/LogAggregator.java`: HTTP server (`com.sun.net.httpserver.HttpServer`, same JDK-only approach as W00 and W03) that accepts structured log lines via `POST /log` (body: JSON) and serves `GET /logs` (last 100 lines, newest first, JSON array). Use a fixed-capacity ring buffer (the same shape as W13's Python DSA Review `RingBuffer`, this time backing a real service) guarded by a `ReentrantReadWriteLock` or wrapped in `Collections.synchronizedList`. ~80 lines. The operator it plugs into is still Go, that boundary is deliberate: this is a language-agnostic sidecar contract, a container image with two HTTP routes, and the `DistributedJob` CRD only ever references it by image name, never by language.
- [ ] `tools/log-aggregator/Dockerfile`: multi-stage build (`maven:3.9-eclipse-temurin-21` builder → `eclipse-temurin:21-jre-alpine` runtime, same shape as W00's), `EXPOSE 8080`.
- [ ] Build and load into the kind cluster from W19:
  ```bash
  docker build -t log-aggregator:latest tools/log-aggregator
  kind load docker-image log-aggregator:latest --name pd-systems
  ```
- [ ] Set `sidecarImage: log-aggregator:latest` on your `DistributedJob` (`code/operator/config/sample.yaml`), then reapply:
  ```bash
  kubectl apply -f code/operator/config/sample.yaml
  kubectl get pod -l job-name=my-job -o jsonpath='{.items[0].spec.containers[*].name}'
  # expect: main sidecar
  ```
- [ ] Confirm the two containers share a network namespace: the point of the sidecar pattern, not just "two containers exist":
  ```bash
  POD=$(kubectl get pod -l job-name=my-job -o jsonpath='{.items[0].metadata.name}')
  kubectl exec $POD -c main -- wget -qO- --post-data='{"msg":"hello from main container"}' localhost:8080/log
  kubectl exec $POD -c main -- wget -qO- localhost:8080/logs
  # the posted line should come back
  ```

**Minimum bar:** the `DistributedJob` Pod runs two containers; the main container reaches the sidecar over `localhost` with no Service or DNS involved; a log line posted from the main container round-trips through `GET /logs`.

---

## Reflect

**What the four golden signals are and which ones your DD engine was "blind" to before this week:**

**What tracing reveals that metrics alone can't (think: which operator is slow for which specific inputs):**

**How you'd extend this instrumentation to W13's distributed training setup:**

**What you'd change to have the DD engine actually ship its JSON log lines to the sidecar over `localhost:8080/log` instead of stdout (the exercise above only proves connectivity via a synthetic curl, not the real log path):**

**How does the `ScopedSpan` RAII pattern compare to Rust's `#[instrument]` macro? What did you lose, and did the C++ version teach you anything about span lifetimes the macro was hiding?**

**What I'd do differently:**
