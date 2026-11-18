---
week_number: 11
track: core
title: "Observing the exchange"
status: not-started
---

# Observing the exchange

**Budget:** 10 hours across four study sessions. Shift the sessions to fit your week.

**Start with:** The uniform and skewed workloads and the earlier per-partition metrics.

## Read

Revisit DDIA 2e Ch. 2 on tail latency. Read [The Tail at Scale](https://research.google/pubs/the-tail-at-scale/) and explain which worker determines query completion in your graph.

## Sessions

1. [ ] **Choose signals.** Instrument partition bytes, queue wait, spill bytes and query duration. Distinguish per-stage timings from end-to-end latency.

   Expose per-partition row/byte counts, queue wait, spill bytes and query duration. Include partition and run identifiers in the lab report, with separate units for bytes and elapsed time. Check a tiny forced-spill run against its measured file sizes.
2. [ ] **Reproduce skew.** Run uniform and skewed inputs at the same worker count and visualize per-partition work.

   Run uniform and hot-key workloads at equal input size and worker count. Keep pre-aggregation fixed. Save per-partition load and elapsed time and identify whether the busiest partition is also the slowest.
3. [ ] **Check a histogram.** Record a raw timing sample. Compare its p99 with histogram estimates using two bucket choices and explain the difference.

   Record individual end-to-end durations. Compute nearest-rank p99 and the bucket containing that percentile under fine and coarse bucket boundaries. Explain how the available sample size limits the estimate.
4. [ ] **Write the diagnosis.** Save the workload seed, metrics, and one diagnosis that another person can reproduce.

   Save commands and a CSV or JSON run report with workload seed, worker count, queue capacity, spill bytes and latency. Document one hypothesis supported by the measurements for the deployment and capstone.

**Minimum deliverable:** A reproducible run whose metrics locate the skew bottleneck.

**Handoff:** W14 runs the same engine and workload across two worker processes.
