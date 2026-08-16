---
week_number: 15
status: not-started
---

# W15: Observability: Metrics, Tracing, Logging

> **Arc:** Infrastructure · **Language:** Java + Go
> **Budget:** about 10 hours. The Minimum bar is what a bad week looks like, not the target.

## What you'll build

Instrument your W05 shuffle (Java) with Prometheus metrics, OpenTelemetry traces, and structured JSON logs. Deploy it to the kind cluster (W00 stack). Build a Grafana dashboard with four panels that show operator behavior in real time.

**Scenario:** a dashboard with a confident-looking p99 line is worse than no dashboard at all if the number on it is wrong, because now everyone trusts it. Histogram bucket boundaries are the single easiest way to make that happen silently, and the exercise below shows you exactly how.

**Prerequisites:** W00 (kind + Prometheus + Grafana), W05 (the shuffle).

---

## Read

- [ ] **Burns, *Designing Distributed Systems*, 2nd ed., Chapter 3** (The Sidecar Pattern): read before Part 3 below. You're about to build a log-aggregator sidecar and wire it into a node Pod of the `TrainJob` from W14 without ever naming what you're doing; this chapter names it, and walks through the same "modular container with its own small API, composed alongside a main container it knows nothing about" design your log aggregator follows.
- [ ] **Burns, *Designing Distributed Systems*, 2nd ed., Chapter 14** (Monitoring and Observability Patterns): the chapter this unit is named after. Logging, metrics, basic versus advanced request monitoring, alerting, tracing, and aggregation, in the same order you are about to build them. Read it before Part 1 rather than after, because it is the only reading here that treats all three signals as one system instead of three tools.
- [ ] Optional: **Burns Ch.5** (Adapters). Its two hands-on sections are Prometheus monitoring and normalizing mismatched log formats with fluentd, which is precisely what your log-aggregator sidecar does by hand in Part 3. Worth it if you want to see the pattern named before you implement it.
- [ ] [The Tail at Scale](https://research.google/pubs/the-tail-at-scale/) (Dean & Barroso, CACM 2013): six pages, read all of them, and this is the unit's second study reading. It is the paper that explains why a service made of a hundred healthy components can still be slow for most users, why the 99th percentile of a component becomes the median of a request that fans out, and what the available mitigations are (hedged requests, tied requests, micro-partitioning). Of everything in this curriculum this is the paper most likely to let you answer a hard performance question correctly and immediately, because it gives you the arithmetic of fan-out rather than an intuition about it.
- [ ] [Prometheus data model + metric types](https://prometheus.io/docs/concepts/data_model/): Counter vs Gauge vs Histogram vs Summary. Know when to use each. When to NOT use a Summary. (~20 min)
- [ ] Optional: [OpenTelemetry concepts](https://opentelemetry.io/docs/concepts/): read "Signals > Traces" and "Signals > Metrics". Understand what a Span is, what attributes are for, how traces differ from metrics. (~25 min)
- [ ] **DDIA Chapter 2** (2nd ed.), second pass, **the response-time and percentile sections only**: "Describing Performance," "Latency and Response Time," and the percentile material through tail latency and tail latency amplification. You skimmed the whole chapter in W00 for the vocabulary; this time read those sections at study depth and skip the reliability, scalability, and maintainability material you already have. Now you have a running system, a histogram you are about to configure wrong on purpose, and a p99 that is about to lie to you, which is what those pages are describing.
- [ ] Optional: [Google SRE Book, Chapter 6: Monitoring Distributed Systems](https://sre.google/sre-book/monitoring-distributed-systems/): free online, the four golden signals. Covers much the same ground as DDIA Ch.2, so read it only if you want that material from an operations angle rather than a design one. (~25 min)

**Depth: study The Tail at Scale, plus the named percentile sections of DDIA Ch.2.** Those two are the unit's deep reading. Burns Ch.14 and the Prometheus data model page are short reads. OpenTelemetry concepts, Burns Ch.3, Burns Ch.5, and the SRE chapter are skims.

**Key question:** Why are histograms better than averages for latency? What does p99 tell you that avg hides?

---

## Code

### Part 1: Instrument the shuffle (Java)

Add observability to `code/shuffle/` (your W05 shuffle). Skew is not something you reason about from source, it is something you see in metrics.

- [ ] Add dependencies via Maven (`pom.xml`):
  - [`io.prometheus:prometheus-metrics-core`](https://github.com/prometheus/client_java) (the same Prometheus Java client W00 already uses) plus `prometheus-metrics-exporter-httpserver`, which gives you a standalone metrics HTTP server without wiring `/metrics` into your own `HttpServer` by hand this time. Note the artifact names: the pre-1.0 client used `simpleclient_*`, and a lot of tutorials still show that. The 1.x API is what you want, and it reads slightly differently, `builder()` rather than `build()` and `labelValues()` rather than `labels()`
  - [OpenTelemetry Java SDK](https://opentelemetry.io/docs/languages/java/): `io.opentelemetry:opentelemetry-sdk` plus the OTLP exporter artifact
  - No JSON library needed for the log lines below; five fields is small enough to build by hand

**Before you add a label, know what it costs.** **Cardinality** is the number of distinct time series a metric produces, and it is the product of the distinct values of every label you attach. A counter with one label of ten values is ten series; add a second label of a thousand values and it is ten thousand. Prometheus holds series in memory, so **cardinality explosion**, the usual name for what happens when someone labels a metric by user ID, request ID, or a raw error string, is the single most common way a working Prometheus deployment falls over. The rule that follows: labels are for values you would group by, from a set you can name in advance. `partition` below is a good label because there are as many values as partitions and you chose that number. A record's key would be a catastrophic one.

- [ ] `Metrics.java`: define and register the following against the default `PrometheusRegistry` (the 1.x replacement for the old `CollectorRegistry`):
  - `records_processed_total`: `Counter`, labelled by reduce partition, incremented per record
  - `reduce_task_duration_seconds`: `Histogram` (buckets: 10ms, 50ms, 100ms, 500ms, 2s), wall time per reduce task
  - `spill_file_bytes`: `Histogram`, size of each map-side spill file
  - `active_partitions`: `Gauge`, reduce tasks currently running
  - Start an `HTTPServer` (from `prometheus-metrics-exporter-httpserver`) on port 9091 serving `/metrics`. The Prometheus client ships its own server because exposing the registry is all it has to do.
- [ ] `TracingSetup.java` + `ScopedSpan.java`: initialize an OpenTelemetry `SdkTracerProvider` with an OTLP exporter. Java has no attribute-macro sugar for this, but it has `AutoCloseable` plus try-with-resources. Write `ScopedSpan implements AutoCloseable`: the constructor starts a span, `close()` ends it, and try-with-resources guarantees `close()` runs when the block exits, including on an exception. The span ends when the enclosing block exits, on the success path and the exception path alike:

```java
try (var span = new ScopedSpan(tracer, "consolidate")) {
    // ... consolidate logic ...
    span.setAttribute("input", inputSize);
    span.setAttribute("output", outputSize);
}
```

Wrap `MapTask.run`, `ReduceTask.run`, and the fetch step this way, recording partition id and record count as span attributes. A trace over one job then shows you the straggler directly: every reduce span short except one.

- [ ] `Logging.java`: a small helper that builds one structured JSON log line by hand (`String.format` or a `StringBuilder`, five fields doesn't need a library) and writes it to stdout:

```json
{"level":"INFO","ts":"2026-10-19T10:00:00Z","op":"consolidate","input":1000,"output":42,"duration_ms":3}
```

Replace any `System.out.println` in the shuffle with calls through this helper.

- [ ] Run the W05 Part 1 shuffle over the skewed dataset from W05 Part 2, generated with the same arguments (`--scale 0.5 --skew 1.2 --seed 42`). The exponent matters here for the same reason it mattered there: at a flatter value the duration histogram has one mode instead of two and there is nothing to see. Verify `/metrics` at `localhost:9091`, and confirm `reduce_task_duration_seconds` is visibly bimodal: a cluster of fast tasks and one slow one. That shape *is* the skew, and recognising it on a dashboard is the transferable skill.

### Break it, then decide

- [ ] Temporarily reconfigure `reduce_task_duration_seconds`'s buckets to something wildly mismatched with reality, say `1, 5, 10, 30, 60` (seconds) for tasks that actually complete in tens of milliseconds. Re-run the W05 shuffle over a Zipf-distributed key set and query `histogram_quantile(0.99, rate(reduce_task_duration_seconds_bucket[5m]))` in Prometheus. Every observation lands in the lowest bucket, so the p99 estimate comes back as some number near your smallest boundary regardless of whether real reduce tasks take 20ms or 900ms, the histogram has no resolution in the range that actually matters. Put the original buckets back and confirm the p99 number now moves when you change the workload. Which failure would you rather ship: buckets too coarse in the range that matters (what you just saw), or too many buckets, adding memory and cardinality cost for resolution nobody queries? Say which you'd default to when instrumenting a component you don't yet know the real latency distribution of.

### Part 2: Grafana dashboard

- [ ] Containerize the instrumented shuffle (`Dockerfile`: the same multi-stage `maven:3.9-eclipse-temurin-21` builder → `eclipse-temurin:21-jre-alpine` runtime shape as W00 and W13; k8s `Deployment` + `ServiceMonitor`)
- [ ] Deploy to kind: `kind load docker-image shuffle:latest --name pd-systems && kubectl apply -f k8s/`
- [ ] In Grafana, create a dashboard with 4 panels:
  - `rate(records_processed_total[1m])` by partition: throughput per reduce task, where the skew is immediately visible as one line far above the rest
  - `histogram_quantile(0.99, rate(reduce_task_duration_seconds_bucket[5m]))`: p99 task duration (stat)
  - `reduce_task_duration_seconds` bucket heatmap: the bimodal shape over time
  - `spill_file_bytes`: map-side output size distribution (graph)
- [ ] Export the dashboard as `config/grafana-dashboard.json` (Grafana → Share → Export)

**Minimum bar:** the instrumented shuffle is deployed to kind, Prometheus is scraping it, and the Grafana dashboard's four panels show the skewed run, with the bimodal duration histogram visible. Plus the bucket-mismatch experiment above run and reverted.

### Part 3: Go log aggregator, wired into W14's TrainJob (optional, stretch)

> Parts 1 and 2 are the unit. The sidecar is a satisfying build and a genuinely different pattern, but it is a second sitting, not the same one.

- [ ] `tools/log-aggregator/main.go`: HTTP server (`net/http`, standard library) that accepts structured log lines via `POST /log` (body: JSON) and serves `GET /logs` (last 100 lines, newest first, JSON array). Back it with a fixed-capacity ring buffer guarded by a `sync.RWMutex` rather than a plain `sync.Mutex`: `POST /log` is a rare write and `GET /logs` can be a frequent read, and a plain mutex serializes readers behind each other even though none of them mutate anything.
- [ ] The 100-line cap means anything logging faster than something reads `/logs` silently evicts the oldest lines. This is W04's `Drop` policy again. Decide whether that is acceptable for a debugging aid, or whether `POST /log` should block or reject once the buffer is full, and say what a blocking `POST /log` would cost the trainer container sharing the Pod.
- [ ] `tools/log-aggregator/Dockerfile`: multi-stage build (`golang:1.26` builder → `gcr.io/distroless/static` runtime, the same shape as W00's), `EXPOSE 8080`.
- [ ] Build and load into the kind cluster from W14:

```bash
docker build -t log-aggregator:latest tools/log-aggregator
kind load docker-image log-aggregator:latest --name pd-systems
```

- [ ] Add the sidecar to your W14 `TrainJob`'s node Pod template (`code/operator/config/train-job.yaml`), as a second entry alongside the existing trainer container:

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

Spark Operator's `spec.executor.sidecars` field is the equivalent mechanism on the other operator from W14, worth knowing exists if you want to try the same thing there.

**Minimum bar (Part 3):** the node Pod runs two containers; the trainer container reaches the sidecar over `localhost` with no Service or DNS involved; a log line posted from the trainer round-trips through `GET /logs`.

**If you also stood up Spark Operator in W14:** Kubeflow's Spark Operator supports the same idea natively too, `spec.driver.sidecars` and `spec.executor.sidecars` on a `SparkApplication`. Wiring your log aggregator in there as well is optional, not required for the minimum bar, but worth doing if you want the comparison: a batch job's driver/executor Pods are short-lived, so the sidecar's job there is closer to "capture logs before the Pod disappears" than the standing-cluster case above.

---

## Reflect

**Prediction versus measurement.** Fill the predictions in *before* you run anything, and do not edit them afterwards. The gap is where calibration comes from.

| Quantity | Predicted | Measured | Which term I got wrong |
|----------|-----------|----------|------------------------|
| | | | |

Copy anything worth carrying into [MEASUREMENTS.md](../MEASUREMENTS.md).

**What the four golden signals are and which ones your DD engine was "blind" to before this unit:**

**What your p99 looked like with mismatched histogram buckets, and coarse-buckets vs. too-many-buckets, which would you default to and why (from Break it, then decide above)?**

**Silently drop the oldest log line under a burst, or make `POST /log` block/reject instead, and what would blocking cost the trainer container?**

**What tracing reveals that metrics alone can't (think: which operator is slow for which specific inputs):**

**How you'd extend this instrumentation to W09's distributed training setup:**

**What you'd change to have the DD engine actually ship its JSON log lines to the sidecar over `localhost:8080/log` instead of stdout (the exercise above only proves connectivity via a synthetic curl, not the real log path):**

**What did writing `ScopedSpan` by hand, and leaning on try-with-resources to close it, teach you about span lifetimes that an auto-instrumentation agent would have hidden from you?**

**Why `RWMutex` instead of a plain `Mutex` for the ring buffer, concretely, in terms of `POST /log` vs `GET /logs` traffic, and what would you actually observe under load if you swapped it for a plain `Mutex` (a correctness bug, or something else)?**

**What I'd do differently:**
