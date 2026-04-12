# Codebase Exploration Report: DAG, Dependency/Impact Logic & CSV Ingestion

**Date:** 2026-04-06 | **Scope:** Complete backend + frontend systems

---

## 1. DAG Tree Display & Rendering (Frontend)

### React Flow Components

| File | Purpose | Key Functions/Components |
|------|---------|--------------------------|
| `/frontend/src/components/tracer/dependency-impact-dag.tsx` | **Main DAG orchestrator component** - Uses React Query to fetch trace data and ReactFlow for rendering. Two queries run in parallel (dependencies & impacts). | `DependencyImpactDAGInner()` - manages node/edge state and layout |
| `/frontend/src/components/tracer/dag-helpers.tsx` | **DAG layout & conversion logic** - Converts backend trace responses to ReactFlow elements; applies Dagre LR layout for positioning | `traceToDAGElements()` - merges dep + impact traces into nodes/edges; `layoutDAG()` - applies Dagre left-to-right layout |
| `/frontend/src/components/tracer/dag-node.tsx` | **ReactFlow node renderer** - Renders individual nodes with topology colors, icons, badges (Source/Level), handles click events | Visual styling with TOPOLOGY_CONFIG colors (Electrical/orange, Cooling/cyan, Spatial/purple, Whitespace/green) |
| `/frontend/src/components/tracer/dag-edge.tsx` | **ReactFlow edge renderer** - Smooth step paths with animated arrows. Upstream=cyan, Downstream=orange, Local=gray dashed | `TracerEdge` - wraps BaseEdge with glow layer |
| `/frontend/src/components/tracer/dag-search.tsx` | **Node search + filter UI** - Typeahead search with type filtering. Fetches from `/api/blueprints/nodes` | Type filter dropdown (Rack, RPP, UPS, etc.); debounced search |
| `/frontend/src/components/tracer/dag-detail-popup.tsx` | **Node detail modal** - Displays on node click | Links to node details page |
| `/frontend/src/components/tracer/dag-types.ts` | **TypeScript interfaces** - API response types matching backend JSON | `TraceResponse`, `TraceLevelGroup`, `TraceLocalGroup`, `TracedNode`, `TracerNodeData` |

### DAG Rendering Strategy

1. **Data Flow:**
   - Query 1: `GET /api/trace/dependencies/{nodeId}?levels=2&include_local=true` → upstream + local
   - Query 2: `GET /api/trace/impacts/{nodeId}?levels=2` → downstream + load
   - Both run in parallel via React Query

2. **Node Grouping:**
   - Source node always at ring 0 (center, with amber glow)
   - **Upstream nodes** arranged left of source, grouped by level (L1, L2, etc.)
   - **Local nodes** grouped in a dashed "Local" container around source (vertical stack)
   - **Downstream nodes** arranged right of source, grouped by level
   - **Load nodes** appear as secondary connections from source

3. **Layout Engine:**
   - Uses **Dagre** (`@dagrejs/dagre`) with `rankdir: "LR"` (left-to-right)
   - Group nodes (Local container) treated as single unit in layout
   - NodeWidth/Height: 180×72px (adjustable for responsive design)
   - Edge markers: arrows at ends, different colors per relationship type

---

## 2. Dependency & Impact API Endpoints (Backend Handlers)

### Tracer Handler

**File:** `/backend/internal/handler/tracer_handler.go`

| Endpoint | Method | Purpose | Handler |
|----------|--------|---------|---------|
| `POST /api/models/ingest` | POST | Ingest 3 model CSVs (Capacity Nodes, Dependencies, Impacts) | `IngestModels()` → calls `ModelIngestionService.IngestAll()`, refreshes tracer lookups |
| `GET /api/trace/dependencies/:nodeId` | GET | Fetch upstream + local dependencies | `TraceDependencies(nodeId, levels, include_local)` → calls `DependencyTracer.TraceDependencies()` |
| `GET /api/trace/impacts/:nodeId` | GET | Fetch downstream + load impacts | `TraceImpacts(nodeId, levels)` → calls `DependencyTracer.TraceImpacts()` |
| `GET /api/models/capacity-nodes` | GET | List capacity node type metadata | `ListCapacityNodes()` → returns `CapacityNodeType` records |

### Blueprint Handler

**File:** `/backend/internal/handler/blueprint_handler.go`

| Endpoint | Method | Purpose | Notes |
|----------|--------|---------|-------|
| `GET /api/blueprints/types` | GET | List all blueprint types (domains) | Returns `BlueprintType` array |
| `GET /api/blueprints/nodes` | GET | Search nodes by type/search term | Supports pagination, filtering |
| `GET /api/blueprints/nodes/:nodeId` | GET | Get single node + memberships | Returns node details + topology memberships |
| `GET /api/blueprints/edges` | GET | List edges for a blueprint type | Requires `type` query param |
| `GET /api/blueprints/tree/:typeSlug` | GET | Get recursive tree structure | For hierarchical topology display |
| `POST /api/blueprints/ingest` | POST | Ingest blueprint CSVs | Discovers domains, parses Nodes/Edges CSVs |

---

## 3. Dependency & Impact Tracing Service

### Core Service Logic

**File:** `/backend/internal/service/dependency_tracer.go` (311 lines)

```go
type DependencyTracer struct {
  repo       *repository.TracerRepository
  topoLookup map[string]string // nodeType → topology name
  slugLookup map[string]string // topology name → blueprint_type slug
}
```

**Key Methods:**

1. **`TraceDependencies(nodeID, maxLevels, includeLocal)`**
   - Finds all nodes that the given node depends on (upstream)
   - Also finds local dependencies (same topology, direct neighbors)
   - Returns structured `TraceResponse` with source + upstream groups + local groups
   - **Fallback logic:** If no upstream found directly, traces from node's downstream children (handles Rack→RACKPDU case)

2. **`TraceImpacts(nodeID, maxLevels)`**
   - Finds all nodes impacted by changes to the given node (downstream)
   - Also finds load nodes (spatial parents of electrical descendants)
   - Returns structured `TraceResponse` with source + downstream groups + load groups
   - **Two-strategy load resolution:**
     - Strategy 1: Walk downstream in infra topologies (finds Racks in Cooling edges)
     - Strategy 2: Find spatial ancestors of deep downstream nodes (finds Load nodes via spatial-topology)

3. **`RefreshLookups()`**
   - Rebuilds topology mappings from `capacity_node_types` table
   - Maps blueprint_type names → slugs with case-insensitive prefix matching
   - Called after CSV ingestion to sync cached lookups

**Helper Functions:** `/backend/internal/service/dependency_tracer_helpers.go` (69 lines)

- `groupDepRules()` - separates dependency rules by topology, groups as Upstream/Local
- `groupImpactRules()` - separates impact rules by topology, groups as Downstream/Load
- `filterByTypes()` - keeps only nodes in allowed type set
- `groupByLevel()` - groups traced nodes by their graph level

---

## 4. Tracer Repository (Database Queries)

**File:** `/backend/internal/repository/tracer_repository.go`

| Method | SQL Pattern | Purpose |
|--------|------------|---------|
| `FindUpstreamNodes(sourceDBID, typeSlug, maxLevel)` | WITH RECURSIVE upstream AS (walk parent edges) | Recursively traces parent edges up to maxLevel hops |
| `FindDownstreamNodes(sourceDBID, typeSlug, maxLevel)` | WITH RECURSIVE downstream AS (walk child edges) | Recursively traces child edges up to maxLevel hops |
| `FindSpatialAncestorsOfType(nodeDBIDs, nodeTypes)` | WITH RECURSIVE ancestors AS (walk spatial-topology edges) | Finds spatial parents (Rack, Row, Zone) of given nodes |
| `FindLocalNodes(sourceDBID, typeSlug)` | Direct edge neighbors (both directions) | Returns direct neighbors in given topology (level=0) |
| `GetDependencyRules(nodeType)` | SELECT FROM dependency_rules WHERE node_type=? | Returns all dependency rules for a type |
| `GetImpactRules(nodeType)` | SELECT FROM impact_rules WHERE node_type=? | Returns all impact rules for a type |

**Key Design:**
- All recursive queries use **blueprint_edges** table with **blueprint_types** slug filtering
- **Spatial topology** (slug: `'spatial-topology'`) is special for finding Load nodes
- Queries return `TracedNode` struct: `{ID, NodeID, Name, NodeType, Level, ParentNodeID}`
- Level tracks hop distance; ParentNodeID enables edge reconstruction

---

## 5. CSV Ingestion Logic

### Model CSV Parsers

**File:** `/backend/internal/service/model_csv_parser.go`

**Three CSV Types with Parsers:**

#### 1. **Capacity Nodes.csv**
```
Columns: NodeType | Topology | IsCapacityNode | ActiveConstraint
Function: ParseCapacityNodesCSV()
Returns: []CapacityNodeTypeRow
```
- 26 node types (24 main + 2 bundle types)
- Maps node types to topologies (Electrical System, Cooling System, Spatial Topology, Whitespace Blueprint)
- Flags which types are capacity domains or active constraints

#### 2. **Dependencies.csv**
```
Columns: NodeType | DependencyNodeType | RelationshipType | TopologicalRelationship | UpstreamLevel
Function: ParseDependenciesCSV()
Returns: []DependencyRuleRow
```
- **152 rows** of type-level dependency rules
- `TopologicalRelationship`: "Upstream" or "Local"
- `UpstreamLevel`: hop distance (1-6) or null for local
- Examples:
  - `Rack depends on RPP (Upstream, L1)`
  - `Rack depends on RDHx (Local)`

#### 3. **Impacts.csv**
```
Columns: NodeType | ImpactNodeType | TopologicalRelationship | DownstreamLevel
Function: ParseImpactsCSV()
Returns: []ImpactRuleRow
```
- **140 rows** of type-level impact rules
- `TopologicalRelationship`: "Downstream" or "Load"
- `DownstreamLevel`: hop distance (1-6) or null for load
- Examples:
  - `RPP impacts Rack PDU (Downstream, L1)`
  - `RPP impacts Rack (Load)`

### Model Ingestion Service

**File:** `/backend/internal/service/model_ingestion_service.go`

```go
type ModelIngestionService struct {
  db *gorm.DB
}

// IngestAll(basePath)
// 1. Parse 3 CSVs from basePath/Capacity Nodes.csv, Dependencies.csv, Impacts.csv
// 2. Upsert in single transaction:
//    - capacity_node_types (upsert by node_type)
//    - dependency_rules (upsert by node_type + dependency_node_type)
//    - impact_rules (upsert by node_type + impact_node_type)
// 3. Returns ModelIngestionSummary with counts + duration
```

### Blueprint CSV Parsers

**File:** `/backend/internal/service/blueprint_csv_parser.go`

**Two CSV Types per Domain:**

1. **Nodes.csv**
   - Columns: NodeID | Name | Role | OrgPath | NodeType
   - Parser: `ParseNodesCSV()`
   - Discovers node instances in a blueprint domain

2. **Edges.csv**
   - Columns: FromName | FromNodeID | FromOrgPath | ToName | ToNodeID | ToOrgPath
   - Parser: `ParseEdgesCSV()`
   - Defines connectivity between nodes (parent→child relationships)

**Domain Discovery:**
- `DiscoverDomains(basePath)` scans for subdirectories (e.g., "Electrical system_Blueprint")
- `FindCSVFile(domainPath, kind)` locates "Node" or "Edge" CSVs in domain folder
- Ingestion service processes each domain, creating BlueprintType + BlueprintNodes + BlueprintEdges

---

## 6. Models (Data Structures)

### Type-Level Rules

| Model | File | Columns | Purpose |
|-------|------|---------|---------|
| `DependencyRule` | `/backend/internal/model/dependency_rule.go` | NodeType, DependencyNodeType, RelationshipType, TopologicalRelationship, UpstreamLevel | Type-level dependency (e.g., "Rack depends on RPP") |
| `ImpactRule` | `/backend/internal/model/impact_rule.go` | NodeType, ImpactNodeType, TopologicalRelationship, DownstreamLevel | Type-level impact (e.g., "RPP impacts Rack PDU") |
| `CapacityNodeType` | `/backend/internal/model/capacity_node_type.go` | NodeType, Topology, IsCapacityNode, ActiveConstraint | Type metadata (26 rows) |

### Instance-Level (Blueprint)

| Model | File | Columns | Purpose |
|-------|------|---------|---------|
| `BlueprintType` | `/backend/internal/model/blueprint_type.go` | ID, Slug, Name, FolderName | Blueprint domain (e.g., "electrical-system", "cooling-system") |
| `BlueprintNode` | `/backend/internal/model/blueprint_node.go` | ID, NodeID (string), Name, NodeType, NodeRole | Node instance in a domain |
| `BlueprintEdge` | `/backend/internal/model/blueprint_edge.go` | ID, FromNodeID, ToNodeID, BlueprintTypeID | Edge between nodes in a domain |
| `BlueprintNodeMembership` | `/backend/internal/model/blueprint_node_membership.go` | ID, BlueprintNodeID, BlueprintTypeID, OrgPath | Maps nodes to domains (many-to-many) |

---

## 7. Upstream vs. Downstream Placement Rules

### Determining Direction

**Dependency Rules determine UPSTREAM placement:**
- `TopologicalRelationship = "Upstream"` → node goes left of source
- `UpstreamLevel` field (1-6) determines distance/ring placement
- Local dependencies (`TopologicalRelationship = "Local"`) group around source

**Impact Rules determine DOWNSTREAM placement:**
- `TopologicalRelationship = "Downstream"` → node goes right of source
- `DownstreamLevel` field (1-6) determines distance/ring placement
- Load impacts (`TopologicalRelationship = "Load"`) group around source

### Example Flow (Rack node):

```
Dependencies.csv: Rack → RPP (Upstream, L1)
Impact.csv: RPP → Rack (Load)

Frontend DAG:
  [RPP L1] ← [Rack(SRC)] → [Rack PDU L1]
```

### Level Assignment Logic

1. **Rule engine selects rules** based on node type from `dependency_rules` or `impact_rules`
2. **Topology lookup** converts rule's target type to topology name via `CapacityNodeType`
3. **Slug resolution** maps topology name to blueprint_type slug
4. **Recursive query** executes with maxLevels limit, returns nodes at each level
5. **Filtering** keeps only nodes matching rule's allowed types
6. **Fallback logic** (for dependencies): if upstream empty, tries node's downstream children

---

## 8. File Tree Summary

```
backend/
├── internal/
│   ├── handler/
│   │   ├── tracer_handler.go         [API endpoints for trace + ingest]
│   │   └── blueprint_handler.go      [API endpoints for blueprints]
│   ├── service/
│   │   ├── dependency_tracer.go      [Trace logic + topology lookups]
│   │   ├── dependency_tracer_helpers.go [Rule grouping, filtering, leveling]
│   │   ├── model_csv_parser.go       [Dependencies/Impacts/Capacity parsing]
│   │   ├── model_ingestion_service.go [CSV ingestion orchestration]
│   │   ├── blueprint_csv_parser.go   [Blueprint Nodes/Edges parsing]
│   │   └── blueprint_ingestion_service.go [Blueprint ingestion]
│   ├── repository/
│   │   ├── tracer_repository.go      [Recursive SQL queries for tracing]
│   │   └── blueprint_repository.go   [Blueprint CRUD operations]
│   ├── model/
│   │   ├── dependency_rule.go        [Type-level rules]
│   │   ├── impact_rule.go            [Type-level rules]
│   │   ├── capacity_node_type.go     [Type metadata]
│   │   ├── blueprint_*.go            [Instance models]
│   │   └── user.go                   [Auth]
│   ├── router/router.go              [Route definitions]
│   └── middleware/auth_middleware.go [JWT auth]
├── cmd/server/main.go                [Entry point]
└── testdata/
    ├── models/                       [Model CSV test fixtures]
    └── blueprint/                    [Blueprint test CSVs]

frontend/src/
├── app/tracer/
│   ├── page.tsx                      [Tracer page]
│   └── layout.tsx
├── components/tracer/
│   ├── dependency-impact-dag.tsx     [Main DAG component]
│   ├── dag-helpers.tsx               [Layout + conversion logic]
│   ├── dag-types.ts                  [TypeScript interfaces]
│   ├── dag-node.tsx                  [Node renderer]
│   ├── dag-edge.tsx                  [Edge renderer]
│   ├── dag-search.tsx                [Search UI]
│   └── dag-detail-popup.tsx          [Detail modal]
├── lib/
│   ├── api-client.ts                 [HTTP client]
│   └── query-client.ts               [React Query setup]
└── hooks/
    └── use-api.ts                    [API hooks]

blueprint/
├── Dependencies.csv                  [152 type-level rules]
├── Impacts.csv                       [140 type-level rules]
├── Capacity Nodes.csv                [26 type metadata]
└── Node & Edge/
    ├── Electrical system_Blueprint/  [Nodes/Edges CSVs]
    ├── Cooling system_Blueprint/
    ├── Spatial Topology_Blueprint/
    ├── Whitespace Blueprint/
    └── Operational infrastructure_Blueprint/
```

---

## 9. Key Functions/Types Reference

### Backend (Go)

**Type-Level Rules & Lookups:**
- `DependencyRule` - one row = "A depends on B (Upstream, L1)"
- `ImpactRule` - one row = "A impacts B (Downstream, L2)"
- `CapacityNodeType` - NodeType → Topology mapping
- `DependencyTracer.topoLookup` - NodeType → Topology string
- `DependencyTracer.slugLookup` - Topology → BlueprintType.Slug

**Tracing Methods:**
- `TraceDependencies()` - returns upstream + local (4 rule groups)
- `TraceImpacts()` - returns downstream + load (4 rule groups)
- `RefreshLookups()` - rebuilds mappings after CSV ingest

**Repository Queries:**
- `FindUpstreamNodes()` - recursive WITH query, parent edge walk
- `FindDownstreamNodes()` - recursive WITH query, child edge walk
- `FindSpatialAncestorsOfType()` - recursive spatial-topology walk
- `FindLocalNodes()` - direct neighbors

**CSV Parsing:**
- `ParseDependenciesCSV()` - reads Dependencies.csv
- `ParseImpactsCSV()` - reads Impacts.csv
- `ParseCapacityNodesCSV()` - reads Capacity Nodes.csv
- `ModelIngestionService.IngestAll()` - orchestrates 3-CSV upsert
- `BlueprintIngestionService.IngestAll()` - orchestrates blueprint domain CSVs

### Frontend (React/TypeScript)

**Main Component:**
- `DependencyImpactDAGInner()` - orchestrates trace queries + layout
- `traceToDAGElements()` - converts API trace to ReactFlow nodes/edges
- `layoutDAG()` - applies Dagre LR layout positioning

**UI Components:**
- `TracerNode` - renders individual nodes with topology styling
- `TracerEdge` - renders animated edges with glows
- `DAGSearch` - typeahead + type filter for node selection

**Response Types:**
- `TraceResponse` - root response (source + upstream/downstream/local/load)
- `TraceLevelGroup` - nodes at specific level in topology
- `TraceLocalGroup` - level-less groups (local/load)
- `TracedNode` - single node result (id, node_id, name, node_type, level)

---

## 10. CSV Data Files

| File | Location | Format | Use |
|------|----------|--------|-----|
| `Dependencies.csv` | `/blueprint/` | Type-level rules | Defines upstream dependencies for all 26 node types |
| `Impacts.csv` | `/blueprint/` | Type-level rules | Defines downstream impacts for all 26 node types |
| `Capacity Nodes.csv` | `/blueprint/` | Type metadata | Maps 26 node types to 4 topologies + capacity flags |
| Domain Nodes CSVs | `/blueprint/Node & Edge/*/` | Instance data | Actual nodes in each topology (e.g., "Rack-1", "UPS-A") |
| Domain Edges CSVs | `/blueprint/Node & Edge/*/` | Instance connectivity | Edges connecting nodes within a topology |

---

## Summary

**DAG System has 3 layers:**

1. **Type-Level Rules** (CSV → DB): Dependency & Impact rules define which types relate + at what level
2. **Instance Tracing** (Go service): Resolves rules to actual node instances via recursive SQL queries
3. **Frontend Rendering** (React): Converts trace response to ReactFlow DAG with Dagre layout

**Upstream vs Downstream** is determined purely by `TopologicalRelationship` field in CSV rules:
- "Upstream" + UpstreamLevel → left side of DAG
- "Downstream" + DownstreamLevel → right side of DAG
- "Local" + "Load" → grouped near source node

---

**Report Generated:** 2026-04-06 | **Thoroughness:** Complete
