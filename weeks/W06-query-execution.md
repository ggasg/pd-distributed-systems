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

1. [ ] **Set a baseline.** Time the Java row iterator for the existing filter and count. Use DuckDB SQL to check the result and inspect its physical plan.

   Use the existing Java iterator path; keep storage and input fixtures the same between Java runs. Assert that DuckDB produces the same sorted purchase counts and capture its plan. Measure DuckDB separately from the controlled Java comparison.
2. [ ] **Introduce batches.** Add a batch scan and filter. Keep any mutable buffers inside operators and test the same semantics for empty inputs and short final batches.

   Implement `Iterator<EventBatch> scanBatches(int maxRows)` and `EventBatch filterBatch(EventBatch batch, Predicate<Event> predicate)`. Reject nonpositive sizes, preserve input order, and emit the shorter final batch. Keep buffers owned by the operator or document their lifetime explicitly. Test empty input, all rows filtered, and a partial final batch.
3. [ ] **Measure.** Hold the worker count and workload fixed. Warm up the JVM and compare throughput and allocation across batch sizes.

   Measure rows per second for row and batch paths at one worker first, then at several batch sizes. Separate preparation from timed execution, consume the result, and report allocation with the same profiler and settings or record why unavailable.
4. [ ] **Keep the reference.** Make both modes runnable with the same query. Record the comparison conditions and keep the iterator as a correctness oracle.

   Expose row and batch modes through the same entry point and return the same grouped-count type. Add equivalence tests on both skew fixtures so W07 can change physical operators without changing required results.

**Minimum deliverable:** Equivalent row and batch results plus a controlled throughput comparison.

**Handoff:** W07 chooses physical operators using a logical plan and checks each choice against the reference.
