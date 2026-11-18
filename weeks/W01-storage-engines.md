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

1. [ ] **Specify the record.** Use a Java record and pure encode/decode functions for a fixed-width benchmark record. Test round trips and truncated input.

   Use `record Write(long sequence, Event event)`. Implement `byte[] encode(Write write)` and `Write decode(byte[] bytes)` with an explicit byte order and stable numeric event-type codes. Reject wrong record lengths and unknown codes. Test round trips and a truncated record.
2. [ ] **Build two paths.** Write a fixed-slot store and an append log. Show that both return the latest value for a key.

   Both stores expose `void put(Write write)`, `Optional<Event> get(long eventId)`, and `List<Event> scanLatest()`. Accept strictly increasing sequence numbers within a store. Fix a dense positive ID range for the fixed-slot file. For the append log, rebuild the latest-offset index on reopen. Test overwrites, a missing ID and a reopen. Return scan rows in event-ID order.
3. [ ] **Measure.** Run the same seeded write workload against both. Use equivalent flush policies, report them, warm up the JVM, and measure point reads as well.

   Run the same seeded sequence of repeated updates against both stores. Use the same explicit `FileChannel.force` policy, JVM warmup, input count and lookup IDs. Report median write and point-read rates over repeated runs and include final file size.
4. [ ] **Export the scan.** Expose a scan of the latest rows and show how compaction changes the number of bytes scanned on repeated updates.

   Implement offline `compact(Path source, Path destination)` for the append log. Write only the latest record per ID in ascending sequence order to a new file. Keep the maximum sequence so the compacted log accepts only newer writes. Compare reopened scans and point lookups before and after. Report bytes read while rebuilding each index.

**Minimum deliverable:** A tested scan and a comparison with documented durability settings.

**Handoff:** W03 uses sequenced writes from this store. W02 later uses the scan in the batch graph.
