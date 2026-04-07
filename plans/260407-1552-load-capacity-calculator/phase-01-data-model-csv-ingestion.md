# Phase 1: Data Model + CSV Ingestion

## Context Links
- Brainstorm: `plans/reports/brainstorm-260407-1542-load-capacity-calculator.md`
- CSV source: `blueprint/ISET capacity - rack load flow.csv` (657 rows, 35 columns)
- Existing parser pattern: `backend/internal/service/model_csv_parser.go`
- Existing ingestion pattern: `backend/internal/service/model_ingestion_service.go`
- Existing blueprint parser: `backend/internal/service/blueprint_csv_parser.go` (uses `ReadCSV()`)

## Overview
- **Priority**: P1 — foundation for all subsequent phases
- **Status**: pending
- **Description**: Create `NodeVariable` model, CSV parser for capacity data, ingestion service, repository, and wire into GORM auto-migration.

## Key Insights
- CSV has 35 columns but is highly sparse — most nodes use 3-5 columns
- Key-value `node_variables` table avoids wide sparse columns and allows future variable additions without migration
- Existing `ReadCSV()` in `blueprint_csv_parser.go` handles lazy quotes + variable field count — reuse it
- Column mapping is node-type-specific: same variable name (e.g. `design_capacity`) maps to different CSV columns per node type
- CSV path is under `blueprint/` dir (same as ModelDir config: `./blueprint`)

## Requirements

### Functional
- New `node_variables` table with composite unique constraint on (node_id, variable_name)
- Parse all 35 CSV columns into standardized variable names per node type
- Upsert (ON CONFLICT) semantics — idempotent re-ingestion
- Only import non-empty numeric values (skip blank cells)
- Track source: `csv_import` for raw data, `computed` for Phase 2 aggregates

### Non-functional
- Transaction-wrapped ingestion
- Log unmapped/skipped columns for debugging
- No breaking changes to existing models or endpoints

## Architecture

```
capacity_csv_parser.go
  -> ParseCapacityFlowCSV(filePath) -> []CapacityFlowRow

capacity_ingestion_service.go
  -> IngestCapacityCSV(filePath) -> CapacityIngestionSummary
     Uses capacity_repository.go -> UpsertNodeVariable()

node_variable.go (model)
  -> NodeVariable struct with GORM tags

database.go
  -> Add &model.NodeVariable{} to AutoMigrate list
```

## Related Code Files

### Files to CREATE
| File | Est. Lines | Purpose |
|------|-----------|---------|
| `backend/internal/model/node_variable.go` | ~25 | GORM model |
| `backend/internal/service/capacity_csv_parser.go` | ~150 | Parse CSV, map columns to variables |
| `backend/internal/service/capacity_ingestion_service.go` | ~80 | Orchestrate parse + upsert |
| `backend/internal/repository/capacity_repository.go` | ~60 | DB operations for node_variables |

### Files to MODIFY
| File | Change |
|------|--------|
| `backend/internal/database/database.go` | Add `&model.NodeVariable{}` to `Migrate()` |
| `backend/internal/config/config.go` | Add `CapacityCSV` field (optional, can reuse `ModelDir`) |

## Implementation Steps

### Step 1: Create NodeVariable model
File: `backend/internal/model/node_variable.go`

```go
package model

import "time"

// NodeVariable stores per-node capacity metrics as key-value pairs.
// Sparse data (35 CSV columns, most nodes use 3-5) makes KV more efficient than wide table.
type NodeVariable struct {
    ID           uint      `gorm:"primaryKey" json:"id"`
    NodeID       string    `gorm:"uniqueIndex:idx_nv_node_var;size:255;not null" json:"node_id"`
    VariableName string    `gorm:"uniqueIndex:idx_nv_node_var;size:100;not null" json:"variable_name"`
    Value        float64   `gorm:"not null" json:"value"`
    Unit         string    `gorm:"size:20;default:'kW'" json:"unit"`
    Source       string    `gorm:"size:20;not null" json:"source"` // csv_import | computed
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

### Step 2: Register model in auto-migration
File: `backend/internal/database/database.go`

Add `&model.NodeVariable{}` to the `db.AutoMigrate(...)` call in `Migrate()`.

### Step 3: Create capacity repository
File: `backend/internal/repository/capacity_repository.go`

```go
// CapacityRepository handles node_variables DB operations.
type CapacityRepository struct {
    db *gorm.DB
}

func NewCapacityRepository(db *gorm.DB) *CapacityRepository

// UpsertNodeVariable upserts a single node variable (ON CONFLICT update value, unit, source).
func (r *CapacityRepository) UpsertNodeVariable(tx *gorm.DB, nv *model.NodeVariable) error

// GetNodeVariables returns all variables for a given node_id.
func (r *CapacityRepository) GetNodeVariables(nodeID string) ([]model.NodeVariable, error)

// GetNodeVariablesByName returns a specific variable across all nodes (e.g. "utilization_pct").
func (r *CapacityRepository) GetNodeVariablesByName(varName string) ([]model.NodeVariable, error)

// DeleteBySource deletes all node_variables with given source (for re-computation).
func (r *CapacityRepository) DeleteBySource(tx *gorm.DB, source string) error

// BulkUpsert upserts a batch of node variables in a single transaction.
func (r *CapacityRepository) BulkUpsert(tx *gorm.DB, vars []model.NodeVariable) error
```

Use `clause.OnConflict` with `Columns: idx_nv_node_var` pattern — same as existing `blueprint_repository.go`.

### Step 4: Create capacity CSV parser
File: `backend/internal/service/capacity_csv_parser.go`

**Key design**: Build a column-index map from the CSV header, then for each row:
1. Read `node_id` (col 0), `node_type` (col 1), `name` (col 2)
2. Based on `node_type`, check the relevant CSV columns
3. For each non-empty numeric value, produce a `CapacityVariable` with standardized name + unit

**Column mapping table** (from brainstorm, verified against CSV header):

| Col# | CSV Column | Var Name | Unit | Node Types |
|------|-----------|----------|------|------------|
| 3 | Rack_Circuit_Capacity_(design) | design_capacity | kW | Rack |
| 4 | Rack_Circuit_Capacity_(rated) | rated_capacity | kW | Rack |
| 5 | Rack_LiquidCool_Fraction | liquid_cool_fraction | fraction | Rack |
| 6 | Rack_AirCool_Fraction | air_cool_fraction | fraction | Rack |
| 7 | Allocated_ITLoad | allocated_load | kW | Rack, RPP, Row |
| 8 | Allocated_LiquidCool_Load | allocated_liquid_load | kW | Rack, Air Zone |
| 9 | Allocated_AirCool_Load | allocated_air_load | kW | Rack, Air Zone |
| 10 | Rack_Count | rack_count | count | RPP, Air Zone, Row |
| 11 | Rated_Cooling_Capacity | rated_capacity | kW | Air Cooling Unit, CDU |
| 12 | Pump_Capacity | pump_capacity | kW | CDU |
| 13 | RPP_Panel_Capacity_(design) | design_capacity | kW | RPP |
| 14 | RPP_BreakerPole_Capacity | breaker_pole_capacity | count | RPP |
| 15 | RPP_Panel_Capacity_(rated) | rated_capacity | kW | RPP |
| 16 | AirZone_Cooling_Capacity | design_capacity | kW | Air Zone |
| 17 | Rack_DeltaT_(design) | rack_delta_t | degC | Air Zone |
| 18 | Maximum_Air_Flow | max_air_flow | CFM | Air Zone |
| 19 | ColdAisle_Supply_Temp | cold_aisle_supply_temp | degC | Air Zone |
| 20 | Rack_Inlet_Setpoint | rack_inlet_setpoint | degC | Air Zone |
| 21 | LL_Cooling_Capacity | design_capacity | kW | Liquid Loop |
| 22 | Loop_Supply_Setpoint | loop_supply_setpoint | degC | Liquid Loop |
| 23 | Loop_DeltaT_(design) | loop_delta_t | degC | Liquid Loop |
| 24 | Maximum_Flow_Rate | max_flow_rate | LPM | Liquid Loop |
| 25 | RPDU_Operational_Cap | rated_capacity | kW | Room PDU |
| 26 | RPDU_Design_Capacity | design_capacity | kW | Room PDU |
| 27 | RPDU_Transformer_Rating | transformer_rating | kVA | Room PDU |
| 28 | UPS_Operational_Cap | rated_capacity | kW | UPS |
| 29 | UPS_Design_Capacity | design_capacity | kW | UPS |
| 30 | UPS_Rated_Capacity | ups_rated_capacity | kW | UPS |
| 31 | RDHx_MaxHeat_Removal | rated_capacity | kW | RDHx |
| 32 | DTC_MaxCool_Capacity | rated_capacity | kW | DTC |
| 33 | Capacity_Envelope | design_capacity | kW | Capacity Cell, Room Bundle, UPS Bundle, Room PDU Bundle |
| 34 | Power_Capacity | power_capacity | kW | Capacity Cell, Room Bundle, UPS Bundle, Room PDU Bundle |
| 35 | Thermal_Capacity | thermal_capacity | kW | Capacity Cell, Room Bundle, UPS Bundle, Room PDU Bundle |

**Parser struct**:
```go
type CapacityFlowRow struct {
    NodeID   string
    NodeType string
    Name     string
    Variables []CapacityVariable
}

type CapacityVariable struct {
    VariableName string
    Value        float64
    Unit         string
}

func ParseCapacityFlowCSV(filePath string) ([]CapacityFlowRow, error)
```

**Implementation approach**:
- Use `ReadCSV(filePath)` from `blueprint_csv_parser.go` (already handles edge cases)
- Build header index map: `colMap[headerName] = colIndex`
- Define per-node-type column mappings as a `map[string][]columnMapping` struct
- For each data row, look up node_type, iterate its column mappings, parse non-empty float values
- Use `strconv.ParseFloat` with `strings.TrimSpace`

### Step 5: Create capacity ingestion service
File: `backend/internal/service/capacity_ingestion_service.go`

```go
type CapacityIngestionSummary struct {
    NodesProcessed    int      `json:"nodes_processed"`
    VariablesUpserted int      `json:"variables_upserted"`
    RowsSkipped       int      `json:"rows_skipped"`
    Errors            []string `json:"errors,omitempty"`
    DurationMs        int64    `json:"duration_ms"`
}

type CapacityIngestionService struct {
    repo *repository.CapacityRepository
    db   *gorm.DB
}

func NewCapacityIngestionService(repo *repository.CapacityRepository, db *gorm.DB) *CapacityIngestionService

// IngestCSV parses capacity CSV and upserts raw variables.
func (s *CapacityIngestionService) IngestCSV(filePath string) (*CapacityIngestionSummary, error)
```

**Logic**:
1. Call `ParseCapacityFlowCSV(filePath)`
2. Wrap in `db.Transaction`
3. For each row, for each variable: upsert with `source = "csv_import"`
4. Return summary with counts

### Step 6: Wire into main.go (deferred to Phase 3)
Wiring the handler/router happens in Phase 3. This phase only creates the model, parser, repo, and service.

## Todo List
- [ ] Create `node_variable.go` model
- [ ] Add to `database.go` AutoMigrate
- [ ] Create `capacity_repository.go` with UpsertNodeVariable, GetNodeVariables, DeleteBySource, BulkUpsert
- [ ] Create `capacity_csv_parser.go` with ParseCapacityFlowCSV
- [ ] Create `capacity_ingestion_service.go` with IngestCSV
- [ ] Verify `go build` compiles
- [ ] Manual smoke test: run ingestion, check DB has ~2000+ node_variables rows

## Success Criteria
- `go build` passes with no errors
- All 657 CSV rows parsed (minus any with empty node_id)
- Correct variable mapping: Rack nodes get design_capacity, rated_capacity, allocated_load, etc.
- Idempotent: running twice produces same result (no duplicates)
- Only non-empty values stored

## Risk Assessment
| Risk | Severity | Mitigation |
|------|----------|-----------|
| CSV column order changes | Medium | Use header-based index mapping, not hardcoded positions |
| Float parsing edge cases | Low | TrimSpace + skip empty; log parse errors |
| node_id in CSV not in blueprint_nodes | Medium | Store anyway (node_variables has no FK constraint); log warning |

## Security Considerations
- No auth needed for ingestion endpoint (matches existing pattern: `POST /api/blueprints/ingest` is public)
- No user input beyond file path (server-side CSV)
