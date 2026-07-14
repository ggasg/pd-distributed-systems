# Code Directory

One subdirectory per week. Each is a self-contained project with its own build file.

```
code/
├── hello-metrics/          # W00: Go service + k8s manifests
│   ├── main.go
│   ├── Dockerfile
│   └── k8s/
│       ├── deployment.yaml
│       └── service-monitor.yaml
│
├── lsm/                    # W01: Go
│   ├── memtable.go
│   ├── sstable.go
│   ├── lsm_tree.go
│   ├── lsm_tree_test.go
│   └── go.mod
│
├── encoding/               # W02: Go
│   ├── varint.go
│   ├── row_store.go
│   ├── column_store.go
│   ├── benchmark.go        # or cmd/benchmark/main.go
│   └── go.mod
│
├── mapreduce/              # W03: Go
│   ├── mapreduce.go        # Mapper/Reducer interfaces
│   ├── runner.go
│   ├── word_count.go
│   ├── pagerank.go
│   └── go.mod
│   # HTTP coordinator lives in tools/job_coordinator/
│
├── clocks/                 # W04: Go
│   ├── vector_clock.go
│   ├── message.go
│   ├── node.go
│   ├── causal_delivery_test.go
│   └── go.mod
│
├── streaming/               # W05: C++ (CMake)
│   ├── include/streaming/
│   │   ├── event.hpp
│   │   ├── watermark.hpp
│   │   ├── aggregator.hpp      # TumblingWindowAggregator
│   │   └── processor.hpp       # StreamProcessor + StreamItem = std::variant<Event, Watermark>
│   ├── src/
│   │   ├── aggregator.cpp
│   │   └── processor.cpp
│   ├── tests/
│   │   └── processor_test.cpp  # GoogleTest
│   └── CMakeLists.txt
│
├── timely-toy/             # W06: C++ (CMake)
│   ├── include/timely_toy/
│   │   ├── timestamp.hpp
│   │   ├── pointstamp.hpp
│   │   ├── operator.hpp        # Operator base class, MapOperator, SinkOperator
│   │   └── progress_tracker.hpp
│   ├── src/
│   │   └── progress_tracker.cpp
│   ├── tests/
│   │   └── progress_tracker_test.cpp   # GoogleTest
│   └── CMakeLists.txt
│
├── dd-scratch/             # W07: C++ (CMake, header-only where templated)
│   ├── include/dd_scratch/
│   │   ├── update.hpp                 # template struct Update<K, V>  (Part 1)
│   │   ├── collection.hpp             # template class Collection<K, V>  (Part 1)
│   │   ├── full_recompute_view.hpp    # FullRecomputeView  (Part 2)
│   │   └── materialized_view.hpp      # IncrementalAggregateView  (Part 2)
│   ├── src/
│   │   ├── word_count.cpp             # Part 1
│   │   ├── full_recompute_view.cpp    # Part 2
│   │   └── materialized_view.cpp      # Part 2
│   ├── benchmark/
│   │   └── mv_benchmark.cpp           # Part 2: incremental vs. full-recompute latency, Release build
│   ├── comparisons/                   # Part 2: same orders/region-revenue model, tested against real local OSS systems
│   │   ├── clickhouse_mv.sql          # local ClickHouse server, real materialized view
│   │   └── spark_stateful_agg.py      # local-mode Spark Structured Streaming, stateful aggregation
│   └── CMakeLists.txt
│
├── query-exec/             # W08: C++ (CMake)
│   ├── include/query_exec/
│   │   ├── row_executor.hpp
│   │   ├── column_filter.hpp
│   │   ├── column_project.hpp
│   │   └── hash_join.hpp
│   ├── src/
│   │   ├── column_filter.cpp
│   │   ├── column_project.cpp
│   │   └── hash_join.cpp
│   ├── benchmark/
│   │   └── benchmark.cpp       # build Release; see W08 for why that's not optional
│   └── CMakeLists.txt
│
├── feature-pipeline/       # W09: Python
│   ├── feature_store.py
│   ├── pipeline.py
│   └── requirements.txt
│
├── distributed-training/   # W10: Python + Go tool
│   ├── mlp.py
│   ├── ring_allreduce.py
│   ├── worker.py
│   ├── train.py
│   └── requirements.txt
│   # Go tool lives in tools/grad_server/
│
├── actor-training/         # W11: Python + Ray
│   ├── model.py             # PyTorch CNN
│   ├── worker_actor.py      # @ray.remote TrainerWorker
│   ├── parameter_server_actor.py  # @ray.remote ParameterServer
│   ├── train.py
│   ├── compare.py           # sequential vs W10 ring-allreduce vs Ray actors
│   └── requirements.txt
│
├── gpu-gemm/               # W12: Python/Numba + C fallback
│   ├── naive_gemm.py
│   ├── tiled_gemm.py
│   ├── benchmark.py
│   ├── roofline.py
│   ├── gemm_fallback.c     # no-GPU fallback
│   └── requirements.txt
│
├── attention/              # W13: Python
│   ├── attention.py        # MultiHeadAttention
│   ├── kv_cache.py
│   ├── benchmark.py
│   └── requirements.txt
│
├── snapshot/                # W14: Go
│   ├── channel.go
│   ├── message.go
│   ├── node.go
│   ├── coordinator.go
│   ├── snapshot_test.go
│   └── go.mod
│
├── capstone/                # W15: your choice of language
│   ├── README.md           # required: design doc
│   └── ...
│
├── operator/                # W16: Go
│   ├── api/v1/
│   │   ├── types.go
│   │   └── register.go
│   ├── controllers/
│   │   └── reconciler.go
│   ├── config/
│   │   ├── crd.yaml
│   │   └── sample.yaml
│   ├── main.go
│   └── go.mod
│
├── dd-scratch/             # W17: extends W07 (C++)
│   ├── include/dd_scratch/
│   │   ├── metrics.hpp          # prometheus-cpp
│   │   ├── tracing_setup.hpp    # opentelemetry-cpp, ScopedSpan RAII helper
│   │   └── logging.hpp          # nlohmann::json structured log lines
│   └── src/
│       ├── metrics.cpp
│       ├── tracing_setup.cpp
│       └── logging.cpp
│
└── capstone-platform/      # W18 (optional): Go + Python, combines W09+W10+W13+W14+W16+W17
    ├── train_worker.py
    ├── checkpoint_coordinator.py
    ├── serve.py
    ├── operator/            # extends code/operator/ from W16
    └── README.md            # required: design doc
```

---

## Build Commands

**C++ (CMake):**
```bash
cmake -S . -B build -DCMAKE_TOOLCHAIN_FILE=$VCPKG_ROOT/scripts/buildsystems/vcpkg.cmake
cmake --build build
ctest --test-dir build                              # GoogleTest suite, where present
cmake -S . -B build -DCMAKE_BUILD_TYPE=Release && cmake --build build && ./build/benchmark
                                                      # Release build required for anything timed (W08)
```

**Python:**
```bash
python -m venv venv && source venv/bin/activate
pip install -r requirements.txt
python main.py
```

**Go:**
```bash
go mod tidy
go run main.go
go test ./...
go build -o bin/app .
```

---

## Notes

- Keep each project buildable in isolation. No shared parent build file.
- Code in this directory is the "lab." It's meant to be written, broken, and rewritten.
- The `tools/` directory (at repo root) holds automation scripts that aren't part of a specific week's deliverable
- C++ projects (W05–W08, W17) keep tests in a separate `tests/` directory using GoogleTest, the CMake-project convention — unlike Rust's inline `#[cfg(test)] mod tests`, C++ test files are separate translation units registered in `CMakeLists.txt` via `gtest_discover_tests`
