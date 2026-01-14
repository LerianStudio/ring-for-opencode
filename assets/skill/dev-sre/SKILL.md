---
name: "ring:dev-ring:sre"
description: |
  Gate 2 of the development cycle. VALIDATES that observability was correctly implemented
  by developers. Does NOT implement observability code - only validates it.
license: MIT
compatibility:
  platforms:
    - opencode

metadata:
  version: "1.0.0"
  author: "Lerian Studio"
  category: development
  trigger:
    - Gate 2 of development cycle
    - Gate 0 (Implementation) complete with observability code
    - Gate 1 (DevOps) setup complete
    - Service needs observability validation (logging, tracing)
  sequence:
    after: [ring:dev-devops]
    before: [ring:dev-testing]
---

# SRE Validation (Gate 2)

## Overview

This skill VALIDATES that observability was correctly implemented by developers:
- Structured logging with trace correlation
- OpenTelemetry tracing instrumentation
- Code instrumentation coverage (90%+ required)
- Context propagation for distributed tracing

## Role Clarification

**Developers IMPLEMENT observability. SRE VALIDATES it.**

| Who | Responsibility |
|-----|----------------|
| **Developers** (Gate 0) | IMPLEMENT observability following Ring Standards |
| **SRE Agent** (Gate 2) | VALIDATE that observability is correctly implemented |
| **Implementation Agent** | FIX issues found by SRE (if any) |

**If observability is missing or incorrect:**
1. SRE reports issues with severity levels
2. This skill dispatches fixes to the implementation agent
3. SRE re-validates after fixes
4. Max 3 iterations, then escalate to user

## Required Input

- `unit_id`: Task/subtask being validated
- `language`: go / typescript / python
- `service_type`: api / worker / batch / cli / library
- `implementation_agent`: Agent that performed Gate 0
- `implementation_files`: Files created/modified in Gate 0

## Dispatch SRE Agent

```yaml
Task:
  subagent_type: "ring:sre"
  description: "Validate observability for [unit_id]"
  prompt: |
    VALIDATE Observability Implementation

    ## Context
    - Language: [language]
    - Service Type: [service_type]
    - Files to Validate: [implementation_files]

    ## Validation Checklist

    ### 0. FORBIDDEN Logging Patterns (Check FIRST)
    | Language | FORBIDDEN | Search For |
    |----------|-----------|------------|
    | Go | fmt.Println, log.Printf | grep in *.go |
    | TypeScript | console.log | grep in *.ts |

    If ANY FORBIDDEN pattern found: Severity CRITICAL, Verdict FAIL

    ### 1. Structured Logging
    - Uses lib-commons logger (Go) or lib-common-js (TypeScript)
    - JSON format with timestamp, level, message
    - trace_id correlation in logs

    ### 2. Instrumentation Coverage (90%+ required)
    Count spans in Handlers, Services, Repositories

    ### 3. Context Propagation
    For external calls: InjectHTTPContext, InjectGRPCContext

    ## Required Output

    ### Validation Summary
    | Check | Status | Evidence |
    |-------|--------|----------|
    | Structured Logging | ✅/❌ | [file:line] |
    | Tracing Enabled | ✅/❌ | [file:line] |
    | Instrumentation ≥90% | ✅/❌ | [X%] |

    ### Instrumentation Coverage Table
    | Layer | Instrumented | Total | Coverage |
    |-------|--------------|-------|----------|
    | Handlers | X | Y | Z% |
    | Services | X | Y | Z% |
    | **TOTAL** | X | Y | **Z%** |

    ### Issues Found (if any)
    - Severity: CRITICAL/HIGH/MEDIUM/LOW
    - File: [path:line]
    - Fix Required By: [implementation_agent]

    ### Verdict
    **ALL CHECKS PASSED:** YES/NO
    **Instrumentation Coverage:** [X%]
```

## Severity Calibration

| Severity | Scenario | Action |
|----------|----------|--------|
| **CRITICAL** | Missing ALL observability | Return to Gate 0 |
| **CRITICAL** | fmt.Println instead of JSON logs | Return to Gate 0 |
| **HIGH** | Instrumentation 50-89% | Fix and re-validate |
| **MEDIUM** | Missing context propagation | Fix and re-validate |
| **LOW** | Minor logging improvements | Note for future |

## Output Schema

```markdown
## Validation Result
**Status:** PASS/FAIL/NEEDS_FIXES
**Iterations:** [N]
**Instrumentation Coverage:** [X%]

## Instrumentation Coverage
| Layer | Instrumented | Total | Coverage |
|-------|--------------|-------|----------|
| Handlers | X | Y | Z% |
| Services | X | Y | Z% |
| Repositories | X | Y | Z% |
| **TOTAL** | X | Y | **Z%** |

## Issues Found
[List or "None"]

## Handoff to Next Gate
- SRE validation: COMPLETE
- Logging: ✅ Structured JSON with trace_id
- Tracing: ✅ OpenTelemetry instrumented
- Ready for Gate 3 (Testing): YES
```

## Pressure Resistance

| User Says | Response |
|-----------|----------|
| "Skip SRE validation" | "Observability is MANDATORY. Dispatching SRE agent now." |
| "90% coverage is too high" | "90% is the Ring Standard minimum. Cannot lower." |
| "Will add instrumentation later" | "Instrumentation is part of implementation. Fix now." |

## Component Type Requirements

| Type | JSON Logs | Tracing | Instrumentation |
|------|-----------|---------|-----------------|
| **API Service** | REQUIRED | REQUIRED | 90%+ |
| **Background Worker** | REQUIRED | REQUIRED | 90%+ |
| **CLI Tool** | REQUIRED | N/A | N/A |
| **Library** | N/A | N/A | N/A |
