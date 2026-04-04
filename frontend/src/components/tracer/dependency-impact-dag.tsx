// @ts-nocheck
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
} from "@xyflow/react"
import { Loader2, Minus, Plus, Layers } from "lucide-react"
import { apiFetch } from "@/lib/api-client"
import { ThemeToggle } from "@/components/dashboard/theme-toggle"
import { traceToDAGElements, layoutDAG } from "./dag-helpers"
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
  const [popupData, setPopupData] = useState<TracerNodeData | null>(null)
  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])
  const { fitView } = useReactFlow()

  // Fetch dependency trace (upstream + local)
  const depQuery = useQuery({
    queryKey: ["trace-deps", selectedNodeId, depth],
    queryFn: () =>
      apiFetch<ApiWrapper<TraceResponse>>(
        `/api/trace/dependencies/${selectedNodeId}?levels=${depth}&include_local=true`
      ).then((res) => res.data),
    enabled: !!selectedNodeId,
    staleTime: 60_000,
  })

  // Fetch impact trace (downstream + load)
  const impactQuery = useQuery({
    queryKey: ["trace-impact", selectedNodeId, depth],
    queryFn: () =>
      apiFetch<ApiWrapper<TraceResponse>>(
        `/api/trace/impacts/${selectedNodeId}?levels=${depth}`
      ).then((res) => res.data),
    enabled: !!selectedNodeId,
    staleTime: 60_000,
  })

  // Sync graph when both queries settle
  useEffect(() => {
    if (!selectedNodeId) {
      setNodes([])
      setEdges([])
      return
    }
    // Wait until at least one has data (other may 404 gracefully)
    if (depQuery.isLoading && impactQuery.isLoading) return

    const dep = depQuery.data ?? null
    const impact = impactQuery.data ?? null
    const { nodes: rawNodes, edges: rawEdges } = traceToDAGElements(dep, impact)
    const { nodes: laidNodes, edges: laidEdges } = layoutDAG(rawNodes, rawEdges)
    // Inject click handler into each node's data
    const nodesWithClick = laidNodes.map((n) => ({
      ...n,
      data: { ...n.data, onNodeClick: (d: TracerNodeData) => setPopupData(d) },
    }))
    setNodes(nodesWithClick)
    setEdges(laidEdges)
    // Center view after ReactFlow measures new nodes (needs slight delay)
    setTimeout(() => fitView({ padding: 0.2, duration: 400 }), 50)
  }, [
    selectedNodeId,
    depQuery.data,
    impactQuery.data,
    depQuery.isLoading,
    impactQuery.isLoading,
  ])

  const handleSelect = useCallback((nodeId: string) => {
    setSelectedNodeId(nodeId)
  }, [])

  const handleClear = useCallback(() => {
    setSelectedNodeId(null)
    setNodes([])
    setEdges([])
  }, [])

  const isLoading = !!selectedNodeId && (depQuery.isFetching || impactQuery.isFetching)
  const isEmpty = !selectedNodeId

  return (
    <div className="relative h-screen overflow-hidden bg-card flex flex-col">
      {/* Top toolbar: depth + search + theme toggle in one row */}
      <div className="relative z-10 flex items-center gap-2 px-3 py-2 border-b border-border bg-card shrink-0">
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

      <div className="flex-1 relative">
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
          proOptions={{ hideAttribution: true }}
        >
          <Background gap={20} size={1} color="hsl(var(--border) / 0.3)" />
          <Controls
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
