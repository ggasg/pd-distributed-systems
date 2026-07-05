---
week_number: 7
status: not-started
---

# W07 — Differential Dataflow

> **Arc:** Streaming and Dataflow · **Language:** Scala

## What you'll build
A simplified Differential Dataflow engine from scratch in Scala: the `(key, value, time, diff)` data model, `map`/`filter`/`count` operators that produce only deltas, and an input loop that demonstrates incremental word count and graph reachability — without touching any external library.

---

## Read
- [ ] [Differential Dataflow](https://www.cidrdb.org/cidr2013/Papers/CIDR13_Paper111.pdf) (McSherry et al., CIDR 2013) — read Sections 1–3. Section 2 defines the data model (collections as functions from time to multisets of changes). Section 3 defines the operators.
- [ ] DD Rust source — skim for design intuition only (not to write Rust): [`src/collection.rs`](https://github.com/TimelyDataflow/differential-dataflow/blob/master/src/collection.rs) (the `Collection` type), [`src/operators/join.rs`](https://github.com/TimelyDataflow/differential-dataflow/blob/master/src/operators/join.rs) (how incremental join works)

**Key question:** What is a "difference" in DD? How does `(key, value, time, diff)` encode both additions and retractions? Walk through what happens to the word count index when you retract a document.

---

## Code

Project: `code/dd-scratch/` (Scala 2.13, sbt)

**Core data model:**

- [ ] `Update.scala` — case class `Update[K, V](key: K, value: V, time: Int, diff: Int)` where `diff` is `+1` (addition) or `-1` (retraction)
- [ ] `Collection.scala` — wraps a `List[Update[K, V]]`; implements:
  - `map[K2, V2](f: (K, V) => (K2, V2)): Collection[K2, V2]` — transform each update, preserve diff
  - `filter(p: (K, V) => Boolean): Collection[K, V]` — drop updates where predicate is false
  - `consolidate: Collection[K, V]` — merge updates with the same (key, value, time), sum their diffs, drop zero-diff entries

**Word count:**

- [ ] `WordCount.scala` — given a `Collection[Int, String]` (document id → document text):
  - `flatMap` each document into words: emit one `Update[String, Unit](word, (), time, diff)` per word
  - `count` by key: consolidate, then group by key and sum diffs to get current count per word
  - In a loop: add document at t=1 (diff=+1), print counts; retract it at t=2 (diff=-1), print updated counts; add a different document at t=3
  - Only print the delta each round — what changed — not the full state

**Graph reachability:**

- [ ] `Reachability.scala` — given a `Collection[Int, Int]` of directed edges (src → dst):
  - Start with a set of root nodes (e.g. `{0}`)
  - One iteration: for each reachable node r, for each edge (r → dst), emit dst as reachable
  - Run 3 iterations manually (no loop abstraction needed); print newly reachable nodes per iteration
  - Then: add edge (0→1) at t=1, (1→2) at t=2, retract (0→1) at t=3; re-run and print which nodes are reachable after each change

**Constraints:** no external libraries beyond sbt. All state is immutable (`List`, `Map`). No mutable vars outside the input loop.

---

## 🐍 Python DSA Review (optional)

**defaultdict as a multiset + sorted consolidation** — a DD `Collection` is a map from keys to signed integer multiplicities. Consolidation collapses duplicates and drops zeros.

```python
from collections import defaultdict

# consolidate.py — core operation in every DD collection
def consolidate(updates: list[tuple]) -> dict:
    """
    updates: list of (key, diff) where diff is +1 (add) or -1 (retract)
    Returns: dict of key → net_diff, with zeros removed
    """
    counts: dict = defaultdict(int)
    for key, diff in updates:
        counts[key] += diff
    return {k: v for k, v in counts.items() if v != 0}

# Test: add 3 "apple", retract 2 → net +1
updates = [("apple", 1), ("apple", 1), ("apple", 1), ("apple", -1), ("apple", -1),
           ("banana", 1), ("banana", -1)]  # banana nets to 0 → dropped
result = consolidate(updates)
assert result == {"apple": 1}

# sorted_merge_dd.py — merge two sorted update streams (used in DD operator composition)
def merge_updates(a: list[tuple], b: list[tuple]) -> list[tuple]:
    """Merge two sorted (key, diff) streams, consolidating as we go."""
    combined = sorted(a + b, key=lambda x: x[0])
    return list(consolidate(combined).items())

assert merge_updates([("a", 1), ("b", 1)], [("a", -1), ("c", 1)]) == [("b", 1), ("c", 1)]
```

**Connection:** `Collection.consolidate()` in your Scala DD engine does exactly this — the Python version makes the data model concrete before you reason about it across epochs and time lattices.

---

## Reflect

**What clicked:**

**What surprised me:**

**What happens inside your engine when you retract a record?**

**What your implementation is missing compared to the real DD (hint: arrangements):**

**How this connects to a query or computation engine you've worked with:**

**What I'd do differently:**
