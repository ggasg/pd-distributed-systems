---
week_number: 13
track: core
title: "Capstone: a two-worker Java query engine"
status: not-started
---

# Capstone: a two-worker Java query engine

**Budget:** 10 hours across four study sessions. Shift the sessions to fit your week.

**Start with:** The tested operator graph, fixtures and measurements from the core labs.

## Read

Review your Dryad graph and Spark plan notes. Write down the physical plan and the partition you expect to limit completion.

## Sessions

1. [ ] **Choose the query.** Fix one scan, filter, shuffle and grouped aggregate over the commerce fixture. Save the single-worker reference.

   Use the reference purchase-count query and a fixed seed and fixture size. Choose a fixture with qualifying rows in at least two partitions and save sorted `customer_id,count` results in one-worker mode. Record partitions, queue capacity and pre-aggregation setting.
2. [ ] **Run across processes.** Execute the same graph with two workers and verify exact equality with the reference.

   Run the same plan with two distinct worker processes over the deployment protocol. The coordinator runs scan, filter and row routing; workers aggregate assigned partitions. Compare output bytes after sorting and prove both workers contributed. Test an empty partition.
3. [ ] **Measure skew.** Compare uniform and skewed runs using the same scale and worker count. Report partition load, spill bytes and query latency.

   Run equally sized uniform and hot-key inputs at the same two-worker configuration. Keep pre-aggregation off for the first comparison. Report partition bytes, spill bytes, elapsed time and worker imbalance. Attribute coordinator and worker time separately where possible.
4. [ ] **Review Java and FP.** Write a short design note on where pure transformations helped, where local mutation was useful, and what the measurements imply for your next runtime design.

   Deliver a short design note and runnable commands. Point to one pure plan or state function, one local mutable buffer, and one concurrency limit. Explain which measured bottleneck you would investigate before changing the runtime design.

**Minimum deliverable:** A reproducible two-worker query, correct results, a measured bottleneck, and a design note.

**Handoff:** Use the engine and design note as a starting point for your own query-engine project.
