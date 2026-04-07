---
title: "Trace Full API Test Suite"
description: "Integration and unit tests for /trace/full endpoint using real PostgreSQL and httptest"
status: complete
priority: P1
effort: 6h
branch: main
tags: [testing, go, integration, trace-api]
created: 2026-04-07
---

# Trace Full API Test Suite

## Goal
End-to-end test coverage for `GET /api/trace/full/:nodeId` — the core DAG tracing endpoint. Validates response schema, business logic (partial failures, bridge fallback, load strategies), and edge cases.

## Phases

| # | Phase | Status | Effort | File |
|---|-------|--------|--------|------|
| 1 | Test Infrastructure (DB helper + fixtures) | complete | 2h | [phase-01](phase-01-test-infrastructure.md) |
| 2 | Integration Tests (httptest + real DB) | complete | 2.5h | [phase-02](phase-02-integration-tests.md) |
| 3 | Service Unit Tests (TraceFull composition) | complete | 1.5h | [phase-03](phase-03-service-unit-tests.md) |

## Architecture

```
backend/internal/
  testutil/                      <- NEW: shared test infrastructure
    test_db.go                   <- DB connect, migrate, cleanup
    test_fixtures.go             <- Seed topology graph data
  handler/
    tracer_handler_test.go       <- NEW: httptest integration tests
  service/
    dependency_tracer_test.go    <- NEW: TraceFull composition tests
    dependency_tracer_helpers_test.go  <- EXISTING (keep as-is)
```

## Key Decisions
- **Real PostgreSQL** for integration tests (no mocks per user preference)
- **Standard `testing` package** only (matches codebase convention)
- **Test DB**: `app_test` database, created/dropped per test suite run
- **Fixture data**: Minimal but realistic topology graph (Rack -> RPP -> UPS chain + spatial + cooling)
- **Skip Playwright E2E** for now — no existing setup, low ROI at this stage

## Files to Create

| File | Package | Purpose | Est. Lines |
|------|---------|---------|-----------|
| `backend/internal/testutil/test_db.go` | testutil | DB connect, migrate, truncate, cleanup | ~80 |
| `backend/internal/testutil/test_fixtures.go` | testutil | Seed minimal topology graph | ~120 |
| `backend/internal/handler/tracer_handler_test.go` | handler | 10 httptest integration tests | ~190 |
| `backend/internal/service/dependency_tracer_test.go` | service | 9 service-level tests | ~180 |

## Test Coverage Summary

| Area | Tests | What's validated |
|------|-------|-----------------|
| HTTP layer | 10 | Status codes, JSON schema, param parsing, 404 |
| TraceFull merge | 4 | Partial failures, merge of deps+impacts, source fields |
| Levels/depth | 3 | Default=2, cap=10, filtering by level |
| Load strategies | 2 | Capacity vs non-capacity, 3-strategy merge |
| Bridge fallback | 1 | Spatial/whitespace walk to find cross-topology nodes |
| Local deps | 1 | Direct neighbor lookup in cooling topology |

## Dependencies
- PostgreSQL running locally
- `app_test` database must be creatable by test user
- All tests skip gracefully via `t.Skip()` if DB unavailable
