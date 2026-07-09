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

score = `w * r^t`

| Stage | rate | Threshold (high/low) | Behavior |
|---|---|---|---|
| Prefrontal cortex | 0.85 | 0.7 / 0.5 | promote to hippocampus when score > high; evict from ring buffer when score < low |
| Hippocampus | 0.95 | 0.7 / 0.3 | promote to neocortex when score > high; forget when score < low |
| Neocortex | 0.98 | 0 / 0.6 | high n/a; keep only if final score > low at session end |

Items are forgotten when score drops below a stage's low threshold (or promoted when score rises above its high threshold). Higher `r` means slower decay; more frequent reuse keeps score elevated.

## System Architecture

- **Memory Reader** — tool available to agents. Takes an arbitrary query (e.g. a plan, a user prompt, a problem description) and searches the neocortex for keyword matches, returning memories ranked by relevance score. The agent uses these scores to decide how much weight to give each memory in its context.
- **Memory Writer** — a separate agent process triggered by a hook — receives the main agent's context, analyzes it, and appends structured entries to memory.
