---
week_number: 13
status: not-started
---

# W13: Fault Tolerance and Snapshots

> **Arc:** Distributed ML Systems · **Language:** Java
> **Budget:** about 5 hours. Hit the Minimum bar first; everything past it is optional.

## What you'll build
Chandy-Lamport distributed snapshot in Java: 3 simulated nodes with FIFO channels, one node triggers a snapshot, all nodes record their local state and in-flight channel contents. Assert the recorded global state is consistent.

**Scenario:** an incident report reads "we took a snapshot for disaster recovery and the restored state doesn't balance, some money appears out of nowhere." The algorithm is provably correct under one assumption, FIFO channels, and the exercise below is where you watch the proof stop applying the moment that assumption doesn't hold.

---

## Read
- [ ] [Distributed Snapshots: Determining Global States of Distributed Systems](https://dl.acm.org/doi/10.1145/214451.214456) (Chandy & Lamport, 1985): 10 pages. Read all of it. The algorithm is in Section 3. A "marker" is just a special message; that's the whole trick.
- [ ] Optional: [Lightweight Asynchronous Snapshots for Distributed Dataflows](https://arxiv.org/abs/1506.08603) (Carbone et al., 2015): Flink's ABS algorithm. Read Sections 1–4. Understand how they extend Chandy-Lamport for cyclic dataflow graphs with barriers.
- [ ] Optional, context: **DDIA Chapter 10** (2nd ed.), Consistency and Consensus, specifically the section on linearizability. Chandy-Lamport doesn't give you linearizability, it gives you a *consistent cut* (a recorded state that could have occurred at one instant, even if it never literally did); the chapter is useful precisely because it draws that distinction sharply, so you don't walk away from this unit conflating "consistent snapshot" with the stronger guarantees Ch. 10 covers.
- [ ] [In Search of an Understandable Consensus Algorithm](https://raft.github.io/raft.pdf) (Ongaro & Ousterhout, USENIX ATC 2014, **free PDF**): the Raft paper. Not implemented anywhere in this curriculum, a deliberate scope call, but it's the algorithm underneath etcd, which is what actually holds the Kubernetes control plane consistent, including the cluster you'll deploy to in W14. Read Sections 1–5 (the formal proof in Section 9 is skippable). It answers a different question than Chandy-Lamport does: Chandy-Lamport gives you a consistent snapshot of state that already exists; Raft is how a cluster agrees on what that state *is* in the first place. You'll watch this algorithm run directly in W14.

**Depth: study Chandy-Lamport.** Ten pages, you implement the algorithm, and the correctness argument is the point. The Raft paper is a read, Sections 1 to 5, and you will watch it run in W14 rather than build it. Flink's ABS paper and DDIA Ch.10 are skims.

**Key question:** Chandy-Lamport requires FIFO channels. What breaks if channels can reorder messages? How does Flink's barrier approach handle this?

---

## Code

Project: `code/snapshot/` (Java 21, Maven)

- [ ] `Channel.java`: a FIFO channel between two nodes. Java's `LinkedBlockingQueue` already gives you FIFO delivery for free, thread-safe, no extra locking required. Wrap it in a small `Channel` type only if you need to intercept messages for the snapshot-recording logic (recording what arrives after a marker but before the channel's own marker). Don't reinvent a queue the JDK already gives you natively.
- [ ] `Message.java`: this is the one place this unit gets to fix something the Go version of this exercise explicitly calls out as a real weakness: no compiler-enforced exhaustiveness over message types. Java 21 gives you a genuine sum type here: `sealed interface Message permits DataMessage, Marker {}`, with `record DataMessage(String fromNode, String toNode, int value) implements Message {}` and `record Marker(String fromNode, String toNode, int snapshotId) implements Message {}`. Dispatch with an exhaustive pattern-matching `switch`, and use record patterns (JEP 440, finalized in the same Java 21 release as the switch itself) to deconstruct each message directly in the case label instead of binding the whole object and calling accessors on the next line:
  ```java
  switch (msg) {
      case DataMessage(var from, var to, var value) -> handleData(from, to, value);
      case Marker(var from, var to, var snapshotId) -> handleMarker(from, to, snapshotId);
  }
  ```
  Because `Message` is `sealed` and permits exactly these two types, the compiler requires the `switch` to cover both, no `default` branch, and no way to compile the program if you add a third message type later and forget to update every switch over `Message`. That's not a stylistic nicety, it's exactly the cost the Go version of this unit names directly and asks you to notice by its absence; here you get it enforced instead of merely noticed. The record pattern is the same idea one level deeper: `DataMessage` and `Marker` are both records, so the compiler already knows their exact shape, and the pattern lets you say so directly in the case label rather than writing `d.fromNode()`, `d.toNode()`, `d.value()` by hand.
- [ ] `Node.java`: each node runs on its own virtual thread with:
  - Local state: a running integer sum (incremented by incoming `DataMessage.value()`)
  - Snapshot logic: when a `Marker` arrives on channel C, if this is the *first* marker: record local state, start recording all other incoming channels; when markers arrive on all other channels: finalize snapshot (record channel states)
  - When initiating a snapshot: record own state, send markers on all outgoing channels
- [ ] `Coordinator.java`: wires 3 nodes in a ring (0 → 1 → 2 → 0), starts a virtual thread per node, injects a sequence of data messages, then triggers a snapshot from node 0, and waits (via a `CountDownLatch` or by joining the virtual threads) for all nodes to report their recorded states
- [ ] `SnapshotTest.java`: JUnit 5: inject 10 data messages (total sum = 55), trigger a snapshot mid-stream, assert: (1) sum of all recorded local states + sum of all in-flight channel states = total messages sent so far; (2) snapshot completes without deadlock. Java has no single built-in flag equivalent to `go test -race`; using immutable messages (records) and a proper concurrent queue instead of hand-rolled shared mutable state eliminates most of that bug class by construction rather than by detecting it after the fact, which is worth noting as a different, not strictly worse, way of getting to the same confidence.

**Constraints:** no third-party concurrency libraries: `java.util.concurrent` (`BlockingQueue`, virtual threads, `CountDownLatch`) is enough for this. Focus on correctness of the marker protocol, not performance.

**Minimum bar:** a 3-node run records a consistent cut that your assertions accept, and you have demonstrated concretely what breaks when the FIFO channel assumption is removed. Flink's barrier variant is reading, not building.

**Break it, then decide:**
- [ ] Add a way to deliberately violate FIFO on one channel: buffer two consecutive `DataMessage`s and enqueue them in swapped order instead of send order. Trigger a snapshot spanning that reordering and check `SnapshotTest.java`'s consistency assertion (sum of recorded states plus in-flight channel states should equal total messages sent so far). Confirm it now fails, or worse, silently produces a "consistent" snapshot that doesn't actually match any state the system was ever really in. This is the concrete cost of the FIFO assumption the Key Question above asks about abstractly; here you're looking at the actual broken invariant.
- [ ] `LinkedBlockingQueue` gives you FIFO for free as long as everything goes through one queue per channel, which is true in this simulation but not guaranteed on a real network (packets can take different routes, retries can reorder). Would you defend against that by having each message carry a per-channel sequence number and having the receiver detect and reject out-of-order delivery (extra bookkeeping, but catches the problem explicitly), or by relying on the transport layer to guarantee FIFO per connection the way TCP already does (no extra code, but now Chandy-Lamport's correctness is silently resting on a property of a layer underneath it that this exercise never touches)? Say which you'd pick for a real system and why.

## Reflect

**What clicked:**

**What surprised me:**

**What "consistent global state" actually means and why it's useful:**

**What actually happened to the consistency assertion when you broke FIFO ordering, and sequence numbers vs. trusting the transport, which did you pick (from Break it, then decide above)?**

**How fault tolerance is handled in a system you know (checkpointing vs replay vs idempotency):**

**Now that `Message` is a sealed interface with exhaustive `switch`, try commenting out the `Marker` case and rebuilding. What does the compiler do, and how does that compare to what would have happened in a language without sum types?**

**What did deconstructing `DataMessage`/`Marker` directly in the switch's case labels (record patterns) save you from writing, compared to binding the whole object and calling accessors?**

**What I'd do differently:**
