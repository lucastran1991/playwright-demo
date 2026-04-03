# Brainstorm: Blueprint CSV Ingestion Backend Model

## Problem Statement
Ingest 6 blueprint domain CSVs (Cooling, Electrical, Spatial, Operational, Whitespace, Deployment) into PostgreSQL. Each domain has Nodes + Edges files representing hierarchical tree structures of data center infrastructure.

## Raw Data Analysis

### CSV Schemas (uniform across all domains)
- **Nodes**: `Node ID, Node Name, Node Role, Org Path, Node Type`
- **Edges**: `From Node Name, From Node ID, From Node Org Path, To Node Name, To Node ID, To Node Org Path`

### Key Observations
- Node Role column is always empty across all 6 domains
- Org Path encodes hierarchy as `/`-separated string
- Edges represent parent→child containment (trees, not graphs)
- ~5,300 total unique nodes, ~6,300 total edges
- **Massive cross-domain overlap**: same Node IDs appear in multiple domains (e.g., 1,191 shared between Operational and Spatial). Same physical asset viewed through different blueprint lenses.

### Data Volume Per Domain
| Domain | Nodes | Edges |
|--------|-------|-------|
| Electrical | 1,943 | 2,643 |
| Spatial | 1,291 | 1,292 |
| Operational | 1,212 | 1,196 |
| Cooling | 591 | 876 |
| Whitespace | 283 | 293 |
| Deployment | 17 | 19 |

## Agreed Design Decisions
1. **Unified nodes** -- single `blueprint_nodes` table. One row per unique Node ID.
2. **Adjacency list** -- edges table defines parent→child. Use recursive CTEs for tree queries.
3. **Generic loader** -- one service auto-discovers CSV files from blueprint dir by folder name.
4. **Upsert ingestion** -- idempotent via ON CONFLICT DO UPDATE.
5. **REST API trigger** -- `POST /api/blueprints/ingest` endpoint.

## Proposed Database Models

### Table: `blueprint_types`
Represents the 6 domain categories.

```go
type BlueprintType struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string    `gorm:"uniqueIndex;size:100;not null"` // e.g. "Cooling System"
    Slug      string    `gorm:"uniqueIndex;size:100;not null"` // e.g. "cooling-system"
    FolderName string   `gorm:"size:255;not null"`             // e.g. "Cooling system_Blueprint"
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### Table: `blueprint_nodes`
Unified node table. One row per unique Node ID across all domains.

```go
type BlueprintNode struct {
    ID        uint      `gorm:"primaryKey"`
    NodeID    string    `gorm:"uniqueIndex;size:255;not null"` // e.g. "RACK-R1-Z1-R1-01"
    Name      string    `gorm:"size:500;not null"`             // e.g. "Rack 1 in ROW-R1-Z1-R1"
    NodeType  string    `gorm:"index;size:100"`                // e.g. "Rack", "Server"
    NodeRole  string    `gorm:"size:100"`                      // currently empty, future use
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### Table: `blueprint_node_memberships`
Maps which nodes belong to which blueprint domains + stores domain-specific org path.

```go
type BlueprintNodeMembership struct {
    ID              uint      `gorm:"primaryKey"`
    BlueprintTypeID uint      `gorm:"index;not null"`
    BlueprintNodeID uint      `gorm:"index;not null"`
    OrgPath         string    `gorm:"size:1000"`               // domain-specific hierarchy path
    CreatedAt       time.Time
    UpdatedAt       time.Time

    BlueprintType BlueprintType `gorm:"foreignKey:BlueprintTypeID"`
    BlueprintNode BlueprintNode `gorm:"foreignKey:BlueprintNodeID"`
}
// Unique constraint: (BlueprintTypeID, BlueprintNodeID)
```

### Table: `blueprint_edges`
Parent→child relationships, scoped to a blueprint domain.

```go
type BlueprintEdge struct {
    ID              uint      `gorm:"primaryKey"`
    BlueprintTypeID uint      `gorm:"index;not null"`
    FromNodeID      uint      `gorm:"index;not null"`          // FK to blueprint_nodes.id
    ToNodeID        uint      `gorm:"index;not null"`          // FK to blueprint_nodes.id
    CreatedAt       time.Time
    UpdatedAt       time.Time

    BlueprintType BlueprintType `gorm:"foreignKey:BlueprintTypeID"`
    FromNode      BlueprintNode `gorm:"foreignKey:FromNodeID"`
    ToNode        BlueprintNode `gorm:"foreignKey:ToNodeID"`
}
// Unique constraint: (BlueprintTypeID, FromNodeID, ToNodeID)
```

## ER Diagram (ASCII)

```
blueprint_types          blueprint_node_memberships         blueprint_nodes
┌──────────────┐        ┌─────────────────────────┐       ┌─────────────────┐
│ id (PK)      │───┐    │ id (PK)                 │    ┌──│ id (PK)         │
│ name         │   └───>│ blueprint_type_id (FK)   │    │  │ node_id (UQ)    │
│ slug         │        │ blueprint_node_id (FK)   │<───┘  │ name            │
│ folder_name  │        │ org_path                 │       │ node_type       │
│ created_at   │        │ created_at               │       │ node_role       │
│ updated_at   │        │ updated_at               │       │ created_at      │
└──────────────┘        └─────────────────────────┘       │ updated_at      │
       │                                                   └─────────────────┘
       │                blueprint_edges                           │  │
       │               ┌──────────────────────┐                   │  │
       └──────────────>│ id (PK)              │                   │  │
                       │ blueprint_type_id(FK) │                   │  │
                       │ from_node_id (FK)     │<──────────────────┘  │
                       │ to_node_id (FK)       │<─────────────────────┘
                       │ created_at            │
                       │ updated_at            │
                       └──────────────────────┘
```

## Ingestion Flow

```
POST /api/blueprints/ingest
  │
  ├─ 1. Scan blueprint/Node & Edge/ for subdirectories
  ├─ 2. For each folder: upsert BlueprintType by folder name
  ├─ 3. Parse *_Nodes.csv → upsert BlueprintNode by node_id
  │      └─ upsert BlueprintNodeMembership (type + node + org_path)
  ├─ 4. Parse *_Edges.csv → resolve from/to node_ids → upsert BlueprintEdge
  └─ 5. Return summary: {domains: 6, nodes_upserted: N, edges_upserted: M}
```

## API Design

```
POST   /api/blueprints/ingest              -- trigger full ingestion
GET    /api/blueprints/types               -- list all blueprint domains
GET    /api/blueprints/nodes?type=cooling  -- list nodes, filterable by domain
GET    /api/blueprints/nodes/:nodeId       -- single node with all memberships
GET    /api/blueprints/edges?type=cooling  -- edges for a domain
GET    /api/blueprints/tree/:typeSlug      -- recursive tree for a domain
```

## Implementation Considerations

### Why unified nodes + membership join table
- Same rack ID appears in Cooling, Electrical, Spatial, Operational, Whitespace
- Unified table avoids data duplication, enables cross-domain queries ("show all blueprints this rack belongs to")
- Membership table stores domain-specific org_path (differs per domain for same node)
- Trade-off: slightly more complex ingestion logic vs much cleaner data model

### Upsert Strategy
- Nodes: `ON CONFLICT (node_id) DO UPDATE SET name, node_type, updated_at`
- Memberships: `ON CONFLICT (blueprint_type_id, blueprint_node_id) DO UPDATE SET org_path`
- Edges: `ON CONFLICT (blueprint_type_id, from_node_id, to_node_id) DO NOTHING`
- Wrap each domain in a transaction for atomicity

### Performance
- Total data is small (~5K nodes, ~6K edges). No need for batch processing or async jobs
- Indexes on: `node_id`, `blueprint_type_id`, `from_node_id`, `to_node_id`, `node_type`
- Recursive CTE for tree queries will be fast at this scale

### Node Type Conflict
- Same node can have different Node Types across domains (rare but possible)
- Decision: store the first-seen type, or last-write-wins on upsert
- Could add node_type to membership table if domain-specific types needed later

## Risk Assessment
- **Low risk**: data volume is small, schema is uniform, no real-time requirements
- **Medium risk**: Node Type conflicts across domains -- monitor during ingestion, log warnings
- **Low risk**: CSV format changes -- generic loader should validate headers

## Success Criteria
- All 6 domains ingested without errors
- Idempotent: re-running produces same result
- Cross-domain queries work (find all blueprints for a given node)
- Tree traversal via recursive CTE returns correct hierarchy

## Next Steps
1. Create GORM models in `backend/internal/model/`
2. Add migration to `database.go`
3. Build generic CSV parser service in `backend/internal/service/`
4. Add ingestion handler + routes
5. Write integration tests
