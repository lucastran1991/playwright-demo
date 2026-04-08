import dagre from "@dagrejs/dagre"
import { Zap, Droplets, Building2, Package } from "lucide-react"
import { MarkerType, type Node, type Edge } from "@xyflow/react"
import type { TraceResponse, TracerNodeData } from "./dag-types"

// Utilization color: green <60%, yellow 60-80%, red >80%
export function getUtilColor(pct: number): string {
  if (pct > 80) return "#EF4444"
  if (pct >= 60) return "#EAB308"
  return "#22C55E"
}

// Dark-theme optimized topology colors
export const TOPOLOGY_CONFIG: Record<string, { color: string; bg: string; icon: React.ReactNode }> = {
  electrical: { color: "#F97316", bg: "rgba(249,115,22,0.1)", icon: <Zap className="h-3 w-3" /> },
  cooling:    { color: "#06B6D4", bg: "rgba(6,182,212,0.1)",  icon: <Droplets className="h-3 w-3" /> },
  spatial:    { color: "#8B5CF6", bg: "rgba(139,92,246,0.1)", icon: <Building2 className="h-3 w-3" /> },
  whitespace: { color: "#10B981", bg: "rgba(16,185,129,0.1)", icon: <Package className="h-3 w-3" /> },
  default:    { color: "#6B7280", bg: "rgba(107,114,128,0.1)", icon: <Package className="h-3 w-3" /> },
}

export function getTopologyKey(topology: string): string {
  const lower = topology.toLowerCase()
  if (lower.includes("electrical")) return "electrical"
  if (lower.includes("cooling")) return "cooling"
  if (lower.includes("spatial")) return "spatial"
  if (lower.includes("whitespace")) return "whitespace"
  return "default"
}

// Filter trace response to only include groups matching selected topologies.
// Source node is always kept. Groups with non-matching topologies are removed.
export function filterTraceByTopologies(
  response: TraceResponse | null,
  selectedTopos: Set<string>
): TraceResponse | null {
  if (!response || selectedTopos.size === 0) return response

  const matchTopo = (topology: string) => selectedTopos.has(getTopologyKey(topology))

  return {
    source: response.source,
    upstream: response.upstream?.filter((g) => matchTopo(g.topology)),
    local: response.local?.filter((g) => matchTopo(g.topology)),
    downstream: response.downstream?.filter((g) => matchTopo(g.topology)),
    load: response.load?.filter((g) => matchTopo(g.topology)),
    capacity: response.capacity,
  }
}

// Direction colors: distinct from topology colors, used for outer border + edges
export const DIRECTION_COLORS: Record<string, string> = {
  upstream: "#3B82F6",   // blue
  downstream: "#EF4444", // red
  local: "#64748B",      // slate
  load: "#EC4899",       // pink
  source: "#EAB308",     // yellow
}

// Edge styles per direction
const UPSTREAM_STYLE = { stroke: DIRECTION_COLORS.upstream, strokeWidth: 2.5 }
const DOWNSTREAM_STYLE = { stroke: DIRECTION_COLORS.downstream, strokeWidth: 2.5 }
const LOCAL_STYLE = { stroke: DIRECTION_COLORS.local, strokeWidth: 1.5, strokeDasharray: "6 3" }
const LOAD_STYLE = { stroke: DIRECTION_COLORS.load, strokeWidth: 1.5, strokeDasharray: "4 4" }

const UPSTREAM_MARKER = { type: MarkerType.ArrowClosed, color: DIRECTION_COLORS.upstream, width: 16, height: 16 }
const DOWNSTREAM_MARKER = { type: MarkerType.ArrowClosed, color: DIRECTION_COLORS.downstream, width: 16, height: 16 }
const LOCAL_MARKER = { type: MarkerType.ArrowClosed, color: DIRECTION_COLORS.local, width: 12, height: 12 }
const LOAD_MARKER = { type: MarkerType.ArrowClosed, color: DIRECTION_COLORS.load, width: 12, height: 12 }

const NODE_WIDTH = 180
const NODE_HEIGHT = 72

function makeNode(id: string, name: string, nodeType: string, topology: string, ring: number, isSource: boolean, isLocal: boolean, level?: number, parentId?: string, direction: "upstream" | "downstream" | "local" | "load" | "source" = "upstream"): Node {
  const data: TracerNodeData = { nodeId: id, name, nodeType, topology, isSource, isLocal, ring, level: level ?? 0, direction }
  const node: Node = { id, type: "tracerNode", position: { x: 0, y: 0 }, data }
  if (parentId) {
    node.parentId = parentId
    node.extent = "parent"
  }
  return node
}

export function traceToDAGElements(
  response: TraceResponse | null
): { nodes: Node[]; edges: Edge[] } {
  if (!response) return { nodes: [], edges: [] }

  const source = response.source
  if (!source) return { nodes: [], edges: [] }

  const nodesMap = new Map<string, Node>()
  const edges: Edge[] = []
  const localNodes: Node[] = []

  // Source node (ring 0)
  const sourceTopo = source.topology || "Electrical System"
  nodesMap.set(source.node_id, makeNode(source.node_id, source.name, source.node_type, sourceTopo, 0, true, false, 0, undefined, "source"))

  // Upstream deps -- chain edges using parent_node_id (L2→L1→source)
  if (response.upstream) {
    const sortedUpstream = [...response.upstream].sort((a, b) => a.level - b.level)
    for (const group of sortedUpstream) {
      for (const n of group.nodes) {
        if (!nodesMap.has(n.node_id)) {
          nodesMap.set(n.node_id, makeNode(n.node_id, n.name, n.node_type, group.topology, group.level, false, false, group.level, undefined, "upstream"))
        }
        const rawTarget = n.parent_node_id ?? source.node_id
        const target = nodesMap.has(rawTarget) || rawTarget === source.node_id ? rawTarget : source.node_id
        const edgeId = `dep-${n.node_id}-${target}`
        if (!edges.some((e) => e.id === edgeId)) {
          edges.push({ id: edgeId, source: n.node_id, target, type: "tracerEdge", style: UPSTREAM_STYLE, animated: true, markerEnd: UPSTREAM_MARKER })
        }
      }
    }
  }

  // Local deps -> grouped around source
  if (response.local) {
    for (const group of response.local) {
      for (const n of group.nodes) {
        if (!nodesMap.has(n.node_id)) {
          const localNode = makeNode(n.node_id, n.name, n.node_type, group.topology, 1, false, true, 0, undefined, "local")
          nodesMap.set(n.node_id, localNode)
          localNodes.push(localNode)
        }
        const edgeId = `local-${source.node_id}-${n.node_id}`
        if (!edges.some((e) => e.id === edgeId)) {
          edges.push({ id: edgeId, source: source.node_id, target: n.node_id, type: "tracerEdge", style: LOCAL_STYLE, markerEnd: LOCAL_MARKER, sourceHandle: "source-bottom", targetHandle: "local-target" })
        }
      }
    }
  }

  // Downstream impacts -- chain edges using parent_node_id (source→L1→L2)
  if (response.downstream) {
    const sortedDownstream = [...response.downstream].sort((a, b) => a.level - b.level)
    for (const group of sortedDownstream) {
      for (const n of group.nodes) {
        if (!nodesMap.has(n.node_id)) {
          nodesMap.set(n.node_id, makeNode(n.node_id, n.name, n.node_type, group.topology, group.level, false, false, group.level, undefined, "downstream"))
        }
        const rawFrom = n.parent_node_id ?? source.node_id
        const from = nodesMap.has(rawFrom) || rawFrom === source.node_id ? rawFrom : source.node_id
        const edgeId = `impact-${from}-${n.node_id}`
        if (!edges.some((e) => e.id === edgeId)) {
          edges.push({ id: edgeId, source: from, target: n.node_id, type: "tracerEdge", style: DOWNSTREAM_STYLE, animated: true, markerEnd: DOWNSTREAM_MARKER })
        }
      }
    }
  }

  // Load impacts — distinct purple dashed edges, skip nodes already in downstream
  if (response.load) {
    for (const group of response.load) {
      for (const n of group.nodes) {
        if (!nodesMap.has(n.node_id)) {
          nodesMap.set(n.node_id, makeNode(n.node_id, n.name, n.node_type, group.topology, 1, false, false, 0, undefined, "load"))
        }
        const edgeId = `load-${source.node_id}-${n.node_id}`
        if (!edges.some((e) => e.id === edgeId)) {
          edges.push({ id: edgeId, source: source.node_id, target: n.node_id, type: "tracerEdge", style: LOAD_STYLE, markerEnd: LOAD_MARKER })
        }
      }
    }
  }

  // Enrich nodes with capacity data from response
  const capacityMap = response.capacity ?? {}
  for (const [nodeId, node] of nodesMap) {
    if (capacityMap[nodeId]) {
      ;(node.data as TracerNodeData).capacity = capacityMap[nodeId]
    }
  }

  // No group container -- local nodes positioned after Dagre layout in layoutDAG
  const allNodes = Array.from(nodesMap.values())
  return { nodes: allNodes.map((n) => ({ ...n, zIndex: 10 })), edges }
}

// Dagre RL layout: downstream LEFT, upstream RIGHT, local nodes BELOW source
export function layoutDAG(nodes: Node[], edges: Edge[]): { nodes: Node[]; edges: Edge[] } {
  if (nodes.length === 0) return { nodes, edges }

  // Separate local nodes from Dagre layout -- they get positioned manually below source
  const dagreNodes = nodes.filter((n) => (n.data as TracerNodeData).direction !== "local")
  const localNodes = nodes.filter((n) => (n.data as TracerNodeData).direction === "local")

  if (dagreNodes.length === 0) return { nodes, edges }

  const g = new dagre.graphlib.Graph()
  g.setDefaultEdgeLabel(() => ({}))
  g.setGraph({ rankdir: "RL", ranksep: 120, nodesep: 50, marginx: 30, marginy: 30 })

  for (const node of dagreNodes) {
    g.setNode(node.id, { width: NODE_WIDTH, height: NODE_HEIGHT })
  }

  // Only add edges between dagre nodes (skip local edges)
  const dagreIds = new Set(dagreNodes.map((n) => n.id))
  for (const edge of edges) {
    if (dagreIds.has(edge.source) && dagreIds.has(edge.target)) {
      g.setEdge(edge.source, edge.target)
    }
  }

  dagre.layout(g)

  const laid = dagreNodes.map((node) => {
    const pos = g.node(node.id)
    return { ...node, position: { x: pos.x - NODE_WIDTH / 2, y: pos.y - NODE_HEIGHT / 2 } }
  })

  // Position local nodes in a row below the source node
  if (localNodes.length > 0) {
    const sourceNode = laid.find((n) => (n.data as TracerNodeData).isSource)
    const sourceX = sourceNode?.position.x ?? 0
    const sourceY = sourceNode?.position.y ?? 0
    const LOCAL_GAP = 24
    const LOCAL_Y_OFFSET = 140 // vertical gap below source
    const totalWidth = localNodes.length * NODE_WIDTH + (localNodes.length - 1) * LOCAL_GAP
    const startX = sourceX + NODE_WIDTH / 2 - totalWidth / 2 // center below source

    const positionedLocals = localNodes.map((ln, i) => ({
      ...ln,
      position: { x: startX + i * (NODE_WIDTH + LOCAL_GAP), y: sourceY + LOCAL_Y_OFFSET },
    }))
    return { nodes: [...laid, ...positionedLocals], edges }
  }

  return { nodes: [...laid, ...localNodes], edges }
}
