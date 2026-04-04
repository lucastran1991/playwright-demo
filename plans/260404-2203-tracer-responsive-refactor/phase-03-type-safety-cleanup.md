# Phase 3: Type Safety Cleanup

## Overview
- **Priority:** P1
- **Status:** Complete
- Remove `@ts-nocheck` from 3 files and add proper TypeScript types

## Related Code Files
- Modify: `frontend/src/components/tracer/dependency-impact-dag.tsx`
- Modify: `frontend/src/components/tracer/dag-node.tsx`
- Modify: `frontend/src/components/tracer/dag-edge.tsx`

## Implementation Steps

### 3.1 `dag-edge.tsx`
1. Remove `// @ts-nocheck`
2. Add proper typed props using XyFlow's `EdgeProps` type:
   ```ts
   import type { EdgeProps } from "@xyflow/react"
   export default function TracerEdge({ id, sourceX, sourceY, targetX, targetY, sourcePosition, targetPosition, style = {}, markerEnd }: EdgeProps) {
   ```

### 3.2 `dag-node.tsx`
1. Remove `// @ts-nocheck`
2. Props already typed via `{ data: TracerNodeData }` — verify no TS errors remain
3. If XyFlow expects `NodeProps`, use:
   ```ts
   import type { NodeProps } from "@xyflow/react"
   export default function TracerNode({ data }: NodeProps<Node<TracerNodeData>>) {
   ```

### 3.3 `dependency-impact-dag.tsx`
1. Remove `// @ts-nocheck`
2. Fix any type errors surfaced — likely around `setNodes`/`setEdges` generics and the `onNodeClick` injection in the `useEffect`
3. Type the `ApiWrapper` properly or replace with direct typing

## Todo
- [x] dag-edge.tsx: remove @ts-nocheck, add EdgeProps type
- [x] dag-node.tsx: remove @ts-nocheck, use NodeProps generic
- [x] dependency-impact-dag.tsx: remove @ts-nocheck, fix type errors
- [x] Run `pnpm tsc --noEmit` to verify zero type errors in tracer files

## Success Criteria
- No `@ts-nocheck` in any tracer file
- `pnpm tsc --noEmit` passes with no errors in tracer components
- No runtime behavior changes
