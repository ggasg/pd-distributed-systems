# Resources

All papers, books, and docs referenced in this curriculum. Organized by week. Free links provided where available.

---

## Books (referenced across multiple weeks)

| Book | Author | Weeks | Notes |
|------|--------|-------|-------|
| [Designing Data-Intensive Applications (DDIA)](https://dataintensive.net) | Kleppmann (2017) | W01, W02, W04, W05 | The single most useful book for this curriculum. Buy it. |
| [The Art of Multiprocessor Programming](https://www.amazon.com/dp/0123705916) | Herlihy & Shavit | W04, W14 | For concurrency primitives and correctness |
| [Observability Engineering](https://www.oreilly.com/library/view/observability-engineering/9781492076438/) | Majors, Fong-Jones, Miranda | W17 | O'Reilly; pairs well with the Google SRE chapter |
| [Google SRE Book](https://sre.google/sre-book/table-of-contents/) | Google | W17 | **Free online.** Read Ch. 6 (Monitoring Distributed Systems) |

---

## W00: Infrastructure Setup

- [kind docs](https://kind.sigs.k8s.io/): Kubernetes in Docker
- [kube-prometheus-stack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack): Helm chart for Prometheus + Grafana
- [Prometheus Go client library](https://github.com/prometheus/client_golang)

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
- [Timely Dataflow (Rust implementation)](https://github.com/TimelyDataflow/timely-dataflow): source code, even if you don't use Rust

---

## W07: Differential Dataflow

- [Differential Dataflow](https://github.com/frankmcsherry/blog/blob/master/posts/2015-09-29.md): McSherry (2013/2015), blog post (**free**), more accessible than the formal paper
- [Differential Dataflow (formal paper)](https://dl.acm.org/doi/10.1145/2588555.2610364): McSherry, Murray et al., CIDR 2013
- [Differential Dataflow (Rust implementation)](https://github.com/TimelyDataflow/differential-dataflow): source reference, even if you implement in Scala

---

## W08: Query Execution

- [Volcano, An Extensible and Parallel Query Evaluation System](https://dl.acm.org/doi/10.1109/69.273032): Graefe (1994), the iterator model; skim for the `open/next/close` interface
- [MonetDB/X100: Hyper-Pipelining Query Execution](https://www.cidrdb.org/cidr2005/papers/P19.pdf): Boncz, Zukowski, Nes, CIDR 2005 (**free PDF**), the vectorized execution paper
- [An Overview of Query Optimization in Relational Systems](https://dl.acm.org/doi/10.1145/275487.275492): Chaudhuri (1998), optional background

---

## W09: ML Data Pipelines

- [Hidden Technical Debt in Machine Learning Systems](https://papers.nips.cc/paper_files/paper/2015/file/86df7dcfd896fcaf2674f757a2463eba-Paper.pdf): Sculley et al., NeurIPS 2015 (**free PDF**)
- [Delta Lake: High-Performance ACID Table Storage](https://www.vldb.org/pvldb/vol13/p3411-armbrust.pdf): Armbrust et al., VLDB 2020 (**free PDF**)
- [DuckDB docs](https://duckdb.org/docs/): for the SQL-on-Parquet layer in the feature store
- [Apache Parquet spec](https://parquet.apache.org/docs/file-format/): understand the columnar format your feature store writes

---

## W10: Distributed Training

- [Horovod: fast and easy distributed deep learning in TensorFlow](https://arxiv.org/abs/1802.05799): Sergeev & Del Balso (2018) (**free on arXiv**), focus on Section 3 (ring-allreduce)
- [PyTorch Distributed: Experiences on Accelerating Data Parallel Training](https://arxiv.org/abs/2006.15704): Li et al. (2020) (**free on arXiv**), how DDP actually works
- [PyTorch DDP source](https://github.com/pytorch/pytorch/blob/main/torch/distributed/distributed_c10d.py): `all_reduce` function

---

## W11: The Actor Model and Ray

- [A Universal Modular ACTOR Formalism for Artificial Intelligence](https://www.ijcai.org/Proceedings/73/Papers/027B.pdf): Hewitt, Bishop, Steiger, IJCAI 1973 (**free PDF**), the original actor model paper
- [Ray: A Distributed Framework for Emerging AI Applications](https://www.usenix.org/system/files/osdi18-moritz.pdf): Moritz et al., OSDI 2018 (**free PDF via USENIX**), Section 3 is the unified task/actor programming model
- [Ray Core: Actors docs](https://docs.ray.io/en/latest/ray-core/actors.html): official Ray documentation on the actor API

---

## W12: GPU Memory and Compute

- [CUDA C++ Programming Guide](https://docs.nvidia.com/cuda/cuda-c-programming-guide/): NVIDIA docs; read Chapters 1–3 (Architecture, Programming Model, Memory Hierarchy)
- [Roofline: An Insightful Visual Performance Model](https://people.eecs.berkeley.edu/~kubitron/cs252/handouts/papers/RooflineVyNoYellow.pdf): Williams, Waterman, Patterson, CACM 2009 (**free PDF**)
- [Numba CUDA docs](https://numba.readthedocs.io/en/stable/cuda/index.html): Python GPU programming

---

## W13: Attention and KV Cache

- [Attention Is All You Need](https://arxiv.org/abs/1706.03762): Vaswani et al. (2017) (**free on arXiv**), the transformer paper
- [FlashAttention: Fast and Memory-Efficient Exact Attention](https://arxiv.org/abs/2205.14135): Dao et al. (2022) (**free on arXiv**), read the intro and Section 2
- [Efficient Memory Management for Large Language Model Serving with PagedAttention](https://arxiv.org/abs/2309.06180): Kwon et al. (2023) (**free on arXiv**)
- [The Illustrated Transformer](https://jalammar.github.io/illustrated-transformer/): Jay Alammar (free blog), visual intro before the paper

---

## W14: Fault Tolerance and Snapshots

- [Distributed Snapshots: Determining Global States of Distributed Systems](https://dl.acm.org/doi/10.1145/214451.214456): Chandy & Lamport (1985), 10 pages; read all of it
- [Lightweight Asynchronous Snapshots for Distributed Dataflows](https://arxiv.org/abs/1506.08603): Carbone et al. (2015) (**free on arXiv**), Flink's ABS algorithm

---

## W15: Capstone

No required reading. You're synthesizing earlier weeks.

---

## W16: Kubernetes Operators

- [Kubernetes Operators docs](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/): official k8s docs
- [controller-runtime](https://pkg.go.dev/sigs.k8s.io/controller-runtime): Go library for writing operators
- [Kubebuilder Book](https://book.kubebuilder.io/): Chapters 1–3 only
- [Programming Kubernetes](https://www.oreilly.com/library/view/programming-kubernetes/9781492047094/): Hausenblas & Schimanski (O'Reilly), optional deep dive

---

## W17: Observability: Metrics, Tracing, Logging

- [Prometheus data model](https://prometheus.io/docs/concepts/data_model/) + [metric types](https://prometheus.io/docs/concepts/metric_types/)
- [OpenTelemetry concepts](https://opentelemetry.io/docs/concepts/)
- [Google SRE Book, Chapter 6: Monitoring Distributed Systems](https://sre.google/sre-book/monitoring-distributed-systems/): **free online**
- [Grafana dashboarding docs](https://grafana.com/docs/grafana/latest/dashboards/)

---

## W18: Grand Capstone (optional)

No new required reading. This week synthesizes W09, W10, W13, W14, W16, and W17. Revisit those weeks' resources as needed.

---

## What to Read If You Have Extra Time

These aren't required but give you broader context:

- [The Google File System](https://dl.acm.org/doi/10.1145/945445.945450): Ghemawat et al., SOSP 2003, the original scale-out storage paper
- [Bigtable: A Distributed Storage System for Structured Data](https://dl.acm.org/doi/10.1145/1365815.1365816): Chang et al., OSDI 2006
- [Spanner](https://dl.acm.org/doi/10.1145/2491245): Corbett et al. (2012), full read after W04
- [Amazon Dynamo](https://dl.acm.org/doi/10.1145/1294261.1294281): DeCandia et al., SOSP 2007, consistent hashing, vector clocks in practice
- [CAP Twelve Years Later](https://www.infoq.com/articles/cap-twelve-years-later-how-the-rules-have-changed/): Brewer (2012), free
- [CRDT: Conflict-free Replicated Data Types](https://hal.science/hal-00932836): Shapiro et al. (2011), natural follow-on to W04
