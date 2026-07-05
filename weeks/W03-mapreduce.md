---
week_number: 3
status: not-started
---

# W03 — MapReduce and Its Limits

> **Arc:** Data Systems Internals · **Language:** Java 21

## What you'll build
A minimal MapReduce framework in Java 21 using virtual threads. Run two jobs through it: word count and iterative PageRank. Measure the disk I/O cost per PageRank iteration — that number is the concrete motivation for everything in Arc 2.

---

## Read
- [ ] [MapReduce: Simplified Data Processing on Large Clusters](https://research.google/pubs/mapreduce-simplified-data-processing-on-large-clusters/) (Dean & Ghemawat, OSDI 2004) — read Sections 1–4. The programming model is simple; pay attention to the fault tolerance mechanism and why it requires materializing intermediate state.
- [ ] [Resilient Distributed Datasets: A Fault-Tolerant Abstraction for In-Memory Cluster Computing](https://www.usenix.org/system/files/conference/nsdi12/nsdi12-final138.pdf) (Zaharia et al., NSDI 2012) — read Sections 1–3. This is the Spark paper. The key argument is in Section 1: why MapReduce forces iterative algorithms to write to disk between every iteration.

**Key question:** PageRank converges after ~50 iterations on typical graphs. In MapReduce, what happens between each iteration, and why does that make it slow? Write down the concrete I/O cost in terms of graph size G.

---

## Code

Project: `code/mapreduce/` (Java 21, virtual threads)

**The framework:**

- [ ] `MapReduceJob.java` — generic class `MapReduceJob<K1,V1,K2,V2,V3>` with:
  - `map(K1 key, V1 value, Emitter<K2,V2> emitter)` — override this
  - `reduce(K2 key, List<V2> values)` — override this, returns `V3`
- [ ] `MapReduceRunner.java` — executes a job:
  - Map phase: spawn one virtual thread per input split (use `List<String>` lines as splits, chunk into groups of 1000), collect `(K2, V2)` pairs
  - Shuffle phase: group pairs by key into `Map<K2, List<V2>>`
  - Reduce phase: spawn one virtual thread per key, run reduce, collect results
  - Write intermediate shuffle data to a temp file between map and reduce (this is the point — make the I/O cost visible)
- [ ] `Serializer.java` — serialize/deserialize `Map<String, List<Integer>>` to/from a temp file (use Java serialization or JSON via `com.fasterxml.jackson`)

**Job 1 — Word Count:**

- [ ] `WordCountJob.java` — map: emit `(word, 1)` per word; reduce: sum the 1s
- [ ] Run on a large text file (download [Wikipedia dump excerpt](https://dumps.wikimedia.org/enwiki/latest/) or use any large `.txt`; aim for >10MB). Print top 20 words by frequency.

**Job 2 — Iterative PageRank:**

- [ ] `PageRankJob.java` — one MapReduce iteration of PageRank: map emits `(destination, rank/out_degree)` for each outgoing edge; reduce sums contributions + applies damping factor `0.85`
- [ ] `PageRankRunner.java` — runs PageRankJob for 10 iterations over a hardcoded 1000-node graph (random edges, average degree 5). After each iteration, print: iteration number, sum of rank changes (convergence), **bytes written to disk for the shuffle file**.
- [ ] In comments: calculate what the disk I/O would be at 1M nodes. This is the argument for keeping intermediate state in memory (Spark) or as a live dataflow (Naiad/DD).

**Go automation tool:**

- [ ] `tools/job_coordinator/main.go` — a Go HTTP server that accepts job submissions and tracks status. Endpoints: `POST /job` (accepts `{"type": "wordcount"|"pagerank", "input": "path"}`, returns `{"job_id": "..."}`), `GET /job/{id}` (returns status + result when done). The Java MR runner calls this server to report completion. Keep it under 100 lines — use only `net/http` and `encoding/json` from stdlib.

This is your first Go program. Notice how little boilerplate HTTP serving requires in Go compared to Java.

---

## 🐍 Python DSA Review (optional)

**Hash maps (groupBy) + adjacency list BFS** — the shuffle phase is a groupBy; PageRank needs a graph.

```python
from collections import defaultdict, deque

# map_reduce.py — the shuffle phase in pure Python
def group_by(pairs: list[tuple]) -> dict:
    groups = defaultdict(list)
    for k, v in pairs:
        groups[k].append(v)
    return dict(groups)

# graph.py — adjacency list + BFS (PageRank iteration needs this)
def bfs(graph: dict, start) -> set:
    visited, q = {start}, deque([start])
    while q:
        node = q.popleft()
        for neighbor in graph.get(node, []):
            if neighbor not in visited:
                visited.add(neighbor)
                q.append(neighbor)
    return visited

# Test groupBy (the "shuffle" step)
pairs = [("a", 1), ("b", 2), ("a", 3), ("b", 4)]
assert group_by(pairs) == {"a": [1, 3], "b": [2, 4]}

# Test BFS on a 4-node graph
graph = {"A": ["B", "C"], "B": ["D"], "C": [], "D": []}
assert bfs(graph, "A") == {"A", "B", "C", "D"}
```

**Connection:** the Java `MapReduceRunner` groups intermediate key-value pairs — that's `group_by`. `PageRank.java` iterates over a graph — that's the adjacency list pattern. Getting these right in Python first clarifies the algorithm before adding virtual threads and disk I/O.

---

## Reflect

**What clicked:**

**What surprised me:**

**Disk I/O per PageRank iteration on your 1000-node graph:** __ KB

**Extrapolated to 1M nodes:** __ GB per iteration × 10 iterations = __ GB total

**Why this problem disappears in Differential Dataflow:**

**What I'd do differently:**
