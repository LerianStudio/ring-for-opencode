---
name: "ring:pre-dev-feature"
description: Lightweight 4-gate pre-dev workflow for small features (<2 days)
agent: plan
subtask: false
---

Use the `ring:pre-dev-feature` skill to plan this feature.

This command initiates the **Small Track** pre-development workflow (4 gates):
1. **Research** (Gate 0)
2. **PRD** (Gate 1)
3. **TRD** (Gate 2)
4. **Tasks** (Gate 3)

Argument hint: `[feature-name]`

## Execution

1. Analyze the user's request (feature name, context).
2. Load and execute the `ring:pre-dev-feature` skill.
3. Follow the skill's instructions strictly.
