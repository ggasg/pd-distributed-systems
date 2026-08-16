---
week_number: 3
status: not-started
---

# W03: Clocks, Causality, Time, and Unreliable Networks

> **Arc:** Storage, Batch, and Failure · **Language:** Go
> **Budget:** about 10 hours. The Minimum bar is what a bad week looks like, not the target.

## What you'll build

Vector clocks + causal message delivery in Go, then a failure detector over the same three nodes. 3 simulated nodes, each its own goroutine, communicating over channels. Assert that no node delivers a message before the messages it causally depends on, then watch your detector confidently declare a perfectly healthy node dead.

Everything in this unit follows from one fact: the network is unreliable. Messages arrive out of order, which vector clocks fix. A crashed node and a slow node look identical, which nothing fixes and which the failure detector below makes you confront. Retrying because of that ambiguity produces duplicates, which idempotency fixes. The rest of the curriculum keeps returning to all three.

**Scenario:** this is the same problem a distributed chat or collaborative-editing app has to solve so a reply never renders before the message it's replying to, even when the network delivers things out of order. Your test suite proves ordering; the exercise below proves the other real failure mode nobody writes a test for on the first pass.

---

## Read

- [ ] DDIA Ch.9 (2nd ed.): unreliable clocks and causality. Focus on "Unreliable Clocks" and "Knowledge, Truth, and Lies". The "Ordering Guarantees" material lives in Ch.10 (Consistency and Consensus), not here, so don't go looking for it in this chapter.
- [ ] [Time, Clocks, and the Ordering of Events in a Distributed System](https://lamport.azurewebsites.net/pubs/time-clocks.pdf) (Lamport, 1978): 11 pages. Read all of it. This paper is the foundation.
- [ ] Optional: [Spanner: Google's Globally Distributed Database](https://dl.acm.org/doi/10.1145/2491245) (Corbett et al., 2012): read only Section 3 (TrueTime API, ~3 pages). Understand how they use bounded clock uncertainty.

- [ ] DDIA Ch.9 again, this time the "Timeouts and Unbounded Delays" section specifically. It is short, and it contains the sentence this unit's second half is built on: over an asynchronous network you cannot distinguish a node that has crashed from one that is merely slow, because the evidence is identical in both cases. Every timeout you will ever set is a guess about that ambiguity.
- [ ] Optional: [Unreliable Failure Detectors for Reliable Distributed Systems](https://dl.acm.org/doi/10.1145/226643.226647) (Chandra & Toueg, JACM 1996; free copies are easy to find). Theory-heavy, and you do not need all of it. What is worth taking is the framing: a failure detector is allowed to be wrong, and the interesting question is not "is it correct" but "how wrong, how often, and how fast." That reframing is more useful in practice than any specific algorithm.

**Depth: study Lamport 1978.** Eleven pages, you implement what it describes, and it rewards a slow pass more than almost anything else in this curriculum. DDIA Ch.9 is a read. Spanner's TrueTime section, Fidge, and Chandra & Toueg are skims; the last one especially, take the framing and leave the proofs.

**The vocabulary Lamport defines**, which the rest of the unit uses constantly:

- **Happens-before**, written `a → b`. The relation the paper builds everything on, created by these rules and nothing else:
  - `a` and `b` are events in the same process and `a` comes first
  - `a` is the sending of a message and `b` is its receipt
  - transitivity: `a → b` and `b → c` gives `a → c`
- **Concurrent**: neither `a → b` nor `b → a`. The word most often misread. It does not mean the events happened at the same instant; it means nothing in the system's message history establishes an order between them, so no observer is entitled to claim one.
- **Partial order**: an ordering under which some pairs are incomparable, which is exactly the situation above.
- **Total order**: orders every pair, which is what you get if you break ties by process ID. Useful, but no longer means anything causal.

Keep the last two apart. The Key question below turns on the distinction, and your `Concurrent()` function is a direct encoding of it.

**Key question:** Lamport clocks establish a partial order. What does vector clocks give you that Lamport clocks don't?

**Second key question:** You send a heartbeat every second and declare a node dead after missing three. What is the shortest network hiccup that makes you wrong, and what is the longest a genuinely dead node stays undetected? Both answers fall out of the same number, which is the whole problem with tuning it.

---

## Code

Project: `code/clocks/` (Go modules)

### Step 1: `vector_clock.go`

- [ ] `type VectorClock struct { counts map[string]int }`, with four functions that return a new value rather than mutating:
  ```go
  func Increment(vc VectorClock, node string) VectorClock
  func Merge(a, b VectorClock) VectorClock
  func HappensBefore(a, b VectorClock) bool
  func Concurrent(a, b VectorClock) bool
  ```
- [ ] Copy before writing. `maps.Clone` gives you a shallow copy in one call; you still write the changed key on the copy yourself. Nothing in Go's type system stops a caller from mutating the `counts` map directly, so treat any function that writes without copying first as a bug.

### Step 2: `message.go`

- [ ] `type Message struct { SenderID string; Payload string; Timestamp VectorClock }`

### Step 3: `node.go`

- [ ] Each node runs on its own goroutine with its own inbound channel, `make(chan Message, N)`. An unbuffered or small buffered channel gives you FIFO delivery per sender for free.
- [ ] On send: increment the node's own entry, attach the clock, send on the recipient's channel.
- [ ] On receive: buffer the message in a local slice. Deliver only when every causally-prior message has already been delivered.
- [ ] Re-check the buffer after each delivery, since delivering one message can unblock another.

### Step 4: `failure_detector.go`

- [ ] Each node sends a heartbeat on a `time.Ticker` to every other node, tracks `lastHeard[nodeID] time.Time`, and marks a peer `SUSPECTED` when the gap exceeds the timeout.
- [ ] Mark, do not remove. A suspected node can come back and your data structure should let it.
- [ ] Make the timeout a parameter rather than a constant. You are about to tune it deliberately.

### Step 5: `causal_delivery_test.go`

- [ ] Go's standard `testing` package. Node A sends m1; node B receives m1 and sends m2, causally after m1; node C receives m2 before m1. Assert that C buffers m2 until m1 arrives, then delivers m1 then m2 in order.
- [ ] Control send order directly to force the reordering. No `time.Sleep`, and no relying on goroutine scheduling, or the test proves nothing on the runs where the scheduler happens to cooperate.
- [ ] Verify with `go test -race ./...`. The race detector matters here specifically: three goroutines sharing clock state is the exact shape of bug it catches.

**Minimum bar:** the causal delivery test passes, and your failure detector marks a delayed-but-healthy node `SUSPECTED` with both timeout numbers written down. The adaptive detector, the deduplication fix, and the reflect answers are all worth doing and none of them are the bar.

---

## Break it, then decide

### The detector

- [ ] Do not crash a node. Instead, make one node sleep for longer than the detector's timeout in the middle of its normal work, then resume as if nothing happened. Watch the other two mark it `SUSPECTED` while it is alive, healthy, and about to send its next heartbeat. Nothing is broken. The detector did exactly what you told it to, and it was wrong, because the information it needs does not exist on the network.
- [ ] Now tune it in both directions and record what each costs. Halve the timeout: detection gets faster and false suspicions get more common, so you start killing healthy work. Triple it: false suspicions go away and a genuinely dead node now holds up everything depending on it for three times as long. Write down both numbers you measured. There is no setting that avoids both, only a choice about which error you would rather make, and the right choice depends entirely on what the caller does with the answer.
- [ ] **Your call:** a fixed timeout treats a datacenter-local peer and a cross-region peer identically, which is obviously wrong, and it also treats a network that is currently healthy the same as one that has been jittery for the last minute. Production systems mostly use an adaptive detector instead, phi-accrual being the common one, which tracks the recent distribution of heartbeat arrival times and outputs a suspicion level rather than a boolean. Stay with your fixed timeout and write down what specifically you are giving up, plus a sketch of how you'd keep the last N inter-arrival gaps and suspect only when the current gap falls well outside their spread. Implementing that is a fine stretch if the unit left you room, but the sketch is what's required. Either way, say what the *caller* should do with a suspicion, because "suspected" is only useful if somebody acts on it differently than they would on "dead."

### Duplicates (optional stretch)

Because you cannot tell dead from slow, you retry. Because you retry, messages arrive twice. That is arithmetic rather than a flaw in anybody's implementation, and it is what the delivery guarantees are named after:

- **At-most-once**: you never retry, so you lose messages.
- **At-least-once**: you retry, so you get duplicates.
- **Exactly-once**: not a thing on the wire. It is always at-least-once delivery paired with a receiver that deduplicates or is idempotent, so duplicates stop mattering.

> The vocabulary above is the part that matters and you now have it. Running the experiment is a second sitting.
- [ ] Send the exact same `Message` value onto a node's inbound channel twice in a row (simulating a network-level retry that redelivers a message the sender thinks was lost). Nothing in `node.go`'s delivery logic checks whether a message with this sender and this exact `VectorClock` has already been delivered, so watch what happens: does the receiving node's own clock get merged and incremented twice for one logical message? If your causal-delivery test suite doesn't already cover this, it's because deduplication and causal ordering are two different problems that are easy to conflate; ordering says *when* a message may be delivered, not *whether* it already was.
- [ ] Given that gap, would you fix it by having each node track the highest sequence number it's seen per sender (a small, bounded amount of extra state) and drop repeats, or by making `Merge` itself idempotent so a duplicate merge is harmless even if delivery isn't deduplicated? Pick one, and say concretely what each approach costs you that the other doesn't.

---

## Rehearse it in Python first (optional, 20 minutes)

> Writing the vector clock in Python first means that when the Go version misbehaves, you already know whether the problem is the algorithm or the syntax. Skip it if the algorithm is already obvious to you.

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

**Prediction versus measurement.** Fill the predictions in *before* you run anything, and do not edit them afterwards. The gap is where calibration comes from.

| Quantity | Predicted | Measured | Which term I got wrong |
|----------|-----------|----------|------------------------|
| | | | |

Copy anything worth carrying into [MEASUREMENTS.md](../MEASUREMENTS.md).

**What clicked:**

**What surprised me:**

**What happens if a node crashes? Does the vector clock approach break?**

**Duplicate delivery: sequence-number dedup or idempotent merge, and what each costs (from Break it, then decide above):**

**How a system you know handles ordering or timestamps (what clock does it use?):**

**The two timeout numbers you measured: shortest hiccup that produced a false suspicion, and longest a dead node went undetected. Which error would you rather make, and for what caller?**

**Which of at-most-once, at-least-once, or at-least-once-plus-dedup does your duplicate fix actually give you, stated precisely?**

**Where else in this curriculum have you already seen a timeout standing in for knowledge nobody has? (You will hit this again in W09, W11, W14, and W16; note it here so it's familiar when you do.)**

**What I'd do differently:**
