# Code Directory

One subdirectory per week. Each is a self-contained project with its own build file.

```
code/
├── hello-metrics/          # W00 — Go service + k8s manifests
│   ├── main.go
│   ├── Dockerfile
│   └── k8s/
│       ├── deployment.yaml
│       └── service-monitor.yaml
│
├── lsm/                    # W01 — Java 21
│   ├── src/main/java/
│   │   ├── MemTable.java
│   │   ├── SSTable.java
│   │   ├── LSMTree.java
│   │   └── LSMTreeTest.java
│   └── pom.xml             # or build.gradle
│
├── encoding/               # W02 — Java 21
│   ├── src/main/java/
│   │   ├── Varint.java
│   │   ├── RowStore.java
│   │   ├── ColumnStore.java
│   │   └── Benchmark.java
│   └── pom.xml
│
├── mapreduce/              # W03 — Java 21 + Go tool
│   ├── src/main/java/
│   │   ├── MapReduceJob.java
│   │   ├── MapReduceRunner.java
│   │   ├── WordCount.java
│   │   └── PageRank.java
│   └── pom.xml
│   # Go tool lives in tools/job_coordinator/
│
├── clocks/                 # W04 — Java 21
│   ├── src/main/java/
│   │   ├── VectorClock.java
│   │   ├── Message.java
│   │   ├── Node.java
│   │   └── CausalDeliveryTest.java
│   └── pom.xml
│
├── streaming/              # W05 — Scala 2.13
│   ├── src/main/scala/
│   │   ├── Event.scala
│   │   ├── Watermark.scala
│   │   ├── TumblingWindowAggregator.scala
│   │   └── StreamProcessor.scala
│   └── build.sbt
│
├── timely-toy/             # W06 — Scala 2.13
│   ├── src/main/scala/
│   │   ├── Timestamp.scala
│   │   ├── Pointstamp.scala
│   │   ├── Operator.scala
│   │   ├── ProgressTracker.scala
│   │   └── DataflowTest.scala
│   └── build.sbt
│
├── dd-scratch/             # W07 — Scala 2.13
│   ├── src/main/scala/
│   │   ├── Update.scala
│   │   ├── Collection.scala
│   │   ├── WordCount.scala
│   │   └── Reachability.scala
│   └── build.sbt
│
├── query-exec/             # W08 — Scala 2.13
│   ├── src/main/scala/
│   │   ├── RowExecutor.scala
│   │   ├── ColumnFilter.scala
│   │   ├── ColumnProject.scala
│   │   ├── HashJoin.scala
│   │   └── Benchmark.scala
│   └── build.sbt
│
├── feature-pipeline/       # W09 — Python
│   ├── feature_store.py
│   ├── pipeline.py
│   └── requirements.txt
│
├── distributed-training/   # W10 — Python + Go tool
│   ├── mlp.py
│   ├── ring_allreduce.py
│   ├── worker.py
│   ├── train.py
│   └── requirements.txt
│   # Go tool lives in tools/grad_server/
│
├── gpu-gemm/               # W11 — Python/Numba + C fallback
│   ├── naive_gemm.py
│   ├── tiled_gemm.py
│   ├── benchmark.py
│   ├── roofline.py
│   ├── gemm_fallback.c     # no-GPU fallback
│   └── requirements.txt
│
├── attention/              # W12 — Python
│   ├── attention.py        # MultiHeadAttention
│   ├── kv_cache.py
│   ├── benchmark.py
│   └── requirements.txt
│
├── snapshot/               # W13 — Java 21 (+ optional Go)
│   ├── src/main/java/
│   │   ├── Channel.java
│   │   ├── Message.java
│   │   ├── Node.java
│   │   ├── Coordinator.java
│   │   └── SnapshotTest.java
│   └── pom.xml
│
├── capstone/               # W14 — your choice of language
│   ├── README.md           # required: design doc
│   └── ...
│
├── operator/               # W15 — Go
│   ├── api/v1/
│   │   ├── types.go
│   │   └── register.go
│   ├── controllers/
│   │   └── reconciler.go
│   ├── config/
│   │   ├── crd.yaml
│   │   └── sample.yaml
│   ├── main.go
│   └── go.mod
│
├── dd-scratch/             # W16 — extends W07
│   └── metrics/
│       ├── DDMetrics.scala
│       ├── tracing/DDTracer.scala
│       └── logging/Log.scala
│
└── capstone-platform/      # W17 (optional) — Go + Python, combines W09+W10+W12+W13+W15+W16
    ├── train_worker.py
    ├── checkpoint_coordinator.py
    ├── serve.py
    ├── operator/            # extends code/operator/ from W15
    └── README.md            # required: design doc
```

---

## Build Commands

**Java (Maven):**
```bash
mvn compile
mvn test
mvn exec:java -Dexec.mainClass=YourMainClass
```

**Java (Gradle):**
```bash
./gradlew build
./gradlew run
```

**Scala (SBT):**
```bash
sbt compile
sbt run
sbt test
sbt "runMain com.example.Benchmark"
```

**Python:**
```bash
python -m venv venv && source venv/bin/activate
pip install -r requirements.txt
python main.py
```

**Go:**
```bash
go mod tidy
go run main.go
go test ./...
go build -o bin/app .
```

---

## Notes

- Keep each project buildable in isolation — no shared parent build file
- Code in this directory is the "lab" — it's meant to be written, broken, and rewritten
- The `tools/` directory (at repo root) holds automation scripts that aren't part of a specific week's deliverable
