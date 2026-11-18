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

   Annotate the two physical plans with their data movement.

**Minimum deliverable:** One semantics-preserving rewrite, a working partitioned join, and annotated Spark plans.

**Handoff:** The streaming unit keeps the partitioner and reuses the aggregate on records that arrive over time.

## Java and Scala comparison (optional, 60–90 minutes)

Finish the Java rewrite in Session 2 first. A *predicate* is the condition in a filter. Moving a predicate below an inner join can reduce the rows reaching the join. A filter on the left input can only move there if the left input has all the columns it needs.

- [ ] **Find the rule in Spark.**

   1. Build two small relations using the PySpark environment from the core lab. Give the left side `(id, active, amount)` rows `(1, true, 300)` and `(1, false, 100)`. Give the right side `(id, limit_cents)` rows `(1, 200)`, `(1, 250)` and `(2, 100)`.
   2. Join on `l.id = r.id` and filter on `l.active AND l.amount > r.limit_cents`. Predict two output rows for ID 1, one for each matching right row. Run the query and save the optimized logical plan from `EXPLAIN EXTENDED`. Look for the left-only condition on the left branch and the condition using both inputs at or above the join.
   3. In [Catalyst's optimizer source](https://github.com/apache/spark/blob/master/sql/catalyst/src/main/scala/org/apache/spark/sql/catalyst/optimizer/Optimizer.scala), search for `PushPredicateThroughJoin`. Use the release tag closest to your Spark version. Read the predicate-splitting logic and the inner-join case. In a few sentences, describe where Spark places each condition and how that compares with your Java rule. If Spark folds a tiny literal relation, read the same rows from two tiny CSV files with explicit schemas and register those DataFrames as temporary views.

- [ ] **Port the rule and its tests.**

   1. In a separate Scala 3 scratch project, model only `Scan`, `Filter`, and `InnerJoin` with cases of an enum or a sealed trait with case classes. Give attributes a relation name as well as a column name so left and right `id` are distinct.
   2. Implement `def pushDownFilters(plan: LogicalPlan): LogicalPlan` with pattern matching. Run the same Java cases in both versions. A left-only predicate moves left, a right-only predicate moves right, and a mixed predicate stays above the join in your toy rule. Test the conditions separately if your Java rule does not split conjunctions.
   3. Check plan shape and confirm that applying the rewrite twice gives the same plan. If reusing result fixtures, compare rows with their counts so duplicate matches survive the check. Include the two-match example above and an empty input. Write down which implementation makes it easier to see why a predicate can move and what each compiler flags if a plan case is missing.

   **Done when:** You have an annotated Spark plan, the same small rewrite tests in Java and Scala, and a short comparison you can refer to when choosing how to model a new operator.
