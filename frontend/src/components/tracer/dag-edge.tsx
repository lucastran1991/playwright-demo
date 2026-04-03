// @ts-nocheck
'use client'

import { BaseEdge, getSmoothStepPath } from "@xyflow/react"

interface TracerEdgeProps {
  id: string
  sourceX: number
  sourceY: number
  targetX: number
  targetY: number
  sourcePosition: any
  targetPosition: any
  style?: React.CSSProperties
  data?: { label?: string }
}

export default function TracerEdge({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  style = {},
  data,
}: TracerEdgeProps) {
  const [edgePath] = getSmoothStepPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
    borderRadius: 8,
  })

  const strokeColor = (style.stroke as string) || "#6B7280"

  return (
    <>
      {/* Glow layer */}
      <BaseEdge
        id={`${id}-glow`}
        path={edgePath}
        style={{
          ...style,
          strokeWidth: (style.strokeWidth as number ?? 2) + 4,
          strokeOpacity: 0.15,
          strokeDasharray: style.strokeDasharray as string | undefined,
        }}
      />
      {/* Main edge */}
      <BaseEdge
        id={id}
        path={edgePath}
        style={style}
      />
      {/* Optional topology label */}
      {data?.label && (
        <text>
          <textPath
            href={`#${id}`}
            startOffset="50%"
            textAnchor="middle"
            style={{ fontSize: 9, fill: strokeColor, opacity: 0.7 }}
          >
            {data.label}
          </textPath>
        </text>
      )}
    </>
  )
}
