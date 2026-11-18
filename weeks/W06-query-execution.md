---
week_number: 6
track: core
title: "Query execution and batching"
status: not-started
---

# Query execution and batching

**Budget:** 10 hours across four study sessions. Shift the sessions to fit your week.

**Start with:** The scan and Exchange from earlier units.

## Read

Read DDIA 2e Ch. 3 on declarative queries and [MonetDB/X100](https://www.cidrdb.org/cidr2005/papers/P19.pdf) on execution granularity.

## Sessions

1. [ ] **Set a baseline.** Time the Java row iterator for the existing filter and count. Use DuckDB SQL to check the result and inspect its physical plan.

   Use the existing Java iterator path; keep storage and input fixtures the same between Java runs. Assert that DuckDB produces the same sorted purchase counts and capture its plan. Measure DuckDB separately from the controlled Java comparison.
2. [ ] **Introduce batches.** Add a batch scan and filter. Keep any mutable buffers inside operators and test the same semantics for empty inputs and short final batches.

   Implement `Iterator<EventBatch> scanBatches(int maxRows)` and `EventBatch filterBatch(EventBatch batch, Predicate<Event> predicate)`. Reject nonpositive sizes, preserve input order, and emit the shorter final batch. Keep buffers owned by the operator or document their lifetime explicitly. Test empty input, all rows filtered, and a partial final batch.
3. [ ] **Measure.** Hold the worker count and workload fixed. Warm up the JVM and compare throughput and allocation across batch sizes.

   Measure rows per second for row and batch paths at one worker first, then at several batch sizes. Separate preparation from timed execution, consume the result, and report allocation with the same profiler and settings or record why unavailable.
4. [ ] **Keep the reference.** Make both modes runnable with the same query. Record the comparison conditions and keep the iterator as a correctness oracle.

   Expose row and batch modes through the same entry point and return the same grouped-count type. Add equivalence tests on both skew fixtures so W07 can change physical operators without changing required results.

**Minimum deliverable:** Equivalent row and batch results plus a controlled throughput comparison.

**Handoff:** W07 chooses physical operators using a logical plan and checks each choice against the reference.

## Numeric correctness (optional, 60–90 minutes)

- [ ] **Build an average you can merge across workers.**

   Start with the purchase filter and customer grouping from the core lab. Each input amount is an integer number of cents. A *partial aggregate* is the summary a worker sends for its share of the rows. Here it contains a sum and a count, so the final average gives every purchase equal weight.

   1. **Work the example by hand.** One worker sees `100` cents. Another sees `200, 300, 400`. Their partial states are `(100, 1)` and `(900, 3)`. Together they give `1,000 / 4 = 250` cents, or `2.50` dollars. Averaging the workers' averages would give `2.00` dollars because it weights the two workers equally.
   2. **Implement the state.** Define an immutable `AverageState(BigInteger sumCents, long count)`. Start at `(0, 0)`. Make `add(long cents)` and `merge(AverageState other)` return new states. Add counts with `Math.addExact`. In `Optional<BigDecimal> finish()`, return empty for zero rows. Otherwise convert the exact sum to dollars and divide by the exact count with scale `2` and `RoundingMode.HALF_EVEN`. Round only here.
   3. **Check rounding and empty state.** `[100, 101]` cents should yield `1.00` dollars, `[101, 102]` should yield `1.02`, and `[-101, -102]` should yield `-1.02`. Merging with the empty state should leave a nonempty state unchanged. A grouped query with no purchases should return no customer rows.
   4. **Compare execution paths.** First test two partial states in one process. Then run the same purchase query through the row, batch, and W05 two-worker paths. Use the same exact state for all three. Try a different split and reverse the order of the partials. If a result changes, inspect where the code divides or rounds.

   **Done when:** A checked-in fixture and tests show the expected amounts, exact agreement between execution paths, and the wrong result from averaging partition averages.
