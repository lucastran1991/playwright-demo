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
import { Loader2 } from "lucide-react"
import { apiFetch } from "@/lib/api-client"
import { traceToDAGElements, layoutDAG } from "./dag-helpers"
import TracerNode from "./dag-node"
import TracerEdge from "./dag-edge"
import DAGSearch from "./dag-search"
import type { TraceResponse } from "./dag-types"

// Register custom node and edge types
const nodeTypes = { tracerNode: TracerNode }
const edgeTypes = { tracerEdge: TracerEdge }

interface ApiWrapper<T> {
  data: T
}

function DependencyImpactDAGInner() {
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null)
  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])
  const { fitView } = useReactFlow()

  // Fetch dependency trace (upstream + local)
  const depQuery = useQuery({
    queryKey: ["trace-deps", selectedNodeId],
    queryFn: () =>
      apiFetch<ApiWrapper<TraceResponse>>(
        `/api/trace/dependencies/${selectedNodeId}?levels=2&include_local=true`
      ).then((res) => res.data),
    enabled: !!selectedNodeId,
    staleTime: 60_000,
  })

  // Fetch impact trace (downstream + load)
  const impactQuery = useQuery({
    queryKey: ["trace-impact", selectedNodeId],
    queryFn: () =>
      apiFetch<ApiWrapper<TraceResponse>>(
        `/api/trace/impacts/${selectedNodeId}?levels=2`
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
    setNodes(laidNodes)
    setEdges(laidEdges)
    // Center the view after React commits new nodes
    requestAnimationFrame(() => fitView({ padding: 0.25, duration: 300 }))
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
    <div className="relative h-[calc(100vh-8rem)] rounded-lg border border-border overflow-hidden bg-card">
      {/* Search bar overlay */}
      <DAGSearch onSelect={handleSelect} onClear={handleClear} />

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

      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        fitView
        fitViewOptions={{ padding: 0.2 }}
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
