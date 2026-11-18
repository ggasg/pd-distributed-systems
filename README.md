# Distributed Data Systems

A self-study curriculum in distributed query engines. Build one Scala project from a local scan to a query that shuffles data across worker processes. Use Python and SQL to compare with Spark and DuckDB. The optional Python ML extension applies the same ideas to feature pipelines, training and serving.

Open [Home](Home.md) in Obsidian and follow [SETUP](SETUP.md). Each unit has four study sessions and an artifact that a later unit uses. Budget about ten hours per core unit, with a shorter setup. A unit can span more than one calendar week.

| Slot | Core unit | Deliverable |
|---|---|---|
| Setup | [Scala project](weeks/W00-setup.md) | Runnable test and fixture |
| 1 | [Storage](weeks/W01-storage-engines.md) | Scan and write benchmark |
| 2 | [Clocks and failure detection](weeks/W03-clocks.md) | Deterministic message harness |
| 3 | [Replication and failover](weeks/replication-and-failover.md) | Lag and acknowledged-write tests |
| 4 | [Batch dataflow](weeks/W02-mapreduce.md) | Operator graph and reference result |
| 5 | [Shuffle and skew](weeks/W05-shuffle.md) | Exchange and per-partition measurements |
| 6 | [Execution](weeks/W06-query-execution.md) | Row and batch operators |
| 7 | [Planning](weeks/W07-query-planning.md) | Plan rewrite and partitioned join |
| 8 | [Streaming dataflow](weeks/W04-streaming.md) | Windowed counts and deterministic late-event trace |
| 9 | [Snapshots and replay](weeks/W13-fault-tolerance.md) | Consistent-cut and restart tests |
| 10 | [Versioned tables](weeks/W08-ml-pipelines.md) | Two table versions and a SQL diff |
| 11 | [Observability](weeks/W15-observability.md) | Reproducible skew diagnosis |
| 12 | [Deployment](weeks/W14-kubernetes-operators.md) | Two worker processes and a failure diagnosis |
| 13 | [Query-engine capstone](weeks/W16-capstone-platform.md) | Reproducible distributed query and design note |

The filename prefixes are original vault identifiers. Use the table and dashboard for study order.

## Optional ML extension

Start after the versioned-table lab. Read [Hidden Technical Debt in Machine Learning Systems](weeks/ml-feature-contracts.md) and change a feature contract before moving on to [all-reduce](weeks/W09-distributed-training.md), [Ray actors](weeks/W11-actor-model-ray.md), [model parallelism](weeks/W10-parallelism-strategies.md), and [cache-aware serving](weeks/W12-attention.md).

Keep experiment results in [Measurements](MEASUREMENTS.md). [Resources](RESOURCES.md) indexes the reading. The [code map](code/README.md) describes the project you will build.
