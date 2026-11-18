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
2. [ ] **Implement the collective.** Implement reduce-scatter and all-gather on NumPy arrays. Test uneven chunk splits and compare to a direct sum.
3. [ ] **Train on two workers.** Reuse the model and split the batch. Compare one update to the reference, then run several training steps.
4. [ ] **Compare bytes.** Count transferred bytes for a direct all-to-all implementation and the ring at 2, 4 and 8 workers.

**Minimum deliverable:** A matching reference update and a byte comparison with both protocols measured.

**Handoff:** Use the same model, shards and update rule in the Ray comparison.
