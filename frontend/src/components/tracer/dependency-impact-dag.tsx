'use client'

import "@xyflow/react/dist/style.css"
import { useState, useEffect, useCallback } from "react"
import { useQuery } from "@tanstack/react-query"
import {
  ReactFlow,
  ReactFlowProvider,
  useReactFlow,
  Background,
  Controls,
  useNodesState,
  useEdgesState,
  type Node,
  type Edge,
} from "@xyflow/react"
import { Loader2, Minus, Plus, Layers, Zap, Droplets, Building2, Package } from "lucide-react"
import { apiFetch } from "@/lib/api-client"
import { ThemeToggle } from "@/components/dashboard/theme-toggle"
import { traceToDAGElements, layoutDAG, filterTraceByTopologies } from "./dag-helpers"
import TracerNode from "./dag-node"
import TracerEdge from "./dag-edge"
import DAGSearch from "./dag-search"
import DagDetailPopup from "./dag-detail-popup"
import type { TraceResponse, TracerNodeData } from "./dag-types"

// Register custom node and edge types
const nodeTypes = { tracerNode: TracerNode }
const edgeTypes = { tracerEdge: TracerEdge }

interface ApiWrapper<T> {
  data: T
}

function DependencyImpactDAGInner() {
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null)
  const [depth, setDepth] = useState(2)
  const [selectedTopos, setSelectedTopos] = useState<Set<string>>(new Set(["electrical", "cooling"]))
  const [popupData, setPopupData] = useState<TracerNodeData | null>(null)
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([])
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([])
  const { fitView } = useReactFlow()

  // Fetch combined trace (upstream + local + downstream + load)
  const traceQuery = useQuery({
    queryKey: ["trace-full", selectedNodeId, depth],
    queryFn: () =>
      apiFetch<ApiWrapper<TraceResponse>>(
        `/api/trace/full/${selectedNodeId}?levels=${depth}`
      ).then((res) => res.data),
    enabled: !!selectedNodeId,
    staleTime: 60_000,
  })

  // Sync graph when query settles
  useEffect(() => {
    if (!selectedNodeId) {
      setNodes([])
      setEdges([])
      return
    }
    if (traceQuery.isLoading) return

    const trace = traceQuery.data ?? null
    const filtered = filterTraceByTopologies(trace, selectedTopos)
    const { nodes: rawNodes, edges: rawEdges } = traceToDAGElements(filtered)
    const { nodes: laidNodes, edges: laidEdges } = layoutDAG(rawNodes, rawEdges)
    const nodesWithClick = laidNodes.map((n) => ({
      ...n,
      data: { ...n.data, onNodeClick: (d: TracerNodeData) => setPopupData(d) },
    }))
    setNodes(nodesWithClick)
    setEdges(laidEdges)
    setTimeout(() => fitView({ padding: 0.2, duration: 400 }), 50)
  }, [selectedNodeId, traceQuery.data, traceQuery.isLoading, selectedTopos])

  const handleSelect = useCallback((nodeId: string) => {
    setSelectedNodeId(nodeId)
  }, [])

  const handleClear = useCallback(() => {
    setSelectedNodeId(null)
    setNodes([])
    setEdges([])
  }, [])

  const toggleTopo = useCallback((key: string) => {
    setSelectedTopos((prev) => {
      const next = new Set(prev)
      if (next.has(key)) next.delete(key)
      else next.add(key)
      return next
    })
  }, [])

  const isLoading = !!selectedNodeId && traceQuery.isFetching
  const isEmpty = !selectedNodeId

  const topoChips = [
    { key: "electrical", label: "Elec", icon: <Zap className="h-3 w-3" />, color: "#F97316", bg: "rgba(249,115,22,0.15)" },
    { key: "cooling", label: "Cool", icon: <Droplets className="h-3 w-3" />, color: "#06B6D4", bg: "rgba(6,182,212,0.15)" },
    { key: "spatial", label: "Spatial", icon: <Building2 className="h-3 w-3" />, color: "#8B5CF6", bg: "rgba(139,92,246,0.15)" },
    { key: "whitespace", label: "WS", icon: <Package className="h-3 w-3" />, color: "#10B981", bg: "rgba(16,185,129,0.15)" },
  ]

  return (
    <div className="relative h-screen h-dvh overflow-hidden bg-card flex flex-col">
      {/* Top toolbar: depth + search + theme toggle in one row */}
      <div className="relative z-10 flex items-center gap-1.5 sm:gap-2 px-2 sm:px-3 py-1.5 sm:py-2 border-b border-border bg-card shrink-0">
        {/* Depth control */}
        <div className="flex items-center gap-1 rounded-lg border border-border bg-card px-1.5 py-1 shrink-0">
          <Layers className="h-3.5 w-3.5 text-muted-foreground" />
          <span className="text-xs font-medium text-muted-foreground hidden sm:inline">Depth</span>
          <button
            onClick={() => setDepth((d) => Math.max(1, d - 1))}
            disabled={depth <= 1}
            className="h-6 w-6 flex items-center justify-center rounded hover:bg-muted disabled:opacity-30 transition-colors"
          >
            <Minus className="h-3 w-3" />
          </button>
          <span className="text-sm font-bold w-4 text-center">{depth}</span>
          <button
            onClick={() => setDepth((d) => Math.min(6, d + 1))}
            disabled={depth >= 6}
            className="h-6 w-6 flex items-center justify-center rounded hover:bg-muted disabled:opacity-30 transition-colors"
          >
            <Plus className="h-3 w-3" />
          </button>
        </div>

        {/* Search bar - takes remaining space */}
        <DAGSearch onSelect={handleSelect} onClear={handleClear} />

        {/* Topology filter chips */}
        <div className="flex items-center gap-1 shrink-0">
          {topoChips.map((t) => {
            const active = selectedTopos.has(t.key)
            return (
              <button
                key={t.key}
                onClick={() => toggleTopo(t.key)}
                className="flex items-center gap-1 px-1.5 py-0.5 rounded-md text-xs font-medium border transition-all"
                style={{
                  borderColor: active ? t.color : "hsl(var(--border))",
                  backgroundColor: active ? t.bg : "transparent",
                  color: active ? t.color : "hsl(var(--muted-foreground))",
                  opacity: active ? 1 : 0.5,
                }}
              >
                {t.icon}
                <span className="hidden sm:inline">{t.label}</span>
              </button>
            )
          })}
        </div>

        {/* Theme toggle */}
        <div className="shrink-0">
          <ThemeToggle />
        </div>
      </div>

      {/* Loading overlay */}
      {isLoading && (
        <div className="absolute inset-0 z-20 flex items-center justify-center bg-background/50 backdrop-blur-sm">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      )}

      {/* Empty state */}
      {isEmpty && (
        <div className="absolute inset-0 flex items-center justify-center">
          <div className="text-center space-y-2">
            <p className="text-muted-foreground text-sm">
              Search for a node to trace dependencies and impacts
            </p>
            <p className="text-muted-foreground/60 text-xs">
              Upstream dependencies shown in blue, downstream impacts in red
            </p>
          </div>
        </div>
      )}

      <div className="flex-1 relative touch-none">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          nodeTypes={nodeTypes}
          edgeTypes={edgeTypes}
          defaultViewport={{ x: 0, y: 0, zoom: 0.8 }}
          minZoom={0.2}
          maxZoom={2}
          panOnScroll={false}
          panOnDrag
          zoomOnPinch
          nodesDraggable={false}
          nodesConnectable={false}
          elementsSelectable={false}
          proOptions={{ hideAttribution: true }}
        >
          <Background gap={20} size={1} color="hsl(var(--border) / 0.3)" />
          <Controls
            position="bottom-right"
            className="!bg-card !border-border !shadow-lg [&>button]:!bg-card [&>button]:!border-border [&>button]:!fill-foreground"
          />
        </ReactFlow>
      </div>

      {/* Node detail popup */}
      {popupData && <DagDetailPopup data={popupData} onClose={() => setPopupData(null)} />}
    </div>
  )
}

// Wrap with ReactFlowProvider for context
export default function DependencyImpactDAG() {
  return (
    <ReactFlowProvider>
      <DependencyImpactDAGInner />
    </ReactFlowProvider>
  )
}
