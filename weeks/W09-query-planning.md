---
week_number: 9
status: not-started
---

# W09: Rule-Based Query Planning in Scala

> **Arc:** Streaming and Dataflow · **Language:** Scala

## What you'll build
A toy version of Spark's Catalyst optimizer: a logical query plan represented as an algebraic data type (Scala case classes), a small set of rewrite rules expressed as pattern-matching partial functions, and a generic `transform` combinator that applies rules recursively across the tree. This is the capstone to Arc 2 — W05–W08 built the individual operators (windowed aggregation, dataflow progress tracking, incremental views, vectorized execution); this week builds the thing that arranges and rewrites those operators into an optimized plan, the way a real MPP engine does.

**Note on why Scala, specifically:** this isn't an arbitrary FP-language choice. Spark itself is written in Scala, and Catalyst — its query optimizer — is genuinely built the way this week has you build your toy version: case classes for plan nodes, pattern matching for rewrite rules, a `transform` combinator for tree recursion. Reading Catalyst's real source while writing a toy version in the same language gets back something Arc 2 lost when it moved off Rust for W05–W08: a reference implementation you can actually read in your own build language, not just a paper.

---

## Read
- [ ] [Spark SQL: Relational Data Processing in Spark](https://people.csail.mit.edu/matei/papers/2015/sigmod_spark_sql.pdf) (Armbrust et al., SIGMOD 2015): read Sections 1–4. Section 4 describes Catalyst directly — the tree representation, rules, and the batches they're organized into (analysis, logical optimization, physical planning).
- [ ] [Catalyst source: `TreeNode.scala`](https://github.com/apache/spark/blob/master/sql/catalyst/src/main/scala/org/apache/spark/sql/catalyst/trees/TreeNode.scala): skim `transform`, `transformDown`, and `transformUp`. This is the real version of the combinator you're about to build a simplified copy of.
- [ ] [Catalyst source: predicate pushdown rule](https://github.com/apache/spark/blob/master/sql/catalyst/src/main/scala/org/apache/spark/sql/catalyst/optimizer/Optimizer.scala): search for `PushDownPredicates` (or a similarly named rule in the current file). This is a production version of the exact rewrite you'll implement below.

**Key question:** Catalyst's `transform` takes a `PartialFunction[LogicalPlan, LogicalPlan]` — a function that's only defined for some inputs. Why is a partial function the right abstraction here, instead of a total function that has to handle every possible plan shape explicitly?

---

## Code

Project: `code/query-planner/` (Scala 3, sbt)

**Plan representation:**

- [ ] `Expr.scala`: a small expression ADT as sealed case classes: `case class Column(name: String) extends Expr`, `case class Literal(value: Int) extends Expr`, `case class GreaterThan(left: Expr, right: Expr) extends Expr`, `case class And(left: Expr, right: Expr) extends Expr`
- [ ] `LogicalPlan.scala`: a small plan ADT, also sealed case classes: `case class Scan(table: String, columns: List[String]) extends LogicalPlan`, `case class Filter(predicate: Expr, child: LogicalPlan) extends LogicalPlan`, `case class Project(columns: List[String], child: LogicalPlan) extends LogicalPlan`, `case class Join(left: LogicalPlan, right: LogicalPlan, condition: Expr) extends LogicalPlan`

**The rewrite engine:**

- [ ] `TreeTransform.scala`: a generic `transformDown` method on `LogicalPlan` — given a `PartialFunction[LogicalPlan, LogicalPlan]`, apply it to the current node (if defined), then recurse into children, reconstructing the tree bottom-up. This is a deliberately smaller version of Catalyst's real `transformDown`; you don't need to handle every generic tree-manipulation edge case Catalyst does, just enough to run the two rules below.
- [ ] `rules/PushDownFilter.scala`: a rule (`PartialFunction[LogicalPlan, LogicalPlan]`) that rewrites `Filter(pred, Project(cols, child))` into `Project(cols, Filter(pred, child))` when `pred` only references columns present in `cols` — pushing the filter below the projection so it runs over fewer rows, sooner. This is the same idea as the real Catalyst `PushDownPredicates` rule you read above, applied to one specific plan shape instead of the general case.
- [ ] `rules/ConstantFold.scala`: a rule that rewrites `GreaterThan(Literal(a), Literal(b))` directly to a boolean-equivalent `Literal`, and simplifies `And(x, Literal(true-equivalent))` to just `x` — constant folding, applied via the same pattern-matching mechanism.
- [ ] `Optimizer.scala`: runs both rules to a fixed point — repeatedly `transformDown` the plan with each rule until a full pass produces no change. This mirrors Catalyst's own batch-to-fixed-point execution model.

**Verify:**

- [ ] A test (ScalaTest or MUnit, your choice) that builds `Filter(GreaterThan(Column("age"), Literal(18)), Project(List("age", "name"), Scan("users", List("age", "name", "email"))))`, runs it through `Optimizer`, and asserts the result has the `Filter` pushed below the `Project`.

**Minimum bar:** the two rules above compose correctly (running the optimizer on a plan that needs both constant folding and filter pushdown produces a fully rewritten plan in one call to `Optimizer.run`, not just when applied individually).

---

## Reflect

**What clicked:**

**What surprised me:**

**Where did Scala's pattern matching make a rewrite rule read almost exactly like the tree shape it matches — and where did that break down?**

**What does `transformDown` guarantee about rule application order that `transformUp` wouldn't, and why does that matter for filter pushdown specifically?**

**What real Catalyst does that your toy optimizer doesn't (hint: cost-based optimization, not just rule-based):**

**What I'd do differently:**
