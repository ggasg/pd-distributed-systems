---
week_number: 6
status: not-started
---

# W06 — Naiad and Timely Dataflow

> **Arc:** Streaming and Dataflow · **Language:** Scala

## What you'll build
A toy timely dataflow graph in Scala: two operators connected by edges, timestamps with (epoch, iteration) pairs, progress tracking via pointstamp dominance, and notification when a frontier advances.

---

## Read
- [ ] [Naiad: A Timely Dataflow System](https://dl.acm.org/doi/10.1145/2517349.2522738) (Murray et al., SOSP 2013) — read Sections 1–4 carefully. Section 2 defines the computation model. Section 3 defines the progress tracking protocol — this is the heart of it.
- [ ] Skim the [timely-dataflow Rust crate README](https://github.com/TimelyDataflow/timely-dataflow) — read enough to understand how `operator`, `notify_at`, and `frontier` are used in practice

**Key question:** What is a pointstamp? How does pointstamp dominance let nodes know when they've seen all messages for a given timestamp?

---

## Code

Project: `code/timely-toy/` (Scala 3, sbt)

- [ ] `Timestamp.scala` — case class `Timestamp(epoch: Int, iteration: Int)` with a `happensBefore` relation: `(e1, i1) < (e2, i2)` iff `e1 < e2 || (e1 == e2 && i1 < i2)` (total order for this toy; Naiad uses partial order)
- [ ] `Pointstamp.scala` — case class `Pointstamp(location: Int, timestamp: Timestamp)`. Implement `couldResultIn(other: Pointstamp, graph: Graph): Boolean` — conservative check based on graph paths
- [ ] `Operator.scala` — trait with `onMessage(msg: Message)` and `onNotification(ts: Timestamp)`. Two concrete operators: `MapOperator` (transforms messages) and `SinkOperator` (prints output)
- [ ] `ProgressTracker.scala` — maintains outstanding event counts per pointstamp; when a count drops to zero and no pointstamp could-result-in it, fires `onNotification` for that timestamp
- [ ] `DataflowTest.scala` — wire two operators: source → map → sink; send 3 messages at epoch 0; send a "done with epoch 0" signal; assert sink's `onNotification(Timestamp(0, 0))` fires after all messages are processed

**Constraints:** single-threaded. Focus on correctness of the progress tracking logic, not performance.

---

## 🐍 Python DSA Review (optional)

**Topological sort (Kahn's algorithm)** — dataflow operator scheduling requires topological ordering. Naiad's progress tracking operates on a DAG of operators.

```python
from collections import defaultdict, deque

# topo_sort.py — Kahn's algorithm: O(V + E)
def topological_sort(nodes: list, edges: list[tuple]) -> list:
    in_degree = defaultdict(int)
    adj = defaultdict(list)
    for u, v in edges:
        adj[u].append(v)
        in_degree[v] += 1

    # Start with all nodes that have no incoming edges
    q = deque(n for n in nodes if in_degree[n] == 0)
    order = []
    while q:
        node = q.popleft()
        order.append(node)
        for neighbor in adj[node]:
            in_degree[neighbor] -= 1
            if in_degree[neighbor] == 0:
                q.append(neighbor)

    if len(order) != len(nodes):
        raise ValueError("Cycle detected — not a valid dataflow DAG")
    return order

# Test: simple 4-operator pipeline
# input → filter → map → output
nodes = ["input", "filter", "map", "output"]
edges = [("input","filter"), ("filter","map"), ("map","output")]
assert topological_sort(nodes, edges) == ["input", "filter", "map", "output"]

# Test: two parallel branches merging
nodes2 = ["src", "A", "B", "join"]
edges2 = [("src","A"), ("src","B"), ("A","join"), ("B","join")]
result = topological_sort(nodes2, edges2)
assert result.index("src") < result.index("A") < result.index("join")
assert result.index("src") < result.index("B") < result.index("join")
```

**Connection:** Naiad's `ProgressTracker` schedules operator notifications in an order consistent with the dataflow graph — that's topological ordering. Your Scala `Operator` trait assumes operators fire in a valid schedule; this is where that schedule comes from.

---

## Reflect

**What clicked:**

**What surprised me:**

**What would break if you removed the could-result-in check?**

**How this maps to a dataflow or streaming system you've encountered:**

**What I'd do differently:**
