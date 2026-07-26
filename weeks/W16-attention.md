---
week_number: 16
status: not-started
---

# W16: Attention and KV Cache

> **Arc:** Distributed ML & Compute · **Language:** Python (NumPy only)

## What you'll build
Multi-head self-attention forward pass, then extend it with a KV cache for autoregressive generation. No PyTorch. Benchmark the memory usage and latency of cached vs uncached generation.

**Scenario:** an inference server serving more than one conversation at a time is the actual, everyday shape of this problem, and it's also exactly where a naive KV cache can quietly let one user's context leak into another user's response. That failure mode is built directly into this week's code so you find it here, not in an incident report.

---

## Read
- [ ] [Attention Is All You Need](https://arxiv.org/abs/1706.03762) (Vaswani et al., 2017): read Sections 3.2–3.5 (the attention mechanism and multi-head attention). Skip the training details.
- [ ] [FlashAttention: Fast and Memory-Efficient Exact Attention with IO-Awareness](https://arxiv.org/abs/2205.14135) (Dao et al., 2022): read Sections 1–3. The key insight is tiling attention to avoid materializing the full N×N attention matrix. You won't implement this, but you need to understand *why* naive attention is memory-bound.
- [ ] [Efficient Memory Management for Large Language Model Serving with PagedAttention](https://arxiv.org/abs/2309.06180) (Kwon et al., 2023): read Sections 1–4. Understand why KV cache fragmentation is a problem and how paging solves it.

**Key question:** Naive attention is O(N²) in memory. Where exactly does this quadratic memory come from? Draw the computation graph and label which tensors are the bottleneck.

**Optional: Burns, *Designing Distributed Systems*, 2nd ed., Chapter 15** (AI Inference and Serving). The three papers above are about the compute and memory mechanics inside one forward pass; Burns' chapter is the layer above that, hosting and distributing a model as a service. "Hosting a Model" and "Distributing a Model" are the production framing for what your `KVCache` and its memory/latency tradeoff exist to support.

---

## Code

Project: `code/attention/` (Python 3.11+, NumPy only)

Model config: `d_model=64`, `n_heads=4`, `d_head=16`, `seq_len=32`, `vocab_size=256`.

- [ ] `attention.py`: `MultiHeadAttention` class:
  - `__init__`: initialize `W_q`, `W_k`, `W_v`, `W_o` as random NumPy arrays (shape `[d_model, d_model]`)
  - `scaled_dot_product(Q, K, V, mask=None)`: compute `softmax(QK^T / sqrt(d_head)) V`; apply causal mask (upper triangle = -inf) when `mask=True`
  - `forward(X)`: split into heads, apply SDPA per head, concatenate, project with `W_o`; input shape `[batch, seq, d_model]`
- [ ] `kv_cache.py`: `KVCache` class:
  - Stores past keys and values per layer: `Dict[int, Tuple[np.ndarray, np.ndarray]]`
  - `update(layer_id, new_k, new_v)`: concatenates new K/V to cached K/V along the sequence dimension
  - `get(layer_id)`: returns full K/V including cache
- [ ] `generate.py`: autoregressive generation:
  - Without cache: re-run full attention over the entire sequence each step (O(N²) per step)
  - With cache: run attention only for the new token against the cached K/V (O(N) per step)
  - Generate 20 tokens from a random start token; measure wall time and peak memory (`tracemalloc`) for both approaches
- [ ] `benchmark.py`: print comparison table: tokens generated, time (ms), peak memory (MB), for cached vs uncached

**Break it, then decide:**
- [ ] `KVCache` as specified stores past keys and values keyed only by `layer_id`. Simulate two concurrent "conversations" sharing one `KVCache` instance: generate a few tokens for prompt A, then interleave a few tokens for a completely unrelated prompt B, calling `update`/`get` on the same cache object for both. Because the cache has no notion of which request a K/V pair belongs to, B's tokens get concatenated onto A's cached keys and values at the same `layer_id`, so a later step generating for A ends up attending over tokens from B's prompt too. Confirm this by printing what `get(layer_id)` returns after the interleaving and checking whether it contains tokens from both prompts.
- [ ] Fix it by keying the cache by `(request_id, layer_id)` instead of `layer_id` alone, and decide: would you give each concurrent request its own `KVCache` instance (simple, no shared bookkeeping, but nothing enforces that a caller can't still pass the wrong instance to the wrong request), or one shared `KVCache` keyed internally by request ID (a single object every request goes through, so isolation is enforced in one place instead of trusted to every caller)? Implement whichever you pick, and re-run the interleaved test above to confirm A's generation no longer sees any of B's tokens.

---

## Reflect

**What clicked:**

**What surprised me:**

**Benchmark results:**
- Uncached 20 tokens: __ ms, __ MB peak
- Cached 20 tokens: __ ms, __ MB peak

**What PagedAttention solves that your KVCache doesn't:**

**Per-request cache instances or one cache keyed by request ID, and why (from Break it, then decide above)?**

**What I'd do differently:**
