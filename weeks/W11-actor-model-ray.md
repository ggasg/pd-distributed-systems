---
week_number: 3
track: extension
title: "ML extension: actors and coordination"
status: not-started
---

# ML extension: actors and coordination

**Budget:** 10 hours across four study sessions. Shift the sessions to fit your week.

**Start with:** The model, manifest and reference update from distributed training.

## Read

Read [Ray](https://www.usenix.org/system/files/osdi18-moritz.pdf) §3 and sketch which state belongs to each actor.

## Sessions

1. [ ] **Keep the experiment fixed.** Use the same model, initial weights, shards, batch order and optimizer as the ring run.

   Load the saved initial weights, shard assignment, optimizer and batch order. Run the sequential oracle again and save its per-round update.
2. [ ] **Implement actors.** Use Ray worker actors and a parameter server. Tag gradient submissions with worker and round identifiers.

   Define `compute_gradient(round_id, weights)` and `submit(round_id, worker_id, gradient_sum, sample_count)`. Accept one contribution per worker per round, reject stale or duplicate submissions explicitly and apply the weighted update only after all contributors arrive.
3. [ ] **Compare.** Check a synchronized update against the reference. Compare wall time and transferred bytes under the same workload, and distinguish coordination API differences from topology differences.

   Check gradients and updated weights against the oracle within the same tolerance as W09. Report comparable payload counts and wall time with warmup and synchronization. Describe the change from peer ring to central parameter server when interpreting the results.
4. [ ] **Delay one worker.** Set an explicit round deadline and observe how the coordinator handles a late submission.

   Delay one actor beyond the deadline. Mark the round failed, discard its partial contributions and reject the late message. Assert the model weights did not change and record the timeline.

**Minimum deliverable:** Equivalent updates and a documented failure trace.

**Handoff:** Return to the tested ring for model parallelism.
