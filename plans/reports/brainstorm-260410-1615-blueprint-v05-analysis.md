# Brainstorm: Blueprint UX/Backend v0.5 Analysis

**Date:** 2026-04-10
**Source:** `blueprint/Blueprint_UX_backend_v0.5 (Additional BE).pptx.pdf` (103 pages)
**Status:** Analysis complete

## Document Structure (103 pages)

| Section | Pages | Content |
|---|---|---|
| User Workflows | 2-6 | Manual, partial auto, full auto-placement workflows |
| UX/UI | 7-23 | Focus/dimming, labels, badges, dependency tracer, load drag-drop |
| Backend (WIP) | 26-40 | 8 backend modules, Load Capacity Calculator steps, variables, margins |
| Whitespace Packing | 40-44 | RAPs Builder → CC Packer → CBD Packer → SBO Builder pipeline |
| Deployment Planning | 45-57 | Deployment blueprint, scenarios, reservations, conflicts, RST/RCST |
| OAKw Evaluator | 58-71 | Full 10-step spec with exact formulas, premium tables, examples |
| Variable Reclarification | 72-80 | Semantic anchoring, context scopes, consolidation rules |
| Topology (WIP) | 84-100 | Spatial, electrical, cooling, whitespace, deployment topology diagrams |
| Old/Reference | 100-103 | Older reservation specs |

---

## Key Findings — What's New vs Codebase

### 1. OAKw Evaluator — EXACT SPEC NOW AVAILABLE

PDF pg 58-71 provides complete specification. Critical corrections from earlier transcript-based estimates:

| Item | Transcript estimate | PDF spec (authoritative) |
|---|---|---|
| Density tiers | 4 tiers (<10, 10-30, 30-60, >60) | **5 tiers** (<=15, 15-30, 30-60, 60-120, >=120) |
| Power redundancy | N, N+1, 2N | **N, N+1, A/B feeds, 2N** (4 tiers) |
| Cooling redundancy | N, N+1, N+2 | **N, N+1, N+2, N+N dual loop** (4 tiers) |
| Cooling mode | Air, Air+Liquid, Full Liquid | **Air, Air+RDHx, Air+DTC, Full Liquid DTC** |
| Coherence factor | 1/N approach (guessed) | **Tier-count: 1=1.0, 2=0.98, 3+=0.92** |
| SPR values | Unknown | **CC=5%, RB=2%, RPDU=3%, UPS=3%** |
| Rack.OAKw formula | Additive (QV + premiums) | **Multiplicative: QV × (1+APR_CC) × (1+APR_RB) × ...** |
| CCAC formula | Proportional (guessed) | **Exact: CC.Envelope × (Row.RCC_design / CC.RCC_design)** |

**Updated `docs/oak-value-calculator-explain.md`** with all corrected values.

### 2. Eight Backend Modules Defined

| # | Module | Status in Codebase | Complexity |
|---|---|---|---|
| 1 | Load-Capacity Calculator | **Partially implemented** (basic bottom-up aggregation exists) | High — needs PP/PD versioning, margins, upstream/subdomain margins |
| 2 | Variable data/models | **Partially implemented** (node_variables sparse KV exists) | Medium — needs variable classification (Reference/Input/Metric) |
| 3 | OP and Metrics | Not implemented | Low-Medium |
| 4 | OAKw Evaluator | Not implemented | High |
| 5 | Auto-placement: Optionality Value Optimizer | Not implemented | Very High |
| 6 | Auto-placement: Capacity Packer | Not implemented | Very High |
| 7 | Whitespace Packing (Blueprint Compiler) | Not implemented | Very High — 4-stage pipeline |
| 8 | Topology update manager | Not implemented | Medium |

### 3. Load Capacity Calculator — Much More Complex Than Current Implementation

Current codebase does simple bottom-up aggregation. PDF specifies **5 major computation phases**:

1. **Capacity** — Roll up RCC_design, rack positions, compute CC Power/Thermal/Envelope
2. **PP/PD Load** — Versioned loads (Placement Plan vs Draft), air/liquid split
3. **Consolidated Load** — Planned Load (public) vs Potential Load (private per PD)
4. **Margins** — Local, Device, Cell, Power, Thermal margins
5. **Upstream/Subdomain Margins** — Reference tables, topology-aware

**Major gap: PP/PD versioning**. Current system has single-version loads. PDF requires multi-version: PP.Allocated_ITLoad, PD.Allocated_ITLoad per draft. Every upstream rollup inherits multiple versions.

### 4. Whitespace Packing — New 4-Stage Pipeline

```
RAPs Builder → CC Packer → CBD Packer → SBO Builder
```

- **RAPs Builder**: Resolved Association Paths for Racks. Lookup table: rack → {rooms, RPPs, AZ, LL, Room PDU, UPS}
- **CC Packer**: Group racks into Capacity Cells (smallest group not sharing {AZ, LL, RPP} with outside racks)
- **CBD Packer**: Group CCs into Containment Bundles (Room Bundle, Room PDU Bundle, UPS Bundle)
- **SBO Builder**: Build Composite SBOs within each family, output `sbo_association_summary`

Each stage has detailed requirements docs (Google Docs links in PDF).

### 5. Scenario Model — Core Architecture Concept

| Scenario | Code | Description |
|---|---|---|
| Whitespace | S_w | Everything free, no reservations |
| Planned | S_0 | All Placement Plans included, no drafts |
| Draft | S_(PD) | Planned + one specific Placement Draft |

**Scenario-dependent data**: All loads, margins, RSTs, RCSCs, OAKw values. This is a **cross-cutting concern** affecting almost every calculation.

### 6. Reservation System — Complex State Machine

- **Reservable entities**: Racks + Base SBOs (CCs, CBDs)
- **RST states**: Free, Reserved, Partially Reserved
- **RCST (conflicts)**: Hard Conflict, Soft Conflict, None
- **5-step reservation flow**: Validate → Status change → Update rack RST → Reconcile SBO RST → Update conflicts
- PDs with Hard Conflicts cannot become PPs
- RST is scenario-dependent (public RST for S_0, private RST for each S_PD)

### 7. UX/UI Features — Priority Matrix

| Feature | Task Priority | Feature Importance |
|---|---|---|
| Focus & Dimming (frame selection) | H | H |
| Persistent Labels | M | M |
| Expanding Labels → Badges | L | M |
| Actions from Badges | L | M |
| Dependency Tracer enhancements | L | H |
| Load Drag & Drop | L | M |

### 8. Variable Reclarification — Important Architecture Shift

- Variables are **Globally Anchored**, **Locally Anchored**, or **Local Only**
- Most DC variables are locally anchored or local only
- **Context Scope** replaces topology-based consolidation
- Consolidation Rule = Variable + Context Scope + Consolidation Logic
- Recommended: Start with Approach A (simple node type list), migrate to Approach B (declarative context definition) later

---

## Gap Analysis: Current Codebase vs Full Spec

### Already Implemented
- Blueprint CSV ingestion + node_variables storage
- Basic bottom-up load aggregation (single version)
- Dependency tracer (backend: upstream/downstream/local, DAG view)
- Capacity summary API endpoints

### Major Gaps (ordered by dependency)

1. **Whitespace Packing** — RAPs, CC Packer, CBD Packer, SBO Builder. Creates the whitespace topology that everything else depends on.
2. **Scenario Model** — Core construct needed before multi-version loads, reservations, OAKw.
3. **Reservation System** — RST/RCST state machine, 5-step flows for reserve/release.
4. **Load Capacity Calculator v2** — PP/PD versioning, air/liquid split, margins, upstream/subdomain.
5. **OAKw Evaluator** — Steps 1-10, depends on Whitespace Packing (SBO data) + Reservations (RST).
6. **Deployment Planning** — Cluster/PP/PD hierarchy, free-floating racks, PD→PP promotion.
7. **Auto-placement** — Optionality Value Optimizer + Capacity Packer (last, most complex).

### Not Urgent / Future
- Topology update manager
- Branch Circuit modification
- Context Scope Approach B (declarative)

---

## Recommendations

### For OAKw (scoped to steps 1-5 as previously agreed)
PDF provides exact formulas — **no more ambiguity**. Steps 1-5 can be implemented with current data IF quality attributes (power_redundancy, cooling_redundancy, cooling_mode) are added to CSV/DB.

Key formula corrections applied to `docs/oak-value-calculator-explain.md`:
- 5 density tiers, 4 power tiers, 4 cooling red tiers, 4 cooling mode tiers
- Coherence factor: tier-count based (1.0 / 0.98 / 0.92)
- Multiplicative APR formula (not additive)
- SPR values: 5% / 2% / 3% / 3%

### For broader roadmap
The full system is massive. Suggest phased approach:
1. **Phase 1** (current): OAKw steps 1-5 core calculator
2. **Phase 2**: Whitespace Packing pipeline (enables SBO data)
3. **Phase 3**: Scenario Model + Reservation system
4. **Phase 4**: OAKw steps 6-10 (needs phases 2+3)
5. **Phase 5**: Load Capacity Calculator v2 (PP/PD versioning)
6. **Phase 6**: Auto-placement engine

---

## Unresolved Questions

1. **CMP table incomplete** — Compound Membership Premium values for count=2,3,n are blank in PDF. Need exact multipliers.
2. **Entanglement discount values** — Rules described but exact numeric factors not specified.
3. **Custom Factor** — 5th quality dimension mentioned but undefined.
4. **Variable source for quality attributes** — PDF references Google Sheets tab [Variables]. Need access to determine which CSV columns map to quality attributes.
5. **Capacity Envelope formula** — PDF shows `MIN(Power_Capacity, Thermal_Capacity)`. Current codebase uses `design_capacity`. Need to verify if this is already computed or needs new calculation.
