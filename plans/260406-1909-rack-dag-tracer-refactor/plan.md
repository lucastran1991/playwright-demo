---
status: complete
branch: main
created: 2026-04-06
completed: 2026-04-06
slug: rack-dag-tracer-refactor
---

# Rack DAG Tracer Refactor

## Summary

Fix cross-topology dependency/impact resolution for Rack and spatial/whitespace nodes. Eliminate duplicate nodes, fix Load resolution bug, add combined `/trace/full` API.

## Context

- [Brainstorm Report](../reports/brainstorm-260406-1909-rack-dependency-impact-dag-fix.md)
- Backend: Go/Gin/GORM/PostgreSQL (port 8889)
- Frontend: Next.js/ReactFlow/Dagre
- Approach: Spatial-bridge traversal (no schema changes)

## Phases

| # | Phase | Status | Priority | Effort |
|---|-------|--------|----------|--------|
| 1 | [Fix data + SQL bugs](phase-01-fix-data-and-sql-bugs.md) | complete | critical | small |
| 2 | [Refactor TraceDependencies](phase-02-refactor-trace-dependencies.md) | complete | high | medium |
| 3 | [Refactor TraceImpacts](phase-03-refactor-trace-impacts.md) | complete | high | medium |
| 4 | [Add /trace/full API](phase-04-add-trace-full-api.md) | complete | medium | small |
| 5 | [Frontend DAG fixes](phase-05-frontend-dag-fixes.md) | complete | high | medium |

## Dependencies

- Phase 1 unblocks all other phases ✓
- Phase 2 and 3 can run in parallel ✓
- Phase 4 depends on 2 + 3 ✓
- Phase 5 depends on 4 ✓

## Key Files

### Backend
- `backend/internal/service/dependency_tracer.go` (311 lines) — main service
- `backend/internal/service/dependency_tracer_helpers.go` (69 lines) — grouping/filtering
- `backend/internal/repository/tracer_repository.go` (179 lines) — SQL queries
- `backend/internal/handler/tracer_handler.go` (97 lines) — HTTP handlers
- `backend/internal/router/router.go` — route registration
- `blueprint/Impacts.csv` — impact rules data
- `blueprint/Dependencies.csv` — dependency rules data

### Frontend
- `frontend/src/components/tracer/dag-helpers.tsx` (219 lines) — layout + graph building
- `frontend/src/components/tracer/dependency-impact-dag.tsx` (204 lines) — main component
- `frontend/src/components/tracer/dag-types.ts` (48 lines) — TypeScript types
