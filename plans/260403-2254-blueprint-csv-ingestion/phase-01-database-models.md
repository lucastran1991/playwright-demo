# Phase 1: Database Models

## Context Links
- [Plan Overview](plan.md)
- [Brainstorm Report](../reports/brainstorm-260403-2249-blueprint-ingestion-model.md)
- Existing pattern: `backend/internal/model/user.go`

## Overview
- **Priority**: P1 (blocking all other phases)
- **Status**: completed
- **Description**: Create 4 GORM model structs following existing conventions (json tags, gorm tags, time.Time fields)

## Key Insights
- Node IDs are strings like "RACK-R1-Z1-R1-01", not integers
- Same node appears in multiple domains -- unified table with membership join
- Node Role is always empty but kept for future use
- Unique constraints needed: `(blueprint_type_id, blueprint_node_id)` on memberships, `(blueprint_type_id, from_node_id, to_node_id)` on edges

## Requirements

### Functional
- 4 model files: BlueprintType, BlueprintNode, BlueprintNodeMembership, BlueprintEdge
- All models include ID, CreatedAt, UpdatedAt
- Foreign key relationships with GORM tags
- JSON tags on all exported fields

### Non-functional
- Each file under 200 lines
- Follow snake_case file naming
- Match existing model conventions (see `user.go`)

## Architecture
```
blueprint_types 1──N blueprint_node_memberships N──1 blueprint_nodes
blueprint_types 1──N blueprint_edges
blueprint_nodes 1──N blueprint_edges (from_node_id)
blueprint_nodes 1──N blueprint_edges (to_node_id)
```

## Related Code Files

### Files to Create
- `backend/internal/model/blueprint_type.go`
- `backend/internal/model/blueprint_node.go`
- `backend/internal/model/blueprint_node_membership.go`
- `backend/internal/model/blueprint_edge.go`

### Files to Reference
- `backend/internal/model/user.go` -- conventions

## Implementation Steps

1. Create `blueprint_type.go` with BlueprintType struct
   - Fields: ID, Name (uniqueIndex), Slug (uniqueIndex), FolderName, CreatedAt, UpdatedAt
   - Slug derived from folder name (lowercase, hyphens)

2. Create `blueprint_node.go` with BlueprintNode struct
   - Fields: ID, NodeID (uniqueIndex), Name, NodeType (indexed), NodeRole, CreatedAt, UpdatedAt
   - NodeID is the string identifier from CSV

3. Create `blueprint_node_membership.go` with BlueprintNodeMembership struct
   - Fields: ID, BlueprintTypeID (FK), BlueprintNodeID (FK), OrgPath, CreatedAt, UpdatedAt
   - Composite unique index on (BlueprintTypeID, BlueprintNodeID)
   - GORM relationship tags for BlueprintType and BlueprintNode

4. Create `blueprint_edge.go` with BlueprintEdge struct
   - Fields: ID, BlueprintTypeID (FK), FromNodeID (FK), ToNodeID (FK), CreatedAt, UpdatedAt
   - Composite unique index on (BlueprintTypeID, FromNodeID, ToNodeID)
   - GORM relationship tags for all FKs

## Todo List
- [x] Create blueprint_type.go
- [x] Create blueprint_node.go
- [x] Create blueprint_node_membership.go
- [x] Create blueprint_edge.go
- [x] Verify all models compile: `go build ./...`

## Success Criteria
- All 4 files compile without errors
- GORM tags match the brainstorm ER diagram
- Unique constraints defined for upsert support
- JSON tags present on all fields

## Risk Assessment
- **Low**: Node Type conflict across domains -- using last-write-wins on upsert (acceptable for now)
- **Mitigation**: Log warning during ingestion if node_type differs

## Security Considerations
- No user input involved in model definitions
- Foreign keys enforce referential integrity
