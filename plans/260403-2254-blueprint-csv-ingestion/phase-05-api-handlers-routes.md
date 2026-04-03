# Phase 5: API Handlers & Routes

## Context Links
- [Plan Overview](plan.md)
- [Phase 4: Ingestion Service](phase-04-ingestion-service.md)
- Existing: `backend/internal/handler/auth_handler.go`, `backend/internal/router/router.go`

## Overview
- **Priority**: P2
- **Status**: completed
- **Description**: REST endpoints for triggering ingestion and querying blueprint data

## Key Insights
- Follow existing handler pattern: struct with service dependency, methods per endpoint
- Use `response.Success()` and `response.Error()` from `pkg/response`
- Blueprint base path should come from config (env var `BLUEPRINT_DIR`)
- Tree endpoint uses recursive CTE via repository

## Requirements

### Functional
- `POST /api/blueprints/ingest` -- trigger full ingestion, return summary
- `GET /api/blueprints/types` -- list all blueprint domains
- `GET /api/blueprints/nodes` -- list nodes, optional `?type=slug` filter
- `GET /api/blueprints/nodes/:nodeId` -- single node with memberships
- `GET /api/blueprints/edges` -- edges, required `?type=slug` filter
- `GET /api/blueprints/tree/:typeSlug` -- recursive tree for a domain

### Non-functional
- Handler file under 200 lines
- Pagination on list endpoints (limit/offset, default limit=100)
- Consistent error responses using `pkg/response`

## Architecture

```
blueprint_handler.go
  ├── Ingest()       POST /api/blueprints/ingest
  ├── ListTypes()    GET  /api/blueprints/types
  ├── ListNodes()    GET  /api/blueprints/nodes
  ├── GetNode()      GET  /api/blueprints/nodes/:nodeId
  ├── ListEdges()    GET  /api/blueprints/edges
  └── GetTree()      GET  /api/blueprints/tree/:typeSlug
```

## Related Code Files

### Files to Create
- `backend/internal/handler/blueprint_handler.go`

### Files to Modify
- `backend/internal/router/router.go` -- add blueprint routes
- `backend/cmd/server/main.go` -- wire BlueprintHandler dependencies
- `backend/internal/config/config.go` -- add BlueprintDir field

## Implementation Steps

1. Add `BlueprintDir` to config:
   ```go
   // in config.go
   BlueprintDir string `env:"BLUEPRINT_DIR" envDefault:"./blueprint/Node & Edge"`
   ```

2. Create `blueprint_handler.go`:
   ```go
   type BlueprintHandler struct {
       ingestionService *service.BlueprintIngestionService
       repo             *repository.BlueprintRepository
       blueprintDir     string
   }
   
   func NewBlueprintHandler(svc *service.BlueprintIngestionService, repo *repository.BlueprintRepository, dir string) *BlueprintHandler
   ```

3. Implement handler methods:
   - `Ingest(c *gin.Context)`: call `ingestionService.IngestAll(h.blueprintDir)`, return summary
   - `ListTypes(c *gin.Context)`: query all BlueprintTypes, return list
   - `ListNodes(c *gin.Context)`: optional `type` query param, paginate
   - `GetNode(c *gin.Context)`: param `:nodeId` (string node_id), preload memberships
   - `ListEdges(c *gin.Context)`: required `type` query param, paginate
   - `GetTree(c *gin.Context)`: param `:typeSlug`, call repo.GetTree()

4. Register routes in `router.go`:
   ```go
   blueprints := r.Group("/api/blueprints")
   {
       blueprints.POST("/ingest", blueprintHandler.Ingest)
       blueprints.GET("/types", blueprintHandler.ListTypes)
       blueprints.GET("/nodes", blueprintHandler.ListNodes)
       blueprints.GET("/nodes/:nodeId", blueprintHandler.GetNode)
       blueprints.GET("/edges", blueprintHandler.ListEdges)
       blueprints.GET("/tree/:typeSlug", blueprintHandler.GetTree)
   }
   ```

5. Wire in `main.go`:
   ```go
   blueprintRepo := repository.NewBlueprintRepository(db)
   blueprintParser := service.NewBlueprintCSVParser()
   ingestionSvc := service.NewBlueprintIngestionService(blueprintRepo, blueprintParser)
   blueprintHandler := handler.NewBlueprintHandler(ingestionSvc, blueprintRepo, cfg.BlueprintDir)
   
   r := router.Setup(authHandler, blueprintHandler, cfg.JWTSecret)
   ```

6. Update `router.Setup` signature to accept `*handler.BlueprintHandler`

## Todo List
- [x] Add BlueprintDir to config.go
- [x] Create blueprint_handler.go with all 6 methods
- [x] Update router.go with blueprint routes
- [x] Update main.go with dependency wiring
- [x] Test endpoints with curl/httpie
- [x] Verify ingestion returns correct summary

## Success Criteria
- POST /api/blueprints/ingest returns summary JSON
- GET endpoints return correct filtered data
- Tree endpoint returns nested hierarchy
- Pagination works on list endpoints
- Error responses follow existing convention

## Risk Assessment
- **Low**: Long-running ingestion blocks HTTP response -- acceptable at current data volume (<1s expected)
- **Mitigation**: If ever slow, add async job queue (YAGNI for now)

## Security Considerations
- Ingest endpoint should be protected (admin only) or rate-limited
- For MVP: no auth required on blueprint endpoints (matches current app setup where only `/api/auth/me` is protected)
- Future: add role-based access control
