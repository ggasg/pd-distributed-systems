---
week_number: 3
status: not-started
---

# W03: Clocks, Causality, Time, and Unreliable Networks

> **Arc:** Storage, Batch, and Failure · **Language:** Go
> **Budget:** about 10 hours. The Minimum bar is what a bad week looks like, not the target.

## What you'll build
Vector clocks + causal message delivery in Go, then a failure detector over the same three nodes. 3 simulated nodes, each its own goroutine, communicating over channels. Assert that no node delivers a message before the messages it causally depends on, then watch your detector confidently declare a perfectly healthy node dead.

This unit is really about one thing with three faces. The network is unreliable, and that single fact gives you: messages arriving out of order (which vector clocks fix), no way to tell a crashed node from a slow one (which nothing fixes, and which the failure detector below makes you confront), and duplicate messages when you retry because of the previous problem (which idempotency fixes). They are not three topics. They are one topic seen from three angles, and the rest of this curriculum keeps returning to them.

**Scenario:** this is the same problem a distributed chat or collaborative-editing app has to solve so a reply never renders before the message it's replying to, even when the network delivers things out of order. Your test suite proves ordering; the exercise below proves the other real failure mode nobody writes a test for on the first pass.

---

## Read
- [ ] DDIA Ch.9 (2nd ed.): unreliable clocks and causality. Focus on "Unreliable Clocks" and "Knowledge, Truth, and Lies". The "Ordering Guarantees" material lives in Ch.10 (Consistency and Consensus), not here, so don't go looking for it in this chapter.
- [ ] [Time, Clocks, and the Ordering of Events in a Distributed System](https://lamport.azurewebsites.net/pubs/time-clocks.pdf) (Lamport, 1978): 11 pages. Read all of it. This paper is the foundation.
- [ ] Optional: [Spanner: Google's Globally Distributed Database](https://dl.acm.org/doi/10.1145/2491245) (Corbett et al., 2012): read only Section 3 (TrueTime API, ~3 pages). Understand how they use bounded clock uncertainty.

- [ ] DDIA Ch.9 again, this time the "Timeouts and Unbounded Delays" section specifically. It is short, and it contains the sentence this unit's second half is built on: over an asynchronous network you cannot distinguish a node that has crashed from one that is merely slow, because the evidence is identical in both cases. Every timeout you will ever set is a guess about that ambiguity.
- [ ] Optional: [Unreliable Failure Detectors for Reliable Distributed Systems](https://dl.acm.org/doi/10.1145/226643.226647) (Chandra & Toueg, JACM 1996; free copies are easy to find). Theory-heavy, and you do not need all of it. What is worth taking is the framing: a failure detector is allowed to be wrong, and the interesting question is not "is it correct" but "how wrong, how often, and how fast." That reframing is more useful in practice than any specific algorithm.

**Depth: study Lamport 1978.** Eleven pages, you implement what it describes, and it rewards a slow pass more than almost anything else in this curriculum. DDIA Ch.9 is a read. Spanner's TrueTime section, Fidge, and Chandra & Toueg are skims; the last one especially, take the framing and leave the proofs.

**The vocabulary Lamport defines, since the rest of the unit uses it constantly.** **Happens-before**, written `a → b`, is the relation the paper builds everything on, and it is defined by exactly three rules: if `a` and `b` are events in the same process and `a` comes first, then `a → b`; if `a` is the sending of a message and `b` is its receipt, then `a → b`; and it is transitive, so `a → b` and `b → c` gives `a → c`. Nothing else creates the relation. Two events are **concurrent** when neither `a → b` nor `b → a`, and this is the word most often misread: it does not mean the events happened at the same instant, it means nothing in the system's message history establishes an order between them, so no observer is entitled to claim one. A **partial order** is exactly that situation, an ordering under which some pairs are simply incomparable. A **total order** orders every pair, which is what you get if you break ties by process ID, and which is useful but no longer means anything causal. Keep that distinction sharp: the Key question below turns on it, and your `concurrent()` function is a direct encoding of it.

**Key question:** Lamport clocks establish a partial order. What does vector clocks give you that Lamport clocks don't?

**Second key question:** You send a heartbeat every second and declare a node dead after missing three. What is the shortest network hiccup that makes you wrong, and what is the longest a genuinely dead node stays undetected? Both answers fall out of the same number, which is the whole problem with tuning it.

---

## Code

Project: `code/clocks/` (Go modules)

- [ ] `vector_clock.go`: `type VectorClock struct { counts map[string]int }`. Implement as functions that return a new `VectorClock` rather than mutating: `func Increment(vc VectorClock, node string) VectorClock`, `func Merge(a, b VectorClock) VectorClock`, `func HappensBefore(a, b VectorClock) bool`, `func Concurrent(a, b VectorClock) bool`. Each of these must build a fresh `map[string]int` and copy every entry across, `maps.Clone` (Go 1.21+, `maps` package) gets you a shallow copy in one call, but you still need to write the mutated copy's changed key yourself afterward. Be honest with yourself about what Go gives you here and what it doesn't: nothing in the type system stops another part of the program from taking your `VectorClock`'s `counts` map and mutating it directly, the way a `record`'s auto-generated immutability would in some other languages. `Increment`/`Merge` returning fresh values is a discipline you're responsible for keeping, not a guarantee the compiler enforces. Treat any function that doesn't copy before writing as a bug, and check for it explicitly when you review your own code.
- [ ] `message.go`: `type Message struct { SenderID string; Payload string; Timestamp VectorClock }`
- [ ] `node.go`: each node runs on its own goroutine with its own inbound channel, `make(chan Message, N)`, an unbuffered or small buffered channel gives you FIFO delivery per sender for free, the direct equivalent of a `BlockingQueue`; on send: increment own entry, attach clock, send the message on the recipient's channel; on receive: buffer messages in a local slice, only deliver when causal preconditions are met (all causally-prior messages already delivered); re-check the buffer every time a new message is delivered, since delivering one message can unblock another
- [ ] `failure_detector.go`: the second half of the unit, over the same three nodes. Each node sends a heartbeat on a `time.Ticker` to every other node. Each node tracks `lastHeard[nodeID] time.Time` and runs a check that marks a peer `SUSPECTED` when the gap exceeds a configurable timeout. Two things worth getting right: mark, do not remove, because a suspected node can come back and your data structure should let it; and make the timeout a parameter rather than a constant, because you are about to tune it deliberately.
- [ ] `causal_delivery_test.go`: Go's standard `testing` package: node A sends m1; node B receives m1 and sends m2 (causally after m1); node C receives m2 before m1; assert C buffers m2 until m1 arrives, then delivers m1 then m2 in order

**Constraints:** simulate 3 nodes in-process, each a goroutine (`go func() { ... }()`). Reorder message delivery in tests by controlling send order directly, not by inserting `time.Sleep` or relying on goroutine scheduling nondeterminism to prove your test cases; make the reordering explicit so the test is deterministic.

**Minimum bar:** the causal delivery test passes, and your failure detector marks a delayed-but-healthy node `SUSPECTED` with both timeout numbers written down. The adaptive detector, the deduplication fix, and the reflect answers are all worth doing and none of them are the bar.

**Break it, then decide (the detector):**
- [ ] Do not crash a node. Instead, make one node sleep for longer than the detector's timeout in the middle of its normal work, then resume as if nothing happened. Watch the other two mark it `SUSPECTED` while it is alive, healthy, and about to send its next heartbeat. Nothing is broken. The detector did exactly what you told it to, and it was wrong, because the information it needs does not exist on the network.
- [ ] Now tune it in both directions and record what each costs. Halve the timeout: detection gets faster and false suspicions get more common, so you start killing healthy work. Triple it: false suspicions go away and a genuinely dead node now holds up everything depending on it for three times as long. Write down both numbers you measured. There is no setting that avoids both, only a choice about which error you would rather make, and the right choice depends entirely on what the caller does with the answer.
- [ ] **Your call:** a fixed timeout treats a datacenter-local peer and a cross-region peer identically, which is obviously wrong, and it also treats a network that is currently healthy the same as one that has been jittery for the last minute. Production systems mostly use an adaptive detector instead, phi-accrual being the common one, which tracks the recent distribution of heartbeat arrival times and outputs a suspicion level rather than a boolean. Stay with your fixed timeout and write down what specifically you are giving up, plus a sketch of how you'd keep the last N inter-arrival gaps and suspect only when the current gap falls well outside their spread. Implementing that is a fine stretch if the unit left you room, but the sketch is what's required. Either way, say what the *caller* should do with a suspicion, because "suspected" is only useful if somebody acts on it differently than they would on "dead."

**Delivery semantics, and why they follow directly from the above.** Because you cannot tell dead from slow, you retry. Because you retry, messages arrive twice. This is not a flaw in anybody's implementation, it is arithmetic, and it is why the three delivery guarantees you will hear named are what they are. **At-most-once** means you never retry, so you lose messages. **At-least-once** means you retry, so you get duplicates. There is no third option on the wire. What people call **exactly-once** is always at-least-once delivery paired with a receiver that either deduplicates or is idempotent, so duplicates stop mattering. Worth being precise about, because "exactly-once" is one of the most oversold phrases in this field, and the next exercise is exactly that situation.

**Break it, then decide (duplicates, optional stretch):**

> The vocabulary above is the part that matters and you now have it. Running the experiment is a second sitting.
- [ ] Send the exact same `Message` value onto a node's inbound channel twice in a row (simulating a network-level retry that redelivers a message the sender thinks was lost). Nothing in `node.go`'s delivery logic checks whether a message with this sender and this exact `VectorClock` has already been delivered, so watch what happens: does the receiving node's own clock get merged and incremented twice for one logical message? If your causal-delivery test suite doesn't already cover this, it's because deduplication and causal ordering are two different problems that are easy to conflate; ordering says *when* a message may be delivered, not *whether* it already was.
- [ ] Given that gap, would you fix it by having each node track the highest sequence number it's seen per sender (a small, bounded amount of extra state) and drop repeats, or by making `Merge` itself idempotent so a duplicate merge is harmless even if delivery isn't deduplicated? Pick one, and say concretely what each approach costs you that the other doesn't.

---

## Rehearse it in Python first (optional, 20 minutes)

> **Why this step exists.** This unit builds in Go, the newest language in the curriculum for most people. Writing the vector clock in Python first means that when the Go version misbehaves, you already know whether the problem is the algorithm or the syntax, which is the most useful thing to know at that moment. This is the only unit that offers the step. Skip it if the algorithm is already obvious to you.

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

**Which of at-most-once, at-least-once, and effectively-once does your duplicate fix actually give you, stated precisely?**

**Where else in this curriculum have you already seen a timeout standing in for knowledge nobody has? (You will hit this again in W09, W11, W14, and W16; note it here so it's familiar when you do.)**

**What I'd do differently:**

---

## Review and articulate

Two steps that exist because self-study has no examiner. Do them at the end of every unit, before marking it done.

- [ ] **Adversarial review.** Hand over three things separately: the number you predicted, the number you measured, and the conclusion you drew. Then ask for the strongest case that the conclusion is *not* supported by the measurement. Do not ask whether you are right; ask what would falsify this. An assistant asked to check your work will tend to find support for your framing, so the prompt has to be adversarial by construction or the exercise is theatre.
- [ ] **Ninety seconds, out loud, timed.** Explain this unit's finding as you would to someone in an interview or a design review: what you measured, what surprised you, and what decision it would change. Articulation under time pressure is a separate skill from understanding, and it is the one that gets tested. If you cannot do it in ninety seconds you do not have the finding yet, you have notes.
