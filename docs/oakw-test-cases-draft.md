# OAKw Evaluator — QA Test Cases Draft

## Features & Test Cases

---

### F1. CC-Allocated Capacity for Row (Row.CCAC)

**Formula:** `Row.CCAC = CC.Capacity_Envelope x (Row.Total_RCC_design / CC.Total_RCC_design)`

| TC# | Description | Input | Expected Output |
|---|---|---|---|
| F1-01 | Basic proportional allocation | CC envelope=500kW, Row-A RCC=200, Row-B RCC=300, CC total=500 | Row-A CCAC=200, Row-B CCAC=300 |
| F1-02 | Sum of all Row.CCAC = CC envelope | CC envelope=400kW, 3 rows with RCC 100/150/250 | Sum = 400kW |
| F1-03 | Envelope < total rack circuit (oversubscribed) | CC envelope=300kW, total RCC=600kW, Row RCC=200 | Row CCAC = 100kW (300 x 200/600) |
| F1-04 | Single row in CC | CC envelope=500kW, 1 row, RCC=600 | Row CCAC = 500kW |
| F1-05 | Row with zero racks | CC envelope=500kW, Row RCC=0 | Row CCAC = 0 |

---

### F2. CC-Allocated Capacity for Rack (Rack.CCAC)

**Formula:** `Rack.CCAC = CC.Capacity_Envelope x (Rack.RCC_design / CC.Total_RCC_design)`

| TC# | Description | Input | Expected Output |
|---|---|---|---|
| F2-01 | Basic rack allocation | CC envelope=400kW, rack RCC=40, CC total=480 | Rack CCAC = 33.33kW |
| F2-02 | Sum of all Rack.CCAC = CC envelope | 12 racks, each RCC=40, CC total=480, envelope=400 | Sum = 400kW |
| F2-03 | Unequal rack sizes | Racks: 10kW, 20kW, 30kW. CC total=60, envelope=48 | 8, 16, 24 |

---

### F3. Row Density Tier

**Rule:** Average rack circuit capacity per row -> bracket

| TC# | Description | Input (avg kW/rack) | Expected Tier | Premium |
|---|---|---|---|---|
| F3-01 | Tier 1 boundary | 15 kW | Tier 1 | 1.00 |
| F3-02 | Tier 2 lower boundary | 15.01 kW | Tier 2 | 1.05 |
| F3-03 | Tier 2 upper boundary | 30 kW | Tier 2 | 1.05 |
| F3-04 | Tier 3 | 30.01 kW | Tier 3 | 1.12 |
| F3-05 | Tier 4 | 60.01 kW | Tier 4 | 1.22 |
| F3-06 | Tier 5 | 120 kW | Tier 5 | 1.30 |
| F3-07 | Mixed density row (avg) | Racks: 10, 20, 60 kW. Avg=30 | Tier 2 | 1.05 |
| F3-08 | Single rack row | 45 kW | Tier 3 | 1.12 |

---

### F4. Row Power Redundancy Tier

| TC# | Description | Input | Expected Premium |
|---|---|---|---|
| F4-01 | N (single feed) | "N" | 1.00 |
| F4-02 | N+1 | "N+1" | 1.03 |
| F4-03 | A/B feeds | "A/B" | 1.08 |
| F4-04 | 2N architecture | "2N" | 1.14 |
| F4-05 | Unknown/missing value | null or "" | Default to 1.00 (N) |

---

### F5. Row Cooling Redundancy Tier

| TC# | Description | Input | Expected Premium |
|---|---|---|---|
| F5-01 | N | "N" | 1.00 |
| F5-02 | N+1 | "N+1" | 1.03 |
| F5-03 | N+2 | "N+2" | 1.08 |
| F5-04 | N+N (dual loop) | "N+N" | 1.14 |
| F5-05 | Unknown/missing value | null or "" | Default to 1.00 (N) |

---

### F6. Row Cooling Mode

| TC# | Description | Input | Expected Premium |
|---|---|---|---|
| F6-01 | Air Only | "Air" | 1.00 |
| F6-02 | Air + RDHx | "Air+RDHx" | 1.03 |
| F6-03 | Air + DTC | "Air+DTC" | 1.07 |
| F6-04 | Full Liquid DTC | "FullLiquid" | 1.07 |
| F6-05 | Unknown/missing value | null or "" | Default to 1.00 (Air) |

---

### F7. Row Quality Premium (QP)

**Formula:** `QP = Σ(attribute_premium x attribute_weight)`
Weights: Density=35%, Power=20%, Cooling Redundancy=15%, Cooling Mode=30%

| TC# | Description | Premiums (D, P, CR, CM) | Expected QP |
|---|---|---|---|
| F7-01 | All base (minimum) | 1.00, 1.00, 1.00, 1.00 | 1.0000 |
| F7-02 | All max | 1.30, 1.14, 1.14, 1.07 | 1.30x0.35 + 1.14x0.20 + 1.14x0.15 + 1.07x0.30 = 1.1750 |
| F7-03 | Mixed typical | 1.12, 1.14, 1.03, 1.07 | 1.12x0.35 + 1.14x0.20 + 1.03x0.15 + 1.07x0.30 = 1.0955 |
| F7-04 | Weights sum to 1.0 | — | 0.35+0.20+0.15+0.30 = 1.00 |

---

### F8. Quality-Weighted Value for Row (Row.QV)

**Formula:** `Row.QV = Row.CCAC x Reference_kW_Rate x Row_QP`

| TC# | Description | Input (CCAC, Rate, QP) | Expected Row.QV |
|---|---|---|---|
| F8-01 | Basic calculation | 200kW, $180, 1.0955 | $39,438.00 |
| F8-02 | Base quality (QP=1.0) | 100kW, $180, 1.0 | $18,000.00 |
| F8-03 | Zero CCAC | 0kW, $180, 1.12 | $0.00 |
| F8-04 | Custom rate | 200kW, $200, 1.05 | $42,000.00 |

---

### F9. Attribute Coherence Factor

**Rule:** Count distinct tiers across rows within CC for 1 dimension

| TC# | Description | Tiers across rows | Expected CF |
|---|---|---|---|
| F9-01 | All same | [Tier3, Tier3, Tier3] | 1.00 |
| F9-02 | Two different | [Tier3, Tier3, Tier2] | 0.98 |
| F9-03 | Three different | [Tier1, Tier3, Tier5] | 0.92 |
| F9-04 | Four different | [Tier1, Tier2, Tier3, Tier4] | 0.92 |
| F9-05 | Single row CC | [Tier3] | 1.00 |

---

### F10. CC Coherence Factor

**Formula:** `CC_CF = Σ(Attribute_CF x Attribute_Weight)`

| TC# | Description | CFs (D, P, CR, CM) | Expected CC_CF |
|---|---|---|---|
| F10-01 | Fully coherent | 1.0, 1.0, 1.0, 1.0 | 1.0000 |
| F10-02 | One dimension mixed | 0.98, 1.0, 1.0, 1.0 | 0.9930 |
| F10-03 | All dimensions 2-tier mixed | 0.98, 0.98, 0.98, 0.98 | 0.9800 |
| F10-04 | All dimensions 3+ tiers | 0.92, 0.92, 0.92, 0.92 | 0.9200 |
| F10-05 | Mixed coherence | 1.0, 0.98, 1.0, 0.92 | 1.0x0.35 + 0.98x0.20 + 1.0x0.15 + 0.92x0.30 = 0.972 |

---

### F11. Quality-Weighted Value for Capacity Cell (CC.QV)

**Formula:** `CC.QV = Σ(Row.QV) x CC_CF`

| TC# | Description | Input | Expected CC.QV |
|---|---|---|---|
| F11-01 | Coherent CC (CF=1.0) | Row values: $39,438 + $39,438. CF=1.0 | $78,876 |
| F11-02 | Mixed CC (CF=0.983) | Row values: $39,438 + $39,438. CF=0.983 | $77,534 |
| F11-03 | Single row | Row value: $50,000. CF=1.0 | $50,000 |
| F11-04 | Discount applied | Sum=$100,000. CF=0.92 | $92,000 |

---

### F12. Quality-Weighted Value for Rack (Rack.QV)

**Formula:** `Rack.QV = CC.QV x (Rack.RCC_design / CC.Total_RCC_design)`

| TC# | Description | Input | Expected Rack.QV |
|---|---|---|---|
| F12-01 | Equal racks | CC.QV=$78,876. 12 racks x 40kW. CC total=480 | $6,573 per rack |
| F12-02 | Sum check | All rack QV sum | = CC.QV |
| F12-03 | Unequal racks | CC.QV=$60,000. Racks: 20kW, 40kW. Total=60 | $20,000 and $40,000 |

---

### F13. Effective SBO Count (Scenario-Dependent)

**Rule:** Partially reserved composite SBOs contribute 0 to count. Free or fully reserved contribute 1.

| TC# | Description | Input | Expected Count |
|---|---|---|---|
| F13-01 | Whitespace (all free) | Base SBO has total_sbo_count=3 in Sw | 3 |
| F13-02 | One composite partially reserved | total_sbo_count=3, 1 composite partially reserved | 2 |
| F13-03 | All composites partially reserved | total_sbo_count=3, 2 composites partially reserved | 1 (self only) |
| F13-04 | Composite fully reserved (not partial) | 1 composite fully reserved | Still counts: contributes 1 |

---

### F14. Entanglement Discount

**Rule:** Penalty for cross-room spanning. Reward for in-room modularity.

| TC# | Description | Input | Expected |
|---|---|---|---|
| F14-01 | CC within 1 room — no penalty | CC in single room | Discount = 1.0 (no penalty) |
| F14-02 | CC spans 2 rooms | CC straddles 2 rooms | Discount < 1.0 (penalty) |
| F14-03 | Multiple CCs in same room (reward) | 3 CCs in Room-1 | Modularity boost > 1.0 |
| F14-04 | Room Bundle spans 2 rooms | RB crosses 2 rooms | Penalty |
| F14-05 | UPS Bundle in single floor (reward) | 2 UPS Bundles in Floor-1 | Floor modularity boost |

---

### F15. Adjusted Premium Rate (APR)

**Formula:** `APR = SPR x Entanglement_Discount x Composite_Boost x RST_Effect`

| TC# | Description | Input | Expected APR |
|---|---|---|---|
| F15-01 | CC free, no discount, no boost | SPR=5%, discount=1.0, boost=1.0, RST=free | 5% |
| F15-02 | CC partially reserved | SPR=5%, RST=partially reserved | **0%** |
| F15-03 | CC with composite boost | SPR=5%, boost=1.5 | 7.5% |
| F15-04 | CC with entanglement penalty | SPR=5%, discount=0.8 | 4% |
| F15-05 | Room Bundle partially reserved | SPR=2%, RST=partially reserved | **0%** |
| F15-06 | All factors combined | SPR=3%, discount=0.9, boost=1.2, RST=free | 3.24% |

---

### F16. Rack OAKw (Final Value)

**Formula:** `Rack.OAKw = Rack.QV x (1+APR_CC) x (1+APR_RB) x (1+APR_RPDU) x (1+APR_UPS)`

| TC# | Description | Input | Expected |
|---|---|---|---|
| F16-01 | Whitespace, all premiums active | QV=$6,573. APR: CC=5%, RB=2%, RPDU=3%, UPS=3% | $6,573 x 1.05 x 1.02 x 1.03 x 1.03 = **$7,485** |
| F16-02 | All premiums zero (free-floating) | QV=$6,573. All APR=0 | **$6,573** |
| F16-03 | CC partially reserved | QV=$6,573. APR_CC=0, others=2%,3%,3% | $6,573 x 1.0 x 1.02 x 1.03 x 1.03 = $7,118 |
| F16-04 | Multiplicative check (not additive) | QV=$10,000. All APR=10% | $10,000 x 1.1^4 = $14,641 (NOT $14,000) |

---

### F17. Remaining Portfolio Value (RPV)

**Rule:** Sum of Rack.OAKw for free racks in scenario.

| TC# | Description | Scenario | Expected |
|---|---|---|---|
| F17-01 | Whitespace (all free) | Sw: 100 racks, each OAKw=$7,000 | RPV = $700,000 |
| F17-02 | Planned (some reserved) | S0: 60 free, 40 reserved. Free OAKw=$7,000 | RPV = $420,000 |
| F17-03 | Draft scenario | S_PD: 50 free, 50 reserved | RPV = sum of 50 free rack OAKw |
| F17-04 | All reserved | 0 free racks | RPV = $0 |

---

### F18. Captured Capacity Value (CCV)

**Rule:** Sum of Rack.OAKw for reserved racks per Placement Plan/Draft.

| TC# | Description | Input | Expected |
|---|---|---|---|
| F18-01 | Single PP | PP-01 has 36 reserved racks | CCV = sum of 36 rack OAKw |
| F18-02 | Multiple PPs | PP-01=20 racks, PP-02=16 racks | CCV_PP01 and CCV_PP02 separate |
| F18-03 | Draft CCV | PD-01 has 10 reserved racks | CCV uses S_PD APRs |
| F18-04 | No reservations | 0 reserved racks | CCV = $0 |

---

### F19. Average CCV per Rack

**Formula:** `Avg CCV/Rack = CCV / number_of_reserved_racks`

| TC# | Description | Input | Expected |
|---|---|---|---|
| F19-01 | Basic | CCV=$778,140, 18 racks | $43,230 |
| F19-02 | Zero racks | CCV=$0, 0 racks | $0 or N/A (avoid div by zero) |

---

### F20. Optionality Impact

**Formula:** `Impact = Σ(Rack.R_OAKw_before) - Σ(Rack.R_OAKw_after)`

| TC# | Description | Expected |
|---|---|---|
| F20-01 | Placement breaks CC containment | Impact negative (value destroyed) |
| F20-02 | Placement fills entire CC | Impact minimal (containment preserved) |
| F20-03 | No placement change | Impact = 0 |

---

### F21. Facility OAKw Rate

**Formula:** `Rate = IT_Facility.OAKw / Facility_Deployable_ITCapacity`

| TC# | Description | Input | Expected |
|---|---|---|---|
| F21-01 | Basic rate | OAKw=$25M, capacity=100,000kW | $250/kW |
| F21-02 | Whitespace only | All racks free | Rate reflects maximum optionality |

---

### F22. CC Capacity Envelope

**Formula:** `Envelope = MIN(Power_Capacity, Thermal_Capacity)`
- Power = Σ(RPP_Panel_Capacity_design)
- Thermal = Σ(AirZone_Cooling_Capacity) + Σ(LL_Cooling_Capacity)

| TC# | Description | Power | Thermal | Expected Envelope |
|---|---|---|---|---|
| F22-01 | Power constrained | 400kW | 500kW | 400kW |
| F22-02 | Thermal constrained | 500kW | 350kW | 350kW |
| F22-03 | Equal | 400kW | 400kW | 400kW |
| F22-04 | No cooling (air zone + LL = 0) | 500kW | 0kW | 0kW |

---

### F23. Rollup Rack Circuit Capacity

**Rule:** Sum RCC_design for all racks within a group (row, CC).

| TC# | Description | Input | Expected |
|---|---|---|---|
| F23-01 | Row rollup | 6 racks x 40kW | 240kW |
| F23-02 | CC rollup | 2 rows: 240kW + 240kW | 480kW |
| F23-03 | Mixed rack sizes | 10, 20, 30, 40 kW | 100kW |

---

### F24. Reservation Status (RST)

**Rule:** Derived from rack-level reservations, cascaded to SBOs.

| TC# | Description | Rack States | Expected SBO RST |
|---|---|---|---|
| F24-01 | All racks free | 12/12 free | CC RST = Free |
| F24-02 | All racks reserved to same PP | 12/12 reserved to PP-01 | CC RST = Reserved |
| F24-03 | Mixed — some free, some reserved | 6 free, 6 reserved | CC RST = **Partially Reserved** |
| F24-04 | All reserved to different PPs | 6 to PP-01, 6 to PP-02 | CC RST = **Partially Reserved** |
| F24-05 | Cascade to Room Bundle | CC-01 partially reserved, CC-02 free | RB RST = Partially Reserved |

---

### F25. Reservation Conflict Status (RCST)

**Rule:** Cross-scenario conflict detection for racks.

| TC# | Description | Input | Expected RCST |
|---|---|---|---|
| F25-01 | Rack in PD, also in PP | Rack reserved in PD-01 and PP-01 | **Hard Conflict** |
| F25-02 | Rack in 2 PDs | Rack reserved in PD-01 and PD-02 | **Soft Conflict** |
| F25-03 | Rack in PP only | Rack reserved only in PP-01 | **None** |
| F25-04 | Rack in PD, no overlap | Rack reserved only in PD-01 | **None** |
| F25-05 | PD with hard conflict → cannot become PP | PD has rack with hard conflict | Block PD→PP promotion |

---

## Summary

| Phase | Features | Test Cases | Needs Reservation? |
|---|---|---|---|
| A (Steps 1-6) | F1-F12 | 43 TCs | No |
| B (Steps 7-8) | F13-F16 | 19 TCs | Yes |
| C (Steps 9-10) | F17-F21 | 13 TCs | Yes |
| D (Support) | F22-F25 | 17 TCs | F24-F25 Yes |
| **Total** | **25 features** | **92 TCs** | — |
