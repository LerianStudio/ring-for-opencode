---
name: ring:react-best-practices
description: React and Next.js performance optimization guidelines. Use when writing, reviewing, or refactoring React/Next.js code to ensure optimal performance patterns.
license: MIT
compatibility: opencode
metadata:
  trigger: Writing/reviewing/refactoring React components, Next.js pages, data fetching, bundle optimization
  skip_when: Non-React projects, simple static content, no performance concerns
  author: Adapted from Vercel Engineering guidelines
  version: 1.0.0
---

# React Best Practices

Comprehensive performance optimization guide for React and Next.js applications. Contains 57 rules across 8 categories, prioritized by impact to guide automated refactoring and code generation.

## When to Apply

Reference these guidelines when:

- Writing new React components or Next.js pages
- Implementing data fetching (client or server-side)
- Reviewing code for performance issues
- Refactoring existing React/Next.js code
- Optimizing bundle size or load times

**Announce at start:** "I'm using the ring:react-best-practices skill to ensure optimal React performance."

## Rule Categories by Priority

| Priority | Category                  | Impact      | Prefix       |
| -------- | ------------------------- | ----------- | ------------ |
| 1        | Eliminating Waterfalls    | CRITICAL    | `async-`     |
| 2        | Bundle Size Optimization  | CRITICAL    | `bundle-`    |
| 3        | Server-Side Performance   | HIGH        | `server-`    |
| 4        | Client-Side Data Fetching | MEDIUM-HIGH | `client-`    |
| 5        | Re-render Optimization    | MEDIUM      | `rerender-`  |
| 6        | Rendering Performance     | MEDIUM      | `rendering-` |
| 7        | JavaScript Performance    | LOW-MEDIUM  | `js-`        |
| 8        | Advanced Patterns         | LOW         | `advanced-`  |

## Quick Reference

### 1. Eliminating Waterfalls (CRITICAL)

**Impact: 2-10x improvement. Waterfalls are the #1 performance killer.**

| Rule                        | Description                                                   |
| --------------------------- | ------------------------------------------------------------- |
| `async-defer-await`         | Move await into branches where actually used                  |
| `async-parallel`            | Use Promise.all() for independent operations                  |
| `async-dependencies`        | Use dependency-based parallelization for partial dependencies |
| `async-api-routes`          | Start promises early, await late in API routes                |
| `async-suspense-boundaries` | Use Suspense to stream content                                |

**Key pattern:**

```typescript
// BAD: Sequential (3 round trips)
const user = await fetchUser();
const posts = await fetchPosts();
const comments = await fetchComments();

// GOOD: Parallel (1 round trip)
const [user, posts, comments] = await Promise.all([
  fetchUser(),
  fetchPosts(),
  fetchComments(),
]);
```

### 2. Bundle Size Optimization (CRITICAL)

**Impact: 200-800ms import cost reduction, faster builds.**

| Rule                       | Description                                 |
| -------------------------- | ------------------------------------------- |
| `bundle-barrel-imports`    | Import directly, avoid barrel files         |
| `bundle-dynamic-imports`   | Use next/dynamic for heavy components       |
| `bundle-defer-third-party` | Load analytics/logging after hydration      |
| `bundle-conditional`       | Load modules only when feature is activated |
| `bundle-preload`           | Preload on hover/focus for perceived speed  |

**Key pattern:**

```typescript
// BAD: Loads 1,583 modules (~2.8s in dev)
import { Check, X, Menu } from "lucide-react";

// GOOD: Loads only 3 modules (~2KB vs ~1MB)
import Check from "lucide-react/dist/esm/icons/check";
import X from "lucide-react/dist/esm/icons/x";
import Menu from "lucide-react/dist/esm/icons/menu";

// ALTERNATIVE: Next.js 13.5+ optimizePackageImports
// next.config.js
module.exports = {
  experimental: {
    optimizePackageImports: ["lucide-react", "@mui/material"],
  },
};
```

### 3. Server-Side Performance (HIGH)

| Rule                       | Description                                     |
| -------------------------- | ----------------------------------------------- |
| `server-auth-actions`      | Authenticate server actions like API routes     |
| `server-cache-react`       | Use React.cache() for per-request deduplication |
| `server-cache-lru`         | Use LRU cache for cross-request caching         |
| `server-dedup-props`       | Avoid duplicate serialization in RSC props      |
| `server-serialization`     | Minimize data passed to client components       |
| `server-parallel-fetching` | Restructure components to parallelize fetches   |
| `server-after-nonblocking` | Use after() for non-blocking operations         |

**Key pattern:**

```typescript
// Server Action authentication (CRITICAL security)
"use server";

export async function deleteUser(userId: string) {
  // ALWAYS check auth inside the action
  const session = await verifySession();
  if (!session) throw unauthorized("Must be logged in");
  if (session.user.role !== "admin" && session.user.id !== userId) {
    throw unauthorized("Cannot delete other users");
  }
  await db.user.delete({ where: { id: userId } });
}
```

### 4. Client-Side Data Fetching (MEDIUM-HIGH)

| Rule                             | Description                                 |
| -------------------------------- | ------------------------------------------- |
| `client-swr-dedup`               | Use SWR for automatic request deduplication |
| `client-event-listeners`         | Deduplicate global event listeners          |
| `client-passive-event-listeners` | Use passive listeners for scroll            |
| `client-localstorage-schema`     | Version and minimize localStorage data      |

### 5. Re-render Optimization (MEDIUM)

| Rule                                 | Description                                     |
| ------------------------------------ | ----------------------------------------------- |
| `rerender-defer-reads`               | Don't subscribe to state only used in callbacks |
| `rerender-memo`                      | Extract expensive work into memoized components |
| `rerender-memo-with-default-value`   | Hoist default non-primitive props               |
| `rerender-dependencies`              | Use primitive dependencies in effects           |
| `rerender-derived-state`             | Subscribe to derived booleans, not raw values   |
| `rerender-derived-state-no-effect`   | Derive state during render, not effects         |
| `rerender-functional-setstate`       | Use functional setState for stable callbacks    |
| `rerender-lazy-state-init`           | Pass function to useState for expensive values  |
| `rerender-simple-expression-in-memo` | Avoid memo for simple primitives                |
| `rerender-move-effect-to-event`      | Put interaction logic in event handlers         |
| `rerender-transitions`               | Use startTransition for non-urgent updates      |
| `rerender-use-ref-transient-values`  | Use refs for transient frequent values          |

**Key pattern:**

```typescript
// BAD: Requires state as dependency, recreated on every change
const addItems = useCallback(
  (newItems: Item[]) => {
    setItems([...items, ...newItems]);
  },
  [items],
);

// GOOD: Stable callback, no stale closures
const addItems = useCallback((newItems: Item[]) => {
  setItems((curr) => [...curr, ...newItems]);
}, []);
```

### 6. Rendering Performance (MEDIUM)

| Rule                                   | Description                            |
| -------------------------------------- | -------------------------------------- |
| `rendering-animate-svg-wrapper`        | Animate div wrapper, not SVG element   |
| `rendering-content-visibility`         | Use content-visibility for long lists  |
| `rendering-hoist-jsx`                  | Extract static JSX outside components  |
| `rendering-svg-precision`              | Reduce SVG coordinate precision        |
| `rendering-hydration-no-flicker`       | Use inline script for client-only data |
| `rendering-hydration-suppress-warning` | Suppress expected mismatches           |
| `rendering-activity`                   | Use Activity component for show/hide   |
| `rendering-conditional-render`         | Use ternary, not && for conditionals   |
| `rendering-usetransition-loading`      | Prefer useTransition for loading state |

### 7. JavaScript Performance (LOW-MEDIUM)

| Rule                        | Description                                    |
| --------------------------- | ---------------------------------------------- |
| `js-batch-dom-css`          | Group CSS changes via classes or cssText       |
| `js-index-maps`             | Build Map for repeated lookups                 |
| `js-cache-property-access`  | Cache object properties in loops               |
| `js-cache-function-results` | Cache function results in module-level Map     |
| `js-cache-storage`          | Cache localStorage/sessionStorage reads        |
| `js-combine-iterations`     | Combine multiple filter/map into one loop      |
| `js-length-check-first`     | Check array length before expensive comparison |
| `js-early-exit`             | Return early from functions                    |
| `js-hoist-regexp`           | Hoist RegExp creation outside loops            |
| `js-min-max-loop`           | Use loop for min/max instead of sort           |
| `js-set-map-lookups`        | Use Set/Map for O(1) lookups                   |
| `js-tosorted-immutable`     | Use toSorted() for immutability                |

### 8. Advanced Patterns (LOW)

| Rule                          | Description                        |
| ----------------------------- | ---------------------------------- |
| `advanced-event-handler-refs` | Store event handlers in refs       |
| `advanced-init-once`          | Initialize app once per app load   |
| `advanced-use-latest`         | useLatest for stable callback refs |

## Code Review Checklist

When reviewing React/Next.js code, check for:

### Critical (Must Fix)

- [ ] No sequential awaits for independent operations
- [ ] No barrel file imports from large libraries
- [ ] Server Actions authenticate internally
- [ ] Heavy components use dynamic imports

### High Priority

- [ ] RSC boundaries minimize serialized data
- [ ] Components structured for parallel fetching
- [ ] React.cache() used for repeated server queries

### Medium Priority

- [ ] Functional setState for state-dependent updates
- [ ] Lazy state initialization for expensive values
- [ ] Derived state calculated during render, not effects
- [ ] Event handlers contain interaction logic, not effects

### Low Priority

- [ ] Static JSX hoisted outside components
- [ ] useTransition for non-urgent updates
- [ ] Set/Map used for O(1) lookups

## Rule Files

Read individual rule files for detailed explanations and code examples:

```
rules/async-parallel.md
rules/bundle-barrel-imports.md
rules/server-auth-actions.md
...
```

Each rule file contains:

- Brief explanation of why it matters
- Incorrect code example with explanation
- Correct code example with explanation
- Impact metrics and references

## Required Patterns

This skill uses these universal patterns:

- **State Tracking:** See `shared-patterns/state-tracking.md`
- **Exit Criteria:** See `shared-patterns/exit-criteria.md`

Apply ALL patterns when using this skill.

## References

- [Vercel React Best Practices](https://github.com/vercel-labs/agent-skills)
- [Next.js Documentation](https://nextjs.org/docs)
- [React Documentation](https://react.dev)
- [SWR Documentation](https://swr.vercel.app)
