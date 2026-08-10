# Environment Setup

Everything you need installed before starting W00. Set this up once; it covers the entire curriculum.

**Version pinning.** The JVM and Python stack is pinned to Databricks Runtime 18.0: Apache Spark 4.1.0, Python 3.12.3, Java 21. DBR is a publicly documented assembly of otherwise independent open-source versions, so aligning to it means what you build locally is version-compatible with a real production runtime. Everything you install is upstream Apache Spark and upstream Python. Go, Flink, and the Kubernetes tooling have no equivalent reference point and are pinned to current upstream releases.

---

## Obsidian

1. Download from [obsidian.md](https://obsidian.md) and install
2. Open vault: **File → Open folder as vault** → select this repo directory
3. Install community plugins (Settings → Community plugins → Browse):
   - **Dataview**: required, powers the Home.md dashboard (enable JavaScript queries)
   - **Tasks**: optional, used for task filtering
   - **Obsidian Git**: optional, auto-syncs vault to GitHub
4. Set `Home.md` as startup note: Settings → General → Default startup page → `Home`

---

## Go 1.26+

Used in W00, W01, W03, and secondary tooling in W09 and W15.

```bash
# macOS
brew install go

go version    # go1.26.x or later
```

Each project is its own module (`go.mod` per directory, no shared workspace file). `go build`, `go test`, and `go run` fetch whatever the project's `go.mod` declares. There is no separate package manager or lockfile beyond `go.mod` and `go.sum`, both maintained by `go get` and `go mod tidy`.

**Before W00, if you are new to Go.** [A Tour of Go](https://go.dev/tour/), about an hour: work through "Basics" (variables, functions, structs, slices, maps) and "Methods and interfaces" through the goroutines and channels section at the end. That covers everything W00 through W03 need. Two idioms carry most of the weight: goroutines plus channels for concurrency (`go func() { ... }()`, `make(chan T)`, `sync.WaitGroup`), and the standard library's `net/http` for every HTTP service in the curriculum. No framework.

**Failure mode to know about.** Nothing stops you from mutating a `map` or slice that another goroutine also holds a reference to, and the compiler will not warn you. The "Constraints" section in each week states where a function must return a fresh copy rather than mutate in place. Run `go test -race` and `go run -race` to catch unsynchronized concurrent access at runtime; W03's tests are the ones where it matters most.

---

## Java 21

Used in W04, W05 Part 1, W13, and W15.

```bash
# macOS
brew install openjdk@21 maven

# Point the shell at it (Homebrew doesn't symlink a versioned JDK onto PATH by default)
echo 'export PATH="/opt/homebrew/opt/openjdk@21/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

java --version    # openjdk 21.x
mvn --version     # Apache Maven 3.9.x, and confirms it's picking up JDK 21
```

Each project is its own Maven project (`pom.xml` per directory, no shared parent build file). `mvn compile`, `mvn test`, and `mvn package` fetch whatever the project's `pom.xml` declares.

**Before W04, if your Java is current.** The only new surface is Java 21 itself. Skim [What's New in Java 21](https://openjdk.org/projects/jdk/21/), about 15 minutes, for two features used directly in W04 and W13: `record` types (they auto-generate `equals`/`hashCode`/`toString` field by field, which matters for array-typed fields), and `sealed` interfaces with exhaustive pattern-matching `switch` plus record patterns (JEP 440).

**Before W04, if you are rusty or newer to Java.** Budget more time: [Java Records](https://docs.oracle.com/en/java/javase/21/language/records.html) and [Pattern Matching for switch](https://docs.oracle.com/en/java/javase/21/language/pattern-matching-switch-statements-and-expressions.html), about 30 minutes each. Generics and collections are unchanged from whatever Java you last wrote.

No Spring and no Kafka anywhere in this curriculum.

---

## Apache Spark 4.1.0 (PySpark)

Used in W02, W05 Part 2, W07, and W14. Local mode throughout: no cluster, no account.

```bash
pip install "pyspark==4.1.0"
```

Spark runs on the JVM regardless of the Python surface, and the Java 21 you installed above satisfies it. There is no second runtime to manage.

**If you hit `InaccessibleObjectException`:** Spark is reaching into internal JDK APIs the module system closed off. Set `--add-opens` via `spark.driver.extraJavaOptions` in your `SparkSession` builder rather than chasing them one stack trace at a time.

**The Spark UI is the tool in three of those four units.** It serves at `localhost:4040` while a driver is alive, so put a `System.in.read()` at the end of `main` when you want to look around after a job finishes.

---

## Apache Flink 2.3.0

Used in W04 Part 1. Nothing to install: Flink's `MiniCluster` runs inside your JVM when you execute a `StreamExecutionEnvironment` job from an IDE or `mvn exec`. Add the dependency to that unit's `pom.xml`:

```xml
<dependency>
  <groupId>org.apache.flink</groupId>
  <artifactId>flink-streaming-java</artifactId>
  <version>2.3.0</version>
</dependency>
```

**Two version notes.** Flink 2.x recommends Java 17 and classifies Java 21 support as beta. Stay on 21; a single local job will not go near the edges that beta status refers to. Separately, Flink 2.0 removed `org.apache.flink.streaming.api.windowing.time.Time` in favour of `java.time.Duration`, so any tutorial showing `Time.seconds(10)` was written for 1.x and will not compile.

---

## Python 3.12

Use [pyenv](https://github.com/pyenv/pyenv) to manage Python versions.

```bash
# macOS
brew install pyenv
pyenv install 3.12.3
pyenv global 3.12.3
python --version   # 3.12.3
```

Install dependencies per arc:

**Arc 3 base (W08–W12):**
```bash
pip install numpy torch torchvision duckdb pyarrow pandas "ray[default]"
```

**W08, ML pipelines:**
```bash
pip install duckdb pyarrow pandas deltalake
```
`deltalake` is delta-rs, a native implementation with a Python binding. Part 2 uses it to open a real Delta transaction log. No JVM, no Spark cluster, no extra setup.

**W09, distributed training:**
```bash
pip install numpy torch           # torch for MNIST loading only
```

**W10, beyond data parallelism:**
```bash
pip install numpy                 # no new dependencies; imports W09's ring_allreduce directly
```

**W11, actor model and Ray:**
```bash
pip install "ray[default]" torch  # torch for the CNN, Ray for actors
```

**W12, attention:**
```bash
pip install numpy                 # NumPy only, no PyTorch this week
```

**W16, optional capstone:**
```bash
pip install mlflow
```
`mlflow server` runs from this install, though W16 deploys it into kind rather than running it on the host.

---

## DuckDB

Used in W06 and W08. Python is DuckDB's primary surface and the one both units use.

```bash
pip install duckdb numpy pyarrow
```

---

## Docker + kind

Used in W00, W14, and W15.

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

## Kubernetes operators (W14)

W14 uses Kubeflow Trainer and the Spark Operator. Both are installs plus reading, not writing, operator source. Register the Spark Operator chart repo ahead of time:

```bash
helm repo add spark-operator https://kubeflow.github.io/spark-operator
helm repo update
```

**Do not pre-install Kubeflow Trainer.** It installs from versioned manifests rather than a stable chart repo, and the manifest paths move between releases. When you reach W14, check the releases page for the current version and follow the project's own installation guide. This is the most common way that unit goes wrong.

---

## Verify Everything

```bash
go version           # go1.26.x or later
java --version       # openjdk 21.x
mvn --version        # Apache Maven 3.9.x
python --version     # 3.12.3
docker --version     # any current release
kind --version       # any release new enough to create a Kubernetes 1.36 cluster
kubectl version      # 1.36.x (1.33 and earlier are end-of-life)
helm version         # 4.x, or 3.21.x if you prefer the 3 line, which is still maintained
```

The Kubernetes tools are not pinned to exact versions because kind, kubectl, and Helm move faster than this document. The constraint that matters is that all three can talk to the same Kubernetes version: target the current stable minor release and take whatever kind, kubectl, and Helm releases support it. The Spark Operator chart W14 installs works on Helm 4.x or 3.x, so use whichever you have.

---

## Recommended IDEs

| Language | IDE |
|----------|-----|
| Go | VS Code + the official Go extension, or GoLand |
| Java | IntelliJ IDEA (Community is fine), or VS Code + Extension Pack for Java |
| Python | VS Code + Pylance, or PyCharm Community |
| All | Neovim with LSP (if you're into that) |
