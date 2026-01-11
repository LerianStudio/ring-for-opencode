---
description: Autonomous two-phase codebase exploration with adaptive agents
agent: build
subtask: true
---

Autonomously discover codebase structure, then explore deeply with adaptive agents.

## Two-Phase Exploration

### Phase 1: Discovery (Meta-Exploration)
Launch 3-4 agents in parallel to understand the codebase:

- **Architecture Discovery**: Identify pattern (hexagonal, layered, microservices, etc.)
- **Component Discovery**: Enumerate major components/modules/services
- **Layer Discovery**: Discover layers within components
- **Organization Discovery**: Understand organizing principle (by layer, feature, domain)

### Phase 2: Deep Dive (Adaptive Exploration)
Based on Phase 1 discoveries, launch N agents (one per component/layer/service):

| What Phase 1 Found | Phase 2 Strategy |
|--------------------|------------------|
| 3 components x 4 layers | Launch 3 agents (one per component) |
| Single component, clear layers | Launch 4 agents (one per layer) |
| 5 microservices | Launch 5 agents (one per service) |

Each agent explores the target within their assigned scope.

### Phase 3: Synthesis
- Integrate Phase 1 structural map with Phase 2 deep dives
- Identify cross-cutting insights
- Document consistent patterns and variations
- Provide actionable implementation guidance

## Output

Comprehensive synthesis document with:
- Executive Summary
- Architecture pattern with evidence
- Component structure with responsibilities
- Deep dive findings per area
- Cross-cutting insights
- Implementation guidance (where to add/modify code)
- Recommended next steps

$ARGUMENTS

Specify the target feature, component, or system to explore (e.g., "account creation", "authentication system").
