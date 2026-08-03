---
week_number: 5
status: not-started
---

# W05: Partitioning and the Shuffle

> **Arc:** Data Movement and Execution · **Language:** Java
> **Budget:** about 5 hours. Hit the Minimum bar first; everything past it is optional.

## What you'll build
A working shuffle in Java: map tasks that split their output into per-partition spill files, reduce tasks that fetch their own partition from every map task, and a partitioner that decides which key goes where. Then you feed it a skewed key distribution and watch one reducer do most of the work.

If "shuffle" is a word you've heard without a firm picture behind it, here it is in one sentence: a shuffle is the all-to-all data movement that happens when every machine holds some of the rows for a key and they all need to end up on one machine so the key can be aggregated. It is the single most expensive thing a distributed query engine does, and it is the mechanism underneath `GROUP BY`, `JOIN`, and `reduceByKey` in every system you're likely to touch.

**Scenario:** a nightly aggregation job that normally finishes in twenty minutes has been taking three hours since last Tuesday. Every executor finished quickly except one, which is still going. Nothing errored, no config changed, and the data volume grew by four percent. This is the most common performance incident in data engineering, and by the end of the unit you will have caused it deliberately and then fixed it.

---

## Read
- [ ] DDIA Ch.7 (2nd ed.), "Sharding": the whole chapter. Key-range vs hash sharding, the hot-spot problem, and rebalancing. Note that Kleppmann is deliberately cool on consistent hashing for databases; read his reasoning rather than assuming consistent hashing is the default right answer.
- [ ] [Spark RDD Programming Guide, "Shuffle operations"](https://spark.apache.org/docs/latest/rdd-programming-guide.html#shuffle-operations): short, concrete, and describes exactly the map-side-write then reduce-side-fetch structure you're about to build. Read the "Performance Impact" subsection twice.

**Depth: study DDIA Ch.7.** You build a partitioner and then reproduce the hot-spot problem the chapter describes, so the two reinforce each other directly. The Spark shuffle page is a short read. Dynamo is an optional skim.

**Key question:** Why does a shuffle write to disk at all, instead of streaming records straight from map tasks to reduce tasks over the network? What breaks if you don't?

---

## Code

Project: `code/shuffle/` (Java 21, Maven)

Data model: `record Record(String key, int value) {}`. The job is a word-count-shaped aggregation, sum of `value` per `key`, deliberately simple so the mechanics stay the subject.

- [ ] `Partitioner.java`: `sealed interface Partitioner permits HashPartitioner, RangePartitioner { int partitionFor(String key); }`. This is the same sealed-interface idiom you used for `StreamItem` in W04, reused rather than newly introduced. `HashPartitioner` is `Math.floorMod(key.hashCode(), numPartitions)`; use `floorMod`, not `%`, because `hashCode()` can be negative and `%` in Java preserves the sign, which would hand you a negative array index. `RangePartitioner` takes a sorted array of split-point keys and binary-searches (`Arrays.binarySearch`) to find the partition.
- [ ] `MapTask.java`: takes a `List<Record>` and a `Partitioner`, and writes one file per reduce partition into `spill/map-<mapId>/part-<partitionId>`. Each file is plain text, one `key,value` pair per line. Writing R files instead of one is the whole point: a reduce task should be able to read only what belongs to it.
- [ ] `ReduceTask.java`: for partition `p`, read `spill/map-*/part-<p>` from every map task, sum values per key, and write `output/part-<p>`. Return the number of records it processed, you will need that number later.
- [ ] `Shuffle.java`: wire M map tasks and R reduce tasks together and run them. Keep it single-JVM and run the tasks on plain threads (or sequentially, which is fine and easier to debug). The network is simulated by the filesystem here, deliberately: real spill-to-disk then fetch is what Spark actually does, so this is a simplification of scale, not of structure.
- [ ] `SkewBench.java`: generate two datasets of the same size, one with uniformly random keys and one Zipf-distributed (a handful of keys taking most of the rows, which is what real data looks like: a few huge customers, one null-ish default value, one bot account). Run the job over both and print records processed per reduce task, plus wall-clock time.

**Constraints:** no Spark, no Hadoop, no external dependencies beyond JUnit 5. Standard library only. Keep every class under 100 lines; if one is growing past that, the shuffle logic has probably leaked into it from somewhere it doesn't belong.

**Minimum bar:** the job runs end to end across M map tasks and R reduce tasks, and you have two numbers: the ratio between the busiest and median reducer on the skewed dataset, and the same ratio after applying one fix. The consistent-hashing review and the broadcast alternative are extras.

**Break it, then decide:**
- [ ] Run `SkewBench` and look at the per-reducer record counts for the Zipf dataset. One reducer should be handling a large multiple of what the others handle, and total wall time should be roughly that one reducer's time, because everyone else finished and waited. Write down the actual ratio you measured. This is the three-hour job from the scenario, reproduced on your laptop in a few seconds.
- [ ] Now fix it with salting: for the handful of hot keys, append a random suffix (`"userA" -> "userA#3"`) so their rows spread across several partitions, aggregate as normal, then run a second, much smaller aggregation pass to combine the salted groups back together. Measure again and confirm the skew flattened. Note what salting cost you: a second pass, and an aggregation that now only works because summing is associative. It would not work this way for, say, a median.
- [ ] **Your call:** salting is not the only fix. If one side of a join is small enough to fit in memory on every worker, you can broadcast it and skip the shuffle entirely. Given a stated memory budget (say 200 MB per worker) and a dimension table you're told is "about 150 MB, growing 10 percent a quarter," decide which you'd ship: salting, which always works but always costs an extra pass, or a broadcast, which is dramatically faster right up until the table stops fitting and the job starts failing with an out-of-memory error in production. Implement whichever you pick, and write down the specific signal you'd want monitored so you find out before your users do.

## Reflect

**What clicked:**

**What surprised me:**

**What ratio did you actually measure between the busiest and the median reduce task on the Zipf dataset?**

**Why does a shuffle write to disk instead of streaming straight to the reducers?**

**Salting or broadcast: which did you implement, and what signal would you monitor to catch the failure mode of the one you chose?**

**Your `HashPartitioner` uses plain modulo, which is correct here because the partition count is fixed for the whole job. DDIA Ch.7 discusses consistent hashing as the alternative and is deliberately cool on it for databases. From the chapter, why does a fixed partition count usually beat a hash ring for a system that shards data rather than caches it?**

**Where would this same partition-then-exchange structure show up in a system you've actually used?**

**What I'd do differently:**
