---
name: production-readiness-audit
description: Comprehensive 27-dimension production readiness audit. Runs in batches of 10 explorers, appending incrementally to a single report file. Categories - Structure (pagination, errors, routes, bootstrap, runtime), Security (auth, IDOR, SQL, validation, rate limiting), Operations (telemetry, health, config, connections, logging), Quality (idempotency, docs, debt, testing, dependencies, performance, concurrency, migrations), Infrastructure (containers, hardening, cicd, async). Produces scored report (0-270) with severity ratings.
allowed-tools: Task, Read, Glob, Grep, Write, TodoWrite
---

# Production Readiness Audit

A comprehensive, multi-agent audit system that evaluates codebase production readiness across **27 dimensions in 5 categories**. This skill runs explorer agents in **batches of 10**, appending results incrementally to a single report file to prevent context bloat while maintaining thorough coverage.

## When This Skill Activates

Use this skill when:

- Preparing for production deployment
- Conducting periodic security/quality reviews
- Onboarding to understand codebase health
- Evaluating technical debt before major releases
- Validating compliance with engineering standards
- Assessing a codebase's maturity level

## Audit Dimensions

The skill evaluates **23 distinct dimensions** across 4 categories, each handled by a specialized explorer agent:

### Category A: Code Structure & Patterns (5 dimensions)

| # | Dimension | Focus Area |
|---|-----------|------------|
| 1 | **Pagination Standards** | Cursor vs offset pagination, limit validation, response structure |
| 2 | **Error Framework** | pkg/assert usage, domain errors, HTTP error responses, error propagation |
| 3 | **Route Organization** | Hexagonal structure, handler construction, route registration |
| 4 | **Bootstrap & Initialization** | Staged startup, cleanup handlers, graceful shutdown |
| 5 | **Runtime Safety** | pkg/runtime usage, panic recovery, production mode handling |

### Category B: Security & Access Control (5 dimensions)

| # | Dimension | Focus Area |
|---|-----------|------------|
| 6 | **Auth Protection** | Route protection, JWT validation, tenant extraction |
| 7 | **IDOR & Access Control** | Ownership verification, tenant isolation, resource authorization |
| 8 | **SQL Safety** | Parameterized queries, identifier escaping, injection prevention |
| 9 | **Input Validation** | Request body validation, query params, VO validation |
| 10 | **Rate Limiting** | Global/per-endpoint limits, tenant/user keying, abuse prevention |

### Category C: Operational Readiness (5 dimensions)

| # | Dimension | Focus Area |
|---|-----------|------------|
| 11 | **Telemetry & Observability** | OpenTelemetry integration, tracing, metrics, structured logging |
| 12 | **Health Checks** | Liveness/readiness probes, dependency health, degraded status |
| 13 | **Configuration Management** | Env var validation, production constraints, secrets handling |
| 14 | **Connection Management** | DB/Redis pool settings, timeouts, replica support |
| 15 | **Logging & PII Safety** | Structured logging, sensitive data protection, log levels |

### Category D: Quality & Maintainability (8 dimensions)

| # | Dimension | Focus Area |
|---|-----------|------------|
| 16 | **Idempotency** | Idempotency keys, retry safety, duplicate prevention |
| 17 | **API Documentation** | Swagger/OpenAPI annotations, response schemas, examples |
| 18 | **Technical Debt** | TODOs, FIXMEs, deprecated code, incomplete implementations |
| 19 | **Testing Coverage** | Co-located tests, mockgen, table-driven tests, integration tests |
| 20 | **Dependency Management** | Pinned versions, CVE scanning, deprecated packages |
| 21 | **Performance Patterns** | N+1 queries, SELECT *, slice pre-allocation, batching |
| 22 | **Concurrency Safety** | Race conditions, goroutine leaks, mutex usage, worker pools |
| 23 | **Migration Safety** | Up/down pairs, CONCURRENTLY indexes, NOT NULL defaults |

### Category E: Infrastructure & Hardening (4 dimensions)

| # | Dimension | Focus Area |
|---|-----------|------------|
| 24 | **Container Security** | Dockerfile best practices, non-root user, multi-stage, image pinning |
| 25 | **HTTP Hardening** | Security headers (HSTS, CSP), cookie attributes, CORS strictness |
| 26 | **CI/CD & Licensing** | Pipeline definitions, license headers, artifact signing, automated tests |
| 27 | **Async Reliability** | DLQs, retry policies, consumer group usage, message durability |

## Execution Protocol

This skill runs **27 explorer agents in batches of 10**, writing results incrementally to a single report file. This prevents context bloat while maintaining comprehensive coverage.

### Output File

All results are appended to: `docs/audits/production-readiness-{YYYY-MM-DD}-{hh:mm}.md`

### Batch Execution Schedule

| Batch | Agents | Category Focus |
|-------|--------|----------------|
| 1 | 1-10 | Pagination, Error Framework, Route Org, Bootstrap, Runtime Safety, Auth, IDOR, SQL Safety, Input Validation, Rate Limiting |
| 2 | 11-20 | Telemetry, Health Checks, Config, Connection Mgmt, Logging/PII, Idempotency, API Docs, Technical Debt, Testing, Dependencies |
| 3 | 21-27 | Performance, Concurrency, Migration Safety, Container Security, HTTP Hardening, CI/CD, Async Reliability + Final Summary |

### Step-by-Step Execution

#### Step 1: Initialize Report File

```markdown
Write to docs/audits/production-readiness-{YYYY-MM-DD}-{hh:mm}.md:

# Production Readiness Audit Report

**Date:** {YYYY-MM-DD}
**Hour:** {hh:mm}
**Codebase:** {project-name}
**Auditor:** Claude Code (Production Readiness Skill v2.0)
**Status:** In Progress...

---
```

#### Step 2: Execute Batch 1 (Agents 1-10)

Launch 10 explorers in parallel:
```
Task(subagent_type="Explore", prompt="<Agent 1: Pagination Standards>")
Task(subagent_type="Explore", prompt="<Agent 2: Error Framework>")
Task(subagent_type="Explore", prompt="<Agent 3: Route Organization>")
Task(subagent_type="Explore", prompt="<Agent 4: Bootstrap & Init>")
Task(subagent_type="Explore", prompt="<Agent 5: Runtime Safety>")
Task(subagent_type="Explore", prompt="<Agent 6: Auth Protection>")
Task(subagent_type="Explore", prompt="<Agent 7: IDOR Protection>")
Task(subagent_type="Explore", prompt="<Agent 8: SQL Safety>")
Task(subagent_type="Explore", prompt="<Agent 9: Input Validation>")
Task(subagent_type="Explore", prompt="<Agent 10: Rate Limiting>")
```

**After completion:** Append results to the report file.

#### Step 3: Execute Batch 2 (Agents 11-20)

Launch 10 explorers in parallel:
```
Task(subagent_type="Explore", prompt="<Agent 11: Telemetry & Observability>")
Task(subagent_type="Explore", prompt="<Agent 12: Health Checks>")
Task(subagent_type="Explore", prompt="<Agent 13: Configuration Management>")
Task(subagent_type="Explore", prompt="<Agent 14: Connection Management>")
Task(subagent_type="Explore", prompt="<Agent 15: Logging & PII Safety>")
Task(subagent_type="Explore", prompt="<Agent 16: Idempotency>")
Task(subagent_type="Explore", prompt="<Agent 17: API Documentation>")
Task(subagent_type="Explore", prompt="<Agent 18: Technical Debt>")
Task(subagent_type="Explore", prompt="<Agent 19: Testing Coverage>")
Task(subagent_type="Explore", prompt="<Agent 20: Dependency Management>")
```

**After completion:** Append results to the report file.

#### Step 4: Execute Batch 3 (Agents 21-27 + Summary)

Launch 7 explorers in parallel:
```
Task(subagent_type="Explore", prompt="<Agent 21: Performance Patterns>")
Task(subagent_type="Explore", prompt="<Agent 22: Concurrency Safety>")
Task(subagent_type="Explore", prompt="<Agent 23: Migration Safety>")
Task(subagent_type="Explore", prompt="<Agent 24: Container Security>")
Task(subagent_type="Explore", prompt="<Agent 25: HTTP Hardening>")
Task(subagent_type="Explore", prompt="<Agent 26: CI/CD & Licensing>")
Task(subagent_type="Explore", prompt="<Agent 27: Async Reliability>")
```

**After completion:** Append results to the report file.

#### Step 5: Finalize Report

1. Read the complete report file
2. Calculate scores for each dimension
3. Generate Executive Summary with totals
4. Prepend Executive Summary to the report
5. Add remediation priorities
6. Present verbal summary to user

---

## Explorer Agent Prompts

### Agent 1: Pagination Standards Auditor

```prompt
Audit pagination implementation across the codebase for production readiness.

**Search Patterns:**
- Files: `**/pagination*.go`, `**/handlers.go`, `**/dto.go`
- Keywords: `limit`, `offset`, `cursor`, `NextCursor`, `PrevCursor`, `HasMore`

**Reference Implementation (GOOD):**
```go
// Cursor-based pagination with proper validation
type CursorResponse struct {
    NextCursor string `json:"next_cursor,omitempty"`
    PrevCursor string `json:"prev_cursor,omitempty"`
    Limit      int    `json:"limit"`
    HasMore    bool   `json:"has_more"`
}

// Limit validation with ceiling
if limit > maxLimit {
    limit = maxLimit // Auto-cap, don't error
}
if limit < 1 {
    return ErrLimitMustBePositive
}
```

**Check For:**
1. Consistent pagination response structure across all list endpoints
2. Maximum limit enforcement (typically 100-200)
3. Cursor-based pagination for real-time data (preferred over offset)
4. Proper error handling for invalid pagination params
5. Default values when params missing

**Severity Ratings:**
- CRITICAL: No limit validation (allows unlimited queries)
- HIGH: Inconsistent pagination structures across endpoints
- MEDIUM: Using offset pagination for frequently-changing data
- LOW: Missing cursor pagination where beneficial

**Output Format:**
```
## Pagination Audit Findings

### Summary
- Total list endpoints: X
- Using cursor pagination: Y
- Using offset pagination: Z
- Missing limit validation: N

### Critical Issues
[file:line] - Description

### Recommendations
1. ...
```
```

### Agent 2: Telemetry & Observability Auditor

```prompt
Audit telemetry and observability implementation for production readiness.

**Search Patterns:**
- Files: `**/observability*.go`, `**/telemetry*.go`, `**/handlers.go`
- Keywords: `NewTrackingFromContext`, `tracer.Start`, `span`, `logger`, `metrics`

**Reference Implementation (GOOD):**
```go
// Handler with proper telemetry
func (h *Handler) DoSomething(c *fiber.Ctx) error {
    ctx := c.UserContext()
    logger, tracer, headerID, _ := libCommons.NewTrackingFromContext(ctx)
    ctx, span := tracer.Start(ctx, "handler.DoSomething")
    defer span.End()

    span.SetAttributes(attribute.String("request_id", headerID))

    // On error
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
    logger.Errorf("operation failed: %v", err)

    return nil
}
```

**Check For:**
1. All handlers start spans with descriptive names
2. Errors recorded to spans before returning
3. Request IDs propagated through context
4. Metrics initialized at startup
5. Structured logging with context (not fmt.Println)
6. Graceful telemetry shutdown

**Severity Ratings:**
- CRITICAL: No tracing in handlers
- HIGH: Errors not recorded to spans
- MEDIUM: Missing request ID propagation
- LOW: Inconsistent span naming conventions

**Output Format:**
```
## Telemetry Audit Findings

### Summary
- Handlers with tracing: X/Y
- Handlers with error recording: X/Y
- Metrics initialization: Yes/No

### Critical Issues
[file:line] - Description

### Recommendations
1. ...
```
```

### Agent 3: Error Framework Auditor

```prompt
Audit error handling framework usage for production readiness.

**Search Patterns:**
- Files: `**/assert*.go`, `**/error*.go`, `**/handlers.go`
- Keywords: `assert.New`, `AssertionError`, `ErrRepo`, `errors.Is`, `errors.As`
- Also search: `panic(`, `log.Fatal`

**Reference Implementation (GOOD):**
```go
// Using pkg/assert for validation
asserter := assert.New(ctx, logger, "module", "operation")
if err := asserter.NotNil(ctx, config, "config required"); err != nil {
    return fmt.Errorf("validation: %w", err)
}

// Domain error types
var (
    ErrNotFound        = errors.New("resource not found")
    ErrInvalidInput    = errors.New("invalid input")
)

// Error mapping in handlers
if errors.Is(err, domain.ErrNotFound) {
    return httputil.NotFoundError(c, span, logger, "resource not found", err)
}
```

**Reference Implementation (BAD):**
```go
// Direct panic in production code
if config == nil {
    panic("config is nil")  // BAD: Use assert or return error
}

// Swallowing errors
result, _ := doSomething()  // BAD: Ignoring error

// Generic error messages
return errors.New("error")  // BAD: Not descriptive
```

**Check For:**
1. pkg/assert used instead of panic for validation
2. Named error variables (sentinel errors) per module
3. Proper error wrapping with %w
4. errors.Is/errors.As for error matching
5. No panic() in non-test production code
6. No swallowed errors (_, err := ignored)

**Severity Ratings:**
- CRITICAL: panic() in production code paths
- CRITICAL: Swallowed errors in critical paths
- HIGH: Generic error messages without context
- MEDIUM: Inconsistent error types across modules
- LOW: Missing error wrapping context

**Output Format:**
```
## Error Framework Audit Findings

### Summary
- Modules using pkg/assert: X
- Panic calls in production: Y
- Swallowed errors: Z

### Critical Issues
[file:line] - Description

### Recommendations
1. ...
```
```

### Agent 4: Route Organization Auditor

```prompt
Audit route organization and handler structure for production readiness.

**Search Patterns:**
- Files: `**/routes.go`, `**/handlers.go`, `internal/**/adapters/http/*.go`
- Keywords: `RegisterRoutes`, `protected(`, `fiber.Router`, `NewHandler`

**Reference Implementation (GOOD):**
```go
// Centralized route registration
func RegisterRoutes(protected func(resource, action string) fiber.Router, handler *Handler) error {
    if handler == nil {
        return errors.New("handler is nil")
    }
    protected("resource", "create").Post("/v1/resources", handler.Create)
    protected("resource", "read").Get("/v1/resources", handler.List)
    protected("resource", "read").Get("/v1/resources/:id", handler.Get)
    return nil
}

// Handler constructor with validation
func NewHandler(deps ...interface{}) (*Handler, error) {
    if dep == nil {
        return nil, ErrNilDependency
    }
    return &Handler{...}, nil
}
```

**Check For:**
1. Hexagonal structure: `internal/{module}/adapters/http/`
2. Centralized route registration per module
3. Handler constructors validate all dependencies
4. Consistent URL patterns (v1, kebab-case, plural resources)
5. All routes use protected() wrapper (no public endpoints)
6. Clear separation: routes.go vs handlers.go

**Severity Ratings:**
- CRITICAL: Unprotected routes (missing auth middleware)
- HIGH: Scattered route definitions
- MEDIUM: Handler accepts nil dependencies
- LOW: Inconsistent URL naming conventions

**Output Format:**
```
## Route Organization Audit Findings

### Summary
- Modules following hexagonal: X/Y
- Routes with protection: X/Y
- Handlers validating deps: X/Y

### Critical Issues
[file:line] - Description

### Recommendations
1. ...
```
```

### Agent 5: Bootstrap & Initialization Auditor

```prompt
Audit application bootstrap and initialization for production readiness.

**Search Patterns:**
- Files: `**/main.go`, `**/init.go`, `**/bootstrap/*.go`
- Keywords: `InitServers`, `startupSucceeded`, `defer`, `cleanup`, `graceful`

**Reference Implementation (GOOD):**
```go
// Staged initialization with cleanup
func InitServers(opts *Options) (*Service, error) {
    startupSucceeded := false
    defer func() {
        if !startupSucceeded {
            cleanupConnections(...)  // Only cleanup on failure
        }
    }()

    // 1. Load config
    cfg, err := loadConfig()
    if err != nil {
        return nil, fmt.Errorf("config: %w", err)
    }

    // 2. Initialize logger
    logger := initLogger(cfg)

    // 3. Initialize telemetry
    telemetry := initTelemetry(cfg, logger)

    // 4. Connect infrastructure (DB, Redis, MQ)
    db, err := connectDB(cfg)
    if err != nil {
        return nil, fmt.Errorf("database: %w", err)
    }

    // 5. Initialize modules in dependency order
    ...

    startupSucceeded = true
    return &Service{...}, nil
}
```

**Check For:**
1. Staged initialization (config → logger → telemetry → infra)
2. Cleanup handlers for failed startup
3. Graceful shutdown support
4. Module initialization in dependency order
5. Error propagation (not just logging and continuing)
6. Production vs development mode handling

**Severity Ratings:**
- CRITICAL: No graceful shutdown
- HIGH: Resources not cleaned up on startup failure
- HIGH: Errors logged but not returned
- MEDIUM: Initialization order issues
- LOW: Missing development mode toggles

**Output Format:**
```
## Bootstrap Audit Findings

### Summary
- Graceful shutdown: Yes/No
- Cleanup on failure: Yes/No
- Staged initialization: Yes/No

### Critical Issues
[file:line] - Description

### Recommendations
1. ...
```
```

### Agent 6: Auth Protection Auditor

```prompt
Audit authentication and authorization implementation for production readiness.

**Search Patterns:**
- Files: `**/auth/*.go`, `**/middleware*.go`, `**/routes.go`
- Keywords: `Authorize`, `protected`, `JWT`, `tenant`, `ExtractToken`

**Reference Implementation (GOOD):**
```go
// Protected route group
protected := func(resource, action string) fiber.Router {
    return auth.ProtectedGroup(api, authClient, tenantExtractor, resource, action)
}

// All routes use protected
protected("contexts", "create").Post("/v1/config/contexts", handler.Create)

// JWT validation
func parseTokenClaims(tokenString string, secret []byte) (jwt.MapClaims, error) {
    parser := jwt.NewParser(jwt.WithValidMethods(validSigningMethods))
    token, err := parser.ParseWithClaims(...)
    if err != nil || !token.Valid {
        return nil, ErrInvalidToken
    }
    // Check expiration
    if exp, ok := claims["exp"].(float64); ok {
        if time.Now().Unix() > int64(exp) {
            return nil, ErrTokenExpired
        }
    }
    return claims, nil
}
```

**Check For:**
1. ALL routes protected (no public endpoints without auth)
2. Resource/action authorization granularity
3. JWT signature validation (not just parsing)
4. Token expiration enforcement
5. Tenant extraction from JWT claims
6. Auth bypass for health/ready endpoints only

**Severity Ratings:**
- CRITICAL: Unprotected data endpoints
- CRITICAL: JWT parsed but not validated
- HIGH: Missing token expiration check
- HIGH: Tenant claims not enforced
- MEDIUM: Overly broad permissions
- LOW: Missing fine-grained actions

**Output Format:**
```
## Auth Protection Audit Findings

### Summary
- Protected routes: X/Y
- JWT validation: Complete/Partial/Missing
- Tenant enforcement: Yes/No

### Critical Issues
[file:line] - Description

### Recommendations
1. ...
```
```

### Agent 7: IDOR & Access Control Auditor

```prompt
Audit IDOR (Insecure Direct Object Reference) protection for production readiness.

**Search Patterns:**
- Files: `**/verifier*.go`, `**/handlers.go`, `**/context.go`
- Keywords: `VerifyOwnership`, `tenantID`, `contextID`, `ParseAndVerify`

**Reference Implementation (GOOD):**
```go
// 4-layer IDOR protection
func ParseAndVerifyContextParam(fiberCtx *fiber.Ctx, verifier ContextOwnershipVerifier) (uuid.UUID, uuid.UUID, error) {
    // 1. UUID format validation
    contextID, err := uuid.Parse(fiberCtx.Params("contextId"))
    if err != nil {
        return uuid.Nil, uuid.Nil, ErrInvalidID
    }

    // 2. Extract tenant from auth context (cannot be spoofed)
    tenantID := auth.GetTenantID(ctx)

    // 3. Database query filtered by tenant
    // 4. Post-query ownership verification
    if err := verifier.VerifyOwnership(ctx, tenantID, contextID); err != nil {
        return uuid.Nil, uuid.Nil, err
    }
    return contextID, tenantID, nil
}

// Verifier implementation
func (v *verifier) VerifyOwnership(ctx context.Context, tenantID, resourceID uuid.UUID) error {
    resource, err := v.query.Get(ctx, tenantID, resourceID)  // Query WITH tenant filter
    if errors.Is(err, sql.ErrNoRows) {
        return ErrNotFound
    }
    if resource.TenantID != tenantID {  // Double-check ownership
        return ErrNotOwned
    }
    return nil
}
```

**Reference Implementation (BAD):**
```go
// BAD: No ownership verification
func GetResource(c *fiber.Ctx) error {
    id := c.Params("id")
    resource, err := repo.FindByID(ctx, id)  // No tenant filter!
    return c.JSON(resource)
}
```

**Check For:**
1. All resource access verifies ownership
2. Tenant ID from JWT context (not request params)
3. Database queries include tenant filter
4. Post-query ownership double-check
5. UUID validation before database lookup
6. Consistent verifier pattern across modules

**Severity Ratings:**
- CRITICAL: Resource access without ownership check
- CRITICAL: Tenant ID from user input (not JWT)
- HIGH: Missing post-query ownership verification
- MEDIUM: Inconsistent verifier implementation
- LOW: Missing UUID format validation

**Output Format:**
```
## IDOR Protection Audit Findings

### Summary
- Modules with verifiers: X/Y
- Multi-tenant filtered queries: X/Y
- Post-query verification: X/Y

### Critical Issues
[file:line] - Description

### Recommendations
1. ...
```
```

### Agent 8: SQL Safety Auditor

```prompt
Audit SQL injection prevention for production readiness.

**Search Patterns:**
- Files: `**/*.postgresql.go`, `**/repository/*.go`, `**/*_repo.go`
- Keywords: `ExecContext`, `QueryContext`, `Exec(`, `Query(`, `$1`, `$2`
- Also search for: String concatenation in SQL: `"SELECT.*" +`, `fmt.Sprintf.*SELECT`

**Reference Implementation (GOOD):**
```go
// Parameterized queries
query := `INSERT INTO resources (id, name, tenant_id) VALUES ($1, $2, $3)`
_, err = tx.ExecContext(ctx, query, id, name, tenantID)

// SQL identifier escaping for dynamic schemas
func QuoteIdentifier(identifier string) string {
    return "\"" + strings.ReplaceAll(identifier, "\"", "\"\"") + "\""
}
schemaQuery := "SET LOCAL search_path TO " + QuoteIdentifier(tenantID)

// Query builder (Squirrel)
query := sq.Select("*").From("resources").Where(sq.Eq{"tenant_id": tenantID})
```

**Reference Implementation (BAD):**
```go
// BAD: String concatenation
query := "SELECT * FROM users WHERE name = '" + name + "'"

// BAD: fmt.Sprintf for values
query := fmt.Sprintf("SELECT * FROM users WHERE id = '%s'", id)

// BAD: Unescaped identifier
query := "SET search_path TO " + tenantID  // SQL injection via tenant
```

**Check For:**
1. All queries use parameterized statements ($1, $2, ...)
2. No string concatenation in SQL queries
3. Dynamic identifiers properly escaped (QuoteIdentifier)
4. Query builders used for complex WHERE clauses
5. No raw SQL with user input

**Severity Ratings:**
- CRITICAL: String concatenation with user input
- CRITICAL: fmt.Sprintf with user values
- HIGH: Unescaped dynamic identifiers
- MEDIUM: Raw SQL where builder would be safer
- LOW: Inconsistent query patterns

**Output Format:**
```
## SQL Safety Audit Findings

### Summary
- Parameterized queries: X/Y
- String concatenation risks: Z
- Identifier escaping: Yes/No

### Critical Issues
[file:line] - Description

### Recommendations
1. ...
```
```

### Agent 9: Runtime Safety Auditor

```prompt
Audit pkg/runtime usage and panic handling for production readiness.

**Search Patterns:**
- Files: `**/runtime/*.go`, `**/recover*.go`, `**/*.go`
- Keywords: `RecoverAndLog`, `RecoverWithPolicy`, `InitPanicMetrics`, `SetProductionMode`
- Also search: `panic(`, `recover()` (manual usage)

**Reference Implementation (GOOD):**
```go
// Bootstrap initialization
runtime.InitPanicMetrics(telemetry.MetricsFactory)
if cfg.EnvName == "production" {
    runtime.SetProductionMode(true)
}

// In HTTP handlers
defer runtime.RecoverAndLogWithContext(ctx, logger, "module", "handler_name")

// In worker goroutines
defer runtime.RecoverWithPolicyAndContext(ctx, logger, "module", "worker", runtime.CrashProcess)

// In background jobs (should retry, not crash)
defer runtime.RecoverWithPolicyAndContext(ctx, logger, "module", "job", runtime.LogAndContinue)
```

**Check For:**
1. pkg/runtime initialized at startup
2. Production mode set based on environment
3. All goroutines have panic recovery
4. Appropriate recovery policies per context
5. Panic metrics enabled for alerting
6. No raw recover() without pkg/runtime

**Severity Ratings:**
- CRITICAL: Goroutines without panic recovery
- HIGH: Missing production mode setting
- HIGH: Raw recover() without proper handling
- MEDIUM: Inconsistent recovery policies
- LOW: Missing panic metrics

**Output Format:**
```
## Runtime Safety Audit Findings

### Summary
- Runtime initialized: Yes/No
- Handlers with recovery: X/Y
- Goroutines with recovery: X/Y

### Critical Issues
[file:line] - Description

### Recommendations
1. ...
```
```

### Agent 10: Input Validation Auditor

```prompt
Audit input validation patterns for production readiness.

**Search Patterns:**
- Files: `**/dto.go`, `**/handlers.go`, `**/value_objects/*.go`
- Keywords: `validate:`, `BodyParser`, `IsValid()`, `Parse`, `required`

**Reference Implementation (GOOD):**
```go
// DTO with validation tags
type CreateRequest struct {
    Name   string `json:"name" validate:"required,min=1,max=255"`
    Type   string `json:"type" validate:"required,oneof=TYPE_A TYPE_B"`
    Amount int    `json:"amount" validate:"gte=0,lte=1000000"`
}

// Handler with body parsing error handling
func (h *Handler) Create(c *fiber.Ctx) error {
    var payload CreateRequest
    if err := c.BodyParser(&payload); err != nil {
        return badRequest(c, span, logger, "invalid request body", err)
    }
    // Validate struct
    if err := h.validator.Struct(payload); err != nil {
        return badRequest(c, span, logger, "validation failed", err)
    }
    ...
}

// Value object with domain validation
func (vo ValueObject) IsValid() bool {
    if vo.value == "" || len(vo.value) > maxLength {
        return false
    }
    return validPattern.MatchString(vo.value)
}
```

**Reference Implementation (BAD):**
```go
// BAD: No validation tags
type Request struct {
    Name string `json:"name"`  // No validation!
}

// BAD: Ignoring body parse error
payload := Request{}
c.BodyParser(&payload)  // Error ignored!

// BAD: No bounds checking
amount := c.QueryInt("amount")  // Could be negative or huge
```

**Check For:**
1. All DTOs have validate: tags on required fields
2. BodyParser errors are handled (not ignored)
3. Query/path params validated before use
4. Numeric bounds enforced (min/max)
5. String length limits enforced
6. Enum values constrained (oneof=)
7. Value objects have IsValid() methods
8. File upload size/type validation

**Severity Ratings:**
- CRITICAL: BodyParser errors ignored
- HIGH: No validation on user input DTOs
- HIGH: Unbounded numeric inputs
- MEDIUM: Missing string length limits
- LOW: Value objects without IsValid()

**Output Format:**
```
## Input Validation Audit Findings

### Summary
- DTOs with validation tags: X/Y
- BodyParser error handling: X/Y
- Value objects with IsValid: X/Y

### Critical Issues
[file:line] - Description

### Recommendations
1. ...
```
```

### Agent 11: Rate Limiting Auditor

```prompt
Audit rate limiting implementation for production readiness.

**Search Patterns:**
- Files: `**/fiber_server.go`, `**/middleware*.go`, `**/limiter*.go`
- Keywords: `limiter.New`, `RateLimiter`, `Max:`, `Expiration:`, `KeyGenerator`

**Reference Implementation (GOOD):**
```go
// Multi-level rate limiting
app.Use(limiter.New(limiter.Config{
    Max:        100,                    // 100 requests
    Expiration: 60 * time.Second,       // per 60 seconds
    KeyGenerator: func(c *fiber.Ctx) string {
        ctx := c.UserContext()
        // Priority: user > tenant > IP
        if userID := getUserID(ctx); userID != "" {
            return "user:" + userID
        }
        if tenantID := getTenantID(ctx); tenantID != "" {
            return "tenant:" + tenantID + ":" + c.IP()
        }
        return "ip:" + c.IP()
    },
    LimitReached: func(c *fiber.Ctx) error {
        return c.Status(429).JSON(fiber.Map{
            "error": "rate limit exceeded",
            "retry_after": 60,
        })
    },
}))

// Endpoint-specific limits for expensive operations
exportGroup.Use(limiter.New(limiter.Config{
    Max:        10,  // Tighter limit for exports
    Expiration: 60 * time.Second,
}))
```

**Check For:**
1. Global rate limiter configured
2. Limits are reasonable (not too high/low)
3. Key generator uses authenticated identity when available
4. 429 response includes retry-after info
5. Expensive endpoints (exports, reports) have tighter limits
6. Rate limit bypass for health checks

**Severity Ratings:**
- CRITICAL: No rate limiting at all
- HIGH: Rate limiting uses only IP (easily bypassed)
- HIGH: Export/report endpoints unprotected
- MEDIUM: Missing retry-after in 429 response
- LOW: Limits not tuned for production load

**Output Format:**
```
## Rate Limiting Audit Findings

### Summary
- Global rate limiter: Yes/No
- Key generator: IP-only / User-aware / Tenant-aware
- Endpoint-specific limits: X endpoints

### Critical Issues
[file:line] - Description

### Recommendations
1. ...
```
```

### Agent 12: Health Checks Auditor

```prompt
Audit health check endpoints for production readiness.

**Search Patterns:**
- Files: `**/fiber_server.go`, `**/health*.go`, `**/routes.go`
- Keywords: `/health`, `/ready`, `/live`, `healthHandler`, `readinessHandler`

**Reference Implementation (GOOD):**
```go
// Liveness probe - always returns healthy if process is running
func healthHandler(c *fiber.Ctx) error {
    return c.SendString("healthy")
}

// Readiness probe - checks all dependencies
func readinessHandler(deps *HealthDependencies) fiber.Handler {
    return func(c *fiber.Ctx) error {
        checks := fiber.Map{}
        status := fiber.StatusOK

        // Required dependency - fails readiness if down
        if err := deps.DB.Ping(c.Context()); err != nil {
            checks["database"] = "unhealthy"
            status = fiber.StatusServiceUnavailable
        } else {
            checks["database"] = "healthy"
        }

        // Optional dependency - reports degraded but doesn't fail
        if deps.Redis != nil {
            if err := deps.Redis.Ping(c.Context()).Err(); err != nil {
                checks["redis"] = "degraded"
            } else {
                checks["redis"] = "healthy"
            }
        }

        return c.Status(status).JSON(fiber.Map{
            "status": statusString(status),
            "checks": checks,
        })
    }
}

// Register without auth middleware
app.Get("/health", healthHandler)
app.Get("/ready", readinessHandler(deps))
```

**Check For:**
1. /health endpoint exists (liveness)
2. /ready endpoint exists (readiness)
3. Health endpoints bypass auth middleware
4. Database connectivity checked in readiness
5. Message queue connectivity checked
6. Optional deps don't fail readiness (just report degraded)
7. Response includes individual check status
8. Appropriate HTTP status codes (200 vs 503)

**Severity Ratings:**
- CRITICAL: No health endpoints at all
- HIGH: No readiness probe (only liveness)
- HIGH: Health endpoints require auth
- MEDIUM: Missing dependency checks in readiness
- LOW: No degraded status for optional deps

**Output Format:**
```
## Health Checks Audit Findings

### Summary
- Liveness endpoint: Yes/No (/path)
- Readiness endpoint: Yes/No (/path)
- Dependencies checked: [list]

### Critical Issues
[file:line] - Description

### Recommendations
1. ...
```
```

### Agent 13: Configuration Management Auditor

```prompt
Audit configuration management for production readiness.

**Search Patterns:**
- Files: `**/config.go`, `**/bootstrap/*.go`, `**/.env*`
- Keywords: `env:`, `envDefault:`, `Validate()`, `LoadConfig`, `production`

**Reference Implementation (GOOD):**
```go
// Config with validation
type Config struct {
    EnvName    string `env:"ENV_NAME" envDefault:"development"`
    DBPassword string `env:"POSTGRES_PASSWORD"`
    AuthEnabled bool  `env:"AUTH_ENABLED" envDefault:"false"`
}

// Production validation
func (c *Config) Validate() error {
    if c.EnvName == "production" {
        // Require auth in production
        if !c.AuthEnabled {
            return errors.New("AUTH_ENABLED must be true in production")
        }
        // Require DB password in production
        if c.DBPassword == "" {
            return errors.New("POSTGRES_PASSWORD required in production")
        }
        // No wildcard CORS in production
        if c.CORSOrigins == "*" {
            return errors.New("CORS_ALLOWED_ORIGINS cannot be * in production")
        }
        // Require TLS for databases
        if c.PostgresSSLMode == "disable" {
            return errors.New("POSTGRES_SSLMODE cannot be disable in production")
        }
    }
    return nil
}

// Load with validation
func LoadConfig() (*Config, error) {
    cfg := &Config{}
    if err := envconfig.Process("", cfg); err != nil {
        return nil, fmt.Errorf("load env: %w", err)
    }
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("validate: %w", err)
    }
    return cfg, nil
}
```

**Check For:**
1. All config loaded from env vars (not hardcoded)
2. Sensible defaults for non-production
3. Production-specific validation exists
4. Auth required in production
5. TLS/SSL required in production
6. No wildcard CORS in production
7. Default credentials rejected in production
8. Secrets not logged during startup
9. Config validation fails fast (at startup)

**Severity Ratings:**
- CRITICAL: Hardcoded secrets in code
- CRITICAL: No production validation
- HIGH: Auth can be disabled in production
- HIGH: TLS not enforced in production
- MEDIUM: Missing sensible defaults
- LOW: Config not validated at startup

**Output Format:**
```
## Configuration Management Audit Findings

### Summary
- Env vars used: X fields
- Production validation: Yes/No
- Constraints enforced: [list]

### Critical Issues
[file:line] - Description

### Recommendations
1. ...
```
```

### Agent 14: Connection Management Auditor

```prompt
Audit database and cache connection management for production readiness.

**Search Patterns:**
- Files: `**/config.go`, `**/database*.go`, `**/redis*.go`, `**/postgres*.go`
- Keywords: `MaxOpenConns`, `MaxIdleConns`, `PoolSize`, `Timeout`, `SetConnMaxLifetime`

**Reference Implementation (GOOD):**
```go
// Database pool configuration
type DBConfig struct {
    MaxOpenConnections int `env:"POSTGRES_MAX_OPEN_CONNS" envDefault:"25"`
    MaxIdleConnections int `env:"POSTGRES_MAX_IDLE_CONNS" envDefault:"5"`
    ConnMaxLifetime    int `env:"POSTGRES_CONN_MAX_LIFETIME_MINS" envDefault:"30"`
}

// Apply pool settings
func ConfigurePool(db *sql.DB, cfg *DBConfig) {
    db.SetMaxOpenConns(cfg.MaxOpenConnections)
    db.SetMaxIdleConns(cfg.MaxIdleConnections)
    db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)
}

// Redis pool configuration
type RedisConfig struct {
    PoolSize       int `env:"REDIS_POOL_SIZE" envDefault:"10"`
    MinIdleConns   int `env:"REDIS_MIN_IDLE_CONNS" envDefault:"2"`
    ReadTimeoutMs  int `env:"REDIS_READ_TIMEOUT_MS" envDefault:"3000"`
    WriteTimeoutMs int `env:"REDIS_WRITE_TIMEOUT_MS" envDefault:"3000"`
    DialTimeoutMs  int `env:"REDIS_DIAL_TIMEOUT_MS" envDefault:"5000"`
}

// Primary + Replica support
type DatabaseConnections struct {
    Primary *sql.DB
    Replica *sql.DB  // Falls back to primary if not configured
}
```

**Check For:**
1. DB connection pool limits configured
2. Redis pool settings configured
3. Connection timeouts set (not infinite)
4. Connection max lifetime set (prevents stale connections)
5. Idle connection limits reasonable
6. Read replica support (for scaling reads)
7. Connection health checks (ping on checkout)
8. Graceful connection shutdown

**Severity Ratings:**
- CRITICAL: No connection pool limits (unbounded connections)
- HIGH: No connection timeouts (hang forever)
- HIGH: No max lifetime (stale connections)
- MEDIUM: Missing read replica support
- LOW: Pool sizes not tuned

**Output Format:**
```
## Connection Management Audit Findings

### Summary
- DB pool configured: Yes/No (max: X, idle: Y)
- Redis pool configured: Yes/No (size: X)
- Timeouts configured: Yes/No
- Replica support: Yes/No

### Critical Issues
[file:line] - Description

### Recommendations
1. ...
```
```

### Agent 15: Logging & PII Safety Auditor

```prompt
Audit logging practices and PII protection for production readiness.

**Search Patterns:**
- Files: `**/*.go`
- Keywords: `logger.`, `log.`, `Errorf`, `Infof`, `WithFields`, `password`, `token`, `secret`
- Also search: `fmt.Print`, `fmt.Println` (should not be used for logging)

**Reference Implementation (GOOD):**
```go
// Structured logging with context
logger, tracer, requestID, _ := libCommons.NewTrackingFromContext(ctx)
logger.WithFields(
    "request_id", requestID,
    "user_id", userID,
    "action", "create_resource",
).Info("resource created")

// Production-safe error logging
if isProduction {
    // Don't include error details that might leak PII
    logger.Errorf("operation failed: status=%d path=%s", code, path)
} else {
    // Development can have full details
    logger.Errorf("operation failed: error=%v", err)
}

// Config DSN without password
func (c *Config) DSN() string {
    // Returns connection string without logging password
    return fmt.Sprintf("host=%s port=%d user=%s dbname=%s",
        c.Host, c.Port, c.User, c.DBName)
}
```

**Reference Implementation (BAD):**
```go
// BAD: fmt.Println for logging
fmt.Println("User logged in:", userEmail)

// BAD: Logging sensitive data
logger.Infof("Login attempt: email=%s password=%s", email, password)

// BAD: Logging full request body (might contain PII)
logger.Debugf("Request body: %+v", requestBody)

// BAD: Not using structured logging
log.Printf("Error: %v", err)
```

**Check For:**
1. Structured logging used (not fmt.Print)
2. Logger obtained from context (request tracking)
3. No passwords/tokens logged
4. Production mode sanitizes error details
5. Request/response bodies not logged raw
6. Log levels appropriate (not everything at INFO)
7. Request IDs included for tracing
8. No PII in log messages (emails, names, etc.)

**Severity Ratings:**
- CRITICAL: Passwords/tokens logged
- CRITICAL: PII logged in production
- HIGH: fmt.Print used instead of logger
- HIGH: Full error details in production
- MEDIUM: Missing request ID in logs
- LOW: Inappropriate log levels

**Output Format:**
```
## Logging & PII Safety Audit Findings

### Summary
- Structured logging: Yes/No
- PII protection: Yes/No
- Production mode: Yes/No

### Critical Issues
[file:line] - Description

### Recommendations
1. ...
```
```

### Agent 16: Idempotency Auditor

```prompt
Audit idempotency implementation for production readiness.

**Search Patterns:**
- Files: `**/idempotency*.go`, `**/value_objects/*.go`, `**/redis/*.go`
- Keywords: `IdempotencyKey`, `TryAcquire`, `MarkComplete`, `SetNX`, `idempotent`

**Reference Implementation (GOOD):**
```go
// Idempotency key value object
type IdempotencyKey string

const (
    idempotencyKeyMaxLength = 128
    idempotencyKeyPattern   = `^[A-Za-z0-9:_-]+$`
)

func (key IdempotencyKey) IsValid() bool {
    s := string(key)
    if s == "" || len(s) > idempotencyKeyMaxLength {
        return false
    }
    return regexp.MustCompile(idempotencyKeyPattern).MatchString(s)
}

// Redis-backed idempotency
type IdempotencyRepository struct {
    client *redis.Client
    ttl    time.Duration  // e.g., 7 days
}

func (r *IdempotencyRepository) TryAcquire(ctx context.Context, key IdempotencyKey) (bool, error) {
    // SetNX is atomic - only first caller wins
    result, err := r.client.SetNX(ctx, r.keyName(key), "acquired", r.ttl).Result()
    return result, err
}

func (r *IdempotencyRepository) MarkComplete(ctx context.Context, key IdempotencyKey) error {
    return r.client.Set(ctx, r.keyName(key), "complete", r.ttl).Err()
}

// Usage in handler
func (h *Handler) ProcessCallback(c *fiber.Ctx) error {
    key := extractIdempotencyKey(c)

    acquired, err := h.idempotency.TryAcquire(ctx, key)
    if err != nil {
        return internalError(c, "idempotency check failed", err)
    }
    if !acquired {
        return c.Status(200).JSON(fiber.Map{"status": "already_processed"})
    }

    // Process...

    h.idempotency.MarkComplete(ctx, key)
    return c.JSON(result)
}
```

**Check For:**
1. Idempotency keys for financial/critical operations
2. Atomic acquire mechanism (SetNX or similar)
3. TTL to prevent unbounded storage
4. Key validation (format, length)
5. Proper state transitions (acquired → complete/failed)
6. Retry-safe (failed operations can be retried)
7. Idempotency for webhook callbacks
8. Idempotency for payment operations

**Severity Ratings:**
- CRITICAL: No idempotency for financial operations
- HIGH: Non-atomic acquire (race conditions)
- HIGH: No TTL (memory leak)
- MEDIUM: Missing key validation
- LOW: No failed state handling

**Output Format:**
```
## Idempotency Audit Findings

### Summary
- Idempotency implemented: Yes/No
- Operations covered: [list]
- Storage backend: Redis/DB/Memory
- TTL configured: X days

### Critical Issues
[file:line] - Description

### Recommendations
1. ...
```
```

### Agent 17: API Documentation Auditor

```prompt
Audit API documentation (Swagger/OpenAPI) for production readiness.

**Search Patterns:**
- Files: `**/main.go`, `**/handlers.go`, `**/dto.go`, `**/swagger/*`
- Keywords: `@Summary`, `@Router`, `@Param`, `@Success`, `@Failure`, `@Security`

**Reference Implementation (GOOD):**
```go
// Main entry with API metadata
// @title           My API
// @version         v1.0.0
// @description     API description
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// Handler with full documentation
// @Summary      Create a resource
// @Description  Creates a new resource with the given parameters
// @Tags         resources
// @Accept       json
// @Produce      json
// @Param        request body CreateRequest true "Resource to create"
// @Success      201 {object} ResourceResponse
// @Failure      400 {object} ErrorResponse "Invalid input"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Forbidden"
// @Failure      500 {object} ErrorResponse "Internal error"
// @Security     BearerAuth
// @Router       /v1/resources [post]
func (h *Handler) Create(c *fiber.Ctx) error { ... }

// DTO with documentation
type CreateRequest struct {
    Name   string `json:"name" example:"my-resource" validate:"required"`
    Type   string `json:"type" example:"TYPE_A" enums:"TYPE_A,TYPE_B"`
    Amount int    `json:"amount" example:"100" minimum:"0" maximum:"1000000"`
}
```

**Check For:**
1. API title, version, description in main.go
2. Security definitions (Bearer token)
3. All endpoints have @Router annotation
4. Request/response types documented
5. All error codes documented (@Failure)
6. Examples in DTOs (example: tag)
7. Enums documented (enums: tag)
8. Parameter constraints documented (minimum, maximum)
9. Tags organize endpoints logically
10. Swagger UI accessible

**Severity Ratings:**
- HIGH: No Swagger annotations at all
- HIGH: Missing security definitions
- MEDIUM: Endpoints without documentation
- MEDIUM: Error responses not documented
- LOW: Missing examples in DTOs
- LOW: Inconsistent tag usage

**Output Format:**
```
## API Documentation Audit Findings

### Summary
- Swagger annotations: Yes/No
- Documented endpoints: X/Y
- Security definitions: Yes/No
- Error responses documented: X/Y

### Critical Issues
[file:line] - Description

### Recommendations
1. ...
```
```

### Agent 18: Technical Debt Auditor

```prompt
Audit technical debt indicators for production readiness.

**Search Patterns (with context):**
- `TODO` - Planned work
- `FIXME` - Known bugs
- `HACK` - Workarounds
- `XXX` - Danger zones
- `deprecated` (case-insensitive)
- `"in a real implementation"` or `"real implementation"`
- `"temporary"` or `"temp fix"`
- `"workaround"`
- `panic("not implemented")`

**Risk Assessment Criteria:**

**Implement Now (High Risk):**
- Security-related TODOs (auth, validation, encryption)
- Error handling TODOs in critical paths
- Data integrity issues
- "FIXME" in production code paths

**Monitor (Medium Risk):**
- Performance optimization TODOs
- Missing rate limiting
- Incomplete logging
- "deprecated" usage without migration plan

**Acceptable Debt (Low Risk):**
- Future feature ideas
- Code style improvements
- Test coverage expansion
- Documentation improvements

**Output Format:**
```
## Technical Debt Audit Findings

### Summary
- Total TODOs: X
- Total FIXMEs: Y
- Deprecated usage: Z
- "Real implementation" markers: N

### HIGH RISK - Implement Now
| File:Line | Type | Description | Risk |
|-----------|------|-------------|------|
| path:123 | TODO | Auth bypass for testing | Security |

### MEDIUM RISK - Monitor
| File:Line | Type | Description | Risk |
|-----------|------|-------------|------|

### LOW RISK - Acceptable Debt
| File:Line | Type | Description | Risk |
|-----------|------|-------------|------|

### Recommendations
1. ...
```
```

### Agent 19: Testing Coverage Auditor

```prompt
Audit test coverage and testing patterns for production readiness.

**Search Patterns:**
- Files: `**/*_test.go`, `**/mocks/**/*.go`, `tests/**/*.go`
- Keywords: `func Test`, `t.Run`, `mock.Mock`, `assert.`, `require.`

**Reference Implementation (GOOD):**
```go
// Co-located test file
// file: handler_test.go (next to handler.go)

func TestHandler_Create(t *testing.T) {
    // Arrange
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mocks.NewMockRepository(ctrl)
    mockRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)

    handler := NewHandler(mockRepo)

    // Act
    result, err := handler.Create(ctx, input)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, expected, result)
}

// Table-driven tests for multiple cases
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid input", "test", false},
        {"empty input", "", true},
        {"too long", strings.Repeat("a", 300), true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := Validate(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

// Integration test with testcontainers
func TestIntegration_CreateResource(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    // Setup container...
}
```

**Check For:**
1. Test files co-located with source (*_test.go)
2. Mocks generated via mockgen (not hand-written)
3. Table-driven tests for multiple cases
4. Integration tests in separate directory or with build tags
5. Test helpers/fixtures organized
6. Assertions use testify (assert/require)
7. Parallel tests where appropriate (t.Parallel())
8. Test cleanup with t.Cleanup() or defer

**Severity Ratings:**
- HIGH: Critical paths without tests
- HIGH: Hand-written mocks (should use mockgen)
- MEDIUM: Missing table-driven tests for validators
- MEDIUM: No integration tests
- LOW: Tests not running in parallel
- LOW: Missing edge case coverage

**Output Format:**
```
## Testing Coverage Audit Findings

### Summary
- Test files found: X
- Modules with tests: X/Y
- Mock generation: mockgen / hand-written
- Integration tests: Yes/No

### Critical Issues
[file:line] - Description

### Recommendations
1. ...
```
```

### Agent 20: Dependency Auditor

```prompt
Audit dependency management for production readiness.

**Search Patterns:**
- Files: `go.mod`, `go.sum`, `**/vendor/**`
- Commands: Run `go list -m -u all` mentally based on go.mod

**Reference Implementation (GOOD):**
```go
// go.mod with pinned versions
module github.com/company/project

go 1.24

require (
    github.com/gofiber/fiber/v2 v2.52.10  // Pinned, not "latest"
    github.com/lib/pq v1.10.9
    go.opentelemetry.io/otel v1.39.0
)

// Indirect deps managed automatically
require (
    github.com/valyala/fasthttp v1.52.0 // indirect
)
```

**Reference Implementation (BAD):**
```go
// BAD: Using replace for production
replace github.com/some/lib => ../local-lib

// BAD: Unpinned versions
require github.com/some/lib latest

// BAD: Very old versions with known CVEs
require github.com/dgrijalva/jwt-go v3.2.0  // Has CVE, use golang-jwt
```

**Check For:**
1. All dependencies pinned (no "latest")
2. No local replace directives in production
3. Known vulnerable packages identified
4. Unused dependencies (not imported anywhere)
5. Major version mismatches
6. Deprecated packages (e.g., dgrijalva/jwt-go → golang-jwt)
7. go.sum exists and is committed

**Known Vulnerable Packages to Flag:**
- github.com/dgrijalva/jwt-go (use golang-jwt/jwt)
- github.com/pkg/sftp < v1.13.5
- golang.org/x/crypto < recent
- golang.org/x/net < recent

**Severity Ratings:**
- CRITICAL: Known CVE in dependency
- HIGH: Local replace directive
- HIGH: Deprecated package with security issues
- MEDIUM: Significantly outdated dependencies
- LOW: Minor version behind

**Output Format:**
```
## Dependency Audit Findings

### Summary
- Total dependencies: X
- Direct dependencies: Y
- Potentially outdated: Z
- Known vulnerabilities: N

### Critical Issues
[package] - Description

### Recommendations
1. ...
```
```

### Agent 21: Performance Auditor

```prompt
Audit performance patterns for production readiness.

**Search Patterns:**
- Files: `**/*.go`
- Keywords: `for.*range`, `append(`, `make(`, `sync.Pool`, `SELECT *`, `N+1`

**Reference Implementation (GOOD):**
```go
// Pre-allocate slices when size is known
items := make([]Item, 0, len(input))  // Capacity hint

// Use sync.Pool for frequently allocated objects
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

// Batch database operations
func (r *Repo) CreateBatch(ctx context.Context, items []Item) error {
    return r.db.WithContext(ctx).CreateInBatches(items, 100).Error
}

// Select only needed columns
func (r *Repo) List(ctx context.Context) ([]Item, error) {
    return r.db.WithContext(ctx).
        Select("id", "name", "status").  // Not SELECT *
        Find(&items).Error
}

// Avoid N+1 with preloading
func (r *Repo) GetWithRelations(ctx context.Context, id uuid.UUID) (*Item, error) {
    return r.db.WithContext(ctx).
        Preload("Children").
        First(&item, id).Error
}
```

**Reference Implementation (BAD):**
```go
// BAD: SELECT * fetches unnecessary data
db.Find(&items)

// BAD: N+1 query pattern
for _, item := range items {
    db.Where("parent_id = ?", item.ID).Find(&children)  // Query per item!
}

// BAD: Growing slice without capacity
var items []Item
for _, input := range inputs {
    items = append(items, transform(input))  // Reallocates repeatedly
}

// BAD: Large allocations in hot path without pooling
func handleRequest() {
    buf := make([]byte, 1<<20)  // 1MB allocation per request
}
```

**Check For:**
1. SELECT * avoided (explicit column selection)
2. N+1 queries prevented (use Preload/joins)
3. Slice pre-allocation when size known
4. sync.Pool for frequent allocations
5. Batch operations for bulk inserts/updates
6. Indexes exist for filtered/sorted columns
7. Connection pooling configured
8. Context timeouts on DB operations

**Severity Ratings:**
- HIGH: N+1 query pattern in production code
- HIGH: SELECT * on large tables
- MEDIUM: Missing slice pre-allocation
- MEDIUM: No batch operations for bulk data
- LOW: Missing sync.Pool optimization
- LOW: Minor inefficiencies

**Output Format:**
```
## Performance Audit Findings

### Summary
- N+1 patterns found: X
- SELECT * usage: Y
- Missing pre-allocations: Z

### Critical Issues
[file:line] - Description

### Recommendations
1. ...
```
```

### Agent 22: Concurrency Auditor

```prompt
Audit concurrency patterns for production readiness.

**Search Patterns:**
- Files: `**/*.go`
- Keywords: `go func`, `sync.Mutex`, `sync.RWMutex`, `chan`, `select {`, `sync.WaitGroup`

**Reference Implementation (GOOD):**
```go
// Mutex protecting shared state
type Cache struct {
    mu    sync.RWMutex
    items map[string]Item
}

func (c *Cache) Get(key string) (Item, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    item, ok := c.items[key]
    return item, ok
}

func (c *Cache) Set(key string, item Item) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.items[key] = item
}

// WaitGroup for goroutine coordination
func processAll(items []Item) error {
    var wg sync.WaitGroup
    errCh := make(chan error, len(items))

    for _, item := range items {
        wg.Add(1)
        go func(i Item) {
            defer wg.Done()
            if err := process(i); err != nil {
                errCh <- err
            }
        }(item)  // Pass item to avoid closure capture
    }

    wg.Wait()
    close(errCh)

    // Collect errors
    for err := range errCh {
        return err
    }
    return nil
}

// Context for cancellation
func worker(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        case item := <-workCh:
            process(item)
        }
    }
}
```

**Reference Implementation (BAD):**
```go
// BAD: Race condition - map access without lock
var cache = make(map[string]Item)
func Get(key string) Item { return cache[key] }  // Concurrent read/write!

// BAD: Goroutine leak - no way to stop
go func() {
    for {
        process()  // Runs forever, no context check
    }
}()

// BAD: Closure captures loop variable
for _, item := range items {
    go func() {
        process(item)  // All goroutines see last item!
    }()
}

// BAD: Unbounded goroutine spawning
for _, item := range millionItems {
    go process(item)  // 1M goroutines!
}
```

**Check For:**
1. Maps protected by mutex when shared
2. Loop variables not captured in closures
3. Goroutines have cancellation (context)
4. WaitGroup used for coordination
5. Bounded concurrency (worker pools)
6. Channels closed by sender
7. Select with default for non-blocking
8. No goroutine leaks (all paths exit)

**Severity Ratings:**
- CRITICAL: Race condition on shared map
- CRITICAL: Goroutine leak (no exit path)
- HIGH: Loop variable capture bug
- HIGH: Unbounded goroutine spawning
- MEDIUM: Missing context cancellation
- LOW: Inefficient locking patterns

**Output Format:**
```
## Concurrency Audit Findings

### Summary
- Goroutine spawns: X locations
- Mutex usage: Y locations
- Potential race conditions: Z

### Critical Issues
[file:line] - Description

### Recommendations
1. ...
```
```

### Agent 23: Migration Safety Auditor

```prompt
Audit database migration safety for production readiness.

**Search Patterns:**
- Files: `migrations/*.sql`, `migrations/*.go`
- Keywords: `DROP`, `ALTER`, `RENAME`, `NOT NULL`, `CREATE INDEX`

**Reference Implementation (GOOD):**
```sql
-- 000001_create_users.up.sql
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email ON users(email);

-- 000001_create_users.down.sql
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;

-- Adding nullable column (safe)
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(50);

-- Adding NOT NULL with default (safe)
ALTER TABLE users ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'active';
```

**Reference Implementation (BAD):**
```sql
-- BAD: Adding NOT NULL without default (locks table, fails if data exists)
ALTER TABLE users ADD COLUMN role VARCHAR(50) NOT NULL;

-- BAD: Non-concurrent index (locks table)
CREATE INDEX idx_users_email ON users(email);

-- BAD: Destructive without IF EXISTS
DROP TABLE users;
DROP COLUMN email;

-- BAD: Renaming column (breaks application)
ALTER TABLE users RENAME COLUMN email TO user_email;
```

**Check For:**
1. All migrations have up AND down files
2. CREATE INDEX uses CONCURRENTLY
3. New NOT NULL columns have DEFAULT
4. DROP/ALTER use IF EXISTS
5. No column renames (add new, migrate data, drop old)
6. No destructive operations in up migrations
7. Migrations are additive (safe rollback)
8. Sequential numbering (no gaps)

**Severity Ratings:**
- CRITICAL: NOT NULL without default
- CRITICAL: Missing down migration
- HIGH: Non-concurrent index creation
- HIGH: Column rename (breaking change)
- MEDIUM: DROP without IF EXISTS
- LOW: Migration naming inconsistency

**Output Format:**
```
## Migration Safety Audit Findings

### Summary
- Total migrations: X
- Up migrations: Y
- Down migrations: Z
- Potentially unsafe: N

### Critical Issues
[file:line] - Description

### Recommendations
1. ...
```
```

### Agent 24: Container Security Auditor

```prompt
Audit container security and Dockerfile best practices for production readiness.

**Search Patterns:**
- Files: `Dockerfile*`, `docker-compose*.yml`, `Makefile`
- Keywords: `FROM`, `USER`, `COPY`, `ADD`, `HEALTHCHECK`

**Reference Implementation (GOOD):**
```dockerfile
# Multi-stage build
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /main cmd/app/main.go

# Distroless or minimal runtime image
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=builder /main .
# Non-root user
USER nonroot:nonroot
# Healthcheck defined
HEALTHCHECK --interval=30s --timeout=3s CMD ["/main", "-health"]
ENTRYPOINT ["/main"]
```

**Check For:**
1. Multi-stage builds (builder vs runtime)
2. Minimal/Distroless runtime images
3. Pinned base image versions (not `latest`)
4. Non-root user execution (`USER nonroot` or numeric ID)
5. `COPY` used instead of `ADD` (unless extracting tar)
6. .dockerignore file exists and excludes secrets/git
7. Sensitive args not passed as build-args (secrets)

**Severity Ratings:**
- CRITICAL: Running as root in production image
- HIGH: Secrets in Dockerfile/history
- MEDIUM: Using `latest` tag
- LOW: Missing HEALTHCHECK in Dockerfile

**Output Format:**
```
## Container Security Audit Findings

### Summary
- Multi-stage build: Yes/No
- Non-root user: Yes/No
- Base image pinned: Yes/No

### Critical Issues
[file:line] - Description

### Recommendations
1. ...
```
```

### Agent 25: HTTP Hardening Auditor

```prompt
Audit HTTP security headers and hardening configuration for production readiness.

**Search Patterns:**
- Files: `**/fiber_server.go`, `**/middleware*.go`, `**/cors.go`
- Keywords: `Helmet`, `CORS`, `CSRF`, `Secure`, `HttpOnly`, `SameSite`

**Reference Implementation (GOOD):**
```go
// Security headers
app.Use(helmet.New(helmet.Config{
    XSSProtection:             "1; mode=block",
    ContentTypeNosniff:        "nosniff",
    XFrameOptions:             "DENY",
    HSTSMaxAge:                31536000,
    HSTSExcludeSubdomains:     false,
    HSTSPreloadEnabled:        true,
    ContentSecurityPolicy:     "default-src 'self'",
}))

// CORS configuration
app.Use(cors.New(cors.Config{
    AllowOrigins:     "https://app.example.com", // Specific origin
    AllowMethods:     "GET,POST,PUT,DELETE",
    AllowCredentials: true,
}))
```

**Check For:**
1. HSTS enabled (Strict-Transport-Security)
2. CSP configured (Content-Security-Policy)
3. X-Frame-Options set to DENY or SAMEORIGIN
4. CORS restricts origins (no `*` in production)
5. Secure cookies (Secure, HttpOnly, SameSite=Strict/Lax)
6. Server banner suppressed (Server: value removed)

**Severity Ratings:**
- CRITICAL: CORS `*` allowed with credentials
- HIGH: Missing HSTS
- MEDIUM: Missing CSP or overly permissive
- LOW: Server banner exposed

**Output Format:**
```
## HTTP Hardening Audit Findings

### Summary
- HSTS enabled: Yes/No
- CORS strictness: Strict/Permissive/Open
- CSP configured: Yes/No

### Critical Issues
[file:line] - Description

### Recommendations
1. ...
```
```

### Agent 26: CI/CD & Licensing Auditor

```prompt
Audit CI/CD pipelines and license compliance for production readiness.

**Search Patterns:**
- Files: `.github/workflows/*.yml`, `.gitlab-ci.yml`, `LICENSE*`
- Keywords: `test`, `lint`, `build`, `docker`, `sign`

**Reference Implementation (GOOD):**
```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - run: go test -race -v ./...
      - run: golangci-lint run
  
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: securego/gosec@master
        with:
          args: ./...
```

**Check For:**
1. CI pipeline exists (GitHub Actions/GitLab CI)
2. Tests and Linting run on PRs
3. Security scanning (gosec, trivy) integrated
4. Artifact signing (cosign/sigstore)
5. **LICENSE File Check**:
   - IF open source: `LICENSE` file MUST exist.
   - IF proprietary: Ensure private repo settings/headers.
   - **ACTION**: Explicitly verify if the project is Open Source or Private.
6. Dependency license check (verify no GPL pollution if proprietary)

**Severity Ratings:**
- CRITICAL: No CI pipeline
- CRITICAL: Tests not running on PR
- HIGH: Open source project missing LICENSE file
- MEDIUM: Missing security scanning
- LOW: Artifacts not signed

**Output Format:**
```
## CI/CD & Licensing Audit Findings

### Summary
- CI Pipeline: Active/Missing
- Security Scans: Yes/No
- License: [Type/Missing] (Context: Open Source/Proprietary)

### Critical Issues
[file:line] - Description

### Recommendations
1. ...
```
```

### Agent 27: Async Reliability Auditor

```prompt
Audit asynchronous processing reliability for production readiness.

**Search Patterns:**
- Files: `**/worker/*.go`, `**/queue/*.go`, `**/kafka/*.go`, `**/rabbitmq/*.go`
- Keywords: `Ack`, `Nack`, `Retry`, `DeadLetter`, `DLQ`, `ConsumerGroup`

**Reference Implementation (GOOD):**
```go
// Reliable consumer with DLQ strategy
func (c *Consumer) Handle(msg *Message) error {
    if err := c.process(msg); err != nil {
        if msg.RetryCount >= maxRetries {
            // Move to Dead Letter Queue
            return c.dlq.Publish(msg)
        }
        // Retry with backoff
        return c.RetryLater(msg, backoff(msg.RetryCount))
    }
    return msg.Ack()
}
```

**Check For:**
1. Dead Letter Queues (DLQ) configured for failed messages
2. Retry policies with exponential backoff
3. Explicit Ack/Nack handling (no auto-ack)
4. Consumer groups for parallel processing
5. Graceful shutdown of consumers (wait for processing to finish)
6. Message durability settings (persistent queues)

**Severity Ratings:**
- CRITICAL: Messages auto-acked before processing
- HIGH: No DLQ for poison messages (infinite loops)
- HIGH: No retry backoff strategy
- MEDIUM: Missing graceful shutdown for workers

**Output Format:**
```
## Async Reliability Audit Findings

### Summary
- Async processing detected: Yes/No
- DLQ configured: Yes/No
- Retry strategy: Yes/No

### Critical Issues
[file:line] - Description

### Recommendations
1. ...
```
```

---

## Consolidated Report Template

After all explorers complete, generate this report:

```markdown
# Production Readiness Audit Report

**Date:** {YYYY-MM-DD}
**Codebase:** {project-name}
**Auditor:** Claude Code (Production Readiness Skill v2.0)

## Executive Summary

### Category A: Code Structure & Patterns

| Dimension | Score | Critical | High | Medium | Low |
|-----------|-------|----------|------|--------|-----|
| 1. Pagination Standards | X/10 | 0 | 0 | 0 | 0 |
| 2. Error Framework | X/10 | 0 | 0 | 0 | 0 |
| 3. Route Organization | X/10 | 0 | 0 | 0 | 0 |
| 4. Bootstrap & Init | X/10 | 0 | 0 | 0 | 0 |
| 5. Runtime Safety | X/10 | 0 | 0 | 0 | 0 |
| **Category A Total** | **X/50** | **0** | **0** | **0** | **0** |

### Category B: Security & Access Control

| Dimension | Score | Critical | High | Medium | Low |
|-----------|-------|----------|------|--------|-----|
| 6. Auth Protection | X/10 | 0 | 0 | 0 | 0 |
| 7. IDOR Protection | X/10 | 0 | 0 | 0 | 0 |
| 8. SQL Safety | X/10 | 0 | 0 | 0 | 0 |
| 9. Input Validation | X/10 | 0 | 0 | 0 | 0 |
| 10. Rate Limiting | X/10 | 0 | 0 | 0 | 0 |
| **Category B Total** | **X/50** | **0** | **0** | **0** | **0** |

### Category C: Operational Readiness

| Dimension | Score | Critical | High | Medium | Low |
|-----------|-------|----------|------|--------|-----|
| 11. Telemetry & Observability | X/10 | 0 | 0 | 0 | 0 |
| 12. Health Checks | X/10 | 0 | 0 | 0 | 0 |
| 13. Configuration Mgmt | X/10 | 0 | 0 | 0 | 0 |
| 14. Connection Mgmt | X/10 | 0 | 0 | 0 | 0 |
| 15. Logging & PII Safety | X/10 | 0 | 0 | 0 | 0 |
| **Category C Total** | **X/50** | **0** | **0** | **0** | **0** |

### Category D: Quality & Maintainability

| Dimension | Score | Critical | High | Medium | Low |
|-----------|-------|----------|------|--------|-----|
| 16. Idempotency | X/10 | 0 | 0 | 0 | 0 |
| 17. API Documentation | X/10 | 0 | 0 | 0 | 0 |
| 18. Technical Debt | X/10 | 0 | 0 | 0 | 0 |
| 19. Testing Coverage | X/10 | 0 | 0 | 0 | 0 |
| 20. Dependency Mgmt | X/10 | 0 | 0 | 0 | 0 |
| 21. Performance | X/10 | 0 | 0 | 0 | 0 |
| 22. Concurrency | X/10 | 0 | 0 | 0 | 0 |
| 23. Migrations | X/10 | 0 | 0 | 0 | 0 |
| **Category D Total** | **X/80** | **0** | **0** | **0** | **0** |

### Category E: Infrastructure & Hardening

| Dimension | Score | Critical | High | Medium | Low |
|-----------|-------|----------|------|--------|-----|
| 24. Container Security | X/10 | 0 | 0 | 0 | 0 |
| 25. HTTP Hardening | X/10 | 0 | 0 | 0 | 0 |
| 26. CI/CD & Licensing | X/10 | 0 | 0 | 0 | 0 |
| 27. Async Reliability | X/10 | 0 | 0 | 0 | 0 |
| **Category E Total** | **X/40** | **0** | **0** | **0** | **0** |

### Overall Score

| Metric | Value |
|--------|-------|
| **Total Score** | **X/270** |
| **Percentage** | **X%** |
| **Critical Issues** | **0** |
| **High Issues** | **0** |
| **Medium Issues** | **0** |
| **Low Issues** | **0** |

### Readiness Classification

| Score Range | Classification | Deployment Recommendation |
|-------------|----------------|---------------------------|
| 243-270 (90%+) | **Production Ready** | Clear to deploy |
| 203-242 (75-89%) | **Ready with Minor Remediation** | Deploy after addressing HIGH issues |
| 135-202 (50-74%) | **Needs Significant Work** | Do not deploy until CRITICAL/HIGH resolved |
| Below 135 (<50%) | **Not Production Ready** | Major remediation required |

**Current Status:** {classification}

## Critical Blockers (Must Fix Before Production)

{List all CRITICAL severity issues from all dimensions}

## High Priority Issues

{List all HIGH severity issues}

## Detailed Findings by Category

### Category A: Code Structure & Patterns

#### 1. Pagination Standards
{Agent 1 output}

#### 2. Error Framework
{Agent 2 output}

#### 3. Route Organization
{Agent 3 output}

#### 4. Bootstrap & Initialization
{Agent 4 output}

#### 5. Runtime Safety
{Agent 5 output}

### Category B: Security & Access Control

#### 6. Auth Protection
{Agent 6 output}

#### 7. IDOR & Access Control
{Agent 7 output}

#### 8. SQL Safety
{Agent 8 output}

#### 9. Input Validation
{Agent 10 output}

#### 10. Rate Limiting
{Agent 11 output}

### Category C: Operational Readiness

#### 11. Telemetry & Observability
{Agent 9 output}

#### 12. Health Checks
{Agent 12 output}

#### 13. Configuration Management
{Agent 13 output}

#### 14. Connection Management
{Agent 14 output}

#### 15. Logging & PII Safety
{Agent 15 output}

### Category D: Quality & Maintainability

#### 16. Idempotency
{Agent 16 output}

#### 17. API Documentation
{Agent 17 output}

#### 18. Technical Debt
{Agent 18 output}

#### 19. Testing Coverage
{Agent 19 output}

#### 20. Dependency Management
{Agent 20 output}

#### 21. Performance Patterns
{Agent 21 output}

#### 22. Concurrency Safety
{Agent 22 output}

#### 23. Migration Safety
{Agent 23 output}

### Category E: Infrastructure & Hardening

#### 24. Container Security
{Agent 24 output}

#### 25. HTTP Hardening
{Agent 25 output}

#### 26. CI/CD & Licensing
{Agent 26 output}

#### 27. Async Reliability
{Agent 27 output}

## Recommended Remediation Order

### 1. Immediate (before any deployment)
- {Critical security issues}
- {Critical data integrity issues}
- {Critical operational gaps}

### 2. Short-term (within 1 sprint)
- {High priority items}

### 3. Medium-term (within 1 quarter)
- {Medium priority items}

### 4. Backlog (track but don't block)
- {Low priority items}

## Appendix A: Files Audited

{List of all files examined with line counts}

## Appendix B: Audit Metadata

| Property | Value |
|----------|-------|
| Audit Duration | X minutes |
| Explorers Launched | 18 |
| Files Examined | X |
| Lines of Code | X |
| Skill Version | 2.0 |
```

---

## Scoring Guide

### Per-Dimension Scoring (0-10 each)

| Score | Criteria |
|-------|----------|
| 10 | Exemplary - could be used as reference implementation |
| 8-9 | Strong - minor improvements possible |
| 6-7 | Adequate - meets basic requirements |
| 4-5 | Concerning - multiple gaps identified |
| 2-3 | Poor - significant remediation needed |
| 0-1 | Critical - fundamentally broken or missing |

### Deductions Per Dimension

- Each CRITICAL issue: -3 points
- Each HIGH issue: -1.5 points
- Each MEDIUM issue: -0.5 points
- Each LOW issue: -0.25 points
- Minimum score: 0 (no negative scores)

### Category Weights

| Category | Dimensions | Max Score |
|----------|------------|-----------|
| A: Code Structure | 5 | 50 |
| B: Security | 5 | 50 |
| C: Operations | 5 | 50 |
| D: Quality | 8 | 80 |
| E: Infrastructure | 4 | 40 |
| **Total** | **27** | **270** |

### Overall Classification

| Score Range | Percentage | Classification |
|-------------|------------|----------------|
| 243-270 | 90%+ | Production Ready |
| 203-242 | 75-89% | Ready with Minor Remediation |
| 135-202 | 50-74% | Needs Significant Work |
| 0-134 | <50% | Not Production Ready |

---

## Usage Example

```
User: /production-readiness-audit
```

---

## Assistant Execution Protocol

When this skill is invoked, follow this exact protocol:

### Step 1: Initialize Todo List

```
TodoWrite: Create todos for all 3 batches + consolidation
```

### Step 2: Launch Parallel Explorers (Batch 1)

**CRITICAL**: Use a SINGLE response with 10 Task tool calls for Structure & Security dimensions.

```
// Batch 1: Agents 1-10
Task(subagent_type="Explore", prompt="<Agent 1: Pagination Standards>")
Task(subagent_type="Explore", prompt="<Agent 2: Error Framework>")
Task(subagent_type="Explore", prompt="<Agent 3: Route Organization>")
Task(subagent_type="Explore", prompt="<Agent 4: Bootstrap & Init>")
Task(subagent_type="Explore", prompt="<Agent 5: Runtime Safety>")
Task(subagent_type="Explore", prompt="<Agent 6: Auth Protection>")
Task(subagent_type="Explore", prompt="<Agent 7: IDOR Protection>")
Task(subagent_type="Explore", prompt="<Agent 8: SQL Safety>")
Task(subagent_type="Explore", prompt="<Agent 9: Input Validation>")
Task(subagent_type="Explore", prompt="<Agent 10: Rate Limiting>")
```

### Step 3: Launch Parallel Explorers (Batch 2)

Launch next 10 agents for Operational & Quality dimensions:

```
// Batch 2: Agents 11-20
Task(subagent_type="Explore", prompt="<Agent 11: Telemetry & Observability>")
Task(subagent_type="Explore", prompt="<Agent 12: Health Checks>")
Task(subagent_type="Explore", prompt="<Agent 13: Configuration Management>")
Task(subagent_type="Explore", prompt="<Agent 14: Connection Management>")
Task(subagent_type="Explore", prompt="<Agent 15: Logging & PII Safety>")
Task(subagent_type="Explore", prompt="<Agent 16: Idempotency>")
Task(subagent_type="Explore", prompt="<Agent 17: API Documentation>")
Task(subagent_type="Explore", prompt="<Agent 18: Technical Debt>")
Task(subagent_type="Explore", prompt="<Agent 19: Testing Coverage>")
Task(subagent_type="Explore", prompt="<Agent 20: Dependency Management>")
```

### Step 4: Launch Parallel Explorers (Batch 3)

Launch remaining 7 agents for Quality & Infrastructure dimensions:

```
// Batch 3: Agents 21-27
Task(subagent_type="Explore", prompt="<Agent 21: Performance Patterns>")
Task(subagent_type="Explore", prompt="<Agent 22: Concurrency Safety>")
Task(subagent_type="Explore", prompt="<Agent 23: Migration Safety>")
Task(subagent_type="Explore", prompt="<Agent 24: Container Security>")
Task(subagent_type="Explore", prompt="<Agent 25: HTTP Hardening>")
Task(subagent_type="Explore", prompt="<Agent 26: CI/CD & Licensing>")
Task(subagent_type="Explore", prompt="<Agent 27: Async Reliability>")
```

Each Task call should include:
- The full explorer prompt from the dimension
- Instruction to search the codebase thoroughly
- Expected output format reminder

### Step 5: Collect Results

As each explorer completes, mark its todo as completed.

### Step 6: Consolidate Report

Once ALL 27 explorers complete:
1. Calculate scores for each dimension (0-10 scale)
2. Calculate category totals (A: /50, B: /50, C: /50, D: /80, E: /40)
3. Calculate overall score (/270)
4. Aggregate critical/high/medium/low counts
5. Determine readiness classification
6. Generate the consolidated report

### Step 7: Write Report

```
Write: docs/audits/production-readiness-{YYYY-MM-DD}.md
```

### Step 7: Present Summary

Provide a verbal summary to the user including:
- Overall score and classification
- Number of critical/high issues
- Top 3 recommendations
- Link to full report

---

## Customization Options

Users can customize the audit:

### Scope Limiting

```
User: /production-readiness-audit --modules=matching,ingestion
```

Only audit specified modules.

### Dimension Selection

```
User: /production-readiness-audit --dimensions=security
```

Run only security-related auditors (6, 7, 8).

### Output Format

```
User: /production-readiness-audit --format=json
```

Output structured JSON instead of markdown.

---

## Integration with CI/CD

This skill can be automated:

1. Run audit on every release branch
2. Block merges if CRITICAL issues exist
3. Track debt trends over time
4. Generate dashboards from JSON output

---

## Reference Patterns Source

The reference implementations in this skill are derived from the Matcher codebase, which serves as the organizational standard for:

- Hexagonal architecture per bounded context
- lib-commons integration (telemetry, database, messaging)
- lib-auth integration (JWT validation, tenant extraction)
- pkg/assert and pkg/runtime usage patterns
- GORM + raw SQL hybrid approach
- Fiber HTTP framework conventions

When auditing other projects, findings are compared against these established patterns
