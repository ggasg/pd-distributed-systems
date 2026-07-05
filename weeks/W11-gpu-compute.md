# W11 — GPU Memory and Compute

> **Arc:** Distributed ML & Compute · **Language:** CUDA C / C

## What you'll build
Two matrix multiplication kernels: (1) naive one-thread-per-output-element, (2) shared memory tiled. Benchmark both on 1024×1024 float32 matrices. Plot on a roofline diagram. If you don't have a GPU: implement cache-blocked GEMM in C with SIMD intrinsics and compare to a naive version.

---

## Read
- [ ] [CUDA C Programming Guide](https://docs.nvidia.com/cuda/cuda-c-programming-guide/) — read Chapters 1–3: thread hierarchy, memory hierarchy (registers, shared memory, L1/L2, global). Stop before Chapter 4. This is the mental model you need.
- [ ] [Roofline: An Insightful Visual Performance Model for Multicore Architectures](https://people.eecs.berkeley.edu/~kubitron/cs252/handouts/papers/RooflineVyNoYellow.pdf) (Williams et al., 2009) — read Sections 1–3. Understand arithmetic intensity and how to locate your kernel on the roofline.

**Key question:** What is arithmetic intensity? Calculate the arithmetic intensity of a naive matmul (flops per byte of memory traffic). Is it compute-bound or memory-bound?

---

## Code

**If you have GPU access:**

Project: `code/gpu-gemm/` (CUDA, nvcc)

- [ ] `naive_gemm.cu` — kernel: each thread computes one output element `C[row][col] = sum_k A[row][k] * B[k][col]`. Launch with `dim3 grid((N+31)/32, (N+31)/32)`, `dim3 block(32, 32)`.
- [ ] `tiled_gemm.cu` — kernel: thread block cooperatively loads a 32×32 tile of A and B into shared memory, computes partial dot products, advances to next tile. Synchronize with `__syncthreads()`.
- [ ] `benchmark.cu` — allocate 1024×1024 float32 matrices on device, warm up (5 runs), time 20 runs with `cudaEventRecord`. Print GFLOPS for both kernels. Verify correctness against CPU reference.
- [ ] `roofline.md` — using your GPU's specs (peak GFLOPS, memory bandwidth from `nvidia-smi`), draw the roofline and plot both kernels. Which bottleneck does tiling address?

**If no GPU (fallback):**

Project: `code/cpu-gemm/` (C, gcc/clang)

- [ ] `naive_gemm.c` — triple loop, row-major layout
- [ ] `blocked_gemm.c` — cache-blocked with tile size 64; reorder loops for cache-friendliness
- [ ] `simd_gemm.c` — use AVX2 intrinsics (`_mm256_fmadd_ps`) for the innermost loop
- [ ] `benchmark.c` — time all three on 1024×1024 float32, print GFLOPS

---

## Reflect

**What clicked:**

**What surprised me:**

**Benchmark results:**
- Naive: __ GFLOPS
- Tiled/blocked: __ GFLOPS
- Speedup: __x
- Where does your kernel sit on the roofline?

**What cuBLAS does that your tiled kernel doesn't:**

**What I'd do differently:**
