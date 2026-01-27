---
name: ring:pre-dev-full
description: Complete 9-gate pre-dev workflow for large features (≥2 days)
---

I'm running the **Full Track** pre-development workflow (9 gates) for your feature.

**This track is for features that have ANY of:**
- ❌ Take ≥2 days to implement
- ❌ Add new external dependencies (APIs, databases, libraries)
- ❌ Create new data models or entities
- ❌ Require multi-service integration
- ❌ Use new architecture patterns
- ❌ Require team collaboration

**If feature is simple (<2 days, existing patterns), use `/pre-dev-feature` instead.**

## Document Organization

All artifacts will be saved to: `docs/pre-dev/<feature-name>/`

**First, let me ask you about your feature:**

Use the AskUserQuestion tool to gather:

**Question 1:** "What is the name of your feature?"
- Header: "Feature Name"
- This will be used for the directory name
- Use kebab-case (e.g., "auth-system", "payment-processing", "file-upload")

---

### Topology Discovery (MANDATORY)

**After getting the feature name, determine the project topology.**

See [shared-patterns/topology-discovery.md](../skills/shared-patterns/topology-discovery.md) for full pattern.

**Question 1.1:** "What is the scope of this feature?"
- Header: "Scope"
- Options:
  - "Fullstack (Recommended)" - Both backend API and frontend UI components
  - "Backend only" - API endpoints, services, data layer, no UI
  - "Frontend only" - UI components, pages, BFF routes, no backend API
- **Processing:**
  - If "Backend only" or "Frontend only" → Skip Q1.2-Q1.4, set `structure: single-repo`
  - If "Fullstack" → Continue to Q1.2

**Question 1.2 (if scope=fullstack):** "How is the codebase organized?"
- Header: "Structure"
- Options:
  - "Single repo (Recommended)" - All code in one repository, same root directory
  - "Monorepo" - Multiple packages in one repo (e.g., packages/api, packages/web)
  - "Multi-repo" - Separate repositories for backend and frontend
- **Processing:**
  - If "Single repo" → Skip Q1.3, continue to Q1.4
  - If "Monorepo" or "Multi-repo" → Continue to Q1.3

**Question 1.3 (if structure=monorepo or multi-repo):** Module paths
- For monorepo: Ask "Backend package path?" and "Frontend package path?"
  - Examples: `packages/api`, `apps/backend`, `services/api`
- For multi-repo: Ask "Backend repo path?" and "Frontend repo path?"
  - Use absolute paths: `/home/user/projects/backend-api`
- **Auto-detection hints:**
  - Run `ls packages/ apps/ libs/ 2>/dev/null` to suggest common monorepo paths
  - Check for `go.mod` or `package.json` in subdirectories

**Question 1.4 (if scope=fullstack):** "How should pre-dev docs be organized?"
- Header: "Docs"
- Options:
  - "Unified (Recommended)" - Single tasks.md with module tags - easier to track
  - "Per-module" - Separate task files per module (backend/, frontend/)

**TopologyConfig Output:**

After topology discovery, construct and display the configuration:

```yaml
Topology Configuration:
  scope: fullstack | backend-only | frontend-only
  structure: single-repo | monorepo | multi-repo
  modules:  # (only if monorepo or multi-repo)
    backend:
      path: [user input]
      language: [auto-detected]
    frontend:
      path: [user input]
      framework: [auto-detected]
  doc_organization: unified | per-module
  doc_placement: unified | per-module | distributed  # NEW: derived from structure
  api_pattern: direct | bff | other  # (only if fullstack)
```

**This TopologyConfig is passed to all gates and persisted in research.md frontmatter.**

---

**Question 2 (CONDITIONAL):** "Does this feature require authentication or authorization?"
- **Auto-detection:** Before asking, check if `go.mod` contains `github.com/LerianStudio/lib-auth`
  - If **found** → Skip this question. Auth is already integrated at project level.
  - If **not found** → Ask this question (new project or project without auth)
- Header: "Auth Requirements"
- Options:
  - "None" - No authentication needed
  - "User authentication only" - Users must log in but no permission checks
  - "User + permissions" - Full user auth with role-based access control
  - "Service-to-service auth" - Machine-to-machine authentication only
  - "Full (user + service-to-service)" - Both user and service auth
- **Note:** For Go services requiring auth, reference `golang.md` → Access Manager Integration section during TRD creation (Gate 3) and Dependency Map (Gate 6)

**Question 3 (CONDITIONAL):** "Is this a licensed product/plugin?"
- **Auto-detection:** Before asking, check if `go.mod` contains `github.com/LerianStudio/lib-license-go`
  - If **found** → Skip this question. Licensing is already integrated at project level.
  - If **not found** → Ask this question (new project or project without licensing)
- Header: "License Requirements"
- Options:
  - "No" - Not a licensed product (open source, internal tool, etc.)
  - "Yes" - Licensed product that requires License Manager integration
- **Note:** For Go services requiring license validation, reference `golang.md` → License Manager Integration section during TRD creation (Gate 3) and Dependency Map (Gate 6)

**Why auto-detection?** Access Manager and License Manager are project-level infrastructure decisions, not feature-level. Once integrated, all features in the project inherit them.

**Question 4 (MANDATORY):** "Does this feature include user interface (UI)?"
- Header: "Has UI?"
- Options:
  - "Yes" - Feature includes pages, screens, forms, dashboards, or any visual interface
  - "No" - Backend-only feature (API, service, CLI, etc.)
- **IMPORTANT:** This question MUST always be asked. Do NOT assume based on feature description keywords.
- **Why mandatory?** Keyword detection is unreliable. Explicitly asking prevents skipping UI configuration.

**Question 5 (MANDATORY if Q4=Yes):** "Which UI component library should be used?"
- **Trigger:** Only ask if Q4 = "Yes"
- **Auto-detection:** Before asking, check `package.json` for existing UI libraries:
  - `@radix-ui/*` packages → shadcn/ui detected
  - `@chakra-ui/*` packages → Chakra UI detected
  - `@headlessui/*` packages → Headless UI detected
  - `@mui/*` or `@material-ui/*` → Material UI detected
  - If **library found** → Still ask, but pre-select detected option and display: "Detected: [library] from package.json"
  - If **not found** → Ask without pre-selection
- Header: "UI Library"
- Options:
  - "shadcn/ui + Radix (Recommended)" - TailwindCSS-based, highly customizable, copy-paste components
  - "Chakra UI" - Full component library with built-in theming system
  - "Headless UI" - Minimal unstyled primitives, Tailwind-native
  - "Material UI" - Google's Material Design components
  - "Ant Design" - Enterprise-grade React components
  - "Custom components only" - Build from scratch, no external UI library
- **Usage:** This choice informs wireframes (component names), TRD (dependencies), and implementation tasks
- **CANNOT be skipped:** Even with auto-detection, user must confirm the choice

**Question 6 (MANDATORY if Q4=Yes):** "Which styling approach should be used?"
- **Trigger:** Only ask if Q4 = "Yes"
- **Auto-detection:** Before asking, check `package.json` for existing styling:
  - `tailwindcss` → TailwindCSS detected
  - `styled-components` → Styled Components detected
  - `@emotion/*` → Emotion detected
  - `sass` or `node-sass` → Sass/SCSS detected
  - If **styling found** → Still ask, but pre-select detected option and display: "Detected: [styling] from package.json"
  - If **not found** → Ask without pre-selection
- Header: "Styling"
- Options:
  - "TailwindCSS (Recommended)" - Utility-first CSS, excellent for design systems
  - "CSS Modules" - Scoped CSS files, traditional approach
  - "Styled Components" - CSS-in-JS with tagged template literals
  - "Sass/SCSS" - CSS preprocessor with variables and nesting
  - "Vanilla CSS" - Plain CSS without preprocessors
- **Usage:** This choice informs wireframes (class naming), component implementation, and code review criteria
- **CANNOT be skipped:** Even with auto-detection, user must confirm the choice

**Question 7 (MANDATORY if scope=fullstack):** "How will the frontend communicate with backend services?"
- **Trigger:** Only ask if topology scope = "Fullstack" (from Q1.1)
- Header: "API Pattern"
- Options:
  - "Direct API calls (Recommended for simple features)" - Frontend calls backend APIs directly
  - "BFF (Backend-for-Frontend) layer" - Frontend calls BFF which aggregates/transforms data
  - "Other (specify)" - Custom pattern (GraphQL, tRPC, custom gateway)
- **Auto-recommend based on feature complexity:**
  - If research.md indicates multiple backend services → Pre-select "BFF"
  - If simple CRUD feature → Pre-select "Direct API calls"
  - Otherwise → No pre-selection
- **Why this matters:** Determines which agent handles frontend tasks:
  - Direct → All frontend tasks use `ring:frontend-engineer`
  - BFF → Data aggregation uses `ring:frontend-bff-engineer-typescript`, UI uses `ring:frontend-engineer`
- **CANNOT be skipped for fullstack features**

**UI Configuration Summary:** After Q4/Q5/Q6/Q7, display:
```
UI Configuration:
- Has UI: Yes/No
- UI Library: [choice] (confirmed by user)
- Styling: [choice] (confirmed by user)
- API Pattern: [direct | bff | other] (only for fullstack)
```

**GATE BLOCKER:** If Q4 = "Yes" but Q5 or Q6 were not answered, DO NOT proceed to Gate 0. Return and ask the missing questions. If scope = "Fullstack" but Q7 was not answered, DO NOT proceed to Gate 0.

This configuration is passed to all subsequent gates and used by:
- **Gate 0 (Research):** product-designer searches for patterns in the chosen library
- **Gate 1 (PRD/UX):** criteria reference available components
- **Gate 2 (Feature Map/UX Design):** wireframes use component names from the library
- **Gate 3 (TRD):** architecture considers UI library constraints
- **Gate 6 (Dependency Map):** lists UI library and styling framework as dependencies
- **Gate 7 (Tasks):** implementation tasks reference specific components

After getting the feature name (and auth/license/UI requirements if applicable), create the directory structure and run the 9-gate workflow:

```bash
mkdir -p docs/pre-dev/<feature-name>
```

## Gate 0: Research Phase

**Skill:** pre-dev-research

1. Determine research mode by asking user or inferring from context:
   - **greenfield**: New capability, no existing patterns
   - **modification**: Extending existing functionality
   - **integration**: Connecting external systems

2. Dispatch 4 research agents in PARALLEL:
   - repo-research-analyst (codebase patterns, file:line refs)
   - best-practices-researcher (web search, Context7)
   - framework-docs-researcher (tech stack, versions)
   - product-designer (UX research, user problem validation, mode: ux-research)

3. Aggregate findings into research document
4. Save to: `docs/pre-dev/<feature-name>/research.md`
5. Run Gate 0 validation checklist
6. Get human approval before proceeding

**Gate 0 Pass Criteria:**
- [ ] Research mode determined and documented
- [ ] All 4 agents dispatched and returned
- [ ] At least one file:line reference (if modification mode)
- [ ] At least one external URL (if greenfield mode)
- [ ] docs/solutions/ knowledge base searched
- [ ] Tech stack versions documented
- [ ] Product/UX research documented

## Gate 1: PRD Creation + UX Validation

**Skill:** pre-dev-prd-creation

1. Ask user to describe the feature (problem, users, business value)
2. Create PRD document with:
   - Problem statement
   - User stories
   - Acceptance criteria
   - Success metrics
   - Out of scope
3. Save to: `docs/pre-dev/<feature-name>/prd.md`
4. Run Gate 1 validation checklist
5. Dispatch product-designer for UX validation (mode: ux-validation)
6. Save to: `docs/pre-dev/<feature-name>/ux-criteria.md`
7. Get human approval before proceeding

**Gate 1 Pass Criteria:**
- [ ] Problem is clearly defined
- [ ] User value is measurable
- [ ] Acceptance criteria are testable
- [ ] Scope is explicitly bounded
- [ ] UX criteria defined (ux-criteria.md created)

## Gate 2: Feature Map Creation + UX Design

**Skill:** pre-dev-feature-map

1. Load PRD from `docs/pre-dev/<feature-name>/prd.md`
2. Load ux-criteria.md from `docs/pre-dev/<feature-name>/ux-criteria.md`
3. Create feature map document with:
   - Feature relationships and dependencies
   - Domain boundaries
   - Integration points
   - Scope visualization
4. Save to: `docs/pre-dev/<feature-name>/feature-map.md`
5. Run Gate 2 validation checklist
6. Dispatch product-designer for UX design (mode: ux-design)
7. Save to: `docs/pre-dev/<feature-name>/user-flows.md`
8. Save to: `docs/pre-dev/<feature-name>/wireframes/` (directory with YAML specs)
9. Get human approval before proceeding

**Gate 2 Pass Criteria:**
- [ ] All features from PRD mapped
- [ ] Relationships are clear
- [ ] Domain boundaries defined
- [ ] Feature interactions documented
- [ ] User flows documented (user-flows.md created)
- [ ] Wireframe specs created (wireframes/ directory)

## Gate 3: TRD Creation

**Skill:** pre-dev-trd-creation

1. Load PRD from `docs/pre-dev/<feature-name>/prd.md`
2. Load Feature Map from `docs/pre-dev/<feature-name>/feature-map.md`
3. Map Feature Map domains to architectural components
4. Create TRD document with:
   - Architecture style (pattern names, not products)
   - Component design (technology-agnostic)
   - Data architecture (conceptual)
   - Integration patterns
   - Security architecture
   - **NO specific tech products**
5. Save to: `docs/pre-dev/<feature-name>/trd.md`
6. Run Gate 3 validation checklist
7. Get human approval before proceeding

**Gate 3 Pass Criteria:**
- [ ] All Feature Map domains mapped to components
- [ ] All PRD features mapped to components
- [ ] Component boundaries are clear
- [ ] Interfaces are technology-agnostic
- [ ] No specific products named

## Gate 4: API Design

**Skill:** pre-dev-api-design

1. Load previous artifacts (PRD, Feature Map, TRD)
2. Create API design document with:
   - Component contracts and interfaces
   - Request/response formats
   - Error handling patterns
   - Integration specifications
3. Save to: `docs/pre-dev/<feature-name>/api-design.md`
4. Run Gate 4 validation checklist
5. Get human approval before proceeding

**Gate 4 Pass Criteria:**
- [ ] All component interfaces defined
- [ ] Contracts are clear and complete
- [ ] Error cases covered
- [ ] Protocol-agnostic (no REST/gRPC specifics yet)

## Gate 5: Data Model

**Skill:** pre-dev-data-model

1. Load previous artifacts
2. Create data model document with:
   - Entity relationships and schemas
   - Data ownership boundaries
   - Access patterns
   - Migration strategy
3. Save to: `docs/pre-dev/<feature-name>/data-model.md`
4. Run Gate 5 validation checklist
5. Get human approval before proceeding

**Gate 5 Pass Criteria:**
- [ ] All entities defined with relationships
- [ ] Data ownership is clear
- [ ] Access patterns documented
- [ ] Database-agnostic (no PostgreSQL/MongoDB specifics yet)

## Gate 6: Dependency Map

**Skill:** pre-dev-dependency-map

1. Load previous artifacts
2. Create dependency map document with:
   - **NOW we select specific technologies**
   - Concrete versions and packages
   - Rationale for each choice
   - Alternative evaluations
3. Save to: `docs/pre-dev/<feature-name>/dependency-map.md`
4. Run Gate 6 validation checklist
5. Get human approval before proceeding

**Gate 6 Pass Criteria:**
- [ ] All technologies selected with rationale
- [ ] Versions pinned (no "latest")
- [ ] Alternatives evaluated
- [ ] Tech stack is complete

## Gate 7: Task Breakdown

**Skill:** pre-dev-task-breakdown

1. Load all previous artifacts (PRD, Feature Map, TRD, API Design, Data Model, Dependency Map)
2. Create task breakdown document with:
   - Value-driven decomposition
   - Each task delivers working software
   - Maximum task size: 2 weeks
   - Dependencies mapped
   - Testing strategy per task
3. Save to: `docs/pre-dev/<feature-name>/tasks.md`
4. Run Gate 7 validation checklist
5. Get human approval before proceeding

**Gate 7 Pass Criteria:**
- [ ] Every task delivers user value
- [ ] No task larger than 2 weeks
- [ ] Dependencies are clear
- [ ] Testing approach defined

## Gate 8: Subtask Creation

**Skill:** pre-dev-subtask-creation

1. Load tasks from `docs/pre-dev/<feature-name>/tasks.md`
2. Create subtask breakdown document with:
   - Bite-sized steps (2-5 minutes each)
   - TDD-based implementation steps
   - Complete code (no placeholders)
   - Zero-context executable
3. Save to: `docs/pre-dev/<feature-name>/subtasks.md`
4. Run Gate 8 validation checklist
5. Get human approval

**Gate 8 Pass Criteria:**
- [ ] Every subtask is 2-5 minutes
- [ ] TDD cycle enforced (test first)
- [ ] Complete code provided
- [ ] Zero-context test passes

## After Completion

Report to human:

```
✅ Full Track (9 gates) complete for <feature-name>

Artifacts created (paths depend on topology.structure):

**For single-repo (all in docs/pre-dev/<feature-name>/):**
- research.md, prd.md, feature-map.md, trd.md
- ux-criteria.md, user-flows.md, wireframes/
- api-design.md, data-model.md, dependency-map.md
- tasks.md, subtasks.md

**For monorepo (distributed by module):**
*Root (shared):* research.md, prd.md, feature-map.md, trd.md, tasks.md (index)
*Backend module:* api-design.md, data-model.md, dependency-map.md, tasks.md
*Frontend module:* ux-criteria.md, user-flows.md, wireframes/, dependency-map.md, tasks.md

**For multi-repo (per-repository):**
*Both repos:* research.md, prd.md, trd.md (synchronized)
*Backend repo:* api-design.md, data-model.md, dependency-map.md, tasks.md
*Frontend repo:* ux-criteria.md, user-flows.md, wireframes/, dependency-map.md, tasks.md

Planning time: 2.5-4.5 hours (comprehensive with UX design)

Next steps:
1. Review artifacts in docs/pre-dev/<feature-name>/
2. Use /worktree to create isolated workspace
3. Use /write-plan to create implementation plan
4. Execute the plan (ui-engineer will use user-flows.md and wireframes/)
```

## Remember

- This is the **Full Track** - comprehensive and thorough
- All 9 gates provide maximum planning depth
- **Gate 0 (Research) runs 4 agents in parallel** for codebase, best practices, framework docs, and UX research
- **Gate 1 creates ux-criteria.md** with UX acceptance criteria
- **Gate 2 creates user-flows.md and wireframes/** with detailed UX design
- Technology decisions happen at Gate 6 (Dependency Map)
- All documents saved to `docs/pre-dev/<feature-name>/`
- Get human approval at each gate before proceeding
- Planning investment (2.5-4.5 hours) pays off during implementation

---

## MANDATORY: Skills Orchestration

**This command orchestrates multiple skills in a 9-gate workflow.**

### Gate Sequence

| Gate | Skill | Purpose | New Outputs |
|------|-------|---------|-------------|
| 0 | `pre-dev-research` | Domain, technical, and UX research (4 agents) | research.md (includes Product/UX Research) |
| 1 | `pre-dev-prd-creation` | Product requirements + UX validation | prd.md, ux-criteria.md |
| 2 | `pre-dev-feature-map` | Feature scope + UX design | feature-map.md, user-flows.md, wireframes/ |
| 3 | `pre-dev-trd-creation` | Technical requirements | trd.md |
| 4 | `pre-dev-api-design` | API contracts | api-design.md |
| 5 | `pre-dev-data-model` | Data architecture | data-model.md |
| 6 | `pre-dev-dependency-map` | Technology selection | dependency-map.md |
| 7 | `pre-dev-task-breakdown` | Task decomposition | tasks.md |
| 8 | `pre-dev-subtask-creation` | Implementation steps | subtasks.md |

### Execution Pattern

```
For each gate:
  Use Skill tool: [gate-skill]
  Wait for human approval
  Proceed to next gate
```

Each skill contains its own:
- Anti-rationalization tables
- Gate pass criteria
- Output format requirements

**Do NOT skip gates.** Each gate builds on the previous gate's output.
