'use client'

import { Handle, Position, type NodeProps, type Node } from "@xyflow/react"
import { TOPOLOGY_CONFIG, getTopologyKey, getUtilColor } from "./dag-helpers"
import type { TracerNodeData } from "./dag-types"

export default function TracerNode({ data }: NodeProps<Node<TracerNodeData>>) {
  const topoKey = getTopologyKey(data.topology)
  const config = TOPOLOGY_CONFIG[topoKey] ?? TOPOLOGY_CONFIG.default
  const { color, bg, icon } = config
  const cap = data.capacity

  const borderStyle = data.isLocal ? "border-dashed" : "border-solid"
  const sourceRing = data.isSource ? "ring-2 ring-amber-400/60 shadow-lg shadow-amber-400/20" : ""
  const showLevel = !data.isSource && !data.isLocal && data.level > 0

  return (
    <div
      className={`relative flex flex-col gap-0.5 rounded-lg px-2.5 sm:px-3 py-1.5 sm:py-2 border-2 ${borderStyle} ${sourceRing} min-w-[140px] sm:min-w-[180px] max-w-[200px] transition-all cursor-pointer hover:scale-[1.03] hover:shadow-lg`}
      style={{ borderColor: color, backgroundColor: bg }}
      onClick={() => data.onNodeClick?.(data)}
    >
      <Handle type="target" position={Position.Left} className="!border-0 !w-2 !h-2 !rounded-full" style={{ background: color }} />

      {/* Level badge - top right */}
      {showLevel && (
        <span className="absolute -top-2 -right-2 px-1.5 py-0.5 rounded-full text-[9px] font-bold bg-card border border-border text-muted-foreground shadow-sm">
          L{data.level}
        </span>
      )}

      {/* Source badge */}
      {data.isSource && (
        <span className="absolute -top-2 -right-2 px-1.5 py-0.5 rounded-full text-[9px] font-bold bg-amber-400 text-black shadow-sm">
          SRC
        </span>
      )}

      {/* Header: icon + type */}
      <div className="flex items-center gap-1.5">
        <span style={{ color }}>{icon}</span>
        <span className="text-[10px] font-medium uppercase tracking-wide text-muted-foreground truncate">
          {data.nodeType}
        </span>
      </div>

      {/* Node ID */}
      <p className="text-xs font-bold leading-tight truncate" style={{ color }}>
        {data.nodeId}
      </p>

      {/* Name */}
      <p className="text-[11px] text-muted-foreground leading-tight truncate">
        {data.name}
      </p>

      {/* Capacity utilization bar */}
      {cap?.utilization_pct != null && (
        <div className="flex items-center gap-1 mt-0.5">
          <div
            className="h-1.5 flex-1 rounded-full bg-muted overflow-hidden"
            title={`${cap.allocated_load ?? 0}/${cap.rated_capacity ?? cap.design_capacity ?? 0} kW`}
          >
            <div
              className="h-full rounded-full transition-all"
              style={{
                width: `${Math.min(cap.utilization_pct, 100)}%`,
                backgroundColor: getUtilColor(cap.utilization_pct),
              }}
            />
          </div>
          <span
            className="text-[9px] font-bold tabular-nums shrink-0"
            style={{ color: getUtilColor(cap.utilization_pct) }}
          >
            {Math.round(cap.utilization_pct)}%
          </span>
        </div>
      )}

      <Handle type="source" position={Position.Right} className="!border-0 !w-2 !h-2 !rounded-full" style={{ background: color }} />
    </div>
  )
}
