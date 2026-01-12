# Universal State Tracking Pattern

Add this section to any multi-step skill:

## State Tracking (MANDATORY)

Create and maintain a status comment:

```
SKILL: [skill-name]
PHASE: [current phase/step]
COMPLETED: [✓ list what's done]
NEXT: [→ what's next]
EVIDENCE: [last verification output]
BLOCKED: [any blockers]
```

**Update after EACH phase/step.**

Example:
```
SKILL: test-driven-development
PHASE: 2 - GREEN (Implementation)
COMPLETED: ✓ Test written ✓ Test failure verified
NEXT: → Write minimal implementation
EVIDENCE: Test fails with "NameError: function not defined"
BLOCKED: None
```

This comment should be included in EVERY response while using the skill.
