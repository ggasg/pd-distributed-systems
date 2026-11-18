---
week_number: 5
track: core
title: "Partitioning, shuffle, and skew"
status: not-started
---

# Partitioning, shuffle, and skew

**Budget:** 10 hours across four study sessions. Shift the sessions to fit your week.

**Start with:** The batch graph and uniform and skewed fixtures.

## Read

Read DDIA 2e Ch. 7 on hash and range sharding and hot spots. Revisit the data movement section of the MapReduce paper.

## Sessions

1. [ ] **Route rows.** Implement pure hash partitioning. Test that equal keys route to the same partition and every row appears once.

   Implement `int partition(long key, int partitions)` with `Math.floorMod(Long.hashCode(key), partitions)`. Reject nonpositive partition counts. Test repeatability, negative keys and identical keys from separate input splits.
2. [ ] **Build Exchange.** Add bounded queues and spill files between map and reduce tasks. Validate the same grouped result at one and two workers.

   Implement `executeShuffle(LogicalPlan query, ExchangeConfig config)` for the W02 count. Keep partition count separate from worker count. Route qualifying rows by customer into bounded queues; spill overflow to files scoped to one execution. Read each spilled row once before finalizing a partition. Bound in-flight data, propagate worker errors and clean up files on success or failure. Force spill with a tiny queue and compare one and two workers to W02.
3. [ ] **Find the straggler.** Hold input size constant, vary the key distribution, and record bytes and duration per partition. Repeat the group-by in Spark and identify the skewed tasks.

   Log per-partition input rows, encoded payload bytes, processing time, queue wait and spill bytes. Use a monotonic timer for durations. Compare uniform and hot-key data at fixed scale and annotate the Spark task view for the same logical query.
4. [ ] **Change the workload.** Try pre-aggregation for this mergeable count and compare the busiest partition before and after. Record queue wait and spill bytes for later instrumentation.

   Pre-aggregate per input split and shuffle `(customerId, partialCount)`. Route the partial by the same key and sum partials at the destination. Check identical results and report how the hottest partition's traffic and duration change. Keep the row-shuffle path selectable for the capstone.

**Minimum deliverable:** A matching reference result and a measured skew bottleneck with per-partition evidence.

**Handoff:** W06 keeps the workload fixed while changing the operator execution model.
