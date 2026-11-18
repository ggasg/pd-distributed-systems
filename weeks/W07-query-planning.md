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

1. [ ] **Separate the plans.** Represent logical operators as Java data and display physical operators with Exchange boundaries explicit.

   Extend the logical plan with `InnerJoin` and named attribute references. Return a physical plan from `PhysicalPlan lower(LogicalPlan plan)`; include exchanges on both sides of the partitioned join. Implement `String explain(PhysicalPlan plan)` with stable operator and key labels.
2. [ ] **Rewrite safely.** Implement one filter pushdown rule for an inner join. Check pre- and post-rewrite results, including duplicate keys and empty inputs.

   Implement `LogicalPlan pushDownFilters(LogicalPlan plan)`. Push a pure predicate above an inner join into the referenced input only when all referenced attributes belong to that side. Leave mixed-input predicates in place. Test plan shape, idempotence, duplicate-key multiplicity and empty inputs. Compare results as multisets.
3. [ ] **Follow the bytes.** Run a partitioned join in Java. Compare Spark physical plans for a small and a larger build side. Record the configuration and inspect the actual executed plan if adaptive execution changes it.

   Implement a partitioned equijoin using W05 routing on both sides. For each key return every matching pair. Test against a nested-loop oracle with duplicate and unmatched keys. In Spark, vary build-side size or the configured broadcast threshold and save both plans and any adaptive changes.
4. [ ] **Explain the choice.** Record where data moves and why. Keep a plan test and a result test for the Java join.

   Annotate the two physical plans with their data movement. As a short optional FP comparison, port only the rewrite and its plan tests to Scala 3 in a separate scratch project. Use the same cases and record which language made the rule easier to read.

**Minimum deliverable:** One semantics-preserving rewrite, a working partitioned join, and annotated Spark plans.

**Handoff:** The streaming unit keeps the partitioner and reuses the aggregate on records that arrive over time.
