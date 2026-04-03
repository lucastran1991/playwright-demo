# Phase 3: Tracer Repository (Recursive CTE Queries)

## Context Links
- Brainstorm CTE example: `plans/reports/brainstorm-260404-0025-capacity-dependency-impact-model.md` (line 152-171)
- Existing repo pattern: `backend/internal/repository/blueprint_repository.go`
- Edge model: `backend/internal/model/blueprint_edge.go` (from_node_id -> to_node_id = parent -> child)

## Overview
- **Priority**: P1
- **Status**: completed
- **Description**: Raw SQL recursive CTE queries for upstream/downstream traversal against blueprint_edges

## Key Insights
- Edge direction: `from_node_id` = parent, `to_node_id` = child
- **Upstream walk**: follow edges where `to_node_id` = current, get `from_node_id` (go to parent)
- **Downstream walk**: follow edges where `from_node_id` = current, get `to_node_id` (go to children)
- Must scope walk to specific blueprint_type (electrical vs cooling)
- Level = hop distance from source node
- Need to map topology names from rules (e.g. "Electrical System") to blueprint_type slugs (e.g. "electrical-system")

## Requirements

### Functional
- `FindUpstreamNodes(sourceDBID uint, blueprintTypeSlug string, maxLevel int) -> []TracedNode`
- `FindDownstreamNodes(sourceDBID uint, blueprintTypeSlug string, maxLevel int) -> []TracedNode`
- `FindLocalNodes(sourceDBID uint, blueprintTypeSlug string) -> []TracedNode` -- direct edge neighbors (both parent and child)
- `FindNodeByNodeID(nodeID string) -> BlueprintNode` -- lookup source node
- `ListCapacityNodeTypes() -> []CapacityNodeType`
- `GetDependencyRules(nodeType string) -> []DependencyRule`
- `GetImpactRules(nodeType string) -> []ImpactRule`

### Non-functional
- Use raw SQL with `db.Raw()` for CTEs (GORM doesn't support recursive CTE natively)
- Return flat result sets -- service layer handles grouping
- Include cycle protection in CTE (track visited IDs)

## Architecture

### TracedNode result struct
```go
type TracedNode struct {
    ID       uint   `json:"id"`
    NodeID   string `json:"node_id"`
    Name     string `json:"name"`
    NodeType string `json:"node_type"`
    Level    int    `json:"level"`
}
```

### Upstream CTE (parent walk)
```sql
WITH RECURSIVE upstream AS (
    -- Base: direct parents of source node in this topology
    SELECT bn.id, bn.node_id, bn.name, bn.node_type, 1 as level
    FROM blueprint_edges be
    JOIN blueprint_nodes bn ON bn.id = be.from_node_id
    JOIN blueprint_types bt ON bt.id = be.blueprint_type_id
    WHERE be.to_node_id = @sourceID AND bt.slug = @typeSlug

    UNION ALL

    -- Recurse: parents of parents
    SELECT bn.id, bn.node_id, bn.name, bn.node_type, u.level + 1
    FROM upstream u
    JOIN blueprint_edges be ON be.to_node_id = u.id
    JOIN blueprint_nodes bn ON bn.id = be.from_node_id
    JOIN blueprint_types bt ON bt.id = be.blueprint_type_id
    WHERE bt.slug = @typeSlug AND u.level < @maxLevel
)
SELECT DISTINCT id, node_id, name, node_type, level FROM upstream
ORDER BY level, node_type, node_id
```

### Downstream CTE (child walk)
```sql
WITH RECURSIVE downstream AS (
    -- Base: direct children of source node in this topology
    SELECT bn.id, bn.node_id, bn.name, bn.node_type, 1 as level
    FROM blueprint_edges be
    JOIN blueprint_nodes bn ON bn.id = be.to_node_id
    JOIN blueprint_types bt ON bt.id = be.blueprint_type_id
    WHERE be.from_node_id = @sourceID AND bt.slug = @typeSlug

    UNION ALL

    -- Recurse: children of children
    SELECT bn.id, bn.node_id, bn.name, bn.node_type, d.level + 1
    FROM downstream d
    JOIN blueprint_edges be ON be.from_node_id = d.id
    JOIN blueprint_nodes bn ON bn.id = be.to_node_id
    JOIN blueprint_types bt ON bt.id = be.blueprint_type_id
    WHERE bt.slug = @typeSlug AND d.level < @maxLevel
)
SELECT DISTINCT id, node_id, name, node_type, level FROM downstream
ORDER BY level, node_type, node_id
```

### Local query (direct neighbors)
No CTE needed -- simple join:
```sql
SELECT bn.id, bn.node_id, bn.name, bn.node_type, 0 as level
FROM blueprint_edges be
JOIN blueprint_nodes bn ON (bn.id = be.from_node_id OR bn.id = be.to_node_id)
JOIN blueprint_types bt ON bt.id = be.blueprint_type_id
WHERE bt.slug = @typeSlug
  AND (be.from_node_id = @sourceID OR be.to_node_id = @sourceID)
  AND bn.id != @sourceID
```

## Related Code Files

### Files to Create
- `backend/internal/repository/tracer_repository.go`

### Files to Read (reference)
- `backend/internal/repository/blueprint_repository.go` -- existing raw SQL pattern (GetTree)

## Implementation Steps

### 1. Create `backend/internal/repository/tracer_repository.go`

```go
type TracerRepository struct {
    db *gorm.DB
}

func NewTracerRepository(db *gorm.DB) *TracerRepository
```

Methods:
1. `FindUpstreamNodes(sourceDBID uint, typeSlug string, maxLevel int) ([]TracedNode, error)` -- upstream CTE
2. `FindDownstreamNodes(sourceDBID uint, typeSlug string, maxLevel int) ([]TracedNode, error)` -- downstream CTE
3. `FindLocalNodes(sourceDBID uint, typeSlug string) ([]TracedNode, error)` -- direct neighbors
4. `GetDependencyRules(nodeType string) ([]model.DependencyRule, error)` -- simple WHERE query
5. `GetImpactRules(nodeType string) ([]model.ImpactRule, error)` -- simple WHERE query
6. `ListCapacityNodeTypes() ([]model.CapacityNodeType, error)` -- list all

### 2. Topology name to slug mapping
Rules CSV uses "Electrical System", blueprint_types uses slug "electrical-system".
Use `FolderToSlug()` from blueprint_csv_parser? No -- topology names don't have "_Blueprint" suffix.
Simpler: lowercase + replace spaces with hyphens. Add helper:
```go
func topologyToSlug(topology string) string {
    return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(topology), " ", "-"))
}
```
But this mapping happens in service layer, not repository. Repository takes slug directly.

## Todo List
- [x] Create tracer_repository.go with TracedNode struct
- [x] Implement FindUpstreamNodes with recursive CTE
- [x] Implement FindDownstreamNodes with recursive CTE
- [x] Implement FindLocalNodes
- [x] Implement rule lookup methods
- [x] Verify `go build` compiles

## Success Criteria
- Upstream CTE returns correct parent chain for a known node
- Downstream CTE returns correct child chain
- Level values match hop distance
- No infinite loops (UNION ALL + level cap prevents cycles)
- File stays under 200 lines

## Risk Assessment
- **Medium**: CTE correctness depends on edge direction convention. Verified: from=parent, to=child.
- **Low**: UNION (not UNION ALL) in CTE deduplicates, but level cap also prevents infinite loops. Using UNION ALL + level cap is sufficient.
- **Medium**: topology-to-slug mapping must be consistent. Service layer handles this.

## Next Steps
- Phase 4 uses these queries to build the full trace response
