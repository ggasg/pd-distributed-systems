---
week_number: 10
track: core
title: "Transactions and versioned tables"
status: not-started
---

# Transactions and versioned tables

**Budget:** 10 hours across four study sessions. Shift the sessions to fit your week.

**Start with:** The final and corrected keyed results from the recovery lab.

## Read

Read DDIA 2e Ch. 5 on encoding and evolution and Ch. 8 on atomicity and isolation. Read the transaction-log design in the [Delta Lake paper](https://www.vldb.org/pvldb/vol13/p3411-armbrust.pdf).

## Sessions

1. [ ] **Define versions.** Decide which keyed updates go into v1 and v2 and predict the row-level diff. Choose a schema change and state which readers should still work.

   Read the exported W13 upserts. Pick a cutoff before the correction for the first snapshot and after for the second. Collapse to the greatest version per `(customer_id,window_start_ms)`. Add a nullable `source_version` field in the second schema and predict which named projections still work.
2. [ ] **Publish tables.** Use Python to write the two versions to a local Delta table. Inspect the added and removed files in the commit log.

   Implement `publish_snapshot(path, rows, schema) -> int` in Python and return the actual Delta commit number. Publish each complete snapshot in a separate commit and record the resulting versions. Inspect both commits for file actions and metadata changes.
3. [ ] **Query both versions.** Read each version explicitly and use SQL to compare rows. Test the reader across the schema change.

   Implement `read_snapshot(path, version)` with explicit time travel. Use a full outer join in SQL on the customer and window key to find additions, removals and changed counts. Verify the old projection against both versions and show whether the new column is available on the old version.
4. [ ] **Trace publication.** Draw the boundary where written files become visible to a reader. Describe what a failed writer leaves behind and what version a repeatable job must record.

   Annotate the commit that makes the second snapshot visible. Record the table location, exact version numbers, and CSV export schema. Keep the original event fixture or trace alongside them for the ML feature-contract lab.

**Minimum deliverable:** Two queryable versions, a SQL diff, and an annotated commit log.

**Handoff:** The ML extension treats versioned tables as feature inputs. The core operations units instrument and deploy the Java query engine.
