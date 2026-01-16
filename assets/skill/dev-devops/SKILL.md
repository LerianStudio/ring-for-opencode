---
name: "ring:dev-devops"
description: |
  Gate 1 of the development cycle. Creates containerization, local development setup,
  and CI/CD configuration using the ring:devops-engineer agent.
license: MIT
compatibility:
  platforms:
    - opencode

metadata:
  version: "1.0.0"
  author: "Lerian Studio"
  category: development
  trigger:
    - Gate 1 of development cycle
    - After implementation complete (Gate 0)
    - Service needs containerization or local dev setup
  sequence:
    after: [ring:dev-implementation]
    before: [ring:dev-sre]
---

# DevOps Setup (Gate 1)

## Overview

This skill creates containerization and local development infrastructure:
- Dockerfile with multi-stage build
- docker-compose.yml for local development
- Environment configuration (.env.example)
- Makefile with standard commands
- CI/CD pipeline configuration

## Role Clarification

**This skill ORCHESTRATES. DevOps agent IMPLEMENTS.**

| Who | Responsibility |
|-----|----------------|
| **This Skill** | Prepare context, dispatch agent, validate output |
| **DevOps Agent** | Create Docker files, compose, Makefile |

## Required Input

- `unit_id`: Task identifier from Gate 0
- `language`: go / typescript / python
- `service_type`: api / worker / batch / cli
- `implementation_files`: Files created in Gate 0

## Dispatch DevOps Agent

```yaml
Task:
  subagent_type: "ring:devops-engineer"
  description: "Create DevOps setup for [unit_id]"
  prompt: |
    Create containerization and local dev setup for this service.

    ## Context
    - Language: [language]
    - Service Type: [service_type]
    - Implementation: [implementation_files]

    ## Standards Reference
    WebFetch: devops.md standards

    ## Required Artifacts

    ### 1. Dockerfile
    - Multi-stage build (builder + runtime)
    - Non-root user
    - Health check
    - Minimal final image

    ### 2. docker-compose.yml
    - Service definition with health checks
    - Database if needed (postgres, redis)
    - Volume mounts for development
    - Environment variables

    ### 3. .env.example
    - All required environment variables
    - Sensible defaults where possible
    - Documentation comments

    ### 4. Makefile
    Required targets: build, lint, test, cover, run, up, down, clean

    ## Output Format
    - Files Created table
    - Verification Results
    - Standards Compliance
```

## Verification

After agent completes, verify:

| Check | Command | Expected |
|-------|---------|----------|
| Docker builds | `docker build -t test .` | Exit 0 |
| Compose valid | `docker-compose config` | Valid YAML |
| Service starts | `docker-compose up -d` | Healthy status |

## Output Schema

```markdown
## DevOps Summary
**Status:** PASS/FAIL
**Artifacts Created:** [count]

## Files Created
| File | Purpose | Lines |
|------|---------|-------|
| Dockerfile | Container image | +X |
| docker-compose.yml | Local dev | +Y |
| .env.example | Configuration | +Z |
| Makefile | Build automation | +W |

## Verification Results
| Check | Status |
|-------|--------|
| Docker builds | ✅/❌ |
| Compose valid | ✅/❌ |
| Service healthy | ✅/❌ |

## Handoff to Next Gate
- DevOps setup: COMPLETE
- Container builds: ✅
- Ready for Gate 2 (SRE): YES
```

## Pressure Resistance

| User Says | Response |
|-----------|----------|
| "Skip Docker, deploy directly" | "Containerization is REQUIRED for consistency." |
| "Compose is overkill" | "Local dev setup is mandatory. Creating docker-compose.yml." |
| "Will add CI/CD later" | "CI/CD is part of Gate 1. Adding now." |

## Anti-Rationalization

| Rationalization | Why Wrong | Required Action |
|-----------------|-----------|-----------------|
| "Dockerfile exists" | Existing ≠ compliant | **Verify against standards** |
| "Manual docker commands work" | Manual ≠ reproducible | **Create Makefile targets** |
| "Local works without compose" | Dev parity is required | **Create docker-compose.yml** |
