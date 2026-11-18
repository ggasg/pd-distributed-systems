---
week_number: 9
track: core
title: "Snapshots and replay"
status: not-started
---

# Snapshots and replay

**Budget:** 10 hours across four study sessions. Shift the sessions to fit your week.

**Start with:** The streaming trace and state transitions, plus the message harness from clocks and replication.

## Read

Read [Chandy and Lamport](https://lamport.azurewebsites.net/pubs/chandy.pdf) on consistent cuts and FIFO channels. Read *Streaming Systems* Ch. 5 on sources, sinks, and side effects. Skim [MillWheel](https://research.google/pubs/millwheel-fault-tolerant-stream-processing-at-internet-scale/) for persistent state.

## Sessions

1. [ ] **Draw the failure.** Mark input progress, operator state and output commits on the trace. Predict the two crash cases around publishing a result.

   Annotate the W04 trace with input offsets, state changes and sink writes. Predict failures immediately before and after sink publication when the durable input position still reflects the previous checkpoint.
2. [ ] **Snapshot the graph.** Add markers to the simulator and record local state and in-flight messages on FIFO channels. Exercise a trace with an in-flight event.

   Extend the W03 simulator with snapshot markers and FIFO channels. Save local state on the first marker, then record messages on each other incoming channel until its marker arrives. Assert that a message sent before the recorded cut and delivered after it appears in the channel state.
3. [ ] **Recover the stream.** Checkpoint input progress and operator state on a single worker. Force a crash after publishing a count but before checkpointing progress. Replay into a sink that replaces a keyed result by deterministic update version.

   Define `Checkpoint(nextOffset, windowState)` and save it atomically for the single-worker replay. Reopen the last good checkpoint after injecting a crash, then replay from its next offset. Upsert sink rows by `(customerId, windowStart)` only when the new version is greater; identical versions must have identical payloads. Persist the sink across restart.
4. [ ] **Verify the boundary.** Compare recovered and uninterrupted final tables. Explain the FIFO assumption in the simulator and which sink assumption makes this replay safe.

   Assert recovered and uninterrupted tables agree in both crash cases, including the late correction. Export ordered upserts from the uninterrupted run as CSV with `customer_id,window_start_ms,count,version`. Explain which experiment records distributed channel state and which actually restarts from durable state.

**Minimum deliverable:** A consistent-cut test and a restart test that preserves the expected final table.

**Handoff:** W08 takes the recovered keyed updates and publishes explicit table versions.
