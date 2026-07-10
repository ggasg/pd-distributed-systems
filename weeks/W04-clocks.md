---
week_number: 4
status: not-started
---

# W04: Clocks, Causality, and Time

> **Arc:** Data Systems Internals · **Language:** Java 21

## What you'll build
Vector clocks + causal message delivery in Java. 3 simulated nodes. Assert that no node delivers a message before the messages it causally depends on.

---

## Read
- [ ] DDIA Ch.8, pp. 291–322: unreliable clocks, ordering guarantees, causality. Pay attention to the "Ordering Guarantees" section.
- [ ] [Time, Clocks, and the Ordering of Events in a Distributed System](https://lamport.azurewebsites.net/pubs/time-clocks.pdf) (Lamport, 1978): 11 pages. Read all of it. This paper is the foundation.
- [ ] [Spanner: Google's Globally Distributed Database](https://dl.acm.org/doi/10.1145/2491245) (Corbett et al., 2012): read only Section 3 (TrueTime API, ~3 pages). Understand how they use bounded clock uncertainty.

**Key question:** Lamport clocks establish a partial order. What does vector clocks give you that Lamport clocks don't?

---

## Code

Project: `code/clocks/` (Java 21)

- [ ] `VectorClock.java`: immutable record holding `int[]` of size N (one entry per node). Implement: `increment(int nodeId)`, `merge(VectorClock other)`, `happensBefore(VectorClock other)`, `concurrent(VectorClock other)`
- [ ] `Message.java`: record with `senderId`, `payload`, `VectorClock timestamp`
- [ ] `Node.java`: each node has its own vector clock; on send: increment own entry, attach clock; on receive: buffer messages, only deliver when causal preconditions are met (all causally prior messages already delivered)
- [ ] `CausalDeliveryTest.java`: test: node A sends m1; node B receives m1 and sends m2 (causally after m1); node C receives m2 before m1; assert C buffers m2 until m1 arrives, then delivers m1 then m2 in order

**Constraints:** simulate 3 nodes in-process. Use `LinkedBlockingQueue` for channels. Reorder message delivery in tests by injecting artificial delay.

---

## 🐍 Python DSA Review (optional)

**Dicts as vector clocks**: a vector clock is just a dict. Implement the three core operations in Python before building the immutable Java record.

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

**Connection:** `VectorClock.java` is this dict, made immutable with a Java record. Writing it as a dict first shows you that the data structure is trivial; the subtlety is in `happensBefore` and `concurrent`.

---

## Reflect

**What clicked:**

**What surprised me:**

**What happens if a node crashes? Does the vector clock approach break?**

**How a system you know handles ordering or timestamps (what clock does it use?):**

**What I'd do differently:**
