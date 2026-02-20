# Reviewer Model Requirement

**Version:** 1.0.0
**Applies to:** All reviewer agents (ring:code-reviewer, ring:business-logic-reviewer, ring:security-reviewer, ring:test-reviewer, ring:nil-safety-reviewer)

---

## Model Requirement: Claude Opus 4.5+

**HARD GATE:** All reviewer agents REQUIRE Claude Opus 4.5 or higher.

### Self-Verification (MANDATORY - Check FIRST)

If you are NOT Claude Opus 4.5+ → **STOP immediately and report:**

```
ERROR: Model requirement not met
Required: Claude Opus 4.5+
Current: [your model]
Action: Cannot proceed. Orchestrator must reinvoke with model="opus"
```

### Orchestrator Requirement

When calling ANY reviewer agent, you MUST specify the model parameter:

```python
Task(subagent_type="ring:{reviewer-name}", model="opus", ...)  # REQUIRED
```

### Why Opus is Required for Reviewers

| Review Capability | Why It Requires Opus |
|------------------|---------------------|
| **Complex code tracing** | Tracing data flows across components, following function calls, understanding state changes |
| **Pattern recognition** | Identifying subtle design patterns, anti-patterns, and inconsistencies |
| **Mental execution** | Walking through code with concrete scenarios to verify correctness |
| **Context integration** | Understanding full file context, adjacent functions, ripple effects |
| **Security analysis** | Identifying attack vectors, OWASP vulnerabilities, cryptographic weaknesses |
| **Business logic verification** | Tracing business rules, edge cases, state machine transitions |

**Domain-Specific Rationale:**

- **Code Reviewer:** Requires tracing algorithmic flow, context propagation, and codebase consistency patterns
- **Business Logic Reviewer:** Requires mental execution analysis with concrete scenarios and full file context
- **Security Reviewer:** Requires deep vulnerability detection, OWASP Top 10 coverage, and cryptographic evaluation
- **Test Reviewer:** Requires analyzing test quality, coverage gaps, and test anti-patterns
- **Nil-Safety Reviewer:** Requires tracing nil propagation through call chains and identifying risk patterns

---

## Enforcement

This is a **HARD GATE** - no reviewer can proceed without Opus model verification.

If invoked with wrong model, the agent MUST:
1. NOT perform any review
2. Output the error message above
3. Return immediately

The orchestrator MUST reinvoke with `model="opus"` parameter.
