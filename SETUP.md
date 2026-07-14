# Environment Setup

Everything you need installed before starting W00. Set this up once; it covers the entire curriculum.

---

## Obsidian

1. Download from [obsidian.md](https://obsidian.md) and install
2. Open vault: **File → Open folder as vault** → select this repo directory
3. Install required community plugins (Settings → Community plugins → Browse):
   - **Dataview**: powers the Home.md dashboard (enable JavaScript queries)
   - **Tasks**: optional; used for task filtering
   - **Obsidian Git**: optional; auto-syncs vault to GitHub
4. Set `Home.md` as startup note: Settings → General → Default startup page → `Home`

---

## Go 1.22+

```bash
# macOS
brew install go
go version   # go1.22.x

# Or download from https://go.dev/dl/
```

**W00–W04, W14, W16, and secondary tooling in W03/W10/W12/W15/W17** use Go — this is the backbone language of the curriculum alongside C++. `go mod init <name>` scaffolds a project; there's no separate package-manager install step, `go build`/`go run`/`go test` fetch whatever `go.mod` declares.

**W16 (Kubernetes Operators) requires Go** — this isn't optional the way secondary tooling elsewhere is. Install `controller-runtime`:
```bash
go get sigs.k8s.io/controller-runtime@v0.18.0
```

**New to Go?** Start here, not at W05. Go's learning curve is short by design — the language spec is deliberately small, there's no ownership model or macro system to internalize — but it's still worth a dedicated pass before W01 rather than learning it while also learning LSM-trees. Work through [A Tour of Go](https://go.dev/tour/) (free, interactive, ~2–3 hours) end to end, then read [Effective Go](https://go.dev/doc/effective_go)'s sections on goroutines, channels, and error handling (~1 hour). That's enough to be productive in W00–W04. The one habit worth building early: Go returns errors as values (`result, err := doThing()`) instead of throwing exceptions — get comfortable checking `err != nil` everywhere, it's idiomatic, not boilerplate to work around.

---

## C++ (C++20, CMake)

```bash
# macOS
xcode-select --install        # ships clang
brew install cmake ninja vcpkg

cmake --version                # 3.25+ recommended
clang++ --version               # or g++ --version if you prefer GCC

# vcpkg (C++ package manager, used for prometheus-cpp / opentelemetry-cpp / GoogleTest in W05–W08, W17)
git clone https://github.com/microsoft/vcpkg ~/vcpkg
~/vcpkg/bootstrap-vcpkg.sh
export VCPKG_ROOT=~/vcpkg
```

Each C++ project (`code/streaming/`, `code/dd-scratch/`, etc.) is its own CMake project. `CMakeLists.txt` declares the target and its dependencies (via `find_package`, with vcpkg supplying the package); the standard build sequence is:
```bash
cmake -S . -B build -DCMAKE_BUILD_TYPE=Release -DCMAKE_TOOLCHAIN_FILE=$VCPKG_ROOT/scripts/buildsystems/vcpkg.cmake
cmake --build build
ctest --test-dir build      # GoogleTest suite, where the project has one
```
Most Arc 2 weeks (W05–W07) declare zero or one dependency (GoogleTest); W17's Prometheus/OTel instrumentation is the one place this arc pulls in real external libraries.

**W05–W08** target C++20 (structured bindings, `std::variant`, concepts where useful) — modern enough to be relevant to the codebases W06 and W08 point you at (PyTorch's autograd engine, DuckDB's execution engine), most of which build against C++17/20 themselves. No nightly/experimental compiler flags needed; mainline `clang` or `gcc` from Homebrew is current enough.

**Already know C++?** If your C++ is from school or an earlier job, most of the syntax will come back fast, but treat this as a refresh into *modern* idioms rather than a cold start — the gap is usually smart pointers (`std::unique_ptr`/`std::shared_ptr` instead of raw `new`/`delete`), move semantics and RAII (resource cleanup tied to scope, the closest thing C++ has to what the borrow checker gave you automatically in a hypothetical Rust track), and STL algorithms/containers (`std::vector`, `std::unordered_map`, `<algorithm>`) instead of hand-rolled arrays and loops. Before starting W05, read [A Tour of C++](https://www.stroustrup.com/tour3.html) (Stroustrup, free chapter previews / short book) Chapters 1 (Basics), 4 (Classes), 5 (Essential Operations — this is where move semantics and RAII live), and 8 (Templates); pair it with cppreference's pages on [smart pointers](https://en.cppreference.com/w/cpp/memory) and [move semantics](https://en.cppreference.com/w/cpp/language/move_constructor). Budget 4–6 hours — less than a from-scratch language, since the syntax and control flow are already familiar, but real time nonetheless for the idioms that changed since you last wrote C++.

**One thing that has no Rust equivalent to complain about:** C++ won't stop you from writing something that compiles but is wrong — no borrow checker catches a dangling reference or a data race for you here. The weeks' "Constraints" sections call out, explicitly, where you're now responsible for a discipline (encapsulation, immutability of returned collections) that used to be compiler-enforced. Read those callouts; they're not boilerplate.

**CMake itself is a real ramp, separate from the language.** Cargo's zero-config "it just builds" experience has no CMake equivalent — expect `CMakeLists.txt` boilerplate and `find_package`/vcpkg wiring to cost real time in W05, even though the C++ language itself is familiar. If a project won't configure, check the vcpkg toolchain file path before anything else; it's the most common first-week failure.

---

## Python 3.11+

Use [pyenv](https://github.com/pyenv/pyenv) to manage Python versions.

```bash
# macOS
brew install pyenv
pyenv install 3.11.9
pyenv global 3.11.9
python --version   # 3.11.x
```

Install dependencies per arc:

**Arc 3 base (W09–W13):**
```bash
pip install numpy torch torchvision duckdb pyarrow pandas "ray[default]"
```

**W09 (ML pipelines):**
```bash
pip install duckdb pyarrow pandas
```

**W10 (distributed training):**
```bash
pip install numpy torch           # torch for MNIST loading only
```

**W11 (actor model / Ray):**
```bash
pip install "ray[default]" torch  # torch for the CNN, Ray for actors
```

**W12 (GPU compute), requires NVIDIA GPU:**
```bash
pip install numba cupy-cuda12x matplotlib
```
No GPU? The week includes a C fallback. Numba's CPU JIT still demonstrates the roofline model.

**W13 (attention):**
```bash
pip install numpy                 # NumPy only, no PyTorch for this week
```

---

## ClickHouse + PySpark (W07 only)

W07's Part 2 comparison exercise runs two real local systems alongside your C++ build — both single-machine, no account, no cluster.

```bash
# ClickHouse: single local server, no cluster
brew install clickhouse

# Spark runs on the JVM regardless of the Python surface — install a JDK if you don't have one
brew install openjdk@17

# PySpark, against your existing Python 3.11 setup — no version conflicts
pip install pyspark
```

Verify:
```bash
clickhouse server &        # then, in another terminal:
clickhouse client --query "SELECT 1"

python -c "from pyspark.sql import SparkSession; SparkSession.builder.master('local[*]').getOrCreate(); print('ok')"
```

Neither tool is needed outside W07 — uninstall or ignore afterward if you'd rather not keep them around.

---

## Docker + kind (W00, W16, W17)

```bash
# Docker Desktop: https://www.docker.com/products/docker-desktop/
# kind (Kubernetes-in-Docker):
brew install kind kubectl helm
```

Verify:
```bash
kind create cluster --name pd-systems
kubectl cluster-info --context kind-pd-systems
kind delete cluster --name pd-systems
```

---

## GPU Setup (W12, optional)

**NVIDIA GPU required.** If you don't have one, skip the Numba CUDA path and use the C fallback.

1. Install [CUDA Toolkit 12.x](https://developer.nvidia.com/cuda-downloads)
2. Verify: `nvcc --version`
3. Install Numba: `pip install numba`
4. Test: `python -c "from numba import cuda; print(cuda.gpus)"`

---

## Verify Everything

```bash
go version           # 1.22.x
cmake --version      # 3.25.x or later
clang++ --version    # or g++ --version
python --version     # 3.11.x or 3.12.x
docker --version     # 25.x or later
kind --version       # 0.22.x or later
kubectl version      # 1.29.x or later
helm version         # 3.14.x or later
```

---

## Recommended IDEs

| Language | IDE |
|----------|-----|
| Go | VS Code + Go extension, or GoLand |
| C++ | VS Code + clangd extension, or CLion |
| Python | VS Code + Pylance, or PyCharm Community |
| All | Neovim with LSP (if you're into that) |
