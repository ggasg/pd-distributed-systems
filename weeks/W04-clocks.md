---
week_number: 4
status: not-started
---

# W04: Clocks, Causality, and Time

> **Arc:** Data Systems Internals · **Language:** Go

## What you'll build
Vector clocks + causal message delivery in Go. 3 simulated nodes, each its own goroutine. Assert that no node delivers a message before the messages it causally depends on.

---

## Read
- [ ] DDIA Ch.8, pp. 291–322: unreliable clocks, ordering guarantees, causality. Pay attention to the "Ordering Guarantees" section.
- [ ] [Time, Clocks, and the Ordering of Events in a Distributed System](https://lamport.azurewebsites.net/pubs/time-clocks.pdf) (Lamport, 1978): 11 pages. Read all of it. This paper is the foundation.
- [ ] [Spanner: Google's Globally Distributed Database](https://dl.acm.org/doi/10.1145/2491245) (Corbett et al., 2012): read only Section 3 (TrueTime API, ~3 pages). Understand how they use bounded clock uncertainty.

**Key question:** Lamport clocks establish a partial order. What does vector clocks give you that Lamport clocks don't?

---

## Code

Project: `code/clocks/` (Go, module)

- [ ] `vector_clock.go`: `type VectorClock map[string]int`. Implement as functions, not methods that mutate: `Increment(vc VectorClock, node string) VectorClock`, `Merge(a, b VectorClock) VectorClock`, `HappensBefore(a, b VectorClock) bool`, `Concurrent(a, b VectorClock) bool` — each returns a **new** map rather than mutating its input. This matters more in Go than it did conceptually elsewhere: Go maps are reference types, so `vc2 := vc1; vc2["A"]++` would silently mutate `vc1` too. Copy explicitly (`maps.Clone` from the `maps` package, or a manual loop) at the top of every function that "changes" a clock.
- [ ] `message.go`: `type Message struct { SenderID string; Payload string; Timestamp VectorClock }`
- [ ] `node.go`: each node is a goroutine with its own inbound `chan Message`; on send: increment own entry, attach clock, send on the recipient's channel; on receive: buffer messages in a local slice, only deliver when causal preconditions are met (all causally-prior messages already delivered) — re-check the buffer every time a new message is delivered, since delivering one message can unblock another
- [ ] `causal_delivery_test.go`: test: node A sends m1; node B receives m1 and sends m2 (causally after m1); node C receives m2 before m1; assert C buffers m2 until m1 arrives, then delivers m1 then m2 in order

**Constraints:** simulate 3 nodes in-process, each a goroutine. Use buffered or unbuffered `chan Message` for channels — try both and see what happens to delivery order when you switch. Reorder message delivery in tests with `time.Sleep` or by controlling send order directly (don't rely on scheduler nondeterminism to prove your test cases; make the reordering explicit).

---

## 🐍 Python DSA Review (optional)

**Dicts as vector clocks**: a vector clock is just a dict. Implement the three core operations in Python before building the Go version.

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

**Connection:** `vector_clock.go` is this dict, made a `map[string]int` with the same copy-on-write discipline enforced by hand instead of by a language guarantee. Writing it in Python first shows you the data structure is trivial; the subtlety is in `happens_before` and `concurrent` — and in Go, in remembering to copy.

---

## Reflect

**What clicked:**

**What surprised me:**

**What happens if a node crashes? Does the vector clock approach break?**

**How a system you know handles ordering or timestamps (what clock does it use?):**

**What I'd do differently:**
