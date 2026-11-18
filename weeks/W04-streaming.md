---
week_number: 8
track: core
title: "Streaming dataflow in Java"
status: not-started
---

# Streaming dataflow in Java

**Budget:** 10 hours across four study sessions. Shift the sessions to fit your week.

**Start with:** The batch partitioner and grouped aggregate, plus the commerce event format.

## Read

Read *Streaming Systems* Ch. 1 on event time and processing time and Ch. 2 on windows, watermarks, triggers and accumulation. Read Sections 1–4 of [The Dataflow Model](https://research.google/pubs/the-dataflow-model-a-practical-approach-to-balancing-correctness-latency-and-cost-in-massive-scale-unbounded-out-of-order-data-processing/). See [the book contents](https://www.oreilly.com/library/view/streaming-systems/9781491983867/preface01.html).

## Sessions

1. [ ] **Predict the outputs.** Write a trace of event arrivals and explicit watermark advances. Mark fixed window boundaries and predict the first output, one late correction, and one event discarded after state expires.

   Keep an ordered trace of event arrivals and watermark advances. Use fixed UTC windows `[start,end)` and a fixed allowed-lateness duration. Predict initial count, corrected count and expired-event outcome. Include an event exactly on a window boundary.
2. [ ] **Implement the transition.** Represent trace inputs with a sealed Java interface with record variants. Write a pure transition from state and input to state and emissions, keyed by customer and event-time window.

   Implement `Transition apply(WindowState state, TraceInput input, WindowConfig config)` with sealed input variants and record outputs. Filter to purchase events, then reuse W05 key routing and the count merge. Tag each emitted row with customer, window start, count and the input offset as its version. Require increasing trace offsets and nondecreasing watermarks.
3. [ ] **Run the trace.** Emit a count at watermark close, replace it when an allowed late event arrives, and evict state at a declared retention boundary. Keep emission and eviction tests separate.

   Emit on the first watermark at or beyond the window end. Apply late events while the watermark is below `end + allowedLateness` and emit a replacement row. Drop events at or beyond expiry. When one watermark jumps across both boundaries, emit the accumulated count before eviction. Test first emission, correction and eviction separately.
4. [ ] **Check the final counts.** Compare against a batch run over the accepted events. Record when the first result arrived and how long state was retained.

   Compare the latest emitted counts with a batch aggregation over accepted events after a final watermark. Keep the trace and output rows as fixtures. Record latency in replay ticks or input offsets and use the same time unit for retention comparisons.

**Minimum deliverable:** One deterministic trace with tested initial counts, corrections and expiry, plus a matching batch reference.

**Handoff:** W13 injects a restart into this exact trace and checks whether the published counts still match.

## Beam and Dataflow comparison (optional, 90–120 minutes locally)

Finish the Java trace first. Beam calls each emission for a window a *pane*. With accumulating panes, a correction includes the previous count plus the newly accepted events.

- [ ] **Replay a late event with Beam Java.**

   1. Create a separate Maven test project from the [Beam Java quickstart](https://beam.apache.org/get-started/quickstart-java/). Use matching versions of the Beam Java SDK and Direct Runner. Run a simple local pipeline before adding windowing. You can finish the local exercise without Google Cloud credentials.
   2. Use [TestStream](https://beam.apache.org/blog/test-stream/) to replay the Java trace. If you need a starting fixture, use one customer, a ten-second window starting at time zero, and five seconds of allowed lateness. Add purchases with event times 1 s and 4 s. Advance the watermark to 10 s, then add a purchase with event time 3 s. Advance to 16 s, then add another purchase with event time 2 s. Finish by advancing the watermark to infinity. Every timestamp here is relative to the same fixed epoch.
   3. Filter purchases, key by customer, apply fixed windows, and count per key. Configure a watermark trigger with a late firing after one element, allowed lateness of five seconds, and accumulating panes. Use `PAssert` to check the emitted counts for the example. Label each output with its pane timing and window. Expect an on-time count of `2` and a late correction to `3`. Assert that no pane has count `4`. Compare by customer and window rather than by arrival order. A final pane may repeat a value.
   4. If you get only one result, check that the on-time watermark advance happens before the late purchase, a late trigger is configured, and the panes accumulate. Save the input schedule and expected outputs as a runnable test.

- [ ] **Follow the graph into a runner.**

   1. Read the [Beam Runner API guide](https://beam.apache.org/contribute/runner-guide/) and [Dataflow pipeline lifecycle](https://docs.cloud.google.com/dataflow/docs/pipeline-lifecycle). Sketch the source, purchase filter, keyed count and sink. Mark where records must be brought together by key and where adjacent transforms might run in one stage.
   2. For a managed run, use a small bounded file and a GCP project with billing enabled. Replace `TestStream` with a file source and keep the purchase filter, event timestamps, window size and keyed count. Compute expected results over every row in that file. Follow the [Dataflow Runner setup](https://beam.apache.org/documentation/runners/dataflow/) to submit it. In the job's execution details, save the stage graph and note one fused group or grouping boundary. Confirm the job finished, the output matches the bounded reference, and you have removed any temporary cloud files you created for the lab.
   3. If you are working locally, annotate the sketch with your predicted stages using the lifecycle guide. Use the Direct Runner test as evidence for lateness behavior. Label the stage sketch as a prediction until you have a managed execution graph to compare.

   **Done when:** You have a passing local trace and an annotated graph. If you ran Dataflow, also save the executed stages and the comparison with the bounded reference.
