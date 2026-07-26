---
week_number: 4
status: not-started
---

# W04: Clocks, Causality, and Time

> **Arc:** Data Systems Internals · **Language:** Go

## What you'll build
Vector clocks + causal message delivery in Go. 3 simulated nodes, each its own goroutine, communicating over channels. Assert that no node delivers a message before the messages it causally depends on.

---

## Read
- [ ] DDIA Ch.9 (2nd ed.): unreliable clocks and causality. Focus on "Unreliable Clocks" and "Knowledge, Truth, and Lies"; the "Ordering Guarantees" material this used to point to actually lives in Ch.10 (Consistency and Consensus), not this chapter, so don't go looking for it here.
- [ ] [Time, Clocks, and the Ordering of Events in a Distributed System](https://lamport.azurewebsites.net/pubs/time-clocks.pdf) (Lamport, 1978): 11 pages. Read all of it. This paper is the foundation.
- [ ] [Spanner: Google's Globally Distributed Database](https://dl.acm.org/doi/10.1145/2491245) (Corbett et al., 2012): read only Section 3 (TrueTime API, ~3 pages). Understand how they use bounded clock uncertainty.

**Key question:** Lamport clocks establish a partial order. What does vector clocks give you that Lamport clocks don't?

---

## Code

Project: `code/clocks/` (Go modules)

- [ ] `vector_clock.go`: `type VectorClock struct { counts map[string]int }`. Implement as functions that return a new `VectorClock` rather than mutating: `func Increment(vc VectorClock, node string) VectorClock`, `func Merge(a, b VectorClock) VectorClock`, `func HappensBefore(a, b VectorClock) bool`, `func Concurrent(a, b VectorClock) bool`. Each of these must build a fresh `map[string]int` and copy every entry across, `maps.Clone` (Go 1.21+, `maps` package) gets you a shallow copy in one call, but you still need to write the mutated copy's changed key yourself afterward. Be honest with yourself about what Go gives you here and what it doesn't: nothing in the type system stops another part of the program from taking your `VectorClock`'s `counts` map and mutating it directly, the way a `record`'s auto-generated immutability would in some other languages. `Increment`/`Merge` returning fresh values is a discipline you're responsible for keeping, not a guarantee the compiler enforces. Treat any function that doesn't copy before writing as a bug, and check for it explicitly when you review your own code.
- [ ] `message.go`: `type Message struct { SenderID string; Payload string; Timestamp VectorClock }`
- [ ] `node.go`: each node runs on its own goroutine with its own inbound channel, `make(chan Message, N)`, an unbuffered or small buffered channel gives you FIFO delivery per sender for free, the direct equivalent of a `BlockingQueue`; on send: increment own entry, attach clock, send the message on the recipient's channel; on receive: buffer messages in a local slice, only deliver when causal preconditions are met (all causally-prior messages already delivered); re-check the buffer every time a new message is delivered, since delivering one message can unblock another
- [ ] `causal_delivery_test.go`: Go's standard `testing` package: node A sends m1; node B receives m1 and sends m2 (causally after m1); node C receives m2 before m1; assert C buffers m2 until m1 arrives, then delivers m1 then m2 in order

**Constraints:** simulate 3 nodes in-process, each a goroutine (`go func() { ... }()`). Reorder message delivery in tests by controlling send order directly, not by inserting `time.Sleep` or relying on goroutine scheduling nondeterminism to prove your test cases; make the reordering explicit so the test is deterministic.

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

**Connection:** `VectorClock` in Go is this dict, made a `map[string]int` wrapped in a struct, with the copy-before-write discipline you have to enforce yourself instead of getting it for free. Writing it in Python first shows you the data structure is trivial; the subtlety is in `HappensBefore` and `Concurrent`, not in the bookkeeping.

---

## Reflect

**What clicked:**

**What surprised me:**

**What happens if a node crashes? Does the vector clock approach break?**

**How a system you know handles ordering or timestamps (what clock does it use?):**

**What I'd do differently:**
