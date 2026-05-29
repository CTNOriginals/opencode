---
name: commit
description: Stage and commit changes using conventional commits format. Use when the user asks to commit, save changes to git, or create a commit.
---

# Commit

Stage and commit changes using conventional commits format. Apply the conventional-commits skill for message format details.

## Grouping

Commit changes in logical chunks:
- Group files in each commit so they relate to each other
- Prefer small commits with few files per commit
- Exception: bulk changes that affect many files in the same way may be one commit

## Self-contained commits

Each commit must be self-contained and able to function on its own. Do not commit anything that loses relevant context because that context is in another commit.

- If Y references or depends on something defined in X, include both X and Y in the same commit, or commit X before Y
- Do not commit Y without X when Y would be broken or meaningless without it
- Order commits so that definitions land before or with their references

## Workflow

1. Run `git status` to see what changed.
2. Stage related files together: `git add <paths>` (avoid `git add -A` when changes span multiple logical commits).
3. Commit with message: `git commit -m "type(scope): description"`.
4. For multi-line body: `git commit -m "type(scope): description" -m "body line 1" -m "body line 2"`.
5. Repeat steps 2-4 for each logical chunk until all changes are committed.
