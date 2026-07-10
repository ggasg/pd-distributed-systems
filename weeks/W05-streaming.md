---
week_number: 5
status: not-started
---

# W05: Stream Processing Primitives

> **Arc:** Streaming and Dataflow · **Language:** Scala

## What you'll build
Tumbling window aggregation from scratch in Scala. No Flink, no Spark. Input: a stream of `(eventTime: Long, value: Int)` tuples. Output: per-window sums, emitted when a watermark advances past the window boundary.

---

## Read
- [ ] DDIA Ch.11: focus on "Processing Streams" section; understand exactly-once semantics and the log as a stream
- [ ] [The Dataflow Model](https://research.google/pubs/the-dataflow-model-a-practical-approach-to-balancing-correctness-latency-and-cost-in-massive-scale-unbounded-out-of-order-data-processing/) (Akidau et al., VLDB 2015): read Sections 1–4. The windowing taxonomy (fixed, sliding, session) and the "What/Where/When/How" framework are the key takeaways.

**Key question:** What is a watermark, exactly? What breaks if your watermark heuristic is too aggressive? What breaks if it's too conservative?

---

## Code

Project: `code/streaming/` (Scala 2.13, sbt)

- [ ] `Event.scala`: case class `Event(eventTime: Long, value: Int)`
- [ ] `Watermark.scala`: case class `Watermark(timestamp: Long)`, represents the assertion "no events with eventTime < timestamp will arrive"
- [ ] `TumblingWindowAggregator.scala`: maintains a `Map[WindowId, Int]` of partial sums; on `Event`: assign to window, add value; on `Watermark`: emit and evict all windows whose end time ≤ watermark timestamp
- [ ] `StreamProcessor.scala`: processes a `Seq[Either[Event, Watermark]]` (mixed stream of events and watermarks) through the aggregator; returns `Seq[(WindowId, Int)]` of completed windows
- [ ] `StreamProcessorTest.scala`: test 1: all events in order, watermarks advance correctly; test 2: out-of-order events arrive before the watermark; test 3: late event arrives after watermark (confirm it's dropped or handled)

**Constraints:** purely functional where possible. No mutable state outside the aggregator class. Use `Long` timestamps (milliseconds).

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

**Connection:** `TumblingWindowAggregator` in Scala holds events in time order; a heap is exactly that priority queue. The monotone deque pattern is how you'd implement a sliding-window aggregate efficiently; your Scala version does this functionally, but the deque makes the O(n) trick explicit.

---

## Reflect

**What clicked:**

**What surprised me:**

**How would you handle late data without dropping it?**

**How a system you've worked with handles event time vs. processing time:**

**What I'd do differently:**
