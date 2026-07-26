---
week_number: 6
status: not-started
---

# W06: Naiad and Timely Dataflow

> **Arc:** Streaming and Dataflow · **Language:** Java

## What you'll build
A toy timely dataflow graph in Java: two operators connected by edges, timestamps with (epoch, iteration) pairs, progress tracking via pointstamp dominance, and notification when a frontier advances.

**Scenario:** a notification firing one instant too early is the dataflow equivalent of a distributed commit that runs before every participant has actually agreed, it looks fine until the one time a message was still in flight. `couldResultIn` is the entire mechanism standing between "probably done" and "provably done," and the exercise below is where you find out what happens when it's gone.

---

## Read
- [ ] [Naiad: A Timely Dataflow System](https://dl.acm.org/doi/10.1145/2517349.2522738) (Murray et al., SOSP 2013): read Sections 1–4 carefully. Section 2 defines the computation model. Section 3 defines the progress tracking protocol; this is the heart of it.
- [ ] [PyTorch Autograd Engine source](https://github.com/pytorch/pytorch/blob/main/torch/csrc/autograd/engine.cpp): not a `timely-dataflow` port, there's no actively maintained JVM continuation of that lineage either, so this is a substitute, not an equivalent, and it's C++ regardless of your own build language this week. But it's a real production codebase solving the same core problem: dependency-counted execution over a DAG, firing a node once its inputs are ready. Read `Engine::execute` and how `ReadyQueue` decides what's runnable next; the parallel to a pointstamp's outstanding count reaching zero is close enough to be worth tracing through by hand.
- [ ] [Ray source: `core_worker.cc`](https://github.com/ray-project/ray/blob/master/src/ray/core_worker/core_worker.cc) and [`task_manager.cc`](https://github.com/ray-project/ray/blob/master/src/ray/core_worker/task_manager.cc): Ray's `CoreWorker` is a real, production dependency-tracking scheduler (also C++, same caveat as above): skim how `TaskManager` tracks pending dependencies and completes/fires a task once they're satisfied, the same "fire when the count of outstanding requirements hits zero" idea behind pointstamp dominance. Ray is Anyscale's product, so this is a direct tie to one of your target companies, and real source rather than a third-party summary.

**Key question:** What is a pointstamp? How does pointstamp dominance let nodes know when they've seen all messages for a given timestamp?

---

## Code

Project: `code/timely-toy/` (Java 21, Maven)

- [ ] `Timestamp.java`: `record Timestamp(int epoch, int iteration) {}` with a `boolean happensBefore(Timestamp other)` method: `epoch < other.epoch() || (epoch == other.epoch() && iteration < other.iteration())` (total order for this toy; Naiad uses a partial order). Records get `equals`/`hashCode` generated for you field-by-field, which is exactly what you want here since both fields are primitive `int`s, no manual hash implementation needed the way a hand-written class would require.
- [ ] `Pointstamp.java`: `record Pointstamp(int location, Timestamp timestamp) {}`. Implement `boolean couldResultIn(Pointstamp other, Graph graph)`, a conservative check based on graph paths. Because it's a record, it's already a valid `HashMap` key with correct `equals`/`hashCode`, unlike the hand-written hash specialization this exercise needs in some other languages.
- [ ] `Operator.java`: `abstract class Operator { abstract void onMessage(Message msg); abstract void onNotification(Timestamp ts); }`. Two concrete subclasses: `MapOperator` (transforms messages) and `SinkOperator` (prints output). This is a direct, one-to-one translation of the "base class with two subclasses" shape: `abstract class` plus `extends` in Java is that pattern's home idiom, unlike languages that only offer composition, this one doesn't require a redesign to express it.
- [ ] `ProgressTracker.java`: maintains outstanding event counts per pointstamp in a `Map<Pointstamp, Integer>`; when a count drops to zero and no pointstamp could-result-in it, fires `onNotification` for that timestamp.
- [ ] Test (`ProgressTrackerTest.java`, JUnit 5): wire two operators: source → map → sink; send 3 messages at epoch 0; send a "done with epoch 0" signal; assert the sink's `onNotification(new Timestamp(0, 0))` fires only after all messages are processed.

**Constraints:** single-threaded. Focus on correctness of the progress-tracking logic, not performance. Wire operator ownership explicitly: `ProgressTracker` (or whatever assembles the graph) holds a `List<Operator>`, and messages get passed by reference the normal Java way, an object reference, not a pointer you have to reason about the lifetime of. Java's garbage collector means there's no ownership-aliasing failure mode to design around here the way there is in a language with manual memory management; the actual discipline this week asks for is keeping the graph's wiring explicit and readable, not defending against a memory bug class that doesn't exist in this language.

**Break it, then decide:**
- [ ] Build a graph with a branch: source feeds two paths of different lengths (say, one direct edge to the sink, one routed through an extra `MapOperator` first) that both eventually reach the same downstream operator. Temporarily make `couldResultIn` always return `false` (pretend no path could still deliver more messages for a timestamp), and send messages down only the longer path. Watch `ProgressTracker` fire `onNotification` for that timestamp as soon as the short path's count hits zero, before the message still traveling the longer path has arrived. That premature notification is exactly the bug `couldResultIn` exists to prevent; put the real check back and confirm the notification now waits correctly.
- [ ] This toy's `couldResultIn` reasons about paths through a fixed, acyclic graph. Naiad's actual claim to fame is handling *cyclic* dataflow, loops for iterative computation, where a message can, in principle, keep coming back around forever. Without implementing loop support, reason through it: would a purely path-based `couldResultIn` even terminate on a graph with a cycle, and if not, what would need to change about how a pointstamp's "could still arrive" question gets answered? You don't need to solve this, just be able to say precisely where the acyclic assumption this toy relies on would break.

---

## 🐍 Python DSA Review (optional)

**Topological sort (Kahn's algorithm)**: dataflow operator scheduling requires topological ordering. Naiad's progress tracking operates on a DAG of operators.

```python
from collections import defaultdict, deque

# topo_sort.py: Kahn's algorithm, O(V + E)
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
        raise ValueError("Cycle detected, not a valid dataflow DAG")
    return order

# Test: simple 4-operator pipeline
# input -> filter -> map -> output
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

**Connection:** Naiad's `ProgressTracker` schedules operator notifications in an order consistent with the dataflow graph; that's topological ordering. Your Java `Operator` hierarchy assumes operators fire in a valid schedule; this is where that schedule comes from.

---

## Reflect

**What clicked:**

**What surprised me:**

**What you actually observed when you disabled `couldResultIn` (which notification fired early, and for what timestamp):**

**Where would a purely path-based `couldResultIn` break down on a cyclic graph?**

**How this maps to a dataflow or streaming system you've encountered:**

**What I'd do differently:**
