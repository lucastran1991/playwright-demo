# Phase 4: Frontend — DAG Capacity Badges

## Context Links
- Phase 3 (dependency): `phase-03-capacity-api.md` (provides API + enriched TraceResponse)
- DAG types: `frontend/src/components/tracer/dag-types.ts`
- DAG node component: `frontend/src/components/tracer/dag-node.tsx`
- DAG helpers: `frontend/src/components/tracer/dag-helpers.tsx`
- Detail popup: `frontend/src/components/tracer/dag-detail-popup.tsx`
- Right panel: `frontend/src/components/tracer/dag-right-panel.tsx`
- Main DAG: `frontend/src/components/tracer/dependency-impact-dag.tsx`
- API client: `frontend/src/lib/api-client.ts`

## Overview
- **Priority**: P2
- **Status**: pending
- **Description**: Extend the existing Tracer DAG to display capacity utilization badges on each node, color-coded by utilization level. Add a capacity detail section to the node detail popup.

## Key Insights
- Phase 3 enriches `TraceResponse` with a `capacity` map: `{ [nodeID]: { design_capacity: 230, rated_capacity: 288, allocated_load: 162, ... } }` — frontend just needs to read this, no extra API call
- Utilization color thresholds: green (<60%), yellow (60-80%), red (>80%) — standard industry convention
- Badge should be compact: `162/288 kW (56%)` or just `56%` with color
- Not all nodes have capacity data — badge only shown when data exists
- Existing `TracerNodeData` extends `Record<string, unknown>` so adding fields is backward-compatible
- IMPORTANT: Must read Next.js docs in `node_modules/next/dist/docs/` per AGENTS.md before writing code

## Requirements

### Functional
- Show utilization badge on each DAG node that has capacity data
- Color code: green (#22C55E) <60%, yellow (#EAB308) 60-80%, red (#EF4444) >80%
- Badge text: `{utilization_pct}%` — compact enough for small nodes
- Node detail popup: show full capacity breakdown (design_capacity, rated_capacity, allocated_load, available_capacity, utilization_pct)
- Tooltip on badge hover: `{allocated_load}/{rated_capacity} kW`

### Non-functional
- No extra API calls — use capacity data from enriched TraceResponse
- Graceful degradation: nodes without capacity data show no badge
- Mobile-friendly: badge must be readable at small viewport

## Architecture

```
TraceResponse.capacity (from API)
       |
  dag-helpers.tsx: traceToDAGElements passes capacity data into TracerNodeData
       |
  dag-node.tsx: renders utilization badge if capacity data present
       |
  dag-detail-popup.tsx: shows full capacity breakdown
```

## Related Code Files

### Files to MODIFY
| File | Change | Est. Lines |
|------|--------|-----------|
| `frontend/src/components/tracer/dag-types.ts` | Add CapacityData interface + capacity field to TracerNodeData + TraceResponse | +15 |
| `frontend/src/components/tracer/dag-helpers.tsx` | Pass capacity data when building nodes in traceToDAGElements | +10 |
| `frontend/src/components/tracer/dag-node.tsx` | Add utilization badge with color coding | +25 |
| `frontend/src/components/tracer/dag-detail-popup.tsx` | Add capacity section to detail view | +40 |

No new files needed — all changes extend existing components.

## Implementation Steps

### Step 1: Extend types
File: `frontend/src/components/tracer/dag-types.ts`

```typescript
// Add to existing file:

export interface CapacityData {
  design_capacity?: number
  rated_capacity?: number
  allocated_load?: number
  available_capacity?: number
  utilization_pct?: number
  // Additional vars (liquid/air split, rack_count, etc.)
  [key: string]: number | undefined
}

// Extend TraceResponse — add capacity field
export interface TraceResponse {
  source: { node_id: string; name: string; node_type: string; topology?: string }
  upstream?: TraceLevelGroup[]
  local?: TraceLocalGroup[]
  downstream?: TraceLevelGroup[]
  load?: TraceLocalGroup[]
  capacity?: Record<string, CapacityData>  // nodeID -> metrics
}

// Extend TracerNodeData — add optional capacity
export interface TracerNodeData extends Record<string, unknown> {
  // ... existing fields ...
  capacity?: CapacityData  // NEW
}
```

### Step 2: Pass capacity data in dag-helpers
File: `frontend/src/components/tracer/dag-helpers.tsx`

Modify `traceToDAGElements` to accept and pass through capacity data:

```typescript
export function traceToDAGElements(
  response: TraceResponse | null
): { nodes: Node[]; edges: Edge[] } {
  // ... existing logic ...
  
  // After building all nodes, enrich with capacity data
  const capacityMap = response.capacity ?? {}
  
  // In makeNode calls or after node creation:
  // For each node in nodesMap, attach capacity data if available
  for (const [nodeId, node] of nodesMap) {
    if (capacityMap[nodeId]) {
      (node.data as TracerNodeData).capacity = capacityMap[nodeId]
    }
  }
  
  // ... rest of existing logic ...
}
```

This is a minimal change — just a loop after the existing node-building logic.

### Step 3: Add utilization badge to dag-node
File: `frontend/src/components/tracer/dag-node.tsx`

Add a badge below the node name, only when capacity data exists:

```tsx
// Utilization color helper
function getUtilColor(pct: number): string {
  if (pct > 80) return "#EF4444"   // red
  if (pct >= 60) return "#EAB308"  // yellow
  return "#22C55E"                  // green
}

// Inside TracerNode component, after the name paragraph:
{data.capacity?.utilization_pct != null && (
  <div className="flex items-center gap-1 mt-0.5">
    <div
      className="h-1.5 flex-1 rounded-full bg-muted overflow-hidden"
      title={`${data.capacity.allocated_load ?? 0}/${data.capacity.rated_capacity ?? 0} kW`}
    >
      <div
        className="h-full rounded-full transition-all"
        style={{
          width: `${Math.min(data.capacity.utilization_pct, 100)}%`,
          backgroundColor: getUtilColor(data.capacity.utilization_pct),
        }}
      />
    </div>
    <span
      className="text-[9px] font-bold tabular-nums shrink-0"
      style={{ color: getUtilColor(data.capacity.utilization_pct) }}
    >
      {Math.round(data.capacity.utilization_pct)}%
    </span>
  </div>
)}
```

This renders:
- A thin progress bar showing utilization level
- A percentage label in the matching color
- Hover tooltip with absolute values (allocated/rated kW)

### Step 4: Add capacity section to detail popup
File: `frontend/src/components/tracer/dag-detail-popup.tsx`

Add capacity detail section after the existing "Status indicators" section:

```tsx
// After the StatusPill grid, before the node ID copyable section:

{data.capacity && (
  <div className="space-y-2">
    <p className="text-[10px] font-semibold uppercase tracking-widest text-muted-foreground">
      Capacity
    </p>
    
    {/* Utilization bar — full width */}
    {data.capacity.utilization_pct != null && (
      <div className="space-y-1">
        <div className="flex justify-between text-xs">
          <span className="text-muted-foreground">Utilization</span>
          <span className="font-bold" style={{ color: getUtilColor(data.capacity.utilization_pct) }}>
            {Math.round(data.capacity.utilization_pct)}%
          </span>
        </div>
        <div className="h-2 rounded-full bg-muted overflow-hidden">
          <div
            className="h-full rounded-full"
            style={{
              width: `${Math.min(data.capacity.utilization_pct, 100)}%`,
              backgroundColor: getUtilColor(data.capacity.utilization_pct),
            }}
          />
        </div>
      </div>
    )}
    
    {/* Metric cards */}
    <div className="grid grid-cols-2 gap-2">
      {data.capacity.rated_capacity != null && (
        <DetailCard label="Rated Capacity" value={`${data.capacity.rated_capacity} kW`} accent={color} />
      )}
      {data.capacity.allocated_load != null && (
        <DetailCard label="Allocated Load" value={`${data.capacity.allocated_load} kW`} accent={color} />
      )}
      {data.capacity.available_capacity != null && (
        <DetailCard label="Available" value={`${data.capacity.available_capacity} kW`} accent="#22C55E" />
      )}
      {data.capacity.design_capacity != null && (
        <DetailCard label="Design Capacity" value={`${data.capacity.design_capacity} kW`} accent={color} />
      )}
    </div>
  </div>
)}
```

Extract `getUtilColor` to a shared location (or define in both files — 3 lines, acceptable duplication for component isolation).

## Todo List
- [ ] Add `CapacityData` interface and `capacity` field to `TraceResponse` in dag-types.ts
- [ ] Add `capacity?: CapacityData` to `TracerNodeData` in dag-types.ts
- [ ] Enrich nodes with capacity data in `traceToDAGElements` in dag-helpers.tsx
- [ ] Add utilization progress bar + percentage badge to dag-node.tsx
- [ ] Add capacity detail section to dag-detail-popup.tsx
- [ ] Add `getUtilColor` helper (in dag-node.tsx and dag-detail-popup.tsx)
- [ ] Verify `pnpm build` compiles (or `pnpm dev` renders)
- [ ] Visual check: nodes show colored utilization bars
- [ ] Visual check: detail popup shows capacity breakdown

## Success Criteria
- Nodes with capacity data show utilization badge (progress bar + %)
- Badge color: green <60%, yellow 60-80%, red >80%
- Nodes without capacity data show no badge (no visual regression)
- Detail popup shows full capacity metrics when clicking a node
- Hover tooltip shows `allocated/rated kW`
- Mobile viewport: badge readable and not clipped
- No TypeScript compilation errors

## Risk Assessment
| Risk | Severity | Mitigation |
|------|----------|-----------|
| TracerNodeData shape change breaks ReactFlow | Low | Extends `Record<string, unknown>`, optional field |
| Capacity data not present in response | Low | All renders guarded with `data.capacity?.` checks |
| Node becomes too tall with badge | Low | Badge is 1 line (~16px), minimal height increase |
| Performance with 100+ nodes rendering badges | Low | Pure CSS progress bars, no JS animation |

## Security Considerations
- No user input — all data from server API response
- No new API calls — piggybacks on existing trace query
