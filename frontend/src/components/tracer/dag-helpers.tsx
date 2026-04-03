import dagre from "@dagrejs/dagre"
import { Zap, Droplets, Building2, Package } from "lucide-react"
import type { Node, Edge } from "@xyflow/react"
import type { TraceResponse, TracerNodeData } from "./dag-types"

// Topology display config
export const TOPOLOGY_CONFIG: Record<string, { color: string; icon: React.ReactNode }> = {
  electrical: { color: "#F59E0B", icon: <Zap className="h-3 w-3" /> },
  cooling: { color: "#3B82F6", icon: <Droplets className="h-3 w-3" /> },
  spatial: { color: "#8B5CF6", icon: <Building2 className="h-3 w-3" /> },
  default: { color: "#6B7280", icon: <Package className="h-3 w-3" /> },
}

// Normalize topology string to config key
export function getTopologyKey(topology: string): string {
  const lower = topology.toLowerCase()
  if (lower.includes("electrical")) return "electrical"
  if (lower.includes("cooling")) return "cooling"
  if (lower.includes("spatial")) return "spatial"
  return "default"
}

// Edge style constants
const DEP_STYLE = { stroke: "#3B82F6", strokeWidth: 2 }
const IMPACT_STYLE = { stroke: "#EF4444", strokeWidth: 2 }
const LOCAL_STYLE = { stroke: "#6B7280", strokeWidth: 1.5, strokeDasharray: "6 3" }

const NODE_WIDTH = 200
const NODE_HEIGHT = 80

// Helper to create a node entry
function makeNode(nodeId: string, name: string, nodeType: string, topology: string, ring: number, isSource: boolean, isLocal: boolean): Node {
  const data: TracerNodeData = { nodeId, name, nodeType, topology, isSource, isLocal, ring }
  return { id: nodeId, type: "tracerNode", position: { x: 0, y: 0 }, data }
}

// Merge dep and impact TraceResponse into ReactFlow nodes + edges
export function traceToDAGElements(
  depResponse: TraceResponse | null,
  impactResponse: TraceResponse | null
): { nodes: Node[]; edges: Edge[] } {
  if (!depResponse && !impactResponse) return { nodes: [], edges: [] }

  const source = depResponse?.source ?? impactResponse?.source
  if (!source) return { nodes: [], edges: [] }

  const nodesMap = new Map<string, Node>()
  const edges: Edge[] = []

  // Source node at center (ring 0)
  nodesMap.set(source.node_id, makeNode(source.node_id, source.name, source.node_type, "source", 0, true, false))

  // Upstream dependencies (ring = level)
  if (depResponse?.upstream) {
    for (const group of depResponse.upstream) {
      for (const n of group.nodes) {
        if (!nodesMap.has(n.node_id)) {
          nodesMap.set(n.node_id, makeNode(n.node_id, n.name, n.node_type, group.topology, group.level, false, false))
        }
        edges.push({ id: `dep-${n.node_id}-${source.node_id}`, source: n.node_id, target: source.node_id, type: "tracerEdge", style: DEP_STYLE, data: { label: group.topology } })
      }
    }
  }

  // Local dependencies (ring 1, local flag)
  if (depResponse?.local) {
    for (const group of depResponse.local) {
      for (const n of group.nodes) {
        if (!nodesMap.has(n.node_id)) {
          nodesMap.set(n.node_id, makeNode(n.node_id, n.name, n.node_type, group.topology, 1, false, true))
        }
        edges.push({ id: `local-${n.node_id}-${source.node_id}`, source: n.node_id, target: source.node_id, type: "tracerEdge", style: LOCAL_STYLE, data: { label: group.topology } })
      }
    }
  }

  // Downstream impacts (ring = level)
  if (impactResponse?.downstream) {
    for (const group of impactResponse.downstream) {
      for (const n of group.nodes) {
        if (!nodesMap.has(n.node_id)) {
          nodesMap.set(n.node_id, makeNode(n.node_id, n.name, n.node_type, group.topology, group.level, false, false))
        }
        edges.push({ id: `impact-${source.node_id}-${n.node_id}`, source: source.node_id, target: n.node_id, type: "tracerEdge", style: IMPACT_STYLE, data: { label: group.topology } })
      }
    }
  }

  // Load impacts (ring 1, local flag)
  if (impactResponse?.load) {
    for (const group of impactResponse.load) {
      for (const n of group.nodes) {
        if (!nodesMap.has(n.node_id)) {
          nodesMap.set(n.node_id, makeNode(n.node_id, n.name, n.node_type, group.topology, 1, false, true))
        }
        edges.push({ id: `load-${source.node_id}-${n.node_id}`, source: source.node_id, target: n.node_id, type: "tracerEdge", style: LOCAL_STYLE, data: { label: group.topology } })
      }
    }
  }

  return { nodes: Array.from(nodesMap.values()), edges }
}

// Dagre horizontal (LR) layout
export function layoutDAG(nodes: Node[], edges: Edge[]): { nodes: Node[]; edges: Edge[] } {
  if (nodes.length === 0) return { nodes, edges }

  const g = new dagre.graphlib.Graph()
  g.setDefaultEdgeLabel(() => ({}))
  g.setGraph({ rankdir: "LR", ranksep: 100, nodesep: 40, marginx: 30, marginy: 30 })

  for (const node of nodes) {
    g.setNode(node.id, { width: NODE_WIDTH, height: NODE_HEIGHT })
  }
  for (const edge of edges) {
    g.setEdge(edge.source, edge.target)
  }

  dagre.layout(g)

  const laid = nodes.map((node) => {
    const pos = g.node(node.id)
    return { ...node, position: { x: pos.x - NODE_WIDTH / 2, y: pos.y - NODE_HEIGHT / 2 } }
  })

  return { nodes: laid, edges }
}
