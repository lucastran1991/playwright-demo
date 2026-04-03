---
title: "Blueprint CSV Ingestion"
description: "Ingest 6 blueprint domain CSVs into PostgreSQL with generic loader, upsert logic, and REST API"
status: completed
priority: P1
effort: 6h
branch: main
tags: [backend, csv, ingestion, blueprint, go]
created: 2026-04-03
completed: 2026-04-03
---

# Blueprint CSV Ingestion

## Overview

Ingest 6 blueprint domain CSVs (Cooling, Electrical, Spatial, Operational, Whitespace, Deployment) into PostgreSQL. Each domain has Nodes + Edges files. Unified node table with membership join table handles cross-domain overlap (~5.3K nodes, ~6.3K edges).

## Context

- [Brainstorm Report](../reports/brainstorm-260403-2249-blueprint-ingestion-model.md)
- CSV source: `blueprint/Node & Edge/` (6 subdirectories)
- Schema: unified nodes + adjacency list edges + membership join table

## Dependencies

- PostgreSQL running with `app_dev` database
- Existing Go backend with Gin/GORM stack

## Phases

| # | Phase | Status | Effort | File |
|---|-------|--------|--------|------|
| 1 | Database Models | completed | 1h | [phase-01](phase-01-database-models.md) |
| 2 | Migration Updates | completed | 0.5h | [phase-02](phase-02-migration-updates.md) |
| 3 | CSV Parser Service | completed | 1.5h | [phase-03](phase-03-csv-parser-service.md) |
| 4 | Ingestion Service | completed | 1.5h | [phase-04](phase-04-ingestion-service.md) |
| 5 | API Handlers & Routes | completed | 1h | [phase-05](phase-05-api-handlers-routes.md) |
| 6 | Testing | completed | 0.5h | [phase-06](phase-06-testing.md) |

## Key Design Decisions

1. Single `blueprint_nodes` table -- one row per unique Node ID across all domains
2. `blueprint_node_memberships` join table stores domain-specific org_path
3. `blueprint_edges` scoped per domain with adjacency list pattern
4. Generic CSV loader auto-discovers folders, no hardcoded domain list
5. Upsert via `ON CONFLICT DO UPDATE` for idempotent re-ingestion
6. All new files use snake_case naming, each under 200 lines

## File Map (new files to create)

```
backend/internal/
  model/blueprint_type.go
  model/blueprint_node.go
  model/blueprint_node_membership.go
  model/blueprint_edge.go
  repository/blueprint_repository.go
  service/blueprint_csv_parser.go
  service/blueprint_ingestion_service.go
  handler/blueprint_handler.go
```

## Success Criteria

- All 6 domains ingested without errors
- Re-running ingestion produces identical state (idempotent)
- Cross-domain node queries work via memberships
- Tree traversal via recursive CTE returns correct hierarchy
