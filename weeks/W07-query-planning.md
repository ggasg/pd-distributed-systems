---
week_number: 7
track: core
title: "Query planning and data movement"
status: not-started
---

# Query planning and data movement

**Budget:** 10 hours across four study sessions. Shift the sessions to fit your week.

**Start with:** The same query, data, and physical operators from W06.

## Read

Read the Catalyst design and physical planning in [Spark SQL](https://people.csail.mit.edu/matei/papers/2015/sigmod_spark_sql.pdf). Revisit the declarative-query section of DDIA 2e Ch. 3.

## Sessions

1. [ ] **Separate the plans.** Represent logical operators as Scala data and display physical operators with Exchange boundaries explicit.
2. [ ] **Rewrite safely.** Implement one filter pushdown rule for an inner join. Check pre- and post-rewrite results, including duplicate keys and empty inputs.
3. [ ] **Follow the bytes.** Run a partitioned join in Scala. Compare Spark physical plans for a small and a larger build side. Record the configuration and inspect the actual executed plan if adaptive execution changes it.
4. [ ] **Explain the choice.** Record where data moves and why. Keep a plan test and a result test for the Scala join.

**Minimum deliverable:** One semantics-preserving rewrite, a working partitioned join, and annotated Spark plans.

**Handoff:** The streaming unit keeps the partitioner and reuses the aggregate on records that arrive over time.
