---
week_number: 10
status: not-started
---

# W10: Aggregation Algebra — Monoids and Semigroups

> **Arc:** Streaming and Dataflow · **Language:** Scala

## What you'll build
A `Semigroup`/`Monoid` typeclass hierarchy in Scala, built from scratch, then used to implement several distributed-style aggregations (sum, max, average, approximate distinct count) that are all correct to combine in any order, from any partial partitioning of the data. This week is the algebraic complement to W09: W09 was about *rearranging* operators safely (query planning); this week is about *combining partial results* safely, which is the other half of what makes an MPP engine's `reduceByKey`/`aggregate`-style operators actually correct at scale.

---

## Read
- [ ] [Algebird `Semigroup.scala`](https://github.com/twitter/algebird/blob/develop/algebird-core/src/main/scala/com/twitter/algebird/Semigroup.scala) and [`Monoid.scala`](https://github.com/twitter/algebird/blob/develop/algebird-core/src/main/scala/com/twitter/algebird/Monoid.scala): Twitter's real, production Scala library built entirely around this idea — "abstract algebra for big data." Read the typeclass definitions and a couple of instances (`IntMonoid`, `MaxMonoid`). Note upfront: Algebird's own published artifacts only target Scala up to 2.13, no Scala 3 port — you're reading this for the idea, not depending on it as a library. Build your own typeclass in current Scala below instead.
- [ ] [Of Algebirds, Monoids, Monads, and Other Bestiary for Large-Scale Data Analytics](https://www.michael-noll.com/blog/2013/12/02/twitter-algebird-monoid-monad-for-large-scala-data-analytics/) (Michael Noll): a much more accessible walkthrough of why this matters for MapReduce-shaped systems specifically, with concrete examples.

**Key question:** A semigroup requires the combining operation to be associative: `(a combine b) combine c == a combine b combine c)`. Why does associativity specifically — not commutativity — determine whether a reduction can be computed as a tree (partial sums combined pairwise across a cluster) instead of strictly left-to-right? Is average a semigroup on its own, or does it need to be represented differently to combine correctly?

---

## Code

Project: `code/agg-algebra/` (Scala 3, sbt)

**The typeclass:**

- [ ] `Semigroup.scala`: `trait Semigroup[A] { def combine(a: A, b: A): A }` — no companion object magic needed yet, just the trait.
- [ ] `Monoid.scala`: `trait Monoid[A] extends Semigroup[A] { def empty: A }` — a semigroup with an identity element, so `combine(a, empty) == a`.
- [ ] `instances/IntInstances.scala`: `given Monoid[Int]` for sum (`combine = _ + _`, `empty = 0`) and a separate `Max` wrapper type with its own `given Monoid[Max]` (`combine = math.max`, `empty = Int.MinValue`) — two different monoids over the same underlying type, which is exactly why Scala's typeclass-with-wrapper-type pattern exists instead of just adding a method to `Int`.
- [ ] `Combine.scala`: a generic `def combineAll[A](xs: List[A])(using m: Monoid[A]): A` that folds a list using the monoid — write it two ways: `xs.foldLeft(m.empty)(m.combine)` (strictly sequential) and a `reduceTree` version that recursively splits the list in half, combines each half, then combines the two results (parallelizable in principle). Test that both produce the same answer for every monoid instance you have — that's associativity paying off directly, not an accident.

**The part that isn't just sum and max:**

- [ ] `Average.scala`: implement average as a monoid *without cheating* — a naive `(a + b) / 2` is not associative (verify this with a failing test first, three values combined two different ways give different answers). The correct approach: represent the accumulator as `case class AvgAcc(sum: Double, count: Int)`, with `combine` adding both fields and `empty = AvgAcc(0, 0)`, and only divide `sum / count` at the very end when you read the result out. This is the general pattern: some aggregations need a richer intermediate representation to become associative, and the final "answer" is a projection off of that, computed once at the end — not that they're just left out of the algebra.
- [ ] `ApproxDistinct.scala`: a deliberately crude approximate-distinct-count monoid using a fixed-size `Set[Int]` capped at, say, 100 elements (`combine` unions and truncates, `empty` is the empty set) — not a real HyperLogLog, just enough to demonstrate that even a lossy, bounded-memory aggregation can still be a correct monoid if `combine` is associative over the representation you chose, which is the actual algebraic property production sketches like HyperLogLog rely on.

**Connect it to Arc 2:**

- [ ] `ConnectToConsolidate.scala`: reimplement W07's `Collection.consolidate()` — which merges updates sharing a `(key, value, time)` and sums their `diff` — as a call to `combineAll` using the `Int` sum monoid, grouped by key. Same operation, now expressed through the general typeclass instead of a one-off `HashMap` sum.

**Minimum bar:** `combineAll` and `reduceTree` agree on every test case; the `Average` monoid's naive-vs-correct distinction is demonstrated with a failing test for the naive version; `ConnectToConsolidate` produces identical output to W07's original `consolidate()` on the same input.

---

## Reflect

**What clicked:**

**What surprised me:**

**Where exactly does associativity break for naive average, concretely — walk through the three-value counterexample:**

**Now that you've built `combineAll` two ways (sequential fold vs. tree-reduce), what does this tell you about why Spark's `reduceByKey` can run partial reductions on each partition before shuffling, instead of shuffling every record first?**

**What I'd do differently:**
