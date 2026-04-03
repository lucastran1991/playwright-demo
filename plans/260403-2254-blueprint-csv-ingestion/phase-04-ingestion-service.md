# Phase 4: Ingestion Service

## Context Links
- [Plan Overview](plan.md)
- [Phase 1: Models](phase-01-database-models.md)
- [Phase 3: CSV Parser](phase-03-csv-parser-service.md)

## Overview
- **Priority**: P1
- **Status**: completed
- **Description**: Orchestrates CSV parsing + database upsert. Wraps each domain in a transaction. Returns ingestion summary.

## Key Insights
- Nodes must be upserted before edges (edges reference node PKs)
- Same node_id may appear in multiple domains -- upsert ensures single row
- Memberships track which domain a node belongs to + domain-specific org_path
- Edges reference internal DB IDs (uint), not CSV string IDs -- need lookup after node upsert
- Data volume is small -- no need for batch inserts or async processing

## Requirements

### Functional
- Upsert BlueprintType by folder name (derive slug from folder name)
- Upsert BlueprintNode by node_id (ON CONFLICT DO UPDATE name, node_type)
- Upsert BlueprintNodeMembership by (type_id, node_id) -- update org_path
- Upsert BlueprintEdge by (type_id, from_node_id, to_node_id) -- DO NOTHING on conflict
- Return summary: domains processed, nodes upserted, edges upserted, errors
- Wrap each domain in a DB transaction

### Non-functional
- File under 200 lines
- Log warnings for node_type conflicts across domains
- Continue processing other domains if one fails

## Architecture

```
blueprint_ingestion_service.go
  └── IngestAll(basePath) -> IngestionSummary
        ├── for each domain folder:
        │   ├── upsertBlueprintType(tx, folderName)
        │   ├── parseAndUpsertNodes(tx, typeID, nodesPath)
        │   │   ├── upsert BlueprintNode
        │   │   └── upsert BlueprintNodeMembership
        │   └── parseAndUpsertEdges(tx, typeID, edgesPath)
        │       ├── resolve node string IDs → DB uint IDs
        │       └── upsert BlueprintEdge
        └── aggregate results

IngestionSummary {
    DomainsProcessed int
    NodesUpserted    int
    EdgesUpserted    int
    Errors           []string
}
```

## Related Code Files

### Files to Create
- `backend/internal/service/blueprint_ingestion_service.go`
- `backend/internal/repository/blueprint_repository.go`

### Files to Reference
- Phase 1 model files
- `backend/internal/service/blueprint_csv_parser.go` (Phase 3)
- `backend/internal/repository/user_repository.go` -- repo conventions

## Implementation Steps

1. Create `blueprint_repository.go` with methods:
   ```go
   type BlueprintRepository struct { db *gorm.DB }
   
   func (r *BlueprintRepository) UpsertType(tx *gorm.DB, bt *model.BlueprintType) error
   func (r *BlueprintRepository) UpsertNode(tx *gorm.DB, node *model.BlueprintNode) error
   func (r *BlueprintRepository) UpsertMembership(tx *gorm.DB, m *model.BlueprintNodeMembership) error
   func (r *BlueprintRepository) UpsertEdge(tx *gorm.DB, e *model.BlueprintEdge) error
   func (r *BlueprintRepository) FindNodeByNodeID(tx *gorm.DB, nodeID string) (*model.BlueprintNode, error)
   func (r *BlueprintRepository) GetTree(typeSlug string) ([]TreeNode, error)  // recursive CTE
   ```
   - Upsert methods use `gorm.DB.Clauses(clause.OnConflict{...})`
   - FindNodeByNodeID used during edge processing to resolve string→uint

2. Create `blueprint_ingestion_service.go`:
   ```go
   type BlueprintIngestionService struct {
       repo   *BlueprintRepository
       parser *BlueprintCSVParser  // or just use package-level funcs
   }
   ```

3. Implement `IngestAll(basePath string) (*IngestionSummary, error)`:
   - Call `DiscoverDomains(basePath)`
   - For each domain:
     - Begin transaction
     - Derive slug from folder name: lowercase, replace spaces/underscores with hyphens, strip "_Blueprint" suffix
     - Upsert BlueprintType
     - Parse Nodes.csv, loop: upsert each node + membership
     - Build in-memory map of `nodeID string → uint` for edge resolution
     - Parse Edges.csv, loop: resolve from/to IDs, upsert edge
     - Commit transaction (rollback on error)
   - Return summary

4. Slug derivation helper:
   ```go
   func folderToSlug(folder string) string {
       // "Cooling system_Blueprint" → "cooling-system"
       // Remove "_Blueprint" or " Blueprint" suffix, lowercase, replace spaces/underscores with hyphens
   }
   ```

5. Slug-to-name helper:
   ```go
   func folderToName(folder string) string {
       // "Cooling system_Blueprint" → "Cooling System"
   }
   ```

## Todo List
- [x] Create blueprint_repository.go with upsert methods
- [x] Create blueprint_ingestion_service.go
- [x] Implement slug/name derivation from folder name
- [x] Implement IngestAll with per-domain transactions
- [x] Implement node ID resolution map for edges
- [x] Add GetTree recursive CTE query in repository
- [x] Verify compilation

## Success Criteria
- All 6 domains processed in single IngestAll call
- Nodes upserted (re-run produces no duplicates)
- Memberships correctly link nodes to domains
- Edges correctly reference internal DB node IDs
- Transaction rollback on failure for individual domain
- Summary reports accurate counts

## Risk Assessment
- **Medium**: Node ID not found during edge resolution (edge references node not in Nodes.csv)
  - **Mitigation**: Log warning, skip that edge, include in error summary
- **Low**: Transaction deadlocks -- data volume too small to matter
- **Low**: Slug collision from different folder names -- unlikely given known data

## Security Considerations
- File paths constructed internally, not from user input
- DB operations use parameterized queries (GORM default)
