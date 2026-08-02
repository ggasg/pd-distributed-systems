---
week_number: 20
status: not-started
---

# W20: Observability: Metrics, Tracing, Logging

> **Arc:** Infrastructure · **Language:** Java + Go

## What you'll build
Instrument your W07 Differential Dataflow engine (Java) with Prometheus metrics, OpenTelemetry traces, and structured JSON logs. Deploy it to the kind cluster (W00 stack). Build a Grafana dashboard with four panels that show operator behavior in real time.

**Scenario:** a dashboard with a confident-looking p99 line is worse than no dashboard at all if the number on it is wrong, because now everyone trusts it. Histogram bucket boundaries are the single easiest way to make that happen silently, and the exercise below shows you exactly how.

**Prerequisites:** W00 (kind + Prometheus + Grafana), W07 (DD engine).

---

## Read
- [ ] **Burns, *Designing Distributed Systems*, 2nd ed., Chapter 3** (The Sidecar Pattern): read before Part 3 below. You're about to build a log-aggregator sidecar and wire it into a node Pod of the `TrainJob` from W19 without ever naming what you're doing; this chapter names it, and walks through the same "modular container with its own small API, composed alongside a main container it knows nothing about" design your log aggregator follows.
- [ ] [Prometheus data model + metric types](https://prometheus.io/docs/concepts/data_model/): Counter vs Gauge vs Histogram vs Summary. Know when to use each. When to NOT use a Summary. (~20 min)
- [ ] [OpenTelemetry concepts](https://opentelemetry.io/docs/concepts/): read "Signals > Traces" and "Signals > Metrics". Understand what a Span is, what attributes are for, how traces differ from metrics. (~25 min)
- [ ] [Google SRE Book, Chapter 6: Monitoring Distributed Systems](https://sre.google/sre-book/monitoring-distributed-systems/): free online. The four golden signals. (~25 min)

**Key question:** Why are histograms better than averages for latency? What does p99 tell you that avg hides?

---

## Code

**Part 1: Instrument the DD Engine (Java)**

Add observability to `code/dd-scratch/` (your W07 Differential Dataflow engine).

- [ ] Add dependencies via Maven (`pom.xml`):
  - [`io.prometheus:simpleclient`](https://github.com/prometheus/client_java) (the same Prometheus Java client W00 already uses) plus `simpleclient_httpserver`, which gives you a standalone metrics HTTP server without wiring `/metrics` into your own `HttpServer` by hand this time
  - [OpenTelemetry Java SDK](https://opentelemetry.io/docs/languages/java/): `io.opentelemetry:opentelemetry-sdk` plus the OTLP exporter artifact
  - No JSON library needed for the log lines below; five fields is small enough to build by hand, the same call W00 made for its two-key response body
- [ ] `Metrics.java`: define and register the following against the default `CollectorRegistry`:
  - `updates_processed_total`: `Counter`, incremented once per `Update` processed
  - `consolidation_duration_seconds`: `Histogram` (buckets: 0.1ms, 1ms, 5ms, 10ms, 50ms), time spent in `consolidate()`
  - `batch_size`: `Histogram` (buckets: 1, 10, 100, 1000, 10000), updates per batch
  - `active_keys`: `Gauge`, current distinct key count in the collection
  - Start an `HTTPServer` (from `simpleclient_httpserver`) on port 9091 serving `/metrics`; this is a different, purpose-built server from the one your own `HttpServer`-based tools use elsewhere in this curriculum, the Prometheus client ships its own because exposing the registry is all it needs to do.
- [ ] `TracingSetup.java` + `ScopedSpan.java`: initialize an OpenTelemetry `SdkTracerProvider` with an OTLP exporter. Java has no attribute-macro sugar for this either, but it has a direct structural equivalent to the RAII pattern this exercise used to reach for in C++: `AutoCloseable` plus try-with-resources. Write `ScopedSpan implements AutoCloseable`: the constructor starts a span, `close()` ends it, and try-with-resources guarantees `close()` runs when the block exits, including on an exception, the same "closes when the enclosing scope ends" guarantee a C++ destructor gives you, just spelled with a different keyword:
  ```java
  try (var span = new ScopedSpan(tracer, "consolidate")) {
      // ... consolidate logic ...
      span.setAttribute("input", inputSize);
      span.setAttribute("output", outputSize);
  }
  ```
  Wrap `map`, `filter`, and `consolidate` in `Collection.java` this way, recording input batch size and output batch size as span attributes.
- [ ] `Logging.java`: a small helper that builds one structured JSON log line by hand (`String.format` or a `StringBuilder`, five fields doesn't need a library) and writes it to stdout:
  ```json
  {"level":"INFO","ts":"2026-10-19T10:00:00Z","op":"consolidate","input":1000,"output":42,"duration_ms":3}
  ```
  Replace any `System.out.println` in the DD engine with calls through this helper.
- [ ] Run `WordCount.java` with 10k document updates. Verify `/metrics` at `localhost:9091`.

**Break it, then decide:** temporarily reconfigure `consolidation_duration_seconds`'s buckets to something wildly mismatched with reality, say `1, 5, 10, 30, 60` (seconds) for an operation that actually completes in single-digit milliseconds. Re-run `WordCount.java` and query `histogram_quantile(0.99, rate(consolidation_duration_seconds_bucket[5m]))` in Prometheus. Every observation lands in the lowest bucket, so the p99 estimate comes back as some number near your smallest boundary regardless of whether real consolidation calls take 0.1ms or 9ms, the histogram has no resolution in the range that actually matters. Put the original buckets (0.1ms-50ms) back and confirm the p99 number now moves when you change the workload. Which failure would you rather ship: buckets too coarse in the range that matters (what you just saw), or too many buckets, adding memory and cardinality cost for resolution nobody queries? Say which you'd default to when instrumenting a component you don't yet know the real latency distribution of.

**Part 2: Grafana Dashboard**

- [ ] Containerize the instrumented DD engine (`Dockerfile`: the same multi-stage `maven:3.9-eclipse-temurin-21` builder → `eclipse-temurin:21-jre-alpine` runtime shape as W00 and W17; k8s `Deployment` + `ServiceMonitor`)
- [ ] Deploy to kind: `kind load docker-image dd-engine:latest --name pd-systems && kubectl apply -f k8s/`
- [ ] In Grafana, create a dashboard with 4 panels:
  - `rate(updates_processed_total[1m])`: update throughput (graph)
  - `histogram_quantile(0.99, rate(consolidation_duration_seconds_bucket[5m]))`: p99 consolidation latency (stat)
  - `batch_size` bucket heatmap: shows batch size distribution over time
  - `active_keys`: gauge value over time (graph)
- [ ] Export the dashboard as `config/grafana-dashboard.json` (Grafana → Share → Export)

**Part 3: Go log aggregator, wired into W19's TrainJob**

- [ ] `tools/log-aggregator/main.go`: HTTP server (`net/http`, standard library, matching the same "no framework needed for two routes" approach every other small service in this curriculum uses) that accepts structured log lines via `POST /log` (body: JSON) and serves `GET /logs` (last 100 lines, newest first, JSON array). Use a fixed-capacity ring buffer (the same shape as W12's Python DSA Review `RingBuffer`, this time backing a real service) guarded by a `sync.RWMutex`, not a plain `sync.Mutex`. A 100-line cap means a real burst, a tight retry loop, a noisy dependency, anything logging faster than something reads `/logs`, silently evicts the oldest lines before anyone sees them. Decide whether that's acceptable for this sidecar's actual job (a debugging aid, not a durability guarantee) or whether you'd rather have `POST /log` block or reject once the buffer's full instead of silently dropping the oldest entry; whichever you pick, say what it would cost the trainer container if `POST /log` ever blocked on a full buffer. That's not a stylistic preference: `POST /log` is a rare write, `GET /logs` can be a frequent read, and a plain `Mutex` serializes every read behind every other read even though none of them mutate anything; `RWMutex`'s `RLock`/`RUnlock` let concurrent readers through together and only blocks them against the occasional writer. Go's runtime doesn't have the virtual-thread-pinning failure mode a `synchronized` block causes on the JVM, goroutines don't get pinned to an OS thread by holding a lock, but an unnecessarily exclusive lock is still a real, measurable throughput cost under concurrent reads, just a different mechanism than pinning.
- [ ] `tools/log-aggregator/Dockerfile`: multi-stage build (`golang:1.22` builder → `gcr.io/distroless/static-debian12` runtime, the same shape as W00's), `EXPOSE 8080`.
- [ ] Build and load into the kind cluster from W19:
  ```bash
  docker build -t log-aggregator:latest tools/log-aggregator
  kind load docker-image log-aggregator:latest --name pd-systems
  ```
- [ ] Add the sidecar to your W19 `TrainJob`'s node Pod template (`code/operator/config/train-job.yaml`), as a second entry alongside the existing trainer container:
  ```yaml
  - name: log-aggregator
    image: log-aggregator:latest
    ports:
      - containerPort: 8080
  ```
  Kubeflow Trainer builds node Pods from the `TrainingRuntime` the job references, so there are two places this can go and the difference is worth understanding rather than guessing at. Adding it to the `TrainJob` affects only this job. Adding it to a `ClusterTrainingRuntime` affects every job that references that runtime, which is how a platform team would actually ship a logging sidecar to everyone at once. Do it on the `TrainJob` first, since a change you can see the blast radius of is the better thing to learn on.

  Reapply and check:
  ```bash
  kubectl apply -f code/operator/config/train-job.yaml
  kubectl get pod -l trainer.kubeflow.org/trainjob-name=<your-job-name> \
    -o jsonpath='{.items[0].spec.containers[*].name}'
  # expect the trainer container and log-aggregator side by side
  ```
- [ ] Confirm the two containers share a network namespace: the point of the sidecar pattern, not just "two containers exist":
  ```bash
  POD=$(kubectl get pod -l trainer.kubeflow.org/trainjob-name=<your-job-name> -o jsonpath='{.items[0].metadata.name}')
  kubectl exec $POD -c <trainer-container> -- wget -qO- --post-data='{"msg":"hello from the trainer"}' localhost:8080/log
  kubectl exec $POD -c <trainer-container> -- wget -qO- localhost:8080/logs
  # the posted line should come back
  ```
  Spark Operator's `spec.executor.sidecars` field is the equivalent mechanism on the other operator from W19, worth knowing exists if you want to try the same thing there.

**Minimum bar:** the node Pod runs two containers; the trainer container reaches the sidecar over `localhost` with no Service or DNS involved; a log line posted from the trainer round-trips through `GET /logs`.

**If you also stood up Spark Operator in W19:** Kubeflow's Spark Operator supports the same idea natively too, `spec.driver.sidecars` and `spec.executor.sidecars` on a `SparkApplication`. Wiring your log aggregator in there as well is optional, not required for the minimum bar, but worth doing if you want the comparison: a batch job's driver/executor Pods are short-lived, so the sidecar's job there is closer to "capture logs before the Pod disappears" than the standing-cluster case above.

---

## Reflect

**What the four golden signals are and which ones your DD engine was "blind" to before this week:**

**What your p99 looked like with mismatched histogram buckets, and coarse-buckets vs. too-many-buckets, which would you default to and why (from Break it, then decide above)?**

**Silently drop the oldest log line under a burst, or make `POST /log` block/reject instead, and what would blocking cost the trainer container?**

**What tracing reveals that metrics alone can't (think: which operator is slow for which specific inputs):**

**How you'd extend this instrumentation to W12's distributed training setup:**

**What you'd change to have the DD engine actually ship its JSON log lines to the sidecar over `localhost:8080/log` instead of stdout (the exercise above only proves connectivity via a synthetic curl, not the real log path):**

**What did writing `ScopedSpan` by hand, and leaning on try-with-resources to close it, teach you about span lifetimes that an auto-instrumentation agent would have hidden from you?**

**Why `RWMutex` instead of a plain `Mutex` for the ring buffer, concretely, in terms of `POST /log` vs `GET /logs` traffic, and what would you actually observe under load if you swapped it for a plain `Mutex` (a correctness bug, or something else)?**

**What I'd do differently:**
