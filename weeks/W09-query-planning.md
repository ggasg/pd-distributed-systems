---
week_number: 9
status: not-started
---

# W09: Query Planning: Rules, Then Costs

> **Arc:** Streaming and Dataflow · **Language:** Scala

## What you'll build
A toy version of Spark's Catalyst optimizer, in two halves that answer different questions.

Part 1 (3 days) is rule-based planning: a logical query plan as an algebraic data type (Scala case classes), rewrite rules as pattern-matching partial functions, and a generic `transform` combinator that applies them recursively across the tree. Every rule here is a rewrite that is always safe, because it preserves the result no matter what the data looks like.

Part 2 (2 days) is cost-based planning, which is a different kind of thing entirely and it is worth being clear about the difference before you start. A cost-based optimizer does not know that its rewrite is correct-and-faster. It *estimates* how much data each plan would move, using statistics about the tables, and picks the cheapest guess. When the statistics are good this is enormously more powerful than rules. When they are wrong it confidently ships a terrible plan and tells you nothing. You'll build both halves and then break the second one on purpose.

This is the capstone to Arc 2: W05 to W08 built the individual operators (windowed aggregation, the shuffle, incremental views, vectorized execution); this week builds the thing that arranges and rewrites those operators into an optimized plan, the way a real MPP engine does.

**Scenario:** every rule you write in Part 1 is a small, local rewrite (push this filter, fold this constant) that has to compose safely with every other rule already in the optimizer, without you re-checking the whole system by hand each time you add one. That's the actual engineering problem a rule-based optimizer solves, and it's where the Part 1 exercises go looking for the gap between "the rule looks right" and "the rule is safe to run automatically, on any plan, forever." Part 2's scenario is the one you're more likely to actually get paged about: a query that ran in two minutes for a year and now runs for six hours, because a table grew and nobody refreshed its statistics.

**Note on why Scala, specifically:** this isn't an arbitrary FP-language choice. Spark itself is written in Scala, and Catalyst (its query optimizer) is genuinely built the way this week has you build your toy version: case classes for plan nodes, pattern matching for rewrite rules, a `transform` combinator for tree recursion. Reading Catalyst's real source while writing a toy version in the same language gets you something W05 to W08 don't have: a reference implementation you can actually read in your own build language, not just a paper.

---

## Before you start (optional, 15–20 min)

A warm-up if it's been a while since you wrote Scala day to day; skip it entirely if you're currently writing Spark jobs and this all looks familiar already.

Write a five-line arithmetic expression evaluator using the exact pattern this week scales up: a sealed ADT of case classes, and a function that pattern-matches over it.

```scala
sealed trait Expr
case class Num(value: Int) extends Expr
case class Add(left: Expr, right: Expr) extends Expr
case class Mul(left: Expr, right: Expr) extends Expr

def eval(e: Expr): Int = e match {
  case Num(v)     => v
  case Add(l, r)  => eval(l) + eval(r)
  case Mul(l, r)  => eval(l) * eval(r)
}

// eval(Add(Num(2), Mul(Num(3), Num(4)))) should be 14
```

If that reads naturally, go straight to `Expr.scala`/`LogicalPlan.scala` below: same shape, a plan tree instead of an arithmetic tree. `TreeTransform.scala` is `eval`'s recursive match generalized into a `transformDown` that rewrites the tree instead of reducing it to a number.

---

## Part 1: Rule-Based Planning (3 days)

### Read
- [ ] Optional but recommended: **DDIA Chapter 3** (2nd ed.), Data Models and Query Languages. Read the "Query Languages for Data" section specifically. Kleppmann draws the declarative-vs-imperative line right where this week's `LogicalPlan` lives: a `Filter`/`Project`/`Join` tree describes *what* result you want, not the loop that computes it, the same distinction the chapter uses to explain why a declarative query language leaves room for an optimizer to rewrite the plan before executing it. That's the whole justification for `PushDownFilter` existing.
- [ ] [Spark SQL: Relational Data Processing in Spark](https://people.csail.mit.edu/matei/papers/2015/sigmod_spark_sql.pdf) (Armbrust et al., SIGMOD 2015): read Sections 1–4. Section 4 describes Catalyst directly: the tree representation, rules, and the batches they're organized into (analysis, logical optimization, physical planning).
- [ ] [Catalyst source: `TreeNode.scala`](https://github.com/apache/spark/blob/master/sql/catalyst/src/main/scala/org/apache/spark/sql/catalyst/trees/TreeNode.scala): skim `transform`, `transformDown`, and `transformUp`. This is the real version of the combinator you're about to build a simplified copy of.
- [ ] [Catalyst source: predicate pushdown rule](https://github.com/apache/spark/blob/master/sql/catalyst/src/main/scala/org/apache/spark/sql/catalyst/optimizer/Optimizer.scala): search for `PushDownPredicates` (or a similarly named rule in the current file). This is a production version of the exact rewrite you'll implement below.

**Key question:** Catalyst's `transform` takes a `PartialFunction[LogicalPlan, LogicalPlan]`, a function that's only defined for some inputs. Why is a partial function the right abstraction here, instead of a total function that has to handle every possible plan shape explicitly?

### Code

Project: `code/query-planner/` (Scala 2.13, sbt)

**Plan representation:**

- [ ] `Expr.scala`: a small expression ADT as sealed case classes: `case class Column(name: String) extends Expr`, `case class Literal(value: Int) extends Expr`, `case class GreaterThan(left: Expr, right: Expr) extends Expr`, `case class And(left: Expr, right: Expr) extends Expr`
- [ ] `LogicalPlan.scala`: a small plan ADT, also sealed case classes: `case class Scan(table: String, columns: List[String]) extends LogicalPlan`, `case class Filter(predicate: Expr, child: LogicalPlan) extends LogicalPlan`, `case class Project(columns: List[String], child: LogicalPlan) extends LogicalPlan`, `case class Join(left: LogicalPlan, right: LogicalPlan, condition: Expr) extends LogicalPlan`. Note that no rule in Part 1 touches `Join` at all. That isn't an oversight, it's the point: rewriting a join is the one decision you cannot make from the plan's shape alone, because it depends on how much data is in each table. That's Part 2's job.

**The rewrite engine:**

- [ ] `TreeTransform.scala`: a generic `transformDown` method on `LogicalPlan`: given a `PartialFunction[LogicalPlan, LogicalPlan]`, apply it to the current node (if defined), then recurse into children, reconstructing the tree bottom-up. This is a deliberately smaller version of Catalyst's real `transformDown`; you don't need to handle every generic tree-manipulation edge case Catalyst does, just enough to run the two rules below.
  A `PartialFunction[A, B]` is just a normal Scala value, you construct one the same way you'd write a `match` block, but without the surrounding `x match { ... }`: `val double: PartialFunction[Int, Int] = { case n if n % 2 == 0 => n * 2 }`. That value has an `isDefinedAt` you can call before applying it, which is what lets `transformDown` skip a node a rule doesn't handle instead of throwing a `MatchError`. Each rule below (`PushDownFilter`, `ConstantFold`) is written exactly this way, a `{ case ... => ... }` block assigned to a `PartialFunction[LogicalPlan, LogicalPlan]`.
- [ ] `rules/PushDownFilter.scala`: a rule (`PartialFunction[LogicalPlan, LogicalPlan]`) that rewrites `Filter(pred, Project(cols, child))` into `Project(cols, Filter(pred, child))` when `pred` only references columns present in `cols`, pushing the filter below the projection so it runs over fewer rows, sooner. This is the same idea as the real Catalyst `PushDownPredicates` rule you read above, applied to one specific plan shape instead of the general case.
- [ ] `rules/ConstantFold.scala`: a rule that rewrites `GreaterThan(Literal(a), Literal(b))` directly to a boolean-equivalent `Literal`, and simplifies `And(x, Literal(true-equivalent))` to just `x`: constant folding, applied via the same pattern-matching mechanism.
- [ ] `Optimizer.scala`: runs both rules to a fixed point: repeatedly `transformDown` the plan with each rule until a full pass produces no change. This mirrors Catalyst's own batch-to-fixed-point execution model.

**Verify:**

- [ ] A test (ScalaTest or MUnit, your choice) that builds `Filter(GreaterThan(Column("age"), Literal(18)), Project(List("age", "name"), Scan("users", List("age", "name", "email"))))`, runs it through `Optimizer`, and asserts the result has the `Filter` pushed below the `Project`.

**Minimum bar (Part 1):** the two rules above compose correctly (running the optimizer on a plan that needs both constant folding and filter pushdown produces a fully rewritten plan in one call to `Optimizer.run`, not just when applied individually).

**Break it, then decide:**
- [ ] Temporarily make `transformDown` call a rule's `apply(node)` directly instead of checking `rule.isDefinedAt(node)` first (or using `applyOrElse`). Run `PushDownFilter` over a bare `Scan` with no `Filter`/`Project` wrapping it, a plan shape the rule was never written to handle. Watch it throw a `MatchError` at runtime instead of just leaving the node alone. Put the `isDefinedAt` check (or `applyOrElse`) back and confirm the same plan now passes through untouched. This is the concrete, operational answer to this week's own Key Question: a partial function lets `transformDown` ask "does this rule apply here?" before running it, so an optimizer with dozens of rules doesn't need one giant rule that explicitly handles every plan shape in existence, including ones it has nothing to say about.
- [ ] `Optimizer.run` currently loops "until a full pass produces no change," with no cap. A real rule you write yourself, or a bug in one, could rewrite `A` to `B` and then `B` back to `A` forever. Would you add a max-iteration limit (Catalyst's real approach) that throws or logs a warning if the plan hasn't stabilized after, say, 100 passes, or is unbounded iteration acceptable here since you control every rule yourself and can just fix a misbehaving one? Make a call; if you add the cap, write a deliberately oscillating test rule to prove it actually stops the loop.

---

## Part 2: Cost-Based Planning (2 days)

Part 1's rules can't answer the question that matters most for a real query: given three tables to join, which two should you join first?

There is no shape-based answer. `Join(Join(a, b), c)` and `Join(a, Join(b, c))` return identical results, so no correctness-preserving rewrite rule can prefer one. But they can differ by orders of magnitude in runtime, because the first one materializes an intermediate result and the second one materializes a different, possibly much larger, intermediate result. Choosing well requires knowing roughly how many rows are in each table and how many survive each join. That is what statistics are for, and reasoning over them is what a cost-based optimizer does.

This is also the half of query optimization that actually breaks in production, which is why it gets its own failure exercise below.

### Read
- [ ] [How Good Are Query Optimizers, Really?](http://www.vldb.org/pvldb/vol9/p204-leis.pdf) (Leis et al., VLDB 2015, **free PDF**): read Sections 1 and 3. This is the paper that measured, across real optimizers, how badly cardinality estimates degrade as joins stack up: errors compound multiplicatively, and by four or five joins the estimate can be off by orders of magnitude. Section 3's error-distribution figures are the single most useful thing to carry out of this week, because they tell you exactly how much to trust a plan you didn't verify.
- [ ] [Spark: Adaptive Query Execution](https://spark.apache.org/docs/latest/sql-performance-tuning.html#adaptive-query-execution): short docs page. Read the three features it lists: coalescing shuffle partitions, switching a join strategy at runtime, and optimizing skewed joins. All three exist for one reason, which is that Spark stopped trusting its own pre-execution estimates and started re-planning using row counts it had actually observed.

**Key question:** A rule-based rewrite is correct regardless of the data. A cost-based rewrite is a bet on a statistic. What does that difference imply about how you'd test each one, and about which failures you'd expect to reach production undetected?

### Code

Same project, `code/query-planner/`. Plain Scala throughout, case classes and recursion; nothing here needs a typeclass or an implicit.

- [ ] `Statistics.scala`: `case class TableStats(rowCount: Long, distinctValues: Map[String, Long])` and a `Catalog = Map[String, TableStats]`. Hand-write a catalog for three tables of deliberately different sizes, say `orders` at 10,000,000 rows, `customers` at 100,000, and `regions` at 50.
- [ ] `Cost.scala`: two functions, and keep them separate because they answer different questions.
  - `estimatedRows(plan: LogicalPlan, catalog: Catalog): Long`, recursing over the tree. `Scan` returns the catalog row count. `Project` passes its child's estimate through unchanged. `Filter` multiplies by a selectivity factor, and for now use the classic hardcoded 0.1 that real optimizers historically used in the absence of a histogram; it is worth knowing that number is folklore, not measurement. `Join` uses the textbook formula, `(rowsLeft * rowsRight) / max(distinctLeft, distinctRight)` on the join key.
  - `cost(plan: LogicalPlan, catalog: Catalog): Long`, the sum of `estimatedRows` across every node in the tree. Summing intermediate cardinalities is exactly the cost function the Leis paper uses, and it captures the thing that actually hurts: a plan is expensive when it builds large intermediate results, regardless of how big the final answer is.
- [ ] `JoinReorder.scala`: given a plan containing a join over three tables, enumerate the possible left-deep orderings (`List.permutations` on the three relations is enough, six candidates), build a candidate plan for each, `cost` each one, and return the cheapest. Print all six with their costs rather than only the winner. Seeing the spread is most of the lesson: with `regions` at 50 rows the best and worst orderings should differ by several orders of magnitude.
- [ ] `CostBasedOptimizer.scala`: run Part 1's rule-based `Optimizer` first, then `JoinReorder` on the result. That ordering is not arbitrary, it's what Catalyst does: rules simplify and shrink the plan, and only then does cost-based planning choose between the alternatives that remain. Pushing a filter down first genuinely changes which join order is cheapest, so running these in the other order would give a worse answer.

**Verify:**

- [ ] A test asserting that with `regions` at 50 rows and `orders` at 10,000,000, the optimizer joins `regions` first rather than starting with the two large tables.

**Minimum bar (Part 2):** `CostBasedOptimizer` picks a different join order than the one you wrote by hand, prints the cost of all six candidates, and you can explain in one sentence why the winner won.

**Break it, then decide:**
- [ ] Change the catalog so `customers` claims 1,000 rows when the plan will actually be run against 5,000,000. Re-run the optimizer. It picks a plan, reports a low estimated cost, and raises nothing: no warning, no confidence interval, no indication that anything is wrong. Now compute the real cost of that chosen plan using the correct row count and compare it against the plan the optimizer would have picked with accurate statistics. Write down both numbers. This is the six-hour job from the scenario, and the reason it is hard to catch is exactly what you just observed: the optimizer's output looks identical whether its inputs were right or wrong.
- [ ] Contrast this with Part 1 deliberately, because it's the distinction the whole week turns on. `PushDownFilter` cannot do this. It is either applicable or not, and when it applies, the result is correct and faster regardless of what's in the tables. Nothing about the data can make it a bad idea. Write two sentences on why that makes rules cheap to trust and cost models expensive to trust, and what that implies about which one you'd add to a system you couldn't easily observe in production.
- [ ] **Your call:** you're designing the policy for what happens when statistics are stale. One option is to refuse to reorder joins when a table's statistics are older than some threshold, falling back to the order the query was written in, which is safe and leaves real performance on the table every day. The other is to always trust the statistics, which is fast whenever they're fresh and catastrophic when they aren't. Implement one of them (a `lastRefreshed` timestamp on `TableStats` and a threshold check is enough), and write down which failure you chose to accept. Then note what a third option would look like, given that this is precisely the question Spark's Adaptive Query Execution answers by declining to commit to an estimate at all.

### The bridge back to W06 (observe, don't build)

You already have PySpark installed locally from W07, so this is a short exercise rather than a project, and it closes a loop this curriculum opened three weeks ago.

- [ ] Write a small PySpark script that joins a skewed dataset against a small one, using the same Zipf-shaped key distribution you generated in W06. Run it twice, once with `spark.sql.adaptive.enabled=false` and once with `true` plus `spark.sql.adaptive.skewJoin.enabled=true`, calling `df.explain("formatted")` each time and timing both.
- [ ] Read the two plans. With AQE off you get a plan fixed before execution, chosen from estimates. With AQE on the plan reports itself as adaptive and Spark has coalesced shuffle partitions and split the skewed ones, using row counts it measured at runtime rather than predicted in advance.
- [ ] Connect it explicitly in your notes: splitting a skewed partition into several is the same fix you implemented by hand as salting in W06, with the same reason it's safe (the aggregation is associative) and the same cost (an extra pass). Spark is doing your W06 exercise automatically, and the only reason it can is that it waited until it had real numbers instead of estimated ones.

---

## Reflect

**What clicked:**

**What surprised me:**

**Where did Scala's pattern matching make a rewrite rule read almost exactly like the tree shape it matches, and where did that break down?**

**What does `transformDown` guarantee about rule application order that `transformUp` wouldn't, and why does that matter for filter pushdown specifically?**

**Building `Filter(GreaterThan(Column("age"), Literal(18)), Scan(...))` doesn't scan anything; it's just a `LogicalPlan` value sitting in memory until something else walks it. What would break about `PushDownFilter`'s ability to rearrange the tree if constructing a plan node ran the scan or filter immediately instead of just recording intent? This is the lazy-versus-eager distinction underneath every query engine, Catalyst included, not just this toy version of it.**

**Did you add a max-iteration cap to `Optimizer.run`, and if so, what did your oscillating test rule actually do before you added it (from Part 1's Break it, then decide)?**

**The six join orderings and their costs. Which won, why, and how far apart were best and worst?**

**Estimated cost versus real cost for the plan chosen from the deliberately wrong `customers` statistic. Both numbers, not a description.**

**Why can a rule-based rewrite never be wrong in the way a cost-based one can, and what does that imply about which you'd add first to a system you can't easily observe in production?**

**Stale-statistics policy: which failure did you choose to accept, and what would you tell the person hit by it?**

**What the two Spark plans looked like with AQE off versus on, and where you can point at your own W06 salting fix inside what AQE did automatically:**

**What real Catalyst still does that your toy optimizer doesn't, now that it has both halves:**

**What I'd do differently:**
