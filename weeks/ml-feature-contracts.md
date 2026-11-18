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

   Draw the lineage from the event fixture and versioned snapshots to feature readers, trainers and serving consumers. Mark the producer, owner, version and contract at each interface. Include one consumer that binds to a floating default.
2. [ ] **Change one feature.** Change a seven-day event count to a thirty-day count without changing its name or type. Compare feature values for the same customers.

   Implement `compute_features(events, as_of_ms, definition_version)` with seven-day and thirty-day definitions of the same integer column. Use UTC half-open intervals ending at the fixed as-of time. Include an event aged eight days so a value changes with no schema change.
3. [ ] **Pin the inputs.** Record the feature definition version and input table version in a Python training manifest. Verify that both versions can still be read explicitly.

   Save dataset identity, table version, feature definition version and as-of time in the training manifest. Demonstrate a pinned rerun producing the same sorted feature rows after the default moves to the new definition. Use W08 table versions for published artifacts and the original trace for raw event-time features.
4. [ ] **Review the debt.** Identify who silently picks up the change, one coupling the manifest exposes, and one risk it leaves open.

   List which consumer changed, which consumer stayed pinned, and where a hidden dependency could still arise. Attach the before/after rows and explain the connection to the assigned paper.

**Minimum deliverable:** A feature diff, dependency sketch and reproducible training manifest.

**Handoff:** Use the manifest pattern to pin model, data, and shard seeds for the distributed training comparison.
