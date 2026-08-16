---
week_number: 99
status: not-started
---

# WXX: [Topic]

> **Arc:** [Arc Name] · **Language:** [Primary Language]
> **Budget:** about [N] hours. The Minimum bar is what a bad week looks like, not the target.

## What you'll build

[One paragraph. Start with: "A [thing] in [language] that [does what]." Be specific enough that someone can start without reading the rest. Name the files they'll create.]

**Scenario:** [The situation at work where this matters. One concrete failure or decision, not a category.]

---

## Read

- [ ] [Paper or book chapter title](URL): Author(s), Year, [what to focus on, estimated time]
- [ ] [Second source](URL): [what to focus on]

**Depth: [study / read / skim] [which source].** [One line on which sources get which treatment.]

**[Vocabulary the unit assumes]**, as a bulleted list rather than a paragraph, if the exercises use terms the reading does not define plainly:

- **[Term]**: [one or two sentences]
- **[Term]**: [one or two sentences]

**Key question:** [One question that forces synthesis of the reading. Should not be answerable by skimming. Something like "Why does X have to be Y, and what breaks if it isn't?"]

---

## Code

Project: `code/[directory]/` ([Language + version])

[Name every artifact the student will produce, up front, before the steps. A table or a short list.]

### Step 1: `[FileName.ext]`

- [ ] [What it does, named fields or methods to implement, any constraints.]

### Step 2: `[FileName.ext]`

- [ ] [What it does.]

### Step 3: verify

- [ ] [Exact command that proves correctness: a test, a benchmark output, a curl. State whether repeated runs are sequential or concurrent, how long to run, and what number to expect.]

**Minimum bar:** [What "done" looks like. Be specific: "X test passes", "benchmark shows Nx speedup", "3 nodes converge".]

---

## Break it, then decide

- [ ] [A deliberate failure with a real production analogue. Say what the symptom looks like and where the student has to be looking to see it.]
- [ ] **Your call:** [A genuine judgment call with defensible answers either way. Ask for a choice, an implementation, and the signal they would monitor to find out the choice was wrong.]

---

## Reflect

**Prediction versus measurement.** Fill the prediction in *before* running anything, and do not edit it afterwards. The gap is the point. Omit this table in units that install rather than measure.

| Quantity | Predicted | Measured | Which term I got wrong |
|----------|-----------|----------|------------------------|
| | | | |

Copy any number worth keeping into [MEASUREMENTS.md](../MEASUREMENTS.md).

**What clicked:**

**What surprised me:**

**[Week-specific question 1, about the core concept]:**

**[Week-specific question 2, connecting to a system you know]:**

**What I'd do differently:**

---

## House style for a unit

Delete this section from the unit you write; it is guidance for the author, not for the student.

- Name every file, endpoint, and config artifact before the steps that build them. No "the two endpoints" without saying which two.
- Ordered `### Step N` headings wherever there is a build sequence.
- Every config file the unit names gets its real content, including the fields that decide whether it works.
- Verification says exactly how: sequential or concurrent, how many, how long, what to expect.
- Quote and link upstream documentation instead of writing your own explanation of a library's behaviour.
- No prose that explains a choice the student does not make, and no commentary on why the curriculum is arranged the way it is.
- No record of the plan's own history: nothing was cut, renamed, or previously done differently.
- Never open a sentence by counting what follows ("Two things to notice", "Three details decide"). Use a list and let it do the counting.
- No em-dashes anywhere.
