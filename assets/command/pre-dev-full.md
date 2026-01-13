---
name: "ring:pre-dev-full"
description: Complete 9-gate pre-dev workflow for large features (≥2 days)
agent: plan
subtask: false
---

Use the `ring:pre-dev-full` skill to plan this feature.

This command initiates the **Full Track** pre-development workflow (9 gates):
1. **Research** (Gate 0)
2. **PRD** (Gate 1)
3. **Feature Map** (Gate 2)
4. **TRD** (Gate 3)
5. **API Design** (Gate 4)
6. **Data Model** (Gate 5)
7. **Dependency Map** (Gate 6)
8. **Tasks** (Gate 7)
9. **Subtasks** (Gate 8)

Argument hint: `[feature-name]`

## Execution

1. Analyze the user's request (feature name, context).
2. Load and execute the `ring:pre-dev-full` skill.
3. Follow the skill's instructions strictly.
