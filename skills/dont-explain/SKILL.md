---
name: dont-explain
description: Execute task without explanation or summary to reduce token usage. Skip preamble, reasoning, and recap.
---

# Don't Explain

Execute the task without explanation or summary. Skip preamble, reasoning, and recap.

## Usage

When invoked, perform the work silently. Output only the result - no "I will...", no "Here's what I did", no "Summary:".

**Example:**

**User:** Fix linting errors in src/main.ts

**Agent:** [Makes changes, outputs only result]

This saves tokens by eliminating redundant narrative.
