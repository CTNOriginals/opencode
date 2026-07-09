# Memory System - Architecture & Project Backlog

This file is a sequential task list. An agent executes tasks in order from top to bottom.
Each task depends on its predecessors being complete. Do not skip ahead.

---

## Environment

- Go 1.26+, `/usr/bin/go`, `go1.26.4-X:nodwarf5 linux/amd64`
- CGO available: GCC 16.1.1 at `/usr/bin/gcc`
- OS: Linux
- OpenCode config: `/home/ctn/.config/opencode/opencode.jsonc`
- Working directory for memory project: `/home/ctn/.config/opencode/custom/memory/`

---

## Directory Structure to Create

```
custom/memory/
  go.mod
  go.sum
  main.go
  internal/
    pfc/
      ring.go
    hippocampus/
      store.go
    neocortex/
      store.go
      search.go
    writer/
      pipeline.go
    sqlite/
      conn.go
      schema.go
  docs/
    spec.md
    implementation.md
    architecture.md (this file)
```

---

## Key Constants (Hardcoded, per v1 constraint)

| Constant | Value | Where Used |
|---|---|---|
| PFC ring buffer size | 50 | pfc/ring.go |
| PFC rate (r) | 0.85 | pfc/ring.go |
| PFC promote threshold | 0.7 | pfc/ring.go |
| PFC evict threshold | 0.5 | pfc/ring.go |
| Hippocampus rate (r) | 0.95 | hippocampus/store.go |
| Hippocampus promote threshold | 0.7 | hippocampus/store.go |
| Hippocampus forget threshold | 0.3 | hippocampus/store.go |
| Neocortex rate (r) | 0.98 | neocortex/store.go |
| Neocortex keep threshold | 0.6 | neocortex/store.go |
| Weight blending ratio | 0.7 old / 0.3 new | writer/pipeline.go |

Scoring formula: `score = weight * rate^tickCount`

---

## Backlog Tasks

### Phase 0: Project Scaffolding

#### Task 0.1: Initialize Go module and install dependencies

- Initialize a Go module at `custom/memory/` with module path `github.com/opencode/memory`.
- Add dependencies: `modernc.org/sqlite` (pure-Go SQLite driver, no CGO needed, `database/sql` compatible) and `github.com/mark3labs/mcp-go` (MCP SDK with stdin/stdout transport support).
- Create the full directory tree under `internal/`.
- Create a placeholder `main.go` with a `package main` and empty `func main()`.

**Why these choices:** `modernc.org/sqlite` avoids CGO even though CGO is available — keeps cross-compilation trivial. `mark3labs/mcp-go` is the simplest community MCP SDK with built-in `ServeStdio()`.

---

### Phase 1: Core Scoring Engine & Prefrontal Cortex

#### Task 1.1: Define MemoryItem struct and CalculateScore function

File: `internal/pfc/ring.go`

Define a struct with fields: `Content` (string), `Score` (float64), `Weight` (float64), `Rate` (float64), `TickCount` (int), `CreatedAt` (time.Time).

Implement a function `CalculateScore(weight, rate float64, tickCount int) float64` that computes `weight * rate^tickCount` using `math.Pow`.

#### Task 1.2: Implement RingBuffer

File: `internal/pfc/ring.go`

A fixed-capacity in-memory ring buffer. Methods needed:

- **New** — create buffer with given size
- **PushFront** — insert item at index 0; if at capacity, drop the last item (tail eviction)
- **Reuse** — given an item index, new content, and an LLM-assessed weight: move item to front, update its content, blend its weight as `0.7*oldWeight + 0.3*llmWeight`, reset TickCount to 0
- **AdvanceTick** — increment TickCount on all items, recalculate Score; remove (evict) items where Score < 0.5; return a slice of items where Score > 0.7 (promoted copies — the originals remain in the buffer)
- **Len** — return current item count
- **Items** — return a snapshot of all items (serialized for the LLM prompt)
- **GetByIndex** — access item by position

Eviction removes the item from the slice. Promotion returns a copy (originals stay in the buffer — items can coexist in PFC and hippocampus per spec).

#### Task 1.3: Manual verification of scoring mechanics

Write a small standalone script (or `_test.go`) that:
1. Creates items at various weights/rates
2. Advances ticks and checks score decay
3. Confirms eviction at < 0.5 and promotion at > 0.7
4. Confirms weight blending on reuse (70/30)

This is not a formal test suite — just enough to validate correctness during development.

---

### Phase 2: SQLite Persistence

#### Task 2.1: Connection Manager

File: `internal/sqlite/conn.go`

Two functions:

- `OpenWriter(path string) (*sql.DB, error)` — opens a SQLite connection with WAL journal mode, `_txlock=immediate`, `synchronous=NORMAL`, `busy_timeout=5000`, `foreign_keys=ON`. Configure the pool to **1 max open connection** (single writer).
- `OpenReader(path string) (*sql.DB, error)` — same pragmas except no `_txlock`. Pool allows **up to 4 concurrent readers**.

Use `sql.Open("sqlite", dsn)` — the `modernc.org/sqlite` driver registers itself as `"sqlite"`.

#### Task 2.2: Schema Creation

File: `internal/sqlite/schema.go`

Write a function `ApplySchema(writer *sql.DB) error` that creates these tables with `CREATE TABLE IF NOT EXISTS`:

- **hippocampus** — columns: `id` (INTEGER PK AUTOINCREMENT), `content` (TEXT), `score` (REAL), `weight` (REAL), `rate` (REAL DEFAULT 0.95), `tick_count` (INTEGER), `created_at` (TEXT), `updated_at` (TEXT). Indexes on `score` and `updated_at`.

- **neocortex** — columns: `id` (INTEGER PK AUTOINCREMENT), `content` (TEXT), `score` (REAL), `rate` (REAL DEFAULT 0.98), `created_at` (TEXT), `updated_at` (TEXT).

- **neocortex_keywords** — junction table: `neocortex_id` (INTEGER, FK -> neocortex.id ON DELETE CASCADE), `keyword` (TEXT). Composite PK on both columns. Index on `keyword` for fast lookup.

#### Task 2.3: Hippocampus Store

File: `internal/hippocampus/store.go`

A `Store` struct holding writer and reader `*sql.DB` references. Constructor takes both.

Methods needed:
- **Insert** — add a new row, return the ID
- **Get** — fetch a single row by ID
- **ListAboveScore** — return all rows where score > threshold (used for promotion check)
- **Delete** — delete by ID
- **UpdateScore** — update score and tick_count
- **AdvanceTicks** — increment tick_count for all rows, recalculate scores; for rows where score > 0.7, return them as a promotion batch (caller inserts into neocortex); delete rows where score < 0.3

#### Task 2.4: Neocortex Store

File: `internal/neocortex/store.go`

A `Store` struct holding writer and reader `*sql.DB` references.

Methods needed:
- **Insert** — insert into `neocortex` and `neocortex_keywords` in a single transaction; keywords arrive as a `[]string`
- **Get** — fetch row + associated keywords; return both as a struct
- **Delete** — delete from neocortex (FK cascade handles keywords)
- **UpdateScore** — update score
- **DeleteBelowScore** — remove rows where score < threshold, return count

#### Task 2.5: Neocortex Search + Keyword Extraction

File: `internal/neocortex/search.go`

**Hybrid search** (`Search(query string, limit int)`):

1. Parse the query into individual words (lowercase, split on whitespace/punctuation, filter stop words like "the", "a", "is", etc.). Keep the most meaningful 3-5 keywords.
2. **First pass** — keyword match via the `neocortex_keywords` junction table: `SELECT DISTINCT n.* FROM neocortex n JOIN neocortex_keywords nk ON nk.neocortex_id = n.id WHERE nk.keyword IN (...) ORDER BY n.score DESC LIMIT ?`
3. **Second pass (re-rank)** — since v1 has no vector embeddings, re-rank by counting how many query words appear in the content text. Adjusted score = `score * (1 + 0.1 * matchCount)`. Sort by adjusted score.
4. Return top N results.
5. If no keyword matches, fall back to `LIKE '%word%'` on `neocortex.content` for each query word.

**Statistical keyword extraction** (`extractKeywords(text string) []string`):
- Lowercase the text
- Tokenize on whitespace/punctuation
- Filter stop words and words shorter than 3 characters
- Return unique keywords, max 10

This is a fallback for when LLM keyword extraction fails in the writer pipeline.

---

### Phase 3: Writer Pipeline

#### Task 3.1: LLM Caller Interface

File: `internal/writer/pipeline.go`

Define `type LLMFunc func(prompt string) (string, error)`. This abstracts the LLM call so the pipeline can be tested with a stub.

The prompt sent to the LLM should include the current PFC items (their content, scores, weights) plus the latest turn context. The LLM must respond with JSON structured as:

```
{
  "reuse_decisions": [
    {"pfc_index": 0, "updated_content": "...", "new_weight": 0.8}
  ],
  "new_items": [
    {"content": "...", "weight": 0.7, "rate": 0.85}
  ]
}
```

The real implementation can call opencode via `exec.Command` or hit an LLM API directly. A stub returning hardcoded JSON is useful during development.

#### Task 3.2: Pipeline Orchestration

File: `internal/writer/pipeline.go`

A `Pipeline` struct that holds a reference to the PFC ring buffer, hippocampus store, neocortex store, and an LLM caller.

A single method `ProcessTurn(turnContext string) error` orchestrates the full per-turn pipeline:

1. **Build LLM prompt** — serialize current PFC items + new context, call LLM, parse reuse decisions and new items from JSON response.
2. **Process reuse decisions** — for each, call RingBuffer.Reuse() to move to front, update content, blend weight.
3. **Add new items** — for each, calculate initial score (= weight since tickCount=0), call RingBuffer.PushFront().
4. **Advance PFC ticks** — call RingBuffer.AdvanceTick(), which returns promoted items (score > 0.7).
5. **Promote to hippocampus** — insert each promoted item into the hippocampus store.
6. **Advance hippocampus ticks** — call Hippocampus.AdvanceTicks(). For each promoted item (score > 0.7): extract keywords via LLM (fall back to `extractKeywords`), insert into neocortex store.
7. **Forget neocortex items** — call Neocortex.DeleteBelowScore(0.6).

Error handling: all errors are logged to stderr via `log.Printf` but never returned or panicked. Memory is non-critical; failures must not break the agent.

#### Task 3.3: LLM-based keyword extraction

File: `internal/writer/pipeline.go` (or separate file under `internal/neocortex/`)

A function that takes an LLM caller and a content string, prompts: "Extract 3-5 keywords from this text that describe what it is about. Return as a JSON array of strings." Parses the JSON response. On any error, falls back to the statistical `extractKeywords` from Task 2.5.

---

### Phase 4: MCP Server & Reader Tool

#### Task 4.1: Build the MCP server

File: `main.go`

Wires everything together:

- Open the SQLite database (path from `MEMORY_DB_PATH` env var, default `memory.db`)
- Apply schema
- Create a Neocortex store
- Initialize an MCP server with `name: "memory-server"`, `version: "1.0.0"`
- Register a `read_memory` tool: accepts a `query` (string, required) parameter, calls `neocortex.Search(query, 10)`, returns ranked results as formatted text
- Register an optional `write_memory` tool: accepts `content` (string, required) and optional `keywords` (string, comma-separated). If keywords omitted, extracts via LLM or statistical fallback. Inserts into neocortex.
- Start serving on stdin/stdout via the MCP SDK's `ServeStdio()`

Key behaviors:
- All logging goes to stderr (stdout is reserved for MCP JSON-RPC messages)
- Errors in tool handlers return `mcp.NewToolResultError(msg)` — never crash
- The server reads JSON-RPC messages from stdin, writes responses to stdout

#### Task 4.2: Build and smoke-test

```bash
cd /home/ctn/.config/opencode/custom/memory
go build -o memory-server .
```

Test by piping an MCP initialize message:

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}' | ./memory-server
```

Should respond with a JSON-RPC response on stdout. After that, pipe a `tools/list` message to confirm the `read_memory` tool is exposed.

---

### Phase 5: Hook Integration

#### Task 5.1: Register MCP server in opencode config

Edit `/home/ctn/.config/opencode/opencode.jsonc` to add the memory server under `mcp`:

```jsonc
{
    "$schema": "https://opencode.ai/config.json",
    "default_agent": "plan",
    "mcp": {
        "memory-server": {
            "type": "local",
            "command": ["./memory-server"],
            "cwd": "/home/ctn/.config/opencode/custom/memory",
            "enabled": true
        }
    }
}
```

This makes the `read_memory` and `write_memory` tools visible to opencode agents.

#### Task 5.2: Wire session_completed hook

Add to `opencode.jsonc`:

```jsonc
{
    // ... existing config above ...
    "experimental": {
        "hook": {
            "session_completed": [
                {
                    "command": ["/home/ctn/.config/opencode/custom/memory/write-session.sh"]
                }
            ]
        }
    }
}
```

Create `custom/memory/write-session.sh` (executable). For now, it should be a stub that logs to `memory-hook.log`. Later it can perform batch consolidation: query hippocampus for items above threshold, promote them to neocortex, etc. The hook fires once per opencode session.

**Note on future alternatives:** The current opencode config schema does not have a per-turn hook. Two approaches exist:

1. **Option A (implemented here):** Use `session_completed` hook — fires once per session. Suitable for batch consolidation at session boundaries.
2. **Option B (future):** Have the MCP server run the writer pipeline as a background goroutine triggered after every tool call, or have the agent invoke `write_memory` explicitly during its turn. This gives per-turn memory without a framework hook.

---

## End-to-End Verification

1. Build and start `./memory-server` (runs in foreground, reads stdin).
2. From a separate terminal, verify protocol handshake by piping initialize + tools/list messages.
3. Insert a memory via SQLite CLI or `write_memory` tool invocation.
4. Query via `read_memory` — should return the inserted memory ranked by relevance.
5. Restart opencode, verify the memory server appears in the MCP tool list.
6. Complete an opencode session and confirm `memory-hook.log` shows the hook fired.

---

## Project Files Summary

| File | What it contains | Created in |
|---|---|---|
| `go.mod` | Module definition, dependencies | Task 0.1 |
| `go.sum` | Dependency checksums | Task 0.1 |
| `main.go` | MCP server, tool definitions, startup | Task 4.1 |
| `internal/pfc/ring.go` | MemoryItem, RingBuffer, CalculateScore | Tasks 1.1, 1.2 |
| `internal/sqlite/conn.go` | OpenWriter, OpenReader | Task 2.1 |
| `internal/sqlite/schema.go` | ApplySchema, DDL statements | Task 2.2 |
| `internal/hippocampus/store.go` | Hippocampus CRUD, AdvanceTicks | Task 2.3 |
| `internal/neocortex/store.go` | Neocortex CRUD, keyword junction | Task 2.4 |
| `internal/neocortex/search.go` | Hybrid Search, extractKeywords | Task 2.5 |
| `internal/writer/pipeline.go` | Pipeline struct, ProcessTurn, LLMFunc | Tasks 3.1, 3.2, 3.3 |
| `write-session.sh` | Session completion hook script | Task 5.2 |
| `opencode.jsonc` | MCP server config + hooks | Tasks 5.1, 5.2 |
