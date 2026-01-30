# React Best Practices

**Version 1.0.0**  
Adapted from Vercel Engineering Guidelines  
January 2026

> **Note:**  
> This document is for AI agents and LLMs to follow when maintaining,  
> generating, or refactoring React and Next.js codebases. Humans  
> may also find it useful, but guidance here is optimized for automation  
> and consistency by AI-assisted workflows.

---

## Abstract

Comprehensive performance optimization guide for React and Next.js applications, designed for AI agents and LLMs. Contains 57 rules across 8 categories, prioritized by impact from critical (eliminating waterfalls, reducing bundle size) to incremental (advanced patterns). Each rule includes detailed explanations, real-world examples comparing incorrect vs. correct implementations, and specific impact metrics to guide automated refactoring and code generation.

---

## Table of Contents

1. [Eliminating Waterfalls](#1-eliminating-waterfalls) - **CRITICAL**
2. [Bundle Size Optimization](#2-bundle-size-optimization) - **CRITICAL**
3. [Server-Side Performance](#3-server-side-performance) - **HIGH**
4. [Client-Side Data Fetching](#4-client-side-data-fetching) - **MEDIUM-HIGH**
5. [Re-render Optimization](#5-re-render-optimization) - **MEDIUM**
6. [Rendering Performance](#6-rendering-performance) - **MEDIUM**
7. [JavaScript Performance](#7-javascript-performance) - **LOW-MEDIUM**
8. [Advanced Patterns](#8-advanced-patterns) - **LOW**

---

## 1. Eliminating Waterfalls

**Impact: CRITICAL**

Waterfalls are the #1 performance killer. Each sequential await adds full network latency. Eliminating them yields the largest gains.

### 1.1 Promise.all() for Independent Operations

**Impact: CRITICAL (2-10x improvement)**

When async operations have no interdependencies, execute them concurrently using `Promise.all()`.

**Incorrect: sequential execution, 3 round trips**

```typescript
const user = await fetchUser()
const posts = await fetchPosts()
const comments = await fetchComments()
```

**Correct: parallel execution, 1 round trip**

```typescript
const [user, posts, comments] = await Promise.all([
  fetchUser(),
  fetchPosts(),
  fetchComments()
])
```

### 1.2 Defer Await Until Needed

**Impact: HIGH (avoids blocking unused code paths)**

Move `await` operations into the branches where they're actually used.

**Incorrect: blocks both branches**

```typescript
async function handleRequest(userId: string, skipProcessing: boolean) {
  const userData = await fetchUserData(userId)
  
  if (skipProcessing) {
    return { skipped: true }  // Still waited for userData
  }
  
  return processUserData(userData)
}
```

**Correct: only blocks when needed**

```typescript
async function handleRequest(userId: string, skipProcessing: boolean) {
  if (skipProcessing) {
    return { skipped: true }  // Returns immediately
  }
  
  const userData = await fetchUserData(userId)
  return processUserData(userData)
}
```

### 1.3 Prevent Waterfall Chains in API Routes

**Impact: CRITICAL (2-10x improvement)**

Start independent operations immediately, even if you don't await them yet.

**Incorrect: config waits for auth, data waits for both**

```typescript
export async function GET(request: Request) {
  const session = await auth()
  const config = await fetchConfig()
  const data = await fetchData(session.user.id)
  return Response.json({ data, config })
}
```

**Correct: auth and config start immediately**

```typescript
export async function GET(request: Request) {
  const sessionPromise = auth()
  const configPromise = fetchConfig()
  const session = await sessionPromise
  const [config, data] = await Promise.all([
    configPromise,
    fetchData(session.user.id)
  ])
  return Response.json({ data, config })
}
```

### 1.4 Strategic Suspense Boundaries

**Impact: HIGH (faster initial paint)**

Use Suspense boundaries to show wrapper UI faster while data loads.

**Incorrect: wrapper blocked by data fetching**

```tsx
async function Page() {
  const data = await fetchData()  // Blocks entire page
  
  return (
    <div>
      <Sidebar />
      <Header />
      <DataDisplay data={data} />
      <Footer />
    </div>
  )
}
```

**Correct: wrapper shows immediately**

```tsx
function Page() {
  return (
    <div>
      <Sidebar />
      <Header />
      <Suspense fallback={<Skeleton />}>
        <DataDisplay />
      </Suspense>
      <Footer />
    </div>
  )
}

async function DataDisplay() {
  const data = await fetchData()
  return <div>{data.content}</div>
}
```

---

## 2. Bundle Size Optimization

**Impact: CRITICAL**

Reducing initial bundle size improves Time to Interactive and Largest Contentful Paint.

### 2.1 Avoid Barrel File Imports

**Impact: CRITICAL (200-800ms import cost, slow builds)**

Import directly from source files instead of barrel files.

**Incorrect: imports entire library**

```tsx
import { Check, X, Menu } from 'lucide-react'
// Loads 1,583 modules, takes ~2.8s extra in dev
```

**Correct: imports only what you need**

```tsx
import Check from 'lucide-react/dist/esm/icons/check'
import X from 'lucide-react/dist/esm/icons/x'
import Menu from 'lucide-react/dist/esm/icons/menu'
// Loads only 3 modules (~2KB vs ~1MB)
```

**Alternative: Next.js 13.5+**

```js
// next.config.js
module.exports = {
  experimental: {
    optimizePackageImports: ['lucide-react', '@mui/material']
  }
}
```

### 2.2 Dynamic Imports for Heavy Components

**Impact: CRITICAL (directly affects TTI and LCP)**

Use `next/dynamic` to lazy-load large components.

**Incorrect: Monaco bundles with main chunk ~300KB**

```tsx
import { MonacoEditor } from './monaco-editor'
```

**Correct: Monaco loads on demand**

```tsx
import dynamic from 'next/dynamic'

const MonacoEditor = dynamic(
  () => import('./monaco-editor').then(m => m.MonacoEditor),
  { ssr: false }
)
```

### 2.3 Defer Non-Critical Third-Party Libraries

**Impact: MEDIUM (loads after hydration)**

Analytics, logging, and error tracking don't block user interaction.

**Correct: loads after hydration**

```tsx
import dynamic from 'next/dynamic'

const Analytics = dynamic(
  () => import('@vercel/analytics/react').then(m => m.Analytics),
  { ssr: false }
)
```

### 2.4 Preload Based on User Intent

**Impact: MEDIUM (reduces perceived latency)**

```tsx
function EditorButton({ onClick }: { onClick: () => void }) {
  const preload = () => {
    void import('./monaco-editor')
  }

  return (
    <button
      onMouseEnter={preload}
      onFocus={preload}
      onClick={onClick}
    >
      Open Editor
    </button>
  )
}
```

---

## 3. Server-Side Performance

**Impact: HIGH**

### 3.1 Authenticate Server Actions Like API Routes

**Impact: CRITICAL (prevents unauthorized access)**

Server Actions are exposed as public endpoints. Always verify authentication inside each action.

**Incorrect: no authentication check**

```typescript
'use server'

export async function deleteUser(userId: string) {
  await db.user.delete({ where: { id: userId } })
}
```

**Correct: authentication inside the action**

```typescript
'use server'

export async function deleteUser(userId: string) {
  const session = await verifySession()
  
  if (!session) {
    throw new Error('Unauthorized')
  }
  
  if (session.user.role !== 'admin' && session.user.id !== userId) {
    throw new Error('Cannot delete other users')
  }
  
  await db.user.delete({ where: { id: userId } })
}
```

### 3.2 Per-Request Deduplication with React.cache()

**Impact: MEDIUM (deduplicates within request)**

```typescript
import { cache } from 'react'

export const getCurrentUser = cache(async () => {
  const session = await auth()
  if (!session?.user?.id) return null
  return await db.user.findUnique({
    where: { id: session.user.id }
  })
})
```

### 3.3 Minimize Serialization at RSC Boundaries

**Impact: HIGH (reduces data transfer size)**

Only pass fields that the client actually uses.

**Incorrect: serializes all 50 fields**

```tsx
async function Page() {
  const user = await fetchUser()
  return <Profile user={user} />
}
```

**Correct: serializes only needed fields**

```tsx
async function Page() {
  const user = await fetchUser()
  return <Profile name={user.name} avatar={user.avatar} />
}
```

### 3.4 Use after() for Non-Blocking Operations

**Impact: MEDIUM (faster response times)**

```tsx
import { after } from 'next/server'

export async function POST(request: Request) {
  await updateDatabase(request)
  
  after(async () => {
    logUserAction({ ... })  // Runs after response sent
  })
  
  return Response.json({ status: 'success' })
}
```

---

## 4. Client-Side Data Fetching

**Impact: MEDIUM-HIGH**

### 4.1 Use SWR for Automatic Deduplication

**Impact: MEDIUM-HIGH (automatic deduplication)**

```tsx
import useSWR from 'swr'

function UserList() {
  const { data: users } = useSWR('/api/users', fetcher)
  return <List items={users ?? []} />
}
// Multiple instances share one request
```

### 4.2 Use Passive Event Listeners

**Impact: MEDIUM (eliminates scroll delay)**

```typescript
document.addEventListener('touchstart', handler, { passive: true })
document.addEventListener('wheel', handler, { passive: true })
```

### 4.3 Version and Minimize localStorage Data

**Impact: MEDIUM (prevents schema conflicts)**

```typescript
const VERSION = 'v2'

function saveConfig(config: Config) {
  try {
    localStorage.setItem(`userConfig:${VERSION}`, JSON.stringify(config))
  } catch {}
}
```

---

## 5. Re-render Optimization

**Impact: MEDIUM**

### 5.1 Use Functional setState Updates

**Impact: MEDIUM (prevents stale closures)**

**Incorrect: requires state as dependency**

```tsx
const addItems = useCallback((newItems: Item[]) => {
  setItems([...items, ...newItems])
}, [items])  // Recreated on every change
```

**Correct: stable callback**

```tsx
const addItems = useCallback((newItems: Item[]) => {
  setItems(curr => [...curr, ...newItems])
}, [])  // Never recreated
```

### 5.2 Calculate Derived State During Rendering

**Impact: MEDIUM (avoids redundant renders)**

**Incorrect: redundant state and effect**

```tsx
const [firstName, setFirstName] = useState('First')
const [fullName, setFullName] = useState('')

useEffect(() => {
  setFullName(firstName + ' ' + lastName)
}, [firstName, lastName])
```

**Correct: derive during render**

```tsx
const [firstName, setFirstName] = useState('First')
const fullName = firstName + ' ' + lastName
```

### 5.3 Use Lazy State Initialization

**Impact: MEDIUM (avoids wasted computation)**

**Incorrect: runs on every render**

```tsx
const [settings, setSettings] = useState(
  JSON.parse(localStorage.getItem('settings') || '{}')
)
```

**Correct: runs only once**

```tsx
const [settings, setSettings] = useState(() => {
  const stored = localStorage.getItem('settings')
  return stored ? JSON.parse(stored) : {}
})
```

### 5.4 Use Transitions for Non-Urgent Updates

**Impact: MEDIUM (maintains UI responsiveness)**

```tsx
import { startTransition } from 'react'

const handler = () => {
  startTransition(() => setScrollY(window.scrollY))
}
```

### 5.5 Extract to Memoized Components

**Impact: MEDIUM (enables early returns)**

```tsx
const UserAvatar = memo(function UserAvatar({ user }: { user: User }) {
  const id = useMemo(() => computeAvatarId(user), [user])
  return <Avatar id={id} />
})

function Profile({ user, loading }: Props) {
  if (loading) return <Skeleton />  // Skips avatar computation
  return <UserAvatar user={user} />
}
```

---

## 6. Rendering Performance

**Impact: MEDIUM**

### 6.1 Use Explicit Conditional Rendering

**Impact: MEDIUM (prevents falsy value bugs)**

**Incorrect: renders "0" when count is 0**

```tsx
{count && <span>{count} messages</span>}
```

**Correct: explicit check**

```tsx
{count > 0 ? <span>{count} messages</span> : null}
```

### 6.2 Hoist Static JSX Elements

**Impact: LOW-MEDIUM**

```tsx
// Outside component - created once
const EMPTY_STATE = <div className="empty">No items</div>

function List({ items }: Props) {
  if (!items.length) return EMPTY_STATE
  return <ul>{items.map(...)}</ul>
}
```

### 6.3 CSS content-visibility for Long Lists

**Impact: MEDIUM**

```css
.list-item {
  content-visibility: auto;
  contain-intrinsic-size: 0 50px;
}
```

---

## 7. JavaScript Performance

**Impact: LOW-MEDIUM**

### 7.1 Use Set/Map for O(1) Lookups

**Incorrect: O(n) lookup**

```tsx
{selectedIds.includes(user.id) && <CheckIcon />}
```

**Correct: O(1) lookup**

```tsx
const selectedSet = useMemo(() => new Set(selectedIds), [selectedIds])
{selectedSet.has(user.id) && <CheckIcon />}
```

### 7.2 Early Return from Functions

```typescript
function process(data: Data | null) {
  if (!data) return null
  if (!data.valid) return { error: 'Invalid' }
  
  // Main logic only runs when needed
  return transform(data)
}
```

### 7.3 Cache Property Access in Loops

```typescript
// Cache length for large loops
for (let i = 0, len = items.length; i < len; i++) {
  // ...
}
```

### 7.4 Use toSorted() for Immutability

```typescript
// Creates new array, doesn't mutate original
const sorted = items.toSorted((a, b) => a.name.localeCompare(b.name))
```

---

## 8. Advanced Patterns

**Impact: LOW**

### 8.1 Initialize App Once, Not Per Mount

```typescript
let initialized = false

function App() {
  useEffect(() => {
    if (!initialized) {
      initialized = true
      initializeAnalytics()
    }
  }, [])
}
```

### 8.2 Store Event Handlers in Refs

For handlers that need latest values without causing re-renders:

```typescript
const handlerRef = useRef(handler)
handlerRef.current = handler

useEffect(() => {
  const listener = (e: Event) => handlerRef.current(e)
  window.addEventListener('resize', listener)
  return () => window.removeEventListener('resize', listener)
}, [])
```

---

## Code Review Checklist

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

---

## References

- [Vercel React Best Practices](https://github.com/vercel-labs/agent-skills)
- [Next.js Documentation](https://nextjs.org/docs)
- [React Documentation](https://react.dev)
- [SWR Documentation](https://swr.vercel.app)
- [You Might Not Need an Effect](https://react.dev/learn/you-might-not-need-an-effect)
