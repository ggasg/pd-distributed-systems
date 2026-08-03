---
week_number: 11
status: not-started
---

# W11: The Actor Model and Ray

> **Arc:** Distributed ML & Compute · **Language:** Python (Ray)
> **Budget:** about 5 hours. Hit the Minimum bar first; everything past it is optional.

## What you'll build
A Ray-based actor system in Python that retrains the same data-parallel MNIST job from W09, this time using stateful actors (a pool of `TrainerWorker` actors plus a `ParameterServer` actor) instead of raw sockets, backed by a real PyTorch model instead of a hand-rolled NumPy MLP. You'll benchmark it against W09's implementation to see what the actor abstraction buys you, and what it costs.

**Scenario:** the actor model just fixed the exact hang you caused on purpose in W09, concurrent writers to shared state can't race anymore. It did nothing, though, about a worker that simply never shows up to a round, and that's a different failure with a different fix, which is what the exercise below is actually about.

---

## Read
- [ ] Optional: [A Universal Modular ACTOR Formalism for Artificial Intelligence](https://www.ijcai.org/Proceedings/73/Papers/027B.pdf) (Hewitt, Bishop, Steiger, IJCAI 1973): read Sections 1–3. This is the original actor model paper, no Erlang or Akka involved. The core claim: an actor is a computational primitive that has private state, communicates only via asynchronous messages, and processes one message at a time. Everything Ray does with actors is this idea, 45 years later.
- [ ] [Ray: A Distributed Framework for Emerging AI Applications](https://www.usenix.org/system/files/osdi18-moritz.pdf) (Moritz et al., OSDI 2018): read Section 3 (Programming Model) closely, skim the rest. Focus on how Ray unifies stateless *tasks* and stateful *actors* under one API, and why that distinction exists at all.

**A note on why this week uses Ray specifically.** Ray moved to the PyTorch Foundation, under the Linux Foundation, in October 2025, alongside PyTorch and vLLM. Anyscale, the company originally behind it, was acquired by Nscale in 2026, and the project's neutral governance is the reason that acquisition changes nothing about what you're learning here: the framework is community-governed and runs on any infrastructure. So this week is not a bet on a vendor. It is the only mainstream way to work with the actor model in Python ML infrastructure, and the actor model is the actual subject. Ray appears exactly once in this curriculum, here, for that reason; the orchestration layer in W14 and W16 deliberately uses the vendor-neutral Kubeflow Trainer instead, so no single framework carries more of this plan than it has earned.

**Depth: read Section 3 of the Ray paper.** Hewitt 1973 is a skim; it is worth seeing the original framing, but the actor model's content here comes from using it, not from the paper.

**Key question:** Ray's programming model gives you both stateless tasks and stateful actors. Why does distributed training specifically need actors and not just tasks? Concretely, what state would you lose between calls if every worker were a stateless task instead of an actor?

---

## Code

Project: `code/actor-training/` (Python 3.13+, `ray`, `torch`)

Scenario: reimplement W09's data-parallel training job, but with two changes: workers are Ray actors instead of raw processes talking over sockets, and the model is a real PyTorch `nn.Module` instead of a hand-rolled NumPy MLP.

- [ ] `model.py`: a small PyTorch CNN for MNIST (2 conv layers + 1 linear layer is enough). Plain `torch.nn.Module`, nothing distributed here.
- [ ] `worker_actor.py`: `@ray.remote class TrainerWorker`. Holds a local copy of the model and optimizer as actor state (set once in `__init__`, mutated across calls, this is the part a stateless task couldn't do). Methods: `compute_gradients(batch) -> grads`, `apply_gradients(grads) -> None`, `set_weights(weights) -> None`, `get_weights() -> weights`.
- [ ] `parameter_server_actor.py`: `@ray.remote class ParameterServer`. Holds the canonical model weights. Methods: `push_gradients(worker_id, grads) -> None` (accumulates gradients from each worker; once all N workers for this round have pushed, averages them and updates its own weights), `pull_weights() -> weights`. Note that Ray guarantees calls to a single actor are processed one at a time, so `push_gradients` from concurrent workers can safely mutate shared accumulator state with no explicit lock.
- [ ] `train.py`: `ray.init()`, spin up N `TrainerWorker` actors and 1 `ParameterServer` actor. Each round: workers `pull_weights()`, run a local batch through `compute_gradients()`, `push_gradients()` to the server; once the server has averaged, workers `pull_weights()` again for the next round. Train for 5 epochs on MNIST, same shard split as W09.
- [ ] `compare.py`: run the same training job three ways and print wall-clock time plus final accuracy for each: (1) single-process sequential baseline, (2) W09's raw-socket ring-allreduce (import directly from `code/distributed-training/`), (3) this week's Ray actor-based parameter server.

**Verify:** all three implementations in `compare.py` converge to comparable accuracy (within a couple percentage points). Run `ray list actors` (or `ray.util.state.list_actors()`) during training and confirm N+1 actors are alive.

**Minimum bar:** N≥2 `TrainerWorker` actors plus 1 `ParameterServer` actor coordinate correctly with no shared memory between them: all communication happens through actor method calls, matching the message-passing model from the 1973 paper, not through globals or files.

**Break it, then decide:**
- [ ] With 3+ `TrainerWorker` actors running, kill one mid-round: `ray.kill(worker_handles[i], no_restart=True)`. Watch `ParameterServer.push_gradients()`, which waits until all N workers for the round have pushed before averaging. With one worker gone for good, that condition can never become true, so the round (and the whole training loop) hangs forever, waiting on an actor that no longer exists. The actor model made concurrent writes safe; it did nothing to make "everyone eventually shows up" a guarantee.
- [ ] This is the same class of problem as W09's dead worker, and underneath it the same ambiguity you built a detector for in W03: a worker that has not pushed gradients is indistinguishable from one that is about to. You now have a real design choice available that raw sockets didn't make convenient: would you have `ParameterServer` proceed once it's heard from a quorum (say, N-1 of N) within a timeout, averaging only the gradients that arrived and accepting the result is now slightly stale, or keep waiting for all N and treat a permanently missing worker as a fatal error requiring the whole job to restart? The first is closer to real asynchronous/stale-SGD parameter servers; the second is simpler and matches what synchronous SGD actually requires for its convergence guarantees to hold. Pick one, implement a timeout-based quorum in `push_gradients`, and say what accuracy or convergence cost you'd expect to pay for the choice you didn't make.

## Reflect

**What clicked:**

**What surprised me:**

**Why does the actor abstraction make `push_gradients` safe without explicit locks, when W09's raw-socket workers needed you to reason about message ordering yourself?**

**Where does a single `ParameterServer` actor start to become a bottleneck at scale, and how does that connect back to why W09's ring-allreduce avoids a central coordinator entirely?**

**Quorum-with-timeout or wait-for-all-N, and what did that decision cost you (from Break it, then decide above)?**

**What I'd do differently:**
