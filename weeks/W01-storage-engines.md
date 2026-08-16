---
week_number: 1
status: not-started
---

# W01: Storage Engines and the Cost of a Write

> **Arc:** Storage, Batch, and Failure · **Language:** Go
> **Budget:** about 10 hours, a good half of it reading. The Minimum bar is what a bad week looks like, not the target.

## What you'll build

Not a storage engine. An in-place write path, an append-only write path, and a stopwatch.

DDIA Ch.4's central claim is that appending to a log is dramatically faster than updating records in place, and that this single fact is why LSM-trees exist and why write-heavy systems are built the way they are. This unit has you measure that claim on your own machine rather than take it on faith, then watch the bill come due on the read side.

**Scenario:** someone proposes switching a write-heavy service from an update-in-place store to a log-structured one, and asks you what it buys and what it costs. You should be able to answer with a ratio you have personally measured and a specific, named cost, not with the phrase "better write throughput."

---

## Read

- [ ] **DDIA Ch.4** (2nd ed.), Storage and Retrieval. Focus on the LSM-tree versus B-tree comparison and the section on SSTables. You do not need the full index taxonomy, and you can skip the vector-embeddings section unless it interests you; W12's attention work is the retrieval-side counterpart if it does.
- [ ] Optional: [LevelDB source](https://github.com/google/leveldb), 20 minutes, `db/memtable.h` and nothing else. Read it to see what a real MemTable looks like, not to understand it. C++, and the syntax is irrelevant here.

**Depth: read DDIA Ch.4, do not study it.** A careful read is the right level here; the measurement below is what makes it stick, not a third pass through the prose. Everything else here is a skim.

**The vocabulary this unit measures in**, which the chapter uses without always naming:

- **Sequential I/O** writes or reads bytes at consecutive offsets. **Random I/O** jumps between scattered offsets. The gap between them is why a log wins, and it survives the move from spinning disks to SSDs for different reasons at each layer, so state which layer you mean when you explain it.
- **Write amplification**: bytes actually written to the device over bytes your application asked to write. This is where an update in place quietly loses, because rewriting a whole page to change one record is amplification.
- **Read amplification**: the same ratio on the read side. It is the bill the log-structured design defers rather than avoids, since a key may now live in any of several files.
- **Page cache**: the kernel's in-memory copy of file data. A write that returns quickly has often only reached memory.
- **fsync**: the call that forces those pages to durable storage, and therefore where the real cost of durability shows up. A benchmark that never calls fsync is measuring the page cache.

**Key question to answer before coding:** Why is appending to a file faster than updating a record in place, given that both end up writing the same number of bytes? Answer in terms of what the disk is physically doing, using the terms above, then keep your answer and check it against your measurement.

---

## Code

Project: `code/storage-bench/` (Go modules, standard library only)

Roughly sixty lines across the three files below. File I/O, a loop, and `time.Since`. No data structures to get right.

### Step 1: `inplace.go`

- [ ] A fixed-slot file. Allocate a file of `N` fixed-width records up front (`key` padded to 16 bytes, `value` padded to 48, so record `i` lives at offset `i*64`). To write key `i`, `Seek` to its offset and `Write` 64 bytes. Every write is a seek followed by a small write, which is the shape of an in-place update.

### Step 2: `appendlog.go`

- [ ] An append-only file. To write key `i`, append a record to the end. Never seek. Later records for the same key sit after earlier ones, and the newest one wins by convention rather than by overwriting anything.

### Step 3: `bench.go`

- [ ] Write 100,000 records through each path, in random key order, and print wall-clock time and throughput in records per second for both.
- [ ] Use random key order, not sequential. Sequential keys would let the in-place path write sequentially too, which hides the entire effect being measured.
- [ ] Call `f.Sync()` once after the 100,000 writes, inside the timed section, on both paths. Without it you are timing the page cache rather than the disk, and the two paths look far more alike than they are. Syncing once per run rather than per write is the choice being made here; note it next to your numbers, because it is the difference between measuring throughput and measuring durability cost.
- [ ] Run each path three times and take the median. A single run on a laptop is noise.

**Minimum bar:** two throughput numbers, measured by you, on the same 100,000 records in random order, with the ratio between them written down. That is the unit. If the ratio surprises you, say so in Reflect.

---

## Break it, then decide

- [ ] Now read. Pick 1,000 keys at random and fetch each one from both files, timing the lookups. The in-place file is trivial: seek to `i*64`, read 64 bytes, done, one disk access per key. The append log has no index, so a correct read has to scan for the *last* record matching that key, which means reading the whole file. Measure both. You have just paid for the write throughput you gained, and the size of the bill is the number in front of you. This is the LSM bargain in its entirety, and every real engine is an attempt to buy the read performance back without giving up the write side.
- [ ] Update the same 1,000 keys 100 times each through the append log, then look at the file size on disk. It holds 100,000 records to represent 1,000 live keys, and nothing in your code will ever reclaim the other 99 percent. That is why compaction exists. You are not going to build it, and you should be able to say precisely what it would do and what it would cost.
- [ ] **Your call:** you have measured a write win and a read loss. Real engines buy the read side back two ways: an in-memory index over the log, which costs memory proportional to the key count, or a Bloom filter per file, which costs very little memory and tells you only that a key is *definitely not* present. Given the numbers you just measured, say which you would add first and what workload would change your mind. There is a defensible answer either way, and the point is that you are now arguing from your own measurements rather than from the chapter.

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

**Write throughput, in-place versus append-only, and the ratio:**

**Read latency for 1,000 random keys, both paths. What did the write win actually cost?**

**Your answer to the Key question before you measured, and whether the measurement agreed with it:**

**Index or Bloom filter first, and what workload would change your mind:**

**What I'd do differently:**
