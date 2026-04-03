## Phase 1: Install Dependencies + Types + Helpers

### Context Links
- [PipelineDAG.tsx](../../PipelineDAG.tsx) -- Dagre layout pattern (line 232-242)
- [api-client.ts](../../frontend/src/lib/api-client.ts) -- existing `apiFetch` wrapper
- [Brainstorm](../reports/brainstorm-260404-0118-dependency-impact-dag-ui.md)

### Overview
- **Priority**: P1 (blocking all other phases)
- **Status**: pending
- **Description**: Install @xyflow/react + @dagrejs/dagre, create type definitions and helper functions for Dagre layout + API response mapping.

### Key Insights
- PipelineDAG.tsx uses `Dagre.graphlib.Graph` with `rankdir: 'LR'` -- same pattern needed here
- `apiFetch` from `@/lib/api-client` already handles error throwing and JSON parsing
- API responses have `upstream[]`, `local[]`, `downstream[]`, `load[]` groups with level + topology + nodes

### Requirements

**Functional**
- TypeScript types covering all API response shapes and internal DAG types
- `traceToDAGElements(depResponse, impactResponse)` -- merge both API responses into ReactFlow nodes + edges
- `layoutDAG(nodes, edges)` -- Dagre LR layout identical to PipelineDAG.tsx pattern
- Color/icon maps for node topologies

**Non-functional**
- Each file under 200 lines
- Reusable helpers, no coupling to React components

### Architecture

```
dag-types.ts      -- pure types, no runtime code
dag-helpers.tsx    -- layout + mapping functions + icon/color constants
```

### Related Code Files

**Create:**
- `frontend/src/components/tracer/dag-types.ts`
- `frontend/src/components/tracer/dag-helpers.tsx`

### Implementation Steps

#### Step 1: Install dependencies
```bash
cd frontend && pnpm add @xyflow/react @dagrejs/dagre
```

Also install Dagre types if available:
```bash
pnpm add -D @types/dagre  # only if @dagrejs/dagre doesn't ship types
```

#### Step 2: Create `dag-types.ts`

Define these types:

```ts
// API response types (match backend JSON)
export interface TraceNode {
  node_id: string
  name: string
  node_type: string    // "electrical_panel", "cooling_unit", etc.
  topology?: string    // "electrical", "cooling", "spatial"
}

export interface TraceSourceNode extends TraceNode {
  description?: string
}

export interface TraceLevelGroup {
  level: number
  topology: string
  nodes: TraceNode[]
}

export interface TraceLocalGroup {
  topology: string
  nodes: TraceNode[]
}

export interface TraceResponse {
  source: TraceSourceNode
  upstream?: TraceLevelGroup[]
  local?: TraceLocalGroup[]
  downstream?: TraceLevelGroup[]
  load?: TraceLocalGroup[]
}

// Search API response
export interface SearchNode {
  node_id: string
  name: string
  node_type: string
}

export interface SearchResponse {
  data: SearchNode[]
}

// Internal DAG types (for ReactFlow)
export type TracerNodeData = {
  nodeId: string
  name: string
  nodeType: string
  topology: string
  level: number           // 0 = source, positive = downstream, negative = upstream
  isSource: boolean
  isLocal: boolean
}

export type EdgeDirection = "dependency" | "impact" | "local"
```

#### Step 3: Create `dag-helpers.tsx`

**3a. Topology color + icon maps**

Follow PipelineDAG.tsx icon pattern (lucide-react). Map topology to colors:
```ts
export const TOPOLOGY_COLORS: Record<string, { bg: string; border: string; text: string }> = {
  electrical: { bg: "bg-orange-500/10", border: "border-orange-500/40", text: "text-orange-400" },
  cooling:    { bg: "bg-blue-500/10",   border: "border-blue-500/40",   text: "text-blue-400" },
  spatial:    { bg: "bg-green-500/10",   border: "border-green-500/40",  text: "text-green-400" },
}

// Default fallback for unknown topologies
const DEFAULT_COLOR = { bg: "bg-gray-500/10", border: "border-gray-500/40", text: "text-gray-400" }
```

Icons using lucide-react (Zap for electrical, Droplets for cooling, etc.):
```ts
export const TOPOLOGY_ICONS: Record<string, React.ReactNode> = {
  electrical: <Zap size={16} />,
  cooling: <Droplets size={16} />,
  spatial: <Building size={16} />,  // or similar
}
```

**3b. `traceToDAGElements` function**

Signature:
```ts
export function traceToDAGElements(
  depResponse: TraceResponse | null,
  impactResponse: TraceResponse | null
): { nodes: Node<TracerNodeData>[]; edges: Edge[] }
```

Logic:
1. Create source node from `depResponse.source` (or `impactResponse.source`)
2. Iterate `depResponse.upstream[]` -- for each group, each node becomes a ReactFlow node; edge: `upstreamNode -> sourceNode` (blue)
3. Iterate `depResponse.local[]` -- each node becomes a ReactFlow node; edge: dashed, bidirectional feel
4. Iterate `impactResponse.downstream[]` -- each node becomes a ReactFlow node; edge: `sourceNode -> downstreamNode` (red)
5. Iterate `impactResponse.load[]` -- similar to local but for impacts
6. Deduplicate by `node_id` (same node may appear in both responses)
7. Return `{ nodes, edges }` with positions at `{x:0, y:0}` (Dagre will set real positions)

Edge styling constants (reuse PipelineDAG pattern):
```ts
const DEP_EDGE_STYLE = { stroke: "#3B82F6", strokeWidth: 2 }   // blue
const IMPACT_EDGE_STYLE = { stroke: "#EF4444", strokeWidth: 2 }  // red
const LOCAL_EDGE_STYLE = { stroke: "#6B7280", strokeWidth: 1.5, strokeDasharray: "6 3" }
```

Arrow markers:
```ts
const DEP_MARKER = { type: MarkerType.ArrowClosed, color: "#3B82F6", width: 14, height: 14 }
const IMPACT_MARKER = { type: MarkerType.ArrowClosed, color: "#EF4444", width: 14, height: 14 }
```

**3c. `layoutDAG` function**

Copy pattern from PipelineDAG.tsx line 232-242:
```ts
export function layoutDAG(nodes: Node[], edges: Edge[], nodeWidth = 200, nodeHeight = 80) {
  const g = new Dagre.graphlib.Graph().setDefaultEdgeLabel(() => ({}))
  g.setGraph({ rankdir: "LR", nodesep: 40, ranksep: 100, marginx: 30, marginy: 30 })
  nodes.forEach((n) => g.setNode(n.id, { width: nodeWidth, height: nodeHeight }))
  edges.forEach((e) => g.setEdge(e.source, e.target))
  Dagre.layout(g)
  return {
    nodes: nodes.map((n) => ({
      ...n,
      position: {
        x: g.node(n.id).x - nodeWidth / 2,
        y: g.node(n.id).y - nodeHeight / 2,
      },
    })),
    edges,
  }
}
```

### Todo List
- [ ] Run `pnpm add @xyflow/react @dagrejs/dagre` in frontend/
- [ ] Create `dag-types.ts` with all API + internal types
- [ ] Create `dag-helpers.tsx` with topology maps, `traceToDAGElements`, `layoutDAG`
- [ ] Verify types compile: `cd frontend && pnpm tsc --noEmit`

### Success Criteria
- `pnpm tsc --noEmit` passes with new files
- Types match actual API response shapes
- `layoutDAG` produces valid positioned nodes

### Risk Assessment
- **Low**: @xyflow/react may need `@ts-nocheck` for React 19 compatibility (same as PipelineDAG.tsx)
- **Low**: @dagrejs/dagre types may need `@types/dagre` dev dependency

### Next Steps
- Phase 2 imports types + helpers to build custom node/edge components
