# Ring Architecture Optimization Implementation Plan

> **For Agents:** REQUIRED SUB-SKILL: Use executing-plans to implement this plan task-by-task.

**Goal:** Reduce ~40% code duplication across reviewer agents, clarify agent/skill boundaries, add missing commands, and standardize composition hierarchy.

**Architecture:** Extract duplicated patterns to shared-patterns/, create a reviewer base template, update agents to reference patterns via explicit "MUST READ" sections, add commands for orphan skills.

**Tech Stack:** Markdown files, YAML frontmatter, Ring plugin architecture

**Global Prerequisites:**
- Environment: macOS/Linux with shell access
- Tools: Text editor, git
- Access: Write access to `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/`
- State: Clean working tree on main branch

**Verification before starting:**
```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git status  # Expected: clean
ls assets/agent/*.md | wc -l  # Expected: 16
ls assets/skill/shared-patterns/*.md | wc -l  # Expected: 16
ls assets/command/*.md | wc -l  # Expected: 16
```

---

## Historical Precedent

**Query:** "ring architecture duplication agents skills"
**Index Status:** Empty (new project)

No historical data available. This is normal for new projects.
Proceeding with standard planning approach.

---

## Executive Summary

### Current State Analysis

| Component | Count | Size Range | Issues |
|-----------|-------|------------|--------|
| Agents | 16 | 4-42KB | 5 reviewers have ~40% duplicated content |
| Skills | 29 | 2-15KB | Some orphaned (no commands) |
| Commands | 16 | 1-3KB | Missing commands for commonly-used skills |
| Shared Patterns | 16 | 0.6-13KB | Exist but not properly referenced |

### Duplication Breakdown (Reviewer Agents)

| Section | Lines/Agent | Duplication % | Action |
|---------|-------------|---------------|--------|
| Model Requirements | ~40 | 90% | Extract to shared pattern |
| Orchestrator Boundary | ~40 | 95% | Already exists - remove inline duplication |
| Blocker Criteria | ~60 | 70% | Keep structure, parameterize |
| Severity Calibration | ~50 | 80% | Reference existing pattern |
| Pass/Fail Criteria | ~30 | 95% | Extract to shared pattern |
| Pressure Resistance | ~50 | 85% | Reference existing pattern |
| Anti-Rationalization | ~80 | 60% | Keep domain-specific, reference universal |
| When Not Needed | ~25 | 90% | Reference existing pattern |
| Time Budget | ~20 | 95% | Extract to shared pattern |
| Remember section | ~20 | 70% | Keep domain-specific only |

**Estimated Reduction:** ~350 lines per reviewer agent (5 agents x 350 = 1,750 lines total)

---

## Phase 1: Extract Missing Shared Patterns

### Task 1.1: Create Pass/Fail Criteria Shared Pattern

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/skill/shared-patterns/reviewer-pass-fail-criteria.md`

**Prerequisites:**
- None

**Step 1: Create the shared pass/fail criteria pattern**

```markdown
# Reviewer Pass/Fail Criteria

**Version:** 1.0.0
**Applies to:** All reviewer agents

---

## Verdict Rules (NON-NEGOTIABLE)

**REVIEW FAILS if:**
- 1 or more Critical issues found (NO EXCEPTIONS)
- 3 or more High issues found (NO EXCEPTIONS)
- Code does not meet domain-specific quality standards

**REVIEW PASSES if:**
- 0 Critical issues (REQUIRED)
- Fewer than 3 High issues (REQUIRED)
- All High issues have clear remediation plan
- Domain-specific requirements satisfied

**NEEDS DISCUSSION if:**
- Major deviations from plan that might be improvements (unclear if intentional)
- Original plan/requirements have issues
- Unclear requirements (CANNOT review without clarity)

---

## Enforcement

**IMPORTANT:** You CANNOT mark PASS if there are Critical issues or 3+ High issues. You CANNOT mark FAIL without documenting specific issues with file:line references.

These criteria are NON-NEGOTIABLE and apply to ALL reviewer agents regardless of domain.
```

**Step 2: Verify the file was created**

Run: `cat /Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/skill/shared-patterns/reviewer-pass-fail-criteria.md | head -20`

**Expected output:**
```
# Reviewer Pass/Fail Criteria

**Version:** 1.0.0
**Applies to:** All reviewer agents
```

**Step 3: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add assets/skill/shared-patterns/reviewer-pass-fail-criteria.md && git commit -m "$(cat <<'EOF'
feat(shared-patterns): add reviewer pass/fail criteria pattern

Extract common pass/fail rules used by all 5 reviewer agents to reduce
duplication. Part of architecture optimization effort.
EOF
)"
```

---

### Task 1.2: Create Time Budget Shared Pattern

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/skill/shared-patterns/reviewer-time-budget.md`

**Prerequisites:**
- None

**Step 1: Create the shared time budget pattern**

```markdown
# Reviewer Time Budget

**Version:** 1.0.0
**Applies to:** All reviewer agents

---

## Standard Time Allocations

| Feature Size | LOC Range | Time Budget |
|--------------|-----------|-------------|
| Simple | < 200 LOC | 10-15 minutes |
| Medium | 200-500 LOC | 20-30 minutes |
| Large | > 500 LOC | 45-60 minutes |

---

## Domain Adjustments

Some review domains require additional time:

| Domain | Adjustment | Reason |
|--------|------------|--------|
| Security | +50% | Thoroughness prevents catastrophic breaches |
| Business Logic | +25% if complex domain | Domain understanding is critical |
| Nil Safety | +25% for large call chains | Tracing requires completeness |

---

## When Time Budget Exceeded

If you cannot complete the review within allocated time:

1. Document what you've reviewed
2. List areas of uncertainty
3. Recommend human review for complex areas
4. Do NOT rush to meet budget - quality over speed

---

## Anti-Rationalization

| Rationalization | Response |
|-----------------|----------|
| "Time's up, ship it" | **NO.** Document incomplete areas, use NEEDS_DISCUSSION |
| "Close enough" | **NO.** Incomplete review = wasted review |
```

**Step 2: Verify the file was created**

Run: `ls -la /Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/skill/shared-patterns/reviewer-time-budget.md`

**Expected output:** File exists with reasonable size

**Step 3: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add assets/skill/shared-patterns/reviewer-time-budget.md && git commit -m "$(cat <<'EOF'
feat(shared-patterns): add reviewer time budget pattern

Extract common time budget guidelines used by all reviewers.
EOF
)"
```

---

### Task 1.3: Create Reviewer Base Template

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/skill/shared-patterns/reviewer-base-template.md`

**Prerequisites:**
- Tasks 1.1, 1.2 completed

**Step 1: Create the base template that documents all required patterns**

```markdown
# Reviewer Base Template

**Version:** 1.0.0
**Purpose:** Documents required patterns and structure for all reviewer agents

---

## Required Pattern Loading

**MANDATORY:** All reviewer agents MUST reference these shared patterns:

| Pattern | File | What It Covers |
|---------|------|----------------|
| Model Requirements | `reviewer-model-requirement.md` | Opus 4.5+ requirement, self-verification |
| Orchestrator Boundary | `reviewer-orchestrator-boundary.md` | REPORT don't FIX principle |
| Blocker Criteria | `reviewer-blocker-criteria.md` | When to STOP and escalate |
| Severity Calibration | `reviewer-severity-calibration.md` | CRITICAL/HIGH/MEDIUM/LOW |
| Pass/Fail Criteria | `reviewer-pass-fail-criteria.md` | Verdict determination rules |
| Pressure Resistance | `reviewer-pressure-resistance.md` | Resist pressure to skip |
| Anti-Rationalization | `reviewer-anti-rationalization.md` | Don't rationalize skipping |
| When Not Needed | `reviewer-when-not-needed.md` | Minimal review conditions |
| Quality Feedback | `reviewer-quality-feedback.md` | High-quality feedback standards |
| Output Schema | `reviewer-output-schema-core.md` | Required output sections |
| Time Budget | `reviewer-time-budget.md` | Time allocation guidelines |

---

## Agent Structure Template

Every reviewer agent MUST follow this structure:

```
1. YAML Frontmatter (description, mode: subagent)
2. # [Domain] Reviewer ([Aspect])
3. ## Your Role
4. ## Shared Patterns (MUST Read) - Table linking to patterns
5. ## Model Requirements - Brief, reference pattern
6. ## Orchestrator Boundary - Brief, reference pattern
7. ## Standards Loading - Domain-specific if needed
8. ## Blocker Criteria - Domain-specific table
9. ## Review Scope - Domain-specific
10. ## Review Checklist - Domain-specific HIGH PRIORITY
11. ## [Domain-Specific Sections] - The meat of the agent
12. ## Severity Calibration - Reference pattern + domain examples
13. ## Pass/Fail Criteria - Reference pattern
14. ## Output Format - Domain-specific template
15. ## When Not Needed - Reference pattern + domain specifics
16. ## Communication Protocol - Domain-specific templates
17. ## Pressure Resistance - Reference pattern
18. ## Anti-Rationalization - Reference pattern + domain table
19. ## Time Budget - Reference pattern
20. ## Remember - Domain-specific key points
21. ## Standards Compliance Report - Required fields
```

---

## Pattern Reference Syntax

Use this format to reference patterns:

```markdown
## Orchestrator Boundary

**HARD GATE:** This reviewer REPORTS issues. It does NOT fix them.

See [reviewer-orchestrator-boundary.md](../skill/shared-patterns/reviewer-orchestrator-boundary.md) for:
- Why reviewers CANNOT edit files
- How orchestrator dispatches fixes
- Anti-rationalization for "I'll just fix it" temptation

[Domain-specific additions here if needed]
```

---

## Domain-Specific Content

Each reviewer agent adds unique value through:

1. **Domain-specific checklists** - What to check in their area
2. **Domain-specific examples** - Code patterns good/bad for their domain
3. **Domain-specific anti-patterns** - What to watch for
4. **Domain-specific severity mapping** - How issues map to severities
5. **Domain-specific output sections** - Additional fields for their domain

**Keep:** All domain expertise
**Remove:** Generic reviewer infrastructure (use patterns instead)
```

**Step 2: Verify creation**

Run: `wc -l /Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/skill/shared-patterns/reviewer-base-template.md`

**Expected output:** ~100 lines

**Step 3: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add assets/skill/shared-patterns/reviewer-base-template.md && git commit -m "$(cat <<'EOF'
feat(shared-patterns): add reviewer base template

Documents required structure and patterns for all reviewer agents.
Enables consistent agent creation and reduces duplication.
EOF
)"
```

---

## Phase 2: Refactor Reviewer Agents

### Task 2.1: Refactor code-reviewer.md

**Files:**
- Modify: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/agent/code-reviewer.md`

**Prerequisites:**
- Phase 1 completed

**Step 1: Read the current file to understand structure**

This task involves significant refactoring. The code-reviewer.md is ~930 lines. We will:

1. Keep the unique domain content:
   - Review Checklist (Plan Alignment, Algorithmic Flow, Code Quality, Architecture, Documentation, Performance, AI Slop Detection)
   - Algorithmic Flow Examples (Mental Walking)
   - Examples of Well-Architected Code
   - Automated Tools Recommendations

2. Replace duplicated sections with references:
   - Model Requirements: Replace with brief + reference
   - Orchestrator Boundary: Replace with brief + reference
   - Blocker Criteria: Keep domain table, reference base
   - Severity Calibration: Keep domain examples, reference base
   - Pass/Fail Criteria: Replace with reference
   - Pressure Resistance: Replace with reference
   - Anti-Rationalization: Keep domain additions, reference base
   - When Not Needed: Replace with reference + domain additions
   - Time Budget: Replace with reference

**Step 2: Create the refactored version**

The refactored agent should follow this structure (showing key changes):

```markdown
---
description: "Foundation Review: Reviews code quality, architecture, design patterns, algorithmic flow, and maintainability. Runs in parallel with business-logic-reviewer and security-reviewer for fast feedback."
mode: subagent
---

# Code Reviewer (Foundation)

You are a Senior Code Reviewer conducting **Foundation** review.

## Your Role

**Position:** Parallel reviewer (runs simultaneously with business-logic-reviewer and security-reviewer)
**Purpose:** Review code quality, architecture, and maintainability
**Independence:** Review independently - do not assume other reviewers will catch issues outside your domain

**Critical:** You are one of five parallel reviewers. Your findings will be aggregated with business logic and security findings for comprehensive feedback.

---

## Shared Patterns (MUST Read)

**MANDATORY:** Before proceeding, load and follow these shared patterns:

| Pattern | Path | What It Covers |
|---------|------|----------------|
| Base Template | [reviewer-base-template.md](../skill/shared-patterns/reviewer-base-template.md) | Required structure for all reviewers |
| Model Requirements | [reviewer-model-requirement.md](../skill/shared-patterns/reviewer-model-requirement.md) | Opus 4.5+ requirement |
| Orchestrator Boundary | [reviewer-orchestrator-boundary.md](../skill/shared-patterns/reviewer-orchestrator-boundary.md) | REPORT don't FIX |
| Blocker Criteria | [reviewer-blocker-criteria.md](../skill/shared-patterns/reviewer-blocker-criteria.md) | When to STOP |
| Severity Calibration | [reviewer-severity-calibration.md](../skill/shared-patterns/reviewer-severity-calibration.md) | Severity levels |
| Pass/Fail Criteria | [reviewer-pass-fail-criteria.md](../skill/shared-patterns/reviewer-pass-fail-criteria.md) | Verdict rules |
| Pressure Resistance | [reviewer-pressure-resistance.md](../skill/shared-patterns/reviewer-pressure-resistance.md) | Resist pressure |
| Anti-Rationalization | [reviewer-anti-rationalization.md](../skill/shared-patterns/reviewer-anti-rationalization.md) | Don't skip |
| When Not Needed | [reviewer-when-not-needed.md](../skill/shared-patterns/reviewer-when-not-needed.md) | Minimal review |
| Quality Feedback | [reviewer-quality-feedback.md](../skill/shared-patterns/reviewer-quality-feedback.md) | Feedback quality |
| Output Schema | [reviewer-output-schema-core.md](../skill/shared-patterns/reviewer-output-schema-core.md) | Output format |
| Time Budget | [reviewer-time-budget.md](../skill/shared-patterns/reviewer-time-budget.md) | Time allocation |
| AI Slop Detection | [ai-slop-detection.md](../skill/shared-patterns/ai-slop-detection.md) | AI code quality |

**You MUST read these patterns. They contain NON-NEGOTIABLE requirements.**

---

## Model Requirements

**HARD GATE:** See [reviewer-model-requirement.md](../skill/shared-patterns/reviewer-model-requirement.md)

This agent requires a capable model for comprehensive code quality analysis including:
- Architecture pattern recognition and SOLID principle verification
- Algorithmic complexity analysis
- Maintainability assessment across multiple files
- Context propagation tracing through call chains

---

## Orchestrator Boundary

**HARD GATE:** See [reviewer-orchestrator-boundary.md](../skill/shared-patterns/reviewer-orchestrator-boundary.md)

**Summary:** You REPORT issues. You DO NOT fix them.

| Your Responsibility | Your Prohibition |
|---------------------|------------------|
| IDENTIFY issues with file:line references | CANNOT use Edit tool |
| CLASSIFY severity (CRITICAL/HIGH/MEDIUM/LOW) | CANNOT use Create tool |
| EXPLAIN problem and impact | CANNOT modify code directly |
| RECOMMEND remediation (show example code) | CANNOT "just fix this quickly" |

---

## Standards Loading

**Status:** Not applicable for this reviewer agent.

This agent reviews using universal software engineering principles (SOLID, DRY, separation of concerns). Language-specific standards are applied by implementation agents.

---

## Blocker Criteria

**Base:** See [reviewer-blocker-criteria.md](../skill/shared-patterns/reviewer-blocker-criteria.md)

**Domain-Specific Decisions:**

| Decision Type | Code Quality Examples | Action |
|---------------|----------------------|--------|
| **Can Decide** | Severity classification, code quality issues, architectural violations, missing tests/docs | Proceed independently |
| **MUST Escalate** | Unclear requirements, conflicting architectural decisions, major plan deviations | STOP and report |
| **CANNOT Override** | Security vulnerabilities = CRITICAL, Data corruption = CRITICAL, 3+ High = FAIL | NON-NEGOTIABLE |

---

## Review Scope

**MANDATORY: Before starting, determine what to review:**

1. Check for planning documents: `PLAN.md`, `requirements.md`, `PRD.md`, `TRD.md`
2. Identify changed files: `git diff` for incremental, full module for comprehensive
3. Understand context: Read plan/requirements FIRST, then implementation

**HARD GATE: If scope is unclear, ask the user before proceeding.**

---

## Review Checklist

**MANDATORY: Work through ALL areas systematically. CANNOT skip any category.**

### 1. Plan Alignment Analysis
- [ ] Compare implementation against planning document
- [ ] Identify deviations from planned approach
- [ ] Assess whether deviations are improvements or problems
- [ ] Verify all planned functionality is implemented
- [ ] Check for scope creep

### 2. Algorithmic Flow & Implementation Correctness HIGH PRIORITY

**"Mental Walking" - Trace execution flow:**

#### Data Flow & Algorithm Correctness
- [ ] Trace data from inputs through processing to outputs
- [ ] Verify transformations are correct and complete
- [ ] Check data reaches all intended destinations
- [ ] Validate algorithm logic matches intent
- [ ] Ensure state transitions in correct order

#### Context Propagation
- [ ] Request/correlation IDs propagated through entire flow
- [ ] User context passed to all operations
- [ ] Transaction context maintained
- [ ] Error context preserved
- [ ] Trace/span context for distributed tracing

#### Codebase Consistency
- [ ] Follows existing patterns
- [ ] Error handling matches conventions
- [ ] Resource cleanup matches patterns
- [ ] Naming conventions consistent

#### Message/Event Distribution
- [ ] Messages sent to all required queues/topics
- [ ] Event handlers properly registered
- [ ] No silent failures in dispatch
- [ ] Acknowledgment/retry logic in place

#### Cross-Cutting Concerns
- [ ] Logging at appropriate points
- [ ] Metrics/monitoring instrumented
- [ ] Feature flags checked appropriately

#### State Management
- [ ] State updates atomic where required
- [ ] State changes properly sequenced
- [ ] Rollback/compensation logic for failures

### 3. Code Quality Assessment
- [ ] Language conventions and style guides
- [ ] Error handling (try-catch, propagation)
- [ ] Type safety
- [ ] Defensive programming (null checks, validation)
- [ ] Code organization (SRP, DRY)
- [ ] Naming conventions
- [ ] No magic numbers

#### Dead Code Detection
- [ ] No `_ = variable` no-op assignments
- [ ] No unused variables or imports
- [ ] No unused type definitions
- [ ] No unreachable code
- [ ] No commented-out code blocks

### 4. Architecture & Design Review
- [ ] SOLID principles followed
- [ ] Separation of concerns
- [ ] Loose coupling
- [ ] Clean integration with existing systems
- [ ] Scalability considerations
- [ ] No circular dependencies

#### Cross-Package Duplication
- [ ] Helper functions not duplicated
- [ ] Shared utilities extracted to common package
- [ ] Test helpers shared via testutil

### 5. Documentation & Readability
- [ ] Functions/methods have comments
- [ ] Complex logic explained
- [ ] Public APIs documented
- [ ] README updated if needed

### 6. Performance & Maintainability
- [ ] No N+1 queries, inefficient loops
- [ ] Memory leaks prevented
- [ ] Appropriate logging
- [ ] Configuration externalized

### 7. AI Slop Detection MANDATORY

**See [ai-slop-detection.md](../skill/shared-patterns/ai-slop-detection.md)**

- [ ] ALL new dependencies verified to exist in registry
- [ ] No morpheme-spliced package names
- [ ] New code matches existing patterns
- [ ] No single-implementation interfaces
- [ ] No "future use" abstractions
- [ ] All changes mentioned in requirements

---

## Domain-Specific Severity Examples

| Severity | Code Quality Examples |
|----------|----------------------|
| **CRITICAL** | Memory leaks, infinite loops, broken core functionality, incorrect state sequencing, critical data flow breaks |
| **HIGH** | Missing error handling, type safety violations, SOLID violations, missing context propagation, incomplete data flow |
| **MEDIUM** | Code duplication, unclear naming, missing documentation, complex logic needing refactoring |
| **LOW** | Style guide deviations, additional test cases, minor refactoring |

---

## Output Format

**MANDATORY: Use this exact structure:**

```markdown
# Code Quality Review (Foundation)

## VERDICT: [PASS | FAIL | NEEDS_DISCUSSION]

## Summary
[2-3 sentences about overall code quality]

## Issues Found
- Critical: [N]
- High: [N]
- Medium: [N]
- Low: [N]

---

## Critical Issues

### [Issue Title]
**Location:** `file.ts:123-145`
**Category:** [Architecture | Quality | Testing | Documentation]

**Problem:** [Clear description]
**Impact:** [Why this matters]
**Example:**
```[language]
// Current problematic code
```
**Recommendation:**
```[language]
// Suggested fix
```

---

## High Issues
[Same format]

## Medium Issues
[More concise]

## Low Issues
[Bullet list]

---

## What Was Done Well
- [Positive observation]
- [Good practice]

---

## Next Steps

**If PASS:** Review complete, findings aggregated with other reviewers
**If FAIL:** Fix issues, re-run all 5 reviewers
**If NEEDS DISCUSSION:** [Specific questions]
```

---

## Algorithmic Flow Examples

[KEEP ALL EXISTING EXAMPLES - these are domain-specific value]

### Example 1: Missing Context Propagation
[Keep existing code example]

### Example 2: Inconsistent Logging Pattern
[Keep existing code example]

### Example 3: Missing Message Distribution
[Keep existing code example]

### Example 4: Incomplete Data Flow
[Keep existing code example]

### Example 5: Incorrect State Sequencing
[Keep existing code example]

### Example 6: Missing Error Context
[Keep existing code example]

---

## Examples of Well-Architected Code

[KEEP ALL EXISTING EXAMPLES]

---

## When Not Needed

**Base:** See [reviewer-when-not-needed.md](../skill/shared-patterns/reviewer-when-not-needed.md)

**Additional for Code Quality:**
- Generated config files MUST be checked for correctness
- Lock file changes MUST verify no vulnerabilities
- Revert commits MUST confirm intentional

---

## Pressure Resistance

**See [reviewer-pressure-resistance.md](../skill/shared-patterns/reviewer-pressure-resistance.md)**

**Core Principle:** "I am here to protect code quality. I will review thoroughly, document findings clearly, and apply criteria consistently. I CANNOT compromise on non-negotiable requirements."

---

## Anti-Rationalization

**Base:** See [reviewer-anti-rationalization.md](../skill/shared-patterns/reviewer-anti-rationalization.md)

**Domain-Specific Additions:**

| Rationalization | Required Action |
|-----------------|-----------------|
| "Code is refactoring, logic unchanged" | **Verify behavior preservation through tracing** |
| "Modern frameworks handle this" | **Verify security features enabled correctly** |

---

## Automated Tools Recommendations

[KEEP ALL EXISTING TOOL RECOMMENDATIONS]

---

## Time Budget

**See [reviewer-time-budget.md](../skill/shared-patterns/reviewer-time-budget.md)**

---

## Remember

1. **MUST do "mental walking"** - Trace execution, verify data reaches destinations
2. **MUST check codebase consistency** - Follow established patterns
3. **MUST be thorough but concise** - Actionable issues with locations
4. **MUST provide examples** - Show problem and solution
5. **MUST review independently** - Cannot assume others catch adjacent issues
6. **MUST be specific** - File:line for EVERY issue

---

## Standards Compliance Report

**Required output fields:**
- **VERDICT:** PASS, FAIL, or NEEDS_DISCUSSION
- **Issues Found:** Severity, file:line, description
- **What Was Done Well:** Positive findings
- **Recommendations:** Fixes with example code
- **Evidence/References:** Links to principles violated
```

**Note:** This is a template showing the refactored structure. The actual refactoring should preserve all domain-specific examples (Algorithmic Flow Examples, Well-Architected Code examples, Automated Tools).

**Step 3: Verify line count reduction**

Run: `wc -l /Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/agent/code-reviewer.md`

**Expected output:** ~550-650 lines (down from ~930)

**Step 4: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add assets/agent/code-reviewer.md && git commit -m "$(cat <<'EOF'
refactor(agent): reduce code-reviewer duplication by referencing shared patterns

- Replace inline Model Requirements with reference
- Replace inline Orchestrator Boundary with reference
- Replace inline Pass/Fail Criteria with reference
- Replace inline Pressure Resistance with reference
- Replace inline Anti-Rationalization base with reference
- Replace inline When Not Needed with reference
- Replace inline Time Budget with reference
- Keep all domain-specific content (checklists, examples, tools)

Reduces agent from ~930 to ~600 lines while preserving functionality.
EOF
)"
```

**If Task Fails:**

1. **File not readable:**
   - Check: `ls -la /Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/agent/code-reviewer.md`
   - Fix: Verify path is correct
   - Rollback: `git checkout -- assets/agent/code-reviewer.md`

2. **Shared pattern references broken:**
   - Check: `ls /Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/skill/shared-patterns/`
   - Fix: Ensure Phase 1 completed first
   - Rollback: `git reset --hard HEAD`

---

### Task 2.2: Refactor business-logic-reviewer.md

**Files:**
- Modify: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/agent/business-logic-reviewer.md`

**Prerequisites:**
- Phase 1 completed
- Task 2.1 completed (follow same pattern)

**Step 1: Apply same refactoring pattern**

Follow the same structure as code-reviewer refactoring:
- Replace duplicated sections with references
- Keep domain-specific content:
  - Mental Execution Protocol
  - Full Context Analysis Requirement
  - Domain Model Correctness checklist
  - Business Rule Implementation checklist
  - Data Consistency & Integrity checklist
  - Financial & Regulatory Correctness checklist
  - Common Business Logic Anti-Patterns
  - Examples of Good Business Logic

**Step 2: Verify line count reduction**

Run: `wc -l /Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/agent/business-logic-reviewer.md`

**Expected output:** ~600-700 lines (down from ~970)

**Step 3: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add assets/agent/business-logic-reviewer.md && git commit -m "$(cat <<'EOF'
refactor(agent): reduce business-logic-reviewer duplication

Apply same pattern as code-reviewer: reference shared patterns,
keep domain-specific Mental Execution Protocol and business examples.
EOF
)"
```

---

### Task 2.3: Refactor security-reviewer.md

**Files:**
- Modify: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/agent/security-reviewer.md`

**Prerequisites:**
- Phase 1 completed
- Task 2.1, 2.2 completed

**Step 1: Apply same refactoring pattern**

Keep domain-specific content:
- OWASP Top 10 Coverage checklist
- Slopsquatting & AI Dependency Hallucination detection
- Authentication & Authorization checklist
- Input Validation & Injection Prevention checklist
- Data Protection & Privacy checklist
- API & Web Security checklist
- Cryptography standards
- Common Vulnerability Patterns examples
- Examples of Secure Code
- Security Tools recommendations

**Step 2: Verify line count reduction**

Run: `wc -l /Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/agent/security-reviewer.md`

**Expected output:** ~650-750 lines (down from ~1020)

**Step 3: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add assets/agent/security-reviewer.md && git commit -m "$(cat <<'EOF'
refactor(agent): reduce security-reviewer duplication

Reference shared patterns, keep OWASP Top 10 checklist, vulnerability
patterns, and security-specific examples.
EOF
)"
```

---

### Task 2.4: Refactor test-reviewer.md

**Files:**
- Modify: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/agent/test-reviewer.md`

**Prerequisites:**
- Phase 1 completed

**Step 1: Apply same refactoring pattern**

Keep domain-specific content:
- Test Anti-Patterns to Detect (9 patterns with examples)
- Edge Case Coverage checklist
- Test Independence checklist
- Assertion Quality checklist
- Mock Appropriateness checklist
- Recommended Test Additions Template

**Step 2: Verify line count reduction**

Run: `wc -l /Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/agent/test-reviewer.md`

**Expected output:** ~350-400 lines (down from ~550)

**Step 3: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add assets/agent/test-reviewer.md && git commit -m "$(cat <<'EOF'
refactor(agent): reduce test-reviewer duplication

Reference shared patterns, keep test anti-patterns and
domain-specific checklists.
EOF
)"
```

---

### Task 2.5: Refactor nil-safety-reviewer.md

**Files:**
- Modify: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/agent/nil-safety-reviewer.md`

**Prerequisites:**
- Phase 1 completed

**Step 1: Apply same refactoring pattern**

Keep domain-specific content:
- Tracing Methodology (4-step process)
- Language-Specific Patterns (Go and TypeScript tables)
- API Response Initialization checklist
- Nil Risk Trace Template
- Go-Specific Examples
- TypeScript-Specific Examples

**Step 2: Verify line count reduction**

Run: `wc -l /Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/agent/nil-safety-reviewer.md`

**Expected output:** ~380-420 lines (down from ~580)

**Step 3: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add assets/agent/nil-safety-reviewer.md && git commit -m "$(cat <<'EOF'
refactor(agent): reduce nil-safety-reviewer duplication

Reference shared patterns, keep nil tracing methodology and
language-specific examples.
EOF
)"
```

---

## Phase 3: Fix Path Inconsistencies

### Task 3.1: Standardize shared pattern paths

**Files:**
- Modify: All files in `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/agent/`
- Modify: All files in `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/skill/*/SKILL.md`

**Prerequisites:**
- Phase 2 completed

**Step 1: Search for inconsistent paths**

Run: `grep -r "shared-patterns" /Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/ | grep -E "\.\./skills/" | head -20`

**Expected output:** Files with `../skills/shared-patterns/` (incorrect)

**Step 2: Fix path references**

The correct path from agents is `../skill/shared-patterns/` (singular "skill").

Run: `find /Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets -name "*.md" -exec grep -l "../skills/shared-patterns" {} \;`

For each file found, replace `../skills/shared-patterns/` with `../skill/shared-patterns/`.

**Step 3: Verify no broken references**

Run: `grep -r "../skills/shared-patterns" /Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/`

**Expected output:** No results

**Step 4: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add -A && git commit -m "$(cat <<'EOF'
fix(paths): standardize shared-patterns path references

Fix inconsistent paths: ../skills/ -> ../skill/ (singular)
EOF
)"
```

---

## Phase 4: Add Missing Commands

### Task 4.1: Create /tdd command

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/command/tdd.md`

**Prerequisites:**
- Skill exists: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/skill/test-driven-development/SKILL.md`

**Step 1: Create the command file**

```markdown
---
name: tdd
description: "Run test-driven development workflow (RED-GREEN-REFACTOR)"
agent: build
---

# /tdd Command

Invokes the test-driven-development skill for RED-GREEN-REFACTOR workflow.

## Usage

```
/tdd [feature description]
```

## What It Does

1. Loads the test-driven-development skill
2. Guides you through:
   - RED: Write failing test first
   - GREEN: Minimal implementation to pass
   - REFACTOR: Clean up while tests pass

## Examples

```
/tdd add user validation
/tdd implement discount calculation
```

## Related Skills

- test-driven-development
- verification-before-completion
```

**Step 2: Verify file created**

Run: `cat /Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/command/tdd.md`

**Expected output:** Command file content

**Step 3: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add assets/command/tdd.md && git commit -m "$(cat <<'EOF'
feat(command): add /tdd command for test-driven development

Provides slash command access to the test-driven-development skill.
EOF
)"
```

---

### Task 4.2: Create /debug command

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/command/debug.md`

**Prerequisites:**
- Skill exists: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/skill/systematic-debugging/SKILL.md`

**Step 1: Create the command file**

```markdown
---
name: debug
description: "Run systematic debugging workflow (investigate -> analyze -> hypothesize -> fix)"
agent: build
---

# /debug Command

Invokes the systematic-debugging skill for structured bug investigation.

## Usage

```
/debug [bug description or error message]
```

## What It Does

1. Loads the systematic-debugging skill
2. Guides through four phases:
   - **Investigate:** Gather evidence, reproduce issue
   - **Analyze:** Pattern analysis, root cause tracing
   - **Hypothesize:** Form and test theories
   - **Fix:** Implement solution with verification

## Examples

```
/debug "TypeError: Cannot read property 'id' of undefined"
/debug users not receiving email notifications
```

## Related Skills

- systematic-debugging
- root-cause-tracing
```

**Step 2: Verify and commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add assets/command/debug.md && git commit -m "$(cat <<'EOF'
feat(command): add /debug command for systematic debugging

Provides slash command access to the systematic-debugging skill.
EOF
)"
```

---

### Task 4.3: Create /trace command

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/command/trace.md`

**Prerequisites:**
- Skill exists: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/skill/root-cause-tracing/SKILL.md`

**Step 1: Create the command file**

```markdown
---
name: trace
description: "Trace bug from error location back through call stack to root cause"
agent: build
---

# /trace Command

Invokes the root-cause-tracing skill for backward call-chain analysis.

## Usage

```
/trace [error location or symptom]
```

## What It Does

1. Loads the root-cause-tracing skill
2. Traces backward through:
   - Error manifestation point
   - Call stack analysis
   - Data flow tracing
   - Original trigger identification

## Examples

```
/trace src/api/handler.go:45 nil pointer
/trace why is user.balance negative
```

## Related Skills

- root-cause-tracing
- systematic-debugging
```

**Step 2: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add assets/command/trace.md && git commit -m "$(cat <<'EOF'
feat(command): add /trace command for root cause tracing

Provides slash command access to the root-cause-tracing skill.
EOF
)"
```

---

### Task 4.4: Create /verify command

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/command/verify.md`

**Prerequisites:**
- Skill exists: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/skill/verification-before-completion/SKILL.md`

**Step 1: Create the command file**

```markdown
---
name: verify
description: "Run verification checks before marking work complete"
agent: build
---

# /verify Command

Invokes the verification-before-completion skill to ensure work is actually done.

## Usage

```
/verify
```

## What It Does

1. Loads the verification-before-completion skill
2. Requires evidence-first completion:
   - Run verification commands
   - Confirm output matches expectations
   - Block completion claims without evidence

## When to Use

- Before claiming a task is complete
- Before committing "finished" work
- Before marking PR as ready for review

## Related Skills

- verification-before-completion
- dev-validation
```

**Step 2: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add assets/command/verify.md && git commit -m "$(cat <<'EOF'
feat(command): add /verify command for completion verification

Provides slash command to ensure work is verified before completion.
EOF
)"
```

---

### Task 4.5: Create /defense command

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/command/defense.md`

**Prerequisites:**
- Skill exists: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/skill/defense-in-depth/SKILL.md`

**Step 1: Create the command file**

```markdown
---
name: defense
description: "Apply defense-in-depth validation pattern to code"
agent: build
---

# /defense Command

Invokes the defense-in-depth skill for multi-layer validation.

## Usage

```
/defense [code path or feature]
```

## What It Does

1. Loads the defense-in-depth skill
2. Validates data at EVERY layer:
   - Input validation at entry points
   - Business rule validation in domain
   - Persistence validation at storage
   - Output validation at exit points

## When to Use

- Adding new data flows
- Reviewing validation coverage
- After security review findings

## Related Skills

- defense-in-depth
- security-reviewer (agent)
```

**Step 2: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add assets/command/defense.md && git commit -m "$(cat <<'EOF'
feat(command): add /defense command for defense-in-depth validation

Provides slash command for multi-layer validation pattern.
EOF
)"
```

---

## Phase 5: Document Architecture Rules

### Task 5.1: Create Architecture Rules Document

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/docs/ARCHITECTURE_RULES.md`

**Prerequisites:**
- Phase 1-4 completed

**Step 1: Create the architecture rules document**

```markdown
# Ring Architecture Rules

This document defines the rules for Ring's component architecture.

---

## Component Hierarchy

```
Commands (thin wrappers)
    ↓ invoke
Skills (orchestration workflows)
    ↓ dispatch
Agents (FAT specialists)
    ↓ reference
Shared Patterns (reusable content)
```

---

## Component Definitions

### Agents

**What they are:** FAT, self-contained specialists (400-900 lines)

**What they do:**
- Execute specific tasks (review, implement, plan)
- Contain domain expertise
- Reference shared patterns for cross-cutting concerns

**Rules:**
- MUST remain specialists with deep domain knowledge
- MUST reference shared patterns (not duplicate)
- MUST have `mode: subagent` in frontmatter
- CANNOT be thin wrappers
- CANNOT orchestrate other agents

**Location:** `assets/agent/<name>.md`

### Skills

**What they are:** Orchestration workflows

**What they do:**
- Coordinate multi-step processes
- Dispatch agents for specialized work
- Manage state across steps

**Rules:**
- MUST orchestrate, not implement
- MUST dispatch agents for specialized work
- MUST have metadata in frontmatter
- CANNOT contain deep domain knowledge (that goes in agents)

**Location:** `assets/skill/<name>/SKILL.md`

### Commands

**What they are:** Thin slash-command wrappers

**What they do:**
- Provide user-friendly access to skills
- Map `/command` to skill invocation

**Rules:**
- MUST be thin (< 50 lines)
- MUST have `agent: build` in frontmatter
- MUST invoke exactly one skill
- CANNOT contain business logic

**Location:** `assets/command/<name>.md`

### Shared Patterns

**What they are:** Reusable content blocks

**What they do:**
- Define cross-cutting concerns
- Provide consistent standards
- Reduce duplication across agents

**Rules:**
- MUST be referenced (not duplicated) by agents
- MUST be versioned
- MUST be self-contained
- CANNOT contain domain-specific content (that stays in agents)

**Location:** `assets/skill/shared-patterns/<name>.md`

---

## Naming Conventions

| Component | Pattern | Example |
|-----------|---------|---------|
| Agent | `<role>.md` | `code-reviewer.md` |
| Skill | `<action>/SKILL.md` | `writing-plans/SKILL.md` |
| Command | `<verb>.md` | `commit.md` |
| Shared Pattern | `<scope>-<topic>.md` | `reviewer-severity-calibration.md` |

---

## Duplication Rules

### What MUST be in Shared Patterns

- Model requirements
- Orchestrator boundary rules
- Severity calibration
- Pass/fail criteria
- Pressure resistance
- Anti-rationalization (universal parts)
- Time budget guidelines

### What MUST stay in Agents

- Domain-specific checklists
- Domain-specific examples
- Domain-specific anti-patterns
- Domain-specific output sections
- Domain-specific severity mappings

---

## Path Reference Rules

**From agents to shared patterns:**
```markdown
[pattern-name.md](../skill/shared-patterns/pattern-name.md)
```

**From skills to shared patterns:**
```markdown
[pattern-name.md](../shared-patterns/pattern-name.md)
```

**Never use:**
- `../skills/` (wrong - use singular `skill`)
- Absolute paths
- URL references for local patterns

---

## Adding New Components

### New Agent Checklist

- [ ] Follows `reviewer-base-template.md` structure (if reviewer)
- [ ] References all required shared patterns
- [ ] Contains domain-specific content only (no duplication)
- [ ] Has proper frontmatter (`description`, `mode: subagent`)
- [ ] Uses correct path references

### New Skill Checklist

- [ ] Has `SKILL.md` in dedicated directory
- [ ] Has proper frontmatter metadata
- [ ] Dispatches agents for specialized work
- [ ] Documents trigger conditions
- [ ] Has corresponding command (if commonly used)

### New Command Checklist

- [ ] Is thin wrapper (< 50 lines)
- [ ] Has `agent: build` in frontmatter
- [ ] Invokes exactly one skill
- [ ] Has usage examples
- [ ] Documents related skills

### New Shared Pattern Checklist

- [ ] Has version number
- [ ] Documents which agents use it
- [ ] Is self-contained
- [ ] Contains cross-cutting concern only
- [ ] No domain-specific content
```

**Step 2: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add docs/ARCHITECTURE_RULES.md && git commit -m "$(cat <<'EOF'
docs: add architecture rules for Ring components

Defines component hierarchy, naming conventions, duplication rules,
and checklists for adding new components.
EOF
)"
```

---

## Phase 6: Code Review Checkpoint

### Task 6.1: Run Code Review

**Prerequisites:**
- All previous phases completed

**Step 1: Dispatch reviewers**

REQUIRED SUB-SKILL: Use requesting-code-review to run all reviewers in parallel.

**Step 2: Handle findings by severity**

- **Critical/High/Medium:** Fix immediately, re-run reviewers
- **Low:** Add `TODO(review):` comments
- **Cosmetic:** Add `FIXME(nitpick):` comments

**Step 3: Proceed when**

- Zero Critical/High/Medium issues remain
- All Low/Cosmetic issues documented in code

---

## Summary of Changes

### Files Created (8)

| File | Purpose |
|------|---------|
| `assets/skill/shared-patterns/reviewer-pass-fail-criteria.md` | Extracted pass/fail rules |
| `assets/skill/shared-patterns/reviewer-time-budget.md` | Extracted time guidelines |
| `assets/skill/shared-patterns/reviewer-base-template.md` | Template for all reviewers |
| `assets/command/tdd.md` | /tdd command |
| `assets/command/debug.md` | /debug command |
| `assets/command/trace.md` | /trace command |
| `assets/command/verify.md` | /verify command |
| `assets/command/defense.md` | /defense command |
| `docs/ARCHITECTURE_RULES.md` | Architecture documentation |

### Files Modified (5)

| File | Change |
|------|--------|
| `assets/agent/code-reviewer.md` | Reduced ~350 lines |
| `assets/agent/business-logic-reviewer.md` | Reduced ~350 lines |
| `assets/agent/security-reviewer.md` | Reduced ~350 lines |
| `assets/agent/test-reviewer.md` | Reduced ~150 lines |
| `assets/agent/nil-safety-reviewer.md` | Reduced ~160 lines |

### Metrics

| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| Total reviewer agent lines | ~4,050 | ~2,700 | ~33% reduction |
| Commands | 16 | 21 | +5 new commands |
| Shared patterns | 16 | 19 | +3 new patterns |
| Path inconsistencies | Multiple | 0 | Fixed |

---

## Verification Checklist

Before marking complete:

- [ ] All shared patterns created and valid markdown
- [ ] All reviewer agents refactored and reference patterns correctly
- [ ] Path inconsistencies fixed (`../skill/` not `../skills/`)
- [ ] All new commands created with proper frontmatter
- [ ] Architecture rules documented
- [ ] Code review completed
- [ ] All commits made with descriptive messages
- [ ] `git log --oneline | head -15` shows ~12-15 new commits
