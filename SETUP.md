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

## Java 21

**macOS (Homebrew):**
```bash
brew install openjdk@21
# Link it so the system sees it:
sudo ln -sfn $(brew --prefix)/opt/openjdk@21/libexec/openjdk.jdk /Library/Java/JavaVirtualMachines/openjdk-21.jdk
java --version   # should print "21.x.x"
```

**Ubuntu / Debian:**
```bash
sudo apt update
sudo apt install openjdk-21-jdk
java --version   # should print "21.x.x"
```

If you have multiple JDKs installed and need to switch:
```bash
# macOS
export JAVA_HOME=$(brew --prefix)/opt/openjdk@21
# Ubuntu
sudo update-alternatives --config java   # pick java-21
```

**W01–W04, W14** use Java 21 with virtual threads (`Thread.ofVirtual()`) and records. Both require Java 21+.

---

## Scala 2.13 + SBT

```bash
# macOS
brew install sbt
sbt --version   # should print sbt script version

# Ubuntu / Debian
echo "deb https://repo.scala-sbt.org/scalasbt/debian all main" | sudo tee /etc/apt/sources.list.d/sbt.list
curl -sL "https://keyserver.ubuntu.com/pks/lookup?op=get&search=0x2EE0EA64E40A89B84B2DF73499E82A75642AC823" | sudo apt-key add -
sudo apt update && sudo apt install sbt
```

Each Scala project (`code/streaming/`, `code/dd-scratch/`, etc.) has its own `build.sbt`. SBT downloads Scala 2.13 on first run.

Minimal `build.sbt` for any Arc 2 project:
```scala
scalaVersion := "2.13.16"
scalacOptions ++= Seq("-deprecation", "-feature", "-language:higherKinds")
```

**W05–W08** use Scala 2.13.16, chosen for compatibility with Spark, Flink, and the broader JVM ecosystem. Recommended IDE: IntelliJ IDEA with the Scala plugin, or VS Code with Metals.

**Scala 2 vs Scala 3 syntax differences** you'll encounter in examples online:

| Concept | Scala 2.13 (use this) | Scala 3 equivalent |
|---------|----------------------|-------------------|
| ADTs | `sealed trait` + `case class` | `enum` |
| Type classes | `implicit val` / `implicit def` | `given` / `using` |
| Braces | always required | optional |

The week files describe *what* to build, not exact syntax. Either version works conceptually.

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

## Go 1.22+

```bash
# macOS
brew install go
go version   # go1.22.x

# Or download from https://go.dev/dl/
```

Go appears as a secondary/tooling language in W03, W10, W12, W14, W15, W16, W17. You don't need Go to complete any arc; it's always optional or a stretch goal except W16.

**W16 (Kubernetes Operators) requires Go.** Install `controller-runtime`:
```bash
go get sigs.k8s.io/controller-runtime@v0.18.0
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
java --version       # 21.x
sbt --version        # 1.9.x or later
python --version     # 3.11.x or 3.12.x
go version           # 1.22.x
docker --version     # 25.x or later
kind --version       # 0.22.x or later
kubectl version      # 1.29.x or later
helm version         # 3.14.x or later
```

---

## Recommended IDEs

| Language | IDE |
|----------|-----|
| Java | IntelliJ IDEA Community (free) |
| Scala | IntelliJ IDEA + Scala plugin, or VS Code + Metals |
| Python | VS Code + Pylance, or PyCharm Community |
| Go | VS Code + Go extension, or GoLand |
| All | Neovim with LSP (if you're into that) |
