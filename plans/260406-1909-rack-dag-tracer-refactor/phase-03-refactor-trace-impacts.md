# Phase 3: Refactor TraceImpacts

## Context

- [Plan overview](plan.md)
- Depends on: Phase 1 (SQL fixes + Rack impact rules)
- Can run in parallel with Phase 2

## Overview

- **Priority:** High
- **Status:** Complete
- **Description:** Fix Load resolution so infrastructure nodes (UPS, RPP) correctly find impacted spatial/whitespace nodes (Rack, Row, Zone, Capacity Cell). Use spatial-bridge in reverse.

## Key Insights

Phase 1 fixes `FindSpatialAncestorsOfType` SQL. But the Load resolution logic in `TraceImpacts` (lines 236-291) has structural issues:
- Strategy 1 walks downstream in infra topologies looking for load node TYPES — but load types (Rack, Row) don't appear in electrical/cooling edges as node_types matching the type filter
- Strategy 2 finds spatial ANCESTORS of downstream nodes — correct direction (Rack is spatial parent of RackPDU)
- But Strategy 2 only runs when Strategy 1 finds nothing, and Strategy 1 may find spurious matches preventing Strategy 2

After Phase 1 Rack impact rules are added, Rack also needs its OWN impact tracing to work (downstream = Servers, RackPDU).

## Related Code Files

### Modify
- `backend/internal/service/dependency_tracer.go` — refactor TraceImpacts Load section (lines 198-293)
- `backend/internal/repository/tracer_repository.go` — potentially add FindSpatialDescendantsOfType

### Reference
- `blueprint/Impacts.csv` — impact rules with Downstream/Load relationships
- `backend/internal/service/dependency_tracer_helpers.go` — groupImpactRules

## Implementation Steps

### Step 1: Always Run Both Load Strategies

**File:** `backend/internal/service/dependency_tracer.go`

Current code runs Strategy 2 only `if len(loadGroupMap) == 0`. Fix: always run BOTH strategies and merge results.

Replace lines 236-291 (the entire `if len(allLoadTypes) > 0` block):

```go
if len(allLoadTypes) > 0 {
    loadMaxLevels := 10
    loadGroupMap := make(map[string][]repository.TracedNode)
    seen := make(map[string]bool)

    // Strategy 1: walk downstream in infra topologies to find load-typed nodes
    for slug := range infraSlugSet {
        nodes, err := t.repo.FindDownstreamNodes(node.ID, slug, loadMaxLevels)
        if err != nil {
            continue
        }
        for _, n := range nodes {
            if allLoadTypes[n.NodeType] && !seen[n.NodeID] {
                seen[n.NodeID] = true
                loadTopo := t.lookupTopology(n.NodeType)
                loadGroupMap[loadTopo] = append(loadGroupMap[loadTopo], n)
            }
        }
    }

    // Strategy 2: find spatial ancestors/descendants of downstream nodes
    // that match load types. Always run — catches cases Strategy 1 misses.
    var allDownstreamDBIDs []uint
    for slug := range infraSlugSet {
        deepNodes, err := t.repo.FindDownstreamNodes(node.ID, slug, loadMaxLevels)
        if err != nil {
            continue
        }
        for _, n := range deepNodes {
            allDownstreamDBIDs = append(allDownstreamDBIDs, n.ID)
        }
    }

    loadTypeSlice := make([]string, 0, len(allLoadTypes))
    for nt := range allLoadTypes {
        loadTypeSlice = append(loadTypeSlice, nt)
    }

    if len(allDownstreamDBIDs) > 0 && len(loadTypeSlice) > 0 {
        // 2a: spatial ancestors (Rack is parent of RackPDU in spatial)
        spatialLoads, err := t.repo.FindSpatialAncestorsOfType(allDownstreamDBIDs, loadTypeSlice)
        if err == nil {
            for _, n := range spatialLoads {
                if !seen[n.NodeID] {
                    seen[n.NodeID] = true
                    loadTopo := t.lookupTopology(n.NodeType)
                    loadGroupMap[loadTopo] = append(loadGroupMap[loadTopo], n)
                }
            }
        }

        // 2b: spatial descendants (Zone contains Rack as child in spatial)
        spatialDescs, err := t.repo.FindSpatialDescendantsOfType(allDownstreamDBIDs, loadTypeSlice)
        if err == nil {
            for _, n := range spatialDescs {
                if !seen[n.NodeID] {
                    seen[n.NodeID] = true
                    loadTopo := t.lookupTopology(n.NodeType)
                    loadGroupMap[loadTopo] = append(loadGroupMap[loadTopo], n)
                }
            }
        }
    }

    for topo, nodes := range loadGroupMap {
        resp.Load = append(resp.Load, TraceLocalGroup{Topology: topo, Nodes: nodes})
    }
}
```

### Step 2: Add FindSpatialDescendantsOfType Repository Method

**File:** `backend/internal/repository/tracer_repository.go`

Mirror of FindSpatialAncestorsOfType but walks DOWN instead of UP:

```go
// FindSpatialDescendantsOfType walks down spatial edges from the given node IDs
// and returns distinct descendant nodes whose node_type is in the given set.
func (r *TracerRepository) FindSpatialDescendantsOfType(nodeDBIDs []uint, nodeTypes []string) ([]TracedNode, error) {
    if len(nodeDBIDs) == 0 || len(nodeTypes) == 0 {
        return nil, nil
    }
    var nodes []TracedNode
    err := r.db.Raw(`
        WITH RECURSIVE descendants AS (
            SELECT bn.id, bn.node_id, bn.name, bn.node_type, 1 as level
            FROM blueprint_edges be
            JOIN blueprint_nodes bn ON bn.id = be.to_node_id
            JOIN blueprint_types bt ON bt.id = be.blueprint_type_id
            WHERE be.from_node_id IN ?
              AND bt.slug = 'spatial-topology'

            UNION ALL

            SELECT bn.id, bn.node_id, bn.name, bn.node_type, d.level + 1
            FROM descendants d
            JOIN blueprint_edges be ON be.from_node_id = d.id
            JOIN blueprint_nodes bn ON bn.id = be.to_node_id
            JOIN blueprint_types bt ON bt.id = be.blueprint_type_id
            WHERE bt.slug = 'spatial-topology' AND d.level < 5
        )
        SELECT DISTINCT ON (id) id, node_id, name, node_type, level
        FROM descendants
        WHERE node_type IN ?
        ORDER BY id, level
    `, nodeDBIDs, nodeTypes).Scan(&nodes).Error
    return nodes, err
}
```

### Step 3: Handle Rack's Own Impact Tracing

With Phase 1 adding Rack impact rules (`Rack → Rack PDU (Downstream L1), Rack → Server (Downstream L1)`), the Rack impact trace needs to find these children.

Rack has spatial edges to RackPDU and Servers. But `groupImpactRules` will map:
- Rack PDU → topology "Electrical System" → downstream["Electrical System"]
- Server → topology ??? (Server might not be in CapacityNodeTypes)

Check if "Server" exists in `Capacity Nodes.csv`. If not, add it or handle the fallback in `lookupTopology`.

The downstream loop will call `FindDownstreamNodes(Rack.ID, "electrical-system", maxLevels)` — this WILL find RackPDU because Rack→RackPDU edge exists in electrical. 

For Servers: they're only in spatial topology. The downstream loop won't find them via electrical slug. Two options:
- Add Server to Capacity Nodes.csv with topology "Spatial Topology"
- OR handle spatial downstream in the impact code similar to the spatial-bridge in Phase 2

**Recommended:** Add Server to Capacity Nodes.csv:
```csv
Server,Spatial Topology,False,False
```

Then the downstream loop will also try spatial-topology slug, finding Rack→Server edges.

### Step 4: Verify UPS Impact Load Resolution

```bash
curl -s "http://localhost:8889/api/trace/impacts/UPS-01?levels=3" | python3 -c "
import sys, json
d = json.load(sys.stdin)['data']
print('downstream:', sum(len(g['nodes']) for g in d.get('downstream', [])), 'nodes')
print('load groups:', len(d.get('load', [])))
for g in d.get('load', []):
    types = list(set(n['node_type'] for n in g['nodes']))
    print(f'  {g[\"topology\"]}: {types} ({len(g[\"nodes\"])} nodes)')
"
```

Expected: load includes Rack nodes (Spatial Topology) found as spatial parents of RackPDU downstream nodes.

### Step 5: Verify Rack Impact Tracing

```bash
curl -s "http://localhost:8889/api/trace/impacts/RACK-R1-Z1-R1-01?levels=3" | python3 -c "
import sys, json
d = json.load(sys.stdin)['data']
for g in d.get('downstream', []):
    types = [n['node_type'] for n in g['nodes']]
    print(f'  L{g[\"level\"]}: {types}')
"
```

Expected: downstream includes Rack PDU (L1) and possibly Server (L1).

## Todo List

- [x] Refactor Load resolution: always run both strategies
- [x] Add FindSpatialDescendantsOfType to tracer_repository.go
- [x] Add Server to Capacity Nodes.csv (if not present)
- [x] Verify UPS impacts include Load nodes (Rack, Row, Zone)
- [x] Verify Rack impacts include downstream (RackPDU, Server)
- [x] Verify RPP impacts still work correctly
- [x] Run unit tests

## Success Criteria

1. UPS impacts: `load` contains Rack, Row, Zone nodes from Spatial Topology
2. Rack impacts: `downstream` contains Rack PDU, Server nodes
3. RPP impacts: no regressions (downstream + load unchanged)
4. All tests pass

## Risk Assessment

- **Medium risk:** Always running both strategies increases query count. But loadMaxLevels=10 is already expensive — one extra spatial walk is marginal.
- **Low risk:** Adding FindSpatialDescendantsOfType is a clean addition, mirrors existing pattern.
