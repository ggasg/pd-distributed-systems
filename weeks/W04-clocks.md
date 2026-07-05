# W04 — Clocks, Causality, and Time

> **Arc:** Data Systems Internals · **Language:** Java 21

## What you'll build
Vector clocks + causal message delivery in Java. 3 simulated nodes. Assert that no node delivers a message before the messages it causally depends on.

---

## Read
- [ ] DDIA Ch.8, pp. 291–322 — unreliable clocks, ordering guarantees, causality. Pay attention to the "Ordering Guarantees" section.
- [ ] [Time, Clocks, and the Ordering of Events in a Distributed System](https://lamport.azurewebsites.net/pubs/time-clocks.pdf) (Lamport, 1978) — 11 pages. Read all of it. This paper is the foundation.
- [ ] [Spanner: Google's Globally Distributed Database](https://dl.acm.org/doi/10.1145/2491245) (Corbett et al., 2012) — read only Section 3 (TrueTime API, ~3 pages). Understand how they use bounded clock uncertainty.

**Key question:** Lamport clocks establish a partial order. What does vector clocks give you that Lamport clocks don't?

---

## Code

Project: `code/clocks/` (Java 21)

- [ ] `VectorClock.java` — immutable record holding `int[]` of size N (one entry per node). Implement: `increment(int nodeId)`, `merge(VectorClock other)`, `happensBefore(VectorClock other)`, `concurrent(VectorClock other)`
- [ ] `Message.java` — record with `senderId`, `payload`, `VectorClock timestamp`
- [ ] `Node.java` — each node has its own vector clock; on send: increment own entry, attach clock; on receive: buffer messages, only deliver when causal preconditions are met (all causally prior messages already delivered)
- [ ] `CausalDeliveryTest.java` — test: node A sends m1; node B receives m1 and sends m2 (causally after m1); node C receives m2 before m1; assert C buffers m2 until m1 arrives, then delivers m1 then m2 in order

**Constraints:** simulate 3 nodes in-process. Use `LinkedBlockingQueue` for channels. Reorder message delivery in tests by injecting artificial delay.

---

## Reflect

**What clicked:**

**What surprised me:**

**What happens if a node crashes — does the vector clock approach break?**

**How this connects to Materialize's timestamp system:**

**What I'd do differently:**
