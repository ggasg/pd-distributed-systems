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
2. [ ] **Build Exchange.** Add bounded queues and spill files between map and reduce tasks. Validate the same grouped result at one and two workers.
3. [ ] **Find the straggler.** Hold input size constant, vary the key distribution, and record bytes and duration per partition. Repeat the group-by in Spark and identify the skewed tasks.
4. [ ] **Change the workload.** Try pre-aggregation for this mergeable count and compare the busiest partition before and after. Record queue wait and spill bytes for later instrumentation.

**Minimum deliverable:** A matching reference result and a measured skew bottleneck with per-partition evidence.

**Handoff:** W06 keeps the workload fixed while changing the operator execution model.
