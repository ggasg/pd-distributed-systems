---
week_number: 0
track: core
title: "Getting the Java 21 project running"
status: not-started
---

# Getting the Java 21 project running

**Budget:** 4 hours across four study sessions. Shift the sessions to fit your week.

**Start with:** A local checkout and an Obsidian vault.

## Read

Skim DDIA 2e Ch. 1 for operational versus analytical workloads and Ch. 2 for reliability and latency.

## Sessions

1. [ ] **Set up the tools.** Use [SETUP](../SETUP.md) to install Zulu 21 CA and Maven. Create the project structure described in the [code map](../code/README.md).

   Create a Maven project with compiler release 21, UTF-8, and JUnit 5 test discovery. Run `mvn test` from `code/query-engine/` and confirm at least one test ran. Record `java -version` and `mvn -version`, and confirm Maven is using Zulu 21.
2. [ ] **Write a fixture.** Check in a tiny synthetic commerce event file and its expected grouped result. Use integer cents in order records.

   Create `events.csv` with the shared header and three rows: customer 10 has two purchases, customer 20 has one view. Use positive unique event IDs and UTC epoch milliseconds. Save `expected-counts.csv` with the single purchase count `10,2`.
3. [ ] **Run it twice.** Write a Java test that loads the fixture and checks the expected grouped count. Run it twice from a fresh process.

   Implement `List<Event> readEvents(Path path)` as a fixture loader. Require the declared header and five fields, and report malformed rows with their line number. For this fixture, reject missing fields and unknown event types. Assert that filtering to purchases and grouping by customer yields the saved result.
4. [ ] **Record the invocation.** Record the JDK and Maven versions and the test command in the project README.

   Document the working directory and command for a clean test run. Preserve the CSV and result under `src/test/resources` so the next lab can load the same events without recreating them.

**Minimum deliverable:** The fixture test runs from a fresh checkout with one documented command.

**Handoff:** W01 uses the fixture to exercise the record encoder and scan.
