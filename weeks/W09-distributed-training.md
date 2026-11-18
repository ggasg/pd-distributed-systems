---
week_number: 2
track: extension
title: "ML extension: distributed training and all-reduce"
status: not-started
---

# ML extension: distributed training and all-reduce

**Budget:** 10 hours across four study sessions. Shift the sessions to fit your week.

**Start with:** The run-manifest pattern and Python environment from the feature-contract lab.

## Read

Read [Horovod](https://arxiv.org/abs/1802.05799) §3 and the design overview of [PyTorch Distributed](https://arxiv.org/abs/2006.15704).

## Sessions

1. [ ] **Pin a baseline.** Use a small PyTorch linear classifier and fixed data and shard seeds. Make one full-batch gradient your reference.

   Create a deterministic fixture with initial weights, labeled minibatch, shard assignments and optimizer settings. Save the single-process full-batch gradient and one parameter update as the oracle.
2. [ ] **Implement the collective.** Implement reduce-scatter and all-gather on NumPy arrays. Test uneven chunk splits and compare to a direct sum.

   Implement `ring_allreduce(local_arrays, transport)` as reduce-scatter followed by all-gather. Return one copy of the summed tensor at every rank. Define contiguous chunk boundaries for lengths not divisible by rank count; test against a direct sum and tag messages by round, phase and chunk.
3. [ ] **Train on two workers.** Reuse the model and split the batch. Compare one update to the reference, then run several training steps.

   Run two distinct workers and sum local gradient sums and sample counts. Divide by the global sample count before updating so uneven shards match the oracle. Check gradients and weights with a declared tolerance.
4. [ ] **Compare bytes.** Count transferred bytes for a direct all-to-all implementation and the ring at 2, 4 and 8 workers.

   Count payload bytes once per send for the ring and direct all-to-all at ranks 2, 4 and 8. Use the same tensor, dtype and instrumentation. Publish manifest, initial weights and shard split for the actor lab.

**Minimum deliverable:** A matching reference update and a byte comparison with both protocols measured.

**Handoff:** Use the same model, shards and update rule in the Ray comparison.
