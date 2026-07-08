---
week_number: 0
status: not-started
---

# W00 — Infrastructure Setup

> **Pre-week:** Complete before W01 begins · **Language:** Go + shell

## What you'll build
A local Kubernetes cluster (kind) with a working observability stack (Prometheus + Grafana). By end of week, you can deploy any of your weekly code artifacts to kind and see their metrics in Grafana. This stack is your running lab — you'll return to it in W15 and W16.

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
- [ ] Verify Prometheus at `localhost:9090` — run a query: `up`

---

## Code

Project: `code/hello-metrics/` (Go)

A minimal Go HTTP service that exposes Prometheus metrics, deployed to kind. This is the pattern every service you build from here on can follow.

- [ ] `main.go` — Go HTTP server with two routes:
  - `GET /` → responds `{"status":"ok"}` and increments `request_count_total`
  - `GET /metrics` → served by `promhttp.Handler()`
  - Metrics: `request_count_total` (Counter), `request_duration_seconds` (Histogram with buckets 5ms–500ms)
  - Use `github.com/prometheus/client_golang/prometheus` and `promhttp`
- [ ] `Dockerfile` — multi-stage build:
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
- [ ] `k8s/deployment.yaml` — `Deployment` (1 replica, image `hello-metrics:latest`, imagePullPolicy: Never) + `Service` (ClusterIP, port 8080)
- [ ] `k8s/service-monitor.yaml` — `ServiceMonitor` resource (so Prometheus scrapes `/metrics` every 15s)
- [ ] Build and deploy:
  ```bash
  docker build -t hello-metrics:latest .
  kind load docker-image hello-metrics:latest --name pd-systems
  kubectl apply -f k8s/
  ```
- [ ] Verify: port-forward the service, send 20 requests with `curl`, query `request_count_total` in Prometheus — see the counter.
- [ ] In Grafana, create a panel: `rate(request_count_total[1m])`. Save the dashboard.

---

## Reflect

**What a ServiceMonitor is and why it exists instead of editing prometheus.yaml directly:**

**What Helm does that raw kubectl apply doesn't:**

**What metrics you'd expose on a real distributed system component (you'll instrument one for real in W16):**
