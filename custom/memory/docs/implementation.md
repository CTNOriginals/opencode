# Memory System - Implementation Plan

## Architecture Summary

```
opencode agent
    |
    | (post-turn hook)
    v
MCP Server (Go, stdin/stdout JSON-RPC)
    |
    +--> Writer: analyze turn -> update PFC -> promote to HC -> promote to NC
    |
    +--> Reader: query neocortex by keywords, return ranked results
    |
    v
SQLite (WAL mode, separate reader/writer connections)
```

## Data Flow (per turn)

1. Agent finishes a turn
2. Hook sends `{latest_turn, pfc_state}` to MCP server
3. Writer's LLM call: analyze context, decide what to remember, extract keywords, assign initial rate/weight
4. **PFC stage**:
   - LLM compares new context against existing PFC items (semantic matching)
   - Reused items: reset ticks, update content, blend weight (70/30 favor original)
   - New items: add with LLM-assigned rate/weight, score starts at `w * r^0 = w`
   - All items advance 1 tick (global counter)
   - Evict items where score < 0.5 (low threshold)
   - Promote items where score > 0.7 (high threshold) to hippocampus
5. **Hippocampus stage**:
   - Items promoted from PFC stored in SQLite
   - Items where score < 0.3: forgotten (deleted)
   - Items where score > 0.7: promoted to neocortex
6. **Neocortex stage**:
   - LLM evaluates hippocampal items for usefulness
   - Useful items stored with LLM-extracted keywords (hybrid: LLM + statistical backup)
   - Items where final score < 0.6 at evaluation: forgotten

## SQLite Schema

```sql
-- Hippocampus (short-term)
CREATE TABLE hippocampus (
    id TEXT PRIMARY KEY,
    content TEXT NOT NULL,
    score REAL NOT NULL,
    weight REAL NOT NULL,
    rate REAL NOT NULL,
    tick_count INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Neocortex (long-term)
CREATE TABLE neocortex (
    id TEXT PRIMARY KEY,
    content TEXT NOT NULL,
    score REAL NOT NULL,
    keywords TEXT NOT NULL,  -- JSON array
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_accessed DATETIME
);

CREATE INDEX idx_neocortex_keywords ON neocortex(keywords);
```

## PFC Item Struct (in-memory)

```go
type PFCItem struct {
    ID          string
    Content     string
    Score       float64
    Weight      float64
    Rate        float64
    TickCount   int
    CreatedAt   time.Time
}
```

## Scoring Formula

```
score = weight * rate^tick_count
```

- New item: `score = weight * rate^0 = weight`
- After 1 tick: `score = weight * rate^1`
- On reuse: reset `tick_count = 0`, blend `weight = 0.7 * old_weight + 0.3 * llm_weight`

## Key Decisions

| Decision | Choice |
|---|---|
| Runtime | Go, separate process |
| IPC | stdin/stdout JSON-RPC |
| Persistence | SQLite (WAL mode) |
| Tool exposure | MCP server |
| PFC storage | In-memory ring buffer |
| Ticks | Global counter, 1 per agent turn |
| Content format | Plain text |
| Error handling | Fail silently + log |
| Writer LLM | Single call per turn |
| PFC matching | LLM semantic matching |
| Weight blending | 70/30 favor original |
| Project structure | Simple package |

## Deferred to Post-v1

- MCP tool names/signatures
- MCP library selection
- Config file format
- Testing strategy
