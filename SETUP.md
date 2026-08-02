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

## Go 1.22+ (modules)

```bash
# macOS
brew install go

go version    # go1.22.x or later
```

**W00–W04, W08, and secondary tooling in W03/W12/W15/W20** use Go, this curriculum's one deliberately introduced new language. Each project is its own module (`go.mod` per directory, no shared workspace file, same isolation principle as the Java/Scala projects); `go build`/`go test`/`go run` fetch whatever the project's `go.mod` declares. There's no separate package manager or lockfile format to learn beyond `go.mod`/`go.sum`, both maintained automatically by `go get` and `go mod tidy`.

**New to Go?** Budget real time before W00, this is the curriculum's genuinely new component, kept deliberately gentle in scope. [A Tour of Go](https://go.dev/tour/) (~1 hour): work through "Basics" (variables, functions, structs, slices, maps) and "Methods and interfaces" (through the goroutines/channels section at the end); that covers everything W00–W04 and W08 need. The two idioms this curriculum leans on hardest: goroutines plus channels for concurrency (`go func() { ... }()`, `make(chan T)`, `sync.WaitGroup`), and the standard library's own `net/http` for every small HTTP service, no framework, ever, in this curriculum. `testing.B` (`go test -bench=.`) is the other one worth knowing before W02 and W08, Go's built-in microbenchmark harness.

**One thing with no compiler safety net:** Go's garbage collector means you don't have C++'s memory-safety failure modes, but nothing stops you from mutating a `map` or slice that another goroutine also holds a reference to. The weeks' "Constraints" sections call out, explicitly, where a function is expected to return a fresh copy rather than mutate in place, a discipline you're responsible for keeping, not one the compiler enforces. Read those callouts; they're not boilerplate. Go's own race detector (`go test -race`, `go run -race`) catches unsynchronized concurrent access at runtime, worth running against W03's and W04's tests specifically.

---

## Java 21 (Maven)

```bash
# macOS
brew install openjdk@21 maven

# Point the shell at it (Homebrew doesn't symlink a versioned JDK onto PATH by default)
echo 'export PATH="/opt/homebrew/opt/openjdk@21/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

java --version    # openjdk 21.x
mvn --version     # Apache Maven 3.9.x, and confirms it's picking up JDK 21
```

**W05–W07, W17** use Java: three weeks in the middle of the streaming/dataflow arc plus fault tolerance, chosen specifically where Java's sealed interfaces and record patterns give a real, compiler-enforced advantage (W05's `StreamItem`, W17's `Message`) rather than by default. Each project is its own Maven project (`pom.xml` per directory, no shared parent build file, same isolation principle as the Go/Scala projects); `mvn compile`/`mvn test`/`mvn package` fetch whatever the project's `pom.xml` declares.

**Already know Java?** If your Java is production-grade (per Gaston's own background, "advanced" and already used for Map/Reduce-style big-data work), this is close to a zero-ramp module: the only genuinely new surface is Java 21 itself, not the language you already know. Skim before W05: `record` types for immutable data (they auto-generate `equals`/`hashCode`/`toString`, but field-by-field, which matters for array-typed fields), and `sealed` interfaces with exhaustive pattern-matching `switch` plus record patterns (JEP 440), both used directly in W05 and W17. [What's New in Java 21](https://openjdk.org/projects/jdk/21/) (official release notes) covers both in about 15 minutes. No Spring, no Kafka anywhere in this curriculum, by design.

**Rusty, or newer to Java?** Budget more time before W05: [Java Records](https://docs.oracle.com/en/java/javase/21/language/records.html) and [Pattern Matching for switch](https://docs.oracle.com/en/java/javase/21/language/pattern-matching-switch-statements-and-expressions.html) (official Oracle guides, ~30 min each) cover the two idioms this curriculum leans on hardest. The rest, generics, collections, is unlikely to have changed much from whatever Java you last wrote.

---

## Kubernetes Operators: Kubeflow Trainer + Spark Operator + Kueue (W19)

No new toolchain to install for this one specifically, you already have Go from the section above, and W19 itself is installs plus reading, not writing, operator source. Register the Spark Operator chart repo ahead of time:
```bash
helm repo add spark-operator https://kubeflow.github.io/spark-operator
helm repo update
```

Kubeflow Trainer and Kueue both install from versioned manifests rather than a stable chart repo, and both move their manifest paths between releases. Don't pre-install them from a version pinned here; when you reach W19, check each project's releases page for the current version and follow its own installation guide. The week says so too, and this is the most common way that part goes wrong.

By the time you reach W19 you'll have written Go in five other weeks (W00–W04, W08), so reading `kubeflow/trainer`'s and `kubeflow/spark-operator`'s real reconciler source here is no longer a cold start, it's the payoff for the Go you've already been writing, applied to two real, production codebases instead of a toy exercise.

---

## Scala 2.13 (sbt)

```bash
# macOS
brew install coursier/formulas/coursier && cs setup
# cs setup installs a JDK if needed, plus sbt and scalac

sbt --version    # sbt 1.9.x or later
```

Each Scala project (`code/query-planner/`, `code/agg-algebra/`) is its own sbt project: `build.sbt` pins `scalaVersion := "2.13.14"` (or the latest 2.13.x patch), so the project's Scala version is fixed regardless of whatever `cs setup` installed as your global default. `sbt compile`/`sbt test`/`sbt run` fetch whatever `build.sbt` declares, a zero-config experience.

**W09–W10 target Scala 2.13, not 3.** This is a deliberate match, not an oversight. 2.13 is what Apache Spark itself is built and published against today (Spark 4.x still compiles Catalyst and the rest of the codebase on 2.13; there is no Spark-on-Scala-3 build), and it's what Algebird (W10's real-world reference) publishes for. Writing these two weeks in 2.13 means the case classes, pattern matching, and `implicit`-based typeclasses you're using are exactly what you'd see reading real Catalyst or Algebird source, not a newer dialect neither project has adopted.

**Already know Scala from Spark?** This is the lowest-ramp module in the curriculum, and now a near-zero one: 2.13 is almost certainly the exact Scala version you already write in production Spark jobs, so there's no syntax delta to review at all. Case classes, pattern matching, and `implicit` typeclass instances are patterns you already use, just without naming the underlying algebra ("this is a semigroup," "this rewrite rule is a partial function over the plan tree") explicitly. Budget closer to zero prep; go straight to W09. Scala's depth in this curriculum is deliberately capped, just enough to reach the distributed-systems concept each week is about, not a vehicle for deep FP mastery; that's intentionally a separate plan. If your Scala is genuinely rusty or this is a first real exposure, [Scala Book](https://docs.scala-lang.org/overviews/scala-book/introduction.html) (scala-lang.org's official 2.13-era guide) chapters on classes, traits, and implicits cover what these two weeks need; budget 2–3 hours instead.

Recommended IDE: IntelliJ IDEA with the Scala plugin, the standard choice for Spark/Scala work and likely already familiar if you've done production Scala.

**New to Scala, or want a warm-up before W09 specifically?** See the drill in [W09](weeks/W09-query-planning.md)'s "Before you start" section: a 15–20 minute case-class-and-pattern-matching exercise scoped to exactly what that week needs, meant to be done the day you start W09, not months ahead of it. W10 has its own equivalent warm-up for the `implicit`-parameter-resolution mechanism specifically.

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

**Arc 3 Python base (W11–W16):**
```bash
pip install numpy torch torchvision duckdb pyarrow pandas "ray[default]"
```

**W11 (ML pipelines):**
```bash
pip install duckdb pyarrow pandas deltalake
```
`deltalake` is delta-rs, a native implementation with a Python binding. Part 2 of that week uses it to open a real Delta transaction log; there's no JVM and no Spark cluster involved, so this is a plain `pip install` with no extra setup.

**W12 (distributed training):**
```bash
pip install numpy torch           # torch for MNIST loading only
```

**W13 (beyond data parallelism):**
```bash
pip install numpy                 # no new dependencies; imports W12's ring_allreduce directly
```

**W14 (actor model / Ray):**
```bash
pip install "ray[default]" torch  # torch for the CNN, Ray for actors
```

**W15 (GPU compute), C is the primary path, no NVIDIA GPU needed:**
```bash
pip install matplotlib   # only for the roofline plot; the kernels themselves are plain C
```
`xcode-select --install` (ships `clang`) is all you need to build `code/cpu-gemm/`; no separate toolchain, no ISA-specific intrinsics to hand-write. If you have NVIDIA GPU access and want the optional CUDA path instead, install `pip install numba numpy matplotlib` (see the GPU Setup section below).

**W16 (attention):**
```bash
pip install numpy                 # NumPy only, no PyTorch for this week
```

---

## ClickHouse + PySpark (W07)

W07's Part 2 comparison exercise runs two real local systems alongside your Java build; both single-machine, no account, no cluster.

```bash
# ClickHouse: single local server, no cluster
brew install clickhouse

# Spark runs on the JVM regardless of the Python surface; install a JDK if you don't have one
brew install openjdk@17

# PySpark, against your existing Python 3.11 setup; no version conflicts
pip install pyspark
```

Verify:
```bash
clickhouse server &        # then, in another terminal:
clickhouse client --query "SELECT 1"

python -c "from pyspark.sql import SparkSession; SparkSession.builder.master('local[*]').getOrCreate(); print('ok')"
```

Neither tool is needed outside W07; uninstall or ignore afterward if you'd rather not keep them around.

---

## Docker + kind (W00, W19, W20)

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

## GPU Setup (W15, optional; the default path is CPU-only C, see the Python section above)

This section only applies if you have NVIDIA GPU access and want to do W15's optional CUDA path instead of the default C one. The default path needs nothing beyond `clang`/`gcc` (via `xcode-select --install`): compile with `-O3 -march=native` and check `clang -Rpass=loop-vectorize` (or `gcc -fopt-info-vec-optimized`) confirms the blocked loop vectorized, no NVIDIA hardware, no intrinsics API, involved.

1. Install [CUDA Toolkit 12.x](https://developer.nvidia.com/cuda-downloads)
2. Verify: `nvcc --version`
3. Install Numba: `pip install numba`
4. Test: `python -c "from numba import cuda; print(cuda.gpus)"`

---

## Verify Everything

```bash
go version           # go1.22.x or later
java --version       # openjdk 21.x
mvn --version        # Apache Maven 3.9.x
sbt --version        # 1.9.x or later, builds Scala 2.13 per-project via build.sbt
python --version     # 3.11.x or 3.12.x
clang --version      # or gcc --version, for W15's C kernels
docker --version     # 25.x or later
kind --version       # 0.22.x or later
kubectl version      # 1.29.x or later
helm version         # 3.14.x or later, needed for W19's Spark Operator install
```

---

## Recommended IDEs

| Language | IDE |
|----------|-----|
| Go | VS Code + the official Go extension, or GoLand |
| Java | IntelliJ IDEA (Community is fine), or VS Code + Extension Pack for Java |
| Scala | IntelliJ IDEA + Scala plugin |
| Python | VS Code + Pylance, or PyCharm Community |
| All | Neovim with LSP (if you're into that) |
