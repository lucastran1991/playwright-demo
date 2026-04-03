## Phase 4: Main DAG Component + Page Route + Sidebar Nav

### Context Links
- [PipelineDAG.tsx](../../PipelineDAG.tsx) -- main component pattern (line 382-471)
- [dashboard layout](../../frontend/src/app/(dashboard)/layout.tsx) -- existing layout
- [sidebar-nav.tsx](../../frontend/src/components/dashboard/sidebar-nav.tsx) -- nav items array
- All Phase 1-3 components

### Overview
- **Priority**: P1
- **Status**: pending
- **Description**: Wire everything together -- main DAG component orchestrating search + ReactFlow + data fetching, page route at `/dashboard/tracer`, and sidebar nav entry.

### Key Insights
- PipelineDAG uses `useNodesState`/`useEdgesState` + `useEffect` sync pattern -- reuse exactly
- `ReactFlowProvider` must wrap the inner component (same pattern)
- `fitView` after data changes via `requestAnimationFrame` (PipelineDAG line 416-417)
- Page is a simple server component that renders the client DAG component
- Sidebar already uses a `navItems` array -- just append one entry

### Requirements

**Functional**
- Select node via search -> parallel fetch deps + impacts -> merge -> layout -> render DAG
- Empty state: show centered message "Search for a node to trace dependencies"
- Loading state: spinner overlay
- Error state: inline error message
- ReactFlow controls (zoom, fit) + background grid

**Non-functional**
- Main DAG component under 150 lines
- Page route under 20 lines

### Related Code Files

**Create:**
- `frontend/src/components/tracer/dependency-impact-dag.tsx` (~150 lines)
- `frontend/src/app/(dashboard)/tracer/page.tsx` (~20 lines)

**Modify:**
- `frontend/src/components/dashboard/sidebar-nav.tsx` -- add Tracer nav item

### Implementation Steps

#### Step 1: Create `dependency-impact-dag.tsx`

```tsx
// @ts-nocheck
'use client'

import { useMemo, useEffect, useCallback, useState } from 'react'
import {
  ReactFlow, ReactFlowProvider, useReactFlow,
  Background, Controls,
  useNodesState, useEdgesState,
  ConnectionLineType,
} from '@xyflow/react'
import { useQuery } from '@tanstack/react-query'
import { apiFetch } from '@/lib/api-client'
import { DagSearch } from './dag-search'
import { TracerNode } from './dag-node'
import { TracerEdge } from './dag-edge'
import { traceToDAGElements, layoutDAG } from './dag-helpers'
import type { TraceResponse } from './dag-types'
import '@xyflow/react/dist/style.css'
```

**Inner component (`TracerDAGInner`):**

State:
```tsx
const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null)
const nodeTypes = useMemo(() => ({ tracer: TracerNode }), [])
const edgeTypes = useMemo(() => ({ tracerEdge: TracerEdge }), [])
const { fitView } = useReactFlow()
```

Data fetching -- two parallel TanStack queries:
```tsx
const { data: depData, isLoading: depLoading } = useQuery({
  queryKey: ['tracer', 'dependencies', selectedNodeId],
  queryFn: () => apiFetch<TraceResponse>(
    `/api/trace/dependencies/${selectedNodeId}?levels=2&include_local=true`
  ),
  enabled: !!selectedNodeId,
})

const { data: impactData, isLoading: impactLoading } = useQuery({
  queryKey: ['tracer', 'impacts', selectedNodeId],
  queryFn: () => apiFetch<TraceResponse>(
    `/api/trace/impacts/${selectedNodeId}?levels=2`
  ),
  enabled: !!selectedNodeId,
})
```

Merge + layout:
```tsx
const { nodes: layoutedNodes, edges: layoutedEdges } = useMemo(() => {
  if (!depData && !impactData) return { nodes: [], edges: [] }
  const { nodes, edges } = traceToDAGElements(depData ?? null, impactData ?? null)
  return layoutDAG(nodes, edges)
}, [depData, impactData])
```

ReactFlow state sync (same pattern as PipelineDAG lines 408-417):
```tsx
const [nodes, setNodes, onNodesChange] = useNodesState(layoutedNodes)
const [edges, setEdges, onEdgesChange] = useEdgesState(layoutedEdges)

useEffect(() => {
  setNodes(layoutedNodes)
  setEdges(layoutedEdges)
  requestAnimationFrame(() => fitView({ padding: 0.25, duration: 300 }))
}, [layoutedNodes, layoutedEdges, setNodes, setEdges, fitView])
```

Handlers:
```tsx
const handleSelect = useCallback((nodeId: string) => setSelectedNodeId(nodeId), [])
const handleClear = useCallback(() => setSelectedNodeId(null), [])
```

Loading state:
```tsx
const isLoading = depLoading || impactLoading
```

**Render:**
```tsx
return (
  <div className="relative w-full h-[calc(100vh-8rem)] rounded-xl border border-border overflow-hidden">
    <DagSearch onSelect={handleSelect} onClear={handleClear} />

    {isLoading && (
      <div className="absolute inset-0 z-20 flex items-center justify-center bg-background/50 backdrop-blur-[1px]">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )}

    {!selectedNodeId && !isLoading && (
      <div className="absolute inset-0 flex items-center justify-center text-muted-foreground">
        <p>Search for a node to trace dependencies and impacts</p>
      </div>
    )}

    <ReactFlow
      nodes={nodes}
      edges={edges}
      onNodesChange={onNodesChange}
      onEdgesChange={onEdgesChange}
      nodeTypes={nodeTypes}
      edgeTypes={edgeTypes}
      connectionLineType={ConnectionLineType.SmoothStep}
      fitView
      fitViewOptions={{ padding: 0.25 }}
      proOptions={{ hideAttribution: true }}
      nodesDraggable={false}
      nodesConnectable={false}
      elementsSelectable={false}
      minZoom={0.2}
      maxZoom={2}
    >
      <Background gap={20} size={1} color="hsl(var(--border) / 0.3)" />
      <Controls
        showInteractive={false}
        className="!bg-card !border-border !shadow-lg [&>button]:!bg-card [&>button]:!border-border [&>button]:!fill-foreground"
      />
    </ReactFlow>
  </div>
)
```

**Wrapper component:**
```tsx
export default function DependencyImpactDAG() {
  return (
    <ReactFlowProvider>
      <TracerDAGInner />
    </ReactFlowProvider>
  )
}
```

#### Step 2: Create `page.tsx`

```tsx
import DependencyImpactDAG from '@/components/tracer/dependency-impact-dag'

export const metadata = {
  title: 'Dependency Tracer',
  description: 'Trace node dependencies and downstream impacts',
}

export default function TracerPage() {
  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-bold">Dependency Tracer</h1>
        <p className="text-muted-foreground text-sm">
          Search for a node to visualize its upstream dependencies and downstream impacts
        </p>
      </div>
      <DependencyImpactDAG />
    </div>
  )
}
```

#### Step 3: Update `sidebar-nav.tsx`

Add to `navItems` array (between Dashboard and Settings):
```tsx
{
  title: "Tracer",
  href: "/dashboard/tracer",
  icon: (
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-4 w-4">
      <circle cx="12" cy="12" r="1" />
      <path d="M20.2 20.2c2.04-2.03.02-7.36-4.5-11.9-4.54-4.52-9.87-6.54-11.9-4.5-2.04 2.03-.02 7.36 4.5 11.9 4.54 4.52 9.87 6.54 11.9 4.5Z" />
      <path d="M15.7 15.7c4.52-4.54 6.54-9.87 4.5-11.9-2.03-2.04-7.36-.02-11.9 4.5-4.52 4.54-6.54 9.87-4.5 11.9 2.03 2.04 7.36.02 11.9-4.5Z" />
    </svg>
  ),
},
```

Alternatively, use lucide-react `Network` icon if simpler:
```tsx
import { Network } from 'lucide-react'
// ...
icon: <Network className="h-4 w-4" />,
```

Note: existing sidebar uses inline SVGs. Match the pattern OR switch to lucide if team prefers consistency. Recommend lucide for simplicity since project already depends on it.

### Todo List
- [ ] Create `dependency-impact-dag.tsx` with ReactFlow + TanStack Query
- [ ] Create `tracer/page.tsx` route
- [ ] Add Tracer nav item to `sidebar-nav.tsx`
- [ ] Verify full page renders: `pnpm dev` + navigate to `/dashboard/tracer`
- [ ] Test search -> select -> DAG renders flow
- [ ] Verify `pnpm build` passes (no SSR issues with ReactFlow)

### Success Criteria
- `/dashboard/tracer` accessible via sidebar nav
- Empty state shows when no node selected
- Selecting a node triggers parallel API calls
- DAG renders with correct layout (upstream left, source center, downstream right)
- Loading overlay during fetch
- `pnpm build` succeeds (ReactFlow is client-only, page uses client component correctly)

### Risk Assessment
- **Medium**: ReactFlow + SSR -- must ensure `dependency-impact-dag.tsx` is only rendered client-side. The `'use client'` directive handles this, but verify `pnpm build` doesn't fail.
- **Low**: `fitView` timing -- `requestAnimationFrame` delay should suffice (proven in PipelineDAG)
- **Low**: Height calculation `h-[calc(100vh-8rem)]` may need adjustment based on topbar height

### Security Considerations
- No auth on trace endpoints (design decision)
- Node IDs passed to API are from server responses, not raw user input
- `encodeURIComponent` already applied in search component

### Next Steps
- Manual testing of full flow
- Optional enhancements: click node to re-center trace on it, level depth selector, legend
