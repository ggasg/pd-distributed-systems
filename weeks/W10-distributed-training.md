---
week_number: 10
status: not-started
---

# W10: Distributed Training

> **Arc:** Distributed ML & Compute · **Language:** Python
> **Budget:** about 5 hours. Hit the Minimum bar first; everything past it is optional.

## What you'll build
Data-parallel training using Python multiprocessing and raw sockets, built entirely around the distributed mechanics rather than the model. You're given a small, already-implemented 2-layer MLP (forward, backward, and the ReLU/softmax/cross-entropy math all provided); deriving backpropagation by hand is real, valuable work that belongs to a dedicated ML/AI track, not this one. Two workers each train on half of MNIST using that provided model, exchange gradients via allreduce (ring-allreduce), and converge to the same result. No PyTorch distributed, no Horovod.

**Scenario:** a training run that's been going for six hours stalls with no error, no crash, no log line, it's just stuck. One worker died mid-step and the other is blocked on a socket read that will never return. This is a real, common way distributed training jobs waste GPU-hours, and it's built directly into this week's implementation so you can watch it happen on a small, safe scale.

---

## Read
- [ ] [Horovod: fast and easy distributed deep learning in TensorFlow](https://arxiv.org/abs/1802.05799) (Sergeev & Del Balso, 2018): focus on Section 3 (ring-allreduce algorithm). Understand exactly what bytes are being sent and why ring topology uses bandwidth efficiently.
- [ ] Optional: PyTorch DDP source, read [`torch/distributed/distributed_c10d.py`](https://github.com/pytorch/pytorch/blob/main/torch/distributed/distributed_c10d.py), specifically the `all_reduce` function and its docstring. You don't need to understand the CUDA path, just the concept.
- [ ] [NCCL: Collective Operations](https://docs.nvidia.com/deeplearning/nccl/user-guide/docs/usage/collectives.html): a short reference page, read for vocabulary rather than API detail. The operations you'll care about are `AllReduce`, `ReduceScatter`, and `AllGather`. The one fact worth carrying out of it: an allreduce is not a primitive. It is a reduce-scatter followed by an all-gather, which is exactly the two phases you're about to implement, and knowing they have standard names makes every distributed-training doc you read afterwards legible.

**Depth: study Section 3 of Horovod.** You implement ring-allreduce from it and then verify your byte counts against its math, which is the tightest paper-to-code loop in the curriculum. The NCCL collectives page is a short read for vocabulary. The DDP source is a skim.

**Key question:** Why is ring-allreduce more bandwidth-efficient than a parameter server for large gradients? Work out the math for N workers and a gradient of size G. You should land on each worker sending roughly 2G(N-1)/N bytes, and you should be able to say why that quantity barely changes as N grows.

---

## Code

Project: `code/distributed-training/` (Python 3.13+)

Dependencies: `numpy`, `torch` (for data loading only, no `torch.distributed`), `socket`, `multiprocessing`.

Model: 2-layer MLP on MNIST (784 → 128 → 10). Implemented in NumPy only.

**Given, not built:** `mlp.py` is provided as a starter file, an `MLP` class with `forward(X)`, `backward(X, Y)`, `params()` (returns list of weight arrays), and `apply_grads(grads)` already implemented (ReLU + softmax + cross-entropy, no PyTorch). Read it once so you know what `worker.py` is calling; you won't need to modify it. Deriving backpropagation by hand is a legitimate exercise on its own, it's just not this week's exercise: the thing actually being tested here is whether your ring-allreduce implementation correctly and efficiently synchronizes gradients across two independent processes, and handing you a working model keeps every hour of this week pointed at that question instead of splitting time with calculus.

- [ ] `ring_allreduce.py`: implement ring-allreduce for a list of NumPy arrays:
  - Each worker has a rank and knows the total number of workers
  - Reduce-scatter phase: each worker sends a chunk to the next, receives and adds. After N-1 steps every worker owns the fully summed version of exactly one chunk.
  - All-gather phase: each worker sends its finished chunk around the ring. After another N-1 steps every worker has all of them.
  - Result: every worker has the sum of all workers' arrays
  - Use TCP sockets for communication (each worker binds a port; worker 0 initiates)
  - Count bytes as you go: add a module-level counter that every `send` increments, so you can report actual bytes on the wire per worker rather than the number you expected
- [ ] `naive_allreduce.py`: the obvious implementation, for contrast. Every worker sends its full gradient to every other worker, and each one sums the N arrays it receives locally. This is correct, it is about ten lines, and it is what most people write first. Instrument it with the same byte counter.
- [ ] `compare_allreduce.py`: run both against the same gradient at N = 2, 4, and 8 simulated workers, and print bytes sent per worker for each. Ring should be flat as N grows while naive climbs linearly. Confirm your measured numbers match the 2G(N-1)/N formula from the Read section, and if they don't, the discrepancy is usually chunk padding when the gradient size doesn't divide evenly by N, which is worth finding rather than rounding away.
- [ ] `worker.py`: each worker: loads its shard of MNIST, runs forward + backward, calls `ring_allreduce` on gradients, updates params. Runs for 5 epochs.
- [ ] `train.py`: launches 2 workers via `multiprocessing.Process`, assigns rank 0 and rank 1, waits for both to complete. Prints final train accuracy per worker (should be similar).

**Constraints:** no `torch.nn`, no `torch.optim`, no `torch.distributed`. Use `multiprocessing` not threads (GIL). Sockets must be real TCP, not shared memory.

**Minimum bar:** two workers train to comparable accuracy with gradients synchronized by your own ring-allreduce, and you have bytes-on-the-wire per worker for naive versus ring at N = 2, 4, and 8. The Go gradient server is a secondary tool, not the bar.

**Break it, then decide:**
- [ ] Mid-training, `kill -9` one worker's process partway through a `ring_allreduce` call (right after it's sent its chunk but before it's received the reply). Watch the surviving worker: it's blocked on a socket `recv()` that will never be satisfied, so the whole job hangs indefinitely rather than crashing or erroring. Confirm this by timing out yourself (Ctrl-C) after a minute, since nothing in the current implementation will do it for you.
- [ ] What you just watched is W03's ambiguity with real money attached: the surviving worker cannot tell whether its peer died or is merely slow, and the only instrument available is a timeout you have to choose. With only 2 workers, there's no way to "route around" the dead one, a ring-allreduce with one member missing isn't a smaller ring, it's a broken one. Given that, is a socket read timeout (fail the whole step loudly and let the caller decide whether to restart both workers from the last checkpoint) the right fix here, or is that only a stopgap that stops mattering once you're past 2 workers, where a real system could exclude a dead node and re-derive a smaller ring instead of failing the whole job? Add a timeout to `ring_allreduce.py`'s socket calls either way, and write down at what worker count you think "fail the whole step" stops being good enough.

**Go gradient server (secondary tool):**

- [ ] `tools/grad_server/main.go`: replace the raw-socket allreduce with a Go HTTP gradient aggregation server, using `net/http` (standard library, no framework). Python workers POST their gradients as JSON arrays to `POST /gradients` (include `{"rank": 0, "gradients": [[...]]})`); once all workers have posted, the server averages them and returns the result. Python workers GET `/gradients/averaged` to fetch the result. Use a `sync.WaitGroup` (or poll a size check under a `sync.Mutex`) to collect worker submissions safely across the goroutines handling each request. Keep under 100 lines.

This is a realistic pattern: Go handles the coordination service, Python handles the ML compute.

## Reflect

**What clicked:**

**What surprised me:**

**How many bytes does each worker send per epoch?**

**Measured bytes per worker, naive vs ring, at N = 2, 4, and 8 (and did they match 2G(N-1)/N):**

**What PyTorch DDP does that you didn't implement (and why it matters at scale):**

**At what worker count does "fail the whole step" stop being good enough, and what would a real fix need instead (from Break it, then decide above)?**

**What I'd do differently:**
