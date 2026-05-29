---
name: reflect
description: Stress-test proposals via multi-agent adversarial critique. Use when a proposal needs deep scrutiny, adversarial review, or the user asks to "reflect on" a suggestion.
---

# Reflect

Stress-test a proposal by spawning 6 adversarial critic subagents in parallel, consolidating their objections via a debate moderator, and producing a revised proposal with disposition.

## When to Apply

- User asks to "reflect on", "stress-test", "adversarially review", or "red-team" a suggestion or proposal
- You have made a substantive recommendation and need to validate it before delivering
- An important decision needs deep scrutiny from multiple angles

## Setup

1. Extract the proposal to be reflected on. Clarify it if needed.
2. Determine round count:
   - Default: 1 round (produce + moderate once)
   - User can pre-specify: "reflect on that with 3 rounds"
   - User can request more after seeing result: "another round"
3. Prepare a context bundle:
   - The proposal text
   - Relevant code snippets, file paths, and background
   - Constraints, requirements, or assumptions
   - If round 2+: include the moderator's previous consolidated output

## Round Execution

### Step 1: Spawn Critics (parallel)

Use the Task tool to launch all 6 critics simultaneously. Optimize each prompt with the compose-prompt skill first.

| subagent_type | Persona |
|---|---|
| critic-pessimist | Pessimist — failure modes |
| critic-security | Security — vulnerabilities |
| critic-maintainability | Maintainability — technical debt |
| critic-simplicity | Simplicity — over-engineering |
| critic-edgecases | Edge-Case — boundary conditions |
| critic-wildcard | Wildcard — blind spots |

Each critic prompt must include:
- The proposal
- The context bundle
- Round number / total rounds
- For round 2+: the moderator's consolidated table and instructions

### Step 2: Moderate

Collect all critic outputs. Spawn the debate moderator (critic-moderator) with:
- The proposal
- All 6 objection tables
- Round number / total rounds

### Step 3: Iterate or Finalize

- If more rounds remain: go to Step 1, passing the moderator's consolidated table and instructions to each critic
- If final round: proceed to Final Output

### Step 4: Produce Final Output

Synthesize the moderator's final consolidated table into:

```
## Reflection Results

### Objections Disposition
| # | Objection | Severity | Source | Disposition | Rationale |
|---|-----------|----------|--------|-------------|-----------|
| 1 | ... | Critical | Security | Accepted | Redesigning auth flow |
| 2 | ... | Medium | Simplicity | Dismissed | Future extensibility required |

### Revised Proposal
[revised version with changes highlighted]

### Changes Made
- change 1
- change 2

### Unresolved Objections
- objections dismissed with reasoning
```

## Constraints

- Always apply the compose-prompt skill to optimize critic prompts before spawning
- Never edit files — critics and moderator are read-only
- Keep context bundles concise and relevant
- If user requests specific critics ("skip Simplicity, add Performance"), adjust the roster accordingly
- Critics and moderator output table-only format (Dont-Explain)
