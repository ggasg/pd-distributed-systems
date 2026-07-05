---
week_number: 10
status: not-started
---

# W10 — Distributed Training

> **Arc:** Distributed ML & Compute · **Language:** Python

## What you'll build
Data-parallel training from scratch using Python multiprocessing and raw sockets — no PyTorch distributed, no Horovod. Two workers each train on half of MNIST, exchange gradients via allreduce (ring-allreduce), and converge to the same model.

---

## Read
- [ ] [Horovod: fast and easy distributed deep learning in TensorFlow](https://arxiv.org/abs/1802.05799) (Sergeev & Del Balso, 2018) — focus on Section 3 (ring-allreduce algorithm). Understand exactly what bytes are being sent and why ring topology uses bandwidth efficiently.
- [ ] PyTorch DDP source — read [`torch/distributed/distributed_c10d.py`](https://github.com/pytorch/pytorch/blob/main/torch/distributed/distributed_c10d.py), specifically the `all_reduce` function and its docstring. You don't need to understand the CUDA path — just the concept.

**Key question:** Why is ring-allreduce more bandwidth-efficient than a parameter server for large gradients? Work out the math for N workers and a gradient of size G.

---

## Code

Project: `code/distributed-training/` (Python 3.11+)

Dependencies: `numpy`, `torch` (for data loading only — no `torch.distributed`), `socket`, `multiprocessing`.

Model: 2-layer MLP on MNIST (784 → 128 → 10). Implemented in NumPy only.

- [ ] `mlp.py` — `MLP` class with `forward(X)`, `backward(X, Y)`, `params()` (returns list of weight arrays), `apply_grads(grads)`. Use ReLU + softmax + cross-entropy. No PyTorch.
- [ ] `ring_allreduce.py` — implement ring-allreduce for a list of NumPy arrays:
  - Each worker has a rank and knows the total number of workers
  - Scatter-reduce phase: each worker sends a chunk to the next, receives and adds
  - All-gather phase: each worker sends the reduced chunk, receives and places
  - Result: every worker has the sum of all workers' arrays
  - Use TCP sockets for communication (each worker binds a port; worker 0 initiates)
- [ ] `worker.py` — each worker: loads its shard of MNIST, runs forward + backward, calls `ring_allreduce` on gradients, updates params. Runs for 5 epochs.
- [ ] `train.py` — launches 2 workers via `multiprocessing.Process`, assigns rank 0 and rank 1, waits for both to complete. Prints final train accuracy per worker (should be similar).

**Constraints:** no `torch.nn`, no `torch.optim`, no `torch.distributed`. Use `multiprocessing` not threads (GIL). Sockets must be real TCP, not shared memory.

**Go gradient server (secondary tool):**

- [ ] `tools/grad_server/main.go` — replace the raw socket allreduce with a Go HTTP gradient aggregation server. Python workers POST their gradients as JSON arrays to `POST /gradients` (include `{"rank": 0, "gradients": [[...]]})`); once all workers have posted, the server averages them and returns the result. Python workers GET `/gradients/averaged` to fetch the result. Use `sync.WaitGroup` and a `Mutex`-protected map to collect worker submissions. Keep under 100 lines.

This is a realistic pattern: Go handles the coordination service, Python handles the ML compute.

---

## 🐍 Python DSA Review (optional)

**Ring buffer (circular array)** — ring-allreduce passes gradient chunks around a logical ring of workers. A ring buffer is the underlying data structure for any fixed-capacity FIFO queue without allocation.

```python
# ring_buffer.py
class RingBuffer:
    def __init__(self, capacity: int):
        self._buf = [None] * capacity
        self._cap = capacity
        self._head = 0   # index of oldest item
        self._size = 0

    def push(self, item) -> None:
        tail = (self._head + self._size) % self._cap
        self._buf[tail] = item
        if self._size < self._cap:
            self._size += 1
        else:
            self._head = (self._head + 1) % self._cap  # overwrite oldest

    def pop(self):
        if self._size == 0:
            raise IndexError("pop from empty RingBuffer")
        item = self._buf[self._head]
        self._head = (self._head + 1) % self._cap
        self._size -= 1
        return item

    def __len__(self): return self._size

# Test: capacity 3, push 4 items → oldest evicted
rb = RingBuffer(3)
for i in range(4): rb.push(i)
assert len(rb) == 3
assert rb.pop() == 1  # 0 was evicted

# Ring-allreduce mental model: N workers in a ring, each holds one chunk
# After N-1 scatter-reduce steps, every worker has partial sum of its chunk
# After N-1 all-gather steps, every worker has full gradient
def ring_indices(rank: int, n_workers: int, n_steps: int) -> list[int]:
    """Which chunk indices does worker `rank` send at each step?"""
    return [(rank - step) % n_workers for step in range(n_steps)]

assert ring_indices(0, 4, 4) == [0, 3, 2, 1]
```

**Connection:** `ring_allreduce.py` sends gradient chunks in exactly this circular pattern. The ring buffer is also the right data structure for the log-aggregator sidecar in W16 — fixed memory, O(1) push/pop.

---

## Reflect

**What clicked:**

**What surprised me:**

**How many bytes does each worker send per epoch?**

**What PyTorch DDP does that you didn't implement (and why it matters at scale):**

**What I'd do differently:**
