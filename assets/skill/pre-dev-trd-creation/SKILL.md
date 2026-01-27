---
name: ring:pre-dev-trd-creation
description: |
  Gate 3: Technical architecture document - defines HOW/WHERE with technology-agnostic
  patterns before concrete implementation choices.

trigger: |
  - PRD passed Gate 1 (required)
  - Feature Map passed Gate 2 (if Large Track)
  - About to design technical architecture
  - Tempted to specify "PostgreSQL" instead of "Relational Database"

skip_when: |
  - PRD not validated → complete Gate 1 first
  - Architecture already documented → proceed to API Design
  - Pure business requirement change → update PRD

sequence:
  after: [pre-dev-prd-creation, pre-dev-feature-map]
  before: [pre-dev-api-design, pre-dev-task-breakdown]
---

# TRD Creation - Architecture Before Implementation

## Foundational Principle

**Architecture decisions (HOW/WHERE) must be technology-agnostic patterns before concrete implementation choices.**

Specifying technologies in TRD creates:
- Vendor lock-in before evaluating alternatives
- Architecture coupled to specific products
- Technology decisions made without full dependency analysis

**The TRD answers**: HOW we'll architect the solution and WHERE components will live.
**The TRD never answers**: WHAT specific products, frameworks, versions, or packages we'll use.

---

## ⛔ HARD BLOCK: Tech Stack Definition (Step 0)

**This is a HARD GATE. Do NOT proceed without defining the tech stack.**

### Step 0.1: Auto-Detect or Ask User

**Auto-detection:** `go.mod` exists → Go | `package.json` with react/next → Frontend TS | `package.json` with express/fastify/nestjs → Backend TS

**If ambiguous, AskUserQuestion:** "What is the primary technology stack?" Options: Go (Backend), TypeScript (Backend), TypeScript (Frontend), Full-Stack TypeScript

### Step 0.2: Load Ring Standards via WebFetch

| Standard | URL | Purpose |
|----------|-----|---------|
| **golang.md** | `https://raw.githubusercontent.com/LerianStudio/ring/main/dev-team/docs/standards/golang.md` | Go patterns, DDD |
| **typescript.md** | `https://raw.githubusercontent.com/LerianStudio/ring/main/dev-team/docs/standards/typescript.md` | TS patterns, async |
| **frontend.md** | `https://raw.githubusercontent.com/LerianStudio/ring/main/dev-team/docs/standards/frontend.md` | React, Next.js, a11y |
| **devops.md** | `https://raw.githubusercontent.com/LerianStudio/ring/main/dev-team/docs/standards/devops.md` | Docker, CI/CD |
| **sre.md** | `https://raw.githubusercontent.com/LerianStudio/ring/main/dev-team/docs/standards/sre.md` | Health checks, logging |

| Tech Stack | Load |
|------------|------|
| Go Backend | golang.md + devops.md + sre.md |
| TypeScript Backend | typescript.md + devops.md + sre.md |
| TypeScript Frontend | frontend.md + devops.md |
| Full-Stack TypeScript | typescript.md + frontend.md + devops.md + sre.md |

### Step 0.3: Read PROJECT_RULES.md

Check: `docs/PROJECT_RULES.md` → `docs/STANDARDS.md` (legacy) → STOP if not found

### Step 0.4: Analyze PRD and Suggest Technologies

Read PRD, extract requirements, suggest technologies that address each, present to user for confirmation. Document in TRD metadata for Gate 6 to create PROJECT_RULES.md.

**AskUserQuestion:** "What deployment model?" Options: Cloud, On-Premise, Hybrid

### Step 0.5: Document in TRD Metadata

TRD header must include: `feature`, `gate: 3`, `deployment.model`, `tech_stack.primary`, `tech_stack.standards_loaded[]`, `project_technologies[]` (category, prd_requirement, choice, rationale per technology decision)

This metadata flows to Gates 4-6.

### Pressure Resistance for Step 0

| Pressure | Response |
|----------|----------|
| "Tech stack doesn't matter for architecture" | "Architecture patterns vary by language. Go patterns ≠ TypeScript patterns. Define stack first." |
| "We'll decide tech stack later" | "Later = Dependency Map. But architecture NOW needs to know capabilities. Define stack." |
| "Just use generic patterns" | "Generic patterns miss stack-specific best practices. 5 min to define saves rework." |
| "Skip to save time" | "Skipping causes Gates 4-6 to ask again. Define once here, inherit everywhere." |

---

## Mandatory Workflow

| Phase | Activities |
|-------|------------|
| **1. Analysis (After Step 0)** | PRD (Gate 1) required; Feature Map (Gate 2) optional; identify NFRs (performance, security, scalability); map domains to components |
| **2. Architecture Definition** | Choose style (Microservices, Modular Monolith, Serverless); design components with boundaries; define interfaces; model data architecture; plan integration patterns; design security |
| **3. Gate 3 Validation** | All domains mapped; component boundaries clear; interfaces technology-agnostic; data ownership explicit; quality attributes achievable; no specific products named |

## Explicit Rules

### ✅ DO Include
System architecture style (patterns, not products), component design with responsibilities, data architecture (ownership, flows - conceptual), API design (contracts, not protocols), security architecture (layers, threat model), integration patterns (sync/async, not tools), performance targets, deployment topology (logical)

### ❌ NEVER Include
Technology products (PostgreSQL, Redis, Kafka), framework versions (Fiber v2, React 18), language specifics (Go 1.24, Node.js 20), cloud services (AWS RDS, Azure Functions), packages (bcrypt, zod, prisma), container orchestration (Kubernetes, ECS), CI/CD details, IaC specifics

### Technology Abstraction Rules

| Element | Say This (✅) | Not This (❌) |
|---------|--------------|---------------|
| Database | "Relational Database" | "PostgreSQL 16" |
| Cache | "In-Memory Cache" | "Redis" or "Valkey" |
| Message Queue | "Message Broker" | "RabbitMQ" |
| Object Storage | "Blob Storage" | "MinIO" or "S3" |
| Web Framework | "HTTP Router" | "Fiber" or "Express" |
| Auth | "JWT-based Authentication" | "specific library" |

## Rationalization Table

| Excuse | Reality |
|--------|---------|
| "Everyone knows we use PostgreSQL" | Assumptions prevent proper evaluation. Stay abstract. |
| "Just mentioning the tech stack for context" | Context belongs in Dependency Map. Keep TRD abstract. |
| "The team needs to know what we're using" | They'll know in Dependency Map. TRD is patterns only. |
| "It's obvious we need Redis here" | Obvious ≠ documented. Abstract to "cache", decide later. |
| "I'll save time by specifying frameworks now" | You'll waste time when better options emerge. Wait. |
| "But our project template requires X" | Templates are implementation. TRD is architecture. Separate. |
| "The dependency is critical to the design" | Then describe the *capability* needed, not the product. |
| "Stakeholders expect to see technology choices" | Stakeholders see them in Dependency Map. Not here. |
| "Architecture decisions depend on technology X" | Then your architecture is too coupled. Redesign abstractly. |
| "We already decided on the tech stack" | Decisions without analysis are assumptions. Validate later. |

## Red Flags - STOP

If you catch yourself writing any of these in a TRD, **STOP**:

- Specific product names with version numbers
- Package manager commands (npm install, go get, pip install)
- Cloud provider service names (RDS, Lambda, Cloud Run, etc.)
- Framework-specific terms (Fiber middleware, React hooks, Express routers)
- Container/orchestration specifics (Docker, K8s, ECS)
- Programming language version constraints
- Infrastructure service names (CloudFront, Cloudflare, Fastly)
- CI/CD tool names (GitHub Actions, CircleCI, Jenkins)

**When you catch yourself**: Replace the product name with the capability it provides. "PostgreSQL 16" → "Relational Database with ACID guarantees"

## Gate 3 Validation Checklist

| Category | Requirements |
|----------|--------------|
| **Architecture Completeness** | All PRD features mapped; DDD boundaries; single clear responsibilities; stable interfaces |
| **Data Design** | Ownership explicit; models support PRD; consistency strategy defined; flows documented |
| **Quality Attributes** | Performance targets set; security addressed; scalability path clear; reliability defined |
| **Integration Readiness** | External deps identified (by capability); patterns selected (not tools); errors considered; versioning strategy exists |
| **Technology Agnostic** | Zero product names; capabilities abstract; patterns named not implementations; can swap tech without redesign |

**Gate Result:** ✅ PASS → API Design | ⚠️ CONDITIONAL (remove product names) | ❌ FAIL (too coupled)

## Common Violations

| Violation | Wrong | Correct |
|-----------|-------|---------|
| **Tech in Architecture** | `Language: Go 1.24+, Framework: Fiber v2.52+, Database: PostgreSQL 16` | `Style: Modular Monolith, Pattern: Hexagonal, Data Tier: Relational DB, Key-value store, Object storage` |
| **Framework in Components** | `Fiber middleware for JWT, bcrypt for passwords, passport.js for OAuth2` | `Auth Component: Purpose, Inbound (HTTP endpoints), Outbound (persistence, notifications), Security (token-based, industry-standard hashing)` |
| **Cloud Services in Deployment** | `Compute: AWS ECS Fargate, Database: AWS RDS, Cache: ElastiCache` | `Compute: Container-based stateless, Data Tier: Managed DB with backup, Performance: Distributed caching, Traffic: Load balanced with health checks` |

## Pagination Strategy (Required for List Endpoints)

If feature includes list/browse, decide during TRD:

| Strategy | Best For | Trade-off | Performance |
|----------|----------|-----------|-------------|
| **Cursor-Based** | >10k records, infinite scroll, real-time | Can't jump pages | O(1) |
| **Page-Based (Offset)** | <10k records, admin interfaces | Degrades with large offsets | O(n) |
| **Page-Based + Total Count** | "Page X of Y" UI | Additional COUNT query | 2 queries |
| **No Pagination** | Very small bounded datasets (<100) | All data at once | - |

Document in TRD: `API Patterns → Pagination → Strategy + Rationale`

## Authentication/Authorization Architecture (If Required)

If feature requires authentication or authorization (as determined in Question 2 of pre-dev command):

| Auth Type | TRD Description (Abstract) | Implementation Reference |
|-----------|---------------------------|-------------------------|
| User authentication only | "Token-based authentication with stateless validation" | For Go: `golang.md` → Access Manager Integration |
| User + permissions | "Token-based authentication with role-based access control (RBAC)" | For Go: `golang.md` → Access Manager Integration |
| Service-to-service | "Machine-to-machine authentication with client credentials" | For Go: `golang.md` → Access Manager Integration (GetApplicationToken) |
| Full (user + S2S) | "Dual-layer authentication: user tokens for end-users, client credentials for services" | For Go: `golang.md` → Access Manager Integration |

**Document in TRD:** `Security Architecture → Authentication/Authorization → Strategy + Implementation Reference`

**Key Implementation Pattern (for TRD reference):**
- Every protected endpoint requires middleware authorization
- Pattern: `auth.Authorize(applicationName, resource, action)` on each route
- Engineers will implement per-route protection following the referenced standard

**Note for Go Services:** Lerian's Access Manager (plugin-auth + identity + lib-auth) is the standard authentication system. Reference `golang.md` → Access Manager Integration section in the TRD so engineers know where to find implementation patterns including route middleware protection.

## License Manager Architecture (If Required)

If feature is a licensed product/plugin (as determined in Question 3 of pre-dev command):

| License Type | TRD Description (Abstract) | Implementation Reference |
|--------------|---------------------------|-------------------------|
| Single-org (global) | "Global license validation at service startup with fail-fast behavior" | For Go: `golang.md` → License Manager Integration |
| Multi-org | "Per-request license validation with organization context" | For Go: `golang.md` → License Manager Integration |

**Document in TRD:** `Security Architecture → Licensing → Strategy + Implementation Reference`

**Key Architecture Pattern (for TRD reference):**
- License validation as global middleware (applied early in chain)
- Fail-fast on startup: service refuses to start without valid license
- Graceful shutdown integration for license manager resources
- Built-in skip paths for health/readiness endpoints

**Note for Go Services:** Lerian's License Manager (lib-license-go) is the standard licensing system. Reference `golang.md` → License Manager Integration section in the TRD so engineers know where to find implementation patterns including global middleware and graceful shutdown.

## Frontend-Backend Integration Pattern (If Fullstack)

**⛔ HARD GATE:** If the feature is fullstack (`topology.scope: fullstack`), this section is MANDATORY in the TRD.

### Step 1: Read api_pattern from research.md

The api_pattern was determined during Topology Discovery (Q7) and persisted in research.md frontmatter.

```yaml
# From research.md frontmatter
topology:
  scope: fullstack
  api_pattern: direct | bff | other
```

### Step 2: Document Pattern in TRD

**TRD must include an `## Integration Patterns` section:**

```markdown
## Integration Patterns

### Frontend-Backend Communication

**Pattern:** [direct | bff | other]

**Rationale:** [Why this pattern was chosen]

**Architecture Implications:**
- [List architectural decisions driven by this pattern]
```

### Pattern-Specific Documentation

**If `api_pattern: direct`:**

```markdown
### Frontend-Backend Communication

**Pattern:** Direct API calls

**Rationale:** Single backend service, simple CRUD operations, no data aggregation needed.

**Architecture Implications:**
- Frontend components call backend API directly via fetch/axios
- No intermediate layer required
- Authentication tokens managed client-side (httpOnly cookies recommended)
- Error handling at component level

**Data Flow:**
Frontend Component → Backend API → Database
```

**If `api_pattern: bff`:**

```markdown
### Frontend-Backend Communication

**Pattern:** BFF (Backend-for-Frontend) layer

**Rationale:** [Multiple backend services | Complex data aggregation | Sensitive keys to hide | Request optimization needed]

**Architecture Implications:**
- Frontend calls BFF API routes (Next.js API Routes recommended)
- BFF aggregates data from multiple backend services
- Sensitive API keys stored server-side in BFF
- Response transformation happens in BFF layer
- Frontend receives optimized, frontend-specific data shapes

**Data Flow:**
Frontend Component → BFF API Route → Backend Service(s) → Database(s)

**BFF Responsibilities:**
- Data aggregation from multiple services
- Response transformation for frontend consumption
- Authentication token management (httpOnly cookies with Secure and SameSite attributes)
- Rate limiting and caching
- Error normalization
```

### Rationalization Table for Integration Patterns

| Excuse | Reality |
|--------|---------|
| "API pattern doesn't affect architecture" | Pattern determines data flow, error handling, and layer responsibilities. Document it. |
| "We can decide direct vs BFF later" | Architecture depends on this choice. Decide now to inform component design. |
| "BFF is overkill for our feature" | If research.md says `api_pattern: bff`, honor the decision. It was made with context. |
| "Just use direct for simplicity" | Simplicity isn't always correct. Follow the topology decision. |

## Design System & Styling (For Frontend Features)

**⛔ HARD GATE:** If the feature includes any user-facing UI, this section is MANDATORY.

### Step 1: Detect UI Library

**Auto-detection from package.json:**
| Package Present | UI Library |
|-----------------|------------|
| `@lerianstudio/sindarian-ui` | Sindarian UI (Radix-based) |
| `@radix-ui/*` | Radix UI Primitives |
| `@shadcn/ui` or `shadcn` in devDeps | shadcn/ui |
| `@chakra-ui/react` | Chakra UI |
| `@headlessui/react` | Headless UI |
| `@mui/material` | Material UI |
| None detected | Ask user for choice |

**If no UI library detected, AskUserQuestion:**
"What UI component library will this project use?"
Options: shadcn/ui (Recommended), Chakra UI, Material UI, Headless UI, Custom components

### Step 2: Document Styling Configuration

**TRD must include a `## Design System Configuration` section:**

```markdown
## Design System Configuration

### UI Library
- **Library:** [Detected or chosen library]
- **Version:** [Package version from package.json]

### CSS Framework
- **Framework:** [TailwindCSS v4 / CSS Modules / Styled Components / etc.]
- **Config File:** [tailwind.config.ts / postcss.config.js / etc.]

### Theme Integration
- **CSS Variables Required:** Yes/No
- **Dark Mode:** prefers-color-scheme / class-based / not supported
- **Source Directive:** @source path for component styles

### Required CSS Imports
List CSS files that MUST be imported in globals.css:
- `@import "tailwindcss";`
- `@import "@library/dist/components/ui/button/styles.css";`
- etc.

### Theme Variables
Document required CSS custom properties:
- Color scale (zinc, shadcn)
- Spacing variables
- Component-specific variables (button, input, dialog)
```

### Step 3: Component Availability Matrix

**TRD MUST document which components exist in the chosen UI library:**

```markdown
### Component Availability

| Component Needed | Available in Library | Notes |
|------------------|---------------------|-------|
| Button | ✅ Yes | Variants: primary, secondary, outline |
| Dialog/Modal | ✅ Yes | Use DialogTrigger pattern |
| Form | ✅ Yes | Requires Form context wrapper |
| Input | ⚠️ Partial | Requires FormField context |
| IconButton | ✅ Yes | Separate component for icon-only |
| Toast | ✅ Yes | Requires Toaster provider |
| Table | ✅ Yes | Use TableHeader, TableBody, etc. |

### Missing Components (Must Create)
- [ ] Component X - not available, need custom implementation
```

### Step 4: Variant Mapping

**TRD MUST document available variants to prevent implementation errors:**

```markdown
### Button Variants (Example)

| Design Intent | Correct Variant | WRONG (Don't Use) |
|---------------|-----------------|-------------------|
| Primary action | `variant="primary"` | `variant="default"` |
| Secondary action | `variant="secondary"` | - |
| Cancel/neutral | `variant="outline"` | `variant="ghost"` |
| Destructive | `variant="primary" className="bg-red-600"` | `variant="destructive"` |
| Icon-only | Use `<IconButton>` | `<Button size="icon">` |
```

### Rationalization Table for Design System

| Excuse | Reality |
|--------|---------|
| "Styling is implementation detail" | Styling bugs cause white screens and broken UX. Document upfront. |
| "Developers will figure out CSS" | Missing CSS variables = hours of debugging. Specify requirements. |
| "All UI libraries work the same" | They don't. Form patterns, variants, and contexts vary significantly. |
| "We'll add dark mode later" | CSS architecture for dark mode must be set from the start. |
| "Component variants are obvious" | They're not. `ghost` vs `plain` vs `outline` varies by library. |
| "Just use the library's defaults" | Defaults may not match design. Document exact variants needed. |

### Red Flags - STOP

If reviewing a TRD for a UI feature and you see NONE of these, **STOP and add them**:

- No UI library specified
- No CSS framework documented
- No theme variables listed
- No component availability matrix
- No variant mapping

### Gate 3 Validation Addition for UI Features

| Category | Requirements |
|----------|--------------|
| **Design System** | UI library specified; CSS framework documented; theme variables listed; component availability verified; variant mapping complete |

## ADR Template

```markdown
**ADR-00X: [Pattern Name]**
- **Context**: [Problem needing solution]
- **Options**: [List with trade-offs - no products]
- **Decision**: [Selected pattern]
- **Rationale**: [Why this pattern]
- **Consequences**: [Impact of decision]
```

## Confidence Scoring

| Factor | Points | Criteria |
|--------|--------|----------|
| Pattern Match | 0-40 | Exact before: 40, Similar: 25, Novel: 10 |
| Complexity Management | 0-30 | Simple proven: 30, Moderate: 20, High: 10 |
| Risk Level | 0-30 | Low proven: 30, Moderate mitigated: 20, High accepted: 10 |

**Action:** 80+ present autonomously | 50-79 present options | <50 request clarification

---

## Document Placement

**trd.md is a shared document** - it defines architecture for the entire feature.

| Structure | trd.md Location |
|-----------|-----------------|
| single-repo | `docs/pre-dev/{feature}/trd.md` |
| monorepo | `docs/pre-dev/{feature}/trd.md` (root) |
| multi-repo | Write to BOTH repos |

**Multi-repo handling:**

```bash
# Read topology from research.md frontmatter
if [[ "$structure" == "multi-repo" ]]; then
    # Write to both repositories
    mkdir -p "{backend.path}/docs/pre-dev/{feature}"
    mkdir -p "{frontend.path}/docs/pre-dev/{feature}"

    # Write TRD to primary (backend)
    # Then copy to frontend
    cp "{backend.path}/docs/pre-dev/{feature}/trd.md" "{frontend.path}/docs/pre-dev/{feature}/trd.md"
fi
```

**Sync footer for multi-repo:**
```markdown
---
**Sync Status:** Architecture document maintained in both repositories.
```

---

## Output & After Approval

**Output to:**
- **single-repo/monorepo:** `docs/pre-dev/{feature-name}/trd.md`
- **multi-repo:** Both `{backend.path}/docs/pre-dev/{feature}/trd.md` AND `{frontend.path}/docs/pre-dev/{feature}/trd.md`

1. ✅ Lock TRD - architecture patterns are now reference
2. 🎯 Use as input for API Design (`pre-dev-api-design`)
3. 🚫 Never add technologies retroactively
4. 📋 Keep architecture/implementation strictly separated

## The Bottom Line

**If you wrote a TRD with specific technology products, delete those sections and rewrite abstractly.**

The TRD is architecture patterns only. Period. No product names. No versions. No frameworks.

Technology choices go in Dependency Map. That's the next phase. Wait for it.

**Stay abstract. Stay flexible. Make technology decisions in the next phase with full analysis.**
