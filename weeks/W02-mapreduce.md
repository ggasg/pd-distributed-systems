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
2. [ ] **Build the reference.** Represent the operators as Scala data and interpret them in one process. Keep filter predicates and the aggregation merge deterministic.
3. [ ] **Compare with Spark.** Run the same group-by in PySpark. Inspect the physical plan and stage view. Mark which edges pipeline rows and which cross a shuffle boundary.
4. [ ] **Save the workload.** Check in a small uniform fixture and a fixed-seed skewed variant. Record the expected totals and sketch the physical graphs.

**Minimum deliverable:** A Scala reference result, a matching Spark result, and an annotated graph.

**Handoff:** W05 replaces the single-process grouping boundary with an Exchange.
