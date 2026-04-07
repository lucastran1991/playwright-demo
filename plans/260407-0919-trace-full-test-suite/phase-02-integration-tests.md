---
phase: 2
title: "Integration Tests — httptest + Real DB"
status: complete
priority: P1
effort: 2.5h
depends_on: [phase-01]
---

# Phase 2: Integration Tests (httptest + Real DB)

## Context Links
- [TracerHandler](../../backend/internal/handler/tracer_handler.go) — HTTP layer
- [Router setup](../../backend/internal/router/router.go) — route registration
- [Response format](../../backend/pkg/response/response.go) — `{"data": ...}` / `{"error": ...}`
- [Phase 1 — test infra](phase-01-test-infrastructure.md)

## Overview
Test the `/api/trace/full/:nodeId` endpoint end-to-end via `net/http/httptest`. Real PostgreSQL, real router, real service. Validates HTTP status codes, response schema, and data correctness against known fixture topology.

## Key Insights
- Response envelope: `{"data": TraceResponse}` on success, `{"error": "msg"}` on failure
- `parseIntParam` defaults `levels` to 2, max 10
- TraceFull merges TraceDependencies + TraceImpacts; handles partial failures
- Source node always present in response, even when no upstream/downstream
- Handler checks `err.Error()` contains "node not found" for 404 routing

## Requirements

### Functional
- Test happy path: known node returns correct upstream/downstream/local/load sections
- Test 404: unknown nodeId returns `{"error": "node not found: ..."}`
- Test levels param: `?levels=1` limits recursion depth
- Test default levels: omitted param defaults to 2
- Test max levels cap: `?levels=99` capped to 10
- Test response schema structure matches frontend expectations

### Non-functional
- Tests must clean up between runs (truncate + re-seed)
- Tests must skip gracefully if DB unavailable
- Each test function focused on one behavior

## Test Setup Pattern

```go
// In TestMain or per-test setup:
// 1. SetupTestDB (from testutil)
// 2. Wire: TracerRepository -> DependencyTracer -> TracerHandler -> router.Setup
// 3. SeedTraceFixtures
// 4. Create httptest.Server from router
// 5. Run tests against server URL
// 6. Cleanup
```

Router wiring needs dummy auth handler + blueprint handler. Create minimal stubs (nil is fine for handlers not under test — Gin will panic only if those routes are hit).

**Simpler approach**: Build a minimal `gin.Engine` with just the trace routes instead of calling `router.Setup`. Avoids needing auth/blueprint handler stubs.

```go
func setupTestRouter(tracerHandler *handler.TracerHandler) *gin.Engine {
    gin.SetMode(gin.TestMode)
    r := gin.New()
    trace := r.Group("/api/trace")
    trace.GET("/full/:nodeId", tracerHandler.TraceFull)
    trace.GET("/dependencies/:nodeId", tracerHandler.TraceDependencies)
    trace.GET("/impacts/:nodeId", tracerHandler.TraceImpacts)
    return r
}
```

## Test Cases

### File: `backend/internal/handler/tracer_handler_test.go`

Split into two files if exceeding 200 lines.

#### 1. `TestTraceFull_HappyPath_RPP` (~30 lines)
- Source: `RPP-01` (has both dependency + impact rules)
- Seed: RACKPDU-01 depends on RPP-01 (downstream), RPP-01 depends on UPS-01 (upstream)
- Assert:
  - Status 200
  - `data.source.node_id` == "RPP-01"
  - `data.source.node_type` == "RPP"
  - `data.downstream` is non-empty (contains RACKPDU-01)
  - `data.upstream` is non-empty (contains UPS-01)
  - `data.source.topology` == "Electrical System"

#### 2. `TestTraceFull_HappyPath_RackPDU` (~30 lines)
- Source: `RACKPDU-01` (has upstream deps: RPP, UPS)
- Assert:
  - Status 200
  - `data.upstream` contains RPP-01 at level 1
  - `data.local` contains RDHx-01 (cooling local dep) — or empty if no local rule for RackPDU
  - `data.downstream` may be empty (RackPDU is leaf in electrical)

#### 3. `TestTraceFull_NotFound` (~15 lines)
- Source: `NONEXISTENT-99`
- Assert:
  - Status 404
  - `error` field contains "node not found"

#### 4. `TestTraceFull_LevelsParam` (~25 lines)
- Source: `RACKPDU-01` with `?levels=1`
- Assert:
  - Only level-1 upstream nodes returned (RPP-01 yes, UPS-01/GEN-01 no)
- Source: `RACKPDU-01` with `?levels=3`
- Assert:
  - Level-2 upstream nodes included (UPS-01)

#### 5. `TestTraceFull_DefaultLevels` (~20 lines)
- Source: `RACKPDU-01` (no levels param)
- Assert:
  - Levels defaults to 2 (RPP-01 at L1 + UPS-01 at L2)

#### 6. `TestTraceFull_LevelsCapped` (~15 lines)
- Source: `RACKPDU-01` with `?levels=99`
- parseIntParam caps at 10, verify no panic, valid response

#### 7. `TestTraceFull_ResponseSchema` (~30 lines)
- Unmarshal full response into typed struct
- Verify JSON field names match frontend expectations:
  - `source`, `upstream`, `downstream`, `local`, `load`
  - Each TraceLevelGroup has `level`, `topology`, `nodes`
  - Each TracedNode has `id`, `node_id`, `name`, `node_type`, `level`, `parent_node_id`

#### 8. `TestTraceFull_EmptyUpstream` (~15 lines)
- Source: `GEN-01` (Generator — top of electrical chain, no further upstream)
- Assert: `data.upstream` is empty/null

#### 9. `TestTraceFull_LoadSection_CapacityNode` (~25 lines)
- Source: `RPP-01` (IsCapacityNode=true)
- Impact rules include Load for Rack
- Assert: `data.load` is present and contains Rack nodes
- Validates 3-strategy load collection works via spatial edges

#### 10. `TestTraceFull_LoadSection_NonCapacityNode` (~15 lines)
- Source: `RACKPDU-01` (Rack PDU, IsCapacityNode=false)
- Assert: `data.load` is empty/null (Load only for capacity nodes)

## JSON Parsing Strategy

Use `encoding/json` to unmarshal into typed structs matching the API response:

```go
type apiResponse struct {
    Data  *traceData `json:"data"`
    Error string     `json:"error"`
}

type traceData struct {
    Source     sourceNode       `json:"source"`
    Upstream   []levelGroup    `json:"upstream"`
    Local      []localGroup    `json:"local"`
    Downstream []levelGroup    `json:"downstream"`
    Load       []localGroup    `json:"load"`
}

type sourceNode struct {
    NodeID   string `json:"node_id"`
    Name     string `json:"name"`
    NodeType string `json:"node_type"`
    Topology string `json:"topology"`
}

type levelGroup struct {
    Level    int          `json:"level"`
    Topology string      `json:"topology"`
    Nodes    []tracedNode `json:"nodes"`
}

type localGroup struct {
    Topology string      `json:"topology"`
    Nodes    []tracedNode `json:"nodes"`
}

type tracedNode struct {
    ID           uint    `json:"id"`
    NodeID       string  `json:"node_id"`
    Name         string  `json:"name"`
    NodeType     string  `json:"node_type"`
    Level        int     `json:"level"`
    ParentNodeID *string `json:"parent_node_id"`
}
```

## Related Code Files

### Files to Create
- `backend/internal/handler/tracer_handler_test.go` — integration tests

### Files NOT to Modify
- All existing code files remain untouched

## Implementation Steps

1. Create `tracer_handler_test.go` in `handler` package
2. Add test response structs (reusable across tests)
3. Implement `TestMain` or shared setup function:
   - Call `testutil.SetupTestDB`
   - Wire TracerRepository -> DependencyTracer -> TracerHandler
   - Build minimal gin router with trace routes only
   - Seed fixtures
4. Implement each test case as a separate `TestXxx` function
5. Use `httptest.NewRecorder` + `router.ServeHTTP` (no actual server needed)
6. Parse JSON response, assert fields
7. If file exceeds 200 lines, extract response structs + helpers into `tracer_handler_test_helpers_test.go`

## Helper Functions to Create

```go
// doTraceRequest performs GET /api/trace/full/:nodeId and returns parsed response.
func doTraceRequest(t *testing.T, router *gin.Engine, nodeID string, queryParams string) (*httptest.ResponseRecorder, apiResponse)

// findNodeInGroups searches level groups for a node by node_id.
func findNodeInGroups(groups []levelGroup, nodeID string) *tracedNode

// findNodeInLocalGroups searches local groups for a node by node_id.
func findNodeInLocalGroups(groups []localGroup, nodeID string) *tracedNode
```

## Todo List
- [ ] Create test response struct types
- [ ] Implement test setup (DB + wiring + router + fixtures)
- [ ] Test: TraceFull happy path RPP
- [ ] Test: TraceFull happy path RackPDU
- [ ] Test: TraceFull 404 not found
- [ ] Test: levels param filtering
- [ ] Test: default levels
- [ ] Test: levels cap at 10
- [ ] Test: response schema validation
- [ ] Test: empty upstream (Generator)
- [ ] Test: load section for capacity node
- [ ] Test: no load section for non-capacity node
- [ ] Verify all tests pass: `go test -v ./internal/handler/ -run TestTraceFull`
- [ ] Check file size, split if >200 lines

## Success Criteria
- All 10 test cases pass with real PostgreSQL
- Tests skip cleanly when DB unavailable
- Response schema matches what frontend consumes
- No test data leaks between test functions
- File(s) under 200 lines each

## Risk Assessment
- **Fixture data doesn't trigger bridge fallback**: The spatial edges (RACK-01 -> RACKPDU-01 in spatial) should enable bridge path. If not, add more spatial edges.
- **Test ordering**: Each test seeds fresh data via truncate + re-seed, so order doesn't matter.
- **Gin test mode suppresses logs**: Set `gin.SetMode(gin.TestMode)` in setup.

## Next Steps
- Phase 3 adds service-level unit tests for TraceFull composition logic
