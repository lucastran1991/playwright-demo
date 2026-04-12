# Brainstorm: OAK Value Calculator

**Date:** 2026-04-10
**Source:** `docs/new_record.txt` (transcript)
**Status:** Brainstorm complete, no plan requested

## Problem Statement

Build OAK (Optionality Adjusted Kilowatt) Value Calculator — quantifies dollar value per kilowatt based on placement quality, containment structure, and future optionality in data center whitespace.

## Agreed Scope: Steps 1-5 Only (Non-Scenario)

Quality-adjusted rack value calculation. No SBO premiums, no reservations, no scenario comparison.

### 10-Step Overview (steps 6-10 deferred)

| Step | Description | Output | In scope? |
|------|------------|--------|-----------|
| 1 | Allocate capacity cell kW to rows by rack circuit proportion | Row allocated kW | Yes |
| 2 | Determine row quality profile (4 dimensions) | Quality profile | Yes |
| 3 | Row kW × $180 base × quality premium | Quality-adjusted row value | Yes |
| 4 | Sum rows + coherence factor discount | Capacity cell quality value | Yes |
| 5 | Distribute CC value to racks proportionally | Quality-adjusted rack value | Yes |
| 6 | SBO membership → standard premium rate | SPR | No |
| 7 | Composite boost + structural discounts | APR | No |
| 8 | Apply APR premiums to rack value | Final Rack OAK | No |
| 9 | Sum free rack OAK values | Remaining portfolio value | No |
| 10 | Sum reserved rack OAK values | Captured capacity value | No |

## Key Decisions

- **SBO Builder**: Not yet implemented, not in scope
- **Reservations**: No system exists, not in scope
- **Constants**: Hardcoded in Go code
- **Quality attributes**: Not in current CSV/DB — need to extend CSV parser
- **Premium brackets**: Use values from transcript as defaults
- **Output**: Full visualization (deferred until backend is solid)

## Quality Premium System

### 4 Dimensions with Weights
| Dimension | Weight | Example values from transcript |
|-----------|--------|-------------------------------|
| Density | 35% | Bracket-based on avg rack circuit capacity |
| Power redundancy | 20% | N=base, N+1, 2N=1.14x |
| Cooling redundancy | 15% | N=base, N+1, N+2=1.03x |
| Cooling mode | 30% | Air=base, air+liquid, full liquid=1.07x |

### Calculation
- Each row gets a value per dimension based on bracket
- Weighted average: `density×0.35 + power_red×0.20 + cool_red×0.15 + cool_mode×0.30`
- Result = single quality premium multiplier per row

### Reference values from transcript
- Density (high): 1.12
- Power 2N: 1.14
- Cooling N+2: 1.03
- Cooling full liquid: 1.07

## Architecture

### Computation flow
```
capacity_cell.envelope_kw
  → proportional allocation by rack_circuit_capacity → row_allocated_kw
  → × $180 × quality_premium → quality_adjusted_row_value
  → sum(rows) × coherence_factor → capacity_cell_quality_value
  → proportional distribution → rack_quality_adjusted_value (= OAK step 5)
```

### Coherence factor
- Measures uniformity of rows within capacity cell across 4 quality dimensions
- All same → factor = 1.0 (no discount)
- Mixed → factor < 1.0 (discount)
- Exact formula not specified in transcript — recommend variance-based approach

### Reuse from existing codebase
- `node_variables` table (sparse KV) for storing computed OAK values
- `capacity_repository.go` for spatial hierarchy queries
- `load_capacity_calculator.go` as pattern reference
- CSV parser extension for quality attributes

### New components needed
- `oak_value_calculator.go` — core calculation engine
- CSV parser extension — quality attribute columns
- API endpoints: compute trigger, per-node query, summary
- UI: heatmap + drill-down (later phase)

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Quality attributes not in CSV yet | Blocks calculation | Extend parser + test CSV |
| Transcript ambiguity on exact brackets | Medium | Use mentioned values, easy to tune |
| Coherence factor formula unspecified | Low | Implement variance-based discount |
| SBO/reservation integration later | Low now | Design struct to accept optional premiums |

## Unresolved Questions

1. Exact density bracket thresholds (what kW ranges map to low/medium/high/ultra)?
2. Full premium value table for all 4 dimensions × all brackets?
3. Coherence factor — exact formula? Is it per-dimension variance or cross-dimension?
4. When quality attributes are added to CSV, what column names/format?
5. SBO Builder spec — needed before steps 6-10 can be planned
