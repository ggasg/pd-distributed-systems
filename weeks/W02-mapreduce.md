---
week_number: 2
status: not-started
---

# W02: MapReduce and Its Limits

> **Arc:** Storage, Batch, and Failure · **Language:** Go
> **Budget:** about 5 hours. Hit the Minimum bar first; everything past it is optional.

## What you'll build
A minimal MapReduce framework in Go using goroutines and channels. Run two jobs through it: word count and iterative PageRank. Measure the disk I/O cost per PageRank iteration; that number is the concrete motivation for everything in Arc 2.

**Scenario:** the original MapReduce paper spends a full section on what happens when a worker dies mid-job, because at Google's scale a job with a thousand mappers will lose one eventually. Your framework runs in one process on one machine, so "a worker dies" doesn't mean the same thing here, which is exactly what the exercise below makes you reckon with.

**Note on why Go, specifically:** this isn't an arbitrary choice. MIT's 6.824/6.5840 (Distributed Systems), the field's most widely used academic treatment of exactly this material, has students build MapReduce in Go as its first lab, then Raft and a sharded key-value store on top of it in later labs. Goroutines also preserve the actual lesson this unit is built around better than most alternatives would: Go's scheduler multiplexes many cheap, garbage-collected goroutines onto a small number of OS threads, the same "spawn thousands of tasks without exhausting the OS" idea, just Go's own long-standing contribution to the field rather than a recent JVM addition.

---

## Read
- [ ] **DDIA Chapter 11** (2nd ed.): Batch Processing. Read this first, before either paper: it's the same MapReduce material and the Spark/RDD argument, told as one continuous narrative instead of two papers written four years apart, and it explicitly frames MapReduce as one point on a spectrum (Unix pipes → MapReduce → dataflow engines like Spark and Flink) rather than a standalone system. That framing is the throughline for the rest of Arc 1 and all of Arc 2.
- [ ] [MapReduce: Simplified Data Processing on Large Clusters](https://research.google/pubs/mapreduce-simplified-data-processing-on-large-clusters/) (Dean & Ghemawat, OSDI 2004): read Sections 1–4. The programming model is simple; pay attention to the fault tolerance mechanism and why it requires materializing intermediate state. Note what makes re-executing a failed task safe at all: the map and reduce functions are deterministic and the output commit is an atomic rename, so running a task twice produces the same result as running it once. That is at-least-once execution plus an idempotent commit, which W03 names properly, and it is the whole reason MapReduce can recover by simply retrying.
- [ ] Optional: [Resilient Distributed Datasets: A Fault-Tolerant Abstraction for In-Memory Cluster Computing](https://www.usenix.org/system/files/conference/nsdi12/nsdi12-final138.pdf) (Zaharia et al., NSDI 2012): read Sections 1–3. This is the Spark paper. The key argument is in Section 1: why MapReduce forces iterative algorithms to write to disk between every iteration.
- [ ] Optional: [MIT 6.5840 Lab 1: MapReduce](https://pdos.csail.mit.edu/6.824/labs/lab-mr.html): the assignment this unit's exercise is directly inspired by. Worth skimming the spec even though this unit's build is smaller in scope (single-process, not a real distributed worker pool), it's the canonical version of the same problem.

**Depth: study Section 3 of the MapReduce paper** (the fault-tolerance mechanism), since it carries this unit's actual argument about why intermediate state gets materialized. DDIA Ch.11 is a read. The RDD paper, the 2010 Spark paper, and the MIT lab are skims.

**Key question:** PageRank converges after ~50 iterations on typical graphs. In MapReduce, what happens between each iteration, and why does that make it slow? Write down the concrete I/O cost in terms of graph size G.

---

## Code

Project: `code/mapreduce/` (Go modules)

**Minimum bar:** the framework (`mapreduce.go` + `runner.go`) and word count running end to end, plus the written I/O calculation for PageRank at 1M nodes. Actually building and running PageRank is stretch, and so is the coordinator service.

A working *concrete* pipeline clears this unit. You don't need a beautifully generic, reusable `Mapper`/`Reducer` abstraction; that's a nice-to-have, not the point. The point is making the disk I/O cost visible.

**The framework:**

- [ ] `mapreduce.go`: two function types, so job definitions can be plain functions instead of interfaces implemented by a struct:
  - `type Mapper func(key, value string) []Pair` where `type Pair struct { Key, Value string }`
  - `type Reducer func(key string, values []string) string`
  - `Mapper` returns its output pairs directly rather than writing into a channel passed in as an argument. A function that returns its results is easier to test in isolation than one that writes into something it was handed.
- [ ] `runner.go`: `func Run(mapper Mapper, reducer Reducer, splits []string) map[string]string`. This is the fiddly part of the unit, and the specific way it's fiddly is worth naming directly:
  ```go
  results := make(chan []Pair, len(splits))
  var wg sync.WaitGroup
  for _, split := range splits {
      wg.Add(1)
      go func(s string) {
          defer wg.Done()
          results <- mapper(s, "")
      }(split)
  }
  go func() {
      wg.Wait()
      close(results)
  }()

  var allPairs []Pair
  for pairs := range results {
      allPairs = append(allPairs, pairs...)
  }
  // shuffle phase below
  ```
  One goroutine per split (cheap enough to spawn thousands without a thread pool), a `sync.WaitGroup` to know when every mapper has finished, and a separate goroutine that closes the `results` channel only after `wg.Wait()` returns. That last part is the one genuinely easy way to get this wrong: closing `results` too early, before every goroutine has sent its output, panics on the next send; closing it too late, or not at all, means the `for pairs := range results` loop below never terminates because the channel is never marked done. The `WaitGroup` plus a dedicated closer goroutine is the idiomatic Go fix, worth understanding by writing it once rather than copying it. The reduce phase follows the same shape: one goroutine per key instead of per split.
  - Shuffle phase: group pairs by key. There's no `groupBy` builtin in Go's standard library, this is a plain loop building a `map[string][]string` by hand, one of the few places Go asks you to write out what other languages give you as a one-liner:
    ```go
    shuffled := make(map[string][]string)
    for _, p := range allPairs {
        shuffled[p.Key] = append(shuffled[p.Key], p.Value)
    }
    ```
  - Write intermediate shuffle data to a temp file between map and reduce (this is the point: make the I/O cost visible). Go's `encoding/gob` (`gob.NewEncoder`/`gob.NewDecoder`) is the quickest path for structured data; if you'd rather write human-readable output while you're debugging, a hand-rolled line-based text format works too and makes `cat`-ing the temp file genuinely useful.

**Job 1: Word Count**

- [ ] `wordcount.go`: `Map` returns one `Pair{word, "1"}` per word; `Reduce` sums the "1"s
- [ ] Run on a large text file (download [Wikipedia dump excerpt](https://dumps.wikimedia.org/enwiki/latest/) or use any large `.txt`; aim for >10MB). Print top 20 words by frequency.

**Job 2: Iterative PageRank**

**Optional, stretch: iterative PageRank.** This is where the unit's argument lands, that MapReduce forces a disk round-trip between every iteration, but you can reach that conclusion from the word-count run plus the arithmetic below if time is short. Do the calculation either way; build it only if you have a second sitting.

- [ ] `pagerank.go`: one MapReduce iteration of PageRank: `Map` returns `(destination, rank/out_degree)` for each outgoing edge; `Reduce` sums contributions + applies damping factor `0.85`
- [ ] `pagerank_runner.go`: runs the PageRank job for 10 iterations over a hardcoded 1000-node graph (random edges, average degree 5). The part that isn't obvious: each iteration's `Run()` output (new ranks per node) becomes next iteration's input, but the edge list itself doesn't change round to round. Keep the graph structure separate from the ranks, and rebuild the per-iteration input by pairing current ranks with the fixed edge list. After each iteration, print: iteration number, sum of rank changes (convergence), **bytes written to disk for the shuffle file**
- [ ] In a comment: calculate what the disk I/O would be at 1M nodes. This is the argument for keeping intermediate state in memory rather than round-tripping it to disk between iterations, which is what Spark's RDDs are for.

**Stretch goal (optional): coordinator service**

- [ ] Optional: `tools/job_coordinator/main.go`: an HTTP server that accepts job submissions and tracks status, using `net/http` (standard library, no framework needed for two routes). Endpoints: `POST /job` (accepts `{"type": "wordcount"|"pagerank", "input": "path"}`, returns `{"job_id": "..."}`), `GET /job/{id}` (returns status + result when done). Your `runner.go` calls this server to report completion. Keep it under 100 lines; `encoding/json` handles the request/response bodies, small enough that hand-writing them would be more code, not less.

If you have time left after the minimum bar: this is a small rep of the master/worker split from the MapReduce paper itself (a coordinator process tracking job state separate from the workers doing the compute), and a preview of the shape you'll build for real in W14, where a control-plane process reconciling state against reported worker status is most of what a Kubernetes operator's reconcile loop does. Skipping it costs you nothing required; it's here because the parallel to W14 is worth having if the unit's going well.

**Break it, then decide:**
- [ ] Pick one mapper goroutine (say, the one handling the last split) and make it `panic("simulated worker crash")` instead of returning normally. Run the word count job. In Go, an unrecovered panic in any goroutine crashes the entire process, not just that goroutine, so watch what actually happens to the other splits that were still running concurrently. This is the real reason the MapReduce paper's fault tolerance re-executes just the failed task on another worker instead of letting one failure take down the whole job; your framework currently has no such isolation.
- [ ] Given that your framework runs single-process rather than across real distributed workers, is it worth adding a `recover()` around each mapper goroutine that logs the failure and re-runs just that split, or would that just be theater, since it wouldn't demonstrate the actual hard part (detecting a worker is gone over the network, deciding it's really dead and not just slow, reassigning its work) that a real distributed MapReduce has to solve? W03 builds a failure detector against exactly that ambiguity, so this is a question you'll get to answer with code rather than prose in W03. Make a call, and if you decide it's worth adding, implement it; if not, write down specifically what a real fix would need that a single-process `recover()` can't give you.

---

## Rehearse it in Python first (optional, 20 minutes)

> **Why this exists, and when it stops.** This unit builds in Go, which is the one language here you are still learning. Writing the shuffle's groupBy and PageRank's graph walk in Python first means that when the Go version misbehaves you already know whether the problem is the algorithm or the syntax, which is the single most useful thing to know at that moment. These sections appear only in the Go units (W02, W03, W06) and stop after W06, by which point Go should no longer be the thing in your way. Skip it whenever the algorithm is already obvious to you.

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

**Connection:** `runner.go` groups intermediate key-value pairs by hand; that's `group_by`. `pagerank.go` iterates over a graph; that's the adjacency list pattern. Getting these right in Python first clarifies the algorithm before adding goroutines and disk I/O.

---

## Reflect

**What clicked:**

**What surprised me:**

**Disk I/O per PageRank iteration on your 1000-node graph:** __ KB

**Extrapolated to 1M nodes:** __ GB per iteration × 10 iterations = __ GB total

**What would have to change for this cost to disappear entirely:**

**What I'd do differently:**
