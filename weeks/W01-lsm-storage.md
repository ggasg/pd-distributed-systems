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
- [ ] DDIA Ch.3, pp. 70–99 — focus on B-Trees vs LSM-Trees; understand SSTables, compaction strategies, and bloom filters
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

## Reflect
<!-- Fill in at the end of the week -->

**What clicked:**

**What surprised me:**

**How this connects to my current role:**

**What I'd do differently:**
