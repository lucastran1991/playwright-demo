'use client'

import { X } from "lucide-react"
import { TOPOLOGY_CONFIG, getTopologyKey } from "./dag-helpers"
import type { TracerNodeData } from "./dag-types"

interface Props {
  data: TracerNodeData
  onClose: () => void
}

export default function DagDetailPopup({ data, onClose }: Props) {
  const topoKey = getTopologyKey(data.topology)
  const config = TOPOLOGY_CONFIG[topoKey] ?? TOPOLOGY_CONFIG.default
  const { color, icon } = config

  const roleLabel = data.isSource ? "Source Node" : data.isLocal ? "Local Dependency" : `Level ${data.level} Dependency`

  return (
    <div
      className="absolute inset-0 z-30 flex items-center justify-center bg-black/40 backdrop-blur-[2px]"
      onClick={onClose}
    >
      <div
        className="w-[calc(100%-1rem)] sm:w-[calc(100%-2rem)] max-w-[360px] max-h-[80dvh] overflow-x-hidden overflow-y-auto rounded-2xl border border-border bg-card shadow-2xl animate-in zoom-in-95 duration-150"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header with topology color bar */}
        <div className="h-1.5" style={{ backgroundColor: color }} />
        <div className="px-5 py-4 flex items-start gap-3">
          <div className="mt-0.5 p-2 rounded-lg" style={{ backgroundColor: `${color}20` }}>
            <span style={{ color }}>{icon}</span>
          </div>
          <div className="flex-1 min-w-0">
            <p className="text-[10px] font-semibold uppercase tracking-widest" style={{ color }}>
              {data.nodeType}
            </p>
            <h3 className="text-base font-bold text-foreground truncate">{data.nodeId}</h3>
            <p className="text-xs text-muted-foreground truncate">{data.name}</p>
          </div>
          <button
            onClick={onClose}
            className="shrink-0 w-7 h-7 flex items-center justify-center rounded-md hover:bg-muted text-muted-foreground"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        {/* Details grid */}
        <div className="px-5 pb-4 space-y-3">
          {/* Role & topology */}
          <div className="grid grid-cols-2 gap-2">
            <DetailCard label="Role" value={roleLabel} accent={color} />
            <DetailCard label="Topology" value={data.topology || "N/A"} accent={color} />
          </div>

          {/* Status indicators */}
          <div className="grid grid-cols-2 sm:grid-cols-3 gap-2">
            <StatusPill
              label={data.isSource ? "SOURCE" : data.isLocal ? "LOCAL" : `L${data.level}`}
              color={data.isSource ? "#FBBF24" : data.isLocal ? "#6B7280" : color}
            />
            <StatusPill
              label={topoKey.toUpperCase()}
              color={color}
            />
            <StatusPill
              label={`Ring ${data.ring}`}
              color="#8B5CF6"
            />
          </div>

          {/* Node ID copyable */}
          <div className="rounded-lg bg-muted/50 px-3 py-2 flex items-center gap-2">
            <span className="text-[10px] text-muted-foreground shrink-0">ID</span>
            <code className="text-xs font-mono font-medium text-foreground truncate flex-1">
              {data.nodeId}
            </code>
            <button
              onClick={() => navigator.clipboard.writeText(data.nodeId)}
              className="text-[10px] text-primary hover:underline shrink-0"
            >
              Copy
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

function DetailCard({ label, value, accent }: { label: string; value: string; accent: string }) {
  return (
    <div className="rounded-lg border border-border/50 px-3 py-2">
      <p className="text-[9px] font-semibold uppercase tracking-wider text-muted-foreground">{label}</p>
      <p className="text-sm font-medium truncate" style={{ color: accent }}>{value}</p>
    </div>
  )
}

function StatusPill({ label, color }: { label: string; color: string }) {
  return (
    <div
      className="rounded-full px-2.5 py-1 text-center text-[10px] font-bold tracking-wide"
      style={{ backgroundColor: `${color}15`, color, border: `1px solid ${color}30` }}
    >
      {label}
    </div>
  )
}
