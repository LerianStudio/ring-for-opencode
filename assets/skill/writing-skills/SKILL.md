---
name: "ring:writing-skills"
description: |
  TDD for process documentation - write test cases (pressure scenarios), watch
  baseline fail, write skill, iterate until bulletproof against rationalization.
license: MIT
compatibility: opencode
metadata:
  trigger: "Creating a new skill, editing an existing skill, skill needs to resist rationalization"
  skip_when: "Writing pure reference skill (API docs), skill has no compliance costs"
---

# Writing Skills

## Related Skills

**Complementary:** ring:testing-skills-with-subagents

---

## Overview

**Writing skills IS Test-Driven Development applied to process documentation.**

You write test cases (pressure scenarios with subagents), watch them fail (baseline behavior), write the skill (documentation), watch tests pass (agents comply), and refactor (close loopholes).

**Core principle:** If you didn't watch an agent fail without the skill, you don't know if the skill teaches the right thing.

**REQUIRED BACKGROUND:** You MUST understand ring:test-driven-development before using this skill. That skill defines the fundamental RED-GREEN-REFACTOR cycle. This skill adapts TDD to documentation.

## What is a Skill?

A **skill** is a reference guide for proven techniques, patterns, or tools. Skills help future agent instances find and apply effective approaches.

**Skills are:** Reusable techniques, patterns, tools, reference guides

**Skills are NOT:** Narratives about how you solved a problem once

## TDD Mapping for Skills

| TDD Concept | Skill Creation |
|-------------|----------------|
| **Test case** | Pressure scenario with subagent |
| **Production code** | Skill document (SKILL.md) |
| **Test fails (RED)** | Agent violates rule without skill (baseline) |
| **Test passes (GREEN)** | Agent complies with skill present |
| **Refactor** | Close loopholes while maintaining compliance |
| **Write test first** | Run baseline scenario BEFORE writing skill |
| **Watch it fail** | Document exact rationalizations agent uses |
| **Minimal code** | Write skill addressing those specific violations |
| **Watch it pass** | Verify agent now complies |
| **Refactor cycle** | Find new rationalizations -> plug -> re-verify |

The entire skill creation process follows RED-GREEN-REFACTOR.

## When to Create a Skill

**Create when:**
- Technique wasn't intuitively obvious to you
- You'd reference this again across projects
- Pattern applies broadly (not project-specific)
- Others would benefit

**Don't create for:**
- One-off solutions
- Standard practices well-documented elsewhere
- Project-specific conventions (put in project docs)

## Skill Types

### Technique
Concrete method with steps to follow (condition-based-waiting)

### Pattern
Way of thinking about problems (flatten-with-flags, test-invariants)

### Reference
API docs, syntax guides, tool documentation

## Directory Structure

`skills/skill-name/SKILL.md` (required) + optional supporting files. Flat namespace.

**Separate files for:** Heavy reference (100+ lines), reusable tools. **Keep inline:** Principles, code patterns (<50 lines).

## SKILL.md Structure

**Frontmatter (YAML):**
- Only fields supported: `name`, `description`, `license`, `compatibility`, `metadata`
- Max 1024 characters for description
- `name`: Use letters, numbers, and hyphens only
- `description`: Third-person, includes BOTH what it does AND when to use it

```markdown
---
name: skill-name-with-hyphens
description: |
  Use when [triggers/symptoms] - [what it does, third person]
license: MIT
compatibility: opencode
metadata:
  trigger: "specific triggers"
  skip_when: "when to skip"
---
# Skill Name
## Overview (1-2 sentences), ## When to Use (symptoms, NOT to use)
## Core Pattern (before/after code), ## Quick Reference (table for scanning)
## Implementation (inline or link), ## Common Mistakes, ## Real-World Impact (optional)
```

## The Iron Law (Same as TDD)

```
NO SKILL WITHOUT A FAILING TEST FIRST
```

This applies to NEW skills AND EDITS to existing skills.

Write skill before testing? Delete it. Start over.
Edit skill without testing? Same violation.

**No exceptions:**
- Not for "simple additions"
- Not for "just adding a section"
- Not for "documentation updates"
- Don't keep untested changes as "reference"
- Don't "adapt" while running tests
- Delete means delete

## RED-GREEN-REFACTOR for Skills

| Phase | Action |
|-------|--------|
| **RED** | Run pressure scenario WITHOUT skill -> document choices/rationalizations verbatim |
| **GREEN** | Write skill addressing specific failures -> verify agent complies |
| **REFACTOR** | Find new rationalizations -> add counters -> re-test until bulletproof |

**REQUIRED SUB-SKILL:** Use ring:testing-skills-with-subagents for pressure scenarios, pressure types, hole-plugging, meta-testing.

## Bulletproofing Skills Against Rationalization

Skills that enforce discipline (like TDD) need to resist rationalization. Agents are smart and will find loopholes when under pressure.

### Close Every Loophole Explicitly

Don't just state rule - forbid specific workarounds:
- **BAD:** `Write code before test? Delete it.`
- **GOOD:** Add `Delete it. Start over.` + explicit `No exceptions:` list

### Address "Spirit vs Letter" Arguments

Add foundational principle early:

```markdown
**Violating the letter of the rules is violating the spirit of the rules.**
```

### Build Rationalization Table

Capture rationalizations from baseline testing. Every excuse agents make goes in the table:

```markdown
| Excuse | Reality |
|--------|---------|
| "Too simple to test" | Simple code breaks. Test takes 30 seconds. |
| "I'll test after" | Tests passing immediately prove nothing. |
```

### Create Red Flags List

Make it easy for agents to self-check when rationalizing:

```markdown
## Red Flags - STOP and Start Over

- Code before test
- "I already manually tested it"
- "This is different because..."

**All of these mean: Delete code. Start over with TDD.**
```

## Skill Creation Checklist (TDD Adapted)

**Use todowrite tool for each phase.**

| Phase | Requirements |
|-------|--------------|
| **RED** | 3+ pressure scenarios, run WITHOUT skill, document rationalizations verbatim |
| **GREEN** | Name (letters/numbers/hyphens), YAML frontmatter (<1024 chars), description starts "Use when...", third person, keywords, address baseline failures, one excellent example, verify compliance |
| **REFACTOR** | New rationalizations -> add counters, build rationalization table, create red flags, re-test |
| **Quality** | Flowchart only if non-obvious, quick ref table, common mistakes, no narrative |
| **Deploy** | Commit and push, consider contributing PR |

## The Bottom Line

**Creating skills IS TDD for process documentation.**

Same Iron Law: No skill without failing test first.
Same cycle: RED (baseline) -> GREEN (write skill) -> REFACTOR (close loopholes).
Same benefits: Better quality, fewer surprises, bulletproof results.

If you follow TDD for code, follow it for skills. It's the same discipline applied to documentation.
