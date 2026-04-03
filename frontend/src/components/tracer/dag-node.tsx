// @ts-nocheck
'use client'

import { Handle, Position } from "@xyflow/react"
import { TOPOLOGY_CONFIG, getTopologyKey } from "./dag-helpers"
import type { TracerNodeData } from "./dag-types"

interface TracerNodeProps {
  data: TracerNodeData
}

export default function TracerNode({ data }: TracerNodeProps) {
  const topoKey = getTopologyKey(data.topology)
  const config = TOPOLOGY_CONFIG[topoKey] ?? TOPOLOGY_CONFIG.default
  const color = config.color
  const icon = config.icon

  const borderClass = data.isLocal
    ? "border-dashed border-2"
    : "border border-solid"

  const sourceHighlight = data.isSource
    ? "ring-2 ring-yellow-400/50 shadow-lg shadow-yellow-400/20"
    : ""

  return (
    <div
      className={`
        relative flex flex-col gap-1 rounded-lg bg-card px-3 py-2
        ${borderClass} ${sourceHighlight}
        min-w-[180px] max-w-[200px]
      `}
      style={{ borderColor: color }}
    >
      {/* Target handle - left side for LR layout */}
      <Handle
        type="target"
        position={Position.Left}
        className="!border-0 !w-2 !h-2 !rounded-full"
        style={{ background: color }}
      />

      {/* Header: topology icon + node_type */}
      <div className="flex items-center gap-1.5">
        <span style={{ color }}>{icon}</span>
        <span className="text-[10px] font-medium uppercase tracking-wide text-muted-foreground truncate">
          {data.nodeType}
        </span>
        {data.isSource && (
          <span className="ml-auto text-[10px] font-semibold text-yellow-500">SOURCE</span>
        )}
      </div>

      {/* node_id - bold */}
      <p className="text-xs font-bold leading-tight truncate" style={{ color }}>
        {data.nodeId}
      </p>

      {/* name - muted, smaller */}
      <p className="text-[11px] text-muted-foreground leading-tight truncate">
        {data.name}
      </p>

      {/* Source handle - right side for LR layout */}
      <Handle
        type="source"
        position={Position.Right}
        className="!border-0 !w-2 !h-2 !rounded-full"
        style={{ background: color }}
      />
    </div>
  )
}
