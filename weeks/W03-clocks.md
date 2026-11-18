---
week_number: 2
track: core
title: "Clocks, causality, and failure detection"
status: not-started
---

# Clocks, causality, and failure detection

**Budget:** 10 hours across four study sessions. Shift the sessions to fit your week.

**Start with:** Sequenced writes from the storage lab.

## Read

Read DDIA 2e Ch. 9 on unreliable clocks and unbounded delays. Study [Lamport on event ordering](https://dl.acm.org/doi/10.1145/359545.359563). Sketch two concurrent writes and one causal chain.

## Sessions

1. [ ] **Model the network.** Use sealed Java interfaces and records for messages and immutable node state. Write a pure transition from state plus input to new state plus outbound messages.

   Define sealed message and input interfaces with records for arrivals, local writes and ticks. Implement `Transition step(NodeState state, NodeInput input)`, returning the new state and an ordered list of outgoing messages. Do not read a wall clock inside it. Use defensive copies for collections stored in state.
2. [ ] **Schedule deliveries.** Use a logical test clock and deterministic event queue. Implement Lamport ticks and add vector timestamps for the concurrent-write example.

   Implement `Trace run(Scenario scenario)` over a queue ordered by logical tick and insertion number. On receive, Lamport time becomes `max(local, received) + 1`; vector time takes the componentwise maximum and increments the receiver. Assert send-before-receive and incomparable vectors for independent writes.
3. [ ] **Delay a healthy peer.** Add heartbeats and a suspicion timeout. Build two traces that look identical to the detector for the first N ticks, with a crashed peer in one and a delayed peer in the other.

   Track `lastHeardTick` per peer. Emit a suspicion on crossing the configured threshold and clear it on a later heartbeat. Compare identical observation prefixes for crashed and delayed peers. Use the same input ordering for ties at a timeout boundary.
4. [ ] **Save both traces.** Assert the causal ordering and the false suspicion, and record how changing the timeout changes detection delay.

   Check in both schedules and their assertions. Report suspicion tick and any recovery tick at two timeout values. Expose delay, drop and stop-node events for the next lab.

**Minimum deliverable:** A reusable simulator with causal and timeout tests.

**Handoff:** The replication lab uses the same clock, messages, and failure schedules.
