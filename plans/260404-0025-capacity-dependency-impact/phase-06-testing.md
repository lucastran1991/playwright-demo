# Phase 6: Testing

## Context Links
- Existing test pattern: `backend/internal/model/blueprint_type_test.go`
- All phase files in `plans/260404-0025-capacity-dependency-impact/`

## Overview
- **Priority**: P2
- **Status**: completed
- **Description**: Unit tests for CSV parsing, integration tests for ingestion + tracer

## Key Insights
- CSV parser tests are pure functions -- easy to unit test with fixture data
- Ingestion + tracer tests need a real database (per project rules: no mocks)
- Keep test files focused: one test file per source file

## Requirements

### Functional
- Test all 3 CSV parsers with valid + malformed input
- Test ingestion idempotency (run twice, same result)
- Test upstream/downstream CTE queries with known topology
- Test handler responses (status codes, response shape)

### Non-functional
- Use real PostgreSQL for integration tests
- No mocks per project rules
- Each test file under 200 lines

## Related Code Files

### Files to Create
- `backend/internal/service/model_csv_parser_test.go`
- `backend/internal/service/model_ingestion_service_test.go` (if DB available)
- `backend/internal/service/dependency_tracer_test.go` (if DB available)

## Implementation Steps

### 1. CSV Parser Tests (`model_csv_parser_test.go`)

Test cases for each parser:
- **ParseCapacityNodesCSV**:
  - Valid file: correct row count, bool parsing
  - "True"/"False"/"true"/"TRUE" all parse correctly
  - Empty NodeType rows skipped
  - Wrong column count returns error
- **ParseDependenciesCSV**:
  - Valid file: correct row count, nullable int parsing
  - Empty UpstreamLevel -> nil
  - "1" -> *int(1)
  - Empty rows skipped
- **ParseImpactsCSV**:
  - Valid file: correct row count
  - Empty DownstreamLevel -> nil

Use `t.TempDir()` + write test CSV files inline.

### 2. Integration Tests (require DB)

Use build tag `//go:build integration` to separate from unit tests.

**Ingestion test**:
- Connect to test DB
- Run IngestAll with real CSV files
- Assert row counts match expected
- Run again, assert same counts (idempotent)

**Tracer test**:
- Requires blueprint data + model data ingested
- Trace a known rack node
- Assert upstream results contain expected node types at expected levels

### 3. Test fixture CSV data

Create small inline CSVs for unit tests:
```go
capacityCSV := `Node Type,Topology,Capacity Node (Capacity Domain),ActiveConstraint
Rack PDU,Electrical System,False,False
RPP,Electrical System,True,True`
```

## Todo List
- [x] Create model_csv_parser_test.go with unit tests
- [x] Create test fixtures (inline CSV strings)
- [x] Create integration test files (build-tagged)
- [x] Run `go test ./...` and verify all pass

## Success Criteria
- All CSV parser unit tests pass
- Integration tests pass against real DB (when available)
- `go test ./internal/service/...` exits 0
- No mocked data

## Risk Assessment
- **Low**: CSV parser tests are straightforward pure-function tests
- **Medium**: integration tests require DB setup -- may need test container or CI config
- **Low**: test fixture data is small and deterministic

## Next Steps
- After all tests pass, feature is ready for code review
