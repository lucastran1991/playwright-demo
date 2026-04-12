# Brainstorm: OAKw Full System (Steps 1-10)

**Date:** 2026-04-10
**Sources:** 4 PDF specs (RAPs Builder, CC Packer, CBD Packer, SBO Builder), Blueprint v0.5, Data Center Models v3 Excel, oakw-test-cases-draft.md, oak-value-calculator-explain.md
**Status:** Brainstorm complete — ready for planning

---

## Problem Statement

Build OAKw (Optionality Adjusted Kilowatt) evaluator — quantifies dollar value per kW for every rack in a data center, factoring in infrastructure quality, containment structure coherence, SBO membership optionality, and reservation impact across scenarios.

## Agreed Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Language | **All Go** | Single-language codebase, union-find/combinatorics trivial in Go |
| Phasing | **Steps 1-5 first**, pipeline + steps 6-10 later | Get value fast, SBO pipeline is prerequisite for steps 6-10 |
| Scenarios | **Full system** design, S_w first | Design schema for S_w/S_0/S_PD now, implement incrementally |
| Quality data | **Hardcode defaults Phase 1** | power_redundancy/cooling_redundancy/cooling_mode not in data model yet |
| Density | **Derive from existing data** | avg Rack_Circuit_Capacity_(design) per row — trivial |
| Reservation DB | **Design schema now, build Phase 2** | Document tables for scenarios/RST/RCST but no migration yet |

---

## Full System Architecture

### Pipeline Overview

```
Phase 1 (Immediate)                    Phase 2 (Later)
═══════════════════                    ═════════════════
                                       RAPs Builder (Go)
                                         ↓
                                       CC Packer (Go)
                                         ↓
                                       CBD Packer (Go)
                                         ↓
                                       SBO Builder (Go)
                                         ↓
Steps 1-5: OAKw Core Calculator  →→→  Steps 6-10: SBO Premium + Scenarios
  (non-scenario-dependent)              (scenario-dependent)
```

### Data Flow — Full 10 Steps

```
INPUTS:
  CC.Capacity_Envelope              ← MIN(Power_Capacity, Thermal_Capacity)
  Rack.RCC_design                   ← Rack_Circuit_Capacity_(design) from CSV
  Row quality attributes            ← Hardcoded Phase 1 / CSV Phase 2
  SBO associations                  ← SBO Builder output (Phase 2)
  Reservation status                ← Reservation tables (Phase 2)

STEP 1: CC → Row kW allocation
  Row.CCAC = CC.Envelope × (Row.Total_RCC / CC.Total_RCC)

STEP 2: Row quality profile (4 dimensions)
  density_premium     = f(avg RCC_design per row)     [derivable]
  power_redundancy    = lookup(row config)             [hardcoded Phase 1]
  cooling_redundancy  = lookup(row config)             [hardcoded Phase 1]
  cooling_mode        = lookup(row config)             [hardcoded Phase 1]

STEP 3: Quality-adjusted row value
  QP = Σ(premium_i × weight_i)
  Row.QV = Row.CCAC × $180 × QP

STEP 4: CC quality value + coherence
  Attribute_CF = f(distinct tier count across rows)
  CC_CF = Σ(Attr_CF × weight)
  CC.QV = Σ(Row.QV) × CC_CF

STEP 5: Distribute to racks
  Rack.QV = CC.QV × (Rack.RCC / CC.Total_RCC)
  ──────────── Phase 1 boundary ────────────

STEP 6: SBO membership → SPR          [needs SBO Builder]
  SPR per family: CC=5%, RB=2%, RPDU=3%, UPS=3%

STEP 7: Adjusted premium rate          [needs SBO + reservations]
  APR = SPR × Entanglement_Discount × Composite_Boost × RST_Effect

STEP 8: Final rack OAKw               [multiplicative]
  Rack.OAKw = Rack.QV × (1+APR_CC) × (1+APR_RB) × (1+APR_RPDU) × (1+APR_UPS)

STEP 9: Remaining Portfolio Value      [scenario-dependent]
  RPV = Σ(free rack OAKw)

STEP 10: Captured Capacity Value       [scenario-dependent]
  CCV = Σ(reserved rack OAKw per PP/PD)
```

---

## Phase 1: OAKw Steps 1-5 (Immediate)

### What We Need

| Input | Source | Status |
|---|---|---|
| CC.Capacity_Envelope | Computed: MIN(Power_Capacity, Thermal_Capacity) | Need to compute from AirZone/LL/RPP rollups |
| Rack.RCC_design | `Rack_Circuit_Capacity_(design)` in capacity CSV | Already ingested as node_variable |
| Row→Rack membership | Spatial topology edges | Already in blueprint_edges |
| CC→Row→Rack hierarchy | Whitespace topology | Needs CC Packer output OR manual CC assignment |
| Quality attributes | Hardcoded defaults | Phase 1 placeholder |

### Critical Dependency: CC Assignment

Steps 1-5 require knowing which racks belong to which Capacity Cell. Options:

**Option A — Implement CC Packer first (recommended for correctness)**
Build a simplified CC Packer in Go (union-find on shared AZ/LL/RPP). This gives real CC assignments from existing blueprint data. CC Packer is well-spec'd and self-contained.

**Option B — Manual/CSV CC assignment**
Add a CC assignment column to capacity CSV. Quick but fragile and requires manual data entry.

**Recommendation: Option A** — CC Packer is foundational for everything (OAKw, CBD Packer, SBO Builder). Building it first unlocks the entire downstream pipeline.

### Revised Phase 1 Scope

```
1. CC Packer (Go)          ← union-find on shared AZ/LL/RPP from RAP data
2. OAKw Steps 1-5          ← uses CC Packer output + existing capacity data
3. API endpoints            ← expose OAKw values per rack/CC/row
```

### Implementation Pattern

Follow existing `LoadCapacityCalculator` pattern:
- New `OakValueCalculator` struct with `ComputeAll()` method
- Top-down computation (CC → Row → Rack), inverse of existing bottom-up
- Store results in `node_variables` table with `source = "oak_computed"`
- New variables: `oak_quality_value`, `oak_quality_premium`, `oak_coherence_factor`, `oak_ccac`

### Constants (hardcoded in Go)

```go
const BaseKwValue = 180.0 // $/kW

var QualityWeights = map[string]float64{
    "density":             0.35,
    "power_redundancy":    0.20,
    "cooling_redundancy":  0.15,
    "cooling_mode":        0.30,
}

var DensityBrackets = []struct{ MaxKw, Premium float64 }{
    {15, 1.00}, {30, 1.05}, {60, 1.12}, {120, 1.22}, {math.MaxFloat64, 1.30},
}

// Phase 1 defaults — replace with real data later
var DefaultQualityProfile = map[string]float64{
    "power_redundancy":    1.00, // N (single feed)
    "cooling_redundancy":  1.00, // N
    "cooling_mode":        1.00, // Air Only
}
```

### New Go Files

```
backend/internal/service/
  oak_value_calculator.go          ← core calculation engine (steps 1-5)
  oak_value_constants.go           ← premium brackets, weights, SPR values

backend/internal/service/
  cc_packer.go                     ← union-find CC assignment from RAP data

backend/internal/handler/
  oak_handler.go                   ← API endpoints

backend/internal/model/
  capacity_cell.go                 ← CC model (surrogate_id, display_id, rack membership)
```

### API Endpoints (Phase 1)

```
POST   /api/oak/compute              # Trigger OAKw computation
GET    /api/oak/racks                 # All rack OAKw values (paginated)
GET    /api/oak/racks/:rackId         # Single rack OAKw detail
GET    /api/oak/capacity-cells        # CC-level summary (QV, CF, rack count)
GET    /api/oak/summary               # Facility-level summary (total QV, avg QV/rack)
```

---

## Phase 2: Whitespace Pipeline + Steps 6-10 (Deferred)

### Pipeline Modules (All Go)

| Module | Input | Output | Complexity |
|---|---|---|---|
| RAPs Builder | blueprint edges + TRAS config | rack → target associations | High (TRAS traversal) |
| CC Packer | RAP output | CC assignments + context | Medium (union-find) |
| CBD Packer | CC context | Bundle assignments + context | Medium (3x union-find) |
| SBO Builder | Base SBOs from CC+CBD | Composite SBOs + counts | Medium-High (combinatorics) |

Note: For Phase 1, we build a simplified CC Packer that directly uses blueprint edges (skip full RAPs Builder). Phase 2 implements the full RAPs Builder for proper TRAS-based traversal.

### Scenario Model Schema (Design Only — Phase 1)

```sql
-- Scenarios
CREATE TABLE scenarios (
    id            SERIAL PRIMARY KEY,
    code          VARCHAR(20) UNIQUE,     -- 'S_w', 'S_0', 'S_PD_001'
    scenario_type VARCHAR(20),            -- 'whitespace', 'planned', 'draft'
    name          VARCHAR(100),
    placement_draft_id INTEGER REFERENCES placement_drafts(id),
    created_at    TIMESTAMP DEFAULT NOW()
);

-- Placement hierarchy
CREATE TABLE clusters (
    id   SERIAL PRIMARY KEY,
    name VARCHAR(100)
);

CREATE TABLE placement_plans (
    id         SERIAL PRIMARY KEY,
    cluster_id INTEGER REFERENCES clusters(id),
    name       VARCHAR(100),
    status     VARCHAR(20) DEFAULT 'active'  -- active, archived
);

CREATE TABLE placement_drafts (
    id                SERIAL PRIMARY KEY,
    cluster_id        INTEGER REFERENCES clusters(id),
    name              VARCHAR(100),
    status            VARCHAR(20) DEFAULT 'draft',  -- draft, promoted, rejected
    promoted_to_pp_id INTEGER REFERENCES placement_plans(id)
);

-- Reservations
CREATE TABLE rack_reservations (
    id                 SERIAL PRIMARY KEY,
    rack_node_id       VARCHAR(255) REFERENCES blueprint_nodes(node_id),
    placement_plan_id  INTEGER REFERENCES placement_plans(id),
    placement_draft_id INTEGER REFERENCES placement_drafts(id),
    -- exactly one of pp/pd should be set
    CHECK (
        (placement_plan_id IS NOT NULL AND placement_draft_id IS NULL) OR
        (placement_plan_id IS NULL AND placement_draft_id IS NOT NULL)
    )
);

-- SBO reservation status (derived)
CREATE TABLE sbo_reservation_status (
    id              SERIAL PRIMARY KEY,
    scenario_id     INTEGER REFERENCES scenarios(id),
    sbo_id          VARCHAR(255),
    sbo_type        VARCHAR(50),     -- CC, ROOM_BUNDLE, ROOM_PDU_BUNDLE, UPS_BUNDLE, COMPOSITE_*
    rst             VARCHAR(30),     -- 'free', 'reserved', 'partially_reserved'
    free_rack_count INTEGER,
    total_rack_count INTEGER
);

-- Rack conflict status (derived)
CREATE TABLE rack_conflict_status (
    id              SERIAL PRIMARY KEY,
    rack_node_id    VARCHAR(255),
    rcst            VARCHAR(20),     -- 'none', 'soft_conflict', 'hard_conflict'
    conflicting_pp  INTEGER[],       -- PP IDs causing conflict
    conflicting_pd  INTEGER[]        -- PD IDs causing conflict
);
```

### Deferred Gaps (Placeholder Values)

| Gap | Placeholder | When to Resolve |
|---|---|---|
| CMP multipliers (count=2,3,n) | `boost = 1.0 + 0.1*(count-1)` linear | When product provides exact table |
| Entanglement discount values | `discount = 1.0` (no penalty) | When product provides exact factors |
| Custom Factor (5th dimension) | Omitted (weight=0) | When product defines scope |
| Quality attribute CSV columns | Hardcoded defaults per row | When data team adds columns to CSV |

---

## Test Coverage

92 test cases defined in `docs/oakw-test-cases-draft.md` (also CSV at `docs/oakw-test-cases.csv`):

| Phase | Features | Test Cases | Phase 1? |
|---|---|---|---|
| A (Steps 1-5) | F1-F12 | 43 TCs | YES |
| B (Steps 6-8) | F13-F16 | 19 TCs | Phase 2 |
| C (Steps 9-10) | F17-F21 | 13 TCs | Phase 2 |
| D (Support) | F22-F25 | 17 TCs | F22-F23 Phase 1, F24-F25 Phase 2 |

Phase 1 test scope: ~50 TCs (F1-F12 + F22-F23).

---

## Risk Assessment

| Risk | Impact | Mitigation |
|---|---|---|
| CC Packer needs RAP data that doesn't exist yet | High | Build simplified CC Packer using direct blueprint edges (AZ/LL/RPP → Rack edges exist) |
| Quality attributes hardcoded = inaccurate values | Medium | Calculator logic is correct; values are config. Easy to plug in real data later. |
| CC assignment may differ from manual expectations | Medium | Add validation: every rack assigned, no rack in multiple CCs. Display ID for UX. |
| Phase 2 scope is massive (4 pipeline modules + scenarios + reservations) | High | Break into sub-phases: 2a=full CC Packer+CBD, 2b=SBO Builder, 2c=scenarios+reservations, 2d=steps 6-10 |
| Multiplicative APR can produce unexpected values | Low | Add sanity bounds. Test case F16-04 explicitly verifies multiplicative behavior. |

---

## Unresolved Questions

1. **Row→Rack spatial edges** — Do blueprint edges contain Row→Rack parent-child? Need to verify. If not, Row may need to be derived from rack position/naming.
2. **CC Packer without full RAPs** — Can we build reliable CC assignments using only direct blueprint edges (Rack→AZ, Rack→LL, Rack→RPP)? Or do we need the full TRAS-based RAPs Builder?
3. **Capacity_Envelope computation** — Is this already computed by LoadCapacityCalculator, or does it need to be added? Excel defines it as MIN(Power_Capacity, Thermal_Capacity) where Power = Σ(RPP_Panel_Capacity_design), Thermal = Σ(AirZone_Cooling_Capacity) + Σ(LL_Cooling_Capacity).
4. **Phase 2 sub-phasing** — Exact boundaries between 2a/2b/2c/2d need planning when Phase 1 completes.
