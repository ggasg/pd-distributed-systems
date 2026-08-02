---
week_number: 15
status: not-started
---

# W15: GPU Memory and Compute

> **Arc:** Distributed ML & Compute · **Language:** C (primary) · Python/Numba CUDA optional, only if you have NVIDIA GPU access

## What you'll build
Two GEMM kernels in C: naive triple-loop, then cache-blocked (tile size 64, `i-k-j` loop order). Benchmark both on 1024×1024 float32 matrices, compare `-O2` against `-O3 -march=native` to measure what compiler auto-vectorization buys you, and plot the results on a roofline diagram. No GPU needed, and no hand-written SIMD intrinsics either: the compiler's vectorizer does that job portably, on Apple Silicon/NEON as much as on x86/AVX2. If you have NVIDIA GPU access, the same exercise as CUDA kernels via Numba is below, but it's optional and not required to complete this week.

**Scenario:** 1024 is a suspiciously convenient number, it's exactly 16 tiles of 64. Real matrices don't arrive sized to flatter your tiling scheme, and a blocked GEMM that only works when dimensions divide evenly is a bug waiting for whichever input size ships it to production.

---

## Read
- [ ] [Roofline: An Insightful Visual Performance Model for Multicore Architectures](https://people.eecs.berkeley.edu/~kubitron/cs252/handouts/papers/RooflineVyNoYellow.pdf) (Williams et al., 2009): read Sections 1–3. Understand arithmetic intensity and how to place your kernel on the roofline.
- [ ] Optional, background only: [CUDA C Programming Guide](https://docs.nvidia.com/cuda/cuda-c-programming-guide/), Chapters 1–3 (thread hierarchy, memory hierarchy). Useful context for how the industry-standard model maps onto what you're doing on CPU, registers/shared/global on GPU versus registers/L1-L2/main memory here, but not required reading if you're only doing the CPU path.

**Key question:** Calculate the arithmetic intensity of a naive matmul (flops per byte of memory traffic). Is it compute-bound or memory-bound on your hardware?

---

## Code

**Primary path (C, no GPU required):**

Project: `code/cpu-gemm/` (C, gcc/clang)

- [ ] `naive_gemm.c`: triple loop, row-major layout
- [ ] `blocked_gemm.c`: cache-blocked with tile size 64; reorder loops to `i-k-j` for cache-friendliness
- [ ] `benchmark.c`: time both kernels on 1024×1024 float32, print GFLOPS. Build `blocked_gemm.c` twice, once with `-O2` and once with `-O3 -march=native`, and compare the two GFLOPS numbers to see what auto-vectorization buys you. Confirm it actually vectorized with `clang -Rpass=loop-vectorize` or `gcc -fopt-info-vec-optimized` on the blocked loop, no need to write or verify any ISA-specific intrinsics by hand.
- [ ] `roofline.py`: small `matplotlib` script, reads the GFLOPS numbers `benchmark.c` prints and plots them against arithmetic intensity, same roofline diagram either path produces.

**Break it, then decide:**
- [ ] Run `blocked_gemm.c` against a naive triple-loop reference on a matrix size that does *not* divide evenly by your tile size, 1000×1000 instead of 1024×1024. Compare outputs element-by-element instead of trusting that "it ran without crashing" means "it's correct." If your tiled loops assume every tile is a full 64×64 block, the remainder rows and columns at the edges either get computed wrong or silently skipped, wrong numbers, no crash, no warning, exactly the kind of bug that survives a benchmark that only ever runs on convenient sizes.
- [ ] Fix it one of two ways and say why you picked it: pad both input matrices up to the next multiple of 64 with zeros before tiling (simple, reuses the exact same inner loop, costs some wasted compute and memory on the padding), or add explicit boundary handling to the blocked loops so the last tile in each dimension can be a partial tile (no waste, more branching in code that was fast partly because it had none). Which one keeps the loop's vectorization intact, and did you check?

**If you have NVIDIA GPU access (optional, not required):**

Project: `code/gpu-gemm/` (Python 3.13+, `numba-cuda`, `numpy`, `matplotlib`)

- [ ] `naive_gemm.py`: Numba CUDA kernel: `@cuda.jit` decorator, each thread computes one output element `C[row, col] = sum_k A[row, k] * B[k, col]`. Launch with `blocks = (N//32, N//32)`, `threads = (32, 32)`.
- [ ] `tiled_gemm.py`: Numba CUDA kernel with shared memory: use `cuda.shared.array(shape=(32, 32), dtype=float32)` for tiles of A and B; cooperatively load tiles, sync with `cuda.syncthreads()`, accumulate partial results.
- [ ] `benchmark.py`: allocate 1024×1024 float32 arrays, warm up (5 runs), time 20 runs using `cuda.event_elapsed_time`. Print GFLOPS for both kernels. Verify correctness against `numpy.matmul`.
- [ ] `roofline.py`: use `matplotlib` to draw the roofline: x-axis arithmetic intensity (flops/byte), y-axis attainable GFLOPS. Plot both kernels as points. Annotate with memory bandwidth and peak compute from `nvidia-smi`.

**Go automation tool:**

- [ ] `tools/bench_runner/main.go`: a small Go CLI (`os/exec`'s `exec.Command` to launch the benchmark subprocess and capture its stdout) that parses GFLOPS from the output and appends results as a row to `results.csv`. Usage: `go run . --kernel naive --runs 20` (`flag` package for the CLI args, standard library, no dependency needed for a program this small). Keep it under 80 lines: a quick return to the `net/http`/CLI style, not new territory; you last wrote something this shape for W12's gradient server.

---

## Reflect

**What clicked:**

**What surprised me:**

**Benchmark results:**
- Naive: __ GFLOPS
- Blocked (`-O2`): __ GFLOPS
- Blocked (`-O3 -march=native`): __ GFLOPS
- Vectorization speedup: __x
- Where does your kernel sit on the roofline?

**What a vendor-tuned BLAS library (Apple's Accelerate/vDSP, Intel MKL, or cuBLAS on GPU) does that your blocked kernel doesn't:**

**Padding or boundary handling for non-multiple-of-tile-size matrices, and did your fix stay vectorized (from Break it, then decide above)?**

**What I'd do differently:**
