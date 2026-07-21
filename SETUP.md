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

**W00–W04, W17, and secondary tooling in W03/W13/W15/W20** use Java: this is the backbone language of the curriculum alongside C++. Each project is its own Maven project (`pom.xml` per directory, no shared parent build file, same isolation principle as the C++/Scala projects); `mvn compile`/`mvn test`/`mvn package` fetch whatever the project's `pom.xml` declares.

**Already know Java?** If your Java is production-grade (per Gaston's own background, "advanced" and already used for Map/Reduce-style big-data work), this is close to a zero-ramp module: the only genuinely new surface is Java 21 itself, not the language you already know. Skim before W01: `record` types for immutable data (they auto-generate `equals`/`hashCode`/`toString`, but field-by-field, which matters for array-typed fields, see W01's callout), virtual threads (`Thread.ofVirtual()`, `Executors.newVirtualThreadPerTaskExecutor()`, cheap enough to use one per task instead of pooling), and `sealed` interfaces with exhaustive pattern-matching `switch` (W17 uses this directly). [What's New in Java 21](https://openjdk.org/projects/jdk/21/) (official release notes) covers all three in about 20 minutes. No Spring, no Kafka anywhere in this curriculum, by design: every HTTP service uses the JDK's own `com.sun.net.httpserver.HttpServer`, deliberately avoiding framework overhead for services this small.

**Rusty, or newer to Java?** Budget more time before W01: [Java Records](https://docs.oracle.com/en/java/javase/21/language/records.html) and [Virtual Threads](https://docs.oracle.com/en/java/javase/21/core/virtual-threads.html) (official Oracle guides, ~30 min each) cover the two idioms this curriculum leans on hardest. The rest, generics, the Streams API, collections, is unlikely to have changed much from whatever Java you last wrote.

---

## Kubernetes Operators: KubeRay + Spark Operator (W19)

No language toolchain to install for this one, no Go, no SDK. W19 has you deploy two real operators to the kind cluster via Helm (installed below in Docker + kind) and read their source on GitHub; nothing here gets compiled locally. Register both chart repos ahead of time so W19 itself is just `helm install`:
```bash
helm repo add kuberay https://ray-project.github.io/kuberay-helm/
helm repo add spark-operator https://kubeflow.github.io/spark-operator
helm repo update
```

**Never touched Go?** That's fine, this curriculum never asks you to write any. W19 links you to specific files inside `ray-project/kuberay` and `kubeflow/spark-operator`; GitHub's web view is enough to read them. If you'd rather browse and grep the source locally instead, `brew install go` gets you a working `go doc`/`gofmt`-aware setup, but it's entirely optional, install it only if reading on GitHub feels limiting.

---

## C++ (C++17/20, CMake)

```bash
# macOS
xcode-select --install        # ships clang
brew install cmake ninja vcpkg

cmake --version                # 3.25+ recommended
clang++ --version               # or g++ --version if you prefer GCC

# vcpkg (C++ package manager, used for prometheus-cpp / opentelemetry-cpp / GoogleTest in W05–W08, W20)
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
Most Arc 2 weeks (W05–W07) declare zero or one dependency (GoogleTest); W20's Prometheus/OTel instrumentation is the one place this arc pulls in real external libraries.

**W05–W08** target C++17/20 (structured bindings and `std::variant` from C++17; concepts, the one addition specific to C++20, where useful). This is modern enough to be relevant to the codebases W06 and W08 point you at, PyTorch's autograd engine and DuckDB's execution engine, both of which build against C++17/20 themselves. No nightly/experimental compiler flags needed; mainline `clang` or `gcc` from Homebrew is current enough.

**Already know C++?** If your C++ is from school or an earlier job, most of the syntax will come back fast, but treat this as a refresh into *modern* idioms rather than a cold start. The gap is usually smart pointers (`std::unique_ptr`/`std::shared_ptr` instead of raw `new`/`delete`), move semantics and RAII (resource cleanup tied to scope, C++'s core answer to manual `new`/`delete` bookkeeping), and STL algorithms/containers (`std::vector`, `std::unordered_map`, `<algorithm>`) instead of hand-rolled arrays and loops. Before starting W05, read [A Tour of C++](https://www.stroustrup.com/tour3.html) (Stroustrup, free chapter previews / short book) Chapters 1 (Basics), 4 (Classes), 5 (Essential Operations, where move semantics and RAII live), and 8 (Templates); pair it with cppreference's pages on [smart pointers](https://en.cppreference.com/w/cpp/memory) and [move semantics](https://en.cppreference.com/w/cpp/language/move_constructor). Budget 4–6 hours: less than a from-scratch language, since the syntax and control flow are already familiar, but real time nonetheless for the idioms that changed since you last wrote C++.

**One thing that has no compiler safety net:** C++ won't stop you from writing something that compiles but is wrong; nothing catches a dangling reference or a data race for you at compile time. The weeks' "Constraints" sections call out, explicitly, where you're responsible for a discipline (encapsulation, immutability of returned collections) that no compiler enforces for you here. Read those callouts; they're not boilerplate.

**CMake itself is a real ramp, separate from the language.** There's no zero-config "it just builds" experience here; expect `CMakeLists.txt` boilerplate and `find_package`/vcpkg wiring to cost real time in W05, even though the C++ language itself is familiar. If a project won't configure, check the vcpkg toolchain file path before anything else; it's the most common first-week failure.

---

## Scala 2.13 (sbt)

```bash
# macOS
brew install coursier/formulas/coursier && cs setup
# cs setup installs a JDK if needed, plus sbt and scalac

sbt --version    # sbt 1.9.x or later
```

Each Scala project (`code/query-planner/`, `code/agg-algebra/`) is its own sbt project: `build.sbt` pins `scalaVersion := "2.13.14"` (or the latest 2.13.x patch), so the project's Scala version is fixed regardless of whatever `cs setup` installed as your global default. `sbt compile`/`sbt test`/`sbt run` fetch whatever `build.sbt` declares, a zero-config experience unlike CMake's `find_package`/vcpkg wiring for C++.

**W09–W10 target Scala 2.13, not 3.** This is a deliberate match, not an oversight. 2.13 is what Apache Spark itself is built and published against today (Spark 4.x still compiles Catalyst and the rest of the codebase on 2.13; there is no Spark-on-Scala-3 build), and it's what Algebird (W10's real-world reference) publishes for. Writing these two weeks in 2.13 means the case classes, pattern matching, and `implicit`-based typeclasses you're using are exactly what you'd see reading real Catalyst or Algebird source, not a newer dialect neither project has adopted.

**Already know Scala from Spark?** This is the lowest-ramp module in the curriculum, and now a near-zero one: 2.13 is almost certainly the exact Scala version you already write in production Spark jobs, so there's no syntax delta to review at all. Case classes, pattern matching, and `implicit` typeclass instances are patterns you already use, just without naming the underlying algebra ("this is a semigroup," "this rewrite rule is a partial function over the plan tree") explicitly. Budget closer to zero prep; go straight to W09. If your Scala is genuinely rusty or this is a first real exposure, [Scala Book](https://docs.scala-lang.org/overviews/scala-book/introduction.html) (scala-lang.org's official 2.13-era guide) chapters on classes, traits, and implicits cover what these two weeks need; budget 2–3 hours instead.

Recommended IDE: IntelliJ IDEA with the Scala plugin, the standard choice for Spark/Scala work and likely already familiar if you've done production Scala.

**New to Scala, or want a warm-up before W09 specifically?** See the drill in [W09](weeks/W09-query-planning.md)'s "Before you start" section: a 15–20 minute case-class-and-pattern-matching exercise scoped to exactly what that week needs, meant to be done the day you start W09, not months ahead of it.

**W12's Scala project is different from W09/W10's:** those two are dependency-free toy projects; W12's `scala/build.sbt` pulls in real Spark (`libraryDependencies += "org.apache.spark" %% "spark-sql" % "3.5.1"`), so the first `sbt compile` there downloads Spark's full dependency tree and will take noticeably longer than anything in W09/W10. Any JDK 8/11/17 works (the same one `cs setup` installed, or the `openjdk@17` installed for W07 below both satisfy Spark 3.5.x). Pin the identical version string in `python/requirements.txt` (`pyspark==3.5.1`); the whole point of the week is comparing two runtimes on the same Spark release, not two different releases.

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

**Arc 3 Python base (W11, W13–W16):**
```bash
pip install numpy torch torchvision duckdb pyarrow pandas "ray[default]"
```

**W11 (ML pipelines):**
```bash
pip install duckdb pyarrow pandas
```

**W12 (PySpark vs. Scala Spark):**
```bash
pip install pyspark==3.5.1        # match the exact version in scala/build.sbt
```
Same Spark install as the ClickHouse + PySpark section below; if you already set that up for W07, you only need to confirm the version matches what W12's `build.sbt` pins, not reinstall from scratch.

**W13 (distributed training):**
```bash
pip install numpy torch           # torch for MNIST loading only
```

**W14 (actor model / Ray):**
```bash
pip install "ray[default]" torch  # torch for the CNN, Ray for actors
```

**W15 (GPU compute), requires NVIDIA GPU for the CUDA path:**
```bash
pip install numba numpy matplotlib
```
No GPU? Skip this install; use the C fallback (`code/cpu-gemm/`) instead: cache-blocked + AVX2 SIMD GEMM, no Python dependencies beyond a working gcc/clang.

**W16 (attention):**
```bash
pip install numpy                 # NumPy only, no PyTorch for this week
```

---

## ClickHouse + PySpark (W07, PySpark reused in W12)

W07's Part 2 comparison exercise runs two real local systems alongside your C++ build; both single-machine, no account, no cluster. The PySpark install here is reused for W12; just confirm the version matches what W12's `scala/build.sbt` pins (see the Python section above), reinstalling with `pip install pyspark==<version>` if it doesn't.

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

## GPU Setup (W15, optional)

**NVIDIA GPU required.** If you don't have one, skip the Numba CUDA path and use the C fallback.

1. Install [CUDA Toolkit 12.x](https://developer.nvidia.com/cuda-downloads)
2. Verify: `nvcc --version`
3. Install Numba: `pip install numba`
4. Test: `python -c "from numba import cuda; print(cuda.gpus)"`

---

## Verify Everything

```bash
java --version       # openjdk 21.x
mvn --version        # Apache Maven 3.9.x
cmake --version      # 3.25.x or later
clang++ --version    # or g++ --version
sbt --version        # 1.9.x or later, builds Scala 2.13 per-project via build.sbt
python --version     # 3.11.x or 3.12.x
docker --version     # 25.x or later
kind --version       # 0.22.x or later
kubectl version      # 1.29.x or later
helm version         # 3.14.x or later, needed for W19's KubeRay + Spark Operator installs
```
`go version` is only relevant if you chose to install Go for local source browsing (see the Kubernetes Operators section above); it isn't part of the required toolchain.

---

## Recommended IDEs

| Language | IDE |
|----------|-----|
| Java | IntelliJ IDEA (Community is fine), or VS Code + Extension Pack for Java |
| Go (optional, source-reading only) | GitHub's web view is enough; VS Code + Go extension if you installed Go locally |
| C++ | VS Code + clangd extension, or CLion |
| Scala | IntelliJ IDEA + Scala plugin |
| Python | VS Code + Pylance, or PyCharm Community |
| All | Neovim with LSP (if you're into that) |
