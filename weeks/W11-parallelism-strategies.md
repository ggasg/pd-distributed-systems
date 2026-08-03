---
week_number: 11
status: not-started
---

# W11: Beyond Data Parallelism

> **Arc:** Distributed ML & Compute · **Language:** Python
> **Budget:** about 5 hours. Hit the Minimum bar first; everything past it is optional.

## What you'll build
Three ways to split a model, rather than the data, across two processes: tensor parallelism (cut a single matrix multiply in half), pipeline parallelism (put different layers on different workers), and sharded optimizer state (each worker keeps only its slice of the optimizer's bookkeeping). You'll reuse the ring-allreduce you wrote in W10 as the communication layer for all three.

Here's the framing, plainly. W10 was **data parallelism**: every worker holds a complete copy of the model and they split the data between them. That works right up until the model itself no longer fits in one worker's memory, at which point it stops being an option at all and you have to cut the model up instead. Everything this week is about how you cut it, and what each cut costs you in communication.

**Scenario:** you have a model that trains fine on one machine and a bigger one that does not fit at all. Somebody asks which parallelism strategy to use, and the honest answer depends on numbers you don't have yet: how much has to cross the network per step, and how much of the time each worker spends waiting. This week you measure both on something small enough to see clearly.

---

## Read
- [ ] [Megatron-LM](https://arxiv.org/abs/1909.08053) (Shoeybi et al., 2019): read Sections 1 and 3. Section 3 is the whole idea: split one weight matrix by columns, split the next by rows, and the two cuts compose so that a full MLP block needs exactly one all-reduce instead of two. That trick is the reason tensor parallelism is practical.
- [ ] [GPipe](https://arxiv.org/abs/1811.06965) (Huang et al., NeurIPS 2019): read Sections 1 to 3. The key concept is the "bubble," the fraction of time a worker spends idle waiting for the stage before it, and how splitting each batch into microbatches shrinks it.
- [ ] [ZeRO](https://arxiv.org/abs/1910.02054) (Rajbhandari et al., SC 2020): read Section 5 and the stage table. You do not need the full memory analysis. What you want is the three stages: shard the optimizer state, then the gradients, then the parameters themselves, each stage saving more memory and costing more communication.
- [ ] Optional: [PyTorch FSDP](https://arxiv.org/abs/2304.11277) (Zhao et al., VLDB 2023): what ZeRO looks like once it's a production API people actually call. Useful if you want to see how the theory above landed in the framework.

**Depth: study Section 3 of Megatron-LM.** The column-then-row composition is the one non-obvious idea this week, and you implement it. GPipe and ZeRO are reads, and you only need the named sections. FSDP is an optional skim.

**Key question:** Tensor parallelism communicates once per layer; pipeline parallelism communicates once per stage boundary. Given that, which one would you put inside a single machine across its GPUs, and which one across machines on a slower network? The answer follows directly from how often each one talks.

---

## Code

Project: `code/parallelism/` (Python 3.13+)

Dependencies: `numpy`, `multiprocessing`. You'll import `ring_allreduce` and `all_gather` directly from `code/distributed-training/` (W10), so this week builds on real code you already wrote rather than a fresh abstraction.

**Given, not built:** `layers.py` is provided, a `Linear` class with `forward` and `backward` and a GeLU activation, all NumPy. Same principle as W10: the calculus is not what's being tested here.

**Part 1: Tensor parallelism**

- [ ] `tensor_parallel.py`: implement `ColumnParallelLinear` and `RowParallelLinear` for two workers.
  - Column-parallel: split the weight matrix `A` by columns, so worker `i` holds `A_i`. Each worker computes `X @ A_i` on the full input and the outputs are concatenated. No communication needed in the forward pass, the halves are just two pieces of one wider output.
  - Row-parallel: split the weight matrix `B` by rows and the input by columns, so worker `i` computes `X_i @ B_i`. Each worker now holds a *partial sum* of the correct output, and an all-reduce is what turns those partials into the real answer.
- [ ] `mlp_block.py`: chain them the way Megatron does, column-parallel then GeLU then row-parallel, and confirm the whole block needs exactly one all-reduce. Assert the result matches a single-process reference implementation to within floating-point tolerance.

**Part 2: Pipeline parallelism**

- [ ] `pipeline_parallel.py`: split a 4-layer MLP into two stages, layers 1 and 2 on worker 0, layers 3 and 4 on worker 1. Forward activations cross the stage boundary over a socket; gradients cross back the other way. Start with one batch at a time.
- [ ] `microbatch.py`: split each batch into `M` microbatches and feed them into the pipeline back to back, so worker 1 can start on microbatch 1 while worker 0 is already working on microbatch 2.
- [ ] `bubble.py`: instrument both workers with timers that record idle time, then sweep `M` from 1 to 16 and print measured bubble fraction alongside the theoretical `(S-1)/(M+S-1)`. With two stages and `M=1` you should measure close to 50 percent idle, which is a genuinely shocking number the first time you see it, and it should fall off quickly as `M` rises.

**Part 3: Sharded optimizer state**

- [ ] `shard_optimizer.py`: implement ZeRO stage 1 for SGD with momentum. Each worker keeps momentum buffers for only its shard of the parameters, applies its own slice of the update, and then all-gathers the updated parameters so everyone ends the step with a complete model. Print peak memory per worker (`tracemalloc` is enough) against a non-sharded baseline, and confirm the loss curve is identical, because this is purely a memory optimization and must not change the math at all.

**Minimum bar:** Parts 1 and 2 only. The Megatron block matches a single-process reference and needs one all-reduce, and you have measured bubble fraction at three values of M against the theoretical curve. Part 3 (ZeRO) is the part to drop if the week runs out, since it is a memory optimization that provably does not change the math, which makes it the least costly thing to postpone.

**Break it, then decide:**
- [ ] In `RowParallelLinear`, comment out the all-reduce. The program does not crash, does not warn, and produces output of exactly the right shape. It is simply wrong, because every worker is holding a partial sum. Compare against the single-process reference and look at how wrong it actually is, then let training run a few steps and watch the loss go somewhere strange rather than error. Silent numerical corruption in a distributed layer is one of the genuinely hard bug classes in this field, and the lesson worth taking is that the shape check you'd normally trust tells you nothing here.
- [ ] Set `M=1` in the pipeline and confirm you measure roughly the theoretical bubble. Then reason about what would happen with 8 stages and `M=1`: the formula says 87.5 percent idle, and that is why nobody runs pipeline parallelism without microbatching.
- [ ] **Your call:** you're given a model where a single layer's weight matrix is too large to fit on one device. Tensor parallelism solves this directly but has to communicate inside every layer, so it needs fast interconnect. Pipeline parallelism communicates only at stage boundaries, which is far less traffic, but it cannot split an individual layer at all and it wastes time in the bubble. Pick one for this specific constraint, implement it as your `mlp_block.py`'s default path, and write down the interconnect assumption your choice depends on. Then say what you'd change if you were told the workers were in different data centers.

---

## Rehearse it in Python first (optional, 20 minutes)

**Balanced array partitioning (binary search on the answer)**: deciding which layers go on which pipeline stage is exactly the problem of splitting a list into `k` contiguous chunks while minimizing the largest chunk. Get this wrong and one stage becomes the bottleneck for the entire pipeline.

```python
# balance_stages.py
def min_max_chunk(costs: list[int], k: int) -> int:
    """Split `costs` into k contiguous chunks, minimizing the largest chunk sum."""
    def chunks_needed(limit: int) -> int:
        count, running = 1, 0
        for c in costs:
            if running + c > limit:
                count += 1
                running = c
            else:
                running += c
        return count

    lo, hi = max(costs), sum(costs)   # answer is somewhere in this range
    while lo < hi:
        mid = (lo + hi) // 2
        if chunks_needed(mid) <= k:
            hi = mid                   # mid is achievable, try smaller
        else:
            lo = mid + 1               # mid is too tight
    return lo

# Test: 4 layers of equal cost across 2 stages splits evenly
assert min_max_chunk([10, 10, 10, 10], 2) == 20

# Test: one heavy layer dominates no matter how you cut
assert min_max_chunk([1, 1, 100, 1], 2) == 102

# Test: more stages than layers is still well defined
assert min_max_chunk([5, 5, 5], 3) == 5
```

**Connection:** your `pipeline_parallel.py` splits 4 layers into 2 stages by hand, which is fine at that size. The second test above is the case worth sitting with: when one layer costs vastly more than the others, no contiguous split balances the pipeline, and the bubble stops being fixable by adding microbatches. That is the point at which a real system reaches for tensor parallelism *inside* that one heavy layer, combining both strategies, which is what large training runs actually do.

---

## Reflect

**What clicked:**

**What surprised me:**

**Measured bubble fraction at M = 1, 4, and 16, against the theoretical (S-1)/(M+S-1):**

**How wrong was the output when you removed the row-parallel all-reduce, and would any shape or type check have caught it?**

**Peak memory per worker, sharded vs unsharded optimizer state:**

**Which strategy did you pick for the too-large-layer scenario, and what interconnect assumption does it rest on?**

**What I'd do differently:**
