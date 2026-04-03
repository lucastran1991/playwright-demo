---
title: "Dependency & Impact DAG UI"
description: "ReactFlow+Dagre DAG visualization page at /dashboard/tracer for tracing node dependencies and impacts"
status: pending
priority: P1
effort: 6h
branch: main
tags: [frontend, reactflow, dagre, dag, visualization, tracer]
created: 2026-04-04
---

# Dependency & Impact DAG UI

## Overview

Add `/dashboard/tracer` page with search bar + ReactFlow DAG showing upstream dependencies (left) -> source (center) -> downstream impacts (right) for any selected capacity node.

## Context

- [Brainstorm Report](../reports/brainstorm-260404-0118-dependency-impact-dag-ui.md)
- [Reference Component](../../PipelineDAG.tsx) -- working ReactFlow+Dagre pattern
- Backend APIs already built and ready

## Dependencies

- New packages: `@xyflow/react`, `@dagrejs/dagre`
- Existing: TanStack Query, lucide-react, Tailwind CSS 4

## File Map

### New Files
| File | Purpose | Est. Lines |
|------|---------|-----------|
| `frontend/src/app/(dashboard)/tracer/page.tsx` | Route page | ~20 |
| `frontend/src/components/tracer/dag-types.ts` | TypeScript types | ~40 |
| `frontend/src/components/tracer/dag-helpers.tsx` | Dagre layout + API->ReactFlow mapping | ~120 |
| `frontend/src/components/tracer/dag-node.tsx` | Custom ReactFlow node | ~90 |
| `frontend/src/components/tracer/dag-edge.tsx` | Custom ReactFlow edge | ~50 |
| `frontend/src/components/tracer/dag-search.tsx` | Autocomplete search | ~100 |
| `frontend/src/components/tracer/dependency-impact-dag.tsx` | Main DAG component | ~150 |

### Modified Files
| File | Change |
|------|--------|
| `frontend/src/components/dashboard/sidebar-nav.tsx` | Add "Tracer" nav item with Network icon |

## Phases

| # | Phase | Status | File |
|---|-------|--------|------|
| 1 | Install deps + types + helpers | pending | [phase-01](phase-01-types-and-helpers.md) |
| 2 | Custom node + edge components | pending | [phase-02](phase-02-node-and-edge.md) |
| 3 | Search component | pending | [phase-03](phase-03-search.md) |
| 4 | Main DAG + page route + sidebar | pending | [phase-04](phase-04-main-dag-and-route.md) |

## Key Decisions

1. LR Dagre layout (upstream left -> source center -> downstream right)
2. Color coding: electrical=orange, cooling=blue, spatial=green
3. Edge colors: blue=dependency, red=impact, dashed=local
4. `@ts-nocheck` for @xyflow/react (React 19 types not yet released, same as PipelineDAG.tsx)
5. No auth needed for read endpoints
6. Default levels=2 to avoid crowding; expandable later

## Success Criteria

- Search node by partial node_id or name
- DAG renders within 500ms of selection
- Clear upstream/source/downstream layout
- Color coding distinguishes node topologies
- Works on tablet+ screens
