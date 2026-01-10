# Reviewer Model Requirement

**Version:** 1.0.0
**Applies to:** All reviewer agents (code-reviewer, business-logic-reviewer, security-reviewer, test-reviewer, nil-safety-reviewer)

---

## Model Requirement: High-Capability Model Required

**HARD GATE:** All reviewer agents REQUIRE a high-capability model (e.g., Claude Opus, GPT-4, or equivalent).

### Self-Verification (MANDATORY - Check FIRST)

If you are NOT running on a high-capability model -> **STOP immediately and report:**

```
ERROR: Model requirement not met
Required: High-capability model (Claude Opus, GPT-4, or equivalent)
Action: Cannot proceed. Orchestrator must reinvoke with appropriate model
```

### Orchestrator Requirement

When dispatching ANY reviewer agent, ensure the model has sufficient capability for:
- Complex multi-file code analysis
- Security vulnerability detection
- Business logic verification
- Deep pattern recognition

### Why High-Capability Models are Required for Reviewers

| Review Capability | Why It Requires High Capability |
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

This is a **HARD GATE** - no reviewer can proceed without appropriate model capability.

If invoked with insufficient model capability, the agent MUST:
1. NOT perform any review
2. Output the error message above
3. Return immediately

The orchestrator MUST reinvoke with an appropriate high-capability model.
