# Phase 2: Load Calculator — Bottom-Up Aggregation

## Context Links
- Brainstorm: `plans/reports/brainstorm-260407-1542-load-capacity-calculator.md`
- Phase 1 (dependency): `phase-01-data-model-csv-ingestion.md`
- Existing tracer: `backend/internal/service/dependency_tracer.go` (TraceImpacts logic)
- Spatial queries: `backend/internal/repository/tracer_repository.go` (FindSpatialDescendantsOfType)
- Capacity node types: `backend/internal/model/capacity_node_type.go` (IsCapacityNode, ActiveConstraint)

## Overview
- **Priority**: P1 — core computation engine
- **Status**: pending
- **Description**: Compute bottom-up load aggregation from leaf Rack nodes up through RPP, Room PDU, UPS, Air Zone, Liquid Loop, etc. Store computed metrics (available_capacity, utilization_pct) in node_variables with `source=computed`.

## Key Insights
- **Double-counting prevention is critical**: Only aggregate from leaf Rack nodes, never from intermediate nodes. RPP's allocated_load = SUM(its child racks' allocated_load), not SUM(child RPPs).
- Existing `FindSpatialDescendantsOfType` already walks spatial edges to find Rack nodes under a parent — reuse this.
- Aggregation order matters: compute Rack metrics first (they have raw CSV data), then aggregate upward.
- Two parallel aggregation chains: **Power** (Rack -> RPP -> Room PDU -> UPS) and **Cooling** (Rack -> Air Zone -> ACU, Rack -> Liquid Loop -> CDU).
- The calculator should be called after CSV ingestion completes (IngestCSV -> ComputeAggregates).

## Requirements

### Functional
- Compute for every Rack: `available_capacity = rated_capacity - allocated_load`, `utilization_pct = (allocated_load / rated_capacity) * 100`
- For each capacity node with `ActiveConstraint=true`: aggregate child Rack loads
- Power chain: RPP sums its Racks' allocated_load; Room PDU sums its Racks'; UPS sums its Racks'
- Cooling chain: Air Zone sums its Racks' allocated_air_load; Liquid Loop sums allocated_liquid_load
- Capacity Cell / Bundles: sum all contained Racks' allocated_load
- Store all computed values with `source=computed`
- Re-computation: delete all `source=computed` rows first, then recompute

### Non-functional
- Must complete in <5s for 657 nodes (tiny dataset)
- Idempotent: delete computed + reinsert
- Transaction-wrapped

## Architecture

```
ComputeAggregates()
  |
  1. Delete all source=computed rows
  2. Compute Rack-level metrics (from CSV raw data)
  3. For each capacity node type (RPP, Room PDU, UPS, Air Zone, LL, CDU, CC, Bundles):
     a. Find all nodes of this type
     b. For each node, find descendant Rack nodes via spatial topology
     c. Sum their allocated_load (power) or allocated_air/liquid_load (cooling)
     d. Compute available_capacity, utilization_pct
     e. Upsert computed variables
```

### How to find "which Racks belong to this node"

Use spatial hierarchy. For each capacity node (e.g. RPP-R1-1):
1. Get its DB ID from blueprint_nodes
2. Call `FindSpatialDescendantsOfType([]uint{nodeDBID}, []string{"Rack"})` — walks spatial edges downward
3. For each found Rack, look up its allocated_load from node_variables

This is the same pattern the tracer already uses for Load impact resolution.

**Why spatial, not electrical?** Because electrical topology connects RPP -> BreakerPole -> RackPDU, not RPP -> Rack directly. Spatial topology has the containment hierarchy: Room -> Zone -> Row -> Rack.

**Fallback for nodes not in spatial tree**: Some capacity nodes (e.g. UPS) sit high in electrical topology. For these, walk electrical downstream to find RPPs/Room PDUs, then use spatial to find their Racks. This is the same multi-strategy approach from `TraceImpacts`.

## Related Code Files

### Files to CREATE
| File | Est. Lines | Purpose |
|------|-----------|---------|
| `backend/internal/service/load_capacity_calculator.go` | ~180 | Core aggregation logic |

### Files to MODIFY
| File | Change |
|------|--------|
| `backend/internal/service/capacity_ingestion_service.go` | Add `ComputeAggregates()` call after IngestCSV |
| `backend/internal/repository/capacity_repository.go` | Add `GetVariableMap`, `GetNodesByType` helpers |
| `backend/internal/repository/tracer_repository.go` | Possibly reuse existing spatial queries (no changes if sufficient) |

## Implementation Steps

### Step 1: Add repository helpers
File: `backend/internal/repository/capacity_repository.go`

```go
// GetVariableMap returns a map[nodeID]map[varName]float64 for quick lookup.
func (r *CapacityRepository) GetVariableMap(source string) (map[string]map[string]float64, error)

// GetNodeIDsByType returns all blueprint_node IDs + DB IDs for a given node_type.
func (r *CapacityRepository) GetNodeIDsByType(nodeType string) ([]struct{ ID uint; NodeID string }, error)
```

### Step 2: Create load capacity calculator
File: `backend/internal/service/load_capacity_calculator.go`

```go
type LoadCapacityCalculator struct {
    capRepo    *repository.CapacityRepository
    tracerRepo *repository.TracerRepository
    db         *gorm.DB
}

func NewLoadCapacityCalculator(
    capRepo *repository.CapacityRepository,
    tracerRepo *repository.TracerRepository,
    db *gorm.DB,
) *LoadCapacityCalculator

// ComputeAll deletes prior computed values, then computes and stores all aggregates.
func (c *LoadCapacityCalculator) ComputeAll() (*ComputeSummary, error)
```

**ComputeAll logic:**

```
1. tx := db.Begin()
2. Delete all node_variables WHERE source='computed'
3. Load all csv_import variables into memory: varMap[nodeID][varName] = value
4. Load capacity_node_types for IsCapacityNode + ActiveConstraint metadata

--- STEP A: Rack-level derived metrics ---
5. For each Rack node in varMap:
   - rated = varMap[rackID]["rated_capacity"]
   - allocated = varMap[rackID]["allocated_load"]
   - available = rated - allocated
   - utilPct = (allocated / rated) * 100 (guard: rated > 0)
   - Upsert: available_capacity, utilization_pct with source=computed

--- STEP B: Aggregate capacity nodes ---
6. Define aggregation config:
   aggregations = []struct{
     NodeType     string
     LoadVarName  string   // which rack variable to sum ("allocated_load", "allocated_air_load", etc.)
     CapVarName   string   // which variable holds this node's capacity ("rated_capacity")
   }{
     {"RPP", "allocated_load", "rated_capacity"},
     {"Room PDU", "allocated_load", "rated_capacity"},
     {"UPS", "allocated_load", "rated_capacity"},
     {"Air Zone", "allocated_air_load", "design_capacity"},
     {"Liquid Loop", "allocated_liquid_load", "design_capacity"},
     {"Air Cooling Unit", "allocated_air_load", "rated_capacity"},
     {"CDU", "allocated_liquid_load", "rated_capacity"},
     {"RDHx", "allocated_liquid_load", "rated_capacity"},
     {"DTC", "allocated_liquid_load", "rated_capacity"},
     {"Capacity Cell", "allocated_load", "design_capacity"},
     {"Room Bundle", "allocated_load", "design_capacity"},
     {"UPS Bundle", "allocated_load", "design_capacity"},
     {"Room PDU Bundle", "allocated_load", "design_capacity"},
   }

7. For each aggregation config:
   a. Get all nodes of this type: GetNodeIDsByType(nodeType)
   b. For each node:
      - Find descendant Racks via FindSpatialDescendantsOfType
      - Sum the LoadVarName across those Racks (from varMap)
      - aggregated_load = sum
      - capacity = varMap[nodeID][CapVarName] (from CSV)
      - available = capacity - aggregated_load
      - utilPct = (aggregated_load / capacity) * 100
      - Upsert: allocated_load (computed), available_capacity, utilization_pct

8. tx.Commit()
```

**Special cases:**
- **Row nodes**: aggregate allocated_load from child Racks (spatial descendants), but Row has no own capacity — only store aggregated allocated_load, not utilization
- **Nodes with no spatial Rack descendants**: log warning, skip (capacity = 0/0 = N/A)
- **Guard against division by zero**: if rated_capacity == 0, set utilization_pct = 0

### Step 3: Wire calculator into ingestion service
File: `backend/internal/service/capacity_ingestion_service.go`

Add `calculator *LoadCapacityCalculator` field. After `IngestCSV` completes, call `calculator.ComputeAll()`.

```go
type CapacityIngestionService struct {
    repo       *repository.CapacityRepository
    calculator *LoadCapacityCalculator
    db         *gorm.DB
}

func (s *CapacityIngestionService) IngestCSV(filePath string) (*CapacityIngestionSummary, error) {
    // ... parse + upsert raw values ...
    
    // Compute aggregates after raw import
    computeSummary, err := s.calculator.ComputeAll()
    if err != nil {
        summary.Errors = append(summary.Errors, "compute aggregates: "+err.Error())
    }
    summary.ComputedVariables = computeSummary.VariablesComputed
    return summary, nil
}
```

### Step 4: Handle UPS/high-level nodes with no direct spatial path to Racks

For nodes like UPS that don't have direct spatial children:
1. Walk electrical downstream from UPS to find Room PDUs
2. For each Room PDU, find its spatial descendant Racks
3. Sum those Racks' loads

Implementation: use `tracerRepo.FindDownstreamNodes(nodeDBID, "electrical-system", 5)` to find downstream nodes, filter for types that DO have spatial Rack children (RPP, Room PDU), then aggregate.

**Simpler alternative (preferred)**: Since we already compute RPP/Room PDU allocated_load in Step B, for UPS we can sum the already-computed allocated_load of its downstream Room PDUs instead of going back to Racks. BUT this creates ordering dependency.

**Decision**: Use the "find all descendant Racks" approach uniformly. It avoids ordering issues and the dataset is small (657 nodes). For UPS: walk spatial+whitespace down via `FindBridgeNodesViaSpatial` if needed, or walk electrical down to Room PDU, then spatial down from Room PDU to Racks.

## Todo List
- [ ] Add `GetVariableMap` and `GetNodeIDsByType` to capacity_repository.go
- [ ] Create `load_capacity_calculator.go` with `ComputeAll()`
- [ ] Implement Rack-level derived metrics (available_capacity, utilization_pct)
- [ ] Implement power chain aggregation (RPP, Room PDU, UPS)
- [ ] Implement cooling chain aggregation (Air Zone, LL, ACU, CDU, RDHx, DTC)
- [ ] Implement bundle/CC aggregation
- [ ] Handle UPS/high-level nodes via electrical+spatial hybrid walk
- [ ] Wire calculator into capacity_ingestion_service.go
- [ ] Verify `go build` compiles
- [ ] Manual smoke: check RPP utilization = sum(rack loads) / RPP rated_capacity

## Success Criteria
- All Rack nodes have computed available_capacity + utilization_pct
- RPP allocated_load = SUM of its child Racks' allocated_load (spot-check 2-3 RPPs)
- UPS allocated_load = SUM of all Racks under its electrical subtree
- No double-counting: each Rack counted once per aggregation parent
- Division by zero handled: 0 capacity -> 0% utilization

## Risk Assessment
| Risk | Severity | Mitigation |
|------|----------|-----------|
| Double-counting across overlapping hierarchies | High | Always aggregate from leaf Racks only, use `seen` set per parent |
| Spatial topology doesn't connect UPS to Racks | Medium | Use electrical->spatial hybrid walk as fallback |
| Ordering dependency in aggregation | Medium | Uniform "sum from Racks" approach avoids ordering |
| Performance with many spatial queries | Low | 657 nodes, ~20 capacity node types — trivial workload |

## Next Steps
- Phase 3 exposes computed data via API endpoints
