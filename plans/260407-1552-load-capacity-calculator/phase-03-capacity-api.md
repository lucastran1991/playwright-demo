# Phase 3: Capacity API Endpoints

## Context Links
- Brainstorm: `plans/reports/brainstorm-260407-1542-load-capacity-calculator.md`
- Phase 1: `phase-01-data-model-csv-ingestion.md` (model + ingestion)
- Phase 2: `phase-02-load-calculator.md` (computed metrics)
- Existing handler pattern: `backend/internal/handler/tracer_handler.go`
- Existing router: `backend/internal/router/router.go`
- Main wiring: `backend/cmd/server/main.go`

## Overview
- **Priority**: P1
- **Status**: pending
- **Description**: Create capacity HTTP handler with 4 endpoints, extend `/trace/full` to include capacity data, wire everything in main.go and router.

## Key Insights
- Follow exact same handler/router pattern as `TracerHandler` — inject services + repo, use `response.Success/Error`
- The `/trace/full` response is the main integration point for frontend — enriching traced nodes with capacity data is the highest-value API change
- Keep capacity endpoints separate from trace endpoints (different domain concern)
- Capacity data fetched via repo, not re-computed on each request (pre-computed in Phase 2)

## Requirements

### Functional
- `POST /api/capacity/ingest` — trigger CSV ingestion + computation
- `GET /api/capacity/nodes/:nodeId` — single node capacity metrics
- `GET /api/capacity/summary` — aggregate stats across all capacity nodes
- `GET /api/capacity/nodes` — list nodes with capacity data (filter by type, min_utilization)
- Extend `TraceResponse` to include capacity data per traced node

### Non-functional
- Follow existing response format: `{ "data": ... }` wrapper
- Error handling: 404 for unknown node, 500 for DB errors
- No auth required (matches existing public endpoints)

## Architecture

```
capacity_handler.go
  ├── IngestCapacity   POST /api/capacity/ingest
  ├── GetNodeCapacity  GET  /api/capacity/nodes/:nodeId
  ├── GetSummary       GET  /api/capacity/summary
  └── ListCapacityNodes GET /api/capacity/nodes

dependency_tracer.go (modified)
  └── TraceFull enriches response with capacity map
```

## Related Code Files

### Files to CREATE
| File | Est. Lines | Purpose |
|------|-----------|---------|
| `backend/internal/handler/capacity_handler.go` | ~100 | HTTP handlers for capacity endpoints |

### Files to MODIFY
| File | Change |
|------|--------|
| `backend/internal/router/router.go` | Add capacity route group, accept CapacityHandler param |
| `backend/cmd/server/main.go` | Wire CapacityRepository, IngestionService, Calculator, Handler |
| `backend/internal/service/dependency_tracer.go` | Add capacity enrichment to TraceResponse |
| `backend/internal/service/dependency_tracer.go` | Add `CapacityMetrics` field to `SourceNode` or `TraceResponse` |
| `backend/internal/repository/capacity_repository.go` | Add query helpers for summary + filtered list |

## Implementation Steps

### Step 1: Add repository query helpers
File: `backend/internal/repository/capacity_repository.go`

```go
// NodeCapacity is the API-facing capacity summary for a single node.
type NodeCapacity struct {
    NodeID          string             `json:"node_id"`
    NodeType        string             `json:"node_type"`
    Name            string             `json:"name"`
    Capacity        map[string]float64 `json:"capacity"`        // varName -> value
    Units           map[string]string  `json:"units"`           // varName -> unit
}

// GetNodeCapacity returns capacity metrics for a single node.
func (r *CapacityRepository) GetNodeCapacity(nodeID string) (*NodeCapacity, error)

// GetCapacitySummary returns aggregate stats: total nodes, avg utilization, overloaded count.
type CapacitySummary struct {
    TotalNodes      int     `json:"total_nodes"`
    AvgUtilization  float64 `json:"avg_utilization_pct"`
    OverloadedNodes int     `json:"overloaded_nodes"` // utilization > 100%
    HighUtilNodes   int     `json:"high_util_nodes"`  // utilization > 80%
    TotalCapacity   float64 `json:"total_capacity_kw"`
    TotalLoad       float64 `json:"total_load_kw"`
}
func (r *CapacityRepository) GetCapacitySummary() (*CapacitySummary, error)

// ListCapacityNodes returns paginated nodes with capacity data.
func (r *CapacityRepository) ListCapacityNodes(nodeType string, minUtil float64, limit, offset int) ([]NodeCapacity, int64, error)

// GetCapacityMapForNodes returns capacity data for a batch of node IDs (for trace enrichment).
func (r *CapacityRepository) GetCapacityMapForNodes(nodeIDs []string) (map[string]map[string]float64, error)
```

**GetCapacitySummary SQL sketch:**
```sql
SELECT
  COUNT(DISTINCT node_id) as total_nodes,
  AVG(CASE WHEN variable_name = 'utilization_pct' THEN value END) as avg_util,
  COUNT(DISTINCT CASE WHEN variable_name = 'utilization_pct' AND value > 100 THEN node_id END) as overloaded,
  COUNT(DISTINCT CASE WHEN variable_name = 'utilization_pct' AND value > 80 THEN node_id END) as high_util,
  SUM(CASE WHEN variable_name = 'rated_capacity' AND source = 'csv_import' THEN value ELSE 0 END) as total_cap,
  SUM(CASE WHEN variable_name = 'allocated_load' AND source = 'csv_import' THEN value ELSE 0 END) as total_load
FROM node_variables
```

### Step 2: Create capacity handler
File: `backend/internal/handler/capacity_handler.go`

```go
type CapacityHandler struct {
    ingestionService *service.CapacityIngestionService
    repo             *repository.CapacityRepository
    csvPath          string // path to capacity CSV
}

func NewCapacityHandler(
    svc *service.CapacityIngestionService,
    repo *repository.CapacityRepository,
    csvPath string,
) *CapacityHandler
```

**Handlers:**

```go
// IngestCapacity handles POST /api/capacity/ingest
func (h *CapacityHandler) IngestCapacity(c *gin.Context) {
    summary, err := h.ingestionService.IngestCSV(h.csvPath)
    if err != nil {
        response.Error(c, 500, "Capacity ingestion failed: "+err.Error())
        return
    }
    response.Success(c, 200, summary)
}

// GetNodeCapacity handles GET /api/capacity/nodes/:nodeId
func (h *CapacityHandler) GetNodeCapacity(c *gin.Context) {
    nodeID := c.Param("nodeId")
    cap, err := h.repo.GetNodeCapacity(nodeID)
    if err != nil {
        response.Error(c, 404, "No capacity data for node: "+nodeID)
        return
    }
    response.Success(c, 200, cap)
}

// GetSummary handles GET /api/capacity/summary
func (h *CapacityHandler) GetSummary(c *gin.Context) {
    summary, err := h.repo.GetCapacitySummary()
    if err != nil {
        response.Error(c, 500, "Failed to get capacity summary")
        return
    }
    response.Success(c, 200, summary)
}

// ListCapacityNodes handles GET /api/capacity/nodes
func (h *CapacityHandler) ListCapacityNodes(c *gin.Context) {
    nodeType := c.Query("type")
    minUtil := parseFloatParam(c, "min_utilization", 0)
    limit := parseIntParam(c, "limit", 50, 500)
    offset := parseIntParam(c, "offset", 0, 10000)
    
    nodes, total, err := h.repo.ListCapacityNodes(nodeType, minUtil, limit, offset)
    if err != nil {
        response.Error(c, 500, "Failed to list capacity nodes")
        return
    }
    response.Success(c, 200, gin.H{"nodes": nodes, "total": total})
}
```

Note: Reuse `parseIntParam` from `tracer_handler.go` — extract to a shared `handler` package helper or duplicate (2 lines, acceptable).

### Step 3: Update router
File: `backend/internal/router/router.go`

Add `capacityHandler *handler.CapacityHandler` param to `Setup()`.

```go
func Setup(authHandler *handler.AuthHandler, blueprintHandler *handler.BlueprintHandler,
    tracerHandler *handler.TracerHandler, capacityHandler *handler.CapacityHandler,
    jwtSecret, corsOrigin string) *gin.Engine {
    // ... existing routes ...
    
    // Capacity endpoints
    capacity := r.Group("/api/capacity")
    {
        capacity.POST("/ingest", capacityHandler.IngestCapacity)
        capacity.GET("/nodes/:nodeId", capacityHandler.GetNodeCapacity)
        capacity.GET("/nodes", capacityHandler.ListCapacityNodes)
        capacity.GET("/summary", capacityHandler.GetSummary)
    }
}
```

### Step 4: Wire in main.go
File: `backend/cmd/server/main.go`

```go
// After existing wiring...
capacityRepo := repository.NewCapacityRepository(db)
calculator := service.NewLoadCapacityCalculator(capacityRepo, tracerRepo, db)
capacityIngestionSvc := service.NewCapacityIngestionService(capacityRepo, calculator, db)
capacityCSVPath := filepath.Join(cfg.ModelDir, "ISET capacity - rack load flow.csv")
capacityHandler := handler.NewCapacityHandler(capacityIngestionSvc, capacityRepo, capacityCSVPath)

r := router.Setup(authHandler, blueprintHandler, tracerHandler, capacityHandler, cfg.JWTSecret, cfg.CORSOrigin)
```

### Step 5: Enrich /trace/full with capacity data
File: `backend/internal/service/dependency_tracer.go`

**Option A (preferred — minimal change)**: Add `Capacity` field to `TraceResponse`:

```go
type TraceResponse struct {
    Source     SourceNode                `json:"source"`
    Upstream   []TraceLevelGroup         `json:"upstream,omitempty"`
    Local      []TraceLocalGroup         `json:"local,omitempty"`
    Downstream []TraceLevelGroup         `json:"downstream,omitempty"`
    Load       []TraceLocalGroup         `json:"load,omitempty"`
    Capacity   map[string]map[string]float64 `json:"capacity,omitempty"` // nodeID -> {varName -> value}
}
```

**Enrichment in TraceFull:**
```go
func (t *DependencyTracer) TraceFull(nodeID string, maxLevels int) (*TraceResponse, error) {
    // ... existing logic ...
    
    // Enrich with capacity data if capacity repo is available
    if t.capRepo != nil {
        allNodeIDs := collectAllNodeIDs(resp) // source + all upstream/downstream/local/load node_ids
        capMap, err := t.capRepo.GetCapacityMapForNodes(allNodeIDs)
        if err == nil && len(capMap) > 0 {
            resp.Capacity = capMap
        }
    }
    return resp, nil
}
```

This requires injecting `capRepo` into `DependencyTracer`. Add optional field:
```go
type DependencyTracer struct {
    repo       *repository.TracerRepository
    capRepo    *repository.CapacityRepository // optional, nil = no capacity enrichment
    // ... existing fields ...
}
```

Update `NewDependencyTracer` to accept optional `capRepo` (or add a setter method to avoid breaking existing callers):
```go
func (t *DependencyTracer) SetCapacityRepo(repo *repository.CapacityRepository) {
    t.capRepo = repo
}
```

Then in `main.go`:
```go
depTracer := service.NewDependencyTracer(tracerRepo)
depTracer.SetCapacityRepo(capacityRepo) // enable capacity enrichment
```

### Step 6: Helper to collect all node IDs from TraceResponse

```go
func collectAllNodeIDs(resp *TraceResponse) []string {
    seen := map[string]bool{resp.Source.NodeID: true}
    for _, groups := range [][]TraceLevelGroup{resp.Upstream, resp.Downstream} {
        for _, g := range groups {
            for _, n := range g.Nodes {
                seen[n.NodeID] = true
            }
        }
    }
    for _, groups := range [][]TraceLocalGroup{resp.Local, resp.Load} {
        for _, g := range groups {
            for _, n := range g.Nodes {
                seen[n.NodeID] = true
            }
        }
    }
    ids := make([]string, 0, len(seen))
    for id := range seen {
        ids = append(ids, id)
    }
    return ids
}
```

## Todo List
- [ ] Add `GetNodeCapacity`, `GetCapacitySummary`, `ListCapacityNodes`, `GetCapacityMapForNodes` to capacity_repository.go
- [ ] Create `capacity_handler.go` with 4 endpoint handlers
- [ ] Update `router.go` to add capacity route group
- [ ] Wire CapacityRepository, Calculator, IngestionService, Handler in main.go
- [ ] Add `SetCapacityRepo` to DependencyTracer
- [ ] Add `Capacity` field to `TraceResponse` + enrichment in `TraceFull`
- [ ] Add `collectAllNodeIDs` helper
- [ ] Verify `go build` compiles
- [ ] Manual smoke: curl endpoints, verify JSON response shape

## Success Criteria
- `POST /api/capacity/ingest` returns summary with node/variable counts
- `GET /api/capacity/nodes/RPP-R1-1` returns capacity metrics
- `GET /api/capacity/summary` returns aggregate stats
- `GET /api/capacity/nodes?type=Rack&min_utilization=80` filters correctly
- `GET /api/trace/full/RPP-R1-1` response includes `capacity` map for all traced nodes
- Existing trace endpoints still work unchanged (backward compatible)

## Risk Assessment
| Risk | Severity | Mitigation |
|------|----------|-----------|
| Breaking TraceResponse JSON for frontend | Medium | `Capacity` field is `omitempty` — absent when capRepo is nil |
| Router Setup signature change | Low | Update all callers (only main.go) |
| Performance of batch capacity lookup | Low | Single query with `WHERE node_id IN (...)`, indexed |

## Security Considerations
- Ingestion endpoint is public (consistent with existing `/api/blueprints/ingest`)
- No user-controlled file paths — CSV path is server config
- Query params sanitized via GORM parameterized queries
