---
week_number: 16
status: not-started
---

# W16: Fault Tolerance and Snapshots

> **Arc:** Distributed ML & Compute · **Language:** Go

## What you'll build
Chandy-Lamport distributed snapshot in Go: 3 simulated nodes with FIFO channels, one node triggers a snapshot, all nodes record their local state and in-flight channel contents. Assert the recorded global state is consistent.

---

## Read
- [ ] [Distributed Snapshots: Determining Global States of Distributed Systems](https://dl.acm.org/doi/10.1145/214451.214456) (Chandy & Lamport, 1985): 10 pages. Read all of it. The algorithm is in Section 3. A "marker" is just a special message; that's the whole trick.
- [ ] [Lightweight Asynchronous Snapshots for Distributed Dataflows](https://arxiv.org/abs/1506.08603) (Carbone et al., 2015): Flink's ABS algorithm. Read Sections 1–4. Understand how they extend Chandy-Lamport for cyclic dataflow graphs with barriers.
- [ ] Optional, context: **DDIA Chapter 9**, Consistency and Consensus — specifically the section on linearizability. Chandy-Lamport doesn't give you linearizability, it gives you a *consistent cut* (a recorded state that could have occurred at one instant, even if it never literally did); the chapter is useful precisely because it draws that distinction sharply, so you don't walk away from this week conflating "consistent snapshot" with the stronger guarantees Ch. 9 covers.

**Key question:** Chandy-Lamport requires FIFO channels. What breaks if channels can reorder messages? How does Flink's barrier approach handle this?

---

## Code

Project: `code/snapshot/` (Go, module)

- [ ] `channel.go`: a FIFO channel between two nodes. Go's built-in `chan Message` already gives you FIFO delivery for free — wrap it in a small `Channel` type only if you need to intercept messages for the snapshot-recording logic (recording what arrives after a marker but before the channel's own marker). Don't reinvent a queue Go already gives you natively.
- [ ] `message.go`: Go has no sum types, so this is the one place its type system is honestly weaker than what you'd get elsewhere — no compiler-enforced exhaustiveness. Define an interface `Message interface { isMessage() }` and two structs implementing it, `DataMessage{FromNode, ToNode, Value int}` and `Marker{FromNode, ToNode, SnapshotID int}`. Dispatch with a type switch: `switch m := msg.(type) { case DataMessage: ...; case Marker: ... }`. If you add a third message type later, nothing forces you to update every switch — that's a real cost of this approach, worth noticing rather than glossing over.
- [ ] `node.go`: each node is a goroutine with:
  - Local state: a running integer sum (incremented by incoming `DataMessage.Value`)
  - Snapshot logic: when a `Marker` arrives on channel C, if this is the *first* marker: record local state, start recording all other incoming channels; when markers arrive on all other channels: finalize snapshot (record channel states)
  - When initiating a snapshot: record own state, send markers on all outgoing channels
- [ ] `coordinator.go`: wires 3 nodes in a ring (0 → 1 → 2 → 0), starts a goroutine per node, injects a sequence of data messages, then triggers a snapshot from node 0, and waits (via a `sync.WaitGroup` or a completion channel) for all nodes to report their recorded states
- [ ] `snapshot_test.go`: inject 10 data messages (total sum = 55), trigger a snapshot mid-stream, assert: (1) sum of all recorded local states + sum of all in-flight channel states = total messages sent so far; (2) snapshot completes without deadlock — run with `go test -race` to catch any channel misuse

**Constraints:** no third-party concurrency libraries — `chan`, `sync.WaitGroup`, and `select` are enough for this. Focus on correctness of the marker protocol, not performance.

---

## 🐍 Python DSA Review (optional)

**FIFO queue + BFS for snapshot reachability**: Chandy-Lamport requires FIFO channels; BFS verifies the snapshot is consistent (every in-flight message is accounted for).

```python
from collections import deque

# fifo_channel.py: what channel.go wraps a Go `chan` around — a FIFO queue with O(1) ops
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

**Connection:** Go's native `chan` already is this `deque`, with thread safety built in for free. The BFS is how you'd write `snapshot_test.go`'s consistency assertion if you wanted to do it graph-theoretically rather than just summing integers.

---

## Reflect

**What clicked:**

**What surprised me:**

**What "consistent global state" actually means and why it's useful:**

**How fault tolerance is handled in a system you know (checkpointing vs replay vs idempotency):**

**Where the lack of exhaustiveness checking on your `Message` type switch actually bit you, if it did:**

**What I'd do differently:**
