// API response types (match Go backend JSON)
export interface TracedNode {
  id: number
  node_id: string
  name: string
  node_type: string
  level: number
  parent_node_id?: string
}

export interface TraceLevelGroup {
  level: number
  topology: string
  nodes: TracedNode[]
}

export interface TraceLocalGroup {
  topology: string
  nodes: TracedNode[]
}

export interface TraceResponse {
  source: { node_id: string; name: string; node_type: string; topology?: string }
  upstream?: TraceLevelGroup[]
  local?: TraceLocalGroup[]
  downstream?: TraceLevelGroup[]
  load?: TraceLocalGroup[]
}

export interface SearchNode {
  id: number
  node_id: string
  name: string
  node_type: string
}

// Internal ReactFlow node data — must extend Record<string, unknown> for ReactFlow compat
export interface TracerNodeData extends Record<string, unknown> {
  nodeId: string
  name: string
  nodeType: string
  topology: string
  isSource: boolean
  isLocal: boolean
  ring: number // distance from source: 0=source, 1=level1, 2=level2, etc.
  level: number // upstream/downstream level (0 for source/local)
  direction: "upstream" | "downstream" | "local" | "load" | "source"
  onNodeClick?: (data: TracerNodeData) => void
}
