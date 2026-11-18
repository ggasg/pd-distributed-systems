---
week_number: 8
track: core
title: "Streaming dataflow in Scala"
status: not-started
---

# Streaming dataflow in Scala

**Budget:** 10 hours across four study sessions. Shift the sessions to fit your week.

**Start with:** The batch partitioner and grouped aggregate, plus the commerce event format.

## Read

Read *Streaming Systems* Ch. 1 on event time and processing time and Ch. 2 on windows, watermarks, triggers and accumulation. Read Sections 1–4 of [The Dataflow Model](https://research.google/pubs/the-dataflow-model-a-practical-approach-to-balancing-correctness-latency-and-cost-in-massive-scale-unbounded-out-of-order-data-processing/). See [the book contents](https://www.oreilly.com/library/view/streaming-systems/9781491983867/preface01.html).

## Sessions

1. [ ] **Predict the outputs.** Write a trace of event arrivals and explicit watermark advances. Mark fixed window boundaries and predict the first output, one late correction, and one event discarded after state expires.
2. [ ] **Implement the transition.** Represent trace inputs with a Scala enum. Write a pure transition from state and input to state and emissions, keyed by customer and event-time window.
3. [ ] **Run the trace.** Emit a count at watermark close, replace it when an allowed late event arrives, and evict state at a declared retention boundary. Keep emission and eviction tests separate.
4. [ ] **Check the final counts.** Compare against a batch run over the accepted events. Record when the first result arrived and how long state was retained.

**Minimum deliverable:** One deterministic trace with tested initial counts, corrections and expiry, plus a matching batch reference.

**Handoff:** W13 injects a restart into this exact trace and checks whether the published counts still match.
