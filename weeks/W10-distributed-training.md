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

---

## Reflect

**What clicked:**

**What surprised me:**

**How many bytes does each worker send per epoch?**

**What PyTorch DDP does that you didn't implement (and why it matters at scale):**

**What I'd do differently:**
