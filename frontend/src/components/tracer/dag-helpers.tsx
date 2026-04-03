import dagre from "@dagrejs/dagre"
import { Zap, Droplets, Building2, Package } from "lucide-react"
import { MarkerType, type Node, type Edge } from "@xyflow/react"
import type { TraceResponse, TracerNodeData } from "./dag-types"

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

// Edge styles: upstream = cyan, downstream = orange, local = gray dashed
const UPSTREAM_STYLE = { stroke: "#06B6D4", strokeWidth: 2.5 }
const DOWNSTREAM_STYLE = { stroke: "#F97316", strokeWidth: 2.5 }
const LOCAL_STYLE = { stroke: "#6B7280", strokeWidth: 1.5, strokeDasharray: "6 3" }

const UPSTREAM_MARKER = { type: MarkerType.ArrowClosed, color: "#06B6D4", width: 16, height: 16 }
const DOWNSTREAM_MARKER = { type: MarkerType.ArrowClosed, color: "#F97316", width: 16, height: 16 }
const LOCAL_MARKER = { type: MarkerType.ArrowClosed, color: "#6B7280", width: 12, height: 12 }

const NODE_WIDTH = 200
const NODE_HEIGHT = 80
const GROUP_PADDING = 20

function makeNode(id: string, name: string, nodeType: string, topology: string, ring: number, isSource: boolean, isLocal: boolean, level?: number, parentId?: string): Node {
  const data: TracerNodeData = { nodeId: id, name, nodeType, topology, isSource, isLocal, ring, level: level ?? 0 }
  const node: Node = { id, type: "tracerNode", position: { x: 0, y: 0 }, data }
  if (parentId) {
    node.parentId = parentId
    node.extent = "parent"
  }
  return node
}

export function traceToDAGElements(
  depResponse: TraceResponse | null,
  impactResponse: TraceResponse | null
): { nodes: Node[]; edges: Edge[] } {
  if (!depResponse && !impactResponse) return { nodes: [], edges: [] }

  const source = depResponse?.source ?? impactResponse?.source
  if (!source) return { nodes: [], edges: [] }

  const nodesMap = new Map<string, Node>()
  const edges: Edge[] = []
  const localNodes: Node[] = []

  // Source node (ring 0)
  const sourceTopo = source.topology || "Electrical System"
  nodesMap.set(source.node_id, makeNode(source.node_id, source.name, source.node_type, sourceTopo, 0, true, false))

  // Upstream deps
  if (depResponse?.upstream) {
    for (const group of depResponse.upstream) {
      for (const n of group.nodes) {
        if (!nodesMap.has(n.node_id)) {
          nodesMap.set(n.node_id, makeNode(n.node_id, n.name, n.node_type, group.topology, group.level, false, false, group.level))
        }
        edges.push({ id: `dep-${n.node_id}-${source.node_id}`, source: n.node_id, target: source.node_id, type: "tracerEdge", style: UPSTREAM_STYLE, animated: true, markerEnd: UPSTREAM_MARKER })
      }
    }
  }

  // Local deps -> will be grouped around source
  if (depResponse?.local) {
    for (const group of depResponse.local) {
      for (const n of group.nodes) {
        if (!nodesMap.has(n.node_id)) {
          const localNode = makeNode(n.node_id, n.name, n.node_type, group.topology, 1, false, true, 0)
          nodesMap.set(n.node_id, localNode)
          localNodes.push(localNode)
        }
        edges.push({ id: `local-${n.node_id}-${source.node_id}`, source: n.node_id, target: source.node_id, type: "tracerEdge", style: LOCAL_STYLE, markerEnd: LOCAL_MARKER })
      }
    }
  }

  // Downstream impacts
  if (impactResponse?.downstream) {
    for (const group of impactResponse.downstream) {
      for (const n of group.nodes) {
        if (!nodesMap.has(n.node_id)) {
          nodesMap.set(n.node_id, makeNode(n.node_id, n.name, n.node_type, group.topology, group.level, false, false, group.level))
        }
        edges.push({ id: `impact-${source.node_id}-${n.node_id}`, source: source.node_id, target: n.node_id, type: "tracerEdge", style: DOWNSTREAM_STYLE, animated: true, markerEnd: DOWNSTREAM_MARKER })
      }
    }
  }

  // Load impacts
  if (impactResponse?.load) {
    for (const group of impactResponse.load) {
      for (const n of group.nodes) {
        if (!nodesMap.has(n.node_id)) {
          nodesMap.set(n.node_id, makeNode(n.node_id, n.name, n.node_type, group.topology, 1, false, true, 0))
        }
        edges.push({ id: `load-${source.node_id}-${n.node_id}`, source: source.node_id, target: n.node_id, type: "tracerEdge", style: LOCAL_STYLE, markerEnd: LOCAL_MARKER })
      }
    }
  }

  // If local deps exist, create a group node containing source + local deps
  const allNodes = Array.from(nodesMap.values())
  if (localNodes.length > 0) {
    const groupId = `group-local-${source.node_id}`
    const cols = Math.min(localNodes.length + 1, 3)
    const rows = Math.ceil((localNodes.length + 1) / cols)
    const groupW = cols * (NODE_WIDTH + GROUP_PADDING) + GROUP_PADDING
    const groupH = rows * (NODE_HEIGHT + GROUP_PADDING) + GROUP_PADDING + 24 // +24 for label

    const groupNode: Node = {
      id: groupId,
      type: "group",
      position: { x: 0, y: 0 },
      data: { label: "Local Dependencies" },
      style: {
        width: groupW,
        height: groupH,
        border: "1.5px dashed #6B7280",
        borderRadius: "12px",
        backgroundColor: "rgba(107,114,128,0.05)",
        padding: "8px",
      },
    }

    // Position children relative to group
    const sourceInGroup = nodesMap.get(source.node_id)!
    sourceInGroup.parentId = groupId
    sourceInGroup.extent = "parent"
    sourceInGroup.position = { x: GROUP_PADDING, y: GROUP_PADDING + 24 }

    localNodes.forEach((ln, i) => {
      ln.parentId = groupId
      ln.extent = "parent"
      const col = (i + 1) % cols
      const row = Math.floor((i + 1) / cols)
      ln.position = { x: GROUP_PADDING + col * (NODE_WIDTH + GROUP_PADDING), y: GROUP_PADDING + 24 + row * (NODE_HEIGHT + GROUP_PADDING) }
    })

    // Group must come first in nodes array
    return { nodes: [groupNode, ...allNodes], edges }
  }

  return { nodes: allNodes, edges }
}

// Dagre LR layout
export function layoutDAG(nodes: Node[], edges: Edge[]): { nodes: Node[]; edges: Edge[] } {
  if (nodes.length === 0) return { nodes, edges }

  // Only layout top-level nodes (no parentId) with Dagre
  const topLevel = nodes.filter((n) => !n.parentId)
  const children = nodes.filter((n) => !!n.parentId)

  if (topLevel.length === 0) return { nodes, edges }

  const g = new dagre.graphlib.Graph()
  g.setDefaultEdgeLabel(() => ({}))
  g.setGraph({ rankdir: "LR", ranksep: 100, nodesep: 40, marginx: 30, marginy: 30 })

  for (const node of topLevel) {
    const w = node.style?.width ? Number(node.style.width) : NODE_WIDTH
    const h = node.style?.height ? Number(node.style.height) : NODE_HEIGHT
    g.setNode(node.id, { width: w, height: h })
  }

  // Only add edges between top-level nodes (or group nodes)
  for (const edge of edges) {
    const srcTop = topLevel.find((n) => n.id === edge.source || children.some((c) => c.id === edge.source && c.parentId === n.id))
    const tgtTop = topLevel.find((n) => n.id === edge.target || children.some((c) => c.id === edge.target && c.parentId === n.id))
    if (srcTop && tgtTop) {
      g.setEdge(srcTop.id, tgtTop.id)
    }
  }

  dagre.layout(g)

  const laid = topLevel.map((node) => {
    const pos = g.node(node.id)
    const w = node.style?.width ? Number(node.style.width) : NODE_WIDTH
    const h = node.style?.height ? Number(node.style.height) : NODE_HEIGHT
    return { ...node, position: { x: pos.x - w / 2, y: pos.y - h / 2 } }
  })

  return { nodes: [...laid, ...children], edges }
}
