# Code Review: Blueprint CSV Ingestion

**Reviewer:** code-reviewer | **Date:** 2026-04-03 | **Score: 7.5/10**

## Scope

- Files: 12 (8 new, 4 modified)
- LOC: ~729 (new code)
- Focus: All new blueprint ingestion feature files + modified wiring

## Overall Assessment

Solid implementation. Clean separation of concerns (model/repo/service/handler), idiomatic Go, proper use of GORM transactions and upserts. Several security and correctness issues need attention before production.

---

## Critical Issues

### C1. POST /ingest is unauthenticated -- anyone can trigger full re-ingestion

**File:** `router/router.go:39-47`

Blueprint routes are under a public group with comment "public for MVP". The `/ingest` endpoint triggers disk I/O and heavy DB writes. An attacker could DoS the system by spamming POST requests.

**Fix:** Move `POST /ingest` behind `middleware.AuthRequired` at minimum. Consider admin-only middleware.

```go
// Move ingest to protected group or add admin check
protected.POST("/blueprints/ingest", blueprintHandler.Ingest)
```

### C2. Path traversal risk in `BlueprintDir` config

**File:** `config/config.go:37`

`BLUEPRINT_DIR` is read from env and passed directly to `os.ReadDir` / `os.Open`. If a malicious env value is set (or if later the path comes from user input), directory traversal is possible.

**Current risk:** Low (env-only), but no validation that the path is within expected bounds.

**Fix:** Add `filepath.Clean` and validate the resolved path is under an expected root:

```go
BlueprintDir: filepath.Clean(getEnv("BLUEPRINT_DIR", "./blueprint/Node & Edge")),
```

### C3. Raw SQL in GetTree uses parameterized query -- OK but verify

**File:** `blueprint_repository.go:149-169`

The raw SQL uses `?` placeholders (parameterized), so no SQL injection risk here. GORM's `Raw()` with `?` is safe. **No action needed** -- this is correctly handled.

---

## High Priority

### H1. GORM `clause.OnConflict` + `Create` may not return ID after upsert (INSERT ... ON CONFLICT DO NOTHING)

**File:** `blueprint_repository.go:44-49` (UpsertEdge)

When `DoNothing: true` is used with `Create`, PostgreSQL returns 0 rows affected on conflict, and GORM may NOT populate `e.ID`. This is fine for edges since the ID isn't used downstream. But for `UpsertNode` (line 28-33), the returned `node.ID` IS used to build `nodeIDMap`.

**Risk:** If `DoUpdates` is used (as in UpsertNode), PostgreSQL RETURNING clause works and ID is populated. Verified: UpsertNode uses `DoUpdates`, so ID should be returned. **OK for nodes.**

**Remaining concern:** If a node already exists with the same `node_id` but the update columns (name, node_type, node_role) are identical, some GORM+PG combos may still not return the ID. Consider adding a `RETURNING id` clause or doing a find-after-upsert.

**Fix (defensive):**
```go
func (r *BlueprintRepository) UpsertNode(tx *gorm.DB, node *model.BlueprintNode) error {
    err := tx.Clauses(clause.OnConflict{
        Columns:   []clause.Column{{Name: "node_id"}},
        DoUpdates: clause.AssignmentColumns([]string{"name", "node_type", "node_role", "updated_at"}),
    }).Create(node).Error
    if err != nil {
        return err
    }
    // Ensure ID is populated after upsert
    if node.ID == 0 {
        return tx.Where("node_id = ?", node.NodeID).First(node).Error
    }
    return nil
}
```

### H2. N+1 upsert pattern -- each node/edge is an individual INSERT

**File:** `blueprint_ingestion_service.go:87-107`

Each node triggers 2 DB calls (UpsertNode + UpsertMembership). For a domain with 10K nodes, that's 20K+ queries in a single transaction.

**Fix (later):** Batch upserts using `tx.CreateInBatches` with `clause.OnConflict`, or use raw SQL `INSERT ... VALUES (...), (...) ON CONFLICT ...`.

**Severity:** Not blocking for MVP with small datasets, but will be a bottleneck at scale.

### H3. Repository file exceeds 200 LOC limit

**File:** `blueprint_repository.go` -- 220 lines

Per project rules, files should stay under 200 lines.

**Fix:** Extract `GetTree` (and `TreeNode` type) into a separate file like `blueprint_tree_repository.go`. The tree logic is self-contained (~80 lines).

### H4. Inconsistent response format in handler

**File:** `blueprint_handler.go`

- `ListTypes` and `GetTree` use `response.Success(c, ...)` which wraps in `{"data": ...}`
- `ListNodes`, `GetNode`, `ListEdges` use `c.JSON(...)` with manual `gin.H{"data": ..., "total": ...}`

Existing convention (auth_handler.go) mixes both too, but new code should be consistent. Consider a `response.SuccessWithMeta` helper.

---

## Medium Priority

### M1. Silent row skipping in CSV parser -- no logging

**File:** `blueprint_csv_parser.go:92-94, 124-125`

Malformed rows are silently skipped. In production, this could mask data quality issues.

**Fix:** Add a counter or log warning for skipped rows:
```go
log.Printf("WARNING: skipping malformed row %d in %s", i+2, filePath)
```

### M2. `_ = i` suppressor is a code smell

**File:** `blueprint_csv_parser.go:102`

The `_ = i` exists only to suppress an unused variable warning from the `for i, rec := range` loop. The variable `i` is used nowhere.

**Fix:** Use `for _, rec := range records[1:]` and track row number separately if needed for logging (see M1).

### M3. Edge resolution silently skips missing cross-domain nodes

**File:** `blueprint_ingestion_service.go:122-139`

If a from/to node doesn't exist in the current domain OR the DB, the edge is silently skipped with a log warning. This is intentional but should be tracked in `IngestionSummary`.

**Fix:** Add `SkippedEdges int` to `IngestionSummary`.

### M4. TreeNode cycle detection missing

**File:** `blueprint_repository.go:192-206`

The recursive `buildChildren` function has no cycle detection. If data has a cycle (A->B->A), this will infinite loop / stack overflow.

**Fix:** Add a visited set:
```go
visited := make(map[string]bool)
buildChildren = func(nodeID string) []TreeNode {
    if visited[nodeID] { return nil }
    visited[nodeID] = true
    // ... rest of logic
}
```

### M5. No input validation on `typeSlug` and `nodeId` path params

**File:** `blueprint_handler.go:61, 93`

Path params are passed directly to DB queries. While GORM parameterizes them (safe from injection), there's no format validation. A slug like `../../../etc` would just return empty results, but validation is good practice.

---

## Low Priority

### L1. `FolderToSlug` could use `regexp` for cleaner hyphen collapsing

**File:** `blueprint_csv_parser.go:148-150`

The `for strings.Contains` loop works but a regex would be cleaner. Not blocking.

### L2. Blueprint routes comment says "public for MVP" -- add TODO/ticket reference

**File:** `router/router.go:39`

Should reference a ticket or plan phase for when auth will be added.

### L3. `DiscoverDomains` doesn't validate folder contents

Folders without CSV files will cause errors downstream. Consider pre-validating.

---

## Positive Observations

- Clean model definitions with proper GORM tags and composite unique indexes
- Transaction wrapping per domain -- partial failure won't corrupt other domains
- Proper error wrapping with `%w` throughout
- `filepath.Join` used correctly (no string concatenation for paths)
- Pagination with sane defaults and max limit of 1000
- `LazyQuotes` and `FieldsPerRecord = -1` -- good defensive CSV parsing
- Follows existing codebase conventions (struct layout, naming, DI pattern)
- `FolderToSlug`/`FolderToName` are well-thought-out for display vs URL usage
- Build compiles cleanly with no errors

---

## Recommended Actions (Priority Order)

1. **[Critical]** Add auth middleware to POST /ingest endpoint
2. **[High]** Add defensive ID check after UpsertNode (H1)
3. **[High]** Split `blueprint_repository.go` -- extract tree logic to stay under 200 LOC
4. **[Medium]** Add cycle detection in tree builder (M4)
5. **[Medium]** Track skipped rows/edges in IngestionSummary (M1, M3)
6. **[Medium]** Remove `_ = i` smell, use proper row tracking (M2)
7. **[Low]** Standardize response format across all handler methods (H4)

## Metrics

| Metric | Value |
|--------|-------|
| Build | Pass |
| LOC (new) | ~729 |
| Files over 200 LOC | 1 (blueprint_repository.go: 220) |
| SQL Injection Risk | None (parameterized) |
| Auth Coverage | Missing on /ingest |
| Test Coverage | N/A (no tests yet) |

## Unresolved Questions

1. Is `UpsertNode` reliably returning the DB-assigned ID after ON CONFLICT DO UPDATE on all PG versions / GORM versions in use? Needs integration test verification.
2. What's the expected dataset size? If >5K nodes per domain, batch upsert (H2) becomes important.
3. When will auth be added to blueprint endpoints? The "public for MVP" comment needs a timeline.
