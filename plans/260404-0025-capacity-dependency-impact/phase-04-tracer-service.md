# Phase 4: Dependency Tracer Service

## Context Links
- Brainstorm resolution algorithm: `plans/reports/brainstorm-260404-0025-capacity-dependency-impact-model.md` (line 92-110)
- Repository methods: `plans/260404-0025-capacity-dependency-impact/phase-03-tracer-repository.md`

## Overview
- **Priority**: P1
- **Status**: completed
- **Description**: Orchestrates type-level rules + topology queries to resolve actual dependency/impact instances for a given node

## Key Insights
- Service is the brain: reads rules, determines which topologies to query, filters results by allowed node types
- Repository returns ALL nodes at each level; service filters to only those matching rule's target type
- Topology name -> slug conversion happens here
- Multiple rules can apply to same source type (e.g. Rack has ~15 dependency rules across 2 topologies)

## Requirements

### Functional
- `TraceDependencies(nodeID string, maxLevels int, includeLocal bool) -> TraceResponse`
- `TraceImpacts(nodeID string, maxLevels int, loadScope string) -> TraceResponse`
- Group results by topology and level
- Filter upstream/downstream results to only node types matching rules
- Handle "Local" dependencies separately from "Upstream"
- Return source node info in response

### Non-functional
- Single struct, inject TracerRepository
- Under 200 lines

## Architecture

### Response structs
```go
type TraceResponse struct {
    Source    SourceNode         `json:"source"`
    Upstream []TraceLevelGroup  `json:"upstream,omitempty"`
    Local    []TraceLocalGroup  `json:"local,omitempty"`
    Downstream []TraceLevelGroup `json:"downstream,omitempty"`
    Load     []TraceLocalGroup  `json:"load,omitempty"`
}

type SourceNode struct {
    NodeID   string `json:"node_id"`
    Name     string `json:"name"`
    NodeType string `json:"node_type"`
}

type TraceLevelGroup struct {
    Level    int          `json:"level"`
    Topology string       `json:"topology"`
    Nodes    []TracedNode `json:"nodes"`
}

type TraceLocalGroup struct {
    Topology string       `json:"topology"`
    Nodes    []TracedNode `json:"nodes"`
}
```

### TraceDependencies algorithm
```
1. Look up source node by nodeID -> get DB ID + node_type
2. Get dependency rules WHERE node_type = source's type
3. Separate rules into upstream vs local
4. For upstream rules:
   a. Group by topology (e.g. "Electrical System", "Cooling System")
   b. For each topology:
      - Convert topology to slug (e.g. "electrical-system")
      - Call repo.FindUpstreamNodes(sourceDBID, slug, maxLevels)
      - Filter results: keep only nodes whose node_type appears in this topology's rules
      - Group by level
5. For local rules (if includeLocal):
   a. Group by topology
   b. For each topology:
      - Call repo.FindLocalNodes(sourceDBID, slug)
      - Filter to matching node types
6. Build and return TraceResponse
```

### TraceImpacts algorithm
```
1. Look up source node
2. Get impact rules WHERE node_type = source's type
3. Separate into downstream vs load
4. For downstream rules:
   a. Group by topology
   b. For each topology:
      - Call repo.FindDownstreamNodes(sourceDBID, slug, maxLevels)
      - Filter to matching impact node types
      - Group by level
5. For load rules:
   a. Similar to local -- find connected load nodes
   b. loadScope param determines how far to look (future: may need special handling)
6. Build and return TraceResponse
```

### Topology to slug helper
```go
func topologyToSlug(topology string) string {
    s := strings.ToLower(strings.TrimSpace(topology))
    s = strings.ReplaceAll(s, " ", "-")
    return s
}
```

## Related Code Files

### Files to Create
- `backend/internal/service/dependency_tracer.go`

### Files to Read (reference)
- `backend/internal/repository/tracer_repository.go` (Phase 3)

## Implementation Steps

### 1. Define response structs at top of file

### 2. Create DependencyTracer struct
```go
type DependencyTracer struct {
    repo *repository.TracerRepository
}

func NewDependencyTracer(repo *repository.TracerRepository) *DependencyTracer
```

### 3. Implement TraceDependencies
- Lookup source node via repo (reuse BlueprintRepository.FindNodeByNodeID or add to TracerRepository)
- Query dependency rules
- Group rules by topological_relationship (Upstream vs Local) then by topology
- For each topology group: run CTE query, filter, group by level
- Assemble response

### 4. Implement TraceImpacts
- Same pattern but with impact rules + downstream CTE
- Load impacts: for now treat same as local (find direct neighbors of matching type)

### 5. Add topologyToSlug helper

## Todo List
- [x] Define response structs
- [x] Implement DependencyTracer struct + constructor
- [x] Implement TraceDependencies
- [x] Implement TraceImpacts
- [x] Add topologyToSlug helper
- [x] Verify `go build` compiles

## Success Criteria
- TraceDependencies returns grouped upstream results filtered to correct node types
- Level numbers match the UpstreamLevel from rules CSV
- Local dependencies included when flag is true
- TraceImpacts returns downstream + load results correctly
- File under 200 lines

## Risk Assessment
- **Medium**: filtering logic correctness -- must match CTE level with rule's UpstreamLevel. The CTE returns ALL upstream nodes at each hop level; rules specify which node types to expect at which level. If a node type appears at a different level than the rule says, the filtering decision matters. Current approach: use CTE level (actual hop distance), not rule level. Rule level is informational.
- **Medium**: "Local" resolution strategy still ambiguous. Using direct edge neighbors as pragmatic first approach.
- **Low**: load scope parameter -- for MVP, ignore scope and return all load-type impacts. Can refine later (YAGNI).

## Next Steps
- Phase 5 exposes this service via HTTP handlers
