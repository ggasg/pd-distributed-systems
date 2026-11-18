---
week_number: 13
track: core
title: "Capstone: a two-worker Scala query engine"
status: not-started
---

# Capstone: a two-worker Scala query engine

**Budget:** 10 hours across four study sessions. Shift the sessions to fit your week.

**Start with:** The tested operator graph, fixtures and measurements from the core labs.

## Read

Review your Dryad graph and Spark plan notes. Write down the physical plan and the partition you expect to limit completion.

## Sessions

1. [ ] **Choose the query.** Fix one scan, filter, shuffle and grouped aggregate over the commerce fixture. Save the single-worker reference.
2. [ ] **Run across processes.** Execute the same graph with two workers and verify exact equality with the reference.
3. [ ] **Measure skew.** Compare uniform and skewed runs using the same scale and worker count. Report partition load, spill bytes and query latency.
4. [ ] **Review Scala and FP.** Write a short design note on where pure transformations helped, where local mutation was useful, and what the measurements imply for your next runtime design.

**Minimum deliverable:** A reproducible two-worker query, correct results, a measured bottleneck, and a design note.

**Handoff:** Use the engine and design note as a starting point for your own query-engine project.
