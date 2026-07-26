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
- [ ] [Algebird `Semigroup.scala`](https://github.com/twitter/algebird/blob/develop/algebird-core/src/main/scala/com/twitter/algebird/Semigroup.scala) and [`Monoid.scala`](https://github.com/twitter/algebird/blob/develop/algebird-core/src/main/scala/com/twitter/algebird/Monoid.scala): Twitter's real, production Scala library built entirely around this idea: "abstract algebra for big data." Read the typeclass definitions and a couple of instances (`IntMonoid`, `MaxMonoid`), and compare them against the much smaller version given to you below. Algebird publishes for Scala 2.13, the same version this week targets, so nothing here is read-only-but-not-runnable; you could `libraryDependencies += "com.twitter" %% "algebird-core" % "0.13.10"` and use it directly if you wanted to.
- [ ] [Of Algebirds, Monoids, Monads, and Other Bestiary for Large-Scale Data Analytics](https://www.michael-noll.com/blog/2013/12/02/twitter-algebird-monoid-monad-for-large-scala-data-analytics/) (Michael Noll): a much more accessible walkthrough of why this matters for MapReduce-shaped systems specifically, with concrete examples.

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

**What I'd do differently:**
