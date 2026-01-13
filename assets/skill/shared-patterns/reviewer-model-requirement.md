# Reviewer Capability Requirements

**Version:** 1.1.0
**Applies to:** All reviewer agents (code-reviewer, business-logic-reviewer, security-reviewer, test-reviewer, nil-safety-reviewer)

---

## Required Capabilities for Reviewers

Reviewer agents require capabilities for:
- Complex multi-file code analysis
- Security vulnerability detection
- Business logic verification
- Deep pattern recognition

### Why These Capabilities Matter

| Review Capability | What It Requires |
|------------------|------------------|
| **Complex code tracing** | Tracing data flows across components, following function calls, understanding state changes |
| **Pattern recognition** | Identifying subtle design patterns, anti-patterns, and inconsistencies |
| **Mental execution** | Walking through code with concrete scenarios to verify correctness |
| **Context integration** | Understanding full file context, adjacent functions, ripple effects |
| **Security analysis** | Identifying attack vectors, OWASP vulnerabilities, cryptographic weaknesses |
| **Business logic verification** | Tracing business rules, edge cases, state machine transitions |

**Domain-Specific Requirements:**

- **Code Reviewer:** Requires tracing algorithmic flow, context propagation, and codebase consistency patterns
- **Business Logic Reviewer:** Requires mental execution analysis with concrete scenarios and full file context
- **Security Reviewer:** Requires deep vulnerability detection, OWASP Top 10 coverage, and cryptographic evaluation
- **Test Reviewer:** Requires analyzing test quality, coverage gaps, and test anti-patterns
- **Nil-Safety Reviewer:** Requires tracing nil propagation through call chains and identifying risk patterns
