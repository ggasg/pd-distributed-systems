---
week_number: 9
status: not-started
---

# W09: Distributed Training

> **Arc:** Distributed ML Systems · **Language:** Python
> **Budget:** about 10 hours. The Minimum bar is what a bad week looks like, not the target.

## What you'll build

Data-parallel training using Python multiprocessing and raw sockets, built around the distributed mechanics rather than the model. Two workers each train on half of MNIST with the same small MLP, exchange gradients via ring-allreduce, and converge to the same result. No PyTorch distributed, no Horovod.

**Scenario:** a training run that's been going for six hours stalls with no error, no crash, no log line, it's just stuck. One worker died mid-step and the other is blocked on a socket read that will never return. This is a real, common way distributed training jobs waste GPU-hours, and it's built directly into this unit's implementation so you can watch it happen on a small, safe scale.

---

## Read

- [ ] [Horovod: fast and easy distributed deep learning in TensorFlow](https://arxiv.org/abs/1802.05799) (Sergeev & Del Balso, 2018): focus on Section 3 (ring-allreduce algorithm). Understand exactly what bytes are being sent and why ring topology uses bandwidth efficiently.
- [ ] Optional: PyTorch DDP source, read [`torch/distributed/distributed_c10d.py`](https://github.com/pytorch/pytorch/blob/main/torch/distributed/distributed_c10d.py), specifically the `all_reduce` function and its docstring. You don't need to understand the CUDA path, just the concept.
- [ ] [NCCL: Collective Operations](https://docs.nvidia.com/deeplearning/nccl/user-guide/docs/usage/collectives.html): a short reference page, read for vocabulary rather than API detail. The operations you'll care about are `AllReduce`, `ReduceScatter`, and `AllGather`. The one fact worth carrying out of it: an allreduce is not a primitive. It is a reduce-scatter followed by an all-gather, which is exactly the two phases you're about to implement, and knowing they have standard names makes every distributed-training doc you read afterwards legible.

**Depth: study Section 3 of Horovod.** You implement ring-allreduce from it and then verify your byte counts against its math, which is the tightest paper-to-code loop in the curriculum. The NCCL collectives page is a short read for vocabulary. The DDP source is a skim.

**The vocabulary Horovod assumes**, worth having before you read it:

- **Data parallelism**: every worker holds a complete copy of the model and they split the training data between them, so each computes gradients over a different slice and they combine the results. W10 is where the model stops fitting and you cut it up instead.
- **Collective operation**: a communication pattern involving every worker at once rather than a point-to-point send.
  - **Broadcast**: sends one worker's buffer to all of them.
  - **Reduce**: combines every worker's buffer with an operator like sum, leaving the result on one worker.
  - **All-reduce**: the same combination, leaving the result on every worker. This is what averaging gradients needs.
  - **All-gather**: gives every worker the concatenation of all buffers, without combining them.
  - **Reduce-scatter**: combines like reduce, leaving each worker holding only its own slice of the result.
- **Ring-allreduce** is not a separate operation. It is an implementation of all-reduce as a reduce-scatter followed by an all-gather around a ring.
- **Rank**: a worker's integer identity, zero through N-1. **World size**: N, the total worker count.

**Key question:** Why is ring-allreduce more bandwidth-efficient than a parameter server for large gradients? Work out the math for N workers and a gradient of size G. You should land on each worker sending roughly 2G(N-1)/N bytes, and you should be able to say why that quantity barely changes as N grows.

---

## Code

Project: `code/distributed-training/` (Python 3.12+)

Dependencies: `numpy`, `torch` (for data loading only, no `torch.distributed`), `socket`, `multiprocessing`.

Model: 2-layer MLP on MNIST (784 → 128 → 10). Implemented in NumPy only.

### Step 1: `mlp.py`

The backprop math is not this unit's exercise. Write this file once from any standard reference, or reuse an implementation you already have, and do not spend the unit on it.

- [ ] An `MLP` class in NumPy with exactly this interface, since `worker.py` calls all four:
  ```python
  forward(X)            # logits
  backward(X, Y)        # returns list of gradient arrays, same order as params()
  params()              # returns list of weight arrays
  apply_grads(grads)    # in-place update from a list in params() order
  ```
- [ ] ReLU, softmax, cross-entropy. No PyTorch.

### Step 2: `ring_allreduce.py`

- [ ] Ring-allreduce over a list of NumPy arrays. Each worker knows its rank and the world size.
- [ ] Reduce-scatter phase: each worker sends a chunk to the next, receives and adds. After N-1 steps every worker owns the fully summed version of exactly one chunk.
- [ ] All-gather phase: each worker sends its finished chunk around the ring. After another N-1 steps every worker has all of them, so every worker ends with the sum of all workers' arrays.
- [ ] Real TCP sockets. Each worker binds a port; worker 0 initiates.
- [ ] Add a module-level byte counter that every `send` increments, so you report bytes actually on the wire rather than the number you expected.

### Step 3: `naive_allreduce.py`

- [ ] The obvious version, for contrast: every worker sends its full gradient to every other worker and sums the N arrays it receives locally. Correct, about ten lines, and what most people write first. Instrument it with the same byte counter.

### Step 4: `compare_allreduce.py`

- [ ] Run both against the same gradient at N = 2, 4, and 8 simulated workers, and print bytes sent per worker for each. Ring should stay flat as N grows while naive climbs linearly.
- [ ] Confirm the measured numbers match the 2G(N-1)/N formula from the Read section. A mismatch is usually chunk padding, when the gradient size does not divide evenly by N. Find it rather than rounding it away.

### Step 5: `worker.py` and `train.py`

- [ ] `worker.py`: load this worker's shard of MNIST, run forward and backward, call `ring_allreduce` on the gradients, update params. Five epochs.
- [ ] `train.py`: launch 2 workers via `multiprocessing.Process` with ranks 0 and 1, wait for both, and print final train accuracy per worker. The two should be close.

### Step 6: `tools/grad_server/main.go` (optional)

- [ ] Replace the raw-socket allreduce with a Go HTTP gradient aggregation server using `net/http`, standard library only. Python workers `POST /gradients` with `{"rank": 0, "gradients": [[...]]}`; once all workers have posted, the server averages them and workers `GET /gradients/averaged` to fetch the result.
- [ ] Collect submissions safely across the goroutines handling each request, with a `sync.WaitGroup` or a size check under a `sync.Mutex`. Keep it under 100 lines.

Go for the coordination service and Python for the ML compute is a common production split.

**Constraints:** no `torch.nn`, no `torch.optim`, no `torch.distributed`. Use `multiprocessing` rather than threads, because of the GIL. Sockets must be real TCP, not shared memory.

**Minimum bar:** two workers train to comparable accuracy with gradients synchronized by your own ring-allreduce, and you have bytes-on-the-wire per worker for naive versus ring at N = 2, 4, and 8. The Go gradient server is optional, not the bar.

---

## Break it, then decide

- [ ] Mid-training, `kill -9` one worker's process partway through a `ring_allreduce` call (right after it's sent its chunk but before it's received the reply). Watch the surviving worker: it's blocked on a socket `recv()` that will never be satisfied, so the whole job hangs indefinitely rather than crashing or erroring. Confirm this by timing out yourself (Ctrl-C) after a minute, since nothing in the current implementation will do it for you.
- [ ] What you just watched is W03's ambiguity with real money attached: the surviving worker cannot tell whether its peer died or is merely slow, and the only instrument available is a timeout you have to choose. With only 2 workers, there's no way to "route around" the dead one, a ring-allreduce with one member missing isn't a smaller ring, it's a broken one. Given that, is a socket read timeout (fail the whole step loudly and let the caller decide whether to restart both workers from the last checkpoint) the right fix here, or is that only a stopgap that stops mattering once you're past 2 workers, where a real system could exclude a dead node and re-derive a smaller ring instead of failing the whole job? Add a timeout to `ring_allreduce.py`'s socket calls either way, and write down at what worker count you think "fail the whole step" stops being good enough.

---

## Reflect

**Prediction versus measurement.** Fill the predictions in *before* you run anything, and do not edit them afterwards. The gap is where calibration comes from.

| Quantity | Predicted | Measured | Which term I got wrong |
|----------|-----------|----------|------------------------|
| | | | |

Copy anything worth carrying into [MEASUREMENTS.md](../MEASUREMENTS.md).

**What clicked:**

**What surprised me:**

**How many bytes does each worker send per epoch?**

**Measured bytes per worker, naive vs ring, at N = 2, 4, and 8 (and did they match 2G(N-1)/N):**

**What PyTorch DDP does that you didn't implement (and why it matters at scale):**

**At what worker count does "fail the whole step" stop being good enough, and what would a real fix need instead (from Break it, then decide above)?**

**What I'd do differently:**
