---
week_number: 6
status: not-started
---

# W06: Naiad and Timely Dataflow

> **Arc:** Streaming and Dataflow · **Language:** C++

## What you'll build
A toy timely dataflow graph in C++: two operators connected by edges, timestamps with (epoch, iteration) pairs, progress tracking via pointstamp dominance, and notification when a frontier advances.

---

## Read
- [ ] [Naiad: A Timely Dataflow System](https://dl.acm.org/doi/10.1145/2517349.2522738) (Murray et al., SOSP 2013): read Sections 1–4 carefully. Section 2 defines the computation model. Section 3 defines the progress tracking protocol; this is the heart of it.
- [ ] [PyTorch Autograd Engine source](https://github.com/pytorch/pytorch/blob/main/torch/csrc/autograd/engine.cpp): not a `timely-dataflow` port — there's no actively maintained C++ continuation of that lineage, so this is a substitute, not an equivalent. But it's a real production codebase solving the same core problem: dependency-counted execution over a DAG, firing a node once its inputs are ready. Read `Engine::execute` and how `ReadyQueue` decides what's runnable next; the parallel to a pointstamp's outstanding count reaching zero is close enough to be worth tracing through by hand.
- [ ] [Ray Core Worker & Task Execution](https://deepwiki.com/ray-project/ray/2.1-core-worker-and-task-execution): Ray's `CoreWorker` is a real, production C++ dependency-tracking scheduler — a task fires once its arguments (its dependencies) are available, the same "fire when the count of outstanding requirements hits zero" idea behind pointstamp dominance. Ray is Anyscale's product, so this is a direct tie to one of your target companies, not an analogy borrowed from elsewhere.

**Key question:** What is a pointstamp? How does pointstamp dominance let nodes know when they've seen all messages for a given timestamp?

---

## Code

Project: `code/timely-toy/` (C++, CMake + GoogleTest)

- [ ] `include/timely_toy/timestamp.hpp`: `struct Timestamp { int32_t epoch; int32_t iteration; };` with a `bool happens_before(const Timestamp& other) const` method: `(epoch < other.epoch) || (epoch == other.epoch && iteration < other.iteration)` (total order for this toy; Naiad uses a partial order). Implement `operator==` as well — you'll need it for hashing and equality in the progress tracker.
- [ ] `include/timely_toy/pointstamp.hpp`: `struct Pointstamp { size_t location; Timestamp timestamp; };`. Implement `bool could_result_in(const Pointstamp& other, const Graph& graph) const`, a conservative check based on graph paths. Provide a `std::hash<Pointstamp>` specialization (or an equivalent comparator), since instances of this type become map keys.
- [ ] `include/timely_toy/operator.hpp`: `class Operator { public: virtual void on_message(const Message& msg) = 0; virtual void on_notification(const Timestamp& ts) = 0; virtual ~Operator() = default; };`. Two concrete subclasses: `MapOperator` (transforms messages) and `SinkOperator` (prints output). Wire them as `std::unique_ptr<Operator>` when building the graph.
- [ ] `include/timely_toy/progress_tracker.hpp` + `src/progress_tracker.cpp`: `class ProgressTracker` maintains outstanding event counts per pointstamp in a `std::unordered_map<Pointstamp, int32_t, PointstampHash>`; when a count drops to zero and no pointstamp could-result-in it, fires `on_notification` for that timestamp.
- [ ] Test (`tests/progress_tracker_test.cpp`, GoogleTest): wire two operators: source → map → sink; send 3 messages at epoch 0; send a "done with epoch 0" signal; assert the sink's `on_notification({0, 0})` fires only after all messages are processed.

**Constraints:** single-threaded. Focus on correctness of the progress-tracking logic, not performance. If you find yourself reaching for `std::shared_ptr` with operators holding pointers back into each other, stop and reconsider the ownership shape first — for a two-operator toy graph you can almost always thread ownership through explicitly (`ProgressTracker` owns the operators in a `std::vector<std::unique_ptr<Operator>>`; messages get passed by `const&` or by value at the call site) instead of reaching for shared, aliased ownership as a default. Rust made the undisciplined version (`Rc<RefCell<_>>` fighting the borrow checker) a compile error; in C++ the equivalent will compile fine and you'll pay for the aliasing bugs at runtime instead. Treat the same discipline as a design rule you enforce yourself, not a guarantee the language gives you.

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

**Connection:** Naiad's `ProgressTracker` schedules operator notifications in an order consistent with the dataflow graph; that's topological ordering. Your C++ `Operator` hierarchy assumes operators fire in a valid schedule; this is where that schedule comes from.

---

## Reflect

**What clicked:**

**What surprised me:**

**What would break if you removed the could-result-in check?**

**How this maps to a dataflow or streaming system you've encountered:**

**What I'd do differently:**
