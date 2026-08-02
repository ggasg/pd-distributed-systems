---
week_number: 10
status: not-started
---

# W10: Aggregation Algebra: Monoids and Semigroups

> **Arc:** Streaming and Dataflow · **Language:** Scala

## What you'll build
A small `Semigroup`/`Monoid` typeclass, given rather than built from scratch this time, used to implement several distributed-style aggregations (sum, max, average, approximate distinct count) that are all correct to combine in any order, from any partial partitioning of the data. Writing the typeclass hierarchy itself is standard FP-course material your dedicated Scala/Haskell track will cover properly; this week starts from a working one and spends its time on the actual distributed-systems argument: why associativity, specifically, is what lets a reduction be computed as a tree across a cluster instead of strictly left-to-right, and where an aggregation that looks obviously correct (like averaging averages) silently isn't. This week is the algebraic complement to W09: W09 was about *rearranging* operators safely (query planning); this week is about *combining partial results* safely, which is the other half of what makes an MPP engine's `reduceByKey`/`aggregate`-style operators actually correct at scale.

**Scenario:** somewhere on a real cluster, partial averages computed on different partitions have to be combined into one final number, correctly, no matter which partitions happen to land on which executor or in what order the shuffle delivers them. "Just average the averages" is the obvious thing to reach for and the wrong answer; this week is built around making that wrongness impossible to miss.

---

## Before you start (optional, 15–20 min)

A warm-up if typeclasses and `implicit` parameters aren't something you reach for often; skip it if `implicit val` and having an instance supplied for you automatically already feels natural.

Write the same mechanism in miniature, no `Semigroup`/`Monoid` names yet, just the parameter-resolution trick this whole week depends on:

```scala
trait Combinable[A] {
  def combine(a: A, b: A): A
}

implicit val intAdd: Combinable[Int] = (a, b) => a + b

def combineAll[A](xs: List[A])(implicit c: Combinable[A]): A =
  xs.reduce(c.combine)

// combineAll(List(1, 2, 3, 4)) should be 10, with no Combinable passed explicitly
```

The `(implicit c: Combinable[A])` parameter list is the whole trick: the compiler looks for exactly one `Combinable[Int]` in scope and supplies it, so you never write `combineAll(xs, intAdd)` by hand. `Semigroup`/`Monoid` below are this same shape with one more method (`empty`) and more instances, nothing conceptually new past this point.

If that reads naturally, skim the provided `Semigroup.scala`/`Monoid.scala` below and go straight to `Combine.scala`.

---

## Read
- [ ] [Spark: User Defined Aggregate Functions](https://spark.apache.org/docs/latest/sql-ref-functions-udf-aggregate.html) and the [`Aggregator` scaladoc](https://spark.apache.org/docs/latest/api/scala/org/apache/spark/sql/expressions/Aggregator.html). This is the week's central reading and the reason it matters. To write a type-safe aggregation in Spark you must supply four things, and three of them are the algebra this week is about under different names: `zero` is the monoid's identity, `merge` is the semigroup's `combine` and Spark requires it to be associative, and `finish` is the projection you compute once at the end. Spark's own documented example for this API is average, which is exactly the exercise below. You are not learning an abstraction that happens to resemble a real system; you are learning the interface a real system makes you implement.
- [ ] [Spark source: `RDD.scala`](https://github.com/apache/spark/blob/master/core/src/main/scala/org/apache/spark/rdd/RDD.scala): find `treeAggregate` and read its signature and comment. It asks you for `seqOp` and `combOp` separately, and it combines partial results in a tree of configurable depth rather than a single pass back to the driver. That is `reduceTree` below, in production, and the reason it is allowed to exist is the associativity requirement stated in the previous reading.
- [ ] Optional: [`cats.kernel.Semigroup`](https://github.com/typelevel/cats/blob/main/kernel/src/main/scala/cats/kernel/Semigroup.scala) and [`Monoid`](https://github.com/typelevel/cats/blob/main/kernel/src/main/scala/cats/kernel/Monoid.scala). Two small files, the general-purpose Scala versions of what you're given below. Worth a glance for the shape; the surrounding library is deep FP territory and belongs to your separate Scala and Haskell track, not here.

**Key question:** A semigroup requires the combining operation to be associative: `(a combine b) combine c == a combine b combine c)`. Why does associativity specifically, not commutativity, determine whether a reduction can be computed as a tree (partial sums combined pairwise across a cluster) instead of strictly left-to-right? Is average a semigroup on its own, or does it need to be represented differently to combine correctly?

---

## Code

Project: `code/agg-algebra/` (Scala 2.13, sbt)

**Given, not built:** `Semigroup.scala`, `Monoid.scala`, and `instances/IntInstances.scala` are provided as starter files this week rather than exercises. `Semigroup[A]` (`trait Semigroup[A] { def combine(a: A, b: A): A }`) and `Monoid[A]` (adds `def empty: A`, an identity element so `combine(a, empty) == a`) are standard FP-course material, and building the typeclass hierarchy itself isn't this week's point, writing it yourself is exactly the kind of exercise your dedicated FP track will have you do properly. `IntInstances.scala` gives you `implicit val intSumMonoid: Monoid[Int]` (`combine = _ + _`, `empty = 0`) and a `Max` wrapper type with its own `implicit val maxMonoid: Monoid[Max]` (`combine = math.max`, `empty = Int.MinValue`), two different monoids over the same underlying type, worth noticing why that needs a wrapper type rather than just adding a method to `Int`, even though you didn't write it yourself this time. Read all three files before moving on; everything below is built on top of them.
- [ ] `Combine.scala`: a generic `def combineAll[A](xs: List[A])(implicit m: Monoid[A]): A` that folds a list using the monoid. Write it two ways: `xs.foldLeft(m.empty)(m.combine)` (strictly sequential) and a `reduceTree` version that recursively splits the list in half, combines each half, then combines the two results (parallelizable in principle). Test that both produce the same answer for every monoid instance you have; that's associativity paying off directly, not an accident.

**The part that isn't just sum and max:**

- [ ] `Average.scala`: implement average as a monoid *without cheating*: a naive `(a + b) / 2` is not associative (verify this with a failing test first, three values combined two different ways give different answers). The correct approach: represent the accumulator as `case class AvgAcc(sum: Double, count: Int)`, with `combine` adding both fields and `empty = AvgAcc(0, 0)`, and only divide `sum / count` at the very end when you read the result out. This is the general pattern: some aggregations need a richer intermediate representation to become associative, and the final "answer" is a projection off of that, computed once at the end, not that they're just left out of the algebra.
- [ ] *(optional, stretch)* `ApproxDistinct.scala`: a deliberately crude approximate-distinct-count monoid using a fixed-size `Set[Int]` capped at, say, 100 elements (`combine` unions and truncates, `empty` is the empty set), not a real HyperLogLog, just enough to demonstrate that even a lossy, bounded-memory aggregation can still be a correct monoid if `combine` is associative over the representation you chose, which is the actual algebraic property production sketches like HyperLogLog rely on. Not required for the minimum bar below; `Average.scala` already carries the week's core "richer representation, associative by construction" lesson. If you do build it: is truncating by an arbitrary rule (say, keeping the 100 smallest hash values) still a correct monoid, or does the truncation rule itself have to be chosen carefully for `combine` to stay associative? Pick a truncation rule, then check by hand whether combining three capped sets two different ways gives the same result.

**Connect it to Arc 2:**

- [ ] `ConnectToConsolidate.scala`: reimplement W07's `Collection.consolidate()` (which merges updates sharing a `(key, value, time)` and sums their `diff`) as a call to `combineAll` using the `Int` sum monoid, grouped by key. Same operation, now expressed through the general typeclass instead of a one-off `HashMap` sum.

**Minimum bar:** `combineAll` and `reduceTree` agree on every test case; the `Average` monoid's naive-vs-correct distinction is demonstrated with a failing test for the naive version; `ConnectToConsolidate` produces identical output to W07's original `consolidate()` on the same input.

---

## Reflect

**What clicked:**

**What surprised me:**

**Where exactly does associativity break for naive average, concretely? Walk through the three-value counterexample:**

**Now that you've built `combineAll` two ways (sequential fold vs. tree-reduce), what does this tell you about why Spark's `reduceByKey` can run partial reductions on each partition before shuffling, instead of shuffling every record first?**

**Map your `Average.scala` onto Spark's `Aggregator` interface by name: which of your pieces is `zero`, which is `merge`, and which is `finish`? What would go wrong in a real Spark job if you implemented `merge` as the naive average?**

**W06 had you fix skew by salting hot keys and aggregating in two passes. Which property of your combining function made that safe, and name an aggregation where the same trick would have silently produced a wrong answer.**

**What I'd do differently:**
