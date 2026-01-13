---
name: "ring:exploring-codebase"
description: |
  Autonomous two-phase codebase exploration - first discovers natural perspectives
  (layers, components, boundaries), then dispatches adaptive deep-dive explorers
  based on what was discovered. Synthesizes findings into actionable insights.
license: MIT
compatibility: opencode
metadata:
  trigger: "Need to understand how feature/system works, starting on unfamiliar codebase, planning multi-layer changes"
  skip_when: "Pure reference lookup, checking if file exists, reading known error location"
---

# Autonomous Two-Phase Codebase Exploration

## Related Skills

**Sequence after:** ring:brainstorming
**Sequence before:** ring:writing-plans, ring:executing-plans

---

## Overview

Traditional exploration assumes structure upfront or explores sequentially. This skill takes an autonomous two-phase approach: **discover** the natural perspectives of the codebase first, then **deep dive** into each discovered perspective with targeted explorers.

**Core principle:** Let the codebase reveal its own structure, then explore each structure element thoroughly with adaptive parallel agents.

**MANDATORY ANNOUNCEMENT at start:**

"I'm using the exploring-codebase skill to autonomously discover and explore the codebase structure.

Before proceeding, I've checked the Red Flags table and confirmed:
- [X] Production pressure makes me WANT to skip discovery -> Using skill anyway
- [X] I think I 'already know' the structure -> Discovery will validate assumptions
- [X] This seems like a simple question -> Location without context is incomplete
- [X] Colleague gave me high-level info -> Discovery finds what they forgot

The skill's core principle: **When pressure is highest, systematic approach matters most.**"

## Red Flags: When You're About to Make a Mistake

**STOP and use this skill if you catch yourself thinking:**

| Red Flag Thought | What It Means | Do This Instead |
|------------------|---------------|-----------------|
| "I already know this architecture" | Dunning-Kruger | Run discovery to validate assumptions |
| "Grep is faster for this simple question" | Optimizing for feeling productive | One exploration > multiple follow-ups |
| "Production is down, no time for process" | Panic mode | High stakes demand MORE rigor |
| "Colleague told me the structure" | Trusting abstractions | Discovery finds what they forgot |
| "Being pragmatic means skipping this" | Conflating speed with value | Real pragmatism = doing it right |
| "This is overkill for..." | Underestimating complexity | Incomplete understanding compounds |
| "I'll explore progressively if I get stuck" | Reactive vs proactive | Discovery prevents getting stuck |
| "Let me just quickly check..." | Ad-hoc investigation trap | Systematic > ad-hoc |

**If 2+ red flags triggered: YOU NEED THIS SKILL.**

## The Two-Phase Flow

### Phase 1: Discovery Pass (Meta-Exploration)
**Goal:** Understand "What IS this codebase?"

Launch 3-4 **discovery agents** to identify:
- Architecture pattern (hexagonal, layered, microservices, etc.)
- Major components/modules
- Natural boundaries and layers
- Organization principles
- Key technologies and frameworks

**Output:** Structural map of the codebase

### Phase 2: Deep Dive Pass (Adaptive Exploration)
**Goal:** Understand "How does [target] work in each discovered area?"

Based on Phase 1 discoveries, launch **N targeted explorers** (where N adapts):
- One explorer per discovered perspective/component/layer
- Each explorer focuses on the target within their scope
- Number and type of explorers match codebase structure

**Output:** Comprehensive understanding of target across all perspectives

## When to Use

**Decision flow:**
- Need codebase understanding? -> Is it trivial (single file/function)? -> Yes = Use Read/Grep directly
- No -> Is it unfamiliar territory or spans multiple areas? -> Yes = **Two-phase exploration**
- Are you about to make changes spanning multiple components? -> Yes = **Two-phase exploration**

**Use when:**
- Understanding how a feature works in an unfamiliar codebase
- Starting work on new component/service
- Planning architectural changes
- Need to find where to implement new functionality
- User asks "how does X work?" for complex X in unknown codebase

**Don't use when:**
- Pure reference lookup: "What's the signature of function X?"
- File existence check: "Does utils.go exist?"
- Reading known error location: "Show me line 45 of errors.go"

## Process

Copy this checklist to track progress:

```
Two-Phase Exploration Progress:
- [ ] Phase 0: Scope Definition (exploration target identified)
- [ ] Phase 1: Discovery Pass (structure discovered - 3-4 agents)
- [ ] Phase 2: Deep Dive Pass (N adaptive explorers launched)
- [ ] Phase 3: Result Collection (all agents completed)
- [ ] Phase 4: Synthesis (discovery + deep dive integrated)
- [ ] Phase 5: Action Recommendations (next steps identified)
```

## Phase 0: Scope Definition

**Step 0.1: Identify Exploration Target**

From user request, extract:
- **Core subject:** What feature/system/component to explore?
- **Context clue:** Why are they asking? (planning change, debugging, learning)
- **Depth needed:** Surface understanding or comprehensive dive?

**Step 0.2: Set Exploration Boundaries**

Define scope to keep agents focused:
- **Include:** Directories/components relevant to target
- **Exclude:** Build config, vendor code, generated files (unless specifically needed)
- **Target specificity:** "account creation" vs "entire onboarding service"

## Phase 1: Discovery Pass (Meta-Exploration)

**Goal:** Discover the natural structure of THIS codebase

**Step 1.1: Launch Discovery Agents in Parallel**

**CRITICAL: Single message with 3-4 Task tool calls**

Dispatch discovery agents simultaneously:

```
Task(subagent_type="Explore", description="Architecture discovery",
     prompt="[Architecture Discovery prompt]")

Task(subagent_type="Explore", description="Component discovery",
     prompt="[Component Discovery prompt]")

Task(subagent_type="Explore", description="Layer discovery",
     prompt="[Layer Discovery prompt]")

Task(subagent_type="Explore", description="Organization discovery",
     prompt="[Organization Discovery prompt]")
```

## Phase 2: Deep Dive Pass (Adaptive Exploration)

**Goal:** Explore target within each discovered perspective

Based on Phase 1 discoveries, launch **N targeted explorers** (where N = number of discovered perspectives)

**CRITICAL: Single message with N Task tool calls**

## Phase 3-5: Collection, Synthesis, Recommendations

- Collect all agent results
- Integrate discovery + deep dive findings
- Provide context-aware next steps based on user goal

## Common Mistakes

| Bad | Good |
|-----|------|
| Skip discovery, assume structure | Always run Phase 1 discovery first |
| Use same deep dive agents for all codebases | Adapt Phase 2 agents based on Phase 1 |
| Accept vague discoveries | Require file:line evidence |
| Run explorers sequentially | Dispatch all in parallel (per phase) |
| Skip synthesis step | Always integrate discovery + deep dive |
| Provide raw dumps | Synthesize into actionable guidance |
| Use for single file lookup | Use Read/Grep instead |

## Integration with Other Skills

| Skill | When to use together |
|-------|----------------------|
| **ring:brainstorming** | Use ring:exploring-codebase in Phase 1 (Understanding) to gather context |
| **ring:writing-plans** | Use ring:exploring-codebase before creating implementation plans |
| **ring:executing-plans** | Use ring:exploring-codebase if plan execution reveals gaps |

## Key Principles

| Principle | Application |
|-----------|-------------|
| **Discover, then dive** | Phase 1 discovery informs Phase 2 exploration |
| **Adaptive parallelization** | Number and type of agents matches structure |
| **Evidence-based** | All discoveries backed by file:line references |
| **Autonomous** | Codebase reveals its own structure |
| **Synthesis required** | Raw outputs must be integrated |
| **Action-oriented** | Always end with next steps |
| **Quality gates** | Verify each phase before proceeding |

## Required Patterns

This skill uses these universal patterns:
- **State Tracking:** See `shared-patterns/state-tracking.md`
- **Failure Recovery:** See `shared-patterns/failure-recovery.md`
- **Todo list:** See `shared-patterns/todowrite-integration.md`

Apply ALL patterns when using this skill.
