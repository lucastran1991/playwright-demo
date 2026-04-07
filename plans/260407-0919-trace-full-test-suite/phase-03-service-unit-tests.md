---
phase: 3
title: "Service Unit Tests — TraceFull Composition Logic"
status: complete
priority: P2
effort: 1.5h
depends_on: [phase-01]
---

# Phase 3: Service Unit Tests — TraceFull Composition

## Context Links
- [TraceFull method](../../backend/internal/service/dependency_tracer.go) — lines 356-384
- [TraceDependencies](../../backend/internal/service/dependency_tracer.go) — lines 105-221
- [TraceImpacts](../../backend/internal/service/dependency_tracer.go) — lines 224-353
- [Existing helper tests](../../backend/internal/service/dependency_tracer_helpers_test.go)
- [Phase 1 — test infra](phase-01-test-infrastructure.md)

## Overview
Test the `DependencyTracer.TraceFull()` composition logic at the service layer using real DB. Focuses on merge behavior, partial failure handling, and edge cases that are harder to isolate via HTTP integration tests.

## Key Insights

### TraceFull Composition Logic (lines 356-384)
```
1. Call TraceDependencies(nodeID, maxLevels, true)
2. Call TraceImpacts(nodeID, maxLevels)
3. If BOTH fail -> return depErr
4. If only deps fail -> log warning, continue with impacts
5. If only impacts fail -> log warning, continue with deps
6. Merge: resp = depResp, then overlay impResp.Downstream + impResp.Load
7. If depResp was nil, use impResp as base
```

This merge logic has subtle edge cases worth testing directly.

### Load Strategy (3-prong)
1. Walk downstream in infra topologies with extended depth (loadMaxLevels=10)
2. Find spatial ancestors of downstream nodes
3. Find spatial descendants of downstream nodes

All three merge with dedup via `seen` map.

## Test Cases

### File: `backend/internal/service/dependency_tracer_test.go`

#### 1. `TestTraceFull_MergesUpstreamAndDownstream` (~25 lines)
- Source: RPP-01 (has both dep rules for upstream + impact rules for downstream)
- Assert:
  - Response has Upstream from TraceDependencies
  - Response has Downstream from TraceImpacts
  - Source node is consistent (same NodeID, Name, NodeType)

#### 2. `TestTraceFull_NodeNotFound` (~10 lines)
- Source: "DOES-NOT-EXIST"
- Assert: returns error containing "node not found"

#### 3. `TestTraceFull_NodeWithNoDependencyRules` (~20 lines)
- Source: GEN-01 (Generator — no dependency rules, only impact rules)
- Assert:
  - No error returned (partial success)
  - Upstream is empty
  - Downstream may have results if impact rules exist for Generator

#### 4. `TestTraceFull_NodeWithNoImpactRules` (~20 lines)
- Source: RACKPDU-01 (has dependency rules but likely no impact rules for Rack PDU)
- Assert:
  - No error
  - Upstream populated (RPP, UPS)
  - Downstream empty

#### 5. `TestTraceFull_SourceNodeFields` (~15 lines)
- Source: RPP-01
- Assert source fields:
  - `NodeID` == "RPP-01"
  - `Name` == "RPP Panel 1"
  - `NodeType` == "RPP"
  - `Topology` == "Electrical System"

#### 6. `TestTraceFull_LevelsLimitUpstream` (~20 lines)
- Source: RACKPDU-01 with maxLevels=1
- Assert: only level-1 nodes (RPP-01), no level-2 (UPS-01)
- Source: RACKPDU-01 with maxLevels=2
- Assert: level-1 (RPP-01) + level-2 (UPS-01) present

#### 7. `TestTraceFull_LoadOnlyForCapacityNodes` (~20 lines)
- Source: RPP-01 (IsCapacityNode=true) -> Load present
- Source: RACKPDU-01 (IsCapacityNode=false) -> Load empty
- Validates the `isCapacityNode` guard in TraceImpacts

#### 8. `TestTraceDependencies_BridgeFallback` (~25 lines)
- Source: CC-01 (Capacity Cell — connected to Rack via whitespace, Rack connected to RACKPDU via spatial, RACKPDU in electrical)
- Assert: upstream trace via bridge finds electrical nodes through spatial/whitespace walk
- This specifically tests FindBridgeNodesViaSpatial path

#### 9. `TestTraceDependencies_LocalNodes` (~20 lines)
- Source: node with Local dependency rules (e.g., a node that has RDHx as Local dep)
- Assert: `Local` section populated with direct edge neighbors in cooling topology

## Related Code Files

### Files to Create
- `backend/internal/service/dependency_tracer_test.go` — service-level tests

### Files NOT to Modify
- All existing files remain untouched (including `dependency_tracer_helpers_test.go`)

## Implementation Steps

1. Create `dependency_tracer_test.go` in `service` package
2. Implement shared setup:
   - `testutil.SetupTestDB` for DB connection
   - `testutil.SeedTraceFixtures` for graph data
   - Create `DependencyTracer` instance via `NewDependencyTracer(tracerRepo)`
3. Implement each test function
4. Call service methods directly (no HTTP layer)
5. Assert on `TraceResponse` struct fields
6. If file exceeds 200 lines, extract setup into `dependency_tracer_test_setup_test.go`

## Test Setup Pattern

```go
func setupTracer(t *testing.T) (*DependencyTracer, func()) {
    t.Helper()
    db, cleanup := testutil.SetupTestDB(t)
    testutil.TruncateAll(db)
    testutil.SeedTraceFixtures(t, db)

    repo := repository.NewTracerRepository(db)
    tracer := NewDependencyTracer(repo)
    return tracer, cleanup
}
```

Note: `NewDependencyTracer` calls `RefreshLookups()` internally, which loads CapacityNodeTypes and BlueprintTypes from DB. This is why real DB is needed even at service level.

## Helper Functions

```go
// hasNodeInUpstream checks if a node_id exists in any upstream level group.
func hasNodeInUpstream(resp *TraceResponse, nodeID string) bool

// hasNodeInDownstream checks if a node_id exists in any downstream level group.
func hasNodeInDownstream(resp *TraceResponse, nodeID string) bool

// hasNodeInLoad checks if a node_id exists in any load group.
func hasNodeInLoad(resp *TraceResponse, nodeID string) bool

// getUpstreamLevel returns the level of a node in upstream, or -1 if not found.
func getUpstreamLevel(resp *TraceResponse, nodeID string) int
```

## Todo List
- [ ] Create `dependency_tracer_test.go`
- [ ] Implement shared setup function
- [ ] Test: merges upstream + downstream
- [ ] Test: node not found
- [ ] Test: no dependency rules (partial success)
- [ ] Test: no impact rules (partial success)
- [ ] Test: source node fields
- [ ] Test: levels limit upstream depth
- [ ] Test: load only for capacity nodes
- [ ] Test: bridge fallback via spatial
- [ ] Test: local nodes
- [ ] Verify: `go test -v ./internal/service/ -run TestTrace`
- [ ] Check file size, split if >200 lines

## Success Criteria
- All 9 test cases pass with real PostgreSQL
- Tests skip when DB unavailable
- Bridge fallback path exercised (CC-01 -> spatial -> electrical)
- Load 3-strategy logic validated (capacity vs non-capacity)
- No overlap with existing `dependency_tracer_helpers_test.go` tests

## Risk Assessment
- **Bridge fallback test depends on precise fixture edges**: If CC-01 -> RACK-01 (whitespace) + RACK-01 -> RACKPDU-01 (spatial) + RACKPDU-01 in electrical edges don't exist, bridge test fails. Phase 1 fixture design accounts for this.
- **DependencyTracer.RefreshLookups needs real DB**: Cannot mock at service level without refactoring to interface. Real DB is the right choice here.
- **Fixture adjustment may be needed**: If test assertions don't match actual trace results, revisit Phase 1 fixtures. The fixture design is based on code reading but may need tuning during implementation.

## Fixture Dependencies

This table shows which fixtures each test relies on (all from Phase 1 seed):

| Test | Nodes needed | Edges needed | Rules needed |
|------|-------------|-------------|-------------|
| MergesUpstreamAndDownstream | RPP-01, UPS-01, RACKPDU-01 | electrical chain | dep+impact for RPP |
| NodeNotFound | none | none | none |
| NodeWithNoDependencyRules | GEN-01 | electrical | impact for Generator (if any) |
| NodeWithNoImpactRules | RACKPDU-01, RPP-01, UPS-01 | electrical | dep for RackPDU |
| SourceNodeFields | RPP-01 | none | none |
| LevelsLimitUpstream | RACKPDU-01, RPP-01, UPS-01 | electrical chain | dep for RackPDU |
| LoadOnlyForCapacityNodes | RPP-01, RACKPDU-01, RACK-01 | electrical+spatial | impact for RPP |
| BridgeFallback | CC-01, RACK-01, RACKPDU-01 | whitespace+spatial+electrical | dep for CC |
| LocalNodes | node w/ Local rule | cooling edges | local dep rule |

## Next Steps
- After all tests pass, run full suite: `go test ./internal/...`
- Consider adding test coverage reporting: `go test -cover ./internal/...`
