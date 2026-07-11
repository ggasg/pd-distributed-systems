---
week_number: 1
status: not-started
---

# W01: LSM-Trees and Storage Engines

> **Arc:** Data Systems Internals · **Language:** Go

## What you'll build
A minimal LSM-tree: MemTable (in-memory sorted buffer) → SSTable (sorted file on disk) → merge read path. No compaction. Passes put/get/scan tests.

---

## Read
- [ ] DDIA Ch.3, pp. 70–99: focus on B-Trees vs LSM-Trees; understand SSTables, compaction strategies, and bloom filters. (DDIA is a book, not a free PDF. See [RESOURCES.md](../RESOURCES.md) if you don't have a copy yet; it's referenced again in W02, W04, and W05.)
- [ ] LevelDB source (30 min skim): [`db/memtable.h`](https://github.com/google/leveldb/blob/main/db/memtable.h), [`db/version_set.cc`](https://github.com/google/leveldb/blob/main/db/version_set.cc): read to see how it's actually done, not to understand every line
- [ ] Optional, if you want the Go-native version of the same idea: skim [Pebble's memtable](https://github.com/cockroachdb/pebble) or [Badger](https://github.com/dgraph-io/badger) source — both are production LSM engines written in Go, and both use skip lists where this week uses a plain sorted slice

**Key question to answer before coding:** Why does an LSM-tree have better write throughput than a B-tree, and what's the cost?

---

## Code

Project: `code/lsm/` (Go, module)

- [ ] `memtable.go`: `MemTable` struct holding entries sorted by key. Go has no built-in sorted map, so keep a `[]entry` slice sorted by key and use `sort.Search` to find insertion/lookup points (O(n) insert is fine here — MemTables are small and bounded before flush). A real engine would use a skip list for O(log n) insert; note in a comment that you're trading that away for simplicity. Methods: `Put(key, value []byte)`, `Get(key []byte) ([]byte, bool)`, `Scan() []entry`
- [ ] `sstable.go`: writes sorted entries to a binary file (key-length, key bytes, value-length, value bytes) using `encoding/binary`; reads via sequential scan or binary search on an in-memory sparse index built at open time
- [ ] `lsm_tree.go`: `LSMTree` struct — writes go to the `MemTable`; when it exceeds a threshold (e.g. 1MB), flush to a new `SSTable` file; reads check the `MemTable` first, then `SSTable`s from newest to oldest
- [ ] `lsm_tree_test.go`: table-driven tests, Go's `testing` package: (1) put 10k keys, get them back; (2) overwrite a key, confirm latest value wins; (3) force a `MemTable` flush, confirm reads still work across the SSTable boundary

**Constraints:** standard library only — no external modules. Implement your own byte-slice comparator (`bytes.Compare` is fine to use; don't reach for a third-party sorted-map library).

---

## 🐍 Python DSA Review (optional)

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

**Connection:** `bisect_left` is what an SSTable does on every point read; Go's `sort.Search` is the same binary search generalized to any predicate. `merge_sstables` is what compaction does when collapsing multiple sorted runs. Your Go `LSMTree` does this, but Python makes the algorithm visible in 10 lines.

---

## Reflect
<!-- Fill in at the end of the week -->

**What clicked:**

**What surprised me:**

**How this connects to a system you've worked with or currently build:**

**What I'd do differently:**
