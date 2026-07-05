---
week_number: 13
status: not-started
---

# W13 — Fault Tolerance and Snapshots

> **Arc:** Distributed ML & Compute · **Language:** Java 21

## What you'll build
Chandy-Lamport distributed snapshot in Java: 3 simulated nodes with FIFO channels, one node triggers a snapshot, all nodes record their local state and in-flight channel contents. Assert the recorded global state is consistent.

---

## Read
- [ ] [Distributed Snapshots: Determining Global States of Distributed Systems](https://dl.acm.org/doi/10.1145/214451.214456) (Chandy & Lamport, 1985) — 10 pages. Read all of it. The algorithm is in Section 3. A "marker" is just a special message — that's the whole trick.
- [ ] [Lightweight Asynchronous Snapshots for Distributed Dataflows](https://arxiv.org/abs/1506.08603) (Carbone et al., 2015) — Flink's ABS algorithm. Read Sections 1–4. Understand how they extend Chandy-Lamport for cyclic dataflow graphs with barriers.

**Key question:** Chandy-Lamport requires FIFO channels. What breaks if channels can reorder messages? How does Flink's barrier approach handle this?

---

## Code

Project: `code/snapshot/` (Java 21, virtual threads)

- [ ] `Channel.java` — FIFO blocking queue between two nodes (`LinkedBlockingQueue<Message>`). Supports sending regular messages and markers. Records messages received after marker for snapshot purposes.
- [ ] `Message.java` — sealed interface with two implementations: `DataMessage(int fromNode, int toNode, int value)` and `Marker(int fromNode, int toNode, int snapshotId)`
- [ ] `Node.java` — each node has:
  - Local state: a running integer sum (incremented by incoming `DataMessage.value`)
  - Snapshot logic: when a Marker arrives on channel C, if this is the *first* marker: record local state, start recording all other incoming channels; when markers arrive on all other channels: finalize snapshot (record channel states)
  - When initiating a snapshot: record own state, send markers on all outgoing channels
- [ ] `Coordinator.java` — wires 3 nodes in a ring (0→1→2→0), starts virtual thread per node, injects a sequence of data messages, then triggers snapshot from node 0, waits for all nodes to report their recorded states
- [ ] `SnapshotTest.java` — inject 10 data messages (total sum = 55), trigger snapshot mid-stream, assert: (1) sum of all recorded local states + sum of all in-flight channel states = total messages sent so far; (2) snapshot completes without deadlock

**Go rewrite (optional stretch goal):**

Once the Java version works, rewrite `Node` and `Coordinator` in Go using goroutines and channels instead of virtual threads and `LinkedBlockingQueue`. Go's `chan` is a natural fit for FIFO message channels. Compare the two implementations: which is more readable? Which makes the algorithm clearer? (~150 lines of Go total)

---

## 🐍 Python DSA Review (optional)

**FIFO queue + BFS for snapshot reachability** — Chandy-Lamport requires FIFO channels; BFS verifies the snapshot is consistent (every in-flight message is accounted for).

```python
from collections import deque

# fifo_channel.py — what Channel.java wraps: a FIFO queue with O(1) ops
channel: deque = deque()
channel.append("msg1")   # enqueue — O(1)
channel.append("msg2")
assert channel.popleft() == "msg1"  # dequeue — O(1), unlike list.pop(0) which is O(n)

# snapshot_verify.py — after snapshot, BFS from initiator to confirm all reachable
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

# 3-node ring: 0 → 1 → 2 → 0
ring = {0: [1], 1: [2], 2: [0]}
assert all_nodes_recorded(ring, 0, {0, 1, 2})
assert not all_nodes_recorded(ring, 0, {0, 1})  # node 2 missed
```

**Connection:** `Channel.java` is a `LinkedBlockingQueue` — that's this `deque` with thread safety. The BFS is how you'd write `SnapshotTest.java`'s consistency assertion if you wanted to do it graph-theoretically rather than just summing integers.

---

## Reflect

**What clicked:**

**What surprised me:**

**What "consistent global state" actually means and why it's useful:**

**How fault tolerance is handled in your current role (checkpointing vs replay):**

**What I'd do differently:**
