---
week_number: 0
status: not-started
---

# W00: Infrastructure Setup

> **Pre-week:** Complete before W01 begins · **Language:** Java + shell

## What you'll build
A local Kubernetes cluster (kind) with a working observability stack (Prometheus + Grafana). By end of week, you can deploy any of your weekly code artifacts to kind and see their metrics in Grafana. This stack is your running lab. You'll return to it in W19 and W20.

---

## Read (before anything else)

- [ ] **DDIA Chapter 1**: Reliable, Scalable, and Maintainable Applications. Not tied to this week's build; read it because it's the vocabulary the entire rest of the curriculum assumes. "Reliable," "scalable," and "maintainable" get used loosely everywhere; Kleppmann defines each precisely in about 20 pages, and every later week's design trade-offs (why W07 trims a corner ClickHouse also trims, why W17 cares about a *consistent* snapshot, why W19's reconcile loop is level-triggered) are instances of these three properties in tension. Read once here, then let it recede into the background.

---

## Install

- [ ] Install Docker Desktop (or Podman Desktop)
- [ ] `brew install kind kubectl helm`
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

Project: `code/hello-metrics/` (Java 21, Maven)

A minimal Java HTTP service that exposes Prometheus metrics, deployed to kind. This is the pattern every service you build from here on can follow.

- [ ] `pom.xml`: a single dependency on the current Prometheus Java client (check [the client's GitHub](https://github.com/prometheus/client_java) for the current artifact coordinates and add it to `pom.xml`; the exact group/artifact has changed as the library evolved, worth confirming rather than assuming). No web framework: the HTTP server itself comes from the JDK, not a dependency.
- [ ] `Main.java`: uses `com.sun.net.httpserver.HttpServer` (built into the JDK, `jdk.httpserver` module, no Spring, no external web framework needed for two routes) with two registered contexts, plus two metric objects shared by both routes.

  **Setup: two shared metrics**
  Before either route runs, create two objects once, at startup, using the Prometheus client library:
  - a Counter named `request_count_total`
  - a Histogram named `request_duration_seconds`, with bucket boundaries spanning roughly 5ms to 500ms (seven boundaries: `0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5`)

  Store both as `final` fields, created exactly once, and reuse those same objects on every request. Never construct a new Counter or Histogram inside a handler; a fresh object on every request would reset to zero each time and nothing would ever accumulate. This is the same discipline the Go version of this exercise required, and the same mistake (rebuilding the metric per request) produces the same silent bug in any language: a `/metrics` endpoint that always reads zero.

  **`GET /`**
  - Response: `{"status":"ok"}`
  - Format: JSON, written by hand (a two-key object doesn't need a JSON library). The Counter and Histogram from Setup aren't involved in building this response body.
  - Side effect: increments the shared Counter, and observes its own response time into the shared Histogram (record a start time when the request comes in, observe the elapsed seconds right before responding).

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
  - The Prometheus client library ships a writer that renders its registry in this exact text format; call it from inside your `/metrics` handler rather than building the text yourself. You wire it into your own `HttpServer` route, the same shape as the Go version wiring `promhttp.Handler()` into its own router: the library hands you the response body, you own the route.

**The full pipeline:** your Java code registers and updates the two metrics → they render as text at `/metrics` → the `ServiceMonitor` (below) tells Prometheus to scrape that text every 15s and store it as a time series → Grafana panels (later in this week) query Prometheus (never your Java service, never `/metrics` directly) to draw graphs. Nothing "goes into" Grafana; it only reads what Prometheus already collected.
- [ ] `Dockerfile`: multi-stage build:
  ```dockerfile
  FROM maven:3.9-eclipse-temurin-21 AS builder
  WORKDIR /app
  COPY . .
  RUN mvn -q package

  FROM eclipse-temurin:21-jre-alpine
  COPY --from=builder /app/target/hello-metrics.jar /hello-metrics.jar
  EXPOSE 8080
  CMD ["java", "-jar", "/hello-metrics.jar"]
  ```
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

---

## Reflect

**What a ServiceMonitor is and why it exists instead of editing prometheus.yaml directly:**

**What Helm does that raw kubectl apply doesn't:**

**What metrics you'd expose on a real distributed system component (you'll instrument one for real in W20):**
