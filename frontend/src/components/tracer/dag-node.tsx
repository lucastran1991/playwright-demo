'use client'

import { Handle, Position, type NodeProps, type Node } from "@xyflow/react"
import { TOPOLOGY_CONFIG, getTopologyKey, getUtilColor, DIRECTION_COLORS } from "./dag-helpers"
import type { TracerNodeData } from "./dag-types"

export default function TracerNode({ data }: NodeProps<Node<TracerNodeData>>) {
  const topoKey = getTopologyKey(data.topology)
  const config = TOPOLOGY_CONFIG[topoKey] ?? TOPOLOGY_CONFIG.default
  const { color, bg, icon } = config
  const cap = data.capacity

  const dirColor = DIRECTION_COLORS[data.direction] ?? DIRECTION_COLORS.source
  const sourceRing = data.isSource ? "shadow-lg shadow-amber-400/20" : ""
  const showLevel = !data.isSource && !data.isLocal && data.level > 0

  // RL layout: target handle on RIGHT (incoming from right), source handle on LEFT (outgoing to left)
  // Local nodes: target handle on TOP (receives from source above)
  // Source node: extra bottom handle for local connections
  const isLocal = data.direction === "local"

  return (
    <div
      className={`relative flex flex-col gap-0.5 px-2.5 sm:px-3 py-1.5 sm:py-2 border-2 border-solid ${sourceRing} min-w-[140px] sm:min-w-[180px] max-w-[200px] transition-all cursor-pointer hover:scale-[1.03] hover:shadow-lg`}
      style={{
        borderColor: color,
        backgroundColor: bg,
        outline: `2.5px solid ${dirColor}`,
        outlineOffset: "2px",
        borderRadius: "8px",
      }}
      onClick={() => data.onNodeClick?.(data)}
    >
      {/* Horizontal handles for upstream/downstream (RL layout) */}
      <Handle type="target" position={Position.Right} className="!border-0 !w-2 !h-2 !rounded-full" style={{ background: color }} />
      <Handle type="source" position={Position.Left} className="!border-0 !w-2 !h-2 !rounded-full" style={{ background: color }} />

      {/* Local node: top handle to receive edge from source */}
      {isLocal && (
        <Handle id="local-target" type="target" position={Position.Top} className="!border-0 !w-2 !h-2 !rounded-full" style={{ background: dirColor }} />
      )}

      {/* Source node: bottom handle to send edges to local nodes */}
      {data.isSource && (
        <Handle id="source-bottom" type="source" position={Position.Bottom} className="!border-0 !w-2 !h-2 !rounded-full" style={{ background: DIRECTION_COLORS.local }} />
      )}

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
    </div>
  )
}
