---
title: "Capacity Nodes, Dependency Rules, Impact Rules Ingestion + Dependency Tracer API"
description: "Ingest 3 model CSVs into PostgreSQL and build Dependency Tracer API for upstream/downstream resolution"
status: completed
priority: P1
effort: 6h
branch: main
tags: [backend, csv-ingestion, api, dependency-tracer, postgresql]
created: 2026-04-04
completed: 2026-04-04
---

# Capacity Nodes, Dependency & Impact Model + Tracer API

## Summary

Ingest 3 type-level metadata CSVs (Capacity Nodes, Dependencies, Impacts) and build a Dependency Tracer API that resolves actual node instances from type-level rules against the existing blueprint topology.

## Context

- Brainstorm report: `plans/reports/brainstorm-260404-0025-capacity-dependency-impact-model.md`
- Existing blueprint ingestion: `backend/internal/service/blueprint_ingestion_service.go`
- Blueprint topology tables: `blueprint_nodes`, `blueprint_edges`, `blueprint_types`, `blueprint_node_memberships`

## Phases

| # | Phase | Status | Effort | File |
|---|-------|--------|--------|------|
| 1 | Database models + migration | completed | 30m | [phase-01](phase-01-database-models.md) |
| 2 | Model CSV parser + ingestion service | completed | 1.5h | [phase-02](phase-02-csv-ingestion.md) |
| 3 | Tracer repository (recursive CTE) | completed | 1.5h | [phase-03](phase-03-tracer-repository.md) |
| 4 | Dependency tracer service | completed | 1h | [phase-04](phase-04-tracer-service.md) |
| 5 | API handlers + routes + wiring | completed | 1h | [phase-05](phase-05-api-handlers-routes.md) |
| 6 | Testing | completed | 30m | [phase-06](phase-06-testing.md) |

## Key Dependencies

- Requires blueprint data already ingested (nodes, edges exist in DB)
- Reuses existing `readCSV()` from `blueprint_csv_parser.go`
- Config already has `BlueprintDir` -- model CSVs live in same root `./blueprint/`

## New Files

```
backend/internal/model/capacity_node_type.go
backend/internal/model/dependency_rule.go
backend/internal/model/impact_rule.go
backend/internal/service/model_csv_parser.go
backend/internal/service/model_ingestion_service.go
backend/internal/service/dependency_tracer.go
backend/internal/handler/tracer_handler.go
backend/internal/repository/tracer_repository.go
```

## Modified Files

```
backend/internal/database/database.go        -- add 3 models to AutoMigrate
backend/internal/router/router.go             -- add tracer routes
backend/cmd/server/main.go                    -- wire tracer dependencies
backend/internal/config/config.go             -- add ModelDir config
```

## Unresolved Questions

1. **Local dependency resolution**: exact strategy for finding "local" nodes (direct edge neighbors vs siblings under same parent). Current plan: direct edge neighbors in same topology.
2. **Cross-domain tracing**: whether upstream walk runs independently per topology or merged. Current plan: independent queries per topology, merged in response.
3. **Load impact scope**: confirmation that row-level scope means "highlight all racks in the row". Current plan: yes.
