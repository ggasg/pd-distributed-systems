---
week_number: 1
track: extension
title: "ML extension: feature contracts and technical debt"
status: not-started
---

# ML extension: feature contracts and technical debt

**Budget:** 6 hours across four study sessions. Shift the sessions to fit your week.

**Start with:** The versioned tables and SQL diff from the core.

## Read

Read Sculley et al., [Hidden Technical Debt in Machine Learning Systems](https://papers.nips.cc/paper_files/paper/2015/file/86df7dcfd896fcaf2674f757a2463eba-Paper.pdf), in full. Focus on entanglement and CACE, undeclared consumers, data dependencies, and glue code.

## Sessions

1. [ ] **Trace the consumers.** Draw the path from event tables through feature generation to training and serving. Mark who reads each output and which interfaces are implicit.
2. [ ] **Change one feature.** Change a seven-day event count to a thirty-day count without changing its name or type. Compare feature values for the same customers.
3. [ ] **Pin the inputs.** Record the feature definition version and input table version in a Python training manifest. Verify that both versions can still be read explicitly.
4. [ ] **Review the debt.** Identify who silently picks up the change, one coupling the manifest exposes, and one risk it leaves open.

**Minimum deliverable:** A feature diff, dependency sketch and reproducible training manifest.

**Handoff:** Use the manifest pattern to pin model, data, and shard seeds for the distributed training comparison.
