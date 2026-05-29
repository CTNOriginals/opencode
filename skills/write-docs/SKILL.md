---
name: write-docs
description: Write maintainable documentation by minimizing specific file references
---

# Write Maintainable Documentation

Avoid file-specific references in permanent documentation unless <= 2 are critically relevant.

## Strategy

**Instead of:**
```markdown
See src/auth/handlers.go, src/auth/middleware.go, src/auth/tokens.go, src/auth/validation.go
```

**Use:**
```markdown
See auth-related files in src/auth/
```

Specific references create maintenance debt. When files move, rename, or refactor, permanent docs must be updated. General directions survive changes.

## Exception

INDEX.md and navigation documents are exempt - their purpose is to reference and organize files. Use specific references freely there.
