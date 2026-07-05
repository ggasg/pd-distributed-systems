# W03 — Raft Consensus

> **Arc:** Data Systems Internals · **Language:** Java 21

## What you'll build
Raft leader election + log replication in Java 21. 5 nodes simulated in-process using virtual threads. No persistence, no snapshots. Tests confirm a leader is elected and a log entry committed after a majority acknowledges it.

---

## Read
- [ ] [In Search of an Understandable Consensus Algorithm](https://raft.github.io/raft.pdf) (Ongaro & Ousterhout, 2014) — read Sections 1–5 carefully; skip the formal proof in Section 8 on first read. Figure 2 is the spec — keep it open while coding.
- [ ] Skim the [Raft visualization](https://raft.github.io/) — run through leader election and log replication scenarios to build intuition before you write a line of code

**Key question:** What happens if a network partition puts the old leader in the minority? Walk through the exact sequence of events.

---

## Code

Project: `code/raft/` (Java 21, virtual threads)

- [ ] `RaftNode.java` — state machine with roles (Follower, Candidate, Leader), currentTerm, votedFor, log (list of entries), commitIndex, lastApplied
- [ ] `RpcChannel.java` — in-process message passing using `LinkedBlockingQueue`; each node gets a channel; simulate network delay with random sleep (0–50ms)
- [ ] `LeaderElection.java` — implement `startElection()`: increment term, vote for self, send `RequestVote` RPCs to all peers via virtual threads, tally responses, transition to Leader if majority
- [ ] `LogReplication.java` — implement `appendEntries()`: leader sends log entries to followers; follower appends if term and prevLogIndex/prevLogTerm match; leader advances commitIndex when majority acknowledges
- [ ] `RaftTest.java` — test 1: start 5 nodes, wait for leader election, assert exactly one leader; test 2: client sends a command to leader, assert it's committed and applied on majority

**Constraints:** use `Thread.ofVirtual().start()` for each node's event loop. No external libraries. Message types as Java records.

---

## Reflect

**What clicked:**

**What surprised me:**

**The hardest part to implement correctly:**

**How this connects to Materialize:**

**What I'd do differently:**
