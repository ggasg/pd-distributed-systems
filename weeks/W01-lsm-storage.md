---
week_number: 1
status: not-started
---

# W01: LSM-Trees and Storage Engines

> **Arc:** Data Systems Internals · **Language:** Java

## What you'll build
A minimal LSM-tree: MemTable (in-memory sorted buffer) → SSTable (sorted file on disk) → merge read path. No compaction. Passes put/get/scan tests.

---

## Read
- [ ] DDIA Ch.3, pp. 70–99: focus on B-Trees vs LSM-Trees; understand SSTables, compaction strategies, and bloom filters. (DDIA is a book, not a free PDF. See [RESOURCES.md](../RESOURCES.md) if you don't have a copy yet; it's referenced again in W02, W04, and W05.)
- [ ] LevelDB source (30 min skim): [`db/memtable.h`](https://github.com/google/leveldb/blob/main/db/memtable.h), [`db/version_set.cc`](https://github.com/google/leveldb/blob/main/db/version_set.cc): read to see how it's actually done, not to understand every line
- [ ] Optional, if you want the JVM-native version of the same idea: skim [Apache Cassandra's storage engine source](https://github.com/apache/cassandra/tree/trunk/src/java/org/apache/cassandra/db): a real, production LSM engine written in Java, and it literally calls its on-disk files SSTables, the same name this week uses

**Key question to answer before coding:** Why does an LSM-tree have better write throughput than a B-tree, and what's the cost?

---

## Code

Project: `code/lsm/` (Java 21, Maven)

- [ ] `MemTable.java`: holds entries sorted by key. Java has a real sorted map (`TreeMap`, a red-black tree) that would make this trivial; deliberately don't reach for it. Keep a plain `List<Entry>` sorted by key and use `Collections.binarySearch` (or a hand-rolled binary search) to find insertion and lookup points (O(n) insert is fine here; MemTables are small and bounded before flush). A real engine would use a skip list or a genuine balanced tree for O(log n) insert; note in a comment that you're trading that away for simplicity. `record Entry(byte[] key, byte[] value) {}` is a natural fit for the entry type, but be aware records auto-generate `equals`/`hashCode` field-by-field, and for `byte[]` fields that compares by array *reference*, not contents; override `equals`/`hashCode` yourself if you need content equality, or use `List<Byte>` instead if you'd rather not think about it. Methods: `put(byte[] key, byte[] value)`, `Optional<byte[]> get(byte[] key)`, `List<Entry> scan()`
- [ ] `SSTable.java`: writes sorted entries to a binary file (key-length, key bytes, value-length, value bytes) using `DataOutputStream`; reads via sequential scan or binary search on an in-memory sparse index built at open time, using `DataInputStream` on the read side
- [ ] `LSMTree.java`: writes go to the `MemTable`; when it exceeds a threshold (e.g. 1MB), flush to a new `SSTable` file; reads check the `MemTable` first, then `SSTable`s from newest to oldest
- [ ] `LSMTreeTest.java`: JUnit 5, parameterized where it helps: (1) put 10k keys, get them back; (2) overwrite a key, confirm latest value wins; (3) force a `MemTable` flush, confirm reads still work across the SSTable boundary

**Constraints:** JDK standard library only, no external dependencies. Implement your own byte-array comparator (`Arrays.compare(byte[], byte[])` is fine to use; don't reach for a third-party sorted-map library, and don't reach for `TreeMap` either, per above).

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

**Connection:** `bisect_left` is what an SSTable does on every point read; `Collections.binarySearch` is the same binary search generalized to any `Comparable`. `merge_sstables` is what compaction does when collapsing multiple sorted runs. Your Java `LSMTree` does this, but Python makes the algorithm visible in 10 lines.

---

## Reflect
<!-- Fill in at the end of the week -->

**What clicked:**

**What surprised me:**

**How this connects to a system you've worked with or currently build:**

**What I'd do differently:**
