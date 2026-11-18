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

1. [ ] **Model the network.** Use Scala enums for messages and immutable node state. Write a pure transition from state plus input to new state plus outbound messages.
2. [ ] **Schedule deliveries.** Use a logical test clock and deterministic event queue. Implement Lamport ticks and add vector timestamps for the concurrent-write example.
3. [ ] **Delay a healthy peer.** Add heartbeats and a suspicion timeout. Build two traces that look identical to the detector for the first N ticks, with a crashed peer in one and a delayed peer in the other.
4. [ ] **Save both traces.** Assert the causal ordering and the false suspicion, and record how changing the timeout changes detection delay.

**Minimum deliverable:** A reusable simulator with causal and timeout tests.

**Handoff:** The replication lab uses the same clock, messages, and failure schedules.
