'use client'

import { useCallback } from "react"
import { Minus, Plus, Layers, Zap, Droplets, Building2, Package, Download, X } from "lucide-react"
import { ThemeToggle } from "@/components/dashboard/theme-toggle"
import { apiFetch } from "@/lib/api-client"
import * as XLSX from "xlsx"
import type { TraceResponse } from "./dag-types"

interface ApiWrapper<T> {
  data: T
}

// Node types to export as separate sheets
const EXPORT_MODELS = [
  { type: "Rack", label: "Rack" },
  { type: "RPP", label: "RPP" },
  { type: "Room PDU", label: "Room PDU" },
  { type: "UPS", label: "UPS" },
  { type: "BESS", label: "BESS" },
  { type: "Switch Gear", label: "Switch Gear" },
  { type: "Utility Feed", label: "Utility Feed" },
  { type: "Generator", label: "Generator" },
  { type: "Air Zone", label: "Air Zone" },
  { type: "Liquid Loop", label: "Liquid Loop" },
  { type: "CDU", label: "CDU" },
  { type: "Cooling Distribution", label: "Cooling Dist" },
  { type: "Cooling Plant", label: "Cooling Plant" },
]

const TOPO_CHIPS = [
  { key: "electrical", label: "Electrical", icon: <Zap className="h-3.5 w-3.5" />, color: "#F97316", bg: "rgba(249,115,22,0.15)" },
  { key: "cooling", label: "Cooling", icon: <Droplets className="h-3.5 w-3.5" />, color: "#06B6D4", bg: "rgba(6,182,212,0.15)" },
  { key: "spatial", label: "Spatial", icon: <Building2 className="h-3.5 w-3.5" />, color: "#8B5CF6", bg: "rgba(139,92,246,0.15)" },
  { key: "whitespace", label: "Whitespace", icon: <Package className="h-3.5 w-3.5" />, color: "#10B981", bg: "rgba(16,185,129,0.15)" },
]

interface DagRightPanelProps {
  open: boolean
  onClose: () => void
  depth: number
  onDepthChange: (d: number) => void
  selectedTopos: Set<string>
  onTopoToggle: (key: string) => void
  selectedNodeId: string | null
}

// Flatten a TraceResponse into CSV-style rows for a sheet
function traceToRows(resp: TraceResponse) {
  const rows: Record<string, string>[] = []
  const src = resp.source

  const addNodes = (direction: string, groups: { level?: number; topology?: string; nodes: { node_id: string; name: string; node_type: string; parent_node_id?: string }[] }[]) => {
    for (const g of groups) {
      for (const n of g.nodes) {
        rows.push({
          src_node_id: src.node_id,
          src_node_name: src.name,
          direction,
          level: String(g.level ?? 0),
          topology: g.topology ?? "",
          node_id: n.node_id,
          node_name: n.name,
          node_type: n.node_type,
          parent_node_id: n.parent_node_id ?? "",
        })
      }
    }
  }

  // Source row
  rows.push({
    src_node_id: src.node_id, src_node_name: src.name, direction: "source",
    level: "0", topology: src.topology ?? "", node_id: src.node_id,
    node_name: src.name, node_type: src.node_type, parent_node_id: "",
  })

  if (resp.upstream) addNodes("upstream", resp.upstream)
  if (resp.local) addNodes("local", resp.local)
  if (resp.downstream) addNodes("downstream", resp.downstream)
  if (resp.load) addNodes("load", resp.load)

  return rows
}

export default function DagRightPanel({ open, onClose, depth, onDepthChange, selectedTopos, onTopoToggle, selectedNodeId }: DagRightPanelProps) {
  const handleExportXlsx = useCallback(async () => {
    if (!selectedNodeId) return

    // Fetch the current node's trace
    const resp = await apiFetch<ApiWrapper<TraceResponse>>(
      `/api/trace/full/${selectedNodeId}?levels=${depth}`
    ).then((r) => r.data)

    const wb = XLSX.utils.book_new()

    // Main trace sheet
    const mainRows = traceToRows(resp)
    const mainWs = XLSX.utils.json_to_sheet(mainRows)
    XLSX.utils.book_append_sheet(wb, mainWs, resp.source.node_type.slice(0, 31))

    // Additional sheets: trace each unique upstream/downstream node type
    const tracedTypes = new Map<string, string>() // node_type → first node_id
    for (const section of [resp.upstream, resp.downstream]) {
      for (const g of section ?? []) {
        for (const n of g.nodes) {
          if (!tracedTypes.has(n.node_type)) {
            tracedTypes.set(n.node_type, n.node_id)
          }
        }
      }
    }

    // Fetch trace for each unique model type (up to 10 to avoid hammering API)
    const entries = Array.from(tracedTypes.entries()).slice(0, 10)
    for (const [nodeType, nodeId] of entries) {
      try {
        const modelResp = await apiFetch<ApiWrapper<TraceResponse>>(
          `/api/trace/full/${nodeId}?levels=${depth}`
        ).then((r) => r.data)
        const rows = traceToRows(modelResp)
        if (rows.length > 0) {
          const ws = XLSX.utils.json_to_sheet(rows)
          // Sheet name max 31 chars, no special chars
          const sheetName = nodeType.replace(/[/\\?*[\]]/g, "").slice(0, 31)
          XLSX.utils.book_append_sheet(wb, ws, sheetName)
        }
      } catch {
        // Skip failed traces silently
      }
    }

    XLSX.writeFile(wb, `trace-${selectedNodeId}.xlsx`)
  }, [selectedNodeId, depth])

  return (
    <div
      className={`absolute top-0 right-0 h-full z-30 bg-card border-l border-border shadow-xl transition-transform duration-300 ${
        open ? "translate-x-0" : "translate-x-full"
      }`}
      style={{ width: 260 }}
    >
      {/* Panel header */}
      <div className="flex items-center justify-between px-3 py-2 border-b border-border">
        <span className="text-sm font-semibold">Settings</span>
        <button onClick={onClose} className="h-6 w-6 flex items-center justify-center rounded hover:bg-muted">
          <X className="h-4 w-4" />
        </button>
      </div>

      <div className="p-3 space-y-4 overflow-y-auto" style={{ maxHeight: "calc(100% - 41px)" }}>
        {/* Depth control */}
        <div>
          <label className="text-xs font-medium text-muted-foreground mb-1.5 block">Depth Level</label>
          <div className="flex items-center gap-1.5 rounded-lg border border-border bg-card px-2 py-1.5">
            <Layers className="h-3.5 w-3.5 text-muted-foreground" />
            <button
              onClick={() => onDepthChange(Math.max(1, depth - 1))}
              disabled={depth <= 1}
              className="h-6 w-6 flex items-center justify-center rounded hover:bg-muted disabled:opacity-30"
            >
              <Minus className="h-3 w-3" />
            </button>
            <span className="text-sm font-bold w-6 text-center">{depth}</span>
            <button
              onClick={() => onDepthChange(Math.min(6, depth + 1))}
              disabled={depth >= 6}
              className="h-6 w-6 flex items-center justify-center rounded hover:bg-muted disabled:opacity-30"
            >
              <Plus className="h-3 w-3" />
            </button>
          </div>
        </div>

        {/* Topology filters */}
        <div>
          <label className="text-xs font-medium text-muted-foreground mb-1.5 block">Topologies</label>
          <div className="space-y-1">
            {TOPO_CHIPS.map((t) => {
              const active = selectedTopos.has(t.key)
              return (
                <button
                  key={t.key}
                  onClick={() => onTopoToggle(t.key)}
                  className="w-full flex items-center gap-2 px-2 py-1.5 rounded-md text-xs font-medium border transition-all"
                  style={{
                    borderColor: active ? t.color : "hsl(var(--border))",
                    backgroundColor: active ? t.bg : "transparent",
                    color: active ? t.color : "hsl(var(--muted-foreground))",
                    opacity: active ? 1 : 0.5,
                  }}
                >
                  {t.icon}
                  {t.label}
                </button>
              )
            })}
          </div>
        </div>

        {/* Export */}
        <div>
          <label className="text-xs font-medium text-muted-foreground mb-1.5 block">Export</label>
          <button
            onClick={handleExportXlsx}
            disabled={!selectedNodeId}
            className="w-full flex items-center gap-2 px-2 py-1.5 rounded-md text-xs font-medium border border-border hover:bg-muted disabled:opacity-30 transition-colors"
          >
            <Download className="h-3.5 w-3.5" />
            Download XLSX
          </button>
        </div>

        {/* Theme */}
        <div>
          <label className="text-xs font-medium text-muted-foreground mb-1.5 block">Theme</label>
          <ThemeToggle />
        </div>
      </div>
    </div>
  )
}
