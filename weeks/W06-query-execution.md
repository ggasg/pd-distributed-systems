---
week_number: 6
track: core
title: "Query execution and batching"
status: not-started
---

# Query execution and batching

**Budget:** 10 hours across four study sessions. Shift the sessions to fit your week.

**Start with:** The scan and Exchange from earlier units.

## Read

Read DDIA 2e Ch. 3 on declarative queries and [MonetDB/X100](https://www.cidrdb.org/cidr2005/papers/P19.pdf) on execution granularity.

## Sessions

1. [ ] **Set a baseline.** Time the Scala row iterator for the existing filter and count. Use DuckDB SQL to check the result and inspect its physical plan.
2. [ ] **Introduce batches.** Add a batch scan and filter. Keep any mutable buffers inside operators and test the same semantics for empty inputs and short final batches.
3. [ ] **Measure.** Hold the worker count and workload fixed. Warm up the JVM and compare throughput and allocation across batch sizes.
4. [ ] **Keep the reference.** Make both modes runnable with the same query. Record the comparison conditions and keep the iterator as a correctness oracle.

**Minimum deliverable:** Equivalent row and batch results plus a controlled throughput comparison.

**Handoff:** W07 chooses physical operators using a logical plan and checks each choice against the reference.
