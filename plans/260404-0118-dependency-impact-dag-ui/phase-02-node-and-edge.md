## Phase 2: Custom Node + Edge Components

### Context Links
- [PipelineDAG.tsx](../../PipelineDAG.tsx) -- DomainNode (line 135-200), FlowEdge (line 206-226)
- [dag-types.ts](../../frontend/src/components/tracer/dag-types.ts) -- from Phase 1
- [dag-helpers.tsx](../../frontend/src/components/tracer/dag-helpers.tsx) -- topology colors/icons from Phase 1

### Overview
- **Priority**: P1
- **Status**: pending
- **Description**: Create custom ReactFlow node and edge components for the tracer DAG. Simplified versions of PipelineDAG's DomainNode and FlowEdge.

### Key Insights
- DomainNode in PipelineDAG shows KPIs, trends, hover tooltips -- tracer node is simpler (just identity info)
- FlowEdge glow effect from PipelineDAG is a nice touch -- keep it
- Source node needs visual distinction (thicker border, slight scale)
- Handles: Left=target, Right=source (same as PipelineDAG for LR layout)

### Requirements

**Functional**
- `dag-node.tsx`: renders node_id, name, node_type with topology icon + color coding. Source node highlighted.
- `dag-edge.tsx`: smooth step edge with glow, colored by direction (blue/red/dashed)

**Non-functional**
- Each file under 100 lines
- `'use client'` directive on both (ReactFlow requires client components)
- Use `@ts-nocheck` at top (React 19 + @xyflow/react compat, matching PipelineDAG.tsx)

### Related Code Files

**Create:**
- `frontend/src/components/tracer/dag-node.tsx` (~90 lines)
- `frontend/src/components/tracer/dag-edge.tsx` (~50 lines)

**Read (imports from Phase 1):**
- `frontend/src/components/tracer/dag-types.ts`
- `frontend/src/components/tracer/dag-helpers.tsx`

### Implementation Steps

#### Step 1: Create `dag-node.tsx`

Structure (follow DomainNode pattern from PipelineDAG.tsx):

```tsx
// @ts-nocheck
'use client'

import { Handle, Position } from '@xyflow/react'
import type { TracerNodeData } from './dag-types'
import { TOPOLOGY_COLORS, TOPOLOGY_ICONS } from './dag-helpers'
```

Component receives `{ data }: { data: TracerNodeData }`.

**Layout:**
```
+----------------------------------+
| [Icon] node_type                 |
| node_id (bold, truncated)        |
| name (muted, smaller)            |
+----------------------------------+
```

**Styling logic:**
1. Get color set from `TOPOLOGY_COLORS[data.topology]` with fallback
2. If `data.isSource` -- add ring/glow effect, slightly larger (`min-w-[220px]` vs `min-w-[180px]`)
3. If `data.isLocal` -- dashed border
4. Base classes: `rounded-xl border-2 px-3 py-2.5 transition-all duration-200 hover:scale-[1.03] cursor-default`
5. Background: use `theme-card` class (matches PipelineDAG)

**Handles:**
- `<Handle type="target" position={Position.Left} />` -- same styling as PipelineDAG
- `<Handle type="source" position={Position.Right} />` -- accent colored for source node

**Source node highlight:**
```tsx
const ringClass = data.isSource
  ? 'ring-2 ring-yellow-400/50 shadow-lg shadow-yellow-400/10'
  : ''
```

#### Step 2: Create `dag-edge.tsx`

Follow FlowEdge from PipelineDAG.tsx (lines 206-226) exactly, but simplified (no flowValue label needed):

```tsx
// @ts-nocheck
'use client'

import { BaseEdge, getSmoothStepPath, EdgeLabelRenderer } from '@xyflow/react'
```

Component props: standard ReactFlow edge props.

**Rendering:**
1. Compute path with `getSmoothStepPath({ sourceX, sourceY, targetX, targetY, sourcePosition, targetPosition, borderRadius: 16 })`
2. Render glow layer: same path with +4 strokeWidth, 0.08 opacity, blur(4px) filter
3. Render main edge with `style` and `markerEnd` from edge data
4. Optional label (topology name) using `EdgeLabelRenderer` -- only if `data?.label` exists

**Label styling** (from PipelineDAG):
```tsx
className="text-[7px] font-medium text-muted-foreground bg-background/80 px-1.5 py-0.5 rounded border border-border/30 pointer-events-none whitespace-nowrap"
```

### Todo List
- [ ] Create `dag-node.tsx` with topology-based styling + source highlight
- [ ] Create `dag-edge.tsx` with glow effect + direction coloring
- [ ] Verify both compile: `cd frontend && pnpm tsc --noEmit`

### Success Criteria
- Node displays node_id, name, type with correct topology color
- Source node visually distinct from dependency/impact nodes
- Edge glow effect renders smoothly
- Both components export correctly for ReactFlow `nodeTypes`/`edgeTypes` maps

### Risk Assessment
- **Low**: Tailwind classes may need verification against project's Tailwind 4 config
- **Low**: `theme-card` class assumed from PipelineDAG -- verify it exists in project CSS or replace with equivalent

### Next Steps
- Phase 3 builds search component
- Phase 4 wires nodes/edges into main DAG component via `nodeTypes`/`edgeTypes` useMemo
