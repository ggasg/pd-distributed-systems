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
2. [ ] **Snapshot the graph.** Add markers to the simulator and record local state and in-flight messages on FIFO channels. Exercise a trace with an in-flight event.
3. [ ] **Recover the stream.** Checkpoint input progress and operator state on a single worker. Force a crash after publishing a count but before checkpointing progress. Replay into a sink that replaces a keyed result by deterministic update version.
4. [ ] **Verify the boundary.** Compare recovered and uninterrupted final tables. Explain the FIFO assumption in the simulator and which sink assumption makes this replay safe.

**Minimum deliverable:** A consistent-cut test and a restart test that preserves the expected final table.

**Handoff:** W08 takes the recovered keyed updates and publishes explicit table versions.
