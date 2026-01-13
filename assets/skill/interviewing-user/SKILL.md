---
name: "ring:interviewing-user"
description: |
  Proactive requirements gathering - systematically interviews the user to uncover
  ambiguities, preferences, and constraints BEFORE implementation begins.
license: MIT
compatibility: opencode
metadata:
  trigger: "Claude detects ambiguity, multiple valid paths exist, task involves architecture decisions"
  skip_when: "Requirements crystal clear, detailed specifications provided, following existing plan"
  sequence_before: "ring:brainstorming, ring:writing-plans"
---

# Interviewing User for Requirements

## Related Skills

**Similar:** ring:brainstorming
**Uses:** ring:doubt-triggered-questions

---

## Overview

Proactively surface and resolve ambiguities by systematically interviewing the user BEFORE implementation begins. This prevents wasted effort from incorrect assumptions.

**Core principle:** It's better to ask 5 questions upfront than to rewrite code 3 times.

**Announce at start:** "I'm using the interviewing-user skill to gather requirements before we begin."

## Quick Reference

| Phase | Key Activities | Output |
|-------|---------------|--------|
| **1. Context Analysis** | Analyze task, identify ambiguities | Ambiguity inventory |
| **2. Question Clustering** | Group questions by category | Prioritized question list |
| **3. Structured Interview** | Ask questions using question tool | User responses |
| **4. Understanding Summary** | Synthesize and confirm | Validated Understanding |
| **5. Proceed or Iterate** | User confirms or clarifies | Green light to proceed |

## The Process

Copy this checklist to track progress:

```
Interview Progress:
- [ ] Phase 1: Context Analysis (ambiguities identified)
- [ ] Phase 2: Question Clustering (questions prioritized)
- [ ] Phase 3: Structured Interview (questions asked and answered)
- [ ] Phase 4: Understanding Summary (presented to user)
- [ ] Phase 5: Proceed or Iterate (user confirmed)
```

### Phase 1: Context Analysis

**BEFORE asking any questions**, analyze:

1. **What the user explicitly stated** - Extract concrete requirements
2. **What the codebase implies** - Patterns, conventions, existing solutions
3. **What remains ambiguous** - Gaps between stated and implied
4. **What decisions I must make** - Architecture, behavior, constraints

**Create an Ambiguity Inventory:**

```
Ambiguity Inventory:
- Architecture: [list unclear architectural decisions]
- Behavior: [list unclear behavioral requirements]
- Constraints: [list unclear constraints or limitations]
- Preferences: [list unclear user preferences]
- Integration: [list unclear integration points]
```

### Phase 2: Question Clustering

Group questions by category and prioritize:

| Priority | Category | Criteria |
|----------|----------|----------|
| **P0** | Blocking | Cannot proceed without answer |
| **P1** | Architecture | Affects overall structure |
| **P2** | Behavior | Affects user-facing functionality |
| **P3** | Preferences | Affects style, not correctness |

**Question Budget:**
- **Maximum 4 questions per question tool call** (tool limitation)
- **Maximum 3 rounds of questions** (respect user's time)
- **Prefer fewer, higher-quality questions**

### Phase 3: Structured Interview

Use `question tool` tool with well-structured options:

**Question Quality Checklist:**
- [ ] Shows what I already know (evidence of exploration)
- [ ] Explains why I'm uncertain (the genuine conflict)
- [ ] Provides 2-4 concrete options with descriptions
- [ ] Options are mutually exclusive or clearly labeled as multi-select

### Phase 4: Understanding Summary

After gathering responses, synthesize into a **Validated Understanding**:

```markdown
## Validated Understanding

### What We're Building
[1-2 sentence summary of the goal]

### Key Decisions Made
| Decision | Choice | Rationale |
|----------|--------|-----------|
| [Topic] | [Selected option] | [Why this was chosen] |

### Constraints Confirmed
- [Constraint 1]
- [Constraint 2]

### Out of Scope (Explicit)
- [Thing we're NOT doing]

### Assumptions (If Any)
- [Assumption]: [What would invalidate this]
```

**Present this to the user for confirmation.**

### Phase 5: Proceed or Iterate

**Confirmation Gate:**

Understanding is NOT confirmed until user explicitly says:
- "Confirmed" / "Correct" / "That's right"
- "Proceed" / "Let's do it" / "Go ahead"
- "Yes" (in response to "Is this correct?")

**These do NOT mean confirmation:**
- Silence
- "Interesting" / "I see"
- Questions about the summary
- "What about X?" (that's requesting changes)

**If not confirmed:** Return to Phase 3 with targeted follow-up questions.

## When to Auto-Trigger This Skill

Claude SHOULD invoke this skill automatically when:

1. **Ambiguity count > 3** - More than 3 unclear decisions
2. **Architecture choice unclear** - Multiple valid patterns, no codebase precedent
3. **User request is high-level** - "Build me X" without specifics
4. **Previous implementation was rejected** - Indicates misunderstanding
5. **Task spans multiple domains** - Frontend + backend + infrastructure

Claude should NOT auto-trigger when:
- Task is a simple bug fix with clear reproduction
- User provided detailed specifications
- Following an existing plan
- Single question would suffice

## Anti-Patterns

| Anti-Pattern | Why It's Wrong | Correct Approach |
|--------------|----------------|------------------|
| Asking without exploring first | Wastes user's time | Explore codebase THEN ask |
| Open-ended questions only | Hard to answer, vague responses | Provide concrete options |
| Too many questions at once | Overwhelming | Max 4 per round, max 3 rounds |
| Asking about things user already said | Shows you weren't listening | Re-read conversation first |
| Asking preferences when conventions exist | Codebase already answers | Follow existing patterns |
| Skipping summary phase | User can't correct misunderstandings | Always present Validated Understanding |

## Integration with Other Skills

| Skill | Relationship |
|-------|--------------|
| `ring:doubt-triggered-questions` | Use for single questions during work; use ring:interviewing-user for systematic upfront gathering |
| `ring:brainstorming` | Interview first to gather requirements, THEN brainstorm solutions |
| `ring:writing-plans` | Interview first to clarify scope, THEN create plan |

## Exit Criteria

Interview is complete when ALL of these are true:

- [ ] All P0 (blocking) questions answered
- [ ] All P1 (architecture) questions answered
- [ ] Validated Understanding presented
- [ ] User explicitly confirmed understanding
- [ ] No remaining ambiguities that affect correctness

## Key Principles

| Principle | Application |
|-----------|-------------|
| **Explore before asking** | 30 seconds of exploration can save a question |
| **Structured choices** | Use question tool with 2-4 concrete options |
| **Show your work** | Include what you found and why you're uncertain |
| **Respect time** | Max 3 rounds, max 4 questions per round |
| **Confirm understanding** | Always present summary for validation |
| **Iterate if needed** | Unclear confirmation = ask follow-up |
