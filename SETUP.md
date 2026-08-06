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

## Why these versions

Where this curriculum touches the JVM data stack, the versions are pinned to match **Databricks Runtime 18.0**: Apache Spark 4.1.0, Python 3.12.3, and Java 21. That is not vendor allegiance, it is a free calibration point. DBR is a widely deployed, publicly documented assembly of otherwise independent open-source versions, so aligning to it means what you build locally is version-compatible with a real production runtime rather than with nothing in particular. Everything you depend on is upstream Apache Spark and upstream Python; the runtime is only the reference that says which combination people actually run together.

Go, Flink, and the Kubernetes tooling have no such reference and are pinned to current upstream releases instead.

---

## Go 1.26+ (modules)

```bash
# macOS
brew install go

go version    # go1.26.x or later
```

**W00, W01, W03, and secondary tooling in W09/W15** use Go, this curriculum's one deliberately introduced new language. Each project is its own module (`go.mod` per directory, no shared workspace file, same isolation principle as the Java projects); `go build`/`go test`/`go run` fetch whatever the project's `go.mod` declares. There's no separate package manager or lockfile format to learn beyond `go.mod`/`go.sum`, both maintained automatically by `go get` and `go mod tidy`.

**New to Go?** Budget real time before W00, this is the curriculum's genuinely new component, kept deliberately gentle in scope. [A Tour of Go](https://go.dev/tour/) (~1 hour): work through "Basics" (variables, functions, structs, slices, maps) and "Methods and interfaces" (through the goroutines/channels section at the end); that covers everything W00 through W03 need. The two idioms this curriculum leans on hardest: goroutines plus channels for concurrency (`go func() { ... }()`, `make(chan T)`, `sync.WaitGroup`), and the standard library's own `net/http` for every small HTTP service, no framework, ever, in this curriculum.

**Why Go's footprint is three units rather than five.** W02 and W06 used to be Go builds. Both now measure a real engine (Spark and DuckDB respectively) instead of reimplementing one, which took their builds with them. Go still carries W00's service, W01's write-path benchmark, and W03's concurrency work, which is enough that W14's reconciler reading is not a cold start.

**One thing with no compiler safety net:** Go's garbage collector means you don't have manual-memory failure modes, but nothing stops you from mutating a `map` or slice that another goroutine also holds a reference to. The weeks' "Constraints" sections call out, explicitly, where a function is expected to return a fresh copy rather than mutate in place, a discipline you're responsible for keeping, not one the compiler enforces. Read those callouts; they're not boilerplate. Go's own race detector (`go test -race`, `go run -race`) catches unsynchronized concurrent access at runtime, worth running against W03's tests specifically.

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

**W02 and W04 through W07, plus W13 and W15** use Java. That is the whole of Arc 2 plus fault tolerance and instrumentation, and it splits into two kinds of work. Where you build something, Java is chosen because its sealed interfaces and record patterns give a real, compiler-enforced advantage (W04's `Policy`, W05's `Partitioner`, W13's `Message`) rather than by default. Where you drive an engine (Spark in W02, W05 Part 2, and W07; Flink in W04 Part 1; DuckDB over JDBC in W06), Java is simply the one driver language for all of them, which keeps the arc on a single stack.

Each project is its own Maven project (`pom.xml` per directory, no shared parent build file, same isolation principle as the Go projects); `mvn compile`/`mvn test`/`mvn package` fetch whatever the project's `pom.xml` declares.

**A note on the `_2.13` suffix you will see in Spark artifact names.** Spark is written in Scala and publishes per-Scala-version artifacts, so its Java-facing JARs are still named `spark-sql_2.13`. You are not writing Scala anywhere in this curriculum, and you do not need Scala or sbt installed. The suffix names the language Spark itself was built in, nothing more.

**Spark on Java 21 needs `--add-opens`.** Spark reaches into internal JDK APIs that the module system closed off. Add the flags to your Maven `exec` configuration or run configuration up front; the alternative is discovering them one `InaccessibleObjectException` at a time.

**Already know Java?** If your Java is production-grade (per Gaston's own background, "advanced" and already used for Map/Reduce-style big-data work), this is close to a zero-ramp module: the only genuinely new surface is Java 21 itself, not the language you already know. Skim before W04: `record` types for immutable data (they auto-generate `equals`/`hashCode`/`toString`, but field-by-field, which matters for array-typed fields), and `sealed` interfaces with exhaustive pattern-matching `switch` plus record patterns (JEP 440), both used directly in W04 and W13. [What's New in Java 21](https://openjdk.org/projects/jdk/21/) (official release notes) covers both in about 15 minutes. No Spring, no Kafka anywhere in this curriculum, by design.

**Rusty, or newer to Java?** Budget more time before W04: [Java Records](https://docs.oracle.com/en/java/javase/21/language/records.html) and [Pattern Matching for switch](https://docs.oracle.com/en/java/javase/21/language/pattern-matching-switch-statements-and-expressions.html) (official Oracle guides, ~30 min each) cover the two idioms this curriculum leans on hardest. The rest, generics, collections, is unlikely to have changed much from whatever Java you last wrote.

---

## Kubernetes Operators: Kubeflow Trainer + Spark Operator (W14)

No new toolchain to install for this one specifically, you already have Go from the section above, and W14 itself is installs plus reading, not writing, operator source. Register the Spark Operator chart repo ahead of time:
```bash
helm repo add spark-operator https://kubeflow.github.io/spark-operator
helm repo update
```

Kubeflow Trainer installs from versioned manifests rather than a stable chart repo, and the manifest paths move between releases. Don't pre-install it from a version pinned here; when you reach W14, check the releases page for the current version and follow the project's own installation guide. The unit says so too, and this is the most common way that part goes wrong.

By the time you reach W14 you'll have written Go in three other weeks (W00 through W03), so reading `kubeflow/trainer`'s and `kubeflow/spark-operator`'s real reconciler source here is no longer a cold start, it's the payoff for the Go you've already been writing, applied to two real, production codebases instead of a toy exercise.

---

## Apache Spark 4.1.0 (via Maven, W02, W05, W07, W14)

Nothing to install. Spark is a Maven dependency in the projects that use it, and it runs in local mode inside your own JVM.

```xml
<!-- in each Spark unit's pom.xml -->
<dependency>
  <groupId>org.apache.spark</groupId>
  <artifactId>spark-sql_2.13</artifactId>
  <version>4.1.0</version>
</dependency>
```

Four units drive Spark, each asking it a different question: W02 reads a stage DAG to see materialization cost, W05 Part 2 reads the task duration distribution to find a straggler, W07 reads `EXPLAIN` output to catch a join strategy changing, and W14 packages a job into an image and submits it to the Spark Operator. Local mode throughout, no cluster, no account.

**The Spark UI is the actual tool** in three of those four. It serves at `localhost:4040` while a driver is alive, so put a `System.in.read()` at the end of `main` when you want to look around after a job finishes.

---

## Apache Flink 2.3.0 (via Maven, W04 Part 1)

Also nothing to install. Flink's `MiniCluster` runs inside your JVM when you execute a `StreamExecutionEnvironment` job from an IDE or `mvn exec`, which is all W04 needs.

```xml
<dependency>
  <groupId>org.apache.flink</groupId>
  <artifactId>flink-streaming-java</artifactId>
  <version>2.3.0</version>
</dependency>
```

**Two version notes worth knowing before you start.** Flink 2.x recommends Java 17 and classifies Java 21 support as beta. Stay on 21: the rest of this curriculum's JVM stack is pinned there to match DBR 18, and a single local job will not go near the edges that beta status refers to. Separately, Flink 2.0 removed the old `org.apache.flink.streaming.api.windowing.time.Time` class in favour of `java.time.Duration`, so any tutorial showing `Time.seconds(10)` was written for 1.x and will not compile.


---

## Python 3.12 (matching DBR 18)

Use [pyenv](https://github.com/pyenv/pyenv) to manage Python versions.

```bash
# macOS
brew install pyenv
pyenv install 3.12.3
pyenv global 3.12.3
python --version   # 3.12.3, matching DBR 18
```

Install dependencies per arc:

**Arc 3 Python base (W08–W12):**
```bash
pip install numpy torch torchvision duckdb pyarrow pandas "ray[default]"
```

**W08 (ML pipelines):**
```bash
pip install duckdb pyarrow pandas deltalake
```
`deltalake` is delta-rs, a native implementation with a Python binding. Part 2 of that week uses it to open a real Delta transaction log; there's no JVM and no Spark cluster involved, so this is a plain `pip install` with no extra setup.

**W09 (distributed training):**
```bash
pip install numpy torch           # torch for MNIST loading only
```

**W10 (beyond data parallelism):**
```bash
pip install numpy                 # no new dependencies; imports W09's ring_allreduce directly
```

**W16 (optional capstone):**
```bash
pip install mlflow
```
Only needed if you do the optional capstone. `mlflow server` runs from this same install; W16 deploys it into kind rather than running it on the host.

**W11 (actor model / Ray):**
```bash
pip install "ray[default]" torch  # torch for the CNN, Ray for actors
```

**W12 (attention):**
```bash
pip install numpy                 # NumPy only, no PyTorch for this week
```

---

## No PySpark

Spark appears in four units and all of them drive it through the Java API and Maven, so there is nothing to `pip install`. If you already have PySpark on your machine from other work it will not conflict; the units simply do not use it.

---

## Docker + kind (W00, W14, W15)

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

## Verify Everything

```bash
go version           # go1.26.x or later
java --version       # openjdk 21.x, which is what DBR 18 runs (Zulu 21)
mvn --version        # Apache Maven 3.9.x
python --version     # 3.12.3
docker --version     # any current release
kind --version       # any release new enough to create a Kubernetes 1.36 cluster
kubectl version      # 1.36.x (1.33 and earlier are end-of-life)
helm version         # 4.x, or 3.21.x if you prefer the 3 line, which is still maintained
```

**Why the Kubernetes tools aren't pinned to exact versions above:** kind, kubectl, and Helm move faster than this document will be updated, and a stale pinned number is worse than no number because it looks authoritative. The constraint that actually matters is that they can all talk to the same Kubernetes version. Target the current stable Kubernetes minor release, and take whatever kind, kubectl, and Helm releases support it. Helm 4 is the current major line; the Spark Operator chart W14 installs works on either 4.x or 3.x, so use whichever you already have.

---

## Recommended IDEs

| Language | IDE |
|----------|-----|
| Go | VS Code + the official Go extension, or GoLand |
| Java | IntelliJ IDEA (Community is fine), or VS Code + Extension Pack for Java |
| Python | VS Code + Pylance, or PyCharm Community |
| All | Neovim with LSP (if you're into that) |
