---
phase: 1
title: "Test Infrastructure — DB Helper & Fixtures"
status: complete
priority: P1
effort: 2h
---

# Phase 1: Test Infrastructure

## Context Links
- [TraceResponse struct](../../backend/internal/service/dependency_tracer.go) — lines 13-20
- [Database setup](../../backend/internal/database/database.go)
- [Models](../../backend/internal/model/)
- [Existing test pattern](../../backend/internal/service/dependency_tracer_helpers_test.go)

## Overview
Create shared test utilities: a test DB connection helper and fixture seeding functions that build a minimal but realistic topology graph. All subsequent test phases depend on this.

## Key Insights
- GORM AutoMigrate handles schema creation — reuse `database.Migrate()`
- 8 tables with FKs: BlueprintType, BlueprintNode, BlueprintNodeMembership, BlueprintEdge, CapacityNodeType, DependencyRule, ImpactRule, User
- Tests need a known topology graph to assert against deterministic results
- Config requires DB_USER, DB_PASSWORD, DB_NAME, JWT_SECRET — test helper must set env vars or provide defaults

## Requirements

### Functional
- Connect to a test PostgreSQL database (`app_test`)
- Auto-migrate all tables before tests
- Truncate all tables between test runs (not drop — schema stays)
- Seed a minimal topology graph that exercises all trace paths
- Provide cleanup function for `defer` usage in tests

### Non-functional
- Test helper must be importable from any test package (`handler`, `service`)
- Must not conflict with production database
- Must handle "DB not available" gracefully (skip tests with `t.Skip()`)

## Fixture Data Design

The minimal graph must cover these trace scenarios:

### Blueprint Types (4 rows)
| Name | Slug |
|------|------|
| Electrical System | electrical-system |
| Cooling System | cooling-system |
| Spatial Topology | spatial-topology |
| Whitespace Blueprint | whitespace |

### Blueprint Nodes (10 rows — minimal chain)
| node_id | name | node_type |
|---------|------|-----------|
| UPS-01 | UPS Unit 1 | UPS |
| RPP-01 | RPP Panel 1 | RPP |
| RACKPDU-01 | Rack PDU 1 | Rack PDU |
| RACK-01 | Rack 1 | Rack |
| GEN-01 | Generator 1 | Generator |
| RDHX-01 | RDHx Unit 1 | RDHx |
| AIRZONE-01 | Air Zone 1 | Air Zone |
| ROW-01 | Row 1 | Row |
| ZONE-01 | Zone 1 | Zone |
| CC-01 | Capacity Cell 1 | Capacity Cell |

### Blueprint Edges (electrical chain: GEN -> UPS -> RPP -> RACKPDU)
| blueprint_type slug | from | to | purpose |
|---------------------|------|----|---------|
| electrical-system | GEN-01 | UPS-01 | upstream level 2 |
| electrical-system | UPS-01 | RPP-01 | upstream level 1 |
| electrical-system | RPP-01 | RACKPDU-01 | downstream level 1 |
| cooling-system | RDHX-01 | AIRZONE-01 | cooling local |
| spatial-topology | ZONE-01 | ROW-01 | spatial parent-child |
| spatial-topology | ROW-01 | RACK-01 | spatial parent-child |
| spatial-topology | RACK-01 | RACKPDU-01 | spatial (for bridge) |
| whitespace | CC-01 | RACK-01 | whitespace bridge |

### Capacity Node Types (subset — 6 rows)
| node_type | topology | is_capacity_node |
|-----------|----------|-----------------|
| UPS | Electrical System | true |
| RPP | Electrical System | true |
| Rack PDU | Electrical System | false |
| Generator | Electrical System | false |
| RDHx | Cooling System | true |
| Rack | Spatial Topology | false |
| Air Zone | Cooling System | false |

### Dependency Rules (for Rack PDU source node)
| node_type | dependency_node_type | relationship | topo_relationship | upstream_level |
|-----------|---------------------|-------------|-------------------|---------------|
| Rack PDU | RPP | depends_on | Upstream | 1 |
| Rack PDU | UPS | depends_on | Upstream | 2 |
| Rack PDU | RDHx | depends_on | Local | nil |

### Impact Rules (for RPP source node)
| node_type | impact_node_type | topo_relationship | downstream_level |
|-----------|-----------------|-------------------|-----------------|
| RPP | Rack PDU | Downstream | 1 |
| RPP | Rack | Load | nil |

## Related Code Files

### Files to Create
- `backend/internal/testutil/test_db.go` — DB connection, migrate, truncate, cleanup
- `backend/internal/testutil/test_fixtures.go` — Seed functions for topology graph

### Files NOT to Modify
- All existing code files remain untouched

## Implementation Steps

### 1. Create `testutil` package directory
```
backend/internal/testutil/
```

### 2. Create `test_db.go` (~80 lines)
```go
// Package testutil provides shared test infrastructure.
package testutil

// SetupTestDB connects to app_test, runs migrations, returns db + cleanup func.
// Skips test if PostgreSQL is not available.
func SetupTestDB(t *testing.T) (*gorm.DB, func())

// TruncateAll truncates all tables in dependency-safe order.
func TruncateAll(db *gorm.DB) error
```

Key details:
- Use env var `TEST_DB_NAME` with default `app_test`
- Reuse `database.Migrate()` for schema
- Truncate order: edges -> memberships -> nodes -> types -> rules (FK-safe)
- Use `TRUNCATE ... CASCADE` for simplicity
- Cleanup function closes DB connection
- Call `t.Skip("PostgreSQL not available")` if connection fails

### 3. Create `test_fixtures.go` (~120 lines)
```go
// SeedTraceFixtures inserts the minimal topology graph for trace tests.
// Returns a map of node_id -> DB ID for assertions.
func SeedTraceFixtures(t *testing.T, db *gorm.DB) map[string]uint
```

Key details:
- Insert BlueprintTypes first (FK dependency)
- Insert BlueprintNodes
- Insert BlueprintEdges using DB IDs from node inserts
- Insert CapacityNodeTypes
- Insert DependencyRules + ImpactRules
- Return `map[string]uint` mapping node_id strings to DB uint IDs
- Use `t.Helper()` for clean error traces
- Use `t.Fatal()` on seed failures (tests cannot proceed with bad data)

## Todo List
- [ ] Create `backend/internal/testutil/` directory
- [ ] Implement `test_db.go` with SetupTestDB and TruncateAll
- [ ] Implement `test_fixtures.go` with SeedTraceFixtures
- [ ] Verify compilation: `go build ./internal/testutil/...`
- [ ] Smoke test: write a trivial test in testutil that connects and seeds

## Success Criteria
- `SetupTestDB` connects to test DB or skips gracefully
- `SeedTraceFixtures` inserts all rows without FK violations
- `TruncateAll` clears all data without errors
- Both functions importable from `handler` and `service` test packages
- All code compiles: `go vet ./...` passes

## Risk Assessment
- **PostgreSQL not running in CI**: Mitigated by `t.Skip()` — tests degrade gracefully
- **FK constraint violations during seed**: Mitigated by careful insertion order
- **Test DB name collision**: Use dedicated `app_test` name, never production `app_dev`

## Next Steps
- Phase 2 uses `SetupTestDB` + `SeedTraceFixtures` for httptest integration tests
- Phase 3 uses fixtures for service-level unit tests
