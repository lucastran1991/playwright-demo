# Phase 5: API Handlers + Routes + Wiring

## Context Links
- Existing handler pattern: `backend/internal/handler/blueprint_handler.go`
- Router: `backend/internal/router/router.go`
- Wiring: `backend/cmd/server/main.go`
- Response helpers: `backend/pkg/response/response.go`

## Overview
- **Priority**: P1
- **Status**: completed
- **Description**: HTTP handlers for model ingestion + trace endpoints, route registration, dependency wiring in main.go

## Key Insights
- Follow exact same patterns as BlueprintHandler
- Model ingestion endpoint is protected (requires auth)
- Trace + capacity-nodes endpoints are public reads
- TracerHandler needs both ModelIngestionService and DependencyTracer

## Requirements

### Functional
- `POST /api/models/ingest` -- trigger ingestion of 3 model CSVs (protected)
- `GET /api/models/capacity-nodes` -- list all capacity node types (public)
- `GET /api/trace/dependencies/:nodeId` -- trace upstream dependencies (public)
- `GET /api/trace/impacts/:nodeId` -- trace downstream impacts (public)

### Non-functional
- Standard error responses using `response.Error()`
- Query param parsing with sensible defaults
- Handler file under 200 lines

## Architecture

### Endpoint Details

#### POST /api/models/ingest
- Auth: required (JWT)
- Body: none (reads from configured directory)
- Response: `ModelIngestionSummary`
- Errors: 500 if ingestion fails

#### GET /api/models/capacity-nodes
- Auth: none
- Response: `[]CapacityNodeType`

#### GET /api/trace/dependencies/:nodeId
- Auth: none
- Params:
  - `:nodeId` (path) -- blueprint node_id string (e.g. "RACK-R1-Z1-R1-01")
  - `levels` (query) -- max upstream levels, default 2, max 10
  - `include_local` (query) -- "true"/"false", default "true"
- Response: `TraceResponse`
- Errors: 404 if node not found, 400 if invalid params

#### GET /api/trace/impacts/:nodeId
- Auth: none
- Params:
  - `:nodeId` (path) -- blueprint node_id string
  - `levels` (query) -- max downstream levels, default 2, max 10
  - `load_scope` (query) -- "rack"|"row"|"zone"|"room", default "" (all)
- Response: `TraceResponse`
- Errors: 404 if node not found

## Related Code Files

### Files to Create
- `backend/internal/handler/tracer_handler.go`

### Files to Modify
- `backend/internal/router/router.go` -- add model + trace route groups
- `backend/cmd/server/main.go` -- wire TracerRepository, DependencyTracer, ModelIngestionService, TracerHandler
- `backend/internal/config/config.go` -- add ModelDir field

## Implementation Steps

### 1. Create `backend/internal/handler/tracer_handler.go`

```go
type TracerHandler struct {
    ingestionService *service.ModelIngestionService
    tracer           *service.DependencyTracer
    repo             *repository.TracerRepository
    modelDir         string
}

func NewTracerHandler(
    svc *service.ModelIngestionService,
    tracer *service.DependencyTracer,
    repo *repository.TracerRepository,
    modelDir string,
) *TracerHandler
```

Methods:
- `IngestModels(c *gin.Context)` -- call ingestionService.IngestAll(modelDir)
- `ListCapacityNodes(c *gin.Context)` -- call repo.ListCapacityNodeTypes()
- `TraceDependencies(c *gin.Context)` -- parse params, call tracer.TraceDependencies()
- `TraceImpacts(c *gin.Context)` -- parse params, call tracer.TraceImpacts()

Query param parsing:
```go
func parseIntParam(c *gin.Context, key string, defaultVal, maxVal int) int {
    if v, err := strconv.Atoi(c.Query(key)); err == nil && v > 0 {
        if v > maxVal { return maxVal }
        return v
    }
    return defaultVal
}
```

### 2. Update `backend/internal/router/router.go`

Add TracerHandler parameter to `Setup()`:
```go
func Setup(authHandler, blueprintHandler, tracerHandler, jwtSecret) *gin.Engine
```

Add route groups:
```go
// Public model read endpoints
models := r.Group("/api/models")
{
    models.GET("/capacity-nodes", tracerHandler.ListCapacityNodes)
}

// Public trace endpoints
trace := r.Group("/api/trace")
{
    trace.GET("/dependencies/:nodeId", tracerHandler.TraceDependencies)
    trace.GET("/impacts/:nodeId", tracerHandler.TraceImpacts)
}

// Add to protected group:
protected.POST("/models/ingest", tracerHandler.IngestModels)
```

### 3. Update `backend/cmd/server/main.go`

Add after blueprint wiring:
```go
tracerRepo := repository.NewTracerRepository(db)
modelIngestionSvc := service.NewModelIngestionService(db)
depTracer := service.NewDependencyTracer(tracerRepo)
tracerHandler := handler.NewTracerHandler(modelIngestionSvc, depTracer, tracerRepo, cfg.ModelDir)

r := router.Setup(authHandler, blueprintHandler, tracerHandler, cfg.JWTSecret)
```

### 4. Update `backend/internal/config/config.go`

Add field:
```go
ModelDir string
```

In Load():
```go
ModelDir: getEnv("MODEL_DIR", "./blueprint"),
```

## Todo List
- [x] Add ModelDir to config.go
- [x] Create tracer_handler.go with 4 handler methods
- [x] Update router.go -- add tracerHandler param + routes
- [x] Update main.go -- wire all new dependencies
- [x] Verify `go build` compiles
- [x] Test endpoints with curl

## Success Criteria
- All 4 endpoints respond correctly
- POST /models/ingest requires auth token
- GET endpoints return proper JSON responses
- Invalid nodeId returns 404
- Query params have sensible defaults

## Risk Assessment
- **Low**: follows established handler/router pattern exactly
- **Low**: router.Setup signature change is backward-compatible addition

## Next Steps
- Phase 6: write tests for the full pipeline
