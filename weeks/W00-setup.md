---
week_number: 0
track: core
title: "Getting the Scala project running"
status: not-started
---

# Getting the Scala project running

**Budget:** 4 hours across four study sessions. Shift the sessions to fit your week.

**Start with:** A local checkout and an Obsidian vault.

## Read

Skim DDIA 2e Ch. 1 for operational versus analytical workloads and Ch. 2 for reliability and latency.

## Sessions

1. [ ] **Set up the tools.** Use [SETUP](../SETUP.md) to install a JDK and sbt. Create the project structure described in the [code map](../code/README.md).
2. [ ] **Write a fixture.** Check in a tiny synthetic commerce event file and its expected grouped result. Use integer cents in order records.
3. [ ] **Run it twice.** Write a Scala test that loads the fixture and checks the expected grouped count. Run it twice from a fresh process.
4. [ ] **Record the invocation.** Record the JDK, Scala and sbt versions and the test command in the project README.

**Minimum deliverable:** The fixture test runs from a fresh checkout with one documented command.

**Handoff:** W01 uses the fixture to exercise the record encoder and scan.
