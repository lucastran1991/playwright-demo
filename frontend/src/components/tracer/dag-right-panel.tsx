'use client'

import { useCallback, useState } from "react"
import { Minus, Plus, Layers, Zap, Droplets, Building2, Package, Download, X, Loader2 } from "lucide-react"
import { ThemeToggle } from "@/components/dashboard/theme-toggle"

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

export default function DagRightPanel({ open, onClose, depth, onDepthChange, selectedTopos, onTopoToggle, selectedNodeId }: DagRightPanelProps) {
  const [exporting, setExporting] = useState(false)

  // Download bulk XLSX from backend — one sheet per capacity node type
  const handleExportXlsx = useCallback(async () => {
    setExporting(true)
    try {
      const apiBase = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8889"
      const res = await fetch(`${apiBase}/api/trace/export/xlsx?levels=${depth}`)
      if (!res.ok) throw new Error("Export failed")
      const blob = await res.blob()
      const url = URL.createObjectURL(blob)
      const a = document.createElement("a")
      a.href = url
      a.download = "trace-all-models.xlsx"
      a.click()
      URL.revokeObjectURL(url)
    } catch (err) {
      console.error("XLSX export failed:", err)
    } finally {
      setExporting(false)
    }
  }, [depth])

  // Download single-node CSV export
  const handleExportCsv = useCallback(async () => {
    if (!selectedNodeId) return
    const apiBase = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8889"
    const res = await fetch(`${apiBase}/api/trace/full/${selectedNodeId}/export?levels=${depth}`)
    if (!res.ok) return
    const blob = await res.blob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement("a")
    a.href = url
    a.download = `trace-${selectedNodeId}.csv`
    a.click()
    URL.revokeObjectURL(url)
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
          <div className="space-y-1">
            <button
              onClick={handleExportXlsx}
              disabled={exporting}
              className="w-full flex items-center gap-2 px-2 py-1.5 rounded-md text-xs font-medium border border-border hover:bg-muted disabled:opacity-50 transition-colors"
            >
              {exporting ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Download className="h-3.5 w-3.5" />}
              {exporting ? "Exporting..." : "Download All (XLSX)"}
            </button>
            <button
              onClick={handleExportCsv}
              disabled={!selectedNodeId}
              className="w-full flex items-center gap-2 px-2 py-1.5 rounded-md text-xs font-medium border border-border hover:bg-muted disabled:opacity-30 transition-colors"
            >
              <Download className="h-3.5 w-3.5" />
              Current Node (CSV)
            </button>
          </div>
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
