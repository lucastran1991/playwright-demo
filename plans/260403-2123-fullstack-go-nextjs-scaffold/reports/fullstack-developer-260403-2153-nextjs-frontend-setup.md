# Phase Implementation Report

## Executed Phase
- Phase: Next.js 15 frontend scaffold setup
- Plan: /Users/mac/studio/playwright-demo/plans/260403-2123-fullstack-go-nextjs-scaffold
- Status: completed

## Files Modified / Created

| File | Action | Notes |
|------|--------|-------|
| `frontend/` | created | Next.js 16.2.2 project via `pnpm create next-app` |
| `frontend/src/app/globals.css` | updated | Design tokens from design-guidelines.md (oklch, light + dark) |
| `frontend/src/app/layout.tsx` | updated | DM Sans + JetBrains Mono fonts, ThemeProvider + QueryProvider wrapping |
| `frontend/src/providers/theme-provider.tsx` | created | next-themes class-based dark mode |
| `frontend/src/providers/query-provider.tsx` | created | React Query context provider |
| `frontend/src/lib/query-client.ts` | created | Singleton QueryClient, 5min stale time, SSR-safe |
| `frontend/src/lib/api-client.ts` | created | Typed fetch wrapper for Go backend |
| `frontend/src/types/index.ts` | created | User, AuthResponse interfaces |
| `frontend/.env.local` | created | Local dev env vars (not committed) |
| `frontend/.env.example` | created | Template for env setup |

## Dependencies Installed
- `@tanstack/react-query` + `devtools` 5.96.2
- `next-auth@beta` 5.0.0-beta.30
- `react-hook-form` 7.72.1 + `@hookform/resolvers` + `zod`
- `next-themes` 0.4.6
- Shadcn/ui (button, input, label, card) — form component CLI silently skipped (registry issue), can be added later with `pnpm dlx shadcn@latest add form`

## Tasks Completed
- [x] Create Next.js 15 project with TypeScript, Tailwind v4, ESLint, App Router, src/ dir
- [x] Install all required dependencies
- [x] Initialize Shadcn/ui with defaults (Tailwind v4 detected, CSS variables)
- [x] Add base Shadcn components (button, input, label, card)
- [x] Update globals.css with design tokens (oklch colors, light + dark mode)
- [x] Create theme-provider.tsx
- [x] Create query-provider.tsx
- [x] Create lib/query-client.ts
- [x] Create lib/api-client.ts
- [x] Create types/index.ts
- [x] Update root layout with DM Sans + JetBrains Mono, wrap with providers
- [x] Create .env.local and .env.example
- [x] Build verification — passed clean

## Tests Status
- Type check: pass (TypeScript clean)
- Build: pass (`pnpm build` — 4/4 static pages generated)
- Unit tests: n/a (no tests in scope for this phase)

## Issues Encountered
1. `DM_Sans` does not accept `"vietnamese"` subset in next/font — only `"latin"` and `"latin-ext"` are valid. Fixed by using `"latin-ext"` which covers extended Latin characters including Vietnamese diacritics.
2. `shadcn add form` silently completed without creating the file — likely a registry issue with the current Shadcn version. The `form` component depends on `@radix-ui/react-label` which is already present via the `label` component.

## Next Steps
- Add `shadcn form` component manually or retry: `pnpm dlx shadcn@latest add form --overwrite`
- Backend API (Go) must be running on port 8080 for `apiFetch` calls to resolve
- `AUTH_SECRET` in `.env.local` must be rotated before any production deploy
