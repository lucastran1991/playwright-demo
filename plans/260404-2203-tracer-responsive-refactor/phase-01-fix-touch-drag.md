# Phase 1: Fix Touch Drag on Small Devices

## Overview
- **Priority:** P0
- **Status:** Complete
- Touch panning broken/unreliable on mobile because nodes consume touch events and `nodesDraggable` defaults to `true`

## Root Cause Analysis
1. ReactFlow `nodesDraggable` defaults to `true` — touching a node initiates node drag instead of canvas pan
2. Nodes are 180-200px wide, consuming most of the viewport on mobile — very little empty canvas to touch
3. DAG layout is auto-computed by Dagre — dragging individual nodes is meaningless (breaks layout)
4. `panOnDrag={true}` only works when touching empty canvas, not nodes

## Key Insight
Setting `nodesDraggable={false}` makes all touch/mouse drags on nodes propagate to the canvas pan handler. This is the correct behavior for a read-only DAG visualization.

## Related Code Files
- `frontend/src/components/tracer/dependency-impact-dag.tsx` (ReactFlow props)

## Implementation Steps

1. Add `nodesDraggable={false}` to `<ReactFlow>` props — disables individual node dragging so all drag gestures pan the canvas
2. Add `nodesConnectable={false}` — prevents accidental edge creation on touch
3. Add `elementsSelectable={false}` — prevents selection box interfering with pan on touch
4. Keep `panOnDrag={true}` and `zoomOnPinch={true}` — these are correct
5. Keep `touch-none` CSS class — required so browser doesn't intercept touch for scroll

## Todo
- [x] Add `nodesDraggable={false}` to ReactFlow
- [x] Add `nodesConnectable={false}` to ReactFlow
- [x] Add `elementsSelectable={false}` to ReactFlow
- [x] Verify touch pan works on mobile viewport via chrome-devtools

## Success Criteria
- Single-finger drag on any part of canvas (including over nodes) pans the view
- Pinch-to-zoom works
- Node tap still opens detail popup
- No accidental node repositioning
