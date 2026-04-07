---
title: "Load-Capacity Calculator"
description: "Ingest rack/capacity CSV, compute bottom-up load aggregation, expose capacity API + DAG badges"
status: complete
priority: P1
effort: 12h
branch: main
tags: [backend, capacity, calculator, csv, frontend, dag]
created: 2026-04-07
---

# Load-Capacity Calculator

## Goal
Ingest per-node capacity data from CSV, compute bottom-up load aggregation, expose capacity metrics via API, and display utilization badges on existing Tracer DAG.

## Context
- Brainstorm report: `plans/reports/brainstorm-260407-1542-load-capacity-calculator.md`
- Source CSV: `blueprint/ISET capacity - rack load flow.csv` (657 nodes, 35 columns)
- Backend: Go/Gin/GORM clean architecture
- Frontend: Next.js + ReactFlow DAG

## Phases

| # | Phase | Status | Est. | Files |
|---|-------|--------|------|-------|
| 1 | [Data Model + CSV Ingestion](phase-01-data-model-csv-ingestion.md) | complete | 2h | model, parser, ingestion svc, repo, migration |
| 2 | [Load Calculator](phase-02-load-calculator.md) | complete | 3h | calculator service, spatial queries |
| 3 | [Capacity API](phase-03-capacity-api.md) | complete | 2h | handler, router, trace integration |
| 4 | [Frontend DAG Badges](phase-04-frontend-dag-badges.md) | complete | 3h | types, node, panel, detail popup |
| 5 | [Testing](phase-05-testing.md) | complete | 2h | integration + unit tests |

## Key Dependencies
- Phase 2 depends on Phase 1 (needs node_variables table + data)
- Phase 3 depends on Phase 2 (needs computed metrics)
- Phase 4 depends on Phase 3 (needs API endpoints)
- Phase 5 runs after Phase 2 (backend tests) and Phase 4 (frontend smoke)

## Architecture Summary
```
CSV -> Parser -> node_variables table (raw, source=csv_import)
                        |
                 Calculator (bottom-up aggregation)
                        |
                 node_variables table (computed, source=computed)
                        |
                 Capacity API -> Frontend DAG badges
```
