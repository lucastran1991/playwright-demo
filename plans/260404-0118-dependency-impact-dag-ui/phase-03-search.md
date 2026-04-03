## Phase 3: Search Component

### Context Links
- [use-api.ts](../../frontend/src/hooks/use-api.ts) -- existing TanStack Query pattern
- [api-client.ts](../../frontend/src/lib/api-client.ts) -- `apiFetch` wrapper
- [dag-types.ts](../../frontend/src/components/tracer/dag-types.ts) -- `SearchNode`, `SearchResponse`

### Overview
- **Priority**: P1
- **Status**: pending
- **Description**: Autocomplete search bar overlaying ReactFlow canvas. Debounced query to `/api/blueprints/nodes`, dropdown with results, selection triggers trace.

### Key Insights
- TanStack Query already in project -- use `useQuery` with `enabled` flag for debounce pattern
- No need for a separate debounce lib; use `useState` + `useEffect` with setTimeout (KISS)
- Search overlay must be z-10+ to sit above ReactFlow canvas

### Requirements

**Functional**
- Text input with search/magnifying glass icon (lucide `Search`)
- Debounced API call (300ms) on input change
- Dropdown shows matching nodes: `node_id | name | node_type`
- Click result fires `onSelect(nodeId: string)` callback
- Clear button to reset search + DAG
- Loading spinner during fetch

**Non-functional**
- Under 100 lines
- Keyboard accessible (arrow keys, Enter to select -- stretch goal, not required for v1)

### Related Code Files

**Create:**
- `frontend/src/components/tracer/dag-search.tsx` (~100 lines)

**Read:**
- `frontend/src/components/tracer/dag-types.ts` -- SearchNode type
- `frontend/src/lib/api-client.ts` -- apiFetch

### Architecture

```
dag-search.tsx
  state: query (raw input), debouncedQuery (after 300ms)
  useQuery({ queryKey: ["tracer", "search", debouncedQuery], queryFn: ... , enabled: debouncedQuery.length >= 2 })
  renders: input + dropdown list
  callback: onSelect(nodeId) -> parent handles trace
```

### Implementation Steps

#### Step 1: Create `dag-search.tsx`

```tsx
'use client'

import { useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Search, X, Loader2 } from 'lucide-react'
import { apiFetch } from '@/lib/api-client'
import type { SearchNode } from './dag-types'
```

**Props:**
```ts
interface DagSearchProps {
  onSelect: (nodeId: string) => void
  onClear: () => void
}
```

**Debounce pattern:**
```tsx
const [query, setQuery] = useState('')
const [debouncedQuery, setDebouncedQuery] = useState('')

useEffect(() => {
  const timer = setTimeout(() => setDebouncedQuery(query), 300)
  return () => clearTimeout(timer)
}, [query])
```

**TanStack Query:**
```tsx
const { data: results, isLoading } = useQuery({
  queryKey: ['tracer', 'search', debouncedQuery],
  queryFn: () => apiFetch<{ data: SearchNode[] }>(
    `/api/blueprints/nodes?type=&limit=20&q=${encodeURIComponent(debouncedQuery)}`
  ).then(r => r.data),
  enabled: debouncedQuery.length >= 2,
})
```

Note: verify actual query param name from backend. Brainstorm says `?type=&limit=20` but search term param may be `q` or `search` or `name` -- check API.

**Dropdown visibility:**
```tsx
const showDropdown = debouncedQuery.length >= 2 && (isLoading || (results && results.length > 0))
```

**Selection handler:**
```tsx
function handleSelect(node: SearchNode) {
  setQuery(node.name || node.node_id)
  onSelect(node.node_id)
  // dropdown closes because query changes but we don't re-search
}
```

**Render structure:**
```
<div className="absolute top-4 left-1/2 -translate-x-1/2 z-10 w-[400px]">
  <div className="relative">
    <Search icon (left side of input)>
    <input ... />
    {query && <X button (right side, calls onClear + setQuery(''))>}
    {isLoading && <Loader2 spinner>}
  </div>
  {showDropdown && (
    <div className="absolute top-full mt-1 w-full bg-card border rounded-lg shadow-xl max-h-[300px] overflow-y-auto">
      {results.map(node => (
        <button onClick={() => handleSelect(node)} className="w-full text-left px-3 py-2 hover:bg-muted ...">
          <span className="font-mono text-xs">{node.node_id}</span>
          <span className="text-sm">{node.name}</span>
          <span className="text-xs text-muted-foreground">{node.node_type}</span>
        </button>
      ))}
    </div>
  )}
</div>
```

**Styling notes:**
- Use `bg-card` or `theme-card` for dropdown background
- `backdrop-blur-sm` on input container for glass effect over ReactFlow
- Border: `border border-border`
- Focus ring on input: `focus:ring-2 focus:ring-primary/50`

### Todo List
- [ ] Create `dag-search.tsx` with debounced autocomplete
- [ ] Verify search API query param name against backend
- [ ] Test dropdown renders + selection works
- [ ] Verify `pnpm tsc --noEmit` passes

### Success Criteria
- Typing 2+ chars triggers debounced API call
- Results appear in dropdown within 300ms + network time
- Clicking result fires `onSelect` with correct node_id
- Clear button resets input and calls `onClear`
- Dropdown dismisses after selection

### Risk Assessment
- **Medium**: Search API query param name unknown -- need to verify against backend. Brainstorm shows `?type=&limit=20` but no search term param specified. Check backend route handler.
- **Low**: Dropdown may need click-outside-to-close behavior. Can add later if needed.

### Security Considerations
- `encodeURIComponent` on search input to prevent injection
- No auth needed per design decisions

### Next Steps
- Phase 4 integrates search into main DAG component
