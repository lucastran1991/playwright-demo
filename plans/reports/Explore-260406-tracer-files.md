# Dependency/Impact Tracer — File Inventory & Signatures

## Backend Service Layer

### /backend/internal/service/dependency_tracer.go (311 lines)
**Key Types & Functions:**
- `TraceResponse` struct — top-level response with Source, Upstream, Local, Downstream, Load
- `SourceNode` struct — identifies traced node (NodeID, Name, NodeType, Topology)
- `TraceLevelGroup` struct — groups nodes by level & topology
- `TraceLocalGroup` struct — groups nodes by topology (no level)
- `DependencyTracer` struct — main tracer with repo + lookups
  - `NewDependencyTracer(repo *TracerRepository) *DependencyTracer`
  - `RefreshLookups()` — reloads topology/slug mappings from DB
  - `TraceDependencies(nodeID string, maxLevels int, includeLocal bool) (*TraceResponse, error)` — upstream + local
  - `TraceImpacts(nodeID string, maxLevels int) (*TraceResponse, error)` — downstream + load
  - `resolveSlug(topology string) string` — maps topology → blueprint slug
  - `lookupTopology(nodeType string) string` — maps nodeType → topology

### /backend/internal/service/dependency_tracer_helpers.go (69 lines)
**Helper Functions:**
- `groupDepRules(rules []DependencyRule) (upstream, local map[string]map[string]bool)` — separates dependency rules
- `groupImpactRules(rules []ImpactRule) (downstream, load map[string]map[string]bool)` — separates impact rules
- `filterByTypes(nodes []TracedNode, allowed map[string]bool) []TracedNode` — filters by type whitelist
- `groupByLevel(nodes []TracedNode, topology string) []TraceLevelGroup` — groups by level

### /backend/internal/service/dependency_tracer_helpers_test.go (296 lines)
**Test Coverage:**
- `TestFilterByTypes_*` (4 tests) — empty allowed set, all filtered, partial filter, empty list
- `TestGroupByLevel_*` (5 tests) — single level, multiple levels, empty nodes, preserves data
- `TestGroupDepRules_UpstreamAndLocal` — mocked tracer, dependency rules grouping
- `TestGroupImpactRules_DownstreamAndLoad` — mocked tracer, impact rules grouping

---

## Backend Repository Layer

### /backend/internal/repository/tracer_repository.go (179 lines)
**Core Type & Methods:**
- `TracedNode` struct — result from tracing (ID, NodeID, Name, NodeType, Level, ParentNodeID)
- `TracerRepository` struct — wraps *gorm.DB
  - `NewTracerRepository(db *gorm.DB) *TracerRepository`
  - `FindUpstreamNodes(sourceDBID uint, typeSlug string, maxLevel int) ([]TracedNode, error)` — recursive parent walk
  - `FindDownstreamNodes(sourceDBID uint, typeSlug string, maxLevel int) ([]TracedNode, error)` — recursive child walk
  - `FindSpatialAncestorsOfType(nodeDBIDs []uint, nodeTypes []string) ([]TracedNode, error)` — spatial parent search
  - `FindLocalNodes(sourceDBID uint, typeSlug string) ([]TracedNode, error)` — direct neighbors
  - `GetDependencyRules(nodeType string) ([]DependencyRule, error)`
  - `GetImpactRules(nodeType string) ([]ImpactRule, error)`
  - `ListCapacityNodeTypes() ([]CapacityNodeType, error)`
  - `FindNodeByStringID(nodeID string) (*BlueprintNode, error)`
  - `ListBlueprintTypes() ([]BlueprintType, error)`

---

## Backend Handler Layer

### /backend/internal/handler/tracer_handler.go (97 lines)
**Type & Methods:**
- `TracerHandler` struct — wraps ingestion service, tracer, repo, modelDir
  - `NewTracerHandler(svc *ModelIngestionService, tracer *DependencyTracer, repo *TracerRepository, modelDir string) *TracerHandler`
  - `IngestModels(c *gin.Context)` — POST /api/models/ingest (triggers CSV ingestion, refreshes lookups)
  - `ListCapacityNodes(c *gin.Context)` — GET /api/models/capacity-nodes
  - `TraceDependencies(c *gin.Context)` — GET /api/trace/dependencies/:nodeId (with ?levels=X&include_local=true)
  - `TraceImpacts(c *gin.Context)` — GET /api/trace/impacts/:nodeId (with ?levels=X)
  - `parseIntParam(key string, defaultVal, maxVal int) int` — query param parsing helper

---

## Backend Router Layer

### /backend/internal/router/router.go (69 lines)
**Route Registrations:**
- **Public trace endpoints:**
  - `GET /api/trace/dependencies/:nodeId` → `tracerHandler.TraceDependencies`
  - `GET /api/trace/impacts/:nodeId` → `tracerHandler.TraceImpacts`
- **Public model endpoints:**
  - `GET /api/models/capacity-nodes` → `tracerHandler.ListCapacityNodes`
  - `POST /api/models/ingest` → `tracerHandler.IngestModels`

---

## Backend Model Definitions

### /backend/internal/model/blueprint_type.go (13 lines)
```go
type BlueprintType struct {
  ID         uint      `gorm:"primaryKey" json:"id"`
  Name       string    `gorm:"uniqueIndex;size:100;not null" json:"name"`
  Slug       string    `gorm:"uniqueIndex;size:100;not null" json:"slug"`
  FolderName string    `gorm:"size:255;not null" json:"folder_name"`
  CreatedAt  time.Time `json:"created_at"`
  UpdatedAt  time.Time `json:"updated_at"`
}
```

### /backend/internal/model/blueprint_node.go (15 lines)
```go
type BlueprintNode struct {
  ID        uint      `gorm:"primaryKey" json:"id"`
  NodeID    string    `gorm:"uniqueIndex;size:255;not null" json:"node_id"`
  Name      string    `gorm:"size:500;not null" json:"name"`
  NodeType  string    `gorm:"index;size:100" json:"node_type"`
  NodeRole  string    `gorm:"size:100" json:"node_role,omitempty"`
  CreatedAt time.Time `json:"created_at"`
  UpdatedAt time.Time `json:"updated_at"`
}
```

### /backend/internal/model/blueprint_edge.go (17 lines)
```go
type BlueprintEdge struct {
  ID              uint      `gorm:"primaryKey" json:"id"`
  BlueprintTypeID uint      `gorm:"uniqueIndex:idx_edge_type_from_to;not null" json:"blueprint_type_id"`
  FromNodeID      uint      `gorm:"uniqueIndex:idx_edge_type_from_to;index;not null" json:"from_node_id"`
  ToNodeID        uint      `gorm:"uniqueIndex:idx_edge_type_from_to;index;not null" json:"to_node_id"`
  CreatedAt       time.Time `json:"created_at"`
  UpdatedAt       time.Time `json:"updated_at"`
  BlueprintType   BlueprintType `gorm:"foreignKey:BlueprintTypeID" json:"blueprint_type,omitempty"`
  FromNode        BlueprintNode `gorm:"foreignKey:FromNodeID" json:"from_node,omitempty"`
  ToNode          BlueprintNode `gorm:"foreignKey:ToNodeID" json:"to_node,omitempty"`
}
```

### /backend/internal/model/capacity_node_type.go (15 lines)
```go
type CapacityNodeType struct {
  ID               uint      `gorm:"primaryKey" json:"id"`
  NodeType         string    `gorm:"uniqueIndex;size:100;not null" json:"node_type"`
  Topology         string    `gorm:"size:100;not null" json:"topology"`
  IsCapacityNode   bool      `gorm:"not null;default:false" json:"is_capacity_node"`
  ActiveConstraint bool      `gorm:"not null;default:false" json:"active_constraint"`
  CreatedAt        time.Time `json:"created_at"`
  UpdatedAt        time.Time `json:"updated_at"`
}
```

### /backend/internal/model/dependency_rule.go (16 lines)
```go
type DependencyRule struct {
  ID                      uint      `gorm:"primaryKey" json:"id"`
  NodeType                string    `gorm:"uniqueIndex:idx_dep_rule_type_pair;index;size:100;not null" json:"node_type"`
  DependencyNodeType      string    `gorm:"uniqueIndex:idx_dep_rule_type_pair;size:100;not null" json:"dependency_node_type"`
  RelationshipType        string    `gorm:"size:20;not null" json:"relationship_type"`
  TopologicalRelationship string    `gorm:"size:20;not null" json:"topological_relationship"`
  UpstreamLevel           *int      `json:"upstream_level"`
  CreatedAt               time.Time `json:"created_at"`
  UpdatedAt               time.Time `json:"updated_at"`
}
```

### /backend/internal/model/impact_rule.go (15 lines)
```go
type ImpactRule struct {
  ID                      uint      `gorm:"primaryKey" json:"id"`
  NodeType                string    `gorm:"uniqueIndex:idx_impact_rule_type_pair;index;size:100;not null" json:"node_type"`
  ImpactNodeType          string    `gorm:"uniqueIndex:idx_impact_rule_type_pair;size:100;not null" json:"impact_node_type"`
  TopologicalRelationship string    `gorm:"size:20;not null" json:"topological_relationship"`
  DownstreamLevel         *int      `json:"downstream_level"`
  CreatedAt               time.Time `json:"created_at"`
  UpdatedAt               time.Time `json:"updated_at"`
}
```

---

## Frontend Components

### /frontend/src/components/tracer/dag-types.ts (48 lines)
**TypeScript Interfaces (match Go JSON):**
- `TracedNode` — API response node with id, node_id, name, node_type, level, parent_node_id
- `TraceLevelGroup` — level, topology, nodes[]
- `TraceLocalGroup` — topology, nodes[] (no level)
- `TraceResponse` — source + upstream[], local[], downstream[], load[]
- `SearchNode` — search result (id, node_id, name, node_type)
- `TracerNodeData extends Record<string, unknown>` — ReactFlow node data (nodeId, name, nodeType, topology, isSource, isLocal, ring, level, onNodeClick callback)

### /frontend/src/components/tracer/dag-node.tsx (59 lines)
**Custom ReactFlow Node Component:**
- `TracerNode({ data })` — renders tracer node with:
  - Topology-based coloring (from dag-helpers config)
  - Source badge (amber ring)
  - Level badge (top-right, for L1+)
  - Type label + node ID + name
  - Dashed border for local nodes
  - Click handler → setPopupData

### /frontend/src/components/tracer/dag-edge.tsx (48 lines)
**Custom ReactFlow Edge Component:**
- `TracerEdge({ id, sourceX, sourceY, targetX, targetY, ... })` — renders:
  - Glow layer (lighter, +4px stroke width)
  - Main edge with arrowhead marker

### /frontend/src/components/tracer/dependency-impact-dag.tsx (204 lines)
**Main DAG Visualization:**
- `DependencyImpactDAG()` — wrapper with ReactFlowProvider
- `DependencyImpactDAGInner()` — main component with:
  - Depth control (1-6, default 2) via Minus/Plus buttons
  - `DAGSearch` for node selection
  - Theme toggle
  - React Query hooks for dependencies & impacts traces
  - `traceToDAGElements()` + `layoutDAG()` to convert API responses to ReactFlow nodes/edges
  - Popup detail view on node click
  - Empty state message
  - Loading overlay

### /frontend/src/components/tracer/dag-detail-popup.tsx (110 lines)
**Node Detail Popup:**
- `DagDetailPopup({ data, onClose })` — modal overlay showing:
  - Topology color bar (header)
  - Icon + node type + node ID + name
  - Role label (Source/Local/Level X)
  - Status pills (SOURCE/LOCAL/LX, topology key, ring)
  - Copyable node ID

### /frontend/src/components/tracer/dag-search.tsx (171 lines)
**Node Search & Filter:**
- `DAGSearch({ onSelect, onClear })` — search bar with:
  - Debounced query (300ms)
  - Type filter dropdown (predefined NODE_TYPES list)
  - React Query search against `/api/blueprints/nodes?search=...&node_type=...`
  - Dropdown results showing node_id, name, node_type badge
  - Clear button (text + icon)

### /frontend/src/components/tracer/dag-helpers.tsx (219 lines)
**Topology Config & Graph Utilities:**
- `TOPOLOGY_CONFIG` — maps topology key → { color, bg, icon }
  - electrical: #F97316 (orange), cooling: #06B6D4 (cyan), spatial: #8B5CF6 (purple), whitespace: #10B981 (green)
- `getTopologyKey(topology: string): string` — returns config key based on topology name
- **Edge/Marker styles:**
  - UPSTREAM_STYLE: cyan, 2.5px
  - DOWNSTREAM_STYLE: orange, 2.5px
  - LOCAL_STYLE: gray, 1.5px dashed
- `makeNode()` — creates ReactFlow node with TracerNodeData
- `traceToDAGElements(depResponse, impactResponse)` — converts API responses to ReactFlow nodes+edges:
  - Source node (ring 0, isSource=true)
  - Upstream deps (blue edges, chained by parent_node_id)
  - Local deps (gray dashed, grouped with source in parent container)
  - Downstream impacts (orange edges, chained from parent)
  - Load impacts (gray dashed edges)
- `layoutDAG(nodes, edges)` — Dagre LR layout:
  - rankdir=LR, ranksep=120, nodesep=50
  - Only layouts top-level nodes (no parentId)

### Frontend Component Totals
- **Individual files:** dag-types (48) + dag-node (59) + dag-edge (48) + dependency-impact-dag (204) + dag-detail-popup (110) + dag-search (171) + dag-helpers (219) = **859 lines**

---

## Test Files

### /backend/internal/service/dependency_tracer_helpers_test.go
- **9 test functions** covering filterByTypes, groupByLevel, groupDepRules, groupImpactRules
- **Lines:** 296

---

## Summary

| Layer | File | Lines | Key Type/Function |
|-------|------|-------|-------------------|
| Service | dependency_tracer.go | 311 | DependencyTracer, TraceDependencies, TraceImpacts |
| Service | dependency_tracer_helpers.go | 69 | groupDepRules, groupImpactRules, filterByTypes |
| Service | dependency_tracer_helpers_test.go | 296 | 9 unit tests |
| Repository | tracer_repository.go | 179 | TracerRepository, FindUpstreamNodes, FindDownstreamNodes, etc. |
| Handler | tracer_handler.go | 97 | TracerHandler, routes: /trace/dependencies, /trace/impacts |
| Router | router.go | 69 | Route setup for /api/trace/* and /api/models/ingest |
| Models | 7 files | ~94 total | BlueprintType, BlueprintNode, BlueprintEdge, CapacityNodeType, DependencyRule, ImpactRule, User |
| Frontend | 7 files | 859 total | DependencyImpactDAG, TracerNode, dag-helpers with Dagre layout |

**Total tracer-specific code: ~1650 lines backend + 859 lines frontend**
