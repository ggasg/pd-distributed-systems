---
week_number: 10
status: not-started
---

# W10: Beyond Data Parallelism

> **Arc:** Distributed ML Systems · **Language:** Python
> **Budget:** about 10 hours. The Minimum bar is what a bad week looks like, not the target.

## What you'll build

Splitting the model rather than the data, across two processes, three ways: tensor parallelism cuts a single matrix multiply in half, pipeline parallelism puts different layers on different workers, and sharded optimizer state gives each worker only its slice of the optimizer's bookkeeping. All three reuse the ring-allreduce you wrote in W09 as their communication layer.

W09 was **data parallelism**: every worker holds a complete copy of the model and they split the data between them. That works right up until the model itself no longer fits in one worker's memory, at which point you have to cut the model up instead. This unit is about how you cut it, and what each cut costs in communication.

**Scenario:** you have a model that trains fine on one machine and a bigger one that does not fit at all. Somebody asks which parallelism strategy to use, and the honest answer depends on numbers you don't have yet: how much has to cross the network per step, and how much of the time each worker spends waiting. This unit you measure both on something small enough to see clearly.

---

## Read

- [ ] [Megatron-LM](https://arxiv.org/abs/1909.08053) (Shoeybi et al., 2019): read Sections 1 and 3. Section 3 is the whole idea: split one weight matrix by columns, split the next by rows, and the two cuts compose so that a full MLP block needs exactly one all-reduce instead of two. That trick is the reason tensor parallelism is practical.
- [ ] [GPipe](https://arxiv.org/abs/1811.06965) (Huang et al., NeurIPS 2019): read Sections 1 to 3. The key concept is the "bubble," the fraction of time a worker spends idle waiting for the stage before it, and how splitting each batch into microbatches shrinks it.
- [ ] [ZeRO](https://arxiv.org/abs/1910.02054) (Rajbhandari et al., SC 2020): read Section 5 and the stage table. You do not need the full memory analysis. What you want is the three stages: shard the optimizer state, then the gradients, then the parameters themselves, each stage saving more memory and costing more communication.
- [ ] Optional: [PyTorch FSDP](https://arxiv.org/abs/2304.11277) (Zhao et al., VLDB 2023): what ZeRO looks like once it's a production API people actually call. Useful if you want to see how the theory above landed in the framework.

**Depth: study Section 3 of Megatron-LM.** The column-then-row composition is the one non-obvious idea in this unit, and you implement it. GPipe and ZeRO are reads, and you only need the named sections. FSDP is an optional skim.

**Key question:** Tensor parallelism communicates once per layer; pipeline parallelism communicates once per stage boundary. Given that, which one would you put inside a single machine across its GPUs, and which one across machines on a slower network? The answer follows directly from how often each one talks.

---

## Code

Project: `code/parallelism/` (Python 3.12+)

Dependencies: `numpy`, `multiprocessing`. You'll import `ring_allreduce` and `all_gather` directly from `code/distributed-training/` (W09), so this unit builds on real code you already wrote rather than a fresh abstraction.

**Write once, then leave alone:** `layers.py`, a NumPy `Linear` class with `forward` and `backward` plus a GeLU activation. As in W09, the calculus is not what this unit tests. Reuse an implementation if you have one.

**Dimensions, because the numbers you measure depend entirely on them.** Use `d_model = 1024`, a hidden width of `4096` (the 4x expansion Megatron's MLP block uses), and a batch of 256, with synthetic random inputs. There is no dataset here; the tensors are the workload.

The reason to pin these: every measurement in this unit is a ratio between compute and communication, and at small dimensions communication wins by default and tells you nothing. Before trusting any bubble number in Part 2, time one stage's forward pass and time one socket round-trip, and print both. You want per-stage compute to be at least an order of magnitude above the round-trip. If it is not, you are measuring socket latency rather than the bubble: it will not track the theoretical `(S-1)/(M+S-1)` curve, and the mismatch will look like a bug in your pipeline when the dimensions are the problem. Widen the layers until the ratio holds, and write down the two timings next to your results.

### Part 1: Tensor parallelism

- [ ] `tensor_parallel.py`: implement `ColumnParallelLinear` and `RowParallelLinear` for two workers.
  - Column-parallel: split the weight matrix `A` by columns, so worker `i` holds `A_i`. Each worker computes `X @ A_i` on the full input and the outputs are concatenated. No communication needed in the forward pass, the halves are just two pieces of one wider output.
  - Row-parallel: split the weight matrix `B` by rows and the input by columns, so worker `i` computes `X_i @ B_i`. Each worker now holds a *partial sum* of the correct output, and an all-reduce is what turns those partials into the real answer.
- [ ] `mlp_block.py`: chain them the way Megatron does, column-parallel then GeLU then row-parallel, and confirm the whole block needs exactly one all-reduce. Assert the result matches a single-process reference implementation to within floating-point tolerance.

### Part 2: Pipeline parallelism

- [ ] `pipeline_parallel.py`: split a 4-layer MLP into two stages, layers 1 and 2 on worker 0, layers 3 and 4 on worker 1. Forward activations cross the stage boundary over a socket; gradients cross back the other way. Start with one batch at a time.
- [ ] `microbatch.py`: split each batch into `M` microbatches and feed them into the pipeline back to back, so worker 1 can start on microbatch 1 while worker 0 is already working on microbatch 2.
- [ ] `balance.py`: your pipeline splits four equal layers across two stages, which needs no thought. Real stacks are not equal, so write the general version: given a list of per-layer costs and a stage count `k`, split the list into `k` contiguous chunks minimising the largest chunk sum. Binary search on the answer, about fifteen lines. Test it on `[10, 10, 10, 10]` with `k=2`, which splits evenly, and then on `[1, 1, 100, 1]` with `k=2`, which cannot: the answer is 102 no matter where you cut. That second case is the one to sit with, because it means no contiguous split balances the pipeline and adding microbatches will not save you. It is the point at which a real system reaches for tensor parallelism *inside* the heavy layer, which is Part 1, and combining the two is what large training runs actually do.
- [ ] `bubble.py`: instrument both workers with timers that record idle time, then sweep `M` from 1 to 16 and print measured bubble fraction alongside the theoretical `(S-1)/(M+S-1)`. With two stages and `M=1` you should measure close to 50 percent idle, and it should fall off quickly as `M` rises.

### Part 3: Sharded optimizer state

- [ ] `shard_optimizer.py`: implement ZeRO stage 1 for SGD with momentum. Each worker keeps momentum buffers for only its shard of the parameters, applies its own slice of the update, and then all-gathers the updated parameters so everyone ends the step with a complete model. Print peak memory per worker (`tracemalloc` is enough) against a non-sharded baseline, and confirm the loss curve is identical, because this is purely a memory optimization and must not change the math at all.

**Minimum bar:** Parts 1 and 2 only. The Megatron block matches a single-process reference and needs one all-reduce, and you have measured bubble fraction at three values of M against the theoretical curve. Part 3 (ZeRO) is the part to drop if the unit runs out, since it is a memory optimization that provably does not change the math, which makes it the least costly thing to postpone.

---

## Break it, then decide

- [ ] In `RowParallelLinear`, comment out the all-reduce. The program does not crash, does not warn, and produces output of exactly the right shape. It is simply wrong, because every worker is holding a partial sum. Compare against the single-process reference and look at how wrong it actually is, then let training run a few steps and watch the loss go somewhere strange rather than error. Silent numerical corruption in a distributed layer is one of the genuinely hard bug classes in this field, and the lesson worth taking is that the shape check you'd normally trust tells you nothing here.
- [ ] Set `M=1` in the pipeline and confirm you measure roughly the theoretical bubble. Then reason about what would happen with 8 stages and `M=1`: the formula says 87.5 percent idle, and that is why nobody runs pipeline parallelism without microbatching.
- [ ] **Your call:** you're given a model where a single layer's weight matrix is too large to fit on one device. Tensor parallelism solves this directly but has to communicate inside every layer, so it needs fast interconnect. Pipeline parallelism communicates only at stage boundaries, which is far less traffic, but it cannot split an individual layer at all and it wastes time in the bubble. Pick one for this specific constraint, implement it as your `mlp_block.py`'s default path, and write down the interconnect assumption your choice depends on. Then say what you'd change if you were told the workers were in different data centers.

## Reflect

**Prediction versus measurement.** Fill the predictions in *before* you run anything, and do not edit them afterwards. The gap is where calibration comes from.

| Quantity | Predicted | Measured | Which term I got wrong |
|----------|-----------|----------|------------------------|
| | | | |

Copy anything worth carrying into [MEASUREMENTS.md](../MEASUREMENTS.md).

**What clicked:**

**What surprised me:**

**Measured bubble fraction at M = 1, 4, and 16, against the theoretical (S-1)/(M+S-1):**

**How wrong was the output when you removed the row-parallel all-reduce, and would any shape or type check have caught it?**

**Peak memory per worker, sharded vs unsharded optimizer state:**

**Which strategy did you pick for the too-large-layer scenario, and what interconnect assumption does it rest on?**

**What I'd do differently:**
