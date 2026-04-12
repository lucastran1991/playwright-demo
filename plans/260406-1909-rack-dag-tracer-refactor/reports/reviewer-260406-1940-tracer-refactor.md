# Code Review: Rack DAG Tracer Refactor

## Scope

- **Files:** 9 changed (4 backend, 2 data, 3 frontend)
- **LOC:** +353 / -102
- **Focus:** Cross-topology dependency/impact resolution, spatial-bridge traversal, combined `/trace/full` API, frontend single-call DAG
- **Build:** Go compiles clean, TypeScript compiles clean

## Overall Assessment

Solid refactor. The spatial-bridge traversal pattern is well-designed and the 3-strategy Load resolution is thorough. SQL changes are correct. The combined `/trace/full` API reduces frontend complexity from 2 queries to 1. A few medium-priority issues around dead code, error swallowing, and React lint warnings.

---

## Critical Issues

None found.

---

## High Priority

### H1. `TraceFull` silently drops one error when only one of dep/impact fails

**File:** `backend/internal/service/dependency_tracer.go:316-338`

```go
depResp, depErr := t.TraceDependencies(nodeID, maxLevels, true)
impResp, impErr := t.TraceImpacts(nodeID, maxLevels)

if depErr != nil && impErr != nil {
    return nil, depErr
}
```

If `TraceDependencies` succeeds but `TraceImpacts` fails with a real DB error (not "node not found"), the error is silently swallowed. This is acceptable for partial results, but the caller gets no signal that half the data is missing.

**Recommendation:** Log the non-nil error when only one fails:

```go
if depErr != nil && impErr == nil {
    log.Printf("WARNING: TraceFull dep trace failed for %s: %v", nodeID, depErr)
} else if impErr != nil && depErr == nil {
    log.Printf("WARNING: TraceFull impact trace failed for %s: %v", nodeID, impErr)
}
```

**Impact:** Debugging production issues -- silent data gaps are hard to diagnose.

### H2. Double downstream query in TraceImpacts for load resolution

**File:** `backend/internal/service/dependency_tracer.go:259-274`

The load resolution block re-queries `FindDownstreamNodes` for each slug in `infraSlugSet` with `loadMaxLevels=10`, even though the same slugs were already queried at lines 239-246 with `maxLevels`. This means the downstream traversal runs twice per topology -- once for the response and once for load collection.

**Recommendation:** Cache downstream results from the first pass and reuse, or collect DB IDs during the initial downstream traversal:

```go
// During initial downstream pass, also collect IDs
var allDownstreamDBIDs []uint
for topo, allowedTypes := range downstreamByTopo {
    // ...existing query...
    for _, n := range nodes {
        allDownstreamDBIDs = append(allDownstreamDBIDs, n.ID)
    }
}
```

Note: the second pass uses `loadMaxLevels=10` (deeper), so results may differ from first pass with `maxLevels` (user-specified, default 2). This is intentional but worth documenting via comment explaining why a deeper pass is needed.

**Impact:** Performance -- N+1 recursive CTE queries per trace call.

---

## Medium Priority

### M1. `HasEdgesInTopology` is dead code

**File:** `backend/internal/repository/tracer_repository.go:124-132`

Added but never called anywhere in the service layer. `FindBridgeNodesViaSpatial` already filters via `EXISTS` subquery, making this redundant.

**Recommendation:** Remove or add a `// TODO: remove if unused after validation` comment. Dead code increases maintenance surface.

### M2. useEffect missing dependency array entries (React lint)

**File:** `frontend/src/components/tracer/dependency-impact-dag.tsx:73`

```
React Hook useEffect has missing dependencies: 'fitView', 'setEdges', and 'setNodes'
```

`setNodes`, `setEdges` from `useNodesState`/`useEdgesState` are stable references (like `useState` setters), and `fitView` from `useReactFlow` is also stable. Including them silences the lint warning without behavioral change.

**Recommendation:**
```ts
}, [selectedNodeId, traceQuery.data, traceQuery.isLoading, setNodes, setEdges, fitView])
```

### M3. `handleClear` useCallback has empty deps but uses `setNodes`/`setEdges`

**File:** `frontend/src/components/tracer/dependency-impact-dag.tsx:79-83`

React Compiler skipped optimization due to mismatch. Same fix as M2 -- include stable setters in deps:

```ts
const handleClear = useCallback(() => {
    setSelectedNodeId(null)
    setNodes([])
    setEdges([])
}, [setNodes, setEdges])
```

### M4. `direction` field added to TracerNodeData but never consumed

**File:** `frontend/src/components/tracer/dag-types.ts:47`, `dag-node.tsx`

The `direction` field was added to the data model and set in `makeNode`, but `dag-node.tsx` never reads it. This is fine if planned for future use (e.g. directional badges), but currently YAGNI.

**Recommendation:** If immediate use planned, note it. Otherwise consider removing until needed, to keep the interface minimal.

### M5. Hardcoded spatial-topology slug in SQL

**Files:** `tracer_repository.go` lines 104, 113, 145, 154, 182, 191

The string `'spatial-topology'` is hardcoded in 4 different queries. If the slug changes, all must be updated.

**Recommendation:** Extract to a package-level constant:

```go
const spatialTopologySlug = "spatial-topology"
```

### M6. `FindSpatialAncestorsOfType` and `FindSpatialDescendantsOfType` hardcode maxLevel=5

**File:** `tracer_repository.go:113, 191`

Both use `a.level < 5` hardcoded in SQL. `FindBridgeNodesViaSpatial` accepts `maxDepth` as a parameter. Inconsistency -- if spatial hierarchies grow deeper, these queries silently truncate.

**Recommendation:** Add a `maxDepth` parameter to match `FindBridgeNodesViaSpatial` signature, or at minimum extract the constant and document the limit.

---

## Low Priority

### L1. `edges.some()` for dedup is O(n^2)

**File:** `frontend/src/components/tracer/dag-helpers.tsx:75, 108`

```ts
if (!edges.some((e) => e.id === edgeId)) {
```

For large graphs this is quadratic. A `Set<string>` for edge IDs would be O(1) lookup.

**Impact:** Negligible for current graph sizes (typically <100 edges), but worth noting for scale.

### L2. Non-deterministic Load group ordering

**File:** `backend/internal/service/dependency_tracer.go:307-309`

```go
for topo, nodes := range loadGroupMap {
```

Go map iteration is non-deterministic. Response ordering varies between calls for the same input. Consider sorting `resp.Load` by topology name before returning for deterministic API output.

---

## Edge Cases Found by Scout

1. **Cycle in spatial hierarchy:** `FindSpatialAncestorsOfType` and `FindBridgeNodesViaSpatial` use `UNION ALL` with level limit, which prevents infinite loops. Good.

2. **Bridge node is also the source node:** If spatial descendants of a node include the node itself (self-loop), `FindBridgeNodesViaSpatial` would include it. The `WHERE be.from_node_id = ?` base case only picks children, not self. Safe.

3. **Empty `infraSlugSet` when node has no topology mapping:** `lookupTopology` falls back to "Electrical System", so `infraSlugSet` always has at least one entry. However, if `resolveSlug("Electrical System")` returns empty (no matching blueprint type), `infraSlugSet` remains empty and Load resolution skips entirely. This is a data integrity edge case, not a code bug.

4. **`allDownstreamDBIDs` can be very large:** If a high-level node (Utility Feed) has thousands of downstream nodes, the `IN ?` clause in `FindSpatialAncestorsOfType` could generate a massive SQL IN list. PostgreSQL handles large IN lists but performance degrades.

5. **Bridge selection: "most upstream results" may not be best:** The spatial-bridge loop picks the bridge with `len(candidate) > len(bestFiltered)`. This is a heuristic -- it picks quantity over relevance. If bridge A returns 3 relevant nodes and bridge B returns 4 irrelevant-but-allowed nodes, B wins. This is acceptable given the domain but worth documenting.

---

## Positive Observations

- Clean separation: repo handles SQL, service handles logic, handler handles HTTP. Good layering.
- `DISTINCT ON (id)` with `ORDER BY id, level` correctly picks minimum-level row per node. PostgreSQL semantics used correctly.
- Guard clauses (`if len(nodeDBIDs) == 0`) prevent empty IN clause SQL errors.
- Frontend simplification from 2 queries to 1 reduces race conditions and complexity.
- `seen` map in load resolution correctly prevents duplicate load nodes across 3 strategies.
- Load edge styling (purple dashed) properly distinguished from local (gray dashed).

---

## Recommended Actions (Prioritized)

1. **H1** -- Add warning logs in `TraceFull` for partial failures
2. **H2** -- Document why double downstream query is needed (depth difference), or cache first-pass results
3. **M1** -- Remove `HasEdgesInTopology` dead code
4. **M2/M3** -- Fix useEffect/useCallback dependency arrays
5. **M5** -- Extract `spatial-topology` constant
6. **M6** -- Parameterize maxDepth in spatial ancestor/descendant queries

---

## Metrics

- **Type Coverage:** 100% (TypeScript strict, all interfaces typed)
- **Test Coverage:** Not assessed (no test files changed)
- **Linting Issues:** 2 (React hooks exhaustive-deps warning + Compiler skip)
- **Build Status:** Go clean, TS clean

---

## Unresolved Questions

1. Is `HasEdgesInTopology` planned for future use, or was it left over from an earlier approach?
2. Should `TraceFull` return a `warnings` field to surface partial failures to the frontend?
3. Is the hardcoded `loadMaxLevels=10` in TraceImpacts sufficient for all topologies, or should it be configurable?
