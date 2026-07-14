---
week_number: 5
status: not-started
---

# W05: Stream Processing Primitives

> **Arc:** Streaming and Dataflow · **Language:** C++

## What you'll build
Tumbling window aggregation from scratch in C++. No Flink, no Spark. Input: a stream of `(event_time: int64_t, value: int32_t)` tuples. Output: per-window sums, emitted when a watermark advances past the window boundary.

---

## Read
- [ ] DDIA Ch.11: focus on "Processing Streams" section; understand exactly-once semantics and the log as a stream
- [ ] [The Dataflow Model](https://research.google/pubs/the-dataflow-model-a-practical-approach-to-balancing-correctness-latency-and-cost-in-massive-scale-unbounded-out-of-order-data-processing/) (Akidau et al., VLDB 2015): read Sections 1–4. The windowing taxonomy (fixed, sliding, session) and the "What/Where/When/How" framework are the key takeaways.

**Key question:** What is a watermark, exactly? What breaks if your watermark heuristic is too aggressive? What breaks if it's too conservative?

---

## Code

Project: `code/streaming/` (C++, CMake + GoogleTest)

- [ ] `include/streaming/event.hpp`: `struct Event { int64_t event_time; int32_t value; };` — a plain aggregate type, no custom constructors needed
- [ ] `include/streaming/watermark.hpp`: `struct Watermark { int64_t timestamp; };`, represents the assertion "no events with event_time < timestamp will arrive"
- [ ] `include/streaming/aggregator.hpp` + `src/aggregator.cpp`: `class TumblingWindowAggregator` holds a private `std::unordered_map<WindowId, int32_t> sums_`; `void on_event(const Event& e)`: assign to window, add value; `std::vector<std::pair<WindowId, int32_t>> on_watermark(const Watermark& w)`: emit and evict all windows whose end time ≤ watermark timestamp
- [ ] `include/streaming/processor.hpp` + `src/processor.cpp`: `using StreamItem = std::variant<Event, Watermark>;`, representing a mixed stream of events and watermarks; `class StreamProcessor { public: std::vector<std::pair<WindowId, int32_t>> process(const std::vector<StreamItem>& items); }` runs the stream through the aggregator via `std::visit` and returns completed windows
- [ ] Tests, in `tests/processor_test.cpp` (GoogleTest): test 1: all events in order, watermarks advance correctly; test 2: out-of-order events arrive before the watermark; test 3: late event arrives after watermark (confirm it's dropped or handled)

**Constraints:** state lives only inside `TumblingWindowAggregator` — no global mutable state, no free-floating `static` variables. Prefer `<algorithm>` calls (`std::transform`, `std::copy_if`, range-based `for`) over manual index loops where they're at least as readable; the aggregator itself will need real mutation of `sums_`, and that's fine — the discipline here is *encapsulated* mutation (private state, `const`-qualified accessors), not zero mutation. One honest difference from the original Rust plan: Rust's borrow checker enforced that encapsulation for you at compile time. C++ won't stop anything outside the class from reaching into `sums_` if you make it public, or stop you from mutating through a `const&` via `const_cast`. Treat encapsulation here as a design discipline you're now responsible for, not a compiler guarantee. Use `int64_t` timestamps (milliseconds), and mark any method that doesn't mutate `sums_` as `const`.

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

**Connection:** `TumblingWindowAggregator` in C++ holds events in time order inside a hash map; a heap is exactly the priority queue that ordering implies. The monotone deque pattern is how you'd implement a sliding-window aggregate efficiently; your C++ version does this with an explicit `unordered_map`, but the deque makes the O(n) trick explicit.

---

## Reflect

**What clicked:**

**What surprised me:**

**How would you handle late data without dropping it?**

**How a system you've worked with handles event time vs. processing time:**

**What I'd do differently:**
