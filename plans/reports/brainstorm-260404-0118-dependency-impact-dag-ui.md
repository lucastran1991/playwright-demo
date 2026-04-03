# Brainstorm: Dependency & Impact DAG UI Page

## Problem Statement
Add a frontend page at `/dashboard/tracer` with a search bar (top center) and ReactFlow DAG visualization showing upstream dependencies (left) -> source node (center) -> downstream impacts (right) for any selected node.

## Decisions

| Decision | Choice |
|----------|--------|
| Library | ReactFlow + Dagre (like PipelineDAG.tsx reference) |
| Data flow | Search node -> call both trace APIs -> render bidirectional DAG |
| Route | `/dashboard/tracer` (under dashboard layout) |
| Search UX | Autocomplete from `/api/blueprints/nodes` API |
| Auth | Public read (no auth needed) |

## Architecture

### Data Flow
```
User types in search bar
  -> debounced GET /api/blueprints/nodes?type=&limit=20 (autocomplete)
  -> user selects node
  -> parallel: GET /api/trace/dependencies/:nodeId?levels=3&include_local=true
              GET /api/trace/impacts/:nodeId?levels=3
  -> merge responses into DAG nodes + edges
  -> Dagre layout (LR direction)
  -> ReactFlow renders
```

### DAG Layout
```
[Upstream L3] -> [Upstream L2] -> [Upstream L1] -> [SOURCE] -> [Downstream L1] -> [Downstream L2]
                                   [Local deps]
```

### Node Styling (reuse PipelineDAG patterns)
- **Source node**: larger, highlighted border (accent color)
- **Electrical nodes**: orange/red theme (Zap icon)
- **Cooling nodes**: blue/cyan theme (Droplets icon)
- **Spatial nodes**: green theme
- **Local deps**: dashed border, positioned below source

### Edge Styling
- **Upstream (dependency)**: blue edges, arrows pointing toward source
- **Downstream (impact)**: red/orange edges, arrows pointing away from source
- **Local**: dashed edges
- Edge labels show topology name

### Search Component
- Positioned top-center, z-10 overlay on ReactFlow
- Input with magnifying glass icon
- Debounced API call (300ms) to `/api/blueprints/nodes?limit=20`
- Dropdown showing: `node_id | name | type` per result
- Click result -> triggers trace
- Use TanStack Query for caching

## New Dependencies
```bash
cd frontend && pnpm add @xyflow/react @dagrejs/dagre
```

## File Map

### New Files
```
frontend/src/
  app/(dashboard)/tracer/
    page.tsx                           -- route page, metadata
  components/tracer/
    dependency-impact-dag.tsx          -- main ReactFlow component (~200 lines)
    dag-node.tsx                       -- custom node component (~80 lines)
    dag-edge.tsx                       -- custom edge component (~40 lines)
    dag-search.tsx                     -- autocomplete search bar (~100 lines)
    dag-helpers.tsx                    -- layout, API->ReactFlow mapping (~100 lines)
    dag-types.ts                       -- TypeScript types (~30 lines)
```

### Modified Files
```
frontend/src/components/dashboard/sidebar-menu.tsx  -- add "Tracer" nav item
```

## Component Breakdown

### `dag-types.ts`
```ts
interface TracerNode { node_id: string; name: string; node_type: string; level: number; }
interface TraceResponse { source: SourceNode; upstream?: TraceLevelGroup[]; local?: TraceLocalGroup[]; downstream?: TraceLevelGroup[]; load?: TraceLocalGroup[]; }
```

### `dag-search.tsx`
- Input + dropdown overlay
- `useQuery` to fetch nodes on input change (debounced)
- Renders list of matches
- onSelect callback triggers parent to load trace

### `dependency-impact-dag.tsx` (main)
- State: selectedNodeId, traceData (dep + impact)
- `useQuery` for dependencies, `useQuery` for impacts (enabled when nodeId set)
- Converts trace responses -> ReactFlow nodes + edges via helpers
- Dagre layout
- ReactFlow render with Background, Controls
- Reuses dag-node, dag-edge components

### `dag-node.tsx`
- Simplified version of PipelineDAG's DomainNode
- Shows: icon (by topology), node_id, name, node_type
- Color-coded by topology (electrical=orange, cooling=blue, spatial=green)
- Source node has special highlight
- Click opens detail popup (optional)

### `dag-edge.tsx`
- Smooth step edges with glow effect (from PipelineDAG)
- Color by direction: blue=dependency, red=impact, dashed=local

### `dag-helpers.tsx`
- `traceToDAGNodes(depResponse, impResponse)` -- merge both responses into unified node list + edge list
- `layoutDAG(nodes, edges)` -- Dagre layout (LR direction)
- Icon/color maps by node type

## API Response -> DAG Mapping

```
Dependencies response:
  upstream[]: each group has level + topology + nodes[]
  local[]: each group has topology + nodes[]

Impacts response:
  downstream[]: each group has level + topology + nodes[]
  load[]: each group has topology + nodes[]

Mapping:
  1. Source node -> center DAG node (id = source.node_id)
  2. Each upstream node -> DAG node, edge: upstream_node -> source
  3. Each downstream node -> DAG node, edge: source -> downstream_node
  4. Each local node -> DAG node, edge: local <-> source (dashed)
  5. Deduplicate nodes (same node_id can appear in both dep + impact)
```

## Risk Assessment
- **Low**: ReactFlow + Dagre well-proven (PipelineDAG.tsx is working reference)
- **Medium**: large trace results (100+ nodes) may crowd the DAG. Mitigation: default levels=2, user can increase.
- **Low**: @xyflow/react types pending for React 19 (reference uses @ts-nocheck). May need same workaround.
- **Low**: TanStack Query already in project, no new patterns needed.

## Success Criteria
- Search any node by typing partial node_id or name
- DAG renders within 500ms of selection
- Upstream (left) -> Source (center) -> Downstream (right) layout clear
- Color coding distinguishes Electrical vs Cooling topology
- Levels visible in left-to-right positioning
- Local dependencies shown below/adjacent to source
- Responsive: works on tablet+ screens
