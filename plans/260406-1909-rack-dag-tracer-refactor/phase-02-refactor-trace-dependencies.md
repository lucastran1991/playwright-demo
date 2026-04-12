# Phase 2: Refactor TraceDependencies

## Context

- [Plan overview](plan.md)
- Depends on: Phase 1 (SQL fixes must land first)

## Overview

- **Priority:** High
- **Status:** Complete
- **Description:** Replace the fragile fallback mechanism with explicit spatial-bridge traversal for cross-topology dependency resolution.

## Key Insights

Current fallback (lines 136-152 in `dependency_tracer.go`) works for Rack because Rack→RackPDU edge exists in electrical. But it:
- Produces `parent_node_id` referencing nodes NOT in the result set (RACKPDU filtered out)
- Doesn't work for Whitespace nodes (no electrical/cooling edges at all)
- Uses hop-count levels that don't match CSV UpstreamLevel values

The spatial-bridge approach: detect source's topology, walk spatial hierarchy to find bridge nodes in target topology, then walk upstream from those bridge nodes.

## Related Code Files

### Modify
- `backend/internal/service/dependency_tracer.go` — refactor TraceDependencies (lines 102-176)
- `backend/internal/repository/tracer_repository.go` — add new spatial bridge query

### Reference
- `blueprint/Dependencies.csv` — UpstreamLevel values for level normalization
- `blueprint/Capacity Nodes.csv` — node type → topology mapping
- `backend/internal/service/dependency_tracer_helpers.go` — groupDepRules logic

## Architecture

### Current Flow (Broken for Whitespace)
```
Source (any topology)
  → groupDepRules → upstream[topo] = {allowed types}
  → FindUpstreamNodes(source, topo_slug, maxLevels)
  → if empty: fallback via FindDownstreamNodes then FindUpstreamNodes from children
  → filterByTypes
```

### New Flow (Spatial-Bridge)
```
Source (any topology)
  → groupDepRules → upstream[topo] = {allowed types}
  → Is source in target topology? (check if source has edges in topo_slug)
    YES → FindUpstreamNodes(source, topo_slug, maxLevels) [existing path]
    NO  → findBridgeNodes(source, topo_slug) → returns infra nodes reachable via spatial
        → FindUpstreamNodes(bridge_node, topo_slug, maxLevels)
        → shift levels to account for bridge distance
  → filterByTypes
```

## Implementation Steps

### Step 1: Add FindBridgeNodesViaSpatial Repository Method

**File:** `backend/internal/repository/tracer_repository.go`

Add new method that walks DOWN from source in spatial-topology to find nodes that have edges in the target topology:

```go
// FindBridgeNodesViaSpatial walks down from sourceDBID in spatial-topology
// and returns nodes that also have edges in the target topology slug.
// Used to find electrical/cooling nodes reachable from spatial/whitespace sources.
func (r *TracerRepository) FindBridgeNodesViaSpatial(sourceDBID uint, targetSlug string, maxDepth int) ([]TracedNode, error) {
    var nodes []TracedNode
    err := r.db.Raw(`
        WITH RECURSIVE spatial_desc AS (
            SELECT bn.id, bn.node_id, bn.name, bn.node_type, 1 as level
            FROM blueprint_edges be
            JOIN blueprint_nodes bn ON bn.id = be.to_node_id
            JOIN blueprint_types bt ON bt.id = be.blueprint_type_id
            WHERE be.from_node_id = ? AND bt.slug = 'spatial-topology'

            UNION ALL

            SELECT bn.id, bn.node_id, bn.name, bn.node_type, sd.level + 1
            FROM spatial_desc sd
            JOIN blueprint_edges be ON be.from_node_id = sd.id
            JOIN blueprint_nodes bn ON bn.id = be.to_node_id
            JOIN blueprint_types bt ON bt.id = be.blueprint_type_id
            WHERE bt.slug = 'spatial-topology' AND sd.level < ?
        )
        SELECT DISTINCT ON (sd.id) sd.id, sd.node_id, sd.name, sd.node_type, sd.level
        FROM spatial_desc sd
        WHERE EXISTS (
            SELECT 1 FROM blueprint_edges be2
            JOIN blueprint_types bt2 ON bt2.id = be2.blueprint_type_id
            WHERE bt2.slug = ?
              AND (be2.from_node_id = sd.id OR be2.to_node_id = sd.id)
        )
        ORDER BY sd.id, sd.level
    `, sourceDBID, maxDepth, targetSlug).Scan(&nodes).Error
    return nodes, err
}
```

This finds spatial descendants of the source that also participate in the target topology's edges. For Rack, it finds RackPDU (spatial child that has electrical edges). For Capacity Cell, it finds Rack (spatial child that has electrical/cooling edges).

### Step 2: Add HasEdgesInTopology Repository Helper

**File:** `backend/internal/repository/tracer_repository.go`

Quick check whether a node has any edges in a given topology:

```go
// HasEdgesInTopology returns true if the node has at least one edge in the given topology.
func (r *TracerRepository) HasEdgesInTopology(nodeDBID uint, typeSlug string) bool {
    var count int64
    r.db.Raw(`
        SELECT COUNT(*) FROM blueprint_edges be
        JOIN blueprint_types bt ON bt.id = be.blueprint_type_id
        WHERE bt.slug = ? AND (be.from_node_id = ? OR be.to_node_id = ?)
    `, typeSlug, nodeDBID, nodeDBID).Scan(&count)
    return count > 0
}
```

### Step 3: Refactor TraceDependencies Upstream Loop

**File:** `backend/internal/service/dependency_tracer.go` — replace lines 122-155

Replace the current upstream loop + fallback with spatial-bridge-aware logic:

```go
for topo, allowedTypes := range upstreamByTopo {
    slug := t.resolveSlug(topo)
    if slug == "" {
        continue
    }

    // Direct path: source has edges in this topology
    if t.repo.HasEdgesInTopology(node.ID, slug) {
        nodes, err := t.repo.FindUpstreamNodes(node.ID, slug, maxLevels)
        if err != nil {
            log.Printf("WARNING: upstream trace failed for %s in %s: %v", nodeID, topo, err)
            continue
        }
        filtered := filterByTypes(nodes, allowedTypes)
        resp.Upstream = append(resp.Upstream, groupByLevel(filtered, topo)...)
        continue
    }

    // Spatial-bridge path: find bridge nodes via spatial hierarchy
    bridgeNodes, err := t.repo.FindBridgeNodesViaSpatial(node.ID, slug, 3)
    if err != nil || len(bridgeNodes) == 0 {
        continue
    }

    for _, bridge := range bridgeNodes {
        upNodes, err := t.repo.FindUpstreamNodes(bridge.ID, slug, maxLevels)
        if err != nil {
            continue
        }
        // Include the bridge node itself if it's an allowed type
        if allowedTypes[bridge.NodeType] {
            upNodes = append([]repository.TracedNode{
                {ID: bridge.ID, NodeID: bridge.NodeID, Name: bridge.Name,
                 NodeType: bridge.NodeType, Level: bridge.Level,
                 ParentNodeID: &node.NodeID},
            }, upNodes...)
        }
        // Shift levels by bridge distance
        for i := range upNodes {
            upNodes[i].Level += bridge.Level
        }
        filtered := filterByTypes(upNodes, allowedTypes)
        resp.Upstream = append(resp.Upstream, groupByLevel(filtered, topo)...)
    }
}
```

### Step 4: Remove Old Fallback Code

Delete the entire fallback block (lines 136-152 in current code):

```go
// DELETE THIS BLOCK:
// If no matching upstream found, also trace from this node's direct children
// in the same topology...
```

It's replaced by the spatial-bridge path in Step 3.

### Step 5: Handle Whitespace Bridge Chains

For Whitespace nodes (Capacity Cell → Room Bundle → Room PDU Bundle), the spatial bridge needs to also check whitespace edges. Add a parallel bridge check:

In `dependency_tracer.go`, after the spatial bridge attempt, if still no results:

```go
// If spatial bridge found nothing, try whitespace bridge
if len(bridgeNodes) == 0 {
    whiteSlug := t.resolveSlug("Whitespace Blueprint")
    if whiteSlug != "" {
        bridgeNodes, err = t.repo.FindBridgeNodesViaSpatial(node.ID, slug, 5)
        // ... same pattern as above
    }
}
```

Actually, a cleaner approach: make `FindBridgeNodesViaSpatial` accept a list of bridge topology slugs to search, not just 'spatial-topology'. For the initial version, try spatial first, then whitespace. But keep it simple — this can be extended later.

### Step 6: Verify Rack Dependencies Still Work

```bash
# Should return full upstream chain for both cooling + electrical
curl -s "http://localhost:8889/api/trace/dependencies/RACK-R1-Z1-R1-01?levels=6&include_local=true" | python3 -c "
import sys, json
d = json.load(sys.stdin)['data']
print('upstream groups:', len(d.get('upstream', [])))
for g in d.get('upstream', []):
    types = [n['node_type'] for n in g['nodes']]
    print(f'  L{g[\"level\"]} {g[\"topology\"]}: {types}')
print('local groups:', len(d.get('local', [])))
"
```

Expected:
- Cooling: Air Zone (L1), Liquid Loop (L1), ACU/CDU (L2), CoolingDist (L3), CoolingPlant (L4)
- Electrical: RPP (L2), Room PDU (L3), UPS (L4), BESS/SWGR (L5), Gen/Util (L6)
- Local: RDHx
- NO duplicate node_ids

### Step 7: Test Whitespace Node Dependencies

```bash
# Capacity Cell should now find cross-topology deps via spatial bridge
curl -s "http://localhost:8889/api/trace/dependencies/CC-R1R2?levels=6&include_local=true" | python3 -m json.tool | head -40
```

## Todo List

- [x] Add FindBridgeNodesViaSpatial to tracer_repository.go
- [x] Add HasEdgesInTopology to tracer_repository.go
- [x] Refactor TraceDependencies upstream loop with spatial-bridge path
- [x] Remove old fallback code
- [x] Verify Rack dependencies (cooling + electrical chains)
- [x] Verify no duplicate node_ids in upstream
- [x] Test Capacity Cell cross-topology dependencies
- [x] Run unit tests: `cd backend && go test ./internal/...`

## Success Criteria

1. Rack deps: same results as before but with no duplicate nodes and correct parent chains
2. Capacity Cell deps: returns non-empty upstream (electrical + cooling via Rack spatial bridge)
3. No regressions for electrical mid-chain nodes (UPS, RPP, Room PDU)
4. All unit tests pass

## Risk Assessment

- **Medium risk:** Replacing fallback changes behavior for all node types. Must test electrical mid-chain nodes too.
- **Mitigation:** `HasEdgesInTopology` check ensures direct path is preferred when available — only spatial bridge for cross-topology cases.
