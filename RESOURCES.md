# Resources

All papers, books, and docs referenced in this curriculum. Organized by week. Free links provided where available.

---

## Books (referenced across multiple weeks)

| Book | Author | Weeks | Notes |
|------|--------|-------|-------|
| [Designing Data-Intensive Applications (DDIA)](https://dataintensive.net) | Kleppmann (2017) | W00, W01, W02, W03, W04, W05, W17 (optional), W18 (Option A), W21 (optional) | The single most useful book for this curriculum. Buy it. |
| [The Art of Multiprocessor Programming](https://www.amazon.com/dp/0123705916) | Herlihy & Shavit | W04, W17 | For concurrency primitives and correctness |
| [Observability Engineering](https://www.oreilly.com/library/view/observability-engineering/9781492076438/) | Majors, Fong-Jones, Miranda | W20 | O'Reilly; pairs well with the Google SRE chapter |
| [Google SRE Book](https://sre.google/sre-book/table-of-contents/) | Google | W20 | **Free online.** Read Ch. 6 (Monitoring Distributed Systems) |

---

## W00: Infrastructure Setup

- **DDIA Chapter 1**: Reliable, Scalable, and Maintainable Applications. Read before anything else in the curriculum; it isn't tied to this week's build specifically.
- [kind docs](https://kind.sigs.k8s.io/): Kubernetes in Docker
- [kube-prometheus-stack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack): Helm chart for Prometheus + Grafana
- [Prometheus Java client library](https://github.com/prometheus/client_java)

---

## W01: LSM-Trees and Storage Engines

- **DDIA Chapter 3**: Storage and Retrieval (SSTables, LSM-Trees, B-Trees)
- [LevelDB source code](https://github.com/google/leveldb): the canonical LSM implementation; read `db/memtable.h`, `table/table.cc`
- [The Log-Structured Merge-Tree (LSM-tree)](https://www.cs.umb.edu/~poneil/lsmtree.pdf): O'Neil et al. (1996), the original paper

---

## W02: Encoding and Wire Formats

- **DDIA Chapter 4**: Encoding and Evolution
- [Protocol Buffers Encoding](https://protobuf.dev/programming-guides/encoding/): Google docs; the varint encoding spec
- [MessagePack spec](https://github.com/msgpack/msgpack/blob/master/spec.md): binary JSON alternative worth understanding

---

## W03: MapReduce and Its Limits

- **DDIA Chapter 10**: Batch Processing. Read this first: it tells the MapReduce-to-Spark story as one continuous narrative and frames it as a point on a spectrum, not a standalone system.
- [MapReduce: Simplified Data Processing on Large Clusters](https://static.googleusercontent.com/media/research.google.com/en//archive/mapreduce-osdi04.pdf): Dean & Ghemawat, OSDI 2004 (**free PDF**)
- [Resilient Distributed Datasets (Spark)](https://www.usenix.org/system/files/conference/nsdi12/nsdi12-final138.pdf): Zaharia et al., NSDI 2012 (**free PDF**), the "why MapReduce isn't enough" paper
- [Spark: Cluster Computing with Working Sets](https://people.csail.mit.edu/matei/papers/2010/hotcloud_spark.pdf): Zaharia et al. (2010) (**free PDF**), shorter, read first

---

## W04: Clocks, Causality, and Time

- [Time, Clocks, and the Ordering of Events in a Distributed System](https://dl.acm.org/doi/10.1145/359545.359563): Lamport (1978) (**ACM DL; 10 pages**), read all of it
- **DDIA Chapter 8**: The Trouble with Distributed Systems (clocks, NTP, monotonic clocks)
- [Spanner: Google's Globally Distributed Database](https://dl.acm.org/doi/10.1145/2491245): Corbett et al. (2012), TrueTime section only (Sections 3 + 5)
- [Detecting Causal Relationships in Distributed Computations](https://zoo.cs.yale.edu/classes/cs426/2012/lab/bib/fidge88timestamps.pdf): Fidge (1988) (**free PDF**), vector clocks

---

## W05: Stream Processing Primitives

- [The Dataflow Model](https://research.google/pubs/the-dataflow-model-a-practical-approach-to-balancing-correctness-latency-and-cost-in-massive-scale-unbounded-out-of-order-data-processing/): Akidau et al., VLDB 2015 (**free PDF via Google Research**), the paper behind Apache Beam and Flink's model
- **DDIA Chapter 11**: Stream Processing (watermarks, windows, exactly-once)
- [Streaming 101](https://www.oreilly.com/radar/the-world-beyond-batch-streaming-101/): Akidau (O'Reilly blog), free, accessible intro before the paper

---

## W06: Naiad and Timely Dataflow

- [Naiad: A Timely Dataflow System](https://dl.acm.org/doi/10.1145/2517349.2522738): Murray et al., SOSP 2013 (**ACM DL**; also [free via MSR](https://www.microsoft.com/en-us/research/wp-content/uploads/2013/11/naiad_sosp2013.pdf))
- [Timely Dataflow (Rust implementation)](https://github.com/TimelyDataflow/timely-dataflow): `timely-toy/` was originally planned as a simplified version of this, in the same language. Since Arc 2 moved to C++, this stays a conceptual reference only; no build target reads it anymore.
- [PyTorch Autograd Engine source](https://github.com/pytorch/pytorch/blob/main/torch/csrc/autograd/engine.cpp): the actual C++ reference for `timely-toy/`, a production dependency-counted DAG scheduler and the closest real analogue to Naiad's progress tracking in a language you're writing this arc in
- [Ray source: `core_worker.cc`](https://github.com/ray-project/ray/blob/master/src/ray/core_worker/core_worker.cc) and [`task_manager.cc`](https://github.com/ray-project/ray/blob/master/src/ray/core_worker/task_manager.cc): Ray's `CoreWorker` (C++) fires tasks once their dependencies are satisfied: a second, directly target-company-relevant analogue (Anyscale), read from source rather than a third-party summary

---

## W07: Differential Dataflow and Incremental View Maintenance

- [Differential Dataflow](https://github.com/frankmcsherry/blog/blob/master/posts/2015-09-29.md): McSherry (2013/2015), blog post (**free**), more accessible than the formal paper
- [Differential Dataflow (formal paper)](https://dl.acm.org/doi/10.1145/2588555.2610364): McSherry, Murray et al., CIDR 2013 (Part 1 reading, Sections 1–2 only)
- [Differential Dataflow (Rust implementation)](https://github.com/TimelyDataflow/differential-dataflow): optional, for context. Your `Update`/`Collection` types in `dd-scratch/` were originally planned as simplified versions of `collection.rs`. No maintained C++ continuation of this lineage exists, so this stays a reading-only reference.
- [ClickHouse Materialized Views](https://clickhouse.com/docs/en/guides/developer/cascading-materialized-views): Part 2 required reading. An insert trigger, not a retraction-aware incrementally maintained view. You'll install ClickHouse locally and build one yourself in the exercise.
- [Spark Structured Streaming: arbitrary stateful operations](https://spark.apache.org/docs/latest/structured-streaming-programming-guide.html#arbitrary-stateful-operations): Part 2 required reading. Per-key state maintained and updated incrementally between micro-batches. Runs entirely in local mode via `pip install pyspark`, no Databricks account or proprietary docs needed.
- [Snowflake Dynamic Tables](https://docs.snowflake.com/en/user-guide/dynamic-tables-about): optional, not required. Closed-source SaaS with no self-hosted option, so it's excluded from the hands-on comparison the same way ClickHouse would have been if it weren't locally installable.
- [pg_ivm](https://github.com/sraoss/pg_ivm): optional stretch. A real, actively maintained PostgreSQL extension for true incremental view maintenance, closer in spirit to DD than ClickHouse or Spark's approach, but requires building a Postgres extension from source (PGXS) rather than a single-binary or `pip install`.

---

## W08: Query Execution

- [Volcano, An Extensible and Parallel Query Evaluation System](https://dl.acm.org/doi/10.1109/69.273032): Graefe (1994), the iterator model; skim for the `open/next/close` interface
- [MonetDB/X100: Hyper-Pipelining Query Execution](https://www.cidrdb.org/cidr2005/papers/P19.pdf): Boncz, Zukowski, Nes, CIDR 2005 (**free PDF**), the vectorized execution paper
- [An Overview of Query Optimization in Relational Systems](https://dl.acm.org/doi/10.1145/275487.275492): Chaudhuri (1998), optional background
- [DuckDB execution engine source](https://github.com/duckdb/duckdb/tree/main/src/execution): optional but recommended. A real, actively maintained vectorized query engine in C++, already in your stack via W11's feature store.
- [Announcing Photon](https://www.databricks.com/blog/2021/06/17/announcing-photon-public-preview-the-next-generation-query-engine-on-the-databricks-lakehouse-platform.html): optional, context only (a free public blog post, not something you install or test against). Databricks' vectorized engine, written from the ground up in C++, built to replace JVM-based Spark execution for the exact row-vs-vectorized reasons this week benchmarks.
- [ClickHouse execution pipeline source](https://github.com/ClickHouse/ClickHouse/tree/master/src/Processors): optional. `IProcessor` and `Chunk`-based batching, a second C++ production reference with a different pipeline design than DuckDB or Photon.

---

## W09: Rule-Based Query Planning in Scala

- [Spark SQL: Relational Data Processing in Spark](https://people.csail.mit.edu/matei/papers/2015/sigmod_spark_sql.pdf): Armbrust et al., SIGMOD 2015 (**free PDF**), Section 4 describes Catalyst directly
- [Catalyst source: `TreeNode.scala`](https://github.com/apache/spark/blob/master/sql/catalyst/src/main/scala/org/apache/spark/sql/catalyst/trees/TreeNode.scala): the real `transform`/`transformDown`/`transformUp` combinators
- [Catalyst source: `Optimizer.scala`](https://github.com/apache/spark/blob/master/sql/catalyst/src/main/scala/org/apache/spark/sql/catalyst/optimizer/Optimizer.scala): search for `PushDownPredicates`, the production version of this week's rewrite rule

---

## W10: Aggregation Algebra: Monoids and Semigroups

- [Algebird `Semigroup.scala`](https://github.com/twitter/algebird/blob/develop/algebird-core/src/main/scala/com/twitter/algebird/Semigroup.scala) and [`Monoid.scala`](https://github.com/twitter/algebird/blob/develop/algebird-core/src/main/scala/com/twitter/algebird/Monoid.scala): Twitter's real Scala library built around this idea, published for Scala 2.13, the same version W10 targets, so it's directly usable, not just readable. The exercise builds its own typeclass from scratch on purpose; read Algebird as prior art, not a dependency.
- [Of Algebirds, Monoids, Monads, and Other Bestiary for Large-Scale Data Analytics](https://www.michael-noll.com/blog/2013/12/02/twitter-algebird-monoid-monad-for-large-scala-data-analytics/): Michael Noll, an accessible walkthrough with concrete MapReduce-shaped examples

---

## W11: ML Data Pipelines

- [Hidden Technical Debt in Machine Learning Systems](https://papers.nips.cc/paper_files/paper/2015/file/86df7dcfd896fcaf2674f757a2463eba-Paper.pdf): Sculley et al., NeurIPS 2015 (**free PDF**)
- [Delta Lake: High-Performance ACID Table Storage](https://www.vldb.org/pvldb/vol13/p3411-armbrust.pdf): Armbrust et al., VLDB 2020 (**free PDF**)
- [DuckDB docs](https://duckdb.org/docs/): for the SQL-on-Parquet layer in the feature store
- [Apache Parquet spec](https://parquet.apache.org/docs/file-format/): understand the columnar format your feature store writes
- Optional, for the memory exercise: [pandas PyArrow-backed dtypes (`dtype_backend`)](https://pandas.pydata.org/docs/user_guide/pyarrow.html) and [`pyarrow.parquet.ParquetFile.iter_batches`](https://arrow.apache.org/docs/python/generated/pyarrow.parquet.ParquetFile.html) docs

---

## W12: PySpark vs. Scala Spark: Where the JVM Boundary Costs You

- [Introducing Pandas UDF for PySpark](https://www.databricks.com/blog/2017/10/30/introducing-vectorized-udfs-for-pyspark.html): Databricks engineering blog (2017), names the py4j-serialization mechanism before you go measure it yourself
- [Apache Spark docs: Pandas UDFs (a.k.a. Vectorized UDFs)](https://spark.apache.org/docs/latest/api/python/user_guide/sql/arrow_pandas.html): current reference for `pandas_udf`, needed for the optional stretch arm
- Recall W09's Spark SQL/Catalyst paper (Armbrust et al., SIGMOD 2015): this week runs the physical plan you read about there

---

## W13: Distributed Training

- [Horovod: fast and easy distributed deep learning in TensorFlow](https://arxiv.org/abs/1802.05799): Sergeev & Del Balso (2018) (**free on arXiv**), focus on Section 3 (ring-allreduce)
- [PyTorch Distributed: Experiences on Accelerating Data Parallel Training](https://arxiv.org/abs/2006.15704): Li et al. (2020) (**free on arXiv**), how DDP actually works
- [PyTorch DDP source](https://github.com/pytorch/pytorch/blob/main/torch/distributed/distributed_c10d.py): `all_reduce` function

---

## W14: The Actor Model and Ray

- [A Universal Modular ACTOR Formalism for Artificial Intelligence](https://www.ijcai.org/Proceedings/73/Papers/027B.pdf): Hewitt, Bishop, Steiger, IJCAI 1973 (**free PDF**), the original actor model paper
- [Ray: A Distributed Framework for Emerging AI Applications](https://www.usenix.org/system/files/osdi18-moritz.pdf): Moritz et al., OSDI 2018 (**free PDF via USENIX**), Section 3 is the unified task/actor programming model
- [Ray Core: Actors docs](https://docs.ray.io/en/latest/ray-core/actors.html): official Ray documentation on the actor API

---

## W15: GPU Memory and Compute

- [CUDA C++ Programming Guide](https://docs.nvidia.com/cuda/cuda-c-programming-guide/): NVIDIA docs; read Chapters 1–3 (Architecture, Programming Model, Memory Hierarchy)
- [Roofline: An Insightful Visual Performance Model](https://people.eecs.berkeley.edu/~kubitron/cs252/handouts/papers/RooflineVyNoYellow.pdf): Williams, Waterman, Patterson, CACM 2009 (**free PDF**)
- [Numba CUDA docs](https://numba.readthedocs.io/en/stable/cuda/index.html): Python GPU programming

---

## W16: Attention and KV Cache

- [Attention Is All You Need](https://arxiv.org/abs/1706.03762): Vaswani et al. (2017) (**free on arXiv**), the transformer paper
- [FlashAttention: Fast and Memory-Efficient Exact Attention](https://arxiv.org/abs/2205.14135): Dao et al. (2022) (**free on arXiv**), read the intro and Section 2
- [Efficient Memory Management for Large Language Model Serving with PagedAttention](https://arxiv.org/abs/2309.06180): Kwon et al. (2023) (**free on arXiv**)
- [The Illustrated Transformer](https://jalammar.github.io/illustrated-transformer/): Jay Alammar (free blog), visual intro before the paper

---

## W17: Fault Tolerance and Snapshots

- [Distributed Snapshots: Determining Global States of Distributed Systems](https://dl.acm.org/doi/10.1145/214451.214456): Chandy & Lamport (1985), 10 pages; read all of it
- [Lightweight Asynchronous Snapshots for Distributed Dataflows](https://arxiv.org/abs/1506.08603): Carbone et al. (2015) (**free on arXiv**), Flink's ABS algorithm
- **DDIA Chapter 9** (optional): Consistency and Consensus, the linearizability section specifically. It sharpens the distinction between "consistent cut" (what Chandy-Lamport gives you) and linearizability (a stronger guarantee it doesn't).

---

## W18: Capstone

No required reading. You're synthesizing earlier weeks. **If you choose Option A** (distributed KV store): **DDIA Chapter 5**, Replication. Read "Leaders and Followers" before writing `promote()`.

---

## W19: Kubernetes Operators

- [Kubernetes Operators docs](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/): official k8s docs
- [controller-runtime](https://pkg.go.dev/sigs.k8s.io/controller-runtime): Go library for writing operators
- [Kubebuilder Book](https://book.kubebuilder.io/): Chapters 1–3 only
- [Programming Kubernetes](https://www.oreilly.com/library/view/programming-kubernetes/9781492047094/): Hausenblas & Schimanski (O'Reilly), optional deep dive

---

## W20: Observability: Metrics, Tracing, Logging

- [Prometheus data model](https://prometheus.io/docs/concepts/data_model/) + [metric types](https://prometheus.io/docs/concepts/metric_types/)
- [OpenTelemetry concepts](https://opentelemetry.io/docs/concepts/)
- [Google SRE Book, Chapter 6: Monitoring Distributed Systems](https://sre.google/sre-book/monitoring-distributed-systems/): **free online**
- [Grafana dashboarding docs](https://grafana.com/docs/grafana/latest/dashboards/)
- [prometheus-cpp](https://github.com/jupp0r/prometheus-cpp): the C++ Prometheus client used to instrument the W07 DD engine
- [OpenTelemetry C++ SDK](https://opentelemetry.io/docs/languages/cpp/): official docs, used for the `ScopedSpan` tracing setup

---

## W21: Grand Capstone (optional)

No required reading tied to this week's build. This week synthesizes W11, W13, W16, W17, W19, and W20; revisit those weeks' resources as needed. **DDIA Chapter 12** (The Future of Data Systems) is optional but a fitting bookend: the book's own synthesis chapter, on unbundling databases into composable derived-data systems, read in the week you're doing exactly that.

---

## What to Read If You Have Extra Time

These aren't required but give you broader context:

- [The Google File System](https://dl.acm.org/doi/10.1145/945445.945450): Ghemawat et al., SOSP 2003, the original scale-out storage paper
- [Bigtable: A Distributed Storage System for Structured Data](https://dl.acm.org/doi/10.1145/1365815.1365816): Chang et al., OSDI 2006
- [Spanner](https://dl.acm.org/doi/10.1145/2491245): Corbett et al. (2012), full read after W04
- [Amazon Dynamo](https://dl.acm.org/doi/10.1145/1294261.1294281): DeCandia et al., SOSP 2007, consistent hashing, vector clocks in practice
- [CAP Twelve Years Later](https://www.infoq.com/articles/cap-twelve-years-later-how-the-rules-have-changed/): Brewer (2012), free
- [CRDT: Conflict-free Replicated Data Types](https://hal.science/hal-00932836): Shapiro et al. (2011), natural follow-on to W04
