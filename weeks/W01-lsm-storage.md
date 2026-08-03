---
week_number: 1
status: not-started
---

# W01: LSM-Trees and Storage Engines

> **Arc:** Data Systems Internals · **Language:** Go
> **Budget:** about 5 hours. Hit the Minimum bar first; everything past it is optional.

## What you'll build
A minimal LSM-tree: MemTable (in-memory sorted buffer) → SSTable (sorted file on disk) → merge read path. No compaction. Passes put/get/scan tests.

**Scenario:** it passes every put/get/scan test below, which is exactly the trap. A storage engine can be "correct" on the happy path and still lose data or degrade badly because of what it doesn't do yet. Both gaps in this week's build are made deliberately visible at the end instead of quietly ignored.

---

## Read
- [ ] DDIA Ch.4 (2nd ed.): focus on B-Trees vs LSM-Trees; understand SSTables, compaction strategies, and bloom filters. Optional, AI-adjacent extension in this edition: the chapter's new Vector Embeddings section, if you want the retrieval-side counterpart to W13's attention work. (DDIA is a book, not a free PDF. See [RESOURCES.md](../RESOURCES.md) if you don't have a copy yet; it's referenced again in W02, W03, and W04.)
- [ ] LevelDB source (30 min skim): [`db/memtable.h`](https://github.com/google/leveldb/blob/main/db/memtable.h), [`db/version_set.cc`](https://github.com/google/leveldb/blob/main/db/version_set.cc): read to see how it's actually done, not to understand every line. LevelDB is C++, this is the algorithm, not the syntax, worth reading regardless of your build language.
- [ ] Optional: [BadgerDB source](https://github.com/dgraph-io/badger): the JVM-native Cassandra pointer from earlier versions of this week doesn't apply anymore, this is its Go replacement, and arguably a closer match: a real, actively maintained, pure-Go LSM-tree key-value store. Skim `memtable.go` and `levels.go` (or their current equivalents, the exact filenames drift as the project evolves): a production Go engine solving the exact problem this week builds a toy version of, in the same language you're about to write it in.

**Depth: study DDIA Ch.4.** You are implementing the mechanism it describes, so this is the one to sit with. LevelDB and BadgerDB are skims: open them to see how a real engine shapes the same idea, close them after twenty minutes. The 1996 LSM paper is optional history.

**Key question to answer before coding:** Why does an LSM-tree have better write throughput than a B-tree, and what's the cost?

---

## Code

Project: `code/lsm/` (Go modules)

- [ ] `memtable.go`: holds entries sorted by key. Go's standard library has no built-in sorted map; deliberately don't reach for a third-party one. Keep a plain `[]Entry` sorted by key and use `sort.Search` (Go's binary search over any sorted slice, you supply the comparison) to find insertion and lookup points (O(n) insert is fine here; MemTables are small and bounded before flush). A real engine would use a skip list or a genuine balanced tree for O(log n) insert; note in a comment that you're trading that away for simplicity. `type Entry struct { Key, Value []byte }` is the natural fit; byte slices compare by content with `bytes.Equal`, not by reference, so you don't need the equals/hashCode caution a records-based version in another language would. Methods: `Put(key, value []byte)`, `Get(key []byte) ([]byte, bool)`, `Scan() []Entry`
- [ ] `sstable.go`: writes sorted entries to a binary file (key-length, key bytes, value-length, value bytes) using `encoding/binary` (`binary.Write` for fixed-width lengths); reads via sequential scan or binary search on an in-memory sparse index built at open time, using `encoding/binary`'s `binary.Read` on the read side
- [ ] `lsm_tree.go`: writes go to the `MemTable`; when it exceeds a threshold (e.g. 1MB), flush to a new `SSTable` file; reads check the `MemTable` first, then `SSTable`s from newest to oldest
- [ ] `lsm_tree_test.go`: Go's standard `testing` package, table-driven where it helps: (1) put 10k keys, get them back; (2) overwrite a key, confirm latest value wins; (3) force a `MemTable` flush, confirm reads still work across the SSTable boundary

**Constraints:** standard library only, no external dependencies beyond what's already in `go.mod`. Implement your own byte-slice comparator (`bytes.Compare` is fine to use; don't reach for a third-party sorted-map package).

**Minimum bar:** `put`, `get`, and `scan` pass through the full path, MemTable to SSTable to merged read. That is the whole unit. No compaction, no WAL, no bloom filter, no levelled merge; those are the gaps the exercise below has you name rather than close, and naming them accurately is worth more here than building them.

**Break it, then decide:**
- [ ] Durability check: `Put` 100 keys, confirm they haven't hit your flush threshold yet (still sitting only in the `MemTable`), then kill the process before calling any explicit flush and restart a fresh `LSMTree` pointed at the same directory. Confirm those 100 keys are gone. Nothing in this week's build writes an intentions log before an entry lands in the `MemTable`, so anything not yet flushed to an `SSTable` doesn't survive a crash. That's not a bug in your code, it's a real simplification; production engines cut this corner differently, by appending every write to a WAL before touching the `MemTable`, so a crash replays the log instead of losing data.
- [ ] Lower your flush threshold (say, 4KB instead of 1MB) so you accumulate several `SSTable`s quickly, then time a `Get()` for a key that doesn't exist anywhere. It has to check the `MemTable`, then every `SSTable` oldest to newest, since nothing tells it early that a given table can't contain the key. Given that cost, and given this build has neither a WAL nor a bloom filter, if you could ship only one improvement before this went to production, durability (the WAL from the bullet above) or read performance on misses (a bloom filter), which would you pick? There's a defensible answer either way depending on the workload; write down which one and why in Reflect.

---

## Rehearse it in Python first (optional, 20 minutes)

> **Why this exists, and when it stops.** This unit builds in Go, which is the one language here you are still learning. Writing the SSTable read and the compaction merge in Python first means that when the Go version misbehaves you already know whether the problem is the algorithm or the syntax, which is the single most useful thing to know at that moment. These sections appear only in the Go units (W01, W02, W03, W07) and stop after W07, by which point Go should no longer be the thing in your way. Skip it whenever the algorithm is already obvious to you.

**Binary search + sorted k-way merge**: the two algorithms inside every SSTable read and compaction.

```python
# binary_search.py: implement bisect_left from scratch
def bisect_left(arr, target):
    lo, hi = 0, len(arr)
    while lo < hi:
        mid = (lo + hi) // 2
        if arr[mid] < target: lo = mid + 1
        else: hi = mid
    return lo

# sorted_merge.py: merge two sorted lists of (key, value) pairs (SSTable compaction)
def merge_sstables(a, b):
    result, i, j = [], 0, 0
    while i < len(a) and j < len(b):
        if a[i][0] <= b[j][0]: result.append(a[i]); i += 1
        else: result.append(b[j]); j += 1
    return result + a[i:] + b[j:]

# Test: merge two sorted sstables, later key wins on tie
a = [("apple", 1), ("mango", 3)]
b = [("apple", 2), ("grape", 4)]
assert merge_sstables(a, b) == [("apple", 1), ("apple", 2), ("grape", 4), ("mango", 3)]
```

**Connection:** `bisect_left` is what an SSTable does on every point read; `sort.Search` is the same binary search generalized to any monotonic predicate. `merge_sstables` is what compaction does when collapsing multiple sorted runs. Your Go `LSMTree` does this, but Python makes the algorithm visible in 10 lines.

---

## Reflect
<!-- Fill in at the end of the week -->

**What clicked:**

**What surprised me:**

**How this connects to a system you've worked with or currently build:**

**WAL or bloom filter first, and why (from Break it, then decide above):**

**What I'd do differently:**
