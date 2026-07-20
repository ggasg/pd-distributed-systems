---
week_number: 3
status: not-started
---

# W03: MapReduce and Its Limits

> **Arc:** Data Systems Internals · **Language:** Java

## What you'll build
A minimal MapReduce framework in Java using virtual threads. Run two jobs through it: word count and iterative PageRank. Measure the disk I/O cost per PageRank iteration; that number is the concrete motivation for everything in Arc 2.

---

## Read
- [ ] **DDIA Chapter 10**: Batch Processing. Read this first, before either paper: it's the same MapReduce material and the Spark/RDD argument, told as one continuous narrative instead of two papers written four years apart, and it explicitly frames MapReduce as one point on a spectrum (Unix pipes → MapReduce → dataflow engines like Spark and Flink) rather than a standalone system. That framing is the throughline for the rest of Arc 1 and all of Arc 2.
- [ ] [MapReduce: Simplified Data Processing on Large Clusters](https://research.google/pubs/mapreduce-simplified-data-processing-on-large-clusters/) (Dean & Ghemawat, OSDI 2004): read Sections 1–4. The programming model is simple; pay attention to the fault tolerance mechanism and why it requires materializing intermediate state.
- [ ] [Resilient Distributed Datasets: A Fault-Tolerant Abstraction for In-Memory Cluster Computing](https://www.usenix.org/system/files/conference/nsdi12/nsdi12-final138.pdf) (Zaharia et al., NSDI 2012): read Sections 1–3. This is the Spark paper. The key argument is in Section 1: why MapReduce forces iterative algorithms to write to disk between every iteration.

**Key question:** PageRank converges after ~50 iterations on typical graphs. In MapReduce, what happens between each iteration, and why does that make it slow? Write down the concrete I/O cost in terms of graph size G.

---

## Code

Project: `code/mapreduce/` (Java 21, Maven)

**Minimum bar:** the framework (`MapReduce.java` + `Runner.java`), word count running end to end, and PageRank running 10 iterations with disk I/O measured and printed. The coordinator service is a stretch goal, not required; see below.

A working *concrete* pipeline clears this week. You don't need a beautifully generic, reusable `Mapper`/`Reducer` abstraction; that's a nice-to-have, not the point. The point is making the disk I/O cost visible.

**The framework:**

- [ ] `MapReduce.java`: two functional interfaces, so job definitions can be plain lambdas instead of full classes:
  - `@FunctionalInterface interface Mapper { List<Pair> map(String key, String value); }` where `record Pair(String key, String value) {}`
  - `@FunctionalInterface interface Reducer { String reduce(String key, List<String> values); }`
  - `map` returns its output pairs directly rather than writing into a passed-in "emitter." A function that returns its results is easier to test in isolation than one that mutates something it was handed, in Java exactly as much as anywhere else.
- [ ] `Runner.java`: `static Map<String, String> run(Mapper mapper, Reducer reducer, List<String> splits)`. This is the fiddly part of the week in every language's version of this exercise, but the specific way it's fiddly is different in Java than in Go, worth naming directly:
  ```java
  try (var executor = Executors.newVirtualThreadPerTaskExecutor()) {
      List<Callable<List<Pair>>> tasks = splits.stream()
          .map(split -> (Callable<List<Pair>>) () -> mapper.map(split, ""))
          .toList();
      List<Future<List<Pair>>> futures = executor.invokeAll(tasks);
      List<Pair> allPairs = new ArrayList<>();
      for (var future : futures) {
          allPairs.addAll(future.get());
      }
      // shuffle phase below
  }
  ```
  `ExecutorService.invokeAll` submits one virtual thread per split (the direct equivalent of one goroutine per split) and blocks until every one of them has finished, returning their results directly. There's no separate "close the channel when everyone's done" step to get wrong, because there's no channel: the executor's own lifecycle *is* the synchronization. This sidesteps the exact bug class the Go version of this exercise is built to teach (closing a channel before every writer is done, which panics or deadlocks depending on timing); it's worth being honest that this isn't a case of Java being harder, here Java's higher-level concurrency API removes an entire category of mistake by construction, a genuine, specific place where the language comparison runs the other way. The reduce phase follows the same shape: one virtual thread per key instead of per split.
  - Shuffle phase: group pairs by key. Java's Streams API makes this a one-liner instead of a hand-rolled loop, and it's a genuinely functional-style way to express a groupBy: `Map<String, List<String>> shuffled = allPairs.stream().collect(Collectors.groupingBy(Pair::key, Collectors.mapping(Pair::value, Collectors.toList())));`
  - Write intermediate shuffle data to a temp file between map and reduce (this is the point: make the I/O cost visible). Java's built-in `Serializable` plus `ObjectOutputStream`/`ObjectInputStream` is the quickest path; if you'd rather write human-readable output while you're debugging, a hand-rolled line-based text format works too and makes `cat`-ing the temp file genuinely useful.

**Job 1: Word Count**

- [ ] `WordCount.java`: `map` returns one `Pair(word, "1")` per word; `reduce` sums the "1"s
- [ ] Run on a large text file (download [Wikipedia dump excerpt](https://dumps.wikimedia.org/enwiki/latest/) or use any large `.txt`; aim for >10MB). Print top 20 words by frequency.

**Job 2: Iterative PageRank**

- [ ] `PageRank.java`: one MapReduce iteration of PageRank: `map` returns `(destination, rank/out_degree)` for each outgoing edge; `reduce` sums contributions + applies damping factor `0.85`
- [ ] `PageRankRunner.java`: runs the PageRank job for 10 iterations over a hardcoded 1000-node graph (random edges, average degree 5). The part that isn't obvious: each iteration's `run()` output (new ranks per node) becomes next iteration's input, but the edge list itself doesn't change round to round. Keep the graph structure separate from the ranks, and rebuild the per-iteration input by pairing current ranks with the fixed edge list. After each iteration, print: iteration number, sum of rank changes (convergence), **bytes written to disk for the shuffle file**
- [ ] In a comment: calculate what the disk I/O would be at 1M nodes. This is the argument for keeping intermediate state in memory (Spark) or as a live dataflow (Naiad/DD).

**Stretch goal (optional): coordinator service**

- [ ] `tools/job_coordinator/JobCoordinator.java`: an HTTP server that accepts job submissions and tracks status, using `com.sun.net.httpserver.HttpServer` (built into the JDK, no framework needed for two routes). Endpoints: `POST /job` (accepts `{"type": "wordcount"|"pagerank", "input": "path"}`, returns `{"job_id": "..."}`), `GET /job/{id}` (returns status + result when done). Your `Runner.java` calls this server to report completion. Keep it under 100 lines, JDK only, no JSON library required at this size (a two-or-three-field object is easy enough to hand-write).

If you have time left after the minimum bar: this is a small rep of the master/worker split from the MapReduce paper itself (a coordinator process tracking job state separate from the workers doing the compute), and a preview of the shape you'll build for real in W19, where a control-plane process reconciling state against reported worker status is most of what a Kubernetes operator's reconcile loop does. Skipping it costs you nothing required; it's here because the parallel to W19 is worth having if the week's going well.

---

## 🐍 Python DSA Review (optional)

**Hash maps (groupBy) + adjacency list BFS**: the shuffle phase is a groupBy; PageRank needs a graph.

```python
from collections import defaultdict, deque

# map_reduce.py: the shuffle phase in pure Python
def group_by(pairs: list[tuple]) -> dict:
    groups = defaultdict(list)
    for k, v in pairs:
        groups[k].append(v)
    return dict(groups)

# graph.py: adjacency list + BFS (PageRank iteration needs this)
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

**Connection:** `Runner.java` groups intermediate key-value pairs; that's `group_by`, and it's the same operation Java's `Collectors.groupingBy` does above, just spelled out by hand here. `PageRank.java` iterates over a graph; that's the adjacency list pattern. Getting these right in Python first clarifies the algorithm before adding virtual threads and disk I/O.

---

## Reflect

**What clicked:**

**What surprised me:**

**Disk I/O per PageRank iteration on your 1000-node graph:** __ KB

**Extrapolated to 1M nodes:** __ GB per iteration × 10 iterations = __ GB total

**Why this problem disappears in Differential Dataflow:**

**What I'd do differently:**
