## Behavior

- Never offer unsolicited follow-ups like "would you like me to implement X" or "should I also do Y". Only act or answer when the user asks.
- Execute tasks silently. Skip explanatory preamble, "I will...", reasoning summaries, and recaps. Output only the result or the requested deliverable.
- Be concise and direct in all responses.

## Agent Hierarchy

When delegating work to subagents, use the agent best suited for the task:

| Agent | Use when |
|-------|----------|
| manager | Overseeing entire projects; coordinates multiple project-leaders for large initiatives |
| project-leader | Complex multi-step tasks requiring planning, delegation, and end-to-end oversight |
| researcher | Investigating how something works, finding where code lives, gathering context |
| documenter | Creating or updating project docs |
| developer | Building features, adding packages, extending the codebase |
| tester | Running tests, analyzing failures, fixing test issues |
| verifier | Validating completed work, confirming tests pass |
| reviewer | Reviewing PRs, changes, or when the user asks for a code review |
| critic-* | Arguing against proposals from specific angles (Pessimist, Security, Maintainability, Simplicity, Edge-Cases, Wildcard, Moderator). Used by the reflect skill. |

## Shared Instructions

All agents apply these instructions:

- Follow the rules in this file where applicable.
- Apply skills when working on relevant tasks. Skills provide direct guidance. For example, apply the `compose-prompt` skill when writing prompts, the `commit` skill when committing changes, or the `reflect` skill when stress-testing proposals.
- Supply all required context in the task prompt when delegating to subagents. Include file contents, code snippets, error messages, and background information. Instruct subagents to use this context instead of reading files.
- Delegate documentation updates to the documenter subagent. Do not handle docs yourself unless you are the documenter.
- Use the explore subagent for broad codebase searches; use the researcher subagent for deeper investigation and synthesis.
- Match the code style of the surrounding project. For this Neovim config, that means tabs for indentation (check existing files before writing new ones).

## Go Conventions

When working with Go files, apply the following conventions:

### Variable Declarations

Prefer `var` over `:=` for clarity and readability.

**Exception:** Use `:=` in `for range` loops for loop-scoped variables:
```go
for i, val := range items {
}
```

### Variable Names

Use descriptive names; avoid single characters or combinations thereof.

**Exception:** Common loop indices like `i` and `j` for loop counters.

### Type Declaration Prefixes

Prefix type declarations with a letter indicating the kind of type:
- **Type alias:** `T` (e.g. `TUserID`, `TStatus`)
- **Struct:** `S` (e.g. `SUser`, `SConfig`)
- **Map:** `M` (e.g. `MUserByID`, `MConfig`)
- **Interface:** `I` (e.g. `IHandler`, `IRepository`)

```go
type SUser struct {
    ID   TUserID
    Name string
}
```

### Package Entry Point

Name the main entry point file the same as the package for easy location.
**Example:** For an `events` package, the entry point is `events/events.go`.

### Plain Text Conventions

Use common ASCII characters in all code, comments, docs, and commit messages.
- No em-dashes or en-dashes; use a normal hyphen (`-`)
- No special unicode; use ASCII equivalents for punctuation
