---
week_number: 4
track: core
title: "Batch dataflow and execution graphs"
status: not-started
---

# Batch dataflow and execution graphs

**Budget:** 10 hours across four study sessions. Shift the sessions to fit your week.

**Start with:** The storage scan and a commerce event fixture.

## Read

Read DDIA 2e Ch. 11 on batch processing, [MapReduce](https://research.google/pubs/mapreduce-simplified-data-processing-on-large-clusters/) on execution, and [Dryad](https://www.microsoft.com/en-us/research/publication/dryad-distributed-data-parallel-programs-from-sequential-building-blocks/) on vertices and channels.

## Sessions

1. [ ] **Draw the graph.** Sketch a scan, filter, and grouped count. Mark the point where rows with the same key must meet.

   Draw `Scan → Filter(event_type = PURCHASE) → CountByCustomer`. Label input partitions and the edge at which the count needs all rows for each customer. Use the W01 scan as a source and compare it with the original CSV fixture.
2. [ ] **Build the reference.** Represent the operators as Java data and interpret them in one process. Keep filter predicates and the aggregation merge deterministic.

   Represent the plan as a sealed interface with record variants. Implement `Map<Long, Long> execute(LogicalPlan plan, Catalog catalog)` for this top-level grouped query. The catalog supplies named scans. Count only qualifying rows, omit absent groups and return an empty map on empty input. Expose a pure count merge by key for later use.
3. [ ] **Compare with Spark.** Run the same group-by in PySpark. Inspect the physical plan and stage view. Mark which edges pipeline rows and which cross a shuffle boundary.

   Execute the same filter and grouped count in PySpark and normalize both outputs to sorted `customer_id,count` pairs. Save the physical plan and identify the exchange and aggregation stages. Assert exact equality on the small fixture.
4. [ ] **Save the workload.** Check in a small uniform fixture and a fixed-seed skewed variant. Record the expected totals and sketch the physical graphs.

   Specify `generate_events(rows, customers, seed, distribution)` for uniform and hot-key fixtures. Assign unique event IDs and keep row count and purchase fraction comparable across distributions. Store generator arguments, result totals and the plan diagram for W05.

**Minimum deliverable:** A Java reference result, a matching Spark result, and an annotated graph.

**Handoff:** W05 replaces the single-process grouping boundary with an Exchange.
