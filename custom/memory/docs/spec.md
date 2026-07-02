# Memory System

A multi-stage memory pipeline for opencode agents. Agents are read-only consumers of memory; writes happen automatically via a separate process.

## Pipeline Stages

### 1. Prefrontal Cortex (active context)

Ring buffer of `{ score, content }` structs.

- New items are pushed to the front. When the buffer is full, the tail item is evicted (natural forgetting).
- On reuse: item is moved back to the front; score and content are updated with any new relevant context.
- When an item's score crosses a promotion threshold, it is sent to the hippocampus.

### 2. Hippocampus / Encoding (short-term)
- Receives context from the prefrontal cortex.
- Encodes it into short-term storage for longer retention.
- Items can remain in the prefrontal cortex concurrently; subsequent revisits may update the hippocampal entry.

### 3. Neocortex (long-term)
- At session end, hippocampal contents are evaluated for usefulness.
- Useful items are further encoded and stored in long-term memory.
- Storage is a key-value map: **keys** = keywords defining the memory, **values** = memory content.

## Scoring System

Each stage maintains a score for every memory item. Score determines retention and promotion.

- `t` = ticks since last reuse (1 tick = 1 prefrontal cortex update cycle)
- `r` = rate (base retention, configurable per stage)
- `w` = weight (importance of the item, e.g. mistakes-to-learn-from have higher weight)

```
every tick:
  t += 1
  score -= r * (t / 100) * w

on reuse:
  score += r * w
  t = 0
```

Items are forgotten when score drops below a stage-specific threshold. Higher `r` means slower decay; more frequent reuse keeps score elevated.

## System Architecture

- **Memory Reader** — tool available to agents. Takes an arbitrary query (e.g. a plan, a user prompt, a problem description) and searches the neocortex for keyword matches, returning memories ranked by relevance score. The agent uses these scores to decide how much weight to give each memory in its context.
- **Memory Writer** — a separate agent process triggered by a hook — receives the main agent's context, analyzes it, and appends structured entries to memory.
