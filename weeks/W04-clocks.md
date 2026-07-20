---
week_number: 4
status: not-started
---

# W04: Clocks, Causality, and Time

> **Arc:** Data Systems Internals · **Language:** Java

## What you'll build
Vector clocks + causal message delivery in Java. 3 simulated nodes, each its own virtual thread. Assert that no node delivers a message before the messages it causally depends on.

---

## Read
- [ ] DDIA Ch.8, pp. 291–322: unreliable clocks, ordering guarantees, causality. Pay attention to the "Ordering Guarantees" section.
- [ ] [Time, Clocks, and the Ordering of Events in a Distributed System](https://lamport.azurewebsites.net/pubs/time-clocks.pdf) (Lamport, 1978): 11 pages. Read all of it. This paper is the foundation.
- [ ] [Spanner: Google's Globally Distributed Database](https://dl.acm.org/doi/10.1145/2491245) (Corbett et al., 2012): read only Section 3 (TrueTime API, ~3 pages). Understand how they use bounded clock uncertainty.

**Key question:** Lamport clocks establish a partial order. What does vector clocks give you that Lamport clocks don't?

---

## Code

Project: `code/clocks/` (Java 21, Maven)

- [ ] `VectorClock.java`: `record VectorClock(Map<String, Integer> counts) {}`. Implement as pure functions that return a new `VectorClock` rather than mutating: `static VectorClock increment(VectorClock vc, String node)`, `static VectorClock merge(VectorClock a, VectorClock b)`, `static boolean happensBefore(VectorClock a, VectorClock b)`, `static boolean concurrent(VectorClock a, VectorClock b)`. In the compact constructor of the record, wrap the incoming map with `Map.copyOf(counts)`: this makes the stored map genuinely immutable, not just conventionally treated as if it were. That's a real, concrete improvement over the discipline this exercise requires in some other languages, where a plain mutable map is a reference type and accidentally sharing one between two clocks is a silent bug you have to remember not to make; here, calling `.put()` on the stored map simply throws `UnsupportedOperationException`, so the mistake is caught immediately instead of corrupting state quietly.
- [ ] `Message.java`: `record Message(String senderId, String payload, VectorClock timestamp) {}`
- [ ] `Node.java`: each node runs on its own virtual thread with its own inbound `BlockingQueue<Message>` (a `LinkedBlockingQueue` gives FIFO delivery, the direct equivalent of a channel); on send: increment own entry, attach clock, put the message on the recipient's queue; on receive: buffer messages in a local list, only deliver when causal preconditions are met (all causally-prior messages already delivered); re-check the buffer every time a new message is delivered, since delivering one message can unblock another
- [ ] `CausalDeliveryTest.java`: JUnit 5 test: node A sends m1; node B receives m1 and sends m2 (causally after m1); node C receives m2 before m1; assert C buffers m2 until m1 arrives, then delivers m1 then m2 in order

**Constraints:** simulate 3 nodes in-process, each a virtual thread (`Thread.ofVirtual().start(...)` or `Executors.newVirtualThreadPerTaskExecutor()`). Reorder message delivery in tests by controlling send order directly, not by inserting sleeps or relying on scheduler nondeterminism to prove your test cases; make the reordering explicit so the test is deterministic.

---

## 🐍 Python DSA Review (optional)

**Dicts as vector clocks**: a vector clock is just a dict. Implement the three core operations in Python before building the Java version.

```python
# vector_clock.py
VClock = dict  # {node_id: int}

def increment(vc: VClock, node: str) -> VClock:
    result = dict(vc)
    result[node] = result.get(node, 0) + 1
    return result

def merge(a: VClock, b: VClock) -> VClock:
    keys = set(a) | set(b)
    return {k: max(a.get(k, 0), b.get(k, 0)) for k in keys}

def happens_before(a: VClock, b: VClock) -> bool:
    keys = set(a) | set(b)
    return (all(a.get(k, 0) <= b.get(k, 0) for k in keys)
            and any(a.get(k, 0) < b.get(k, 0) for k in keys))

def concurrent(a: VClock, b: VClock) -> bool:
    return not happens_before(a, b) and not happens_before(b, a)

# Tests
vc0 = {}
vc1 = increment(vc0, "A")        # {"A": 1}
vc2 = increment(vc1, "A")        # {"A": 2}
vc3 = increment(vc0, "B")        # {"B": 1}

assert happens_before(vc1, vc2)
assert not happens_before(vc2, vc1)
assert concurrent(vc2, vc3)      # A and B are causally independent
assert merge(vc2, vc3) == {"A": 2, "B": 1}
```

**Connection:** `VectorClock.java` is this dict, made a `Map<String, Integer>` wrapped in a record, with the copy-on-write discipline enforced by the type system instead of by hand. Writing it in Python first shows you the data structure is trivial; the subtlety is in `happensBefore` and `concurrent`, not in the bookkeeping.

---

## Reflect

**What clicked:**

**What surprised me:**

**What happens if a node crashes? Does the vector clock approach break?**

**How a system you know handles ordering or timestamps (what clock does it use?):**

**What I'd do differently:**
