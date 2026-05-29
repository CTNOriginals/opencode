---
name: terminal-usage
description: Use the bash tool efficiently within agent workflows
---

# Terminal Usage

## Bash Tool Management

Use at most one shell session per agent. Reuse the initial terminal instead of opening additional ones.

## Command Guidelines

Do NOT use `$(cmd)` syntax or backticks for command substitution. Instead, write content to temporary files and reference them, or use direct value passing.

### Examples

**Avoid:**
```bash
# Command substitution
git commit -m "$(echo 'some message')"
gh pr create --body "$(cat file.txt)"
```

**Prefer:**
```bash
# Direct value passing
git commit -m "feat(api): add endpoint"

# File-based approach for multi-line content
gh pr create --title "PR(feature > main): add feature" --body-file temp-pr-body.txt
```
