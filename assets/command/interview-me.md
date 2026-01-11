---
description: Proactive requirements gathering through structured user interview
agent: build
subtask: false
---

Surface ambiguities and gather requirements BEFORE implementation begins. Instead of you anticipating what might be misunderstood, proactively interview to build a complete picture.

## Process

### 1. Context Analysis
Analyze the task/topic and identify ambiguities.

### 2. Question Clustering
Group questions by priority:
- Blocking questions first
- Architecture decisions
- Behavior specifications
- Preferences last

### 3. Structured Interview
Ask questions using structured choices (2-4 options each):
- Maximum 4 questions per round
- Maximum 3 rounds total
- Fewer is better - explore before asking

### 4. Understanding Summary
Present a "Validated Understanding" document:

```markdown
## Validated Understanding

### What We're Building
[1-2 sentence summary]

### Key Decisions Made
| Decision | Choice | Rationale |
|----------|--------|-----------|

### Constraints Confirmed
- [List of constraints]

### Out of Scope (Explicit)
- [Things we're NOT doing]
```

### 5. Confirmation
User must explicitly confirm before proceeding.

## When to Use

| Scenario | Use? |
|----------|------|
| New feature with vague requirements | Yes |
| Wrong assumptions made previously | Yes |
| Multiple valid approaches, unclear which | Yes |
| Simple bug fix with clear reproduction | No |
| Following existing detailed plan | No |

$ARGUMENTS

Optionally specify the topic to gather requirements for. If omitted, analyze current context.
