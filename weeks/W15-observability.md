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
2. [ ] **Reproduce skew.** Run uniform and skewed inputs at the same worker count and visualize per-partition work.
3. [ ] **Check a histogram.** Record a raw timing sample. Compare its p99 with histogram estimates using two bucket choices and explain the difference.
4. [ ] **Write the diagnosis.** Save the workload seed, metrics, and one diagnosis that another person can reproduce.

**Minimum deliverable:** A reproducible run whose metrics locate the skew bottleneck.

**Handoff:** W14 runs the same engine and workload across two worker processes.
