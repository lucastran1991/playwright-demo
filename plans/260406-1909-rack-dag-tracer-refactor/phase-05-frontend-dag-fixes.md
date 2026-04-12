# Phase 5: Frontend DAG Fixes

## Context

- [Plan overview](plan.md)
- Depends on: Phase 4 (needs /trace/full API)

## Overview

- **Priority:** High
- **Status:** Complete
- **Description:** Switch frontend to use combined `/trace/full` API. Fix duplicate node handling, improve topology-separated layout, and distinguish Load nodes visually.

## Key Insights

Current frontend issues:
1. Makes 2 API calls (deps + impacts) — race conditions, doubled loading state
2. `traceToDAGElements` doesn't deduplicate nodes that appear in both upstream and downstream
3. All upstream nodes (cooling + electrical) share same edge style — no topology separation
4. Load nodes use same `LOCAL_STYLE` as local deps — visually indistinguishable
5. Dagre LR layout gets confused when nodes have conflicting rank constraints from duplicate edges

## Related Code Files

### Modify
- `frontend/src/components/tracer/dependency-impact-dag.tsx` — switch to single API call
- `frontend/src/components/tracer/dag-helpers.tsx` — fix graph building + layout
- `frontend/src/components/tracer/dag-types.ts` — add direction field to TracerNodeData

### Reference
- `frontend/src/components/tracer/dag-node.tsx` — node rendering (may need Load indicator)
- `frontend/src/components/tracer/dag-edge.tsx` — edge rendering

## Implementation Steps

### Step 1: Switch to Single /trace/full API Call

**File:** `frontend/src/components/tracer/dependency-impact-dag.tsx`

Replace the two separate useQuery calls (lines 44-63) with one:

```tsx
const traceQuery = useQuery({
  queryKey: ["trace-full", selectedNodeId, depth],
  queryFn: () =>
    apiFetch<ApiWrapper<TraceResponse>>(
      `/api/trace/full/${selectedNodeId}?levels=${depth}`
    ).then((res) => res.data),
  enabled: !!selectedNodeId,
  staleTime: 60_000,
})
```

Update the useEffect (lines 66-94) to use `traceQuery.data` instead of merging dep + impact:

```tsx
useEffect(() => {
  if (!selectedNodeId) { setNodes([]); setEdges([]); return }
  if (traceQuery.isLoading) return

  const trace = traceQuery.data ?? null
  const { nodes: rawNodes, edges: rawEdges } = traceToDAGElements(trace)
  const { nodes: laidNodes, edges: laidEdges } = layoutDAG(rawNodes, rawEdges)
  const nodesWithClick = laidNodes.map((n) => ({
    ...n,
    data: { ...n.data, onNodeClick: (d: TracerNodeData) => setPopupData(d) },
  }))
  setNodes(nodesWithClick)
  setEdges(laidEdges)
  setTimeout(() => fitView({ padding: 0.2, duration: 400 }), 50)
}, [selectedNodeId, traceQuery.data, traceQuery.isLoading])
```

Update isLoading:
```tsx
const isLoading = !!selectedNodeId && traceQuery.isFetching
```

### Step 2: Simplify traceToDAGElements Signature

**File:** `frontend/src/components/tracer/dag-helpers.tsx`

Change from two parameters to one:

```tsx
export function traceToDAGElements(
  response: TraceResponse | null
): { nodes: Node[]; edges: Edge[] } {
  if (!response) return { nodes: [], edges: [] }

  const source = response.source
  if (!source) return { nodes: [], edges: [] }

  // ... rest of function uses response.upstream, response.downstream, etc.
}
```

### Step 3: Add Load Edge Style

**File:** `frontend/src/components/tracer/dag-helpers.tsx`

Add distinct style for Load edges (currently uses LOCAL_STYLE which is gray dashed):

```tsx
const LOAD_STYLE = { stroke: "#8B5CF6", strokeWidth: 1.5, strokeDasharray: "4 4" }
const LOAD_MARKER = { type: MarkerType.ArrowClosed, color: "#8B5CF6", width: 12, height: 12 }
```

Use purple to match Spatial topology color, distinguishing from gray local deps.

### Step 4: Add Direction to TracerNodeData

**File:** `frontend/src/components/tracer/dag-types.ts`

Add a `direction` field so nodes know their role in the graph:

```tsx
export interface TracerNodeData extends Record<string, unknown> {
  nodeId: string
  name: string
  nodeType: string
  topology: string
  isSource: boolean
  isLocal: boolean
  ring: number
  level: number
  direction: "upstream" | "downstream" | "local" | "load" | "source"
  onNodeClick?: (data: TracerNodeData) => void
}
```

### Step 5: Deduplicate Nodes in traceToDAGElements

**File:** `frontend/src/components/tracer/dag-helpers.tsx`

The nodesMap already deduplicates by node_id (first write wins). But after Phase 1 SQL fix, duplicates should be gone from the API. Still, add a safety check for Load nodes that might also appear in downstream:

```tsx
// Load impacts — skip nodes already in downstream
if (response.load) {
  for (const group of response.load) {
    for (const n of group.nodes) {
      if (!nodesMap.has(n.node_id)) {
        nodesMap.set(n.node_id, makeNode(
          n.node_id, n.name, n.node_type, group.topology,
          1, false, true, 0, undefined, "load"
        ))
      }
      edges.push({
        id: `load-${source.node_id}-${n.node_id}`,
        source: source.node_id,
        target: n.node_id,
        type: "tracerEdge",
        style: LOAD_STYLE,
        markerEnd: LOAD_MARKER,
      })
    }
  }
}
```

### Step 6: Update makeNode to Include Direction

**File:** `frontend/src/components/tracer/dag-helpers.tsx`

Update `makeNode` signature and usage:

```tsx
function makeNode(
  id: string, name: string, nodeType: string, topology: string,
  ring: number, isSource: boolean, isLocal: boolean,
  level?: number, parentId?: string,
  direction: "upstream" | "downstream" | "local" | "load" | "source" = "upstream"
): Node {
  const data: TracerNodeData = {
    nodeId: id, name, nodeType, topology,
    isSource, isLocal, ring, level: level ?? 0, direction,
  }
  // ...
}
```

Update all makeNode calls to pass the correct direction.

### Step 7: Visual Indicator for Load Nodes in dag-node.tsx

**File:** `frontend/src/components/tracer/dag-node.tsx`

Add a subtle visual distinction for Load nodes (e.g., dashed border):

```tsx
// Inside the node component, check data.direction === "load"
const isLoad = data.direction === "load"
// Apply dashed border or different opacity
```

Keep it simple — just a border style change, no major redesign.

### Step 8: Verify DAG Layout

Test with multiple node types:

1. **Rack**: upstream LEFT (cooling + electrical chains), downstream RIGHT (RackPDU, Server)
2. **UPS**: upstream LEFT (BESS, SWGR, Gen, Util), downstream RIGHT (Room PDU, RPP, RackPDU), load RIGHT-BOTTOM (Rack, Row, Zone)
3. **RPP**: upstream LEFT (Room PDU, UPS, ...), downstream RIGHT (RackPDU), load RIGHT-BOTTOM (Rack, Row)

## Todo List

- [x] Switch dependency-impact-dag.tsx to single /trace/full API call
- [x] Simplify traceToDAGElements to single TraceResponse param
- [x] Add LOAD_STYLE + LOAD_MARKER edge styling
- [x] Add direction field to TracerNodeData
- [x] Update makeNode with direction parameter
- [x] Deduplicate Load nodes vs downstream nodes
- [x] Add subtle visual indicator for Load nodes in dag-node.tsx
- [x] Test Rack DAG: upstream LEFT, downstream RIGHT
- [x] Test UPS DAG: both sides populated correctly
- [x] Test RPP DAG: no regressions
- [x] Compile check: `cd frontend && pnpm build`

## Success Criteria

1. Single API call for DAG view (no race conditions)
2. Upstream nodes LEFT, downstream RIGHT, load visually distinct
3. No duplicate nodes in rendered graph
4. Load nodes have purple dashed edges (distinct from gray local)
5. Frontend builds without errors

## Risk Assessment

- **Low risk:** Frontend changes are self-contained, no backend impact
- **Medium risk:** Changing traceToDAGElements signature breaks the existing two-param call — must update all callers simultaneously
- **Mitigation:** Single file change since only dependency-impact-dag.tsx calls traceToDAGElements
