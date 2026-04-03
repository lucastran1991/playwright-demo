# Phase 6: Testing

## Context Links
- [Plan Overview](plan.md)
- All prior phase files
- Existing test pattern: none yet (no test files in codebase)

## Overview
- **Priority**: P2
- **Status**: completed
- **Description**: Integration tests for CSV parsing, ingestion logic, and API endpoints

## Key Insights
- No existing test files in the codebase -- establish test conventions
- Data volume is small enough for real DB integration tests
- CSV test fixtures can use small subsets of actual data
- Per development rules: no mocks for DB, use real PostgreSQL

## Requirements

### Functional
- Test CSV parser: valid files, malformed headers, empty rows, missing files
- Test ingestion: upsert idempotency, cross-domain node sharing, edge resolution
- Test API endpoints: ingest trigger, list/filter/tree queries
- Test slug derivation from folder names

### Non-functional
- Use Go standard `testing` package
- Test files co-located with source (Go convention: `_test.go` suffix)
- Use test DB (separate from dev DB or use transactions + rollback)

## Architecture

```
Tests to create:
  service/blueprint_csv_parser_test.go     -- unit tests for parsing
  service/blueprint_ingestion_service_test.go -- integration tests with DB
  handler/blueprint_handler_test.go        -- HTTP endpoint tests
```

## Related Code Files

### Files to Create
- `backend/internal/service/blueprint_csv_parser_test.go`
- `backend/internal/service/blueprint_ingestion_service_test.go`
- `backend/internal/handler/blueprint_handler_test.go`
- `backend/testdata/blueprint/` -- small CSV fixtures

### Files to Reference
- All Phase 1-5 implementation files

## Implementation Steps

1. Create test fixtures in `backend/testdata/blueprint/`:
   - `TestDomain_Blueprint/Nodes.csv` -- 5-10 sample rows
   - `TestDomain_Blueprint/Edges.csv` -- 3-5 sample edges
   - `BadHeaders/Nodes.csv` -- malformed header for error testing

2. Create `blueprint_csv_parser_test.go`:
   - `TestDiscoverDomains` -- discovers test fixtures
   - `TestParseNodesCSV_Valid` -- parses sample nodes
   - `TestParseNodesCSV_BadHeaders` -- returns error
   - `TestParseEdgesCSV_Valid` -- parses sample edges
   - `TestParseEdgesCSV_EmptyRows` -- skips empty rows
   - `TestFolderToSlug` -- tests slug derivation for all 6 known folder names

3. Create `blueprint_ingestion_service_test.go`:
   - Setup: connect to test DB, run migrations
   - `TestIngestAll_FirstRun` -- ingests fixtures, verify counts
   - `TestIngestAll_Idempotent` -- run twice, verify no duplicates
   - `TestIngestAll_CrossDomainNode` -- same node in 2 domains, verify single row + 2 memberships
   - Teardown: clean up test data

4. Create `blueprint_handler_test.go`:
   - Use `httptest.NewRecorder` + Gin test mode
   - `TestIngestEndpoint` -- POST returns summary
   - `TestListTypes` -- GET returns domain list
   - `TestListNodes_WithFilter` -- GET with ?type=slug
   - `TestGetTree` -- GET recursive tree structure

## Todo List
- [x] Create test CSV fixtures
- [x] Write CSV parser unit tests
- [x] Write ingestion integration tests
- [x] Write handler/endpoint tests
- [x] Run `go test ./...` -- all pass
- [x] Check test coverage

## Success Criteria
- All tests pass with `go test ./...`
- CSV parser tests cover valid, invalid, and edge cases
- Ingestion tests verify idempotency with real DB
- Handler tests verify HTTP status codes and response shapes
- No mocked database -- real PostgreSQL used

## Risk Assessment
- **Medium**: Test DB setup complexity -- need separate DB or transaction rollback strategy
  - **Mitigation**: Use `app_test` database, or wrap each test in a transaction that rolls back
- **Low**: Test fixtures diverge from real CSV format
  - **Mitigation**: Copy real rows, just fewer of them

## Security Considerations
- Test DB credentials should not be committed (use env vars or .env.test)
- Test fixtures contain no sensitive data (infrastructure topology only)
