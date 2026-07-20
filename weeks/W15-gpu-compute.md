---
week_number: 15
status: not-started
---

# W15: GPU Memory and Compute

> **Arc:** Distributed ML & Compute · **Language:** Python (Numba) · **Fallback:** C

## What you'll build
Two matrix multiplication kernels: (1) naive one-thread-per-output-element, (2) shared memory tiled. Written in Python using Numba's CUDA JIT. Benchmark both on 1024×1024 float32 matrices and plot on a roofline diagram. If no GPU: cache-blocked GEMM in C with SIMD intrinsics.

---

## Read
- [ ] [CUDA C Programming Guide](https://docs.nvidia.com/cuda/cuda-c-programming-guide/): read Chapters 1–3: thread hierarchy, memory hierarchy (registers, shared memory, L1/L2, global). The concepts apply regardless of whether you write CUDA in C or Python.
- [ ] [Roofline: An Insightful Visual Performance Model for Multicore Architectures](https://people.eecs.berkeley.edu/~kubitron/cs252/handouts/papers/RooflineVyNoYellow.pdf) (Williams et al., 2009): read Sections 1–3. Understand arithmetic intensity and how to place your kernel on the roofline.

**Key question:** Calculate the arithmetic intensity of a naive matmul (flops per byte of memory traffic). Is it compute-bound or memory-bound on your hardware?

---

## Code

**If you have GPU access:**

Project: `code/gpu-gemm/` (Python 3.11+, `numba`, `numpy`, `matplotlib`)

- [ ] `naive_gemm.py`: Numba CUDA kernel: `@cuda.jit` decorator, each thread computes one output element `C[row, col] = sum_k A[row, k] * B[k, col]`. Launch with `blocks = (N//32, N//32)`, `threads = (32, 32)`.
- [ ] `tiled_gemm.py`: Numba CUDA kernel with shared memory: use `cuda.shared.array(shape=(32, 32), dtype=float32)` for tiles of A and B; cooperatively load tiles, sync with `cuda.syncthreads()`, accumulate partial results.
- [ ] `benchmark.py`: allocate 1024×1024 float32 arrays, warm up (5 runs), time 20 runs using `cuda.event_elapsed_time`. Print GFLOPS for both kernels. Verify correctness against `numpy.matmul`.
- [ ] `roofline.py`: use `matplotlib` to draw the roofline: x-axis arithmetic intensity (flops/byte), y-axis attainable GFLOPS. Plot both kernels as points. Annotate with memory bandwidth and peak compute from `nvidia-smi`.

**If no GPU (fallback):**

Project: `code/cpu-gemm/` (C, gcc/clang)

- [ ] `naive_gemm.c`: triple loop, row-major layout
- [ ] `blocked_gemm.c`: cache-blocked with tile size 64; reorder loops to `i-k-j` for cache-friendliness
- [ ] `simd_gemm.c`: use AVX2 intrinsics (`_mm256_fmadd_ps`) for the innermost loop
- [ ] `benchmark.c`: time all three on 1024×1024 float32, print GFLOPS

**Java automation tool:**

- [ ] `tools/bench_runner/BenchRunner.java`: a small Java CLI (`ProcessBuilder` to launch the benchmark subprocess and capture its stdout) that parses GFLOPS from the output and appends results as a row to `results.csv`. Usage: `java BenchRunner.java --kernel naive --runs 20` (single-file source-code execution, no build step needed for a program this small, one of Java 21's genuinely beginner-friendly features). Keep it under 80 lines: a quick return to the JDK HttpServer/CLI style, not new territory; you last wrote something this shape for W13's gradient server.

---

## Reflect

**What clicked:**

**What surprised me:**

**Benchmark results:**
- Naive: __ GFLOPS
- Tiled/blocked: __ GFLOPS
- Speedup: __x
- Where does your kernel sit on the roofline?

**What cuBLAS / vendor libraries do that your tiled kernel doesn't:**

**What I'd do differently:**
