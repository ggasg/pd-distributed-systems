# Code Directory

One subdirectory per week. Each is a self-contained project with its own build file.

```
code/
├── hello-metrics/          # W00: Go service (modules) + k8s manifests
│   ├── go.mod
│   ├── main.go
│   ├── Dockerfile
│   └── k8s/
│       ├── deployment.yaml
│       └── service-monitor.yaml
│
├── storage-bench/          # W01: Go (modules)
│   ├── inplace.go           # fixed-slot file, seek + write per record
│   ├── appendlog.go         # append-only file, never seeks
│   ├── bench.go             # 100k records in random order, both paths, timed
│   └── go.mod
├── mapreduce/              # W02: Go (modules)
│   ├── mapreduce.go        # Mapper/Reducer function types
│   ├── runner.go
│   ├── wordcount.go
│   ├── pagerank.go
│   ├── pagerank_runner.go
│   └── go.mod
│   # HTTP coordinator lives in tools/job_coordinator/
│
├── clocks/                 # W03: Go (modules)
│   ├── vector_clock.go
│   ├── message.go
│   ├── node.go
│   ├── failure_detector.go      # heartbeats, lastHeard, SUSPECTED on timeout
│   ├── causal_delivery_test.go
│   └── go.mod
│
├── streaming/               # W04: Java (Maven)
│   ├── Event.java               # Part 1
│   ├── Watermark.java           # Part 1
│   ├── StreamItem.java          # Part 1: sealed interface permits Event, Watermark
│   ├── TumblingWindowAggregator.java   # Part 1
│   ├── StreamProcessor.java     # Part 1
│   ├── BoundedQueue.java        # Part 2: fixed-capacity queue between source and aggregator
│   ├── Source.java              # Part 2: configurable arrival rate
│   ├── SlowAggregator.java      # Part 2: configurable service rate
│   ├── Policy.java              # Part 2: sealed interface permits Block, Drop, Spill
│   ├── BackpressureBench.java   # Part 2: all three policies against one overload
│   ├── StreamProcessorTest.java # JUnit 5
│   └── pom.xml
│
├── shuffle/                # W05: Java (Maven)
│   ├── Record.java              # record Record(String key, int value)
│   ├── Partitioner.java         # sealed interface permits HashPartitioner, RangePartitioner
│   ├── HashPartitioner.java
│   ├── RangePartitioner.java
│   ├── MapTask.java             # writes spill/map-<id>/part-<p>, one file per reduce partition
│   ├── ReduceTask.java          # fetches spill/map-*/part-<p>, aggregates, writes output/part-<p>
│   ├── Shuffle.java             # wires M map tasks to R reduce tasks
│   ├── SkewBench.java           # uniform vs Zipf keys; per-reducer counts and wall time
│   ├── ShuffleTest.java         # JUnit 5
│   └── pom.xml
│
├── dd-scratch/             # W06: Java (Maven)
│   ├── Update.java              # record Update<K, V>  (Part 1, GIVEN)
│   ├── Collection.java          # class Collection<K, V>  (Part 1, GIVEN)
│   ├── WordCount.java           # Part 1, the only Part 1 build target
│   ├── FullRecomputeView.java   # Part 2
│   ├── IncrementalAggregateView.java  # Part 2
│   ├── MvBenchmark.java         # Part 2: incremental vs. full-recompute latency
│   ├── comparisons/                   # Part 2: same orders/region-revenue model, tested against real local OSS systems
│   │   ├── clickhouse_mv.sql          # local ClickHouse server, real materialized view
│   │   └── spark_stateful_agg.py      # local-mode Spark Structured Streaming, stateful aggregation
│   └── pom.xml
│
├── query-exec/             # W07: Go (modules)
│   ├── row_executor.go
│   ├── column_filter.go
│   ├── column_project.go
│   ├── hash_join.go
│   ├── benchmark_test.go       # go test -bench=. -benchmem
│   └── go.mod
│
├── query-planner/          # W08: Scala (sbt)
│   ├── src/main/scala/
│   │   ├── Expr.scala               # Column, Literal, GreaterThan, And
│   │   ├── LogicalPlan.scala        # Scan, Filter, Project, Join
│   │   ├── TreeTransform.scala      # generic transformDown combinator
│   │   ├── rules/
│   │   │   ├── PushDownFilter.scala
│   │   │   └── ConstantFold.scala
│   │   ├── Optimizer.scala          # Part 1: runs rules to a fixed point
│   │   ├── Statistics.scala         # Part 2: TableStats + Catalog
│   │   ├── Cost.scala               # Part 2: estimatedRows + cost (sum of intermediates)
│   │   ├── JoinReorder.scala        # Part 2: enumerate 3-table orderings, cost each
│   │   └── CostBasedOptimizer.scala # Part 2: rules first, then join reordering
│   ├── src/test/scala/
│   │   ├── OptimizerSpec.scala      # ScalaTest or MUnit
│   │   └── JoinReorderSpec.scala
│   ├── aqe_skew_join.py             # Part 2 bridge: PySpark, AQE off vs on, reuses W06's install
│   └── build.sbt
│
├── feature-pipeline/       # W09: Python
│   ├── feature_store.py
│   ├── pipeline.py
│   ├── requirements.txt
│   # Optional: generate_events_large.py, memory_naive.py, memory_chunked.py,
│   # memory_columnar.py (evidence-based memory exercise, no new dependencies)
│
├── distributed-training/   # W10: Python + Go tool
│   ├── mlp.py                   # GIVEN starter file
│   ├── ring_allreduce.py
│   ├── naive_allreduce.py       # send-everything-to-everyone baseline, for the byte comparison
│   ├── compare_allreduce.py     # bytes on the wire per worker at N = 2, 4, 8
│   ├── worker.py
│   ├── train.py
│   └── requirements.txt
│   # Go tool lives in tools/grad_server/
│
├── parallelism/            # W11: Python (imports W10's ring_allreduce directly)
│   ├── layers.py                # GIVEN starter file: Linear + GeLU, NumPy
│   ├── tensor_parallel.py       # ColumnParallelLinear, RowParallelLinear
│   ├── mlp_block.py             # Megatron-style column-then-row composition, one all-reduce
│   ├── pipeline_parallel.py     # 4 layers, 2 stages, activations over sockets
│   ├── microbatch.py
│   ├── bubble.py                # measured vs theoretical (S-1)/(M+S-1)
│   ├── shard_optimizer.py       # ZeRO stage 1: sharded momentum, all-gather params
│   └── requirements.txt
│
├── actor-training/         # W12: Python + Ray
│   ├── model.py             # PyTorch CNN
│   ├── worker_actor.py      # @ray.remote TrainerWorker
│   ├── parameter_server_actor.py  # @ray.remote ParameterServer
│   ├── train.py
│   ├── compare.py           # sequential vs W10 ring-allreduce vs Ray actors
│   └── requirements.txt
│
├── attention/              # W13: Python
│   ├── attention.py        # Part 1: MultiHeadAttention, GIVEN starter file
│   ├── kv_cache.py         # Part 1
│   ├── generate.py         # Part 1: cached vs uncached generation
│   ├── benchmark.py        # Part 1
│   ├── replica.py          # Part 2: model + request-keyed cache, tracks prefill tokens
│   ├── workload.py         # Part 2: interleaved multi-turn conversations
│   ├── router.py           # Part 2: RoundRobinRouter vs CacheAwareRouter
│   ├── bench_routing.py    # Part 2: prefill tokens recomputed, per router
│   └── requirements.txt
│
├── snapshot/                # W14: Java (Maven)
│   ├── Channel.java
│   ├── Message.java
│   ├── Node.java
│   ├── Coordinator.java
│   ├── SnapshotTest.java
│   └── pom.xml
│
├── spark-k8s-job/          # W15 Part 2b: Scala (sbt), submitted to the Spark Operator
│   ├── build.sbt            # Scala 2.13, spark-sql marked "provided"
│   ├── src/main/scala/
│   │   └── Main.scala       # deliberately boring: SparkSession + one groupBy/agg
│   ├── Dockerfile           # FROM apache/spark:<version>, COPY the thin JAR in
│   └── README.md            # the sbt package -> docker build -> kind load sequence
│
├── operator/                # W15: deployment configs only, no code built. Kubeflow Trainer, Kubeflow's
│   │                        # and Spark Operator are installed per SETUP.md; you deploy CRs
│   │                        # against them, not author a controller
│   └── config/
│       ├── train-job.yaml     # TrainJob CR (Kubeflow Trainer); W16 edits this to add a sidecar container
│       ├── spark-pi.yaml      # SparkApplication CR, built-in example, used as a smoke test
│       ├── spark-job.yaml     # SparkApplication CR running your own JAR from spark-k8s-job/
│
├── dd-scratch/             # W16: extends W06 (Java)
│   ├── Metrics.java             # Prometheus Java client
│   ├── TracingSetup.java        # OpenTelemetry Java SDK
│   ├── ScopedSpan.java          # AutoCloseable, try-with-resources span helper
│   └── Logging.java             # hand-rolled structured JSON log lines
│   # Go sidecar lives in tools/log-aggregator/, wired into the W15 TrainJob's node Pod template
│
└── capstone-platform/      # W17 (optional): Python, combines W09+W10+W13+W14+W16, deploys as a TrainJob via W15's Trainer
    ├── train_worker.py
    ├── checkpoint_coordinator.py
    ├── serve.py
    ├── config/
    │   ├── train-job.yaml     # training workers as TrainJob nodes; reuses W15's Trainer install
    │   └── mlflow.yaml        # MLflow Deployment + Service, the registry Part 5 registers into
    └── README.md            # required: design doc
```

---

## Build Commands

**Go (modules):**
```bash
go build ./...
go test ./...
go test -bench=. -benchmem ./...                    # W07's benchmark_test.go files
go run .
go vet ./...                                         # catch common mistakes before they cost you a debugging session
```

**Java (Maven):**
```bash
mvn compile
mvn test
mvn package                                          # produces target/<name>.jar
java -jar target/<name>.jar
java SomeFile.java                                   # single-file source execution, no build step
```

**Scala (sbt):**
```bash
sbt compile
sbt test
sbt run
```

**Python:**
```bash
python -m venv venv && source venv/bin/activate
pip install -r requirements.txt
python main.py
```

**Kubernetes (W15, W17):**
```bash
# Kubeflow Trainer installs from versioned manifests, not a chart repo;
# pin the current release from each project's own installation guide (see SETUP.md)
helm install spark-operator spark-operator/spark-operator --namespace spark-operator --create-namespace
kubectl apply -f code/operator/config/train-job.yaml
kubectl apply -f code/operator/config/spark-pi.yaml -n spark-operator
```
No `go build` specific to W15/W17 themselves: those two weeks deploy CRs against operators you installed, not code you compiled. Go is very much built everywhere else in this curriculum, W00–W03, W07, and every secondary tool.

---

## Notes

- Keep each project buildable in isolation. No shared parent build file.
- Code in this directory is the "lab." It's meant to be written, broken, and rewritten.
- The `tools/` directory (at repo root) holds automation scripts that aren't part of a specific week's deliverable
- Go projects (W00–W03, W07, and the secondary tools in `tools/`) put tests in `<name>_test.go` files sitting directly alongside the code they test, no separate `tests/` directory, that's Go's own toolchain convention (`go test` discovers `_test.go` files automatically in the same package), a fourth layout convention alongside the other languages here, not a departure from how the rest of this repo organizes things
- Scala projects (W08 and W15's `spark-k8s-job`) follow the standard sbt layout (`src/main/scala`, `src/test/scala`), a different convention from Go's in-place tests, that's just how sbt expects things laid out
- Java projects (W04–W06, W14, and the W16 DD-engine instrumentation) each have their own `pom.xml`, one Maven project per directory, no shared parent POM, same isolation as every other language here. Source files sit flat in the project root rather than under the conventional `src/main/java/...` package tree; these are small, single-package exercises, and skipping the package hierarchy keeps the file listing above honest about what's actually in each directory
