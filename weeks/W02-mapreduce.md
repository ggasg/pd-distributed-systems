---
week_number: 2
status: not-started
---

# W02: MapReduce and Its Limits

> **Arc:** Storage, Batch, and Failure · **Language:** Python (PySpark)
> **Budget:** about 10 hours. The Minimum bar is what a bad week looks like, not the target.

## What you'll build
Not a MapReduce framework. An iterative job on real Spark, run twice, and the arithmetic that explains the gap between the two runs.

MapReduce's defining cost is that intermediate state has to hit disk between stages, which means an iterative algorithm pays that cost once per iteration. The RDD paper's entire argument is that this is the thing worth fixing. This unit has you measure the size of that argument on your own machine, in the engine that was built to win it.

**Why not write the framework?** An earlier version of this unit had you build a `Mapper`/`Reducer` abstraction, a hand-rolled shuffle, and a temp-file round trip in Go. That is real work, and at the end of it you own a MapReduce framework, which is a thing you will never write again and never operate. Hadoop is legacy and Spark superseded it. What you actually need is the cost model, precise enough to predict which jobs will be slow before you run them, plus the ability to look at a stage DAG and see where the money goes. Spark gives you both, and it gives you the second one for free because it draws you the picture.

**Scenario:** a colleague's nightly model-refresh job runs fifty iterations of an algorithm over the same dataset and takes six hours. They ask whether more executors will fix it. The answer depends entirely on whether the job is re-reading its input every iteration, and you can tell from the stage DAG in about thirty seconds once you know what to look at.

---

## Read
- [ ] **DDIA Chapter 11** (2nd ed.): Batch Processing. Read this first, before the papers: it is the same MapReduce material and the Spark/RDD argument told as one continuous narrative instead of two papers written eight years apart, and it explicitly frames MapReduce as one point on a spectrum (Unix pipes, then MapReduce, then dataflow engines like Spark and Flink) rather than a standalone system. That framing is the throughline for the rest of Arc 1 and all of Arc 2.
- [ ] [MapReduce: Simplified Data Processing on Large Clusters](https://research.google/pubs/mapreduce-simplified-data-processing-on-large-clusters/) (Dean & Ghemawat, OSDI 2004): read Sections 1 to 4. The programming model is simple; pay attention to the fault-tolerance mechanism and why it requires materializing intermediate state. Note what makes re-executing a failed task safe at all: the map and reduce functions are deterministic and the output commit is an atomic rename, so running a task twice produces the same result as running it once. That is at-least-once execution plus an idempotent commit, which W03 names properly, and it is the whole reason MapReduce can recover by simply retrying.
- [ ] [Resilient Distributed Datasets: A Fault-Tolerant Abstraction for In-Memory Cluster Computing](https://www.usenix.org/system/files/conference/nsdi12/nsdi12-final138.pdf) (Zaharia et al., NSDI 2012): read Sections 1 to 3. This is the Spark paper, and it is now required rather than optional, because its Section 1 argument is precisely the thing you are about to measure. It also introduces lineage, which is what makes the `checkpoint` decision at the bottom of this unit a real decision rather than a formality.
- [ ] Optional: **Burns, *Designing Distributed Systems*, 2nd ed., Chapter 13** (Coordinated Batch Processing). Short, and it gives you the vocabulary the papers assume: join as barrier synchronization, and reduce as a pattern rather than a function name. Useful mainly because "the barrier is the expensive part" is the same observation you are about to make from a stage DAG.

**Depth: study Section 3 of the MapReduce paper** (the fault-tolerance mechanism) and **Section 1 of the RDD paper** (why iterative algorithms suffer). Those two carry the unit. DDIA Ch.11 is a read.

**Key question, and do this before you run anything:** PageRank converges after roughly 50 iterations on typical graphs. In MapReduce, what happens to the intermediate state between each iteration, and why does that make it slow? Write down the concrete I/O cost in terms of graph size G and iteration count N. Keep the number. You will check it against a real DAG in a moment.

---

## Code

Project: `code/batch-spark/` (Python 3.12, PySpark 4.1.0)

PySpark, because it is the API Spark is actually driven with. The engine, the physical plan, and the stage DAG are identical whichever surface you use, so the only thing the choice changes is how much ceremony sits between you and the measurement. That argues for the surface with the most examples, the most documentation, and the least typing.

**Minimum bar:** the written I/O calculation from the Key question, plus two wall-clock numbers for the same 10-iteration job, cached and uncached, plus one sentence naming what the Spark UI showed you was different between them. That is the unit.

**Setup:**

- [ ] `pip install "pyspark==4.1.0"`. Spark runs on the JVM regardless of the Python surface, so the Java 21 you installed for W04 satisfies it; there is no second JVM to manage.
- [ ] If you hit `InaccessibleObjectException` on the first run, that is Spark reaching into internal JDK APIs the module system closed off. Set `--add-opens` flags via `spark.driver.extraJavaOptions` in your `SparkSession` builder rather than discovering them one stack trace at a time.

**The job:**

- [ ] `graph_gen.py`: generate a random directed graph, 100,000 nodes, average out-degree 5, written once to Parquet as `(src, dst)` pairs. Fixed seed, so both runs below see identical input.
- [ ] `pagerank.py`: ten iterations of PageRank using the DataFrame API. Each iteration joins ranks to edges, divides each rank by its out-degree, aggregates contributions per destination, and applies damping at 0.85. The shape that matters: the edge list never changes across iterations, and the ranks do. Keep them as separate DataFrames.
- [ ] Run it once with no caching at all. Then run it again with the edge DataFrame cached (`edges.cache()` before the loop, and force materialization with a `count()` so the caching actually happens before the first iteration rather than during it).
- [ ] Record wall-clock time for both.

**Read the DAG. This is the actual unit:**

- [ ] While the job runs, open the Spark UI at `localhost:4040`. Keep it open after the job finishes; Spark serves it until the driver exits, so put an `input()` call at the end of the script if you need time to look.
- [ ] In the **Stages** tab, count the stages for the uncached run. Then count them for the cached run. The difference is the shape of the argument you read in the RDD paper, drawn for you.
- [ ] Open one iteration's stage and find **Shuffle Write** and **Shuffle Read** in the summary metrics. Multiply by iterations. Compare that number to the I/O cost you predicted in the Key question above. If you were wrong, work out which term you got wrong before reading on; being wrong here is more instructive than being right, because the usual error is forgetting that the edge list gets re-read too.
- [ ] In the **SQL / DataFrame** tab, open the query plan for one iteration and find the `Exchange` nodes. Every one of them is an all-to-all data movement. W05 builds one of those by hand and W07 is about choosing how many you get.

**Break it, then decide:**

- [ ] Push the loop from 10 iterations to 100 on the uncached version and watch what happens to the *planning* time, separately from the execution time. Spark tracks lineage, the full recipe for recomputing any DataFrame from its inputs, and that lineage grows with every iteration. Long enough chains stop being merely slow and start failing outright with a `StackOverflowError` in the driver during plan construction, before a single row is processed. This is a real production failure mode in iterative Spark jobs and it surprises people, because the job worked fine at 20 iterations.
- [ ] **Your call:** there are two fixes and they are not the same. `cache()` keeps the *data* around but preserves the lineage, so recovery after an executor loss is still possible by recomputation, and the driver's plan keeps growing. `checkpoint()` writes the data to reliable storage and *truncates* the lineage, so the plan stays small and recovery is a read rather than a recomputation, but you have paid a full write to disk and you cannot recover if that storage is lost. Given that you are running 100 iterations and the input is cheap to regenerate, say which you would use and at what iteration interval. Then say what would change your answer if the input were an expensive upstream join rather than a generated graph.
- [ ] The MapReduce paper recovers from a dead worker by re-running its task, which is safe because the function is deterministic and the commit is an atomic rename. Spark's lineage recovery is the same idea generalized. Write down what breaks in both schemes if the map function is *not* deterministic, for example if it reads the current time or samples a random number without a fixed seed. This is the assumption underneath every retry in this curriculum, and W03 is where you meet it again from the delivery-semantics side.

---

## Reflect
<!-- Fill in at the end of the unit -->

**Prediction versus measurement.** Fill the predictions in *before* you run anything, and do not edit them afterwards. The gap is where calibration comes from.

| Quantity | Predicted | Measured | Which term I got wrong |
|----------|-----------|----------|------------------------|
| | | | |

Copy anything worth carrying into [MEASUREMENTS.md](../MEASUREMENTS.md).


**What clicked:**

**What surprised me:**

**Predicted disk I/O per iteration, from the Key question:** __

**Actual Shuffle Write per iteration, from the Spark UI:** __ , and where my prediction was wrong:

**Wall clock, 10 iterations, uncached versus cached:** __ s versus __ s

**What the stage count showed that the wall-clock number did not:**

**`cache()` or `checkpoint()` at 100 iterations, and what would change my mind:**

---

## Review and articulate

Two steps that exist because self-study has no examiner. Do them at the end of every unit, before marking it done.

- [ ] **Adversarial review.** Hand over three things separately: the number you predicted, the number you measured, and the conclusion you drew. Then ask for the strongest case that the conclusion is *not* supported by the measurement. Do not ask whether you are right; ask what would falsify this. An assistant asked to check your work will tend to find support for your framing, so the prompt has to be adversarial by construction or the exercise is theatre.
- [ ] **Ninety seconds, out loud, timed.** Explain this unit's finding as you would to someone in an interview or a design review: what you measured, what surprised you, and what decision it would change. Articulation under time pressure is a separate skill from understanding, and it is the one that gets tested. If you cannot do it in ninety seconds you do not have the finding yet, you have notes.
