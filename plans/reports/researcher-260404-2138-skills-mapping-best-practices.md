---
name: Skills Activation Mapping for Development Rules
description: Mandatory skill activation rules by context (frontend, backend, testing) for this project
type: reference
---

# Skills Activation Mapping for Playwright-Demo

## Context: Tech Stack
- Backend: Go 1.25, Gin v1.12, GORM, PostgreSQL
- Frontend: Next.js 16, React 19.2, Tailwind v4, shadcn/ui, TanStack Query v5, NextAuth.js v5
- Testing: Go tests, Playwright (planned)
- Deployment: PM2, EC2

---

## 1. Frontend File Edits (TypeScript/React)

| Trigger | Mandatory Skills | Rationale |
|---------|-----------------|-----------|
| Any `.tsx` / `.ts` file | `frontend-development`, `docs-seeker` | React 19 has breaking changes. Lookup latest Next.js 16 upgrade guide before coding. |
| Component styling | `ui-styling`, `web-design-guidelines` | Tailwind v4 has breaking CSS variable syntax. Always reference design-guidelines.md + shadcn overrides. |
| Forms + validation | `react-best-practices`, `sequential-thinking` | React Hook Form + Zod patterns; validate schema structure before implementation. |
| Server/Client split | `docs-seeker` | Next.js 16 async APIs are mandatory. Lookup App Router patterns for request handling. |
| API data fetching | `tanstack` | TanStack Query v5 uses `isPending` not `isLoading`. Verify hook signatures. |
| Authentication logic | `sequential-thinking` | NextAuth.js v5 session management + security (HttpOnly, SameSite, auth secret rotation). |

**Key Insight:** Frontend work is context-heavy. `docs-seeker` is non-negotiable because training data will be stale on:
- React 19 hooks (useEffectEvent, useActionState)
- Next.js 16 Turbopack default + async Request APIs
- NextAuth.js v5 session strategy (JWT vs database sessions)

---

## 2. Backend File Edits (Go)

| Trigger | Mandatory Skills | Rationale |
|---------|-----------------|-----------|
| Any `.go` file | `backend-development`, `debug` | Go 1.25 linting + error handling patterns. Debug skill for race conditions. |
| GORM models/queries | `databases`, `sequential-thinking` | GORM connection pooling, N+1 queries, transaction patterns. Verify schema before ORM call. |
| Gin handlers | `backend-development` | Gin middleware chains, error responses, input validation binding. |
| JWT/Auth middleware | `sequential-thinking` | Token refresh logic, claim validation, error wrapping. |
| API contracts | `code-review` | Cross-layer validation (handler → service → repo). Ensure contracts match frontend expectations. |

**Key Insight:** Go ecosystem is stable but Gin + GORM integration patterns can be subtle (e.g., database connection pooling, lazy loading gotchas). Use `debug` for concurrency issues.

---

## 3. Testing (Unit + E2E)

| Trigger | Mandatory Skills | Rationale |
|---------|-----------------|-----------|
| Go unit tests | `test`, `debug` | Table-driven tests, mocking repos, assertion clarity. |
| Playwright E2E | `test`, `web-testing`, `chrome-devtools` | Browser automation, selectors, wait strategies, visual regression. |
| Coverage gaps | `debug`, `sequential-thinking` | Analyze root causes of uncovered branches. Don't just add tests. |
| Integration tests | `databases` | Real DB fixtures, transaction cleanup, race conditions. |

---

## 4. Styling & Responsive Design

| Trigger | Mandatory Skills | Rationale |
|---------|-----------------|-----------|
| Mobile-first layout | `ui-styling`, `web-design-guidelines` | Tailwind v4: unprefixed utilities apply mobile. Use `sm:/md:/lg:` modifiers for breakpoints. |
| Dark mode | `web-design-guidelines` | CSS variables via next-themes. Reference oklch color palette in design-guidelines.md. |
| Shadcn overrides | `ui-styling`, `web-design-guidelines` | Custom radius, shadows, spacing must align with design tokens (px-4 mobile, px-6 tablet, px-8 desktop). |

---

## 5. Database Work

| Trigger | Mandatory Skills | Rationale |
|---------|-----------------|-----------|
| Schema changes | `databases`, `sequential-thinking` | GORM migrations, backward compatibility, FK constraints. Validate migrations in dev first. |
| Performance tuning | `databases`, `debug` | Query profiling, index strategy, connection pooling. Use `psql` to analyze query plans. |
| Data integrity | `sequential-thinking` | Transaction design, constraint enforcement. Trace through service → repo layers. |

---

## 6. Debugging & Bug Fixes

| Trigger | Mandatory Skills | Rationale |
|---------|-----------------|-----------|
| Any production issue | `debug`, `chrome-devtools` (frontend) or `sequential-thinking` (backend) | Reproduce first, isolate root cause, verify fix. |
| React rendering bugs | `chrome-devtools`, `react-best-practices` | DevTools Profiler, component tree, hook dependencies. |
| Go panics/races | `debug`, `sequential-thinking` | Stack traces, race detector output, concurrency analysis. |

---

## 7. DevOps & Deployment

| Trigger | Mandatory Skills | Rationale |
|---------|-----------------|-----------|
| PM2 config / EC2 setup | `devops` | Environment variables, process management, scaling. |
| Docker / container builds | `devops` | Multi-stage builds, layer caching, security scanning. |

---

## 8. Cross-Cutting Concerns

| Trigger | Mandatory Skills | Rationale |
|---------|-----------------|-----------|
| Code review (all files) | `code-review` | Verify against code-standards.md. Check DRY, naming, error handling. |
| Architecture decisions | `sequential-thinking`, `problem-solving` | Multi-layer trace (handler → service → repo). Consider performance impact. |
| Documentation updates | `web-design-guidelines` (for design changes) | Keep docs/ in sync. Update code-standards.md + design-guidelines.md when patterns change. |
| Visual explanations | `ui-ux-pro-max`, `mermaidjs-v11` | Complex flows: use `/preview --diagram` to save visuals to plans/visuals/. |

---

## 9. Documentation Lookups (Mandatory via docs-seeker)

**ALWAYS check before coding:**

1. **React 19 Changes**
   - New hooks: useEffectEvent, useActionState, useFormState, useFormStatus
   - Action directives, Server Actions in Next.js 16
   - Training data (Feb 2025) will miss v19.2 features

2. **Next.js 16 Upgrade**
   - Async Request APIs (headers(), cookies(), searchParams) — no synchronous access
   - Turbopack default for next dev/build
   - React 19 integration + Canary features (View Transitions, Activity)

3. **NextAuth.js v5 Session Strategy**
   - JWT vs database sessions trade-off
   - Auth secret rotation (invalidates all sessions)
   - Server Action patterns for secure auth

4. **TanStack Query v5**
   - isPending (not isLoading)
   - useSuspenseQuery, suspense-ready hooks
   - Devtools integration

5. **Tailwind v4**
   - CSS variable syntax changes
   - Container queries (@container)
   - oklch color model (design-guidelines.md)

---

## 10. Skill Activation Decision Tree

```
File edited?
  ├─ .go → backend-development, databases (if GORM), debug
  ├─ .tsx/.ts → frontend-development, docs-seeker (ALWAYS), ui-styling (if CSS)
  ├─ .md → web-design-guidelines (if design), docs-seeker (if referencing patterns)
  
Stuck on design?
  └─ web-design-guidelines + ui-styling + design-guidelines.md

Debugging?
  ├─ Frontend → chrome-devtools, react-best-practices
  └─ Backend → debug, sequential-thinking

Writing tests?
  └─ test + (chrome-devtools OR sequential-thinking per language)

Building visuals?
  └─ mermaidjs-v11 + ui-ux-pro-max
```

---

## Summary Table

| Context | Primary | Secondary | Lookup First |
|---------|---------|-----------|--------------|
| Frontend component | frontend-development | react-best-practices | docs-seeker (React 19, Next.js 16) |
| Styling | ui-styling | web-design-guidelines | design-guidelines.md + Tailwind docs |
| Forms | react-best-practices | sequential-thinking | React Hook Form + Zod patterns |
| Backend handler | backend-development | code-review | — |
| Database | databases | sequential-thinking | GORM patterns + schema |
| Testing | test | debug | — |
| Authentication | sequential-thinking | react-best-practices | NextAuth.js v5 docs + patterns |
| Debugging | debug | (language-specific) | — |
| Responsive design | ui-styling | web-design-guidelines | Tailwind v4 breakpoint system |

---

## Key Takeaways

1. **docs-seeker is non-negotiable for frontend.** React 19 and Next.js 16 have material breaking changes.
2. **Backend is stable but requires precision.** GORM connection pooling and Gin error handling patterns are subtle.
3. **Testing must be real.** No mocks for database tests; use real fixtures + cleanup.
4. **Design consistency enforced.** Always cross-reference design-guidelines.md for color, spacing, responsive breaks.
5. **Sequential thinking for cross-layer work.** Handler → Service → Repo trace prevents architectural violations.

---

## Unresolved Questions

- Should we pre-configure React Compiler in next.config.ts (experimental feature, not enabled by default)?
- What's the plan for Playwright E2E test coverage? Should it be gated on all PRs?
- Should we enforce code-review skill on ALL code changes or only major features?
