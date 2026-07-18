---
week_number: 14
status: not-started
---

# W14: The Actor Model and Ray

> **Arc:** Distributed ML & Compute · **Language:** Python (Ray)

## What you'll build
A Ray-based actor system in Python that retrains the same data-parallel MNIST job from W13, this time using stateful actors (a pool of `TrainerWorker` actors plus a `ParameterServer` actor) instead of raw sockets, backed by a real PyTorch model instead of a hand-rolled NumPy MLP. You'll benchmark it against W13's implementation to see what the actor abstraction buys you, and what it costs.

---

## Read
- [ ] [A Universal Modular ACTOR Formalism for Artificial Intelligence](https://www.ijcai.org/Proceedings/73/Papers/027B.pdf) (Hewitt, Bishop, Steiger, IJCAI 1973): read Sections 1–3. This is the original actor model paper, no Erlang or Akka involved. The core claim: an actor is a computational primitive that has private state, communicates only via asynchronous messages, and processes one message at a time. Everything Ray does with actors is this idea, 45 years later.
- [ ] [Ray: A Distributed Framework for Emerging AI Applications](https://www.usenix.org/system/files/osdi18-moritz.pdf) (Moritz et al., OSDI 2018): read Section 3 (Programming Model) closely, skim the rest. Focus on how Ray unifies stateless *tasks* and stateful *actors* under one API, and why that distinction exists at all. Ray is Anyscale's product; this week is hands-on with a target company's core framework, not a paper read at a distance.

**Key question:** Ray's programming model gives you both stateless tasks and stateful actors. Why does distributed training specifically need actors and not just tasks? Concretely, what state would you lose between calls if every worker were a stateless task instead of an actor?

---

## Code

Project: `code/actor-training/` (Python 3.11+, `ray`, `torch`)

Scenario: reimplement W13's data-parallel training job, but with two changes: workers are Ray actors instead of raw processes talking over sockets, and the model is a real PyTorch `nn.Module` instead of a hand-rolled NumPy MLP.

- [ ] `model.py`: a small PyTorch CNN for MNIST (2 conv layers + 1 linear layer is enough). Plain `torch.nn.Module`, nothing distributed here.
- [ ] `worker_actor.py`: `@ray.remote class TrainerWorker`. Holds a local copy of the model and optimizer as actor state (set once in `__init__`, mutated across calls, this is the part a stateless task couldn't do). Methods: `compute_gradients(batch) -> grads`, `apply_gradients(grads) -> None`, `set_weights(weights) -> None`, `get_weights() -> weights`.
- [ ] `parameter_server_actor.py`: `@ray.remote class ParameterServer`. Holds the canonical model weights. Methods: `push_gradients(worker_id, grads) -> None` (accumulates gradients from each worker; once all N workers for this round have pushed, averages them and updates its own weights), `pull_weights() -> weights`. Note that Ray guarantees calls to a single actor are processed one at a time, so `push_gradients` from concurrent workers can safely mutate shared accumulator state with no explicit lock.
- [ ] `train.py`: `ray.init()`, spin up N `TrainerWorker` actors and 1 `ParameterServer` actor. Each round: workers `pull_weights()`, run a local batch through `compute_gradients()`, `push_gradients()` to the server; once the server has averaged, workers `pull_weights()` again for the next round. Train for 5 epochs on MNIST, same shard split as W13.
- [ ] `compare.py`: run the same training job three ways and print wall-clock time plus final accuracy for each: (1) single-process sequential baseline, (2) W13's raw-socket ring-allreduce (import directly from `code/distributed-training/`), (3) this week's Ray actor-based parameter server.

**Verify:** all three implementations in `compare.py` converge to comparable accuracy (within a couple percentage points). Run `ray list actors` (or `ray.util.state.list_actors()`) during training and confirm N+1 actors are alive.

**Minimum bar:** N≥2 `TrainerWorker` actors plus 1 `ParameterServer` actor coordinate correctly with no shared memory between them: all communication happens through actor method calls, matching the message-passing model from the 1973 paper, not through globals or files.

---

## 🐍 Python DSA Review (optional)

**Toy actor mailbox (single-threaded message loop)**: before Ray hides the mechanics from you, build the simplest possible version by hand. This is exactly what makes actor state safe to mutate without locks: mutation only ever happens from inside a loop that processes one message at a time.

```python
# toy_actor.py
from collections import deque

class ToyActor:
    def __init__(self):
        self.mailbox = deque()
        self.state = 0  # mutable actor state, only ever touched inside process_one()

    def send(self, op: str, value=None) -> None:
        self.mailbox.append((op, value))

    def process_one(self):
        if not self.mailbox:
            return None
        op, value = self.mailbox.popleft()
        if op == "add":
            self.state += value
            return None
        if op == "get":
            return self.state

a = ToyActor()
a.send("add", 5)
a.send("add", 3)
a.send("get")
a.process_one()          # state = 5
a.process_one()          # state = 8
assert a.process_one() == 8   # processed strictly in order, no locks needed
```

**Connection:** `ParameterServer.push_gradients()` is this same pattern at scale. Ray gives every actor its own mailbox and guarantees method calls to one actor process sequentially, so averaging gradients from concurrent workers is exactly as safe as `process_one()` mutating `self.state` above. Compare this to W13's `ring_allreduce.py`, where you had to reason about socket message ordering by hand; the actor model buys you that ordering guarantee for free, at the cost of the parameter server becoming a potential bottleneck (more on that in Reflect).

---

## Reflect

**What clicked:**

**What surprised me:**

**Why does the actor abstraction make `push_gradients` safe without explicit locks, when W13's raw-socket workers needed you to reason about message ordering yourself?**

**Where does a single `ParameterServer` actor start to become a bottleneck at scale, and how does that connect back to why W13's ring-allreduce avoids a central coordinator entirely?**

**What I'd do differently:**
