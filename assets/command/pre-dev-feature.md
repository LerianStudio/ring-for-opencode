---
name: ring:pre-dev-feature
description: Lightweight 4-gate pre-dev workflow for small features (<2 days)
---

I'm running the **Small Track** pre-development workflow (4 gates) for your feature.

**This track is for features that:**
- ✅ Take <2 days to implement
- ✅ Use existing architecture patterns
- ✅ Don't add new external dependencies
- ✅ Don't create new data models/entities
- ✅ Don't require multi-service integration
- ✅ Can be completed by a single developer

**If any of the above are false, use `/pre-dev-full` instead.**

## Document Organization

All artifacts will be saved to: `docs/pre-dev/<feature-name>/`

**First, let me ask you about your feature:**

Use the AskUserQuestion tool to gather:

**Question 1:** "What is the name of your feature?"
- Header: "Feature Name"
- This will be used for the directory name
- Use kebab-case (e.g., "user-logout", "email-validation", "rate-limiting")

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
- **Note:** For Go services requiring auth, reference `golang.md` → Access Manager Integration section during TRD creation (Gate 2)

**Question 3 (CONDITIONAL):** "Is this a licensed product/plugin?"
- **Auto-detection:** Before asking, check if `go.mod` contains `github.com/LerianStudio/lib-license-go`
  - If **found** → Skip this question. Licensing is already integrated at project level.
  - If **not found** → Ask this question (new project or project without licensing)
- Header: "License Requirements"
- Options:
  - "No" - Not a licensed product (open source, internal tool, etc.)
  - "Yes" - Licensed product that requires License Manager integration
- **Note:** For Go services requiring license validation, reference `golang.md` → License Manager Integration section during TRD creation (Gate 2)

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

**Question 7 (MANDATORY if Q4=Yes AND new project):** "What is the target accessibility level?"
- **Trigger:** Only ask if Q4 = "Yes" AND (no existing `package.json` OR no existing a11y config)
- **Auto-detection:** Check for existing accessibility configuration:
  - `@axe-core/*` in package.json → Skip, note "Detected: accessibility testing configured"
  - `pa11y` in package.json → Skip, note "Detected: accessibility testing configured"
  - `.lighthouserc.*` file exists → Skip, note "Detected: Lighthouse CI configured"
  - If **config found** → Skip question, use detected level
  - If **not found** → Ask question
- Header: "Accessibility"
- Options:
  - "WCAG 2.1 AA (Recommended)" - Standard compliance, 4.5:1 contrast for normal text, 3:1 for large text
  - "WCAG 2.1 AAA" - Enhanced compliance, 7:1 contrast for normal text, 4.5:1 for large text
  - "Basic" - Semantic HTML only, no strict contrast requirements
- **Usage:** Informs contrast validation in design-system.md, component requirements, code review criteria
- **Why this matters:** Accessibility requirements affect color choices, component design, and testing requirements

**Question 8 (MANDATORY if Q4=Yes AND new project):** "Does the application need dark mode support?"
- **Trigger:** Only ask if Q4 = "Yes" AND (no existing `package.json` OR no dark mode detected)
- **Auto-detection:** Check for existing dark mode:
  - `dark:` classes in existing components → Skip, note "Detected: dark mode classes in use"
  - CSS variables with `.dark` class in globals.css → Skip, note "Detected: dark mode CSS variables"
  - Theme provider (next-themes, etc.) in layout → Skip, note "Detected: theme provider configured"
  - If **dark mode found** → Skip question, note existing configuration
  - If **not found** → Ask question
- Header: "Dark Mode"
- Options:
  - "Light + Dark (Recommended)" - Full theme support with system preference detection
  - "Light only" - Single light theme, no dark mode
  - "Dark only" - Single dark theme (for specific apps like developer tools)
- **Usage:** Informs design-system.md color tokens, CSS architecture, component variants
- **Why this matters:** Dark mode requires dual color palettes and affects contrast calculations

**Question 9 (MANDATORY if Q4=Yes AND new project):** "What is the primary brand color?"
- **Trigger:** Only ask if Q4 = "Yes" AND (no existing `package.json` OR no brand colors detected)
- **Auto-detection:** Check for existing brand colors:
  - `tailwind.config.*` with custom `primary` color → Skip, note "Detected: primary color in Tailwind config"
  - CSS variable `--primary` or `--color-primary` in globals.css → Skip, note "Detected: primary color CSS variable"
  - If **brand color found** → Skip question, note existing configuration
  - If **not found** → Ask question
- Header: "Brand Color"
- Options:
  - "Blue" - Professional, trustworthy (finance, enterprise, healthcare)
  - "Purple" - Creative, innovative (tech startups, design tools)
  - "Green" - Growth, sustainability (fintech, eco, wellness)
  - "Orange" - Energy, action (marketplaces, social, food)
  - "Custom (specify hex)" - Custom brand color (user provides hex code)
- **Usage:** Informs design-system.md primary palette, derived colors, semantic mappings
- **Why this matters:** Brand color defines the primary visual identity and derived color palette

**Question 10 (MANDATORY if Q4=Yes AND new project):** "What typography style fits the brand?"
- **Trigger:** Only ask if Q4 = "Yes" AND (no existing `package.json` OR no custom fonts detected)
- **Auto-detection:** Check for existing typography:
  - Font imports in `layout.tsx` or `_app.tsx` → Skip, note "Detected: [font] configured"
  - `@fontsource/*` packages in package.json → Skip, note "Detected: [font] in dependencies"
  - Google Fonts in `next/font/google` imports → Skip, note "Detected: [font] from Google Fonts"
  - If **fonts found** → Skip question, note existing configuration
  - If **not found** → Ask question
- Header: "Typography"
- Options:
  - "Modern Tech (Geist) - Recommended" - Clean, technical feel, excellent for SaaS/developer tools
  - "Contemporary (Satoshi)" - Friendly, approachable, good for consumer apps
  - "Editorial (Cabinet Grotesk)" - Bold, statement-making, good for marketing/media
  - "Professional (General Sans)" - Neutral, enterprise-grade, good for B2B
- **Usage:** Informs design-system.md typography scale, font loading strategy, component styles
- **Why this matters:** Typography affects readability, brand perception, and page load performance
- **Note:** Avoid generic fonts (Inter, Roboto, Arial) per frontend standards

**Question 11 (MANDATORY if scope=fullstack):** "How will the frontend communicate with backend services?"
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

**UI Configuration Summary:** After Q4-Q11, display:
```
UI Configuration:
- Has UI: Yes/No
- UI Library: [choice] (confirmed by user)
- Styling: [choice] (confirmed by user)
- Accessibility: [WCAG AA | WCAG AAA | Basic] (for new projects)
- Dark Mode: [Light + Dark | Light only | Dark only] (for new projects)
- Brand Color: [Blue | Purple | Green | Orange | #hex] (for new projects)
- Typography: [Geist | Satoshi | Cabinet Grotesk | General Sans] (for new projects)
- API Pattern: [direct | bff | other] (only for fullstack)
```

**GATE BLOCKER:** If Q4 = "Yes" but Q5 or Q6 were not answered, DO NOT proceed to Gate 0. Return and ask the missing questions. For new projects with UI, Q7-Q10 are also required.

This configuration is passed to all subsequent gates and used by:
- **Gate 0 (Research):** product-designer searches for patterns in the chosen library
- **Gate 1 (PRD/UX):** wireframes use component names from the library
- **Gate 2 (TRD):** architecture considers UI library constraints
- **Gate 3 (Tasks):** implementation tasks reference specific components

After getting the feature name (and auth/license/UI requirements if applicable), create the directory structure and run the 4-gate workflow:

```bash
mkdir -p docs/pre-dev/<feature-name>
```

## Gate 0: Research Phase (Lightweight)

**Skill:** pre-dev-research

Even small features benefit from quick research:

1. Determine research mode (usually **modification** for small features)
2. Dispatch 4 research agents in PARALLEL (quick mode):
   - repo-research-analyst (codebase patterns)
   - best-practices-researcher (external best practices)
   - framework-docs-researcher (tech stack docs)
   - product-designer (UX research, mode: ux-research)
3. Save to: `docs/pre-dev/<feature-name>/research.md`
4. Get human approval before proceeding

**Gate 0 Pass Criteria (Small Track):**
- [ ] Research mode determined
- [ ] Existing patterns identified (if any)
- [ ] No conflicting implementations found
- [ ] Product/UX research documented

**Note:** For very simple changes, Gate 0 can be abbreviated - focus on checking for existing patterns.

## Gate 1: PRD Creation + UX Validation

**Skill:** pre-dev-prd-creation

1. Ask user to describe the feature (what problem does it solve, who are the users, what's the business value)
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
7. **If feature has UI:** product-designer also creates wireframes
8. Save to: `docs/pre-dev/<feature-name>/wireframes/`
9. Get human approval before proceeding

**Gate 1 Pass Criteria:**
- [ ] Problem is clearly defined
- [ ] User value is measurable
- [ ] Acceptance criteria are testable
- [ ] Scope is explicitly bounded
- [ ] UX criteria defined (ux-criteria.md created)
- [ ] Wireframes created (if feature has UI)
- [ ] User flows documented (if feature has UI)

## Gate 2: TRD Creation (Skipping Feature Map)

**Skill:** pre-dev-trd-creation

1. Load PRD from `docs/pre-dev/<feature-name>/prd.md`
2. Note: No Feature Map exists (small track) - map PRD features directly to components
3. Create TRD document with:
   - Architecture style (pattern names, not products)
   - Component design (technology-agnostic)
   - Data architecture (conceptual)
   - Integration patterns
   - Security architecture
   - **NO specific tech products** (use "Relational Database" not "PostgreSQL")
4. Save to: `docs/pre-dev/<feature-name>/trd.md`
5. Run Gate 2 validation checklist
6. Get human approval before proceeding

**Gate 2 Pass Criteria:**
- [ ] All PRD features mapped to components
- [ ] Component boundaries are clear
- [ ] Interfaces are technology-agnostic
- [ ] No specific products named

## Gate 3: Task Breakdown (Skipping API/Data/Deps)

**Skill:** pre-dev-task-breakdown

1. Load PRD from `docs/pre-dev/<feature-name>/prd.md`
2. Load TRD from `docs/pre-dev/<feature-name>/trd.md`
3. Note: No Feature Map, API Design, Data Model, or Dependency Map exist (small track)
4. Create task breakdown document with:
   - Value-driven decomposition
   - Each task delivers working software
   - Maximum task size: 2 weeks
   - Dependencies mapped
   - Testing strategy per task
5. Save to: `docs/pre-dev/<feature-name>/tasks.md`
6. Run Gate 3 validation checklist
7. Get human approval

**Gate 3 Pass Criteria:**
- [ ] Every task delivers user value
- [ ] No task larger than 2 weeks
- [ ] Dependencies are clear
- [ ] Testing approach defined

## After Completion

Report to human:

```
✅ Small Track (4 gates) complete for <feature-name>

Artifacts created (paths depend on topology.structure):

**For single-repo:**
- docs/pre-dev/<feature-name>/research.md
- docs/pre-dev/<feature-name>/prd.md
- docs/pre-dev/<feature-name>/ux-criteria.md
- docs/pre-dev/<feature-name>/wireframes/
- docs/pre-dev/<feature-name>/trd.md
- docs/pre-dev/<feature-name>/tasks.md

**For monorepo:**
- docs/pre-dev/<feature-name>/research.md (root - shared)
- docs/pre-dev/<feature-name>/prd.md (root - shared)
- docs/pre-dev/<feature-name>/trd.md (root - shared)
- docs/pre-dev/<feature-name>/tasks.md (root - index)
- {frontend.path}/docs/pre-dev/<feature-name>/ux-criteria.md
- {frontend.path}/docs/pre-dev/<feature-name>/wireframes/
- {frontend.path}/docs/pre-dev/<feature-name>/tasks.md (filtered)
- {backend.path}/docs/pre-dev/<feature-name>/tasks.md (filtered)

**For multi-repo:**
- Both repos: research.md, prd.md, trd.md (synchronized)
- Frontend repo: ux-criteria.md, wireframes/, tasks.md
- Backend repo: tasks.md

Skipped from full workflow:
- Feature Map (features simple enough to map directly)
- API Design (no new APIs)
- Data Model (no new data structures)
- Dependency Map (no new dependencies)
- Subtask Creation (tasks small enough already)

Next steps:
1. Review artifacts in docs/pre-dev/<feature-name>/
2. Use /worktree to create isolated workspace
3. Use /write-plan to create implementation plan
4. Execute the plan
```

## Remember

- This is the **Small Track** - lightweight and fast
- **Gate 0 (Research) checks for existing patterns** even for small features
- If feature grows during planning, switch to `/pre-dev-full`
- All documents saved to `docs/pre-dev/<feature-name>/`
- Get human approval at each gate
- Technology decisions happen later in Dependency Map (not in this track)

---

## MANDATORY: Skills Orchestration

**This command orchestrates multiple skills in a 4-gate workflow.**

### Gate Sequence

| Gate | Skill | Purpose | New Outputs |
|------|-------|---------|-------------|
| 0 | `pre-dev-research` | Domain, technical, and UX research (4 agents) | research.md (includes Product/UX Research) |
| 1 | `pre-dev-prd-creation` | Product requirements + UX validation + wireframes | prd.md, ux-criteria.md, wireframes/ (if UI) |
| 2 | `pre-dev-trd-creation` | Technical requirements | trd.md |
| 3 | `pre-dev-task-breakdown` | Task decomposition | tasks.md |

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
