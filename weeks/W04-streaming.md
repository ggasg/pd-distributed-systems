---
week_number: 4
status: not-started
---

# W04: Stream Processing Primitives

> **Arc:** Data Movement and Execution · **Language:** Java (Flink) / Python (Spark comparison)
> **Budget:** about 10 hours. The Minimum bar is what a bad week looks like, not the target.

## What you'll build

Part 1: event-time windowing on Flink, driven until it drops data on purpose, then compared against how Spark Structured Streaming answers the same question. Part 2: the one thing no framework will show you, which is what happens when events arrive faster than you can process them, built by hand because comparing the three responses side by side is the entire exercise.

Part 1 is about *time*: which events belong to which window, and when it is safe to declare a window finished. Part 2 is about *rate*: what you do when events arrive faster than you finish windows. These are independent problems and a stream processor has to solve both, but only the first one has a satisfying answer, and only the first one has a good production implementation to learn from.

Flink exposes the Dataflow model's event-time machinery directly: watermark strategy, allowed lateness, and late-data side outputs are separate, visible knobs rather than one policy baked in. Spark Structured Streaming answers the same question with fewer knobs, and the difference between the two is the lesson in Part 1.

**Scenario:** the aggregator behind a real-time revenue dashboard. A mobile client that was offline for ten minutes reconnects and sends its buffered events, late by definition, since your watermark moved on without them. Whether those events count is a product decision, not a technical one, and both frameworks below force you to make it, in different words.

---

## Vocabulary

Read this before Part 1. Every term below appears in the exercises, in the Flink API, or in both, with the authoritative source linked.

**Event time** is the timestamp the event carries, assigned when it happened on the device or service that produced it. **Processing time** is the wall clock on the machine handling it. **Ingestion time** is the wall clock when it entered the system. Event time is the only one of the three that gives the same answer when you replay the same data tomorrow, which is why every correctness argument in this unit is made in event time. Reference: [Flink, Timely Stream Processing](https://nightlies.apache.org/flink/flink-docs-stable/docs/concepts/time/).

**Watermark.** A marker flowing inline with the data, carrying a timestamp `t`, which asserts that no further event with a timestamp at or below `t` should arrive. It is an assertion, not a fact: it is a heuristic the system makes on your behalf, it can be wrong, and everything that follows in Part 1 is about what happens when it is. A watermark is what lets an operator decide a window is finished and emit a result, since without one an event-time window has no way of knowing whether the stream is done with it or merely quiet. Reference: [Flink, Generating Watermarks](https://nightlies.apache.org/flink/flink-docs-stable/docs/dev/datastream/event-time/generating_watermarks/).

**Out-of-orderness bound.** How far behind the largest timestamp seen so far the watermark is held, which is the knob you set with `forBoundedOutOfOrderness`. It buys tolerance for late arrival and pays for it in latency, since every window now waits that much longer before firing. Setting it is a direct trade of completeness against delay, and there is no value that avoids the trade.

**Window.** A finite bucket carved out of an infinite stream so an aggregation can terminate. **Tumbling** windows are fixed-size and non-overlapping, so each event lands in exactly one. **Sliding** windows are fixed-size with a separate slide interval, so they overlap and an event can land in several. **Session** windows have no fixed size at all: they group events separated by less than a configured gap of inactivity, so their boundaries come from the data rather than from the clock. Reference: [Flink, Windows](https://nightlies.apache.org/flink/flink-docs-stable/docs/dev/datastream/operators/windows/).

**Trigger.** The condition that decides when a window emits. In this unit the default event-time trigger fires when the watermark passes the end of the window, but a window can fire more than once, which is exactly what `allowedLateness` causes.

**Allowed lateness.** A grace period past the end of a window during which a late event is still accepted and the window fires again with a corrected result. Default is zero. The cost is that a downstream consumer sees a number it already processed change.

**Late data and side output.** An event whose timestamp falls in a window the watermark has already passed, beyond any allowed lateness, is late and is dropped silently by default. A **side output** is a second, separately tagged output stream you can route those events to instead, which keeps the main output immutable and makes reconciliation an explicit downstream job rather than an invisible loss.

**Stateful operator.** An operator that keeps information between events, such as the running sum inside an open window, rather than mapping each event independently. State is what makes windowing possible and what makes failure recovery a problem worth solving. Reference: [Flink, Stateful Stream Processing](https://nightlies.apache.org/flink/flink-docs-stable/docs/concepts/stateful-stream-processing/).

**Checkpoint.** A consistent snapshot of every operator's state plus the source positions it corresponds to, taken periodically and automatically, so that after a failure the job restarts from that snapshot instead of from nothing. Flink takes checkpoints by injecting **barriers** into the stream, which are markers that flow with the data and tell each operator when to snapshot. That is Chandy-Lamport adapted to a dataflow graph, and **W13 is where you implement the underlying algorithm**; here you only need to know that checkpointing is what makes a streaming job's state survive a crash. A **savepoint** is the same snapshot triggered by you rather than by a timer, used for upgrades and rescaling, and it does not expire when newer checkpoints complete. A **state backend** is where that state and those snapshots are actually kept. References: [Flink, Checkpointing](https://nightlies.apache.org/flink/flink-docs-stable/docs/dev/datastream/fault-tolerance/checkpointing/), [Savepoints](https://nightlies.apache.org/flink/flink-docs-stable/docs/ops/state/savepoints/), [State Backends](https://nightlies.apache.org/flink/flink-docs-stable/docs/ops/state/state_backends/).

**Incremental processing.** Computing over only what arrived since the last run, carrying state forward, rather than recomputing over the whole input. It is the property that makes a streaming aggregation affordable: a tumbling sum updates a running total per event instead of rescanning the window. **Continuous** execution processes each event as it arrives, which is Flink's model. **Micro-batch** execution collects events into small batches on a timer and runs a batch job over each, which is Spark Structured Streaming's default model, and it is why the two engines below give you different latency and different answers about when a result is final. The same word means something related but distinct at the table level in W08, where an incremental read consumes only the new commits in a Delta log rather than rescanning the table.

**Delivery semantics** (at-most-once, at-least-once, and at-least-once paired with dedup) are defined in [W03](W03-clocks.md) and are not redefined here. DDIA Ch.12's exactly-once material assumes them, so if that distinction is fuzzy, reread that section of W03 before starting.

**Backpressure** is defined and measured in Part 2 below, where you build the three possible responses to it by hand.

---

## Part 1: Windows and Watermarks

### Read

- [ ] DDIA Ch.12 (2nd ed.): focus on the "Processing Streams" section; understand exactly-once semantics and the log as a stream.
- [ ] [The Dataflow Model](https://research.google/pubs/the-dataflow-model-a-practical-approach-to-balancing-correctness-latency-and-cost-in-massive-scale-unbounded-out-of-order-data-processing/) (Akidau et al., VLDB 2015): read Sections 1 to 4. The windowing taxonomy (fixed, sliding, session) and the What/Where/When/How framework are the key takeaways, and Flink's API below is close to a direct implementation of them.
- [ ] [Flink: Generating Watermarks](https://nightlies.apache.org/flink/flink-docs-stable/docs/dev/datastream/event-time/generating_watermarks/): the API reference for what you are about to configure. Skim it now, return to it when something drops.

**Depth: read DDIA Ch.12 and Sections 1 to 4 of the Dataflow Model.** The Flink docs pages are references rather than readings.

**Key question:** Restate what a watermark asserts, in your own words and without looking back at the Vocabulary section. Then answer the part the definition does not give you: what breaks if your heuristic is too aggressive, and what breaks if it is too conservative? Answer before you configure anything, then check yourself against the dropped-record count Flink reports.

### Code

Project: `code/streaming/` (Java 21, Maven, Flink 2.3.0)

**Use the Java DataStream API here.** PyFlink's DataStream API lags the Java one on exactly the features this unit turns: watermark strategies, allowed lateness, and late-data side outputs.

**A version note:** Flink 2.x recommends Java 17 and treats Java 21 as beta. Stay on 21, which is what the rest of the JVM stack here is pinned to; a single-job local Flink run will not go anywhere near the edges that beta status is about. Also note that Flink 2.0 removed the old `Time` class in favour of `java.time.Duration`, so tutorials showing `Time.seconds(10)` predate the version you are running.

**Data:** roughly twenty events, written out by hand. The exercise depends on controlling precisely which event arrives when, down to the individual timestamp, which a generator would take away. The event shape matches the shared commerce dataset the later data units use (see [code/README.md](../code/README.md)), so `eventTime` is an `event_ts` and `value` is an order amount, but you are writing the list, not sampling one.

- [ ] `Event.java`: `record Event(long eventTime, int value) {}`, plus a bounded source that emits a scripted sequence you control exactly. Do not use a random or wall-clock-driven source. A hardcoded list replayed in a fixed order is the correct design here. Keep it small enough to reason about on paper: two or three tumbling windows' worth of events is plenty, and you need to be able to say what the correct sum is for each window without running anything.
- [ ] `FlinkWindows.java`: assign timestamps and watermarks with `WatermarkStrategy.forBoundedOutOfOrderness(Duration.ofSeconds(5))`, key the stream, apply `TumblingEventTimeWindows.of(Duration.ofSeconds(10))`, and sum. Print each window as it fires, with its start and end.

**Same job, one knob changed per run:**

- [ ] **1. In order.** Everything arrives in event-time order. Windows fire as the watermark crosses each boundary. Confirm the sums and, more importantly, confirm *when* each one appears relative to the events that follow it.
- [ ] **2. Out of order, inside the bound.** Shuffle the events so some arrive up to 4 seconds late, still inside the 5-second out-of-orderness bound. The sums should be identical to run 1. This is what the bound buys you, and it is worth seeing that it costs nothing but latency.
- [ ] **3. Late, past the bound.** Send one genuinely valid event 30 seconds after its window closed, exactly like the reconnecting mobile client. It is silently discarded. Find Flink's `numLateRecordsDropped` metric and confirm the count. The important detail is that your output looked completely normal: correct-looking sums, no error, no warning, and a number on the dashboard that is now quietly wrong.
- [ ] **4. The fixes, compared side by side.** Add `allowedLateness(Duration.ofSeconds(60))` and re-run: the window re-fires and emits a *corrected* sum, which means a downstream consumer sees a value it already processed change. Then add `sideOutputLateData(lateTag)` and re-run: late events are routed to a separate stream and the main output stays immutable, leaving reconciliation as an explicit downstream step. Run both. Look at both output streams.

**Then the comparison, which is the point of Part 1:**

- [ ] `spark_windows.py`: the same aggregation in Spark Structured Streaming (PySpark), using `withWatermark("eventTime", "5 seconds")` and a tumbling `window(...)`. Same scripted input, same window size, same bound. You are comparing what the two engines do, not what the two APIs look like.
- [ ] Run it against case 3, the late event. Note what Spark gives you and what it does not: there is no side-output equivalent for late records, and the output mode (`append` versus `update`) decides whether you see corrected results at all or only finalised ones. Write down which of Flink's four behaviours above Spark can reproduce and which it cannot.

**Minimum bar (Part 1):** all four Flink runs done, with the dropped-record count from run 3 and both fixes from run 4 observed in real output. Plus one paragraph on what the Spark version does differently. Writing the Spark job is the bar; tuning it is not.

### Break it, then decide

- [ ] **Your call:** for the revenue dashboard, pick `allowedLateness` or a side output and defend it. Be concrete about who pays: allowed lateness means every downstream consumer must handle a previously-emitted number changing, which is a contract change, not a config change. A side output means the main number stays stable and somebody has to own the reconciliation job, which in practice means it is written once and never looked at again. Say which failure you would rather explain to whoever reads the dashboard.
- [ ] Now widen the out-of-orderness bound from 5 seconds to 5 minutes and re-run case 1. Nothing is dropped any more. Say precisely what you paid for that, in a number, and why a stream processor cannot simply set this to infinity and be correct.

---

## Part 2: Backpressure and Flow Control

Everything in Part 1 assumed you could keep up. Drop that assumption and a different problem appears, one with no correct answer.

Flink and Spark both implement backpressure, and both pick one policy and apply it for you. This half is hand-built because the comparison between the policies is the lesson, and no framework offers you the choice.

A producer sends at some rate. Your aggregator processes at some rate. If the first number is larger than the second, the queue between them grows, and it grows forever. This is worth stating carefully because it is the part people get wrong: a bigger buffer does not fix it. A bigger buffer only changes how long it takes before you notice. If arrivals genuinely outpace service, no amount of memory saves you, because the gap compounds every second.

The tool for reasoning about this is **Little's Law**, which is one line and applies to every queue you will ever build: `L = λW`. The average number of items in the system equals the arrival rate times the average time each item spends there. It turns vague statements into arithmetic. If 1,000 events arrive per second and each takes 5ms to process, you have 5 items in flight on average. Push arrivals to 10,000 per second without making processing faster and you need 50 in flight, and if your buffer holds 20, you now know exactly what happens and roughly when.

When the producer outruns you, the options are:

- **Block** the producer until you catch up, which pushes the problem upstream to whoever is feeding it.
- **Drop** events, which bounds memory and makes your answers wrong.
- **Spill** to somewhere larger like disk, which defers the same choice with more room.

There is no fourth. Every streaming system you will meet implements some combination of these.

### Read

- [ ] [Flink: Network Stack and Backpressure](https://nightlies.apache.org/flink/flink-docs-stable/docs/ops/monitoring/back_pressure/): short and concrete. Read how Flink detects backpressure and how it propagates upstream through the job graph rather than being handled locally. That propagation is the important part: a slow operator eventually slows the source, which is a feature, not a failure. Note which of the three policies below that is, since Flink made the choice for you.
- [ ] Optional: **Burns, *Designing Distributed Systems*, 2nd ed., Chapter 12** (Event-Driven Batch Processing). "Work Stealing" is a further response to overload that Part 2 does not give you: rather than block, drop, spill, or add workers, let idle workers take work from busy ones, which fixes imbalance but not insufficiency. And "Errors, Priority, and Retry" is the question of what happens to the events you did not drop but could not process, which is the case Part 2 quietly assumes away.
- [ ] Optional: **Burns, *Designing Distributed Systems*, 2nd ed., Chapter 11** (Work Queue Systems). Read "Dynamic Scaling of the Workers" and "The Multiworker Pattern." Part 2 gives you three ways to respond to overload with a fixed number of workers; this chapter is the fourth option you did not have, which is to add workers, and it is worth knowing what that fixes and what it does not. It does nothing about a source that is faster than any number of workers.

**Key question:** Given Little's Law, if your arrival rate exceeds your service rate, what buffer size makes the system stable? Answer honestly, then say what a buffer is actually for, since it clearly is not that.

### Code

Project: `code/backpressure/` (Java 21, Maven, no framework)

Use `ArrayBlockingQueue` from the JDK as the queue between producer and consumer. Do not hand-roll one; the queue is not the lesson here.

- [ ] `Rates.java`: a producer that emits events at a configurable rate, and a consumer with a configurable delay per event, so you can set arrival rate above service rate deliberately and watch what follows.
- [ ] `Policy.java`: `sealed interface Policy permits Block, Drop, Spill`. The same sealed-interface-plus-exhaustive-switch idiom W05 and W13 use, applied here to the three responses above. Implement `Block` and `Drop`; `Spill` is optional:
  - `Block`: `queue.put(event)`, which blocks the producer thread when full. Measure how far the *source* falls behind its intended rate.
  - `Drop`: `queue.offer(event)`, which returns false instead of blocking. Count what you discarded, and separately compute how wrong the final window sums are compared to a run where nothing was dropped. That second number is the one that matters and the one nobody measures.
  - `Spill` (optional): write overflow events to a file and drain them back when the queue has room, measuring peak file size and the added latency for a spilled event. `Block` and `Drop` already give you the trade-off; spilling is the one whose behaviour you can predict without running it.
- [ ] `BackpressureBench.java`: run all three policies against the same overload, and print for each: events processed, events lost, peak memory or file size, source lag, and the error in the final aggregate.
  - Events lost and the error in the final aggregate are different reductions over the same sequence, and you cannot make a second pass over a stream you have already consumed. [`Collectors.teeing`](https://docs.oracle.com/en/java/javase/21/docs/api/java.base/java/util/stream/Collectors.html#teeing(java.util.stream.Collector,java.util.stream.Collector,java.util.function.BiFunction)) runs both collectors over one pass and merges the results.

**Minimum bar (Part 2):** `Block` and `Drop` run against the same overload, and you have three numbers for each: events lost, source lag, and the error each introduces in the final aggregate. That last number is the point of the exercise.

### Break it, then decide

- [ ] Set the arrival rate to exactly the service rate and run for a while. It looks stable. Now raise arrivals by 10 percent, which is the kind of change a normal traffic day produces, and watch the queue depth. It does not settle at a higher level, it climbs continuously. Confirm with Little's Law that this was predictable from the two rates alone, before you ran anything.
- [ ] **Your call:** dropping events makes the dashboard number quietly wrong and nobody downstream can tell. Blocking keeps it correct and makes it stale, and if the source is a shared queue that other consumers also read from, your blocking becomes their problem. Spilling keeps it correct and bounded but adds a latency spike exactly when the system is already stressed. Pick one, implement it as the default in `BackpressureBench`, and write down the specific number you would alert on to find out it was the wrong choice.
- [ ] Finally, connect the two halves: Flink chose one of these three for you and propagates it to the source. Given what you measured, say what that choice costs on the revenue dashboard, and whether you would want the ability to override it.

**Where these policies show up again.** W05's shuffle writes partitioned spill files to disk between map and reduce, which is `Spill`. W15's log-aggregator ring buffer silently evicts the oldest line under load, which is `Drop`. W12's queue-depth threshold, where you stop sending to a cache-holding replica once its queue is too deep, is `Block` expressed as routing.

## Reflect

**Prediction versus measurement.** Fill the predictions in *before* you run anything, and do not edit them afterwards. The gap is where calibration comes from.

| Quantity | Predicted | Measured | Which term I got wrong |
|----------|-----------|----------|------------------------|
| | | | |

Copy anything worth carrying into [MEASUREMENTS.md](../MEASUREMENTS.md).

**What clicked:**

**What surprised me:**

**`numLateRecordsDropped` from run 3, and what your output looked like while it was happening:**

**`allowedLateness` or side output for the revenue dashboard, and who pays for your choice:**

**What Spark Structured Streaming could and could not reproduce of Flink's four behaviours:**

**What widening the bound to 5 minutes cost, in a number:**

**Your three policies, side by side: events lost, source lag, peak memory or spill size, and the error in the final aggregate for each.**

**What buffer size makes an over-subscribed system stable, and what is a buffer actually for?**

**Which policy did you ship for the revenue dashboard, and what number would you alert on to learn you were wrong?**

**What I'd do differently:**
