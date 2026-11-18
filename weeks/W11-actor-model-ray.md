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
2. [ ] **Implement actors.** Use Ray worker actors and a parameter server. Tag gradient submissions with worker and round identifiers.
3. [ ] **Compare.** Check a synchronized update against the reference. Compare wall time and transferred bytes under the same workload, and distinguish coordination API differences from topology differences.
4. [ ] **Delay one worker.** Set an explicit round deadline and observe how the coordinator handles a late submission.

**Minimum deliverable:** Equivalent updates and a documented failure trace.

**Handoff:** Return to the tested ring for model parallelism.
