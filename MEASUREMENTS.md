# Measurements

Numbers you have personally measured, on your own hardware, with the conditions that produced them.

**Why this file exists.** Answering a performance question on the spot is not a memory trick. It works because you carry a handful of reference points you actually measured, and you interpolate from them. Somebody asks whether adding executors will fix a slow job, and you can answer in thirty seconds because you once measured what a shuffle costs and what re-reading an input costs, and you know which one this smells like.

Without this file those numbers exist, but scattered across sixteen Reflect sections you will never search. With it they are one page you can reread before an interview or a design review.

**Rules, and they matter more than the format:**

1. **Only numbers you measured yourself.** A figure from a paper or a blog post is not a measurement, it is a citation. Keep those in the units; keep this page first-hand.
2. **Record the conditions.** A throughput number without hardware, data size, and configuration is not reusable, because you cannot tell whether a new situation is like the old one.
3. **Record what you predicted first.** The gap between prediction and measurement is the single most valuable column here. It is where calibration comes from, and it is the reason to write the prediction down before running anything rather than after.
4. **Keep the surprises.** A number that matched expectations teaches you little. A number that was off by 10x teaches you which term in your model was wrong, and that is the thing you will still remember in a year.

---

## Log

| Unit | Quantity | Predicted | Measured | Conditions | What the gap taught me | Date |
|------|----------|-----------|----------|------------|------------------------|------|
| W01 | Write throughput, in-place vs append-only | | | 100k records, random key order, median of 3 | | |
| W01 | Read latency, 1000 random keys, both paths | | | | | |
| W02 | Disk I/O per PageRank iteration | | | 100k nodes, avg out-degree 5 | | |
| W02 | Wall clock, 10 iterations, cached vs uncached | | | | | |
| W03 | Shortest hiccup that produces a false suspicion | | | 3 nodes, heartbeat 1s | | |
| W03 | Longest a dead node stays undetected | | | same detector, same timeout | | |
| W04 | `numLateRecordsDropped`, event 30s past the bound | | | 5s out-of-orderness, 10s tumbling window | | |
| W04 | Error in final aggregate under Drop | | | arrival 10% above service rate | | |
| W05 | Max-to-median task duration, Zipf keys | | | local Spark, groupBy | | |
| W05 | Same ratio after salting | | | | | |
| W06 | Rows/sec, hand loop vs Stream vs DuckDB at `threads=1` | | | 10M rows, Parquet | | |
| W06 | Cost of the join spilling to disk | | | build side grown past memory | | |
| W07 | Bytes moved, broadcast vs shuffle, 20 workers | | | 10 GB fact, 100 MB dimension | | |
| W09 | Bytes on the wire per worker, naive vs ring | | | N = 2, 4, 8 | | |
| W10 | Pipeline bubble fraction, measured vs theoretical | | | 3 values of M | | |
| W12 | Prefill tokens recomputed, round-robin vs cache-aware | | | 2 replicas | | |
| W15 | p99 from a correctly bucketed histogram vs a wrong one | | | | | |

Add rows as you go. The table above is a starting scaffold, not a limit; anything you measured belongs here.

---

## Reference points worth carrying

Fill these in as the units produce them. These are the ones that come up most often in real conversations, so they are the ones worth being able to state without looking.

- **A shuffle costs:** ___
- **Re-reading an input per iteration costs:** ___
- **Vectorized execution buys:** ___ (and parallelism separately buys ___)
- **A ring-allreduce moves:** ___ bytes per worker, versus ___ naive
- **A broadcast join beats a shuffle join until the small side exceeds:** ___
- **The cheapest useful failure detector timeout is bounded below by:** ___
