# Brainstorm: Rack Dependency/Impact DAG Fix

## Problem Statement

The DAG tracer for Rack (and other spatial/whitespace leaf nodes) has multiple issues:
1. Dependencies and impacts display on the same side (all LEFT)
2. Impact API returns empty for Rack (no impact rules in Impacts.csv)
3. Whitespace nodes (Capacity Cell, Room Bundle, etc.) return completely empty for both APIs
4. Duplicate nodes (SWGR at L5 + L6) create confusing cross-edges
5. UPS impact Load resolution fails (FindSpatialAncestorsOfType not finding Rack/Row/Zone)

## Root Cause Analysis

### 1. Missing Impact Rules for Rack

Impacts.csv has NO rows where `Node Type = "Rack"`. Rack appears only as a LOAD TARGET (column 2), never as a source. Since Rack contains Servers and Rack PDUs as spatial children, it SHOULD have impact rules.

**Missing rules needed in Impacts.csv:**
```
Rack,Server,Downstream,1
Rack,Rack PDU,Downstream,1  (already exists as Load but not Downstream)
```

### 2. Cross-Topology Edge Discovery

Current code topology: Rack has edges in 3 topologies:
- **Spatial**: Row → Rack → RackPDU, Servers
- **Electrical**: Rack → RackPDU (cross-ref), RPP → RackPDU
- **Cooling**: AirZone → Rack (cross-ref)

Dependencies work via fallback:
```
Rack (no upstream electrical edges)
→ fallback: FindDownstreamNodes("electrical") → finds RackPDU
→ then: FindUpstreamNodes(RackPDU, "electrical") → RPP → Room PDU → UPS → ...
→ levels shifted +1
```

But this fallback has issues:
- Creates `parent_node_id = "RACKPDU-*"` which is NOT in the node set (filtered out by type rules)
- Frontend falls back to connecting RPP → source directly, losing the chain context

### 3. Duplicate Nodes from GROUP BY parent_node_id

`FindUpstreamNodes` SQL groups by `(id, node_id, name, node_type, parent_node_id)`. Same node appears at different levels with different parents:
```
SWGR-01, level=5, parent=UPS-01 (direct: SWGR→UPS)
SWGR-01, level=6, parent=BESS-01 (via: SWGR→BESS→UPS)
```

Frontend creates node once but 2 edges. Dagre gets conflicting rank signals.

### 4. UPS Impact Load Resolution Bug

Strategy 2 (`FindSpatialAncestorsOfType`) should find Rack as spatial parent of RACKPDU. API test confirms `load: []` for UPS impacts — 0 load groups despite 48 RACKPDU downstream nodes. Likely a SQL or GORM binding issue with the `IN ?` clause for large ID sets.

### 5. Whitespace Nodes Completely Broken

Capacity Cell, Room Bundle, Room PDU Bundle, UPS Bundle return empty for both APIs. They have edges ONLY in whitespace topology, but dependency rules point to Electrical/Cooling types. The tracer queries electrical/cooling edge sets where these nodes don't exist.

## Evaluated Approaches

### A. Spatial-Bridge Traversal (SELECTED)

Modify tracer to use spatial hierarchy as bridge for cross-topology resolution.

**Deps flow (Rack):**
1. Source: Rack (Spatial)
2. Walk UP spatial: Rack → Row → Zone
3. Walk DOWN spatial from Zone: find RPP, AIRZONE, LL sibling-children
4. Walk UP electrical from RPP: RPP → Room PDU → UPS → ...
5. Walk UP cooling from AIRZONE: AirZone → ACU → CDist → CPlant

**Impact flow (UPS):**
1. Source: UPS (Electrical)
2. Walk DOWN electrical: UPS → Room PDU → RPP → RackPDU
3. From RackPDU, find spatial parent Zone/Rack
4. Walk DOWN spatial: find Rack, Row children
5. Return as Load nodes

**Pros:** No schema changes, leverages existing edges, correct by construction
**Cons:** Multi-step traversal is complex, needs careful level numbering

### B. Flatten to Unified Trace API

Single API using type-level rules + spatial containment for direct instance matching.

**Pros:** Conceptually cleaner, matches CSV model directly
**Cons:** Bigger rewrite, new spatial query functions needed, loses edge chain context

### C. Cross-Topology Link Table

Precompute and store cross-topology links at ingestion time.

**Pros:** Fastest queries, explicit
**Cons:** New table + migration, ingestion must re-run on any change

## Recommended Solution

### Approach A + combined API endpoint

#### Phase 1: Fix Data Issues
- Add missing Rack impact rules to Impacts.csv
- Fix FindSpatialAncestorsOfType SQL (DISTINCT ON + GROUP BY conflict, or GORM IN binding)
- Remove duplicate node rows from SQL results (GROUP BY without parent_node_id, use array_agg for parents)

#### Phase 2: Refactor TraceDependencies
- Replace fallback mechanism with explicit spatial-bridge traversal
- When source is spatial/whitespace: walk spatial hierarchy to find infra topology bridge nodes
- When source is electrical/cooling: use existing edge-walking (works already)
- Normalize levels using CSV UpstreamLevel values instead of hop counts

#### Phase 3: Refactor TraceImpacts
- Fix Load resolution: ensure FindSpatialAncestorsOfType works for large ID sets
- Add Rack → downstream impact rules
- Bridge from electrical downstream to spatial hierarchy for Load nodes

#### Phase 4: Add `/api/trace/full/:nodeId` Endpoint
- Returns combined `{upstream, local, downstream, load}` in one call
- Frontend uses single API call for DAG view
- Keep existing `/dependencies` and `/impacts` for backward compatibility

#### Phase 5: Frontend DAG Fixes
- Handle duplicate parent edges (deduplicate by node_id, keep shortest level)
- Separate upstream by topology for cleaner Dagre layout (cooling chain vs electrical chain)
- Fix Load nodes rendering (currently use LOCAL_STYLE, should be distinct)

## Implementation Considerations

### Edge Direction Convention
Confirmed: `from_node_id = parent/upstream`, `to_node_id = child/downstream`. Consistent across all topologies.

### Level Normalization
Current: SQL hop count (distance from source in edge graph)
CSV: Logical level (type-level abstraction)
These differ when edges skip levels. Should use CSV levels for display, SQL levels for traversal.

### Performance
Rack dependency trace involves: 1 spatial walk + 1 electrical walk + 1 cooling walk = 3 recursive CTEs. Acceptable for single-node trace. For bulk traces, consider precomputation.

## Success Metrics
1. Rack trace returns correct upstream (Cooling + Electrical) AND downstream (Servers, RackPDU)
2. Whitespace nodes (Capacity Cell, etc.) return cross-topology dependencies via spatial bridge
3. UPS impacts include Load nodes (Rack, Row, Zone)
4. No duplicate nodes in DAG
5. DAG displays upstream LEFT, downstream RIGHT, local around source
6. Single `/trace/full` API available for frontend

## Risk Assessment
- **Spatial hierarchy assumptions**: Bridge traversal assumes spatial parent of electrical node maps to the correct zone/room. Validated with test-cases CSV.
- **Level numbering**: Shifted levels (+1 from fallback) won't match CSV levels. Need explicit level mapping.
- **Performance**: Additional recursive CTEs per trace. Monitor query time.

## Unresolved Questions
1. Should Whitespace nodes (Capacity Cell) trace through Room Bundle → Room PDU Bundle → Room PDU → electrical chain? Or through spatial bridge?
2. Should `/trace/full` be a POST (with body params) or GET (with query params)?
3. Should the combined API deduplicate nodes that appear in both dep and impact responses?
