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
- [ ] **Burns, *Designing Distributed Systems*, 2nd ed., Chapter 3** (The Sidecar Pattern): read before Part 3 below. You're about to build a log-aggregator sidecar and wire it into a worker Pod on the KubeRay cluster from W19 without ever naming what you're doing; this chapter names it, and walks through the same "modular container with its own small API, composed alongside a main container it knows nothing about" design your `LogAggregator.java` follows.
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

**Part 3: Java log aggregator, wired into W19's RayCluster**

- [ ] `tools/log-aggregator/LogAggregator.java`: HTTP server (`com.sun.net.httpserver.HttpServer`, same JDK-only approach as W00 and W03, handlers dispatched on virtual threads the same way W13's gradient server assumes) that accepts structured log lines via `POST /log` (body: JSON) and serves `GET /logs` (last 100 lines, newest first, JSON array). Use a fixed-capacity ring buffer (the same shape as W13's Python DSA Review `RingBuffer`, this time backing a real service) guarded by a `ReentrantReadWriteLock`, not `Collections.synchronizedList`. That's not a stylistic preference: `synchronizedList` guards every call with a `synchronized` block internally, and a `synchronized` block **pins** the calling virtual thread to its carrier (platform) thread for as long as it's held, silently defeating the reason you're using virtual threads at all. Under enough concurrent requests, a pinned handler can starve the carrier pool the same way a blocked platform thread used to, just with a less obvious cause. `ReentrantReadWriteLock` is a `java.util.concurrent` lock, not a `synchronized` block, so it doesn't pin. ~80 lines. The cluster it plugs into is managed by KubeRay, a real operator you didn't write; that boundary is deliberate, this is a language-agnostic sidecar contract, a container image with two HTTP routes, wired in by editing a Pod template the same way any production sidecar gets attached.
- [ ] `tools/log-aggregator/Dockerfile`: multi-stage build (`maven:3.9-eclipse-temurin-21` builder → `eclipse-temurin:21-jre-alpine` runtime, same shape as W00's), `EXPOSE 8080`.
- [ ] Build and load into the kind cluster from W19:
  ```bash
  docker build -t log-aggregator:latest tools/log-aggregator
  kind load docker-image log-aggregator:latest --name pd-systems
  ```
- [ ] Add the sidecar directly to your W19 `RayCluster`'s worker group Pod template (`code/operator/config/ray-cluster.yaml`), a second entry under `workerGroupSpecs[0].template.spec.containers`, alongside the existing `ray-worker` container:
  ```yaml
  - name: log-aggregator
    image: log-aggregator:latest
    ports:
      - containerPort: 8080
  ```
  Reapply and check:
  ```bash
  kubectl apply -f code/operator/config/ray-cluster.yaml
  kubectl get pod -l ray.io/group=small-group -o jsonpath='{.items[0].spec.containers[*].name}'
  # expect: ray-worker log-aggregator
  ```
  This is the same mechanism a real sidecar uses in production, KubeRay itself exposes an equivalent shortcut for exactly this (a `workerSidecarContainers` Helm value that injects a sidecar into every worker Pod cluster-wide instead of one CR at a time), worth knowing exists even though editing the Pod template directly, as you just did, is the better exercise for seeing what's actually happening.
- [ ] Confirm the two containers share a network namespace: the point of the sidecar pattern, not just "two containers exist":
  ```bash
  POD=$(kubectl get pod -l ray.io/group=small-group -o jsonpath='{.items[0].metadata.name}')
  kubectl exec $POD -c ray-worker -- wget -qO- --post-data='{"msg":"hello from the ray worker"}' localhost:8080/log
  kubectl exec $POD -c ray-worker -- wget -qO- localhost:8080/logs
  # the posted line should come back
  ```

**Minimum bar:** the worker Pod runs two containers; the `ray-worker` container reaches the sidecar over `localhost` with no Service or DNS involved; a log line posted from `ray-worker` round-trips through `GET /logs`.

**If you also stood up Spark Operator in W19:** Kubeflow's Spark Operator supports the same idea natively too, `spec.driver.sidecars` and `spec.executor.sidecars` on a `SparkApplication`. Wiring your log aggregator in there as well is optional, not required for the minimum bar, but worth doing if you want the comparison: a batch job's driver/executor Pods are short-lived, so the sidecar's job there is closer to "capture logs before the Pod disappears" than the standing-cluster case above.

---

## Reflect

**What the four golden signals are and which ones your DD engine was "blind" to before this week:**

**What tracing reveals that metrics alone can't (think: which operator is slow for which specific inputs):**

**How you'd extend this instrumentation to W13's distributed training setup:**

**What you'd change to have the DD engine actually ship its JSON log lines to the sidecar over `localhost:8080/log` instead of stdout (the exercise above only proves connectivity via a synthetic curl, not the real log path):**

**How does the `ScopedSpan` RAII pattern compare to Rust's `#[instrument]` macro? What did you lose, and did the C++ version teach you anything about span lifetimes the macro was hiding?**

**What virtual thread pinning is, concretely, in terms of your `LogAggregator`'s ring buffer lock, and why it wouldn't show up as a correctness bug in testing, only as a throughput problem under load:**

**What I'd do differently:**
