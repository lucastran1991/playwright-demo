# Phase 1: Fix Data + SQL Bugs

## Context

- [Plan overview](plan.md)
- [Brainstorm report](../reports/brainstorm-260406-1909-rack-dependency-impact-dag-fix.md)

## Overview

- **Priority:** Critical
- **Status:** Complete
- **Description:** Fix foundational data and SQL issues that block all other phases.

## Key Insights

Three independent bugs that each cause visible failures:
1. Impacts.csv missing Rack as source node type → impact API returns empty for Rack
2. `FindUpstreamNodes` GROUP BY includes `parent_node_id` → duplicate rows for same node
3. `FindSpatialAncestorsOfType` SQL has `DISTINCT ON` + `GROUP BY` conflict → Load resolution fails silently

## Related Code Files

### Modify
- `blueprint/Impacts.csv` — add Rack impact rules
- `backend/internal/repository/tracer_repository.go` — fix SQL queries (lines 29-57, 94-124)

### Reference (read-only)
- `blueprint/Dependencies.csv` — verify completeness
- `blueprint/Capacity Nodes.csv` — node type → topology mapping
- `backend/internal/service/dependency_tracer.go` — understand how SQL results are consumed

## Implementation Steps

### Step 1: Add Missing Rack Impact Rules to Impacts.csv

Rack should impact its spatial/electrical children. Add these rows at the end of Impacts.csv (before the blank trailing line):

```csv
Rack,Rack PDU,Downstream,1
Rack,Rack PDU Outlet,Downstream,2
Rack,Server,Downstream,1
```

Rack doesn't impact upstream infrastructure — it only impacts what's physically inside it.

### Step 2: Fix FindUpstreamNodes Duplicate Rows

**File:** `backend/internal/repository/tracer_repository.go` lines 29-57

**Problem:** `GROUP BY id, node_id, name, node_type, parent_node_id` produces multiple rows for the same node when reachable via different parents (e.g., SWGR at L5 via UPS + L6 via BESS).

**Fix:** Remove `parent_node_id` from GROUP BY. Use `MIN(level)` and pick the parent from the shortest path. Replace the final SELECT:

```sql
-- BEFORE (broken):
SELECT id, node_id, name, node_type, MIN(level) as level, parent_node_id
FROM upstream
GROUP BY id, node_id, name, node_type, parent_node_id
ORDER BY level, node_type, node_id

-- AFTER (fixed):
SELECT DISTINCT ON (id) id, node_id, name, node_type, level, parent_node_id
FROM upstream
ORDER BY id, level, parent_node_id
```

This keeps ONE row per node: the one with the smallest level (shortest path), with a deterministic parent.

### Step 3: Fix FindDownstreamNodes Duplicate Rows (Same Pattern)

**File:** `backend/internal/repository/tracer_repository.go` lines 60-87

Apply the same fix as Step 2:

```sql
-- BEFORE:
SELECT id, node_id, name, node_type, MIN(level) as level, parent_node_id
FROM downstream
GROUP BY id, node_id, name, node_type, parent_node_id
ORDER BY level, node_type, node_id

-- AFTER:
SELECT DISTINCT ON (id) id, node_id, name, node_type, level, parent_node_id
FROM downstream
ORDER BY id, level, parent_node_id
```

### Step 4: Fix FindSpatialAncestorsOfType SQL

**File:** `backend/internal/repository/tracer_repository.go` lines 94-124

**Problem:** Current SQL has `DISTINCT ON (id)` combined with `GROUP BY` and `MIN(level)`. These conflict — `DISTINCT ON` operates before `GROUP BY` in PostgreSQL's execution model when both are present, causing unpredictable results.

**Fix:** Remove `GROUP BY` and `MIN()`, rely solely on `DISTINCT ON` with proper ordering:

```sql
-- BEFORE (broken):
SELECT DISTINCT ON (id) id, node_id, name, node_type, MIN(level) as level
FROM ancestors
WHERE node_type IN ?
GROUP BY id, node_id, name, node_type
ORDER BY id, level

-- AFTER (fixed):
SELECT DISTINCT ON (id) id, node_id, name, node_type, level
FROM ancestors
WHERE node_type IN ?
ORDER BY id, level
```

`DISTINCT ON (id)` + `ORDER BY id, level` picks the row with the smallest level per unique id.

### Step 5: Re-ingest Model CSVs

After modifying Impacts.csv, trigger re-ingestion:

```bash
curl -X POST http://localhost:8889/api/models/ingest
```

### Step 6: Verify Fixes

Test each fix independently:

```bash
# Rack impacts should now return downstream nodes
curl -s "http://localhost:8889/api/trace/impacts/RACK-R1-Z1-R1-01?levels=3" | python3 -m json.tool

# UPS impacts should include Load nodes (Rack, Row, Zone)
curl -s "http://localhost:8889/api/trace/impacts/UPS-01?levels=3" | python3 -c "
import sys, json
d = json.load(sys.stdin)['data']
print('load groups:', len(d.get('load', [])))
for g in d.get('load', []):
    print(f'  {g[\"topology\"]}: {[n[\"node_type\"] for n in g[\"nodes\"][:5]]}')"

# Rack deps should NOT have duplicate SWGR entries
curl -s "http://localhost:8889/api/trace/dependencies/RACK-R1-Z1-R1-01?levels=6" | python3 -c "
import sys, json
d = json.load(sys.stdin)['data']
ids = [n['node_id'] for g in d.get('upstream', []) for n in g['nodes']]
dupes = [x for x in set(ids) if ids.count(x) > 1]
print('duplicates:', dupes if dupes else 'none')"
```

## Todo List

- [x] Add Rack impact rules to Impacts.csv
- [x] Fix FindUpstreamNodes: DISTINCT ON instead of GROUP BY parent_node_id
- [x] Fix FindDownstreamNodes: same pattern
- [x] Fix FindSpatialAncestorsOfType: remove GROUP BY, keep DISTINCT ON
- [x] Re-ingest model CSVs
- [x] Verify: Rack impacts non-empty
- [x] Verify: UPS impacts include Load nodes
- [x] Verify: No duplicate upstream nodes for Rack
- [x] Run existing tests: `cd backend && go test ./internal/service/...`

## Success Criteria

1. `GET /api/trace/impacts/RACK-*` returns `downstream` with Rack PDU, Server nodes
2. `GET /api/trace/impacts/UPS-01` returns `load` with Rack, Row, Zone nodes
3. `GET /api/trace/dependencies/RACK-*` returns zero duplicate node_ids in upstream
4. All existing unit tests pass

## Risk Assessment

- **Low risk:** CSV changes are additive (new rows only)
- **Medium risk:** SQL changes affect all trace queries — must test both upstream and downstream paths
- **Mitigation:** Run verification commands after each SQL change before moving to next step
