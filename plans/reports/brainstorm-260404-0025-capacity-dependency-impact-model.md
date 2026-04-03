# Brainstorm: Capacity Nodes, Dependency & Impact Model

## Problem Statement
Ingest 3 model CSVs (Capacity Nodes, Dependencies, Impacts) into PostgreSQL and build a Dependency Tracer API that resolves actual node instances from type-level rules against the ingested blueprint topology.

## Input Data Analysis

### 1. Capacity Nodes.csv (24 rows)
| Column | Type | Description |
|--------|------|-------------|
| Node Type | string | e.g. "Rack", "RPP", "UPS" |
| Topology | string | Which blueprint domain: Electrical System, Cooling System, Spatial Topology, Whitespace Blueprint |
| Capacity Node (Capacity Domain) | bool | Tracks supply/load balance |
| ActiveConstraint | bool | Actively limits capacity of downstream |

Key insight: this is **metadata about node types**, not about individual nodes. It classifies which node types are capacity domains and which are active vs passive constraints.

### 2. Dependencies.csv (147 rows)
| Column | Type | Description |
|--------|------|-------------|
| Node Type | string | Source node type (e.g. "Rack") |
| Dependency Node Type | string | What it depends on (e.g. "RPP") |
| Relationship Type | string | Always "Dependency" |
| Topological Relationship | string | "Upstream" or "Local" |
| Upstream Level | int/null | Distance from source (1=direct, 6=furthest). Null for Local. |

Key insight: type-level rules. "Every Rack depends on every RPP that's upstream of it." Resolution against actual topology needed at query time.

### 3. Impacts.csv (118 rows)
| Column | Type | Description |
|--------|------|-------------|
| Node Type | string | Source node type (e.g. "UPS") |
| Impact Node Type | string | What it impacts (e.g. "Rack") |
| Topological Relationship | string | "Downstream" or "Load" |
| Downstream Level | int/null | Distance downstream. Null for Load impacts. |

Key insight: inverse of dependencies. "If UPS fails, these node types downstream are impacted."

## Agreed Design Decisions
1. **Ingest 3 CSVs** + build Dependency Tracer API
2. **Backend resolves** actual instances from type-level rules against topology
3. Reuse existing blueprint tables (blueprint_nodes, blueprint_edges, blueprint_types, blueprint_node_memberships) for topology traversal

## Proposed Database Models

### Table: `capacity_node_types`
Metadata about node types -- which are capacity domains and constraints.

```go
type CapacityNodeType struct {
    ID               uint   `gorm:"primaryKey"`
    NodeType         string `gorm:"uniqueIndex;size:100;not null"` // e.g. "Rack"
    Topology         string `gorm:"size:100;not null"`             // e.g. "Electrical System"
    IsCapacityNode   bool   `gorm:"not null;default:false"`
    ActiveConstraint bool   `gorm:"not null;default:false"`
}
```

### Table: `dependency_rules`
Type-level dependency rules.

```go
type DependencyRule struct {
    ID                     uint   `gorm:"primaryKey"`
    NodeType               string `gorm:"index;size:100;not null"`
    DependencyNodeType     string `gorm:"size:100;not null"`
    TopologicalRelationship string `gorm:"size:20;not null"` // "Upstream" or "Local"
    UpstreamLevel          *int   `gorm:""`                   // null for Local
}
// Composite unique: (NodeType, DependencyNodeType)
```

### Table: `impact_rules`
Type-level impact rules.

```go
type ImpactRule struct {
    ID                     uint   `gorm:"primaryKey"`
    NodeType               string `gorm:"index;size:100;not null"`
    ImpactNodeType         string `gorm:"size:100;not null"`
    TopologicalRelationship string `gorm:"size:20;not null"` // "Downstream" or "Load"
    DownstreamLevel        *int   `gorm:""`                   // null for Load
}
// Composite unique: (NodeType, ImpactNodeType)
```

## Dependency Tracer API Design

### Resolution Algorithm
Given a source node (e.g. "RACK-R1-Z1-R1-01"):
1. Look up its `node_type` from `blueprint_nodes`
2. Query `dependency_rules` WHERE `node_type` = source's type
3. For each rule, find actual nodes:
   - **Upstream**: walk UP the blueprint edges from source node, filter by `DependencyNodeType`, respecting `UpstreamLevel`
   - **Local**: find sibling/contained nodes of matching type in same spatial scope (same parent in topology)
4. Return grouped by topology (Electrical vs Cooling) with level info

### API Endpoints

```
POST /api/blueprints/models/ingest    -- ingest 3 model CSVs (protected)

GET /api/blueprints/trace/dependencies/:nodeId
    ?levels=2                          -- default 2, how many upstream levels
    ?include_local=true                -- include local dependencies
    Response: { electrical: [...], cooling: [...] }

GET /api/blueprints/trace/impacts/:nodeId
    ?levels=2                          -- default 2 downstream levels
    ?load_scope=rack                   -- rack|row|zone|room|capacity_cell
    Response: { downstream: [...], load: [...] }

GET /api/blueprints/capacity-nodes     -- list all capacity node types
    Response: [{ node_type, topology, is_capacity, active_constraint }]
```

### Trace Response Shape
```json
{
  "source": { "node_id": "RACK-R1-Z1-R1-01", "name": "...", "node_type": "Rack" },
  "dependencies": {
    "upstream": [
      { "level": 1, "topology": "Electrical System", "nodes": [
        { "node_id": "RPP-R1-A2", "name": "...", "node_type": "RPP" }
      ]},
      { "level": 2, "topology": "Electrical System", "nodes": [
        { "node_id": "RPDU-R1-A", "name": "...", "node_type": "Room PDU" }
      ]}
    ],
    "local": [
      { "topology": "Electrical System", "nodes": [...] },
      { "topology": "Cooling System", "nodes": [...] }
    ]
  }
}
```

## Resolution Strategy: Upstream Walk

The key challenge: how to find "all electrical nodes upstream to the Rack PDUs of this Rack."

**Approach**: Use `blueprint_edges` from existing ingested data.
1. Start from source node
2. Walk UP edges (find parent) in the relevant blueprint domain (Electrical or Cooling)
3. At each step, check if parent's `node_type` matches any `DependencyNodeType` from the rules
4. Track level = distance from source
5. Stop at requested level depth

This is a BFS/recursive CTE query scoped to a specific `blueprint_type`.

```sql
-- Example: find upstream dependencies for a rack in electrical system
WITH RECURSIVE upstream AS (
    SELECT bn.id, bn.node_id, bn.node_type, 1 as level
    FROM blueprint_edges be
    JOIN blueprint_nodes bn ON bn.id = be.from_node_id
    JOIN blueprint_types bt ON bt.id = be.blueprint_type_id
    WHERE be.to_node_id = :source_db_id AND bt.slug = 'electrical-system'
    
    UNION ALL
    
    SELECT bn.id, bn.node_id, bn.node_type, u.level + 1
    FROM upstream u
    JOIN blueprint_edges be ON be.to_node_id = u.id
    JOIN blueprint_nodes bn ON bn.id = be.from_node_id
    JOIN blueprint_types bt ON bt.id = be.blueprint_type_id
    WHERE bt.slug = 'electrical-system' AND u.level < :max_level
)
SELECT * FROM upstream
WHERE node_type IN (SELECT dependency_node_type FROM dependency_rules WHERE node_type = :source_type)
```

## File Map (new files)

```
backend/internal/
  model/capacity_node_type.go
  model/dependency_rule.go
  model/impact_rule.go
  service/model_csv_parser.go        -- parse 3 model CSVs
  service/model_ingestion_service.go -- ingest into DB
  service/dependency_tracer.go       -- resolve actual dependencies/impacts
  handler/tracer_handler.go          -- API endpoints
  repository/tracer_repository.go    -- upstream walk queries
```

## Implementation Considerations

### Why separate tables (not reuse blueprint_edges)
- Dependency/Impact rules are **type-level metadata**, not instance-level topology edges
- They describe "Rack type depends on RPP type" -- a different semantic than "this rack is a child of this row"
- Rules persist even if topology changes; they're configuration data

### Performance
- Rules tables are tiny (147 + 118 + 24 rows) -- always cached
- Upstream walk is the expensive part: recursive CTE per query
- For ~2K nodes, CTE depth 6 is fast (<50ms)
- If needed later: materialize dependency map on topology change (YAGNI for now)

### Tracing Levels (from PDF p.21)
- Default: show 2 levels upstream, 2 levels downstream
- User-configurable via query param
- Local dependencies have no level (all shown)
- Load impacts can be scoped: rack, row, zone, room, capacity cell

### Multiple Source Nodes (from PDF p.21)
- Frontend can call trace API for each TSN (Traced Source Node)
- Backend stateless -- each call is independent
- "Remove tracing" = frontend clears all accumulated traces

## Risk Assessment
- **Medium**: upstream walk correctness depends on blueprint edge data quality. Edges must accurately represent parent→child in each topology.
- **Low**: rules CSV format changes -- header validation catches this
- **Medium**: "Local" dependency resolution is ambiguous. For Rack→Rack PDU (local), need to find Rack PDUs connected to this specific rack via edges. Strategy: find nodes that are direct parents/children/siblings of the source in the relevant topology.

## Success Criteria
- 3 CSVs ingested without errors, idempotent
- GET /trace/dependencies/:nodeId returns correct upstream nodes for any rack
- Upstream levels match the `Upstream Level` column from CSV
- Local dependencies correctly resolved
- Impact trace returns correct downstream + load-impacted nodes

## Unresolved Questions
1. **Local dependency resolution**: How exactly to find "local" nodes? Is it direct edge neighbors? Siblings under same parent? Need to clarify for Rack→RDHx, Rack→DTC specifically.
2. **Cross-domain tracing**: A rack has dependencies in BOTH Electrical and Cooling. Does the upstream walk need to happen in both topologies simultaneously, or can they be independent queries?
3. **Load impact scope**: When user selects "Row" scope for load impacts -- does that mean "if any rack in the row is impacted, highlight the whole row"? The PDF says yes, but confirming.
