# Lab code

Create one Maven project under `code/query-engine/`. The weekly notes specify the exercises. Implement each addition in the same project and run the earlier tests as you go.

Use `src/main/java` and `src/test/java`. Group code by its role once the files grow: storage and codecs, message simulation and replication, plans and execution, streaming and replay, and process entry points. Keep reference implementations and deterministic test traces in the project.

Keep Python comparison scripts under `code/experiments/` and Python extension labs under `code/ml/`. Share fixtures and result formats with Java through files. Record commands and dependency versions alongside the scripts.

## Shared fixtures

The small commerce event fixture should have `event_id`, `customer_id`, `event_type`, `value_cents`, and `event_ts_ms`. Use a header, positive numeric event and customer IDs, event types `PURCHASE` and `VIEW`, signed integer cents, and nonnegative UTC epoch milliseconds. Use the corresponding Java `Event` record in both the storage and execution labs. Keep a tiny checked-in CSV and a table of expected counts. Use a fixed generator seed for larger uniform and skewed variants. The streaming trace adds explicit watermark advances and can reuse the same event schema.

## Boundaries to test

- Encoding and operator logic should be testable independently of files, clocks, queues and network connections.
- The row interpreter is the reference for the batched and partitioned paths.
- For process execution, ship an explicit operator or task description and versioned records. Keep the first protocol to the fixed capstone query and two workers.
- For replay, use deterministic trace inputs and compare with an uninterrupted run.

Keep records immutable at module boundaries and give mutable buffers a single owner. Use Java 21 record patterns and exhaustive switches for protocol and plan variants. For blocking worker I/O use virtual threads with an explicit in-flight limit. Give CPU-heavy operators a fixed worker budget.
