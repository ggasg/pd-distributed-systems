# W05 — Stream Processing Primitives

> **Arc:** Streaming and Dataflow · **Language:** Scala

## What you'll build
Tumbling window aggregation from scratch in Scala. No Flink, no Spark. Input: a stream of `(eventTime: Long, value: Int)` tuples. Output: per-window sums, emitted when a watermark advances past the window boundary.

---

## Read
- [ ] DDIA Ch.11 — focus on "Processing Streams" section; understand exactly-once semantics and the log as a stream
- [ ] [The Dataflow Model](https://research.google/pubs/the-dataflow-model-a-practical-approach-to-balancing-correctness-latency-and-cost-in-massive-scale-unbounded-out-of-order-data-processing/) (Akidau et al., VLDB 2015) — read Sections 1–4. The windowing taxonomy (fixed, sliding, session) and the "What/Where/When/How" framework are the key takeaways.

**Key question:** What is a watermark, exactly? What breaks if your watermark heuristic is too aggressive? What breaks if it's too conservative?

---

## Code

Project: `code/streaming/` (Scala 3, sbt)

- [ ] `Event.scala` — case class `Event(eventTime: Long, value: Int)`
- [ ] `Watermark.scala` — case class `Watermark(timestamp: Long)` — represents the assertion "no events with eventTime < timestamp will arrive"
- [ ] `TumblingWindowAggregator.scala` — maintains a `Map[WindowId, Int]` of partial sums; on `Event`: assign to window, add value; on `Watermark`: emit and evict all windows whose end time ≤ watermark timestamp
- [ ] `StreamProcessor.scala` — processes a `Seq[Either[Event, Watermark]]` (mixed stream of events and watermarks) through the aggregator; returns `Seq[(WindowId, Int)]` of completed windows
- [ ] `StreamProcessorTest.scala` — test 1: all events in order, watermarks advance correctly; test 2: out-of-order events arrive before the watermark; test 3: late event arrives after watermark (confirm it's dropped or handled)

**Constraints:** purely functional where possible. No mutable state outside the aggregator class. Use `Long` timestamps (milliseconds).

---

## Reflect

**What clicked:**

**What surprised me:**

**How would you handle late data without dropping it?**

**How this connects to Materialize's approach to time:**

**What I'd do differently:**
