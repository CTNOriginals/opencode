---
name: conventional-commits
description: Write standardized commit messages using the Conventional Commits specification
---

# Conventional Commits

Use Conventional Commits for all commit messages.

## Format

```
type(scope): description

[optional body]

[optional footer(s)]
```

## Types

- **feat**: New feature
- **fix**: Bug fix
- **docs**, **style**, **refactor**, **perf**, **test**, **build**, **ci**, **chore**: Other changes

## Scope

Mandatory. Use a compact noun or path for context (e.g., `feat(api):` or `docs(agents):`). Keep scope short.

## Breaking Changes

Use `!` after type/scope: `feat(api)!: require Node 8+ for ES2015` or add footer: `BREAKING CHANGE: description`.

## Title Format

Write descriptions in "why" manner: focus on benefit, outcome, or reason - not implementation detail.

**Hard rule:** The commit title MUST NOT exceed 50 characters.

## Examples

```
feat(auth): enable users to sign in
fix(api): prevent duplicate responses from concurrent requests
docs(agents): add agent documentation
feat(api)!: require Node 8+ for ES2015 support
```
