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

**W00–W04, W14, W16, and secondary tooling in W03/W10/W12/W15/W17** use Go — this is the backbone language of the curriculum alongside Rust. `go mod init <name>` scaffolds a project; there's no separate package-manager install step, `go build`/`go run`/`go test` fetch whatever `go.mod` declares.

**W16 (Kubernetes Operators) requires Go** — this isn't optional the way secondary tooling elsewhere is. Install `controller-runtime`:
```bash
go get sigs.k8s.io/controller-runtime@v0.18.0
```

**New to Go?** Start here, not at W05. Go's learning curve is short by design — the language spec is deliberately small, there's no ownership model or macro system to internalize — but it's still worth a dedicated pass before W01 rather than learning it while also learning LSM-trees. Work through [A Tour of Go](https://go.dev/tour/) (free, interactive, ~2–3 hours) end to end, then read [Effective Go](https://go.dev/doc/effective_go)'s sections on goroutines, channels, and error handling (~1 hour). That's enough to be productive in W00–W04. The one habit worth building early: Go returns errors as values (`result, err := doThing()`) instead of throwing exceptions — get comfortable checking `err != nil` everywhere, it's idiomatic, not boilerplate to work around.

---

## Rust (stable, 2021 edition)

```bash
# macOS / Linux
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source "$HOME/.cargo/env"
rustc --version   # rustc 1.8x.x or later
cargo --version
```

Each Rust project (`code/streaming/`, `code/dd-scratch/`, etc.) is its own Cargo project. `cargo new --lib <name>` scaffolds it; `Cargo.toml` lists dependencies. `cargo build` fetches whatever the project declares, and most Arc 2 weeks declare zero external crates.

**W05–W08** use stable Rust, 2021 edition, no nightly features. Recommended IDE: VS Code with the rust-analyzer extension, or RustRover.

**New to Rust?** This is the real ramp in the curriculum — Go doesn't prepare you for the borrow checker any more than any other language would, since ownership has no equivalent outside Rust (and a handful of niche languages). Don't expect W01–W04's Go to have softened this jump; budget it as a fresh investment. Before starting W05, work through [The Rust Book](https://doc.rust-lang.org/book/) (free) Chapters 4 (Ownership), 5 (Structs), 6 (Enums and Pattern Matching), 10 (Traits), and 13 (Closures and Iterators) — those five chapters map almost directly onto what W05–W08 need. Budget 6–8 hours if this is genuinely your first time past `borrow checker` errors — meaningfully longer than the Go ramp, and that asymmetry is real, not a formality.

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
rustc --version      # 1.8x.x or later
cargo --version
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
| Rust | VS Code + rust-analyzer, or RustRover |
| Python | VS Code + Pylance, or PyCharm Community |
| All | Neovim with LSP (if you're into that) |
