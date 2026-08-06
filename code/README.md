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
├── batch-spark/            # W02: Java (Maven, Spark 4.1.0, local mode)
│   ├── GraphGen.java            # 100k-node random graph to Parquet, fixed seed
│   ├── PageRank.java            # 10 iterations, DataFrame API, cached and uncached runs
│   └── pom.xml
│   # No framework to build. The artifact is the stage DAG and the I/O arithmetic.
│
├── clocks/                 # W03: Go (modules)
│   ├── vector_clock.go
│   ├── message.go
│   ├── node.go
│   ├── failure_detector.go      # heartbeats, lastHeard, SUSPECTED on timeout
│   ├── causal_delivery_test.go
│   └── go.mod
│
├── streaming/               # W04 Part 1: Java (Maven, Flink 2.3.0, MiniCluster)
│   ├── Event.java               # record Event(long eventTime, int value)
│   ├── FlinkWindows.java        # watermark strategy, tumbling window, allowedLateness, side output
│   ├── SparkWindows.java        # same aggregation in Structured Streaming, for the comparison
│   └── pom.xml
│
├── backpressure/            # W04 Part 2: Java (Maven, no framework)
│   ├── Rates.java               # configurable arrival rate and service rate
│   ├── Policy.java              # sealed interface permits Block, Drop, Spill
│   ├── BackpressureBench.java   # all three policies against one overload
│   └── pom.xml
│
├── shuffle/                # W05 Part 1: Java (Maven, standard library only)
│   ├── Record.java              # record Record(String key, int value)
│   ├── Partitioner.java         # sealed interface permits HashPartitioner, RangePartitioner
│   ├── HashPartitioner.java
│   ├── RangePartitioner.java
│   ├── MapTask.java             # writes spill/map-<id>/part-<p>, one file per reduce partition
│   ├── ReduceTask.java          # fetches spill/map-*/part-<p>, aggregates, writes output/part-<p>
│   ├── Shuffle.java             # wires M map tasks to R reduce tasks
│   ├── ShuffleTest.java         # JUnit 5
│   └── pom.xml
│
├── shuffle-skew/           # W05 Part 2: Java (Maven, Spark 4.1.0, local mode)
│   ├── SkewJob.java             # uniform vs Zipf keys; diagnosis happens in the Spark UI
│   └── pom.xml
│
├── query-exec/             # W06: Java (Maven, DuckDB JDBC)
│   ├── RowAtATime.java          # the Volcano-model baseline you write
│   ├── DuckDbRun.java           # same query as SQL; SET threads=1 before measuring
│   └── pom.xml
│
├── query-plans/            # W07: Java (Maven, Spark 4.1.0, local mode)
│   ├── Fixtures.java            # three lopsided Parquet tables + ANALYZE TABLE
│   ├── Explain.java             # EXPLAIN FORMATTED / EXPLAIN COST across four experiments
│   └── pom.xml
│   # Nothing is built here. The deliverable is four plan diffs.
│
├── feature-pipeline/       # W08: Python
│   ├── feature_store.py
│   ├── pipeline.py
│   ├── requirements.txt
│   # Optional: generate_events_large.py, memory_naive.py, memory_chunked.py,
│   # memory_columnar.py (evidence-based memory exercise, no new dependencies)
│
├── distributed-training/   # W09: Python + Go tool
│   ├── mlp.py                   # GIVEN starter file
│   ├── ring_allreduce.py
│   ├── naive_allreduce.py       # send-everything-to-everyone baseline, for the byte comparison
│   ├── compare_allreduce.py     # bytes on the wire per worker at N = 2, 4, 8
│   ├── worker.py
│   ├── train.py
│   └── requirements.txt
│   # Go tool lives in tools/grad_server/
│
├── parallelism/            # W10: Python (imports W09's ring_allreduce directly)
│   ├── layers.py                # GIVEN starter file: Linear + GeLU, NumPy
│   ├── tensor_parallel.py       # ColumnParallelLinear, RowParallelLinear
│   ├── mlp_block.py             # Megatron-style column-then-row composition, one all-reduce
│   ├── pipeline_parallel.py     # 4 layers, 2 stages, activations over sockets
│   ├── microbatch.py
│   ├── bubble.py                # measured vs theoretical (S-1)/(M+S-1)
│   ├── shard_optimizer.py       # ZeRO stage 1: sharded momentum, all-gather params
│   └── requirements.txt
│
├── actor-training/         # W11: Python + Ray
│   ├── model.py             # PyTorch CNN
│   ├── worker_actor.py      # @ray.remote TrainerWorker
│   ├── parameter_server_actor.py  # @ray.remote ParameterServer
│   ├── train.py
│   ├── compare.py           # sequential vs W09 ring-allreduce vs Ray actors
│   └── requirements.txt
│
├── attention/              # W12: Python
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
├── snapshot/                # W13: Java (Maven)
│   ├── Channel.java
│   ├── Message.java
│   ├── Node.java
│   ├── Coordinator.java
│   ├── SnapshotTest.java
│   └── pom.xml
│
├── spark-k8s-job/          # W14 Part 2b: Java (Maven), submitted to the Spark Operator
│   ├── pom.xml              # spark-sql_2.13 with <scope>provided</scope>
│   ├── src/main/java/
│   │   └── Main.java        # deliberately boring: SparkSession + one groupBy/agg
│   ├── Dockerfile           # FROM apache/spark:<version>, COPY the thin JAR in
│   └── README.md            # the mvn package -> docker build -> kind load sequence
│
├── operator/                # W14: deployment configs only, no code built. Kubeflow Trainer, Kubeflow's
│   │                        # and Spark Operator are installed per SETUP.md; you deploy CRs
│   │                        # against them, not author a controller
│   └── config/
│       ├── train-job.yaml     # TrainJob CR (Kubeflow Trainer); W15 edits this to add a sidecar container
│       ├── spark-pi.yaml      # SparkApplication CR, built-in example, used as a smoke test
│       ├── spark-job.yaml     # SparkApplication CR running your own JAR from spark-k8s-job/
│
└── capstone-platform/      # W16 (optional): Python, combines W08+W09+W12+W13+W15, deploys as a TrainJob via W14's Trainer
    ├── train_worker.py
    ├── checkpoint_coordinator.py
    ├── serve.py
    ├── config/
    │   ├── train-job.yaml     # training workers as TrainJob nodes; reuses W14's Trainer install
    │   └── mlflow.yaml        # MLflow Deployment + Service, the registry Part 5 registers into
    └── README.md            # required: design doc
```

---

## Build Commands

**Go (modules):**
```bash
go build ./...
go test ./...
go test -bench=. -benchmem ./...                    # W06's benchmark_test.go files
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

**Python:**
```bash
python -m venv venv && source venv/bin/activate
pip install -r requirements.txt
python main.py
```

**Kubernetes (W14, W16):**
```bash
# Kubeflow Trainer installs from versioned manifests, not a chart repo;
# pin the current release from each project's own installation guide (see SETUP.md)
helm install spark-operator spark-operator/spark-operator --namespace spark-operator --create-namespace
kubectl apply -f code/operator/config/train-job.yaml
kubectl apply -f code/operator/config/spark-pi.yaml -n spark-operator
```
No `go build` specific to W14/W16 themselves: those two weeks deploy CRs against operators you installed, not code you compiled. Go is still built in W00 through W03 and in every secondary tool.

---

## Notes

- Keep each project buildable in isolation. No shared parent build file.
- Code in this directory is the "lab." It's meant to be written, broken, and rewritten.
- The `tools/` directory (at repo root) holds automation scripts that aren't part of a specific week's deliverable
- Go projects (W00 through W03, and the secondary tools in `tools/`) put tests in `<name>_test.go` files sitting directly alongside the code they test, no separate `tests/` directory, that's Go's own toolchain convention (`go test` discovers `_test.go` files automatically in the same package), a different layout convention from the other languages here, not a departure from how the rest of this repo organizes things
- Java projects each have their own `pom.xml`, one Maven project per directory, no shared parent POM, same isolation as every other language here. Source files sit flat in the project root rather than under the conventional `src/main/java/...` package tree; these are small, single-package exercises, and skipping the package hierarchy keeps the file listing above honest about what's actually in each directory. The one exception is `spark-k8s-job/`, which uses the standard Maven layout because the JAR it produces has to be loadable by a Spark image
- **Several projects here build nothing and that is deliberate.** `batch-spark/`, `query-plans/`, and `shuffle-skew/` exist to run a query against a real engine and read what it tells you. Their deliverable is a plan diff, a stage DAG, or a task duration distribution, not a passing test suite
