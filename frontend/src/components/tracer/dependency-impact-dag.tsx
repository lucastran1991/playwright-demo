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
import { Loader2, Settings } from "lucide-react"
import { apiFetch } from "@/lib/api-client"
import { traceToDAGElements, layoutDAG, filterTraceByTopologies } from "./dag-helpers"
import TracerNode from "./dag-node"
import TracerEdge from "./dag-edge"
import DAGSearch from "./dag-search"
import DagDetailPopup from "./dag-detail-popup"
import DagRightPanel from "./dag-right-panel"
import type { TraceResponse, TracerNodeData } from "./dag-types"

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
  const [panelOpen, setPanelOpen] = useState(false)
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([])
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([])
  const { fitView } = useReactFlow()

  const traceQuery = useQuery({
    queryKey: ["trace-full", selectedNodeId, depth],
    queryFn: () =>
      apiFetch<ApiWrapper<TraceResponse>>(
        `/api/trace/full/${selectedNodeId}?levels=${depth}`
      ).then((res) => res.data),
    enabled: !!selectedNodeId,
    staleTime: 60_000,
  })

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

  return (
    <div className="relative h-screen h-dvh overflow-hidden bg-card flex flex-col">
      {/* Top toolbar: search + panel toggle */}
      <div className="relative z-10 flex items-center gap-1.5 sm:gap-2 px-2 sm:px-3 py-1.5 sm:py-2 border-b border-border bg-card shrink-0">
        <DAGSearch onSelect={handleSelect} onClear={handleClear} />
        <button
          onClick={() => setPanelOpen((o) => !o)}
          className="h-8 w-8 flex items-center justify-center rounded-lg border border-border hover:bg-muted transition-colors shrink-0"
          title="Settings"
        >
          <Settings className="h-4 w-4" />
        </button>
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
              Downstream impacts (left, red) | Upstream dependencies (right, blue)
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

        {/* Right settings panel */}
        <DagRightPanel
          open={panelOpen}
          onClose={() => setPanelOpen(false)}
          depth={depth}
          onDepthChange={setDepth}
          selectedTopos={selectedTopos}
          onTopoToggle={toggleTopo}
          selectedNodeId={selectedNodeId}
        />
      </div>

      {popupData && <DagDetailPopup data={popupData} onClose={() => setPopupData(null)} />}
    </div>
  )
}

export default function DependencyImpactDAG() {
  return (
    <ReactFlowProvider>
      <DependencyImpactDAGInner />
    </ReactFlowProvider>
  )
}
