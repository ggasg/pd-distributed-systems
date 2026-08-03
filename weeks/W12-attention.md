---
week_number: 12
status: not-started
---

# W12: Attention, KV Cache, and Cache-Aware Routing

> **Arc:** Distributed ML & Compute · **Language:** Python (NumPy only)
> **Budget:** about 5 hours. Hit the Minimum bar first; everything past it is optional.

## What you'll build
A KV cache for autoregressive generation (Part 1), then a router that has to decide which of two replicas a request should go to, given that each one holds a different cache (Part 2).

Part 1 builds the cache against a given multi-head self-attention implementation rather than one you derive from scratch: the attention forward pass itself is standard transformer mechanics that belongs to a dedicated ML/AI track, not this one. Its actual subject is what happens to memory and latency once you start caching past keys and values across generation steps, and where a naive cache can leak state between requests it was never supposed to share. No PyTorch.

Part 2 is the same cache one level up. Once you have more than one replica of a model, the cache stops being an implementation detail and becomes the thing that decides where a request should be sent, because sending it to the wrong replica means throwing away work that was already done.

**Scenario:** an inference server serving more than one conversation at a time is the actual, everyday shape of this problem. It's exactly where a naive KV cache can quietly let one user's context leak into another user's response, which is Part 1's failure mode, and it's also where a perfectly ordinary round-robin load balancer silently doubles your latency by routing turn two of a conversation to a replica that has never seen turn one, which is Part 2's.

---

## Part 1: The Cache Itself

### Read
- [ ] [Attention Is All You Need](https://arxiv.org/abs/1706.03762) (Vaswani et al., 2017): read Sections 3.2–3.5 (the attention mechanism and multi-head attention) to understand the mechanism `attention.py` gives you below, not to derive it yourself. Skip the training details.
- [ ] Optional: [FlashAttention: Fast and Memory-Efficient Exact Attention with IO-Awareness](https://arxiv.org/abs/2205.14135) (Dao et al., 2022): read Sections 1–3. The key insight is tiling attention to avoid materializing the full N×N attention matrix. You won't implement this, but you need to understand *why* naive attention is memory-bound.
- [ ] [Efficient Memory Management for Large Language Model Serving with PagedAttention](https://arxiv.org/abs/2309.06180) (Kwon et al., 2023): read Sections 1–4. Understand why KV cache fragmentation is a problem and how paging solves it.

**Depth: read Sections 3.2 to 3.5 of Attention Is All You Need and Sections 1 to 4 of PagedAttention.** No study reading: multi-head attention is given to you, and the cache and router you build are not described in any of these papers. FlashAttention, Burns Ch.15, and the Gateway API posts are skims.

**Key question:** Naive attention is O(N²) in memory. Where exactly does this quadratic memory come from? Draw the computation graph and label which tensors are the bottleneck.

**Optional: Burns, *Designing Distributed Systems*, 2nd ed., Chapter 15** (AI Inference and Serving). The three papers above are about the compute and memory mechanics inside one forward pass; Burns' chapter is the layer above that, hosting and distributing a model as a service. "Hosting a Model" and "Distributing a Model" are the production framing for what your `KVCache` and its memory/latency tradeoff exist to support, and it's a good warm-up for Part 2.

### Code

Project: `code/attention/` (Python 3.13+, NumPy only)

Model config: `d_model=64`, `n_heads=4`, `d_head=16`, `seq_len=32`, `vocab_size=256`.

**Given, not built:** `attention.py`'s `MultiHeadAttention` class is provided as a starter file: `__init__` (random `W_q`/`W_k`/`W_v`/`W_o` projections, shape `[d_model, d_model]`), `scaled_dot_product(Q, K, V, mask=None)` (`softmax(QK^T / sqrt(d_head)) V`, with an optional causal mask), and `forward(X)` (split into heads, apply SDPA per head, concatenate, project with `W_o`). Read it closely enough to know what shape `forward(X)` expects and returns, since `kv_cache.py` and `generate.py` both call into it directly, but you won't need to modify it. Deriving this mechanism yourself is real, valuable work; it's just not what this week is testing.

- [ ] `kv_cache.py`: `KVCache` class:
  - Stores past keys and values per layer: `Dict[int, Tuple[np.ndarray, np.ndarray]]`
  - `update(layer_id, new_k, new_v)`: concatenates new K/V to cached K/V along the sequence dimension
  - `get(layer_id)`: returns full K/V including cache
- [ ] `generate.py`: autoregressive generation:
  - Without cache: re-run full attention over the entire sequence each step (O(N²) per step)
  - With cache: run attention only for the new token against the cached K/V (O(N) per step)
  - Generate 20 tokens from a random start token; measure wall time and peak memory (`tracemalloc`) for both approaches
- [ ] Print the comparison from inside `generate.py` rather than a separate file: tokens generated, time in ms, and peak memory in MB, cached versus uncached. Two numbers side by side is the whole deliverable and it does not need its own module.

**Break it, then decide:**
- [ ] `KVCache` as specified stores past keys and values keyed only by `layer_id`. Simulate two concurrent "conversations" sharing one `KVCache` instance: generate a few tokens for prompt A, then interleave a few tokens for a completely unrelated prompt B, calling `update`/`get` on the same cache object for both. Because the cache has no notion of which request a K/V pair belongs to, B's tokens get concatenated onto A's cached keys and values at the same `layer_id`, so a later step generating for A ends up attending over tokens from B's prompt too. Confirm this by printing what `get(layer_id)` returns after the interleaving and checking whether it contains tokens from both prompts.
- [ ] Fix it by keying the cache by `(request_id, layer_id)` instead of `layer_id` alone, and decide: would you give each concurrent request its own `KVCache` instance (simple, no shared bookkeeping, but nothing enforces that a caller can't still pass the wrong instance to the wrong request), or one shared `KVCache` keyed internally by request ID (a single object every request goes through, so isolation is enforced in one place instead of trusted to every caller)? Implement whichever you pick, and re-run the interleaved test above to confirm A's generation no longer sees any of B's tokens.

---

## Part 2: Routing to a Cache

You just made the cache per-request. That change has a consequence that only shows up once there is more than one copy of your model running, and it is the thing this part is about.

Picture a chat service. A user sends a message, your server runs prefill over the whole conversation so far, caches the keys and values, and generates a reply. The user sends a follow-up. If that follow-up lands on the same replica, the cache is already there and you only prefill the new message. If it lands on a different replica, that replica has never seen this conversation, so it prefills the entire history again from scratch. Same answer, several times the latency, for no reason other than which machine picked up the request.

An ordinary load balancer cannot know any of this. Round-robin is correct, fair, and completely wrong here, because it treats replicas as interchangeable when the whole point is that they are not: **each one holds different state**. Routing that accounts for this is called cache-aware routing, and it is one of the fastest-moving parts of production inference infrastructure right now.

This part is deliberately small and entirely local. No Kubernetes, no GPU, no serving framework. Two replicas are two Python objects in one process, and the workload is simulated. The goal is to see the effect clearly, not to operate the real thing.

### Read
- [ ] [Introducing Gateway API Inference Extension](https://kubernetes.io/blog/2025/06/05/introducing-gateway-api-inference-extension/) (Kubernetes blog, June 2025): short and readable. Note what it says the router needs to know: KV cache utilization and which LoRA adapters a replica has loaded. Both are facts about a replica's *state*, which is precisely what a normal load balancer is designed not to care about.
- [ ] Optional: [KV cache aware routing with llm-d](https://developers.redhat.com/articles/2025/10/07/master-kv-cache-aware-routing-llm-d-efficient-ai-inference) (Red Hat). Describes the same idea running against real vLLM replicas, with reported improvements of up to 3x on time-to-first-token.

**Key question:** Round-robin balancing assumes every backend can serve every request equally well. Name exactly which assumption a KV cache breaks, and say whether the same problem would exist for a stateless service like an image resizer.

### Code

Same project, `code/attention/`. Still NumPy and the standard library.

- [ ] `replica.py`: a `Replica` class wrapping your Part 1 model and its request-keyed `KVCache`. Two methods, `prefill(request_id, tokens)` and `decode_step(request_id)`, plus two things to track: `has_cache_for(request_id) -> bool`, and a running counter of how many tokens it has spent on prefill. That counter is your whole metric, so make it honest.
- [ ] `router.py`: two routers over a list of replicas, both about ten lines.
  - `RoundRobinRouter`: hand each request to the next replica in order. This is what you get by default from essentially every load balancer in existence.
  - `CacheAwareRouter`: ask each replica whether it already holds a cache for this conversation, and prefer that one. Fall back to the least-loaded replica when nobody does.
- [ ] `bench_routing.py`: generate the workload and run it through both routers. The workload is a handful of multi-turn conversations, say 6 with 5 turns each, arriving interleaved so no conversation's turns are adjacent, with each turn appending to a growing history. That's what makes recomputed prefill expensive and increasingly so, and it's a dozen lines, so keep it here rather than in its own module. Run the identical workload through both routers with 2 replicas, and print total prefill tokens computed for each, plus the per-replica breakdown. Round-robin should be recomputing history on most turns; cache-aware should be recomputing almost none of it. Report the ratio.

**Minimum bar (Part 2):** a measured number, not an argument. Total prefill tokens under round-robin versus cache-aware, on the same workload, with the gap explained in one sentence.

**Break it, then decide:**
- [ ] Cache-aware routing sends every turn of a conversation to the same replica, which is the point. Now give one conversation a much longer history and many more turns than the others, and watch where it goes: that replica keeps getting picked, its queue grows, and the other one sits idle holding nothing useful. You have just recreated W05's skew problem one layer up. The cause is identical, a routing key whose distribution is uneven, and so is the symptom, one worker doing most of the work while the rest wait.
- [ ] **Your call:** these two goals genuinely conflict and no policy satisfies both. Pure cache affinity gives the best time-to-first-token and the worst load balance. Pure round-robin gives perfect balance and throws away cache constantly. Implement the middle option: prefer the replica holding the cache *unless* its queue depth exceeds some threshold, at which point you accept the prefill cost and send the request elsewhere. Pick the threshold, say what happens to your prefill-token number when you do, and name the metric you'd put on a dashboard to know whether the threshold was set wrong in either direction.

### Where this lives in the real world (read only)

Worth knowing, because it connects two weeks that otherwise sit apart. The production version of `router.py` is not part of the model server at all. It's a component of the Kubernetes control plane: the Gateway API Inference Extension is a Go project, and llm-d's Endpoint Picker is the same idea running as a Kubernetes-native service. Routing has to live there because it needs facts about every replica's state, and the thing that already tracks every replica is the control plane.

That is the same layer W14 is about, and it's the honest answer to why this curriculum has you write Go at all. The tensors are C++ and the model is Python, but deciding *which* replica gets a request, and which GPU that replica runs on, is Go, and it is where an inference platform is actually engineered.

---

## Reflect

**What clicked:**

**What surprised me:**

**Benchmark results:**
- Uncached 20 tokens: __ ms, __ MB peak
- Cached 20 tokens: __ ms, __ MB peak

**What PagedAttention solves that your KVCache doesn't:**

**Per-request cache instances or one cache keyed by request ID, and why (from Part 1's Break it, then decide)?**

**Total prefill tokens, round-robin vs cache-aware, on the same workload. Both numbers and the ratio.**

**Which assumption does a KV cache break that round-robin balancing relies on, and would a stateless service have the same problem?**

**Your queue-depth threshold: what number, what it cost you in recomputed prefill, and what you'd monitor to find out it was wrong.**

**Where the skew you caused in Part 2 is the same problem as W05's, and where it genuinely differs:**

**What I'd do differently:**
