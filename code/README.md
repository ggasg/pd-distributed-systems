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
├── streaming/               # W05: Rust
│   ├── src/
│   │   ├── lib.rs
│   │   ├── event.rs
│   │   ├── watermark.rs
│   │   ├── aggregator.rs       # TumblingWindowAggregator
│   │   └── processor.rs        # StreamProcessor + StreamItem enum; #[cfg(test)] tests live here
│   └── Cargo.toml
│
├── timely-toy/             # W06: Rust
│   ├── src/
│   │   ├── lib.rs
│   │   ├── timestamp.rs
│   │   ├── pointstamp.rs
│   │   ├── operator.rs         # Operator trait, MapOperator, SinkOperator
│   │   └── progress_tracker.rs
│   └── Cargo.toml
│
├── dd-scratch/             # W07: Rust
│   ├── src/
│   │   ├── lib.rs
│   │   ├── update.rs
│   │   ├── collection.rs
│   │   ├── word_count.rs
│   │   └── reachability.rs
│   └── Cargo.toml
│
├── query-exec/             # W08: Rust
│   ├── src/
│   │   ├── lib.rs
│   │   ├── row_executor.rs
│   │   ├── column_filter.rs
│   │   ├── column_project.rs
│   │   └── hash_join.rs
│   ├── src/bin/
│   │   └── benchmark.rs        # cargo run --release --bin benchmark
│   └── Cargo.toml
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
├── dd-scratch/             # W17: extends W07
│   └── src/
│       ├── metrics.rs
│       ├── tracing_setup.rs
│       └── logging.rs
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

**Rust (Cargo):**
```bash
cargo build
cargo test
cargo run                        # default bin target, if any
cargo run --release --bin benchmark   # named bin target; use --release for anything timed
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
- Rust projects (W05–W08, W17) keep unit tests inline via `#[cfg(test)] mod tests { ... }` at the bottom of the relevant file, in the Rust convention — no separate `*Test.rs` files
