# Data Center Models v3 — Reference Guide

**Source:** `blueprint/Data Center Models v3_Feb 2026.xlsx` (14 sheets)
**Date extracted:** 2026-04-10

---

## 1. Node Types (NTList)

68 node types across 9 topology domains.

| Domain | Seq | Node Types |
|---|---|---|
| **Spatial** | a | Server, Storage Device, Networking Device, Rack, Row, Aisle, Zone, Room, Floor, Electrical Room, Mechanical Room, Cooling Yard, Generator Yard, Transformer Yard, Roof, IT Facility, Site |
| **Electrical** | b | Device PSU, Rack PDU Outlet, Rack PDU, RPP, Room PDU, UPS, BESS, Switch Gear, Generator, Transformer, Utility Feed, Lighting, Mechanical Power Panel |
| **Cooling** | c | RDHx, DTC, Liquid Loop, Air Zone, CDU, Air Cooling Unit, Cooling Distribution, Cooling Plant, Chiller, Condenser Chilled Water Pump, Condenser Water Pump, Cooling Tower Cell |
| **Whitespace** | d | Capacity Cell, Room Bundle, Room PDU Bundle, UPS Bundle, Containment Bundles |
| **Deployment** | e | Deployment, Cluster, Placement Plan, Placement Draft |
| **AI Infrastructure** | f | Fabric Domain, Inference Application, Compute Node, Job, Job Placement, Model, Model Replica, Queue, Project |
| **Asset Inventory** | g | Asset Class, Asset Type, ATS/STS, Secondary Pump |
| **Pod** | h | Pod |
| **Operational** | i | Business Unit, Team, Vendor |

---

## 2. Topology Schemes (TopologySchemes)

### Spatial Topology (Used=T)
```
Site → IT Facility → Floor → Room → Zone → Row → Rack → {Server, Networking Device, Storage Device, Rack PDU}
                   → Roof
                   → Generator Yard → Generator
                   → Cooling Yard → {Chiller, Cooling Tower Cell}
                   → Transformer Yard → Transformer
         Floor → Electrical Room
         Floor → Mechanical Room → Air Cooling Unit
         Zone → {Aisle, RPP, CDU}
```

### Electrical System (Used=T)
```
{Utility Feed, Generator} → Switch Gear → {UPS, Mechanical Power Panel, BESS}
BESS → {UPS, Mechanical Power Panel}
UPS → Room PDU → RPP → Rack PDU → Rack PDU Outlet → Device PSU → {Server, Storage, Networking}
Rack → Rack PDU (spatial containment)
Mechanical Power Panel → {Chiller, Cooling Tower Cell, Condenser Water Pump, Condenser Chilled Water Pump, Air Cooling Unit, CDU, Secondary Pump, Lighting}
```

### Cooling System (Used=T)
```
Cooling Plant → Cooling Distribution → {Air Cooling Unit, CDU}
                                      → {Chiller, Cooling Tower Cell, Condenser Water Pump, Condenser Chilled Water Pump}
Air Cooling Unit → Air Zone → Rack
CDU → Liquid Loop → Rack
Rack → {RDHx, DTC}
```

### Whitespace Blueprint (Used=T)
```
Site → Containment Bundles → {Room Bundle, Room PDU Bundle, UPS Bundle} → Capacity Cell → Rack
```
CC has 3 parents: Room Bundle, Room PDU Bundle, UPS Bundle (belongs to one of each type).

### Deployment Blueprint (Used=T)
```
Deployment → Cluster → {Placement Plan, Placement Draft} → {Capacity Cell, Rack}
```

### Others (Used=F mostly)
- **AI Infrastructure Fabric**: Cluster → Fabric Domain → Compute Node → {Model Replica, Job Placement}
- **Operational Infrastructure**: mirror of Spatial + Business Unit → Team → Project
- **Asset Inventory**: Asset Class → Asset Type → physical devices
- **Pod Operations**: Pod → Rack

---

## 3. Capacity Nodes

Nodes participating in capacity domain calculations:

| Node Type | Topology | Is Capacity Node | Active Constraint |
|---|---|---|---|
| RPP | Electrical | Yes | Yes |
| Room PDU | Electrical | Yes | Yes |
| UPS | Electrical | Yes | Yes |
| BESS | Electrical | Yes | Yes |
| Utility Feed | Electrical | Yes | Yes |
| Generator | Electrical | Yes | Yes |
| Air Zone | Cooling | Yes | Yes |
| Liquid Loop | Cooling | Yes | Yes |
| Cooling Distribution | Cooling | Yes | Yes |
| Cooling Plant | Cooling | Yes | Yes |
| Rack | Spatial | Yes | (no active constraint) |
| Capacity Cell | Whitespace | Yes | (passive constraint) |
| Room Bundle | Whitespace | Yes | (passive constraint) |
| Room PDU Bundle | Whitespace | Yes | (passive constraint) |
| UPS Bundle | Whitespace | Yes | (passive constraint) |

Non-capacity nodes: Rack PDU, Switch Gear, RDHx, DTC, Air Cooling Unit, CDU, Row, Zone, Aisle.

---

## 4. Edge Types (EdgeTypes)

Key constraints (Min_Link - Max_Link):

| Topology | Parent → Child | Cardinality |
|---|---|---|
| Spatial | Row → Rack | 1-100 |
| Spatial | Rack → Rack PDU | 1-10 |
| Electrical | RPP → Rack PDU | 1-100 |
| Electrical | Room PDU → RPP | 1-100 |
| Electrical | UPS → Room PDU | 1-100 |
| Cooling | Air Zone → Rack | 1-100 |
| Cooling | Liquid Loop → Rack | 1-100 |
| Whitespace | Capacity Cell → Rack | up to 10000 |
| Whitespace | Room/RPDUBundle/UPSBundle → CC | 1-100 |
| Deployment | PP/PD → Rack | up to 10000 |
| Deployment | Cluster → PP | max 1 |
| Deployment | Cluster → PD | max 25 |

---

## 5. Dependencies

Dependency rules define upstream relationships. Organized by source node:

**Rack dependencies (electrical + cooling):**
- Electrical: RPP(L1) → Room PDU(L2) → UPS(L3) → BESS(L4) → Switch Gear(L5) → {Utility Feed, Generator}(L6)
- Cooling: {Air Zone, Liquid Loop}(L1) → {Air Cooling Unit, CDU}(L2) → Cooling Distribution(L3) → Cooling Plant(L4)
- Local: RDHx, DTC

**Capacity Cell dependencies:**
- Local: RPP, Air Zone, Liquid Loop, RDHx, DTC
- Upstream: Room PDU(L1) → UPS(L2) → BESS(L3) → Switch Gear(L4) → {Utility Feed, Generator}(L5)
- Upstream cooling: Air Cooling Unit(L1) → CDU(L1) → Cooling Distribution(L2) → Cooling Plant(L3)

**Bundle dependencies** follow same pattern but Local includes more types as bundled scope grows.

---

## 6. Impacts

Impact rules define downstream "load" relationships. Key insight: **Load impacts cross topology boundaries**.

Upstream electrical/cooling nodes impact not just electrical downstream but also spatial/whitespace nodes:
- RPP impacts: Rack PDU(downstream L1) + {Rack, Row, Zone, CC, RB, RPDU-B, UPS-B}(load)
- Air Zone impacts: {Rack, Row, Zone, CC, RB, RPDU-B, UPS-B}(load)
- UPS impacts: Room PDU(L1) + RPP(L2) + Rack PDU(L3) + {Rack, Row, Zone, CC, RB, RPDU-B, UPS-B}(load)

This means: a UPS outage impacts all whitespace bundles whose member CCs have racks served by that UPS.

---

## 7. Variable Taxonomy (VT)

96 variable types across 5 functional domains:

| Domain | Classes | Count | Examples |
|---|---|---|---|
| **Power** | Input, Metric, Reference, Event | 42 | Power_Usage, Power_Headroom, Power_Fitness_Index, Breaker_PowerLimit |
| **Thermal** | Input, Metric, Reference, Event | 21 | Temperature, Thermal_Headroom, Thermal_Fitness_Index |
| **Cooling** | Input, Metric, Reference | 26 | Fan_Speed, Liquid_SupplyT_Margin, Cooling_Efficiency_Index |
| **Compute** | Metric, Reference | 4 | Batchability_PowerProxy_Factor, Batching_Fitness_Index |
| **Composite** | Metric | 3 | Node_Fitness_Index, IT_OperatingMargin, TrueCost_Power |

Variable Classes: **Input** (telemetry/measured), **Metric** (computed), **Reference** (static config), **Event** (triggered).

---

## 8. Variables (Updated) — Capacity-Relevant Subset

### Rack Variables
| Variable | Class | Description |
|---|---|---|
| Rack_Circuit_Capacity_(rated) | Reference | Branch circuit breaker limit |
| Rack_Circuit_Capacity_(design) | Reference | After derating |
| Allocated_ITLoad | Input | kW load allocated to rack |
| TheorMax_ITLoad | Input | Sum of peak device loads |
| Rack_AirCool_Fraction | Reference | Fraction cooled by air |
| Rack_LiquidCool_Fraction | Reference | Fraction cooled by liquid |
| Allocated_AirCool_Load | Metric | Derived: load × air fraction |
| Allocated_LiquidCool_Load | Metric | Derived: load × liquid fraction |
| Local_Capacity_Margin | Metric | Headroom at rack level |
| Upstream_Capacity_Margin | Metric | Min headroom across upstream |

### Row Variables
| Variable | Class |
|---|---|
| Rack_Circuit_Capacity_(design) | Reference (rollup from racks) |
| Rack_Positions | Reference |
| Rack_Count | Metric |
| Allocated_ITLoad | Metric (rollup) |

### Capacity Cell Variables
| Variable | Class | Notes |
|---|---|---|
| Capacity_Envelope | Metric | MIN(Power_Capacity, Thermal_Capacity) |
| Power_Capacity | Metric | Σ(RPP_Panel_Capacity_design) for member RPPs |
| Thermal_Capacity | Metric | Σ(AirZone_Cooling_Capacity) + Σ(LL_Cooling_Capacity) |
| Rack_Positions | Metric | Total rack positions in CC |
| Rack_Count | Metric | Actual racks |
| Allocated_ITLoad | Metric | Rollup from racks |
| Cell_Margin / Cell_Margin% | Metric | Envelope - load |
| Cell_Utilization% | Metric | Load / envelope |
| Power_Margin / Power_Margin% | Metric | Per-scope margins |
| Thermal_Margin / Thermal_Margin% | Metric | Per-scope margins |

### Bundle Variables (Room Bundle / Room PDU Bundle / UPS Bundle)
Same as CC plus:
| Variable | Class |
|---|---|
| Bundle_Margin / Bundle_Margin% | Metric |
| Bundle_Utilization% | Metric |

### Infrastructure Variables
| Node Type | Key Variables |
|---|---|
| RPP | RPP_Panel_Capacity_(rated/design), RPP_BreakerPole_Capacity |
| Room PDU | RPDU_Transformer_Rating, RPDU_Design_Capacity, RPDU_Operational_Cap |
| UPS | UPS_Rated_Capacity, UPS_Design_Capacity, UPS_Operational_Cap |
| Air Zone | AirZone_Cooling_Capacity, Maximum_Air_Flow, Rack_Inlet_Setpoint, ColdAisle_Supply_Temp |
| Liquid Loop | LL_Cooling_Capacity, Maximum_Flow_Rate, Loop_Supply_Setpoint, Loop_DeltaT_(design) |
| Air Cooling Unit | Rated_Cooling_Capacity, Maximum_Air_Flow |
| CDU | Rated_Cooling_Capacity, Pump_Capacity |
| RDHx | RDHx_MaxHeat_Removal |
| DTC | DTC_MaxCool_Capacity |

### Facility-Level
| Variable | Description |
|---|---|
| Facility_Deployable_ITCapacity | Total deployable IT capacity for facility |

---

## 9. Context Scopes (ContextScope)

Defines how variables consolidate across topologies:

| Context Group | Contexts | Consolidation Logic |
|---|---|---|
| **Power usage** | Electrical (topology), RkGroup (subgraph) | Smart Inclusive |
| **Rack rollup** | Spatial, Whitespace, Placement Plan, Placement Draft, Electrical, Air Cooling, Liquid Cooling | Standard Rollup |

"Smart Inclusive" = includes all power along path. "Standard Rollup" = sum descendants.

**Key for OAKw:** Rack rollup has Whitespace context — rolls up through CC → Bundle hierarchy. Also has separate PP and PD contexts for scenario-dependent rollups.

---

## 10. iDOT (Instance Data Object Template)

Runtime/telemetry variables per node type. Key entries:

| Node Type | DOT Variables |
|---|---|
| **Rack** | Power_Usage, Power_Fitness_Index, Thermal_Fitness_Index, Cooling_Efficiency_Index, Upstream_PowerRisk_Index + ~40 time-windowed metrics |
| **Server** | Power_Usage, Power/Thermal_Fitness_Index + 20 metrics |
| **Liquid Loop** | LiquidCooling_Fitness_Index, Liquid_SupplyT_Margin, Pump_Saturation + 8 metrics |
| **Compute Node** | Node_Fitness_Index, Batching_Fitness_Index |
| **Switch Gear / PDU / Rack PDU** | Breaker_Headroom |

---

## 11. iSET (Instance SET Template)

Static configuration fields per node type. Key entries:

| Node Type | SET Name | Fields |
|---|---|---|
| **Rack** | Operating Constraints | Rated_Power_Capacity, Breaker_PowerLimit.Inherited, Power_Derating_Factor, Effective_Power_Limit, Thermal_Limit |
| **Rack** | NFI Metric Weights | 11 weight fields (power/thermal/cooling/upstream) |
| **Rack** | Capacity | RCC_(rated/design), Allocated_ITLoad, TheorMax, AirCool/LiquidCool fractions, margins |
| **Row** | Capacity | RCC_(design), Rack_Positions, Rack_Count, Allocated_ITLoad |
| **RPP** | Capacity | RPP_Panel_Capacity_(rated/design), BreakerPole_Capacity, margins |
| **Air Zone** | Capacity | AirZone_Cooling_Capacity, Maximum_Air_Flow, Rack_Inlet_Setpoint, ColdAisle_Supply_Temp |
| **Liquid Loop** | Capacity | LL_Cooling_Capacity, Maximum_Flow_Rate, Loop_Supply_Setpoint, Loop_DeltaT_(design) |
| **Room PDU** | Capacity | RPDU_Transformer_Rating, Design_Capacity, Operational_Cap |
| **UPS** | Capacity | UPS_Rated/Design_Capacity, Operational_Cap |
| **CC** | Capacity + Margins | Capacity_Envelope, Power/Thermal_Capacity, Cell/Power/Thermal Margins (3 scopes: local, subdomain, upstream) |
| **Room/RPDU/UPS Bundle** | Capacity + Margins | Same structure as CC with Bundle_Margin instead of Cell_Margin |

---

## 12. NFI Data (Node Fitness Index)

147 unique telemetry/metadata parameter definitions for runtime monitoring:

| Data Type | Description | Count |
|---|---|---|
| Metric (Telemetry) | Real-time sensor readings | ~25 (power, temp, fan speed, pump) |
| Metadata (Static) | Device configuration/limits | ~45 (breaker limits, derating, weights) |
| Feature T1-T4 | Computed features at different time windows | ~77 (headroom, ramp rate, burstiness, coupling) |

---

## 13. Key Observations for OAKw Implementation

1. **Row→Rack spatial edge confirmed**: TopologySchemes shows `Row → Rack` (1-100 cardinality). OAKw step 1 can traverse this.

2. **CC→Rack whitespace edge confirmed**: `Capacity Cell → Rack` (up to 10000). CC packer output creates these edges.

3. **CC has 3 bundle parents**: In Whitespace Blueprint, CC's parent is {Room Bundle, Room PDU Bundle, UPS Bundle}. Each CC belongs to exactly one of each bundle type.

4. **PP/PD→Rack edges exist**: Deployment Blueprint shows `Placement Plan → Rack` and `Placement Draft → Rack`. Reservations create these edges.

5. **Capacity_Envelope formula ingredients available**: RPP_Panel_Capacity_(design) → Power_Capacity, AirZone_Cooling_Capacity + LL_Cooling_Capacity → Thermal_Capacity.

6. **Quality attributes NOT in data model**: No power_redundancy, cooling_redundancy, cooling_mode variables. Must be hardcoded or derived from topology structure.

7. **Scenario-aware rollups defined**: ContextScope has separate Placement Plan and Placement Draft rollup paths — confirms multi-scenario architecture.

8. **Impact rules cross topologies**: RPP/Air Zone/UPS impacts propagate to CC/Bundle level. This is the mechanism for upstream capacity margin at whitespace level.
