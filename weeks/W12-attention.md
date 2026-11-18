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
2. [ ] **Cache keys and values.** Process a prefix once and verify that incremental decoding matches the full-prefix reference within a declared tolerance.
3. [ ] **Route the requests.** Compare round-robin and prefix-affinity routing over the same interleaved conversations across two replicas.
4. [ ] **Measure locality.** Record recomputed prefill tokens and load per replica. Add a test that prevents cache entries from leaking across conversations.

**Minimum deliverable:** Matching cached and uncached outputs and a routing comparison on a fixed trace.

**Handoff:** Record which memory, latency or load constraint would govern a real deployment.
