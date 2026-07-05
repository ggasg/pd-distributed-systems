# W06 — Naiad and Timely Dataflow

> **Arc:** Streaming and Dataflow · **Language:** Scala

## What you'll build
A toy timely dataflow graph in Scala: two operators connected by edges, timestamps with (epoch, iteration) pairs, progress tracking via pointstamp dominance, and notification when a frontier advances.

---

## Read
- [ ] [Naiad: A Timely Dataflow System](https://dl.acm.org/doi/10.1145/2517349.2522738) (Murray et al., SOSP 2013) — read Sections 1–4 carefully. Section 2 defines the computation model. Section 3 defines the progress tracking protocol — this is the heart of it.
- [ ] Skim the [timely-dataflow Rust crate README](https://github.com/TimelyDataflow/timely-dataflow) — read enough to understand how `operator`, `notify_at`, and `frontier` are used in practice

**Key question:** What is a pointstamp? How does pointstamp dominance let nodes know when they've seen all messages for a given timestamp?

---

## Code

Project: `code/timely-toy/` (Scala 3, sbt)

- [ ] `Timestamp.scala` — case class `Timestamp(epoch: Int, iteration: Int)` with a `happensBefore` relation: `(e1, i1) < (e2, i2)` iff `e1 < e2 || (e1 == e2 && i1 < i2)` (total order for this toy; Naiad uses partial order)
- [ ] `Pointstamp.scala` — case class `Pointstamp(location: Int, timestamp: Timestamp)`. Implement `couldResultIn(other: Pointstamp, graph: Graph): Boolean` — conservative check based on graph paths
- [ ] `Operator.scala` — trait with `onMessage(msg: Message)` and `onNotification(ts: Timestamp)`. Two concrete operators: `MapOperator` (transforms messages) and `SinkOperator` (prints output)
- [ ] `ProgressTracker.scala` — maintains outstanding event counts per pointstamp; when a count drops to zero and no pointstamp could-result-in it, fires `onNotification` for that timestamp
- [ ] `DataflowTest.scala` — wire two operators: source → map → sink; send 3 messages at epoch 0; send a "done with epoch 0" signal; assert sink's `onNotification(Timestamp(0, 0))` fires after all messages are processed

**Constraints:** single-threaded. Focus on correctness of the progress tracking logic, not performance.

---

## Reflect

**What clicked:**

**What surprised me:**

**What would break if you removed the could-result-in check?**

**How this directly maps to what Materialize does:**

**What I'd do differently:**
