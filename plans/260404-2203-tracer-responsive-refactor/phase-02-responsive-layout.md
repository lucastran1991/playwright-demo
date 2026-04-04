# Phase 2: Responsive Layout Refactor

## Overview
- **Priority:** P0
- **Status:** Complete
- Apply mobile-first responsive patterns across all tracer components using Tailwind v4 breakpoints

## Key Insights
- Tailwind v4: unprefixed = mobile, `sm:` = 640px+, `md:` = 768px+, `lg:` = 1024px+
- `h-dvh` handles mobile browser chrome (address bar) better than `h-screen`
- ReactFlow Controls should be repositioned to avoid toolbar overlap on mobile
- Detail popup should go near-full-screen on mobile

## Related Code Files
- Modify: `frontend/src/app/tracer/layout.tsx`
- Modify: `frontend/src/components/tracer/dependency-impact-dag.tsx`
- Modify: `frontend/src/components/tracer/dag-node.tsx`
- Modify: `frontend/src/components/tracer/dag-search.tsx`
- Modify: `frontend/src/components/tracer/dag-detail-popup.tsx`

## Implementation Steps

### 2.1 Layout — `layout.tsx`
1. Change `h-screen` to `h-dvh` for proper mobile viewport height

### 2.2 Main DAG — `dependency-impact-dag.tsx`
1. Change outer `h-screen` to `h-dvh`
2. Toolbar: wrap depth control and search in a responsive row
   - On mobile (< sm): hide "Depth" text label (already done), reduce padding
   - On sm+: show full toolbar as-is
3. ReactFlow Controls: add `position="bottom-right"` and custom responsive styles
4. Empty state text: adjust font sizes for mobile
5. Loading overlay: keep as-is (works responsive already)

### 2.3 Search — `dag-search.tsx`
1. Search results dropdown: add `absolute left-0 right-0` for full-width on mobile
2. Type filter dropdown: position below search on mobile, right-aligned on sm+
3. Input padding: reduce on mobile for more text space
4. Type filter button: icon-only on mobile (hide text), show text on `sm:`

### 2.4 Node — `dag-node.tsx`
1. Reduce `min-w-[180px]` to `min-w-[140px]` — nodes slightly smaller on mobile
2. Keep `max-w-[200px]` — cap still reasonable
3. Font sizes already small, no change needed

### 2.5 Detail Popup — `dag-detail-popup.tsx`
1. Mobile: `w-[calc(100%-1rem)]` with `max-w-[360px]` on `sm:`
2. Add `max-h-[80dvh] overflow-y-auto` for very small screens
3. Status pills: `grid-cols-2 sm:grid-cols-3` to prevent overflow on narrow screens

### 2.6 Helpers — `dag-helpers.tsx`
1. Reduce `NODE_WIDTH` from 200 to 180 for tighter DAG layout
2. Reduce `ranksep` from 160 to 120 for more compact horizontal spacing on mobile
3. These are initial positions — `fitView()` auto-scales anyway

## Todo
- [x] layout.tsx: `h-screen` → `h-dvh`
- [x] dependency-impact-dag.tsx: `h-screen` → `h-dvh`, toolbar padding
- [x] dag-search.tsx: type filter icon-only on mobile, full-width dropdowns
- [x] dag-node.tsx: smaller min-width
- [x] dag-detail-popup.tsx: full-width on mobile, scrollable, 2-col pills
- [x] dag-helpers.tsx: tighter layout constants

## Success Criteria
- Toolbar usable on 360px viewport without overflow
- DAG canvas takes full remaining height with `dvh`
- Detail popup readable on mobile without horizontal scroll
- No layout shift when mobile browser chrome appears/disappears

## Risk Assessment
- `h-dvh` has good browser support (96%+) but fallback to `h-screen` if needed via `h-screen h-dvh` pattern
- Smaller node width may truncate long node IDs — acceptable since truncation already exists
