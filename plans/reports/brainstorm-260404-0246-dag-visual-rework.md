# Brainstorm: DAG Tracer Visual Rework

## Problem Statement
Current DAG tracer shows nodes with basic topology colors but lacks: proper dark-theme palette, level badges, grouped local dependencies, distinct upstream/downstream edge colors, and topology info on source node.

## Decisions

| Decision | Choice |
|----------|--------|
| Color scheme | Custom dark-theme optimized palette |
| Local dependencies | Grouped sub-graph (ReactFlow Group node) |
| Level display | Badge on node corner ("L1", "L2") |
| Edge distinction | Color-only: upstream=one color, downstream=another |
| Backend change | Add topology to source node in trace response |

## Design Spec

### Dark-Theme Color Palette
```
Electrical nodes: border #F97316 (orange-500), bg #F97316/10
Cooling nodes:    border #06B6D4 (cyan-500),   bg #06B6D4/10
Spatial nodes:    border #8B5CF6 (violet-500),  bg #8B5CF6/10
Whitespace nodes: border #10B981 (emerald-500), bg #10B981/10
Source node:      ring #FBBF24 (amber-400), bg amber-400/5

Upstream edges:   stroke #06B6D4 (cyan-500)    -- deps flow INTO source
Downstream edges: stroke #F97316 (orange-500)  -- impacts flow OUT of source
Local edges:      stroke #6B7280 (gray-500), dashed
```

### Node Card Layout
```
+--------------------------------------+
| [Icon] NODE_TYPE            [L2] |  <- level badge top-right
| NODE_ID (bold, topology color)       |
| name (muted)                         |
+--------------------------------------+
```
- Source node: amber ring + "SOURCE" badge, slightly larger
- Level badge: small rounded pill, e.g. "L1" in muted style
- Topology icon + color on border

### Grouped Local Dependencies
Use ReactFlow `type: 'group'` parent node to create a bordered rectangle containing source + its local deps.

```
+-- Local Dependencies ----------------+
|  [RDHx node]   [SOURCE node]         |
|  [DTC node]                          |
+--------------------------------------+
```

Group node:
- Dashed border, gray, labeled "Local ({topology})"
- Source node + local dep nodes are children (`parentId` set to group id)
- Group auto-sizes to contain children
- Only shown when local deps exist

### Level Badges
- Small pill in top-right corner of node card
- Format: "L1", "L2", "L3"
- Source = no level badge (or "SRC")
- Local deps = no level badge
- Muted color, doesn't dominate visual hierarchy

### Edge Colors
- **Upstream (dependency)**: cyan (#06B6D4) -- flows INTO source from left
- **Downstream (impact)**: orange (#F97316) -- flows OUT of source to right
- **Local**: gray dashed (#6B7280)
- All non-local edges: animated + arrow markers

### Backend Change: Add topology to source
In `dependency_tracer.go`, after looking up the source node, also look up its topology from `capacity_node_types`:

```go
// In TraceDependencies and TraceImpacts
resp.Source = SourceNode{
    NodeID:   node.NodeID,
    Name:     node.Name,
    NodeType: node.NodeType,
    Topology: t.lookupTopology(node.NodeType), // NEW
}
```

Update `SourceNode` struct:
```go
type SourceNode struct {
    NodeID   string `json:"node_id"`
    Name     string `json:"name"`
    NodeType string `json:"node_type"`
    Topology string `json:"topology"` // NEW
}
```

Frontend `TraceResponse.source` type also needs `topology?: string`.

## Files to Change

### Backend (2 files)
- `backend/internal/service/dependency_tracer.go` -- add topology to SourceNode struct + both trace methods
- No new files needed

### Frontend (4 files)
- `frontend/src/components/tracer/dag-types.ts` -- add topology to source type
- `frontend/src/components/tracer/dag-helpers.tsx` -- new color palette, group node creation for local deps, level badge data
- `frontend/src/components/tracer/dag-node.tsx` -- level badge UI, updated colors
- `frontend/src/components/tracer/dag-edge.tsx` -- (minimal change, colors come from helpers)

## Implementation Considerations

### ReactFlow Group Nodes
- Group node = a node with `type: 'group'` and no custom component (uses default)
- Child nodes set `parentId: groupId` and `extent: 'parent'`
- Child positions are RELATIVE to group
- Group must be added to nodes BEFORE children
- Need to calculate group dimensions based on children count

### Risk
- **Medium**: ReactFlow group nodes with Dagre layout can conflict. Dagre doesn't know about parent-child positioning. May need to layout group children separately from the main graph.
- **Mitigation**: Layout the main graph with Dagre (source + upstream + downstream), then position local deps manually around the source node within the group.

## Success Criteria
- Dark theme colors look good on both light and dark mode
- Level badges visible but not dominant
- Local deps clearly grouped around source
- Upstream (cyan) vs downstream (orange) instantly distinguishable
- Source node prominently highlighted (amber ring)
