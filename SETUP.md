# Setup

## Start here

1. Open this directory as a local vault in Obsidian. Enable Dataview and its JavaScript queries for the [Home dashboard](Home.md). The [README study order](README.md) also works as a plain Markdown index.
2. Install JDK 21 and Maven. Check `java -version` and `mvn -version`. Configure the project compiler release to 21 and pin the build and test dependencies. From `code/query-engine/`, use `mvn test`. Set up the Python and SQL comparisons independently of the Java build.
3. Set a first-study-week date in [config](config.md) when you want the calendar view. The date is independent of task completion.
4. Open the [setup unit](weeks/W00-setup.md) and create the fixture and test project.

## Install when the unit needs it

- **Batch and planning.** In a Python virtual environment, install a PySpark release and record its version. Check its [installation requirements](https://spark.apache.org/docs/latest/api/python/getting_started/install.html) for the compatible Python and Java versions. Run local mode and confirm the Spark UI is reachable while the process runs.
- **Query execution and tables.** Add DuckDB and PyArrow to the relevant experiment environment. For the table lab use the [Python Delta Lake API](https://delta-io.github.io/delta-rs/), and record the version alongside your fixture.
- **Deployment.** Install a container runtime, [kind](https://kind.sigs.k8s.io/) and kubectl. First run the coordinator and workers locally. Use kind once you can rerun the query and capture metrics.
- **ML extension.** Use a separate environment and record the Python, NumPy, PyTorch and Ray versions needed by your exercises.

Check in small deterministic fixtures and keep larger generated data out of Git.
