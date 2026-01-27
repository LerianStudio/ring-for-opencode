# Topology Discovery Pattern

Reusable pattern for discovering project structure at the start of pre-dev workflows.

---

## When to Use

**MANDATORY** at the START of pre-dev workflow (before Gate 0 research).

This pattern ensures:
- Correct working directories are identified
- Docs are organized appropriately
- Tasks are tagged for proper execution context

---

## Questions Flow

```
Q1: Feature scope?
├─ backend-only → Set scope, skip Q2-Q5, use single directory
├─ frontend-only → Set scope, skip Q2-Q5, use single directory
└─ fullstack → Continue to Q2

Q2: Repository structure? (only if fullstack)
├─ single-repo → Set structure, skip Q3, unified docs
├─ monorepo → Continue to Q3
└─ multi-repo → Continue to Q3

Q3: Module paths (only if monorepo or multi-repo)
├─ Backend path: user input
└─ Frontend path: user input

Q4: Doc organization? (only if fullstack)
├─ unified → Single tasks.md with target tags
└─ per-module → Separate task files per module

Q5: API Pattern? (only if scope=fullstack)
├─ direct → Set api_pattern: direct
├─ bff → Set api_pattern: bff
└─ other → Ask for specification, set api_pattern: other
```

---

## AskUserQuestion Implementations

### Q1: Feature Scope

```json
{
  "question": "What is the scope of this feature?",
  "header": "Scope",
  "multiSelect": false,
  "options": [
    {
      "label": "Fullstack (Recommended)",
      "description": "Both backend API and frontend UI components"
    },
    {
      "label": "Backend only",
      "description": "API endpoints, services, data layer, no UI"
    },
    {
      "label": "Frontend only",
      "description": "UI components, pages, BFF routes, no backend API"
    }
  ]
}
```

**Processing:**
- "Fullstack" → `scope: fullstack`, continue to Q2
- "Backend only" → `scope: backend-only`, skip to output
- "Frontend only" → `scope: frontend-only`, skip to output

### Q2: Repository Structure

**Only ask if `scope: fullstack`**

```json
{
  "question": "How is the codebase organized?",
  "header": "Structure",
  "multiSelect": false,
  "options": [
    {
      "label": "Single repo (Recommended)",
      "description": "All code in one repository, same root directory"
    },
    {
      "label": "Monorepo",
      "description": "Multiple packages in one repo (e.g., packages/api, packages/web)"
    },
    {
      "label": "Multi-repo",
      "description": "Separate repositories for backend and frontend"
    }
  ]
}
```

**Processing:**
- "Single repo" → `structure: single-repo`, skip Q3, continue to Q4
- "Monorepo" → `structure: monorepo`, continue to Q3
- "Multi-repo" → `structure: multi-repo`, continue to Q3

### Q3: Module Paths

**Only ask if `structure: monorepo` or `structure: multi-repo`**

For monorepo, ask via follow-up prompts:
- "What is the backend package path? (e.g., packages/api, apps/backend)"
- "What is the frontend package path? (e.g., packages/web, apps/frontend)"

For multi-repo, ask via follow-up prompts:
- "What is the absolute path to the backend repository?"
- "What is the absolute path to the frontend repository?"

**Auto-detection hints:**

```bash
# Monorepo detection
ls packages/ apps/ libs/ 2>/dev/null | head -10

# Look for package.json or go.mod in subdirectories
find . -maxdepth 3 -name "package.json" -o -name "go.mod" | head -10
```

### Q4: Doc Organization

**Only ask if `scope: fullstack`**

```json
{
  "question": "How should the pre-dev documentation be organized?",
  "header": "Docs",
  "multiSelect": false,
  "options": [
    {
      "label": "Unified (Recommended)",
      "description": "Single tasks.md with module tags - easier to track progress"
    },
    {
      "label": "Per-module",
      "description": "Separate task files per module (backend/, frontend/) - better for separate teams"
    }
  ]
}
```

**Processing:**
- "Unified" → `doc_organization: unified`
- "Per-module" → `doc_organization: per-module`

### Q5: API Pattern

**Only ask if `scope: fullstack`**

```json
{
  "question": "How will the frontend communicate with backend services?",
  "header": "API Pattern",
  "multiSelect": false,
  "options": [
    {
      "label": "Direct API calls (Recommended for simple features)",
      "description": "Frontend calls backend APIs directly. Best for single backend service, no aggregation needed."
    },
    {
      "label": "BFF (Backend-for-Frontend) layer",
      "description": "Frontend calls BFF which aggregates/transforms data from multiple backends. Best for complex data needs."
    },
    {
      "label": "Other (specify)",
      "description": "Custom pattern - describe your approach"
    }
  ]
}
```

**Processing:**
- "Direct API calls" → `api_pattern: direct`
- "BFF" → `api_pattern: bff`
- "Other" → Ask for specification, set `api_pattern: other`

**When to Recommend Each Pattern:**

| Pattern | Recommend When |
|---------|---------------|
| **Direct** | Single backend service, simple CRUD, no aggregation, <3 API calls per page |
| **BFF** | Multiple backend services, data aggregation needed, complex transformations, sensitive keys to hide |
| **Other** | GraphQL federation, tRPC, custom gateway, existing patterns |

---

## Output: TopologyConfig

After completing the questions flow, construct the TopologyConfig object:

### Backend-only or Frontend-only

```yaml
topology:
  scope: backend-only  # or frontend-only
  structure: single-repo
```

### Single-repo Fullstack

```yaml
topology:
  scope: fullstack
  structure: single-repo
  doc_organization: unified  # or per-module
  doc_placement: unified     # derived from structure
  api_pattern: direct        # or bff, other
```

### Monorepo Fullstack

```yaml
topology:
  scope: fullstack
  structure: monorepo
  modules:
    backend:
      path: packages/api        # From Q3
      language: golang          # Auto-detected or asked
    frontend:
      path: packages/web        # From Q3
      framework: nextjs         # Auto-detected or asked
  doc_organization: unified     # From Q4
  doc_placement: per-module     # derived from structure
  api_pattern: bff              # From Q5
```

### Multi-repo Fullstack

```yaml
topology:
  scope: fullstack
  structure: multi-repo
  modules:
    backend:
      path: /home/user/projects/backend-api    # From Q3
      language: typescript                      # Auto-detected
    frontend:
      path: /home/user/projects/frontend-app   # From Q3
      framework: react                          # Auto-detected
  doc_organization: per-module                  # From Q4
  doc_placement: distributed                    # derived from structure
  api_pattern: direct                           # From Q5
```

---

## Language/Framework Auto-Detection

### Backend Language Detection

```bash
# Check for Go
if [ -f "go.mod" ] || [ -f "{backend_path}/go.mod" ]; then
  language="golang"
fi

# Check for TypeScript/Node
if [ -f "package.json" ] || [ -f "{backend_path}/package.json" ]; then
  # Check if it's a backend (has express, fastify, nest, etc.)
  if grep -q '"express"\|"fastify"\|"@nestjs"\|"hono"' package.json; then
    language="typescript"
  fi
fi
```

### Frontend Framework Detection

```bash
# Check package.json for framework
if grep -q '"next"' package.json; then
  framework="nextjs"
elif grep -q '"react"' package.json && ! grep -q '"next"' package.json; then
  framework="react"
elif grep -q '"vue"' package.json; then
  framework="vue"
elif grep -q '"@angular/core"' package.json; then
  framework="angular"
fi
```

---

## Persistence

The TopologyConfig MUST be persisted in the `research.md` frontmatter:

```yaml
---
feature: my-feature-name
gate: 0
date: 2026-01-21
topology:
  scope: fullstack
  structure: monorepo
  modules:
    backend:
      path: packages/api
      language: golang
    frontend:
      path: packages/web
      framework: nextjs
  doc_organization: unified
  api_pattern: bff
---

# Research: My Feature Name

...
```

All subsequent gates MUST read the topology from the research.md frontmatter.

---

## Integration Points

| Workflow Step | How Topology is Used |
|---------------|---------------------|
| Gate 0: Research | Persist config, dispatch agents per module path |
| Gate 1: PRD | Include module-specific requirements |
| Gate 2/3: TRD | Architecture per module |
| Gate 7: Task Breakdown | Tag tasks with `target:` and `working_directory:` |
| Execution | Context switching between modules |

---

## Defaults and Skip Conditions

| Condition | Default | Rationale |
|-----------|---------|-----------|
| Scope not fullstack | Skip Q2-Q5 | Single directory, no module coordination needed |
| Structure is single-repo | Skip Q3 | No separate paths |
| User selects "Other" | Ask follow-up | Allow custom configurations |
| API pattern "Other" | Ask specification | Support custom patterns (GraphQL, tRPC, etc.) |

---

## Documentation Path Resolution

### Path Resolution Function

Skills MUST use this logic to determine where to write documents:

```python
def get_doc_path(doc_type: str, feature_name: str, topology: dict) -> str | list[str]:
    """
    Returns the path(s) where a document should be written.

    Args:
        doc_type: One of 'research', 'prd', 'trd', 'ux-criteria', 'wireframes',
                  'api-design', 'data-model', 'dependency-map', 'tasks'
        feature_name: The feature name in kebab-case
        topology: The TopologyConfig dictionary

    Returns:
        Single path string, or list of paths for multi-repo shared docs
    """
    structure = topology.get('structure', 'single-repo')
    modules = topology.get('modules', {})

    # Single-repo: all docs in one place
    if structure == 'single-repo':
        return f"docs/pre-dev/{feature_name}/"

    # Shared documents (research, prd, trd)
    if doc_type in ['research', 'prd', 'trd']:
        if structure == 'monorepo':
            return f"docs/pre-dev/{feature_name}/"
        else:  # multi-repo
            # Return both paths - document must be copied to both repos
            backend_path = modules.get('backend', {}).get('path', '.')
            frontend_path = modules.get('frontend', {}).get('path', '.')
            return [
                f"{backend_path}/docs/pre-dev/{feature_name}/",
                f"{frontend_path}/docs/pre-dev/{feature_name}/"
            ]

    # Frontend documents (ux-criteria, wireframes)
    if doc_type in ['ux-criteria', 'wireframes']:
        frontend_path = modules.get('frontend', {}).get('path', '.')
        return f"{frontend_path}/docs/pre-dev/{feature_name}/"

    # Backend documents (api-design, data-model)
    if doc_type in ['api-design', 'data-model']:
        backend_path = modules.get('backend', {}).get('path', '.')
        return f"{backend_path}/docs/pre-dev/{feature_name}/"

    # Split documents (dependency-map, tasks)
    if doc_type in ['dependency-map', 'tasks']:
        # Return paths for both modules - skill handles split logic
        if structure == 'monorepo':
            backend_path = modules.get('backend', {}).get('path', '.')
            frontend_path = modules.get('frontend', {}).get('path', '.')
            return {
                'index': f"docs/pre-dev/{feature_name}/",
                'backend': f"{backend_path}/docs/pre-dev/{feature_name}/",
                'frontend': f"{frontend_path}/docs/pre-dev/{feature_name}/"
            }
        else:  # multi-repo
            backend_path = modules.get('backend', {}).get('path', '.')
            frontend_path = modules.get('frontend', {}).get('path', '.')
            return {
                'backend': f"{backend_path}/docs/pre-dev/{feature_name}/",
                'frontend': f"{frontend_path}/docs/pre-dev/{feature_name}/"
            }

    # Default fallback
    return f"docs/pre-dev/{feature_name}/"
```

### Document Classification

| Document | Classification | Placement Rule |
|----------|---------------|----------------|
| research.md | Shared | Root (monorepo) or both repos (multi-repo) |
| prd.md | Shared | Root (monorepo) or both repos (multi-repo) |
| trd.md | Shared | Root (monorepo) or both repos (multi-repo) |
| ux-criteria.md | Frontend | Frontend module/repo path |
| wireframes/ | Frontend | Frontend module/repo path |
| api-design.md | Backend | Backend module/repo path |
| data-model.md | Backend | Backend module/repo path |
| dependency-map.md | Split | Index at root, module-specific at module paths |
| tasks.md | Split | Index at root, filtered tasks at module paths |

### Multi-Repo Document Synchronization

For multi-repo with shared documents (research.md, prd.md, trd.md):

1. **Write to primary location first** (backend repo by convention)
2. **Copy to secondary location** (frontend repo)
3. **Include sync note in document footer:**

```markdown
---
**Note:** This document is synchronized across repositories.
Primary: {backend.path}/docs/pre-dev/{feature}/
Mirror: {frontend.path}/docs/pre-dev/{feature}/
```

### Directory Creation

Before writing any document, create the target directory:

```bash
# Single path
mkdir -p "{path}"

# Multi-repo (both paths)
mkdir -p "{backend_path}" "{frontend_path}"
```

---

## Error Handling

| Error | Recovery |
|-------|----------|
| Invalid path provided | Re-prompt with validation hint |
| Path doesn't exist (multi-repo) | Warn but allow (might be created later) |
| Cannot auto-detect language/framework | Ask user explicitly |
| Conflicting detection (e.g., both Go and TS) | Ask user to clarify |
