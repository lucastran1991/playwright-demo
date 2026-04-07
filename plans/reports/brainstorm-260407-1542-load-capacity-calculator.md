# Brainstorm: Load-Capacity Calculator

## Problem Statement

System needs a Load-Capacity Calculator to ingest node capacity data (from `ISET capacity - rack load flow.csv`), store it, compute bottom-up aggregations, and expose capacity metrics through API + existing Tracer DAG UI.

Currently the system can trace dependencies/impacts but has **no capacity data** — no kW values, no utilization metrics, no margin calculations. The blueprint PDF (v0.25) defines this as Backend Module #1, foundation for all subsequent load planning features.

## Decisions Made

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Scope | Load-Capacity Calculator only | Foundation first, placement/drag-drop later |
| Storage | `node_variables` table (key-value) | Sparse data (35 cols, most nodes use 3-5). Flexible. |
| Compute | Pre-compute + store | Fast reads, recompute on ingest. Standard for capacity tools. |
| Propagation | Bottom-up aggregation | Sum child loads to parent. Industry standard. |
| Metrics | Core set (5 metrics) | design_capacity, rated_capacity, allocated_load, available_capacity, utilization_pct |
| Derived data | Raw + derived in same table | `source` column distinguishes csv_import vs computed |
| Frontend | Extend existing Tracer DAG | Add capacity badges to DAG nodes, leverage existing UI |

## Data Analysis

### Source CSV: `ISET capacity - rack load flow.csv`
- 657 nodes, 35 columns, highly sparse
- Node type distribution: 264 Rack, 240 DTC, 31 ACU, 26 Row, 24 RPP, 24 RDHx, 9 Air Zone, 8 Room PDU, 7 LL, 7 CDU, 5 CC, 4 Room Bundle, 3 UPS, 3 Room PDU Bundle, 2 UPS Bundle

### Key columns per node type:
- **Rack**: Rack_Circuit_Capacity_(design/rated), Rack_LiquidCool_Fraction, Rack_AirCool_Fraction, Allocated_ITLoad, Allocated_LiquidCool_Load, Allocated_AirCool_Load
- **RPP**: RPP_Panel_Capacity_(design/rated), RPP_BreakerPole_Capacity, Allocated_ITLoad, Rack_Count
- **Air Zone**: AirZone_Cooling_Capacity, Allocated_AirCool_Load, Rack_Count, thermal params
- **Room PDU / UPS**: RPDU/UPS Design/Rated/Operational Capacity
- **RDHx/DTC**: RDHx_MaxHeat_Removal, DTC_MaxCool_Capacity
- **Row**: Allocated_ITLoad, Rack_Count
- **Capacity Cell / Bundles**: Capacity_Envelope, Power_Capacity, Thermal_Capacity

## Architecture

### Backend

#### 1. New Table: `node_variables`
```sql
CREATE TABLE node_variables (
    id            SERIAL PRIMARY KEY,
    node_id       VARCHAR(255) NOT NULL,     -- FK to blueprint_nodes.node_id
    variable_name VARCHAR(100) NOT NULL,     -- e.g. "design_capacity", "allocated_load"
    value         DOUBLE PRECISION NOT NULL,
    unit          VARCHAR(20) DEFAULT 'kW',  -- kW, %, fraction, count
    source        VARCHAR(20) NOT NULL,      -- 'csv_import' or 'computed'
    created_at    TIMESTAMP DEFAULT NOW(),
    updated_at    TIMESTAMP DEFAULT NOW(),
    UNIQUE(node_id, variable_name)
);
CREATE INDEX idx_node_variables_node_id ON node_variables(node_id);
CREATE INDEX idx_node_variables_name ON node_variables(variable_name);
```

#### 2. CSV Parser: `capacity_csv_parser.go`
- Parse `ISET capacity - rack load flow.csv`
- Map CSV columns to standardized variable names
- Column mapping per node type (only import non-empty values)

Variable name mapping:
| CSV Column | Variable Name | Unit | Applicable Types |
|-----------|--------------|------|-----------------|
| Rack_Circuit_Capacity_(design) | design_capacity | kW | Rack |
| Rack_Circuit_Capacity_(rated) | rated_capacity | kW | Rack |
| Allocated_ITLoad | allocated_load | kW | Rack, RPP, Row |
| RPP_Panel_Capacity_(design) | design_capacity | kW | RPP |
| RPP_Panel_Capacity_(rated) | rated_capacity | kW | RPP |
| RPDU_Design_Capacity | design_capacity | kW | Room PDU |
| RPDU_Operational_Cap | rated_capacity | kW | Room PDU |
| UPS_Design_Capacity | design_capacity | kW | UPS |
| UPS_Rated_Capacity | rated_capacity | kW | UPS |
| AirZone_Cooling_Capacity | design_capacity | kW | Air Zone |
| Rated_Cooling_Capacity | rated_capacity | kW | ACU, CDU |
| RDHx_MaxHeat_Removal | rated_capacity | kW | RDHx |
| DTC_MaxCool_Capacity | rated_capacity | kW | DTC |
| LL_Cooling_Capacity | design_capacity | kW | Liquid Loop |
| Capacity_Envelope | design_capacity | kW | CC, Bundles |
| Power_Capacity | power_capacity | kW | CC, Bundles |
| Thermal_Capacity | thermal_capacity | kW | CC, Bundles |
| Rack_LiquidCool_Fraction | liquid_cool_fraction | fraction | Rack |
| Rack_AirCool_Fraction | air_cool_fraction | fraction | Rack |
| Rack_Count | rack_count | count | RPP, AZ, Row |
| Allocated_LiquidCool_Load | allocated_liquid_load | kW | Rack, AZ |
| Allocated_AirCool_Load | allocated_air_load | kW | Rack, AZ |

#### 3. Ingestion Service: `capacity_ingestion_service.go`
- `IngestCapacityCSV(filePath)` — parse + upsert raw values
- `ComputeAggregates()` — bottom-up aggregation after raw import
- Idempotent (ON CONFLICT upsert)
- Transaction-wrapped

#### 4. Load Calculator: `load_capacity_calculator.go`
Bottom-up aggregation logic:

**Step 1: Leaf load (Rack)**
- `allocated_load` = from CSV (Allocated_ITLoad)
- `available_capacity` = rated_capacity - allocated_load
- `utilization_pct` = (allocated_load / rated_capacity) * 100

**Step 2: Aggregate up via impact rules (Load relationships)**
For each capacity node type with ActiveConstraint=true (RPP, Room PDU, UPS, etc.):
- Find all Load-impacted nodes (Rack, Row, Zone, CC, Bundles) via existing TraceImpacts logic
- `allocated_load` = SUM(child rack allocated_loads) -- only sum direct rack loads to avoid double-counting
- `available_capacity` = rated_capacity - allocated_load
- `utilization_pct` = (allocated_load / rated_capacity) * 100

**Step 3: Cooling aggregation**
For cooling nodes (Air Zone, Liquid Loop, CDU, Cooling Distribution, Cooling Plant):
- Same pattern: sum downstream rack cooling loads
- Air Zone: sum allocated_air_load from racks
- Liquid Loop: sum allocated_liquid_load from racks

**Double-counting prevention:**
- Only aggregate from **direct leaf nodes (Racks)**, not from intermediate nodes
- Use spatial hierarchy to find which racks belong to which RPP/Zone/CC
- This matches how the existing `FindSpatialDescendantsOfType` works

#### 5. API Endpoints

```
POST /api/capacity/ingest          -- Trigger CSV ingest + compute (protected)
GET  /api/capacity/nodes/:nodeId   -- Get capacity metrics for single node
GET  /api/capacity/summary         -- Aggregate summary across all capacity nodes
GET  /api/capacity/nodes           -- List nodes with capacity data (filter by type, min_utilization, etc.)
```

Response format for `/capacity/nodes/:nodeId`:
```json
{
  "node_id": "RPP-R1-1",
  "node_type": "RPP",
  "name": "RPP 1 for ROOM-1-1",
  "capacity": {
    "design_capacity": 230,
    "rated_capacity": 288,
    "allocated_load": 162,
    "available_capacity": 126,
    "utilization_pct": 56.25
  },
  "details": {
    "rack_count": 9,
    "breaker_pole_capacity": 54
  }
}
```

#### 6. Extend `/trace/full` response
Add `capacity` field to each traced node:
```json
{
  "upstream": [...],
  "downstream": [...],
  "local": [...],
  "load": [...],
  "source": {
    "node_id": "RACK-R1-Z1-R1-01",
    "capacity": {
      "design_capacity": 15,
      "rated_capacity": 18,
      "allocated_load": 15,
      "available_capacity": 3,
      "utilization_pct": 83.3
    }
  }
}
```

### Frontend

#### 1. Extend DAG node component
- Show capacity badge on node: `15/18 kW (83%)`
- Color code utilization: green (<60%), yellow (60-80%), red (>80%)
- Tooltip with full metrics on hover

#### 2. Extend right panel
- Add "Capacity" tab alongside existing topology/export controls
- Show detailed capacity breakdown for selected node
- Show upstream constraint chain (which parent is the bottleneck)

### File Map

| File | Action | Est. Lines |
|------|--------|-----------|
| `backend/internal/model/node_variable.go` | NEW | ~30 |
| `backend/internal/service/capacity_csv_parser.go` | NEW | ~120 |
| `backend/internal/service/capacity_ingestion_service.go` | NEW | ~80 |
| `backend/internal/service/load_capacity_calculator.go` | NEW | ~150 |
| `backend/internal/repository/capacity_repository.go` | NEW | ~100 |
| `backend/internal/handler/capacity_handler.go` | NEW | ~80 |
| `backend/internal/router/router.go` | MODIFY | +5 |
| `backend/internal/service/dependency_tracer.go` | MODIFY | +20 (attach capacity to trace response) |
| `frontend/src/components/tracer/dag-node.tsx` | MODIFY | +30 (capacity badge) |
| `frontend/src/components/tracer/dag-types.ts` | MODIFY | +10 (capacity types) |
| `frontend/src/components/tracer/dag-right-panel.tsx` | MODIFY | +40 (capacity tab) |

## Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|-----------|
| Double-counting in aggregation | High | Only aggregate from leaf (Rack) nodes; use spatial hierarchy for grouping |
| CSV column mapping errors | Medium | Validate against known node types; log unmapped columns |
| Performance with 657 nodes | Low | Pre-computed values, indexed queries. Tiny dataset. |
| Schema flexibility | Low | Key-value store handles new variables without migration |
| Aggregation correctness | High | Write integration tests comparing computed values against CSV source data |

## Success Criteria

1. All 657 nodes from CSV ingested with correct variable mapping
2. Core 5 metrics computed for every capacity node
3. Bottom-up aggregation matches manual calculation (spot-check RPP, UPS, Air Zone)
4. `/trace/full` response includes capacity data for each node
5. DAG nodes show utilization badge with color coding
6. Rack load validation: allocated_load <= rated_capacity

## Unresolved Questions

1. **Branch Circuit topology**: Slide 32-33 mentions Branch Circuit as intermediary between RPP and Rack. Current CSV doesn't include Branch Circuit nodes. Do we need to account for this in aggregation, or is it future scope?
2. **Cooling vs Power split**: Should utilization track electrical capacity and cooling capacity separately, or a single combined metric? CSV has both `Power_Capacity` and `Thermal_Capacity` for bundles.
3. **Active Constraint propagation**: When RPP is at 95% utilization, should downstream nodes (racks) show a "constrained" indicator even if the rack itself has room? This is implied by slide 18 but not explicitly spec'd.
4. **CSV update frequency**: Is this a one-time import, or will the CSV be updated regularly? Affects whether we need diff/merge logic.
