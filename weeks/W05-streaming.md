---
week_number: 5
status: not-started
---

# W05: Stream Processing Primitives

> **Arc:** Streaming and Dataflow · **Language:** Java

## What you'll build
Tumbling window aggregation from scratch in Java. No Flink, no Spark. Input: a stream of `(eventTime: long, value: int)` tuples. Output: per-window sums, emitted when a watermark advances past the window boundary.

**Scenario:** think of this as the aggregator behind a real-time revenue dashboard. A mobile client that was offline for ten minutes eventually reconnects and sends its buffered events, late, by definition, since your watermark has already moved on without them. Whether those events count is a real product decision, not just a technical one, and it's the decision this week actually makes you make.

---

## Read
- [ ] DDIA Ch.12 (2nd ed.): focus on "Processing Streams" section; understand exactly-once semantics and the log as a stream
- [ ] [The Dataflow Model](https://research.google/pubs/the-dataflow-model-a-practical-approach-to-balancing-correctness-latency-and-cost-in-massive-scale-unbounded-out-of-order-data-processing/) (Akidau et al., VLDB 2015): read Sections 1–4. The windowing taxonomy (fixed, sliding, session) and the "What/Where/When/How" framework are the key takeaways.

**Key question:** What is a watermark, exactly? What breaks if your watermark heuristic is too aggressive? What breaks if it's too conservative?

---

## Code

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

**Break it, then decide:**
- [ ] Feed a batch of events spanning several windows, but advance the watermark to the maximum event time you've seen so far immediately after each batch, as aggressively as possible. Confirm this forces every window up to that point to close and emit right away. Now have one more, genuinely valid event arrive for an already-closed window (still earlier than "now," just delayed in transit, not maliciously late). Watch it get silently dropped by `onWatermark`'s eviction logic. This is the concrete cost of an aggressive watermark the Key Question above asks about in the abstract; here you're looking at the actual dropped value.
- [ ] Pick one real fix and implement it, rather than leaving "handle late data" as a Reflect answer: either (a) let a late event reopen its window and re-emit a corrected sum downstream (simple, but now a consumer of your output has to handle a value it already saw changing), or (b) route late events to a separate side-output list instead of the main result list, the way Flink actually does, and leave reconciling them as an explicit downstream step rather than a silent correction. Whichever you pick, update `TumblingWindowAggregator` and `StreamProcessorTest.java`'s test 3 to prove it.

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

**What I'd do differently:**
