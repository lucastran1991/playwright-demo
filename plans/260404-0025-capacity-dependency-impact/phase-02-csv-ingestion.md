# Phase 2: Model CSV Parser + Ingestion Service

## Context Links
- CSV parser pattern: `backend/internal/service/blueprint_csv_parser.go`
- Ingestion pattern: `backend/internal/service/blueprint_ingestion_service.go`
- Input CSVs: `blueprint/Capacity Nodes.csv`, `blueprint/Dependencies.csv`, `blueprint/Impacts.csv`

## Overview
- **Priority**: P1
- **Status**: completed
- **Description**: Parse 3 model CSVs and upsert into database. Reuse `readCSV()` from existing parser.

## Key Insights
- CSVs are flat, no domain-folder discovery needed -- 3 known files at fixed paths
- `readCSV()` in `blueprint_csv_parser.go` is unexported -- either export it or duplicate (prefer export)
- Upstream Level and Downstream Level can be empty string (null) -- need careful int parsing
- `Capacity Node (Capacity Domain)` column has "True"/"False" strings, not booleans
- `Relationship Type` column in Dependencies.csv is always "Dependency" -- store but don't rely on it

## Requirements

### Functional
- Parse `Capacity Nodes.csv` -> `[]CapacityNodeType` rows
- Parse `Dependencies.csv` -> `[]DependencyRule` rows
- Parse `Impacts.csv` -> `[]ImpactRule` rows
- Validate CSV headers before parsing
- Skip empty/malformed rows with warnings
- Upsert all rows in a single transaction per CSV
- Idempotent: re-running produces same state

### Non-functional
- Reuse existing `readCSV()` helper (export it as `ReadCSV`)
- Each file under 200 lines

## Architecture

```
blueprint/
  Capacity Nodes.csv  ──> ParseCapacityNodesCSV() ──> []CapacityNodeTypeRow
  Dependencies.csv    ──> ParseDependenciesCSV()   ──> []DependencyRuleRow
  Impacts.csv         ──> ParseImpactsCSV()        ──> []ImpactRuleRow
                                                          │
                              ModelIngestionService.IngestAll()
                                          │
                              DB transaction: upsert all rows
```

## Related Code Files

### Files to Create
- `backend/internal/service/model_csv_parser.go` -- 3 parse functions + row structs
- `backend/internal/service/model_ingestion_service.go` -- orchestrates ingestion

### Files to Modify
- `backend/internal/service/blueprint_csv_parser.go` -- export `readCSV` -> `ReadCSV`

## Implementation Steps

### 1. Export `readCSV` in `blueprint_csv_parser.go`
Rename `readCSV` to `ReadCSV` (line 162). Update all callers in same file.

### 2. Create `backend/internal/service/model_csv_parser.go`

Row structs:
```go
type CapacityNodeTypeRow struct {
    NodeType         string
    Topology         string
    IsCapacityNode   bool
    ActiveConstraint bool
}

type DependencyRuleRow struct {
    NodeType                string
    DependencyNodeType      string
    RelationshipType        string
    TopologicalRelationship string
    UpstreamLevel           *int
}

type ImpactRuleRow struct {
    NodeType                string
    ImpactNodeType          string
    TopologicalRelationship string
    DownstreamLevel         *int
}
```

Parse functions:
- `ParseCapacityNodesCSV(filePath string) ([]CapacityNodeTypeRow, error)`
  - Header validation: expect 4 columns
  - Parse "True"/"False" strings to bool (case-insensitive)
  - Skip rows where NodeType is empty
- `ParseDependenciesCSV(filePath string) ([]DependencyRuleRow, error)`
  - Header validation: expect 5 columns
  - Parse UpstreamLevel: empty string -> nil, otherwise strconv.Atoi
  - Skip rows where NodeType or DependencyNodeType empty
- `ParseImpactsCSV(filePath string) ([]ImpactRuleRow, error)`
  - Header validation: expect 4 columns
  - Parse DownstreamLevel: empty string -> nil, otherwise strconv.Atoi
  - Skip rows where NodeType or ImpactNodeType empty

Helper:
```go
func parseBool(s string) bool {
    return strings.EqualFold(strings.TrimSpace(s), "true")
}

func parseOptionalInt(s string) *int {
    s = strings.TrimSpace(s)
    if s == "" { return nil }
    v, err := strconv.Atoi(s)
    if err != nil { return nil }
    return &v
}
```

### 3. Create `backend/internal/service/model_ingestion_service.go`

```go
type ModelIngestionSummary struct {
    CapacityNodesUpserted  int      `json:"capacity_nodes_upserted"`
    DependencyRulesUpserted int     `json:"dependency_rules_upserted"`
    ImpactRulesUpserted    int      `json:"impact_rules_upserted"`
    Errors                 []string `json:"errors,omitempty"`
    DurationMs             int64    `json:"duration_ms"`
}

type ModelIngestionService struct {
    db *gorm.DB
}

func NewModelIngestionService(db *gorm.DB) *ModelIngestionService
```

`IngestAll(basePath string) (*ModelIngestionSummary, error)`:
1. Build file paths: `filepath.Join(basePath, "Capacity Nodes.csv")` etc.
2. Parse all 3 CSVs (fail fast if any parse fails)
3. Single DB transaction:
   - Upsert capacity nodes (ON CONFLICT node_type DO UPDATE)
   - Upsert dependency rules (ON CONFLICT (node_type, dependency_node_type) DO UPDATE)
   - Upsert impact rules (ON CONFLICT (node_type, impact_node_type) DO UPDATE)
4. Return summary

Upsert pattern (reuse from blueprint_repository.go):
```go
tx.Clauses(clause.OnConflict{
    Columns:   []clause.Column{{Name: "node_type"}},
    DoUpdates: clause.AssignmentColumns([]string{"topology", "is_capacity_node", "active_constraint", "updated_at"}),
}).Create(&model)
```

### 4. CSV file path convention
Files live at `{BlueprintDir}/../` relative to Node & Edge folder, or use a new `MODEL_DIR` env var.

Simplest approach: model CSVs live at `./blueprint/` root (same level as domain folders). Add `ModelDir` to config with default `./blueprint`.

Update `backend/internal/config/config.go`:
```go
ModelDir: getEnv("MODEL_DIR", "./blueprint"),
```

## Todo List
- [x] Export readCSV -> ReadCSV in blueprint_csv_parser.go
- [x] Create model_csv_parser.go with 3 parse functions
- [x] Create model_ingestion_service.go with IngestAll
- [x] Add ModelDir to config.go
- [x] Verify `go build` compiles
- [x] Test with actual CSV files

## Success Criteria
- All 3 CSVs parsed without errors (24 + 147 + 118 rows)
- Upsert idempotent -- running twice yields same row count
- Malformed rows logged as warnings, not fatal errors
- Empty UpstreamLevel/DownstreamLevel correctly stored as NULL

## Risk Assessment
- **Low**: CSV format is stable, headers known
- **Medium**: bool parsing -- CSV uses "True"/"False" not "true"/"false". Using case-insensitive comparison mitigates.
- **Low**: readCSV export is a safe refactor (same package)

## Next Steps
- Phase 3 needs these tables populated to test CTE queries
