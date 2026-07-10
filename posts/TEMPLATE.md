# [WXX] [Title]: What I Built and What I Learned

> **Week:** WXX · **Arc:** [Arc Name] · **Language:** [Language]  
> **Repo:** [link to your code/directory/]

---

## What I Built

[2–3 sentences. Name the specific thing you built, not "I implemented a data structure" but "I built a 3-node simulated distributed snapshot using Chandy-Lamport markers over FIFO channels in Java." A fellow engineer should know exactly what this is after reading these sentences.]

---

## The Core Idea

[3–5 sentences explaining the concept to a peer who hasn't read the paper. Use an analogy if it helps. Avoid jargon unless you immediately define it. The test: could a smart engineer with no background in this topic understand what problem you're solving?]

---

## The Implementation

[Walk through the key design decision or the most interesting part of the code. Include a short snippet if it makes a point clearly. Don't paste the whole file, pick the 5–15 lines that are the heart of the thing.]

```[language]
// The interesting part
```

[Explain what makes this non-obvious. What did you have to think about? What assumption does this code depend on?]

---

## What Surprised Me

[One thing you didn't expect: a bug, a performance result, a conceptual shift, something the paper got wrong or glossed over. This is the most valuable section for other engineers reading your posts.]

---

## How This Connects to Real Systems

[Name a real system (Kafka, Flink, Postgres, DynamoDB, etc.) and explain concretely how what you built appears in it. "This is basically how Kafka's log compaction works" or "FlashAttention solves the same memory bandwidth problem I hit in W12."]

---

## What I'd Do Next

[One concrete follow-on. Not vague like "explore further" but specific like "implement the all-gather phase of ring-allreduce over UDP to see if the extra complexity is worth it."]

---

*This is week X of a [18]-week engineering curriculum on distributed systems. [Link to repo]*
