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

   Write a three-node scenario with a replicated prefix and one new write. Record when the leader persists it, when the client sees acknowledgment, and when each follower would persist it. Crash the leader before one chosen follower receives the write.
2. [ ] **Replicate the log.** Add a leader and two followers. Simulate asynchronous acknowledgments, then require the leader and both followers to persist a write before acknowledgment.

   Extend `step` with `Append(epoch, write)` and `Persisted(epoch, sequence, replicaId)`. Model disk persistence as a distinct event and acknowledge once per sequence. In asynchronous mode wait for leader persistence. In all-replica mode require distinct persistence confirmations from all three nodes in the current epoch.
3. [ ] **Test promotion.** Use scripted promotion with epoch checks at the write boundary. Replay the crash and a stale read under both policies. Measure the write delay when a follower lags.

   Implement scripted `promote(nodeId, newEpoch)` and return the last persisted sequence alongside follower reads. After promotion reject old-epoch writes and acknowledgments at the relevant boundaries. Compare promotion after an asynchronous acknowledgment with promotion after an all-replica acknowledgment. Also test the same early crash under the all-replica policy, where the client has no acknowledgment.
4. [ ] **State the guarantee.** Explain what this simulation assumes about promotion and fencing, and what an election and commit protocol would still have to provide.

   Produce a table of client acknowledgment, promoted replica state, stale-read result and acknowledgment delay. Explain the availability cost of requiring all replicas and state which fencing and promotion decisions the simulator scripts externally.

**Minimum deliverable:** Tests that reproduce one lost acknowledged write under asynchronous replication and verify recovery under the all-replica acknowledgment policy and scripted promotion.

**Handoff:** Keep the failure harness for recovery. The batch unit returns to the storage scan and asks how to distribute computation over its rows.
