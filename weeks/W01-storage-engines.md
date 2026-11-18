---
week_number: 1
track: core
title: "Storage engines and the cost of a write"
status: not-started
---

# Storage engines and the cost of a write

**Budget:** 10 hours across four study sessions. Shift the sessions to fit your week.

**Start with:** The commerce fixture and test runner from setup.

## Read

Read DDIA 2e Ch. 4 on B-trees, LSM-trees and SSTables. Sketch what happens to a write and a point lookup in each.

## Sessions

1. [ ] **Specify the record.** Use a case class and pure encode/decode functions for a fixed-width benchmark record. Test round trips and truncated input.
2. [ ] **Build two paths.** Write a fixed-slot store and an append log. Show that both return the latest value for a key.
3. [ ] **Measure.** Run the same seeded write workload against both. Use equivalent flush policies, report them, warm up the JVM, and measure point reads as well.
4. [ ] **Export the scan.** Expose a scan of the latest rows and show how compaction changes the number of bytes scanned on repeated updates.

**Minimum deliverable:** A tested scan and a comparison with documented durability settings.

**Handoff:** W03 uses sequenced writes from this store. W02 later uses the scan in the batch graph.
