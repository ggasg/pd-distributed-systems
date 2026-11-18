---
week_number: 3
track: core
title: "Replication, lag, and failover"
status: not-started
---

# Replication, lag, and failover

**Budget:** 10 hours across four study sessions. Shift the sessions to fit your week.

**Start with:** The sequenced write log and deterministic message harness.

## Read

Read DDIA 2e Ch. 6 on replication lag and failover, and Ch. 10 on linearizability and consensus. Read [Raft](https://raft.github.io/raft.pdf) §5 for terms, election restrictions, and log commitment.

## Sessions

1. [ ] **Predict lost writes.** Trace a leader crash before a follower receives its last write. Mark the client acknowledgment and each replica position.
2. [ ] **Replicate the log.** Add a leader and two followers. Simulate asynchronous acknowledgments, then require the leader and both followers to persist a write before acknowledgment.
3. [ ] **Test promotion.** Use scripted promotion with epoch checks at the write boundary. Replay the crash and a stale read under both policies. Measure the write delay when a follower lags.
4. [ ] **State the guarantee.** Explain what this simulation assumes about promotion and fencing, and what an election and commit protocol would still have to provide.

**Minimum deliverable:** Tests that reproduce one lost acknowledged write under asynchronous replication and verify recovery under the all-replica acknowledgment policy and scripted promotion.

**Handoff:** Keep the failure harness for recovery. The batch unit returns to the storage scan and asks how to distribute computation over its rows.
