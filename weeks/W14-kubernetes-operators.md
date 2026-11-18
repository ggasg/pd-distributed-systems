---
week_number: 12
track: core
title: "Deploying the query engine"
status: not-started
---

# Deploying the query engine

**Budget:** 10 hours across four study sessions. Shift the sessions to fit your week.

**Start with:** A working Scala engine with a repeatable workload and metrics.

## Read

Read the Kubernetes [controller overview](https://kubernetes.io/docs/concepts/architecture/controller/) and describe how desired state differs from live worker state.

## Sessions

1. [ ] **Establish a local reference.** Add a minimal transport for the fixed query: let the coordinator send a partition task to each worker and receive versioned rows or partial counts. Run with explicit local addresses and test the result against the in-process reference.
2. [ ] **Package the workers.** Build an image that contains the Scala application and document its runtime arguments.
3. [ ] **Deploy to kind.** Use ordinary Kubernetes workloads and Services to run the workers and reach them from the coordinator. Run the same fixture and collect the existing metrics.
4. [ ] **Diagnose one failure.** Stop a worker or misconfigure its address and inspect events and logs. Check what Kubernetes restarts and what the query runtime has to retry.

**Minimum deliverable:** The query returns the expected result with two worker processes, and the failure diagnosis identifies the responsible layer.

**Handoff:** The capstone uses the same topology and workload to compare the single-worker and distributed runs.
