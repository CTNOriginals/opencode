---
name: create-pull-request
description: Create GitHub pull requests using the gh CLI with standardized title format
---

# Create Pull Request

## PR Title Format

For pull request titles, use this format:

```
PR(current > target): description
```

- **current**: The branch being merged
- **target**: The branch being merged into
- **description**: Follows the conventional commit description format

## PR Title Examples

```
PR(infrastructure > main): add core infrastructure components
PR(feature/api > develop): implement user authentication
PR(hotfix/memory-leak > main): fix goroutine leak in event handler
```

## PR Body Creation

Write body text to a temporary file using the Write tool, then reference it with `--body-file`. Clean up temporary files after creation.

## Example

```bash
# Write to file first, then use --body-file
gh pr create --title "PR(feature > main): add feature" --body-file temp-pr-body.txt
```
