---
week_number: 17
status: not-started
---

# W17: Fault Tolerance and Snapshots

> **Arc:** Distributed ML & Compute · **Language:** Java

## What you'll build
Chandy-Lamport distributed snapshot in Java: 3 simulated nodes with FIFO channels, one node triggers a snapshot, all nodes record their local state and in-flight channel contents. Assert the recorded global state is consistent.

---

## Read
- [ ] [Distributed Snapshots: Determining Global States of Distributed Systems](https://dl.acm.org/doi/10.1145/214451.214456) (Chandy & Lamport, 1985): 10 pages. Read all of it. The algorithm is in Section 3. A "marker" is just a special message; that's the whole trick.
- [ ] [Lightweight Asynchronous Snapshots for Distributed Dataflows](https://arxiv.org/abs/1506.08603) (Carbone et al., 2015): Flink's ABS algorithm. Read Sections 1–4. Understand how they extend Chandy-Lamport for cyclic dataflow graphs with barriers.
- [ ] Optional, context: **DDIA Chapter 9**, Consistency and Consensus, specifically the section on linearizability. Chandy-Lamport doesn't give you linearizability, it gives you a *consistent cut* (a recorded state that could have occurred at one instant, even if it never literally did); the chapter is useful precisely because it draws that distinction sharply, so you don't walk away from this week conflating "consistent snapshot" with the stronger guarantees Ch. 9 covers.

**Key question:** Chandy-Lamport requires FIFO channels. What breaks if channels can reorder messages? How does Flink's barrier approach handle this?

---

## Code

Project: `code/snapshot/` (Java 21, Maven)

- [ ] `Channel.java`: a FIFO channel between two nodes. Java's `LinkedBlockingQueue` already gives you FIFO delivery for free, thread-safe, no extra locking required. Wrap it in a small `Channel` type only if you need to intercept messages for the snapshot-recording logic (recording what arrives after a marker but before the channel's own marker). Don't reinvent a queue the JDK already gives you natively.
- [ ] `Message.java`: this is the one place this week gets to fix something the Go version of this exercise explicitly calls out as a real weakness: no compiler-enforced exhaustiveness over message types. Java 21 gives you a genuine sum type here: `sealed interface Message permits DataMessage, Marker {}`, with `record DataMessage(String fromNode, String toNode, int value) implements Message {}` and `record Marker(String fromNode, String toNode, int snapshotId) implements Message {}`. Dispatch with an exhaustive pattern-matching `switch`, and use record patterns (JEP 440, finalized in the same Java 21 release as the switch itself) to deconstruct each message directly in the case label instead of binding the whole object and calling accessors on the next line:
  ```java
  switch (msg) {
      case DataMessage(var from, var to, var value) -> handleData(from, to, value);
      case Marker(var from, var to, var snapshotId) -> handleMarker(from, to, snapshotId);
  }
  ```
  Because `Message` is `sealed` and permits exactly these two types, the compiler requires the `switch` to cover both, no `default` branch, and no way to compile the program if you add a third message type later and forget to update every switch over `Message`. That's not a stylistic nicety, it's exactly the cost the Go version of this week names directly and asks you to notice by its absence; here you get it enforced instead of merely noticed. The record pattern is the same idea one level deeper: `DataMessage` and `Marker` are both records, so the compiler already knows their exact shape, and the pattern lets you say so directly in the case label rather than writing `d.fromNode()`, `d.toNode()`, `d.value()` by hand.
- [ ] `Node.java`: each node runs on its own virtual thread with:
  - Local state: a running integer sum (incremented by incoming `DataMessage.value()`)
  - Snapshot logic: when a `Marker` arrives on channel C, if this is the *first* marker: record local state, start recording all other incoming channels; when markers arrive on all other channels: finalize snapshot (record channel states)
  - When initiating a snapshot: record own state, send markers on all outgoing channels
- [ ] `Coordinator.java`: wires 3 nodes in a ring (0 → 1 → 2 → 0), starts a virtual thread per node, injects a sequence of data messages, then triggers a snapshot from node 0, and waits (via a `CountDownLatch` or by joining the virtual threads) for all nodes to report their recorded states
- [ ] `SnapshotTest.java`: JUnit 5: inject 10 data messages (total sum = 55), trigger a snapshot mid-stream, assert: (1) sum of all recorded local states + sum of all in-flight channel states = total messages sent so far; (2) snapshot completes without deadlock. Java has no single built-in flag equivalent to `go test -race`; using immutable messages (records) and a proper concurrent queue instead of hand-rolled shared mutable state eliminates most of that bug class by construction rather than by detecting it after the fact, which is worth noting as a different, not strictly worse, way of getting to the same confidence.

**Constraints:** no third-party concurrency libraries: `java.util.concurrent` (`BlockingQueue`, virtual threads, `CountDownLatch`) is enough for this. Focus on correctness of the marker protocol, not performance.

---

## 🐍 Python DSA Review (optional)

**FIFO queue + BFS for snapshot reachability**: Chandy-Lamport requires FIFO channels; BFS verifies the snapshot is consistent (every in-flight message is accounted for).

```python
from collections import deque

# fifo_channel.py: what Channel.java wraps a LinkedBlockingQueue around, a FIFO queue with O(1) ops
channel: deque = deque()
channel.append("msg1")   # enqueue, O(1)
channel.append("msg2")
assert channel.popleft() == "msg1"  # dequeue, O(1), unlike list.pop(0) which is O(n)

# snapshot_verify.py: after snapshot, BFS from initiator to confirm all reachable
# nodes have recorded state (consistency check)
def all_nodes_recorded(graph: dict, initiator: str, recorded: set) -> bool:
    """BFS from initiator; every reachable node must be in `recorded`."""
    visited, q = {initiator}, deque([initiator])
    while q:
        node = q.popleft()
        if node not in recorded:
            return False
        for neighbor in graph.get(node, []):
            if neighbor not in visited:
                visited.add(neighbor)
                q.append(neighbor)
    return True

# 3-node ring: 0 to 1 to 2 to 0
ring = {0: [1], 1: [2], 2: [0]}
assert all_nodes_recorded(ring, 0, {0, 1, 2})
assert not all_nodes_recorded(ring, 0, {0, 1})  # node 2 missed
```

**Connection:** Java's `LinkedBlockingQueue` already is this `deque`, with thread safety built in for free. The BFS is how you'd write `SnapshotTest.java`'s consistency assertion if you wanted to do it graph-theoretically rather than just summing integers.

---

## Reflect

**What clicked:**

**What surprised me:**

**What "consistent global state" actually means and why it's useful:**

**How fault tolerance is handled in a system you know (checkpointing vs replay vs idempotency):**

**Now that `Message` is a sealed interface with exhaustive `switch`, try commenting out the `Marker` case and rebuilding. What does the compiler do, and how does that compare to what would have happened in a language without sum types?**

**What did deconstructing `DataMessage`/`Marker` directly in the switch's case labels (record patterns) save you from writing, compared to binding the whole object and calling accessors?**

**What I'd do differently:**
