---
week_number: 5
status: not-started
---

# W05: Stream Processing Primitives

> **Arc:** Streaming and Dataflow · **Language:** Java

## What you'll build
Tumbling window aggregation from scratch in Java (Part 1), then what happens when events arrive faster than you can aggregate them (Part 2). No Flink, no Spark. Input: a stream of `(eventTime: long, value: int)` tuples. Output: per-window sums, emitted when a watermark advances past the window boundary.

Part 1 is about *time*: which events belong to which window, and when it is safe to declare a window finished. Part 2 is about *rate*: what you do when the events are arriving faster than you are finishing windows. These are independent problems and a stream processor has to solve both, but only the first one has a satisfying answer.

**Scenario:** think of this as the aggregator behind a real-time revenue dashboard. A mobile client that was offline for ten minutes eventually reconnects and sends its buffered events, late, by definition, since your watermark has already moved on without them. Whether those events count is a real product decision, not just a technical one, and it's the decision this week actually makes you make.

---

## Part 1: Windows and Watermarks (3 days)

### Read
- [ ] DDIA Ch.12 (2nd ed.): focus on "Processing Streams" section; understand exactly-once semantics and the log as a stream
- [ ] [The Dataflow Model](https://research.google/pubs/the-dataflow-model-a-practical-approach-to-balancing-correctness-latency-and-cost-in-massive-scale-unbounded-out-of-order-data-processing/) (Akidau et al., VLDB 2015): read Sections 1–4. The windowing taxonomy (fixed, sliding, session) and the "What/Where/When/How" framework are the key takeaways.

**Key question:** What is a watermark, exactly? What breaks if your watermark heuristic is too aggressive? What breaks if it's too conservative?

### Code

Project: `code/streaming/` (Java 21, Maven)

- [ ] `Event.java`: `record Event(long eventTime, int value) {}`
- [ ] `Watermark.java`: `record Watermark(long timestamp) {}`, represents the assertion "no events with eventTime < timestamp will arrive"
- [ ] `StreamItem.java`: `sealed interface StreamItem permits Event, Watermark {}`, then have `Event` and `Watermark` each `implements StreamItem`. A mixed stream of events and watermarks is exactly the two-case sum type this shape exists for, and it's the same idiom W17 uses for `Message`: a sealed interface plus an exhaustive pattern-matching `switch`, no `default` branch, no way to add a third stream-item type later and silently forget to handle it somewhere.
- [ ] `TumblingWindowAggregator.java`: holds a private `Map<Long, Integer> sums` (window id to running sum, a `HashMap` is fine, this is single-threaded); `void onEvent(Event e)`: assign to window, add value; `List<Map.Entry<Long, Integer>> onWatermark(Watermark w)`: emit and evict all windows whose end time is `<=` the watermark timestamp
- [ ] `StreamProcessor.java`: `List<Map.Entry<Long, Integer>> process(List<StreamItem> items)` runs the stream through the aggregator, dispatching each item with:
  ```java
  for (StreamItem item : items) {
      switch (item) {
          case Event e -> aggregator.onEvent(e);
          case Watermark w -> results.addAll(aggregator.onWatermark(w));
      }
  }
  ```
  and returns the accumulated completed windows.
- [ ] Tests, `StreamProcessorTest.java` (JUnit 5): test 1: all events in order, watermarks advance correctly; test 2: out-of-order events arrive before the watermark; test 3: late event arrives after watermark (confirm it's dropped or handled)

**Constraints:** state lives only inside `TumblingWindowAggregator`: keep `sums` `private`, expose behavior through methods, not the field itself. Java won't stop you from making `sums` public and reaching in from outside the class, that's a discipline you enforce with access modifiers, the same way encapsulation works in any language; the compiler only helps once you've actually marked the field `private`. Use `long` timestamps (milliseconds, matching `Event.eventTime`).

**Minimum bar (Part 1):** windowed sums emit when the watermark passes the window boundary, and you've implemented one real handling of late data rather than leaving it as an opinion.

**Break it, then decide:**
- [ ] Feed a batch of events spanning several windows, but advance the watermark to the maximum event time you've seen so far immediately after each batch, as aggressively as possible. Confirm this forces every window up to that point to close and emit right away. Now have one more, genuinely valid event arrive for an already-closed window (still earlier than "now," just delayed in transit, not maliciously late). Watch it get silently dropped by `onWatermark`'s eviction logic. This is the concrete cost of an aggressive watermark the Key Question above asks about in the abstract; here you're looking at the actual dropped value.
- [ ] Pick one real fix and implement it, rather than leaving "handle late data" as a Reflect answer: either (a) let a late event reopen its window and re-emit a corrected sum downstream (simple, but now a consumer of your output has to handle a value it already saw changing), or (b) route late events to a separate side-output list instead of the main result list, the way Flink actually does, and leave reconciling them as an explicit downstream step rather than a silent correction. Whichever you pick, update `TumblingWindowAggregator` and `StreamProcessorTest.java`'s test 3 to prove it.

---

## Part 2: Backpressure and Flow Control (2 days)

Everything in Part 1 assumed you could keep up. Drop that assumption and a different problem appears, one that has no correct answer, only three wrong ones you get to choose between.

A producer sends at some rate. Your aggregator processes at some rate. If the first number is larger than the second, the queue between them grows, and it grows forever. This is worth stating carefully because it is the part people get wrong: a bigger buffer does not fix it. A bigger buffer only changes how long it takes before you notice. If arrivals genuinely outpace service, no amount of memory saves you, because the gap compounds every second.

The tool for reasoning about this is **Little's Law**, which is one line and applies to every queue you will ever build: `L = λW`. The average number of items in the system equals the arrival rate times the average time each item spends there. It is worth internalizing because it turns vague statements into arithmetic. If 1,000 events arrive per second and each takes 5ms to process, you have 5 items in flight on average. Push arrivals to 10,000 per second without making processing faster and you need 50 in flight, and if your buffer holds 20, you now know exactly what happens and roughly when.

So when the producer outruns you, you have exactly three options and no others: **block** the producer until you catch up, which pushes the problem upstream to whoever is feeding it; **drop** events, which bounds memory and makes your answers wrong; or **buffer** to somewhere larger like disk, which is really just deferring the same choice with more room. Every streaming system you will meet implements some combination of these three.

### Read
- [ ] [Flink: Network Stack and Backpressure](https://nightlies.apache.org/flink/flink-docs-stable/docs/ops/monitoring/back_pressure/): short and concrete. Read how Flink detects backpressure and how it propagates upstream through the job graph rather than being handled locally. That propagation is the important part: a slow operator eventually slows the source, which is a feature, not a failure.

**Key question:** Given Little's Law, if your arrival rate exceeds your service rate, what buffer size makes the system stable? Answer honestly, then say what a buffer is actually for, since it clearly isn't that.

### Code

Same project.

Use `ArrayBlockingQueue` from the JDK as the queue between producer and consumer. Do not hand-roll one; the queue is not the lesson here.

- [ ] `Rates.java`: a producer that emits events at a configurable rate and a wrapper around your Part 1 aggregator that adds a configurable delay per event. Two knobs in one small file, so you can set arrival rate above service rate deliberately and watch what follows.
- [ ] `Policy.java`: `sealed interface Policy permits Block, Drop, Spill`, the same idiom as `StreamItem`, now applied to the three responses above. Implement all three:
  - `Block`: `queue.put(event)`, which blocks the producer thread when full. Measure how far the *source* falls behind its intended rate.
  - `Drop`: `queue.offer(event)`, which returns false instead of blocking. Count what you discarded, and separately compute how wrong the final window sums are compared to a run where nothing was dropped. That second number is the one that matters and the one nobody measures.
  - `Spill`: write overflow events to a file and drain them back when the queue has room. Measure peak file size and the added latency for a spilled event versus one that went straight through.
- [ ] `BackpressureBench.java`: run all three policies against the same overload, and print for each: events processed, events lost, peak memory or file size, source lag, and the error in the final aggregate.

**Minimum bar (Part 2):** all three policies run against the same overload, and you have four numbers for each: events lost, source lag, peak memory or spill size, and the error each one introduces in the final aggregate. That last number is the point of the exercise.

**Break it, then decide:**
- [ ] Set the arrival rate to exactly the service rate and run for a while. It looks stable. Now raise arrivals by 10 percent, which is the kind of change a normal traffic day produces, and watch the queue depth. It does not settle at a higher level, it climbs continuously. Confirm with Little's Law that this was predictable from the two rates alone, before you ran anything.
- [ ] **Your call:** this week's scenario is the aggregator behind a real-time revenue dashboard. Dropping events makes the number on the dashboard quietly wrong, and nobody downstream can tell. Blocking keeps it correct and makes it stale, and if the source is a shared queue that other consumers also read from, your blocking becomes their problem. Spilling keeps it correct and bounded but adds a latency spike exactly when the system is already stressed. Pick one, implement it as the default in `BackpressureBench`, and write down the specific number you'd alert on to find out it was the wrong choice.

**Where you have already met this.** Once you have the three policies in hand, the rest of the curriculum stops looking like unrelated design decisions: W06's shuffle writes partitioned spill files to disk between map and reduce, which is `Spill`; W20's log-aggregator ring buffer silently evicts the oldest line under load, which is `Drop`; and W16's queue-depth threshold, where you stop sending to a cache-holding replica once its queue is too deep, is `Block` expressed as routing. Same three answers, four different weeks.

---

## 🐍 Python DSA Review (optional)

**Min-heap for event-time ordering + monotone deque for sliding window**: the two data structures behind watermarks and windowed aggregation.

```python
import heapq
from collections import deque

# heap_watermark.py: process events in timestamp order (like a streaming engine does)
events = [(3, "c"), (1, "a"), (2, "b")]  # (timestamp, payload)
heap = []
for ts, payload in events:
    heapq.heappush(heap, (ts, payload))

ordered = []
while heap:
    ordered.append(heapq.heappop(heap))
# [(1, 'a'), (2, 'b'), (3, 'c')], ordered by event time regardless of arrival order

# sliding_max.py: O(n) sliding window maximum using a monotone deque
def sliding_max(nums: list[int], k: int) -> list[int]:
    dq: deque[int] = deque()  # stores indices
    result = []
    for i, x in enumerate(nums):
        while dq and nums[dq[-1]] <= x:
            dq.pop()           # remove smaller elements, they can never be the max
        dq.append(i)
        if dq[0] < i - k + 1:
            dq.popleft()       # evict index outside window
        if i >= k - 1:
            result.append(nums[dq[0]])
    return result

assert sliding_max([1, 3, -1, -3, 5, 3, 6, 7], 3) == [3, 3, 5, 5, 6, 7]
```

**Connection:** `TumblingWindowAggregator` in Java holds events in time order inside a hash map keyed by window; a heap is exactly the priority queue that ordering implies. The monotone deque pattern is how you'd implement a sliding-window aggregate efficiently; your Java version does this with an explicit `HashMap`, but the deque makes the O(n) trick explicit.

---

## Reflect

**What clicked:**

**What surprised me:**

**Which fix did you implement, reopen-and-correct or a side-output, and what does it cost a downstream consumer of your results compared to silently dropping late data?**

**How a system you've worked with handles event time vs. processing time:**

**Your three policies, side by side: events lost, source lag, peak memory or spill size, and the error in the final aggregate for each.**

**What buffer size makes an over-subscribed system stable, and what is a buffer actually for?**

**Which policy did you ship for the revenue dashboard, and what number would you alert on to learn you were wrong?**

**What I'd do differently:**
