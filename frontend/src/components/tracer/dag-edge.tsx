// @ts-nocheck
'use client'

import { BaseEdge, getSmoothStepPath } from "@xyflow/react"

export default function TracerEdge({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  style = {},
  markerEnd,
}) {
  const [edgePath] = getSmoothStepPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
    borderRadius: 8,
  })

  return (
    <>
      {/* Glow layer */}
      <BaseEdge
        id={`${id}-glow`}
        path={edgePath}
        style={{
          ...style,
          strokeWidth: (style.strokeWidth ?? 2) + 4,
          strokeOpacity: 0.15,
          strokeDasharray: style.strokeDasharray,
        }}
      />
      {/* Main edge with arrow */}
      <BaseEdge
        id={id}
        path={edgePath}
        style={style}
        markerEnd={markerEnd}
      />
    </>
  )
}
