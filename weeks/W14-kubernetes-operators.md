---
week_number: 12
track: core
title: "Deploying the query engine"
status: not-started
---

# Deploying the query engine

**Budget:** 10 hours across four study sessions. Shift the sessions to fit your week.

**Start with:** A working Java engine with a repeatable workload and metrics.

## Read

Read the Kubernetes [controller overview](https://kubernetes.io/docs/concepts/architecture/controller/) and describe how desired state differs from live worker state.

## Sessions

1. [ ] **Establish a local reference.** Add a minimal transport for the fixed query: let the coordinator send a partition task to each worker and receive versioned rows or partial counts. Run with explicit local addresses and test the result against the in-process reference.

   Use a versioned request with query ID, partition ID and a bounded payload of rows already routed by W05. Each worker returns matching IDs and grouped counts. The coordinator merges a partition reply once and fails the query on timeout, malformed reply, or a missing partition. Use virtual threads for blocking connections and cap concurrent requests and CPU work separately. Exercise the local protocol with two workers.
2. [ ] **Package the workers.** Build an image that contains the Java application and document its runtime arguments.

   Provide `worker --port N` and `query --input PATH --workers URL1,URL2` entry points. Document the image build, JVM version and how to make the fixture available. Verify the process exits nonzero on query failure.
3. [ ] **Deploy to kind.** Use ordinary Kubernetes workloads and Services to run the workers and reach them from the coordinator. Run the same fixture and collect the existing metrics.

   Deploy two worker endpoints and run the coordinator with both explicit addresses. Use one worker Service per endpoint for the lab, and show in logs or metrics that both workers handled distinct partitions. Run a two-customer purchase fixture with keys that hash to different partitions, then compare with the single-worker result.
4. [ ] **Diagnose one failure.** Stop a worker or misconfigure its address and inspect events and logs. Check what Kubernetes restarts and what the query runtime has to retry.

   Stop one worker during a request and record the coordinator error and the Kubernetes events. Verify no partial aggregate is presented as success. After readiness returns, rerun the whole query and verify exact results.

**Minimum deliverable:** The query returns the expected result with two worker processes, and the failure diagnosis identifies the responsible layer.

**Handoff:** The capstone uses the same topology and workload to compare the single-worker and distributed runs.
