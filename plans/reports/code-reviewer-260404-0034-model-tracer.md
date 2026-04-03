# Code Review: Capacity/Dependency/Impact Model + Dependency Tracer

## Scope
- Files: 8 new, 5 modified
- LOC: ~774 new code
- Focus: All new files + modified wiring
- Build: PASSES (go build ./...)

## Overall Assessment

Solid implementation. Clean separation of concerns (model/repository/service/handler). CSV parsing reuses existing `ReadCSV`. Upsert logic is correct. CTE queries are structurally sound. A few issues worth addressing, mostly around maintainability and one potential correctness bug.

---

## Critical Issues

### C1. Hardcoded `inferTopology()` map -- fragile, will silently misclassify new node types
**File:** `backend/internal/service/dependency_tracer.go:196-217`

The `inferTopology()` function hardcodes a map of node types to topologies. This duplicates data that already exists in the `capacity_node_types` table (which has a `topology` column). Any new node type added to the CSV but missing from this map silently falls through to `"Electrical System"` default.

Also contains a typo: `"Liquid Lood"` should be `"Liquid Loop"` (line 198 -- duplicate key with wrong spelling).

**Fix:** Query `capacity_node_types` table to build the map at tracer creation time, or pass it into `groupDepRules`/`groupImpactRules`. Cache it in the `DependencyTracer` struct.

```go
type DependencyTracer struct {
    repo        *repository.TracerRepository
    topoLookup  map[string]string // nodeType -> topology
}

func NewDependencyTracer(repo *repository.TracerRepository) *DependencyTracer {
    t := &DependencyTracer{repo: repo}
    t.refreshTopoLookup() // load from DB
    return t
}

func (t *DependencyTracer) refreshTopoLookup() {
    types, err := t.repo.ListCapacityNodeTypes()
    if err != nil {
        return // fallback to empty map
    }
    lookup := make(map[string]string, len(types))
    for _, ct := range types {
        lookup[ct.NodeType] = ct.Topology
    }
    t.topoLookup = lookup
}
```

**Impact:** Correctness. Without this, any new node type added to CSVs but not to hardcoded map gets wrong topology, producing incorrect trace results with no warning.

---

## High Priority

### H1. `topologyToSlug()` may not match `FolderToSlug()` output
**File:** `dependency_tracer.go:148-149` vs `blueprint_csv_parser.go:140-151`

`topologyToSlug()` does `strings.ToLower(strings.ReplaceAll(topology, " ", "-"))`.
`FolderToSlug()` strips `_Blueprint` suffix, replaces underscores, collapses `--`.

If topology names in `capacity_node_types` don't exactly match the slug format stored in `blueprint_types.slug`, the CTE queries will return zero results silently.

Example: topology `"Cooling System"` -> slug `"cooling-system"` (OK). But topology `"Whitespace Blueprint"` -> slug `"whitespace-blueprint"`, while folder `"Whitespace_Blueprint"` -> slug `"whitespace"`.

**Fix:** Either normalize consistently using the same function, or join against `blueprint_types.name` instead of matching on slug. At minimum, add a test that verifies all topology values in `capacity_node_types` resolve to valid `blueprint_types.slug` values.

### H2. Errors silently swallowed in trace loops
**File:** `dependency_tracer.go:77-79, 88-90, 126-128, 136-138`

Multiple `continue` statements on error:
```go
nodes, err := t.repo.FindUpstreamNodes(...)
if err != nil {
    continue // error lost
}
```

If a topology slug doesn't match any `blueprint_types` row, this silently returns empty results. User gets a partial response with no indication something failed.

**Fix:** Collect errors and include them in the response, or at minimum log them:
```go
if err != nil {
    log.Printf("WARN: upstream trace failed for topology %s: %v", topo, err)
    continue
}
```

### H3. `dependency_tracer.go` exceeds 200-line limit
**File:** 242 lines, project convention is max 200.

**Fix:** Extract `inferTopology`, `filterByTypes`, `groupByLevel`, `groupDepRules`, `groupImpactRules` into a separate `tracer_helpers.go` file.

---

## Medium Priority

### M1. N+1 upsert pattern in `model_ingestion_service.go`
**File:** `model_ingestion_service.go:55-102`

Each row is upserted individually in a loop. For 24 capacity node types + ~30 rules this is fine. If datasets grow, this becomes a bottleneck.

**Fix:** Not urgent. Batch upserts with `tx.Clauses(clause.OnConflict{...}).Create(&slice)` when GORM supports slice upsert (it does). Current approach is acceptable for <100 rows.

### M2. No input validation on `nodeId` path parameter
**File:** `tracer_handler.go:54,71`

`nodeID := c.Param("nodeId")` passes raw user input directly to DB query. While GORM parameterizes queries (no SQL injection), there's no format validation. Malformed node IDs will just 404, which is fine, but a length check would prevent unnecessarily long strings hitting the DB.

**Fix:** Add basic validation:
```go
nodeID := c.Param("nodeId")
if len(nodeID) > 255 || nodeID == "" {
    response.Error(c, http.StatusBadRequest, "invalid node ID")
    return
}
```

### M3. Error detection via string matching is fragile
**File:** `tracer_handler.go:60,78`

```go
if strings.Contains(err.Error(), "node not found") {
```

If the error message changes, the 404 detection breaks.

**Fix:** Define a sentinel error or custom error type:
```go
var ErrNodeNotFound = errors.New("node not found")
// In tracer: return nil, ErrNodeNotFound
// In handler: if errors.Is(err, service.ErrNodeNotFound)
```

### M4. POST `/models/ingest` leaks internal error details
**File:** `tracer_handler.go:36`

```go
response.Error(c, http.StatusInternalServerError, "Model ingestion failed: "+err.Error())
```

Internal errors (DB connection failures, file system errors) get returned verbatim to the client.

**Fix:** Log the full error, return generic message to client:
```go
log.Printf("ERROR: model ingestion failed: %v", err)
response.Error(c, http.StatusInternalServerError, "Model ingestion failed")
```

---

## Low Priority

### L1. `ReadCSV` was renamed from `readCSV` (now exported)
**File:** `blueprint_csv_parser.go:162`

This is correct and intentional to share with `model_csv_parser.go`. No issue, just noting the breaking change is internal-only (same package).

### L2. Hardcoded CSV file names
**File:** `model_ingestion_service.go:38-49`

`"Capacity Nodes.csv"`, `"Dependencies.csv"`, `"Impacts.csv"` are hardcoded. Acceptable since these are fixed-format model files, but consider making them constants.

---

## Edge Cases Found

1. **Cyclic graph in CTE**: If `blueprint_edges` contains cycles, the recursive CTE could loop. `UNION ALL` with a depth limit (`u.level < ?`) prevents infinite recursion, but cycles at the same level could produce duplicate rows. The `DISTINCT` + `GROUP BY` in the final SELECT mitigates this. Acceptable.

2. **Node exists in multiple topologies**: A node (e.g., Rack) appears in both Electrical and Spatial topologies. The tracer queries each topology independently, which is correct. But `FindNodeByStringID` returns only the `BlueprintNode` (which has no topology). The node type is used to look up rules -- this works because rules are type-level, not topology-level.

3. **Empty trace results**: If a node type has no dependency/impact rules, the response returns empty arrays. This is correct behavior but the client should handle it.

4. **Concurrent ingestion**: Two simultaneous `POST /models/ingest` calls could race on upserts. The transaction + unique constraint upserts handle this correctly (last writer wins).

5. **`topologyToSlug` mismatch**: As noted in H1, the `"Whitespace Blueprint"` topology name could produce a slug that doesn't match any `blueprint_types` entry, causing silent empty results.

---

## Positive Observations

- Clean layer separation: model -> repository -> service -> handler
- Proper use of GORM's `clause.OnConflict` for idempotent upserts
- CTE queries are well-structured with depth limits
- `parseIntParam` with max cap prevents abuse of recursion depth
- Auth middleware correctly protects the ingest endpoint
- CSV parser reuse via exported `ReadCSV`
- All files under 200 LOC except `dependency_tracer.go`

---

## Recommended Actions (Priority Order)

1. **[Critical]** Replace `inferTopology()` with DB-backed lookup from `capacity_node_types` table
2. **[Critical]** Fix typo `"Liquid Lood"` -> remove duplicate entry
3. **[High]** Verify `topologyToSlug()` output matches actual `blueprint_types.slug` values in DB
4. **[High]** Add logging for swallowed errors in trace loops
5. **[High]** Split `dependency_tracer.go` into two files to stay under 200 LOC
6. **[Medium]** Use sentinel error for "node not found" instead of string matching
7. **[Medium]** Don't leak internal error details in ingestion endpoint response
8. **[Medium]** Add basic length validation on `nodeId` parameter
9. **[Low]** Extract CSV filenames as constants

## Metrics

- Type Coverage: N/A (Go, statically typed)
- Test Coverage: 0% (no tests yet, phase 6 pending)
- Linting Issues: 0 (builds clean)
- Files over 200 LOC: 1 (`dependency_tracer.go` at 242)

## Unresolved Questions

1. What are the actual `blueprint_types.slug` values in the DB? Need to verify `topologyToSlug()` produces matching values, especially for "Whitespace Blueprint".
2. Should trace endpoints require authentication? Currently public (unlike ingest which is protected).
3. Is `refreshTopoLookup()` acceptable at startup, or should it refresh per-request (in case capacity node types change)?
