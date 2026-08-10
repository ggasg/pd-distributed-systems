---
week_number: 5
status: not-started
---

# W05: Partitioning and the Shuffle

> **Arc:** Data Movement and Execution · **Language:** Java (Part 1) / Python (Part 2, PySpark)
> **Budget:** about 10 hours. The Minimum bar is what a bad week looks like, not the target.

## What you'll build
Two halves, and they are deliberately different in kind.

**Part 1, the mechanism, which you build.** A working shuffle in Java: map tasks that split their output into per-partition spill files, reduce tasks that fetch their own partition from every map task, and a partitioner that decides which key goes where. This is the one build in Arc 2 that survives intact, because the shuffle is the mechanism every other unit in the arc refers back to, and there is no way to see the write-then-fetch structure from outside a system that does it for you.

**Part 2, the incident, which you diagnose on real Spark.** Reproduce the same skew in Spark, find the straggler task in the Spark UI, and fix it. Building the shuffle teaches you what the machine is doing. Reading the Spark UI is how you will actually find out that it is doing it badly, at four in the afternoon, on someone else's job.

If "shuffle" is a word you've heard without a firm picture behind it, here it is in one sentence: a shuffle is the all-to-all data movement that happens when every machine holds some of the rows for a key and they all need to end up on one machine so the key can be aggregated. It is the single most expensive thing a distributed query engine does, and it is the mechanism underneath `GROUP BY`, `JOIN`, and `reduceByKey` in every system you're likely to touch.

**Scenario:** a nightly aggregation job that normally finishes in twenty minutes has been taking three hours since last Tuesday. Every executor finished quickly except one, which is still going. Nothing errored, no config changed, and the data volume grew by four percent. This is the most common performance incident in data engineering, and by the end of the unit you will have caused it deliberately, found it the way you would at work, and fixed it.

---

## Read
- [ ] DDIA Ch.7 (2nd ed.), "Sharding": the whole chapter. Key-range vs hash sharding, the hot-spot problem, and rebalancing. Note that Kleppmann is deliberately cool on consistent hashing for databases; read his reasoning rather than assuming consistent hashing is the default right answer.
- [ ] [Spark RDD Programming Guide, "Shuffle operations"](https://spark.apache.org/docs/latest/rdd-programming-guide.html#shuffle-operations): short, concrete, and describes exactly the map-side-write then reduce-side-fetch structure you're about to build. Read the "Performance Impact" subsection twice.
- [ ] **Burns, *Designing Distributed Systems*, 2nd ed., Chapter 7** (Sharded Services): read "An Examination of Sharding Functions," "Selecting a Key," and "Hot Sharding Systems." This is the closest thing in the pattern literature to what both halves of this unit do. Kleppmann tells you what sharding is and why hot spots happen; Burns tells you what to do about it in a running system, and his hot-sharding section is the production answer to the straggler you are about to find in the Spark UI.

**Depth: study DDIA Ch.7.** You build a partitioner and then reproduce the hot-spot problem the chapter describes, so the two reinforce each other directly. Burns Ch.7 and the Spark shuffle page are short reads. Dynamo is an optional skim.

**Key question:** Why does a shuffle write to disk at all, instead of streaming records straight from map tasks to reduce tasks over the network? What breaks if you don't?

---

## Part 1: Build the shuffle

Project: `code/shuffle/` (Java 21, Maven)

Data model: `record Record(String key, int value) {}`. The job is a word-count-shaped aggregation, sum of `value` per `key`, deliberately simple so the mechanics stay the subject.

- [ ] `Partitioner.java`: `sealed interface Partitioner permits HashPartitioner, RangePartitioner { int partitionFor(String key); }`. This is the same sealed-interface idiom you used for `StreamItem` in W04, reused rather than newly introduced. `HashPartitioner` is `Math.floorMod(key.hashCode(), numPartitions)`; use `floorMod`, not `%`, because `hashCode()` can be negative and `%` in Java preserves the sign, which would hand you a negative array index. `RangePartitioner` takes a sorted array of split-point keys and binary-searches (`Arrays.binarySearch`) to find the partition.
- [ ] `MapTask.java`: takes a `List<Record>` and a `Partitioner`, and writes one file per reduce partition into `spill/map-<mapId>/part-<partitionId>`. Each file is plain text, one `key,value` pair per line. Writing R files instead of one is the whole point: a reduce task should be able to read only what belongs to it.
- [ ] `ReduceTask.java`: for partition `p`, read `spill/map-*/part-<p>` from every map task, sum values per key, and write `output/part-<p>`. Return the number of records it processed, you will need that number later.
  - Write the aggregation itself as `Collectors.groupingBy(Record::key, Collectors.summingInt(Record::value))`. One line, and it is worth pausing on: **that line is the entire reduce side of a shuffle.** Everything else you are building in Part 1, the partitioner, the spill files, the fetch, exists for one reason, which is that you cannot call it. The data does not fit in one JVM, so the grouping has to be split across machines and the rows for a key have to be brought together first. A shuffle is `groupingBy` for data larger than memory, and having written both versions three lines apart is the clearest way to hold that.
- [ ] `Shuffle.java`: wire M map tasks and R reduce tasks together and run them. Keep it single-JVM and run the tasks on plain threads (or sequentially, which is fine and easier to debug). The network is simulated by the filesystem here, deliberately: real spill-to-disk then fetch is what Spark actually does, so this is a simplification of scale, not of structure.

**Constraints:** no Spark, no Hadoop, no external dependencies beyond JUnit 5 in this half. Standard library only. Keep every class under 100 lines; if one is growing past that, the shuffle logic has probably leaked into it from somewhere it doesn't belong.

**Minimum bar (Part 1):** the job runs end to end across M map tasks and R reduce tasks, over uniformly random keys, and you can point at the `spill/` directory and explain why there are M times R files in it.

**Break it, then decide:**
- [ ] Delete one `spill/map-<id>/part-<p>` file after the map phase and before the reduce phase, then run the reduce. Your `ReduceTask` reads `spill/map-*/part-<p>` with a glob, so it will not fail; it will quietly produce a smaller sum. Nothing errors. This is why real systems track which map outputs exist rather than trusting a directory listing, and it is the same idempotency-and-commit question the MapReduce paper answers with an atomic rename in W02.
- [ ] **Your call:** your `HashPartitioner` uses plain modulo, which is correct here because the partition count is fixed for the whole job. Say what would have to change about the system for that to stop being true, and what DDIA Ch.7 says you would reach for instead. The chapter is deliberately cool on consistent hashing for databases; engage with its reasoning rather than assuming the hash ring is the grown-up answer.

---

## Part 2: Find the skew in real Spark

Project: `code/shuffle-skew/` (Python 3.12, PySpark 4.1.0, local mode)

PySpark here even though Part 1 is Java, and the switch is the point rather than an inconsistency. Part 1 is a mechanism you author, where the language's type system does real work. Part 2 is an engine you drive and a UI you read, where the language is almost invisible and the only thing that matters is getting to the evidence with minimal ceremony.

Your Part 1 shuffle can be made to skew, and you would then be reading per-reducer counts you printed yourself, from a system you wrote, in a format you chose. That teaches the mechanism and nothing about the diagnosis. Here you get the version where the system is not yours and the evidence is where it will be at work.

- [ ] `skew_job.py`: generate two datasets of the same size, one with uniformly random keys and one Zipf-distributed (a handful of keys taking most of the rows, which is what real data looks like: a few huge customers, one null-ish default value, one bot account). Run the same `groupBy` and aggregation over both.
- [ ] Run against the uniform dataset first, so you know what healthy looks like. Open the Spark UI at `localhost:4040`, find the aggregation stage, and look at the **task duration distribution**: min, 25th percentile, median, 75th, max. On uniform data these are close together.
- [ ] Now run against the Zipf dataset and look at the same summary. The max is a large multiple of the median, and stage wall time is essentially that one task's duration, because every other task finished and the stage cannot complete without the straggler. **Write down the max-to-median ratio.** That ratio is the diagnosis, and recognising it on sight is the actual skill this part exists to build.
- [ ] Find the same story a second way, in the **Shuffle Read Size / Records** column per task. One task read far more than its share. Being able to distinguish "one task got more data" from "one task was on a slow machine" is what separates a skew diagnosis from a guess, and these two views are how you tell them apart.

**Minimum bar (Part 2):** the max-to-median task duration ratio on the skewed run, read out of the Spark UI, plus the same ratio after one fix. Knowing which UI panel told you is part of the bar.

**Break it, then decide:**
- [ ] Fix it with salting: for the hot keys, append a random suffix (`"userA"` becomes `"userA#3"`) so their rows spread across several partitions, aggregate, then run a second, much smaller pass to combine the salted groups. Re-run and confirm the task duration spread flattened in the UI. Note what salting cost you: a second shuffle, and an aggregation that only works because summing is associative. It would not work this way for a median.
- [ ] Turn on AQE's skew handling (`spark.sql.adaptive.skewJoin.enabled`) and re-run the unsalted version. Spark splits the oversized partition for you. Compare against your salted result and say what AQE could do automatically and what it could not, and be specific about why a `groupBy` skew and a join skew are not the same problem.
- [ ] **Your call:** given a stated memory budget of 200 MB per worker and a dimension table you are told is "about 150 MB, growing 10 percent a quarter," decide which you would ship: salting, which always works and always costs an extra pass, or a broadcast, which is dramatically faster right up until the table stops fitting and the job starts failing with out-of-memory errors in production. Implement whichever you pick, and write down the specific signal you would want monitored so you find out before your users do. W07 is the unit where that threshold gets crossed on purpose.

## Reflect


**Prediction versus measurement.** Fill the predictions in *before* you run anything, and do not edit them afterwards. The gap is where calibration comes from.

| Quantity | Predicted | Measured | Which term I got wrong |
|----------|-----------|----------|------------------------|
| | | | |

Copy anything worth carrying into [MEASUREMENTS.md](../MEASUREMENTS.md).

**What clicked:**

**What surprised me:**

**Max-to-median task duration ratio on the Zipf run, and which Spark UI panel you read it from:**

**The same ratio after your fix:**

**Why does a shuffle write to disk instead of streaming straight to the reducers?**

**What happened when you deleted a spill file, and what a real system does instead of trusting a directory listing:**

**Salting or broadcast: which did you implement, and what signal would you monitor to catch the failure mode of the one you chose?**

**What AQE's skew handling fixed automatically, what it did not, and why a `groupBy` skew is not a join skew:**

**Your `HashPartitioner` uses plain modulo, which is correct here because the partition count is fixed for the whole job. DDIA Ch.7 discusses consistent hashing as the alternative and is deliberately cool on it for databases. From the chapter, why does a fixed partition count usually beat a hash ring for a system that shards data rather than caches it?**

**Where would this same partition-then-exchange structure show up in a system you've actually used?**

**What I'd do differently:**

---

## Review and articulate

Two steps that exist because self-study has no examiner. Do them at the end of every unit, before marking it done.

- [ ] **Adversarial review.** Hand over three things separately: the number you predicted, the number you measured, and the conclusion you drew. Then ask for the strongest case that the conclusion is *not* supported by the measurement. Do not ask whether you are right; ask what would falsify this. An assistant asked to check your work will tend to find support for your framing, so the prompt has to be adversarial by construction or the exercise is theatre.
- [ ] **Ninety seconds, out loud, timed.** Explain this unit's finding as you would to someone in an interview or a design review: what you measured, what surprised you, and what decision it would change. Articulation under time pressure is a separate skill from understanding, and it is the one that gets tested. If you cannot do it in ninety seconds you do not have the finding yet, you have notes.
