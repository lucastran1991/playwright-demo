# Phase 5: Testing

## Context Links
- Existing test pattern: `backend/internal/handler/tracer_handler_test.go` (TestMain + real DB)
- Test helpers: `backend/internal/handler/tracer_handler_test_helpers_test.go`
- Test DB setup: `backend/internal/testutil/` (SetupTestDBForMain, TruncateAndSeedForMain)
- CSV parser tests: `backend/internal/service/model_csv_parser_test.go`
- Tracer service tests: `backend/internal/service/dependency_tracer_test.go`

## Overview
- **Priority**: P1
- **Status**: pending
- **Description**: Write integration tests for CSV ingestion, unit tests for aggregation logic, and API tests for capacity endpoints. Use real DB (no mocks) per project convention.

## Key Insights
- Project uses **real test DB** — no mocks. `testutil.SetupTestDBForMain()` provides a DB connection, `TruncateAndSeedForMain` seeds blueprint data.
- Tests use `TestMain` for one-time DB setup to avoid parallel package races (see commit `43cf22e`)
- CSV parser tests are pure functions — no DB needed, fast
- Aggregation tests need seeded blueprint_nodes + blueprint_edges (spatial topology) + node_variables (raw CSV data)
- Spot-check computed values against manual calculations from CSV

## Requirements

### Functional
- CSV parser: correct column mapping per node type, skip empty values, handle edge cases
- Aggregation: RPP allocated_load = SUM(child rack loads), utilization_pct correct
- API: correct response shape, 404 for unknown nodes, filter params work
- Integration: full ingest -> compute -> query round-trip

### Non-functional
- Tests must use real DB (no mocks)
- Tests must be idempotent (truncate + seed before each run)
- No `t.Parallel()` for DB-dependent tests within same package

## Architecture

```
Tests by layer:

service/capacity_csv_parser_test.go      — pure function tests (no DB)
service/load_capacity_calculator_test.go  — integration tests (needs DB + seeded data)
handler/capacity_handler_test.go          — API tests (HTTP round-trip, needs DB)
```

## Related Code Files

### Files to CREATE
| File | Est. Lines | Purpose |
|------|-----------|---------|
| `backend/internal/service/capacity_csv_parser_test.go` | ~80 | Parser unit tests |
| `backend/internal/service/load_capacity_calculator_test.go` | ~120 | Aggregation integration tests |
| `backend/internal/handler/capacity_handler_test.go` | ~100 | API endpoint tests |

### Files to MODIFY
| File | Change |
|------|--------|
| `backend/internal/testutil/` | May need to add capacity-specific seed data helper |

## Implementation Steps

### Step 1: CSV Parser Tests (no DB)
File: `backend/internal/service/capacity_csv_parser_test.go`

Use the actual CSV file for integration-style parser tests:

```go
func TestParseCapacityFlowCSV_RealFile(t *testing.T) {
    // Use the real CSV file
    rows, err := ParseCapacityFlowCSV("../../blueprint/ISET capacity - rack load flow.csv")
    if err != nil {
        t.Skipf("CSV not available: %v", err)
    }
    
    if len(rows) != 657 {
        t.Errorf("expected 657 rows, got %d", len(rows))
    }
    
    // Spot-check a known Rack node
    // Find RACK-R1-Z1-R1-01 (or similar known ID from CSV)
    var rackRow *CapacityFlowRow
    for i := range rows {
        if rows[i].NodeID == "RACK-R1-Z1-R1-01" {
            rackRow = &rows[i]
            break
        }
    }
    if rackRow == nil {
        t.Fatal("expected to find RACK-R1-Z1-R1-01")
    }
    
    // Check it has design_capacity, rated_capacity, allocated_load
    varMap := make(map[string]float64)
    for _, v := range rackRow.Variables {
        varMap[v.VariableName] = v.Value
    }
    if _, ok := varMap["design_capacity"]; !ok {
        t.Error("Rack missing design_capacity")
    }
    if _, ok := varMap["rated_capacity"]; !ok {
        t.Error("Rack missing rated_capacity")
    }
    if _, ok := varMap["allocated_load"]; !ok {
        t.Error("Rack missing allocated_load")
    }
}

func TestParseCapacityFlowCSV_EmptyValues(t *testing.T) {
    // Verify nodes only get variables for non-empty CSV columns
    rows, err := ParseCapacityFlowCSV("../../blueprint/ISET capacity - rack load flow.csv")
    if err != nil {
        t.Skipf("CSV not available: %v", err)
    }
    
    // ACU nodes should NOT have Rack_Circuit_Capacity fields
    for _, row := range rows {
        if row.NodeType == "Air Cooling Unit" {
            for _, v := range row.Variables {
                if v.VariableName == "design_capacity" {
                    // ACU gets rated_capacity from Rated_Cooling_Capacity, not design
                    // Verify it's from the correct column
                }
            }
            break
        }
    }
}

func TestParseCapacityFlowCSV_NodeTypeCounts(t *testing.T) {
    rows, _ := ParseCapacityFlowCSV("../../blueprint/ISET capacity - rack load flow.csv")
    typeCounts := make(map[string]int)
    for _, r := range rows {
        typeCounts[r.NodeType]++
    }
    // Verify known distribution from brainstorm
    if typeCounts["Rack"] != 264 {
        t.Errorf("expected 264 Rack nodes, got %d", typeCounts["Rack"])
    }
    if typeCounts["RPP"] != 24 {
        t.Errorf("expected 24 RPP nodes, got %d", typeCounts["RPP"])
    }
}
```

### Step 2: Aggregation Integration Tests (needs DB)
File: `backend/internal/service/load_capacity_calculator_test.go`

Uses `TestMain` pattern. Seeds a small subset of known data, then verifies computed values.

```go
var testDB *gorm.DB

func TestMain(m *testing.M) {
    db, cleanup := testutil.SetupTestDBForMain()
    if db == nil {
        os.Exit(0)
    }
    defer cleanup()
    testDB = db
    
    // Seed: create blueprint nodes, spatial edges, and raw node_variables
    seedCapacityTestData(db)
    
    os.Exit(m.Run())
}

func seedCapacityTestData(db *gorm.DB) {
    // Create a mini topology:
    // RPP-TEST -> (spatial children) -> RACK-TEST-01, RACK-TEST-02
    // RACK-TEST-01: rated_capacity=18, allocated_load=15
    // RACK-TEST-02: rated_capacity=18, allocated_load=12
    // RPP-TEST: rated_capacity=288
    //
    // Expected:
    //   RPP-TEST allocated_load = 15+12 = 27
    //   RPP-TEST utilization = 27/288 = 9.375%
    //   RPP-TEST available = 288-27 = 261
    
    // Insert blueprint_nodes, blueprint_types, blueprint_edges (spatial)
    // Insert node_variables with source=csv_import
}

func TestComputeAll_RackMetrics(t *testing.T) {
    capRepo := repository.NewCapacityRepository(testDB)
    tracerRepo := repository.NewTracerRepository(testDB)
    calc := NewLoadCapacityCalculator(capRepo, tracerRepo, testDB)
    
    summary, err := calc.ComputeAll()
    if err != nil {
        t.Fatalf("ComputeAll failed: %v", err)
    }
    if summary.VariablesComputed == 0 {
        t.Error("expected computed variables > 0")
    }
    
    // Check RACK-TEST-01 metrics
    vars, _ := capRepo.GetNodeVariables("RACK-TEST-01")
    varMap := make(map[string]float64)
    for _, v := range vars {
        varMap[v.VariableName] = v.Value
    }
    
    if varMap["available_capacity"] != 3 { // 18 - 15
        t.Errorf("RACK-TEST-01 available_capacity: want 3, got %f", varMap["available_capacity"])
    }
    utilPct := varMap["utilization_pct"]
    if utilPct < 83.0 || utilPct > 84.0 { // 15/18 * 100 = 83.33
        t.Errorf("RACK-TEST-01 utilization_pct: want ~83.33, got %f", utilPct)
    }
}

func TestComputeAll_RPPAggregation(t *testing.T) {
    capRepo := repository.NewCapacityRepository(testDB)
    
    vars, _ := capRepo.GetNodeVariables("RPP-TEST")
    varMap := make(map[string]float64)
    for _, v := range vars {
        if v.Source == "computed" {
            varMap[v.VariableName] = v.Value
        }
    }
    
    if varMap["allocated_load"] != 27 { // 15 + 12
        t.Errorf("RPP-TEST allocated_load: want 27, got %f", varMap["allocated_load"])
    }
    if varMap["available_capacity"] != 261 { // 288 - 27
        t.Errorf("RPP-TEST available_capacity: want 261, got %f", varMap["available_capacity"])
    }
}

func TestComputeAll_Idempotent(t *testing.T) {
    capRepo := repository.NewCapacityRepository(testDB)
    tracerRepo := repository.NewTracerRepository(testDB)
    calc := NewLoadCapacityCalculator(capRepo, tracerRepo, testDB)
    
    // Run twice
    calc.ComputeAll()
    calc.ComputeAll()
    
    // Check no duplicates — count computed variables for a single node
    var count int64
    testDB.Model(&model.NodeVariable{}).
        Where("node_id = ? AND source = ?", "RPP-TEST", "computed").
        Count(&count)
    
    if count > 3 { // allocated_load, available_capacity, utilization_pct
        t.Errorf("expected <= 3 computed vars for RPP-TEST, got %d (duplicates?)", count)
    }
}
```

### Step 3: API Endpoint Tests
File: `backend/internal/handler/capacity_handler_test.go`

Follow `tracer_handler_test.go` pattern: TestMain sets up router, tests make HTTP requests.

```go
var testCapRouter *gin.Engine

func TestMain(m *testing.M) {
    gin.SetMode(gin.TestMode)
    db, cleanup := testutil.SetupTestDBForMain()
    if db == nil {
        os.Exit(0)
    }
    defer cleanup()
    
    // Seed data + run computation
    seedAndCompute(db)
    
    capRepo := repository.NewCapacityRepository(db)
    h := NewCapacityHandler(nil, capRepo, "") // nil ingestion svc (not testing ingest)
    
    r := gin.New()
    cap := r.Group("/api/capacity")
    cap.GET("/nodes/:nodeId", h.GetNodeCapacity)
    cap.GET("/nodes", h.ListCapacityNodes)
    cap.GET("/summary", h.GetSummary)
    testCapRouter = r
    
    os.Exit(m.Run())
}

func TestGetNodeCapacity_Found(t *testing.T) {
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/api/capacity/nodes/RACK-TEST-01", nil)
    testCapRouter.ServeHTTP(w, req)
    
    if w.Code != 200 {
        t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
    }
    // Parse and verify capacity fields present
}

func TestGetNodeCapacity_NotFound(t *testing.T) {
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/api/capacity/nodes/NONEXISTENT", nil)
    testCapRouter.ServeHTTP(w, req)
    
    if w.Code != 404 {
        t.Fatalf("expected 404, got %d", w.Code)
    }
}

func TestGetSummary(t *testing.T) {
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/api/capacity/summary", nil)
    testCapRouter.ServeHTTP(w, req)
    
    if w.Code != 200 {
        t.Fatalf("expected 200, got %d", w.Code)
    }
    // Verify total_nodes > 0, avg_utilization > 0
}

func TestListCapacityNodes_FilterByType(t *testing.T) {
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/api/capacity/nodes?type=Rack", nil)
    testCapRouter.ServeHTTP(w, req)
    
    if w.Code != 200 {
        t.Fatalf("expected 200, got %d", w.Code)
    }
    // Verify all returned nodes are Rack type
}
```

### Step 4: Spot-check against manual CSV calculation
Add a test that reads the real CSV and manually computes what a known RPP's allocated_load should be, then verifies the calculator produces the same result.

```go
func TestSpotCheck_RPP_Against_CSV(t *testing.T) {
    // This test requires full blueprint + capacity data ingested
    // Run against real DB with full data (integration test)
    
    // 1. Find all Rack nodes that are spatial children of RPP-R1-1
    // 2. Sum their Allocated_ITLoad from CSV
    // 3. Compare with computed allocated_load in node_variables
    // 4. Verify they match
}
```

## Todo List
- [ ] Create `capacity_csv_parser_test.go` with parser unit tests
- [ ] Create `load_capacity_calculator_test.go` with aggregation integration tests
- [ ] Create `capacity_handler_test.go` with API endpoint tests
- [ ] Add test seed data helper for capacity-specific topology
- [ ] Spot-check: RPP aggregation matches manual CSV sum
- [ ] Spot-check: UPS aggregation covers all downstream Racks
- [ ] Verify idempotency: double-run produces same results
- [ ] Run `go test ./...` — all pass
- [ ] Run existing tracer tests — no regressions

## Success Criteria
- All 657 CSV rows parsed correctly (parser test)
- Rack available_capacity = rated - allocated (calculator test)
- RPP allocated_load = SUM(child rack loads) (aggregation test)
- Idempotent: running ComputeAll twice = same result
- API returns correct 200/404 responses
- Existing tracer tests still pass (no regression)
- `go test ./...` exits 0

## Risk Assessment
| Risk | Severity | Mitigation |
|------|----------|-----------|
| Test DB not available in CI | Medium | Tests skip gracefully with `t.Skipf` if DB unavailable |
| Seed data inconsistent with production | Low | Use known fixed node IDs; document expected values |
| Parallel test package conflicts | Medium | Use TestMain pattern per existing convention |

## Next Steps
- After all tests pass, the feature is ready for code review
- Frontend visual QA: manually verify DAG badges render correctly
