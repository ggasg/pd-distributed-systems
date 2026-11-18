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
2. [ ] **Publish tables.** Use Python to write the two versions to a local Delta table. Inspect the added and removed files in the commit log.
3. [ ] **Query both versions.** Read each version explicitly and use SQL to compare rows. Test the reader across the schema change.
4. [ ] **Trace publication.** Draw the boundary where written files become visible to a reader. Describe what a failed writer leaves behind and what version a repeatable job must record.

**Minimum deliverable:** Two queryable versions, a SQL diff, and an annotated commit log.

**Handoff:** The ML extension treats versioned tables as feature inputs. The core operations units instrument and deploy the Scala query engine.
