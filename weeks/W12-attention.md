---
week_number: 5
track: extension
title: "ML extension: attention and cache-aware serving"
status: not-started
---

# ML extension: attention and cache-aware serving

**Budget:** 10 hours across four study sessions. Shift the sessions to fit your week.

**Start with:** The reference-and-measure workflow from parallelism.

## Read

Read the attention section of [Attention Is All You Need](https://arxiv.org/abs/1706.03762) and the motivation for KV-cache management in [PagedAttention](https://arxiv.org/abs/2309.06180).

## Sessions

1. [ ] **Build the reference.** Implement a small causal attention layer with fixed NumPy weights and save its uncached outputs.

   Implement `causal_attention(x, weights)` using fixed Q/K/V projections and masking future positions before softmax. Save prefix outputs with a declared numeric tolerance and test that changing a future token cannot change an earlier output.
2. [ ] **Cache keys and values.** Process a prefix once and verify that incremental decoding matches the full-prefix reference within a declared tolerance.

   Implement `decode_step(token, cache) -> (output, next_cache)`. Append just the new token's key and value and compare every position with the full-prefix reference. Check empty cache and multi-token prefixes.
3. [ ] **Route the requests.** Compare round-robin and prefix-affinity routing over the same interleaved conversations across two replicas.

   Provide `route(request, replicas)` for round robin and prefix affinity. Key cached prefixes by model, conversation and exact token prefix. Replay the same ordered request trace with cold caches and the same capacity.
4. [ ] **Measure locality.** Record recomputed prefill tokens and load per replica. Add a test that prevents cache entries from leaking across conversations.

   Count recomputed prefix tokens and requests per replica. Test that two conversations with identical tokens do not share conversation-scoped state. Report locality savings and the resulting load imbalance.

**Minimum deliverable:** Matching cached and uncached outputs and a routing comparison on a fixed trace.

**Handoff:** Record which memory, latency or load constraint would govern a real deployment.
