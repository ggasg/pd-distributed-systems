# Resources

All papers, books, and docs referenced in this curriculum. Organized by unit. Free links provided where available.

---

## Books (referenced across multiple units)

| Book | Author | Units | Notes |
|------|--------|-------|-------|
| [Designing Data-Intensive Applications (DDIA), 2nd ed.](https://dataintensive.net) | Kleppmann & Riccomini (2026) | Ch.1 pre-curriculum, then W01, W02, W03, W04, W05, W06, W07, W08, W13, W15, W16. 13 of 14 chapters cited; Ch.14 (law and society) is deliberately out of scope for an engineering plan | The single most useful book for this curriculum. Buy it. 2nd edition (Feb 2026) restructures the whole book: 12 chapters became 14, and every chapter after the first is renumbered. Chapter references below are to the 2nd edition. |
| [The Art of Multiprocessor Programming](https://www.amazon.com/dp/0123705916) | Herlihy & Shavit | W03, W13 | For concurrency primitives and correctness |
| [Observability Engineering](https://www.oreilly.com/library/view/observability-engineering/9781492076438/) | Majors, Fong-Jones, Miranda | W15 | O'Reilly; pairs well with the Google SRE chapter |
| [Google SRE Book](https://sre.google/sre-book/table-of-contents/) | Google | W15 (optional) | **Free online.** Ch. 6, optional; overlaps DDIA Ch.2 |
| [Designing Distributed Systems, 2nd ed.](https://www.oreilly.com/library/view/designing-distributed-systems/9781098156343/) | Burns (O'Reilly, Dec 2024) | W02, W04, W05, W06, W12, W13, W14, W15. 10 of 16 chapters cited | The pattern catalogue this curriculum is organised around, and short: 220 pages of mostly self-contained chapters. **Check your edition before using the chapter numbers below.** Every reference here is to the 2nd edition. Microsoft's free sponsored PDF has historically been the 2018 1st edition, whose numbering is different (its Ch.2 is The Sidecar Pattern, and it has no AI Inference chapter at all), so a free copy will send you to the wrong chapter every time |

---

## External Reading Lists

- [A Distributed Systems Reading List](https://dancres.github.io/Pages/) (Dan Creswell): a broad, long-running collection of foundational distributed-systems papers and essays, organized by theme (Google, Amazon, Consensus, Paxos, Gossip Protocols, P2P, and more). Some of it predates this curriculum's focus on ML/AI workloads and isn't a required source anywhere here, but it's a good browsing list once you're past a given unit and want more of that theme. Percolator (W06), Dremel (W06), and Chubby (extra-time list below) were added to this curriculum directly from its "Google" section, which was audited in full against what this curriculum already covers.

---

## Before You Start (outside any unit's budget)

- **DDIA Chapter 1** (2nd ed.), Trade-Offs in Data Systems Architecture. **Depth: skim.** Read this once, whenever you like, before or during W00. It is orientation rather than prerequisite: nothing in any unit depends on having read it, and it contains no mechanism you will implement. What it gives you is the operational-versus-analytical and distributed-versus-single-node framing that the whole curriculum is arranged around, which makes the difference between W01's write path and W06's columnar executor legible as a deliberate contrast rather than two unrelated builds. Forty-five minutes at skim depth. Do not study it.

DDIA Chapter 2 used to sit alongside it here. It now lives in W15, where reliability, scalability, and tail latency have a running system to attach to.

---

## W00: Infrastructure Setup

- [kind docs](https://kind.sigs.k8s.io/): Kubernetes in Docker
- [kube-prometheus-stack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack): Helm chart for Prometheus + Grafana
- [Prometheus Java client library](https://github.com/prometheus/client_java)

---

## W01: Storage Engines and the Cost of a Write

- **DDIA Chapter 4**: Storage and Retrieval. Required, read depth rather than study: the unit measures the chapter's central claim instead of implementing the mechanism. Focus on LSM-trees versus B-trees and the SSTable section
- [LevelDB source code](https://github.com/google/leveldb): optional, 20 minutes, `db/memtable.h` only. What a real MemTable looks like; C++, and the syntax doesn't matter
- [The Log-Structured Merge-Tree (LSM-tree)](https://www.cs.umb.edu/~poneil/lsmtree.pdf): O'Neil et al. (1996), optional history

---

## W02: MapReduce and Its Limits

- **DDIA Chapter 11**: Batch Processing. Read this first: it tells the MapReduce-to-Spark story as one continuous narrative and frames it as a point on a spectrum, not a standalone system.
- [MapReduce: Simplified Data Processing on Large Clusters](https://static.googleusercontent.com/media/research.google.com/en//archive/mapreduce-osdi04.pdf): Dean & Ghemawat, OSDI 2004 (**free PDF**)
- [Resilient Distributed Datasets (Spark)](https://www.usenix.org/system/files/conference/nsdi12/nsdi12-final138.pdf): Zaharia et al., NSDI 2012 (**free PDF**), required. The "why MapReduce isn't enough" paper, and its Section 1 argument is exactly what this unit measures on a real engine. Lineage, introduced here, is what makes the unit's `cache` versus `checkpoint` decision a real one
- [Spark: Cluster Computing with Working Sets](https://people.csail.mit.edu/matei/papers/2010/hotcloud_spark.pdf): Zaharia et al. (2010) (**free PDF**), optional, shorter, read first if the NSDI paper is heavy going
- [Spark Web UI documentation](https://spark.apache.org/docs/latest/web-ui.html): reference rather than reading. The Stages and SQL tabs are where this unit's evidence lives; come back to it when a panel shows you something you cannot name
- **Burns, *Designing Distributed Systems*, 2nd ed., Ch.13**: Coordinated Batch Processing, optional. Join as barrier synchronization, and reduce as a pattern rather than a function name

---

## W03: Clocks, Causality, Time, and Unreliable Networks

- [Time, Clocks, and the Ordering of Events in a Distributed System](https://dl.acm.org/doi/10.1145/359545.359563): Lamport (1978) (**ACM DL; 10 pages**), read all of it
- **DDIA Chapter 9**: The Trouble with Distributed Systems (clocks, NTP, monotonic clocks)
- [Spanner: Google's Globally Distributed Database](https://dl.acm.org/doi/10.1145/2491245): Corbett et al. (2012), TrueTime section only (Sections 3 + 5)
- [Detecting Causal Relationships in Distributed Computations](https://zoo.cs.yale.edu/classes/cs426/2012/lab/bib/fidge88timestamps.pdf): Fidge (1988) (**free PDF**), vector clocks
- **DDIA Chapter 9, "Timeouts and Unbounded Delays"**: required a second time, for the unit's failure-detector half. The one sentence that matters: over an asynchronous network a crashed node and a slow node produce identical evidence
- [Unreliable Failure Detectors for Reliable Distributed Systems](https://dl.acm.org/doi/10.1145/226643.226647): Chandra & Toueg, JACM 1996 (**ACM DL**; free copies are easy to find), optional and theory-heavy. Read for the framing rather than the algorithms: a failure detector is permitted to be wrong, and the useful questions are how wrong, how often, and how fast

---

## W04: Stream Processing Primitives

- [The Dataflow Model](https://research.google/pubs/the-dataflow-model-a-practical-approach-to-balancing-correctness-latency-and-cost-in-massive-scale-unbounded-out-of-order-data-processing/): Akidau et al., VLDB 2015 (**free PDF via Google Research**), the paper behind Apache Beam and Flink's model
- **DDIA Chapter 12**: Stream Processing (watermarks, windows, exactly-once)
- [Streaming 101](https://www.oreilly.com/radar/the-world-beyond-batch-streaming-101/): Akidau (O'Reilly blog), free, accessible intro before the paper
- [Flink: Generating Watermarks](https://nightlies.apache.org/flink/flink-docs-stable/docs/dev/datastream/event-time/generating_watermarks/): Part 1 reference. `forBoundedOutOfOrderness`, `allowedLateness`, and `sideOutputLateData` are the three knobs Part 1 turns
- [Spark Structured Streaming programming guide](https://spark.apache.org/docs/latest/structured-streaming-programming-guide.html): Part 1 reference, the "Handling Late Data and Watermarking" and "Output Modes" sections only. Spark's answer to the same question, and the differences from Flink are the point of the comparison
- [Flink: Network Stack and Backpressure](https://nightlies.apache.org/flink/flink-docs-stable/docs/ops/monitoring/back_pressure/): Part 2 required reading, short. How backpressure is detected and why it propagates upstream through the job graph rather than being absorbed locally. Note which of Part 2's three policies that is: Flink made the choice for you
- **Little's Law** (`L = λW`): no link needed, it is one line, but Part 2 is built on it. Average items in the system equals arrival rate times average time in the system. It is what turns "the queue is filling up" into arithmetic you can do before running anything
- **Burns, *Designing Distributed Systems*, 2nd ed., Ch.11**: Work Queue Systems, optional, Part 2. "Dynamic Scaling of the Workers" is the fourth response to overload that Part 2 does not give you

---

## W05: Partitioning and the Shuffle

- **DDIA Ch.7 (2nd ed.), "Sharding"**: required, the whole chapter. Key-range vs hash sharding, hot spots, rebalancing, and Kleppmann's deliberately skeptical treatment of consistent hashing for databases
- [Spark RDD Programming Guide: Shuffle operations](https://spark.apache.org/docs/latest/rdd-programming-guide.html#shuffle-operations): required, short. The map-side-write then reduce-side-fetch structure this unit builds, described by the system that made it famous
- [Dynamo: Amazon's Highly Available Key-value Store](https://www.allthingsdistributed.com/files/amazon-dynamo-sosp2007.pdf): DeCandia et al., SOSP 2007 (**free PDF**), optional. Section 4.2 is the canonical description of consistent hashing with virtual nodes, the technique the Python DSA Review implements
- **Burns, *Designing Distributed Systems*, 2nd ed., Ch.7**: Sharded Services, required. Sharding functions, selecting a key, and "Hot Sharding Systems," which is the production answer to the straggler you find in the Spark UI

---

## W06: Query Execution

- **DDIA Chapter 3**: Data Models and Query Languages, required. The declarative-versus-imperative argument is why an engine is free to vectorize at all: a `SELECT` does not specify a row-at-a-time loop
- [Volcano, An Extensible and Parallel Query Evaluation System](https://dl.acm.org/doi/10.1109/69.273032): Graefe (1994), the iterator model; skim for the `open/next/close` interface
- [MonetDB/X100: Hyper-Pipelining Query Execution](https://www.cidrdb.org/cidr2005/papers/P19.pdf): Boncz, Zukowski, Nes, CIDR 2005 (**free PDF**), the vectorized execution paper
- [Dremel: Interactive Analysis of Web-Scale Datasets](https://research.google/pubs/dremel-interactive-analysis-of-web-scale-datasets/): Melnik et al., Google, VLDB 2010 (**free PDF**), optional. Columnar storage (the ancestor of Parquet) plus a multi-level serving tree that fans aggregation out across thousands of machines, the distributed-scale continuation of the single-node vectorized-execution argument above. Sourced from [dancres' reading list](https://dancres.github.io/Pages/), Google section.
- [An Overview of Query Optimization in Relational Systems](https://dl.acm.org/doi/10.1145/275487.275492): Chaudhuri (1998), optional background
- [DuckDB execution engine source](https://github.com/duckdb/duckdb/tree/main/src/execution): optional but recommended. A real, actively maintained vectorized query engine in C++, and no longer a foreign artifact: this unit measures it, and W08 depends on it.
- [DuckDB `EXPLAIN ANALYZE`](https://duckdb.org/docs/stable/guides/meta/explain_analyze): reference. Per-operator timing is the artifact this unit is after, and reading it is most of the exercise.
- **Burns, *Designing Distributed Systems*, 2nd ed., Ch.8**: Scatter/Gather, optional. Pairs with Dremel above; "Choosing the Right Number of Leaves" is the partition-count question again

---

## W07: Query Planning: Choosing Where Data Moves

- **DDIA Chapter 3**: carried over from W06 if you skipped it there. A planner only gets to choose a join strategy because you did not tell it which one to use
- [Spark SQL: Relational Data Processing in Spark](https://people.csail.mit.edu/matei/papers/2015/sigmod_spark_sql.pdf): Armbrust et al., SIGMOD 2015 (**free PDF**), required. The physical-planning section is the one that matters: rules first, then cost, and the join strategy is what the cost model is for
- [Spark: Adaptive Query Execution](https://spark.apache.org/docs/latest/sql-performance-tuning.html#adaptive-query-execution): required, short. Dynamically switching join strategies exists because pre-execution estimates are unreliable at scale
- [Spark SQL performance tuning](https://spark.apache.org/docs/latest/sql-performance-tuning.html): required, reference. `autoBroadcastJoinThreshold` is the number this unit has you cross on purpose, and `spark.sql.adaptive.*` is the family of knobs the fourth experiment turns on
- [Spark: `EXPLAIN` syntax](https://spark.apache.org/docs/latest/sql-ref-syntax-qry-explain.html): reference. `FORMATTED` and `COST` are the two modes this unit lives in; the others are worth knowing exist

---

## W08: ML Data Pipelines

- **DDIA Chapter 5**: Encoding and Evolution, required. Backward versus forward compatibility is exactly the question "can a model trained on v1 features read v2, and can a v2 reader still read v1"

- [Hidden Technical Debt in Machine Learning Systems](https://papers.nips.cc/paper_files/paper/2015/file/86df7dcfd896fcaf2674f757a2463eba-Paper.pdf): Sculley et al., NeurIPS 2015 (**free PDF**)
- [Delta Lake: High-Performance ACID Table Storage](https://www.vldb.org/pvldb/vol13/p3411-armbrust.pdf): Armbrust et al., VLDB 2020 (**free PDF**)
- [DuckDB docs](https://duckdb.org/docs/): for the SQL-on-Parquet layer in the feature store
- [Apache Parquet spec](https://parquet.apache.org/docs/file-format/): understand the columnar format your feature store writes
- [Apache Iceberg Table Spec](https://iceberg.apache.org/spec/): Part 2 required reading, "Overview" and "Table Metadata" sections only. The three-level metadata file to manifest list to manifest structure, worth seeing next to Delta's flat commit log so you don't mistake one implementation for the concept
- [delta-rs (`deltalake` Python package)](https://delta-io.github.io/delta-rs/): Part 2's dependency. An independent Rust implementation of the Delta format with a Python binding, so no JVM and no Spark cluster are involved. Chosen for being the shortest path to a real transaction log locally, not because the format matters more than Iceberg's
- Optional, for the memory exercise: [pandas PyArrow-backed dtypes (`dtype_backend`)](https://pandas.pydata.org/docs/user_guide/pyarrow.html) and [`pyarrow.parquet.ParquetFile.iter_batches`](https://arrow.apache.org/docs/python/generated/pyarrow.parquet.ParquetFile.html) docs

---

## W09: Distributed Training

- [Horovod: fast and easy distributed deep learning in TensorFlow](https://arxiv.org/abs/1802.05799): Sergeev & Del Balso (2018) (**free on arXiv**), focus on Section 3 (ring-allreduce)
- [PyTorch Distributed: Experiences on Accelerating Data Parallel Training](https://arxiv.org/abs/2006.15704): Li et al. (2020) (**free on arXiv**), how DDP actually works
- [PyTorch DDP source](https://github.com/pytorch/pytorch/blob/main/torch/distributed/distributed_c10d.py): `all_reduce` function
- [NCCL: Collective Operations](https://docs.nvidia.com/deeplearning/nccl/user-guide/docs/usage/collectives.html): short reference page, read for vocabulary. `AllReduce`, `ReduceScatter`, `AllGather`, and the fact that the first is built from the other two

---

## W10: Beyond Data Parallelism

- [Megatron-LM: Training Multi-Billion Parameter Language Models Using Model Parallelism](https://arxiv.org/abs/1909.08053): Shoeybi et al. (2019) (**free on arXiv**), Sections 1 and 3. Column-parallel then row-parallel composition, and why one all-reduce per MLP block is enough
- [GPipe: Efficient Training of Giant Neural Networks using Pipeline Parallelism](https://arxiv.org/abs/1811.06965): Huang et al., NeurIPS 2019 (**free on arXiv**), Sections 1-3. Microbatching and the bubble
- [ZeRO: Memory Optimizations Toward Training Trillion Parameter Models](https://arxiv.org/abs/1910.02054): Rajbhandari et al., SC 2020 (**free on arXiv**), Section 5 and the stage table. Shard optimizer state, then gradients, then parameters
- [PyTorch FSDP: Experiences on Scaling Fully Sharded Data Parallel](https://arxiv.org/abs/2304.11277): Zhao et al., VLDB 2023 (**free on arXiv**), optional. ZeRO once it became a production API

---

## W11: The Actor Model and Ray

- [A Universal Modular ACTOR Formalism for Artificial Intelligence](https://www.ijcai.org/Proceedings/73/Papers/027B.pdf): Hewitt, Bishop, Steiger, IJCAI 1973 (**free PDF**), the original actor model paper
- [Ray: A Distributed Framework for Emerging AI Applications](https://www.usenix.org/system/files/osdi18-moritz.pdf): Moritz et al., OSDI 2018 (**free PDF via USENIX**), Section 3 is the unified task/actor programming model
- [Ray Core: Actors docs](https://docs.ray.io/en/latest/ray-core/actors.html): official Ray documentation on the actor API

---

## W12: Attention, KV Cache, and Cache-Aware Routing

- **Burns, *Designing Distributed Systems*, 2nd ed., Chapter 15** (optional): AI Inference and Serving. "Hosting a Model" and "Distributing a Model" give the production-serving framing for why the KV cache tradeoff this unit measures matters outside a benchmark script.
- [Attention Is All You Need](https://arxiv.org/abs/1706.03762): Vaswani et al. (2017) (**free on arXiv**), the transformer paper
- [FlashAttention: Fast and Memory-Efficient Exact Attention](https://arxiv.org/abs/2205.14135): Dao et al. (2022) (**free on arXiv**), read the intro and Section 2
- [Efficient Memory Management for Large Language Model Serving with PagedAttention](https://arxiv.org/abs/2309.06180): Kwon et al. (2023) (**free on arXiv**)
- [Introducing Gateway API Inference Extension](https://kubernetes.io/blog/2025/06/05/introducing-gateway-api-inference-extension/): Kubernetes blog (June 2025), Part 2 required reading, short. Inference-aware routing on KV cache utilization and LoRA readiness, i.e. facts about a replica's state that a normal load balancer is built to ignore
- [KV cache aware routing with llm-d](https://developers.redhat.com/articles/2025/10/07/master-kv-cache-aware-routing-llm-d-efficient-ai-inference): Red Hat, Part 2 optional. The same idea against real vLLM replicas, reporting up to 3x on time-to-first-token
- [kubernetes-sigs/gateway-api-inference-extension](https://github.com/kubernetes-sigs/gateway-api-inference-extension): the production version of Part 2's `router.py`, written in Go and running as part of the control plane rather than the model server. Reading only, no build target; it's here to connect this unit to W14
- [The Illustrated Transformer](https://jalammar.github.io/illustrated-transformer/): Jay Alammar (free blog), visual intro before the paper
- **Burns, *Designing Distributed Systems*, 2nd ed., Chapter 6**: Replicated Load-Balanced Services, required for Part 2. "Session Tracked Services" is cache-aware routing written down twenty years early: route to the replica holding the state, and accept worse load balance to get it

---

## W13: Fault Tolerance and Snapshots

- **DDIA Chapter 6**: Replication, required. The curriculum's largest genuine reading gap until now: nothing covered how systems keep more than one copy of data. Read the replication-lag and failover sections; failover is W03's crashed-or-slow ambiguity with writes on the line
- **Burns, *Designing Distributed Systems*, 2nd ed., Ch.16**: Common Failure Patterns, optional skim. Three of its entries are failures this curriculum has you cause on purpose

- [Distributed Snapshots: Determining Global States of Distributed Systems](https://dl.acm.org/doi/10.1145/214451.214456): Chandy & Lamport (1985), 10 pages; read all of it
- [Lightweight Asynchronous Snapshots for Distributed Dataflows](https://arxiv.org/abs/1506.08603): Carbone et al. (2015) (**free on arXiv**), Flink's ABS algorithm
- **DDIA Chapter 10** (optional): Consistency and Consensus, the linearizability section specifically. It sharpens the distinction between "consistent cut" (what Chandy-Lamport gives you) and linearizability (a stronger guarantee it doesn't).
- [In Search of an Understandable Consensus Algorithm](https://raft.github.io/raft.pdf): Ongaro & Ousterhout, USENIX ATC 2014 (**free PDF**), the Raft paper. Not implemented anywhere in this curriculum (see a deliberate scope call), but it's the algorithm underneath etcd, watched directly in W14. Read Sections 1–5.

---

## W14: Operating Kubernetes Operators (Kubeflow Trainer + Spark Operator)

- **Burns, *Designing Distributed Systems*, 2nd ed., Chapter 2**: Important Distributed System Concepts. Read "Idempotency" and "Orchestration and Kubernetes" before deploying either operator; the chapter argues directly for why a reconcile loop has to be idempotent, the same claim this unit's Reflect section asks you to defend.
- [Kubernetes Operators docs](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/): official k8s docs
- [Kubeflow Trainer documentation](https://www.kubeflow.org/docs/components/trainer/): architecture overview plus the `TrainJob`, `TrainingRuntime`, and `ClusterTrainingRuntime` APIs. Trainer v2 unified the older framework-specific CRDs (`PyTorchJob`, `MPIJob`, `JAXJob`, `XGBoostJob`) into one `TrainJob` plus a pluggable runtime, which is why it's the portable choice
- [kubeflow/trainer releases](https://github.com/kubeflow/trainer/releases): pin a current `v2.x.y` before installing; the manifests move between releases, so don't copy a version tag out of a tutorial
- [kubeflow/trainer source](https://github.com/kubeflow/trainer): search for `TrainJobReconciler`; the real reconcile loop this unit has you read, not write
- [Kubeflow Spark Operator documentation](https://kubeflow.github.io/spark-operator/): quick-start guide and the `SparkApplication` API reference
- [Kueue overview](https://kueue.sigs.k8s.io/docs/overview/): fifteen minutes, read-only, for Part 4's written exercise. Gang admission on Kubernetes. The idea is older than Kubernetes (Ousterhout named gang scheduling in 1982, and Slurm and Borg both implement a version of it); Kueue is one current answer, not the concept
- [Running Spark on Kubernetes](https://spark.apache.org/docs/latest/running-on-kubernetes.html): Apache Spark's own docs, for Part 2b. The `local://` scheme, `mainClass`/`mainApplicationFile`, and why your image's Spark version has to match what you compiled against
- [Programming Kubernetes](https://www.oreilly.com/library/view/programming-kubernetes/9781492047094/): Hausenblas & Schimanski (O'Reilly), optional. Covers how operators like these two are actually built; useful context even though this unit has you operate one rather than author one
- [etcd: Set up a local cluster](https://etcd.io/docs/v3.5/dev-guide/local_cluster/): official docs for the 3-member local-cluster bootstrap Part 3 uses (run by hand here instead of via their `Procfile`/`goreman` wrapper)
- [etcd-io/raft](https://github.com/etcd-io/raft): the standalone Raft library etcd actually runs (also vendored into Kubernetes itself, and used by CockroachDB and TiKV); Part 3 has you read `raft.go`'s `becomeLeader`/`campaign`, not the whole file
- Recall W13's Raft paper (Ongaro & Ousterhout, 2014): Part 3 is where you watch the algorithm that paper describes run for real
- **Burns, *Designing Distributed Systems*, 2nd ed., Chapter 10**: Ownership Election. Both operators run leader election so only one replica reconciles; Part 3's etcd work is this chapter's hands-on. "Determining If You Even Need Leader Election" is the useful section

---

## W15: Observability: Metrics, Tracing, Logging

- **Burns, *Designing Distributed Systems*, 2nd ed., Chapter 5**: Adapters, optional. Its hands-on sections are Prometheus monitoring and normalizing log formats with fluentd, which is the Part 3 sidecar
- **Burns, *Designing Distributed Systems*, 2nd ed., Chapter 3**: The Sidecar Pattern. Read before Part 3; names the pattern your log-aggregator sidecar already implements.
- [Prometheus data model](https://prometheus.io/docs/concepts/data_model/) + [metric types](https://prometheus.io/docs/concepts/metric_types/)
- [OpenTelemetry concepts](https://opentelemetry.io/docs/concepts/)
- **DDIA Chapter 2**: Defining Nonfunctional Requirements. Moved here from W00 on purpose: percentiles and tail latency mean something once you have a system emitting them
- [Google SRE Book, Chapter 6: Monitoring Distributed Systems](https://sre.google/sre-book/monitoring-distributed-systems/): **free online**, optional, overlaps DDIA Ch.2 heavily
- [Grafana dashboarding docs](https://grafana.com/docs/grafana/latest/dashboards/)
- [Prometheus Java client](https://github.com/prometheus/client_java): used to instrument the W06 DD engine (same library W00 already uses)
- [OpenTelemetry Java SDK](https://opentelemetry.io/docs/languages/java/): official docs, used for the `ScopedSpan` tracing setup
- [Go `net/http` docs](https://pkg.go.dev/net/http): the standard library package the Part 3 log-aggregator sidecar is built on
- **Burns, *Designing Distributed Systems*, 2nd ed., Chapter 14**: Monitoring and Observability Patterns. Required, and the chapter this unit is named after. Treats logging, metrics, and tracing as one system rather than three tools

---

## W16: Grand Capstone (optional)

No required reading tied to this project's build. It synthesizes W08, W09, W12, W13, W14, and W15; revisit those units' resources as needed.

- [MLflow Model Registry](https://mlflow.org/docs/latest/model-registry.html): for Part 5. Registered models, versions, and aliases. The mental model that transfers: a registry is a commit log with movable pointers, the same shape as the Delta transaction log from W08, applied to models instead of tables. **DDIA Chapter 13** (A Philosophy of Streaming Systems, renamed from "The Future of Data Systems" in the 2nd edition) is optional but a fitting bookend: the book's own synthesis chapter, on unbundling databases into composable derived-data systems, read in the unit you're doing exactly that.

---

## What to Read If You Have Extra Time

These aren't required but give you broader context:

- **DDIA Chapter 8**, Transactions. No unit in this curriculum implements isolation levels or multi-object transactions, so there's no clean place to attach it as required reading, but it's the one DDIA chapter this curriculum otherwise skips entirely, and it's foundational enough to be worth reading on its own rather than forced into an unrelated unit.
- **DDIA Chapter 4** (2nd ed.) also gained a Vector Embeddings section, folded into the storage-and-retrieval chapter alongside full-text and multidimensional indexing. No unit currently builds against it, but it's the most directly AI-relevant new material in the 2nd edition and worth reading given the curriculum's focus on AI workflows; a natural pairing with W12's attention/KV-cache work if you want the retrieval side of the same systems.
- [The Google File System](https://dl.acm.org/doi/10.1145/945445.945450): Ghemawat et al., SOSP 2003, the original scale-out storage paper
- [Bigtable: A Distributed Storage System for Structured Data](https://dl.acm.org/doi/10.1145/1365815.1365816): Chang et al., OSDI 2006
- [The Chubby Lock Service for Loosely-Coupled Distributed Systems](https://research.google/pubs/the-chubby-lock-service-for-loosely-coupled-distributed-systems/): Burrows, Google, OSDI 2006, free PDF. Google's Paxos-based lock and small-file coordination service, the direct conceptual ancestor of etcd and ZooKeeper. A natural pairing with W13's Raft paper and W14's hands-on etcd cluster: same problem (a small, strongly-consistent coordination service other systems depend on), Paxos instead of Raft underneath. Sourced from [dancres' reading list](https://dancres.github.io/Pages/), Google section.
- [Spanner](https://dl.acm.org/doi/10.1145/2491245): Corbett et al. (2012), full read after W03
- [Amazon Dynamo](https://dl.acm.org/doi/10.1145/1294261.1294281): DeCandia et al., SOSP 2007, consistent hashing, vector clocks in practice
- [CAP Twelve Years Later](https://www.infoq.com/articles/cap-twelve-years-later-how-the-rules-have-changed/): Brewer (2012), free
- [CRDT: Conflict-free Replicated Data Types](https://hal.science/hal-00932836): Shapiro et al. (2011), natural follow-on to W03
