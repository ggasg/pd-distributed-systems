---
week_number: 4
track: extension
title: "ML extension: model parallelism"
status: not-started
---

# ML extension: model parallelism

**Budget:** 10 hours across four study sessions. Shift the sessions to fit your week.

**Start with:** The tested collective from distributed training.

## Read

Read [Megatron-LM](https://arxiv.org/abs/1909.08053) on tensor parallelism and [GPipe](https://arxiv.org/abs/1811.06965) on microbatches. Skim [ZeRO](https://arxiv.org/abs/1910.02054) for optimizer-state partitioning.

## Sessions

1. [ ] **Make a reference.** Define a deterministic two-layer NumPy block.

   Implement `reference(x, w1, w2)` for a deterministic two-layer block with an elementwise activation between matrices. Record shapes, dtype, seed and tolerance.
2. [ ] **Split a layer.** Implement column then row partitioning and match the unsplit output using the tested collective.

   Split the first weight matrix by output columns and the second by matching input rows. Apply the activation to each local hidden slice, compute partial outputs and sum them with W09's collective. Match the reference output and record bytes.
3. [ ] **Split the stages.** Simulate a two-stage pipeline and measure idle slots as microbatch count changes.

   Implement `simulate_pipeline(stage_ticks, microbatches)` returning a forward-only fill-and-drain schedule. Each stage executes at most one microbatch per tick and stage two waits for stage one's output for that microbatch. Include equal-duration and imbalanced stages.
4. [ ] **Compare the costs.** Record the collective traffic and the pipeline idle fraction. Identify what would dominate for your workload.

   Report idle slots divided by total stage slots at several microbatch counts. Distinguish simulated scheduling occupancy from measured transport traffic and state where each limits the toy workload.

**Minimum deliverable:** A tensor-parallel result that matches the reference and a measured pipeline schedule.

**Handoff:** The serving lab uses attention to study a different trade-off between memory and computation.
