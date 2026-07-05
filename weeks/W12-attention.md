# W12 — Attention and KV Cache

> **Arc:** Distributed ML & Compute · **Language:** Python (NumPy only)

## What you'll build
Multi-head self-attention forward pass, then extend it with a KV cache for autoregressive generation. No PyTorch. Benchmark the memory usage and latency of cached vs uncached generation.

---

## Read
- [ ] [Attention Is All You Need](https://arxiv.org/abs/1706.03762) (Vaswani et al., 2017) — read Sections 3.2–3.5 (the attention mechanism and multi-head attention). Skip the training details.
- [ ] [FlashAttention: Fast and Memory-Efficient Exact Attention with IO-Awareness](https://arxiv.org/abs/2205.14135) (Dao et al., 2022) — read Sections 1–3. The key insight is tiling attention to avoid materializing the full N×N attention matrix. You won't implement this, but you need to understand *why* naive attention is memory-bound.
- [ ] [Efficient Memory Management for Large Language Model Serving with PagedAttention](https://arxiv.org/abs/2309.06180) (Kwon et al., 2023) — read Sections 1–4. Understand why KV cache fragmentation is a problem and how paging solves it.

**Key question:** Naive attention is O(N²) in memory. Where exactly does this quadratic memory come from? Draw the computation graph and label which tensors are the bottleneck.

---

## Code

Project: `code/attention/` (Python 3.11+, NumPy only)

Model config: `d_model=64`, `n_heads=4`, `d_head=16`, `seq_len=32`, `vocab_size=256`.

- [ ] `attention.py` — `MultiHeadAttention` class:
  - `__init__`: initialize `W_q`, `W_k`, `W_v`, `W_o` as random NumPy arrays (shape `[d_model, d_model]`)
  - `scaled_dot_product(Q, K, V, mask=None)` — compute `softmax(QK^T / sqrt(d_head)) V`; apply causal mask (upper triangle = -inf) when `mask=True`
  - `forward(X)` — split into heads, apply SDPA per head, concatenate, project with `W_o`; input shape `[batch, seq, d_model]`
- [ ] `kv_cache.py` — `KVCache` class:
  - Stores past keys and values per layer: `Dict[int, Tuple[np.ndarray, np.ndarray]]`
  - `update(layer_id, new_k, new_v)` — concatenates new K/V to cached K/V along the sequence dimension
  - `get(layer_id)` — returns full K/V including cache
- [ ] `generate.py` — autoregressive generation:
  - Without cache: re-run full attention over the entire sequence each step (O(N²) per step)
  - With cache: run attention only for the new token against the cached K/V (O(N) per step)
  - Generate 20 tokens from a random start token; measure wall time and peak memory (`tracemalloc`) for both approaches
- [ ] `benchmark.py` — print comparison table: tokens generated, time (ms), peak memory (MB), for cached vs uncached

---

## Reflect

**What clicked:**

**What surprised me:**

**Benchmark results:**
- Uncached 20 tokens: __ ms, __ MB peak
- Cached 20 tokens: __ ms, __ MB peak

**What PagedAttention solves that your KVCache doesn't:**

**What I'd do differently:**
