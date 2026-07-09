---
week_number: 1
status: not-started
---

# W01 — LSM-Trees and Storage Engines

> **Arc:** Data Systems Internals · **Language:** Java 21

## What you'll build
A minimal LSM-tree: MemTable (in-memory sorted buffer) → SSTable (sorted file on disk) → merge read path. No compaction. Passes put/get/scan tests.

---

## Read
- [ ] DDIA Ch.3, pp. 70–99 — focus on B-Trees vs LSM-Trees; understand SSTables, compaction strategies, and bloom filters. (DDIA is a book, not a free PDF — see [RESOURCES.md](../RESOURCES.md) if you don't have a copy yet; it's referenced again in W02, W04, and W05.)
- [ ] LevelDB source (30 min skim): [`db/memtable.h`](https://github.com/google/leveldb/blob/main/db/memtable.h), [`db/version_set.cc`](https://github.com/google/leveldb/blob/main/db/version_set.cc) — read to see how it's actually done, not to understand every line

**Key question to answer before coding:** Why does an LSM-tree have better write throughput than a B-tree, and what's the cost?

---

## Code

Project: `code/lsm/` (Java 21, Maven or Gradle)

- [ ] `MemTable.java` — sorted in-memory buffer using `TreeMap<byte[], byte[]>` with a custom byte-array comparator
- [ ] `SSTable.java` — writes sorted entries to a binary file (key-length, key bytes, value-length, value bytes); reads via sequential scan or binary search on an in-memory index
- [ ] `LSMTree.java` — writes go to MemTable; when MemTable exceeds threshold (e.g. 1MB), flush to a new SSTable file; reads check MemTable first, then SSTables from newest to oldest
- [ ] `LSMTreeTest.java` — test: (1) put 10k keys, get them back; (2) overwrite a key, confirm latest value wins; (3) force MemTable flush, confirm reads still work across SSTable boundary

**Constraints:** no external libraries beyond JUnit. Implement your own byte comparator.

---

## 🐍 Python DSA Review (optional)

**Binary search + sorted k-way merge** — the two algorithms inside every SSTable read and compaction.

```python
# binary_search.py — implement bisect_left from scratch
def bisect_left(arr, target):
    lo, hi = 0, len(arr)
    while lo < hi:
        mid = (lo + hi) // 2
        if arr[mid] < target: lo = mid + 1
        else: hi = mid
    return lo

# sorted_merge.py — merge two sorted lists of (key, value) pairs (SSTable compaction)
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

**Connection:** `bisect_left` is what an SSTable does on every point read. `merge_sstables` is what compaction does when collapsing multiple sorted runs — your Java `LSMTree` does this, but Python makes the algorithm visible in 10 lines.

---

## Reflect
<!-- Fill in at the end of the week -->

**What clicked:**

**What surprised me:**

**How this connects to a system you've worked with or currently build:**

**What I'd do differently:**
