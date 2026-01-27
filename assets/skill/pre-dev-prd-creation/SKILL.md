---
name: ring:pre-dev-prd-creation
description: |
  Gate 1: Business requirements document - defines WHAT/WHY before HOW.
  Creates PRD with problem definition, user stories, success metrics.

trigger: |
  - Starting new product or major feature
  - User asks to "plan", "design", or "architect"
  - About to write code without documented requirements
  - Asked to create PRD or requirements document

skip_when: |
  - PRD already exists and validated → proceed to Gate 2
  - Pure technical task without business impact → TRD directly
  - Bug fix → systematic-debugging

sequence:
  before: [pre-dev-feature-map, pre-dev-trd-creation]
---

# PRD Creation - Business Before Technical

## Foundational Principle

**Business requirements (WHAT/WHY) must be fully defined before technical decisions (HOW/WHERE).**

Mixing business and technical concerns creates:
- Requirements that serve implementation convenience, not user needs
- Technical constraints that limit product vision
- Inability to evaluate alternatives objectively
- Cascade failures when requirements change

**The PRD answers**: WHAT we're building and WHY it matters to users and business.
**The PRD never answers**: HOW we'll build it or WHERE components will live.

## Mandatory Workflow

| Phase | Activities |
|-------|------------|
| **0. Load Research** | Check `docs/pre-dev/{feature}/research.md`; review codebase patterns, best practices, framework constraints, UX research; reference findings with `file:line` notation |
| **1. Problem Discovery** | Define problem without solution bias; identify specific users; quantify pain with metrics/evidence |
| **2. Business Requirements** | Executive summary (3 sentences); user personas (goals, frustrations); user stories (As/I want/So that); success metrics (measurable); scope boundaries (in/out) |
| **3. Gate 1 Validation** | Problem articulated; impact quantified; users identified; features address problem; metrics measurable; scope explicit |
| **4. UX Validation** | Dispatch `product-designer` to validate PRD against user needs and create `ux-criteria.md` |

## Explicit Rules

### ✅ DO Include in PRD
Problem definition and user pain points, user personas (demographics, goals, frustrations), user stories with acceptance criteria, feature requirements (WHAT not HOW), success metrics (adoption, satisfaction, KPIs), scope boundaries (in/out explicitly), go-to-market considerations

### ❌ NEVER Include in PRD
Architecture diagrams or component design, technology choices (languages, frameworks, databases), implementation approaches or algorithms, database schemas or API specifications, code examples or package dependencies, infrastructure needs or deployment strategies, system integration patterns

### Separation Rules
1. **If it's a technology name** → Not in PRD (goes in Dependency Map)
2. **If it's a "how to build"** → Not in PRD (goes in TRD)
3. **If it's implementation** → Not in PRD (goes in Tasks/Subtasks)
4. **If it describes system behavior** → Not in PRD (goes in TRD)

## Rationalization Table

| Excuse | Reality |
|--------|---------|
| "Just a quick technical note won't hurt" | Technical details constrain business thinking. Keep them separate. |
| "Stakeholders need to know it's feasible" | Feasibility comes in TRD after business requirements are locked. |
| "The implementation is obvious" | Obvious to you ≠ obvious to everyone. Separate concerns. |
| "I'll save time by combining PRD and TRD" | You'll waste time rewriting when requirements change. |
| "This is a simple feature, no need for formality" | Simple features still need clear requirements. Follow the process. |
| "I can skip Gate 1, I know it's good" | Gates exist because humans are overconfident. Validate. |
| "The problem is obvious, no need for personas" | Obvious to you ≠ validated with users. Document it. |
| "Success metrics can be defined later" | Defining metrics later means building without targets. Do it now. |
| "I'll just add this one API endpoint detail" | API design is technical architecture. Stop. Keep it in TRD. |
| "But we already decided on PostgreSQL" | Technology decisions come after business requirements. Wait. |
| "CEO/CTO says it's a business constraint" | Authority doesn't change what's technical. Abstract it anyway. |
| "Investors need to see specific vendors/tech" | Show phasing and constraints abstractly. Vendors go in TRD. |
| "This is product scoping, not technical design" | Scope = capabilities. Technology = implementation. Different things. |
| "Mentioning Stripe shows we're being practical" | Mentioning "payment processor" shows the same. Stay abstract. |
| "PRDs can mention tech when it's a constraint" | PRDs mention capabilities needed. TRD maps capabilities to tech. |
| "Context matters - this is for exec review" | Context doesn't override principles. Executives get abstracted version. |

## Security Requirements Discovery (Business Level)

**During PRD creation, identify if the feature requires access control:**

| Business Question | If Yes → Document |
|-------------------|-------------------|
| Does this feature handle user-specific data? | "Users can only access their own [data type]" |
| Are there different user roles with different permissions? | "Admins can [X], regular users can [Y]" |
| Does this feature need to identify who performed an action? | "Audit trail required for [action type]" |
| Does this integrate with other internal services? | "Service must authenticate to [service name]" |
| Are there regulatory requirements (GDPR, PCI-DSS, HIPAA)? | "Must comply with [regulation] for [data type]" |

**What to include in PRD:**
- ✅ "Only authenticated users can access this feature"
- ✅ "Users can only view/edit their own records"
- ✅ "Admin approval required for [action]"
- ✅ "Must track who performed each action"

**What NOT to include in PRD:**
- ❌ "Use JWT tokens" (technology choice → TRD)
- ❌ "Integrate with Access Manager" (architecture → TRD)
- ❌ "OAuth2 flow" (protocol choice → TRD)

**Note:** The TRD (Gate 3) will translate these business requirements into authentication/authorization architecture patterns. For Go services, refer to `golang.md` → Access Manager Integration section during TRD creation.

---

## Red Flags - STOP

If you catch yourself writing or thinking any of these in a PRD, **STOP**:

- Technology product names (PostgreSQL, Redis, Kafka, AWS, etc.)
- Framework or library names (React, Fiber, Express, etc.)
- Words like: "architecture", "component", "service", "endpoint", "schema"
- Phrases like: "we'll use X to do Y" or "the system will store data in Z"
- Code examples or API specifications
- "How we'll implement" or "Technical approach"
- Database table designs or data models
- Integration patterns or protocols

**When you catch yourself**: Move that content to a "technical notes" section to transfer to TRD later. Keep PRD pure business.

## Gate 1 Validation Checklist

| Category | Requirements |
|----------|--------------|
| **Problem Definition** | Problem articulated (1-2 sentences); impact quantified/qualified; users specifically identified; current workarounds documented |
| **Solution Value** | Features address core problem; success metrics measurable; ROI case documented; user value clear per feature |
| **Scope Clarity** | In-scope items explicit; out-of-scope with rationale; assumptions documented; business dependencies identified |
| **Market Fit** | Differentiation clear; value proposition validated; business case sound; go-to-market outlined |

**Gate Result:** ✅ PASS → UX Validation → Feature Map | ⚠️ CONDITIONAL (address gaps) | ❌ FAIL (return to discovery)

## Phase 4: UX Validation + Wireframes

**After PRD passes Gate 1 validation, dispatch product-designer for UX validation:**

```
Task(
  subagent_type="ring:product-designer",
  model="opus",
  prompt="Validate PRD at docs/pre-dev/{feature}/prd.md against user needs. Mode: ux-validation.

  UI Configuration (from pre-dev command):
  - UI Library: {ui_library}  // e.g., shadcn/ui, Chakra UI, or auto-detected
  - Styling: {styling}        // e.g., TailwindCSS, CSS Modules, or auto-detected

  Create ux-criteria.md with: problem validation status, refined personas, UX acceptance criteria (functional, usability, accessibility, responsive).

  If feature has UI components, also create wireframes/ directory with low-fidelity prototypes using the specified UI library components and styling approach."
)
```

**IMPORTANT: Pass UI Configuration to product-designer**

The UI Library and Styling choices (from `/pre-dev-feature` or `/pre-dev-full` questions) MUST be passed to product-designer:
- If user selected a library → Use that library's component names in wireframes
- If auto-detected from package.json → Use the detected library
- If "Custom components only" → Use generic component names

This ensures wireframes reference real components that will be available during implementation.

**UX Validation Outputs:**
- `docs/pre-dev/{feature}/ux-criteria.md` - UX acceptance criteria
- `docs/pre-dev/{feature}/wireframes/` - Low-fidelity prototypes (if feature has UI)
  - `{screen-name}.yaml` - YAML wireframe specification per screen
  - `user-flows.md` - User flow diagrams with state transitions

**UI Detection Rule:**
If PRD contains any of these, feature HAS UI and wireframes are REQUIRED:
- User stories mentioning: "see", "view", "click", "navigate", "page", "screen", "button", "form"
- Features involving: login, dashboard, settings, profile, reports, notifications
- Any user-facing interaction

**UX Validation Checklist:**

| Check | Required | Condition |
|-------|----------|-----------|
| Problem validation status documented | Yes | Always |
| Personas refined based on PRD | Yes | Always |
| Functional UX criteria defined | Yes | Always |
| Usability criteria defined | Yes | Always |
| Accessibility criteria defined | Yes | Always |
| Responsive criteria defined | Yes | Always |
| All PRD user stories have UX criteria | Yes | Always |
| Wireframes created for each screen | Yes | If feature has UI |
| User flows documented | Yes | If feature has UI |
| State coverage table complete | Yes | If feature has UI |

**Wireframe YAML Format:**
```yaml
screen: Screen Name
route: /path
layout: layout-type
components:
  - id: component-id
    type: component-type
    # ... component specs
states:
  default: { ... }
  loading: { ... }
  error: { ... }
responsive:
  mobile: { ... }
  desktop: { ... }
accessibility:
  keyboard: [ ... ]
  screen-reader: [ ... ]
```

### Document Placement (based on topology.structure)

**prd.md placement:**

| Structure | prd.md Location |
|-----------|-----------------|
| single-repo | `docs/pre-dev/{feature}/prd.md` |
| monorepo | `docs/pre-dev/{feature}/prd.md` (root) |
| multi-repo | Write to BOTH repos |

**ux-criteria.md and wireframes/ placement:**

| Structure | Location |
|-----------|----------|
| single-repo | `docs/pre-dev/{feature}/` |
| monorepo | `{frontend.path}/docs/pre-dev/{feature}/` |
| multi-repo | `{frontend.path}/docs/pre-dev/{feature}/` |

**Why frontend path for UX docs?** UX criteria and wireframes are consumed by frontend engineers. Placing them in the frontend module/repo ensures they are discoverable where they'll be used.

**Directory creation for multi-module:**
```bash
# Read topology from research.md frontmatter
# Create appropriate directories:

# For monorepo - frontend module
mkdir -p "{frontend.path}/docs/pre-dev/{feature}"

# For multi-repo - both repos for prd.md, frontend for UX
mkdir -p "{backend.path}/docs/pre-dev/{feature}"
mkdir -p "{frontend.path}/docs/pre-dev/{feature}"
```

**If UX validation fails:**
- Conflicting user needs → Return to Phase 1 (Problem Discovery)
- Missing persona details → Enrich PRD personas
- Unclear acceptance criteria → Iterate with product-designer
- Missing wireframes for UI feature → product-designer must create them

## Common Violations

| Violation | Wrong | Correct |
|-----------|-------|---------|
| **Tech in Features** | "FR-001: Use JWT tokens for session, bcrypt for passwords, OAuth2 with Google" | "FR-001: Users can create accounts and securely log in. Value: Access personalized content. Success: 95% authenticate first attempt" |
| **Implementation in Stories** | "As user, I want to store data in PostgreSQL so queries are fast" | "As user, I want dashboard to load in <2 seconds so I can quickly access information" |
| **Architecture in Problem** | "Our microservices architecture doesn't support real-time notifications" | "Users miss important updates because they must manually refresh. 78% report missing time-sensitive info" |
| **Authority-Based Bypass** | "MVP: Stripe for payments, PostgreSQL (we already use it)" | "Phase 1: Integrate with existing payment vendor (2-week timeline); leverage existing database infrastructure. TRD will document specific vendor selection" |

## Confidence Scoring

| Factor | Points | Criteria |
|--------|--------|----------|
| Market Validation | 0-25 | Direct user feedback: 25, Market research: 15, Assumptions: 5 |
| Problem Clarity | 0-25 | Quantified pain: 25, Qualitative evidence: 15, Hypothetical: 5 |
| Solution Fit | 0-25 | Proven pattern: 25, Adjacent pattern: 15, Novel: 5 |
| Business Value | 0-25 | Clear ROI: 25, Indirect value: 15, Uncertain: 5 |

**Action:** 80+ autonomous | 50-79 present options | <50 ask discovery questions

## Output & After Approval

**Outputs (paths depend on topology.structure):**

| Document | single-repo | monorepo | multi-repo |
|----------|-------------|----------|------------|
| prd.md | `docs/pre-dev/{feature}/` | `docs/pre-dev/{feature}/` | Both repos |
| ux-criteria.md | `docs/pre-dev/{feature}/` | `{frontend.path}/docs/pre-dev/{feature}/` | Frontend repo |
| wireframes/ | `docs/pre-dev/{feature}/` | `{frontend.path}/docs/pre-dev/{feature}/` | Frontend repo |

1. ✅ Lock the PRD - no changes without formal amendment
2. ✅ Lock ux-criteria.md - defines UX acceptance for implementation
3. ✅ Lock wireframes/ - defines visual structure for ui-engineer
4. 🎯 Use all as input for Feature Map (`pre-dev-feature-map`) or TRD (`pre-dev-trd-creation`)
5. 🚫 Never add technical details retroactively
6. 📋 Keep business/technical strictly separated

## The Bottom Line

**If you wrote a PRD with technical details, delete it and start over.**

The PRD is business-only. Period. No exceptions. No "just this once". No "but it's relevant".

Technical details go in TRD. That's the next phase. Wait for it.

**Follow the separation. Your future self will thank you.**
