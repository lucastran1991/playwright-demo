---
type: code-review
date: 2026-04-07
slug: load-capacity
plan: 260407-1552-load-capacity-calculator
---

# Code Review: Load-Capacity Calculator

## Scope
- **New files (6):** model, repository, CSV parser, ingestion service, calculator, handler
- **Modified files (4 backend + 4 frontend):** database, router, main, dependency_tracer, dag-types, dag-helpers, dag-node, dag-detail-popup
- **LOC:** ~767 new backend, ~70 modified frontend
- **Focus:** correctness, security, edge cases, backward compat

## Overall Assessment

Solid feature implementation. Clean architecture following existing patterns (repository/service/handler). CSV parser mapping approach is well-designed for sparse column data. Calculator's spatial-descendant-based aggregation is correct. A few issues need attention, primarily a frontend bug that drops capacity data during topology filtering.

---

## Critical Issues

### 1. `filterTraceByTopologies` drops `capacity` field (BUG)

**File:** `frontend/src/components/tracer/dag-helpers.tsx:34-40`

The function reconstructs `TraceResponse` but omits `response.capacity`. After topology filtering, all capacity data is lost — nodes render without utilization bars.

**Fix:**
```ts
return {
  source: response.source,
  upstream: response.upstream?.filter((g) => matchTopo(g.topology)),
  local: response.local?.filter((g) => matchTopo(g.topology)),
  downstream: response.downstream?.filter((g) => matchTopo(g.topology)),
  load: response.load?.filter((g) => matchTopo(g.topology)),
  capacity: response.capacity, // <-- add this
}
```

---

## High Priority

### 2. `ListCapacityNodes` total count ignores filters

**File:** `backend/internal/repository/capacity_repository.go:179-181`

The `total` count always returns COUNT(DISTINCT node_id) across ALL node_variables, regardless of the `nodeType` or `minUtil` filter applied to the paginated query. Pagination metadata will be incorrect.

**Fix:** Build the count query using the same filter conditions as the data query.

### 3. N+1 query in `ListCapacityNodes`

**File:** `backend/internal/repository/capacity_repository.go:184-191`

For each nodeID returned, `GetNodeCapacity` runs 2 queries (variables + node metadata). With `limit=500`, that's up to 1000 queries per request.

**Fix:** Batch-load variables and node metadata for all nodeIDs in 2 queries, then assemble in memory.

### 4. `IngestCapacity` POST endpoint has no auth

**File:** `backend/internal/router/router.go:65-71`

The entire `/api/capacity` group is public, including `POST /ingest`. This triggers full CSV re-parse + DB writes + computation. Should be behind auth middleware or at least rate-limited.

**Impact:** Any anonymous user can hammer the endpoint, causing expensive DB operations. DoS risk.

### 5. Calculator aggregates `allocated_load` for non-load variables

**File:** `backend/internal/service/load_capacity_calculator.go:141-143`

Line 143 stores `allocated_load` as the computed variable name for ALL aggregation configs, even when `cfg.LoadVar` is `allocated_air_load` or `allocated_liquid_load`. This means an Air Zone's computed load is stored as `allocated_load` rather than `allocated_air_load`.

This could cause confusion when the frontend displays "Allocated Load" for a cooling node. However, since utilization is computed from `totalLoad / capacity`, the math is correct — it's a naming/semantics issue. Consider storing as `computed_total_load` or using `cfg.LoadVar` as the stored variable name to avoid ambiguity.

---

## Medium Priority

### 6. `getUtilColor` duplicated (DRY violation)

**Files:** `dag-node.tsx:8-12` and `dag-detail-popup.tsx:7-11`

Identical function in two files. Extract to `dag-helpers.tsx` or a shared util.

### 7. `GetNodeCapacity` silently ignores metadata lookup error

**File:** `backend/internal/repository/capacity_repository.go:108`

```go
r.db.Raw(`SELECT ... FROM blueprint_nodes WHERE node_id = ? LIMIT 1`, nodeID).Scan(&node)
```

Error is not checked. If the query fails, `node` will be zero-valued and NodeType/Name will be empty in the response.

### 8. `capacity_repository.go` exceeds 200 line limit (212 LOC)

Per project code standards, files should stay under 200 lines. `CapacitySummary` struct + `GetCapacitySummary` + `ListCapacityNodes` could be extracted to a separate file (e.g., `capacity-repository-queries.go`).

### 9. `GetCapacitySummary` SQL counts `high_util_nodes` including overloaded

**File:** `backend/internal/repository/capacity_repository.go:142`

`value > 80` includes nodes with value > 100 (overloaded). If the intent is "80-100% exclusive of overloaded," the condition should be `value > 80 AND value <= 100`. Otherwise the `high_util_nodes` count is a superset of `overloaded_nodes`. Document the intent.

### 10. Ingestion error handling — partial compute failure

**File:** `backend/internal/service/capacity_ingestion_service.go:78-85`

If `ComputeAll()` fails, the error is appended to `summary.Errors` but the function returns `nil` error. The raw CSV data is committed but computed aggregates may be missing. The caller gets HTTP 200 with a partial result. Consider returning an error or at minimum a non-200 status when computation fails.

---

## Low Priority

### 11. `parseFloatParam` in capacity_handler doesn't match existing pattern

**File:** `backend/internal/handler/capacity_handler.go:79-84`

The existing `parseIntParam` (in tracer_handler.go) includes a `maxVal` guard. `parseFloatParam` has no upper bound. A malicious request with `min_utilization=99999999` would return no results but is semantically odd. Minor, no real impact.

### 12. No input validation on `nodeId` path param

**File:** `backend/internal/handler/capacity_handler.go:41`

`c.Param("nodeId")` is passed directly to the repository. GORM parameterizes it, so no SQL injection risk, but there's no length/format validation. An extremely long nodeId would still be sent to DB. Very minor.

---

## Edge Cases Found by Scout

1. **Double-counting risk:** Calculator sums `varMap[rack.NodeID][cfg.LoadVar]` from CSV-imported data only (`GetVariableMap("csv_import")`). Since computed values are deleted first and only CSV values are read, no double-counting occurs. This is correct.

2. **Rack with zero rated_capacity:** Division guard (`if rated > 0`) prevents NaN. `utilization_pct` will be 0.0 for such racks. Frontend handles this gracefully (0% bar).

3. **Frontend null safety:** `cap?.utilization_pct != null` uses `!=` (not `!==`), which correctly catches both `null` and `undefined`. Good.

4. **Topology filter + capacity interaction:** When user deselects all topologies, `selectedTopos.size === 0` returns full unfiltered response (including capacity). This is correct behavior per the early return.

5. **Large IN clause:** `GetCapacityMapForNodes` receives all nodeIDs from a trace response. For large graphs this could be 500+ IDs in a single `WHERE IN`. PostgreSQL handles this fine, but consider batching if traces grow to 10k+ nodes.

6. **Transaction scope:** `ComputeAll` runs inside a single transaction (delete computed + recompute + bulk upsert). If the server crashes mid-computation, the transaction rolls back cleanly. Correct.

---

## Positive Observations

- KV model (`node_variables`) is a pragmatic choice for sparse 35-column data
- Upsert with `ON CONFLICT` prevents stale duplicate data
- `ComputeAll` correctly deletes prior computed values before recomputing (idempotent)
- `FindSpatialDescendantsOfType` reuses the existing recursive CTE pattern from dependency_tracer
- Frontend capacity enrichment is cleanly separated (enrichment happens in `traceToDAGElements`, not in API layer)
- Utilization color coding (green/yellow/red) follows standard thresholds
- `SetCapacityRepo` injection pattern avoids circular dependency between tracer and capacity modules
- `go vet` passes clean

---

## Recommended Actions (Priority Order)

1. **[Critical]** Fix `filterTraceByTopologies` to pass through `capacity` field
2. **[High]** Fix `ListCapacityNodes` total count to respect filters
3. **[High]** Add auth/rate-limiting to `POST /api/capacity/ingest`
4. **[High]** Fix N+1 query in `ListCapacityNodes` with batch loading
5. **[Medium]** Extract shared `getUtilColor` to dag-helpers
6. **[Medium]** Check error from `GetNodeCapacity` metadata lookup
7. **[Medium]** Clarify `high_util_nodes` vs `overloaded_nodes` semantics

## Metrics

- **Type Coverage:** Frontend TS interfaces are well-typed; `CapacityData` uses index signature for extensibility
- **Test Coverage:** No new tests for capacity feature yet (Phase 5 per plan)
- **Linting:** `go vet` clean, no TypeScript errors observed
- **Build:** Backend compiles successfully

## Unresolved Questions

1. Should `allocated_load` be the stored variable name for all aggregation types (air, liquid, total) or should each keep its own name?
2. Should `POST /ingest` be an admin-only endpoint or remain public?
3. Is `high_util_nodes` intended to include overloaded nodes (>100%) or only 80-100% range?
